package vcd

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceVcdNsxtDistributedFirewall() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdNsxtDistributedFirewallCreate,
		ReadContext:   resourceVcdNsxtDistributedFirewallRead,
		UpdateContext: resourceVcdNsxtDistributedFirewallUpdate,
		DeleteContext: resourceVcdNsxtDistributedFirewallDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdNsxtDistributedFirewallImport,
		},

		Schema: map[string]*schema.Schema{
			"org": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Description: "The name of organization to use, optional if defined at provider " +
					"level. Useful when connected as sysadmin working across different organizations",
			},
			"vdc_group_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name of VDC to use, optional if defined at provider level",
			},
			"edge_gateway": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Edge gateway name in which NAT Rule is located",
			},
			"network_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Org or external network name",
			},
		},
	}
}

func resourceVcdNsxtDistributedFirewallCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	vcdClient.lockParentEdgeGtw(d)
	defer vcdClient.unLockParentEdgeGtw(d)

	// orgName := d.Get("org").(string)
	// vdcName := d.Get("vdc").(string)
	// edgeGatewayId := d.Get("edge_gateway_id").(string)

	// nsxtEdge, err := vcdClient.GetNsxtEdgeGatewayById(orgName, vdcName, edgeGatewayId)
	// if err != nil {
	// 	return diag.Errorf("error retrieving Edge Gateway: %s", err)
	// }

	// firewallRulesType := getNsxtFirewallTypes(d)
	// firewallContainer := &types.NsxtFirewallRuleContainer{
	// 	UserDefinedRules: firewallRulesType,
	// }

	// _, err = nsxtEdge.UpdateNsxtFirewall(firewallContainer)
	// if err != nil {
	// 	return diag.Errorf("error creating NSX-T Firewall Rules: %s", err)
	// }

	// // ID is stored as Edge Gateway ID - because this is a "container" for all firewall rules at once and each child
	// // TypeSet element will have a computed ID field for each rule
	// d.SetId(edgeGatewayId)

	return resourceVcdNsxtDistributedFirewallRead(ctx, d, meta)
}

func resourceVcdNsxtDistributedFirewallUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceVcdNsxtDistributedFirewallRead(ctx, d, meta)
}

func resourceVcdNsxtDistributedFirewallRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

func resourceVcdNsxtDistributedFirewallDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

func resourceVcdNsxtDistributedFirewallImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	return []*schema.ResourceData{d}, nil
}
