package vcd

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdIpSpaceUplink() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdIpSpaceUplinkRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Tenant facing name for IP Space Uplink",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "IP Space Uplink description",
			},
			"external_network_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "External Network ID",
			},
			"ip_space_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "IP Space ID",
			},
			"ip_space_type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "IP Space Type",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "IP Space Status",
			},
		},
	}
}

func datasourceVcdIpSpaceUplinkRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	log.Printf("[TRACE] IP Space Uplink datasource read initiated")

	externalNetworkId := d.Get("external_network_id").(string)
	name := d.Get("name").(string)

	ipSpaceUplink, err := vcdClient.GetIpSpaceUplinkByName(externalNetworkId, name)
	if err != nil {
		return diag.Errorf("error finding IP Space Uplink by Name '%s': %s", d.Id(), err)
	}

	d.SetId(ipSpaceUplink.IpSpaceUplink.ID)
	err = setIpSpaceUplinkData(d, ipSpaceUplink.IpSpaceUplink)
	if err != nil {
		return diag.Errorf("error storing IP Space Uplink state: %s", err)
	}

	return nil
}
