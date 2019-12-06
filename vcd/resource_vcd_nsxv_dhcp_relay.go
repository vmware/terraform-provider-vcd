package vcd

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func resourceVcdNsxvDhcpRelay() *schema.Resource {
	return &schema.Resource{
		Create: resourceVcdNsxvDhcpRelayCreate,
		Read:   resourceVcdNsxvDhcpRelayRead,
		Update: resourceVcdNsxvDhcpRelayUpdate,
		Delete: resourceVcdNsxvDhcpRelayDelete,
		Importer: &schema.ResourceImporter{
			State: resourceVcdNsxvDhcpRelayImport,
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
				Description: "Edge gateway name for DHCP relay settings",
			},
			"ip_addresses": {
				Optional:    true,
				Type:        schema.TypeSet,
				Description: "IP addresses ",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"domain_names": {
				Optional:    true,
				Type:        schema.TypeSet,
				Description: "IP addresses ",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"ip_sets": {
				Optional:    true,
				Type:        schema.TypeSet,
				Description: "IP addresses ",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"relay_agent": {
				Required: true,
				MinItems: 1,
				Type:     schema.TypeSet,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"org_network": {
							Required: true,
							Type:     schema.TypeString,
						},
						"gateway_ip_address": {
							Optional: true,
							Computed: true,
							Type:     schema.TypeString,
						},
					},
				},
			},
		},
	}
}

// resourceVcdNsxvDhcpRelayCreate sets up DHCP relay configuration as per supplied schema
// configuration
func resourceVcdNsxvDhcpRelayCreate(d *schema.ResourceData, meta interface{}) error {
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

	dhcpRelayConfig, err := getDhcpRelayType(d, edgeGateway, vdc)
	if err != nil {
		return fmt.Errorf("could not process DHCP relay settings: %s", err)
	}

	_, err = edgeGateway.UpdateDhcpRelay(dhcpRelayConfig)
	if err != nil {
		return fmt.Errorf("unable to update DHCP relay settings for Edge Gateway %s: %s", edgeGateway.EdgeGateway.Name, err)
	}

	// This is not a real object but a settings property on Edge gateway - creating a fake composite
	// ID
	fakeId, err := getDhclRelaySettingsId(edgeGateway)
	if err != nil {
		return fmt.Errorf("could not construct DHCP relay settings ID: %s", err)
	}

	d.SetId(fakeId)

	return nil
}

// resourceVcdNsxvDhcpRelayUpdate is in fact exactly the same as create because there is no object,
// just settings to modify
func resourceVcdNsxvDhcpRelayUpdate(d *schema.ResourceData, meta interface{}) error {
	return resourceVcdNsxvDhcpRelayCreate(d, meta)
}

// resourceVcdNsxvDhcpRelayRead
func resourceVcdNsxvDhcpRelayRead(d *schema.ResourceData, meta interface{}) error {
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

	dhcpRelaySettings, err := edgeGateway.GetDhcpRelay()
	if err != nil {
		return fmt.Errorf("could not read DHCP relay settings: %s", err)
	}

	getDhcpRelayData(d, dhcpRelaySettings, vdc)

	return nil
}

// resourceVcdNsxvDhcpRelayDelete removes DHCP relay configuration by triggering ResetDhcpRelay()
func resourceVcdNsxvDhcpRelayDelete(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)
	vcdClient.lockParentEdgeGtw(d)
	defer vcdClient.unLockParentEdgeGtw(d)

	edgeGateway, err := vcdClient.GetEdgeGatewayFromResource(d, "edge_gateway")
	if err != nil {
		return fmt.Errorf(errorUnableToFindEdgeGateway, err)
	}

	err = edgeGateway.ResetDhcpRelay()
	if err != nil {
		return fmt.Errorf("could not reset DHCP relay settings: %s", err)
	}

	return nil
}

