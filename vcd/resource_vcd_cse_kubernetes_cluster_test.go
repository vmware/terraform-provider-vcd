//go:build cse || ALL

package vcd

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcdCseKubernetesCluster(t *testing.T) {
	preTestChecks(t)

	if cse := os.Getenv("TEST_VCD_CSE"); cse == "" {
		t.Skip("CSE tests deactivated, skipping " + t.Name())
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

	now := time.Now()
	var params = StringMap{
		"Name":              strings.ToLower(t.Name()),
		"OvaCatalog":        testConfig.Cse.OvaCatalog,
		"OvaName":           testConfig.Cse.OvaName,
		"SolutionsOrg":      testConfig.Cse.SolutionsOrg,
		"TenantOrg":         testConfig.Cse.TenantOrg,
		"Vdc":               testConfig.Cse.Vdc,
		"EdgeGateway":       testConfig.Cse.EdgeGateway,
		"Network":           testConfig.Cse.RoutedNetwork,
		"TokenName":         fmt.Sprintf("%s%d%d%d", strings.ToLower(t.Name()), now.Day(), now.Hour(), now.Minute()),
		"TokenFile":         tokenFilename,
		"ControlPlaneCount": 1,
		"NodePoolCount":     1,
	}
	testParamsNotEmpty(t, params)

	step1 := templateFill(testAccVcdCseKubernetesCluster, params)

	params["FuncName"] = t.Name() + "Step2"
	params["ControlPlaneCount"] = 2
	step2 := templateFill(testAccVcdCseKubernetesCluster, params)

	params["FuncName"] = t.Name() + "Step3"
	params["ControlPlaneCount"] = 1
	params["NodePoolCount"] = 2
	step3 := templateFill(testAccVcdCseKubernetesCluster, params)

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
			{
				Config: step1,
				Check: resource.ComposeAggregateTestCheckFunc(
					cacheId.cacheTestResourceFieldValue(clusterName, "id"),
					resource.TestCheckResourceAttrSet(clusterName, "id"),
					resource.TestCheckResourceAttr(clusterName, "name", strings.ToLower(t.Name())),
					resource.TestCheckResourceAttr(clusterName, "state", "provisioned"),
					resource.TestCheckResourceAttrSet(clusterName, "kubeconfig"),
				),
			},
			{
				Config: step2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(clusterName, "id", cacheId.fieldValue),
					resource.TestCheckResourceAttr(clusterName, "state", "provisioned"),
					resource.TestCheckResourceAttrSet(clusterName, "kubeconfig"),
				),
			},
			{
				Config: step3,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(clusterName, "id", cacheId.fieldValue),
					resource.TestCheckResourceAttr(clusterName, "state", "provisioned"),
					resource.TestCheckResourceAttrSet(clusterName, "kubeconfig"),
				),
			},
		},
	})
	postTestChecks(t)
}

// TODO: Test:
// Basic (DONE)
// With machine health checks
// With machine health checks
// Without storage class
// With virtual IP and control plane IPs
// Nodes With vGPU policies
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
  cse_version        = "4.2"
  runtime            = "tkg"
  name               = "{{.Name}}"
  ova_id             = data.vcd_catalog_vapp_template.tkg_ova.id
  org                = data.vcd_org_vdc.vdc.org
  vdc_id             = data.vcd_org_vdc.vdc.id
  network_id         = data.vcd_network_routed_v2.routed.id
  api_token_file	 = vcd_api_token.token.file_name

  control_plane {
    machine_count      = {{.ControlPlaneCount}}
    disk_size_gi       = 20
    sizing_policy_id   = data.vcd_vm_sizing_policy.tkg_small.id
    storage_profile_id = data.vcd_storage_profile.sp.id
  }

  node_pool {
    name               = "node-pool-1"
    machine_count      = {{.NodePoolCount}}
    disk_size_gi       = 20
    sizing_policy_id   = data.vcd_vm_sizing_policy.tkg_small.id
    storage_profile_id = data.vcd_storage_profile.sp.id
  }

  default_storage_class {
	name               = "sc-1"
	storage_profile_id = data.vcd_storage_profile.sp.id
    reclaim_policy     = "delete"
    filesystem         = "ext4"
  }

  auto_repair_on_errors = true
  node_health_check     = true

  operations_timeout_minutes = 0
}
`
