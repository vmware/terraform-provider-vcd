package vcd

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdNsxtFirewall() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceNsxtFirewallRead,
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
				Description: "Edge Gateway ID in which Firewall rules are located",
			},
			"rule": &schema.Schema{
				Type:        schema.TypeList, // Firewall rule order matters
				Computed:    true,
				Description: "List of firewall rules",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": &schema.Schema{
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Firewall Rule ID",
						},
						"name": &schema.Schema{
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Firewall Rule name",
						},
						"direction": &schema.Schema{
							Type:        schema.TypeString,
							Computed:    true,
							Description: "IN OUT IN_OUT",
						},
						"ip_protocol": &schema.Schema{
							Type:        schema.TypeString,
							Computed:    true,
							Description: "IPV4,  IPV6, IPV4_IPV6",
						},
						"enabled": &schema.Schema{
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Firewall Rule name",
						},
						"logging": &schema.Schema{
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Firewall Rule name",
						},
						"action": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"sources": {
							Type:        schema.TypeSet,
							Computed:    true,
							Description: "A set of Source Firewall Group IDs (IP Sets or Security Groups). Leaving it empty means 'Any'",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"destinations": {
							Type:        schema.TypeSet,
							Computed:    true,
							Description: "A set of Destination Firewall Group IDs (IP Sets or Security Groups). Leaving it empty means 'Any'",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"applications": {
							Type:        schema.TypeSet,
							Computed:    true,
							Description: "A set of Application Port Profile IDs. Leaving it empty means 'Any'",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
		},
	}
}

func datasourceNsxtFirewallRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	orgName := d.Get("org").(string)
	vdcName := d.Get("vdc").(string)
	edgeGatewayId := d.Get("edge_gateway_id").(string)

	nsxtEdge, err := vcdClient.GetNsxtEdgeGatewayById(orgName, vdcName, edgeGatewayId)
	if err != nil {
		return diag.Errorf("error retrieving NSX-T Edge Gateway: %s", err)
	}

	fwRules, err := nsxtEdge.GetNsxtFirewall()
	if err != nil {
		return diag.Errorf("error retrieving NSX-T Firewall Rules: %s", err)
	}

	err = setNsxtFirewallData(fwRules.NsxtFirewallRuleContainer.UserDefinedRules, d, edgeGatewayId)
	if err != nil {
		return diag.Errorf("error storing NSX-T Firewall data to schema: %s", err)
	}

	d.SetId(edgeGatewayId)

	return nil
}
