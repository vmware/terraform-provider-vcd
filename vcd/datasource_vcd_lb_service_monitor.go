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
				Description: "vCD organization in which the LB Service Monitor is located",
			},
			"vdc": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "vCD virtual datacenter in which the LB Service Monitor is located",
			},
			"edge_gateway": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Edge gateway name in which the LB Service Monitor is located",
			},
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "LB Service Monitor name",
			},
			"interval": &schema.Schema{
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Interval in seconds at which a server is to be monitored",
			},
			"timeout": &schema.Schema{
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Maximum time in seconds within which a response from the server must be received",
			},
			"max_retries": &schema.Schema{
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Number of times the specified monitoring Method must fail sequentially before the server is declared down",
			},
			"type": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Way in which you want to send the health check request to the server. One of http, https, tcp, icmp, or udp",
			},
			"expected": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "String that the monitor expects to match in the status line of the http or https response (for example, HTTP/1.1)",
			},
			"method": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Method to be used to detect server status. One of OPTIONS, GET, HEAD, POST, PUT, DELETE, TRACE, or CONNECT",
			},
			"url": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "URL to be used in the server status request",
			},
			"send": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Data to be sent",
			},
			"receive": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "String to be matched in the response content",
			},
			"extension": {
				Type:        schema.TypeMap,
				Computed:    true,
				Description: "Advanced monitor parameters as key=value pairs",
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

	readLBMonitor, err := edgeGateway.GetLbServiceMonitor(&types.LbMonitor{Name: d.Get("name").(string)})
	if err != nil {
		return fmt.Errorf("unable to find load balancer service monitor with Name %s: %s", d.Get("name").(string), err)
	}

	d.SetId(readLBMonitor.Id)
	return setLBMonitorData(d, readLBMonitor)
}
