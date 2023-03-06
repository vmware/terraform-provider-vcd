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
	"strings"
)

var DFWElements = []string{
	govcd.DFWElementIpv4,
	govcd.DFWElementNetwork,
	govcd.DFWElementEdge,
	govcd.DFWElementIpSet,
	govcd.DFWElementVirtualMachine,
	govcd.DFWElementVdc,
}

func sourceDef() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the source entity",
			},
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "Type of the source entity (one of Network, Edge, VirtualMachine, IpSet, VDC, Ipv4Address)",
				ValidateFunc: validation.StringInSlice(DFWElements, false),
			},
			"value": {
				Type:        schema.TypeString,
				Required:    true,
				StateFunc:   filterVdcId,
				Description: "Value of the source entity",
			},
		},
	}
}

func destinationDef() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the destination entity",
			},
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "Type of the destination entity (one of Network, Edge, VirtualMachine, IpSet, VDC, Ipv4Address)",
				ValidateFunc: validation.StringInSlice(DFWElements, false),
			},
			"value": {
				Type:        schema.TypeString,
				Required:    true,
				StateFunc:   filterVdcId,
				Description: "Value of the destination entity",
			},
		},
	}
}

func appliedToDef() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the applied-to entity",
			},
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "Type of the applied-to entity (one of Network, Edge, VirtualMachine, IPSet, VDC, Ipv4Address)",
				ValidateFunc: validation.StringInSlice(DFWElements, false),
			},
			"value": {
				Type:        schema.TypeString,
				Required:    true,
				StateFunc:   filterVdcId,
				Description: "Value of the applied-to entity",
			},
		},
	}
}

func applicationDef() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"protocol": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Protocol of the application (one of TCP, UDP, ICMP) (When not using name/value)",
			},
			"source_port": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Source port for this application. Leaving it empty means 'any' port",
			},
			"destination_port": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Destination port for this application. Leaving it empty means 'any' port",
			},
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Name of application (Application, ApplicationGroup)",
			},
			"value": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Value of the application",
				StateFunc:   filterVdcId,
			},
			"type": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Type of application",
			},
		},
	}
}

func resourceVcdNsxvDistributedFirewall() *schema.Resource {
	return &schema.Resource{
		ReadContext:   resourceVcdNsxvDistributedFirewallRead,
		CreateContext: resourceVcdNsxvDistributedFirewallCreate,
		UpdateContext: resourceVcdNsxvDistributedFirewallUpdate,
		DeleteContext: resourceVcdNsxvDistributedFirewallDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdNsxvDistributedFirewallImport,
		},

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
							Type:        schema.TypeInt,
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
							Type:         schema.TypeString,
							Required:     true,
							Description:  "Action of the rule (allow, deny)",
							ValidateFunc: validation.StringInSlice([]string{"allow", "deny"}, false),
						},
						"direction": {
							Type:         schema.TypeString,
							Required:     true,
							Description:  "Direction of the rule (in, out, inout)",
							ValidateFunc: validation.StringInSlice([]string{"in", "out", "inout"}, false),
						},
						"packet_type": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "any",
							Description:  "Packet type of the rule (any, ipv4, ipv6)",
							ValidateFunc: validation.StringInSlice([]string{"any", "ipv4", "ipv6"}, false),
						},
						"source": {
							Type:        schema.TypeSet,
							Optional:    true,
							Description: "List of source traffic for this rule. Leaving it empty means 'any'",
							Elem:        sourceDef(),
						},
						"application": {
							Type:        schema.TypeSet,
							Optional:    true,
							Description: "Application definitions for this rule. Leaving it empty means 'any'",
							Elem:        applicationDef(),
						},
						"exclude_source": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "If set, reverses the content of the source elements",
						},
						"destination": {
							Type:        schema.TypeSet,
							Optional:    true,
							Description: "List of destination traffic for this rule. Leaving it empty means 'any'",
							Elem:        destinationDef(),
						},
						"exclude_destination": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "If set, reverses the content of the destination elements",
						},
						"applied_to": {
							Type:        schema.TypeSet,
							Required:    true,
							Description: "List of elements to which this rule applies",
							Elem:        appliedToDef(),
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

	if err != nil {
		if origin == "datasource" {
			return diag.Errorf("[NSX-V Distributed Firewall DS Read] error retrieving NSX-V Firewall state: %s - %s", err, govcd.ErrorEntityNotFound)
		}
		d.SetId("")
		return nil
	}
	d.SetId(vdcId)
	if configuration == nil { // disabled
		util.Logger.Println("[NSX-V DFW DISABLED]")
		dSet(d, "enabled", false)
		err = d.Set("rule", nil)
		if err != nil {
			return diag.FromErr(err)
		}
		return nil
	}

	util.Logger.Println("[NSX-V DFW START]")
	err = dfwRulesToResource(dfw.Configuration.Layer3Sections.Section.Rule, d)
	if err != nil {
		return diag.Errorf("error setting distributed firewall state: %s", err)
	}
	util.Logger.Println("[NSX-V DFW END]")

	return nil
}

