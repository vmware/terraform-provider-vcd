package vcd

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdProviderVdc() *schema.Resource {

	return &schema.Resource{
		ReadContext: datasourceVcdProviderVdcRead,
		Schema: map[string]*schema.Schema{
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
				Type:     schema.TypeInt,
				Computed: true,
				// FIXME: Investigate the values
				Description: "Status of the Provider VDC (can be 1 \"normal\" or 0 \"????\")",
			},
			"is_enabled": {
				Type:        schema.TypeBool,
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

func datasourceVcdProviderVdcRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	providerVdcName := d.Get("name").(string)
	providerVdc, err := vcdClient.GetProviderVdcByName(providerVdcName)
	if err != nil {
		log.Printf("[DEBUG] Could not find any Provider VDC with name %s: %s", providerVdcName, err)
		return diag.Errorf("could not find any Provider VDC with name %s", providerVdcName, err)
	}

	dSet(d, "name", providerVdc.ProviderVdc.Name)
	dSet(d, "description", providerVdc.ProviderVdc.Description)
	dSet(d, "status", providerVdc.ProviderVdc.Status)
	dSet(d, "is_enabled", providerVdc.ProviderVdc.IsEnabled)

	metadata, err := providerVdc.GetMetadata()
	if err != nil {
		log.Printf("[DEBUG] Error retrieving metadata for Provider VDC: %s", err)
		return diag.Errorf("error retrieving metadata for Provider VDC %s: %s", providerVdcName, err)
	}
	err = d.Set("metadata", getMetadataStruct(metadata.MetadataEntry))
	if err != nil {
		return diag.Errorf("There was an issue when setting metadata into the schema - %s", err)
	}

	d.SetId(providerVdc.ProviderVdc.ID)
	return nil
}
