package vcd

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/kr/pretty"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/util"
)

// ruleComponent can define one of the following:
// * a source
// * a destination
// * an "apply-to" clause
// * a named service
func ruleComponent(label string) map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: fmt.Sprintf("Name of the %s entity", label),
		},
		"type": {
			Type:        schema.TypeString,
			Required:    true,
			Description: fmt.Sprintf("Type of the %s entity (one of Network, Edge, VirtualMachine, IpSet, VDC, Ipv4Address, Ipv6Address)", label),
		},
		"value": {
			Type:        schema.TypeString,
			Required:    true,
			Description: fmt.Sprintf("Value of the %s entity", label),
		},
	}
}
func resourceVcdNsxvDistributedFirewall() *schema.Resource {
	return &schema.Resource{
		ReadContext:   resourceVcdNsxvDistributedFirewallRead,
		UpdateContext: resourceVcdNsxvDistributedFirewallUpdate,
		DeleteContext: resourceVcdNsxvDistributedFirewallDelete,

		Schema: map[string]*schema.Schema{
			"org": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Description: "The name of organization to use, optional if defined at provider " +
					"level. Useful when connected as sysadmin working across different organizations",
			},
			"vdc_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The ID of VDC",
			},
			"enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Description: "When true, it enables the NSX-V distributed firewall. When false, it disables the firewall. " +
					"If this property is false, existing distributed firewall rules will be removed completely.",
			},
			"rule": {
				Type:        schema.TypeList,
				Optional:    true,
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
							Optional:    true,
							Description: "Firewall Rule name",
						},
						"enabled": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     true,
							Description: "Whether the rule is enabled",
						},
						"logged": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Whether the rule traffic is logged",
						},
						"action": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Action of the rule (allow, deny)",
						},
						"direction": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "in",
							Description: "Direction of the rule (in, out, inout)",
						},
						"packet_type": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "Any",
							Description: "Packet type of the rule (any, ipv4, ipv6)",
						},
						"source": {
							Type:        schema.TypeList,
							Optional:    true,
							Description: "List of source traffic for this rule. Leaving it empty means 'any'",
							Elem: &schema.Resource{
								Schema: ruleComponent("source"),
							},
						},
						"named_service": {
							Type:        schema.TypeList,
							Optional:    true,
							Description: "Named service definitions for this rule. Leaving it empty means 'any'",
							Elem: &schema.Resource{
								Schema: ruleComponent("service"),
							},
						},
						"service": {
							Type:        schema.TypeList,
							Optional:    true,
							Description: "Raw service definitions for this rule. Leaving it empty means 'any'",
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"protocol": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "Protocol of the service (one of TCP, UDP, ICMP) (When not using name/value)",
									},
									"source_port": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "Source port for this service. Leaving it empty means 'any' port",
									},
									"destination_port": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "Destination port for this service. Leaving it empty means 'any' port",
									},
									"name": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "Name of service",
									},
									"value": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "Value of the service",
									},
									"type": {
										Type:        schema.TypeString,
										Optional:    true,
										Description: "Type of service",
									},
								},
							},
						},
						"exclude_source": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "If set, reverses the content of the source elements",
						},
						"destination": {
							Type:        schema.TypeList,
							Optional:    true,
							Description: "List of destination traffic for this rule. Leaving it empty means 'any'",
							Elem: &schema.Resource{
								Schema: ruleComponent("destination"),
							},
						},
						"exclude_destination": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "If set, reverses the content of the destination elements",
						},
						"applies_to": {
							Type:        schema.TypeList,
							Optional:    true,
							Description: "List of elements to which this rule applies",
							Elem: &schema.Resource{
								Schema: ruleComponent("apply-to"),
							},
						},
					},
				},
			},
		},
	}
}

func resourceVcdNsxvDistributedFirewallRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	org, err := vcdClient.GetOrgFromResource(d)
	if err != nil {
		return diag.Errorf("[Distributed Firewall DS Read] error retrieving Org: %s", err)
	}

	vdcId := d.Get("vdc_id").(string)
	vdc, err := org.GetVDCById(vdcId, false)
	if err != nil {
		return diag.Errorf("[NSXV Distributed Firewall DS Read] error retrieving VDC: %s", err)
	}

	dfw := govcd.NewNsxvDistributedFirewall(&vcdClient.Client, vdcId)
	enabled, err := dfw.IsEnabled()

	if err != nil {
		return diag.Errorf("[NSXV Distributed Firewall DS Read] error retrieving NSX-V Firewall state: %s", err)
	}
	if !enabled {
		return diag.Errorf("VDC '%s' does not have distributed firewall enabled", vdc.Vdc.Name)
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
	d.SetId(vdc.Vdc.ID)

	return nil
}

func resourceVcdNsxvDistributedFirewallUpdate(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return diag.Errorf("not implemented yet")
}

func resourceVcdNsxvDistributedFirewallDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	vdcId := d.Get("vdc_id").(string)
	if vdcId == "" {
		vdcId = d.Id()
	}
	dfw := govcd.NewNsxvDistributedFirewall(&vcdClient.Client, vdcId)
	err := dfw.Disable()
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}
