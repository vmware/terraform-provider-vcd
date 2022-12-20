package vcd

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdRdeInterface() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdRdeInterfaceRead,
		Schema: map[string]*schema.Schema{
			"namespace": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "A unique namespace associated with the interface",
			},
			"version": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The interface's version. The version should follow semantic versioning rules",
			},
			"vendor": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The interface's version. The version should follow semantic versioning rules",
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The name of the defined interface",
			},
			"readonly": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "True if the entity type cannot be modified",
			},
		},
	}
}

func datasourceVcdRdeInterfaceRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	vendor := d.Get("vendor").(string)
	nss := d.Get("namespace").(string)
	version := d.Get("version").(string)
	di, err := vcdClient.VCDClient.GetDefinedInterface(vendor, nss, version)
	if err != nil {
		return diag.Errorf("could not get any Defined Interface with vendor %s, namespace %s and version %s: %s", vendor, nss, version, err)
	}
	d.SetId(di.DefinedInterface.ID)
	dSet(d, "name", di.DefinedInterface.Name)
	dSet(d, "readonly", di.DefinedInterface.IsReadOnly)
	return nil
}
