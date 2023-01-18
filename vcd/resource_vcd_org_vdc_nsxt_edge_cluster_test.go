//go:build vdc || ALL || functional

package vcd

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcdOrgVdcNsxtEdgeCluster(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	var params = StringMap{
		"VdcName":                   t.Name(),
		"OrgName":                   testConfig.VCD.Org,
		"ProviderVdc":               testConfig.VCD.NsxtProviderVdc.Name,
		"NetworkPool":               testConfig.VCD.NsxtProviderVdc.NetworkPool,
		"ProviderVdcStorageProfile": testConfig.VCD.NsxtProviderVdc.StorageProfile,
		"EdgeCluster":               testConfig.Nsxt.NsxtEdgeCluster,
		"Tags":                      "vdc",
	}
	testParamsNotEmpty(t, params)

	configText1 := templateFill(testAccVcdOrgVdcNsxtEdgeCluster, params)
	params["FuncName"] = t.Name() + "-step2DS"
	configText2 := templateFill(testAccVcdOrgVdcNsxtEdgeClusterDataSource, params)

	params["FuncName"] = t.Name() + "-Update"
	configText3 := templateFill(testAccVcdOrgVdcNsxtEdgeCluster_update, params)

	params["FuncName"] = t.Name() + "-UpdateDS"
	configText4 := templateFill(testAccVcdOrgVdcNsxtEdgeClusterDataSource_update, params)

	debugPrintf("#[DEBUG] CONFIGURATION - Step1: %s", configText1)
	debugPrintf("#[DEBUG] CONFIGURATION - Step2: %s", configText2)
	debugPrintf("#[DEBUG] CONFIGURATION - Step3: %s", configText3)
	debugPrintf("#[DEBUG] CONFIGURATION - Step4: %s", configText4)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckVdcDestroy,
		),
		Steps: []resource.TestStep{
			{
				Config: configText1,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdVdcExists("vcd_org_vdc.with-edge-cluster"),
					resource.TestCheckResourceAttr("vcd_org_vdc.with-edge-cluster", "name", t.Name()),
					resource.TestCheckResourceAttr("vcd_org_vdc.with-edge-cluster", "org", testConfig.VCD.Org),
					resource.TestCheckResourceAttr("vcd_org_vdc.with-edge-cluster", "network_pool_name", testConfig.VCD.NsxtProviderVdc.NetworkPool),
					resource.TestCheckResourceAttr("vcd_org_vdc.with-edge-cluster", "provider_vdc_name", testConfig.VCD.NsxtProviderVdc.Name),
					resource.TestCheckResourceAttr("vcd_org_vdc.with-edge-cluster", "enabled", "true"),
					resource.TestCheckResourceAttrSet("vcd_org_vdc.with-edge-cluster", "edge_cluster_id"),
				),
			},
			{
				Config: configText2,
				Check: resource.ComposeTestCheckFunc(
					resourceFieldsEqual("vcd_org_vdc.with-edge-cluster", "data.vcd_org_vdc.ds", []string{"delete_recursive", "delete_force", "%"}),
				),
			},
			{
				Config: configText3,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vcd_org_vdc.with-edge-cluster", "edge_cluster_id", ""),
				),
			},
			{
				Config: configText4,
				Check: resource.ComposeTestCheckFunc(
					resourceFieldsEqual("vcd_org_vdc.with-edge-cluster", "data.vcd_org_vdc.ds", []string{"delete_recursive", "delete_force", "%"}),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdOrgVdcNsxtEdgeClusterDS = `
# skip-binary-test: resource and data source cannot refer itself in a single file
data "vcd_org_vdc" "ds" {
  org  = "{{.OrgName}}"
  name = "{{.VdcName}}"
}
`

const testAccVcdOrgVdcNsxtEdgeCluster = `
data "vcd_provider_vdc" "pvdc" {
  name = "{{.ProviderVdc}}"
}

data "vcd_nsxt_edge_cluster" "ec" {
	provider_vdc_id = data.vcd_provider_vdc.pvdc.id
	name            = "{{.EdgeCluster}}"
}

resource "vcd_org_vdc" "with-edge-cluster" {
  name = "{{.VdcName}}"
  org  = "{{.OrgName}}"

  allocation_model  = "ReservationPool"
  network_pool_name = "{{.NetworkPool}}"
  provider_vdc_name = data.vcd_provider_vdc.pvdc.name

  edge_cluster_id = data.vcd_nsxt_edge_cluster.ec.id

  compute_capacity {
    cpu {
      allocated = 1024
      limit     = 1024
    }

    memory {
      allocated = 1024
      limit     = 1024
    }
  }

  storage_profile {
    name     = "{{.ProviderVdcStorageProfile}}"
    enabled  = true
    limit    = 10240
    default  = true
  }

  enabled                    = true
  enable_thin_provisioning   = true
  enable_fast_provisioning   = true
  delete_force               = true
  delete_recursive           = true
}
`

const testAccVcdOrgVdcNsxtEdgeClusterDataSource = testAccVcdOrgVdcNsxtEdgeCluster + testAccVcdOrgVdcNsxtEdgeClusterDS

const testAccVcdOrgVdcNsxtEdgeCluster_update = `
data "vcd_provider_vdc" "pvdc" {
  name = "{{.ProviderVdc}}"
}

data "vcd_nsxt_edge_cluster" "ec" {
	provider_vdc_id = data.vcd_provider_vdc.pvdc.id
	name            = "{{.EdgeCluster}}"
}

resource "vcd_org_vdc" "with-edge-cluster" {
  name = "{{.VdcName}}"
  org  = "{{.OrgName}}"

  allocation_model  = "ReservationPool"
  network_pool_name = "{{.NetworkPool}}"
  provider_vdc_name = data.vcd_provider_vdc.pvdc.name

  compute_capacity {
    cpu {
      allocated = 1024
      limit     = 1024
    }

    memory {
      allocated = 1024
      limit     = 1024
    }
  }

  storage_profile {
    name     = "{{.ProviderVdcStorageProfile}}"
    enabled  = true
    limit    = 10240
    default  = true
  }

  enabled                    = true
  enable_thin_provisioning   = true
  enable_fast_provisioning   = true
  delete_force               = true
  delete_recursive           = true
}
`

const testAccVcdOrgVdcNsxtEdgeClusterDataSource_update = testAccVcdOrgVdcNsxtEdgeCluster_update + testAccVcdOrgVdcNsxtEdgeClusterDS
