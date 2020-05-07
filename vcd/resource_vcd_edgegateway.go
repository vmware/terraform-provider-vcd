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

var subAllocationPool = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"start_address": {
			Required: true,
			Type:     schema.TypeString,
			ForceNew: true,
		},
		"end_address": {
			Required: true,
			Type:     schema.TypeString,
			ForceNew: true,
		},
	},
}

var subnetResource = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"gateway": {
			Required:    true,
			ForceNew:    true,
			Description: "Gateway address for a subnet",
			Type:        schema.TypeString,
		},
		"netmask": {
			Required:    true,
			ForceNew:    true,
			Description: "Netmask address for a subnet",
			Type:        schema.TypeString,
		},
		"ip_address": {
			Optional:    true,
			Type:        schema.TypeString,
			ForceNew:    true,
			Description: "IP address on the edge gateway - will be auto-assigned if not defined",
		},
		"use_for_default_route": {
			Optional:    true,
			Default:     false,
			ForceNew:    true,
			Type:        schema.TypeBool,
			Description: "Defines if this subnet should be used as default gateway for edge",
		},
		"suballocate_pool": {
			Optional:    true,
			Type:        schema.TypeSet,
			ForceNew:    true,
			Description: "Define zero or more blocks to sub-allocate pools on the edge gateway",
			Elem:        subAllocationPool,
		},
	},
}

