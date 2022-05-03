package vcd

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func resourceVcdSecurityGroup() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdSecurityGroupCreate,
		ReadContext:   resourceVcdSecurityGroupRead,
		UpdateContext: resourceVcdSecurityGroupUpdate,
		DeleteContext: resourceVcdSecurityGroupDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdSecurityGroupImport,
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
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The name of VDC to use, optional if defined at provider level",
				Deprecated:  "Deprecated in favor of `edge_gateway_id`. Security Group will inherit VDC from parent Edge Gateway.",
			},
			"edge_gateway_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Edge Gateway ID in which security group is located",
			},
			"owner_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ID of VDC or VDC Group",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Security Group name",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Security Group description",
			},
			"member_org_network_ids": {
				Optional:    true,
				Type:        schema.TypeSet,
				Description: "Set of Org VDC network IDs attached to this security group",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"member_vms": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "Set of VM IDs",
				Elem:        nsxtFirewallGroupMemberVms,
			},
		},
	}
}

var nsxtFirewallGroupMemberVms = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"vm_id": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Member VM ID",
		},
		"vm_name": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Member VM Name",
		},
		"vapp_id": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Parent vApp name (if exists) for member VM",
		},
		"vapp_name": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Parent vApp ID (if exists) for member VM",
		},
	},
}

func resourceVcdSecurityGroupCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	parentEdgeGatewayOwnerId, nsxtEdgeGateway, err := getParentEdgeGatewayOwnerIdAndNsxtEdgeGateway(vcdClient, d, "security group create")
	if err != nil {
		return diag.FromErr(err)
	}

	var securityGroup *types.NsxtFirewallGroup
	var vdcOrVdcGroup vdcOrVdcGroupHandler

	org, err := vcdClient.GetOrgFromResource(d)
	if err != nil {
		return diag.Errorf("[nsxt security group create] error retrieving Org: %s", err)
	}

	if govcd.OwnerIsVdcGroup(parentEdgeGatewayOwnerId) {
		vcdClient.lockById(parentEdgeGatewayOwnerId)
		defer vcdClient.unlockById(parentEdgeGatewayOwnerId)
		securityGroup = getNsxtSecurityGroupType(d, parentEdgeGatewayOwnerId)
		vdcOrVdcGroup, err = org.GetVdcGroupById(parentEdgeGatewayOwnerId)
		diag.Errorf("[nsxt security group create] error retrieving VDC Group with ID %s: %s", parentEdgeGatewayOwnerId, err)
	} else {
		vcdClient.lockParentEdgeGtw(d)
		defer vcdClient.unLockParentEdgeGtw(d)
		securityGroup = getNsxtSecurityGroupType(d, d.Get("edge_gateway_id").(string))
		vdcOrVdcGroup, err = org.GetVDCById(parentEdgeGatewayOwnerId, false)
		diag.Errorf("[nsxt security group create] error retrieving VDC with ID %s: %s", parentEdgeGatewayOwnerId, err)
	}

	createdFwGroup, err := nsxtEdgeGateway.CreateNsxtFirewallGroup(securityGroup)
	err = improveErrorMessageOnIncorrectMembership(err, securityGroup, vdcOrVdcGroup)
	if err != nil {
		return diag.Errorf("[nsxt security group create] error creating NSX-T Security Group '%s': %s", securityGroup.Name, err)
	}

	dSet(d, "edge_gateway_id", nsxtEdgeGateway.EdgeGateway.ID)
	d.SetId(createdFwGroup.NsxtFirewallGroup.ID)

	return resourceVcdSecurityGroupRead(ctx, d, meta)
}

func resourceVcdSecurityGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	parentEdgeGatewayOwnerId, nsxtEdgeGateway, err := getParentEdgeGatewayOwnerIdAndNsxtEdgeGateway(vcdClient, d, "security group create")
	if err != nil {
		return diag.FromErr(err)
	}

	var updateSecurityGroup *types.NsxtFirewallGroup
	var vdcOrVdcGroup vdcOrVdcGroupHandler

	org, err := vcdClient.GetOrgFromResource(d)
	if err != nil {
		return diag.Errorf("[nsxt security group update] error retrieving Org: %s", err)
	}

	if govcd.OwnerIsVdcGroup(parentEdgeGatewayOwnerId) {
		vcdClient.lockById(parentEdgeGatewayOwnerId)
		defer vcdClient.unlockById(parentEdgeGatewayOwnerId)
		updateSecurityGroup = getNsxtSecurityGroupType(d, parentEdgeGatewayOwnerId)
		vdcOrVdcGroup, err = org.GetVdcGroupById(parentEdgeGatewayOwnerId)
		diag.Errorf("[nsxt security group update] error retrieving VDC Group with ID %s: %s", parentEdgeGatewayOwnerId, err)
	} else {
		vcdClient.lockParentEdgeGtw(d)
		defer vcdClient.unLockParentEdgeGtw(d)
		updateSecurityGroup = getNsxtSecurityGroupType(d, d.Get("edge_gateway_id").(string))
		vdcOrVdcGroup, err = org.GetVDCById(parentEdgeGatewayOwnerId, false)
		diag.Errorf("[nsxt security group update] error retrieving VDC with ID %s: %s", parentEdgeGatewayOwnerId, err)
	}

	securityGroup, err := nsxtEdgeGateway.GetNsxtFirewallGroupById(d.Id())
	if err != nil {
		return diag.Errorf("[nsxt security group update] error getting NSX-T Security Group: %s", err)
	}

	// Inject existing ID for update
	updateSecurityGroup.ID = d.Id()

	_, err = securityGroup.Update(updateSecurityGroup)
	err = improveErrorMessageOnIncorrectMembership(err, updateSecurityGroup, vdcOrVdcGroup)
	if err != nil {
		return diag.Errorf("[nsxt security group update] error updating NSX-T Security Group '%s': %s", securityGroup.NsxtFirewallGroup.Name, err)
	}

	return resourceVcdSecurityGroupRead(ctx, d, meta)
}

func resourceVcdSecurityGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	adminOrg, err := vcdClient.GetAdminOrgFromResource(d)
	if err != nil {
		return diag.Errorf("[nsxt security group read] error retrieving Org: %s", err)
	}

	parentEdgeGatewayOwnerId, nsxtEdgeGateway, err := getParentEdgeGatewayOwnerIdAndNsxtEdgeGateway(vcdClient, d, "read")
	if err != nil {
		if govcd.ContainsNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	var securityGroup *govcd.NsxtFirewallGroup
	if govcd.OwnerIsVdcGroup(parentEdgeGatewayOwnerId) {
		vdcGroup, err := adminOrg.GetVdcGroupById(parentEdgeGatewayOwnerId)
		if err != nil {
			return diag.Errorf("[nsxt security group resource read] error finding VDC Group by ID '%s': %s", parentEdgeGatewayOwnerId, err)
		}

		securityGroup, err = vdcGroup.GetNsxtFirewallGroupById(d.Id())
		if err != nil {
			if govcd.ContainsNotFound(err) {
				d.SetId("")
				return nil
			}
			return diag.Errorf("[nsxt security group resource read] error getting NSX-T Security Group with ID '%s': %s", d.Id(), err)
		}
	} else {
		securityGroup, err = nsxtEdgeGateway.GetNsxtFirewallGroupById(d.Id())
		if err != nil {
			if govcd.ContainsNotFound(err) {
				d.SetId("")
				return nil
			}
			return diag.Errorf("[nsxt security group resource read] error getting NSX-T Security Group with ID '%s': %s", d.Id(), err)
		}
	}

	err = setNsxtSecurityGroupData(d, securityGroup.NsxtFirewallGroup, parentEdgeGatewayOwnerId)
	if err != nil {
		return diag.Errorf("[nsxt security group resource read] error reading NSX-T Security Group: %s", err)
	}

	// A separate GET call is required to get all associated VMs
	associatedVms, err := securityGroup.GetAssociatedVms()
	if err != nil {
		return diag.Errorf("[nsxt security group resource read] error getting associated VMs for Security Group '%s': %s", securityGroup.NsxtFirewallGroup.Name, err)
	}

	err = setNsxtSecurityGroupAssociatedVmsData(d, associatedVms)
	if err != nil {
		return diag.Errorf("[nsxt security group resource read] error getting associated VMs for Security Group '%s': %s", securityGroup.NsxtFirewallGroup.Name, err)
	}

	return nil
}

func resourceVcdSecurityGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	parentEdgeGatewayOwnerId, nsxtEdgeGateway, err := getParentEdgeGatewayOwnerIdAndNsxtEdgeGateway(vcdClient, d, "security group create")
	if err != nil {
		return diag.FromErr(err)
	}

	if govcd.OwnerIsVdcGroup(parentEdgeGatewayOwnerId) {
		vcdClient.lockById(parentEdgeGatewayOwnerId)
		defer vcdClient.unlockById(parentEdgeGatewayOwnerId)
	} else {
		vcdClient.lockParentEdgeGtw(d)
		defer vcdClient.unLockParentEdgeGtw(d)
	}

	securityGroup, err := nsxtEdgeGateway.GetNsxtFirewallGroupById(d.Id())
	if err != nil {
		return diag.Errorf("[nsxt security group resource delete] error getting NSX-T Security Group: %s", err)
	}

	err = securityGroup.Delete()
	if err != nil {
		return diag.Errorf("[nsxt security group resource delete] error deleting NSX-T Security Group: %s", err)
	}

	d.SetId("")

	return nil
}

func resourceVcdSecurityGroupImport(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	resourceURI := strings.Split(d.Id(), ImportSeparator)
	if len(resourceURI) != 4 {
		return nil, fmt.Errorf("resource name must be specified as org-name.vdc-name.edge_gateway_name.security_group_name or" +
			"as org-name.vdc-group-name.edge_gateway_name.security_group_name")
	}
	orgName, vdcOrVdcGroupName, edgeGatewayName, securityGroupName := resourceURI[0], resourceURI[1], resourceURI[2], resourceURI[3]

	vcdClient := meta.(*VCDClient)

	// define an interface type to match VDC and VDC Groups
	var vdcOrVdcGroup vdcOrVdcGroupHandler
	var adminOrg *govcd.AdminOrg
	_, vdcOrVdcGroup, err := vcdClient.GetOrgAndVdc(orgName, vdcOrVdcGroupName)
	if govcd.ContainsNotFound(err) {
		adminOrg, err = vcdClient.GetAdminOrg(orgName)
		if err != nil {
			return nil, fmt.Errorf("[nsxt security group resource import] error retrieving Admin Org for '%s': %s", orgName, err)
		}

		vdcOrVdcGroup, err = adminOrg.GetVdcGroupByName(vdcOrVdcGroupName)
		if err != nil {
			return nil, fmt.Errorf("[nsxt security group resource import] error finding VDC or VDC Group by name '%s': %s", vdcOrVdcGroupName, err)
		}
	}

	// Lookup Edge Gateway to know parent VDC or VDC Group
	nsxtEdgeGateway, err := vdcOrVdcGroup.GetNsxtEdgeGatewayByName(edgeGatewayName)
	if err != nil {
		return nil, fmt.Errorf("[nsxt security group import] error retrieving Edge Gateway structure: %s", err)
	}

	var securityGroup *govcd.NsxtFirewallGroup
	parentEdgeGatewayOwnerId := nsxtEdgeGateway.EdgeGateway.OwnerRef.ID
	if govcd.OwnerIsVdcGroup(parentEdgeGatewayOwnerId) {
		vdcGroup, err := adminOrg.GetVdcGroupById(parentEdgeGatewayOwnerId)
		if err != nil {
			return nil, fmt.Errorf("[nsxt security group resource import] error finding VDC Group by ID '%s': %s", parentEdgeGatewayOwnerId, err)
		}

		securityGroup, err = vdcGroup.GetNsxtFirewallGroupByName(securityGroupName, types.FirewallGroupTypeSecurityGroup)
		if err != nil {
			return nil, fmt.Errorf("[nsxt security group resource import] error getting NSX-T Security Group '%s': %s", securityGroupName, err)
		}
	} else {
		securityGroup, err = nsxtEdgeGateway.GetNsxtFirewallGroupByName(securityGroupName, types.FirewallGroupTypeSecurityGroup)
		if err != nil {
			return nil, fmt.Errorf("[nsxt security group resource import] unable to find NSX-T Security Group '%s': %s", securityGroupName, err)
		}
	}

	if !securityGroup.IsSecurityGroup() {
		return nil, fmt.Errorf("firewall group '%s' is not a Security Group, but '%s'",
			securityGroup.NsxtFirewallGroup.Name, securityGroup.NsxtFirewallGroup.Type)
	}
	dSet(d, "org", orgName)
	dSet(d, "edge_gateway_id", nsxtEdgeGateway.EdgeGateway.ID)
	d.SetId(securityGroup.NsxtFirewallGroup.ID)

	return []*schema.ResourceData{d}, nil
}

