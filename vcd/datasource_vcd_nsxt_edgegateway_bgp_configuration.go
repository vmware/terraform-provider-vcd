package vcd

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdEdgeBgpConfig() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdEdgeBgpConfigRead,

		Schema: map[string]*schema.Schema{
			"org": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Description: "The name of organization to use, optional if defined at provider " +
					"level. Useful when connected as sysadmin working across different organizations",
			},
			"edge_gateway_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Edge gateway name in which NAT Rule is located",
			},
			"enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Defines if BGP service is enabled",
			},
			"local_as_number": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Autonomous system number",
			},
			"graceful_restart_mode": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Graceful restart configuration on Edge Gateway",
			},
			"graceful_restart_timer": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Maximum time taken (in seconds) for a BGP session to be established after a restart",
			},
			"stale_route_timer": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Maximum time (in seconds) before stale routes are removed when BGP restarts",
			},
			"ecmp_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Defines if ECMP (Equal-cost multi-path routing) is enabled",
			},
		},
	}
}

func datasourceVcdEdgeBgpConfigRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	orgName := d.Get("org").(string)
	edgeGatewayId := d.Get("edge_gateway_id").(string)

	nsxtEdge, err := vcdClient.GetNsxtEdgeGatewayById(orgName, edgeGatewayId)
	if err != nil {
		return diag.Errorf("error retrieving NSX-T Edge Gateway BGP Configuration: %s", err)
	}

	bgpConfig, err := nsxtEdge.GetBgpConfiguration()
	if err != nil {
		return diag.Errorf("error retrieving NSX-T Edge Gateway BGP Configuration: %s", err)
	}

	setEdgeBgpConfigData(d, bgpConfig)

	d.SetId(nsxtEdge.EdgeGateway.ID)

	return nil
}
