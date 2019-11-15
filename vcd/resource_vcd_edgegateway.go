package vcd

import (
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

var subnetResource = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"ip_address": {
			Optional:    true,
			Computed:    true,
			Type:        schema.TypeString,
			Description: "IP address on the edge gateway - will be auto-assigned if not defined",
		},
		"gateway": {
			Required: true,
			Type:     schema.TypeString,
		},
		"netmask": {
			Required: true,
			Type:     schema.TypeString,
		},
		"use_for_default_route": {
			Optional:    true,
			Default:     false,
			Type:        schema.TypeBool,
			Description: "Defines if this subnet should be used as default gateway for edge",
		},
	},
}

var externalNetworkResource = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"name": {
			Required:    true,
			Type:        schema.TypeString,
			Description: "External network name",
		},
		"enable_rate_limit": {
			Optional: true,
			Computed: true,
			// Default:     false,
			Type:        schema.TypeBool,
			Description: "Enable rate limitting",
		},
		"incoming_rate_limit": {
			Optional:    true,
			Computed:    true,
			Type:        schema.TypeFloat,
			Description: "Incoming rate limit (Mbps)",
		},
		"outgoing_rate_limit": {
			Optional:    true,
			Computed:    true,
			Type:        schema.TypeFloat,
			Description: "Outgoing rate limit (Mbps)",
		},
		"subnet": {
			Required: true,
			// ForceNew: true,
			Type:     schema.TypeSet,
			MinItems: 1,
			Elem:     subnetResource,
			// Set:      resourceVcdEdgeGatewayExternalNetworkSubnetHash,
		},
	},
}

func resourceVcdEdgeGateway() *schema.Resource {

	return &schema.Resource{
		Create: resourceVcdEdgeGatewayCreate,
		Read:   resourceVcdEdgeGatewayRead,
		Update: resourceVcdEdgeGatewayUpdate,
		Delete: resourceVcdEdgeGatewayDelete,
		Importer: &schema.ResourceImporter{
			State: resourceVcdEdgeGatewayImport,
		},

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"org": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Description: "The name of organization to use, optional if defined at provider " +
					"level. Useful when connected as sysadmin working across different organizations",
			},
			"vdc": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "The name of VDC to use, optional if defined at provider level",
			},
			"description": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"advanced": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				ForceNew:    true,
				Description: "True if the gateway uses advanced networking. (Enabled by default)",
			},
			"configuration": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: `Configuration of the vShield edge VM for this gateway. One of: compact, full ("Large"), full4 ("Quad Large"), x-large`,
			},
			"ha_enabled": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				ForceNew:    true,
				Description: "Enable high availability on this edge gateway",
			},
			"external_networks": &schema.Schema{
				ConflictsWith: []string{"external_network"},
				Type:          schema.TypeList,
				Optional:      true,
				ForceNew:      true,
				Description:   "A list of external networks to be used by the edge gateway",
				Deprecated:    "Please use the more advanced 'external_network' block(s)",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"default_gateway_network": &schema.Schema{
				ConflictsWith: []string{"external_network"},
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				// Default:       "",
				ForceNew:    true,
				Description: "External network to be used as default gateway. Its name must be included in 'external_networks'. An empty value will skip the default gateway",
			},
			"default_external_network_ip": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "IP address of edge gateway interface which is used as default.",
			},
			"distributed_routing": &schema.Schema{
				Type:             schema.TypeBool,
				Optional:         true,
				Default:          false,
				ForceNew:         true,
				Description:      "If advanced networking enabled, also enable distributed routing",
				DiffSuppressFunc: suppressFalse(),
			},
			"lb_enabled": &schema.Schema{
				Type:        schema.TypeBool,
				Default:     false,
				Optional:    true,
				Description: "Enable load balancing. (Disabled by default)",
			},
			"lb_acceleration_enabled": &schema.Schema{
				Type:        schema.TypeBool,
				Default:     false,
				Optional:    true,
				Description: "Enable load balancer acceleration. (Disabled by default)",
			},
			"lb_logging_enabled": &schema.Schema{
				Type:        schema.TypeBool,
				Default:     false,
				Optional:    true,
				Description: "Enable load balancer logging. (Disabled by default)",
			},
			"lb_loglevel": &schema.Schema{
				Type:         schema.TypeString,
				Default:      "info",
				Optional:     true,
				ValidateFunc: validateCase("lower"),
				Description: "Log level. One of 'emergency', 'alert', 'critical', 'error', " +
					"'warning', 'notice', 'info', 'debug'. ('info' by default)",
			},
			"fw_enabled": &schema.Schema{
				Type:        schema.TypeBool,
				Default:     true,
				Optional:    true,
				Description: "Enable firewall. Default 'true'",
			},
			"fw_default_rule_logging_enabled": &schema.Schema{
				Type:        schema.TypeBool,
				Default:     false,
				Optional:    true,
				Description: "Enable logging for default rule. Default 'false'",
			},
			"fw_default_rule_action": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "deny",
				Description:  "'accept' or 'deny'. Default 'deny'",
				ValidateFunc: validation.StringInSlice([]string{"accept", "deny"}, false),
			},

			"fips_mode_enabled": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "Enable FIPS mode. FIPS mode turns on the cipher suites that comply with FIPS. (False by default)",
			},
			"use_default_route_for_dns_relay": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Computed:    true,
				Description: "If true, default gateway will be used for the edge gateways' default routing and DNS forwarding.(False by default)",
			},
			// "suballocation_pool": {
			// 	Optional:    true,
			// 	Type:        schema.TypeSet,
			// 	Description: "string",
			// 	Elem: &schema.Resource{
			// 		Schema: map[string]*schema.Schema{
			// 			"start_address": {
			// 				Required: true,
			// 				Type:     schema.TypeString,
			// 			},
			// 			"end_address": {
			// 				Optional: true,
			// 				Computed: true,
			// 				Type:     schema.TypeString,
			// 			},
			// 		},
			// 	},
			// },
			"external_network": {
				ConflictsWith: []string{"external_networks", "default_gateway_network"},
				Optional:      true,
				// ForceNew:      true,
				// MinItems:      1,
				Type: schema.TypeSet,
				// Set:           resourceVcdNsxvFirewallRuleServiceHash,
				Elem: externalNetworkResource,
			},
		},
	}
}

