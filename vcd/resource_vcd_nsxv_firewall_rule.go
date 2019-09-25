package vcd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/davecgh/go-spew/spew"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func resourceVcdNsxvFirewall() *schema.Resource {
	return &schema.Resource{
		Create: resourceVcdNsxvFirewallCreate,
		Read:   resourceVcdNsxvFirewallRead,
		Update: resourceVcdNsxvFirewallUpdate,
		Delete: resourceVcdNsxvFirewallDelete,
		Importer: &schema.ResourceImporter{
			State: resourceVcdNsxvFirewallImport,
		},

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
			"edge_gateway": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Edge gateway name in which Firewall Rule is located",
			},
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Firewall rule name",
			},
			"rule_type": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    false,
				Computed:    true,
				Description: "Read only. Possible values 'user', 'internal_high'",
			},
			"rule_tag": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Computed:    true,
				Description: "Optional. Allows to set custom rule tag",
			},
			"description": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Firewall rule description",
			},
			"action": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "accept",
				Description:  "'accept' or 'deny'. Default 'accept'",
				ValidateFunc: validation.StringInSlice([]string{"accept", "deny"}, false),
			},
			"enabled": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				ForceNew:    false,
				Default:     true,
				Description: "Whether the rule should be enabled. Default 'true'",
			},
			"logging_enabled": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				ForceNew:    false,
				Default:     false,
				Description: "Whether logging should be enabled for this rule. Default 'false'",
			},
			"source": {
				MinItems: 1,
				MaxItems: 1,
				Required: true,
				Type:     schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"exclude": {
							Optional: true,
							Type:     schema.TypeBool,
							Default:  false,
							Description: "Rule is applied to traffic coming from all sources " +
								"except for the source you excluded. Default false",
						},
						"ips": {
							Optional:    true,
							Type:        schema.TypeSet,
							Description: "IP address, CIDR, an IP range, or the keyword any",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"network_ids": {
							Optional:    true,
							Type:        schema.TypeSet,
							Description: "string",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},

			"destination": {
				MinItems: 1,
				MaxItems: 1,
				Required: true,
				Type:     schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"exclude": {
							Optional: true,
							Type:     schema.TypeBool,
							Default:  false,
							Description: "Rule is applied to traffic coming from all sources " +
								"except for the source you excluded. Default false",
						},
						"ips": {
							Optional:    true,
							Type:        schema.TypeSet,
							Description: "IP address, CIDR, an IP range, or the keyword any",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"network_ids": {
							Optional:    true,
							Type:        schema.TypeSet,
							Description: "string",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},

			"service": {
				Required: true,
				MinItems: 1,
				MaxItems: 1,
				Type:     schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"protocol": {
							Optional: true,
							Type:     schema.TypeString,
						},
						"port": {
							Optional: true,
							Type:     schema.TypeString,
						},
						"source_port": {
							Optional: true,
							Type:     schema.TypeString,
						},
					},
				},
			},
		},
	}
}

// resourceVcdNsxvFirewallCreate
func resourceVcdNsxvFirewallCreate(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)
	vcdClient.lockParentEdgeGtw(d)
	defer vcdClient.unLockParentEdgeGtw(d)

	edgeGateway, err := vcdClient.GetEdgeGatewayFromResource(d, "edge_gateway")
	if err != nil {
		return fmt.Errorf(errorUnableToFindEdgeGateway, err)
	}

	firewallRule, err := getFirewallRule(d, edgeGateway)
	if err != nil {
		return fmt.Errorf("unable to make firewall rule query: %s", err)
	}

	createdFirewallRule, err := edgeGateway.CreateNsxvFirewall(firewallRule)
	if err != nil {
		return fmt.Errorf("error creating new firewall rule: %s", err)
	}

	d.SetId(createdFirewallRule.ID)
	return resourceVcdNsxvFirewallRead(d, meta)
}

// resourceVcdNsxvFirewallUpdate
func resourceVcdNsxvFirewallUpdate(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)
	vcdClient.lockParentEdgeGtw(d)
	defer vcdClient.unLockParentEdgeGtw(d)

	edgeGateway, err := vcdClient.GetEdgeGatewayFromResource(d, "edge_gateway")
	if err != nil {
		return fmt.Errorf(errorUnableToFindEdgeGateway, err)
	}

	updateFirewallRule, err := getFirewallRule(d, edgeGateway)
	updateFirewallRule.ID = d.Id() // We already know an ID for update and it allows to change name

	if err != nil {
		return fmt.Errorf("could not create firewall rule type for update: %s", err)
	}

	_, err = edgeGateway.UpdateNsxvFirewall(updateFirewallRule)
	if err != nil {
		return fmt.Errorf("unable to update firewall rule with ID %s: %s", d.Id(), err)
	}

	return resourceVcdNsxvFirewallRead(d, meta)
}

