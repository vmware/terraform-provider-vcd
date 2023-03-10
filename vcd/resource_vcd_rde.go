package vcd

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"github.com/vmware/go-vcloud-director/v2/util"
	"log"
	"strconv"
	"strings"
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
				Description: "The name of the Runtime Defined Entity",
			},
			"rde_type_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The Runtime Defined Entity Type ID",
			},
			"external_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
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
			"owner_id": {
				Type:        schema.TypeString,
				Description: "The owner of the Runtime Defined Entity",
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
					"resolved and must be either done by an external component or with an update. The Runtime Defined Entity can't be" +
					"deleted until the entity is resolved.",
				Required: true,
			},
			"resolve_on_removal": {
				Type:        schema.TypeBool,
				Description: "If `true`, the Runtime Defined Entity will be resolved before it gets deleted, to forcefully delete it. Otherwise, destroy will fail if it is not resolved.",
				Default:     false,
				Optional:    true,
			},
			"state": {
				Type: schema.TypeString,
				// Todo: Add PRE_CREATED
				Description: "If the specified JSON in either `input_entity` or `input_entity_url` is correct, the state will be RESOLVED, otherwise it will be RESOLUTION_ERROR. If an entity in an RESOLUTION_ERROR state, it will require to be updated to a correct JSON to be usable",
				Computed:    true,
			},
			// TODO: entity_in_sync
			"metadata_entry": getOpenApiMetadataEntrySchema("Runtime Defined Entity", false),
		},
	}
}

func resourceVcdRdeCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	name := d.Get("name").(string)
	rdeTypeId := d.Get("rde_type_id").(string)

	rdeType, err := vcdClient.GetRdeTypeById(rdeTypeId)
	if err != nil {
		return diag.Errorf("could not retrieve RDE Type with ID %s", rdeTypeId)
	}

	tenantContext := govcd.TenantContext{}
	if vcdClient.Client.IsSysAdmin {
		org, err := vcdClient.GetAdminOrgFromResource(d)
		if err != nil {
			return diag.Errorf("error retrieving org %s: %s", d.Get("org").(string), err)
		}
		tenantContext.OrgId = org.AdminOrg.ID
		tenantContext.OrgName = org.AdminOrg.Name
	}

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
		return diag.Errorf("could not create the Runtime Defined Entity: %s", err)
	}

	// We save the ID immediately as the Resolve operation can fail, but the RDE is already created. If this happens,
	// it should go to the Update operation instead.
	d.SetId(rde.DefinedEntity.ID)

	if d.Get("resolve").(bool) {
		err = rde.Resolve()
		if err != nil {
			return diag.Errorf("could not resolve the Runtime Defined Entity: %s", err)
		}
	}

	// Metadata is only supported since v37.0
	if vcdClient.Client.APIVCDMaxVersionIs(">= 37.0") {
		err = createOrUpdateOpenApiMetadataEntryInVcd(d, rde)
		if err != nil {
			return diag.Errorf("could not create metadata for the Runtime Defined Entity: %s", err)
		}
	} else if _, ok := d.GetOk("metadata_entry"); ok {
		return diag.Errorf("metadata_entry is only supported since VCD 10.4.0")
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
	rde, err := getRde(d, vcdClient)
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
		util.Logger.Printf("[DEBUG] RDE %s is not in RESOLVED state", rde.DefinedEntity.Name)
	}

	jsonEntity, err := jsonToCompactString(rde.DefinedEntity.Entity)
	if err != nil {
		return diag.Errorf("could not save the Runtime Defined Entity JSON into state: %s", err)
	}
	err = d.Set("computed_entity", jsonEntity)
	if err != nil {
		return diag.FromErr(err)
	}

	if rde.DefinedEntity.Org != nil {
		dSet(d, "org_id", rde.DefinedEntity.Org.ID)
	}
	if rde.DefinedEntity.Owner != nil {
		dSet(d, "owner_id", rde.DefinedEntity.Owner.ID)
	}

	// Metadata is only available since API v37.0
	if vcdClient.Client.APIVCDMaxVersionIs(">= 37.0") {
		err = updateOpenApiMetadataInState(d, rde)
		if err != nil {
			return diag.Errorf("could not set metadata for the Runtime Defined Entity: %s", err)
		}
	}

	d.SetId(rde.DefinedEntity.ID)

	return nil
}