// Creates a new edge gateway from a resource definition
func resourceVcdEdgeGatewayCreate(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[TRACE] edge gateway creation initiated")

	vcdClient := meta.(*VCDClient)

	// Making sure the parent entities are available
	orgName := vcdClient.getOrgName(d)
	vdcName := vcdClient.getVdcName(d)

	var missing []string
	if orgName == "" {
		missing = append(missing, "org")
	}
	if vdcName == "" {
		missing = append(missing, "vdc")
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing properties. %v should be given either in the resource or at provider level", missing)
	}

	org, vdc, err := vcdClient.GetOrgAndVdc(orgName, vdcName)
	if err != nil {
		return err
	}
	if org == nil {
		return fmt.Errorf("no valid Organization named '%s' was found", orgName)
	}
	if vdc == nil || vdc.Vdc.HREF == "" || vdc.Vdc.ID == "" || vdc.Vdc.Name == "" {
		return fmt.Errorf("no valid VDC named '%s' was found", vdcName)
	}

	// In version 9.7+ the advanced property is true by default
	if vcdClient.APIVCDMaxVersionIs(">= 32.0") {
		if !d.Get("advanced").(bool) {
			return fmt.Errorf("'advanced' property for vCD 9.7+ must be set to 'true'")
		}
	}

	var gwInterfaces []*types.GatewayInterface

	simpleExtNetworksSlice, simpleExtNetsExist := d.GetOk("external_networks")
	simpleExtDefaultGwNet := d.Get("default_gateway_network")
	if simpleExtNetsExist {
		log.Printf("[TRACE] creating edge gateway using simple 'external_networks' and 'default_gateway_network' fields")
		// Get gateway interfaces from simple structure
		oldExtNetworksSliceString := convertToStringSlice(simpleExtNetworksSlice.([]interface{}))
		gwInterfaces, err = getGatewayExternalNetworks(vcdClient, oldExtNetworksSliceString, simpleExtDefaultGwNet.(string))
		if err != nil {
			return fmt.Errorf("could not process 'external_networks' and 'default_gateway_network': %s", err)
		}
	} else {
		log.Printf("[TRACE] creating edge gateway using advanced 'external_network' blocks")
		// Get gateway interfaces from complex structure
		gwInterfaces, err = getGatewayInterfaces(vcdClient, d.Get("external_network").(*schema.Set))
		if err != nil {
			return fmt.Errorf("could not process 'external_network' block(s): %s", err)
		}
	}

	egwName := d.Get("name").(string)
	egwConfiguration := &types.EdgeGateway{
		Xmlns:       types.XMLNamespaceVCloud,
		Name:        egwName,
		Description: d.Get("description").(string),
		Configuration: &types.GatewayConfiguration{
			UseDefaultRouteForDNSRelay: d.Get("use_default_route_for_dns_relay").(bool),
			FipsModeEnabled:            takeBoolPointer(d.Get("fips_mode_enabled").(bool)),
			HaEnabled:                  d.Get("ha_enabled").(bool),
			GatewayBackingConfig:       d.Get("configuration").(string),
			AdvancedNetworkingEnabled:  d.Get("advanced").(bool),
			DistributedRoutingEnabled:  takeBoolPointer(d.Get("distributed_routing").(bool)),
			GatewayInterfaces: &types.GatewayInterfaces{
				GatewayInterface: gwInterfaces,
			},
			EdgeGatewayServiceConfiguration: &types.GatewayFeatures{},
		},
	}

	edge, err := govcd.CreateAndConfigureEdgeGateway(vcdClient.VCDClient, orgName, vdcName, egwName, egwConfiguration)
	if err != nil {
		log.Printf("[DEBUG] Error creating edge gateway: %s", err)
		return fmt.Errorf("error creating edge gateway: %s", err)
	}
	// Edge gateway creation succeeded therefore we save related fields now to preserve Id.
	// Edge gateway is already created even if further process fails
	log.Printf("[TRACE] flushing edge gateway creation fields")
	err = setEdgeGatewayValues(d, edge, "resource")
	if err != nil {
		return err
	}

	// Only perform load balancer and firewall configuration if gateway is advanced
	if d.Get("advanced").(bool) {
		log.Printf("[TRACE] edge gateway load balancer configuration started")

		err := updateLoadBalancer(d, edge)
		if err != nil {
			return fmt.Errorf("unable to update general load balancer settings: %s", err)
		}

		log.Printf("[TRACE] edge gateway load balancer configured")

		log.Printf("[TRACE] edge gateway firewall configuration started")

		err = updateFirewall(d, edge)
		if err != nil {
			return fmt.Errorf("unable to update firewall settings: %s", err)
		}

		log.Printf("[TRACE] edge gateway firewall configured")

		// update load balancer and firewall configuration in statefile
		err = setEdgeGatewayComponentValues(d, edge)
		if err != nil {
			return err
		}
	}

	// TODO double validate if we need to use partial state here
	// https://www.terraform.io/docs/extend/writing-custom-providers.html#error-handling-amp-partial-state
	d.SetId(edge.EdgeGateway.ID)
	log.Printf("[TRACE] edge gateway created: %#v", edge.EdgeGateway.Name)
	return resourceVcdEdgeGatewayRead(d, meta)
}

