package vcd

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdTmNsxtManager() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdTmNsxtManagerRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of NSX-T Manager",
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Description of NSX-T Manager",
			},
			"username": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Username for authenticating to NSX-T Manager",
			},
			"url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "URL of NSX-T Manager",
			},
			"network_provider_scope": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Network Provider Scope for NSX-T Manager",
			},
			"status": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Status of NSX-T Manager",
			},
		},
	}
}

func datasourceVcdTmNsxtManagerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	nsxtManager, err := vcdClient.GetTmNsxtManagerByName(d.Get("name").(string))
	if err != nil {
		return diag.Errorf("error retrieving NSX-T Manager: %s")
	}

	err = setTmNsxtManagerData(d, nsxtManager.TmNsxtManager)
	if err != nil {
		return diag.Errorf("error storing NSX-T Manager to state: %s", err)
	}

	return nil
}
