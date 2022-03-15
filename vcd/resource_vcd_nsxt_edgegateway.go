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
				ConflictsWith: []string{"owner_id", "starting_vdc_id"},
				Deprecated:    "This field is deprecated in favor of 'owner_id' which supports both - VDC and VDC group IDs",
			},
			"owner_id": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				Description:   "ID of VDC or VDC Group",
				ConflictsWith: []string{"vdc"},
			},
			"starting_vdc_id": {
				Type:          schema.TypeString,
				Optional:      true,
				Description:   "Optional ID of starting VDC if the 'owner_id' is a VDC Group",
				ConflictsWith: []string{"vdc"},
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Edge Gateway name",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Edge Gateway description",
			},
			"dedicate_external_network": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Dedicating the External Network will enable Route Advertisement for this Edge Gateway.",
			},
			"external_network_id": {
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
			"edge_cluster_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Select specific NSX-T Edge Cluster. Will be inherited from external network if not specified",
			},
		},
	}
}

func resourceVcdNsxtEdgeGatewayCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("[TRACE] NSX-T Edge Gateway creation initiated")

	vcdClient := meta.(*VCDClient)

	adminOrg, err := vcdClient.GetAdminOrgFromResource(d)
	if err != nil {
		return diag.Errorf("error getting Org: %s", err)
	}

	nsxtEdgeGatewayType, err := getNsxtEdgeGatewayType(d, vcdClient, true)
	if err != nil {
		return diag.Errorf("could not create NSX-T Edge Gateway type: %s", err)
	}

	createdEdgeGateway, err := adminOrg.CreateNsxtEdgeGateway(nsxtEdgeGatewayType)
	if err != nil {
		return diag.Errorf("error creating NSX-T Edge Gateway: %s", err)
	}

	d.SetId(createdEdgeGateway.EdgeGateway.ID)

	// NSX-T Edge Gateway cannot be directly created in VDC group, but can only be assigned to VDC
	// Group after creation. Function `getNsxtEdgeGatewayType` decided the initial location of VDC,
	// but if the `owner_id` was set to VDC Group for creation - it must be moved to that VDC Group
	// explicitly after creation.
	ownerIdField := d.Get("owner_id").(string)
	if ownerIdField != "" && govcd.OwnerIsVdcGroup(ownerIdField) {
		log.Printf("[TRACE] NSX-T Edge Gateway update - 'owner_id' is specified and is VDC group. Moving it to VDC Group '%s'", ownerIdField)
		_, err := createdEdgeGateway.MoveToVdcOrVdcGroup(ownerIdField)
		if err != nil {
			return diag.Errorf("error assigning NSX-T Edge Gateway to VDC Group: %s", err)
		}
	}

	return resourceVcdNsxtEdgeGatewayRead(ctx, d, meta)
}

func resourceVcdNsxtEdgeGatewayUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("[TRACE] NSX-T Edge Gateway update initiated")

	// `vdc` field is deprecated. `vdc` value should not be changed unless it is removal of the
	// field at all to allow easy migration to `owner_id` path
	if _, new := d.GetChange("vdc"); d.HasChange("vdc") && new.(string) != "" {
		return diag.Errorf("changing 'vdc' field value is not supported. It can only be removed. " +
			"Please use `owner_id` field for moving Edge Gateway to/from VDC Group")
	}

	vcdClient := meta.(*VCDClient)
	adminOrg, err := vcdClient.GetAdminOrgFromResource(d)
	if err != nil {
		return diag.Errorf("error retrieving Org: %s", err)
	}

	edge, err := adminOrg.GetNsxtEdgeGatewayById(d.Id())
	if err != nil {
		return diag.Errorf("could not retrieve NSX-T Edge Gateway: %s", err)
	}

	updatedEdge, err := getNsxtEdgeGatewayType(d, vcdClient, false)
	if err != nil {
		return diag.Errorf("error updating NSX-T Edge Gateway type: %s", err)
	}

	updatedEdge.ID = edge.EdgeGateway.ID
	edge.EdgeGateway = updatedEdge

	_, err = edge.Update(edge.EdgeGateway)
	if err != nil {
		return diag.Errorf("error updating NSX-T Edge Gateway with ID '%s': %s", d.Id(), err)
	}

	return resourceVcdNsxtEdgeGatewayRead(ctx, d, meta)
}

func resourceVcdNsxtEdgeGatewayRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("[TRACE] NSX-T Edge Gateway read initiated")

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
		return diag.Errorf("could not retrieve NSX-T Edge Gateway: %s", err)
	}

	err = setNsxtEdgeGatewayData(edge.EdgeGateway, d)
	if err != nil {
		return diag.Errorf("error setting NSX-T Edge Gateway data: %s", err)
	}
	return nil
}

func resourceVcdNsxtEdgeGatewayDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	log.Printf("[TRACE] edge gateway deletion initiated")

	vcdClient := meta.(*VCDClient)
	org, err := vcdClient.GetOrgFromResource(d)
	if err != nil {
		return diag.Errorf("error retrieving Org: %s", err)
	}

	edge, err := org.GetNsxtEdgeGatewayById(d.Id())
	if err != nil {
		return diag.Errorf("could not retrieve NSX-T Edge Gateway: %s", err)
	}

	err = edge.Delete()
	if err != nil {
		return diag.Errorf("error deleting NSX-T Edge Gateway: %s", err)
	}

	return nil
}

func resourceVcdNsxtEdgeGatewayImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	log.Printf("[TRACE] NSX-T Edge Gateway import initiated")

	resourceURI := strings.Split(d.Id(), ImportSeparator)
	if len(resourceURI) != 3 {
		return nil, fmt.Errorf("resource name must be specified as org-name.vdc-name.nsxt-edge-gw-name or org-name.vdc-group-name.nsxt-edge-gw-name")
	}
	orgName, vdcName, edgeName := resourceURI[0], resourceURI[1], resourceURI[2]

	vcdClient := meta.(*VCDClient)

	// define an interface type to match VDC and VDC Groups
	var vdcOrGroup vdcOrVdcGroupVerifier
	_, vdcOrGroup, err := vcdClient.GetOrgAndVdc(orgName, vdcName)

	// VDC was not found - attempt to find a VDC Group
	if govcd.ContainsNotFound(err) {
		adminOrg, err := vcdClient.GetAdminOrg(orgName)
		if err != nil {
			return nil, fmt.Errorf("error retrieving Admin Org for '%s': %s", orgName, err)
		}

		vdcOrGroup, err = adminOrg.GetVdcGroupByName(vdcName)
		if err != nil {
			return nil, fmt.Errorf("error finding VDC or VDC Group by name '%s': %s", vdcName, err)
		}

	}

	if !vdcOrGroup.IsNsxt() {
		return nil, fmt.Errorf("please use 'vcd_edgegateway' for NSX-V backed VDC")
	}

	edge, err := vdcOrGroup.GetNsxtEdgeGatewayByName(edgeName)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve NSX-T Edge Gateway with ID '%s': %s", d.Id(), err)
	}

	// Only setting Org because VDC is a deprecated field. `owner_id` is set by resourceVcdNsxtEdgeGatewayRead by itself
	dSet(d, "org", orgName)

	d.SetId(edge.EdgeGateway.ID)

	return []*schema.ResourceData{d}, nil
}