// Fetches information about an existing edge gateway for a data definition
func resourceVcdEdgeGatewayRead(d *schema.ResourceData, meta interface{}) error {
	return genericVcdEdgeGatewayRead(d, meta, "resource")
}

func genericVcdEdgeGatewayRead(d *schema.ResourceData, meta interface{}, origin string) error {
	log.Printf("[TRACE] edge gateway read initiated")

	vcdClient := meta.(*VCDClient)

	identifier := d.Id()
	if identifier == "" {
		identifier = d.Get("name").(string)
	}
	if identifier == "" {
		return fmt.Errorf("[edgegateway read] no identifier provided")
	}

	orgName := d.Get("org").(string)
	vdcName := d.Get("vdc").(string)
	_, vdc, err := vcdClient.GetOrgAndVdc(orgName, vdcName)
	if err != nil {
		return fmt.Errorf("[edgegateway read] error retrieving org and vdc: %s", err)
	}
	edgeGateway, err := vdc.GetEdgeGatewayByNameOrId(identifier, false)
	if err != nil {
		if origin == "resource" {
			log.Printf("[edgegateway read] edge gateway %s not found. Removing from state file: %s", identifier, err)
			d.SetId("")
			return nil
		}
		return fmt.Errorf("[edgegateway read] error retrieving edge gateway %s: %s", identifier, err)
	}

	if err := setEdgeGatewayValues(d, *edgeGateway, origin); err != nil {
		return err
	}

	// Only read and set the statefile if the edge gateway is advanced
	if edgeGateway.HasAdvancedNetworking() {
		if err := setLoadBalancerData(d, *edgeGateway); err != nil {
			return err
		}

		if err := setFirewallData(d, *edgeGateway); err != nil {
			return err
		}
	}

	d.SetId(edgeGateway.EdgeGateway.ID)

	log.Printf("[TRACE] edge gateway read completed: %#v", edgeGateway.EdgeGateway)
	return nil
}

