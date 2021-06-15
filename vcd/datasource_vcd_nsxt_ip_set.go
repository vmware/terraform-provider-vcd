package vcd

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
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
			},
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "IP set name",
			},
			"edge_gateway_id": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Edge Gateway ID in which IP Set is located",
			},
			"description": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "IP set description",
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

	nsxtEdgeGateway, err := vcdClient.GetNsxtEdgeGatewayFromResourceById(d, "edge_gateway_id")
	if err != nil {
		return diag.Errorf(errorUnableToFindEdgeGateway, err)
	}

	// Name uniqueness is enforced by VCD for types.FirewallGroupTypeIpSet
	ipSet, err := nsxtEdgeGateway.GetNsxtFirewallGroupByName(d.Get("name").(string), types.FirewallGroupTypeIpSet)
	if err != nil {
		return diag.Errorf("error getting NSX-T IP Set with Name '%s': %s", d.Get("name").(string), err)
	}
	err = setNsxtIpSetData(d, ipSet.NsxtFirewallGroup)
	if err != nil {
		return diag.Errorf("error setting NSX-T IP Set: %s", err)
	}

	d.SetId(ipSet.NsxtFirewallGroup.ID)

	return nil
}
