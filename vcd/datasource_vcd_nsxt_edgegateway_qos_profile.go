package vcd

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdNsxtEdgeGatewayQosProfile() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdNsxtEdgeGatewayQosProfileRead,
		Schema: map[string]*schema.Schema{
			"nsxt_manager_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "ID of NSX-T manager",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of QoS profile in NSX-T manager",
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Description of QoS profile in NSX-T manager",
			},
			"committed_bandwidth": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Committed bandwidth in Mb/s",
			},
			"burst_size": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Burst size in bytes",
			},
			"excess_action": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Excess action",
			},
		},
	}
}

func datasourceVcdNsxtEdgeGatewayQosProfileRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	nsxtManagerId := d.Get("nsxt_manager_id").(string)
	qosProfileName := d.Get("name").(string)

	qosProfile, err := vcdClient.GetNsxtEdgeGatewayQosProfileByDisplayName(nsxtManagerId, qosProfileName)
	if err != nil {
		return diag.Errorf("could not find NSX-T QoS profile by Name '%s' in NSX-T manager %s: %s",
			qosProfileName, nsxtManagerId, err)
	}

	d.SetId(qosProfile.NsxtEdgeGatewayQosProfile.ID)
	dSet(d, "name", qosProfile.NsxtEdgeGatewayQosProfile.DisplayName)
	dSet(d, "excess_action", qosProfile.NsxtEdgeGatewayQosProfile.ExcessAction)
	dSet(d, "description", qosProfile.NsxtEdgeGatewayQosProfile.Description)
	dSet(d, "burst_size", qosProfile.NsxtEdgeGatewayQosProfile.BurstSize)
	dSet(d, "committed_bandwidth", qosProfile.NsxtEdgeGatewayQosProfile.CommittedBandwidth)

	return nil
}
