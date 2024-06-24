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

/*
"name": "VCD Data Solutions",
"type": "PackageRepository",
"image": "harbor-repo.vmware.com/vcdtds/dev/vcd-data-solutions-package-repo:1.4.0-dev.104.g2266851b",
"version": "1.4.0",
"manifests": "apiVersion: v1\nkind: Namespace\nmetadata:\n  name: vcd-ds-system\n  labels:\n    vcd-ds/installation-manifest: \"true\"\n---\napiVersion: v1\nkind: Namespace\nmetadata:\n  name: vcd-ds-workloads\n  labels:\n    vcd-ds/installation-manifest: \"true\"\n---\napiVersion: v1\ndata:\n  .dockerconfigjson: {{{registryCreds}}}\nkind: Secret\nmetadata:\n  name: vcd-ds-registry-creds\n  namespace: vcd-ds-system\n  labels:\n    vcd-ds/installation-manifest: \"true\"\ntype: kubernetes.io/dockerconfigjson\n---\napiVersion: secretgen.carvel.dev/v1alpha1\nkind: SecretExport\nmetadata:\n  name: vcd-ds-registry-creds\n  namespace: vcd-ds-system\n  labels:\n    vcd-ds/installation-manifest: \"true\"\nspec:\n  toNamespaces:\n  - vcd-ds-system\n  - vcd-ds-workloads\n---\napiVersion: v1\nkind: ServiceAccount\nmetadata:\n  name: vcd-data-solutions-install\n  namespace: vcd-ds-system\n  labels:\n    vcd-ds/installation-manifest: \"true\"\n---\napiVersion: rbac.authorization.k8s.io/v1\nkind: ClusterRole\nmetadata:\n  name: vcd-data-solutions-install\n  labels:\n    vcd-ds/installation-manifest: \"true\"\nrules:\n- apiGroups:\n  - \"\"\n  resources:\n  - configmaps\n  - secrets\n  - serviceaccounts\n  - pods\n  verbs:\n  - \"*\"\n- apiGroups:\n  - apps\n  resources:\n  - deployments\n  verbs:\n  - \"*\"\n- apiGroups:\n  - rbac.authorization.k8s.io\n  resources:\n  - clusterrolebindings\n  - clusterroles\n  verbs:\n  - \"*\"\n- apiGroups:\n  - secretgen.carvel.dev\n  resources:\n  - secretexports\n  - secretimports\n  verbs:\n  - \"*\"\n- apiGroups:\n  - cert-manager.io\n  resources:\n  - certificates\n  - issuers\n  verbs:\n  - \"*\"\n---\napiVersion: rbac.authorization.k8s.io/v1\nkind: ClusterRoleBinding\nmetadata:\n  name: vcd-data-solutions-install\n  labels:\n    vcd-ds/installation-manifest: \"true\"\nroleRef:\n  apiGroup: rbac.authorization.k8s.io\n  kind: ClusterRole\n  name: vcd-data-solutions-install\nsubjects:\n- kind: ServiceAccount\n  name: vcd-data-solutions-install\n  namespace: vcd-ds-system\n---\napiVersion: v1\nkind: Secret\nmetadata:\n  name: vcd-data-solutions-install-values\n  namespace: vcd-ds-system\n  labels:\n    vcd-ds/installation-manifest: \"true\"\nstringData:\n  values.yml: |\n    ---\n    vcdURL: {{vcdURL}}\n    vcdOrg: \"{{vcdOrg}}\"\n    clusterID: {{clusterID}}\n    apiToken: {{apiToken}}\n---\napiVersion: packaging.carvel.dev/v1alpha1\nkind: PackageRepository\nmetadata:\n  name: vcd-data-solutions\n  namespace: vcd-ds-system\n  labels:\n    vcd-ds/installation-manifest: \"true\"\nspec:\n  fetch:\n    imgpkgBundle:\n      image: {{dsRepoImage}}\n---\napiVersion: packaging.carvel.dev/v1alpha1\nkind: PackageInstall\nmetadata:\n  name: vcd-data-solutions-install\n  namespace: vcd-ds-system\n  annotations:\n    packaging.carvel.dev/downgradable: \"\"\n  labels:\n    vcd-ds/installation-manifest: \"true\"\nspec:\n  serviceAccountName: vcd-data-solutions-install\n  packageRef:\n    refName: data-solutions.vcloud.vmware.com\n    versionSelection:\n      constraints: \"{{dsVersion}}\"\n      prereleases: {}\n  values:\n  - secretRef:\n      name: vcd-data-solutions-install-values",
"defaultImage": "projects.registry.vmware.com/vcdds/vcd-data-solutions-package-repo:1.4.0",
"defaultVersion": "1.4.0",
"compatibleVersions": ">=1.0.0 <=1.4.0",
"requireVersionCompatibility": true
*/

/*
"name": "MongoDB",
"type": "ChartRepository",
"version": "",
"packageName": "",
"defaultVersion": "1.24.0",
"chartRepository": "",
"compatibleVersions": ">=1.23.0 <1.25.0",
"defaultPackageName": "enterprise-operator",
"defaultChartRepository": "https://mongodb.github.io/helm-charts",
"requireVersionCompatibility": false
*/
