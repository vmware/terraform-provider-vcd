package vcd

import (
	"context"
	"errors"
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
				ForceNew:    true,
				Description: "The name of VDC to use, optional if defined at provider level",
			},
			"edge_gateway_id": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Edge Gateway ID in which security group is located",
			},
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Security Group name",
			},
			"description": &schema.Schema{
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
	vcdClient.lockParentEdgeGtw(d)
	defer vcdClient.unLockParentEdgeGtw(d)

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return diag.Errorf(errorRetrievingOrgAndVdc, err)
	}

	fwGroup := getNsxtSecurityGroupType(d)

	createdFwGroup, err := vdc.CreateNsxtFirewallGroup(fwGroup)
	err = improveErrorMessageOnIncorrectMembership(err, fwGroup, vdc)
	if err != nil {
		return diag.Errorf("error creating NSX-T Security Group '%s': %s", fwGroup.Name, err)
	}

	d.SetId(createdFwGroup.NsxtFirewallGroup.ID)

	return resourceVcdSecurityGroupRead(ctx, d, meta)
}

func resourceVcdSecurityGroupUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	vcdClient.lockParentEdgeGtw(d)
	defer vcdClient.unLockParentEdgeGtw(d)

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return diag.Errorf(errorRetrievingOrgAndVdc, err)
	}

	fwGroup, err := vdc.GetNsxtFirewallGroupById(d.Id())
	if err != nil {
		return diag.Errorf("error getting NSX-T Security Group: %s", err)
	}

	updateFwGroup := getNsxtSecurityGroupType(d)
	// Inject existing ID for update
	updateFwGroup.ID = d.Id()

	_, err = fwGroup.Update(updateFwGroup)
	err = improveErrorMessageOnIncorrectMembership(err, updateFwGroup, vdc)
	if err != nil {
		return diag.Errorf("error updating NSX-T Security Group '%s': %s", fwGroup.NsxtFirewallGroup.Name, err)
	}

	return resourceVcdSecurityGroupRead(ctx, d, meta)
}

func resourceVcdSecurityGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return diag.Errorf(errorRetrievingOrgAndVdc, err)
	}

	fwGroup, err := vdc.GetNsxtFirewallGroupById(d.Id())
	if err != nil {
		if govcd.ContainsNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.Errorf("error getting NSX-T Security Group with ID '%s': %s", d.Id(), err)
	}

	err = setNsxtSecurityGroupData(d, fwGroup.NsxtFirewallGroup)
	if err != nil {
		return diag.Errorf("error reading NSX-T Security Group: %s", err)
	}

	// A separate GET call is required to get all associated VMs
	associatedVms, err := fwGroup.GetAssociatedVms()
	if err != nil {
		return diag.Errorf("error getting associated VMs for Security Group '%s': %s", fwGroup.NsxtFirewallGroup.Name, err)
	}

	err = setNsxtSecurityGroupAssociatedVmsData(d, associatedVms)
	if err != nil {
		return diag.Errorf("error getting associated VMs for Security Group '%s': %s", fwGroup.NsxtFirewallGroup.Name, err)
	}

	return nil
}

func resourceVcdSecurityGroupDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	vcdClient.lockParentEdgeGtw(d)
	defer vcdClient.unLockParentEdgeGtw(d)

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return diag.Errorf(errorRetrievingOrgAndVdc, err)
	}

	fwGroup, err := vdc.GetNsxtFirewallGroupById(d.Id())
	if err != nil {
		return diag.Errorf("error getting NSX-T Security Group: %s", err)
	}

	err = fwGroup.Delete()
	if err != nil {
		return diag.Errorf("error deleting NSX-T Security Group: %s", err)
	}

	d.SetId("")

	return nil
}

func resourceVcdSecurityGroupImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	resourceURI := strings.Split(d.Id(), ImportSeparator)
	if len(resourceURI) != 4 {
		return nil, fmt.Errorf("resource name must be specified as org-name.vdc-name.edge_gateway_name.security_group_name")
	}
	orgName, vdcName, edgeGatewayName, securityGroupName := resourceURI[0], resourceURI[1], resourceURI[2], resourceURI[3]

	vcdClient := meta.(*VCDClient)
	org, err := vcdClient.GetAdminOrg(orgName)
	if err != nil {
		return nil, fmt.Errorf("unable to find Org %s: %s", orgName, err)
	}
	vdc, err := org.GetVDCByName(vdcName, false)
	if err != nil {
		return nil, fmt.Errorf("unable to find VDC %s: %s", vdcName, err)
	}

	if !vdc.IsNsxt() {
		return nil, errors.New("security groups are only supported by NSX-T VDCs")
	}

	edgeGateway, err := vdc.GetNsxtEdgeGatewayByName(edgeGatewayName)
	if err != nil {
		return nil, fmt.Errorf("unable to find Edge Gateway '%s': %s", edgeGatewayName, err)
	}

	securityGroup, err := edgeGateway.GetNsxtFirewallGroupByName(securityGroupName, types.FirewallGroupTypeSecurityGroup)
	if err != nil {
		return nil, fmt.Errorf("unable to find Security Group '%s': %s", securityGroupName, err)
	}

	if !securityGroup.IsSecurityGroup() {
		return nil, fmt.Errorf("firewall group '%s' is not a Security Group, but '%s'",
			securityGroup.NsxtFirewallGroup.Name, securityGroup.NsxtFirewallGroup.Type)
	}

	_ = d.Set("org", orgName)
	_ = d.Set("vdc", vdcName)
	_ = d.Set("edge_gateway_id", edgeGateway.EdgeGateway.ID)
	d.SetId(securityGroup.NsxtFirewallGroup.ID)

	return []*schema.ResourceData{d}, nil
}

func setNsxtSecurityGroupData(d *schema.ResourceData, fw *types.NsxtFirewallGroup) error {
	_ = d.Set("name", fw.Name)
	_ = d.Set("description", fw.Description)

	netIds := make([]string, len(fw.Members))
	for i := range fw.Members {
		netIds[i] = fw.Members[i].ID
	}

	// Convert `member_org_network_ids` to set
	memberNetIds := convertStringsToInterfaceSlice(netIds)
	memberNetSet := schema.NewSet(schema.HashSchema(&schema.Schema{Type: schema.TypeString}), memberNetIds)

	err := d.Set("member_org_network_ids", memberNetSet)
	if err != nil {
		return fmt.Errorf("error setting 'member_org_network_ids': %s", err)
	}

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

func getNsxtSecurityGroupType(d *schema.ResourceData) *types.NsxtFirewallGroup {
	fwGroup := &types.NsxtFirewallGroup{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		EdgeGatewayRef: &types.OpenApiReference{
			ID: d.Get("edge_gateway_id").(string),
		},
		Type: types.FirewallGroupTypeSecurityGroup,
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
func improveErrorMessageOnIncorrectMembership(err error, fwGroup *types.NsxtFirewallGroup, vdc *govcd.Vdc) error {
	// See if error message contains an indication to non Routed networks
	if err != nil && strings.Contains(err.Error(), `No access to entity "com.vmware.vcloud.entity.gateway`) {
		hasAtLeastOneNonRoutedNetwork := false
		for i := range fwGroup.Members {
			orgNet, err2 := vdc.GetOpenApiOrgVdcNetworkById(fwGroup.Members[i].ID)

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
