package vcd

import (
	"context"

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
			},
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Security Group name",
			},
			"edge_gateway_id": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Edge Gateway ID in which security group is located",
			},
			"description": &schema.Schema{
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

	nsxtEdgeGateway, err := vcdClient.GetNsxtEdgeGatewayFromResourceById(d, "edge_gateway_id")
	if err != nil {
		return diag.Errorf(errorUnableToFindEdgeGateway, err)
	}

	// Name uniqueness is enforce by VCD for types.FirewallGroupTypeSecurityGroup
	secGroup, err := nsxtEdgeGateway.GetNsxtFirewallGroupByName(d.Get("name").(string), types.FirewallGroupTypeSecurityGroup)
	if err != nil {
		return diag.Errorf("error getting NSX-T Security Group with ID '%s': %s", d.Id(), err)
	}

	err = setNsxtSecurityGroupData(d, secGroup.NsxtFirewallGroup)
	if err != nil {
		return diag.Errorf("error reading NSX-T Security Group: %s", err)
	}

	// A separate GET call is required to get all associated VMs
	associatedVms, err := secGroup.GetAssociatedVms()
	if err != nil {
		return diag.Errorf("error getting associated VMs for Security Group '%s': %s", secGroup.NsxtFirewallGroup.Name, err)
	}

	err = setNsxtSecurityGroupAssociatedVmsData(d, associatedVms)
	if err != nil {
		return diag.Errorf("error getting associated VMs for Security Group '%s': %s", secGroup.NsxtFirewallGroup.Name, err)
	}

	d.SetId(secGroup.NsxtFirewallGroup.ID)

	return nil
}