// resourceVcdNsxvDhcpRelayImport imports DHCP relay configuration. Because DHCP relay is just a
// settings on edge gateway and not a separate object - the ID actually does not represent any object
func resourceVcdNsxvDhcpRelayImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	resourceURI := strings.Split(d.Id(), ImportSeparator)
	if len(resourceURI) != 3 {
		return nil, fmt.Errorf("resource name must be specified in such way org.vdc.edge-gw")
	}
	orgName, vdcName, edgeName := resourceURI[0], resourceURI[1], resourceURI[2]

	vcdClient := meta.(*VCDClient)
	edgeGateway, err := vcdClient.GetEdgeGateway(orgName, vdcName, edgeName)
	if err != nil {
		return nil, fmt.Errorf(errorUnableToFindEdgeGateway, err)
	}

	fakeId, err := getDhclRelaySettingsId(edgeGateway)
	if err != nil {
		return nil, fmt.Errorf("could not construct DHCP relay settings ID: %s", err)
	}

	d.Set("org", orgName)
	d.Set("vdc", vdcName)
	d.Set("edge_gateway", edgeName)
	d.SetId(fakeId)
	return []*schema.ResourceData{d}, nil
}

// getDhcpRelayType converts resource schema to *types.EdgeDhcpRelay
func getDhcpRelayType(d *schema.ResourceData, edge *govcd.EdgeGateway, vdc *govcd.Vdc) (*types.EdgeDhcpRelay, error) {
	dhcpRelayConfig := &types.EdgeDhcpRelay{}

	// Relay server part
	var (
		listOfIps         []string
		listOfDomainNames []string
		listOfIpSetNames  []string
		listOfIpSetIds    []string
		err               error
	)

	if ipAddresses, ok := d.GetOk("ip_addresses"); ok {
		listOfIps = convertSchemaSetToSliceOfStrings(ipAddresses.(*schema.Set))
	}

	if domainNames, ok := d.GetOk("domain_names"); ok {
		listOfDomainNames = convertSchemaSetToSliceOfStrings(domainNames.(*schema.Set))
	}

	if ipSetNames, ok := d.GetOk("ip_sets"); ok {
		listOfIpSetNames = convertSchemaSetToSliceOfStrings(ipSetNames.(*schema.Set))
		listOfIpSetIds, err = ipSetNamesToIds(listOfIpSetNames, vdc)
		if err != nil {
			return nil, fmt.Errorf("could not lookup supplied IP set IDs by their names: %s", err)
		}
	}

	dhcpRelayServer := &types.EdgeDhcpRelayServer{
		IpAddress:        listOfIps,
		Fqdns:            listOfDomainNames,
		GroupingObjectId: listOfIpSetIds,
	}

	// Add DHCP relay server part to struct
	dhcpRelayConfig.RelayServer = dhcpRelayServer

	// Relay agent part
	relayAgent := d.Get("relay_agent")
	relayAgentsStruct, err := getDhcpRelayAgentsType(relayAgent.(*schema.Set), edge)
	if err != nil {
		return nil, fmt.Errorf("could not process relay agents: %s", err)
	}
	// Add all relay agent values to struct
	dhcpRelayConfig.RelayAgents = &types.EdgeDhcpRelayAgents{Agents: relayAgentsStruct}

	return dhcpRelayConfig, nil

}

// getDhcpRelayAgentsType converts relay_agent configuration blocks to []types.EdgeDhcpRelayAgent
func getDhcpRelayAgentsType(relayAgentsSet *schema.Set, edge *govcd.EdgeGateway) ([]types.EdgeDhcpRelayAgent, error) {
	relayAgentsSlice := relayAgentsSet.List()
	relayAgentsStruct := make([]types.EdgeDhcpRelayAgent, len(relayAgentsSlice))
	if len(relayAgentsSlice) > 0 {
		for index, relayAgent := range relayAgentsSlice {
			relayAgentMap := convertToStringMap(relayAgent.(map[string]interface{}))

			// Lookup vNic index by network name
			orgNetworkName := relayAgentMap["org_network"]
			vNicIndex, _, err := edge.GetAnyVnicIndexByNetworkName(orgNetworkName)
			if err != nil {
				return nil, fmt.Errorf("could not lookup edge gateway interface (vNic) index by network name for network %s: %s", orgNetworkName, err)
			}

			oneRelayAgent := types.EdgeDhcpRelayAgent{
				VnicIndex: vNicIndex,
			}

			if gatewayIp, isSet := relayAgentMap["gateway_ip_address"]; isSet {
				oneRelayAgent.GatewayInterfaceAddress = gatewayIp
			}

			relayAgentsStruct[index] = oneRelayAgent
		}
	}

	return relayAgentsStruct, nil
}

