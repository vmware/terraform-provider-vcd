package vcd

import (
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/vmware/go-vcloud-director/v2/govcd"
)

func resourceVcdEdgeGatewaySettings() *schema.Resource {
	return &schema.Resource{
		Create: resourceVcdEdgeGatewaySettingsCreate,
		Read:   resourceVcdEdgeGatewaySettingsRead,
		Update: resourceVcdEdgeGatewaySettingsUpdate,
		Delete: resourceVcdEdgeGatewaySettingsDelete,
		Importer: &schema.ResourceImporter{
			State: resourceVcdEdgeGatewaySettingsImport,
		},
		Schema: map[string]*schema.Schema{
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
			"edge_gateway_name": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				Description:  "Name of the edge gateway. Required when 'edge_gateway_id' is not set",
				ExactlyOneOf: []string{"edge_gateway_id", "edge_gateway_name"},
			},
			"edge_gateway_id": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				Description:  "ID of the edge gateway. Required when 'edge_gateway_name' is not set",
				ExactlyOneOf: []string{"edge_gateway_id", "edge_gateway_name"},
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
				// Due to a vCD bug, this field can only be changed by a system administrator
			},
			"lb_loglevel": &schema.Schema{
				Type:         schema.TypeString,
				Default:      "info",
				Optional:     true,
				ValidateFunc: validateCase("lower"),
				Description: "Log level. One of 'emergency', 'alert', 'critical', 'error', " +
					"'warning', 'notice', 'info', 'debug'. ('info' by default)",
				// Due to a vCD bug, this field can only be changed by a system administrator
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

func resourceVcdEdgeGatewaySettingsCreate(d *schema.ResourceData, meta interface{}) error {
	return resourceVcdEdgeGatewaySettingsUpdate(d, meta)
}

func getVcdEdgeGateway(d *schema.ResourceData, meta interface{}) (*govcd.EdgeGateway, error) {

	log.Printf("[TRACE] edge gateway settings read initiated")

	vcdClient := meta.(*VCDClient)

	orgName := d.Get("org").(string)
	vdcName := d.Get("vdc").(string)
	_, vdc, err := vcdClient.GetOrgAndVdc(orgName, vdcName)
	if err != nil {
		return nil, fmt.Errorf("[getVcdEdgeGateway] error retrieving org and vdc: %s", err)
	}
	var edgeGateway *govcd.EdgeGateway

	// Preferred identification method is by ID
	identifier := d.Get("edge_gateway_id").(string)
	if identifier == "" {
		// Alternative method is by name
		identifier = d.Get("name").(string)
	}
	if identifier == "" {
		return nil, fmt.Errorf("[getVcdEdgeGateway] no identifier provided")
	}
	edgeGateway, err = vdc.GetEdgeGatewayByNameOrId(identifier, false)
	if err != nil {
		return nil, fmt.Errorf("[getVcdEdgeGateway] edge gateway %s not found: %s", identifier, err)
	}
	return edgeGateway, nil
}

func resourceVcdEdgeGatewaySettingsRead(d *schema.ResourceData, meta interface{}) error {
	edgeGateway, err := getVcdEdgeGateway(d, meta)
	if err != nil {
		log.Printf("[edgegateway settings read] edge gateway not found. Removing from state file: %s", err)
		d.SetId("")
		return nil
	}
	// Only advanced edge gateway can have the settings resource
	if !edgeGateway.HasAdvancedNetworking() {
		return fmt.Errorf("[edge gateway settings read] this resource is only available with advanced edge gateways")
	}
	if err := setLoadBalancerData(d, *edgeGateway); err != nil {
		return err
	}

	if err := setFirewallData(d, *edgeGateway); err != nil {
		return err
	}

	_ = d.Set("edge_gateway_id", edgeGateway.EdgeGateway.ID)
	_ = d.Set("edge_gateway_name", edgeGateway.EdgeGateway.Name)
	d.SetId(edgeGateway.EdgeGateway.ID)

	log.Printf("[TRACE] edge gateway settings read completed: %#v", edgeGateway.EdgeGateway)
	return nil
}

func resourceVcdEdgeGatewaySettingsUpdate(d *schema.ResourceData, meta interface{}) error {
	edgeGateway, err := getVcdEdgeGateway(d, meta)
	if err != nil {
		return err
	}

	if !edgeGateway.HasAdvancedNetworking() {
		return fmt.Errorf("[edge gateway settings update] this resource is only available with advanced edge gateways")
	}

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

	log.Printf("[TRACE] edge gateway settings update completed: %#v", edgeGateway.EdgeGateway)
	return resourceVcdEdgeGatewaySettingsRead(d, meta)
}

func resourceVcdEdgeGatewaySettingsDelete(d *schema.ResourceData, meta interface{}) error {
	return resourceVcdEdgeGatewaySettingsUpdate(d, meta)
}

// resourceVcdEdgeGatewaySettingsImport is responsible for importing the resource.
// The following steps happen as part of import
// 1. The user supplies `terraform import _resource_name_ _the_id_string_` command
// 2. `_the_id_string_` contains a dot formatted path to resource as in the example below
// 3. The functions splits the dot-formatted path and tries to lookup the object
// 4. If the lookup succeeds it sets the ID field for `_resource_name_` resource in statefile
// (the resource must be already defined in .tf config otherwise `terraform import` will complain)
// 5. `terraform refresh` is being implicitly launched. The Read method looks up all other fields
// based on the known ID of object.
//
// Example resource name (_resource_name_): vcd_edgegateway_settings.my-edge-gateway-name
// Example import path (_the_id_string_): org.vdc.my-edge-gw
// Note: the separator can be changed using Provider.import_separator or variable VCD_IMPORT_SEPARATOR
func resourceVcdEdgeGatewaySettingsImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	resourceURI := strings.Split(d.Id(), ImportSeparator)
	if len(resourceURI) != 3 {
		return nil, fmt.Errorf("[resourceVcdEdgeGatewaySettingsImport] resource name must be specified as org-name.vdc-name.edge-gw-name")
	}
	orgName, vdcName, edgeName := resourceURI[0], resourceURI[1], resourceURI[2]

	vcdClient := meta.(*VCDClient)
	edgeGateway, err := vcdClient.GetEdgeGateway(orgName, vdcName, edgeName)
	if err != nil {
		return nil, fmt.Errorf(errorUnableToFindEdgeGateway, err)
	}

	_ = d.Set("org", orgName)
	_ = d.Set("vdc", vdcName)
	_ = d.Set("edge_gateway_name", edgeGateway.EdgeGateway.Name)
	_ = d.Set("edge_gateway_id", edgeGateway.EdgeGateway.ID)
	d.SetId(edgeGateway.EdgeGateway.ID)
	return []*schema.ResourceData{d}, nil
}
