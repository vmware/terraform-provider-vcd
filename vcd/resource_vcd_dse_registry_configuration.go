package vcd

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"github.com/vmware/go-vcloud-director/v2/util"
)

// Name of Data Solutions Operator package. It cannot be published itself, but it is still seen in
// the list.
var defaultDsoName = "VCD Data Solutions" // Data Solutions Operator (DSO) name

var dseContainerRegistry = &schema.Resource{
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

func resourceVcdDseRegistryConfiguration() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdDseRegistryConfigurationCreate,
		ReadContext:   resourceVcdDseRegistryConfigurationRead,
		UpdateContext: resourceVcdDseRegistryConfigurationUpdate,
		DeleteContext: resourceVcdDseRegistryConfigurationDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdDseRegistryConfigurationImport,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "Data Solution Name",
			},
			"use_default_value": {
				Type:          schema.TypeBool,
				Optional:      true,
				Description:   "Use the default settings as provided by the Data Solution",
				ConflictsWith: []string{"package_repository", "chart_repository", "version", "package_name"},
			},
			"package_name": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				Description:  "Package name",
				RequiredWith: []string{"chart_repository"},
			},
			"default_package_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Default Package name provided by Data Solution",
			},
			"version": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Version of package to use",
			},
			"default_version": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Default version provided by Data Solution",
			},
			"chart_repository": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Chart repository to use",
			},
			"default_chart_repository": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Default chart repository provided by Data Solution",
			},
			"type": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Type of Data Solution package. One of 'PackageRepository', 'ChartRepository' ",
			},
			"package_repository": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				Description:   "Package repository to use",
				ConflictsWith: []string{"package_name", "chart_repository"},
			},
			"default_repository": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Default package repository provided by Data Solution",
			},
			"container_registry": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "Container registry configuration - only applicable for 'VCD Data Solutions'",
				Elem:        dseContainerRegistry,
			},
			"compatible_version_constraints": {
				Type:        schema.TypeSet,
				Computed:    true,
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

func resourceVcdDseRegistryConfigurationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceVcdDseRegistryConfigurationCreateUpdate(ctx, d, meta, "CREATE")
}

func resourceVcdDseRegistryConfigurationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceVcdDseRegistryConfigurationCreateUpdate(ctx, d, meta, "UPDATE")
}

func resourceVcdDseRegistryConfigurationCreateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}, operation string) diag.Diagnostics {
	util.Logger.Printf("[TRACE] Data Solution Registry Configuration %s started", operation)
	vcdClient := meta.(*VCDClient)

	dseEntryConfig, err := vcdClient.GetDataSolutionByName(d.Get("name").(string))
	if err != nil {
		return diag.Errorf("error retrieving Data Solution Configuration: %s", err)
	}
	cfg := dseEntryConfig.DataSolution
	packageType := cfg.Spec.Artifacts[0]["type"]

	// Simulate UI button 'Use Default Value' - pass any value that is not nil in the default fields
	if d.Get("use_default_value").(bool) {
		artifacts := dseEntryConfig.DataSolution.Spec.Artifacts[0]

		if artifacts["defaultImage"] != nil {
			cfg.Spec.Artifacts[0]["image"] = artifacts["defaultImage"].(string)
		}

		if artifacts["defaultChartRepository"] != nil {
			cfg.Spec.Artifacts[0]["chartRepository"] = artifacts["defaultChartRepository"].(string)
		}
		if artifacts["defaultVersion"] != nil {
			cfg.Spec.Artifacts[0]["version"] = artifacts["defaultVersion"].(string)
		}

		if artifacts["defaultPackageName"] != nil {
			cfg.Spec.Artifacts[0]["packageName"] = artifacts["defaultPackageName"].(string)
		}

	} else { // user configured options
		// validations for user configurable options
		if packageType == "ChartRepository" && d.Get("package_repository").(string) != "" {
			return diag.Errorf("cannot use 'repository' field for charts, please use 'chart_repository' field")
		}

		if packageType == "PackageRepository" && (d.Get("chart_repository").(string) != "" || d.Get("package_name").(string) != "") {
			return diag.Errorf("cannot use 'chart_repository' and 'package_name' field for packages, please use 'package_repository' field")
		}

		if packageType == "ChartRepository" && (d.Get("package_name") == "" || d.Get("chart_repository") == "" || d.Get("version") == "") {
			return diag.Errorf("Package of type ChartRepository must have 'package_name', 'chart_repository', 'version' set")

		}
		// end of validations for user configurable options

		if d.Get("package_repository").(string) != "" {
			cfg.Spec.Artifacts[0]["image"] = d.Get("package_repository").(string)
		}

		if d.Get("chart_repository").(string) != "" {
			cfg.Spec.Artifacts[0]["chartRepository"] = d.Get("chart_repository").(string)
		}

		if d.Get("package_name").(string) != "" {
			cfg.Spec.Artifacts[0]["packageName"] = d.Get("package_name").(string)
		}
		cfg.Spec.Artifacts[0]["version"] = d.Get("version").(string)

	}

	// 'container_registry' blocks are only configured for DSO (name "VCD Data Solutions")
	containerRegistrySet := d.Get("container_registry").(*schema.Set)
	if len(containerRegistrySet.List()) > 0 {

		if d.Get("name").(string) != defaultDsoName {
			return diag.Errorf("only %s repository can configure container registries", defaultDsoName)
		}

		auths := getRegistryConfigurationType(containerRegistrySet)
		cfg.Spec.DockerConfig.Auths = auths
	}

	updatedDseEntry, err := dseEntryConfig.Update(cfg)
	if err != nil {
		return diag.Errorf("error updating Data Solution repository for '%s': %s", d.Get("name").(string), err)
	}

	err = resolveRdeIfNotYetResolved(d.Get("name").(string), updatedDseEntry)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(dseEntryConfig.RdeId())
	util.Logger.Printf("[TRACE] Data Solution Configuration %s ended", operation)

	return resourceVcdDseRegistryConfigurationRead(ctx, d, meta)
}

func resourceVcdDseRegistryConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericVcdDseRegistryConfigurationRead(ctx, d, meta, "resource")
}

func genericVcdDseRegistryConfigurationRead(_ context.Context, d *schema.ResourceData, meta interface{}, origin string) diag.Diagnostics {
	util.Logger.Printf("[TRACE] Data Solution Registry Configuration read for %s started", origin)
	vcdClient := meta.(*VCDClient)

	configInstance, err := vcdClient.GetDataSolutionByName(d.Get("name").(string))
	if err != nil {
		return diag.Errorf("error retrieving Data Solution Configuration: %s", err)
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

	err = setRegistryConfigurationData(configInstance, d)
	if err != nil {
		return diag.FromErr(err)
	}

	dSet(d, "rde_state", configInstance.DefinedEntity.State())
	d.SetId(configInstance.RdeId())

	return nil
}

func resourceVcdDseRegistryConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	util.Logger.Printf("[TRACE] Data Solution Registry Configuration delete started")
	vcdClient := meta.(*VCDClient)

	dseEntryConfig, err := vcdClient.GetDataSolutionByName(d.Get("name").(string))
	if err != nil {
		return diag.Errorf("error retrieving Data Solution Configuration: %s", err)
	}

	// There is no real deletion once configurations are created, but
	// restoring default values provided by Data Solutions is the closest to deletion.
	cfg := dseEntryConfig.DataSolution
	artifacts := dseEntryConfig.DataSolution.Spec.Artifacts[0]

	if artifacts["defaultImage"] != nil {
		cfg.Spec.Artifacts[0]["image"] = artifacts["defaultImage"].(string)
	}

	if artifacts["defaultChartRepository"] != nil {
		cfg.Spec.Artifacts[0]["chartRepository"] = artifacts["defaultChartRepository"].(string)
	}
	if artifacts["defaultVersion"] != nil {
		cfg.Spec.Artifacts[0]["version"] = artifacts["defaultVersion"].(string)
	}

	if artifacts["defaultPackageName"] != nil {
		cfg.Spec.Artifacts[0]["packageName"] = artifacts["defaultPackageName"].(string)
	}

	// Data Solutions Operator (DSO) repository additionally has registry host configuration, it
	// must also be removed
	if dseEntryConfig.Name() == defaultDsoName {
		cfg.Spec.DockerConfig = &types.DseDockerConfig{Auths: types.DseDockerAuths{}}
	}

	_, err = dseEntryConfig.Update(cfg)
	if err != nil {
		return diag.Errorf("error updating Data Solution R repository details for '%s': %s", d.Get("name").(string), err)
	}

	return nil
}

func resourceVcdDseRegistryConfigurationImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	vcdClient := meta.(*VCDClient)
	util.Logger.Printf("[TRACE] Data Solution Registry Configuration import started with ID %s", d.Id())

	configInstance, err := vcdClient.GetDataSolutionByName(d.Id())
	if err != nil {
		return nil, fmt.Errorf("error retrieving Data Solution Configuration: %s", err)
	}

	dSet(d, "name", d.Id())
	dSet(d, "use_default_value", false)
	d.SetId(configInstance.RdeId())

	return []*schema.ResourceData{d}, nil
}

func setRegistryConfigurationData(configInstance *govcd.DataSolution, d *schema.ResourceData) error {
	if configInstance.DataSolution != nil && configInstance.DataSolution.Spec.DockerConfig != nil && configInstance.DataSolution.Spec.DockerConfig.Auths != nil {
		registryAuthConfigEntries := configInstance.DataSolution.Spec.DockerConfig.Auths
		authConfigSlice := make([]interface{}, len(registryAuthConfigEntries))
		for host, hostConfig := range configInstance.DataSolution.Spec.DockerConfig.Auths {
			singleEntry := make(map[string]interface{})
			singleEntry["host"] = host
			singleEntry["username"] = hostConfig.Username
			singleEntry["password"] = hostConfig.Password
			singleEntry["description"] = hostConfig.Description

			authConfigSlice = append(authConfigSlice, singleEntry)
		}
		authConfigSet := schema.NewSet(schema.HashResource(dseContainerRegistry), authConfigSlice)
		err := d.Set("container_registry", authConfigSet)
		if err != nil {
			return fmt.Errorf("error storing 'container_registry': %s", err)
		}
	}

	return nil
}

func getRegistryConfigurationType(containerRegistrySet *schema.Set) types.DseDockerAuths {
	containerRegistryList := containerRegistrySet.List()
	auths := make(map[string]types.DseDockerAuth)

	for _, entry := range containerRegistryList {
		entryMap := entry.(map[string]interface{})

		host := entryMap["host"].(string)
		description := entryMap["description"].(string)
		username := entryMap["username"].(string)
		password := entryMap["password"].(string)

		authEntry := types.DseDockerAuth{}
		if username != "" {
			authEntry.Username = username
		}

		if password != "" {
			authEntry.Password = password
		}

		if description != "" {
			authEntry.Description = description
		}

		auths[host] = authEntry
	}

	return auths
}

func resolveRdeIfNotYetResolved(name string, dseEntity *govcd.DataSolution) error {
	var err error

	if dseEntity.DefinedEntity.State() != "RESOLVED" {
		err = dseEntity.DefinedEntity.Resolve()
		if err != nil {
			return fmt.Errorf("error resolving Data Solution Config with Name '%s', ID '%s': %s", name, dseEntity.RdeId(), err)
		}
		// error might be nill but there might be RESOLUTION error
		if dseEntity.DefinedEntity.State() == "RESOLUTION_ERROR" {
			return fmt.Errorf("error resolving Data Solution Config with Name '%s', ID '%s':\n\nError message: %s",
				name, dseEntity.RdeId(), dseEntity.DefinedEntity.DefinedEntity.Message)
		}
	}

	return nil
}