// resourceVcdEdgeGatewayUpdate updates general load balancer settings only at the moment
func resourceVcdEdgeGatewayUpdate(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)
	vcdClient.lockEdgeGateway(d)
	defer vcdClient.unlockEdgeGateway(d)

	edgeGateway, err := vcdClient.GetEdgeGatewayFromResource(d, "name")
	if err != nil {
		return nil
	}

	// If edge gateway is advanced - check if load balancer or firewall needs adjustments
	if edgeGateway.HasAdvancedNetworking() {
		if d.HasChange("lb_enabled") || d.HasChange("lb_acceleration_enabled") ||
			d.HasChange("lb_logging_enabled") || d.HasChange("lb_loglevel") {
			err := updateLoadBalancer(d, *edgeGateway)
			if err != nil {
				return err
			}
		}

		if d.HasChange("fw_enabled") || d.HasChange("fw_default_rule_logging_enabled") ||
			d.HasChange("fw_default_rule_action") {
			err := updateFirewall(d, *edgeGateway)
			if err != nil {
				return err
			}
		}
	}

	return resourceVcdEdgeGatewayRead(d, meta)
}

// Deletes a edge gateway, optionally removing all objects in it as well
func resourceVcdEdgeGatewayDelete(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[TRACE] edge gateway delete started")

	vcdClient := meta.(*VCDClient)

	vcdClient.lockEdgeGateway(d)
	defer vcdClient.unlockEdgeGateway(d)

	edgeGateway, err := vcdClient.GetEdgeGatewayFromResource(d, "name")
	if err != nil {
		return fmt.Errorf("error fetching edge gateway details %#v", err)
	}

	err = edgeGateway.Delete(true, true)

	log.Printf("[TRACE] edge gateway deletion completed\n")
	return err
}

// getGatewayExternalNetworks aims to add compatibility layer to go-vcloud-director
// CreateEdgeGateway function which is a wrapper around CreateAndConfigureEdgeGateway. The layer
// resides here so that code can work together getGatewayInterfaces
func getGatewayExternalNetworks(vcdClient *VCDClient, externalNetworks []string, defaultGatewayNetwork string) ([]*types.GatewayInterface, error) {
	gatewayInterfaces := make([]*types.GatewayInterface, len(externalNetworks))
	// Add external networks inside the configuration structure
	for extNetworkIndex, extNetName := range externalNetworks {
		extNet, err := vcdClient.GetExternalNetworkByName(extNetName)
		if err != nil {
			return nil, err
		}

		// Populate the subnet participation only if default gateway was set
		var subnetParticipation *types.SubnetParticipation
		if defaultGatewayNetwork != "" && extNet.ExternalNetwork.Name == defaultGatewayNetwork {
			for _, net := range extNet.ExternalNetwork.Configuration.IPScopes.IPScope {
				if net.IsEnabled {
					subnetParticipation = &types.SubnetParticipation{
						Gateway: net.Gateway,
						Netmask: net.Netmask,
					}
					break
				}
			}
		}
		networkConf := &types.GatewayInterface{
			Name:          extNet.ExternalNetwork.Name,
			DisplayName:   extNet.ExternalNetwork.Name,
			InterfaceType: "uplink",
			Network: &types.Reference{
				HREF: extNet.ExternalNetwork.HREF,
				ID:   extNet.ExternalNetwork.ID,
				Type: "application/vnd.vmware.admin.network+xml",
				Name: extNet.ExternalNetwork.Name,
			},
			UseForDefaultRoute:  defaultGatewayNetwork == extNet.ExternalNetwork.Name,
			SubnetParticipation: []*types.SubnetParticipation{subnetParticipation},
		}

		gatewayInterfaces[extNetworkIndex] = networkConf
	}
	return gatewayInterfaces, nil
}

