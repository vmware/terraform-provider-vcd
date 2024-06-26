package vcd

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdDseRegistryConfiguration() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdDseRegistryConfigurationRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Data Solution Name",
			},
			"type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Type of Data Solution package. One of 'PackageRepository', 'ChartRepository'",
			},
			"package_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Package name",
			},
			"default_package_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Package name when type=ChartRepository",
			},
			"chart_repository": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Chart repository used",
			},
			"default_chart_repository": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Default chart repository provided by Data Solution",
			},
			"version": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Version of package to use",
			},
			"default_version": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Default version provided by Solution",
			},
			"package_repository": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Package repository to use",
			},
			"default_repository": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Default package repository provided by Solution",
			},
			"container_registry": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "Container registry configuration - only applicable for 'VCD Data Solutions'",
				Elem:        datasourceDseContainerRegistry,
			},
			"compatible_version_constraints": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "A set of version compatibility constraints",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"requires_version_compatibility": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Boolean flag if the Data Solution requires version compatibility",
			},
			"rde_state": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Parent RDE state",
			},
		},
	}
}

var datasourceDseContainerRegistry = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"host": {
			Required:    true,
			Type:        schema.TypeString,
			Description: "Registry host",
		},
		"description": {
			Required:    true,
			Type:        schema.TypeString,
			Description: "Registry description",
		},
		"username": {
			Optional:    true,
			Type:        schema.TypeString,
			Description: "Username for registry access",
		},
		"password": {
			Optional:    true,
			Type:        schema.TypeString,
			Description: "Password for registry user",
			Sensitive:   true,
		},
	},
}

func datasourceVcdDseRegistryConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericVcdDseRegistryConfigurationRead(ctx, d, meta, "datasource")
}
