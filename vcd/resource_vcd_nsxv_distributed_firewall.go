package vcd

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"github.com/vmware/go-vcloud-director/v2/util"
)

// ruleComponent can define one of the following:
// * a source
// * a destination
// * an "apply-to" clause
func ruleComponent(label, origin string) map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": {
			Type:        schema.TypeString,
			Required:    origin == "resource",
			Computed:    origin == "datasource",
			Description: fmt.Sprintf("Name of the %s entity", label),
		},
		"type": {
			Type:        schema.TypeString,
			Required:    origin == "resource",
			Computed:    origin == "datasource",
			Description: fmt.Sprintf("Type of the %s entity (one of Network, Edge, VirtualMachine, IpSet, VDC, Ipv4Address, Ipv6Address)", label),
		},
		"value": {
			Type:        schema.TypeString,
			Required:    origin == "resource",
			Computed:    origin == "datasource",
			Description: fmt.Sprintf("Value of the %s entity", label),
		},
	}
}

func resourceVcdNsxvDistributedFirewall() *schema.Resource {
	return &schema.Resource{
		ReadContext:   resourceVcdNsxvDistributedFirewallRead,
		CreateContext: resourceVcdNsxvDistributedFirewallCreate,
		UpdateContext: resourceVcdNsxvDistributedFirewallUpdate,
		DeleteContext: resourceVcdNsxvDistributedFirewallDelete,

		Schema: map[string]*schema.Schema{
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
								Schema: ruleComponent("source", "resource"),
							},
						},
						"service": {
							Type:        schema.TypeList,
							Optional:    true,
							Description: "Service definitions for this rule. Leaving it empty means 'any'",
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
								Schema: ruleComponent("destination", "resource"),
							},
						},
						"exclude_destination": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "If set, reverses the content of the destination elements",
						},
						"applied_to": {
							Type:        schema.TypeList,
							Optional:    true,
							Description: "List of elements to which this rule applies",
							Elem: &schema.Resource{
								Schema: ruleComponent("apply-to", "resource"),
							},
						},
					},
				},
			},
		},
	}
}

func resourceVcdNsxvDistributedFirewallRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericVcdNsxvDistributedFirewallRead(ctx, d, meta, "resource")
}

func genericVcdNsxvDistributedFirewallRead(_ context.Context, d *schema.ResourceData, meta interface{}, origin string) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	vdcId := d.Get("vdc_id").(string)
	dfw := govcd.NewNsxvDistributedFirewall(&vcdClient.Client, vdcId)
	configuration, err := dfw.GetConfiguration()
	//enabled, err := dfw.IsEnabled()

	if err != nil {
		if origin == "datasource" {
			return diag.Errorf("[NSXV Distributed Firewall DS Read] error retrieving NSX-V Firewall state: %s - %s", err, govcd.ErrorEntityNotFound)
		}
		d.SetId("")
		return nil
	}
	d.SetId(vdcId)
	if configuration == nil { // disabled
		dSet(d, "enabled", false)
		err = d.Set("rule", nil)
		if err != nil {
			return diag.FromErr(err)
		}
		return nil
	}

	util.Logger.Println("[NSXV DFW START]")
	err = dfwRulesToResource(dfw.Configuration.Layer3Sections.Section.Rule, d)
	if err != nil {
		return diag.Errorf("error setting distributed firewall state: %s", err)
	}
	util.Logger.Println("[NSXV DFW END]")

	return nil
}

func resourceVcdNsxvDistributedFirewallCreateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}, origin string) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	vdcId := d.Get("vdc_id").(string)
	if vdcId == "" {
		vdcId = d.Id()
	}
	wantToEnable := d.Get("enabled").(bool)
	dfw := govcd.NewNsxvDistributedFirewall(&vcdClient.Client, vdcId)
	isEnabled, err := dfw.IsEnabled()
	if err != nil {
		return diag.FromErr(err)
	}
	if !isEnabled && wantToEnable {
		err := dfw.Enable()
		if err != nil {
			return diag.FromErr(err)
		}
	}
	if !wantToEnable {
		err := dfw.Disable()
		if err != nil {
			return diag.FromErr(err)
		}
		return resourceVcdNsxvDistributedFirewallRead(ctx, d, meta)
	}
	rules, err := resourceToDfwRules(d)
	if err != nil {
		return diag.FromErr(err)
	}
	_, err = dfw.UpdateConfiguration(rules)
	if err != nil {
		return diag.Errorf("error updating distributed firewall configuration: %s", err)
	}
	return resourceVcdNsxvDistributedFirewallRead(ctx, d, meta)
}

func resourceVcdNsxvDistributedFirewallCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceVcdNsxvDistributedFirewallCreateUpdate(ctx, d, meta, "creation")
}

func resourceVcdNsxvDistributedFirewallUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceVcdNsxvDistributedFirewallCreateUpdate(ctx, d, meta, "update")
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

// resourceToDfwRules takes the data from HCL and returns a list of NSXV distributed firewall rules
func resourceToDfwRules(d *schema.ResourceData) ([]types.NsxvDistributedFirewallRule, error) {

	var resultRules []types.NsxvDistributedFirewallRule
	rawRuleList, ok := d.Get("rule").([]interface{}) // []interface{}
	if !ok {
		return nil, fmt.Errorf("[resourceToDfwRules] expected interface slice - got %T", rawRuleList)
	}
	for i, rawRule := range rawRuleList {
		ruleMap, ok := rawRule.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("[resourceToDfwRules] rule %d - expected map[string]interface{} - got %T", i, rawRule)
		}
		excludeSource := ruleMap["exclude_source"].(bool)
		excludeDestination := ruleMap["exclude_destination"].(bool)
		resultRule := types.NsxvDistributedFirewallRule{
			Disabled: !ruleMap["enabled"].(bool),
			Logged:   ruleMap["logged"].(bool),
			Name:     ruleMap["name"].(string),
			Action:   ruleMap["action"].(string),
			//AppliedToList: types.AppliedToList{},
			Direction:  ruleMap["direction"].(string),
			PacketType: ruleMap["packet_type"].(string),
			//SectionID:     nil,
			//Sources:      nil,
			//Destinations: nil,
			Services: nil,
		}
		rawSources := d.Get("source")
		if rawSources != nil {
			var sources []types.Source
			sourceList := rawSources.([]interface{})
			for j, s := range sourceList {
				source, ok := s.(map[string]interface{})
				if !ok {
					return nil, fmt.Errorf("[resourceToDfwRules] rule %d - source %d - expected map[string]interface{} - got %T", i, j, s)
				}
				inputSource := types.Source{
					Name:    source["name"].(string),
					Value:   source["value"].(string),
					Type:    source["type"].(string),
					IsValid: true,
				}
				// When the source is an IP address, the name should be filled with the same value
				_, errs := validation.IsIPAddress(inputSource.Value, "source")
				if errs == nil && inputSource.Name == "" {
					inputSource.Name = inputSource.Value
				}
				sources = append(sources, inputSource)
			}
			resultRule.Sources = &types.Sources{
				Excluded: excludeSource,
				Source:   sources,
			}
		}
		rawDestinations := d.Get("destination")
		if rawDestinations != nil {
			var destinations []types.Destination
			destinationList := rawDestinations.([]interface{})
			for j, dest := range destinationList {
				destination, ok := dest.(map[string]interface{})
				if !ok {
					return nil, fmt.Errorf("[resourceToDfwRules] rule %d - destination %d - expected map[string]interface{} - got %T", i, j, dest)
				}
				inputDestination := types.Destination{
					Name:    destination["name"].(string),
					Value:   destination["value"].(string),
					Type:    destination["type"].(string),
					IsValid: true,
				}
				// When the source is an IP address, the name should be filled with the same value
				_, errs := validation.IsIPAddress(inputDestination.Value, "destination")
				if errs == nil && inputDestination.Name == "" {
					inputDestination.Name = inputDestination.Value
				}
				destinations = append(destinations, inputDestination)
			}
			resultRule.Destinations = &types.Destinations{
				Excluded:    excludeDestination,
				Destination: destinations,
			}
		}
		rawAppliedTo := d.Get("applied_to")
		if rawAppliedTo != nil {
			var appliedTo []types.AppliedTo
			appliedToList := rawAppliedTo.([]interface{})
			for j, a := range appliedToList {
				apply, ok := a.(map[string]interface{})
				if !ok {
					return nil, fmt.Errorf("[resourceToDfwRules] rule %d - applied-to %d - expected map[string]interface{} - got %T", i, j, a)
				}
				inputApplyTo := types.AppliedTo{
					Name:    apply["name"].(string),
					Value:   apply["value"].(string),
					Type:    apply["type"].(string),
					IsValid: true,
				}
				appliedTo = append(appliedTo, inputApplyTo)
			}
			resultRule.AppliedToList = &types.AppliedToList{
				AppliedTo: appliedTo,
			}
		}
		rawServices := d.Get("service")
		if rawServices != nil {
			var services []types.Service
			serviceList := rawServices.([]interface{})
			for j, s := range serviceList {
				service, ok := s.(map[string]interface{})
				if !ok {
					return nil, fmt.Errorf("[resourceToDfwRules] rule %d - service %d - expected map[string]interface{} - got %T", i, j, s)
				}
				sourcePort := service["source_port"].(string)
				destinationPort := service["destination_port"].(string)
				protocol := service["protocol"].(string)

				inputService := types.Service{
					Name:            service["name"].(string),
					Value:           service["value"].(string),
					Type:            service["type"].(string),
					SourcePort:      stringPtrOrNil(sourcePort),
					DestinationPort: stringPtrOrNil(destinationPort),
					Protocol:        getDfwProtocolCode(protocol),
					IsValid:         true,
				}
				services = append(services, inputService)
			}
			resultRule.Services = &types.Services{
				Service: services,
			}
		}
		resultRules = append(resultRules, resultRule)
	}
	return resultRules, nil
}