func getDhcpRelayData(d *schema.ResourceData, edgeRelay *types.EdgeDhcpRelay, vdc *govcd.Vdc) error {

	// DHCP relay server settings
	relayServer := edgeRelay.RelayServer
	// relayServer.

	relayServerIpAddresses := convertToTypeSet(relayServer.IpAddress)
	relayServerIpAddressesSet := schema.NewSet(schema.HashSchema(&schema.Schema{Type: schema.TypeString}), relayServerIpAddresses)
	err := d.Set("ip_addresses", relayServerIpAddressesSet)
	if err != nil {
		return fmt.Errorf("could not save ip_addresses to schema: %s", err)
	}

	relayServerDomainNames := convertToTypeSet(relayServer.Fqdns)
	relayServerDomainNamesSet := schema.NewSet(schema.HashSchema(&schema.Schema{Type: schema.TypeString}), relayServerDomainNames)
	err = d.Set("domain_names", relayServerDomainNamesSet)
	if err != nil {
		return fmt.Errorf("could not save domain_names to schema: %s", err)
	}
	spew.Dump(relayServer.GroupingObjectId)
	ipSetNames, err := ipSetIdsToNames(relayServer.GroupingObjectId, vdc)
	if err != nil {
		return fmt.Errorf("could not find names for all IP set IDs: %s", err)
	}

	relayServerIpSetNames := convertToTypeSet(ipSetNames)
	relayServerIpSetNamesSet := schema.NewSet(schema.HashSchema(&schema.Schema{Type: schema.TypeString}), relayServerIpSetNames)
	err = d.Set("ip_sets", relayServerIpSetNamesSet)
	if err != nil {
		return fmt.Errorf("could not save ip_sets to schema: %s", err)
	}

	return nil
}

// getDhclRelaySettingsId constructs a fake DHCP relay configuration ID which is needed for
// Terraform The ID is in format "edgeGateway.ID:dhcpRelaySettings". Edge Gateway ID is left here
// just in case we ever want to refer this object somewhere.
func getDhclRelaySettingsId(edge *govcd.EdgeGateway) (string, error) {
	if edge.EdgeGateway.ID == "" {
		return "", fmt.Errorf("edge gateway does not have ID populated")
	}

	id := edge.EdgeGateway.ID + ":dhcpRelaySettings"
	return id, nil
}

// BELOW THIS LINE To be removed when other PR is merged
func ipSetNamesToIds(ipSetNames []string, vdc *govcd.Vdc) ([]string, error) {
	ipSetIds := make([]string, len(ipSetNames))

	allIpSets, err := vdc.GetAllNsxvIpSets()
	if err != nil {
		return nil, fmt.Errorf("unable to fetch all IP sets in vDC %s: %s", vdc.Vdc.Name, err)
	}

	for index, ipSetName := range ipSetNames {
		var ipSetFound bool
		for _, ipSet := range allIpSets {
			if ipSet.Name == ipSetName {
				ipSetIds[index] = ipSet.ID
				ipSetFound = true
			}
		}
		// If ID was not found - fail early
		if !ipSetFound {
			return nil, fmt.Errorf("could not find IP set with Name %s", ipSetName)
		}
	}

	return ipSetIds, nil
}

func ipSetIdsToNames(ipSetIds []string, vdc *govcd.Vdc) ([]string, error) {
	ipSetNames := make([]string, len(ipSetIds))

	allIpSets, err := vdc.GetAllNsxvIpSets()
	if err != nil {
		return nil, fmt.Errorf("unable to fetch all IP sets in vDC %s: %s", vdc.Vdc.Name, err)
	}

	for index, ipSetId := range ipSetIds {
		var ipSetFound bool
		for _, ipSet := range allIpSets {
			if ipSet.ID == ipSetId {
				ipSetNames[index] = ipSet.Name
				ipSetFound = true
			}
		}
		// If ID was not found - fail early
		if !ipSetFound {
			return nil, fmt.Errorf("could not find IP set with ID %s", ipSetId)
		}
	}

	return ipSetNames, nil
}
