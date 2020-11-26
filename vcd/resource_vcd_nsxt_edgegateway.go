package vcd

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

var nsxtEdgeSubnetRange = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"start_address": {
			Required: true,
			Type:     schema.TypeString,
		},
		"end_address": {
			Required: true,
			Type:     schema.TypeString,
		},
	},
}

var nsxtEdgeSubnet = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"gateway": {
			Required:    true,
			Description: "Gateway address for a subnet",
			Type:        schema.TypeString,
		},
		"prefix_length": {
			Required:    true,
			Description: "Netmask address for a subnet (e.g. 24 for /24)",
			Type:        schema.TypeInt,
		},
		"primary_ip": {
			Optional:    true,
			Type:        schema.TypeString,
			Description: "Primary IP address for the edge gateway - will be auto-assigned if not defined",
		},
		"allocated_ips": {
			Optional:    true,
			Computed:    true,
			Type:        schema.TypeSet,
			Description: "Define zero or more blocks to sub-allocate pools on the edge gateway",
			Elem:        nsxtEdgeSubnetRange,
		},
	},
}

func resourceVcdNsxtEdgeGateway() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdNsxtEdgeGatewayCreate,
		ReadContext:   resourceVcdNsxtEdgeGatewayRead,
		UpdateContext: resourceVcdNsxtEdgeGatewayUpdate,
		DeleteContext: resourceVcdNsxtEdgeGatewayDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdNsxtEdgeGatewayImport,
		},

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
				Optional:    true,
				Description: "Edge Gateway description",
			},
			"dedicate_external_network": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Dedicating the External Network will enable Route Advertisement for this Edge Gateway.",
			},
			"external_network_id": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "External network ID",
			},
			"subnet": {
				Description: "One or more blocks with external network information to be attached to this gateway's interface",
				Required:    true,
				Type:        schema.TypeSet,
				Elem:        nsxtEdgeSubnet,
			},
			"primary_ip": {
				Computed:    true,
				Type:        schema.TypeString,
				Description: "Primary IP address of edge gateway. Read-only (can be specified in specific subnet)",
			},
			"edge_cluster_id": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Select specific NSX-T Edge Cluster. Will be inherited from external network if not specified",
			},
		},
	}
}

// resourceVcdNsxtEdgeGatewayCreate
func resourceVcdNsxtEdgeGatewayCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("[TRACE] NSX-T edge gateway creation initiated")

	vcdClient := meta.(*VCDClient)
	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error retrieving VDC: %s", err))
	}

	adminOrg, err := vcdClient.GetAdminOrgFromResource(d)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error getting adminOrg: %s", err))
	}

	nsxtEdgeGatewayType, err := getNsxtEdgeGatewayType(d, vdc)
	if err != nil {
		return diag.FromErr(fmt.Errorf("could not create NSX-T edge gateway type: %s", err))
	}

	createdEdgeGateway, err := adminOrg.CreateNsxtEdgeGateway(nsxtEdgeGatewayType)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating NSX-T edge gateway: %s", err))
	}

	d.SetId(createdEdgeGateway.EdgeGateway.ID)

	return resourceVcdNsxtEdgeGatewayRead(ctx, d, meta)
}

// resourceVcdNsxtEdgeGatewayUpdate
func resourceVcdNsxtEdgeGatewayUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("[TRACE] NSX-T edge gateway update initiated")

	vcdClient := meta.(*VCDClient)
	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error retrieving VDC: %s", err))
	}

	adminOrg, err := vcdClient.GetAdminOrgFromResource(d)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error getting adminOrg: %s", err))
	}

	edge, err := adminOrg.GetNsxtEdgeGatewayById(d.Id())
	if err != nil {
		return diag.FromErr(fmt.Errorf("could not retrieve NSX-T edge gateway: %s", err))
	}

	updatedEdge, err := getNsxtEdgeGatewayType(d, vdc)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error updating NSX-T edge gateway type: %s", err))
	}

	updatedEdge.ID = edge.EdgeGateway.ID
	edge.EdgeGateway = updatedEdge

	_, err = edge.Update(edge.EdgeGateway)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error updating edge gateway with ID '%s': %s", d.Id(), err))
	}

	return resourceVcdNsxtEdgeGatewayRead(ctx, d, meta)
}

