//go:build cse || ALL || functional

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

	sshPublicKey := "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCrCI+QkLjgQVqR7c7dJfawJqCslVomo5I25JdolqlteX7RCUq0yncWyS+8MTYWCS03sm1jOroLOeuji8CDKCDCcKwQerJiOFoJS+VOK5xCjJ2u8RBGlIpXNcmIh2VriRJrV7TCKrFMSKLNF4/n83q4gWI/YPf6/dRhpPB72HYrdI4omvRlU4GG09jMmgiz+5Yb8wJEXYMsJni+MwPzFKe6TbMcqjBusDyeFGAhgyN7QJGpdNhAn1sqvqZrW2QjaE8P+4t8RzBo8B2ucyQazd6+lbYmOHq9366LjG160snzXrFzlARc4hhpjMzu9Bcm6i3ZZI70qhIbmi5IonbbVh8t"

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
		"SshPublicKey":       sshPublicKey,
		"AutoRepairOnErrors": true,
		"NodeHealthCheck":    true,
		"Timeout":            150,
		"Autoscaler":         " ",
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
	params["NodePoolCount"] = 0
	params["Autoscaler"] = "    autoscaler_max_replicas = 5\n    autoscaler_min_replicas = 1"
	step6 := templateFill(testAccVcdCseKubernetesCluster, params)
	debugPrintf("#[DEBUG] CONFIGURATION step6: %s", step6)

	params["FuncName"] = t.Name() + "Step7"
	step7 := templateFill(testAccVcdCseKubernetesCluster+testAccVcdCseKubernetesClusterDS, params)
	debugPrintf("#[DEBUG] CONFIGURATION step7: %s", step7)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	vcdClient := createSystemTemporaryVCDConnection()
	cacheId := testCachedFieldValue{}
	clusterName := "vcd_cse_kubernetes_cluster.my_cluster"
	dataWithName := "data.vcd_cse_kubernetes_cluster.with_name_ds"
	dataWithId := "data.vcd_cse_kubernetes_cluster.with_id_ds"
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			org, err := vcdClient.GetOrgByName(testConfig.Cse.TenantOrg)
			if err != nil {
				return fmt.Errorf("could not check cluster deletion: %s", err)
			}
			clusters, err := org.CseGetKubernetesClustersByName(*cseVersion, strings.ToLower(t.Name()))
			if err != nil && !govcd.ContainsNotFound(err) {
				return fmt.Errorf("could not check cluster deletion: %s", err)
			}
			if len(clusters) == 0 || govcd.ContainsNotFound(err) {
				return nil
			}
			return fmt.Errorf("there are still %d clusters with name '%s': %s", len(clusters), clusterName, err)
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
					resource.TestCheckResourceAttr(clusterName, "ssh_public_key", sshPublicKey),
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
					resource.TestCheckResourceAttr(clusterName, "worker_pool.0.autoscaler_max_replicas", "0"),
					resource.TestCheckResourceAttr(clusterName, "worker_pool.0.autoscaler_min_replicas", "0"),
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
					resource.TestCheckResourceAttr(clusterName, "ssh_public_key", sshPublicKey),
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
					resource.TestCheckResourceAttr(clusterName, "worker_pool.0.autoscaler_max_replicas", "0"),
					resource.TestCheckResourceAttr(clusterName, "worker_pool.0.autoscaler_min_replicas", "0"),
					resource.TestCheckResourceAttrPair(clusterName, "default_storage_class.0.storage_profile_id", "data.vcd_storage_profile.sp", "id"),
					resource.TestCheckResourceAttr(clusterName, "default_storage_class.0.name", "sc-1"),
					resource.TestCheckResourceAttr(clusterName, "default_storage_class.0.reclaim_policy", "delete"),
					resource.TestCheckResourceAttr(clusterName, "default_storage_class.0.filesystem", "ext4"),
					resource.TestCheckResourceAttr(clusterName, "pods_cidr", "100.96.0.0/11"),
					resource.TestCheckResourceAttr(clusterName, "services_cidr", "100.64.0.0/13"),
					resource.TestCheckResourceAttr(clusterName, "virtual_ip_subnet", ""),
					resource.TestCheckResourceAttr(clusterName, "auto_repair_on_errors", "false"),
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
					resource.TestCheckResourceAttr(clusterName, "ssh_public_key", sshPublicKey),
					resource.TestCheckResourceAttr(clusterName, "control_plane.0.disk_size_gi", "20"),
					resource.TestCheckResourceAttrPair(clusterName, "control_plane.0.sizing_policy_id", "data.vcd_vm_sizing_policy.tkg_small", "id"),
					resource.TestCheckResourceAttrPair(clusterName, "control_plane.0.storage_profile_id", "data.vcd_storage_profile.sp", "id"),
					resource.TestCheckResourceAttrSet(clusterName, "control_plane.0.ip"), // IP should be assigned after creation as it was not set manually in HCL config
					resource.TestCheckResourceAttr(clusterName, "worker_pool.#", "1"),
					resource.TestCheckResourceAttr(clusterName, "worker_pool.0.name", "worker-pool-1"),
					resource.TestCheckResourceAttr(clusterName, "worker_pool.0.disk_size_gi", "20"),
					resource.TestCheckResourceAttrPair(clusterName, "worker_pool.0.sizing_policy_id", "data.vcd_vm_sizing_policy.tkg_small", "id"),
					resource.TestCheckResourceAttrPair(clusterName, "worker_pool.0.storage_profile_id", "data.vcd_storage_profile.sp", "id"),
					resource.TestCheckResourceAttr(clusterName, "worker_pool.0.autoscaler_max_replicas", "0"),
					resource.TestCheckResourceAttr(clusterName, "worker_pool.0.autoscaler_min_replicas", "0"),
					resource.TestCheckResourceAttrPair(clusterName, "default_storage_class.0.storage_profile_id", "data.vcd_storage_profile.sp", "id"),
					resource.TestCheckResourceAttr(clusterName, "default_storage_class.0.name", "sc-1"),
					resource.TestCheckResourceAttr(clusterName, "default_storage_class.0.reclaim_policy", "delete"),
					resource.TestCheckResourceAttr(clusterName, "default_storage_class.0.filesystem", "ext4"),
					resource.TestCheckResourceAttr(clusterName, "pods_cidr", "100.96.0.0/11"),
					resource.TestCheckResourceAttr(clusterName, "services_cidr", "100.64.0.0/13"),
					resource.TestCheckResourceAttr(clusterName, "virtual_ip_subnet", ""),
					resource.TestCheckResourceAttr(clusterName, "auto_repair_on_errors", "false"),
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
					resource.TestCheckResourceAttr(clusterName, "ssh_public_key", sshPublicKey),
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
					resource.TestCheckResourceAttr(clusterName, "worker_pool.0.autoscaler_max_replicas", "0"),
					resource.TestCheckResourceAttr(clusterName, "worker_pool.0.autoscaler_min_replicas", "0"),
					resource.TestCheckResourceAttrPair(clusterName, "default_storage_class.0.storage_profile_id", "data.vcd_storage_profile.sp", "id"),
					resource.TestCheckResourceAttr(clusterName, "default_storage_class.0.name", "sc-1"),
					resource.TestCheckResourceAttr(clusterName, "default_storage_class.0.reclaim_policy", "delete"),
					resource.TestCheckResourceAttr(clusterName, "default_storage_class.0.filesystem", "ext4"),
					resource.TestCheckResourceAttr(clusterName, "pods_cidr", "100.96.0.0/11"),
					resource.TestCheckResourceAttr(clusterName, "services_cidr", "100.64.0.0/13"),
					resource.TestCheckResourceAttr(clusterName, "virtual_ip_subnet", ""),
					resource.TestCheckResourceAttr(clusterName, "auto_repair_on_errors", "false"),
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
					resource.TestCheckResourceAttr(clusterName, "worker_pool.1.autoscaler_max_replicas", "0"),
					resource.TestCheckResourceAttr(clusterName, "worker_pool.1.autoscaler_min_replicas", "0"),
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
					resource.TestCheckResourceAttr(clusterName, "ssh_public_key", sshPublicKey),
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
					resource.TestCheckResourceAttr(clusterName, "worker_pool.0.autoscaler_max_replicas", "0"),
					resource.TestCheckResourceAttr(clusterName, "worker_pool.0.autoscaler_min_replicas", "0"),
					resource.TestCheckResourceAttrPair(clusterName, "default_storage_class.0.storage_profile_id", "data.vcd_storage_profile.sp", "id"),
					resource.TestCheckResourceAttr(clusterName, "default_storage_class.0.name", "sc-1"),
					resource.TestCheckResourceAttr(clusterName, "default_storage_class.0.reclaim_policy", "delete"),
					resource.TestCheckResourceAttr(clusterName, "default_storage_class.0.filesystem", "ext4"),
					resource.TestCheckResourceAttr(clusterName, "pods_cidr", "100.96.0.0/11"),
					resource.TestCheckResourceAttr(clusterName, "services_cidr", "100.64.0.0/13"),
					resource.TestCheckResourceAttr(clusterName, "virtual_ip_subnet", ""),
					resource.TestCheckResourceAttr(clusterName, "auto_repair_on_errors", "false"),
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
			// Change first worker pool to use the Autoscaler
			{
				Config: step6,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Autoscaler should be enabled now
					resource.TestCheckResourceAttr(clusterName, "worker_pool.0.machine_count", "0"),
					resource.TestCheckResourceAttr(clusterName, "worker_pool.0.autoscaler_max_replicas", "5"),
					resource.TestCheckResourceAttr(clusterName, "worker_pool.0.autoscaler_min_replicas", "1"),

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
					resource.TestCheckResourceAttr(clusterName, "ssh_public_key", sshPublicKey),
					resource.TestCheckResourceAttr(clusterName, "control_plane.0.machine_count", "1"),
					resource.TestCheckResourceAttr(clusterName, "control_plane.0.disk_size_gi", "20"),
					resource.TestCheckResourceAttrPair(clusterName, "control_plane.0.sizing_policy_id", "data.vcd_vm_sizing_policy.tkg_small", "id"),
					resource.TestCheckResourceAttrPair(clusterName, "control_plane.0.storage_profile_id", "data.vcd_storage_profile.sp", "id"),
					resource.TestCheckResourceAttrSet(clusterName, "control_plane.0.ip"), // IP should be assigned after creation as it was not set manually in HCL config
					resource.TestCheckResourceAttr(clusterName, "worker_pool.0.name", "worker-pool-1"),
					resource.TestCheckResourceAttr(clusterName, "worker_pool.0.disk_size_gi", "20"),
					resource.TestCheckResourceAttrPair(clusterName, "worker_pool.0.sizing_policy_id", "data.vcd_vm_sizing_policy.tkg_small", "id"),
					resource.TestCheckResourceAttrPair(clusterName, "worker_pool.0.storage_profile_id", "data.vcd_storage_profile.sp", "id"),
					resource.TestCheckResourceAttr(clusterName, "worker_pool.#", "2"),
					resource.TestCheckResourceAttr(clusterName, "worker_pool.1.machine_count", "1"),
					resource.TestCheckResourceAttr(clusterName, "worker_pool.1.name", "worker-pool-2"),
					resource.TestCheckResourceAttr(clusterName, "worker_pool.1.disk_size_gi", "20"),
					resource.TestCheckResourceAttr(clusterName, "worker_pool.1.autoscaler_max_replicas", "0"),
					resource.TestCheckResourceAttr(clusterName, "worker_pool.1.autoscaler_min_replicas", "0"),
					resource.TestCheckResourceAttrPair(clusterName, "worker_pool.1.sizing_policy_id", "data.vcd_vm_sizing_policy.tkg_small", "id"),
					resource.TestCheckResourceAttrPair(clusterName, "worker_pool.1.storage_profile_id", "data.vcd_storage_profile.sp", "id"),
					resource.TestCheckResourceAttr(clusterName, "node_health_check", "true"),
					resource.TestCheckResourceAttrPair(clusterName, "default_storage_class.0.storage_profile_id", "data.vcd_storage_profile.sp", "id"),
					resource.TestCheckResourceAttr(clusterName, "default_storage_class.0.name", "sc-1"),
					resource.TestCheckResourceAttr(clusterName, "default_storage_class.0.reclaim_policy", "delete"),
					resource.TestCheckResourceAttr(clusterName, "default_storage_class.0.filesystem", "ext4"),
					resource.TestCheckResourceAttr(clusterName, "pods_cidr", "100.96.0.0/11"),
					resource.TestCheckResourceAttr(clusterName, "services_cidr", "100.64.0.0/13"),
					resource.TestCheckResourceAttr(clusterName, "virtual_ip_subnet", ""),
					resource.TestCheckResourceAttr(clusterName, "auto_repair_on_errors", "false"),
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
			// Test data sources. Can't use resourceFieldsEqual function as we need to ignore the "events" TypeList which has an unknown size
			{
				Config: step7,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Data source with name
					resource.TestCheckResourceAttrPair(dataWithName, "id", clusterName, "id"),
					resource.TestCheckResourceAttrPair(dataWithName, "cse_version", clusterName, "cse_version"),
					resource.TestCheckResourceAttrPair(dataWithName, "runtime", clusterName, "runtime"),
					resource.TestCheckResourceAttrPair(dataWithName, "name", clusterName, "name"),
					resource.TestCheckResourceAttrPair(dataWithName, "kubernetes_template_id", clusterName, "kubernetes_template_id"),
					resource.TestMatchResourceAttr(dataWithName, "org_id", regexp.MustCompile(`^urn:vcloud:org:.+$`)),
					resource.TestCheckResourceAttrPair(dataWithName, "vdc_id", clusterName, "vdc_id"),
					resource.TestCheckResourceAttrPair(dataWithName, "network_id", clusterName, "network_id"),
					resource.TestCheckResourceAttrSet(dataWithName, "owner"), // This time the owner can be obtained
					resource.TestCheckResourceAttrPair(dataWithName, "ssh_public_key", clusterName, "ssh_public_key"),
					resource.TestCheckResourceAttrPair(dataWithName, "control_plane.0.disk_size_gi", clusterName, "control_plane.0.disk_size_gi"),
					resource.TestCheckResourceAttrPair(dataWithName, "control_plane.0.sizing_policy_id", clusterName, "control_plane.0.sizing_policy_id"),
					resource.TestCheckResourceAttrPair(dataWithName, "control_plane.0.storage_profile_id", clusterName, "control_plane.0.storage_profile_id"),
					resource.TestCheckResourceAttrPair(dataWithName, "control_plane.0.ip", clusterName, "control_plane.0.ip"),
					resource.TestCheckResourceAttrPair(dataWithName, "worker_pool.#", clusterName, "worker_pool.#"),
					resource.TestCheckResourceAttrPair(dataWithName, "worker_pool.0.name", clusterName, "worker_pool.0.name"),
					resource.TestCheckResourceAttrPair(dataWithName, "worker_pool.0.machine_count", clusterName, "worker_pool.0.machine_count"),
					resource.TestCheckResourceAttrPair(dataWithName, "worker_pool.0.disk_size_gi", clusterName, "worker_pool.0.disk_size_gi"),
					resource.TestCheckResourceAttrPair(dataWithName, "worker_pool.0.sizing_policy_id", clusterName, "worker_pool.0.sizing_policy_id"),
					resource.TestCheckResourceAttrPair(dataWithName, "worker_pool.0.storage_profile_id", clusterName, "worker_pool.0.storage_profile_id"),
					resource.TestCheckResourceAttrPair(dataWithName, "worker_pool.0.autoscaler_max_replicas", clusterName, "worker_pool.0.autoscaler_max_replicas"),
					resource.TestCheckResourceAttrPair(dataWithName, "worker_pool.0.autoscaler_min_replicas", clusterName, "worker_pool.0.autoscaler_min_replicas"),
					resource.TestCheckResourceAttrPair(dataWithName, "default_storage_class.0.storage_profile_id", clusterName, "default_storage_class.0.storage_profile_id"),
					resource.TestCheckResourceAttrPair(dataWithName, "default_storage_class.0.name", clusterName, "default_storage_class.0.name"),
					resource.TestCheckResourceAttrPair(dataWithName, "default_storage_class.0.reclaim_policy", clusterName, "default_storage_class.0.reclaim_policy"),
					resource.TestCheckResourceAttrPair(dataWithName, "default_storage_class.0.filesystem", clusterName, "default_storage_class.0.filesystem"),
					resource.TestCheckResourceAttrPair(dataWithName, "pods_cidr", clusterName, "pods_cidr"),
					resource.TestCheckResourceAttrPair(dataWithName, "services_cidr", clusterName, "services_cidr"),
					resource.TestCheckResourceAttrPair(dataWithName, "virtual_ip_subnet", clusterName, "virtual_ip_subnet"),
					resource.TestCheckResourceAttrPair(dataWithName, "auto_repair_on_errors", clusterName, "auto_repair_on_errors"),
					resource.TestCheckResourceAttrPair(dataWithName, "ssh_public_key", clusterName, "ssh_public_key"),
					resource.TestCheckResourceAttrPair(dataWithName, "node_health_check", clusterName, "node_health_check"),
					resource.TestCheckResourceAttrPair(dataWithName, "kubernetes_version", clusterName, "kubernetes_version"),
					resource.TestCheckResourceAttrPair(dataWithName, "tkg_product_version", clusterName, "tkg_product_version"),
					resource.TestCheckResourceAttrPair(dataWithName, "capvcd_version", clusterName, "capvcd_version"),
					resource.TestCheckResourceAttrPair(dataWithName, "cluster_resource_set_bindings.#", clusterName, "cluster_resource_set_bindings.#"),
					resource.TestCheckResourceAttrPair(dataWithName, "cpi_version", clusterName, "cpi_version"),
					resource.TestCheckResourceAttrPair(dataWithName, "csi_version", clusterName, "csi_version"),
					resource.TestCheckResourceAttrPair(dataWithName, "state", clusterName, "state"),
					resource.TestCheckResourceAttrPair(dataWithName, "kubeconfig", clusterName, "kubeconfig"),
					resource.TestMatchResourceAttr(dataWithName, "events.#", regexp.MustCompile(`^[1-9][0-9]*$`)),

					// Data source with ID
					resource.TestCheckResourceAttrPair(dataWithId, "id", dataWithName, "id"),
					resource.TestCheckResourceAttrPair(dataWithId, "cse_version", dataWithName, "cse_version"),
					resource.TestCheckResourceAttrPair(dataWithId, "runtime", dataWithName, "runtime"),
					resource.TestCheckResourceAttrPair(dataWithId, "name", dataWithName, "name"),
					resource.TestCheckResourceAttrPair(dataWithId, "kubernetes_template_id", dataWithName, "kubernetes_template_id"),
					resource.TestCheckResourceAttrPair(dataWithId, "org_id", dataWithName, "org_id"),
					resource.TestCheckResourceAttrPair(dataWithId, "vdc_id", dataWithName, "vdc_id"),
					resource.TestCheckResourceAttrPair(dataWithId, "network_id", dataWithName, "network_id"),
					resource.TestCheckResourceAttrPair(dataWithId, "owner", dataWithName, "owner"),
					resource.TestCheckResourceAttrPair(dataWithId, "ssh_public_key", dataWithName, "ssh_public_key"),
					resource.TestCheckResourceAttrPair(dataWithId, "control_plane.0.disk_size_gi", dataWithName, "control_plane.0.disk_size_gi"),
					resource.TestCheckResourceAttrPair(dataWithId, "control_plane.0.sizing_policy_id", dataWithName, "control_plane.0.sizing_policy_id"),
					resource.TestCheckResourceAttrPair(dataWithId, "control_plane.0.storage_profile_id", dataWithName, "control_plane.0.storage_profile_id"),
					resource.TestCheckResourceAttrPair(dataWithId, "control_plane.0.ip", dataWithName, "control_plane.0.ip"),
					resource.TestCheckResourceAttrPair(dataWithId, "worker_pool.#", dataWithName, "worker_pool.#"),
					resource.TestCheckResourceAttrPair(dataWithId, "worker_pool.0.name", dataWithName, "worker_pool.0.name"),
					resource.TestCheckResourceAttrPair(dataWithId, "worker_pool.0.machine_count", dataWithName, "worker_pool.0.machine_count"),
					resource.TestCheckResourceAttrPair(dataWithId, "worker_pool.0.disk_size_gi", dataWithName, "worker_pool.0.disk_size_gi"),
					resource.TestCheckResourceAttrPair(dataWithId, "worker_pool.0.sizing_policy_id", dataWithName, "worker_pool.0.sizing_policy_id"),
					resource.TestCheckResourceAttrPair(dataWithId, "worker_pool.0.storage_profile_id", dataWithName, "worker_pool.0.storage_profile_id"),
					resource.TestCheckResourceAttrPair(dataWithId, "worker_pool.0.autoscaler_max_replicas", dataWithName, "worker_pool.0.autoscaler_max_replicas"),
					resource.TestCheckResourceAttrPair(dataWithId, "worker_pool.0.autoscaler_min_replicas", dataWithName, "worker_pool.0.autoscaler_min_replicas"),
					resource.TestCheckResourceAttrPair(dataWithId, "default_storage_class.0.storage_profile_id", dataWithName, "default_storage_class.0.storage_profile_id"),
					resource.TestCheckResourceAttrPair(dataWithId, "default_storage_class.0.name", dataWithName, "default_storage_class.0.name"),
					resource.TestCheckResourceAttrPair(dataWithId, "default_storage_class.0.reclaim_policy", dataWithName, "default_storage_class.0.reclaim_policy"),
					resource.TestCheckResourceAttrPair(dataWithId, "default_storage_class.0.filesystem", dataWithName, "default_storage_class.0.filesystem"),
					resource.TestCheckResourceAttrPair(dataWithId, "pods_cidr", dataWithName, "pods_cidr"),
					resource.TestCheckResourceAttrPair(dataWithId, "services_cidr", dataWithName, "services_cidr"),
					resource.TestCheckResourceAttrPair(dataWithId, "virtual_ip_subnet", dataWithName, "virtual_ip_subnet"),
					resource.TestCheckResourceAttrPair(dataWithId, "auto_repair_on_errors", dataWithName, "auto_repair_on_errors"),
					resource.TestCheckResourceAttrPair(dataWithId, "node_health_check", dataWithName, "node_health_check"),
					resource.TestCheckResourceAttrPair(dataWithId, "ssh_public_key", dataWithName, "ssh_public_key"),
					resource.TestCheckResourceAttrPair(dataWithId, "kubernetes_version", dataWithName, "kubernetes_version"),
					resource.TestCheckResourceAttrPair(dataWithId, "tkg_product_version", dataWithName, "tkg_product_version"),
					resource.TestCheckResourceAttrPair(dataWithId, "capvcd_version", dataWithName, "capvcd_version"),
					resource.TestCheckResourceAttrPair(dataWithId, "cluster_resource_set_bindings.#", dataWithName, "cluster_resource_set_bindings.#"),
					resource.TestCheckResourceAttrPair(dataWithId, "cpi_version", dataWithName, "cpi_version"),
					resource.TestCheckResourceAttrPair(dataWithId, "csi_version", dataWithName, "csi_version"),
					resource.TestCheckResourceAttrPair(dataWithId, "state", dataWithName, "state"),
					resource.TestCheckResourceAttrPair(dataWithId, "kubeconfig", dataWithName, "kubeconfig"),
					resource.TestMatchResourceAttr(dataWithId, "events.#", regexp.MustCompile(`^[1-9][0-9]*$`)),
				),
			},
			{
				ResourceName:      clusterName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					return cacheId.fieldValue, nil
				},
				// Ignore 'api_token_file' and 'operations_timeout_minutes' as these are not computed from VCD, so they are missing
				// after any successful import.
				// Ignore also 'owner' and 'org' as these may not be set in the resource configuration, but they are always
				// set on imports.
				// 'events' is ignored as the list may differ between runs.
				ImportStateVerifyIgnore: []string{"api_token_file", "operations_timeout_minutes", "owner", "org", "events"},
			},
		},
	})
	postTestChecks(t)
}

