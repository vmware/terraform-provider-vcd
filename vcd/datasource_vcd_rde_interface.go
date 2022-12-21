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
				Description: "The vendor name",
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

func datasourceVcdRdeInterfaceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericVcdRdeInterfaceRead(ctx, d, meta, "datasource")
}
