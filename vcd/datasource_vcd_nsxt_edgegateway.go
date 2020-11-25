package vcd

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdNsxtEdgeGateway() *schema.Resource {
	return &schema.Resource{
		Read: resourceVcdNsxtEdgeGatewayRead,
		Schema: map[string]*schema.Schema{
			"org": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Description: "The name of organization to use, optional if defined at provider " +
					"level. Useful when connected as sysadmin working across different organizations",
			},
			"vdc": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "The name of VDC to use, optional if defined at provider level",
			},
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Edge Gateway name",
			},
			"description": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Edge Gateway description",
			},
			"dedicate_external_network": &schema.Schema{
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Dedicating the External Network will enable Route Advertisement for this Edge Gateway.",
			},
			"external_network_id": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "External network ID",
			},
			"subnet": {
				Description: "One or more blocks with external network information to be attached to this gateway's interface",
				Required:    true,
				Type:        schema.TypeSet,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"gateway": {
							Computed:    true,
							Description: "Gateway address for a subnet",
							Type:        schema.TypeString,
						},
						"prefix_length": {
							Computed:    true,
							Description: "Netmask address for a subnet (e.g. 24 for /24)",
							Type:        schema.TypeInt,
						},
						"enabled": {
							Computed:    true,
							Description: "Specifies if the subnet is enabled",
							Type:        schema.TypeBool,
						},
						"primary_ip": {
							Computed:    true,
							Type:        schema.TypeString,
							Description: "IP address on the edge gateway - will be auto-assigned if not defined",
						},
						"allocated_ips": {
							Computed:    true,
							Type:        schema.TypeSet,
							Description: "Define zero or more blocks to sub-allocate pools on the edge gateway",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"start_address": {
										Computed: true,
										Type:     schema.TypeString,
									},
									"end_address": {
										Computed: true,
										Type:     schema.TypeString,
									},
								},
							},
						},
					},
				},
			},
			"primary_ip": {
				Computed:    true,
				Type:        schema.TypeString,
				Description: "Primary IP address of edge gateway. Read-only (can be specified in specific subnet)",
			},
			"edge_cluster_id": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Select specific NSX-T Edge Cluster. Will be inherited from external network if not specified",
			},
		},
	}
}