// resourceVcdNsxtEdgeGatewayRead
func resourceVcdNsxtEdgeGatewayRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("[TRACE] NSX-T edge gateway read initiated")

	vcdClient := meta.(*VCDClient)

	adminOrg, err := vcdClient.GetAdminOrgFromResource(d)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error getting adminOrg: %s", err))
	}

	edge, err := adminOrg.GetNsxtEdgeGatewayById(d.Id())
	if err != nil {
		return diag.FromErr(fmt.Errorf("could not retrieve NSX-T edge gateway: %s", err))
	}

	err = setNsxtEdgeGatewayData(edge.EdgeGateway, d)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading NSX-T edge gateway data: %s", err))
	}
	return nil
}

// resourceVcdNsxtEdgeGatewayDelete
func resourceVcdNsxtEdgeGatewayDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("[TRACE] edge gateway deletion initiated")

	vcdClient := meta.(*VCDClient)

	adminOrg, err := vcdClient.GetAdminOrgFromResource(d)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error getting adminOrg: %s", err))
	}

	edge, err := adminOrg.GetNsxtEdgeGatewayById(d.Id())
	if err != nil {
		return diag.FromErr(fmt.Errorf("could not retrieve NSX-T edge gateway: %s", err))
	}

	err = edge.Delete()
	if err != nil {
		return diag.FromErr(fmt.Errorf("error deleting NSX-T edge gateway: %s", err))
	}

	return nil
}

// resourceVcdNsxtEdgeGatewayImport
func resourceVcdNsxtEdgeGatewayImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	log.Printf("[TRACE] NSX-T edge gateway import initiated")

	resourceURI := strings.Split(d.Id(), ImportSeparator)
	if len(resourceURI) != 3 {
		return nil, fmt.Errorf("resource name must be specified as org-name.vdc-name.nsxt-edge-gw-name")
	}
	orgName, vdcName, edgeName := resourceURI[0], resourceURI[1], resourceURI[2]

	vcdClient := meta.(*VCDClient)
	adminOrg, err := vcdClient.GetAdminOrg(orgName)
	if err != nil {
		return nil, fmt.Errorf("unable to find org %s: %s", orgName, err)
	}

	edge, err := adminOrg.GetNsxtEdgeGatewayByName(edgeName)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve NSX-T edge gateway with ID '%s': %s", d.Id(), err)
	}

	_ = d.Set("org", orgName)
	_ = d.Set("vdc", vdcName)

	d.SetId(edge.EdgeGateway.ID)

	return []*schema.ResourceData{d}, nil
}

// getNsxtEdgeGatewayType
func getNsxtEdgeGatewayType(d *schema.ResourceData, vdc *govcd.Vdc) (*types.OpenAPIEdgeGateway, error) {
	e := types.OpenAPIEdgeGateway{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		EdgeGatewayUplinks: []types.EdgeGatewayUplinks{types.EdgeGatewayUplinks{
			UplinkID:  d.Get("external_network_id").(string),
			Subnets:   types.OpenAPIEdgeGatewaySubnets{getNsxtEdgeGatewayUplinksType(d)},
			Dedicated: d.Get("dedicate_external_network").(bool),
		}},
		// OrgVdcNetworkCount:        0,
		// OpenAPIEdgeGatewayBacking: types.OpenAPIEdgeGatewayBacking{},
		OrgVdc: &types.OpenApiReference{
			ID: vdc.Vdc.ID,
		},
	}

	// Optional edge_cluster_id
	if clusterId, isSet := d.GetOk("edge_cluster_id"); isSet {
		e.EdgeClusterConfig = &types.OpenAPIEdgeGatewayEdgeClusterConfig{
			PrimaryEdgeCluster: types.OpenAPIEdgeGatewayEdgeCluster{
				BackingID: clusterId.(string),
			},
		}
	}

	return &e, nil
}

// getNsxtEdgeGatewayUplinksType
func getNsxtEdgeGatewayUplinksType(d *schema.ResourceData) []types.OpenAPIEdgeGatewaySubnetValue {

	var isPrimaryIpSet bool
	var primaryIpIndex int

	extNetworks := d.Get("subnet").(*schema.Set).List()
	subnetSlice := make([]types.OpenAPIEdgeGatewaySubnetValue, len(extNetworks))

	for index, singleSubnet := range extNetworks {
		subnetMap := singleSubnet.(map[string]interface{})
		singleSubnet := types.OpenAPIEdgeGatewaySubnetValue{
			Gateway:      subnetMap["gateway"].(string),
			PrefixLength: subnetMap["prefix_length"].(int),
			PrimaryIP:    subnetMap["primary_ip"].(string),
		}

		if subnetMap["primary_ip"].(string) != "" {
			isPrimaryIpSet = true
			primaryIpIndex = index
		}

		// Only feed in ip range allocations if they are defined
		if ipRanges := getNsxtEdgeGatewayUplinkRangeTypes(subnetMap); ipRanges != nil {
			singleSubnet.IPRanges = &types.OpenApiIPRanges{ipRanges}
		}

		subnetSlice[index] = singleSubnet
	}

	// VCD API is very odd in how it assigns primary_ip. The defined subnet having primary_ip must be sent to API as
	// first item in JSON list therefore if `primary_ip` was specified one must shuffle slice elements so that the one
	// with primary_ip is first.
	// The order does not really matter for Terraform schema as TypeSet is used, but user must get expected primary_ip.
	if isPrimaryIpSet {
		subnetZero := subnetSlice[0]
		subnetSlice[0] = subnetSlice[primaryIpIndex]
		subnetSlice[primaryIpIndex] = subnetZero
	}

	return subnetSlice
}

