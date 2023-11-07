package vcd

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"github.com/vmware/go-vcloud-director/v2/util"
)

func resourceVcdRde() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdRdeCreate,
		ReadContext:   resourceVcdRdeRead,
		UpdateContext: resourceVcdRdeUpdate,
		DeleteContext: resourceVcdRdeDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdRdeImport,
		},
		Schema: map[string]*schema.Schema{
			"org": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Description: "The name of organization that will own this Runtime Defined Entity, optional if defined at provider " +
					"level. Useful when connected as sysadmin working across different organizations",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the Runtime Defined Entity. It can be non-unique",
			},
			"rde_type_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The Runtime Defined Entity Type ID",
			},
			"external_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true, // It can be populated by a 3rd party
				Description: "An external entity's ID that this Runtime Defined Entity may have a relation to",
			},
			"input_entity_url": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "URL that should point to a JSON representation of the Runtime Defined Entity and is used to initialize/override its contents",
				ExactlyOneOf: []string{"input_entity_url", "input_entity"},
			},
			"input_entity": {
				Type:                  schema.TypeString,
				Optional:              true,
				Description:           "A JSON representation of the Runtime Defined Entity that is defined by the user and is used to initialize/override its contents",
				ExactlyOneOf:          []string{"input_entity_url", "input_entity"},
				DiffSuppressFunc:      hasJsonValueChanged,
				DiffSuppressOnRefresh: true,
			},
			"computed_entity": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "A computed representation of the actual Runtime Defined Entity JSON retrieved from VCD. Useful to see the actual entity contents if it is being changed by a third party in VCD",
			},
			"owner_user_id": {
				Type:        schema.TypeString,
				Description: "The ID of the user that owns the Runtime Defined Entity",
				Computed:    true,
			},
			"org_id": {
				Type:        schema.TypeString,
				Description: "The organization of the Runtime Defined Entity",
				Computed:    true,
			},
			"resolve": {
				Type: schema.TypeBool,
				Description: "If `true`, the Runtime Defined Entity will be resolved by this provider. If `false`, it won't be" +
					"resolved and must be done either by an external component action or by an update. The Runtime Defined Entity can't be" +
					"deleted until the entity is resolved.",
				Required: true,
			},
			"resolve_on_removal": {
				Type: schema.TypeBool,
				Description: "If `true`, the Runtime Defined Entity will be resolved before it gets deleted, to ensure forced deletion." +
					"Destroy will fail if it is not resolved.",
				Default:  false,
				Optional: true,
			},
			"state": {
				Type: schema.TypeString,
				Description: "Specifies whether the entity is correctly resolved or not. When created it will be in PRE_CREATED state. If the entity is correctly validated against its RDE Type schema, the state will be RESOLVED," +
					"otherwise it will be RESOLUTION_ERROR. If an entity resolution ends in a RESOLUTION_ERROR state, it will require to be updated to a correct JSON to be usable",
				Computed: true,
			},
			"entity_in_sync": {
				Type:        schema.TypeBool,
				Description: "If true, `computed_entity` is equal to either `input_entity` or the contents of `input_entity_url`",
				Computed:    true,
			},
		},
	}
}

func resourceVcdRdeCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	name := d.Get("name").(string)
	rdeTypeId := d.Get("rde_type_id").(string)

	rdeType, err := vcdClient.GetRdeTypeById(rdeTypeId)
	if err != nil {
		return diag.Errorf("could not create RDE with name '%s', could not retrieve RDE Type with ID '%s': %s", name, rdeTypeId, err)
	}

	tenantContext := govcd.TenantContext{}
	org, err := vcdClient.GetOrgFromResource(d)
	if err != nil {
		return diag.Errorf("could not create RDE with name '%s', error retrieving Org: %s", name, err)
	}
	tenantContext.OrgId = org.Org.ID
	tenantContext.OrgName = org.Org.Name

	jsonSchema, err := getRdeJson(vcdClient, d)
	if err != nil {
		return diag.FromErr(err)
	}

	rde, err := rdeType.CreateRde(types.DefinedEntity{
		Name:       name,
		ExternalId: d.Get("external_id").(string),
		Entity:     jsonSchema,
	}, &tenantContext)
	if err != nil {
		return diag.Errorf("could not create the Runtime Defined Entity '%s' of type '%s' in Organization '%s': %s", name, rdeTypeId, org.Org.Name, err)
	}

	// We save the ID immediately as the Resolve operation can fail but the RDE is already created.
	d.SetId(rde.DefinedEntity.ID)

	if d.Get("resolve").(bool) {
		err = rde.Resolve()
		if err != nil {
			return diag.Errorf("could not resolve the Runtime Defined Entity '%s' of type '%s' in Organization '%s': %s", name, rdeTypeId, org.Org.Name, err)
		}
	}

	return resourceVcdRdeRead(ctx, d, meta)
}

