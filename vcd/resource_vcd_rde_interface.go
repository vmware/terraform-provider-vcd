package vcd

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"log"
)

func resourceVcdRdeInterface() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdRdeInterfaceCreate,
		ReadContext:   resourceVcdRdeInterfaceRead,
		UpdateContext: resourceVcdRdeInterfaceUpdate,
		DeleteContext: resourceVcdRdeInterfaceDelete,
		Schema: map[string]*schema.Schema{
			"namespace": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true, // Can't update namespace
				Description: "A unique namespace associated with the interface",
			},
			"version": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true, // Can't update version
				Description: "The interface's version. The version should follow semantic versioning rules",
			},
			"vendor": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true, // Can't update vendor
				Description: "The vendor name",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the defined interface",
			},
			// FIXME: It seems this field is always false??????
			"readonly": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				ForceNew:    true, // Can't update readonly
				Description: "True if the entity type cannot be modified. Defaults to false",
			},
		},
	}
}

func resourceVcdRdeInterfaceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
	return genericVcdRdeInterfaceRead(ctx, d, meta, "resource")
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
		log.Printf("[DEBUG] Defined Interface no longer exists. Removing from tfstate")
		d.SetId("")
		return nil
	}
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(di.DefinedInterface.ID)
	dSet(d, "name", di.DefinedInterface.Name) // Although name is required in Create, we always "compute" it on Read
	dSet(d, "readonly", di.DefinedInterface.IsReadOnly)
	return nil
}

// getDefinedInterface retrieves a Defined Interface from VCD with the required attributes from the Terraform config.
func getDefinedInterface(d *schema.ResourceData, meta interface{}) (*govcd.DefinedInterface, error) {
	vcdClient := meta.(*VCDClient)

	vendor := d.Get("vendor").(string)
	nss := d.Get("namespace").(string)
	version := d.Get("version").(string)

	di, err := vcdClient.VCDClient.GetDefinedInterface(vendor, nss, version)
	if err != nil {
		return nil, fmt.Errorf("could not get any Defined Interface with vendor %s, namespace %s and version %s: %s", vendor, nss, version, err)
	}
	return di, nil
}

func resourceVcdRdeInterfaceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	di, err := getDefinedInterface(d, meta)
	if govcd.ContainsNotFound(err) {
		log.Printf("[DEBUG] Defined Interface no longer exists. Removing from tfstate")
		return nil
	}
	if err != nil {
		return diag.FromErr(err)
	}
	err = di.Update(types.DefinedInterface{
		Name:       d.Get("name").(string),
		Namespace:  d.Get("namespace").(string),
		Version:    d.Get("version").(string),
		Vendor:     d.Get("vendor").(string),
		IsReadOnly: d.Get("readonly").(bool),
	})
	if err != nil {
		return diag.Errorf("could not update the Defined Interface: %s", err)
	}
	return resourceVcdRdeInterfaceRead(ctx, d, meta)
}

func resourceVcdRdeInterfaceDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
