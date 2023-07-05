package vcd

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdRdeTypeBehavior() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdRdeTypeBehaviorRead,
		Schema: map[string]*schema.Schema{
			"rde_type_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the RDE Type that owns the Behavior to fetch",
			},
			"behavior_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of either a RDE Interface Behavior or RDE Type Behavior",
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The name of the Behavior",
			},
			"execution": {
				Type:        schema.TypeMap,
				Computed:    true,
				Description: "Execution map of the Behavior",
			},
			"ref": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The Behavior invocation reference to be used for polymorphic behavior invocations",
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "A description specifying the contract of the Behavior",
			},
		},
	}
}

func datasourceVcdRdeTypeBehaviorRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericVcdRdeTypeBehaviorRead(ctx, d, meta, "datasource")
}
