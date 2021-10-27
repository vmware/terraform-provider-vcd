package vcd

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func resourceVcdLBVirtualServer() *schema.Resource {
	return &schema.Resource{
		Create: resourceVcdLBVirtualServerCreate,
		Read:   resourceVcdLBVirtualServerRead,
		Update: resourceVcdLBVirtualServerUpdate,
		Delete: resourceVcdLBVirtualServerDelete,
		Importer: &schema.ResourceImporter{
			State: resourceVcdLBVirtualServerImport,
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
				Description: "Edge gateway name in which the LB Virtual Server is located",
			},
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Unique Virtual Server name",
			},
			"description": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Virtual Server description",
			},
			"enabled": &schema.Schema{
				Type:        schema.TypeBool,
				Default:     true,
				Optional:    true,
				Description: "Defines if the virtual server is enabled",
			},
			"ip_address": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "IP address that the load balancer listens on",
				ValidateFunc: validation.IsIPAddress,
			},
			"protocol": &schema.Schema{
				Type:         schema.TypeString,
				Required:     true,
				Description:  "Protocol that the virtual server accepts",
				ValidateFunc: validateCase("lower"),
			},
			"port": &schema.Schema{
				Type:        schema.TypeInt,
				Required:    true,
				Description: "Port number that the load balancer listens on",
			},
			"enable_acceleration": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Enable virtual server acceleration",
			},
			"connection_limit": &schema.Schema{
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Maximum concurrent connections that the virtual server can process",
			},
			"connection_rate_limit": &schema.Schema{
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Maximum incoming new connection requests per second",
			},
			"app_profile_id": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Application profile ID to be associated with the virtual server",
			},
			"server_pool_id": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The server pool that the load balancer will use",
			},
			"app_rule_ids": &schema.Schema{
				Type:        schema.TypeList,
				Optional:    true,
				Description: "List of attached application rule IDs",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceVcdLBVirtualServerCreate(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)
	vcdClient.lockParentEdgeGtw(d)
	defer vcdClient.unLockParentEdgeGtw(d)

	edgeGateway, err := vcdClient.GetEdgeGatewayFromResource(d, "edge_gateway")
	if err != nil {
		return fmt.Errorf(errorUnableToFindEdgeGateway, err)
	}

	lBVirtualServer, err := getLBVirtualServerType(d)
	if err != nil {
		return fmt.Errorf("unable to make load balancer virtual server query: %s", err)
	}

	createdVirtualServer, err := edgeGateway.CreateLbVirtualServer(lBVirtualServer)
	if err != nil {
		return fmt.Errorf("error creating new load balancer virtual server: %s", err)
	}

	d.SetId(createdVirtualServer.ID)
	return resourceVcdLBVirtualServerRead(d, meta)
}

func resourceVcdLBVirtualServerRead(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)

	edgeGateway, err := vcdClient.GetEdgeGatewayFromResource(d, "edge_gateway")
	if err != nil {
		return fmt.Errorf(errorUnableToFindEdgeGateway, err)
	}

	readVirtualServer, err := edgeGateway.GetLbVirtualServerById(d.Id())
	if err != nil {
		d.SetId("")
		return fmt.Errorf("unable to find load balancer virtual server with ID %s: %s", d.Id(), err)
	}

	return setlBVirtualServerData(d, readVirtualServer)
}

func resourceVcdLBVirtualServerUpdate(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)
	vcdClient.lockParentEdgeGtw(d)
	defer vcdClient.unLockParentEdgeGtw(d)

	edgeGateway, err := vcdClient.GetEdgeGatewayFromResource(d, "edge_gateway")
	if err != nil {
		return fmt.Errorf(errorUnableToFindEdgeGateway, err)
	}

	updateVirtualServerConfig, err := getLBVirtualServerType(d)
	updateVirtualServerConfig.ID = d.Id() // We already know an ID for update and it allows to change name

	if err != nil {
		return fmt.Errorf("could not create load balancer virtual server type for update: %s", err)
	}

	updatedVirtualServer, err := edgeGateway.UpdateLbVirtualServer(updateVirtualServerConfig)
	if err != nil {
		return fmt.Errorf("unable to update load balancer virtual server with ID %s: %s", d.Id(), err)
	}

	return setlBVirtualServerData(d, updatedVirtualServer)
}

func resourceVcdLBVirtualServerDelete(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)
	vcdClient.lockParentEdgeGtw(d)
	defer vcdClient.unLockParentEdgeGtw(d)

	edgeGateway, err := vcdClient.GetEdgeGatewayFromResource(d, "edge_gateway")
	if err != nil {
		return fmt.Errorf(errorUnableToFindEdgeGateway, err)
	}

	err = edgeGateway.DeleteLbVirtualServerById(d.Id())
	if err != nil {
		return fmt.Errorf("error deleting load balancer virtual server: %s", err)
	}

	d.SetId("")
	return nil
}

