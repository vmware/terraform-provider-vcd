package vcd

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func resourceVcdLBServerPool() *schema.Resource {
	return &schema.Resource{
		Create: resourceVcdLBServerPoolCreate,
		Read:   resourceVcdLBServerPoolRead,
		Update: resourceVcdLBServerPoolUpdate,
		Delete: resourceVcdLBServerPoolDelete,
		Importer: &schema.ResourceImporter{
			State: resourceVcdLBServerPoolImport,
		},

		Schema: map[string]*schema.Schema{
			"org": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "vCD organization in which the LB Server Pool is located",
			},
			"vdc": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "vCD virtual datacenter in which the LB Server Pool is located",
			},
			"edge_gateway": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Edge gateway name in which the LB Server Pool is located",
			},
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Unique LB Server Pool name",
			},
			"description": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Server pool description",
			},
			"algorithm": &schema.Schema{
				Type:         schema.TypeString,
				Required:     true,
				Description:  "Balancing method for the service. One of 'ip-hash', 'round-robin', 'uri', 'leastconn', 'url', or 'httpheader'",
				ValidateFunc: validateCase("lower"),
			},
			"algorithm_parameters": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Additional options for load balancing algorithm for httpheader or url algorithms",
			},
			"monitor_id": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Load Balancer Service Monitor Id",
			},
			"enable_transparency": &schema.Schema{
				Type:        schema.TypeBool,
				Default:     false,
				Optional:    true,
				Description: "Makes client IP addresses visible to the backend servers",
			},
			"member": {
				Optional: true,
				ForceNew: false,
				Type:     schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							ForceNew:    false,
							Computed:    true,
							Type:        schema.TypeString,
							Description: "Pool member id (formatted as member-xx, where xx is a number)",
						},
						"condition": &schema.Schema{
							Type:         schema.TypeString,
							ForceNew:     false,
							Required:     true,
							ValidateFunc: validateCase("lower"),
							Description:  "Defines member state. One of enabled, drain, disabled.",
						},
						"name": {
							Required:    true,
							ForceNew:    false,
							Type:        schema.TypeString,
							Description: "Name of pool member",
						},
						"ip_address": {
							Required:    true,
							ForceNew:    false,
							Type:        schema.TypeString,
							Description: "IP address of member in server pool",
						},
						"port": {
							Required:    true,
							ForceNew:    false,
							Type:        schema.TypeInt,
							Description: "Port at which the member is to receive traffic from the load balancer",
						},
						"monitor_port": {
							Required:    true,
							ForceNew:    false,
							Type:        schema.TypeInt,
							Description: "Port at which the member is to receive health monitor requests. Can be the same as port",
						},
						"weight": {
							Required:    true,
							ForceNew:    false,
							Type:        schema.TypeInt,
							Description: "Proportion of traffic this member is to handle. Must be an integer in the range 1-256",
						},
						"min_connections": {
							Optional:    true,
							ForceNew:    false,
							Type:        schema.TypeInt,
							Description: "Minimum number of concurrent connections a member must always accept",
						},
						"max_connections": {
							Optional: true,
							ForceNew: false,
							Type:     schema.TypeInt,
							Description: "The maximum number of concurrent connections the member can handle. If exceeded " +
								"requests are queued and the load balancer waits for a connection to be released",
						},
					},
				},
			},
		},
	}
}

func resourceVcdLBServerPoolCreate(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)
	vcdClient.lockParentEdgeGtw(d)
	defer vcdClient.unLockParentEdgeGtw(d)

	edgeGateway, err := vcdClient.GetEdgeGatewayFromResource(d, "edge_gateway")
	if err != nil {
		return fmt.Errorf(errorUnableToFindEdgeGateway, err)
	}

	LBPool, err := getLBPoolType(d)
	if err != nil {
		return fmt.Errorf("unable to create load balancer server pool type: %s", err)
	}

	createdPool, err := edgeGateway.CreateLBServerPool(LBPool)
	if err != nil {
		return fmt.Errorf("error creating new load balancer server pool: %s", err)
	}

	// We store the values once again because response includes pool member IDs
	if err := setLBPoolData(d, createdPool); err != nil {
		return err
	}
	d.SetId(createdPool.Id)
	return resourceVcdLBServerPoolRead(d, meta)
}

func resourceVcdLBServerPoolRead(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)

	edgeGateway, err := vcdClient.GetEdgeGatewayFromResource(d, "edge_gateway")
	if err != nil {
		return fmt.Errorf(errorUnableToFindEdgeGateway, err)
	}

	readLBPool, err := edgeGateway.ReadLBServerPoolByID(d.Id())
	if err != nil {
		d.SetId("")
		return fmt.Errorf("unable to find load balancer server pool with Id %s: %s", d.Id(), err)
	}

	return setLBPoolData(d, readLBPool)
}

func resourceVcdLBServerPoolUpdate(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)
	vcdClient.lockParentEdgeGtw(d)
	defer vcdClient.unLockParentEdgeGtw(d)

	edgeGateway, err := vcdClient.GetEdgeGatewayFromResource(d, "edge_gateway")
	if err != nil {
		return fmt.Errorf(errorUnableToFindEdgeGateway, err)
	}

	updateLBPoolConfig, err := getLBPoolType(d)
	if err != nil {
		return fmt.Errorf("could not create load balancer server pool type for update: %s", err)
	}

	updatedLBPool, err := edgeGateway.UpdateLBServerPool(updateLBPoolConfig)
	if err != nil {
		return fmt.Errorf("unable to update load balancer server pool with Id %s: %s", d.Id(), err)
	}

	return setLBPoolData(d, updatedLBPool)
}

