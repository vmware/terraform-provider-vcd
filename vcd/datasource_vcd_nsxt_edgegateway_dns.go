package vcd

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdNsxtEdgeGatewayDns() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdNsxtEdgeGatewayDnsRead,
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
				Description: "Edge gateway ID for DNS configuration",
			},
			"enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Status of the DNS Forwarder. Defaults to `true`",
			},
			"listener_ip": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "IP on which the DNS forwarder listens.",
			},
			"snat_rule_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "The value is `true` if a SNAT rule exists for the DNS forwarder.",
			},
			"snat_rule_ip_address": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The external IP address of the SNAT rule. (VCD 10.5.0+)",
			},
			"default_forwarder_zone": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The default forwarder zone.",
				Elem:        defaultForwarderZone,
			},
			"conditional_forwarder_zone": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "Conditional forwarder zone",
				Elem:        conditionalForwarderZone,
			},
		},
	}
}

func datasourceVcdNsxtEdgeGatewayDnsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericVcdNsxtEdgegatewayDnsRead(ctx, d, meta, "datasource")
}
