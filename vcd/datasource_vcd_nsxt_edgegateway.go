package vcd

import (
	"context"
	"fmt"
	"log"

	"github.com/vmware/go-vcloud-director/v2/govcd"

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
				Deprecated:    "This field is deprecated in favor of 'owner_id' which supports both - VDC and VDC group IDs",
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
							Description: "Netmask address for a subnet (e.g. 24 for /24)",
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
		},
	}
}

func datasourceVcdNsxtEdgeGatewayRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	err = valdateIfVdcOrVdcGroupIsNsxt(d, org, vcdClient, inheritedVdcField, vdcField, ownerIdField)
	if err != nil {
		return diag.FromErr(err)
	}

	var edge *govcd.NsxtEdgeGateway
	edgeGatewayName := d.Get("name").(string)
	ownerId := d.Get("owner_id").(string)
	switch {
	case ownerId != "":
		edge, err = org.GetNsxtEdgeGatewayByNameAndOwnerId(edgeGatewayName, ownerId)
		if err != nil {
			return diag.Errorf("error getting NSX-T Edge Gateway:%s", err)
		}
	case ownerId == "":
		edge, err = org.GetNsxtEdgeGatewayByName(edgeGatewayName)
		if err != nil {
			return diag.FromErr(fmt.Errorf("could not retrieve NSX-T edge gateway: %s", err))
		}
	}

	err = setNsxtEdgeGatewayData(edge.EdgeGateway, d)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading NSX-T Edge Gateway data: %s", err))
	}

	d.SetId(edge.EdgeGateway.ID)

	return nil
}

// valdateIfVdcOrVdcGroupIsNsxt picks VDC / VDC Group field by priority and check if it is NSX-T
// backed. If not - ir returns an error.
func valdateIfVdcOrVdcGroupIsNsxt(d *schema.ResourceData, org *govcd.Org, vdcClient *VCDClient, inheritedVdcField, vdcField, ownerIdField string) error {
	ownerId, err := getOwnerId(d, vdcClient, ownerIdField, vdcField, inheritedVdcField)

	if err != nil {
		return fmt.Errorf("error finding VDC ID: %s", err)
	}

	vdcOrGroup, err := getVdcOrVdcGroupVerifierByOwnerId(org, ownerId)
	if err != nil {
		return fmt.Errorf("error identifying VDC or VDC Group by Owner ID '%s' : %s", ownerId, err)
	}

	if !vdcOrGroup.IsNsxt() {
		return fmt.Errorf("this resource only supports NSX-T, but your entity is NSX-V backed")
	}

	return nil
}
