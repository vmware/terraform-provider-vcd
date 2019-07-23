package vcd

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
)

func datasourceVcdLbVirtualServer() *schema.Resource {
	return &schema.Resource{
		Read: datasourceVcdLbVirtualServerRead,
		Schema: map[string]*schema.Schema{
			"org": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "vCD organization in which the Virtual Server is located",
			},
			"vdc": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "vCD virtual datacenter in which the Virtual Server is located",
			},
			"edge_gateway": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Edge gateway name in which the Virtual Server is located",
			},
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Virtual Server name for lookup",
			},
			"description": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Virtual Server description",
			},
			"enabled": &schema.Schema{
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Defines if the virtual server is enabled",
			},
			"ip_address": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "IP address that the load balancer listens on",
			},
			"protocol": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Protocol that the virtual server accepts",
			},
			"port": &schema.Schema{
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Port number that the load balancer listens on",
			},
			"enable_acceleration": &schema.Schema{
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Enable virtual server acceleration",
			},
			"connection_limit": &schema.Schema{
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Maximum concurrent connections that the virtual server can process",
			},
			"connection_rate_limit": &schema.Schema{
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Maximum incoming new connection requests per second",
			},
			"app_profile_id": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Application profile ID to be associated with the virtual server",
			},
			"server_pool_id": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The server pool that the load balancer will use",
			},
			"app_rule_ids": &schema.Schema{
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of attached application rule IDs",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func datasourceVcdLbVirtualServerRead(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)
	edgeGateway, err := vcdClient.GetEdgeGatewayFromResource(d, "edge_gateway")
	if err != nil {
		return fmt.Errorf(errorUnableToFindEdgeGateway, err)
	}

	readLBVirtualServer, err := edgeGateway.ReadLBVirtualServerByName(d.Get("name").(string))
	if err != nil {
		return fmt.Errorf("unable to find load balancer virtual server with Name %s: %s",
			d.Get("name").(string), err)
	}

	d.SetId(readLBVirtualServer.Id)
	return setlBVirtualServerData(d, readLBVirtualServer)
}
