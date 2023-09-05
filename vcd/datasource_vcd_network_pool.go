package vcd

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdNetworkPool() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceNetworkPoolRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of network pool.",
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
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Status of the network pool",
			},
			"network_provider_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Id of the network provider (either VC or NSX-T manager)",
			},
			"network_provider_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Name of the network provider",
			},
			"network_provider_type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Type of network provider",
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
	dSet(d, "status", networkPool.NetworkPool.Status)

	networkProviderType := "vCenter"
	if strings.Contains(networkPool.NetworkPool.ManagingOwnerRef.ID, "nsxtmanager") {
		networkProviderType = "NSX-T manager"
	}
	dSet(d, "network_provider_type", networkProviderType)
	dSet(d, "network_provider_id", networkPool.NetworkPool.ManagingOwnerRef.ID)
	dSet(d, "network_provider_name", networkPool.NetworkPool.ManagingOwnerRef.Name)
	d.SetId(networkPool.NetworkPool.Id)

	return nil
}
