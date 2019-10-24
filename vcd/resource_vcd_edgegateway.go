package vcd

import (
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	"github.com/vmware/go-vcloud-director/v2/govcd"
)

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
				Type:        schema.TypeList,
				Required:    true,
				ForceNew:    true,
				Description: "A list of external networks to be used by the edge gateway",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"default_gateway_network": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				ForceNew:    true,
				Description: "External network to be used as default gateway. Its name must be included in 'external_networks'. An empty value will skip the default gateway",
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
		},
	}
}

// Creates a new edge gateway from a resource definition
func resourceVcdEdgeGatewayCreate(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[TRACE] edge gateway creation initiated")

	vcdClient := meta.(*VCDClient)

	rawExternalNetworks := d.Get("external_networks").([]interface{})
	var externalNetworks []string
	for _, en := range rawExternalNetworks {
		externalNetworks = append(externalNetworks, en.(string))
	}

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

	var gwCreation = govcd.EdgeGatewayCreation{
		ExternalNetworks:          externalNetworks,
		Name:                      d.Get("name").(string),
		OrgName:                   orgName,
		VdcName:                   vdcName,
		Description:               d.Get("description").(string),
		BackingConfiguration:      d.Get("configuration").(string),
		AdvancedNetworkingEnabled: d.Get("advanced").(bool),
		DefaultGateway:            d.Get("default_gateway_network").(string),
		DistributedRoutingEnabled: d.Get("distributed_routing").(bool),
		HAEnabled:                 d.Get("ha_enabled").(bool),
	}

	// In version 9.7+ the advanced property is true by default
	if vcdClient.APIVCDMaxVersionIs(">= 32.0") {
		if !gwCreation.AdvancedNetworkingEnabled {
			return fmt.Errorf("'advanced' property for vCD 9.7+ must be set to 'true'")
		}
	}

	edge, err := govcd.CreateEdgeGateway(vcdClient.VCDClient, gwCreation)
	if err != nil {
		log.Printf("[DEBUG] Error creating edge gateway: %#v", err)
		return fmt.Errorf("error creating edge gateway: %#v", err)
	}
	// Edge gateway creation succeeded therefore we save related fields now to preserve Id.
	// Edge gateway is already created even if further process fails
	log.Printf("[TRACE] flushing edge gateway creation fields")
	err = setEdgeGatewayValues(d, edge)
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

	if err := setEdgeGatewayValues(d, *edgeGateway); err != nil {
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

// Convenience function to fill edge gateway values from resource data
func setEdgeGatewayValues(d *schema.ResourceData, egw govcd.EdgeGateway) error {

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
	var gateways = make(map[string]string)
	var networks []string
	for _, net := range egw.EdgeGateway.Configuration.GatewayInterfaces.GatewayInterface {
		if net.InterfaceType == "uplink" {
			networks = append(networks, net.Network.Name)
			gateways[net.SubnetParticipation.Gateway] = net.Network.Name
		}
	}
	err = d.Set("external_networks", networks)
	if err != nil {
		return err
	}

	_ = d.Set("advanced", egw.EdgeGateway.Configuration.AdvancedNetworkingEnabled)
	_ = d.Set("ha_enabled", egw.EdgeGateway.Configuration.HaEnabled)

	for _, gw := range egw.EdgeGateway.Configuration.GatewayInterfaces.GatewayInterface {
		if gw.SubnetParticipation == nil || gw.SubnetParticipation.Gateway == "" {
			log.Printf("[DEBUG] [setEdgeGatewayValues] gateway %s is missing SubnetParticipation elements: %+#v",
				egw.EdgeGateway.Name, gw)
			return fmt.Errorf("[setEdgeGatewayValues] gateway %s is missing SubnetParticipation elements", egw.EdgeGateway.Name)
		}
		defaultGwNet, ok := gateways[gw.SubnetParticipation.Gateway]
		if ok {
			_ = d.Set("default_gateway_network", defaultGwNet)
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

	if egw.HasAdvancedNetworking() {
		err = setLoadBalancerData(d, egw)
		if err != nil {
			return err
		}

		err = setFirewallData(d, egw)
		if err != nil {
			return err
		}
	}

	d.SetId(egw.EdgeGateway.ID)
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
