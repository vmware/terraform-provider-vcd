package vcd

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdNsxvDistributedFirewall() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdNsxvDistributedFirewallRead,
		Schema: map[string]*schema.Schema{
			"vdc_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of VDC",
			},
			"enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "When true, it enables the NSX-V distributed firewall",
			},
			"rule": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Ordered list of distributed firewall rules. Will be considered only if `enabled` is true",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Firewall Rule ID",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Firewall Rule name",
						},
						"enabled": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether the rule is enabled",
						},
						"logged": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether the rule traffic is logged",
						},
						"action": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Action of the rule (allow, deny)",
						},
						"direction": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Direction of the rule (in, out, inout)",
						},
						"packet_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Packet type of the rule (any, ipv4, ipv6)",
						},
						"source": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "List of source traffic for this rule. Leaving it empty means 'any'",
							Elem: &schema.Resource{
								Schema: ruleComponent("source", "datasource"),
							},
						},
						"service": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "Service definitions for this rule. Leaving it empty means 'any'",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"protocol": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Protocol of the service (one of TCP, UDP, ICMP) (When not using name/value)",
									},
									"source_port": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Source port for this service. Leaving it empty means 'any' port",
									},
									"destination_port": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Destination port for this service. Leaving it empty means 'any' port",
									},
									"name": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Name of service",
									},
									"value": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Value of the service",
									},
									"type": {
										Type:        schema.TypeString,
										Computed:    true,
										Description: "Type of service",
									},
								},
							},
						},
						"exclude_source": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "If set, reverses the content of the source elements",
						},
						"destination": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "List of destination traffic for this rule. Leaving it empty means 'any'",
							Elem: &schema.Resource{
								Schema: ruleComponent("destination", "datasource"),
							},
						},
						"exclude_destination": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "If set, reverses the content of the destination elements",
						},
						"applied_to": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "List of elements to which this rule applies",
							Elem: &schema.Resource{
								Schema: ruleComponent("apply-to", "datasource"),
							},
						},
					},
				},
			},
		},
	}
}

func datasourceVcdNsxvDistributedFirewallRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericVcdNsxvDistributedFirewallRead(ctx, d, meta, "datasource")
}

/*

func datasourceVcdNsxvDistributedFirewallRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	//org, err := vcdClient.GetOrgFromResource(d)
	//if err != nil {
	//	return diag.Errorf("[Distributed Firewall DS Read] error retrieving Org: %s", err)
	//}

	vdcId := d.Get("vdc_id").(string)
	//vdc, err := org.GetVDCById(vdcId, false)
	//if err != nil {
	//	return diag.Errorf("[NSXV Distributed Firewall DS Read] error retrieving VDC: %s", err)
	//}

	dfw := govcd.NewNsxvDistributedFirewall(&vcdClient.Client, vdcId)
	enabled, err := dfw.IsEnabled()

	if err != nil {
		return diag.Errorf("[NSXV Distributed Firewall DS Read] error retrieving NSX-V Firewall state: %s", err)
	}
	if !enabled {
		return diag.Errorf("VDC '%s' does not have distributed firewall enabled", vdcId)
	}
	util.Logger.Println("[NSXV DFW START]")
	configuration, err := dfw.GetConfiguration()
	if err != nil {
		return diag.Errorf("[NSXV Distributed Firewall DS Read] error retrieving NSX-V Firewall Rules: %s", err)
	}
	util.Logger.Printf("%# v\n", pretty.Formatter(configuration))
	util.Logger.Println("[NSXV DFW END]")
	confText, err := json.MarshalIndent(configuration, " ", " ")
	if err != nil {
		return diag.Errorf("[NSXV Distributed Firewall DS Read] error encoding configuration into JSON: %s", err)
	}
	dSet(d, "rules", string(confText))
	d.SetId(vdcId)

	return nil
}


*/
