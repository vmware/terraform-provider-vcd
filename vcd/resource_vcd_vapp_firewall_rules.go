package vcd

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"log"
	"strings"
	"text/tabwriter"
)

func resourceVcdVappFirewallRules() *schema.Resource {
	return &schema.Resource{
		Create: resourceVcdVappFirewallRulesCreate,
		Delete: resourceVappFirewallRulesDelete,
		Read:   resourceVappFirewallRulesRead,
		Update: resourceVcdVappFirewallRulesUpdate,
		Importer: &schema.ResourceImporter{
			State: vappFirewallRulesImport,
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
						"name": &schema.Schema{
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
							Default:      "any",
							ValidateFunc: validation.StringInSlice([]string{"any", "icmp", "tcp", "udp", "tcp&udp"}, true),
							Description:  "Specify the protocols to which the rule should be applied. One of: `any`, `icmp`, `tcp`, `udp`, `tcp&udp`",
						},
						"destination_port": &schema.Schema{
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Destination port to which this rule applies.",
						},
						"destination_ip": &schema.Schema{
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Destination IP address to which the rule applies. A value of `Any` matches any IP address.",
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
							Description:  "The value can be one of: `assigned` - assigned internal IP will be automatically chosen. `NAT`: NATed external IP will be automatically chosen.",
						},
						"destination_vm_nic_id": &schema.Schema{
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "Destination VM NIC ID to which this rule applies.",
						},
						"source_port": &schema.Schema{
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Source port to which this rule applies.",
						},
						"source_ip": &schema.Schema{
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Source IP address to which the rule applies. A value of `Any` matches any IP address.",
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
							Description:  "The value can be one of: `assigned` - assigned internal IP will be automatically chosen. `NAT`: NATed external IP will be automatically chosen.",
						},
						"source_vm_nic_id": &schema.Schema{
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "Source VM NIC ID to which this rule applies.",
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
func resourceVcdVappFirewallRulesCreate(d *schema.ResourceData, meta interface{}) error {
	return resourceVcdVappFirewallRulesUpdate(d, meta)
}

func resourceVcdVappFirewallRulesUpdate(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)
	vapp, err := getVapp(vcdClient, d, meta)
	if err != nil {
		return err
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
		return fmt.Errorf("error setting firewall rules: %s", err)
	}

	d.SetId(vappNetwork.ID)

	return resourceVappFirewallRulesRead(d, meta)
}

func getVapp(vcdClient *VCDClient, d *schema.ResourceData, meta interface{}) (*govcd.VApp, error) {
	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return nil, fmt.Errorf(errorRetrievingOrgAndVdc, err)
	}

	vappId := d.Get("vapp_id").(string)
	vapp, err := vdc.GetVAppById(vappId, false)
	if err != nil {
		return nil, fmt.Errorf("error finding vApp. %s", err)
	}

	return vapp, nil
}

func resourceVappFirewallRulesDelete(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)
	vapp, err := getVapp(vcdClient, d, meta)
	if err != nil {
		return err
	}

	vcdClient.lockParentVappWithName(d, vapp.VApp.Name)
	defer vcdClient.unLockParentVappWithName(d, vapp.VApp.Name)

	err = vapp.RemoveAllNetworkFirewallRules(d.Get("network_id").(string))
	if err != nil {
		log.Printf("[INFO] Error removing firewall rules: %s", err)
		return fmt.Errorf("error removing firewall rules: %s", err)
	}

	return nil
}

func resourceVappFirewallRulesRead(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)
	vapp, err := getVapp(vcdClient, d, meta)
	if err != nil {
		return err
	}

	vappNetwork, err := vapp.GetVappNetworkById(d.Get("network_id").(string), false)
	if err != nil {
		return fmt.Errorf("error finding vApp network. %s", err)
	}

	var rules []map[string]interface{}
	for _, rule := range vappNetwork.Configuration.Features.FirewallService.FirewallRule {
		singleRule := make(map[string]interface{})
		singleRule["name"] = rule.Description
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
	err = d.Set("rule", rules)
	if err != nil {
		return err
	}
	_ = d.Set("default_action", vappNetwork.Configuration.Features.FirewallService.DefaultAction)
	_ = d.Set("log_default_action", vappNetwork.Configuration.Features.FirewallService.LogDefaultAction)

	return nil
}

// getVmIdFromVmVappLocalId returns vm ID using VAppScopedLocalID.
// VAppScopedLocalID is another ID provided in VM entity.
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
			Description:          configuredRule["name"].(string),
			Policy:               configuredRule["policy"].(string),
			Protocols:            protocol,
			DestinationPortRange: strings.ToLower(configuredRule["destination_port"].(string)),
			DestinationIP:        strings.ToLower(configuredRule["destination_ip"].(string)),
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

var errHelpVappNetworkRulesImport = fmt.Errorf(`resource id must be specified in one of these formats:
'org-name.vdc-name.vapp-name.network_name', 'org.vdc-name.vapp-id.network-id' or 
'list@org-name.vdc-name.vapp-name' to get a list of vapp networks with their IDs`)

// vappFirewallRulesImport is responsible for importing the resource.
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
// Example import path (_the_id_string_): org.my_existing_vdc.vapp_name.network_name or org.my_existing_vdc.vapp_id.network_id
// Example list path (_the_id_string_): list@org-name.vdc-name.vapp-name
// Note: the separator can be changed using Provider.import_separator or variable VCD_IMPORT_SEPARATOR
func vappFirewallRulesImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	return vappNetworkRuleImport(d, meta, "vcd_vapp_firewall_rules")
}
func vappNetworkRuleImport(d *schema.ResourceData, meta interface{}, resourceType string) ([]*schema.ResourceData, error) {
	var commandOrgName, orgName, vdcName, vappName string
	resourceURI := strings.Split(d.Id(), ImportSeparator)

	log.Printf("[DEBUG] importing %s resource with provided id %s", resourceType, d.Id())

	if len(resourceURI) != 4 && len(resourceURI) != 3 {
		return nil, errHelpVappNetworkRulesImport
	}
	if strings.Contains(d.Id(), "list@") {
		commandOrgName, vdcName, vappName = resourceURI[0], resourceURI[1], resourceURI[2]
		commandOrgNameSplit := strings.Split(commandOrgName, "@")
		if len(commandOrgNameSplit) != 2 {
			return nil, errHelpVappNetworkRulesImport
		}
		orgName = commandOrgNameSplit[1]
		return listVappNetworksForImport(meta, orgName, vdcName, vappName)
	} else {
		orgName, vdcName, vappId, networkId := resourceURI[0], resourceURI[1], resourceURI[2], resourceURI[3]
		return getNetworkRules(d, meta, orgName, vdcName, vappId, networkId)
	}

}

