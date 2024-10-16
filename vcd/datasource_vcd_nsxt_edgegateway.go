package vcd

import (
	"context"
	"fmt"
	"log"

	"github.com/vmware/go-vcloud-director/v3/govcd"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdNsxtEdgeGateway() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdNsxtEdgeGatewayRead,
		Schema: map[string]*schema.Schema{
			"org": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Description: "The name of organization to use, optional if defined at provider " +
					"level. Useful when connected as sysadmin working across different organizations",
			},
			"vdc": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				Description:   "The name of VDC to use, optional if defined at provider level",
				ConflictsWith: []string{"owner_id"},
				Deprecated:    "This field is deprecated in favor of 'owner_id' which supports both - VDC and VDC Group IDs",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Edge Gateway name",
			},
			"owner_id": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				Description:   "ID of VDC or VDC Group",
				ConflictsWith: []string{"vdc"},
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Edge Gateway description",
			},
			"dedicate_external_network": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Dedicating the External Network will enable Route Advertisement for this Edge Gateway.",
			},
			"deployment_mode": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Edge Gateway deployment mode. One of 'DISTRIBUTED_ONLY', 'ACTIVE_STANDBY'.",
			},
			"external_network_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "External network ID",
			},
			"subnet": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "One or more blocks with external network information to be attached to this gateway's interface",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"gateway": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Gateway address for a subnet",
						},
						"prefix_length": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Prefix length for a subnet (e.g. 24)",
						},
						"primary_ip": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "IP address on the edge gateway",
						},
						"allocated_ips": {
							Type:        schema.TypeSet,
							Computed:    true,
							Description: "One or more blocks to sub-allocate pools on the edge gateway",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"start_address": {
										Type:     schema.TypeString,
										Computed: true,
									},
									"end_address": {
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
			"subnet_with_total_ip_count": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "Exposes IP allocation subnet for this Edge Gateway",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"gateway": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Gateway address for a subnet",
						},
						"primary_ip": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Primary IP address for the Edge Gateway",
						},
						"prefix_length": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Prefix length for a subnet (e.g. 24)",
						},
					},
				},
			},
			"subnet_with_ip_count": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "Exposes IP allocation subnet for this Edge Gateway including allocated IP count",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"gateway": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Gateway address for a subnet",
						},
						"prefix_length": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Prefix length for a subnet (e.g. 24)",
						},
						"primary_ip": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Primary IP address for the Edge Gateway - will be auto-assigned if not defined",
						},
						"allocated_ip_count": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Number of IP addresses to allocate",
						},
					},
				},
			},
			"primary_ip": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Primary IP address of edge gateway",
			},
			"edge_cluster_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "NSX-T Edge Cluster ID.",
			},
			"used_ip_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Number of used IP addresses",
			},
			"unused_ip_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Number of unused IP addresses",
			},
			"ip_count_read_limit": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     defaultReadLimitOfUnusedIps,
				Description: fmt.Sprintf("How many maximum IPs should be reported in 'used_ipcount' and 'unused_ip_count'. Default %d, 0 - unlimited", defaultReadLimitOfUnusedIps),
			},
			"total_allocated_ip_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Total number of IP addresses allocated for this Edge Gateway",
			},
			"use_ip_spaces": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Boolean value that specifies that the Edge Gateway is using IP Spaces",
			},
			"external_network": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "Additional NSX-T Segment Backed networks",
				Elem:        nsxtEdgeExternalNetworksDS,
			},
			"external_network_allocated_ip_count": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Total number of IPs allocated for this Gateway from NSX-T Segment backed External Network uplinks",
			},
			"non_distributed_routing_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "A flag indicating whether non-distributed routing is enabled or not.",
			},
		},
	}
}

var nsxtEdgeExternalNetworksDS = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"external_network_id": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "NSX-T Segment backed External Network ID",
		},
		"gateway": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Gateway IP Address",
		},
		"prefix_length": {
			Type:        schema.TypeInt,
			Computed:    true,
			Description: "Prefix length for a subnet (e.g. 24)",
		},
		"primary_ip": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Primary IP address for the Edge Gateway",
		},
		"allocated_ip_count": {
			Type:        schema.TypeInt,
			Computed:    true,
			Description: "Number of allocated IPs",
		},
	},
}

func datasourceVcdNsxtEdgeGatewayRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("[TRACE] NSX-T edge gateway datasource read initiated")

	vcdClient := meta.(*VCDClient)
	org, err := vcdClient.GetOrgFromResource(d)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error retrieving Org: %s", err))
	}

	// Validate if VDC or VDC Group is NSX-T backed
	inheritedVdcField := vcdClient.Vdc
	vdcField := d.Get("vdc").(string)
	ownerIdField := d.Get("owner_id").(string)

	err = validateIfVdcOrVdcGroupIsNsxt(org, inheritedVdcField, vdcField, ownerIdField)
	if err != nil {
		return diag.FromErr(err)
	}

	var edge *govcd.NsxtEdgeGateway
	edgeGatewayName := d.Get("name").(string)
	switch {
	case ownerIdField != "":
		edge, err = org.GetNsxtEdgeGatewayByNameAndOwnerId(edgeGatewayName, ownerIdField)
		if err != nil {
			return diag.Errorf("error getting NSX-T Edge Gateway:%s", err)
		}
	case ownerIdField == "":
		_, vdc, err := pickVdcIdByPriority(org, inheritedVdcField, vdcField, ownerIdField)
		if err != nil {
			return diag.Errorf("error getting VDC ID: %s", err)
		}

		edge, err = vdc.GetNsxtEdgeGatewayByName(edgeGatewayName)
		if err != nil {
			return diag.FromErr(fmt.Errorf("could not retrieve NSX-T edge gateway: %s", err))
		}
	default:
		return diag.Errorf("error looking up Edge Gateway - switch did not match any cases")
	}

	err = setNsxtEdgeGatewayData(vcdClient, edge, d)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading NSX-T Edge Gateway data: %s", err))
	}

	d.SetId(edge.EdgeGateway.ID)

	return nil
}
