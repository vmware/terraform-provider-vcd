package vcd

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdRdeTypeBehaviorAccessLevel() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdRdeTypeBehaviorAccessLevelRead,
		Schema: map[string]*schema.Schema{
			"rde_type_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of the RDE Type",
			},
			"behavior_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of either a RDE Interface Behavior or RDE Type Behavior",
			},
			"access_level_ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "Set of access level IDs associated to this Behavior",
			},
		},
	}
}

func datasourceVcdRdeTypeBehaviorAccessLevelRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericVcdRdeTypeBehaviorAccessLevelRead(ctx, d, meta)
}
