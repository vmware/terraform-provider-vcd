package vcd

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func datasourceVcdLbServiceMonitor() *schema.Resource {
	return &schema.Resource{
		Read: datasourceVcdLbServiceMonitorRead,
		Schema: map[string]*schema.Schema{
			"org": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "vCD organization in which the Service Monitor is located",
			},
			"vdc": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "vCD virtual datacenter in which the Service Monitor is located",
			},
			"edge_gateway": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Edge gateway name in which the Service Monitor is located",
			},
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Service Monitor name",
			},
			"interval": &schema.Schema{
				Type:     schema.TypeInt,
				Computed: true,
			},
			"timeout": &schema.Schema{
				Type:     schema.TypeInt,
				Computed: true,
			},
			"max_retries": &schema.Schema{
				Type:     schema.TypeInt,
				Computed: true,
			},
			"type": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"expected": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"method": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"url": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"send": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"receive": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"extension": {
				Type:     schema.TypeMap,
				Computed: true,
			},
		},
	}
}

func datasourceVcdLbServiceMonitorRead(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)
	edgeGateway, err := vcdClient.GetEdgeGatewayFromResource(d, "edge_gateway")
	if err != nil {
		return fmt.Errorf(errorUnableToFindEdgeGateway, err)
	}

	readLBMonitor, err := edgeGateway.ReadLBServiceMonitor(&types.LBMonitor{Name: d.Get("name").(string)})
	if err != nil {
		return fmt.Errorf("unable to find load balancer service monitor with Name %s: %s", d.Get("name").(string), err)
	}

	d.SetId(readLBMonitor.ID)
	return setLBMonitorData(d, readLBMonitor)
}