// resourceVcdLBVirtualServerImport is responsible for importing the resource.
// The following steps happen as part of import
// 1. The user supplies `terraform import _resource_name_ _the_id_string_` command
// 2. `_the_id_string_` contains a dot formatted path to resource as in the example below
// 3. The functions splits the dot-formatted path and tries to lookup the object
// 4. If the lookup succeeds it sets the ID field for `_resource_name_` resource in statefile
// (the resource must be already defined in .tf config otherwise `terraform import` will complain)
// 5. `terraform refresh` is being implicitly launched. The Read method looks up all other fields
// based on the known ID of object.
//
// Example resource name (_resource_name_): vcd_lb_virtual_server.my-test-virtual-server
// Example import path (_the_id_string_): org.vdc.edge-gw.existing-virtual-server
// Note: the separator can be changed using Provider.import_separator or variable VCD_IMPORT_SEPARATOR
func resourceVcdLBVirtualServerImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	resourceURI := strings.Split(d.Id(), ImportSeparator)
	if len(resourceURI) != 4 {
		return nil, fmt.Errorf("resource name must be specified as org.vdc.edge-gw.lb-virtual-server")
	}
	orgName, vdcName, edgeName, virtualServerName := resourceURI[0], resourceURI[1], resourceURI[2], resourceURI[3]

	vcdClient := meta.(*VCDClient)
	edgeGateway, err := vcdClient.GetEdgeGateway(orgName, vdcName, edgeName)
	if err != nil {
		return nil, fmt.Errorf(errorUnableToFindEdgeGateway, err)
	}

	readVirtualServer, err := edgeGateway.GetLbVirtualServerByName(virtualServerName)
	if err != nil {
		return []*schema.ResourceData{}, fmt.Errorf("unable to find load balancer virtual server with name %s: %s",
			d.Id(), err)
	}

	dSet(d, "org", orgName)
	dSet(d, "vdc", vdcName)
	dSet(d, "edge_gateway", edgeName)
	dSet(d, "name", virtualServerName)

	d.SetId(readVirtualServer.ID)
	return []*schema.ResourceData{d}, nil
}

// getLBVirtualServerType converts schema.ResourceData to *types.LbVirtualServer and is useful
// for creating API requests
func getLBVirtualServerType(d *schema.ResourceData) (*types.LbVirtualServer, error) {
	lbVirtualServer := &types.LbVirtualServer{
		Name:                 d.Get("name").(string),
		Description:          d.Get("description").(string),
		Enabled:              d.Get("enabled").(bool),
		IpAddress:            d.Get("ip_address").(string),
		Protocol:             d.Get("protocol").(string),
		Port:                 d.Get("port").(int),
		AccelerationEnabled:  d.Get("enable_acceleration").(bool),
		ConnectionLimit:      d.Get("connection_limit").(int),
		ConnectionRateLimit:  d.Get("connection_rate_limit").(int),
		ApplicationProfileId: d.Get("app_profile_id").(string),
		DefaultPoolId:        d.Get("server_pool_id").(string),
	}

	// convert list of app rule ids to slice of strings
	var appRuleIds []string
	rules := d.Get("app_rule_ids").([]interface{})
	for _, rule := range rules {
		ruleString := rule.(string)
		appRuleIds = append(appRuleIds, ruleString)
	}

	lbVirtualServer.ApplicationRuleIds = appRuleIds

	return lbVirtualServer, nil
}

// setlBVirtualServerData sets object state from *types.LbVirtualServer
func setlBVirtualServerData(d *schema.ResourceData, lBVirtualServer *types.LbVirtualServer) error {
	dSet(d, "name", lBVirtualServer.Name)
	dSet(d, "description", lBVirtualServer.Description)
	dSet(d, "enabled", lBVirtualServer.Enabled)
	dSet(d, "ip_address", lBVirtualServer.IpAddress)
	dSet(d, "protocol", lBVirtualServer.Protocol)
	dSet(d, "port", lBVirtualServer.Port)
	dSet(d, "enable_acceleration", lBVirtualServer.AccelerationEnabled)
	dSet(d, "connection_limit", lBVirtualServer.ConnectionLimit)
	dSet(d, "connection_rate_limit", lBVirtualServer.ConnectionRateLimit)
	dSet(d, "app_profile_id", lBVirtualServer.ApplicationProfileId)
	dSet(d, "server_pool_id", lBVirtualServer.DefaultPoolId)
	dSet(d, "app_rule_ids", lBVirtualServer.ApplicationRuleIds)

	return nil
}
