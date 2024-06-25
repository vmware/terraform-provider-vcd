package vcd

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdDseRegistryConfiguration() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdDseRegistryConfigurationRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "Artifact name",
			},
			"type": {
				Type:     schema.TypeString,
				Computed: true,
				Description: "The name of organization to use, optional if defined at provider " +
					"level. Useful when connected as sysadmin working across different organizations",
			},
			"package_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "",
			},
			"default_package_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "",
			},
			"package_repository": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "",
			},
			"default_repository": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "",
			},
			"chart_repository": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "",
			},
			"default_chart_repository": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "",
			},
			"version": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "",
			},
			"default_version": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "",
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
				Description: "",
			},
			"rde_state": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "",
			},
		},
	}
}

func datasourceVcdDseRegistryConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	configInstance, err := vcdClient.GetDataSolutionByName(d.Get("name").(string))
	if err != nil {
		return diag.Errorf("error retrieving DSE Configuration: %s", err)
	}

	artifacts := configInstance.DataSolution.Spec.Artifacts[0]

	artifactType := artifacts["type"].(string)
	dSet(d, "type", artifactType)

	if artifactType == "ChartRepository" {
		dSet(d, "chart_repository", artifacts["chartRepository"].(string))
		dSet(d, "default_chart_repository", artifacts["defaultChartRepository"].(string))

		dSet(d, "default_package_name", artifacts["defaultPackageName"].(string))
		dSet(d, "package_name", artifacts["packageName"].(string))
	}

	if artifactType == "PackageRepository" {
		dSet(d, "package_repository", artifacts["image"].(string))
		dSet(d, "default_repository", artifacts["defaultImage"].(string))
	}

	dSet(d, "version", artifacts["version"].(string))
	dSet(d, "default_version", artifacts["defaultVersion"].(string))

	compatibleVersionsSlice := strings.Split(artifacts["compatibleVersions"].(string), " ")
	err = d.Set("compatible_version_constraints", convertStringsToTypeSet(compatibleVersionsSlice))
	if err != nil {
		return diag.Errorf("error storing 'compatible_version_constraints': %s", err)
	}

	dSet(d, "requires_version_compatibility", artifacts["requireVersionCompatibility"].(bool))

	if configInstance.DefinedEntity.DefinedEntity.State != nil {
		dSet(d, "rde_state", *configInstance.DefinedEntity.DefinedEntity.State)
	}

	d.SetId(configInstance.RdeId())

	return nil
}
