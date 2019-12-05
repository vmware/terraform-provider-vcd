package vcd

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func resourceVcdNsxvDhcpRelay() *schema.Resource {
	return &schema.Resource{
		Create: resourceVcdNsxvDhcpRelayCreate,
		Read:   resourceVcdNsxvDhcpRelayRead,
		// Update: resourceVcdNsxvDhcpRelayUpdate,
		Delete: resourceVcdNsxvDhcpRelayDelete,
		// Importer: &schema.ResourceImporter{
		// 	State: resourceVcdNsxvDhcpRelayImport,
		// },

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
				ForceNew:    true,
				Optional:    true,
				Type:        schema.TypeSet,
				Description: "IP addresses ",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"domain_names": {
				ForceNew:    true,
				Optional:    true,
				Type:        schema.TypeSet,
				Description: "IP addresses ",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"ip_sets": {
				ForceNew:    true,
				Optional:    true,
				Type:        schema.TypeSet,
				Description: "IP addresses ",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"relay_agent": {
				ForceNew: true,
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

// resourceVcdNsxvDhcpRelayCreate
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

	// Because this is not a real object but a settings property - let create a fake composite ID
	fakeId := edgeGateway.EdgeGateway.ID + ":dhcpRelaySettings"
	d.SetId(fakeId)

	return nil
}

// resourceVcdNsxvDhcpRelayUpdate
func resourceVcdNsxvDhcpRelayUpdate(d *schema.ResourceData, meta interface{}) error {
	return nil
}

// resourceVcdNsxvDhcpRelayRead
func resourceVcdNsxvDhcpRelayRead(d *schema.ResourceData, meta interface{}) error {
	return nil
}

// resourceVcdNsxvDhcpRelayDelete
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

// resourceVcdNsxvDhcpRelayImport
func resourceVcdNsxvDhcpRelayImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	return []*schema.ResourceData{d}, nil
}

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

	if ipSetNames, ok := d.GetOk("ipsets"); ok {
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
	if relayAgent, ok := d.GetOk("relay_agent"); ok {
		relayAgentsSet := relayAgent.(*schema.Set)
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
		// Add all relay agent values to struct
		dhcpRelayConfig.RelayAgents = &types.EdgeDhcpRelayAgents{Agents: relayAgentsStruct}
	}

	return dhcpRelayConfig, nil

}

// To be removed when other PR is merged
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