// getRdeJson gets the RDE as JSON from the Terraform configuration
func getRdeJson(vcdClient *VCDClient, d *schema.ResourceData) (map[string]interface{}, error) {
	var jsonRde string
	var err error
	if url, isUrlSet := d.GetOk("input_entity_url"); isUrlSet {
		jsonRde, err = fileFromUrlToString(vcdClient, url.(string), ".json")
		if err != nil {
			return nil, fmt.Errorf("could not download JSON RDE from url %s: %s", url, err)
		}
	} else {
		jsonRde = d.Get("input_entity").(string)
	}

	var unmarshalledJson map[string]interface{}
	err = json.Unmarshal([]byte(jsonRde), &unmarshalledJson)
	if err != nil {
		return nil, err
	}

	return unmarshalledJson, err
}

func resourceVcdRdeRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	rde, err := getRde(d, vcdClient, "resource")
	if govcd.ContainsNotFound(err) {
		log.Printf("[DEBUG] Runtime Defined Entity no longer exists. Removing from tfstate")
		d.SetId("")
		return nil
	}
	if err != nil {
		return diag.FromErr(err)
	}

	dSet(d, "name", rde.DefinedEntity.Name)
	dSet(d, "external_id", rde.DefinedEntity.ExternalId)
	dSet(d, "state", rde.DefinedEntity.State)

	if rde.DefinedEntity.State != nil && *rde.DefinedEntity.State != "RESOLVED" {
		util.Logger.Printf("[DEBUG] RDE '%s' is not in RESOLVED state", rde.DefinedEntity.ID)
	}

	jsonEntity, err := jsonToCompactString(rde.DefinedEntity.Entity)
	if err != nil {
		return diag.Errorf("could not save the RDE '%s' JSON into state: %s", rde.DefinedEntity.ID, err)
	}
	err = d.Set("computed_entity", jsonEntity)
	if err != nil {
		return diag.FromErr(err)
	}

	if rde.DefinedEntity.Org != nil {
		dSet(d, "org_id", rde.DefinedEntity.Org.ID)
	}
	if rde.DefinedEntity.Owner != nil {
		dSet(d, "owner_user_id", rde.DefinedEntity.Owner.ID)
	}

	dSet(d, "entity_in_sync", false)
	// These fields can be empty on imports
	if d.Get("input_entity_url") != "" || d.Get("input_entity") != "" {
		inputJson, err := getRdeJson(vcdClient, d)
		if err != nil {
			return diag.Errorf("error getting JSON from RDE '%s' configuration: %s", rde.DefinedEntity.ID, err)
		}
		inputJsonMarshaled, err := json.Marshal(inputJson)
		if err != nil {
			return diag.Errorf("error marshaling JSON retrieved from RDE '%s' configuration: %s", rde.DefinedEntity.ID, err)
		}
		areJsonEqual, err := areMarshaledJsonEqual([]byte(jsonEntity), inputJsonMarshaled)
		if err != nil {
			return diag.Errorf("error comparing %s with %s of RDE '%s': %s", jsonEntity, inputJsonMarshaled, rde.DefinedEntity.ID, err)
		}
		dSet(d, "entity_in_sync", areJsonEqual)
	}

	d.SetId(rde.DefinedEntity.ID)

	return nil
}

