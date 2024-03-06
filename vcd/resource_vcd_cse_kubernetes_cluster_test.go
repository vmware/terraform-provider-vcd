//go:build cse || ALL

package vcd

import (
	"fmt"
	semver "github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"os"
	"reflect"
	"regexp"
	"strings"
	"testing"
)

func requireCseConfig(t *testing.T, testConfig TestConfig) {
	skippedPrefix := fmt.Sprintf("skipped %s because:", t.Name())
	if cse := os.Getenv("TEST_VCD_CSE"); cse == "" {
		t.Skipf("%s the environment variable TEST_VCD_CSE is not set", skippedPrefix)
	}
	cseConfigValues := reflect.ValueOf(testConfig.Cse)
	cseConfigType := cseConfigValues.Type()
	for i := 0; i < cseConfigValues.NumField(); i++ {
		if cseConfigValues.Field(i).String() == "" {
			t.Skipf("%s the config value '%s' inside 'cse' object of vcd_test_config.json is not set", skippedPrefix, cseConfigType.Field(i).Name)
		}
	}
}

func TestAccVcdCseKubernetesCluster(t *testing.T) {
	preTestChecks(t)
	requireCseConfig(t, testConfig)

	cseVersion, err := semver.NewVersion(testConfig.Cse.Version)
	if err != nil {
		t.Fatal(err)
	}

	v411, err := semver.NewVersion("4.1.1")
	if err != nil {
		t.Fatal(err)
	}

	tokenFilename := getCurrentDir() + t.Name() + ".json"
	defer func() {
		// Clean the API Token file
		if fileExists(tokenFilename) {
			err := os.Remove(tokenFilename)
			if err != nil {
				fmt.Printf("could not delete API token file '%s', please delete it manually", tokenFilename)
			}
		}
	}()

	var params = StringMap{
		"CseVersion":         testConfig.Cse.Version,
		"Name":               strings.ToLower(t.Name()),
		"OvaCatalog":         testConfig.Cse.OvaCatalog,
		"OvaName":            testConfig.Cse.OvaName,
		"KubernetesOva":      "data.vcd_catalog_vapp_template.tkg_ova.id",
		"SolutionsOrg":       testConfig.Cse.SolutionsOrg,
		"TenantOrg":          testConfig.Cse.TenantOrg,
		"Vdc":                testConfig.Cse.TenantVdc,
		"EdgeGateway":        testConfig.Cse.EdgeGateway,
		"Network":            testConfig.Cse.RoutedNetwork,
		"TokenName":          t.Name(),
		"TokenFile":          tokenFilename,
		"ControlPlaneCount":  1,
		"NodePoolCount":      1,
		"ExtraWorkerPool":    " ",
		"PodsCidr":           "100.96.0.0/11",
		"ServicesCidr":       "100.64.0.0/13",
		"AutoRepairOnErrors": true,
		"NodeHealthCheck":    true,
		"Timeout":            150,
	}
	testParamsNotEmpty(t, params)

	step1 := templateFill(testAccVcdCseKubernetesCluster, params)
	debugPrintf("#[DEBUG] CONFIGURATION step1: %s", step1)

	params["FuncName"] = t.Name() + "Step2"
	params["AutoRepairOnErrors"] = "false" // Deactivate it to avoid non-empty plans. Also, it is recommended after cluster creation
	params["ControlPlaneCount"] = 3
	step2 := templateFill(testAccVcdCseKubernetesCluster, params)
	debugPrintf("#[DEBUG] CONFIGURATION step2: %s", step2)

	params["FuncName"] = t.Name() + "Step3"
	params["ControlPlaneCount"] = 1
	params["NodePoolCount"] = 2
	step3 := templateFill(testAccVcdCseKubernetesCluster, params)
	debugPrintf("#[DEBUG] CONFIGURATION step3: %s", step3)

	params["FuncName"] = t.Name() + "Step4"
	params["ControlPlaneCount"] = 1
	params["NodePoolCount"] = 1
	params["NodeHealthCheck"] = false
	step4 := templateFill(testAccVcdCseKubernetesCluster, params)
	debugPrintf("#[DEBUG] CONFIGURATION step4: %s", step4)

	extraWorkerPool := "  worker_pool {\n" +
		"    name               = \"worker-pool-2\"\n" +
		"    machine_count      = 1\n" +
		"    disk_size_gi       = 20\n" +
		"    sizing_policy_id   = data.vcd_vm_sizing_policy.tkg_small.id\n" +
		"    storage_profile_id = data.vcd_storage_profile.sp.id\n" +
		"  }"

	params["FuncName"] = t.Name() + "Step5"
	params["NodeHealthCheck"] = true
	params["ExtraWorkerPool"] = extraWorkerPool
	step5 := templateFill(testAccVcdCseKubernetesCluster, params)
	debugPrintf("#[DEBUG] CONFIGURATION step5: %s", step5)

	params["FuncName"] = t.Name() + "Step6"
	step6 := templateFill(testAccVcdCseKubernetesCluster+testAccVcdCseKubernetesClusterDS, params)
	debugPrintf("#[DEBUG] CONFIGURATION step6: %s", step5)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	cacheId := testCachedFieldValue{}
	clusterName := "vcd_cse_kubernetes_cluster.my_cluster"
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			if cacheId.fieldValue == "" {
				return fmt.Errorf("cached ID '%s' is empty", cacheId.fieldValue)
			}
			conn := testAccProvider.Meta().(*VCDClient)
			_, err := conn.GetRdeById(cacheId.fieldValue)
			if err == nil {
				return fmt.Errorf("cluster with ID '%s' still exists", cacheId.fieldValue)
			}
			return nil
		},
		Steps: []resource.TestStep{
			// Basic scenario of cluster creation
			{
				Config: step1,
				ExpectNonEmptyPlan: func() bool {
					// Auto Repair on Errors gets automatically deactivated after cluster creation since CSE 4.1.1,
					// so it will return a non-empty plan
					if cseVersion.GreaterThanOrEqual(v411) {
						return true
					} else {
						return false
					}
				}(),
				Check: resource.ComposeAggregateTestCheckFunc(
					cacheId.cacheTestResourceFieldValue(clusterName, "id"),
					resource.TestMatchResourceAttr(clusterName, "id", regexp.MustCompile(`^urn:vcloud:entity:vmware:capvcdCluster:.+$`)),
					resource.TestCheckResourceAttr(clusterName, "cse_version", testConfig.Cse.Version),
					resource.TestCheckResourceAttr(clusterName, "runtime", "tkg"),
					resource.TestCheckResourceAttr(clusterName, "name", strings.ToLower(t.Name())),
					resource.TestCheckResourceAttrPair(clusterName, "kubernetes_template_id", "data.vcd_catalog_vapp_template.tkg_ova", "id"),
					resource.TestCheckResourceAttrPair(clusterName, "org", "data.vcd_org_vdc.vdc", "org"),
					resource.TestCheckResourceAttrPair(clusterName, "vdc_id", "data.vcd_org_vdc.vdc", "id"),
					resource.TestCheckResourceAttrPair(clusterName, "network_id", "data.vcd_network_routed_v2.routed", "id"),
					resource.TestCheckNoResourceAttr(clusterName, "owner"), // It is taken from Provider config
					resource.TestCheckResourceAttr(clusterName, "ssh_public_key", ""),
					resource.TestCheckResourceAttr(clusterName, "control_plane.0.machine_count", "1"),
					resource.TestCheckResourceAttr(clusterName, "control_plane.0.disk_size_gi", "20"),
					resource.TestCheckResourceAttrPair(clusterName, "control_plane.0.sizing_policy_id", "data.vcd_vm_sizing_policy.tkg_small", "id"),
					resource.TestCheckResourceAttrPair(clusterName, "control_plane.0.storage_profile_id", "data.vcd_storage_profile.sp", "id"),
					resource.TestCheckResourceAttrSet(clusterName, "control_plane.0.ip"), // IP should be assigned after creation as it was not set manually in HCL config
					resource.TestCheckResourceAttr(clusterName, "worker_pool.#", "1"),
					resource.TestCheckResourceAttr(clusterName, "worker_pool.0.name", "worker-pool-1"),
					resource.TestCheckResourceAttr(clusterName, "worker_pool.0.machine_count", "1"),
					resource.TestCheckResourceAttr(clusterName, "worker_pool.0.disk_size_gi", "20"),
					resource.TestCheckResourceAttrPair(clusterName, "worker_pool.0.sizing_policy_id", "data.vcd_vm_sizing_policy.tkg_small", "id"),
					resource.TestCheckResourceAttrPair(clusterName, "worker_pool.0.storage_profile_id", "data.vcd_storage_profile.sp", "id"),
					resource.TestCheckResourceAttrPair(clusterName, "default_storage_class.0.storage_profile_id", "data.vcd_storage_profile.sp", "id"),
					resource.TestCheckResourceAttr(clusterName, "default_storage_class.0.name", "sc-1"),
					resource.TestCheckResourceAttr(clusterName, "default_storage_class.0.reclaim_policy", "delete"),
					resource.TestCheckResourceAttr(clusterName, "default_storage_class.0.filesystem", "ext4"),
					resource.TestCheckResourceAttr(clusterName, "pods_cidr", "100.96.0.0/11"),
					resource.TestCheckResourceAttr(clusterName, "services_cidr", "100.64.0.0/13"),
					resource.TestCheckResourceAttr(clusterName, "virtual_ip_subnet", ""),
					func() resource.TestCheckFunc {
						// Auto Repair on Errors gets automatically deactivated after cluster creation since CSE 4.1.1
						if cseVersion.GreaterThanOrEqual(v411) {
							return resource.TestCheckResourceAttr(clusterName, "auto_repair_on_errors", "false")
						} else {
							return resource.TestCheckResourceAttr(clusterName, "auto_repair_on_errors", "true")
						}
					}(),
					resource.TestCheckResourceAttr(clusterName, "node_health_check", "true"),
					resource.TestMatchResourceAttr(clusterName, "kubernetes_version", regexp.MustCompile(`^v[0-9]+\.[0-9]+\.[0-9]+\+vmware\.[0-9]$`)),
					resource.TestMatchResourceAttr(clusterName, "tkg_product_version", regexp.MustCompile(`^v[0-9]+\.[0-9]+\.[0-9]+$`)),
					resource.TestMatchResourceAttr(clusterName, "capvcd_version", regexp.MustCompile(`^v[0-9]+\.[0-9]+\.[0-9]+$`)),
					resource.TestMatchResourceAttr(clusterName, "cluster_resource_set_bindings.#", regexp.MustCompile(`^[1-9][0-9]*$`)),
					resource.TestMatchResourceAttr(clusterName, "cpi_version", regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+$`)),
					resource.TestMatchResourceAttr(clusterName, "csi_version", regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+$`)),
					resource.TestCheckResourceAttr(clusterName, "state", "provisioned"),
					resource.TestCheckResourceAttrSet(clusterName, "kubeconfig"),
					resource.TestMatchResourceAttr(clusterName, "events.#", regexp.MustCompile(`^[1-9][0-9]*$`)),
				),
			},
			// Scale the control plane to 3 replicas
			{
				Config: step2,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Control plane should change
					resource.TestCheckResourceAttr(clusterName, "control_plane.0.machine_count", "3"),

					// Other things should remain the same
					cacheId.testCheckCachedResourceFieldValue(clusterName, "id"),
					resource.TestCheckResourceAttr(clusterName, "cse_version", testConfig.Cse.Version),
					resource.TestCheckResourceAttr(clusterName, "runtime", "tkg"),
					resource.TestCheckResourceAttr(clusterName, "name", strings.ToLower(t.Name())),
					resource.TestCheckResourceAttrPair(clusterName, "kubernetes_template_id", "data.vcd_catalog_vapp_template.tkg_ova", "id"),
					resource.TestCheckResourceAttrPair(clusterName, "org", "data.vcd_org_vdc.vdc", "org"),
					resource.TestCheckResourceAttrPair(clusterName, "vdc_id", "data.vcd_org_vdc.vdc", "id"),
					resource.TestCheckResourceAttrPair(clusterName, "network_id", "data.vcd_network_routed_v2.routed", "id"),
					resource.TestCheckNoResourceAttr(clusterName, "owner"), // It is taken from Provider config
					resource.TestCheckResourceAttr(clusterName, "ssh_public_key", ""),
					resource.TestCheckResourceAttr(clusterName, "control_plane.0.disk_size_gi", "20"),
					resource.TestCheckResourceAttrPair(clusterName, "control_plane.0.sizing_policy_id", "data.vcd_vm_sizing_policy.tkg_small", "id"),
					resource.TestCheckResourceAttrPair(clusterName, "control_plane.0.storage_profile_id", "data.vcd_storage_profile.sp", "id"),
					resource.TestCheckResourceAttrSet(clusterName, "control_plane.0.ip"), // IP should be assigned after creation as it was not set manually in HCL config
					resource.TestCheckResourceAttr(clusterName, "worker_pool.#", "1"),
					resource.TestCheckResourceAttr(clusterName, "worker_pool.0.name", "worker-pool-1"),
					resource.TestCheckResourceAttr(clusterName, "worker_pool.0.machine_count", "1"),
					resource.TestCheckResourceAttr(clusterName, "worker_pool.0.disk_size_gi", "20"),
					resource.TestCheckResourceAttrPair(clusterName, "worker_pool.0.sizing_policy_id", "data.vcd_vm_sizing_policy.tkg_small", "id"),
					resource.TestCheckResourceAttrPair(clusterName, "worker_pool.0.storage_profile_id", "data.vcd_storage_profile.sp", "id"),
					resource.TestCheckResourceAttrPair(clusterName, "default_storage_class.0.storage_profile_id", "data.vcd_storage_profile.sp", "id"),
					resource.TestCheckResourceAttr(clusterName, "default_storage_class.0.name", "sc-1"),
					resource.TestCheckResourceAttr(clusterName, "default_storage_class.0.reclaim_policy", "delete"),
					resource.TestCheckResourceAttr(clusterName, "default_storage_class.0.filesystem", "ext4"),
					resource.TestCheckResourceAttr(clusterName, "pods_cidr", "100.96.0.0/11"),
					resource.TestCheckResourceAttr(clusterName, "services_cidr", "100.64.0.0/13"),
					resource.TestCheckResourceAttr(clusterName, "virtual_ip_subnet", ""),
					func() resource.TestCheckFunc {
						// Auto Repair on Errors gets automatically deactivated after cluster creation since CSE 4.1.1
						if cseVersion.GreaterThanOrEqual(v411) {
							return resource.TestCheckResourceAttr(clusterName, "auto_repair_on_errors", "false")
						} else {
							return resource.TestCheckResourceAttr(clusterName, "auto_repair_on_errors", "true")
						}
					}(),
					resource.TestCheckResourceAttr(clusterName, "node_health_check", "true"),
					resource.TestMatchResourceAttr(clusterName, "kubernetes_version", regexp.MustCompile(`^v[0-9]+\.[0-9]+\.[0-9]+\+vmware\.[0-9]$`)),
					resource.TestMatchResourceAttr(clusterName, "tkg_product_version", regexp.MustCompile(`^v[0-9]+\.[0-9]+\.[0-9]+$`)),
					resource.TestMatchResourceAttr(clusterName, "capvcd_version", regexp.MustCompile(`^v[0-9]+\.[0-9]+\.[0-9]+$`)),
					resource.TestMatchResourceAttr(clusterName, "cluster_resource_set_bindings.#", regexp.MustCompile(`^[1-9][0-9]*$`)),
					resource.TestMatchResourceAttr(clusterName, "cpi_version", regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+$`)),
					resource.TestMatchResourceAttr(clusterName, "csi_version", regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+$`)),
					resource.TestCheckResourceAttr(clusterName, "state", "provisioned"),
					resource.TestCheckResourceAttrSet(clusterName, "kubeconfig"),
					resource.TestMatchResourceAttr(clusterName, "events.#", regexp.MustCompile(`^[1-9][0-9]*$`)),
				),
			},
			// Scale down the control plane to 1 replica, scale up worker pool to 2 replicas
			{
				Config: step3,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Changed settings
					resource.TestCheckResourceAttr(clusterName, "control_plane.0.machine_count", "1"),
					resource.TestCheckResourceAttr(clusterName, "worker_pool.0.machine_count", "2"),

					// Other things should remain the same
					cacheId.testCheckCachedResourceFieldValue(clusterName, "id"),
					resource.TestCheckResourceAttr(clusterName, "cse_version", testConfig.Cse.Version),
					resource.TestCheckResourceAttr(clusterName, "runtime", "tkg"),
					resource.TestCheckResourceAttr(clusterName, "name", strings.ToLower(t.Name())),
					resource.TestCheckResourceAttrPair(clusterName, "kubernetes_template_id", "data.vcd_catalog_vapp_template.tkg_ova", "id"),
					resource.TestCheckResourceAttrPair(clusterName, "org", "data.vcd_org_vdc.vdc", "org"),
					resource.TestCheckResourceAttrPair(clusterName, "vdc_id", "data.vcd_org_vdc.vdc", "id"),
					resource.TestCheckResourceAttrPair(clusterName, "network_id", "data.vcd_network_routed_v2.routed", "id"),
					resource.TestCheckNoResourceAttr(clusterName, "owner"), // It is taken from Provider config
					resource.TestCheckResourceAttr(clusterName, "ssh_public_key", ""),
					resource.TestCheckResourceAttr(clusterName, "control_plane.0.disk_size_gi", "20"),
					resource.TestCheckResourceAttrPair(clusterName, "control_plane.0.sizing_policy_id", "data.vcd_vm_sizing_policy.tkg_small", "id"),
					resource.TestCheckResourceAttrPair(clusterName, "control_plane.0.storage_profile_id", "data.vcd_storage_profile.sp", "id"),
					resource.TestCheckResourceAttrSet(clusterName, "control_plane.0.ip"), // IP should be assigned after creation as it was not set manually in HCL config
					resource.TestCheckResourceAttr(clusterName, "worker_pool.#", "1"),
					resource.TestCheckResourceAttr(clusterName, "worker_pool.0.name", "worker-pool-1"),
					resource.TestCheckResourceAttr(clusterName, "worker_pool.0.disk_size_gi", "20"),
					resource.TestCheckResourceAttrPair(clusterName, "worker_pool.0.sizing_policy_id", "data.vcd_vm_sizing_policy.tkg_small", "id"),
					resource.TestCheckResourceAttrPair(clusterName, "worker_pool.0.storage_profile_id", "data.vcd_storage_profile.sp", "id"),
					resource.TestCheckResourceAttrPair(clusterName, "default_storage_class.0.storage_profile_id", "data.vcd_storage_profile.sp", "id"),
					resource.TestCheckResourceAttr(clusterName, "default_storage_class.0.name", "sc-1"),
					resource.TestCheckResourceAttr(clusterName, "default_storage_class.0.reclaim_policy", "delete"),
					resource.TestCheckResourceAttr(clusterName, "default_storage_class.0.filesystem", "ext4"),
					resource.TestCheckResourceAttr(clusterName, "pods_cidr", "100.96.0.0/11"),
					resource.TestCheckResourceAttr(clusterName, "services_cidr", "100.64.0.0/13"),
					resource.TestCheckResourceAttr(clusterName, "virtual_ip_subnet", ""),
					func() resource.TestCheckFunc {
						// Auto Repair on Errors gets automatically deactivated after cluster creation since CSE 4.1.1
						if cseVersion.GreaterThanOrEqual(v411) {
							return resource.TestCheckResourceAttr(clusterName, "auto_repair_on_errors", "false")
						} else {
							return resource.TestCheckResourceAttr(clusterName, "auto_repair_on_errors", "true")
						}
					}(),
					resource.TestCheckResourceAttr(clusterName, "node_health_check", "true"),
					resource.TestMatchResourceAttr(clusterName, "kubernetes_version", regexp.MustCompile(`^v[0-9]+\.[0-9]+\.[0-9]+\+vmware\.[0-9]$`)),
					resource.TestMatchResourceAttr(clusterName, "tkg_product_version", regexp.MustCompile(`^v[0-9]+\.[0-9]+\.[0-9]+$`)),
					resource.TestMatchResourceAttr(clusterName, "capvcd_version", regexp.MustCompile(`^v[0-9]+\.[0-9]+\.[0-9]+$`)),
					resource.TestMatchResourceAttr(clusterName, "cluster_resource_set_bindings.#", regexp.MustCompile(`^[1-9][0-9]*$`)),
					resource.TestMatchResourceAttr(clusterName, "cpi_version", regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+$`)),
					resource.TestMatchResourceAttr(clusterName, "csi_version", regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+$`)),
					resource.TestCheckResourceAttr(clusterName, "state", "provisioned"),
					resource.TestCheckResourceAttrSet(clusterName, "kubeconfig"),
					resource.TestMatchResourceAttr(clusterName, "events.#", regexp.MustCompile(`^[1-9][0-9]*$`)),
				),
			},
			// Scale down the worker pool to 1 replica, disable health check
			{
				Config: step4,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Changed settings
					resource.TestCheckResourceAttr(clusterName, "worker_pool.0.machine_count", "1"),
					resource.TestCheckResourceAttr(clusterName, "node_health_check", "false"),

					// Other things should remain the same
					cacheId.testCheckCachedResourceFieldValue(clusterName, "id"),
					resource.TestCheckResourceAttr(clusterName, "cse_version", testConfig.Cse.Version),
					resource.TestCheckResourceAttr(clusterName, "runtime", "tkg"),
					resource.TestCheckResourceAttr(clusterName, "name", strings.ToLower(t.Name())),
					resource.TestCheckResourceAttrPair(clusterName, "kubernetes_template_id", "data.vcd_catalog_vapp_template.tkg_ova", "id"),
					resource.TestCheckResourceAttrPair(clusterName, "org", "data.vcd_org_vdc.vdc", "org"),
					resource.TestCheckResourceAttrPair(clusterName, "vdc_id", "data.vcd_org_vdc.vdc", "id"),
					resource.TestCheckResourceAttrPair(clusterName, "network_id", "data.vcd_network_routed_v2.routed", "id"),
					resource.TestCheckNoResourceAttr(clusterName, "owner"), // It is taken from Provider config
					resource.TestCheckResourceAttr(clusterName, "ssh_public_key", ""),
					resource.TestCheckResourceAttr(clusterName, "control_plane.0.machine_count", "1"),
					resource.TestCheckResourceAttr(clusterName, "control_plane.0.disk_size_gi", "20"),
					resource.TestCheckResourceAttrPair(clusterName, "control_plane.0.sizing_policy_id", "data.vcd_vm_sizing_policy.tkg_small", "id"),
					resource.TestCheckResourceAttrPair(clusterName, "control_plane.0.storage_profile_id", "data.vcd_storage_profile.sp", "id"),
					resource.TestCheckResourceAttrSet(clusterName, "control_plane.0.ip"), // IP should be assigned after creation as it was not set manually in HCL config
					resource.TestCheckResourceAttr(clusterName, "worker_pool.#", "1"),
					resource.TestCheckResourceAttr(clusterName, "worker_pool.0.name", "worker-pool-1"),
					resource.TestCheckResourceAttr(clusterName, "worker_pool.0.disk_size_gi", "20"),
					resource.TestCheckResourceAttrPair(clusterName, "worker_pool.0.sizing_policy_id", "data.vcd_vm_sizing_policy.tkg_small", "id"),
					resource.TestCheckResourceAttrPair(clusterName, "worker_pool.0.storage_profile_id", "data.vcd_storage_profile.sp", "id"),
					resource.TestCheckResourceAttrPair(clusterName, "default_storage_class.0.storage_profile_id", "data.vcd_storage_profile.sp", "id"),
					resource.TestCheckResourceAttr(clusterName, "default_storage_class.0.name", "sc-1"),
					resource.TestCheckResourceAttr(clusterName, "default_storage_class.0.reclaim_policy", "delete"),
					resource.TestCheckResourceAttr(clusterName, "default_storage_class.0.filesystem", "ext4"),
					resource.TestCheckResourceAttr(clusterName, "pods_cidr", "100.96.0.0/11"),
					resource.TestCheckResourceAttr(clusterName, "services_cidr", "100.64.0.0/13"),
					resource.TestCheckResourceAttr(clusterName, "virtual_ip_subnet", ""),
					func() resource.TestCheckFunc {
						// Auto Repair on Errors gets automatically deactivated after cluster creation since CSE 4.1.1
						if cseVersion.GreaterThanOrEqual(v411) {
							return resource.TestCheckResourceAttr(clusterName, "auto_repair_on_errors", "false")
						} else {
							return resource.TestCheckResourceAttr(clusterName, "auto_repair_on_errors", "true")
						}
					}(),
					resource.TestMatchResourceAttr(clusterName, "kubernetes_version", regexp.MustCompile(`^v[0-9]+\.[0-9]+\.[0-9]+\+vmware\.[0-9]$`)),
					resource.TestMatchResourceAttr(clusterName, "tkg_product_version", regexp.MustCompile(`^v[0-9]+\.[0-9]+\.[0-9]+$`)),
					resource.TestMatchResourceAttr(clusterName, "capvcd_version", regexp.MustCompile(`^v[0-9]+\.[0-9]+\.[0-9]+$`)),
					resource.TestMatchResourceAttr(clusterName, "cluster_resource_set_bindings.#", regexp.MustCompile(`^[1-9][0-9]*$`)),
					resource.TestMatchResourceAttr(clusterName, "cpi_version", regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+$`)),
					resource.TestMatchResourceAttr(clusterName, "csi_version", regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+$`)),
					resource.TestCheckResourceAttr(clusterName, "state", "provisioned"),
					resource.TestCheckResourceAttrSet(clusterName, "kubeconfig"),
					resource.TestMatchResourceAttr(clusterName, "events.#", regexp.MustCompile(`^[1-9][0-9]*$`)),
				),
			},
			// Enable health check, add a worker pool
			{
				Config: step5,
				Check: resource.ComposeAggregateTestCheckFunc(
					// The new worker pool should be present
					resource.TestCheckResourceAttr(clusterName, "worker_pool.#", "2"),
					resource.TestCheckResourceAttr(clusterName, "worker_pool.1.machine_count", "1"),
					resource.TestCheckResourceAttr(clusterName, "worker_pool.1.name", "worker-pool-2"),
					resource.TestCheckResourceAttr(clusterName, "worker_pool.1.disk_size_gi", "20"),
					resource.TestCheckResourceAttrPair(clusterName, "worker_pool.1.sizing_policy_id", "data.vcd_vm_sizing_policy.tkg_small", "id"),
					resource.TestCheckResourceAttrPair(clusterName, "worker_pool.1.storage_profile_id", "data.vcd_storage_profile.sp", "id"),
					resource.TestCheckResourceAttr(clusterName, "node_health_check", "true"),

					// Other things should remain the same
					cacheId.testCheckCachedResourceFieldValue(clusterName, "id"),
					resource.TestCheckResourceAttr(clusterName, "cse_version", testConfig.Cse.Version),
					resource.TestCheckResourceAttr(clusterName, "runtime", "tkg"),
					resource.TestCheckResourceAttr(clusterName, "name", strings.ToLower(t.Name())),
					resource.TestCheckResourceAttrPair(clusterName, "kubernetes_template_id", "data.vcd_catalog_vapp_template.tkg_ova", "id"),
					resource.TestCheckResourceAttrPair(clusterName, "org", "data.vcd_org_vdc.vdc", "org"),
					resource.TestCheckResourceAttrPair(clusterName, "vdc_id", "data.vcd_org_vdc.vdc", "id"),
					resource.TestCheckResourceAttrPair(clusterName, "network_id", "data.vcd_network_routed_v2.routed", "id"),
					resource.TestCheckNoResourceAttr(clusterName, "owner"), // It is taken from Provider config
					resource.TestCheckResourceAttr(clusterName, "ssh_public_key", ""),
					resource.TestCheckResourceAttr(clusterName, "control_plane.0.machine_count", "1"),
					resource.TestCheckResourceAttr(clusterName, "control_plane.0.disk_size_gi", "20"),
					resource.TestCheckResourceAttrPair(clusterName, "control_plane.0.sizing_policy_id", "data.vcd_vm_sizing_policy.tkg_small", "id"),
					resource.TestCheckResourceAttrPair(clusterName, "control_plane.0.storage_profile_id", "data.vcd_storage_profile.sp", "id"),
					resource.TestCheckResourceAttrSet(clusterName, "control_plane.0.ip"), // IP should be assigned after creation as it was not set manually in HCL config
					resource.TestCheckResourceAttr(clusterName, "worker_pool.0.machine_count", "1"),
					resource.TestCheckResourceAttr(clusterName, "worker_pool.0.name", "worker-pool-1"),
					resource.TestCheckResourceAttr(clusterName, "worker_pool.0.disk_size_gi", "20"),
					resource.TestCheckResourceAttrPair(clusterName, "worker_pool.0.sizing_policy_id", "data.vcd_vm_sizing_policy.tkg_small", "id"),
					resource.TestCheckResourceAttrPair(clusterName, "worker_pool.0.storage_profile_id", "data.vcd_storage_profile.sp", "id"),
					resource.TestCheckResourceAttrPair(clusterName, "default_storage_class.0.storage_profile_id", "data.vcd_storage_profile.sp", "id"),
					resource.TestCheckResourceAttr(clusterName, "default_storage_class.0.name", "sc-1"),
					resource.TestCheckResourceAttr(clusterName, "default_storage_class.0.reclaim_policy", "delete"),
					resource.TestCheckResourceAttr(clusterName, "default_storage_class.0.filesystem", "ext4"),
					resource.TestCheckResourceAttr(clusterName, "pods_cidr", "100.96.0.0/11"),
					resource.TestCheckResourceAttr(clusterName, "services_cidr", "100.64.0.0/13"),
					resource.TestCheckResourceAttr(clusterName, "virtual_ip_subnet", ""),
					func() resource.TestCheckFunc {
						// Auto Repair on Errors gets automatically deactivated after cluster creation since CSE 4.1.1
						if cseVersion.GreaterThanOrEqual(v411) {
							return resource.TestCheckResourceAttr(clusterName, "auto_repair_on_errors", "false")
						} else {
							return resource.TestCheckResourceAttr(clusterName, "auto_repair_on_errors", "true")
						}
					}(),
					resource.TestMatchResourceAttr(clusterName, "kubernetes_version", regexp.MustCompile(`^v[0-9]+\.[0-9]+\.[0-9]+\+vmware\.[0-9]$`)),
					resource.TestMatchResourceAttr(clusterName, "tkg_product_version", regexp.MustCompile(`^v[0-9]+\.[0-9]+\.[0-9]+$`)),
					resource.TestMatchResourceAttr(clusterName, "capvcd_version", regexp.MustCompile(`^v[0-9]+\.[0-9]+\.[0-9]+$`)),
					resource.TestMatchResourceAttr(clusterName, "cluster_resource_set_bindings.#", regexp.MustCompile(`^[1-9][0-9]*$`)),
					resource.TestMatchResourceAttr(clusterName, "cpi_version", regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+$`)),
					resource.TestMatchResourceAttr(clusterName, "csi_version", regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+$`)),
					resource.TestCheckResourceAttr(clusterName, "state", "provisioned"),
					resource.TestCheckResourceAttrSet(clusterName, "kubeconfig"),
					resource.TestMatchResourceAttr(clusterName, "events.#", regexp.MustCompile(`^[1-9][0-9]*$`)),
				),
			},
			// Test data sources
			{
				Config: step6,
				Check: resource.ComposeAggregateTestCheckFunc(
					resourceFieldsEqual(clusterName, "data.vcd_cse_kubernetes_cluster.with_id_ds", []string{"org_id", "org", "operations_timeout_minutes", "api_token_file", "events"}),
					resourceFieldsEqual(clusterName, "data.vcd_cse_kubernetes_cluster.with_name_ds", []string{"org_id", "org", "operations_timeout_minutes", "api_token_file", "events"}),
					resourceFieldsEqual("data.vcd_cse_kubernetes_cluster.with_id_ds", "data.vcd_cse_kubernetes_cluster.with_name_ds", []string{"events"}),
				),
			},
			{
				ResourceName:      clusterName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					return cacheId.fieldValue, nil
				},
				// Ignore api_token_file and operations_timeout_minutes as these are not computed from VCD, so they are missing
				// after any successful import.
				// Ignore also owner and org as these may not be set in the resource configuration, but they are always
				// set on imports.
				ImportStateVerifyIgnore: []string{"api_token_file", "operations_timeout_minutes", "owner", "org"},
			},
		},
	})
	postTestChecks(t)
}

