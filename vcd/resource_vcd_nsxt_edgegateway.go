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
			Type:     schema.TypeString,
			Required: true,
		},
		"end_address": {
			Type:     schema.TypeString,
			Required: true,
		},
	},
}

var nsxtEdgeSubnet = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"gateway": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Gateway address for a subnet",
		},
		"prefix_length": {
			Type:        schema.TypeInt,
			Required:    true,
			Description: "Netmask address for a subnet (e.g. 24 for /24)",
		},
		"primary_ip": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Primary IP address for the edge gateway - will be auto-assigned if not defined",
		},
		"allocated_ips": {
			Type:        schema.TypeSet,
			Required:    true,
			MinItems:    1,
			Description: "Define one or more blocks to sub-allocate pools on the edge gateway",
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
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				Description:   "The name of VDC to use, optional if defined at provider level",
				ConflictsWith: []string{"owner_id", "starting_vdc_id"},
				Deprecated:    "This field is deprecated in favor of 'owner_id' which supports both - VDC and VDC group IDs",
			},
			"owner_id": &schema.Schema{
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				Description:   "ID of VDC group or VDC",
				ConflictsWith: []string{"vdc"},
			},
			"starting_vdc_id": &schema.Schema{
				Type:          schema.TypeString,
				Optional:      true,
				Description:   "ID of VDC group or VDC",
				ConflictsWith: []string{"vdc"},
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
				Type:        schema.TypeSet,
				Required:    true,
				Description: "One or more blocks with external network information to be attached to this gateway's interface",
				Elem:        nsxtEdgeSubnet,
			},
			"primary_ip": {
				Type:        schema.TypeString,
				Computed:    true,
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

func resourceVcdNsxtEdgeGatewayCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("[TRACE] NSX-T edge gateway creation initiated")

	vcdClient := meta.(*VCDClient)

	// Result should be owner ID field value
	adminOrg, err := vcdClient.GetAdminOrgFromResource(d)
	if err != nil {
		return diag.Errorf("error getting adminOrg: %s", err)
	}

	nsxtEdgeGatewayType, err := getNsxtEdgeGatewayType(d, vcdClient, true)
	if err != nil {
		return diag.Errorf("could not create NSX-T edge gateway type: %s", err)
	}

	createdEdgeGateway, err := adminOrg.CreateNsxtEdgeGateway(nsxtEdgeGatewayType)
	if err != nil {
		return diag.Errorf("error creating NSX-T edge gateway: %s", err)
	}

	d.SetId(createdEdgeGateway.EdgeGateway.ID)

	// NSX-T Edge Gateway cannot be directly created in VDC group, but can only be assigned to VDC group after creation
	ownerIdField := d.Get("owner_id").(string)
	if ownerIdField != "" && govcd.OwnerIsVdcGroup(ownerIdField) {
		_, err := createdEdgeGateway.MoveToVdcGroup(ownerIdField)
		if err != nil {
			return diag.Errorf("error assigning NSX-T Edge Gateway to VDC Group: %s", err)
		}
	}

	return resourceVcdNsxtEdgeGatewayRead(ctx, d, meta)
}

func resourceVcdNsxtEdgeGatewayUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("[TRACE] NSX-T edge gateway update initiated")

	// We do not allow changing `vdc` field value unless it is removal of the field at all
	if _, new := d.GetChange("vdc"); d.HasChange("vdc") && new.(string) != "" {
		return diag.Errorf("changing 'vdc' field value is not supported. It can only be removed. " +
			"Please use `owner_id` field for moving Edge Gateway to/from VDC Group")
	}

	vcdClient := meta.(*VCDClient)
	org, err := vcdClient.GetOrgFromResource(d)
	if err != nil {
		return diag.Errorf("error retrieving VDC: %s", err)
	}

	edge, err := org.GetNsxtEdgeGatewayById(d.Id())
	if err != nil {
		return diag.Errorf("could not retrieve NSX-T edge gateway: %s", err)
	}

	updatedEdge, err := getNsxtEdgeGatewayType(d, vcdClient, false)
	if err != nil {
		return diag.Errorf("error updating NSX-T edge gateway type: %s", err)
	}

	updatedEdge.ID = edge.EdgeGateway.ID
	edge.EdgeGateway = updatedEdge

	_, err = edge.Update(edge.EdgeGateway)
	if err != nil {
		return diag.Errorf("error updating edge gateway with ID '%s': %s", d.Id(), err)
	}

	return resourceVcdNsxtEdgeGatewayRead(ctx, d, meta)
}

func resourceVcdNsxtEdgeGatewayRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("[TRACE] NSX-T edge gateway read initiated")

	vcdClient := meta.(*VCDClient)

	org, err := vcdClient.GetOrgFromResource(d)
	if err != nil {
		return diag.Errorf("error retrieving VDC: %s", err)
	}

	edge, err := org.GetNsxtEdgeGatewayById(d.Id())
	if err != nil {
		if govcd.ContainsNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.Errorf("could not retrieve NSX-T edge gateway: %s", err)
	}

	err = setNsxtEdgeGatewayData(edge.EdgeGateway, d)
	if err != nil {
		return diag.Errorf("error reading NSX-T edge gateway data: %s", err)
	}
	return nil
}

func resourceVcdNsxtEdgeGatewayDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("[TRACE] edge gateway deletion initiated")

	vcdClient := meta.(*VCDClient)
	org, err := vcdClient.GetOrgFromResource(d)
	if err != nil {
		return diag.Errorf("error retrieving VDC: %s", err)
	}

	edge, err := org.GetNsxtEdgeGatewayById(d.Id())
	if err != nil {
		return diag.Errorf("could not retrieve NSX-T edge gateway: %s", err)
	}

	err = edge.Delete()
	if err != nil {
		return diag.Errorf("error deleting NSX-T edge gateway: %s", err)
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

	_, vdc, err := vcdClient.GetOrgAndVdc(orgName, vdcName)
	if err != nil {
		return nil, fmt.Errorf("unable to find org %s: %s", vdcName, err)
	}

	if vdc.IsNsxv() {
		return nil, fmt.Errorf("please use 'vcd_edgegateway' for NSX-V backed VDC")
	}

	edge, err := vdc.GetNsxtEdgeGatewayByName(edgeName)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve NSX-T edge gateway with ID '%s': %s", d.Id(), err)
	}

	dSet(d, "org", orgName)
	dSet(d, "vdc", vdcName)

	d.SetId(edge.EdgeGateway.ID)

	return []*schema.ResourceData{d}, nil
}

// getNsxtEdgeGatewayType
func getNsxtEdgeGatewayType(d *schema.ResourceData, vcdClient *VCDClient, isCreateOperation bool) (*types.OpenAPIEdgeGateway, error) {
	inheritedVdcField := vcdClient.Vdc
	vdcField := d.Get("vdc").(string)
	ownerIdField := d.Get("owner_id").(string)
	startingVdcId := d.Get("starting_vdc_id").(string)

	ownerId, err := getOwnerId(d, vcdClient, isCreateOperation, ownerIdField, startingVdcId, vdcField, inheritedVdcField)
	if err != nil {
		return nil, err
	}

	edgeGatewayType := types.OpenAPIEdgeGateway{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		EdgeGatewayUplinks: []types.EdgeGatewayUplinks{types.EdgeGatewayUplinks{
			UplinkID:  d.Get("external_network_id").(string),
			Subnets:   types.OpenAPIEdgeGatewaySubnets{Values: getNsxtEdgeGatewayUplinksType(d)},
			Dedicated: d.Get("dedicate_external_network").(bool),
		}},
		OwnerRef: &types.OpenApiReference{ID: ownerId},
	}

	// Optional edge_cluster_id
	if clusterId, isSet := d.GetOk("edge_cluster_id"); isSet {
		edgeGatewayType.EdgeClusterConfig = &types.OpenAPIEdgeGatewayEdgeClusterConfig{
			PrimaryEdgeCluster: types.OpenAPIEdgeGatewayEdgeCluster{
				BackingID: clusterId.(string),
			},
		}
	}

	return &edgeGatewayType, nil
}