// resourceVcdNsxvFirewallRead
func resourceVcdNsxvFirewallRead(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)

	edgeGateway, err := vcdClient.GetEdgeGatewayFromResource(d, "edge_gateway")
	if err != nil {
		return fmt.Errorf(errorUnableToFindEdgeGateway, err)
	}

	readFirewallRule, err := edgeGateway.GetNsxvFirewallById(d.Id())
	if err != nil {
		d.SetId("")
		return fmt.Errorf("unable to find firewall rule with ID %s: %s", d.Id(), err)
	}

	return setFirewallRuleData(d, readFirewallRule, edgeGateway)
}

func resourceVcdNsxvFirewallDelete(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)
	vcdClient.lockParentEdgeGtw(d)
	defer vcdClient.unLockParentEdgeGtw(d)

	edgeGateway, err := vcdClient.GetEdgeGatewayFromResource(d, "edge_gateway")
	if err != nil {
		return fmt.Errorf(errorUnableToFindEdgeGateway, err)
	}

	err = edgeGateway.DeleteNsxvFirewallById(d.Id())
	if err != nil {
		return fmt.Errorf("error deleting firewall rule with id %s: %s", d.Id(), err)
	}

	d.SetId("")
	return nil
}

// resourceVcdNsxvFirewallImport  is responsible for importing the resource.
// The following steps happen as part of import
// 1. The user supplies `terraform import _resource_name_ _the_id_string_` command
// 2. `_the_id_string_` contains a dot formatted path to resource as in the example below
// 3. The functions splits the dot-formatted path and tries to lookup the object
// 4. If the lookup succeeds it set's the ID field for `_resource_name_` resource in statefile
// (the resource must be already defined in .tf config otherwise `terraform import` will complain)
// 5. `terraform refresh` is being implicitly launched. The Read method looks up all other fields
// based on the known ID of object.
//
// Example resource name (_resource_name_): vcd_lb_nsxv_firewall_rule.my-test-fw-rule
// Example import path (_the_id_string_): org.vdc.edge-gw.existing-firewall-rule-id
func resourceVcdNsxvFirewallImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	resourceURI := strings.Split(d.Id(), ".")
	if len(resourceURI) != 4 {
		return nil, fmt.Errorf("resource name must be specified as org.vdc.edge-gw.firewall-rule-id")
	}
	orgName, vdcName, edgeName, firewallRuleId := resourceURI[0], resourceURI[1], resourceURI[2], resourceURI[3]

	vcdClient := meta.(*VCDClient)
	edgeGateway, err := vcdClient.GetEdgeGateway(orgName, vdcName, edgeName)
	if err != nil {
		return nil, fmt.Errorf(errorUnableToFindEdgeGateway, err)
	}

	readFirewallRule, err := edgeGateway.GetNsxvFirewallById(firewallRuleId)
	if err != nil {
		return []*schema.ResourceData{}, fmt.Errorf("unable to find firewall rule with id %s: %s",
			d.Id(), err)
	}

	d.Set("org", orgName)
	d.Set("vdc", vdcName)
	d.Set("edge_gateway", edgeName)

	d.SetId(readFirewallRule.ID)
	return []*schema.ResourceData{d}, nil
}

func setFirewallRuleData(d *schema.ResourceData, rule *types.EdgeFirewallRule, edge govcd.EdgeGateway) error {
	_ = d.Set("name", rule.Name)
	_ = d.Set("description", rule.Description)
	_ = d.Set("enabled", rule.Enabled)
	_ = d.Set("logging_enabled", rule.LoggingEnabled)
	_ = d.Set("action", rule.Action)
	_ = d.Set("rule_tag", rule.RuleTag)
	_ = d.Set("rule_type", rule.RuleTag)

	sourceIpsSlice := convertToTypeSet(rule.Source.IpAddress)
	sourceIpsSet := schema.NewSet(schema.HashSchema(&schema.Schema{Type: schema.TypeString}), sourceIpsSlice)

	vnicGroupIdStrings, err := groupIdStringsToNetworkNames(rule.Source.VnicGroupId, edge)
	if err != nil {
		return err
	}

	sourceNetworksSlice := convertToTypeSet(vnicGroupIdStrings)
	sourceNetworksSet := schema.NewSet(schema.HashSchema(&schema.Schema{Type: schema.TypeString}), sourceNetworksSlice)

	source := make([]interface{}, 1)
	sourceMap := make(map[string]interface{})
	sourceMap["exclude"] = rule.Source.Exclude
	sourceMap["ips"] = sourceIpsSet
	sourceMap["network_ids"] = sourceNetworksSet

	source[0] = sourceMap
	fmt.Println("===to schema===")
	spew.Dump(source)
	_ = d.Set("source", source)

	return nil
}

