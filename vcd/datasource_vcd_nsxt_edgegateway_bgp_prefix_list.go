package vcd

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdEdgeBgpIpPrefixList() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdEdgeBgpIpPrefixListRead,
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
				Description: "Edge gateway ID for BGP IP Prefix List Configuration",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "BGP IP Prefix List name",
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "BGP IP Prefix List description",
			},
			"ip_prefix": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "BGP IP Prefix List entry",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"network": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Network in CIDR notation (e.g. '192.168.100.0/24', '2001:db8::/48')",
						},
						"action": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Action 'PERMIT' or 'DENY'",
						},
						"greater_than_or_equal_to": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Greater than or equal to (ge) subnet mask",
						},
						"less_than_or_equal_to": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Less than or equal to (le) subnet mask",
						},
					},
				},
			},
		},
	}
}

func datasourceVcdEdgeBgpIpPrefixListRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	orgName := d.Get("org").(string)
	edgeGatewayId := d.Get("edge_gateway_id").(string)

	nsxtEdge, err := vcdClient.GetNsxtEdgeGatewayById(orgName, edgeGatewayId)
	if err != nil {
		return diag.Errorf("[bgp ip prefix list read] error retrieving NSX-T Edge Gateway BGP IP Prefix List: %s", err)
	}

	bgpIpPrefixList, err := nsxtEdge.GetBgpIpPrefixListByName(d.Get("name").(string))
	if err != nil {
		return diag.Errorf("[bgp ip prefix list read] error retrieving NSX-T Edge Gateway BGP IP Prefix List: %s", err)
	}

	err = setEdgeBgpIpPrefixListData(d, bgpIpPrefixList.EdgeBgpIpPrefixList)
	if err != nil {
		return diag.Errorf("[bgp ip prefix list read] error storing entity into schema: %s", err)
	}

	d.SetId(bgpIpPrefixList.EdgeBgpIpPrefixList.ID)

	return nil
}