// getRde retrieves a Runtime Defined Entity from VCD with the required attributes from the Terraform config.
func getRde(d *schema.ResourceData, vcdClient *VCDClient) (*govcd.DefinedEntity, error) {
	if d.Id() != "" {
		return vcdClient.GetRdeById(d.Id())
	}

	rdeTypeId := d.Get("rde_type_id").(string)
	name := d.Get("name").(string)

	rdeType, err := vcdClient.GetRdeTypeById(rdeTypeId)
	if err != nil {
		return nil, fmt.Errorf("could not get RDE Type with ID %s", rdeTypeId)
	}

	rdes, err := rdeType.GetRdesByName(name)
	if err != nil {
		return nil, fmt.Errorf("could not get RDE with name %s and RDE Type ID %s", name, rdeTypeId)
	}

	// We perform another GET by ID to retrieve the ETag of the RDE.
	// We pick the first retrieved RDE from the result above.
	return vcdClient.GetRdeById(rdes[0].DefinedEntity.ID)
}

func resourceVcdRdeUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	rde, err := getRde(d, vcdClient)
	if err != nil {
		return diag.FromErr(err)
	}
	jsonEntity, err := getRdeJson(vcdClient, d)
	if err != nil {
		return diag.FromErr(err)
	}

	err = rde.Update(types.DefinedEntity{
		Name:       d.Get("name").(string),
		ExternalId: d.Get("external_id").(string),
		Entity:     jsonEntity,
	})
	if err != nil {
		return diag.Errorf("could not update the Runtime Defined Entity: %s", err)
	}

	if d.Get("resolve").(bool) {
		err = rde.Resolve()
		if err != nil {
			return diag.Errorf("could not resolve the Runtime Defined Entity: %s", err)
		}
	}

	// Metadata is only supported since v37.0
	if vcdClient.Client.APIVCDMaxVersionIs(">= 37.0") {
		err = createOrUpdateOpenApiMetadataEntryInVcd(d, rde)
		if err != nil {
			return diag.Errorf("could not create metadata for the Runtime Defined Entity: %s", err)
		}
	} else if _, ok := d.GetOk("metadata_entry"); ok {
		return diag.Errorf("metadata_entry is only supported since VCD 10.4.0")
	}

	return resourceVcdRdeRead(ctx, d, meta)
}

func resourceVcdRdeDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	rde, err := getRde(d, vcdClient)
	if err != nil {
		return diag.FromErr(err)
	}

	if d.Get("resolve_on_removal").(bool) {
		err = rde.Resolve()
		if err != nil {
			return diag.FromErr(err)
		}
	}

	err = rde.Delete()
	if err != nil {
		return diag.Errorf("could not delete the Runtime Defined Entity: %s", err)
	}
	return nil
}

// resourceVcdRdeImport is responsible for importing the resource.
// The following steps happen as part of import
// 1. The user supplies `terraform import _resource_name_ _the_id_string_` command
// 2. `_the_id_string_` contains a dot formatted path to resource as in the example below
// 3. The functions splits the dot-formatted path and tries to lookup the object
// 4. If the lookup succeeds it set's the ID field for `_resource_name_` resource in state file
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
	case 4: // ie: list@vendor.nss.1.2.3.name
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
	case 6: // ie: VCD_IMPORT_SEPARATOR="_" list@vendor_nss_1.2.3_name
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
	dSet(d, "rde_type_id", rde.DefinedEntity.EntityType)

	return []*schema.ResourceData{d}, nil
}
