package vcd

import (
	"context"
	"log"

	"github.com/vmware/go-vcloud-director/v2/govcd"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdNetworkIsolatedV2() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdNetworkIsolatedV2Read,

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
				Deprecated:    "This field is deprecated in favor of 'owner_id' which supports both - VDC and VDC Group IDs",
				ConflictsWith: []string{"owner_id"},
			},
			"owner_id": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				Description:   "ID of VDC or VDC Group",
				ConflictsWith: []string{"vdc"},
			},
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"name", "filter"},
				Description:  "A unique name for this network (optional if 'filter' is used)",
			},
			"filter": {
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
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Network description",
			},
			"is_shared": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "NSX-V only - share this network with other VDCs in this organization",
			},
			"gateway": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Gateway IP address",
			},
			"prefix_length": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Network prefix",
			},
			"dns1": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "DNS server 1",
			},
			"dns2": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "DNS server 1",
			},
			"dns_suffix": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "DNS suffix",
			},
			"static_ip_pool": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "IP ranges used for static pool allocation in the network",
				Elem:        networkV2IpRangeComputed,
			},
			"metadata": {
				Type:        schema.TypeMap,
				Computed:    true,
				Description: "Key value map of metadata assigned to this network. Key and value can be any string",
			},
		},
	}
}

func datasourceVcdNetworkIsolatedV2Read(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	org, err := vcdClient.GetOrgFromResource(d)
	if err != nil {
		return diag.Errorf("error retrieving Org: %s", err)
	}

	inheritedVdcField := vcdClient.Vdc
	vdcField := d.Get("vdc").(string)
	ownerIdField := d.Get("owner_id").(string)

	if !nameOrFilterIsSet(d) {
		return diag.Errorf(noNameOrFilterError, "vcd_network_isolated_v2")
	}

	networkName := d.Get("name").(string)

	// Try to search by filter if it exists
	var network *govcd.OpenApiOrgVdcNetwork
	filter, hasFilter := d.GetOk("filter")
	switch {
	// User supplied `filter`, search in the `vdc` (in data source or inherited)
	case hasFilter && networkName == "" && (vdcField != "" || inheritedVdcField != ""):
		_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
		if err != nil {
			return diag.Errorf("error getting VDC: %s", err)
		}
		network, err = getOpenApiOrgVdcNetworkByFilter(vdc, filter, "isolated")
		if err != nil {
			return diag.FromErr(err)
		}
	// TODO - XML Query based API does not support VDC Group networks (does not return them)
	// User supplied `filter` and `edge_gateway_id` (search scope can be detected - VDC or VDC Group)
	// case hasFilter && edgeGatewayId != "":
	// 	network, err = getOpenApiOrgVdcNetworkByFilter(vdc, filter, "isolated")
	// 	if err != nil {
	// 		return diag.FromErr(err)
	// 	}
	// User supplied `name` and also `owner_id`
	case ownerIdField != "" && networkName != "":
		network, err = org.GetOpenApiOrgVdcNetworkByNameAndOwnerId(networkName, ownerIdField)
		if err != nil {
			return diag.Errorf("[isolated network read v2] error getting Org VDC network: %s", err)
		}
	// Users supplied only `name` (VDC reference will be used from resource or inherited from provider)
	case networkName != "":
		_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
		if err != nil {
			return diag.Errorf("error getting VDC: %s", err)
		}

		network, err = vdc.GetOpenApiOrgVdcNetworkByName(d.Get("name").(string))
		if err != nil {
			return diag.Errorf("[isolated network read v2] error getting Org VDC network: %s", err)
		}
	default:
		return diag.Errorf("error - not all parameters specified for network lookup")
	}

	if !network.IsIsolated() {
		return diag.Errorf("[isolated network read v2] Org network with name '%s' found, but is not of type Isolated (ISOLATED) (type is '%s')",
			network.OpenApiOrgVdcNetwork.Name, network.GetType())
	}

	err = setOpenApiOrgVdcIsolatedNetworkData(d, network.OpenApiOrgVdcNetwork)
	if err != nil {
		return diag.Errorf("[isolated network read v2] error setting Org VDC network data: %s", err)
	}

	// Metadata is not supported when the network is in a VDC Group
	if !govcd.OwnerIsVdcGroup(network.OpenApiOrgVdcNetwork.OwnerRef.ID) {
		metadata, err := network.GetMetadata()
		if err != nil {
			log.Printf("[DEBUG] Unable to find isolated network v2 metadata: %s", err)
			return diag.Errorf("[isolated network read v2] unable to find Org VDC network metadata %s", err)
		}
		err = d.Set("metadata", getMetadataStruct(metadata.MetadataEntry))
		if err != nil {
			return diag.Errorf("[isolated network read v2] unable to set Org VDC network metadata %s", err)
		}
	}

	d.SetId(network.OpenApiOrgVdcNetwork.ID)

	return nil
}