// getExternalNetworks extracts processes `external_network` blocks with more advanced settings
func getGatewayInterfaces(vcdClient *VCDClient, externalInterfaceSet *schema.Set) ([]*types.GatewayInterface, error) {
	var gatewayInterfaceSlice []*types.GatewayInterface

	extInterfaceList := externalInterfaceSet.List()
	if len(extInterfaceList) > 0 {
		gatewayInterfaceSlice = make([]*types.GatewayInterface, len(extInterfaceList))

		for extInterfaceIndex, extInterface := range extInterfaceList {
			extInterfaceMap := extInterface.(map[string]interface{})
			externalNetworkName := extInterfaceMap["name"].(string)
			externalNetwork, err := vcdClient.GetExternalNetworkByName(externalNetworkName)
			if err != nil {
				return nil, fmt.Errorf("could not look up external network %s by name: %s", externalNetworkName, err)
			}
			var isInterfaceUsedForDefaultRoute bool

			subnetSet := extInterfaceMap["subnet"].(*schema.Set)
			subnetList := subnetSet.List()
			subnetParticipationSlice := make([]*types.SubnetParticipation, len(subnetList))
			if len(subnetList) > 0 {
				for subnetIndex, subnet := range subnetList {
					subnetMap := subnet.(map[string]interface{})
					subnetParticipationSlice[subnetIndex] = &types.SubnetParticipation{
						IPAddress:          subnetMap["ip_address"].(string),
						Gateway:            subnetMap["gateway"].(string),
						Netmask:            subnetMap["netmask"].(string),
						UseForDefaultRoute: subnetMap["use_for_default_route"].(bool),
					}

					if subnetMap["use_for_default_route"].(bool) {
						isInterfaceUsedForDefaultRoute = true
					}
				}
			}

			// Create interface and add it to gatewayInterfaceSlice
			gatewayInterface := &types.GatewayInterface{
				Name:          externalNetwork.ExternalNetwork.Name,
				DisplayName:   externalNetwork.ExternalNetwork.Name,
				InterfaceType: "uplink",
				Network: &types.Reference{
					HREF: externalNetwork.ExternalNetwork.HREF,
					ID:   externalNetwork.ExternalNetwork.ID,
					Type: "application/vnd.vmware.admin.network+xml",
					Name: externalNetwork.ExternalNetwork.Name,
				},
				UseForDefaultRoute:  isInterfaceUsedForDefaultRoute,
				ApplyRateLimit:      extInterfaceMap["enable_rate_limit"].(bool),
				InRateLimit:         extInterfaceMap["incoming_rate_limit"].(float64),
				OutRateLimit:        extInterfaceMap["outgoing_rate_limit"].(float64),
				SubnetParticipation: subnetParticipationSlice,
			}
			// Add populated network interface to the slice
			gatewayInterfaceSlice[extInterfaceIndex] = gatewayInterface
		}
	}

	return gatewayInterfaceSlice, nil
}

func getExternalNetworkData(gatewayInterfaces []*types.GatewayInterface) (*schema.Set, error) {
	externalNetworkSlice := make([]interface{}, len(gatewayInterfaces))
	if len(gatewayInterfaces) > 0 {
		for extNetworkIndex, extNetwork := range gatewayInterfaces {
			extNetworkMap := make(map[string]interface{})
			extNetworkMap["name"] = extNetwork.Network.Name
			extNetworkMap["enable_rate_limit"] = extNetwork.ApplyRateLimit
			extNetworkMap["incoming_rate_limit"] = extNetwork.InRateLimit
			extNetworkMap["outgoing_rate_limit"] = extNetwork.OutRateLimit

			externalNetworkSubnetSlice := make([]interface{}, len(extNetwork.SubnetParticipation))
			for extNetsubnetIndex, extNetsubnet := range extNetwork.SubnetParticipation {
				extNetworkSubnetMap := make(map[string]interface{})
				extNetworkSubnetMap["ip_address"] = extNetsubnet.IPAddress
				extNetworkSubnetMap["gateway"] = extNetsubnet.Gateway
				extNetworkSubnetMap["netmask"] = extNetsubnet.Netmask
				extNetworkSubnetMap["use_for_default_route"] = extNetsubnet.UseForDefaultRoute

				// Make a set and add it to externalNetworkSubnetSlice
				externalNetworkSubnetSlice[extNetsubnetIndex] = extNetworkSubnetMap
			}

			extNetworkMap["subnet"] = schema.NewSet(schema.HashResource(subnetResource), externalNetworkSubnetSlice)
			externalNetworkSlice[extNetworkIndex] = extNetworkMap
		}
	}
	externalNetworkSet := schema.NewSet(schema.HashResource(externalNetworkResource), externalNetworkSlice)

	return externalNetworkSet, nil
}