func getDfwProtocolCode(s string) *int {
	if s == "" {
		return nil
	}
	protocolNumber, ok := govcd.NsxvProtocolCodes[s]
	if !ok {
		return nil
	}
	return &protocolNumber
}

func getDfwProtocolString(i *int) string {
	if i == nil {
		return ""
	}
	for k, v := range govcd.NsxvProtocolCodes {
		if *i == v {
			return k
		}
	}
	return ""
}

// dfwRulesToResource gets the rules from a NSXV distributed firewall and sets the corresponding state
func dfwRulesToResource(rules []types.NsxvDistributedFirewallRule, d *schema.ResourceData) error {
	var rulesList []map[string]interface{}

	for _, rule := range rules {
		ruleMap := make(map[string]interface{})
		ruleMap["enabled"] = !rule.Disabled
		ruleMap["logged"] = rule.Logged
		ruleMap["name"] = rule.Name
		ruleMap["action"] = rule.Action
		ruleMap["direction"] = rule.Direction

		// sources
		if rule.Sources != nil && len(rule.Sources.Source) > 0 {
			var sourceList []map[string]interface{}
			for _, s := range rule.Sources.Source {
				sourceMap := map[string]interface{}{
					"name":  s.Name,
					"type":  s.Type,
					"value": s.Value,
				}
				sourceList = append(sourceList, sourceMap)
			}
			ruleMap["source"] = sourceList
		}
		// destinations
		if rule.Destinations != nil && len(rule.Destinations.Destination) > 0 {
			var destinationList []map[string]interface{}
			for _, dest := range rule.Destinations.Destination {
				destinationMap := map[string]interface{}{
					"name":  dest.Name,
					"type":  dest.Type,
					"value": dest.Value,
				}
				destinationList = append(destinationList, destinationMap)
			}
			ruleMap["destination"] = destinationList
		}
		// services
		if rule.Services != nil && len(rule.Services.Service) > 0 {
			var serviceList []map[string]interface{}
			for _, s := range rule.Services.Service {
				serviceMap := map[string]interface{}{
					"name":             s.Name,
					"type":             s.Type,
					"destination_port": *s.DestinationPort,
					"source_port":      *s.SourcePort,
					"protocol":         getDfwProtocolString(s.Protocol),
					"value":            s.Value,
				}
				serviceList = append(serviceList, serviceMap)
			}
			ruleMap["service"] = serviceList
		}
		// applied-to
		if rule.AppliedToList != nil && len(rule.AppliedToList.AppliedTo) > 0 {
			var appliedToList []map[string]interface{}
			for _, a := range rule.AppliedToList.AppliedTo {
				appliedToMap := map[string]interface{}{
					"name":  a.Name,
					"type":  a.Type,
					"value": a.Value,
				}
				appliedToList = append(appliedToList, appliedToMap)
			}
			ruleMap["applied_to"] = appliedToList
		}

		rulesList = append(rulesList, ruleMap)
	}

	return d.Set("rule", rulesList)
}
