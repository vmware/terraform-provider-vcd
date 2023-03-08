package vcd

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdAlbSettings() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdAlbSettingsRead,

		Schema: map[string]*schema.Schema{
			"org": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Description: "The name of organization to use, optional if defined at provider " +
					"level. Useful when connected as sysadmin working across different organizations",
			},
			"vdc": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The name of VDC to use, optional if defined at provider level",
				Deprecated:  "Edge Gateway will be looked up based on 'edge_gateway_id' field",
			},
			"edge_gateway_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Edge gateway ID",
			},
			"is_active": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Defines if ALB is enabled on Edge Gateway",
			},
			"service_network_specification": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Optional custom network CIDR definition for ALB Service Engine placement (VCD default is 192.168.255.1/25)",
			},
			"ipv6_service_network_specification": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The IPv6 network definition in Gateway CIDR format which will be used by Load Balancer service on Edge (VCD 10.4.0+)",
			},
			"supported_feature_set": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Feature set for ALB in this Edge Gateway. One of 'STANDARD', 'PREMIUM'.",
			},
			"is_transparent_mode_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Enabling transparent mode allows to configure Preserve Client IP on a Virtual Service (VCD 10.4.1+)",
			},
		},
	}
}

func datasourceVcdAlbSettingsRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return vcdAlbSettingsRead(meta, d, "datasource")
}
