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

	err = setNsxtEdgeGatewayData(edge.EdgeGateway, d)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading NSX-T Edge Gateway data: %s", err))
	}

	d.SetId(edge.EdgeGateway.ID)

	return nil
}

// validateIfVdcOrVdcGroupIsNsxt evaluates VDC field priority using pickVdcIdByPriority and then
// checks if that VDC or VDC Group is an NSX-T one and returns an error if not
func validateIfVdcOrVdcGroupIsNsxt(org *govcd.Org, inheritedVdcField, vdcField, ownerIdField string) error {
	usedFieldId, _, err := pickVdcIdByPriority(org, inheritedVdcField, vdcField, ownerIdField)

	if err != nil {
		return fmt.Errorf("error finding VDC ID: %s", err)
	}

	isNsxt, err := isBackedByNsxt(org, usedFieldId)
	if err != nil {
		return fmt.Errorf("error ")
	}

	if !isNsxt {
		return fmt.Errorf("please use 'vcd_edgegateway' for NSX-V backed VDC")
	}

	return nil
}

// pickVdcIdByPriority picks primary field to be used from the specified ones. The priority is such
// * `owner_id`
// * `vdc` at resource level
// * `vdc` inherited from provider configuration
func pickVdcIdByPriority(org *govcd.Org, inheritedVdcField, vdcField, ownerIdField string) (string, *govcd.Vdc, error) {
	if ownerIdField != "" {
		return ownerIdField, nil, nil
	}

	if vdcField != "" {
		vdc, err := org.GetVDCByName(vdcField, false)
		if err != nil {
			return "", nil, fmt.Errorf("error finding VDC '%s': %s", vdc.Vdc.ID, err)
		}
		return vdc.Vdc.ID, vdc, nil
	}

	if inheritedVdcField != "" {
		vdc, err := org.GetVDCByName(inheritedVdcField, false)
		if err != nil {
			return "", nil, fmt.Errorf("error finding VDC '%s': %s", vdc.Vdc.ID, err)
		}
		return vdc.Vdc.ID, vdc, nil
	}

	return "", nil, fmt.Errorf("none of the fields `owner_id`, `vdc` and provider inherited `vdc`")
}

// isBackedByNsxt accepts VDC or VDC Group ID and checks if it is backed by NSX-T
func isBackedByNsxt(org *govcd.Org, vdcOrVdcGroupId string) (bool, error) {
	var vdcOrGroup vdcOrVdcGroupVerifier
	var err error

	switch {
	case govcd.OwnerIsVdc(vdcOrVdcGroupId):
		vdcOrGroup, err = org.GetVDCById(vdcOrVdcGroupId, false)
		if err != nil {
			return false, err
		}
	case govcd.OwnerIsVdcGroup(vdcOrVdcGroupId):
		vdcOrGroup, err = org.GetVdcGroupById(vdcOrVdcGroupId)
		if err != nil {
			return false, err
		}
	default:
		return false, fmt.Errorf("error determining VDC type by ID '%s'", vdcOrVdcGroupId)
	}

	return vdcOrGroup.IsNsxt(), nil
}
