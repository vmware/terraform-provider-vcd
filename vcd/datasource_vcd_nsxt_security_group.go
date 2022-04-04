package vcd

import (
	"context"
	"github.com/vmware/go-vcloud-director/v2/govcd"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func datasourceVcdNsxtSecurityGroup() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdSecurityGroupRead,

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
				Deprecated:  "Deprecated in favor of `edge_gateway_id`. Security Group will inherit VDC from parent Edge Gateway.",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Security Group name",
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
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Security Group description",
			},
			"member_org_network_ids": {
				Computed:    true,
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

func datasourceVcdSecurityGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	org, err := vcdClient.GetOrgFromResource(d)
	if err != nil {
		return diag.Errorf("[nsxt security group read] error retrieving Org: %s", err)
	}

	securityGroupName := d.Get("name").(string)
	edgeGatewayId := d.Get("edge_gateway_id").(string)

	var securityGroup *govcd.NsxtFirewallGroup
	var parentVdcOrVdcGroupId string

	if securityGroupName == "" || edgeGatewayId == "" {
		return diag.Errorf("error - not all parameters specified for NSX-T Security Group lookup")
	}
	// Lookup Edge Gateway to know parent VDC or VDC Group
	anyEdgeGateway, err := org.GetAnyTypeEdgeGatewayById(edgeGatewayId)
	if err != nil {
		return diag.Errorf("[nsxt security group read] error retrieving Edge Gateway structure: %s", err)
	}
	if anyEdgeGateway.IsNsxv() {
		return diag.Errorf("[nsxt security group read] NSX-V Edge Gateway not supported")
	}

	parentVdcOrVdcGroupId = anyEdgeGateway.EdgeGateway.OwnerRef.ID

	if govcd.OwnerIsVdcGroup(parentVdcOrVdcGroupId) {
		vdcGroup, err := org.GetVdcGroupById(parentVdcOrVdcGroupId)
		if err != nil {
			return diag.Errorf("could not retrieve VDC Group with ID '%s': %s", d.Id(), err)
		}

		// Name uniqueness is enforced by VCD for types.FirewallGroupTypeSecurityGroup
		securityGroup, err = vdcGroup.GetNsxtFirewallGroupByName(d.Get("name").(string), types.FirewallGroupTypeSecurityGroup)
		if err != nil {
			return diag.Errorf("[nsxt security group read] error getting NSX-T Security Group with Name '%s': %s", d.Get("name").(string), err)
		}
	} else {
		nsxtEdgeGateway, err := anyEdgeGateway.GetNsxtEdgeGateway()
		if err != nil {
			return diag.Errorf("could not retrieve NSX-T Edge Gateway with ID '%s': %s", d.Id(), err)
		}

		// Name uniqueness is enforced by VCD for types.FirewallGroupTypeSecurityGroup
		securityGroup, err = nsxtEdgeGateway.GetNsxtFirewallGroupByName(d.Get("name").(string), types.FirewallGroupTypeSecurityGroup)
		if err != nil {
			return diag.Errorf("[nsxt security group read] error getting NSX-T Security Group with Name '%s': %s", d.Get("name").(string), err)
		}
	}

	err = setNsxtSecurityGroupData(d, securityGroup.NsxtFirewallGroup, parentVdcOrVdcGroupId)
	if err != nil {
		return diag.Errorf("[nsxt security group read] error setting NSX-T Security Group: %s", err)
	}

	// A separate GET call is required to get all associated VMs
	associatedVms, err := securityGroup.GetAssociatedVms()
	if err != nil {
		return diag.Errorf("[nsxt security group read] error getting associated VMs for Security Group '%s': %s", securityGroup.NsxtFirewallGroup.Name, err)
	}

	err = setNsxtSecurityGroupAssociatedVmsData(d, associatedVms)
	if err != nil {
		return diag.Errorf("[nsxt security group read] error getting associated VMs for Security Group '%s': %s", securityGroup.NsxtFirewallGroup.Name, err)
	}

	d.SetId(securityGroup.NsxtFirewallGroup.ID)

	return nil
}