func resourceVcdNsxvDistributedFirewallCreateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}, origin string) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	vdcId := d.Get("vdc_id").(string)
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
		isEnabled = true
	}
	if !isEnabled {
		return diag.Errorf("distributed firewall for VDC %s needs to be enabled before being configured", vdcId)
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
	return resourceVcdNsxvDistributedFirewallCreateUpdate(ctx, d, meta, "create")
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
	util.Logger.Printf("[INFO] disabling NSX-V distributed firewall for VDC %s\n", vdcId)
	err := dfw.Disable()
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}

// resourceToDfwRules takes the data from HCL and returns a list of NSX-V distributed firewall rules
func resourceToDfwRules(d *schema.ResourceData) ([]types.NsxvDistributedFirewallRule, error) {

	var resultRules []types.NsxvDistributedFirewallRule
	rawRuleList, ok := d.Get("rule").([]interface{})
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
			Disabled:   !ruleMap["enabled"].(bool),
			Logged:     ruleMap["logged"].(bool),
			Name:       ruleMap["name"].(string),
			Action:     ruleMap["action"].(string),
			Direction:  ruleMap["direction"].(string),
			PacketType: ruleMap["packet_type"].(string),
			Services:   nil,
		}
		rawSources, ok := ruleMap["source"]
		if ok && rawSources != nil {
			sourceSet := rawSources.(*schema.Set)
			var sources []types.Source
			for j, s := range sourceSet.List() {
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
				// TODO: find why sourceSet.List() returns an extra empty item
				if inputSource.Name == "" && inputSource.Type == "" {
					continue
				}
				// When the source is an IP address, the name should be filled with the same value
				if inputSource.Name == "" && inputSource.Type == govcd.DFWElementIpv4 {
					inputSource.Name = inputSource.Value
				}
				sources = append(sources, inputSource)
			}
			resultRule.Sources = &types.Sources{
				Excluded: excludeSource,
				Source:   sources,
			}
		}
		rawDestinations, ok := ruleMap["destination"]
		if ok && rawDestinations != nil {
			destinationSet := rawDestinations.(*schema.Set)
			var destinations []types.Destination
			for j, dest := range destinationSet.List() {
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
				// TODO: find why inputDestination.List() returns an extra empty item
				if inputDestination.Name == "" && inputDestination.Type == "" {
					continue
				}
				// When the destination is an IP address, the name should be filled with the same value
				if inputDestination.Name == "" && inputDestination.Type == govcd.DFWElementIpv4 {
					inputDestination.Name = inputDestination.Value
				}
				destinations = append(destinations, inputDestination)
			}
			resultRule.Destinations = &types.Destinations{
				Excluded:    excludeDestination,
				Destination: destinations,
			}
		}
		rawAppliedTo, ok := ruleMap["applied_to"]
		if ok && rawAppliedTo != nil {
			var appliedTo []types.AppliedTo
			appliedToSet := rawAppliedTo.(*schema.Set)
			for j, a := range appliedToSet.List() {
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
				// TODO: find why inputApplyTo.List() returns an extra empty item
				if inputApplyTo.Name == "" && inputApplyTo.Type == "" {
					continue
				}
				appliedTo = append(appliedTo, inputApplyTo)
			}
			resultRule.AppliedToList = &types.AppliedToList{
				AppliedTo: appliedTo,
			}
		}
		rawApplications, ok := ruleMap["application"]
		if ok && rawApplications != nil {
			var applications []types.Service
			applicationSet := rawApplications.(*schema.Set)
			for j, s := range applicationSet.List() {
				application, ok := s.(map[string]interface{})
				if !ok {
					return nil, fmt.Errorf("[resourceToDfwRules] rule %d - application %d - expected map[string]interface{} - got %T", i, j, s)
				}
				sourcePort := application["source_port"].(string)
				destinationPort := application["destination_port"].(string)
				protocol := application["protocol"].(string)

				inputApplication := types.Service{
					Name:            application["name"].(string),
					Value:           application["value"].(string),
					Type:            application["type"].(string),
					SourcePort:      stringPtrOrNil(sourcePort),
					DestinationPort: stringPtrOrNil(destinationPort),
					Protocol:        getDfwProtocolCode(protocol),
					IsValid:         true,
				}
				applications = append(applications, inputApplication)
			}
			resultRule.Services = &types.Services{
				Service: applications,
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

// dfwRulesToResource gets the rules from a NSX-V distributed firewall and sets the corresponding state
func dfwRulesToResource(rules []types.NsxvDistributedFirewallRule, d *schema.ResourceData) error {
	var rulesList []map[string]interface{}

	for _, rule := range rules {
		ruleMap := make(map[string]interface{})
		ruleMap["enabled"] = !rule.Disabled
		ruleMap["logged"] = rule.Logged
		ruleMap["name"] = rule.Name
		ruleMap["action"] = rule.Action
		ruleMap["direction"] = rule.Direction
		ruleMap["packet_type"] = rule.PacketType
		if rule.ID != nil {
			ruleMap["id"] = *rule.ID
		}

		// sources
		excludeSource := false
		if rule.Sources != nil && len(rule.Sources.Source) > 0 {
			var rawSourceList []interface{}
			if rule.Sources.Excluded {
				excludeSource = true
			}
			for _, s := range rule.Sources.Source {
				sourceMap := map[string]interface{}{
					"name":  s.Name,
					"type":  s.Type,
					"value": filterVdcId(s.Value),
				}
				rawSourceList = append(rawSourceList, sourceMap)
			}
			sourceSet := schema.NewSet(schema.HashResource(sourceDef()), rawSourceList)
			ruleMap["source"] = sourceSet
		}
		// destinations
		excludeDestination := false
		if rule.Destinations != nil && len(rule.Destinations.Destination) > 0 {
			var destinationList []interface{}
			if rule.Destinations.Excluded {
				excludeDestination = true
			}
			for _, dest := range rule.Destinations.Destination {
				destinationMap := map[string]interface{}{
					"name":  dest.Name,
					"type":  dest.Type,
					"value": filterVdcId(dest.Value),
				}
				destinationList = append(destinationList, destinationMap)
			}
			destinationSet := schema.NewSet(schema.HashResource(destinationDef()), destinationList)
			ruleMap["destination"] = destinationSet
		}
		// applications
		if rule.Services != nil && len(rule.Services.Service) > 0 {
			var applicationList []interface{}
			for _, s := range rule.Services.Service {
				applicationMap := map[string]interface{}{
					"name":             s.Name,
					"type":             s.Type,
					"destination_port": stringOnNotNil(s.DestinationPort),
					"source_port":      stringOnNotNil(s.SourcePort),
					"protocol":         getDfwProtocolString(s.Protocol),
					"value":            filterVdcId(s.Value),
				}
				applicationList = append(applicationList, applicationMap)
			}
			applicationSet := schema.NewSet(schema.HashResource(applicationDef()), applicationList)
			ruleMap["application"] = applicationSet
		}
		// applied-to
		if rule.AppliedToList != nil && len(rule.AppliedToList.AppliedTo) > 0 {
			var appliedToList []interface{}
			for _, a := range rule.AppliedToList.AppliedTo {
				appliedToMap := map[string]interface{}{
					"name":  a.Name,
					"type":  a.Type,
					"value": filterVdcId(a.Value),
				}
				appliedToList = append(appliedToList, appliedToMap)
			}
			appliedToSet := schema.NewSet(schema.HashResource(appliedToDef()), appliedToList)
			ruleMap["applied_to"] = appliedToSet
		}
		if excludeDestination {
			ruleMap["exclude_destination"] = true
		}
		if excludeSource {
			ruleMap["exclude_source"] = true
		}
		rulesList = append(rulesList, ruleMap)
	}

	return d.Set("rule", rulesList)
}

// stringOnNotNil returns the contents of a string pointer
// if the pointer is nil, returns an empty string
func stringOnNotNil(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func resourceVcdNsxvDistributedFirewallImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	resourceURI := strings.Split(d.Id(), ImportSeparator)

	vcdClient := meta.(*VCDClient)
	var dfw *govcd.NsxvDistributedFirewall

	var vdcId string
	if len(resourceURI) == 1 { // only VDC-ID
		uuid := extractUuid(resourceURI[0])
		if uuid == "" {
			return nil, fmt.Errorf("not a valid ID provided for VDC")
		}
		vdcId = resourceURI[0]
		dfw = govcd.NewNsxvDistributedFirewall(&vcdClient.Client, vdcId)
	} else {
		if len(resourceURI) == 2 { // Org name + VDC name
			orgName := resourceURI[0]
			vdcName := resourceURI[1]
			org, err := vcdClient.GetOrg(orgName)
			if err != nil {
				return nil, err
			}
			vdc, err := org.GetVDCByName(vdcName, false)
			if err != nil {
				return nil, err
			}
			vdcId = vdc.Vdc.ID
			dfw = govcd.NewNsxvDistributedFirewall(&vcdClient.Client, vdcId)
		}
	}
	configuration, err := dfw.GetConfiguration()
	if err != nil {
		return nil, err
	}
	if configuration == nil {
		return nil, fmt.Errorf("distributed firewall for VDC %s is not enabled", dfw.VdcId)
	}
	d.SetId(vdcId)
	dSet(d, "enabled", true)
	dSet(d, "vdc_id", vdcId)
	return []*schema.ResourceData{d}, nil
}