func resourceVcdLBServerPoolDelete(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)
	vcdClient.lockParentEdgeGtw(d)
	defer vcdClient.unLockParentEdgeGtw(d)

	edgeGateway, err := vcdClient.GetEdgeGatewayFromResource(d, "edge_gateway")
	if err != nil {
		return fmt.Errorf(errorUnableToFindEdgeGateway, err)
	}

	err = edgeGateway.DeleteLBServerPoolByID(d.Id())
	if err != nil {
		return fmt.Errorf("error deleting load balancer server pool: %s", err)
	}

	d.SetId("")
	return nil
}

// resourceVcdLBServerPoolImport is responsible for importing the resource.
// The d.Id() field as being passed from `terraform import _resource_name_ _the_id_string_ requires
// a name based dot-formatted path to the object to lookup the object and sets the id of object.
// `terraform import` automatically performs `refresh` operation which loads up all other fields.
//
// Example import path (id): org.vdc.edge-gw.lb-server-pool
func resourceVcdLBServerPoolImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	resourceURI := strings.Split(d.Id(), ".")
	if len(resourceURI) != 4 {
		return nil, fmt.Errorf("resource name must be specified as org.vdc.edge-gw.lb-server-pool")
	}
	orgName, vdcName, edgeName, poolName := resourceURI[0], resourceURI[1], resourceURI[2], resourceURI[3]

	vcdClient := meta.(*VCDClient)
	edgeGateway, err := vcdClient.GetEdgeGateway(orgName, vdcName, edgeName)
	if err != nil {
		return nil, fmt.Errorf(errorUnableToFindEdgeGateway, err)
	}

	readLBPool, err := edgeGateway.ReadLBServerPoolByName(poolName)
	if err != nil {
		return []*schema.ResourceData{}, fmt.Errorf("unable to find load balancer server pool with name %s: %s", d.Id(), err)
	}

	d.Set("org", orgName)
	d.Set("vdc", vdcName)
	d.Set("edge_gateway", edgeName)
	d.Set("name", poolName)

	d.SetId(readLBPool.Id)
	return []*schema.ResourceData{d}, nil
}

// getLBPoolType converts schema.ResourceData to *types.LbPool and is useful
// for creating API requests
func getLBPoolType(d *schema.ResourceData) (*types.LbPool, error) {
	lbPool := &types.LbPool{
		Name:                d.Get("name").(string),
		Description:         d.Get("description").(string),
		Algorithm:           d.Get("algorithm").(string),
		MonitorId:           d.Get("monitor_id").(string),
		Transparent:         d.Get("enable_transparency").(bool),
		AlgorithmParameters: d.Get("algorithm_parameters").(string),
	}

	members, err := getLBPoolMembersType(d)
	if err != nil {
		return nil, err
	}
	lbPool.Members = members

	return lbPool, nil
}

// getLBPoolMembersType converts schema.ResourceData to *types.LbPoolMembers and is useful
// for creating API requests
func getLBPoolMembersType(d *schema.ResourceData) (types.LbPoolMembers, error) {
	var lbPoolMembers types.LbPoolMembers

	members := d.Get("member").([]interface{})
	for _, memberInterface := range members {
		var memberConfig types.LbPoolMember
		member := memberInterface.(map[string]interface{})

		// If we have IDs - then we must insert them for update. Otherwise the update may get mixed
		if member["id"].(string) != "" {
			memberConfig.Id = member["id"].(string)
		}

		memberConfig.Name = member["name"].(string)
		memberConfig.IpAddress = member["ip_address"].(string)
		memberConfig.Port = member["port"].(int)
		memberConfig.MonitorPort = member["monitor_port"].(int)
		memberConfig.Weight = member["weight"].(int)
		memberConfig.MinConn = member["min_connections"].(int)
		memberConfig.MaxConn = member["max_connections"].(int)
		memberConfig.Weight = member["weight"].(int)
		memberConfig.Condition = member["condition"].(string)

		lbPoolMembers = append(lbPoolMembers, memberConfig)
	}

	return lbPoolMembers, nil
}

// setLBPoolData sets object state from *types.LbPool
func setLBPoolData(d *schema.ResourceData, lBpool *types.LbPool) error {
	d.Set("name", lBpool.Name)
	d.Set("description", lBpool.Description)
	d.Set("algorithm", lBpool.Algorithm)
	// Optional attributes may not be necessary
	d.Set("monitor_id", lBpool.MonitorId)
	d.Set("enable_transparency", lBpool.Transparent)
	d.Set("algorithm_parameters", lBpool.AlgorithmParameters)

	return setLBPoolMembersData(d, lBpool.Members)
}

// setLBPoolMembersData sets pool members state from *types.LbPoolMembers
func setLBPoolMembersData(d *schema.ResourceData, lBpoolMembers types.LbPoolMembers) error {

	memberSet := make([]map[string]interface{}, len(lBpoolMembers))
	for index, member := range lBpoolMembers {
		oneMember := make(map[string]interface{})

		oneMember["condition"] = member.Condition
		oneMember["name"] = member.Name
		oneMember["ip_address"] = member.IpAddress
		oneMember["port"] = member.Port
		oneMember["monitor_port"] = member.MonitorPort
		oneMember["weight"] = member.Weight
		oneMember["min_connections"] = member.MinConn
		oneMember["max_connections"] = member.MaxConn
		oneMember["id"] = member.Id

		memberSet[index] = oneMember
	}

	d.Set("member", memberSet)

	return nil
}
