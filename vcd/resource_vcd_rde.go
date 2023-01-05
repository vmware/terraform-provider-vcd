package vcd

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"log"
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
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the Runtime Defined Entity",
			},
			"rde_type_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The type ID of the Runtime Defined Entity",
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
			"state": {
				Type:        schema.TypeString,
				Description: "Every Runtime Defined Entity is created in the \"PRE_CREATED\" state. Once an entity is ready to be validated against its schema, it will transition in another state - RESOLVED, if the entity is valid according to the schema, or RESOLUTION_ERROR otherwise. If an entity in an \"RESOLUTION_ERROR\" state is updated, it will transition to the inital \"PRE_CREATED\" state without performing any validation. If its in the \"RESOLVED\" state, then it will be validated against the entity type schema and throw an exception if its invalid",
				Computed:    true,
			},
			"metadata_entry": getMetadataEntrySchema("Runtime Defined Entity", false),
		},
	}
}

func resourceVcdRdeCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	jsonSchema, err := getRdeJson(d)
	if err != nil {
		return diag.FromErr(err)
	}
	rdeTypeId := d.Get("rde_type_id").(string)
	rdeType, err := vcdClient.VCDClient.GetRdeTypeById(rdeTypeId)
	if err != nil {
		return diag.Errorf("could not find any Runtime Defined Entity type with ID %s: %s", rdeTypeId, err)
	}

	_, err = rdeType.CreateRde(types.DefinedEntity{
		Name:   d.Get("name").(string),
		Entity: jsonSchema,
	})

	if err != nil {
		return diag.Errorf("could not create the Runtime Defined Entity: %s", err)
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
	rde, err := getRde(d, meta)
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

	d.SetId(rde.DefinedEntity.ID)

	return nil
}

// getRde retrieves a Runtime Defined Entity from VCD with the required attributes from the Terraform config.
func getRde(d *schema.ResourceData, meta interface{}) (*govcd.DefinedEntity, error) {
	vcdClient := meta.(*VCDClient)

	rdeTypeId := d.Get("rde_type_id").(string)
	rdeType, err := vcdClient.VCDClient.GetRdeTypeById(rdeTypeId)
	if err != nil {
		return nil, err
	}

	if d.Id() != "" {
		return rdeType.GetRdeById(d.Id())
	}

	return rdeType.GetRdeByName(d.Get("name").(string))
}

func resourceVcdRdeUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rdeType, err := getRde(d, meta)
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

	err = rdeType.Update(types.DefinedEntity{
		Name:       d.Get("name").(string),
		ExternalId: d.Get("external_id").(string),
		Entity:     jsonEntity,
	})
	if err != nil {
		return diag.Errorf("could not update the Runtime Defined Entity: %s", err)
	}
	return resourceVcdRdeRead(ctx, d, meta)
}

func resourceVcdRdeDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	rdeType, err := getRde(d, meta)
	if govcd.ContainsNotFound(err) {
		log.Printf("[DEBUG] Runtime Defined Entity no longer exists. Removing from tfstate")
		return nil
	}
	if err != nil {
		return diag.FromErr(err)
	}
	err = rdeType.Delete()
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
// Example resource name (_resource_name_): vcd_rde_type.outer-type
// Example import path (_the_id_string_): vmware.kubernetes.1.0.0
// Note: the separator can be changed using Provider.import_separator or variable VCD_IMPORT_SEPARATOR
func resourceVcdRdeImport(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	resourceURI := strings.Split(d.Id(), ImportSeparator)
	if len(resourceURI) < 3 {
		return nil, fmt.Errorf("resource identifier must be specified as vendor.namespace.version")
	}
	vendor, namespace, version := resourceURI[0], resourceURI[1], strings.Join(resourceURI[2:], ".")

	vcdClient := meta.(*VCDClient)
	rdeType, err := vcdClient.GetRdeType(vendor, namespace, version)
	if err != nil {
		return nil, fmt.Errorf("error finding Runtime Defined Entity with vendor %s, namespace %s and version %s: %s", vendor, namespace, version, err)
	}

	d.SetId(rdeType.DefinedEntityType.ID)
	return []*schema.ResourceData{d}, nil
}
