package vcd

import (
	"fmt"
	"log"
	"strconv"
	"strings"

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
			"insert_above_rule_id": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "Optional. Allows to insert the firewall rule above some other rule",
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
								"except for the excluded source. Default 'false'",
						},
						"ip_addresses": {
							Optional:    true,
							Type:        schema.TypeSet,
							Description: "IP address, CIDR, an IP range, or the keyword 'any'",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"gateway_interfaces": {
							Optional:    true,
							Type:        schema.TypeSet,
							Description: "'vse', 'internal', 'external' or network name",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"virtual_machines_ids": {
							Optional:    true,
							Type:        schema.TypeSet,
							Description: "Set of VM IDs",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"org_network_ids": {
							Optional:    true,
							Type:        schema.TypeSet,
							Description: "Set of org network IDs",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"ipset_ids": {
							Optional:    true,
							Type:        schema.TypeSet,
							Description: "Set of org network IDs",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"security_group_ids": {
							Optional:    true,
							Type:        schema.TypeSet,
							Description: "Set of org network IDs",
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
							Description: "Rule is applied to traffic going to any destinations " +
								"except for the excluded destination. Default 'false'",
						},
						"ip_addresses": {
							Optional:    true,
							Type:        schema.TypeSet,
							Description: "IP address, CIDR, an IP range, or the keyword 'any'",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"gateway_interfaces": {
							Optional:    true,
							Type:        schema.TypeSet,
							Description: "'vse', 'internal', 'external' or network name",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"virtual_machines_ids": {
							Optional:    true,
							Type:        schema.TypeSet,
							Description: "Set of VM IDs",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"org_network_ids": {
							Optional:    true,
							Type:        schema.TypeSet,
							Description: "Set of org network IDs",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"ipset_ids": {
							Optional:    true,
							Type:        schema.TypeSet,
							Description: "Set of org network IDs",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"security_group_ids": {
							Optional:    true,
							Type:        schema.TypeSet,
							Description: "Set of org network IDs",
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

// setFirewallRuleData is the main function used for setting Terraform schema
func setFirewallRuleData(d *schema.ResourceData, rule *types.EdgeFirewallRule, edge govcd.EdgeGateway) error {
	_ = d.Set("name", rule.Name)
	_ = d.Set("description", rule.Description)
	_ = d.Set("enabled", rule.Enabled)
	_ = d.Set("logging_enabled", rule.LoggingEnabled)
	_ = d.Set("action", rule.Action)
	_ = d.Set("rule_tag", rule.RuleTag)
	_ = d.Set("rule_type", rule.RuleTag)

	// Process and set "source" block
	source, err := getEndpointData(rule.Source, edge)
	if err != nil {
		return fmt.Errorf("could not prepare data for setting 'source' block: %s", err)
	}
	err = d.Set("source", source)
	if err != nil {
		return fmt.Errorf("could not set 'source' block: %s", err)
	}

	// Process and set "destination" block
	destination, err := getEndpointData(rule.Destination, edge)
	if err != nil {
		return fmt.Errorf("could not prepare data for setting 'destination' block: %s", err)
	}
	err = d.Set("destination", destination)
	if err != nil {
		return fmt.Errorf("could not set 'destination' block: %s", err)
	}

	// Process and set "service" blocks

	return nil
}

// getFirewallRule is the main function  used for creating *types.EdgeFirewallRule structure from
// Terraform schema configuration
func getFirewallRule(d *schema.ResourceData, edge govcd.EdgeGateway) (*types.EdgeFirewallRule, error) {
	service := d.Get("service").([]interface{})
	if len(service) != 1 {
		return nil, fmt.Errorf("no service specified")
	}
	serviceMap := convertToStringMap(service[0].(map[string]interface{}))

	sourceEndpoint, err := getFirewallRuleEndpoint(d.Get("source").([]interface{}), edge)
	if err != nil {
		return nil, fmt.Errorf("could not convert 'source' block to API request: %s", err)
	}

	destinationEndpoint, err := getFirewallRuleEndpoint(d.Get("destination").([]interface{}), edge)
	if err != nil {
		return nil, fmt.Errorf("could not convert 'destination' block to API request: %s", err)
	}

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
		Source:      *sourceEndpoint,
		Destination: *destinationEndpoint,
	}

	return firewallRule, nil
}

// getEndpointData formats the nested set structure suitable for d.Set() for
// 'source' and 'destination' blocks in firewall rule
func getEndpointData(endpoint types.EdgeFirewallEndpoint, edge govcd.EdgeGateway) ([]interface{}, error) {
	// Different object types are in the same grouping object tag <groupingObjectId>
	// They can be distinguished by 3rd element in ID
	var (
		endpointNetworks       []string
		endpointVMs            []string
		endpointIpSets         []string
		endpointSecurityGroups []string
	)

	for _, groupingObject := range endpoint.GroupingObjectId {
		switch strings.Split(groupingObject, ":")[2] {
		// Handle org vdc networks
		// Sample ID: urn:vcloud:network:95bffe8e-7e67-452d-abf2-535ac298db2b
		case "network":
			endpointNetworks = append(endpointNetworks, groupingObject)

		// Handle virtual machines
		// Sample ID: urn:vcloud:vm:c0c5a316-fb2d-4f33-a814-3e0fba714c74
		case "vm":
			endpointVMs = append(endpointVMs, groupingObject)

		// Handle ipsets
		// Sample ID: f9daf2da-b4f9-4921-a2f4-d77a943a381c:ipset-2
		case "ipset":
			endpointIpSets = append(endpointIpSets, groupingObject)

		// Handle security groups
		// Sample ID:
		case "security-group":
			endpointSecurityGroups = append(endpointSecurityGroups, groupingObject)

		// Log the group ID if it was not one of above
		default:
			log.Printf("[WARN] Unrecognized grouping object ID: %s", groupingObject)
		}
	}

	// Convert org vdc networks to set
	endpointNetworksSlice := convertToTypeSet(endpointNetworks)
	endpointNetworksSet := schema.NewSet(schema.HashSchema(&schema.Schema{Type: schema.TypeString}), endpointNetworksSlice)

	// Convert virtual machines to set
	endpointVmSlice := convertToTypeSet(endpointVMs)
	endpointVmSet := schema.NewSet(schema.HashSchema(&schema.Schema{Type: schema.TypeString}), endpointVmSlice)

	// Convert ipsets to set
	endpointIpSetSlice := convertToTypeSet(endpointIpSets)
	endpointIpSetSet := schema.NewSet(schema.HashSchema(&schema.Schema{Type: schema.TypeString}), endpointIpSetSlice)

	// Convert security groups to set
	endpointSecurityGroupSlice := convertToTypeSet(endpointSecurityGroups)
	endpointSecurityGroupSet := schema.NewSet(schema.HashSchema(&schema.Schema{Type: schema.TypeString}), endpointSecurityGroupSlice)

	// Convert `ip_addresses` to set
	endpointIpsSlice := convertToTypeSet(endpoint.IpAddress)
	endpointIpsSet := schema.NewSet(schema.HashSchema(&schema.Schema{Type: schema.TypeString}), endpointIpsSlice)

	// Convert `gateway_interfaces` to set
	vnicGroupIdStrings, err := groupIdStringsToNetworkNames(endpoint.VnicGroupId, edge)
	if err != nil {
		return nil, err
	}
	endpointGatewayInterfaceSlice := convertToTypeSet(vnicGroupIdStrings)
	endpointGatewayInterfaceSet := schema.NewSet(schema.HashSchema(&schema.Schema{Type: schema.TypeString}), endpointGatewayInterfaceSlice)

	// Insert all sets into single element block ready to be ('source' or 'destination')
	endpointSlice := make([]interface{}, 1)
	endpointMap := make(map[string]interface{})
	endpointMap["exclude"] = endpoint.Exclude
	endpointMap["ip_addresses"] = endpointIpsSet
	endpointMap["gateway_interfaces"] = endpointGatewayInterfaceSet
	endpointMap["org_network_ids"] = endpointNetworksSet
	endpointMap["virtual_machine_ids"] = endpointVmSet
	endpointMap["security_group_ids"] = endpointSecurityGroupSet
	endpointMap["ipset_ids"] = endpointIpSetSet

	endpointSlice[0] = endpointMap

	return endpointSlice, nil
}

// getFirewallRuleEndpoint processes Terraform schema and converts it to *types.EdgeFirewallEndpoint
// which is useful for 'source' or 'destination' blocks
func getFirewallRuleEndpoint(endpoint []interface{}, edge govcd.EdgeGateway) (*types.EdgeFirewallEndpoint, error) {
	if len(endpoint) != 1 {
		return nil, fmt.Errorf("no source specified")
	}

	// Create empty endpoint structure for populating
	result := &types.EdgeFirewallEndpoint{}

	// Extract 'exclude' field from structure
	endpointMap := endpoint[0].(map[string]interface{})
	endpointExclude := endpointMap["exclude"].(bool)
	result.Exclude = endpointExclude

	// Extract ips and add them to endpoint structure
	endpointIpStrings := convertSchemaSetToSliceOfStrings(endpointMap["ip_addresses"].(*schema.Set))
	result.IpAddress = endpointIpStrings

	// Extract 'gateway_interfaces' names, convert them to vNic indexes and add to the structure
	endpointEdgeInterfaceIdStrings := convertSchemaSetToSliceOfStrings(endpointMap["gateway_interfaces"].(*schema.Set))
	endpointEdgeInterfaceVnicList, err := groupNetworkNamesToIdStrings(endpointEdgeInterfaceIdStrings, edge)
	if err != nil {
		return nil, fmt.Errorf("could not lookup vNic indexes for networks: %s", err)
	}
	result.VnicGroupId = endpointEdgeInterfaceVnicList

	// 'types.EdgeFirewallEndpoint.GroupingObjectId' holds IDs for VMs, org networks, ipsets and Security groups

	// Extract VM IDs from set and add them to endpoint structure
	endpointVmStrings := convertSchemaSetToSliceOfStrings(endpointMap["virtual_machines_ids"].(*schema.Set))
	result.GroupingObjectId = append(result.GroupingObjectId, endpointVmStrings...)

	// Extract org network IDs from set and add them to endpoint structure
	endpointOrgNetworkStrings := convertSchemaSetToSliceOfStrings(endpointMap["org_network_ids"].(*schema.Set))
	result.GroupingObjectId = append(result.GroupingObjectId, endpointOrgNetworkStrings...)

	// Extract ipset IDs from set and add them to endpoint structure
	endpointIpSetStrings := convertSchemaSetToSliceOfStrings(endpointMap["ipset_ids"].(*schema.Set))
	result.GroupingObjectId = append(result.GroupingObjectId, endpointIpSetStrings...)

	// Extract security group IDs from set and add them to endpoint structure
	endpointSecurityGroupStrings := convertSchemaSetToSliceOfStrings(endpointMap["security_group_ids"].(*schema.Set))
	result.GroupingObjectId = append(result.GroupingObjectId, endpointSecurityGroupStrings...)

	return result, nil
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

// groupNetworkNamesToIdStrings iterates over network names and returns vNic ID name list
// (suitable for firewall creation)
func groupNetworkNamesToIdStrings(groupNetworkNames []string, edge govcd.EdgeGateway) ([]string, error) {
	idStrings := make([]string, len(groupNetworkNames))
	for index, networkName := range groupNetworkNames {
		// A list of accepted parameters as strings (not real network names). No need to look them
		// up. Passing these names as they are directly to statefile
		if stringInSlice(networkName, []string{"internal", "external", "vse"}) {
			idStrings[index] = networkName
			continue
		}

		// TODO improve GetVnicIndexByNetworkNameAndType and ensure that only one network with the
		// the same name can be defined in edge gateway
		vNicIndex, err := edge.GetVnicIndexByNetworkNameAndType(networkName, "subinterface")
		if govcd.IsNotFound(err) {
			vNicIndex, err = edge.GetVnicIndexByNetworkNameAndType(networkName, "internal")
			if govcd.IsNotFound(err) {
				vNicIndex, err = edge.GetVnicIndexByNetworkNameAndType(networkName, "uplink")
				if govcd.IsNotFound(err) {
					vNicIndex, err = edge.GetVnicIndexByNetworkNameAndType(networkName, "trunk")
					if govcd.IsNotFound(err) {
						return nil, fmt.Errorf("unable to find network %s interface on edge gateway", networkName)
					}
				}
			}
		}

		if err != nil {
			return nil, fmt.Errorf("error searching for network %s: %s", networkName, err)
		}

		// we found the network - add it to the list
		idStrings[index] = "vnic-" + strconv.Itoa(*vNicIndex)
	}

	return idStrings, nil
}

// stringInSlice checks if a string exists in slice of strings
func stringInSlice(str string, list []string) bool {
	for _, v := range list {
		if v == str {
			return true
		}
	}
	return false
}
