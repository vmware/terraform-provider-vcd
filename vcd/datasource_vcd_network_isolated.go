package vcd

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdNetworkIsolated() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdNetworkIsolatedRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"name", "filter"},
				Description:  "A unique name for this network (optional if 'filter' is used)",
			},
			"org": {
				Type:     schema.TypeString,
				Optional: true,
				Description: "The name of organization to use, optional if defined at provider " +
					"level. Useful when connected as sysadmin working across different organizations",
			},
			"vdc": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The name of VDC to use, optional if defined at provider level",
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Optional description for the network",
			},
			"netmask": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The netmask for the new network",
			},
			"gateway": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The gateway for this network",
			},
			"dns1": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "First DNS server to use",
			},
			"dns2": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Second DNS server to use",
			},
			"dns_suffix": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "A FQDN for the virtual machines on this network",
			},
			"href": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Network Hyper Reference",
			},
			"shared": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Defines if this network is shared between multiple VDCs in the Org",
			},
			"dhcp_pool": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "A range of IPs to issue to virtual machines that don't have a static IP",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"start_address": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The first address in the IP Range",
						},
						"end_address": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The final address in the IP Range",
						},
						"default_lease_time": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The default DHCP lease time to use",
						},
						"max_lease_time": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The maximum DHCP lease time to use",
						},
					},
				},
				Set: resourceVcdNetworkIsolatedDhcpPoolHash,
			},
			"static_ip_pool": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "A range of IPs permitted to be used as static IPs for virtual machines",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"start_address": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The first address in the IP Range",
						},
						"end_address": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The final address in the IP Range",
						},
					},
				},
				Set: resourceVcdNetworkStaticIpPoolHash,
			},
			"filter": {
				Type:        schema.TypeList,
				MaxItems:    1,
				MinItems:    1,
				Optional:    true,
				Description: "Criteria for retrieving a network by various attributes",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name_regex": elementNameRegex,
						"ip":         elementIp,
						"metadata":   elementMetadata,
					},
				},
			},
			"metadata": {
				Type:        schema.TypeMap,
				Computed:    true,
				Description: "Key value map of metadata assigned to this network. Key and value can be any string",
			},
		},
	}
}

func datasourceVcdNetworkIsolatedRead(c context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return diag.Errorf(errorRetrievingOrgAndVdc, err)
	}

	if vdc.IsNsxt() {
		logForScreen("vcd_network_isolated", "WARNING: please use 'vcd_network_isolated_v2' for NSX-T VDCs")
	}

	return genericVcdNetworkIsolatedRead(c, d, meta, "datasource")
}
