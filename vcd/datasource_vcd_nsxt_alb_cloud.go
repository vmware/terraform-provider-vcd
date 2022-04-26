package vcd

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdAlbCloud() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdAlbCloudRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "NSX-T ALB Cloud name",
			},
			"controller_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "NSX-T ALB Importable Cloud ID",
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "NSX-T ALB Cloud description",
			},
			"importable_cloud_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "NSX-T ALB Importable Cloud ID",
			},
			"network_pool_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Network pool name of NSX-T ALB Cloud",
			},
			"network_pool_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Network pool ID NSX-T ALB Cloud",
			},
			"health_status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "NSX-T ALB Cloud health status",
			},
			"health_message": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "NSX-T ALB Cloud detailed health message",
			},
		},
	}
}

func datasourceVcdAlbCloudRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	if !vcdClient.Client.IsSysAdmin {
		return diag.Errorf("this resource is only supported for Providers")
	}

	albCloud, err := vcdClient.GetAlbCloudByName(d.Get("name").(string))
	if err != nil {
		return diag.Errorf("unable to find NSX-T ALB Cloud: %s", err)
	}

	setNsxtAlbCloudData(d, albCloud.NsxtAlbCloud)
	d.SetId(albCloud.NsxtAlbCloud.ID)

	return nil
}
