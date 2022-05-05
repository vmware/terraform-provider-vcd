package vcd

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdAlbController() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdAlbControllerRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "NSX-T ALB Controller name",
			},
			"url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "NSX-T ALB Controller URL",
			},
			"username": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "NSX-T ALB Controller Username",
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "NSX-T ALB Controller description",
			},
			"license_type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "NSX-T ALB License type. One of 'BASIC', 'ENTERPRISE'",
			},
			"version": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "NSX-T ALB Controller version",
			},
		},
	}
}

func datasourceVcdAlbControllerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	if !vcdClient.Client.IsSysAdmin {
		return diag.Errorf("this resource is only supported for Providers")
	}

	albController, err := vcdClient.GetAlbControllerByName(d.Get("name").(string))
	if err != nil {
		return diag.Errorf("unable to find NSX-T ALB Controller by Name '%s': %s",
			d.Get("name").(string), err)
	}

	setNsxtAlbControllerData(d, albController.NsxtAlbController)
	d.SetId(albController.NsxtAlbController.ID)

	return nil
}