var externalNetworkResource = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"name": {
			Required:    true,
			ForceNew:    true,
			Type:        schema.TypeString,
			Description: "External network name",
		},
		"enable_rate_limit": {
			Optional:    true,
			Default:     false,
			ForceNew:    true,
			Type:        schema.TypeBool,
			Description: "Enable rate limiting",
		},
		"incoming_rate_limit": {
			Optional:    true,
			Default:     0,
			ForceNew:    true,
			Type:        schema.TypeFloat,
			Description: "Incoming rate limit (Mbps)",
		},
		"outgoing_rate_limit": {
			Optional:    true,
			Default:     0,
			ForceNew:    true,
			Type:        schema.TypeFloat,
			Description: "Outgoing rate limit (Mbps)",
		},
		"subnet": {
			Optional: true,
			Computed: true,
			ForceNew: true,
			Type:     schema.TypeSet,
			MinItems: 1,
			Elem:     subnetResource,
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
				ForceNew:      true,
				Deprecated:    "Please use 'use_for_default_route' flag in the more advanced 'external_network' block(s)",
				Description:   "External network to be used as default gateway. Its name must be included in 'external_networks'. An empty value will skip the default gateway",
			},
			"default_external_network_ip": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "IP address of edge gateway interface which is used as default.",
			},
			"external_network_ips": {
				Computed:    true,
				Type:        schema.TypeList,
				Description: "List of IP addresses set on edge gateway external network interfaces",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"distributed_routing": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				ForceNew:    true,
				Description: "If advanced networking enabled, also enable distributed routing",
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
				ForceNew:    true,
				Optional:    true,
				Default:     false,
				Description: "Enable FIPS mode. FIPS mode turns on the cipher suites that comply with FIPS. (False by default)",
			},
			"use_default_route_for_dns_relay": &schema.Schema{
				Type:        schema.TypeBool,
				ForceNew:    true,
				Optional:    true,
				Computed:    true,
				Description: "If true, default gateway will be used for the edge gateways' default routing and DNS forwarding.(False by default)",
			},
			"external_network": {
				ConflictsWith: []string{"external_networks", "default_gateway_network"},
				Description:   "One or more blocks with external network information to be attached to this gateway's interface",
				ForceNew:      true,
				Optional:      true,
				Computed:      true,
				Type:          schema.TypeSet,
				Elem:          externalNetworkResource,
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
	if vcdClient.Client.APIVCDMaxVersionIs(">= 32.0") {
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
		gwInterfaces, err = getSimpleGatewayInterfaces(vcdClient, oldExtNetworksSliceString, simpleExtDefaultGwNet.(string))
		if err != nil {
			return fmt.Errorf("could not process 'external_networks' and 'default_gateway_network': %s", err)
		}
	} else {
		log.Printf("[TRACE] creating edge gateway using advanced 'external_network' blocks")
		// Get gateway interfaces from complex structure
		gwInterfaces, err = getGatewayInterfacesType(vcdClient, d.Get("external_network").(*schema.Set))
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
			UseDefaultRouteForDNSRelay: takeBoolPointer(d.Get("use_default_route_for_dns_relay").(bool)),
			HaEnabled:                  takeBoolPointer(d.Get("ha_enabled").(bool)),
			GatewayBackingConfig:       d.Get("configuration").(string),
			AdvancedNetworkingEnabled:  takeBoolPointer(d.Get("advanced").(bool)),
			DistributedRoutingEnabled:  takeBoolPointer(d.Get("distributed_routing").(bool)),
			GatewayInterfaces: &types.GatewayInterfaces{
				GatewayInterface: gwInterfaces,
			},
			EdgeGatewayServiceConfiguration: &types.GatewayFeatures{},
		},
	}

	if fipsModeEnabled, ok := d.GetOkExists("fips_mode_enabled"); ok {
		fipsModeEnabledBool := fipsModeEnabled.(bool)
		log.Printf("[TRACE] edge gateway creation. FIPS mode was set with value %t", fipsModeEnabledBool)
		egwConfiguration.Configuration.FipsModeEnabled = takeBoolPointer(fipsModeEnabledBool)
	}

	edge, err := govcd.CreateAndConfigureEdgeGateway(vcdClient.VCDClient, orgName, vdcName, egwName, egwConfiguration)
	if err != nil {
		log.Printf("[DEBUG] Error creating edge gateway: %s", err)
		return fmt.Errorf("error creating edge gateway: %s", err)
	}
	// Edge gateway creation succeeded therefore we save related fields now to preserve Id.
	// Edge gateway is already created even if further process fails
	log.Printf("[TRACE] flushing edge gateway creation fields")
	err = setEdgeGatewayValues(vcdClient, d, edge, "resource")
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
	log.Printf("[TRACE] edge gateway read initiated from origin %s", origin)

	vcdClient := meta.(*VCDClient)

	orgName := d.Get("org").(string)
	vdcName := d.Get("vdc").(string)
	_, vdc, err := vcdClient.GetOrgAndVdc(orgName, vdcName)
	if err != nil {
		return fmt.Errorf("[edgegateway read] error retrieving org and vdc: %s", err)
	}
	var edgeGateway *govcd.EdgeGateway
	var hasFilter bool
	var filter interface{}

	if origin == "datasource" {

		if !nameOrFilterIsSet(d) {
			return fmt.Errorf(noNameOrFilterError, "vcd_edgegateway")
		}
		filter, hasFilter = d.GetOk("filter")
		if hasFilter {
			edgeGateway, err = getEdgeGatewayByFilter(vdc, filter)
			if err != nil {
				return err
			}
		}
	}
	identifier := d.Id()
	if identifier == "" {
		identifier = d.Get("name").(string)
	}
	if identifier == "" && !hasFilter {
		return fmt.Errorf("[edgegateway read] no identifier provided")
	}
	if edgeGateway == nil {
		edgeGateway, err = vdc.GetEdgeGatewayByNameOrId(identifier, false)
	}
	if err != nil {
		if origin == "resource" {
			log.Printf("[edgegateway read] edge gateway %s not found. Removing from state file: %s", identifier, err)
			d.SetId("")
			return nil
		}
		return fmt.Errorf("[edgegateway read] error retrieving edge gateway %s: %s", identifier, err)
	}

	if err := setEdgeGatewayValues(vcdClient, d, *edgeGateway, origin); err != nil {
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

// getSimpleGatewayInterfaces aims to add compatibility layer to go-vcloud-director
// CreateEdgeGateway function which is a wrapper around CreateAndConfigureEdgeGateway. The layer
// resides here so that code can work together getGatewayInterfaces
func getSimpleGatewayInterfaces(vcdClient *VCDClient, externalNetworks []string, defaultGatewayNetwork string) ([]*types.GatewayInterface, error) {
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

// getGatewayInterfacesType extracts `external_network` blocks with more advanced settings into
// []*types.GatewayInterface
// This is a pretty complicated function with 3 level nesting as it is implied by API structure.
// This structure is documented below.
//
// <GatewayInterface>						<---- maps directly to `external_network` block
// 	<Name>test_external_network</Name>
// 	<DisplayName>test_external_network</DisplayName>
// 	<Network href="...." id="urn:vcloud:network:144bafa4-7cbe-44de-a647-e6e045b7b8c5" type="application/vnd.vmware.admin.network+xml" name="test_external_network"></Network>
// 	<InterfaceType>uplink</InterfaceType>
// 	<SubnetParticipation>					<----- maps to nested `subnet` block(s) inside `external_network`
// 		<Gateway>192.168.30.49</Gateway>
// 		<Netmask>255.255.255.240</Netmask>
// 		<IpAddress>192.168.30.51</IpAddress>
// 		<IpRanges>							<----- maps to `suballocate_pool` block(s) inside `subnet`
// 			<IpRange>
// 				<StartAddress>192.168.30.53</StartAddress>
// 				<EndAddress>192.168.30.55</EndAddress>
// 			</IpRange>
// 			<IpRange>
// 				<StartAddress>192.168.30.58</StartAddress>
// 				<EndAddress>192.168.30.60</EndAddress>
// 			</IpRange>
// 		</IpRanges>
// 		<UseForDefaultRoute>true</UseForDefaultRoute>
// 	</SubnetParticipation>
// 	<SubnetParticipation>				    <---- simple `subnet` block without suballocated pools and automatic IP assignment
// 		<Gateway>292.168.30.49</Gateway>
// 		<Netmask>255.255.255.240</Netmask>
// 		<UseForDefaultRoute>true</UseForDefaultRoute>
// 	</SubnetParticipation>
// 	<ApplyRateLimit>true</ApplyRateLimit>
// 	<InRateLimit>100</InRateLimit>
// 	<OutRateLimit>100</OutRateLimit>
// 	<UseForDefaultRoute>true</UseForDefaultRoute>
// </GatewayInterface>
func getGatewayInterfacesType(vcdClient *VCDClient, externalInterfaceSet *schema.Set) ([]*types.GatewayInterface, error) {
	var gatewayInterfaceSlice []*types.GatewayInterface

	extInterfaceList := externalInterfaceSet.List()
	if len(extInterfaceList) > 0 {
		gatewayInterfaceSlice = make([]*types.GatewayInterface, len(extInterfaceList))
		// Loop over interface definitions one by one and add them to list of interfaces
		for extInterfaceIndex, extInterface := range extInterfaceList {
			extInterfaceMap := extInterface.(map[string]interface{})
			externalNetworkName := extInterfaceMap["name"].(string)
			externalNetwork, err := vcdClient.GetExternalNetworkByName(externalNetworkName)
			if err != nil {
				return nil, fmt.Errorf("could not look up external network %s by name: %s", externalNetworkName, err)
			}
			var isInterfaceUsedForDefaultRoute bool
			var subnetParticipationSlice []*types.SubnetParticipation

			// Create subnet participation definitions for a particular edge gateway
			subnetSchema := extInterfaceMap["subnet"]
			subnetParticipationSlice, isInterfaceUsedForDefaultRoute, err = getGatewayInterfaceSubnetParticipationType(subnetSchema.(*schema.Set))
			if err != nil {
				return nil, fmt.Errorf("unable to create subnet participation definition: %s", err)
			}

			// Try to create a network interface and add it to the slice of gateway interfaces
			gwInterface, err := getGatewayInterfaceType(externalNetwork, extInterfaceMap, isInterfaceUsedForDefaultRoute, subnetParticipationSlice)
			if err != nil {
				return nil, fmt.Errorf("unable to create edge gateway interface definition: %s", err)
			}
			gatewayInterfaceSlice[extInterfaceIndex] = gwInterface
		}
	}

	return gatewayInterfaceSlice, nil
}

// getGatewayInterfaceType gets all interface properties and returns a definition for single
// *types.GatewayInterface
func getGatewayInterfaceType(externalNetwork *govcd.ExternalNetwork, extInterfaceMap map[string]interface{}, usedForDefaultRoute bool, subnetParticipationSlice []*types.SubnetParticipation) (*types.GatewayInterface, error) {
	singleGatewayInterface := &types.GatewayInterface{
		Name:          externalNetwork.ExternalNetwork.Name,
		DisplayName:   externalNetwork.ExternalNetwork.Name,
		InterfaceType: "uplink",
		Network: &types.Reference{
			HREF: externalNetwork.ExternalNetwork.HREF,
			ID:   externalNetwork.ExternalNetwork.ID,
			Type: "application/vnd.vmware.admin.network+xml",
			Name: externalNetwork.ExternalNetwork.Name,
		},
		UseForDefaultRoute:  usedForDefaultRoute,
		ApplyRateLimit:      extInterfaceMap["enable_rate_limit"].(bool),
		InRateLimit:         extInterfaceMap["incoming_rate_limit"].(float64),
		OutRateLimit:        extInterfaceMap["outgoing_rate_limit"].(float64),
		SubnetParticipation: subnetParticipationSlice,
	}

	return singleGatewayInterface, nil
}

// getGatewayInterfaceSubnetParticipationType creates []*types.SubnetParticipation for a particular
func getGatewayInterfaceSubnetParticipationType(subnetSet *schema.Set) ([]*types.SubnetParticipation, bool, error) {
	var isInterfaceUsedForDefaultRoute bool
	subnetList := subnetSet.List()
	subnetParticipationSlice := make([]*types.SubnetParticipation, len(subnetList))

	for subnetIndex, subnet := range subnetList {
		subnetMap := subnet.(map[string]interface{})
		subnetParticipationSlice[subnetIndex] = &types.SubnetParticipation{
			Gateway:            subnetMap["gateway"].(string),
			Netmask:            subnetMap["netmask"].(string),
			UseForDefaultRoute: subnetMap["use_for_default_route"].(bool),
		}
		subnetParticipationSlice[subnetIndex].IPAddress = subnetMap["ip_address"].(string)
		if subnetMap["use_for_default_route"].(bool) {
			isInterfaceUsedForDefaultRoute = true
		}

		// Check if there are any optional suballocated pool definitions (defined as IP ranges) and
		// parse them
		if subnetSchema, ok := subnetMap["suballocate_pool"]; ok {
			ranges, err := getGatewayInterfaceIpRangeType(subnetSchema.(*schema.Set))
			if err != nil {
				return nil, false, fmt.Errorf("unable to prepare sub-allocation pools :%s", err)
			}
			// if there are any ranges returned - add them to subnet
			if len(ranges) > 0 {
				subnetParticipationSlice[subnetIndex].IPRanges = &types.IPRanges{}
				subnetParticipationSlice[subnetIndex].IPRanges.IPRange = ranges
			}

		}
	}
	return subnetParticipationSlice, isInterfaceUsedForDefaultRoute, nil
}

// getGatewayInterfaceIpRangeType creates []*types.IPRange with IP address ranges which are shown as
// IP pool sub-allocations in UI
func getGatewayInterfaceIpRangeType(suballocatePoolSet *schema.Set) ([]*types.IPRange, error) {
	var ipRange []*types.IPRange

	suballocatePoolList := suballocatePoolSet.List()
	if len(suballocatePoolList) > 0 {
		// Allocate nested IP ranges slice size
		ipRange = make([]*types.IPRange, len(suballocatePoolList))

		for suballocatePoolIndex, suballocatePool := range suballocatePoolList {
			suballocatePoolMap := convertToStringMap(suballocatePool.(map[string]interface{}))
			singleRange := &types.IPRange{
				StartAddress: suballocatePoolMap["start_address"],
				EndAddress:   suballocatePoolMap["end_address"],
			}
			// Add single IP range into list
			ipRange[suballocatePoolIndex] = singleRange
		}
	}

	return ipRange, nil
}

// getExternalNetworkData traverses API structure over edge gateway interfaces to unpack such
// hierarchy to `external_network` block(s).
// external_network -> subnet(func getExternalNetworkSubnetTypeSet) ->
// suballocate_pool (func getExternalNetworkIPRangeSubAllocatePoolTypeSet)
// It must only convert to such structure only
// `uplink` interfaces. One uplink exception in the case of distributed routing support (DLR) is an
// `uplink` network `DLR_to_EDGE_%s` which is a transit interface
func getExternalNetworkData(vcdClient *VCDClient, d *schema.ResourceData, gatewayInterfaces []*types.GatewayInterface, origin string) (*schema.Set, error) {
	edgeGatewayName := d.Get("name").(string)
	isDistributedRouter := d.Get("distributed_routing").(bool)

	var externalNetworkSlice []interface{}
	if len(gatewayInterfaces) > 0 {
		for _, extNetwork := range gatewayInterfaces {
			// Only when InterfaceType == "uplink" this interface (vNic) is connected to external
			// network
			if extNetwork.InterfaceType != "uplink" {
				log.Printf("[TRACE] edge gateway - skipping read of network %s because it is not of type 'uplink' (%s)",
					extNetwork.Network.Name, extNetwork.InterfaceType)
				continue
			}
			// One of gateway interfaces can be named `DLR_to_EDGE_%s` where %s=edge_gateway_name
			// (e.g. `DLR_to_EDGE_edge-with-complex-networks`). This interface (vNic) is added as
			// transit interface when distributed logical routing (DLR) is enabled. It is not being
			// defined by user and we do not need to read/set it into `external_network` block.
			if isDistributedRouter && extNetwork.Network.Name == fmt.Sprintf("DLR_to_EDGE_%s", edgeGatewayName) {
				log.Printf("[TRACE] edge gateway - skipping read of uplink interface network %s because it is a DLR interface",
					extNetwork.Network.Name)
				continue
			}

			extNetworkMap := make(map[string]interface{})
			extNetworkMap["name"] = extNetwork.Network.Name

			extNetworkMap["enable_rate_limit"] = extNetwork.ApplyRateLimit
			extNetworkMap["incoming_rate_limit"] = extNetwork.InRateLimit
			extNetworkMap["outgoing_rate_limit"] = extNetwork.OutRateLimit

			if len(extNetwork.SubnetParticipation) > 0 {
				subnet, err := getExternalNetworkSubnetTypeSet(extNetwork.SubnetParticipation, vcdClient, d, extNetwork.Network.Name)
				if err != nil {
					return nil, fmt.Errorf("unable to create subnet structure for storing in statefile: %s", err)
				}
				extNetworkMap["subnet"] = subnet

				externalNetworkSlice = append(externalNetworkSlice, extNetworkMap)
			}
		}
	}
	externalNetworkSet := schema.NewSet(schema.HashResource(externalNetworkResource), externalNetworkSlice)

	return externalNetworkSet, nil
}

// getExternalNetworkSubnetTypeSet creates a Hashicorp TypeSet holding all subnets defined for
// external network (including sub-allocated pools)
func getExternalNetworkSubnetTypeSet(subnetParticipation []*types.SubnetParticipation, vcdClient *VCDClient, d *schema.ResourceData, extNetworkName string) (*schema.Set, error) {
	externalNetworkSubnetSlice := make([]interface{}, len(subnetParticipation))
	for extNetsubnetIndex, extNetsubnet := range subnetParticipation {
		extNetworkSubnetMap := make(map[string]interface{})
		extNetworkSubnetMap["gateway"] = extNetsubnet.Gateway
		extNetworkSubnetMap["netmask"] = extNetsubnet.Netmask
		extNetworkSubnetMap["use_for_default_route"] = extNetsubnet.UseForDefaultRoute

		// IP address is tricky. It is only possible to set the IP address during read if it
		// was actually a used field in .tf configuration. If it was not used, it cannot be
		// set because Terraform will recommend to rebuild whole block every time, because
		// `computed` field causes the hash function for TypeSet to be recomputed ("known
		// after apply") every time.
		wasIpSet, err := wasIpAddressSet(vcdClient, d, extNetworkName, extNetsubnet.Gateway, extNetsubnet.Netmask)
		if err != nil {
			return nil, fmt.Errorf("could not check if IP address was set in configuration: %s", err)
		}
		if wasIpSet {
			extNetworkSubnetMap["ip_address"] = extNetsubnet.IPAddress
		}

		// Check for suballocated ip pools and set them if there are any
		if extNetsubnet.IPRanges != nil && len(extNetsubnet.IPRanges.IPRange) > 0 {
			// Hash externalNetworkSubnetRangeSlice and add it to parent object "subnet"
			extNetworkSubnetMap["suballocate_pool"] = getExternalNetworkIPRangeSubAllocatePoolTypeSet(extNetsubnet.IPRanges.IPRange)
		}

		// Make a set and add it to externalNetworkSubnetSlice
		externalNetworkSubnetSlice[extNetsubnetIndex] = extNetworkSubnetMap
	}

	return schema.NewSet(schema.HashResource(subnetResource), externalNetworkSubnetSlice), nil
}

// getExternalNetworkIPRangeSubAllocatePoolTypeSet creates a Hashicorp TypeSet holding all IP ranges of sub-allocated pools
func getExternalNetworkIPRangeSubAllocatePoolTypeSet(ipRanges []*types.IPRange) *schema.Set {
	externalNetworkSubnetRangeSlice := make([]interface{}, len(ipRanges))
	for ipRangeIndex, ipRange := range ipRanges {
		extNetworkSubnetRangeMap := make(map[string]interface{})
		extNetworkSubnetRangeMap["start_address"] = ipRange.StartAddress
		extNetworkSubnetRangeMap["end_address"] = ipRange.EndAddress

		externalNetworkSubnetRangeSlice[ipRangeIndex] = extNetworkSubnetRangeMap
	}
	// Hash and return IP ranges of sub-allocated pools
	return schema.NewSet(schema.HashResource(subAllocationPool), externalNetworkSubnetRangeSlice)
}

// wasIpAddressSet checks if specific `SubnetParticipation` element had IP address field populated
// in .tf configuration. It is needed to decide whether a computed IP address field should be set,
// because TypeSet is very sensitive to injecting any new data (ends up with new hash and shows
// replacement of whole object)
func wasIpAddressSet(vcdClient *VCDClient, d *schema.ResourceData, extNetworkName, extNetworkGateway, extNetworkNetmask string) (bool, error) {
	// Cache currently set configuration
	gwInterfaces, err := getGatewayInterfacesType(vcdClient, d.Get("external_network").(*schema.Set))
	if err != nil {
		return false, fmt.Errorf("could not read current state configuration: %s", err)
	}

	for _, k := range gwInterfaces {
		if k.Network.Name == extNetworkName {
			for _, kk := range k.SubnetParticipation {
				if kk.Gateway == extNetworkGateway && kk.Netmask == extNetworkNetmask {
					if kk.IPAddress != "" {
						return true, nil
					}
				}
			}
		}
	}
	return false, nil
}

// Convenience function to fill edge gateway values from resource data
func setEdgeGatewayValues(vcdClient *VCDClient, d *schema.ResourceData, egw govcd.EdgeGateway, origin string) error {
	log.Printf("[TRACE] edge gateway read - setting values")
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

	_, simpleExternalNetworksSet := d.GetOk("external_networks")
	// When `external_networks` field was not used - we set a more rich `external_network` block
	// which allows to set multiple used subnets, manual IP addresses for IPs assigned to edge
	// gateway and which subnet should be used as the default one for edge gateway. Data source
	// always gets it populated.
	if !simpleExternalNetworksSet || origin == "datasource" {
		externalNetworkData, err := getExternalNetworkData(vcdClient, d, egw.EdgeGateway.Configuration.GatewayInterfaces.GatewayInterface, origin)
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
	if simpleExternalNetworksSet || origin == "datasource" {
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

	// Populate list of external_network_ip_addresses
	log.Printf("[TRACE] creating edge gateway using simple 'external_networks' and 'default_gateway_network' fields")
	var externalNets []interface{}
	for _, net := range egw.EdgeGateway.Configuration.GatewayInterfaces.GatewayInterface {
		if net.InterfaceType == "uplink" {
			for _, subnet := range net.SubnetParticipation {
				externalNets = append(externalNets, subnet.IPAddress)
			}
		}

	}
	err = d.Set("external_network_ips", externalNets)
	if err != nil {
		return fmt.Errorf("could not set external_network_ip_addresses field: %s", err)
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

	err = d.Set("distributed_routing", egw.EdgeGateway.Configuration.DistributedRoutingEnabled)
	if err != nil {
		return err
	}

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
