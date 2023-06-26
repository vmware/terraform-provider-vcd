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
			"nss": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "A unique namespace associated with the Runtime Defined Entity Interface. Combination of `vendor`, `nss` and `version` must be unique",
			},
			"version": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The Runtime Defined Entity Interface's version. The version follows semantic versioning rules. Combination of `vendor`, `nss` and `version` must be unique",
			},
			"vendor": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The vendor name. Combination of `vendor`, `nss` and `version` must be unique",
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The name of the Runtime Defined Entity Interface",
			},
			"readonly": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "True if the Runtime Defined Entity Interface cannot be modified",
			},
			"behavior": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "Behaviors defined in the Runtime Defined Entity Interface. Only System administrators can read Behaviors",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Name of the Defined Interface Behavior",
						},
						"description": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Description of the Defined Interface Behavior",
						},
						"execution": {
							Type:        schema.TypeMap,
							Computed:    true,
							Description: "Execution map of the Defined Interface Behavior",
						},
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The Defined Interface Behavior ID",
						},
						"ref": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The Behavior invocation reference to be used for polymorphic behavior invocations",
						},
					},
				},
			},
		},
	}
}

func datasourceVcdRdeInterfaceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericVcdRdeInterfaceRead(ctx, d, meta, "datasource")
}
