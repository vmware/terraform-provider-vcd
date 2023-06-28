package vcd

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdRdeInterfaceBehavior() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdRdeInterfaceBehaviorRead,
		Schema: map[string]*schema.Schema{
			"interface_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "ID of the RDE Interface that contains the desired Behavior",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The name of the RDE Interface Behavior to read",
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
				Description: "Description of the Behavior",
			},
		},
	}
}

func datasourceVcdRdeInterfaceBehaviorRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericVcdRdeInterfaceBehaviorRead(ctx, d, meta, "datasource")
}
