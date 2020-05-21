package vcd

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"log"
	"strings"
)

func resourceVcdVappFirewallRules() *schema.Resource {
	return &schema.Resource{
		Create: resourceVcdVapFirewallRulesCreateUpdate,
		Delete: resourceVAppFirewallRulesDelete,
		Read:   resourceVappFirewallRulesRead,
		Update: resourceVcdVapFirewallRulesCreateUpdate,
		Importer: &schema.ResourceImporter{
			State: vappFirewallRuleImport,
		},

		Schema: map[string]*schema.Schema{
			"org": {
				Type:     schema.TypeString,
				Required: false,
				Optional: true,
				ForceNew: true,
				Description: "The name of organization to use, optional if defined at provider " +
					"level. Useful when connected as sysadmin working across different organizations",
			},
			"vdc": {
				Type:        schema.TypeString,
				Required:    false,
				Optional:    true,
				ForceNew:    true,
				Description: "The name of VDC to use, optional if defined at provider level",
			},
			"vapp_id": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "vApp identifier",
			},
			"network_id": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "vApp network identifier",
			},
			"default_action": &schema.Schema{
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"allow", "drop"}, false),
				Description:  "Specifies what to do should none of the rules match. Either `allow` or `drop`",
			},
			"log_default_action": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Flag to enable logging for default action. Default value is false.",
			},
			"rule": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"description": &schema.Schema{
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Rule name",
						},
						"enabled": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     true,
							Description: "'true' value will enable firewall rule",
						},
						"policy": &schema.Schema{
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"drop", "allow"}, false),
							Description:  "One of: `drop` (drop packets that match the rule), `allow` (allow packets that match the rule to pass through the firewall)",
						},
						"protocol": &schema.Schema{
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"any", "icmp", "tcp", "udp", "tcp&udp"}, true),
							Description:  "Specify the protocols to which the rule should be applied. Possible one of: `any`, `icmp`, `tcp`, `udp`, `tcp&udp`",
						},
						"destination_port": &schema.Schema{
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Destination port range to which this rule applies.",
						},
						"destination_ip": &schema.Schema{
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Destination IP address to which the rule applies. A value of Any matches any IP address.",
						},
						"destination_vm_id": &schema.Schema{
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Destination VM identifier",
						},
						"destination_vm_ip_type": &schema.Schema{
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"assigned", "NAT"}, false),
							Description:  "The value can be one of: `assigned` - assigned internal IP be automatically choosen. `NAT`: NATed external IP will be automatically choosen.",
						},
						"destination_vm_nic_id": &schema.Schema{
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "VM NIC ID to which this rule applies.",
						},
						"source_port": &schema.Schema{
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Source port range to which this rule applies.",
						},
						"source_ip": &schema.Schema{
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Source IP address to which the rule applies. A value of Any matches any IP address.",
						},
						"source_vm_id": &schema.Schema{
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Source VM identifier",
						},
						"source_vm_ip_type": &schema.Schema{
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringInSlice([]string{"assigned", "NAT"}, false),
							Description:  "The value can be one of: `assigned` - assigned internal IP be automatically choosen. `NAT`: NATed external IP will be automatically choosen.",
						},
						"source_vm_nic_id": &schema.Schema{
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "VM NIC ID to which this rule applies.",
						},
						"enable_logging": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "'true' value will enable rule logging. Default is false",
						},
					},
				},
			},
		},
	}
}

func resourceVcdVapFirewallRulesCreateUpdate(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return fmt.Errorf(errorRetrievingOrgAndVdc, err)
	}

	vappId := d.Get("vapp_id").(string)
	vapp, err := vdc.GetVAppById(vappId, false)
	if err != nil {
		return fmt.Errorf("error finding vApp. %s", err)
	}

	vcdClient.lockParentVappWithName(d, vapp.VApp.Name)
	defer vcdClient.unLockParentVappWithName(d, vapp.VApp.Name)

	networkId := d.Get("network_id").(string)
	firewallRules, err := expandVappFirewallRules(d, vapp)
	if err != nil {
		return fmt.Errorf("error expanding firewall rules: %s", err)
	}

	vappNetwork, err := vapp.UpdateNetworkFirewallRules(networkId, firewallRules,
		d.Get("default_action").(string), d.Get("log_default_action").(bool))
	if err != nil {
		log.Printf("[INFO] Error setting firewall rules: %s", err)
		return fmt.Errorf("error setting firewall rules: %#v", err)
	}

	d.SetId(vappNetwork.ID)

	return resourceVappFirewallRulesRead(d, meta)
}