// getRde retrieves a Runtime Defined Entity from VCD with the required attributes from the Terraform config.
func getRde(d *schema.ResourceData, vcdClient *VCDClient, origin string) (*govcd.DefinedEntity, error) {
	if d.Id() != "" {
		return vcdClient.GetRdeById(d.Id())
	}

	rdeTypeId := d.Get("rde_type_id").(string)
	name := d.Get("name").(string)

	rdeType, err := vcdClient.GetRdeTypeById(rdeTypeId)
	if err != nil {
		return nil, fmt.Errorf("could not get RDE Type with ID '%s': %s", rdeTypeId, err)
	}

	rdes, err := rdeType.GetRdesByName(name)
	if err != nil {
		return nil, fmt.Errorf("could not get RDEs with name '%s' and RDE Type ID '%s': %s", name, rdeTypeId, err)
	}

	// As RDEs can have many instances with same name and RDE Type, we can't guarantee that we will read the one we want,
	// but at least we try to filter a bit with things we know, like Organization.
	var filteredRdes []*govcd.DefinedEntity
	orgName := d.Get("org")
	for _, rde := range rdes {
		if rde.DefinedEntity.Org != nil && orgName == rde.DefinedEntity.Org.Name {
			filteredRdes = append(filteredRdes, rde)
		}
	}

	if len(filteredRdes) == 0 {
		return nil, fmt.Errorf("no RDEs found with name '%s' and RDE Type ID '%s' in Org '%s': %s", name, rdeTypeId, orgName, govcd.ErrorEntityNotFound)
	}

	// If there is more than one RDE, we retrieve the IDs to give the user some feedback.
	var filteredRdesIds []string
	if len(filteredRdes) > 1 {
		for _, rde := range filteredRdes {
			filteredRdesIds = append(filteredRdesIds, rde.DefinedEntity.ID)
		}
	}

	err = fmt.Errorf("there are %d RDEs with name '%s' and RDE Type ID '%s' in Org '%s': %v", len(filteredRdes), name, rdeTypeId, orgName, filteredRdesIds)
	// We end early with the data source if there is more than one RDE found.
	if origin == "datasource" && len(filteredRdes) > 1 {
		return nil, err
	}

	if len(filteredRdes) > 1 {
		util.Logger.Printf("[WARN] %s: Choosing the first one", err.Error())
	}
	// We just perform another GET by ID to retrieve the ETag of the RDE (it is only returned when we get a specific and unique ID).
	// We pick the first retrieved RDE from the result above. This doesn't guarantee that is the RDE we want, but it's
	// how VCD works.
	return vcdClient.GetRdeById(filteredRdes[0].DefinedEntity.ID)
}

func resourceVcdRdeUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	rde, err := getRde(d, vcdClient, "resource")
	if err != nil {
		return diag.FromErr(err)
	}
	jsonEntity, err := getRdeJson(vcdClient, d)
	if err != nil {
		return diag.FromErr(err)
	}

	err = rde.Update(types.DefinedEntity{
		Name:       d.Get("name").(string),
		EntityType: d.Get("rde_type_id").(string),
		ExternalId: d.Get("external_id").(string),
		Entity:     jsonEntity,
	})
	if err != nil {
		return diag.Errorf("could not update the Runtime Defined Entity '%s' with ID '%s': %s", rde.DefinedEntity.Name, rde.DefinedEntity.ID, err)
	}

	if d.Get("resolve").(bool) {
		err = rde.Resolve()
		if err != nil {
			return diag.Errorf("could not resolve the Runtime Defined Entity '%s' with ID '%s': %s", rde.DefinedEntity.Name, rde.DefinedEntity.ID, err)
		}
	}

	return resourceVcdRdeRead(ctx, d, meta)
}

func resourceVcdRdeDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	rde, err := getRde(d, vcdClient, "resource")
	if err != nil {
		return diag.FromErr(err)
	}

	if d.Get("resolve_on_removal").(bool) {
		err = rde.Resolve()
		if err != nil {
			return diag.Errorf("could not resolve the Runtime Defined Entity before removal '%s' with ID '%s': %s", rde.DefinedEntity.Name, rde.DefinedEntity.ID, err)
		}
	}

	err = rde.Delete()
	if err != nil {
		return diag.Errorf("could not delete the Runtime Defined Entity '%s' with ID '%s': %s", rde.DefinedEntity.Name, rde.DefinedEntity.ID, err)
	}
	return nil
}

