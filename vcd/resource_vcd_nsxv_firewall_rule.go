package vcd

import (
	"bytes"
	"fmt"
	"log"
	"strconv"
	"strings"

	"text/tabwriter"

	"github.com/hashicorp/terraform/helper/hashcode"
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
			"above_rule_id": &schema.Schema{
				Type:        schema.TypeString,
				ForceNew:    true,
				Optional:    true,
				Description: "This firewall rule will be inserted above the referred one",
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
						"virtual_machine_ids": {
							Optional:    true,
							Type:        schema.TypeSet,
							Description: "Set of VM IDs",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"org_networks": {
							Optional:    true,
							Type:        schema.TypeSet,
							Description: "Set of org network names",
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
						"virtual_machine_ids": {
							Optional:    true,
							Type:        schema.TypeSet,
							Description: "Set of VM IDs",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"org_networks": {
							Optional:    true,
							Type:        schema.TypeSet,
							Description: "Set of org network names",
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
				Optional: true,
				MinItems: 1,
				Type:     schema.TypeSet,
				Set:      resourceVcdNsxvFirewallRuleServiceHash,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"protocol": {
							Required: true,
							Type:     schema.TypeString,
						},
						"port": {
							Optional:     true,
							Type:         schema.TypeString,
							ValidateFunc: validateCase("lower"),
						},
						"source_port": {
							Optional:     true,
							Type:         schema.TypeString,
							ValidateFunc: validateCase("lower"),
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

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return fmt.Errorf(errorRetrievingOrgAndVdc, err)
	}

	firewallRule, err := getFirewallRule(d, edgeGateway, vdc)
	if err != nil {
		return fmt.Errorf("unable to make firewall rule query: %s", err)
	}

	// Check if above_rule_id is set
	// aboveRuleId := ""
	// if id, ok := d.GetOk("above_rule_id"); ok {
	// 	aboveRuleId = id.(string)
	// }

	createdFirewallRule, err := edgeGateway.CreateNsxvFirewall(firewallRule, d.Get("above_rule_id").(string))
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

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return fmt.Errorf(errorRetrievingOrgAndVdc, err)
	}

	updateFirewallRule, err := getFirewallRule(d, edgeGateway, vdc)
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

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return fmt.Errorf(errorRetrievingOrgAndVdc, err)
	}

	readFirewallRule, err := edgeGateway.GetNsxvFirewallById(d.Id())
	if err != nil {
		d.SetId("")
		return fmt.Errorf("unable to find firewall rule with ID %s: %s", d.Id(), err)
	}

	return setFirewallRuleData(d, readFirewallRule, edgeGateway, vdc)
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
// 2a. If the `_the_id_string_` contains a dot formatted path to resource as in the example below
// it will try to import it. If it is found - the ID is set
// 2b. If the `_the_id_string_` starts with `list@` and contains path to edge gateway similar to
// `list@org.vdc.edge-gw` then the function lists all firewall rules and their IDs in that edge
// gateway.
// 2c. If the `_the_id_string_` does not match format described neither in '2a' nor in '2b' a
// usage error message is printed
//
// Example resource name (_resource_name_): vcd_lb_nsxv_firewall_rule.my-test-fw-rule
// Example import path (_the_id_string_): org.vdc.edge-gw.existing-firewall-rule-id
func resourceVcdNsxvFirewallImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	var commandOrgName, orgName, vdcName, edgeName, firewallRuleId string
	var listRules, importRule bool

	resourceURI := strings.Split(d.Id(), ".")
	switch len(resourceURI) {
	case 3:
		commandOrgName, vdcName, edgeName = resourceURI[0], resourceURI[1], resourceURI[2]
		commandOrgNameSplit := strings.Split(commandOrgName, "@")
		if len(commandOrgNameSplit) != 2 {
			return nil, fmt.Errorf("resource name must be specified as " +
				"'org.vdc.edge-gw.firewall-rule-id' or 'list@org.vdc.edge-gw' to get a list of rules")
		}
		orgName = commandOrgNameSplit[1]
		listRules = true
	case 4:
		orgName, vdcName, edgeName, firewallRuleId = resourceURI[0], resourceURI[1], resourceURI[2], resourceURI[3]
		importRule = true
	default:
		return nil, fmt.Errorf("resource name must be specified as " +
			"'org.vdc.edge-gw.firewall-rule-id' or 'list@org.vdc.edge-gw' to get a list of rules")
	}

	vcdClient := meta.(*VCDClient)
	edgeGateway, err := vcdClient.GetEdgeGateway(orgName, vdcName, edgeName)
	if err != nil {
		return nil, fmt.Errorf(errorUnableToFindEdgeGateway, err)
	}

	// If the user requested to print rules, try to fetch all of them and print in a user friendly
	// table with both UI and real firewall IDs
	if listRules {
		_, _ = fmt.Fprintln(getTerraformStdout(), "Retrieving all firewall rules")
		allRules, err := edgeGateway.GetAllNsxvFirewallRules()
		if err != nil {
			return nil, fmt.Errorf("unable to retrieve all firewal rules: %s", err)
		}

		writer := tabwriter.NewWriter(getTerraformStdout(), 0, 8, 1, '\t', tabwriter.AlignRight)

		fmt.Fprintln(writer, "UI ID\tID\tName\tAction\tType")
		fmt.Fprintln(writer, "-----\t--\t----\t------\t----")
		for index, rule := range allRules {
			fmt.Fprintf(writer, "%d\t%s\t%s\t%s\t%s\n", (index + 1), rule.ID, rule.Name, rule.Action, rule.RuleType)
		}
		writer.Flush()

		return nil, fmt.Errorf("Resource was not imported! Please use the above ID to format the command as: \n" +
			"terraform import vcd_nsxv_firewall.resource-name org.vdc.edge-gw.firewall-rule-id")
	}

	// Proceed with import
	if importRule {
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

	return nil, nil
}

// setFirewallRuleData is the main function used for setting Terraform schema
func setFirewallRuleData(d *schema.ResourceData, rule *types.EdgeFirewallRule, edge *govcd.EdgeGateway, vdc *govcd.Vdc) error {
	_ = d.Set("name", rule.Name)
	_ = d.Set("enabled", rule.Enabled)
	_ = d.Set("logging_enabled", rule.LoggingEnabled)
	_ = d.Set("action", rule.Action)
	_ = d.Set("rule_tag", rule.RuleTag)
	_ = d.Set("rule_type", rule.RuleType)

	// Process and set "source" block
	source, err := getEndpointData(rule.Source, edge, vdc)
	if err != nil {
		return fmt.Errorf("could not prepare data for setting 'source' block: %s", err)
	}
	err = d.Set("source", source)
	if err != nil {
		return fmt.Errorf("could not set 'source' block: %s", err)
	}

	// Process and set "destination" block
	destination, err := getEndpointData(rule.Destination, edge, vdc)
	if err != nil {
		return fmt.Errorf("could not prepare data for setting 'destination' block: %s", err)
	}
	err = d.Set("destination", destination)
	if err != nil {
		return fmt.Errorf("could not set 'destination' block: %s", err)
	}

	serviceSet, err := getServiceData(rule.Application, edge, vdc)
	// log.Printf("[DEBUG] found vNic index %d for network %s (type %s)", vNicIndex, networkName, networkType)
	// spew.Dump(serviceSet)
	if err != nil {
		return fmt.Errorf("could not prepare data for setting 'service' blocks: %s", err)
	}

	err = d.Set("service", serviceSet)
	if err != nil {
		return fmt.Errorf("could not set 'service' blocks: %s", err)
	}

	// Process and set "service" blocks

	return nil
}

// getFirewallServices extracts service definition from terraform schema and returns it
func getFirewallServices(serviceSet *schema.Set) ([]types.EdgeFirewallApplicationService, error) {
	serviceSlice := serviceSet.List()
	services := make([]types.EdgeFirewallApplicationService, len(serviceSlice))
	if len(services) > 0 {
		for index, service := range serviceSlice {
			serviceMap := convertToStringMap(service.(map[string]interface{}))
			oneService := types.EdgeFirewallApplicationService{
				Protocol:   serviceMap["protocol"],
				Port:       serviceMap["port"],
				SourcePort: serviceMap["source_port"],
			}
			services[index] = oneService
		}
	}
	return services, nil
}

// getFirewallRule is the main function  used for creating *types.EdgeFirewallRule structure from
// Terraform schema configuration
func getFirewallRule(d *schema.ResourceData, edge *govcd.EdgeGateway, vdc *govcd.Vdc) (*types.EdgeFirewallRule, error) {
	sourceEndpoint, err := getFirewallRuleEndpoint(d.Get("source").([]interface{}), edge, vdc)
	if err != nil {
		return nil, fmt.Errorf("could not convert 'source' block to API request: %s", err)
	}

	destinationEndpoint, err := getFirewallRuleEndpoint(d.Get("destination").([]interface{}), edge, vdc)
	if err != nil {
		return nil, fmt.Errorf("could not convert 'destination' block to API request: %s", err)
	}

	services, err := getFirewallServices(d.Get("service").(*schema.Set))
	if err != nil {
		return nil, fmt.Errorf("could not convert services blocks for API request: %s ", err)
	}

	firewallRule := &types.EdgeFirewallRule{
		Name:           d.Get("name").(string),
		Enabled:        d.Get("enabled").(bool),
		LoggingEnabled: d.Get("logging_enabled").(bool),
		Action:         d.Get("action").(string),
		RuleTag:        d.Get("rule_tag").(string),
		Application: types.EdgeFirewallApplication{
			Services: services,
		},
		Source:      *sourceEndpoint,
		Destination: *destinationEndpoint,
	}

	return firewallRule, nil
}

// getEndpointData formats nested set structure suitable for d.Set() for
// 'source' and 'destination' blocks in firewall rule
func getEndpointData(endpoint types.EdgeFirewallEndpoint, edge *govcd.EdgeGateway, vdc *govcd.Vdc) ([]interface{}, error) {
	// Different object types are in the same grouping object tag <groupingObjectId>
	// They can be distinguished by 3rd element in ID
	var (
		endpointNetworks       []string
		endpointVMs            []string
		endpointIpSets         []string
		endpointSecurityGroups []string
	)

	for _, groupingObject := range endpoint.GroupingObjectIds {
		idSplit := strings.Split(groupingObject, ":")
		idLen := len(idSplit)
		subIdSplit := ""
		if idLen == 2 {
			subSplit := strings.Split(idSplit[1], "-")
			if len(subSplit) == 2 {
				subIdSplit = subSplit[0]
			}
		}
		switch {
		// Handle org vdc networks
		// Sample ID: urn:vcloud:network:95bffe8e-7e67-452d-abf2-535ac298db2b
		case idLen == 4 && idSplit[2] == "network":
			endpointNetworks = append(endpointNetworks, groupingObject)

		// Handle virtual machines
		// Sample ID: urn:vcloud:vm:c0c5a316-fb2d-4f33-a814-3e0fba714c74
		case idLen == 4 && idSplit[2] == "vm":
			endpointVMs = append(endpointVMs, groupingObject)

		// Handle ipsets
		// Sample ID: f9daf2da-b4f9-4921-a2f4-d77a943a381c:ipset-2
		case idLen == 2 && subIdSplit == "ipset":
			endpointIpSets = append(endpointIpSets, groupingObject)

		// Handle security groups
		// Sample ID: f9daf2da-b4f9-4921-a2f4-d77a943a381c:securitygroup-11
		case idLen == 2 && subIdSplit == "securitygroup":
			endpointSecurityGroups = append(endpointSecurityGroups, groupingObject)

		// Log the group ID if it was not one of above
		default:
			log.Printf("[WARN] Unrecognized grouping object ID: %s", groupingObject)
		}
	}

	// Convert org vdc network IDs to org network names, then make a set of these network names
	endpointNetworkNames, err := orgNetworksIdsToNames(endpointNetworks, vdc)
	if err != nil {
		return nil, fmt.Errorf("could not convert org network names to IDs: %s", err)
	}
	endpointNetworksSlice := convertToTypeSet(endpointNetworkNames)
	endpointNetworksSet := schema.NewSet(schema.HashSchema(&schema.Schema{Type: schema.TypeString}), endpointNetworksSlice)

	// Convert virtual machine IDs to set
	endpointVmSlice := convertToTypeSet(endpointVMs)
	endpointVmSet := schema.NewSet(schema.HashSchema(&schema.Schema{Type: schema.TypeString}), endpointVmSlice)

	// Convert ipset IDs to set
	endpointIpSetSlice := convertToTypeSet(endpointIpSets)
	endpointIpSetSet := schema.NewSet(schema.HashSchema(&schema.Schema{Type: schema.TypeString}), endpointIpSetSlice)

	// Convert security group IDs to set
	endpointSecurityGroupSlice := convertToTypeSet(endpointSecurityGroups)
	endpointSecurityGroupSet := schema.NewSet(schema.HashSchema(&schema.Schema{Type: schema.TypeString}), endpointSecurityGroupSlice)

	// Convert `ip_addresses` to set
	endpointIpsSlice := convertToTypeSet(endpoint.IpAddresses)
	endpointIpsSet := schema.NewSet(schema.HashSchema(&schema.Schema{Type: schema.TypeString}), endpointIpsSlice)

	// Convert `gateway_interfaces` vNic IDs to network names as the UI does it so
	vnicGroupIdStrings, err := edgeVnicIdStringsToNetworkNames(endpoint.VnicGroupIds, edge)
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
	endpointMap["org_networks"] = endpointNetworksSet
	endpointMap["virtual_machine_ids"] = endpointVmSet
	endpointMap["security_group_ids"] = endpointSecurityGroupSet
	endpointMap["ipset_ids"] = endpointIpSetSet

	endpointSlice[0] = endpointMap

	return endpointSlice, nil
}

// getServiceData formats nested set structure suitable for d.Set() for services blocks
func getServiceData(firewallApplication types.EdgeFirewallApplication, edge *govcd.EdgeGateway, vdc *govcd.Vdc) (*schema.Set, error) {
	serviceSlice := make([]interface{}, len(firewallApplication.Services))

	for index, service := range firewallApplication.Services {
		serviceMap := make(map[string]interface{})
		serviceMap["protocol"] = service.Protocol
		serviceMap["port"] = service.Port
		serviceMap["source_port"] = service.SourcePort

		serviceSlice[index] = serviceMap
	}

	serviceSet := schema.NewSet(resourceVcdNsxvFirewallRuleServiceHash, serviceSlice)

	return serviceSet, nil
}

// getFirewallRuleEndpoint processes Terraform schema and converts it to *types.EdgeFirewallEndpoint
// which is useful for 'source' or 'destination' blocks
func getFirewallRuleEndpoint(endpoint []interface{}, edge *govcd.EdgeGateway, vdc *govcd.Vdc) (*types.EdgeFirewallEndpoint, error) {
	if len(endpoint) != 1 {
		return nil, fmt.Errorf("no source specified")
	}

	result := &types.EdgeFirewallEndpoint{}

	// Extract 'exclude' field from structure
	endpointMap := endpoint[0].(map[string]interface{})
	endpointExclude := endpointMap["exclude"].(bool)
	result.Exclude = endpointExclude

	// Extract ips and add them to endpoint structure
	endpointIpStrings := convertSchemaSetToSliceOfStrings(endpointMap["ip_addresses"].(*schema.Set))
	result.IpAddresses = endpointIpStrings

	// Extract 'gateway_interfaces' names, convert them to vNic indexes and add to the structure
	endpointEdgeInterfaceIdStrings := convertSchemaSetToSliceOfStrings(endpointMap["gateway_interfaces"].(*schema.Set))
	endpointEdgeInterfaceVnicList, err := edgeInterfaceNamesToIdStrings(endpointEdgeInterfaceIdStrings, edge)
	if err != nil {
		return nil, fmt.Errorf("could not lookup vNic indexes for networks: %s", err)
	}
	result.VnicGroupIds = endpointEdgeInterfaceVnicList

	// 'types.EdgeFirewallEndpoint.GroupingObjectId' holds IDs for VMs, org networks, ipsets and Security groups

	// Extract VM IDs from set and add them to endpoint structure
	endpointVmStrings := convertSchemaSetToSliceOfStrings(endpointMap["virtual_machine_ids"].(*schema.Set))
	result.GroupingObjectIds = append(result.GroupingObjectIds, endpointVmStrings...)

	// Extract org network names from set, lookup their IDs and add them to endpoint structure
	endpointOrgNetworkNameStrings := convertSchemaSetToSliceOfStrings(endpointMap["org_networks"].(*schema.Set))
	endpointOrgNetworkIdStrings, err := orgNetworkNamesToIds(endpointOrgNetworkNameStrings, vdc)
	if err != nil {
		return nil, fmt.Errorf("could not lookup network IDs for networks: %s", err)
	}
	result.GroupingObjectIds = append(result.GroupingObjectIds, endpointOrgNetworkIdStrings...)

	// Extract ipset IDs from set and add them to endpoint structure
	endpointIpSetStrings := convertSchemaSetToSliceOfStrings(endpointMap["ipset_ids"].(*schema.Set))
	result.GroupingObjectIds = append(result.GroupingObjectIds, endpointIpSetStrings...)

	// Extract security group IDs from set and add them to endpoint structure
	endpointSecurityGroupStrings := convertSchemaSetToSliceOfStrings(endpointMap["security_group_ids"].(*schema.Set))
	result.GroupingObjectIds = append(result.GroupingObjectIds, endpointSecurityGroupStrings...)

	return result, nil
}

// edgeVnicIdStringsToNetworkNames iterates over vnic IDs in format `vnic-10`, `vnic-x` and converts
// them to network names.
// It passes through 3 network types "internal", "external", "vse" as they are because the API
// accepts such notation.
func edgeVnicIdStringsToNetworkNames(groupIdStrings []string, edge *govcd.EdgeGateway) ([]string, error) {
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

// edgeInterfaceNamesToIdStrings iterates over network names and returns vNic ID name list
// Format: vnic-10, vnic-3, etc. (suitable for firewall creation)
func edgeInterfaceNamesToIdStrings(groupNetworkNames []string, edge *govcd.EdgeGateway) ([]string, error) {
	idStrings := make([]string, len(groupNetworkNames))
	for index, networkName := range groupNetworkNames {
		// A list of accepted parameters as strings (not real network names). No need to look them
		// up. Passing these names as they are directly to statefile because the API accepts them.
		if stringInSlice(networkName, []string{"internal", "external", "vse"}) {
			idStrings[index] = networkName
			continue
		}

		vNicIndex, networkType, err := edge.GetAnyVnicIndexByNetworkName(networkName)
		if err != nil {
			return nil, fmt.Errorf("error searching for network %s: %s", networkName, err)
		}
		// we found the network - add it to the list
		log.Printf("[DEBUG] found vNic index %d for network %s (type %s)", vNicIndex, networkName, networkType)
		idStrings[index] = "vnic-" + strconv.Itoa(*vNicIndex)
	}
	return idStrings, nil
}

// orgNetworkNamesToIds looks up org network ids by their  names.
// Returned ID format: urn:vcloud:network:95bffe8e-7e67-452d-abf2-535ac298db2b
func orgNetworkNamesToIds(networkNames []string, vdc *govcd.Vdc) ([]string, error) {
	orgNetworkIds := make([]string, len(networkNames))
	for index, networkName := range networkNames {
		orgVdcNetwork, err := vdc.GetOrgVdcNetworkByName(networkName, false)
		if err != nil {
			return nil, fmt.Errorf("could not find org network with name %s: %s", networkName, err)
		}
		orgNetworkIds[index] = orgVdcNetwork.OrgVDCNetwork.ID

	}
	return orgNetworkIds, nil
}

// orgNetworksIdsToNames looks up network name by ID
func orgNetworksIdsToNames(networkIds []string, vdc *govcd.Vdc) ([]string, error) {
	orgNetworkNames := make([]string, len(networkIds))
	for index, networkId := range networkIds {
		orgVdcNetwork, err := vdc.GetOrgVdcNetworkById(networkId, false)
		if err != nil {
			return nil, fmt.Errorf("could not find org network with name %s: %s", networkId, err)
		}
		orgNetworkNames[index] = orgVdcNetwork.OrgVDCNetwork.Name

	}
	return orgNetworkNames, nil
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

// resourceVcdNsxvFirewallRuleServiceHash generates a hash for service TypeSet. Its main purpose is to
// avoid hash changes when port or source_port ar left empty or set as 'any'. Having empty port and
// source_port is the same as having "any".
// protocol, port, source_port
func resourceVcdNsxvFirewallRuleServiceHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})

	protocol := strings.ToLower(m["protocol"].(string))
	port := strings.ToLower(m["port"].(string))
	sourcePort := strings.ToLower(m["source_port"].(string))

	if port == "" {
		port = "any"
	}

	if sourcePort == "" {
		sourcePort = "any"
	}

	buf.WriteString(fmt.Sprintf("%s-", protocol))
	buf.WriteString(fmt.Sprintf("%s-", port))
	buf.WriteString(fmt.Sprintf("%s-", sourcePort))

	return hashcode.String(buf.String())
}