// TestAccVcdCseKubernetesClusterCreationWithAutoscaler tests the creation of a cluster with autoscaler enabled
func TestAccVcdCseKubernetesClusterCreationWithAutoscaler(t *testing.T) {
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

	sshPublicKey := "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCrCI+QkLjgQVqR7c7dJfawJqCslVomo5I25JdolqlteX7RCUq0yncWyS+8MTYWCS03sm1jOroLOeuji8CDKCDCcKwQerJiOFoJS+VOK5xCjJ2u8RBGlIpXNcmIh2VriRJrV7TCKrFMSKLNF4/n83q4gWI/YPf6/dRhpPB72HYrdI4omvRlU4GG09jMmgiz+5Yb8wJEXYMsJni+MwPzFKe6TbMcqjBusDyeFGAhgyN7QJGpdNhAn1sqvqZrW2QjaE8P+4t8RzBo8B2ucyQazd6+lbYmOHq9366LjG160snzXrFzlARc4hhpjMzu9Bcm6i3ZZI70qhIbmi5IonbbVh8t"
	clusterName := "test-autoscaler"
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
		"TokenName":          t.Name(),
		"TokenFile":          tokenFilename,
		"ControlPlaneCount":  1,
		"NodePoolCount":      0,
		"ExtraWorkerPool":    " ",
		"PodsCidr":           "100.96.0.0/11",
		"ServicesCidr":       "100.64.0.0/13",
		"SshPublicKey":       sshPublicKey,
		"AutoRepairOnErrors": true,
		"NodeHealthCheck":    true,
		"Timeout":            150,
		"Autoscaler":         "    autoscaler_max_replicas = 5\n    autoscaler_min_replicas = 1",
	}
	testParamsNotEmpty(t, params)

	step1 := templateFill(testAccVcdCseKubernetesCluster, params)
	debugPrintf("#[DEBUG] CONFIGURATION step1: %s", step1)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	vcdClient := createSystemTemporaryVCDConnection()
	cacheId := testCachedFieldValue{}
	clusterResource := "vcd_cse_kubernetes_cluster.my_cluster"
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
					cacheId.cacheTestResourceFieldValue(clusterResource, "id"),
					resource.TestMatchResourceAttr(clusterResource, "id", regexp.MustCompile(`^urn:vcloud:entity:vmware:capvcdCluster:.+$`)),
					resource.TestCheckResourceAttr(clusterResource, "cse_version", testConfig.Cse.Version),
					resource.TestCheckResourceAttr(clusterResource, "runtime", "tkg"),
					resource.TestCheckResourceAttr(clusterResource, "name", clusterName),
					resource.TestCheckResourceAttrPair(clusterResource, "kubernetes_template_id", "data.vcd_catalog_vapp_template.tkg_ova", "id"),
					resource.TestCheckResourceAttrPair(clusterResource, "org", "data.vcd_org_vdc.vdc", "org"),
					resource.TestCheckResourceAttrPair(clusterResource, "vdc_id", "data.vcd_org_vdc.vdc", "id"),
					resource.TestCheckResourceAttrPair(clusterResource, "network_id", "data.vcd_network_routed_v2.routed", "id"),
					resource.TestCheckNoResourceAttr(clusterResource, "owner"), // It is taken from Provider config
					resource.TestCheckResourceAttr(clusterResource, "ssh_public_key", sshPublicKey),
					resource.TestCheckResourceAttr(clusterResource, "control_plane.0.machine_count", "1"),
					resource.TestCheckResourceAttr(clusterResource, "control_plane.0.disk_size_gi", "20"),
					resource.TestCheckResourceAttrPair(clusterResource, "control_plane.0.sizing_policy_id", "data.vcd_vm_sizing_policy.tkg_small", "id"),
					resource.TestCheckResourceAttrPair(clusterResource, "control_plane.0.storage_profile_id", "data.vcd_storage_profile.sp", "id"),
					resource.TestCheckResourceAttrSet(clusterResource, "control_plane.0.ip"), // IP should be assigned after creation as it was not set manually in HCL config
					resource.TestCheckResourceAttr(clusterResource, "worker_pool.#", "1"),
					resource.TestCheckResourceAttr(clusterResource, "worker_pool.0.name", "worker-pool-1"),
					resource.TestCheckResourceAttr(clusterResource, "worker_pool.0.machine_count", "0"),
					resource.TestCheckResourceAttr(clusterResource, "worker_pool.0.autoscaler_max_replicas", "5"),
					resource.TestCheckResourceAttr(clusterResource, "worker_pool.0.autoscaler_min_replicas", "1"),
					resource.TestCheckResourceAttr(clusterResource, "worker_pool.0.disk_size_gi", "20"),
					resource.TestCheckResourceAttrPair(clusterResource, "worker_pool.0.sizing_policy_id", "data.vcd_vm_sizing_policy.tkg_small", "id"),
					resource.TestCheckResourceAttrPair(clusterResource, "worker_pool.0.storage_profile_id", "data.vcd_storage_profile.sp", "id"),
					resource.TestCheckResourceAttrPair(clusterResource, "default_storage_class.0.storage_profile_id", "data.vcd_storage_profile.sp", "id"),
					resource.TestCheckResourceAttr(clusterResource, "default_storage_class.0.name", "sc-1"),
					resource.TestCheckResourceAttr(clusterResource, "default_storage_class.0.reclaim_policy", "delete"),
					resource.TestCheckResourceAttr(clusterResource, "default_storage_class.0.filesystem", "ext4"),
					resource.TestCheckResourceAttr(clusterResource, "pods_cidr", "100.96.0.0/11"),
					resource.TestCheckResourceAttr(clusterResource, "services_cidr", "100.64.0.0/13"),
					resource.TestCheckResourceAttr(clusterResource, "virtual_ip_subnet", ""),
					func() resource.TestCheckFunc {
						// Auto Repair on Errors gets automatically deactivated after cluster creation since CSE 4.1.1
						if cseVersion.GreaterThanOrEqual(v411) {
							return resource.TestCheckResourceAttr(clusterResource, "auto_repair_on_errors", "false")
						} else {
							return resource.TestCheckResourceAttr(clusterResource, "auto_repair_on_errors", "true")
						}
					}(),
					resource.TestCheckResourceAttr(clusterResource, "node_health_check", "true"),
					resource.TestMatchResourceAttr(clusterResource, "kubernetes_version", regexp.MustCompile(`^v[0-9]+\.[0-9]+\.[0-9]+\+vmware\.[0-9]$`)),
					resource.TestMatchResourceAttr(clusterResource, "tkg_product_version", regexp.MustCompile(`^v[0-9]+\.[0-9]+\.[0-9]+$`)),
					resource.TestMatchResourceAttr(clusterResource, "capvcd_version", regexp.MustCompile(`^v[0-9]+\.[0-9]+\.[0-9]+$`)),
					resource.TestMatchResourceAttr(clusterResource, "cluster_resource_set_bindings.#", regexp.MustCompile(`^[1-9][0-9]*$`)),
					resource.TestMatchResourceAttr(clusterResource, "cpi_version", regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+$`)),
					resource.TestMatchResourceAttr(clusterResource, "csi_version", regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+$`)),
					resource.TestCheckResourceAttr(clusterResource, "state", "provisioned"),
					resource.TestCheckResourceAttrSet(clusterResource, "kubeconfig"),
					resource.TestMatchResourceAttr(clusterResource, "events.#", regexp.MustCompile(`^[1-9][0-9]*$`)),
				),
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
		"TokenName":          t.Name(),
		"TokenFile":          tokenFilename,
		"ControlPlaneCount":  1,
		"NodePoolCount":      1,
		"ExtraWorkerPool":    " ",
		"PodsCidr":           "1.2.3.4/24", // This will make the cluster to fail
		"ServicesCidr":       "5.6.7.8/24", // This will make the cluster to fail
		"AutoRepairOnErrors": false,        // This must be false
		"NodeHealthCheck":    false,
		"Autoscaler":         " ",
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
  ssh_public_key         = "{{.SshPublicKey}}"

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
    {{.Autoscaler}}
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
  name = "{{.TenantOrg}}"
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
