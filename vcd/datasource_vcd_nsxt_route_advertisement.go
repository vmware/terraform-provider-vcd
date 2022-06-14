package vcd

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdNsxtRouteAdvertisement() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdNsxtRouteAdvertisementRead,

		Schema: map[string]*schema.Schema{
			"org": {
				Type:     schema.TypeString,
				Optional: true,
				Description: "The name of organization to use, optional if defined at provider " +
					"level. Useful when connected as sysadmin working across different organizations",
			},
			"edge_gateway_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "NSX-T Edge Gateway ID in which route advertisement is located",
			},
			"enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Defines if route advertisement is active",
			},
			"subnets": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "Set of subnets that will be advertised to Tier-0 gateway. Empty means none",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func datasourceVcdNsxtRouteAdvertisementRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	edgeGatewayID := d.Get("edge_gateway_id").(string)

	orgName, err := vcdClient.GetOrgNameFromResource(d)
	if err != nil {
		return diag.Errorf("error when getting Org name - %s", err)
	}

	nsxtEdge, err := vcdClient.GetNsxtEdgeGatewayById(orgName, edgeGatewayID)
	if err != nil {
		return diag.Errorf("error retrieving NSX-T Edge Gateway: %s", err)
	}

	routeAdvertisement, err := nsxtEdge.GetNsxtRouteAdvertisement(true)
	if err != nil {
		return diag.Errorf("error while retrieving route advertisement - %s", err)
	}

	dSet(d, "enabled", routeAdvertisement.Enable)

	subnetSet := convertStringsToTypeSet(routeAdvertisement.Subnets)
	err = d.Set("subnets", subnetSet)
	if err != nil {
		return diag.Errorf("error while setting subnets argument: %s", err)
	}

	d.SetId(edgeGatewayID)

	return nil
}
