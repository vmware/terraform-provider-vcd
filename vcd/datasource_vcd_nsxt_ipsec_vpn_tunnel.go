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
				Description: "IPv4 Address for the endpoint. This has to be a sub-allocated IP on the Edge Gateway.",
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
							Description: "IKE version one of IKE_V1, IKE_V2, IKE_FLEX",
						},
						"ike_encryption_algorithms": &schema.Schema{
							Type:        schema.TypeSet,
							Computed:    true,
							Description: "Encryption algorithms. One of SHA1, SHA2_256, SHA2_384, SHA2_512",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"ike_digest_algorithms": &schema.Schema{
							Type:     schema.TypeSet,
							Computed: true,
							Description: "Secure hashing algorithms to use during the IKE negotiation. One of SHA1, " +
								"SHA2_256, SHA2_384, SHA2_512",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"ike_dh_groups": &schema.Schema{
							Type:     schema.TypeSet,
							Computed: true,
							Description: "Diffie-Hellman groups to be used if Perfect Forward Secrecy is enabled. One " +
								"of GROUP2, GROUP5, GROUP14, GROUP15, GROUP16, GROUP19, GROUP20, GROUP21",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"ike_sa_lifetime": &schema.Schema{
							Type:     schema.TypeInt,
							Computed: true,
							Description: "Security Association life time (in seconds). It is number of seconds " +
								"before the IPsec tunnel needs to reestablish",
						},

						"tunnel_pfs_enabled": &schema.Schema{
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Perfect Forward Secrecy Enabled or Disabled. Default (enabled)",
						},

						"tunnel_df_policy": &schema.Schema{
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Policy for handling defragmentation bit. One of COPY, CLEAR",
						},

						"tunnel_encryption_algorithms": &schema.Schema{
							Type:     schema.TypeSet,
							Computed: true,
							Description: "Encryption algorithms to use in IPSec tunnel establishment. One of AES_128, " +
								"AES_256, AES_GCM_128, AES_GCM_192, AES_GCM_256, NO_ENCRYPTION_AUTH_AES_GMAC_128, " +
								"NO_ENCRYPTION_AUTH_AES_GMAC_192, NO_ENCRYPTION_AUTH_AES_GMAC_256, NO_ENCRYPTION",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"tunnel_digest_algorithms": &schema.Schema{
							Type:     schema.TypeSet,
							Optional: true,
							Description: "Digest algorithms to be used for message digest. One of SHA1, SHA2_256, " +
								"SHA2_384, SHA2_512",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"tunnel_dh_groups": &schema.Schema{
							Type:     schema.TypeSet,
							Computed: true,
							Description: "Diffie-Hellman groups to be used is PFS is enabled. One of GROUP2, GROUP5, " +
								"GROUP14, GROUP15, GROUP16, GROUP19, GROUP20, GROUP21",
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
							Type:     schema.TypeInt,
							Computed: true,
							Description: "Value in seconds of dead probe detection interval. Minimum is 3 seconds and " +
								"the maximum is 60 seconds",
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

	ipSecVpnTunnelName := d.Get("name").(string)
	ipSecVpnTunnel, err := nsxtEdge.GetIpSecVpnTunnelByName(ipSecVpnTunnelName)
	if err != nil {
		return diag.Errorf("error retrieving NSX-T IPsec VPN Tunnel configuration with name '%s;: %s", ipSecVpnTunnelName, err)
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