// getOwnerId looks up correct value for `owner_id`
//
// With the introduction of VDC group support handling VDC reference becomes
// 3 major combinations possible with `owner_id` having some sub-combinations. They also differ for `create` and `update` because
// Edge Gateway cannot be created directly in VDC Group.
// * `vdc` or `owner_id` fields are not set (inherited from `provider` section)
// * `vdc` field is specified
// * `owner_id` field is specified
//   * `owner_id` is VDC
//   * `owner_id` is VDC Group
//     * `owner_id` is VDC Group and `starting_vdc_id` is possibly set
// Whenever owner_id field is set - it takes priority over `vdc` field (set in resource or inherited from provider)
func getOwnerId(d *schema.ResourceData, vcdClient *VCDClient, isCreateOperation bool, ownerIdField string, startingVdcId string, vdcField string, inheritedVdcField string) (string, error) {
	var ownerId string
	switch {
	// Create operation
	// `owner_id` is specified and is VDC Group
	// `starting_vdc_id` is specified
	case isCreateOperation && ownerIdField != "" && govcd.OwnerIsVdcGroup(ownerIdField) && startingVdcId != "":
		log.Printf("[TRACE] NSX-T edge gateway create 'owner_id' field is set and is VDC group. 'starting_vdc_id' is set")
		ownerId = startingVdcId
	// Update operation
	// `owner_id` is specified and is VDC Group. It does not matter if `starting_vdc_id` is specified or not
	case !isCreateOperation && ownerIdField != "" && govcd.OwnerIsVdcGroup(ownerIdField) && startingVdcId != "":
		log.Printf("[TRACE] NSX-T edge gateway update 'owner_id' field is set and is VDC group.")
		ownerId = startingVdcId
	// Create operation
	// `owner_id` is specified and is VDC Group. `starting_vdc_id` is not specified.
	// NSX-T Edge Gateway cannot be created in VDC group therefore we are going to lookup random VDC
	case isCreateOperation && ownerIdField != "" && govcd.OwnerIsVdcGroup(ownerIdField) && startingVdcId == "":
		log.Printf("[TRACE] NSX-T edge gateway create 'owner_id' field is set and is VDC group. 'starting_vdc_id' is not set. Choosing random starting VDC")

		// Lookup Org
		adminOrg, err := vcdClient.GetAdminOrgFromResource(d)
		if err != nil {
			return "", fmt.Errorf("error retrieving Org: %s", err)
		}

		vdcGroup, err := adminOrg.GetVdcGroupById(ownerIdField)
		if err != nil {
			return "", fmt.Errorf("error retrieving VDC group: %s", err)
		}

		if vdcGroup.VdcGroup != nil && len(vdcGroup.VdcGroup.ParticipatingOrgVdcs) > 0 {
			ownerId = vdcGroup.VdcGroup.ParticipatingOrgVdcs[0].VdcRef.ID
		}
	// Update operation
	// `owner_id` is set
	case !isCreateOperation && ownerIdField != "" && govcd.OwnerIsVdcGroup(ownerIdField) && startingVdcId == "":
		log.Printf("[TRACE] NSX-T edge gateway update 'owner_id' field is set and is VDC group. 'starting_vdc_id' is not set. Choosing random starting VDC")

		ownerId = ownerIdField
	//case !isCreateOperation
	//case ownerIdField != "" && govcd.OwnerIsVdc(ownerIdField):
	//	log.Printf("[TRACE] NSX-T edge gateway 'owner_id' field is set and is VDC")
	//	ownerId = ownerIdField
	case vdcField != "":
		log.Printf("[TRACE] NSX-T edge gateway 'vdc' field is set only in resource")

		adminOrg, err := vcdClient.GetAdminOrgFromResource(d)
		if err != nil {
			return "", fmt.Errorf("error retrieving Org: %s", err)
		}

		vdc, err := adminOrg.GetVDCByName(vdcField, false)
		if err != nil {
			return "", fmt.Errorf("error finding VDC '%s': %s", vdcField, err)
		}

		ownerId = vdc.Vdc.ID
	case inheritedVdcField != "" && vdcField == "" && ownerIdField == "":
		log.Printf("[TRACE] NSX-T edge gateway 'vdc' field is inherited from provider. `vdc` and `owner_id` are not set")

		adminOrg, err := vcdClient.GetAdminOrgFromResource(d)
		if err != nil {
			return "", fmt.Errorf("error retrieving Org: %s", err)
		}

		vdc, err := adminOrg.GetVDCByName(inheritedVdcField, false)
		if err != nil {
			return "", fmt.Errorf("error finding VDC '%s': %s", inheritedVdcField, err)
		}

		ownerId = vdc.Vdc.ID

	default:
		return "", fmt.Errorf("error looking up ownerId field")
	}
	return ownerId, nil
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
			singleSubnet.IPRanges = &types.OpenApiIPRanges{Values: ipRanges}
		}

		subnetSlice[index] = singleSubnet
	}

	// VCD API is very odd in how it assigns primary_ip. The defined subnet having primary_ip must be sent to API as
	// first item in JSON list therefore if `primary_ip` was specified in other item than first one must shuffle slice
	// elements so that the one with primary_ip is first.
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
	dSet(d, "name", e.Name)
	dSet(d, "description", e.Description)
	dSet(d, "edge_cluster_id", e.EdgeClusterConfig.PrimaryEdgeCluster.BackingID)
	if len(e.EdgeGatewayUplinks) < 1 {
		return fmt.Errorf("no edge gateway uplinks detected during read")
	}

	dSet(d, "owner_id", e.OwnerRef.ID)

	// NSX-T edge gateways support only 1 uplink. Edge gateway can only be connected to one external network (in NSX-T terms
	// Tier 1 gateway can only be connected to single Tier 0 gateway)
	edgeUplink := e.EdgeGatewayUplinks[0]

	dSet(d, "dedicate_external_network", edgeUplink.Dedicated)
	dSet(d, "external_network_id", edgeUplink.UplinkID)

	// subnets
	subnets := make([]interface{}, 1)
	for _, subnetValue := range edgeUplink.Subnets.Values {

		// Edge Gateway API returns all subnets defined on external network. However, if they don't have "ranges"
		// defined - it means they are not allocated to edge gateway and Terraform should not display it as UI does not
		// display them as well
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
			dSet(d, "primary_ip", subnetValue.PrimaryIP)
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
