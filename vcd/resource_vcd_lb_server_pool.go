package vcd

import (
	"fmt"

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
				Type:     schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"is_enabled": &schema.Schema{
							Type:        schema.TypeBool,
							Required:    true,
							Description: "Defines wether the member is used in pool",
						},
						"name": {
							ForceNew:    true,
							Required:    true,
							Type:        schema.TypeString,
							Description: "Name of pool member",
						},
						"ip_address": {
							Optional: true,
							Type:     schema.TypeString,
						},
						"port": {
							Required: true,
							Type:     schema.TypeString,
						},
						"monitor_port": {
							Required: true,
							Type:     schema.TypeString,
						},
						"weight": {
							Required: true,
							Type:     schema.TypeString,
						},
						"min_connections": {
							Optional: true,
							Type:     schema.TypeString,
						},
						"max_connections": {
							Optional: true,
							Type:     schema.TypeString,
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
		return fmt.Errorf("unable to expand load balancer service monitor: %s", err)
	}

	//tempLock.Lock()
	//defer tempLock.Unlock()

	createdMonitor, err := edgeGateway.CreateLBServiceMonitor(LBPool)
	if err != nil {
		return fmt.Errorf("error creating new load balancer service monitor: %s", err)
	}

	d.SetId(createdMonitor.ID)
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
		return fmt.Errorf("unable to find load balancer service monitor with ID %s: %s", d.Id(), err)
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
		return fmt.Errorf("could not expand monitor for update: %s", err)
	}

	updatedLBPool, err := edgeGateway.UpdateLBServiceMonitor(updateLBPoolConfig)
	if err != nil {
		return fmt.Errorf("unable to update load balancer service monitor with ID %s: %s", d.Id(), err)
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
	err = edgeGateway.DeleteLBServiceMonitor(&types.LBPool{ID: d.Id()})
	if err != nil {
		return fmt.Errorf("error deleting load balancer service monitor: %s", err)
	}

	d.SetId("")
	return nil
}

func resourceVcdLBServerPoolImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	return []*schema.ResourceData{}, nil
}

func expandLBPool(d *schema.ResourceData) (*types.LBMonitor, error) {
	lbMonitor := &types.LBMonitor{
		Name:       d.Get("name").(string),
		Interval:   d.Get("interval").(int),
		Timeout:    d.Get("timeout").(int),
		Type:       d.Get("type").(string), // In API this field is called "mode", while in UI "type"
		MaxRetries: d.Get("max_retries").(int),
		Expected:   d.Get("expected").(string),
		Method:     d.Get("method").(string),
		URI:        d.Get("url").(string),
		Send:       d.Get("send").(string),
		Receive:    d.Get("receive").(string),
		Extension:  expandLBMonitorExtension(d),
	}

	return lbMonitor, nil
}

func flattenLBPool(d *schema.ResourceData, lBmonitor *types.LBMonitor) error {
	d.Set("interval", lBmonitor.Interval)
	d.Set("timeout", lBmonitor.Timeout)
	d.Set("max_retries", lBmonitor.MaxRetries)
	d.Set("type", lBmonitor.Type)
	// Optional attributes may not necessarilly
	d.Set("method", lBmonitor.Method)
	d.Set("url", lBmonitor.URI)
	d.Set("send", lBmonitor.Send)
	d.Set("receive", lBmonitor.Receive)
	d.Set("expected", lBmonitor.Expected)

	if err := flattenLBMonitorExtension(d, lBmonitor); err != nil {
		return err
	}

	return nil
}
