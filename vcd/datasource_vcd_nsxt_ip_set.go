package vcd

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func datasourceVcdNsxtIpSet() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdNsxtIpSetRead,

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
				Deprecated:  "Deprecated in favor of `edge_gateway_id`. IP Set will inherit VDC from parent Edge Gateway.",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "IP Set name",
			},
			"edge_gateway_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Edge Gateway ID in which IP Set is located",
			},
			"owner_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ID of VDC or VDC Group",
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "IP Set description",
			},
			"ip_addresses": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "A set of IP address, CIDR, IP range objects",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func datasourceVcdNsxtIpSetRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	org, err := vcdClient.GetOrgFromResource(d)
	if err != nil {
		return diag.Errorf("[nsxt ip set read] error retrieving Org: %s", err)
	}

	ipSetName := d.Get("name").(string)
	edgeGatewayId := d.Get("edge_gateway_id").(string)

	var ipSet *govcd.NsxtFirewallGroup
	var parentVdcOrVdcGroupId string

	if ipSetName == "" || edgeGatewayId == "" {
		return diag.Errorf("error - not all parameters specified for NSX-T IP Set lookup")
	}
	// Lookup Edge Gateway to know parent VDC or VDC Group
	anyEdgeGateway, err := org.GetAnyTypeEdgeGatewayById(edgeGatewayId)
	if err != nil {
		return diag.Errorf("[nsxt ip set read] error retrieving Edge Gateway structure: %s", err)
	}
	if anyEdgeGateway.IsNsxv() {
		return diag.Errorf("[nsxt ip set read] NSX-V edge gateway not supported")
	}

	parentVdcOrVdcGroupId = anyEdgeGateway.EdgeGateway.OwnerRef.ID

	if govcd.OwnerIsVdcGroup(parentVdcOrVdcGroupId) {
		vdcGroup, err := org.GetVdcGroupById(parentVdcOrVdcGroupId)
		if err != nil {
			return diag.Errorf("could not retrieve VDC Group with ID '%s': %s", d.Id(), err)
		}

		// Name uniqueness is enforced by VCD for types.FirewallGroupTypeIpSet
		ipSet, err = vdcGroup.GetNsxtFirewallGroupByName(d.Get("name").(string), types.FirewallGroupTypeIpSet)
		if err != nil {
			return diag.Errorf("[nsxt ip set read] error getting NSX-T IP Set with Name '%s': %s", d.Get("name").(string), err)
		}
	} else {
		nsxtEdgeGateway, err := anyEdgeGateway.GetNsxtEdgeGateway()
		if err != nil {
			return diag.Errorf("could not retrieve NSX-T Edge Gateway with ID '%s': %s", d.Id(), err)
		}

		// Name uniqueness is enforced by VCD for types.FirewallGroupTypeIpSet
		ipSet, err = nsxtEdgeGateway.GetNsxtFirewallGroupByName(d.Get("name").(string), types.FirewallGroupTypeIpSet)
		if err != nil {
			return diag.Errorf("[nsxt ip set read] error getting NSX-T IP Set with Name '%s': %s", d.Get("name").(string), err)
		}
	}

	err = setNsxtIpSetData(d, ipSet.NsxtFirewallGroup, parentVdcOrVdcGroupId)
	if err != nil {
		return diag.Errorf("error setting NSX-T IP Set: %s", err)
	}

	d.SetId(ipSet.NsxtFirewallGroup.ID)

	return nil
}