func resourceVAppFirewallRulesDelete(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return fmt.Errorf(errorRetrievingOrgAndVdc, err)
	}

	vappId := d.Get("vapp_id").(string)
	vapp, err := vdc.GetVAppById(vappId, false)
	if err != nil {
		return fmt.Errorf("error finding vApp. %s", err)
	}

	vcdClient.lockParentVappWithName(d, vapp.VApp.Name)
	defer vcdClient.unLockParentVappWithName(d, vapp.VApp.Name)

	_, err = vapp.UpdateNetworkFirewallRules(d.Get("network_id").(string), []*types.FirewallRule{},
		d.Get("default_action").(string), d.Get("log_default_action").(bool))
	if err != nil {
		log.Printf("[INFO] Error setting firewall rules: %s", err)
		return fmt.Errorf("error setting firewall rules: %#v", err)
	}

	return nil
}

func resourceVappFirewallRulesRead(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return fmt.Errorf(errorRetrievingOrgAndVdc, err)
	}

	vappId := d.Get("vapp_id").(string)
	vapp, err := vdc.GetVAppById(vappId, false)
	if err != nil {
		return fmt.Errorf("error finding vApp. %s", err)
	}

	vappNetwork, err := vapp.GetVappNetworkById(d.Get("network_id").(string), false)
	if err != nil {
		return fmt.Errorf("error finding vApp network. %s", err)
	}

	var rules []map[string]interface{}
	for _, rule := range vappNetwork.Configuration.Features.FirewallService.FirewallRule {
		singleRule := make(map[string]interface{})
		singleRule["description"] = rule.Description
		singleRule["enabled"] = rule.IsEnabled
		singleRule["policy"] = rule.Policy
		singleRule["protocol"] = getProtocol(*rule.Protocols)
		singleRule["destination_port"] = strings.ToLower(rule.DestinationPortRange)
		singleRule["destination_ip"] = strings.ToLower(rule.DestinationIP)
		if rule.DestinationVM != nil {
			singleRule["destination_vm_id"] = getVmIdFromVmVappLocalId(vapp, rule.DestinationVM.VAppScopedVMID)
			singleRule["destination_vm_nic_id"] = rule.DestinationVM.VMNicID
			singleRule["destination_vm_ip_type"] = rule.DestinationVM.IPType
		}
		singleRule["source_port"] = strings.ToLower(rule.SourcePortRange)
		singleRule["source_ip"] = strings.ToLower(rule.SourceIP)
		if rule.SourceVM != nil {
			singleRule["source_vm_id"] = getVmIdFromVmVappLocalId(vapp, rule.SourceVM.VAppScopedVMID)
			singleRule["source_vm_nic_id"] = rule.SourceVM.VMNicID
			singleRule["source_vm_ip_type"] = rule.SourceVM.IPType
		}
		singleRule["enable_logging"] = rule.EnableLogging
		rules = append(rules, singleRule)
	}
	_ = d.Set("rule", rules)
	_ = d.Set("default_action", vappNetwork.Configuration.Features.FirewallService.DefaultAction)
	_ = d.Set("log_default_action", vappNetwork.Configuration.Features.FirewallService.LogDefaultAction)

	return nil
}

func getVmIdFromVmVappLocalId(vapp *govcd.VApp, vmVappLocalId string) string {
	for _, vm := range vapp.VApp.Children.VM {
		if vm.VAppScopedLocalID == vmVappLocalId {
			return vm.ID
		}
	}
	return ""
}

