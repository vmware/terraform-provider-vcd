//go:build cse || ALL

package vcd

import (
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcdCseKubernetesCluster(t *testing.T) {
	preTestChecks(t)

	if cse := os.Getenv("TEST_VCD_CSE"); cse == "" {
		t.Skip("CSE tests deactivated, skipping " + t.Name())
	}

	var params = StringMap{
		"Name":         strings.ToLower(t.Name()),
		"OvaCatalog":   testConfig.Cse.OvaCatalog,
		"OvaName":      testConfig.Cse.OvaName,
		"SolutionsOrg": testConfig.Cse.SolutionsOrg,
		"TenantOrg":    testConfig.Cse.TenantOrg,
		"Vdc":          testConfig.Cse.Vdc,
		"EdgeGateway":  testConfig.Cse.EdgeGateway,
		"Network":      testConfig.Cse.RoutedNetwork,
		"TokenFile":    getCurrentDir() + t.Name() + ".json",
	}
	testParamsNotEmpty(t, params)

	configText := templateFill(testAccVcdCseKubernetesCluster, params)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check:  resource.ComposeTestCheckFunc(),
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
  name             = "{{.Name}}62"
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
    machine_count      = 1
    disk_size_gi       = 20
    sizing_policy_id   = data.vcd_vm_sizing_policy.tkg_small.id
    storage_profile_id = data.vcd_storage_profile.sp.id
  }

  node_pool {
    name               = "node-pool-1"
    machine_count      = 1
    disk_size_gi       = 20
    sizing_policy_id   = data.vcd_vm_sizing_policy.tkg_small.id
    storage_profile_id = data.vcd_storage_profile.sp.id
  }

  node_pool {
    name               = "node-pool-2"
    machine_count      = 1
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

  pods_cidr     = "100.10.0.0/11"
  services_cidr = "100.90.0.0/11"

  auto_repair_on_errors = false
  node_health_check     = false
}
`
