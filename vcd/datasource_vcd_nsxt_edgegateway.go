package vcd

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdNsxtEdgeGateway() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdNsxtEdgeGatewayRead,
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
							Description: "Netmask address for a subnet (e.g. 24 for /24)",
						},
						"primary_ip": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "IP address on the edge gateway ",
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
			"primary_ip": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Primary IP address of edge gateway",
			},
			"edge_cluster_id": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "NSX-T Edge Cluster ID.",
			},
		},
	}
}

func datasourceVcdNsxtEdgeGatewayRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("[TRACE] NSX-T edge gateway datasource read initiated")

	vcdClient := meta.(*VCDClient)

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error retrieving Org: %s", err))
	}

	if vdc.IsNsxv() {
		return diag.Errorf("please use 'vcd_edgegateway' for NSX-V backed VDC")
	}

	edge, err := vdc.GetNsxtEdgeGatewayByName(d.Get("name").(string))
	if err != nil {
		return diag.FromErr(fmt.Errorf("could not retrieve NSX-T edge gateway: %s", err))
	}

	err = setNsxtEdgeGatewayData(edge.EdgeGateway, d)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading NSX-T edge gateway data: %s", err))
	}

	d.SetId(edge.EdgeGateway.ID)

	return nil
}
