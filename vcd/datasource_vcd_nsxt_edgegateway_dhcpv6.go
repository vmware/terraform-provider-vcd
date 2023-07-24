package vcd

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdNsxtEdgegatewayDhcpV6() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdNsxtEdgegatewayDhcpV6Read,

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
				Description: "Edge gateway ID for DHCPv6 configuration",
			},
			"mode": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "DHCPv6 configuration mode. One of 'SLAAC', 'DHCPv6'",
			},
			"domain_names": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "A set of domain names (only applicable for 'SLAAC' mode)",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"dns_servers": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "A set of DNS Servers (only applicable for 'SLAAC' mode)",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func datasourceVcdNsxtEdgegatewayDhcpV6Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	orgName := d.Get("org").(string)
	edgeGatewayId := d.Get("edge_gateway_id").(string)

	nsxtEdge, err := vcdClient.GetNsxtEdgeGatewayById(orgName, edgeGatewayId)
	if err != nil {
		return diag.Errorf("[dhcpv6 (SLAAC Profile) DS read] error retrieving NSX-T Edge Gateway DHCPv6: %s", err)
	}

	slaacProfile, err := nsxtEdge.GetSlaacProfile()
	if err != nil {
		return diag.Errorf("[dhcpv6 (SLAAC Profile) DS read] error retrieving NSX-T Edge Gateway DHCPv6: %s", err)
	}

	// Rate limiting does not have its own ID - it is a part of Edge Gateway
	d.SetId(edgeGatewayId)
	err = setNsxtEdgeGatewaySlaacProfileData(d, slaacProfile)
	if err != nil {
		return diag.Errorf("error storing state: %s", err)
	}

	return nil
}
