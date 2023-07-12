package vcd

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdNetworkPool() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceNetworkPoolRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of NSX-T manager.",
			},
			"type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Type of the network pool",
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Description of the network pool",
			},
		},
	}
}

func datasourceNetworkPoolRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	resourcePoolName := d.Get("name").(string)

	networkPool, err := vcdClient.GetNetworkPoolByName(resourcePoolName)
	if err != nil {
		return diag.FromErr(err)
	}

	dSet(d, "name", networkPool.NetworkPool.Name)
	dSet(d, "type", networkPool.NetworkPool.PoolType)
	dSet(d, "description", networkPool.NetworkPool.Description)
	d.SetId(networkPool.NetworkPool.Id)

	return nil
}