func getFirewallRule(d *schema.ResourceData, edge govcd.EdgeGateway) (*types.EdgeFirewallRule, error) {
	fmt.Println("===from schema===")
	spew.Dump(d.Get("source"))
	service := d.Get("service").([]interface{})
	if len(service) != 1 {
		return nil, fmt.Errorf("no service specified")
	}
	serviceMap := convertToStringMap(service[0].(map[string]interface{}))

	source := d.Get("source").([]interface{})
	if len(source) != 1 {
		return nil, fmt.Errorf("no source specified")
	}
	sourceMap := source[0].(map[string]interface{})
	sourceExclude := sourceMap["exclude"].(bool)
	sourceIps := sourceMap["ips"].(*schema.Set)
	sourceNetIds := sourceMap["network_ids"].(*schema.Set)
	sourceIpList := sourceIps.List()
	sourceNetIdList := sourceNetIds.List()
	sourceIpStrings := convertToSliceOfStrings(sourceIpList)
	sourceNetIdStrings := convertToSliceOfStrings(sourceNetIdList)

	destination := d.Get("destination").([]interface{})
	if len(destination) != 1 {
		return nil, fmt.Errorf("no destination specified")
	}
	destinationMap := destination[0].(map[string]interface{})
	destinationExclude := destinationMap["exclude"].(bool)
	destinationIps := destinationMap["ips"].(*schema.Set)
	destinationNetIds := destinationMap["network_ids"].(*schema.Set)
	destinationIpList := destinationIps.List()
	destinationNetIdList := destinationNetIds.List()
	destinationIpStrings := convertToSliceOfStrings(destinationIpList)
	destinationNetIdStrings := convertToSliceOfStrings(destinationNetIdList)

	firewallRule := &types.EdgeFirewallRule{
		Name:           d.Get("name").(string),
		Enabled:        d.Get("enabled").(bool),
		LoggingEnabled: d.Get("logging_enabled").(bool),
		Action:         d.Get("action").(string),
		Description:    d.Get("description").(string),
		RuleTag:        d.Get("rule_tag").(string),
		Application: types.EdgeFirewallApplication{
			Service: types.EdgeFirewallApplicationService{
				Protocol:   serviceMap["protocol"],
				Port:       serviceMap["port"],
				SourcePort: serviceMap["source_port"],
			},
		},
		Source: types.EdgeFirewallObject{
			Exclude:     sourceExclude,
			IpAddress:   sourceIpStrings,
			VnicGroupId: sourceNetIdStrings,
		},
		Destination: types.EdgeFirewallObject{
			Exclude:     destinationExclude,
			IpAddress:   destinationIpStrings,
			VnicGroupId: destinationNetIdStrings,
		},
	}

	return firewallRule, nil
}

func stringInSlice(str string, list []string) bool {
	for _, v := range list {
		if v == str {
			return true
		}
	}
	return false
}

// groupIdStringsToNetworkNames iterates over
func groupIdStringsToNetworkNames(groupIdStrings []string, edge govcd.EdgeGateway) ([]string, error) {
	vnicGroupIdStrings := make([]string, len(groupIdStrings))
	for index, value := range groupIdStrings {
		// A list of accepted parameters as strings (not real network names). No need to look them
		// up. Passing these names as they are directly to statefile
		if stringInSlice(value, []string{"internal", "external", "vse"}) {
			vnicGroupIdStrings[index] = value
			continue
		}

		vNicNameSplit := strings.Split(value, "-") // extract index from format 'vnic-10'
		if len(vNicNameSplit) < 2 {
			return []string{}, fmt.Errorf("could not find vNic index from value: %s", value)
		}

		vNicIndex, err := strconv.Atoi(vNicNameSplit[1])
		if err != nil {
			return []string{}, fmt.Errorf("could not convert edge gateway NIC index to int: %s: %s",
				vNicNameSplit[1], err)
		}

		networkName, _, err := edge.GetNetworkNameAndTypeByVnicIndex(vNicIndex)
		if err != nil {
			return []string{}, fmt.Errorf("could not find network name by vNic index %d: %s", vNicIndex, err)
		}
		vnicGroupIdStrings[index] = networkName
	}
	return vnicGroupIdStrings, nil
}