// Convenience function to fill edge gateway values from resource data
func setEdgeGatewayValues(d *schema.ResourceData, egw govcd.EdgeGateway, origin string) error {

	d.SetId(egw.EdgeGateway.ID)
	err := d.Set("name", egw.EdgeGateway.Name)
	if err != nil {
		return err
	}
	err = d.Set("description", egw.EdgeGateway.Description)
	if err != nil {
		return err
	}
	err = d.Set("configuration", egw.EdgeGateway.Configuration.GatewayBackingConfig)
	if err != nil {
		return err
	}

	_, externalNetworksSet := d.GetOk("external_networks")
	// When `external_networks` field was not used - we set a more rich `external_network` block
	// which allows to set multiple used subnets, manual IP addresses for IPs assigned to edge
	// gateway and which subnet should be used as the default one for edge gateway
	if !externalNetworksSet {
		externalNetworkData, err := getExternalNetworkData(egw.EdgeGateway.Configuration.GatewayInterfaces.GatewayInterface)
		if err != nil {
			return fmt.Errorf("[edgegateway read] could not process network interface data: %s", err)
		}

		err = d.Set("external_network", externalNetworkData)
		if err != nil {
			return fmt.Errorf("[edgegateway read] could not set external_network block: %s", err)
		}

	}

	// only if `external_networks` field was used or it is a data source we set the older
	// fields "external_networks"
	if externalNetworksSet || origin == "datasource" {
		var gateways = make(map[string]string)
		var networks []string
		for _, net := range egw.EdgeGateway.Configuration.GatewayInterfaces.GatewayInterface {
			if net.InterfaceType == "uplink" {
				networks = append(networks, net.Network.Name)

				for _, subnet := range net.SubnetParticipation {
					gateways[subnet.Gateway] = net.Network.Name
				}
			}
		}
		err = d.Set("external_networks", networks)
		if err != nil {
			return err
		}
	}

	_ = d.Set("use_default_route_for_dns_relay", egw.EdgeGateway.Configuration.UseDefaultRouteForDNSRelay)
	_ = d.Set("fips_mode_enabled", egw.EdgeGateway.Configuration.FipsModeEnabled)
	_ = d.Set("advanced", egw.EdgeGateway.Configuration.AdvancedNetworkingEnabled)
	_ = d.Set("ha_enabled", egw.EdgeGateway.Configuration.HaEnabled)

	for _, gw := range egw.EdgeGateway.Configuration.GatewayInterfaces.GatewayInterface {
		if len(gw.SubnetParticipation) < 1 {
			log.Printf("[DEBUG] [setEdgeGatewayValues] gateway %s is missing SubnetParticipation elements: %+#v",
				egw.EdgeGateway.Name, gw)

			return fmt.Errorf("[setEdgeGatewayValues] gateway %s is missing SubnetParticipation elements",
				egw.EdgeGateway.Name)
		}

		for _, subnet := range gw.SubnetParticipation {
			// Check if this subnet is used as default gateway and set IP and `default_gateway_network` value
			if subnet.UseForDefaultRoute {
				_ = d.Set("default_gateway_network", gw.Network.Name)
				_ = d.Set("default_external_network_ip", subnet.IPAddress)
			}
		}
	}
	// TODO: Enable this setting after we switch to a higher API version.
	//Based on testing the API does accept (and set) the setting, but upon GET query it omits the DistributedRouting
	// field therefore struct field defaults to false after unmarshaling.
	//This has already been a case for us and then it was proven that API v27.0 does not return field, while API v31.0
	// does return the field. (Thanks, Dainius)
	//err = d.Set("distributed_routing", egw.EdgeGateway.Configuration.DistributedRoutingEnabled)
	//if err != nil {
	//	return err
	//}

	d.SetId(egw.EdgeGateway.ID)
	return nil
}

// setEdgeGatewayComponentValues sets component values to the statefile which are created with
// additional API calls
func setEdgeGatewayComponentValues(d *schema.ResourceData, egw govcd.EdgeGateway) error {
	if egw.HasAdvancedNetworking() {
		err := setLoadBalancerData(d, egw)
		if err != nil {
			return err
		}

		err = setFirewallData(d, egw)
		if err != nil {
			return err
		}
	}
	return nil
}

