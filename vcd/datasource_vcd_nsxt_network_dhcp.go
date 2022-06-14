package vcd

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var datasourceNsxtDhcpPoolSetSchema = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"start_address": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Start address of DHCP pool IP range",
		},
		"end_address": {
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
				Description: "The name of VDC to use, optional if defined at provider level",
				Deprecated:  "Org network will be looked up based on 'org_network_id' field",
			},
			"org_network_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Parent Org VDC network name",
			},
			"pool": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "IP ranges used for DHCP pool allocation in the network",
				Elem:        datasourceNsxtDhcpPoolSetSchema,
			},
			"dns_servers": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "The DNS server IPs to be assigned by this DHCP service. 2 values maximum.",
				MaxItems:    2,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func datasourceVcdOpenApiDhcpRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	org, err := vcdClient.GetOrgFromResource(d)
	if err != nil {
		return diag.Errorf("[NSX-T DHCP pool datasource read] error retrieving Org: %s", err)
	}

	orgNetworkId := d.Get("org_network_id").(string)
	orgVdcNetwork, err := org.GetOpenApiOrgVdcNetworkById(orgNetworkId)
	if err != nil {
		return diag.Errorf("[NSX-T DHCP pool datasource read] error retrieving Org VDC network with ID '%s': %s", orgNetworkId, err)
	}

	pool, err := orgVdcNetwork.GetOpenApiOrgVdcNetworkDhcp()
	if err != nil {
		return diag.Errorf("[NSX-T DHCP pool datasource read] error retrieving DHCP pools for Org network ID '%s': %s",
			d.Id(), err)
	}

	err = setOpenAPIOrgVdcNetworkDhcpData(orgVdcNetwork.OpenApiOrgVdcNetwork.ID, pool.OpenApiOrgVdcNetworkDhcp, d)
	if err != nil {
		return diag.Errorf("[NSX-T DHCP pool datasource read] error setting DHCP pool: %s", err)
	}

	d.SetId(orgVdcNetwork.OpenApiOrgVdcNetwork.ID)

	return nil
}
