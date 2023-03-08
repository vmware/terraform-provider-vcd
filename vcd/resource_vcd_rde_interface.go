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

func resourceVcdRdeInterface() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdRdeInterfaceCreate,
		ReadContext:   resourceVcdRdeInterfaceRead,
		UpdateContext: resourceVcdRdeInterfaceUpdate,
		DeleteContext: resourceVcdRdeInterfaceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdRdeInterfaceImport,
		},
		Schema: map[string]*schema.Schema{
			"nss": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true, // Can't update nss
				ValidateFunc: validateAlphanumericWithUnderscoresAndHyphens(),
				Description:  "A unique namespace associated with the Runtime Defined Entity Interface. Combination of `vendor`, `nss` and `version` must be unique",
			},
			"version": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true, // Can't update version
				Description: "The Runtime Defined Entity Interface's version. The version must follow semantic versioning rules. Combination of `vendor`, `nss` and `version` must be unique",
			},
			"vendor": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true, // Can't update vendor
				ValidateFunc: validateAlphanumericWithUnderscoresAndHyphens(),
				Description:  "The vendor name. Combination of `vendor`, `nss` and `version` must be unique",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the Runtime Defined Entity Interface",
			},
			"readonly": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "True if the Runtime Defined Entity Interface cannot be modified",
			},
		},
	}
}

func resourceVcdRdeInterfaceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	vendor := d.Get("vendor").(string)
	nss := d.Get("nss").(string)
	version := d.Get("version").(string)
	name := d.Get("name").(string)

	_, err := vcdClient.VCDClient.CreateDefinedInterface(&types.DefinedInterface{
		Name:    name,
		Nss:     nss,
		Version: version,
		Vendor:  vendor,
	})
	if err != nil {
		return diag.Errorf("could not create the Runtime Defined Entity Interface with name %s, vendor %s, nss %s and version %s: %s", name, vendor, nss, version, err)
	}
	return resourceVcdRdeInterfaceRead(ctx, d, meta)
}

func resourceVcdRdeInterfaceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericVcdRdeInterfaceRead(ctx, d, meta, "resource")
}

// genericVcdRdeInterfaceRead reads a Defined Interface from VCD and sets the Terraform state accordingly.
// If origin == "datasource", if the referenced Interface doesn't exist, it errors.
// If origin == "resource", if the referenced Interface doesn't exist, it removes it from tfstate and exits normally.
func genericVcdRdeInterfaceRead(_ context.Context, d *schema.ResourceData, meta interface{}, origin string) diag.Diagnostics {
	di, err := getDefinedInterface(d, meta)
	if origin == "resource" && govcd.ContainsNotFound(err) {
		log.Printf("[DEBUG] Runtime Defined Entity Interface no longer exists. Removing from tfstate")
		d.SetId("")
		return nil
	}
	if err != nil {
		return diag.FromErr(err)
	}

	dSet(d, "vendor", di.DefinedInterface.Vendor)
	dSet(d, "nss", di.DefinedInterface.Nss)
	dSet(d, "version", di.DefinedInterface.Version)
	dSet(d, "name", di.DefinedInterface.Name)
	dSet(d, "readonly", di.DefinedInterface.IsReadOnly)
	d.SetId(di.DefinedInterface.ID)

	return nil
}

// getDefinedInterface retrieves a Defined Interface from VCD with the required attributes from the Terraform config.
func getDefinedInterface(d *schema.ResourceData, meta interface{}) (*govcd.DefinedInterface, error) {
	vcdClient := meta.(*VCDClient)

	if d.Id() != "" {
		return vcdClient.VCDClient.GetDefinedInterfaceById(d.Id())
	}

	vendor := d.Get("vendor").(string)
	nss := d.Get("nss").(string)
	version := d.Get("version").(string)

	return vcdClient.VCDClient.GetDefinedInterface(vendor, nss, version)
}

func resourceVcdRdeInterfaceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	di, err := getDefinedInterface(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}
	err = di.Update(types.DefinedInterface{
		Name: d.Get("name").(string), // Only name can be updated
	})
	if err != nil {
		return diag.Errorf("could not update the Runtime Defined Entity Interface: %s", err)
	}
	return resourceVcdRdeInterfaceRead(ctx, d, meta)
}

func resourceVcdRdeInterfaceDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	di, err := getDefinedInterface(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}
	err = di.Delete()
	if err != nil {
		return diag.Errorf("could not delete the Runtime Defined Entity Interface: %s", err)
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
func resourceVcdRdeInterfaceImport(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	resourceURI := strings.Split(d.Id(), ImportSeparator)
	if len(resourceURI) < 3 {
		return nil, fmt.Errorf("resource identifier must be specified as vendor.nss.version")
	}
	vendor, nss, version := resourceURI[0], resourceURI[1], strings.Join(resourceURI[2:], ".")

	vcdClient := meta.(*VCDClient)
	di, err := vcdClient.GetDefinedInterface(vendor, nss, version)
	if err != nil {
		return nil, fmt.Errorf("error finding Runtime Defined Entity Interface with vendor %s, nss %s and version %s: %s", vendor, nss, version, err)
	}

	d.SetId(di.DefinedInterface.ID)
	return []*schema.ResourceData{d}, nil
}
