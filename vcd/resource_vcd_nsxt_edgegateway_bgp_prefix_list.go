package vcd

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func resourceVcdEdgeBgpIpPrefixList() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdEdgeBgpIpPrefixListCreate,
		ReadContext:   resourceVcdEdgeBgpIpPrefixListRead,
		UpdateContext: resourceVcdEdgeBgpIpPrefixListUpdate,
		DeleteContext: resourceVcdEdgeBgpIpPrefixListDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdEdgeBgpIpPrefixListImport,
		},

		Schema: map[string]*schema.Schema{
			"org": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Description: "The name of organization to use, optional if defined at provider " +
					"level. Useful when connected as sysadmin working across different organizations",
			},
			"edge_gateway_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Edge gateway ID for BGP IP Prefix List Configuration",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "BGP IP Prefix List name",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "BGP IP Prefix List description",
			},
			"ip_prefix": {
				Type:     schema.TypeSet,
				Required: true,
				// Cannot create empty container without any IP prefixes
				MinItems:    1,
				Description: "BGP IP Prefix List entry",
				Elem:        edgeBgpIpPrefixListHash,
			},
		},
	}
}

var edgeBgpIpPrefixListHash = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"network": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Network in CIDR notation (e.g. '192.168.100.0/24', '2001:db8::/48')",
		},
		"action": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Action 'PERMIT' or 'DENY'",
			// API error is not friendly (complains that field is not specified at all) therefore
			// enforcing validation
			ValidateFunc: validation.StringInSlice([]string{"PERMIT", "DENY"}, false),
		},
		"greater_than_or_equal_to": {
			Type:        schema.TypeInt,
			Optional:    true,
			Description: "Greater than or equal to subnet mask",
		},
		"less_than_or_equal_to": {
			Type:        schema.TypeInt,
			Optional:    true,
			Description: "Less than or equal to subnet mask",
		},
	},
}

func resourceVcdEdgeBgpIpPrefixListCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	// Handling locks on for BGP configuration is conditional. There are two scenarios:
	// * When the parent Edge Gateway is in a VDC - a lock on parent Edge Gateway must be acquired
	// * When the parent Edge Gateway is in a VDC Group - a lock on parent VDC Group must be acquired
	// To find out parent lock object, Edge Gateway must be looked up and its OwnerRef must be checked
	// Note. It is not safe to do multiple locks in the same resource as it can result in a deadlock
	parentEdgeGatewayOwnerId, _, err := getParentEdgeGatewayOwnerId(vcdClient, d)
	if err != nil {
		return diag.Errorf("[bgp ip prefix list create] error finding parent Edge Gateway: %s", err)
	}

	if govcd.OwnerIsVdcGroup(parentEdgeGatewayOwnerId) {
		vcdClient.lockById(parentEdgeGatewayOwnerId)
		defer vcdClient.unlockById(parentEdgeGatewayOwnerId)
	} else {
		vcdClient.lockParentEdgeGtw(d)
		defer vcdClient.unLockParentEdgeGtw(d)
	}

	orgName := d.Get("org").(string)
	edgeGatewayId := d.Get("edge_gateway_id").(string)

	nsxtEdge, err := vcdClient.GetNsxtEdgeGatewayById(orgName, edgeGatewayId)
	if err != nil {
		return diag.Errorf("[bgp ip prefix list create] error retrieving Edge Gateway: %s", err)
	}

	bgpIpPrefixList := getEdgeBgpIpPrefixListType(d)

	createdBgpIpPrefixList, err := nsxtEdge.CreateBgpIpPrefixList(bgpIpPrefixList)
	if err != nil {
		return diag.Errorf("[bgp ip prefix list create] error updating NSX-T Edge Gateway BGP IP Prefix List: %s", err)
	}

	d.SetId(createdBgpIpPrefixList.EdgeBgpIpPrefixList.ID)

	return resourceVcdEdgeBgpIpPrefixListRead(ctx, d, meta)
}

