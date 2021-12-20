package vcd

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdAlbVirtualService() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdAlbVirtualServiceRead,

		Schema: map[string]*schema.Schema{
			"org": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Description: "The name of organization to use, optional if defined at provider " +
					"level. Useful when connected as sysadmin working across different organizations",
			},
			"vdc": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "The name of VDC to use, optional if defined at provider level",
			},
			"edge_gateway_id": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Edge gateway ID in which ALB Virtual Service is",
			},
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of ALB Virtual Service",
			},
			"description": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Description of ALB Virtual Service",
			},
			"pool_id": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Edge gateway ID in which ALB Pool should be created",
			},
			"service_engine_group_id": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Service Engine Group ID",
			},
			"ca_certificate_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "ID of certificate in library if used",
			},
			"enabled": &schema.Schema{
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Virtual Service is enabled or disabled (default true)",
			},
			"virtual_ip_address": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Virtual IP address (VIP) for Virtual Service",
			},
			"application_profile_type": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "HTTP, HTTPS, L4, L4_TLS",
			},
			"service_port": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"start_port": &schema.Schema{
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Starting port in the range",
						},
						"end_port": &schema.Schema{
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Starting port in the range",
						},
						"ssl_enabled": &schema.Schema{
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Starting port in the range",
						},
						"type": &schema.Schema{
							Type:        schema.TypeString,
							Computed:    true,
							Description: "One of 'TCP_PROXY', 'TCP_FAST_PATH', 'UDP_FAST_PATH'",
						},
					},
				},
			},
		},
	}
}

func datasourceVcdAlbVirtualServiceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return diag.Errorf("error getting Org and VDC: %s", err)
	}

	if vdc.IsNsxv() {
		return diag.Errorf("ALB Pools are only supported on NSX-T. Please use 'vcd_lb_server_pool' for NSX-V load balancers")
	}

	nsxtEdge, err := vdc.GetNsxtEdgeGatewayById(d.Get("edge_gateway_id").(string))
	if err != nil {
		return diag.Errorf("could not retrieve NSX-T nsxtEdge gateway with ID '%s': %s", d.Id(), err)
	}

	albVirtualService, err := vcdClient.GetAlbVirtualServiceByName(nsxtEdge.EdgeGateway.ID, d.Get("name").(string))
	if err != nil {
		return diag.Errorf("could not retrieve NSX-T ALB Virtual Service '%s': %s", d.Get("name").(string), err)
	}

	err = setNsxtAlbVirtualServiceData(d, albVirtualService.NsxtAlbVirtualService)
	if err != nil {
		return diag.Errorf("error setting NSX-T ALB Virtual Service data: %s", err)
	}
	d.SetId(albVirtualService.NsxtAlbVirtualService.ID)

	return nil
}