func TestAccVcdCseKubernetesClusterFailure(t *testing.T) {
	preTestChecks(t)
	requireCseConfig(t, testConfig)

	vcdClient := createSystemTemporaryVCDConnection()

	cseVersion, err := semver.NewVersion(testConfig.Cse.Version)
	if err != nil {
		t.Fatal(err)
	}

	tokenFilename := getCurrentDir() + t.Name() + ".json"
	defer func() {
		// Clean the API Token file
		if fileExists(tokenFilename) {
			err := os.Remove(tokenFilename)
			if err != nil {
				fmt.Printf("could not delete API token file '%s', please delete it manually", tokenFilename)
			}
		}
	}()

	clusterName := "cse-k8s-cluster-failure" // We can't use the test name as it is too long

	var params = StringMap{
		"CseVersion":         testConfig.Cse.Version,
		"Name":               clusterName,
		"OvaCatalog":         testConfig.Cse.OvaCatalog,
		"OvaName":            testConfig.Cse.OvaName,
		"KubernetesOva":      "data.vcd_catalog_vapp_template.tkg_ova.id",
		"SolutionsOrg":       testConfig.Cse.SolutionsOrg,
		"TenantOrg":          testConfig.Cse.TenantOrg,
		"Vdc":                testConfig.Cse.TenantVdc,
		"EdgeGateway":        testConfig.Cse.EdgeGateway,
		"Network":            testConfig.Cse.RoutedNetwork,
		"TokenName":          t.Name() + "3",
		"TokenFile":          tokenFilename,
		"ControlPlaneCount":  1,
		"NodePoolCount":      1,
		"ExtraWorkerPool":    " ",
		"PodsCidr":           "1.2.3.4/24", // This will make the cluster to fail
		"ServicesCidr":       "5.6.7.8/24", // This will make the cluster to fail
		"AutoRepairOnErrors": false,        // This must be false
		"NodeHealthCheck":    false,
		"Timeout":            150,
	}
	testParamsNotEmpty(t, params)

	step1 := templateFill(testAccVcdCseKubernetesCluster, params)
	debugPrintf("#[DEBUG] CONFIGURATION step1: %s", step1)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			org, err := vcdClient.GetOrgByName(testConfig.Cse.TenantOrg)
			if err != nil {
				return fmt.Errorf("could not check cluster deletion: %s", err)
			}
			clusters, err := org.CseGetKubernetesClustersByName(*cseVersion, clusterName)
			if err != nil && !govcd.ContainsNotFound(err) {
				return fmt.Errorf("could not check cluster deletion: %s", err)
			}
			if len(clusters) == 0 || govcd.ContainsNotFound(err) {
				return nil
			}
			return fmt.Errorf("there are still %d clusters with name '%s': %s", len(clusters), clusterName, err)
		},
		Steps: []resource.TestStep{
			{
				Config:      step1,
				ExpectError: regexp.MustCompile(`Kubernetes cluster creation finished, but it is not in 'provisioned' state`),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdCseKubernetesCluster = `
# skip-binary-test - This one requires a very special setup

data "vcd_catalog" "tkg_catalog" {
  org  = "{{.SolutionsOrg}}"
  name = "{{.OvaCatalog}}"
}

data "vcd_catalog_vapp_template" "tkg_ova" {
  org        = data.vcd_catalog.tkg_catalog.org
  catalog_id = data.vcd_catalog.tkg_catalog.id
  name       = "{{.OvaName}}"
}

data "vcd_org_vdc" "vdc" {
  org  = "{{.TenantOrg}}"
  name = "{{.Vdc}}"
}

data "vcd_nsxt_edgegateway" "egw" {
  org      = data.vcd_org_vdc.vdc.org
  owner_id = data.vcd_org_vdc.vdc.id
  name     = "{{.EdgeGateway}}"
}

data "vcd_network_routed_v2" "routed" {
  org             = data.vcd_nsxt_edgegateway.egw.org
  edge_gateway_id = data.vcd_nsxt_edgegateway.egw.id
  name            = "{{.Network}}"
}

data "vcd_vm_sizing_policy" "tkg_small" {
  name = "TKG small"
}

data "vcd_storage_profile" "sp" {
  org  = data.vcd_org_vdc.vdc.org
  vdc  = data.vcd_org_vdc.vdc.name
  name = "*"
}

resource "vcd_api_token" "token" {
  name             = "{{.TokenName}}"
  file_name        = "{{.TokenFile}}"
  allow_token_file = true
}

resource "vcd_cse_kubernetes_cluster" "my_cluster" {
  cse_version            = "{{.CseVersion}}"
  runtime                = "tkg"
  name                   = "{{.Name}}"
  kubernetes_template_id = {{.KubernetesOva}}
  org                    = data.vcd_org_vdc.vdc.org
  vdc_id                 = data.vcd_org_vdc.vdc.id
  network_id             = data.vcd_network_routed_v2.routed.id
  api_token_file	     = vcd_api_token.token.file_name

  control_plane {
    machine_count      = {{.ControlPlaneCount}}
    disk_size_gi       = 20
    sizing_policy_id   = data.vcd_vm_sizing_policy.tkg_small.id
    storage_profile_id = data.vcd_storage_profile.sp.id
  }

  worker_pool {
    name               = "worker-pool-1"
    machine_count      = {{.NodePoolCount}}
    disk_size_gi       = 20
    sizing_policy_id   = data.vcd_vm_sizing_policy.tkg_small.id
    storage_profile_id = data.vcd_storage_profile.sp.id
  }

  {{.ExtraWorkerPool}}

  default_storage_class {
	name               = "sc-1"
	storage_profile_id = data.vcd_storage_profile.sp.id
    reclaim_policy     = "delete"
    filesystem         = "ext4"
  }

  pods_cidr     = "{{.PodsCidr}}"
  services_cidr = "{{.ServicesCidr}}"

  auto_repair_on_errors = {{.AutoRepairOnErrors}}
  node_health_check     = {{.NodeHealthCheck}}

  operations_timeout_minutes = {{.Timeout}}
}
`

const testAccVcdCseKubernetesClusterDS = `
# skip-binary-test - This one requires a very special setup

data "vcd_org" "tenant_org" {
  name = "tenant_org"
}

data "vcd_cse_kubernetes_cluster" "with_id_ds" {
  cluster_id = vcd_cse_kubernetes_cluster.my_cluster.id
}

data "vcd_cse_kubernetes_cluster" "with_name_ds" {
  org_id      = data.vcd_org.tenant_org.id
  cse_version = vcd_cse_kubernetes_cluster.my_cluster.cse_version
  name        = vcd_cse_kubernetes_cluster.my_cluster.name
}
`
