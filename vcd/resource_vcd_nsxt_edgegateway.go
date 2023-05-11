package vcd

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"github.com/vmware/go-vcloud-director/v2/util"
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
			Description: "Prefix length for a subnet (e.g. 24)",
		},
		"primary_ip": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Primary IP address for the edge gateway - will be auto-assigned if not defined",
		},
		"allocated_ips": {
			Type:        schema.TypeSet,
			Optional:    true,
			Computed:    true,
			Description: "Define one or more blocks to sub-allocate pools on the edge gateway",
			Elem:        nsxtEdgeSubnetRange,
		},
	},
}

var nsxtEdgeAutoSubnetAndTotal = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"gateway": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Gateway address for a subnet",
		},
		"primary_ip": {
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			Description: "Primary IP address for the edge gateway - will be auto-assigned if not defined",
		},
		"prefix_length": {
			Type:        schema.TypeInt,
			Required:    true,
			Description: "Prefix length for a subnet (e.g. 24)",
		},
	},
}

var nsxtEdgeAutoAllocatedSubnet = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"gateway": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "Gateway address for a subnet",
		},
		"prefix_length": {
			Type:        schema.TypeInt,
			Required:    true,
			Description: "Prefix length for a subnet (e.g. 24)",
		},
		"primary_ip": {
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			Description: "Primary IP address for the edge gateway - will be auto-assigned if not defined",
		},
		"allocated_ip_count": {
			Type:        schema.TypeInt,
			Required:    true,
			Description: "Number of IP addresses to allocate",
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
				Deprecated:    "This field is deprecated in favor of 'owner_id' which supports both - VDC and VDC Group IDs",
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
			"total_allocated_ip_count": {
				Type:          schema.TypeInt,
				Computed:      true,
				Optional:      true,
				Description:   "Total number of IP addresses allocated for this gateway. Can be set with 'subnet_with_total_ip_count' definitions only",
				RequiredWith:  []string{"subnet_with_total_ip_count"},
				ConflictsWith: []string{"subnet", "subnet_with_ip_count"},
				ValidateFunc:  validation.IntAtLeast(1),
			},
			"subnet_with_total_ip_count": {
				Type:          schema.TypeSet,
				Optional:      true,
				Computed:      true,
				Description:   "Subnet definitions for this Edge Gateway. IP allocation is controlled using 'total_allocated_ip_count'",
				Elem:          nsxtEdgeAutoSubnetAndTotal,
				RequiredWith:  []string{"total_allocated_ip_count"},
				ConflictsWith: []string{"subnet", "subnet_with_ip_count"},
				AtLeastOneOf:  []string{"subnet_with_total_ip_count", "subnet", "subnet_with_ip_count"},
			},
			"subnet_with_ip_count": {
				Type:          schema.TypeSet,
				Optional:      true,
				Computed:      true,
				Description:   "Auto allocation of subnets by using per subnet IP allocation counts",
				Elem:          nsxtEdgeAutoAllocatedSubnet,
				ConflictsWith: []string{"subnet", "subnet_with_total_ip_count"},
				AtLeastOneOf:  []string{"subnet_with_total_ip_count", "subnet", "subnet_with_ip_count"},
			},
			"subnet": {
				Type:          schema.TypeSet,
				Optional:      true,
				Computed:      true,
				Description:   "One or more blocks with external network information to be attached to this gateway's interface including IP allocation ranges",
				Elem:          nsxtEdgeSubnet,
				ConflictsWith: []string{"subnet_with_total_ip_count", "total_allocated_ip_count", "subnet_with_ip_count"},
				AtLeastOneOf:  []string{"subnet_with_total_ip_count", "subnet", "subnet_with_ip_count"},
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

	nsxtEdgeGatewayType, err := getNsxtEdgeGatewayType(d, vcdClient, true, nil, nil)
	if err != nil {
		return diag.Errorf("could not create NSX-T Edge Gateway type: %s", err)
	}

	createdEdgeGateway, err := adminOrg.CreateNsxtEdgeGateway(nsxtEdgeGatewayType)
	if err != nil {
		return diag.Errorf("error creating NSX-T Edge Gateway: %s", err)
	}

	d.SetId(createdEdgeGateway.EdgeGateway.ID)

	// NSX-T Edge Gateway cannot be directly created in VDC Group, but can only be assigned to VDC
	// Group after creation. Function `getNsxtEdgeGatewayType` decided the initial location of VDC,
	// but if the `owner_id` was set to VDC Group for creation - it must be moved to that VDC Group
	// explicitly after creation.
	ownerIdField := d.Get("owner_id").(string)
	if ownerIdField != "" && govcd.OwnerIsVdcGroup(ownerIdField) {
		log.Printf("[TRACE] NSX-T Edge Gateway update - 'owner_id' is specified and is VDC Group. Moving it to VDC Group '%s'", ownerIdField)
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

	allocatedIpCount, err := edge.GetAllocatedIpCount(false)
	if err != nil {
		return diag.Errorf("could not retrieve NSX-T Edge Gateway allocated IP count: %s", err)
	}

	updatedEdge, err := getNsxtEdgeGatewayType(d, vcdClient, false, &allocatedIpCount, edge)
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

func resourceVcdNsxtEdgeGatewayRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	err = setNsxtEdgeGatewayData(edge, d)
	if err != nil {
		return diag.Errorf("error setting NSX-T Edge Gateway data: %s", err)
	}
	return nil
}

func resourceVcdNsxtEdgeGatewayDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

func resourceVcdNsxtEdgeGatewayImport(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	log.Printf("[TRACE] NSX-T Edge Gateway import initiated")

	resourceURI := strings.Split(d.Id(), ImportSeparator)
	if len(resourceURI) != 3 {
		return nil, fmt.Errorf("resource name must be specified as org-name.vdc-name.nsxt-edge-gw-name or org-name.vdc-group-name.nsxt-edge-gw-name")
	}
	orgName, vdcOrVdcGroupName, edgeName := resourceURI[0], resourceURI[1], resourceURI[2]

	vcdClient := meta.(*VCDClient)
	vdcOrVdcGroup, err := lookupVdcOrVdcGroup(vcdClient, orgName, vdcOrVdcGroupName)
	if err != nil {
		return nil, err
	}

	if !vdcOrVdcGroup.IsNsxt() {
		return nil, fmt.Errorf("please use 'vcd_edgegateway' for NSX-V backed VDC")
	}

	edge, err := vdcOrVdcGroup.GetNsxtEdgeGatewayByName(edgeName)
	if err != nil {
		return nil, fmt.Errorf("could not retrieve NSX-T Edge Gateway with ID '%s': %s", d.Id(), err)
	}

	// Only setting Org because VDC is a deprecated field. `owner_id` is set by resourceVcdNsxtEdgeGatewayRead by itself
	dSet(d, "org", orgName)

	d.SetId(edge.EdgeGateway.ID)

	return []*schema.ResourceData{d}, nil
}

// getNsxtEdgeGatewayType creates *types.OpenAPIEdgeGateway from Terraform schema
func getNsxtEdgeGatewayType(d *schema.ResourceData, vcdClient *VCDClient, isCreateOperation bool, allocatedIpCount *int, edgeGateway *govcd.NsxtEdgeGateway) (*types.OpenAPIEdgeGateway, error) {
	inheritedVdcField := vcdClient.Vdc
	vdcField := d.Get("vdc").(string)
	ownerIdField := d.Get("owner_id").(string)
	startingVdcId := d.Get("starting_vdc_id").(string)

	var ownerId string
	var err error

	isUpdateOperation := !isCreateOperation
	if isCreateOperation {
		ownerId, err = getCreateOwnerIdWithStartingVdcId(d, vcdClient, ownerIdField, startingVdcId, vdcField, inheritedVdcField)
	}

	if isUpdateOperation {
		ownerId, err = getOwnerId(d, vcdClient, ownerIdField, vdcField, inheritedVdcField)
	}

	if err != nil {
		return nil, err
	}

	edgeGatewayType := types.OpenAPIEdgeGateway{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		OwnerRef:    &types.OpenApiReference{ID: ownerId},
	}

	// Optional edge_cluster_id
	if clusterId, isSet := d.GetOk("edge_cluster_id"); isSet {
		edgeGatewayType.EdgeClusterConfig = &types.OpenAPIEdgeGatewayEdgeClusterConfig{
			PrimaryEdgeCluster: types.OpenAPIEdgeGatewayEdgeCluster{
				BackingID: clusterId.(string),
			},
		}
	}

	// Handle uplink and IP allocations for create and update in separate functions
	switch {
	case isCreateOperation:
		edgeGatewayType.EdgeGatewayUplinks, err = getNsxtEdgeGatewayUplinksTypeForCreate(d)
		if err != nil {
			return nil, err
		}
	case isUpdateOperation:
		edgeGatewayType.EdgeGatewayUplinks, err = getNsxtEdgeGatewayUplinksTypeForUpdate(d, allocatedIpCount, edgeGateway)
		if err != nil {
			return nil, err
		}
	}

	// Set common uplink values
	edgeGatewayType.EdgeGatewayUplinks[0].UplinkID = d.Get("external_network_id").(string)
	edgeGatewayType.EdgeGatewayUplinks[0].Dedicated = d.Get("dedicate_external_network").(bool)

	return &edgeGatewayType, nil
}

// getNsxtEdgeGatewayUplinksTypeForCreate handles uplink structure in create only operations
func getNsxtEdgeGatewayUplinksTypeForCreate(d *schema.ResourceData) ([]types.EdgeGatewayUplinks, error) {
	_, usingSubnetAllocation := d.GetOk("subnet")
	_, usingAutoSubnetAllocation := d.GetOk("subnet_with_total_ip_count")
	_, usingAutoAllocatedSubnetAllocation := d.GetOk("subnet_with_ip_count")

	log.Printf("[TRACE] NSX-T Edge Gateway creation 'subnet': %t (HasChange %t), 'subnet_with_total_ip_count': %t (HasChange %t), 'subnet_with_ip_count': %t (HasChange %t)",
		usingSubnetAllocation, d.HasChange("subnet"), usingAutoSubnetAllocation, d.HasChange("subnet_with_total_ip_count"), usingAutoAllocatedSubnetAllocation, d.HasChange("subnet_with_ip_count"))

	switch {
	// 'subnet' is specified
	case usingSubnetAllocation:
		return []types.EdgeGatewayUplinks{{Subnets: types.OpenAPIEdgeGatewaySubnets{Values: getNsxtEdgeGatewayUplinksType(d)}}}, nil

	// 'subnet_with_total_ip_count' and 'total_allocated_ip_count' are specified
	case usingAutoSubnetAllocation:
		isPrimaryIpSet, subnetValues := getNsxtEdgeGatewayUplinksTypeAutoSubnets(d)

		uplinks := []types.EdgeGatewayUplinks{{}}
		uplinks[0].Subnets = types.OpenAPIEdgeGatewaySubnets{Values: subnetValues}
		if totalIpCount, isSetTotalIpCount := d.GetOk("total_allocated_ip_count"); isSetTotalIpCount {

			// For create operation, the total_allocated_ip_count cannot be negative and therefore we can utilize QuickAddAllocatedIPCount field
			uplinks[0].QuickAddAllocatedIPCount = totalIpCount.(int)
			// Primary IP is an additional IP address that is not included in the total allocated IP
			// count when used with QuickAddAllocatedIPCount therefore we need to subtract it
			if isPrimaryIpSet {
				uplinks[0].QuickAddAllocatedIPCount = totalIpCount.(int) - 1
			}

		}
		return uplinks, nil

	// 'subnet_with_ip_count' is specified
	case usingAutoAllocatedSubnetAllocation:
		return []types.EdgeGatewayUplinks{{Subnets: types.OpenAPIEdgeGatewaySubnets{Values: getNsxtEdgeGatewayUplinksTypeAutoAllocateSubnets(d)}}}, nil

	}

	return nil, fmt.Errorf("one of the following fields must be set: 'subnet', 'subnet_with_total_ip_count', 'subnet_with_ip_count'")
}

// getNsxtEdgeGatewayUplinksTypeForUpdate handles uplink structure in update only operations
func getNsxtEdgeGatewayUplinksTypeForUpdate(d *schema.ResourceData, currentlyAllocatedIpCount *int, edgeGateway *govcd.NsxtEdgeGateway) ([]types.EdgeGatewayUplinks, error) {
	if edgeGateway == nil {
		return nil, fmt.Errorf("edge gateway cannot be nil")
	}

	if currentlyAllocatedIpCount == nil {
		return nil, fmt.Errorf("currentlyAllocatedIpCount cannot be nil for update operation")
	}

	_, usingSubnetAllocation := d.GetOk("subnet")
	_, usingAutoSubnetAllocation := d.GetOk("subnet_with_total_ip_count")
	_, usingAutoAllocatedSubnetAllocation := d.GetOk("subnet_with_ip_count")

	log.Printf("[TRACE] NSX-T Edge Gateway update 'subnet': %t (HasChange %t), 'subnet_with_total_ip_count': %t (HasChange %t), 'subnet_with_ip_count': %t (HasChange %t)",
		usingSubnetAllocation, d.HasChange("subnet"), usingAutoSubnetAllocation, d.HasChange("subnet_with_total_ip_count"), usingAutoAllocatedSubnetAllocation, d.HasChange("subnet_with_ip_count"))

	switch {
	// 'subnet' is specified
	case d.HasChange("subnet"):
		return []types.EdgeGatewayUplinks{{Subnets: types.OpenAPIEdgeGatewaySubnets{Values: getNsxtEdgeGatewayUplinksType(d)}}}, nil

	// 'subnet_with_total_ip_count' and 'total_allocated_ip_count' are specified
	case d.HasChange("subnet_with_total_ip_count") || d.HasChange("total_allocated_ip_count"):
		_, subnetValues := getNsxtEdgeGatewayUplinksTypeAutoSubnets(d)
		uplinks := []types.EdgeGatewayUplinks{{}}
		uplinks[0].Subnets = types.OpenAPIEdgeGatewaySubnets{Values: subnetValues}
		if desiredtotalIpCount, isSetTotalIpCount := d.GetOk("total_allocated_ip_count"); isSetTotalIpCount {
			// Allocation and deallocation of IPs are distinct operations due to API limitations
			// To decide whether to allocate or deallocate IPs, we need to calculate the balance.
			// Balance is the difference between desired and current total allocated IP count
			// If balance is positive, we need to allocate additional IPs
			// Example: 204 - 200 = 4 (4 IPs to allocate)
			// If balance is negative, we need to deallocate IPs
			// Example: 200 - 204 = -4 (4 IPs to deallocate)
			ipBalance := desiredtotalIpCount.(int) - *currentlyAllocatedIpCount
			util.Logger.Printf("[TRACE] Edge Gateway Ip Balance is '%d' (desired IP count '%d', currently allocated IP count'%d')",
				ipBalance, desiredtotalIpCount.(int), *currentlyAllocatedIpCount)

			// If balance is positive, we can utilize QuickAddAllocatedIPCount field
			if ipBalance > 0 {
				util.Logger.Printf("[TRACE] Edge Gateway Requesting to allocate '%d' IPs", ipBalance)
				uplinks[0].QuickAddAllocatedIPCount = ipBalance
			}

			// If balance is negative, we need to deallocate IPs by modifying the structure
			if ipBalance < 0 {
				util.Logger.Printf("[TRACE] Edge Gateway Modifying structure to deallocate '%d' IPs", -ipBalance)
				uplinks = edgeGateway.EdgeGateway.EdgeGatewayUplinks
				// edgeGatewayStructure := &types.OpenAPIEdgeGateway{
				// 	EdgeGatewayUplinks: uplinks,
				// }

				edgeGatewayStructure := &govcd.NsxtEdgeGateway{
					EdgeGateway: &types.OpenAPIEdgeGateway{
						EdgeGatewayUplinks: uplinks,
					},
				}

				err := edgeGatewayStructure.DeallocateIpCount(-ipBalance)
				if err != nil {
					return nil, fmt.Errorf("error deallocating IPs: %s", err)
				}
			}

		}

		return uplinks, nil

	// 'subnet_with_ip_count' is specified
	case d.HasChange("subnet_with_ip_count"):
		return []types.EdgeGatewayUplinks{{Subnets: types.OpenAPIEdgeGatewaySubnets{Values: getNsxtEdgeGatewayUplinksTypeAutoAllocateSubnets(d)}}}, nil

	// If no changes occur in the 'subnet', 'subnet_with_total_ip_count' or 'subnet_with_ip_count' fields during
	// 'Update', then we pass back original values to the update request
	default:
		return edgeGateway.EdgeGateway.EdgeGatewayUplinks, nil
	}
}

// getCreateOwnerIdWithStartingVdcId defines how `owner_id` is defined for NSX-T Edge Gateway create
// operation.
// Create has a few possible scenarios for getting ownerRef which can be a VDC or a VDC Group.
// NSX-T Edge Gateway cannot be created directly in VDC Group
// * `owner_id` and `starting_vdc_id` are set. `owner_id` is a VDC Group
// * `owner_id` is set and is a VDC Group, but `starting_vdc_id` is not set (create operation must
// lookup any starting VDC in given VDC Group)
// * `vdc` field is set in the resource
// * Neither `vdc`, nor `owner_id` fields are set in the resource. `vdc` is inherited from `provider` section
//
// Note. Only one of `vdc` or `owner_id` (with optional `starting_vdc_id`) can be supplied. This is
// enforced by Terraform schema definition.
func getCreateOwnerIdWithStartingVdcId(d *schema.ResourceData, vcdClient *VCDClient, ownerIdField string, startingVdcId string, vdcField string, inheritedVdcField string) (string, error) {
	var ownerId string

	switch {
	// `owner_id` is specified and is VDC.
	case ownerIdField != "" && govcd.OwnerIsVdc(ownerIdField):
		ownerId = ownerIdField
	// `owner_id` is specified and is VDC Group. `starting_vdc_id` is specified.
	// Initial `owner_id` for create operation should be `starting_vdc_id` which is later going to
	// be moved to a VDC by a separate API call `createdEdgeGateway.MoveToVdcGroup`
	case ownerIdField != "" && govcd.OwnerIsVdcGroup(ownerIdField) && startingVdcId != "":
		log.Printf("[TRACE] NSX-T Edge Gateway create 'owner_id' field is set and is VDC Group. 'starting_vdc_id' is set. Picking 'starting_vdc_id' for create operation")
		ownerId = startingVdcId
	// `owner_id` is specified and is VDC Group. `starting_vdc_id` is not specified.
	// NSX-T Edge Gateway cannot be created in VDC Group therefore we are going to lookup random VDC in specified group
	case ownerIdField != "" && govcd.OwnerIsVdcGroup(ownerIdField) && startingVdcId == "":
		log.Printf("[TRACE] NSX-T Edge Gateway create 'owner_id' field is set and is VDC Group. 'starting_vdc_id' is not set. Choosing random starting VDC")

		// Lookup Org
		org, err := vcdClient.GetOrgFromResource(d)
		if err != nil {
			return "", fmt.Errorf("error retrieving Org: %s", err)
		}

		vdcGroup, err := org.GetVdcGroupById(ownerIdField)
		if err != nil {
			return "", fmt.Errorf("error retrieving VDC Group: %s", err)
		}

		if vdcGroup.VdcGroup != nil && len(vdcGroup.VdcGroup.ParticipatingOrgVdcs) > 0 {
			log.Printf("[TRACE] NSX-T Edge Gateway create 'owner_id' field is set and is VDC Group. 'starting_vdc_id' is not set. Picked starting VDC '%s' (%s)",
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

// getOwnerId defines how `owner_id` is picked from update behaves differently than create as
// `ownerRef` can have a VDC Group set directly. A few scenarios can happen (listed by priority):
// * `owner_id` is set - using it
// * `vdc` field is set in the resource - using it
// * Neither `vdc`, nor `owner_id` fields are set in the resource. `vdc` is inherited from `provider` section
//
// Note. Only one of `vdc` or `owner_id`. This is enforced by Terraform schema definition.
func getOwnerId(d *schema.ResourceData, vcdClient *VCDClient, ownerIdField, vdcField, inheritedVdcField string) (string, error) {
	switch {
	case ownerIdField != "":
		log.Printf("[TRACE] 'owner_id' is set. Using it.")
		return ownerIdField, nil
	case vdcField != "":
		log.Printf("[TRACE] 'vdc' field is set in resource")

		org, err := vcdClient.GetOrgFromResource(d)
		if err != nil {
			return "", fmt.Errorf("error retrieving Org: %s", err)
		}

		vdc, err := org.GetVDCByName(vdcField, false)
		if err != nil {
			return "", fmt.Errorf("error finding VDC '%s': %s", vdcField, err)
		}
		return vdc.Vdc.ID, nil
	case inheritedVdcField != "" && vdcField == "" && ownerIdField == "":
		log.Printf("[TRACE] 'vdc' field is inherited from provider. `vdc` and `owner_id` are not set")

		org, err := vcdClient.GetOrgFromResource(d)
		if err != nil {
			return "", fmt.Errorf("error retrieving Org: %s", err)
		}

		vdc, err := org.GetVDCByName(inheritedVdcField, false)
		if err != nil {
			return "", fmt.Errorf("error finding VDC '%s': %s", inheritedVdcField, err)
		}

		return vdc.Vdc.ID, nil
	}

	return "", fmt.Errorf("error looking up ownerId field owner_id='%s', vdc='%s', inherited vdc='%s'",
		ownerIdField, vdcField, inheritedVdcField)
}

// getNsxtEdgeGatewayUplinksType is used to convert the uplink slice from the schema to the type
// based on 'subnet' field
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

// getNsxtEdgeGatewayUplinksTypeAutoSubnets is used to convert the uplink slice from the schema to
// the type based on 'subnet_with_total_ip_count' field
func getNsxtEdgeGatewayUplinksTypeAutoSubnets(d *schema.ResourceData) (bool, []types.OpenAPIEdgeGatewaySubnetValue) {
	var isPrimaryIpSet bool
	var primaryIpIndex int

	extNetworks := d.Get("subnet_with_total_ip_count").(*schema.Set).List()
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

	return isPrimaryIpSet, subnetSlice
}

// getNsxtEdgeGatewayUplinksTypeAutoAllocateSubnets is used to convert the uplink slice from the
// schema to the type based on 'subnet_with_ip_count' field
func getNsxtEdgeGatewayUplinksTypeAutoAllocateSubnets(d *schema.ResourceData) []types.OpenAPIEdgeGatewaySubnetValue {
	var isPrimaryIpSet bool
	var primaryIpIndex int

	schemaSubnets := d.Get("subnet_with_ip_count").(*schema.Set).List()
	subnetSlice := make([]types.OpenAPIEdgeGatewaySubnetValue, len(schemaSubnets))

	for index, singleSubnet := range schemaSubnets {
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

		singleSubnet.AutoAllocateIPRanges = true // required for allocated_ip_count
		singleSubnet.TotalIPCount = addrOf(subnetMap["allocated_ip_count"].(int))

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
func setNsxtEdgeGatewayData(edgeGateway *govcd.NsxtEdgeGateway, d *schema.ResourceData) error {
	edgeGw := edgeGateway.EdgeGateway
	dSet(d, "name", edgeGw.Name)
	dSet(d, "description", edgeGw.Description)
	dSet(d, "edge_cluster_id", edgeGw.EdgeClusterConfig.PrimaryEdgeCluster.BackingID)
	if len(edgeGw.EdgeGatewayUplinks) < 1 {
		return fmt.Errorf("no edge gateway uplinks detected during read")
	}

	dSet(d, "owner_id", edgeGw.OwnerRef.ID)
	dSet(d, "vdc", edgeGw.OwnerRef.Name)

	// NSX-T Edge Gateways support only 1 uplink. Edge gateway can only be connected to one external network (in NSX-T terms
	// Tier 1 gateway can only be connected to single Tier 0 gateway)
	edgeUplink := edgeGw.EdgeGatewayUplinks[0]

	dSet(d, "dedicate_external_network", edgeUplink.Dedicated)
	dSet(d, "external_network_id", edgeUplink.UplinkID)

	// 'subnet' field
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
	// End of 'subnet' field

	// 'subnet_with_total_ip_count' and 'total_allocated_ip_count' fields
	totalAllocatedIpCount, err := edgeGateway.GetAllocatedIpCount(false)
	if err != nil {
		return fmt.Errorf("error getting NSX-T Edge Gateway total allocated IP count: %s", err)
	}
	err = d.Set("total_allocated_ip_count", totalAllocatedIpCount)
	if err != nil {
		return fmt.Errorf("error setting NSX-T Edge Gateway total allocated IP count: %s", err)
	}

	autoSubnets := make([]interface{}, 1)
	for _, subnetValue := range edgeUplink.Subnets.Values {
		oneSubnet := make(map[string]interface{})
		oneSubnet["gateway"] = subnetValue.Gateway
		oneSubnet["prefix_length"] = subnetValue.PrefixLength
		oneSubnet["primary_ip"] = subnetValue.PrimaryIP

		autoSubnets = append(autoSubnets, oneSubnet)
	}

	autoSubnetSet := schema.NewSet(schema.HashResource(nsxtEdgeAutoSubnetAndTotal), autoSubnets)
	err = d.Set("subnet_with_total_ip_count", autoSubnetSet)
	if err != nil {
		return fmt.Errorf("error setting NSX-T Edge Gateway automatic subnets after read: %s", err)
	}
	// End of 'subnet_with_total_ip_count' and 'total_allocated_ip_count' fields

	// 'subnet_with_ip_count' field
	autoAllocatedSubnets := make([]interface{}, 1)
	for _, subnetValue := range edgeUplink.Subnets.Values {
		oneSubnet := make(map[string]interface{})
		oneSubnet["gateway"] = subnetValue.Gateway
		oneSubnet["prefix_length"] = subnetValue.PrefixLength
		oneSubnet["primary_ip"] = subnetValue.PrimaryIP
		oneSubnet["allocated_ip_count"] = *subnetValue.TotalIPCount

		autoAllocatedSubnets = append(autoAllocatedSubnets, oneSubnet)
	}

	autoAllocatedSubnetSet := schema.NewSet(schema.HashResource(nsxtEdgeAutoAllocatedSubnet), autoAllocatedSubnets)
	err = d.Set("subnet_with_ip_count", autoAllocatedSubnetSet)
	if err != nil {
		return fmt.Errorf("error setting NSX-T Edge Gateway auto allocated subnets after read: %s", err)
	}
	// End of 'subnet_with_ip_count' field

	unusedIps, err := edgeGateway.GetAllUnusedExternalIPAddresses(false)
	if err != nil {
		return fmt.Errorf("error getting NSX-T Edge Gateway unused IPs after read: %s", err)
	}

	usedIps, err := edgeGateway.GetUsedIpAddressSlice(false)
	if err != nil {
		return fmt.Errorf("error getting NSX-T Edge Gateway used IPs after read: %s", err)
	}

	dSet(d, "used_ip_count", len(usedIps))
	dSet(d, "unused_ip_count", len(unusedIps))

	return nil
}

// vdcOrVdcGroupVerifier is an interface to access IsNsxt() and GetNsxtEdgeGatewayByName() on VDC or
// VDC Group method `IsNsxt` (used in isBackedByNsxt and resourceVcdNsxtEdgeGatewayImport)
type vdcOrVdcGroupVerifier interface {
	IsNsxt() bool
	GetNsxtEdgeGatewayByName(name string) (*govcd.NsxtEdgeGateway, error)
}
