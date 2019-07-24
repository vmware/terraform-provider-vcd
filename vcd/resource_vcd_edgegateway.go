package vcd

import (
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
)

func resourceVcdEdgeGateway() *schema.Resource {
	return &schema.Resource{
		Create: resourceVcdEdgeGatewayCreate,
		Read:   resourceVcdEdgeGatewayRead,
		Update: resourceVcdEdgeGatewayUpdate,
		Delete: resourceVcdEdgeGatewayDelete,

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
			},
			"vdc": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
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
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				ForceNew:    true,
				Description: "If advanced networking enabled, also enable distributed routing",
			},
			"lb_enabled": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Enable load balancing",
			},
			"lb_acceleration_enabled": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Enable load balancer acceleration",
			},
			"lb_logging_enabled": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Enable load balancer logging",
			},
			"lb_loglevel": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validateCase("lower"),
				Description: "Log level. One of 'emergency', 'alert', 'critical', 'error', " +
					"'warning', 'notice', 'info', 'debug'",
			},
		},
	}
}

// Creates a new edge gateway from a resource definition
func resourceVcdEdgeGatewayCreate(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[TRACE] edge gateway creation initiated")

	// We use partial mode here because edge gateway creation and configuration consists of two
	// parts (API calls) - creating edge gateway and configuring load balancer. If the second part
	// fails we still want to persist state for the first part
	d.Partial(true)

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
	if (org == govcd.Org{}) || org.Org.HREF == "" || org.Org.ID == "" || org.Org.Name == "" {
		return fmt.Errorf("no valid Organization named '%s' was found", orgName)
	}
	if (vdc == govcd.Vdc{}) || vdc.Vdc.HREF == "" || vdc.Vdc.ID == "" || vdc.Vdc.Name == "" {
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
	// Edge gateway creation succeeded therefore we can flush related fields now. Edge
	// gateway is already created even if further process fails.
	log.Printf("[TRACE] flushing partial edge gateway creation fields")
	err = setEdgeGatewayValues(d, edge)
	setPartialEdgeGatewayValues(d)
	if err != nil {
		return err
	}

	// Only perform general load balancer configuration if settings are set
	if isEdgeGatewayLbConfigured(d) {
		if !d.Get("advanced").(bool) {
			return fmt.Errorf("load balancing cannot be used when advanced networking is disabled")
		}

		log.Printf("[TRACE] edge gateway load balancer configuration started")

		err := updateLoadBalancer(d, edge)
		if err != nil {
			return fmt.Errorf("unable to update general load balancer settings: %s", err)
		}

		// Load balancer configuration succeeded therefore we can flush related fields now
		log.Printf("[TRACE] flushing partial edge gateway load balancer configuration")
		err = setLoadBalancerData(d, edge)
		setPartialLoadBalancerData(d)
		if err != nil {
			return err
		}

		log.Printf("[TRACE] edge gateway load balancer configured")
	}

	// We succeeded in all steps, disabling partial mode. This causes Terraform to save all fields again.
	d.Partial(false)

	d.SetId(edge.EdgeGateway.ID)
	log.Printf("[TRACE] edge gateway created: %#v", edge.EdgeGateway.Name)
	return resourceVcdEdgeGatewayRead(d, meta)
}

// Fetches information about an existing edge gateway for a data definition
func resourceVcdEdgeGatewayRead(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[TRACE] edge gateway read initiated")

	vcdClient := meta.(*VCDClient)

	edgeGateway, err := vcdClient.GetEdgeGatewayFromResource(d, "name")
	if err != nil {
		d.SetId("")
		return nil
	}

	if err := setEdgeGatewayValues(d, edgeGateway); err != nil {
		return err
	}

	// Only read and set the statefile if the edge gateway is advanced
	// and general lb settings are used
	if isEdgeGatewayLbConfigured(d) && edgeGateway.HasAdvancedNetworking() {
		if err := setLoadBalancerData(d, edgeGateway); err != nil {
			return err
		}
	}

	log.Printf("[TRACE] edge gateway read completed: %#v", edgeGateway.EdgeGateway)
	return nil
}

// resourceVcdEdgeGatewayUpdate updates an edge gateway from a resource definition
func resourceVcdEdgeGatewayUpdate(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)
	vcdClient.lockEdgeGateway(d)
	defer vcdClient.unlockEdgeGateway(d)

	edgeGateway, err := vcdClient.GetEdgeGatewayFromResource(d, "name")
	if err != nil {
		return nil
	}

	if d.HasChange("lb_enabled") || d.HasChange("lb_acceleration_enabled") ||
		d.HasChange("lb_logging_enabled") || d.HasChange("lb_loglevel") {
		err := updateLoadBalancer(d, edgeGateway)
		if err != nil {
			return err
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
		d.SetId("")
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
	var networks []string
	for _, net := range egw.EdgeGateway.Configuration.GatewayInterfaces.GatewayInterface {
		if net.InterfaceType == "uplink" {
			networks = append(networks, net.Network.Name)
		}
	}
	err = d.Set("external_networks", networks)
	if err != nil {
		return err
	}

	err = d.Set("advanced", egw.EdgeGateway.Configuration.AdvancedNetworkingEnabled)
	if err != nil {
		return err
	}
	err = d.Set("ha_enabled", egw.EdgeGateway.Configuration.HaEnabled)
	if err != nil {
		return err
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

	return nil
}

// setPartialEdgeGatewayValues uses `d.SetPartial()` to flush edge gateway configuration in partial mode.
// It flushes all values which are set in `setEdgeGatewayValues`
func setPartialEdgeGatewayValues(d *schema.ResourceData) {
	d.SetPartial("id")
	d.SetPartial("name")
	d.SetPartial("description")
	d.SetPartial("configuration")
	d.SetPartial("external_networks")
	d.SetPartial("advanced")
	d.SetPartial("ha_enabled")
}

// setLoadBalancerData is a convenience function to handle load balancer settings on edge gateway
func setLoadBalancerData(d *schema.ResourceData, egw govcd.EdgeGateway) error {
	lb, err := egw.GetLBGeneralParams()
	if err != nil {
		return fmt.Errorf("unable to read general load balancer settings: %s", err)
	}

	d.Set("lb_enabled", lb.Enabled)
	d.Set("lb_acceleration_enabled", lb.AccelerationEnabled)
	d.Set("lb_logging_enabled", lb.Logging.Enable)
	d.Set("lb_loglevel", lb.Logging.LogLevel)

	return nil
}

// setPartialLoadBalancerData uses `d.SetPartial()` to flush edge gateway configuration in partial mode.
// It flushes all values which are set in `setLoadBalancerData`
func setPartialLoadBalancerData(d *schema.ResourceData) {
	d.SetPartial("lb_enabled")
	d.SetPartial("lb_acceleration_enabled")
	d.SetPartial("lb_logging_enabled")
	d.SetPartial("lb_loglevel")
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

// isEdgeGatewayLbConfigured checks if any of load balancer related settings are set
func isEdgeGatewayLbConfigured(d *schema.ResourceData) bool {
	_, existsLbEnabled := d.GetOk("lb_enabled")
	_, existsLbAccelerationEnabled := d.GetOk("lb_acceleration_enabled")
	_, existsLbLoggingEnabled := d.GetOk("lb_logging_enabled")
	_, existsLbLogLevel := d.GetOk("lb_loglevel")

	return existsLbEnabled || existsLbAccelerationEnabled || existsLbLoggingEnabled || existsLbLogLevel
}
