package vcd

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdLbServiceMonitor() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdLbServiceMonitorRead,
		Schema: map[string]*schema.Schema{
			"org": {
				Type:     schema.TypeString,
				Optional: true,
				Description: "The name of organization to use, optional if defined at provider " +
					"level. Useful when connected as sysadmin working across different organizations",
			},
			"vdc": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The name of VDC to use, optional if defined at provider level",
			},
			"edge_gateway": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Edge gateway name in which the LB Service Monitor is located",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "LB Service Monitor name",
			},
			"interval": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Interval in seconds at which a server is to be monitored",
			},
			"timeout": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Maximum time in seconds within which a response from the server must be received",
			},
			"max_retries": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Number of times the specified monitoring Method must fail sequentially before the server is declared down",
			},
			"type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Way in which you want to send the health check request to the server. One of http, https, tcp, icmp, or udp",
			},
			"expected": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "String that the monitor expects to match in the status line of the http or https response (for example, HTTP/1.1)",
			},
			"method": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Method to be used to detect server status. One of OPTIONS, GET, HEAD, POST, PUT, DELETE, TRACE, or CONNECT",
			},
			"url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "URL to be used in the server status request",
			},
			"send": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Data to be sent",
			},
			"receive": {
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

func datasourceVcdLbServiceMonitorRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	edgeGateway, err := vcdClient.GetEdgeGatewayFromResource(d, "edge_gateway")
	if err != nil {
		return diag.Errorf(errorUnableToFindEdgeGateway, err)
	}

	readLBMonitor, err := edgeGateway.GetLbServiceMonitorByName(d.Get("name").(string))
	if err != nil {
		return diag.Errorf("unable to find load balancer service monitor with Name %s: %s", d.Get("name").(string), err)
	}

	d.SetId(readLBMonitor.ID)
	err = setLBMonitorData(d, readLBMonitor)
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}