// getNsxtEdgeGatewayType creates *types.OpenAPIEdgeGateway from Terraform schema
func getNsxtEdgeGatewayType(d *schema.ResourceData, vcdClient *VCDClient, isCreateOperation bool) (*types.OpenAPIEdgeGateway, error) {
	inheritedVdcField := vcdClient.Vdc
	vdcField := d.Get("vdc").(string)
	ownerIdField := d.Get("owner_id").(string)
	startingVdcId := d.Get("starting_vdc_id").(string)

	var ownerId string
	var err error

	if isCreateOperation {
		ownerId, err = getCreateOwnerId(d, vcdClient, ownerIdField, startingVdcId, vdcField, inheritedVdcField)
	}

	if !isCreateOperation {
		ownerId, err = getUpdateOwnerId(d, vcdClient, ownerIdField, startingVdcId, vdcField, inheritedVdcField)
	}

	if err != nil {
		return nil, err
	}

	edgeGatewayType := types.OpenAPIEdgeGateway{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		EdgeGatewayUplinks: []types.EdgeGatewayUplinks{{
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

// getCreateOwnerId defines how `owner_id` is defined for create operations
// Create has a few possible scenarios for getting ownerRef which can be a VDC or a VDC Group.
// NSX-T Edge Gateway cannot be created directly in VDC group
// * `owner_id` and `starting_vdc_id` are set. `owner_id` is a VDC Group
// * `owner_id` is set and is a VDC Group, but `starting_vdc_id` is not set (create operation must
// lookup any starting VDC in given VDC Group)
// * `vdc` field is set in the resource
// * Neither `vdc`, nor `owner_id` fields are set in the resource. `vdc` is inherited from `provider` section
//
// Note. Only one of `vdc` or `owner_id` (with optional `starting_vdc_id`) can be supplied. This is
// enforce by Terraform schema definition.
func getCreateOwnerId(d *schema.ResourceData, vcdClient *VCDClient, ownerIdField string, startingVdcId string, vdcField string, inheritedVdcField string) (string, error) {
	var ownerId string

	switch {
	// `owner_id` is specified and is VDC.
	case ownerIdField != "" && govcd.OwnerIsVdc(ownerIdField):
		ownerId = ownerIdField
	// `owner_id` is specified and is VDC Group. `starting_vdc_id` is specified.
	// Initial `owner_id` for create operation should be `starting_vdc_id` which is later going to
	// be moved to a VDC by a separate API call `createdEdgeGateway.MoveToVdcGroup`
	case ownerIdField != "" && govcd.OwnerIsVdcGroup(ownerIdField) && startingVdcId != "":
		log.Printf("[TRACE] NSX-T Edge Gateway create 'owner_id' field is set and is VDC group. 'starting_vdc_id' is set. Picking 'starting_vdc_id' for create operation")
		ownerId = startingVdcId
	// `owner_id` is specified and is VDC Group. `starting_vdc_id` is not specified.
	// NSX-T Edge Gateway cannot be created in VDC group therefore we are going to lookup random VDC in specified group
	case ownerIdField != "" && govcd.OwnerIsVdcGroup(ownerIdField) && startingVdcId == "":
		log.Printf("[TRACE] NSX-T Edge Gateway create 'owner_id' field is set and is VDC group. 'starting_vdc_id' is not set. Choosing random starting VDC")

		// Lookup Org
		org, err := vcdClient.GetOrgFromResource(d)
		if err != nil {
			return "", fmt.Errorf("error retrieving Org: %s", err)
		}

		vdcGroup, err := org.GetVdcGroupById(ownerIdField)
		if err != nil {
			return "", fmt.Errorf("error retrieving VDC group: %s", err)
		}

		if vdcGroup.VdcGroup != nil && len(vdcGroup.VdcGroup.ParticipatingOrgVdcs) > 0 {
			log.Printf("[TRACE] NSX-T Edge Gateway create 'owner_id' field is set and is VDC group. 'starting_vdc_id' is not set. Picked starting VDC '%s' (%s)",
				vdcGroup.VdcGroup.ParticipatingOrgVdcs[0].VdcRef.Name, vdcGroup.VdcGroup.ParticipatingOrgVdcs[0].VdcRef.ID)
			ownerId = vdcGroup.VdcGroup.ParticipatingOrgVdcs[0].VdcRef.ID
		}
	// `vdc` field is specified in the resource
	case vdcField != "":
		log.Printf("[TRACE] NSX-T Edge Gateway 'vdc' field is set in resource")

		org, err := vcdClient.GetOrgFromResource(d)
		if err != nil {
			return "", fmt.Errorf("error retrieving Org: %s", err)
		}

		vdc, err := org.GetVDCByName(vdcField, false)
		if err != nil {
			return "", fmt.Errorf("error finding VDC '%s': %s", vdcField, err)
		}

		ownerId = vdc.Vdc.ID
	// `vdc` field is not set in the resource itself, but is inherited from `provider`
	case inheritedVdcField != "" && vdcField == "" && ownerIdField == "":
		log.Printf("[TRACE] NSX-T Edge Gateway 'vdc' field is inherited from provider configuration. `vdc` and `owner_id` are not set in resource.")

		org, err := vcdClient.GetOrgFromResource(d)
		if err != nil {
			return "", fmt.Errorf("error retrieving Org: %s", err)
		}

		vdc, err := org.GetVDCByName(inheritedVdcField, false)
		if err != nil {
			return "", fmt.Errorf("error finding VDC '%s': %s", inheritedVdcField, err)
		}

		ownerId = vdc.Vdc.ID
	default:
		return "", fmt.Errorf("error looking up ownerId field owner_id='%s', vdc='%s', starting_vdc_id='%s', inherited vdc='%s'",
			ownerIdField, startingVdcId, vdcField, inheritedVdcField)
	}

	return ownerId, nil
}

// getUpdateOwnerId defines how `owner_id` is defined for update operations
// update behaves differently than create as `ownerRef` can have a VDC group set directly. A few
// scenarios can happen:
// * `owner_id` is set - using it
// * `vdc` field is set in the resource - using it
// * Neither `vdc`, nor `owner_id` fields are set in the resource. `vdc` is inherited from `provider` section
//
// Note. Only one of `vdc` or `owner_id` (with optional `starting_vdc_id`) can be supplied. This is
// enforce by Terraform schema definition.
func getUpdateOwnerId(d *schema.ResourceData, vcdClient *VCDClient, ownerIdField string, startingVdcId string, vdcField string, inheritedVdcField string) (string, error) {
	var ownerId string

	switch {
	case ownerIdField != "":
		log.Printf("[TRACE] NSX-T Edge Gateway update - 'owner_id' is set. Using it.")
		ownerId = ownerIdField

	case vdcField != "":
		log.Printf("[TRACE] NSX-T Edge Gateway update 'vdc' field is set in resource")

		org, err := vcdClient.GetOrgFromResource(d)
		if err != nil {
			return "", fmt.Errorf("error retrieving Org: %s", err)
		}

		vdc, err := org.GetVDCByName(vdcField, false)
		if err != nil {
			return "", fmt.Errorf("error finding VDC '%s': %s", vdcField, err)
		}

		ownerId = vdc.Vdc.ID
	case inheritedVdcField != "" && vdcField == "" && ownerIdField == "":
		log.Printf("[TRACE] NSX-T Edge Gateway update 'vdc' field is inherited from provider. `vdc` and `owner_id` are not set")

		org, err := vcdClient.GetOrgFromResource(d)
		if err != nil {
			return "", fmt.Errorf("error retrieving Org: %s", err)
		}

		vdc, err := org.GetVDCByName(inheritedVdcField, false)
		if err != nil {
			return "", fmt.Errorf("error finding VDC '%s': %s", inheritedVdcField, err)
		}

		ownerId = vdc.Vdc.ID

	default:
		return "", fmt.Errorf("error looking up ownerId field owner_id='%s', vdc='%s', starting_vdc_id='%s', inherited vdc='%s'",
			ownerIdField, startingVdcId, vdcField, inheritedVdcField)
	}
	return ownerId, nil
}

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

// setNsxtEdgeGatewayData stores Terraform schema from a read *types.OpenAPIEdgeGateway type
func setNsxtEdgeGatewayData(edgeGateway *types.OpenAPIEdgeGateway, d *schema.ResourceData) error {
	dSet(d, "name", edgeGateway.Name)
	dSet(d, "description", edgeGateway.Description)
	dSet(d, "edge_cluster_id", edgeGateway.EdgeClusterConfig.PrimaryEdgeCluster.BackingID)
	if len(edgeGateway.EdgeGatewayUplinks) < 1 {
		return fmt.Errorf("no edge gateway uplinks detected during read")
	}

	dSet(d, "owner_id", edgeGateway.OwnerRef.ID)
	dSet(d, "vdc", edgeGateway.OwnerRef.Name)

	// NSX-T Edge Gateways support only 1 uplink. Edge gateway can only be connected to one external network (in NSX-T terms
	// Tier 1 gateway can only be connected to single Tier 0 gateway)
	edgeUplink := edgeGateway.EdgeGatewayUplinks[0]

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
		return fmt.Errorf("error setting NSX-T Edge Gateway subnets after read: %s", err)
	}

	return nil
}

// vdcOrVdcGroupVerifier is an interface to access IsNsxt() and GetNsxtEdgeGatewayByName() on VDC or
// VDC Group method `IsNsxt` (used in isBackedByNsxt and resourceVcdNsxtEdgeGatewayImport)
type vdcOrVdcGroupVerifier interface {
	IsNsxt() bool
	GetNsxtEdgeGatewayByName(name string) (*govcd.NsxtEdgeGateway, error)
}
