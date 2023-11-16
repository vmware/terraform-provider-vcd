package vcd

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func datasourceVcdNsxtEdgeGatewayDns() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdNsxtEdgeGatewayDnsRead,
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
				Description: "Edge gateway ID for DNS configuration",
			},
			"enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Status of the DNS Forwarder.",
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
				Elem:        defaultForwarderZoneDS,
			},
			"conditional_forwarder_zone": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "Conditional forwarder zone",
				Elem:        conditionalForwarderZoneDS,
			},
		},
	}
}

var defaultForwarderZoneDS = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"id": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Unique ID of the forwarder zone.",
		},
		"name": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Name of the forwarder zone.",
		},
		"upstream_servers": {
			Type:        schema.TypeSet,
			Computed:    true,
			Description: "Servers to which DNS requests should be forwarded to.",
			Elem: &schema.Schema{
				Type:         schema.TypeString,
				ValidateFunc: validation.IsIPAddress,
			},
		},
	},
}

var conditionalForwarderZoneDS = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"id": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Unique ID of the forwarder zone.",
		},
		"name": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Name of the forwarder zone.",
		},
		"upstream_servers": {
			Type:        schema.TypeSet,
			Computed:    true,
			Description: "Servers to which DNS requests should be forwarded to.",
			Elem: &schema.Schema{
				Type:         schema.TypeString,
				ValidateFunc: validation.IsIPAddress,
			},
		},
		"domain_names": {
			Type:        schema.TypeSet,
			Computed:    true,
			Description: "Set of domain names on which conditional forwarding is based.",
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
	},
}

func datasourceVcdNsxtEdgeGatewayDnsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericVcdNsxtEdgegatewayDnsRead(ctx, d, meta, "datasource")
}
