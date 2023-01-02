package vcd

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"log"
	"strings"
)

func resourceVcdRdeType() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdRdeTypeCreate,
		ReadContext:   resourceVcdRdeTypeRead,
		UpdateContext: resourceVcdRdeTypeUpdate,
		DeleteContext: resourceVcdRdeTypeDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdRdeTypeImport,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the Runtime Defined Entity type",
			},
			"vendor": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The vendor name for the Runtime Defined Entity type",
			},
			"namespace": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "A unique namespace associated with the Runtime Defined Entity type",
			},
			"version": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The version of the Runtime Defined Entity type. The version string must follow semantic versioning rules",
			},
			"interface_ids": {
				Type: schema.TypeSet,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Required:    true,
				Description: "Set of Defined Interface URNs that this defined entity type is referenced by",
			},
			"schema_url": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "URL that should point to a JSON-Schema valid definition file of the Runtime Defined Entity type",
				AtLeastOneOf: []string{"schema_url", "schema"},
			},
			"schema": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				Description:  "The JSON-Schema valid definition of the Runtime Defined Entity type",
				AtLeastOneOf: []string{"schema_url", "schema"},
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The description of the Runtime Defined Entity type",
			},
			"external_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "An external entity's id that this definition may apply to",
			},
			"hooks": {
				Type: schema.TypeMap,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional:    true,
				Description: "A mapping defining which behaviors should be invoked upon specific lifecycle events, like PostCreate, PostUpdate or PreDelete",
			},
			"inherited_version": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "To be used when creating a new version of a defined entity type. Specifies the version of the type that will be the template for the authorization configuration of the new version. The Type ACLs and the access requirements of the Type Behaviors of the new version will be copied from those of the inherited version. If the value of this property is ‘0’, then the new type version will not inherit another version and will have the default authorization settings, just like the first version of a new type",
			},
			"readonly": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "True if the entity type cannot be modified",
			},
		},
	}
}

func resourceVcdRdeTypeCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	vendor := d.Get("vendor").(string)
	nss := d.Get("namespace").(string)
	version := d.Get("version").(string)
	name := d.Get("name").(string)
	readonly := d.Get("readonly").(bool)

	_, err := vcdClient.VCDClient.CreateDefinedInterface(&types.DefinedInterface{
		Name:       name,
		Namespace:  nss,
		Version:    version,
		Vendor:     vendor,
		IsReadOnly: readonly,
	})
	if err != nil {
		return diag.Errorf("could not create the Defined Interface with name %s, vendor %s, namespace %s and version %s: %s", name, vendor, nss, version, err)
	}
	return resourceVcdRdeTypeRead(ctx, d, meta)
}

func resourceVcdRdeTypeRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericVcdRdeTypeRead(ctx, d, meta, "resource")
}

// genericVcdRdeTypeRead reads a Runtime Defined Entity type from VCD and sets the Terraform state accordingly.
// If origin == "datasource", if the referenced RDE type doesn't exist, it errors.
// If origin == "resource", if the referenced RDE type doesn't exist, it removes it from tfstate and exits normally.
func genericVcdRdeTypeRead(_ context.Context, d *schema.ResourceData, meta interface{}, origin string) diag.Diagnostics {
	di, err := getDefinedInterface(d, meta)
	if origin == "resource" && govcd.ContainsNotFound(err) {
		log.Printf("[DEBUG] Defined Interface no longer exists. Removing from tfstate")
		d.SetId("")
		return nil
	}
	if err != nil {
		return diag.FromErr(err)
	}

	dSet(d, "vendor", di.DefinedInterface.Vendor)
	dSet(d, "namespace", di.DefinedInterface.Namespace)
	dSet(d, "version", di.DefinedInterface.Version)
	dSet(d, "name", di.DefinedInterface.Name)
	dSet(d, "readonly", di.DefinedInterface.IsReadOnly)
	d.SetId(di.DefinedInterface.ID)

	return nil
}

func resourceVcdRdeTypeUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	di, err := getDefinedInterface(d, meta)
	if govcd.ContainsNotFound(err) {
		log.Printf("[DEBUG] Defined Interface no longer exists. Removing from tfstate")
		return nil
	}
	if err != nil {
		return diag.FromErr(err)
	}
	err = di.Update(types.DefinedInterface{
		Name: d.Get("name").(string), // Only name can be updated
	})
	if err != nil {
		return diag.Errorf("could not update the Defined Interface: %s", err)
	}
	return resourceVcdRdeTypeRead(ctx, d, meta)
}

func resourceVcdRdeTypeDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	di, err := getDefinedInterface(d, meta)
	if govcd.ContainsNotFound(err) {
		log.Printf("[DEBUG] Defined Interface no longer exists. Removing from tfstate")
		return nil
	}
	if err != nil {
		return diag.FromErr(err)
	}
	err = di.Delete()
	if err != nil {
		return diag.Errorf("could not delete the Defined Interface: %s", err)
	}
	return nil
}

// resourceVcdRdeInterfaceImport is responsible for importing the resource.
// The following steps happen as part of import
// 1. The user supplies `terraform import _resource_name_ _the_id_string_` command
// 2. `_the_id_string_` contains a dot formatted path to resource as in the example below
// 3. The functions splits the dot-formatted path and tries to lookup the object
// 4. If the lookup succeeds it set's the ID field for `_resource_name_` resource in state file
// (the resource must be already defined in .tf config otherwise `terraform import` will complain)
// 5. `terraform refresh` is being implicitly launched. The Read method looks up all other fields
// based on the known ID of object.
//
// Example resource name (_resource_name_): vcd_rde_interface.outer-interface
// Example import path (_the_id_string_): vmware.kubernetes.1.0.0
// Note: the separator can be changed using Provider.import_separator or variable VCD_IMPORT_SEPARATOR
func resourceVcdRdeTypeImport(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	resourceURI := strings.Split(d.Id(), ImportSeparator)
	if len(resourceURI) < 3 {
		return nil, fmt.Errorf("resource identifier must be specified as vendor.namespace.version")
	}
	vendor, namespace, version := resourceURI[0], resourceURI[1], strings.Join(resourceURI[2:], ".")

	vcdClient := meta.(*VCDClient)
	di, err := vcdClient.GetDefinedInterface(vendor, namespace, version)
	if err != nil {
		return nil, fmt.Errorf("error finding Defined Interface with vendor %s, namespace %s and version %s: %s", vendor, namespace, version, err)
	}

	d.SetId(di.DefinedInterface.ID)
	return []*schema.ResourceData{d}, nil
}
