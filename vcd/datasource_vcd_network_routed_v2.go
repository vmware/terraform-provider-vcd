package vcd

import (
	"context"

	"github.com/vmware/go-vcloud-director/v2/govcd"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdNetworkRoutedV2() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdNetworkRoutedV2Read,

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
			"name": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"name", "filter"},
				Description:  "A unique name for this network (optional if 'filter' is used)",
			},
			"filter": &schema.Schema{
				Type:         schema.TypeList,
				MaxItems:     1,
				MinItems:     1,
				Optional:     true,
				ExactlyOneOf: []string{"name", "filter"},
				Description:  "Criteria for retrieving a network by various attributes",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name_regex": elementNameRegex,
						"ip":         elementIp,
					},
				},
			},
			"edge_gateway_id": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Edge gateway name in which Routed network is located",
			},
			"description": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Network description",
			},
			"interface_type": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Interface type (only for NSX-V networks). One of 'INTERNAL' (default), 'DISTRIBUTED', 'SUBINTERFACE'",
			},
			"gateway": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Gateway IP address",
			},
			"prefix_length": &schema.Schema{
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Network prefix",
			},
			"dns1": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "DNS server 1",
			},
			"dns2": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "DNS server 1",
			},
			"dns_suffix": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "DNS suffix",
			},
			"static_ip_pool": &schema.Schema{
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "IP ranges used for static pool allocation in the network",
				Elem:        networkV2IpRangeComputed,
			},
		},
	}
}

var networkV2IpRangeComputed = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"start_address": &schema.Schema{
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Start address of the IP range",
		},
		"end_address": &schema.Schema{
			Type:        schema.TypeString,
			Computed:    true,
			Description: "End address of the IP range",
		},
	},
}

func datasourceVcdNetworkRoutedV2Read(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return diag.Errorf("[routed network read v2] error retrieving VDC: %s", err)
	}

	if !nameOrFilterIsSet(d) {
		return diag.Errorf(noNameOrFilterError, "vcd_network_routed_v2")
	}

	name := d.Get("name").(string)

	// Try to search by filter if it exists
	var network *govcd.OpenApiOrgVdcNetwork
	filter, hasFilter := d.GetOk("filter")
	if hasFilter && name == "" {
		network, err = getOpenApiOrgVdcNetworkByFilter(vdc, filter, "routed")
		if err != nil {
			return diag.FromErr(err)
		}

	}

	if name != "" {
		network, err = vdc.GetOpenApiOrgVdcNetworkByName(d.Get("name").(string))
		if err != nil {
			return diag.Errorf("[routed network read v2] error getting Org VDC network: %s", err)
		}
	}

	// This check is needed to prevent a coverity warning about using a nil network.
	// In fact, this can't happen, because the schema definition requires either "name" or "filter" to be
	// defined, and both conditions are evaluated above.
	if network == nil {
		return diag.Errorf("undefined network: not found by either name or search criteria")
	}

	err = setOpenApiOrgVdcNetworkData(d, network.OpenApiOrgVdcNetwork)
	if err != nil {
		return diag.Errorf("[routed network read v2] error setting Org VDC network data: %s", err)
	}

	d.SetId(network.OpenApiOrgVdcNetwork.ID)

	return nil
}