func getNetworkRules(d *schema.ResourceData, meta interface{}, orgName, vdcName, vappId, networkId string) ([]*schema.ResourceData, error) {
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

func listVappNetworksForImport(meta interface{}, orgName, vdcName, vappId string) ([]*schema.ResourceData, error) {

	vcdClient := meta.(*VCDClient)
	_, vdc, err := vcdClient.GetOrgAndVdc(orgName, vdcName)
	if err != nil {
		return nil, fmt.Errorf("[vapp network rules import, network list] unable to find VDC %s: %s ", vdcName, err)
	}

	stdout := getTerraformStdout()
	_, _ = fmt.Fprintln(stdout, "Retrieving all vApp networks by name")
	vapp, err := vdc.GetVAppByNameOrId(vappId, false)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve vApp by name: %s", err)
	}

	writer := tabwriter.NewWriter(stdout, 0, 8, 1, '\t', tabwriter.AlignRight)

	fmt.Fprintln(writer, "No\tvApp ID\tID\tName\t")
	fmt.Fprintln(writer, "--\t-------\t--\t----\t")

	for index, vappNetwork := range vapp.VApp.NetworkConfigSection.NetworkConfig {
		uuid, err := govcd.GetUuidFromHref(vappNetwork.Link.HREF, false)
		if err != nil {
			return nil, fmt.Errorf("unable to parse vApp network ID: %s, %s", err, uuid)
		}

		fmt.Fprintf(writer, "%d\t%s\t%s\t%s\n", (index + 1), vapp.VApp.ID, uuid, vappNetwork.NetworkName)
	}
	err = writer.Flush()
	if err != nil {
		return nil, fmt.Errorf("unable to write to stdout: %s", err)
	}

	return nil, fmt.Errorf("resource was not imported! %s", errHelpVappNetworkRulesImport)
}
