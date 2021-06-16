package vcd

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdNsxtIpSecVpnTunnel() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdNsxtIpSecVpnTunnelRead,

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
				ForceNew:    true,
				Description: "The name of VDC to use, optional if defined at provider level",
			},
			"edge_gateway_id": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Edge gateway name in which IP Sec VPN configuration is located",
			},
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of IP Sec VPN configuration",
			},
			"enabled": &schema.Schema{
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Enables or disables this configuration (default true)",
			},
			"description": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Description of NAT rule",
			},
			"pre_shared_key": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Pre-Shared Key (PSK)",
			},
			"local_ip_address": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "IPv4 Address for the endpoint. This has to be a suballocated IP on the Edge Gateway.",
			},
			"local_networks": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "Set of local networks in CIDR format. At least one value is required",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"remote_ip_address": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Public IPv4 Address of the remote device terminating the VPN connection",
			},
			"remote_networks": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "Set of remote networks in CIDR format. Leaving it empty is interpreted as 0.0.0.0/0",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"logging": &schema.Schema{
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Sets whether logging for the tunnel is enabled or not. (default - false)",
			},
			"security_profile_customization": &schema.Schema{
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Security profile customization",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ike_version": &schema.Schema{
							Type:        schema.TypeString,
							Required:    true,
							Description: "",
						},
						"ike_encryption_algorithms": &schema.Schema{
							Type:        schema.TypeSet,
							Computed:    true,
							Description: "",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"ike_digest_algorithms": &schema.Schema{
							Type:        schema.TypeSet,
							Computed:    true,
							Description: "",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"ike_dh_groups": &schema.Schema{
							Type:        schema.TypeSet,
							Computed:    true,
							Description: "",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"ike_sa_lifetime": &schema.Schema{
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Security Association life time (in seconds)",
						},

						"tunnel_pfs_enabled": &schema.Schema{
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Perfect Forward Secrecy",
						},

						"tunnel_df_policy": &schema.Schema{
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Perfect Forward Secrecy",
						},

						"tunnel_encryption_algorithms": &schema.Schema{
							Type:        schema.TypeSet,
							Computed:    true,
							Description: "",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"tunnel_digest_algorithms": &schema.Schema{
							Type:        schema.TypeSet,
							Optional:    true,
							Description: "",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"tunnel_dh_groups": &schema.Schema{
							Type:        schema.TypeSet,
							Computed:    true,
							Description: "",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"tunnel_sa_lifetime": &schema.Schema{
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Security Association life time (in seconds)",
						},
						"dpd_probe_internal": &schema.Schema{
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Dead Peer Detection probe interval (in seconds)",
						},
					},
				},
			},
			// Computed attributes from here
			"security_profile": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
				Description: "Security type which is use for IPsec VPN Tunnel. It will be 'DEFAULT' if nothing is " +
					"customized and 'CUSTOM' if some changes are applied",
			},
			"status": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Overall IPsec VPN Tunnel Status",
			},
			"ike_service_status": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Status for the actual IKE Session for the given tunnel",
			},
			"ike_fail_reason": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Provides more details of failure if the IKE service is not UP",
			},
		},
	}
}

func datasourceVcdNsxtIpSecVpnTunnelRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	orgName := d.Get("org").(string)
	vdcName := d.Get("vdc").(string)
	edgeGatewayId := d.Get("edge_gateway_id").(string)

	nsxtEdge, err := vcdClient.GetNsxtEdgeGatewayById(orgName, vdcName, edgeGatewayId)
	if err != nil {
		return diag.Errorf("error retrieving Edge Gateway: %s", err)
	}

	ipSecVpnTunnel, err := nsxtEdge.GetIpSecVpnByName(d.Get("name").(string))
	if err != nil {
		return diag.Errorf("error retrieving NSX-T IPsec VPN Tunnel configuration for deletion: %s", err)
	}

	// Set general schema for configuration
	err = setNsxtIpSecVpnTunnelData(d, ipSecVpnTunnel.NsxtIpSecVpn)
	if err != nil {
		return diag.Errorf("error storing NSX-T IPsec VPN Tunnel configuration to schema: %s", err)
	}

	d.SetId(ipSecVpnTunnel.NsxtIpSecVpn.ID)

	// Tunnel Security Properties
	tunnelConnectionProperties, err := ipSecVpnTunnel.GetTunnelConnectionProperties()
	if err != nil {
		return diag.Errorf("error reading NSX-T IPsec VPN Tunnel Security Customization: %s", err)
	}

	err = setNsxtIpSecVpnProfileTunnelConfigurationData(d, tunnelConnectionProperties)
	if err != nil {
		return diag.Errorf("error storing NSX-T IPsec VPN Tunnel Security Customization to schema: %s", err)
	}

	// Read tunnel status data from separate endpoint
	tunnelStatus, err := ipSecVpnTunnel.GetStatus()
	if err != nil {
		return diag.Errorf("error reading NSX-T IPsec VPN Tunnel status: %s", err)
	}
	setNsxtIpSecVpnTunnelStatusData(d, tunnelStatus)

	return nil
}
