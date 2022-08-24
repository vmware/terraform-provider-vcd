package vcd

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdProviderVdc() *schema.Resource {

	return &schema.Resource{
		ReadContext: datasourceVcdProviderVdcRead,
		Schema:      map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the Provider VDC",
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Description of the Provider VDC",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				// FIXME: Investigate the values
				Description: "Status of the Provider VDC (can be 1 \"normal\" or 0 \"????\")",
			},
			"is_enabled": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Whether this Provider VDC is enabled or not",
			},
			// FIXME: Investigate the values
			"compute_provider_scope": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "????",
			},
			"highest_supported_hardware_version": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "????",
			},
			"nsxt_manager": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "????",
			},
			"vdc_ids": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "????",
			},
			"storage_container_ids": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "????",
			},
			"external_network_ids": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "????",
			},
			"storage_profile_ids": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "????",
			},
			"resource_pool_ids": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "????",
			},
			"vm_placement_policy_ids": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "????",
			},
			"kubernetes_policy_ids": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "????",
			},
			"metadata": {
				Type:        schema.TypeMap,
				Computed:    true,
				Description: "Key and value pairs for Provider VDC metadata",
			},
		},
	}
}

func datasourceVcdProviderVdcRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}