func resourceVcdEdgeBgpIpPrefixListUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	// Handling locks on for BGP configuration is conditional. There are two scenarios:
	// * When the parent Edge Gateway is in a VDC - a lock on parent Edge Gateway must be acquired
	// * When the parent Edge Gateway is in a VDC Group - a lock on parent VDC Group must be acquired
	// To find out parent lock object, Edge Gateway must be looked up and its OwnerRef must be checked
	// Note. It is not safe to do multiple locks in the same resource as it can result in a deadlock
	parentEdgeGatewayOwnerId, _, err := getParentEdgeGatewayOwnerId(vcdClient, d)
	if err != nil {
		return diag.Errorf("[bgp configuration update] error finding parent Edge Gateway: %s", err)
	}

	if govcd.OwnerIsVdcGroup(parentEdgeGatewayOwnerId) {
		vcdClient.lockById(parentEdgeGatewayOwnerId)
		defer vcdClient.unlockById(parentEdgeGatewayOwnerId)
	} else {
		vcdClient.lockParentEdgeGtw(d)
		defer vcdClient.unLockParentEdgeGtw(d)
	}

	orgName := d.Get("org").(string)
	edgeGatewayId := d.Get("edge_gateway_id").(string)

	nsxtEdge, err := vcdClient.GetNsxtEdgeGatewayById(orgName, edgeGatewayId)
	if err != nil {
		return diag.Errorf("[bgp ip prefix list update] error retrieving Edge Gateway: %s", err)
	}

	bgpIpPrefixList, err := nsxtEdge.GetBgpIpPrefixListById(d.Id())
	if err != nil {
		return diag.Errorf("[bgp ip prefix list update] error retrieving NSX-T Edge Gateway BGP IP Prefix Lis: %s", err)
	}

	bgpIpPrefixList.EdgeBgpIpPrefixList = getEdgeBgpIpPrefixListType(d)
	bgpIpPrefixList.EdgeBgpIpPrefixList.ID = d.Id()

	_, err = bgpIpPrefixList.Update(bgpIpPrefixList.EdgeBgpIpPrefixList)
	if err != nil {
		return diag.Errorf("[bgp ip prefix list update] error updating NSX-T Edge Gateway BGP IP Prefix List: %s", err)
	}

	return resourceVcdEdgeBgpIpPrefixListRead(ctx, d, meta)
}

func resourceVcdEdgeBgpIpPrefixListRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	orgName := d.Get("org").(string)
	edgeGatewayId := d.Get("edge_gateway_id").(string)

	nsxtEdge, err := vcdClient.GetNsxtEdgeGatewayById(orgName, edgeGatewayId)
	if err != nil {
		if govcd.ContainsNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.Errorf("[bgp ip prefix list read] error retrieving NSX-T Edge Gateway BGP IP Prefix List: %s", err)
	}

	bgpIpPrefixList, err := nsxtEdge.GetBgpIpPrefixListById(d.Id())
	if err != nil {
		return diag.Errorf("[bgp ip prefix list read] error retrieving NSX-T Edge Gateway BGP IP Prefix List: %s", err)
	}

	err = setEdgeBgpIpPrefixListData(d, bgpIpPrefixList.EdgeBgpIpPrefixList)
	if err != nil {
		return diag.Errorf("[bgp ip prefix list read] error storing entity into schema: %s", err)
	}

	return nil
}

func resourceVcdEdgeBgpIpPrefixListDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	// Handling locks on for BGP configuration is conditional. There are two scenarios:
	// * When the parent Edge Gateway is in a VDC - a lock on parent Edge Gateway must be acquired
	// * When the parent Edge Gateway is in a VDC Group - a lock on parent VDC Group must be acquired
	// To find out parent lock object, Edge Gateway must be looked up and its OwnerRef must be checked
	// Note. It is not safe to do multiple locks in the same resource as it can result in a deadlock
	parentEdgeGatewayOwnerId, _, err := getParentEdgeGatewayOwnerId(vcdClient, d)
	if err != nil {
		return diag.Errorf("[bgp ip prefix list delete] error finding parent Edge Gateway: %s", err)
	}

	if govcd.OwnerIsVdcGroup(parentEdgeGatewayOwnerId) {
		vcdClient.lockById(parentEdgeGatewayOwnerId)
		defer vcdClient.unlockById(parentEdgeGatewayOwnerId)
	} else {
		vcdClient.lockParentEdgeGtw(d)
		defer vcdClient.unLockParentEdgeGtw(d)
	}

	orgName := d.Get("org").(string)
	edgeGatewayId := d.Get("edge_gateway_id").(string)

	nsxtEdge, err := vcdClient.GetNsxtEdgeGatewayById(orgName, edgeGatewayId)
	if err != nil {
		return diag.Errorf("[bgp ip prefix list delete] error retrieving NSX-T Edge Gateway: %s", err)
	}

	bgpIpPrefixList, err := nsxtEdge.GetBgpIpPrefixListById(d.Id())
	if err != nil {
		return diag.Errorf("[bgp ip prefix list delete] error retrieving NSX-T Edge Gateway BGP Configuration: %s", err)
	}

	err = bgpIpPrefixList.Delete()
	if err != nil {
		return diag.Errorf("[bgp ip prefix list delete] error deleting NSX-T Edge Gateway BGP Configuration: %s", err)
	}

	return nil
}

