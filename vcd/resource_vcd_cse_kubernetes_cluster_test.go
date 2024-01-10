//go:build cse

package vcd

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcdCseKubernetesCluster(t *testing.T) {
	preTestChecks(t)

	var params = StringMap{
		"Name":          t.Name(),
		"OvaCatalog":    testConfig.Cse.OvaCatalog,
		"OvaName":       testConfig.Cse.OvaName,
		"Org":           testConfig.Cse.Org,
		"Vdc":           testConfig.Cse.Vdc,
		"EdgeGateway":   testConfig.Cse.EdgeGateway,
		"Network":       testConfig.Cse.RoutedNetwork,
		"CapVcdVersion": testConfig.Cse.CapVcdVersion,
		"Owner":         testConfig.Cse.Owner,
		"ApiToken":      testConfig.Cse.ApiTokenFile,
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
  org  = "{{.Org}}"
  name = "{{.OvaCatalog}}"
}

data "vcd_catalog_vapp_template" "tkg_ova" {
  org        = data.vcd_catalog.tkg_catalog.org
  catalog_id = data.vcd_catalog.tkg_catalog.id
  name       = "{{.OvaName}}"
}

data "vcd_rde_type" "capvcdcluster_type" {
  vendor  = "vmware"
  nss     = "capvcdCluster"
  version = "{{.CapVcdVersion}}"
}

data "vcd_org_vdc" "vdc" {
  org  = data.vcd_catalog.tkg_catalog.org
  name = "{{.Vdc}}"
}

data "vcd_nsxt_edgegateway" "egw" {
  org  = data.vcd_org_vdc.vdc.org
  name = "{{.EdgeGateway}}"
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

resource "vcd_cse_kubernetes_cluster" "my_cluster" {
  name               = "{{.Name}}"
  ova_id             = data.vcd_catalog_vapp_template.tkg_ova.id
  capvcd_rde_type_id = data.vcd_rde_type.capvcdcluster_type.id
  org                = "{{.Org}}"
  vdc_id             = data.vcd_org_vdc.vdc.id
  network_id         = data.vcd_network_routed_v2.routed.id
  owner              = "{{.Owner}}"
  api_token_file	 = "{{.ApiTokenFile}}"

  control_plane {
    machine_count    = 1
    disk_size        = 20
    sizing_policy_id = data.vcd_vm_sizing_policy.tkg_small.id
  }

  node_pool {
    machine_count      = 1
    disk_size          = 20
    sizing_policy_id   = data.vcd_vm_sizing_policy.tkg_small.id
    storage_profile_id = data.vcd_storage_profile.sp.id
  }

  node_pool {
    name               = "node-pool-1"
    machine_count      = 1
    disk_size          = 20
    sizing_policy_id   = data.vcd_vm_sizing_policy.tkg_small.id
    storage_profile_id = data.vcd_storage_profile.sp.id
  }

  node_pool {
    name               = "node-pool-2"
    machine_count      = 1
    disk_size          = 20
    sizing_policy_id   = data.vcd_vm_sizing_policy.tkg_small.id
    storage_profile_id = data.vcd_storage_profile.sp.id
  }

  storage_class {
	storage_profile_id = data.vcd_storage_profile.sp.id
    name               = "sc-1"
    reclaim_policy     = "delete"
    filesystem         = "ext4"
  }

  pods_cidr     = "100.10.0.0/11"
  services_cidr = "100.90.0.0/11"

  auto_repair_on_errors = true
  node_health_check     = true
}
`