// getNsxtEdgeGatewayUplinkRangeTypes
func getNsxtEdgeGatewayUplinkRangeTypes(subnetMap map[string]interface{}) []types.OpenApiIPRangeValues {
	suballocatePoolSchema := subnetMap["allocated_ips"].(*schema.Set)
	subnetRanges := make([]types.OpenApiIPRangeValues, len(suballocatePoolSchema.List()))

	if len(subnetRanges) == 0 {
		return nil
	}

	for rangeIndex, subnetRange := range suballocatePoolSchema.List() {
		subnetRangeStr := convertToStringMap(subnetRange.(map[string]interface{}))
		oneRange := types.OpenApiIPRangeValues{
			StartAddress: subnetRangeStr["start_address"],
			EndAddress:   subnetRangeStr["end_address"],
		}
		subnetRanges[rangeIndex] = oneRange
	}
	return subnetRanges
}

// setNsxtEdgeGatewayData sets schema
func setNsxtEdgeGatewayData(e *types.OpenAPIEdgeGateway, d *schema.ResourceData) error {
	_ = d.Set("name", e.Name)
	_ = d.Set("description", e.Description)
	_ = d.Set("edge_cluster_id", e.EdgeClusterConfig.PrimaryEdgeCluster.BackingID)
	if len(e.EdgeGatewayUplinks) < 1 {
		return fmt.Errorf("no edge gateway uplinks detected during read")
	}

	// NSX-T edge gateways support only 1 uplink. Edge gateway can only be connected to one external network (in NSX-T terms
	// Tier 1 gateway can only be connected to single Tier 0 gateway)
	edgeUplink := e.EdgeGatewayUplinks[0]

	_ = d.Set("dedicate_external_network", edgeUplink.Dedicated)
	_ = d.Set("external_network_id", edgeUplink.UplinkID)

	// subnets
	subnets := make([]interface{}, 1)
	for _, subnetValue := range edgeUplink.Subnets.Values {

		// Edge Gateway API returns all subnets defined on external network. However if they don't have "ranges" defined
		// - it means they are not allocated to edge gateway and Terraform should not display it as UI does not display
		// them as well
		ipRangeCount := len(subnetValue.IPRanges.Values)
		if ipRangeCount == 0 {
			continue
		}

		oneSubnet := make(map[string]interface{})

		oneSubnet["gateway"] = subnetValue.Gateway
		oneSubnet["prefix_length"] = subnetValue.PrefixLength

		// If primary IP exists - set it to schema and computed variable at the top level for easier access
		if subnetValue.PrimaryIP != "" {
			oneSubnet["primary_ip"] = subnetValue.PrimaryIP
			_ = d.Set("primary_ip", subnetValue.PrimaryIP)
		}

		// Check for allocated IPs
		allIpRanges := make([]interface{}, ipRangeCount)
		for ipRangeIndex, ipRangeValue := range subnetValue.IPRanges.Values {
			oneIpRange := make(map[string]interface{})
			oneIpRange["start_address"] = ipRangeValue.StartAddress
			oneIpRange["end_address"] = ipRangeValue.EndAddress

			allIpRanges[ipRangeIndex] = oneIpRange
		}

		ipRangeSet := schema.NewSet(schema.HashResource(nsxtEdgeSubnetRange), allIpRanges)
		oneSubnet["allocated_ips"] = ipRangeSet
		subnets = append(subnets, oneSubnet)
	}

	subnetSet := schema.NewSet(schema.HashResource(nsxtEdgeSubnet), subnets)

	err := d.Set("subnet", subnetSet)
	if err != nil {
		return fmt.Errorf("error setting subnets after read: %s", err)
	}

	return nil
}