func setNsxtSecurityGroupData(d *schema.ResourceData, fw *types.NsxtFirewallGroup, parentVdcOrVdcGroupId string) error {
	dSet(d, "name", fw.Name)
	dSet(d, "description", fw.Description)

	netIds := make([]string, len(fw.Members))
	for i := range fw.Members {
		netIds[i] = fw.Members[i].ID
	}

	// Convert `member_org_network_ids` to set
	memberNetSet := convertStringsToTypeSet(netIds)

	err := d.Set("member_org_network_ids", memberNetSet)
	if err != nil {
		return fmt.Errorf("error setting 'member_org_network_ids': %s", err)
	}

	dSet(d, "owner_id", parentVdcOrVdcGroupId)

	return nil
}

func setNsxtSecurityGroupAssociatedVmsData(d *schema.ResourceData, fw []*types.NsxtFirewallGroupMemberVms) error {
	memberVmSlice := make([]interface{}, len(fw))
	for index, vmAssociation := range fw {
		singleVm := make(map[string]interface{})

		if vmAssociation.VmRef != nil {
			singleVm["vm_id"] = vmAssociation.VmRef.ID
			singleVm["vm_name"] = vmAssociation.VmRef.Name
		}

		if vmAssociation.VappRef != nil {
			singleVm["vapp_id"] = vmAssociation.VappRef.ID
			singleVm["vapp_name"] = vmAssociation.VappRef.Name
		}

		memberVmSlice[index] = singleVm
	}
	memberVmSet := schema.NewSet(schema.HashResource(nsxtFirewallGroupMemberVms), memberVmSlice)

	return d.Set("member_vms", memberVmSet)
}

func getNsxtSecurityGroupType(d *schema.ResourceData, ownerId string) *types.NsxtFirewallGroup {
	fwGroup := &types.NsxtFirewallGroup{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		OwnerRef:    &types.OpenApiReference{ID: ownerId},
		Type:        types.FirewallGroupTypeSecurityGroup,
	}

	// Expand member networks
	orgNetworkIds := convertSchemaSetToSliceOfStrings(d.Get("member_org_network_ids").(*schema.Set))
	memberReferences := make([]types.OpenApiReference, len(orgNetworkIds))
	for i := range orgNetworkIds {
		memberReferences[i].ID = orgNetworkIds[i]
	}

	if len(memberReferences) > 0 {
		fwGroup.Members = memberReferences
	}

	return fwGroup
}

// improveErrorMessageOnIncorrectMembership checks if error message is similar to the one that is returned when a
// non Routed Org VDC network ID is passed for Security Group membership and adds additional hint.
func improveErrorMessageOnIncorrectMembership(err error, fwGroup *types.NsxtFirewallGroup, vdcOrVdcGroup vdcOrVdcGroupHandler) error {
	// See if error message contains an indication to non Routed networks
	if err != nil && strings.Contains(err.Error(), `No access to entity "com.vmware.vcloud.entity.gateway`) {
		hasAtLeastOneNonRoutedNetwork := false
		for i := range fwGroup.Members {
			orgNet, err2 := vdcOrVdcGroup.GetOpenApiOrgVdcNetworkById(fwGroup.Members[i].ID)

			// when unable to validate network types - just return the original API error `err`
			if err2 != nil {
				return fmt.Errorf("error creating NSX-T Security Group '%s': %s", fwGroup.Name, err)
			}

			if !orgNet.IsRouted() {
				hasAtLeastOneNonRoutedNetwork = true
			}

		}

		if hasAtLeastOneNonRoutedNetwork {
			err = fmt.Errorf("%s - %s", err.Error(), "Not all member network IDs reference to Routed Org networks.")
		}
	}

	return err
}
