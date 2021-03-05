package vcd

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var datasourceNsxtDhcpPoolSetSchema = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"start_address": &schema.Schema{
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Start address of DHCP pool IP range",
		},
		"end_address": &schema.Schema{
			Type:        schema.TypeString,
			Computed:    true,
			Description: "End address of DHCP pool IP range",
		},
	},
}

func datasourceVcdOpenApiDhcp() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdOpenApiDhcpRead,

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
			"org_network_id": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Parent Org VDC network name",
			},
			"pool": &schema.Schema{
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "IP ranges used for DHCP pool allocation in the network",
				Elem:        datasourceNsxtDhcpPoolSetSchema,
			},
		},
	}
}

func datasourceVcdOpenApiDhcpRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	_, vdc, err := vcdClient.GetOrgAndVdcFromResource(d)
	if err != nil {
		return diag.Errorf("[NSX-T DHCP pool datasource read] error retrieving VDC: %s", err)
	}

	orgNetworkId := d.Get("org_network_id").(string)
	orgNetwork, err := vdc.GetOpenApiOrgVdcNetworkById(orgNetworkId)
	if err != nil {
		return diag.Errorf("[NSX-T DHCP pool datasource read] error retrieving Org VDC network with ID '%s': %s", orgNetworkId, err)
	}

	pool, err := vdc.GetOpenApiOrgVdcNetworkDhcp(orgNetwork.OpenApiOrgVdcNetwork.ID)
	if err != nil {
		return diag.Errorf("[NSX-T DHCP pool datasource read] error retrieving DHCP pools for Org network ID '%s': %s",
			orgNetwork.OpenApiOrgVdcNetwork.ID, err)
	}

	err = setOpenAPIOrgVdcNetworkDhcpData(orgNetwork.OpenApiOrgVdcNetwork.ID, pool.OpenApiOrgVdcNetworkDhcp, d)
	if err != nil {
		return diag.Errorf("[NSX-T DHCP pool datasource read] error setting DHCP pool: %s", err)
	}

	d.SetId(orgNetwork.OpenApiOrgVdcNetwork.ID)

	return nil
}
