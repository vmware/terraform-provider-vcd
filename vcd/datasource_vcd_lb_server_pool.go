package vcd

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func datasourceVcdLbServerPool() *schema.Resource {
	return &schema.Resource{
		Read: datasourceVcdLbServerPoolRead,
		Schema: map[string]*schema.Schema{
			"vdc": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"org": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"edge_gateway": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Edge gateway name",
			},
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Service Monitor name",
			},

			"description": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"algorithm": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"algorithm_parameters": {
				Type:     schema.TypeMap,
				Computed: true,
			},
			"monitor_id": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"enable_transparency": &schema.Schema{
				Type:     schema.TypeBool,
				Computed: true,
			},

			"member": {
				Computed: true,
				Type:     schema.TypeList,
				//Set:      lbServerPoolMemberHash,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"condition": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"name": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"ip_address": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"port": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"monitor_port": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"weight": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"min_connections": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"max_connections": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func datasourceVcdLbServerPoolRead(d *schema.ResourceData, meta interface{}) error {
	vcdClient := meta.(*VCDClient)
	edgeGateway, err := vcdClient.GetEdgeGatewayFromResource(d)
	if err != nil {
		return fmt.Errorf(errorUnableToFindEdgeGateway, err)
	}

	readLBPool, err := edgeGateway.ReadLBServerPool(&types.LBPool{Name: d.Get("name").(string)})
	if err != nil {
		return fmt.Errorf("unable to find load balancer server poolwith Name %s: %s", d.Get("name").(string), err)
	}

	d.SetId(readLBPool.ID)
	return flattenLBPool(d, readLBPool)
}
