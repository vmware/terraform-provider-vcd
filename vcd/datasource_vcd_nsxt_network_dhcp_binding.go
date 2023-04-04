package vcd

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdNsxtDhcpBinding() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdNsxtDhcpBindingRead,

		Schema: map[string]*schema.Schema{
			"org": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Description: "The name of organization to use, optional if defined at provider " +
					"level. Useful when connected as sysadmin working across different organizations",
			},
			"org_network_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Parent Org VDC network ID",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of DHCP binding",
			},

			"ip_address": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "IP address of the DHCP binding",
			},
			"mac_address": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "MAC address of the DHCP binding",
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Description of DHCP binding",
			},
			"binding_type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Binding type 'IPV4' or 'IPV6'",
			},
			"dns_servers": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The DNS server IPs to be assigned . 2 values maximum.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"lease_time": {
				Type:     schema.TypeInt,
				Computed: true,
			},
			"dhcp_v4_config": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "IPv4 specific DHCP Binding configuration",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"gateway_ip_address": {
							Computed:    true,
							Type:        schema.TypeString,
							Description: "",
						},
						"hostname": {
							Computed:    true,
							Type:        schema.TypeString,
							Description: "",
						},
					},
				},
			},
			"dhcp_v6_config": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "IPv6 specific DHCP Binding configuration",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"sntp_servers": {
							Computed:    true,
							Type:        schema.TypeSet,
							Description: "",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"dns_servers": {
							Computed:    true,
							Type:        schema.TypeSet,
							Description: "IP address, CIDR, an IP range, or the keyword 'any'",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
		},
	}
}

func datasourceVcdNsxtDhcpBindingRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	org, err := vcdClient.GetOrgFromResource(d)
	if err != nil {
		return diag.Errorf("[NSX-T DHCP binding DS read] error retrieving Org: %s", err)
	}

	orgNetworkId := d.Get("org_network_id").(string)

	// Perform validations to only allow DHCP configuration on NSX-T backed Routed Org VDC networks
	orgVdcNet, err := org.GetOpenApiOrgVdcNetworkById(orgNetworkId)
	if err != nil {
		return diag.Errorf("[NSX-T DHCP binding DS read] error retrieving Org VDC network with ID '%s': %s", orgNetworkId, err)
	}

	bindingName := d.Get("name").(string)
	dhcpBinding, err := orgVdcNet.GetOpenApiOrgVdcNetworkDhcpBindingByName(bindingName)
	if err != nil {
		return diag.Errorf("[NSX-T DHCP binding DS read] error retrieving DHCP binding with ID '%s' for Org VDC network with ID '%s': %s",
			bindingName, orgNetworkId, err)
	}

	d.SetId(dhcpBinding.OpenApiOrgVdcNetworkDhcpBinding.ID)

	if err := setOpenApiOrgVdcNetworkDhcpBindingData(d, dhcpBinding.OpenApiOrgVdcNetworkDhcpBinding); err != nil {
		return diag.Errorf("[NSX-T DHCP binding DS read] error setting DHCP binding data: %s", err)
	}
	return nil
}
