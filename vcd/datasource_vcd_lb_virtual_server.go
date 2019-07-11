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
				Description: "Virtual Server name for lookup",
			},
			"description": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Virtual server description",
			},
			"enabled": &schema.Schema{
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "",
			},
			"ip_address": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "",
			},
			"protocol": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "",
			},
			"port": &schema.Schema{
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "",
			},
			"enable_acceleration": &schema.Schema{
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "",
			},
			"connection_limit": &schema.Schema{
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "",
			},
			"connection_rate_limit": &schema.Schema{
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "",
			},

			"app_rule_id": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "",
			},
			"app_profile_id": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "",
			},
			// TODO - find out if there is a fallback pool option
			"server_pool_id": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "",
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

	d.SetId(readLBVirtualServer.ID)
	return setlBVirtualServerData(d, readLBVirtualServer)
}
