package vcd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
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
			"vdc": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"org": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"edge_gateway": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Edge gateway name",
			},
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Unique Server Pool name",
			},

			"description": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Server pool description",
			},
			"algorithm": &schema.Schema{
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"ROUND-ROBIN", "IP-HASH", "LEASTCONN", "URI", "HTTPHEADER", "URL"}, false),
				Description:  "Load balancing algorithm",
			},
			"algorithm_parameters": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "Additional options for load balancing algorithm",
			},
			"monitor_id": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Load Balancer Service Monitor ID",
			},
			"is_transparent": &schema.Schema{
				Type:        schema.TypeBool,
				Default:     false,
				Optional:    true,
				Description: "Client IP addresses are visible to the backend servers when enabled",
			},

			"member": {
				Optional: true,
				ForceNew: false,
				Type:     schema.TypeSet,
				Set:      lbServerPoolMemberHash,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						//"id": {
						//	ForceNew:    false,
						//	Computed:    true,
						//	Type:        schema.TypeString,
						//	Description: "Pool member id",
						//},
						"condition": &schema.Schema{
							Type:         schema.TypeString,
							ForceNew:     false,
							Required:     true,
							ValidateFunc: validation.StringInSlice([]string{"enabled", "drain", "disabled"}, false),
							Description:  "Defines member state",
						},
						"name": {
							//ForceNew:    true,
							Required:    true,
							ForceNew:    false,
							Type:        schema.TypeString,
							Description: "Name of pool member",
						},
						"ip_address": {
							Optional: true,
							ForceNew: false,
							Type:     schema.TypeString,
						},
						"port": {
							Required: true,
							ForceNew: false,
							Type:     schema.TypeInt,
						},
						"monitor_port": {
							Required: true,
							ForceNew: false,
							Type:     schema.TypeInt,
						},
						"weight": {
							Required: true,
							ForceNew: false,
							Type:     schema.TypeInt,
						},
						"min_connections": {
							Optional: true,
							ForceNew: false,
							Type:     schema.TypeInt,
						},
						"max_connections": {
							Optional: true,
							ForceNew: false,
							Type:     schema.TypeInt,
						},
					},
				},
			},
		},
	}
}

func resourceVcdLBServerPoolCreate(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)
	edgeGateway, err := vcdClient.GetEdgeGatewayFromResource(d)
	if err != nil {
		return fmt.Errorf(errorUnableToFindEdgeGateway, err)
	}

	LBPool, err := expandLBPool(d)
	if err != nil {
		return fmt.Errorf("unable to expand load balancer server pool: %s", err)
	}

	//tempLock.Lock()
	//defer tempLock.Unlock()

	createdPool, err := edgeGateway.CreateLBServerPool(LBPool)
	if err != nil {
		return fmt.Errorf("error creating new load balancer server pool: %s", err)
	}

	d.SetId(createdPool.ID)
	return nil
}

func resourceVcdLBServerPoolRead(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)
	edgeGateway, err := vcdClient.GetEdgeGatewayFromResource(d)
	if err != nil {
		return fmt.Errorf(errorUnableToFindEdgeGateway, err)
	}

	//tempLock.Lock()
	//defer tempLock.Unlock()
	readLBPool, err := edgeGateway.ReadLBServerPool(&types.LBPool{ID: d.Id()})
	if err != nil {
		d.SetId("")
		return fmt.Errorf("unable to find load balancer server pool with ID %s: %s", d.Id(), err)
	}

	return flattenLBPool(d, readLBPool)
}

func resourceVcdLBServerPoolUpdate(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)
	edgeGateway, err := vcdClient.GetEdgeGatewayFromResource(d)
	if err != nil {
		return fmt.Errorf(errorUnableToFindEdgeGateway, err)
	}

	//tempLock.Lock()
	//defer tempLock.Unlock()

	updateLBPoolConfig, err := expandLBPool(d)
	if err != nil {
		return fmt.Errorf("could not expand load balancer server pool for update: %s", err)
	}

	updatedLBPool, err := edgeGateway.UpdateLBServerPool(updateLBPoolConfig)
	if err != nil {
		return fmt.Errorf("unable to update load balancer server pool with ID %s: %s", d.Id(), err)
	}

	if err := flattenLBPool(d, updatedLBPool); err != nil {
		return err
	}

	return nil
}

func resourceVcdLBServerPoolDelete(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)
	edgeGateway, err := vcdClient.GetEdgeGatewayFromResource(d)
	if err != nil {
		return fmt.Errorf(errorUnableToFindEdgeGateway, err)
	}

	//tempLock.Lock()
	//defer tempLock.Unlock()
	err = edgeGateway.DeleteLBServerPool(&types.LBPool{ID: d.Id()})
	if err != nil {
		return fmt.Errorf("error deleting load balancer server pool: %s", err)
	}

	d.SetId("")
	return nil
}

func resourceVcdLBServerPoolImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	return []*schema.ResourceData{}, nil
}

func expandLBPool(d *schema.ResourceData) (*types.LBPool, error) {
	lbPool := &types.LBPool{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		Algorithm:   d.Get("algorithm").(string),
		//AlgorithmParameters: d.Get("algorithm_parameters").(string),
		MonitorId:   d.Get("monitor_id").(string),
		Transparent: d.Get("is_transparent").(bool),
	}

	members, err := expandLBPoolMembers(d)
	if err != nil {
		return nil, err
	}
	lbPool.Members = members

	return lbPool, nil
}

func expandLBPoolMembers(d *schema.ResourceData) (types.LBPoolMembers, error) {
	var lbPoolMembers types.LBPoolMembers

	//members := d.Get("member").([]map[string]interface{})
	members := d.Get("member").(*schema.Set).List()
	//mm := members.([]map[string]interface{})
	for _, memberInterface := range members {

		member := memberInterface.(map[string]interface{})

		var memberConfig types.LBPoolMember

		memberConfig.Name = member["name"].(string)
		memberConfig.IpAddress = member["ip_address"].(string)
		memberConfig.Port = member["port"].(int)
		memberConfig.MonitorPort = member["monitor_port"].(int)
		memberConfig.Weight = member["weight"].(int)
		memberConfig.MinConn = member["min_connections"].(int)
		memberConfig.MaxConn = member["max_connections"].(int)
		memberConfig.MaxConn = member["weight"].(int)
		memberConfig.Condition = member["condition"].(string)

		lbPoolMembers = append(lbPoolMembers, memberConfig)
	}

	return lbPoolMembers, nil
}

func flattenLBPool(d *schema.ResourceData, lBpool *types.LBPool) error {
	d.Set("name", lBpool.Name)
	d.Set("description", lBpool.Description)
	d.Set("algorithm", lBpool.Algorithm)
	d.Set("algorithm_parameters", lBpool.AlgorithmParameters)
	// Optional attributes may not necessarily
	d.Set("monitor_id", lBpool.MonitorId)
	d.Set("is_transparent", lBpool.Transparent)

	err := flattenLBPoolMembers(d, lBpool.Members)
	if err != nil {
		return err
	}
	return nil
}

func flattenLBPoolMembers(d *schema.ResourceData, lBpoolMembers types.LBPoolMembers) error {

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
		oneMember["id"] = member.ID

		memberSet[index] = oneMember
	}

	d.Set("member", memberSet)

	return nil
}

func lbServerPoolMemberHash(v interface{}) int {
	m := v.(map[string]interface{})
	splitID := strings.Split(m["id"].(string), "-")
	intId, _ := strconv.Atoi(splitID[1])

	return intId
}