// resourceVcdRdeImport is responsible for importing the resource.
// The following steps happen as part of import
// 1. The user supplies `terraform import _resource_name_ _the_id_string_` command
// 2. `_the_id_string_` contains a dot formatted path to resource as in the example below
// 3. The function splits the dot-formatted path and tries to look up the object
// 4. If the lookup succeeds it sets the ID field for `_resource_name_` resource in state file
// (the resource must be already defined in .tf config otherwise `terraform import` will complain)
// 5. `terraform refresh` is being implicitly launched. The Read method looks up all other fields
// based on the known ID of object.
//
// Example resource name (_resource_name_): vcd_rde.outer-rde
// Example import path (_the_id_string_): my-rde.vmware.kubernetes.1.0.0
// Note: the separator can be changed using Provider.import_separator or variable VCD_IMPORT_SEPARATOR
func resourceVcdRdeImport(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	vcdClient := meta.(*VCDClient)
	helpError := fmt.Errorf(`resource id must be specified in one of these formats:
'rde-id' to import by RDE id
'vendor.nss.version.name.position' where position is the RDE number as returned by VCD, starting on 1
'list@vendor.nss.version.name' to get a list of RDEs with their respective positions and real IDs`)

	printList := func(vendor, nss, version, name string) error {
		rdes, err := vcdClient.GetRdesByName(vendor, nss, version, name)
		if err != nil {
			return err
		}
		fmt.Printf("Found RDEs with vendor '%s', nss '%s', version '%s' and name '%s':\n", vendor, nss, version, name)
		for _, rde := range rdes {
			fmt.Printf("* %s\n", rde.DefinedEntity.ID)
		}
		return fmt.Errorf("resource was not imported! %s", helpError.Error())
	}

	getRdeInPosition := func(vendor, nss, version, name, position string) (*govcd.DefinedEntity, error) {
		rdes, err := vcdClient.VCDClient.GetRdesByName(vendor, nss, version, name)
		if err != nil {
			return nil, err
		}
		idx, err := strconv.Atoi(position)
		if err != nil {
			return nil, fmt.Errorf("introduced position %v is not a number", position)
		}
		if idx < 1 || idx > len(rdes) {
			return nil, fmt.Errorf("introduced position %d is outside the range of RDEs retrieved: [1, %d]", idx, len(rdes))
		}
		return rdes[idx-1], nil
	}

	log.Printf("[DEBUG] importing vcd_rde resource with provided id %s", d.Id())

	resourceURI := strings.Split(d.Id(), ImportSeparator)
	var rde *govcd.DefinedEntity
	var err error
	switch len(resourceURI) {
	case 1: // ie: urn:vcloud:entity:vendor:nss:a074f9e9-5d76-4f1e-8c37-f4e8b28e51ff
		rde, err = vcdClient.VCDClient.GetRdeById(resourceURI[0])
		if err != nil {
			return nil, err
		}
	case 4: // ie: VCD_IMPORT_SEPARATOR="_" list@vendor_nss_1.2.3_name
		listAndVendorSplit := strings.Split(resourceURI[0], "@")
		if len(listAndVendorSplit) != 2 {
			return nil, helpError
		}
		return nil, printList(listAndVendorSplit[1], resourceURI[1], resourceURI[2], resourceURI[3])
	case 5: // ie: VCD_IMPORT_SEPARATOR="_" vendor_nss_1.2.3_name_1
		rde, err = getRdeInPosition(resourceURI[0], resourceURI[1], resourceURI[2], resourceURI[3], resourceURI[4])
		if err != nil {
			return nil, err
		}
	case 6: // ie: list@vendor.nss.1.2.3.name
		listAndVendorSplit := strings.Split(resourceURI[0], "@")
		if len(listAndVendorSplit) != 2 {
			return nil, helpError
		}
		return nil, printList(listAndVendorSplit[1], resourceURI[1], fmt.Sprintf("%s.%s.%s", resourceURI[2], resourceURI[3], resourceURI[4]), resourceURI[5])
	case 7: // ie: vendor.nss.1.2.3.name.1
		rde, err = getRdeInPosition(resourceURI[0], resourceURI[1], fmt.Sprintf("%s.%s.%s", resourceURI[2], resourceURI[3], resourceURI[4]), resourceURI[5], resourceURI[6])
		if err != nil {
			return nil, err
		}
	default:
		return nil, helpError
	}

	d.SetId(rde.DefinedEntity.ID)
	if rde.DefinedEntity.Org != nil {
		dSet(d, "org", rde.DefinedEntity.Org.Name)
	}
	dSet(d, "rde_type_id", rde.DefinedEntity.EntityType)

	return []*schema.ResourceData{d}, nil
}