// setLoadBalancerData is a convenience function to handle load balancer settings on edge gateway
func setLoadBalancerData(d *schema.ResourceData, egw govcd.EdgeGateway) error {
	lb, err := egw.GetLBGeneralParams()
	if err != nil {
		return fmt.Errorf("unable to read general load balancer settings: %s", err)
	}

	_ = d.Set("lb_enabled", lb.Enabled)
	_ = d.Set("lb_acceleration_enabled", lb.AccelerationEnabled)
	_ = d.Set("lb_logging_enabled", lb.Logging.Enable)
	_ = d.Set("lb_loglevel", lb.Logging.LogLevel)

	return nil
}

// setFirewallData is a convenience function to handle firewall settings on edge gateway
func setFirewallData(d *schema.ResourceData, egw govcd.EdgeGateway) error {
	fw, err := egw.GetFirewallConfig()
	if err != nil {
		return fmt.Errorf("unable to read firewall settings: %s", err)
	}

	_ = d.Set("fw_enabled", fw.Enabled)
	_ = d.Set("fw_default_rule_logging_enabled", fw.DefaultPolicy.LoggingEnabled)
	_ = d.Set("fw_default_rule_action", fw.DefaultPolicy.Action)

	return nil
}

// updateLoadBalancer updates general load balancer configuration
func updateLoadBalancer(d *schema.ResourceData, egw govcd.EdgeGateway) error {
	lbEnabled := d.Get("lb_enabled").(bool)
	lbAccelerationEnabled := d.Get("lb_acceleration_enabled").(bool)
	lbLoggingEnabled := d.Get("lb_logging_enabled").(bool)
	lbLogLevel := d.Get("lb_loglevel").(string)
	_, err := egw.UpdateLBGeneralParams(lbEnabled, lbAccelerationEnabled, lbLoggingEnabled, lbLogLevel)
	if err != nil {
		return fmt.Errorf("unable to update general load balancer settings: %s", err)
	}

	return nil
}

// updateFirewall updates general firewall configuration
func updateFirewall(d *schema.ResourceData, egw govcd.EdgeGateway) error {
	fwEnabled := d.Get("fw_enabled").(bool)
	fwDefaultRuleLogging := d.Get("fw_default_rule_logging_enabled").(bool)
	fwDefaultRuleAction := d.Get("fw_default_rule_action").(string)
	_, err := egw.UpdateFirewallConfig(fwEnabled, fwDefaultRuleLogging, fwDefaultRuleAction)
	if err != nil {
		return fmt.Errorf("unable to update firewall settings: %s", err)
	}

	return nil
}

// resourceVcdEdgeGatewayImport is responsible for importing the resource.
// The following steps happen as part of import
// 1. The user supplies `terraform import _resource_name_ _the_id_string_` command
// 2. `_the_id_string_` contains a dot formatted path to resource as in the example below
// 3. The functions splits the dot-formatted path and tries to lookup the object
// 4. If the lookup succeeds it sets the ID field for `_resource_name_` resource in statefile
// (the resource must be already defined in .tf config otherwise `terraform import` will complain)
// 5. `terraform refresh` is being implicitly launched. The Read method looks up all other fields
// based on the known ID of object.
//
// Example resource name (_resource_name_): vcd_edgegateway.my-edge-gateway
// Example import path (_the_id_string_): org.vdc.my-edge-gw
// Note: the separator can be changed using Provider.import_separator or variable VCD_IMPORT_SEPARATOR
func resourceVcdEdgeGatewayImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	resourceURI := strings.Split(d.Id(), ImportSeparator)
	if len(resourceURI) != 3 {
		return nil, fmt.Errorf("resource name must be specified as org-name.vdc-name.edge-gw-name")
	}
	orgName, vdcName, edgeName := resourceURI[0], resourceURI[1], resourceURI[2]

	vcdClient := meta.(*VCDClient)
	edgeGateway, err := vcdClient.GetEdgeGateway(orgName, vdcName, edgeName)
	if err != nil {
		return nil, fmt.Errorf(errorUnableToFindEdgeGateway, err)
	}

	_ = d.Set("org", orgName)
	_ = d.Set("vdc", vdcName)
	d.SetId(edgeGateway.EdgeGateway.ID)
	return []*schema.ResourceData{d}, nil
}
