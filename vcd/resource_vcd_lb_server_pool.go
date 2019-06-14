package vcd

import (
	"fmt"
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
				ValidateFunc: validation.StringInSlice([]string{"round-robin", "ip-hash", "leastconn", "uri", "httpheader", "url"}, false),
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
			"enable_transparency": &schema.Schema{
				Type:        schema.TypeBool,
				Default:     false,
				Optional:    true,
				Description: "Client IP addresses are visible to the backend servers when enabled",
			},

			"member": {
				Optional: true,
				ForceNew: false,
				Type:     schema.TypeList,
				//Set:      lbServerPoolMemberHash,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							ForceNew:    false,
							Computed:    true,
							Type:        schema.TypeString,
							Description: "Pool member id",
						},
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
	vcdClient.lockParentEdgeGtw(d)
	defer vcdClient.unLockParentEdgeGtw(d)

	edgeGateway, err := vcdClient.GetEdgeGatewayFromResource(d)
	if err != nil {
		return fmt.Errorf(errorUnableToFindEdgeGateway, err)
	}

	LBPool, err := getLBPoolType(d)
	if err != nil {
		return fmt.Errorf("unable to expand load balancer server pool: %s", err)
	}

	createdPool, err := edgeGateway.CreateLBServerPool(LBPool)
	if err != nil {
		return fmt.Errorf("error creating new load balancer server pool: %s", err)
	}

	// We store the values once again because response include pool member IDs
	setLBPoolData(d, createdPool)
	d.SetId(createdPool.ID)
	return nil
}

func resourceVcdLBServerPoolRead(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)
	vcdClient.lockParentEdgeGtw(d)
	defer vcdClient.unLockParentEdgeGtw(d)

	edgeGateway, err := vcdClient.GetEdgeGatewayFromResource(d)
	if err != nil {
		return fmt.Errorf(errorUnableToFindEdgeGateway, err)
	}

	readLBPool, err := edgeGateway.ReadLBServerPool(&types.LBPool{ID: d.Id()})
	if err != nil {
		d.SetId("")
		return fmt.Errorf("unable to find load balancer server pool with ID %s: %s", d.Id(), err)
	}

	return setLBPoolData(d, readLBPool)
}

func resourceVcdLBServerPoolUpdate(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)
	vcdClient.lockParentEdgeGtw(d)
	defer vcdClient.unLockParentEdgeGtw(d)

	edgeGateway, err := vcdClient.GetEdgeGatewayFromResource(d)
	if err != nil {
		return fmt.Errorf(errorUnableToFindEdgeGateway, err)
	}

	updateLBPoolConfig, err := getLBPoolType(d)
	if err != nil {
		return fmt.Errorf("could not expand load balancer server pool for update: %s", err)
	}

	updatedLBPool, err := edgeGateway.UpdateLBServerPool(updateLBPoolConfig)
	if err != nil {
		return fmt.Errorf("unable to update load balancer server pool with ID %s: %s", d.Id(), err)
	}

	if err := setLBPoolData(d, updatedLBPool); err != nil {
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
	resourceURI := strings.Split(d.Id(), ".")
	if len(resourceURI) != 4 {
		return nil, fmt.Errorf("resource name must be specified in such way my-org.my-org-vdc.my-edge-gw.existing-server-pool")
	}
	orgName, vdcName, edgeName, poolName := resourceURI[0], resourceURI[1], resourceURI[2], resourceURI[3]

	vcdClient := meta.(*VCDClient)
	edgeGateway, err := vcdClient.GetEdgeGateway(orgName, vdcName, edgeName)
	if err != nil {
		return nil, fmt.Errorf(errorUnableToFindEdgeGateway, err)
	}

	readLBPool, err := edgeGateway.ReadLBServerPool(&types.LBPool{Name: poolName})
	if err != nil {
		return []*schema.ResourceData{}, fmt.Errorf("unable to find load balancer server pool with name %s: %s", d.Id(), err)
	}

	d.Set("org", orgName)
	d.Set("vdc", vdcName)
	d.Set("edge_gateway", edgeName)
	d.Set("name", poolName)

	d.SetId(readLBPool.ID)
	return []*schema.ResourceData{d}, nil
}

func getLBPoolType(d *schema.ResourceData) (*types.LBPool, error) {
	lbPool := &types.LBPool{
		Name:                d.Get("name").(string),
		Description:         d.Get("description").(string),
		Algorithm:           d.Get("algorithm").(string),
		MonitorId:           d.Get("monitor_id").(string),
		Transparent:         d.Get("enable_transparency").(bool),
		AlgorithmParameters: getLBPoolAlgorithmType(d),
	}

	members, err := getLBPoolMembersType(d)
	if err != nil {
		return nil, err
	}
	lbPool.Members = members

	return lbPool, nil
}

func getLBPoolMembersType(d *schema.ResourceData) (types.LBPoolMembers, error) {
	var lbPoolMembers types.LBPoolMembers

	members := d.Get("member").([]interface{})
	for _, memberInterface := range members {
		var memberConfig types.LBPoolMember
		member := memberInterface.(map[string]interface{})

		// If we have IDs - then we must insert them for update. Otherwise the update may get mixed
		if member["id"].(string) != "" {
			memberConfig.ID = member["id"].(string)
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

func getLBPoolAlgorithmType(d *schema.ResourceData) string {
	var extensionString string
	extension := d.Get("algorithm_parameters").(map[string]interface{})
	for k, v := range extension {
		if k != "" && v != "" { // When key and value are given it must look like "content-type=STRING"
			extensionString += k + "=" + v.(string) + "\n"
		} else { // If only key is specified it does not need equals sign. Like "no-body" extension
			extensionString += k + "\n"
		}
	}
	return extensionString
}

func setLBPoolData(d *schema.ResourceData, lBpool *types.LBPool) error {
	d.Set("name", lBpool.Name)
	d.Set("description", lBpool.Description)
	d.Set("algorithm", lBpool.Algorithm)
	// Optional attributes may not necessarily be set
	d.Set("monitor_id", lBpool.MonitorId)
	d.Set("enable_transparency", lBpool.Transparent)

	err := setLBPoolAlgorithmData(d, lBpool)
	if err != nil {
		return err
	}

	err = setLBPoolMembersData(d, lBpool.Members)
	if err != nil {
		return err
	}
	return nil
}

func setLBPoolMembersData(d *schema.ResourceData, lBpoolMembers types.LBPoolMembers) error {

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

func setLBPoolAlgorithmData(d *schema.ResourceData, lBPool *types.LBPool) error {
	extensionStorage := make(map[string]string)

	if lBPool.AlgorithmParameters != "" {
		kvList := strings.Split(lBPool.AlgorithmParameters, "\n")
		for _, algorithmLine := range kvList {
			// Skip empty lines
			if algorithmLine == "" {
				continue
			}

			// When key=algorithmLine format is present
			if strings.Contains(algorithmLine, "=") {
				keyValue := strings.Split(algorithmLine, "=")
				if len(keyValue) != 2 {
					return fmt.Errorf("unable to flatten extension field %s", algorithmLine)
				}
				// Populate extension data with key value
				extensionStorage[keyValue[0]] = keyValue[1]
				// If there was no "=" sign then it means whole line is just key. Like `no-body`, `linespan`
			} else {
				extensionStorage[algorithmLine] = ""
			}
		}

	}

	d.Set("algorithm_parameters", extensionStorage)
	return nil
}