func resourceVcdEdgeBgpIpPrefixListImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	resourceURI := strings.Split(d.Id(), ImportSeparator)
	if len(resourceURI) != 4 {
		return nil, fmt.Errorf("resource name must be specified as org-name.vdc-or-vdc-group-name.edge_gateway_name.bgp_prefix_list_name")
	}
	orgName, vdcOrVdcGroupName, edgeGatewayName, bgpIpPrefixListName := resourceURI[0], resourceURI[1], resourceURI[2], resourceURI[3]

	vcdClient := meta.(*VCDClient)
	vdcOrVdcGroup, err := lookupVdcOrVdcGroup(vcdClient, orgName, vdcOrVdcGroupName)
	if err != nil {
		return nil, err
	}

	edgeGateway, err := vdcOrVdcGroup.GetNsxtEdgeGatewayByName(edgeGatewayName)
	if err != nil {
		return nil, fmt.Errorf("[bgp ip prefix list import] unable to find Edge Gateway '%s': %s", edgeGatewayName, err)
	}

	bgpPrefixList, err := edgeGateway.GetBgpIpPrefixListByName(bgpIpPrefixListName)
	if err != nil {
		return nil, fmt.Errorf("[bgp ip prefix list import] unable to find BGP IP Prefix List with Name '%s': %s", bgpIpPrefixListName, err)
	}

	dSet(d, "org", orgName)
	dSet(d, "edge_gateway_id", edgeGateway.EdgeGateway.ID)
	d.SetId(bgpPrefixList.EdgeBgpIpPrefixList.ID)

	return []*schema.ResourceData{d}, nil
}

func getEdgeBgpIpPrefixListType(d *schema.ResourceData) *types.EdgeBgpIpPrefixList {
	bgpConfig := &types.EdgeBgpIpPrefixList{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
	}

	ipPrefixSet := d.Get("ip_prefix").(*schema.Set)
	ipPrefixSlice := make([]types.EdgeBgpConfigPrefixListPrefixes, len(ipPrefixSet.List()))
	for prefixIndex, prefix := range ipPrefixSet.List() {
		ipPrefixMap := prefix.(map[string]interface{})

		ipPrefixSlice[prefixIndex] = types.EdgeBgpConfigPrefixListPrefixes{
			Network:            ipPrefixMap["network"].(string),
			Action:             ipPrefixMap["action"].(string),
			GreaterThanEqualTo: ipPrefixMap["greater_than_or_equal_to"].(int),
			LessThanEqualTo:    ipPrefixMap["less_than_or_equal_to"].(int),
		}
	}

	bgpConfig.Prefixes = ipPrefixSlice

	return bgpConfig
}

func setEdgeBgpIpPrefixListData(d *schema.ResourceData, bgpConfig *types.EdgeBgpIpPrefixList) error {
	dSet(d, "name", bgpConfig.Name)
	dSet(d, "description", bgpConfig.Description)

	if len(bgpConfig.Prefixes) > 0 {
		ipPrefixSlice := make([]interface{}, len(bgpConfig.Prefixes))

		for prefixIndex, prefix := range bgpConfig.Prefixes {

			ipPrefixMap := make(map[string]interface{})
			ipPrefixMap["network"] = prefix.Network
			ipPrefixMap["action"] = prefix.Action
			ipPrefixMap["greater_than_or_equal_to"] = prefix.GreaterThanEqualTo
			ipPrefixMap["less_than_or_equal_to"] = prefix.LessThanEqualTo

			ipPrefixSlice[prefixIndex] = ipPrefixMap
		}

		ipPrefixSet := schema.NewSet(schema.HashResource(edgeBgpIpPrefixListHash), ipPrefixSlice)
		err := d.Set("ip_prefix", ipPrefixSet)
		if err != nil {
			return fmt.Errorf("error setting 'ip_prefix' block: %s", err)
		}
	}

	return nil
}
