package vcd

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func datasourceVcdDynamicSecurityGroup() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdDynamicSecurityGroupRead,

		Schema: map[string]*schema.Schema{
			"org": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Description: "The name of organization to use, optional if defined at provider " +
					"level. Useful when connected as sysadmin working across different organizations",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Dynamic Security Group name",
			},
			"vdc_group_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "VDC Group ID in which Dynamic Security Group is located",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Dynamic Security Group description",
			},
			"criteria": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "Criteria to be used to define the Dynamic Security Group",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"rule": {
							Type:        schema.TypeSet,
							Description: "Up to 4 rules can be used to define single criteria",
							Computed:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"type": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "Type of object matching 'VM_TAG' or 'VM_NAME'",
									},
									"operator": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "Operator can be one of 'EQUALS', 'CONTAINS', 'STARTS_WITH', 'ENDS_WITH'",
									},
									"value": {
										Type:        schema.TypeString,
										Required:    true,
										Description: "Filter value",
									},
								},
							},
						},
					},
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

func datasourceVcdDynamicSecurityGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	vdcGroupId := d.Get("vdc_group_id").(string)

	org, err := vcdClient.GetOrgFromResource(d)
	if err != nil {
		return diag.Errorf("[nsxt dynamic security group data source read] error retrieving Org: %s", err)
	}

	vdcGroup, err := org.GetVdcGroupById(vdcGroupId)
	if err != nil {
		return diag.Errorf("error retrieving VDC Group: %s", err)
	}

	securityGroup, err := vdcGroup.GetNsxtFirewallGroupByName(d.Get("name").(string), types.FirewallGroupTypeVmCriteria)
	if err != nil {
		return diag.Errorf("[nsxt dynamic security group data source read] error getting NSX-T dynamic security group: %s", err)
	}

	err = setNsxtDynamicSecurityGroupData(d, securityGroup.NsxtFirewallGroup)
	if err != nil {
		return diag.Errorf("[nsxt security group data source read] error setting NSX-T Security Group: %s", err)
	}

	// A separate GET call is required to get all associated VMs
	associatedVms, err := securityGroup.GetAssociatedVms()
	if err != nil {
		return diag.Errorf("[nsxt dynamic security group data source read] error getting associated VMs for Security Group '%s': %s", securityGroup.NsxtFirewallGroup.Name, err)
	}

	err = setNsxtSecurityGroupAssociatedVmsData(d, associatedVms)
	if err != nil {
		return diag.Errorf("[nsxt dynamic security group data source read] error getting associated VMs for Security Group '%s': %s", securityGroup.NsxtFirewallGroup.Name, err)
	}

	d.SetId(securityGroup.NsxtFirewallGroup.ID)

	return nil
}