func expandVappFirewallRules(d *schema.ResourceData, vapp *govcd.VApp) ([]*types.FirewallRule, error) {
	firewallRules := []*types.FirewallRule{}
	for _, singleRule := range d.Get("rule").([]interface{}) {
		configuredRule := singleRule.(map[string]interface{})

		var protocol *types.FirewallRuleProtocols
		// Allow upper and lower case protocol names
		switch strings.ToLower(configuredRule["protocol"].(string)) {
		case "tcp":
			protocol = &types.FirewallRuleProtocols{
				TCP: true,
			}
		case "udp":
			protocol = &types.FirewallRuleProtocols{
				UDP: true,
			}
		case "icmp":
			protocol = &types.FirewallRuleProtocols{
				ICMP: true,
			}
		case "tcp&udp":
			protocol = &types.FirewallRuleProtocols{
				TCP: true,
				UDP: true,
			}
		default:
			protocol = &types.FirewallRuleProtocols{
				Any: true,
			}
		}
		rule := &types.FirewallRule{
			IsEnabled:            configuredRule["enabled"].(bool),
			MatchOnTranslate:     false,
			Description:          configuredRule["description"].(string),
			Policy:               configuredRule["policy"].(string),
			Protocols:            protocol,
			Port:                 getNumericPort(configuredRule["destination_port"]),
			DestinationPortRange: strings.ToLower(configuredRule["destination_port"].(string)),
			DestinationIP:        strings.ToLower(configuredRule["destination_ip"].(string)),
			SourcePort:           getNumericPort(configuredRule["source_port"]),
			SourcePortRange:      strings.ToLower(configuredRule["source_port"].(string)),
			SourceIP:             strings.ToLower(configuredRule["source_ip"].(string)),
			EnableLogging:        configuredRule["enable_logging"].(bool),
		}

		if configuredRule["destination_vm_id"].(string) != "" {
			vm, err := vapp.GetVMById(configuredRule["destination_vm_id"].(string), false)
			if err != nil {
				return nil, fmt.Errorf("error fetchining VM: %s", err)
			}

			rule.DestinationVM = &types.VMSelection{VAppScopedVMID: vm.VM.VAppScopedLocalID,
				VMNicID: configuredRule["destination_vm_nic_id"].(int), IPType: configuredRule["destination_vm_ip_type"].(string)}
		}
		if configuredRule["source_vm_id"].(string) != "" {
			vm, err := vapp.GetVMById(configuredRule["source_vm_id"].(string), false)
			if err != nil {
				return nil, fmt.Errorf("error fetchining VM: %s", err)
			}

			rule.SourceVM = &types.VMSelection{VAppScopedVMID: vm.VM.VAppScopedLocalID,
				VMNicID: configuredRule["source_vm_nic_id"].(int), IPType: configuredRule["source_vm_ip_type"].(string)}
		}
		firewallRules = append(firewallRules, rule)
	}

	return firewallRules, nil
}

// vappFirewallRuleImport is responsible for importing the resource.
// The following steps happen as part of import
// 1. The user supplies `terraform import _resource_name_ _the_id_string_` command
// 2. `_the_id_string_` contains a dot formatted path to resource as in the example below
// 3. The functions splits the dot-formatted path and tries to lookup the object
// 4. If the lookup succeeds it set's the ID field for `_resource_name_` resource in state file
// (the resource must be already defined in .tf config otherwise `terraform import` will complain)
// 5. `terraform refresh` is being implicitly launched. The Read method looks up all other fields
// based on the known ID of object.
//
// Example resource name (_resource_name_): vcd_vapp_firewall_rules.my_existing_firewall_rules
// Example import path (_the_id_string_): org.my_existing_vdc.vapp_name.network_name
// Note: the separator can be changed using Provider.import_separator or variable VCD_IMPORT_SEPARATOR
func vappFirewallRuleImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	resourceURI := strings.Split(d.Id(), ImportSeparator)
	if len(resourceURI) != 4 {
		return nil, fmt.Errorf("resource name must be specified as rg.my_existing_vdc.vapp_name.network_name")
	}
	orgName, vdcName, vappId, networkId := resourceURI[0], resourceURI[1], resourceURI[2], resourceURI[3]

	vcdClient := meta.(*VCDClient)

	_, vdc, err := vcdClient.GetOrgAndVdc(orgName, vdcName)
	if err != nil {
		return nil, fmt.Errorf(errorRetrievingOrgAndVdc, err)
	}

	vapp, err := vdc.GetVAppByNameOrId(vappId, false)
	if err != nil {
		return nil, fmt.Errorf("error retrieving vApp %s:%s", vappId, err)
	}

	vappNetwork, err := vapp.GetVappNetworkByNameOrId(networkId, false)
	if err != nil {
		return nil, fmt.Errorf("error retrieving vApp network %s:%s", networkId, err)
	}

	if vcdClient.Org != orgName {
		d.Set("org", orgName)
	}
	if vcdClient.Vdc != vdcName {
		d.Set("vdc", vdcName)
	}
	_ = d.Set("vapp_id", vapp.VApp.ID)
	_ = d.Set("network_id", vappNetwork.ID)
	d.SetId(vappNetwork.ID)

	return []*schema.ResourceData{d}, nil
}
