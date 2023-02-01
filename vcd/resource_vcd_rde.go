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
			"rde_type_vendor": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The Runtime Defined Entity Type vendor",
			},
			"rde_type_namespace": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The Runtime Defined Entity Type namespace",
			},
			"rde_type_version": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The Runtime Defined Entity Type version",
			},
			"external_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "An external entity's ID that this Runtime Defined Entity may have a relation to",
			},
			"entity_url": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "URL that should point to a JSON representation of the Runtime Defined Entity. The JSON will be validated against the schema of the RDE type that the entity is an instance of",
				ExactlyOneOf: []string{"entity_url", "entity"},
			},
			"entity": {
				Type:                  schema.TypeString,
				Optional:              true,
				Computed:              true,
				Description:           "A JSON representation of the Runtime Defined Entity. The JSON will be validated against the schema of the RDE type that the entity is an instance of",
				ExactlyOneOf:          []string{"entity_url", "entity"},
				DiffSuppressFunc:      hasJsonValueChanged,
				DiffSuppressOnRefresh: true,
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
			"state": {
				Type:        schema.TypeString,
				Description: "If the specified JSON in either `entity` or `entity_url` is correct, the state will be RESOLVED, otherwise it will be RESOLUTION_ERROR. If an entity in an RESOLUTION_ERROR state, it will require to be updated to a correct JSON to be usable",
				Computed:    true,
			},
			"metadata_entry": getOpenApiMetadataEntrySchema("Runtime Defined Entity", false),
		},
	}
}

func resourceVcdRdeCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	jsonSchema, err := getRdeJson(d)
	if err != nil {
		return diag.FromErr(err)
	}

	name := d.Get("name").(string)
	vendor := d.Get("rde_type_vendor").(string)
	nss := d.Get("rde_type_namespace").(string)
	version := d.Get("rde_type_version").(string)

	// VCD allows to have multiple RDEs with the same name, but this is not compatible with Terraform as there is no
	// other way to unequivocally identify a RDE from a given type.
	// In other words, without this check, the data source could be potentially broken.
	rdes, err := vcdClient.GetRdesByName(vendor, nss, version, name)
	if err == nil && rdes != nil {
		rdeList := make([]string, len(rdes))
		for i, rde := range rdes {
			rdeList[i] = rde.DefinedEntity.ID
		}
		return diag.Errorf("found other Runtime Defined Entities with same name: %v", rdeList)
	}
	if err != nil && !govcd.ContainsNotFound(err) {
		return diag.Errorf("could not create an RDE, failed fetching existing RDEs: %s", err)
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

	rde, err := vcdClient.CreateRde(vendor, nss, version, types.DefinedEntity{
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
func getRdeJson(d *schema.ResourceData) (map[string]interface{}, error) {
	var jsonRde string
	var err error
	if url, isUrlSet := d.GetOk("entity_url"); isUrlSet {
		jsonRde, err = fileFromUrlToString(url.(string), ".json")
		if err != nil {
			return nil, fmt.Errorf("could not download JSON RDE from url %s: %s", url, err)
		}
	} else {
		jsonRde = d.Get("entity").(string)
	}

	var unmarshalledJson map[string]interface{}
	err = json.Unmarshal([]byte(jsonRde), &unmarshalledJson)
	if err != nil {
		return nil, err
	}

	return unmarshalledJson, err
}

func resourceVcdRdeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericVcdRdeRead(ctx, d, meta, "resource")
}

// genericVcdRdeRead reads a Runtime Defined Entity from VCD and sets the Terraform state accordingly.
// If origin == "datasource", if the referenced RDE type doesn't exist, it errors.
// If origin == "resource", if the referenced RDE type doesn't exist, it removes it from tfstate and exits normally.
func genericVcdRdeRead(_ context.Context, d *schema.ResourceData, meta interface{}, origin string) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	rde, err := getRde(d, vcdClient)
	if origin == "resource" && govcd.ContainsNotFound(err) {
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
	err = d.Set("entity", jsonEntity)
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

	vendor := d.Get("rde_type_vendor").(string)
	nss := d.Get("rde_type_namespace").(string)
	version := d.Get("rde_type_version").(string)
	name := d.Get("name").(string)

	rdes, err := vcdClient.GetRdesByName(vendor, nss, version, name)
	if err != nil {
		return nil, err
	}

	// We return the first found RDE as a design decision. Ideally, we should only find more than one RDE with the same
	// name during imports, where Terraform doesn't have any control.
	return rdes[0], nil
}

func resourceVcdRdeUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	rde, err := getRde(d, vcdClient)
	if govcd.ContainsNotFound(err) {
		log.Printf("[DEBUG] Runtime Defined Entity no longer exists. Removing from tfstate")
		return nil
	}
	if err != nil {
		return diag.FromErr(err)
	}
	jsonEntity, err := getRdeJson(d)
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
	if govcd.ContainsNotFound(err) {
		log.Printf("[DEBUG] Runtime Defined Entity no longer exists. Removing from tfstate")
		return nil
	}
	if err != nil {
		return diag.FromErr(err)
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
'vendor.namespace.version.name.position' where position is the RDE number as returned by VCD, starting on 1
'list@vendor.namespace.version.name' to get a list of RDEs with their respective positions and real IDs`)

	printList := func(vendor, namespace, version, name string) error {
		rdes, err := vcdClient.GetRdesByName(vendor, namespace, version, name)
		if err != nil {
			return err
		}
		fmt.Printf("Found RDEs with vendor '%s', namespace '%s', version '%s' and name '%s':\n", vendor, namespace, version, name)
		for _, rde := range rdes {
			fmt.Printf("* %s\n", rde.DefinedEntity.ID)
		}
		return fmt.Errorf("resource was not imported! %s", helpError.Error())
	}

	getRdeInPosition := func(vendor, namespace, version, name, position string) (*govcd.DefinedEntity, error) {
		rdes, err := vcdClient.VCDClient.GetRdesByName(vendor, namespace, version, name)
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
	case 1: // ie: urn:vcloud:entity:vendor:namespace:a074f9e9-5d76-4f1e-8c37-f4e8b28e51ff
		rde, err = vcdClient.VCDClient.GetRdeById(resourceURI[0])
		if err != nil {
			return nil, err
		}
	case 4: // ie: list@vendor.namespace.1.2.3.name
		listAndVendorSplit := strings.Split(resourceURI[0], "@")
		if len(listAndVendorSplit) != 2 {
			return nil, helpError
		}
		return nil, printList(listAndVendorSplit[1], resourceURI[1], resourceURI[2], resourceURI[3])
	case 5: // ie: VCD_IMPORT_SEPARATOR="_" vendor_namespace_1.2.3_name_1
		rde, err = getRdeInPosition(resourceURI[0], resourceURI[1], resourceURI[2], resourceURI[3], resourceURI[4])
		if err != nil {
			return nil, err
		}
	case 6: // ie: VCD_IMPORT_SEPARATOR="_" list@vendor_namespace_1.2.3_name
		listAndVendorSplit := strings.Split(resourceURI[0], "@")
		if len(listAndVendorSplit) != 2 {
			return nil, helpError
		}
		return nil, printList(listAndVendorSplit[1], resourceURI[1], fmt.Sprintf("%s.%s.%s", resourceURI[2], resourceURI[3], resourceURI[4]), resourceURI[5])
	case 7: // ie: vendor.namespace.1.2.3.name.1
		rde, err = getRdeInPosition(resourceURI[0], resourceURI[1], fmt.Sprintf("%s.%s.%s", resourceURI[2], resourceURI[3], resourceURI[4]), resourceURI[5], resourceURI[6])
		if err != nil {
			return nil, err
		}
	default:
		return nil, helpError
	}

	d.SetId(rde.DefinedEntity.ID)
	rdeTypeParts := strings.Split(rde.DefinedEntity.EntityType, ":")
	if len(rdeTypeParts) < 5 {
		return nil, fmt.Errorf("error in recovered RDE, its Type is incorrect: %s", rde.DefinedEntity.EntityType)
	}
	dSet(d, "rde_type_vendor", rdeTypeParts[3])
	dSet(d, "rde_type_namespace", rdeTypeParts[4])
	dSet(d, "rde_type_version", rdeTypeParts[5])

	return []*schema.ResourceData{d}, nil
}
