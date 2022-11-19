//go:build vdc || ALL || functional
// +build vdc ALL functional

package vcd

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcdOrgVdcWithAllComputePolicies(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	var params = StringMap{
		"VmGroup":                   testConfig.VCD.NsxtProviderVdc.PlacementPolicyVmGroup,
		"VdcName":                   t.Name(),
		"OrgName":                   testConfig.VCD.Org,
		"ProviderVdc":               testConfig.VCD.NsxtProviderVdc.Name,
		"NetworkPool":               testConfig.VCD.NsxtProviderVdc.NetworkPool,
		"ProviderVdcStorageProfile": testConfig.VCD.NsxtProviderVdc.StorageProfile,
		"FuncName":                  t.Name(),
	}
	testParamsNotEmpty(t, params)

	resourceName := "vcd_org_vdc." + t.Name()
	configText := templateFill(testAccCheckVcdVdcAllComputePolicies, params)
	params["FuncName"] = t.Name() + "-Update"
	configText2 := templateFill(testAccCheckVcdVdcAllComputePolicies_update, params)
	debugPrintf("#[DEBUG] CONFIGURATION - creation: %s", configText)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckVdcDestroy,
			testAccCheckComputePolicyDestroyed("placement1", "placement"),
			testAccCheckComputePolicyDestroyed("sizing1", "sizing"),
			testAccCheckComputePolicyDestroyed("sizing2", "sizing"),
		),
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdVdcExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", t.Name()),
					resource.TestCheckResourceAttr(resourceName, "org", testConfig.VCD.Org),
					resource.TestCheckResourceAttr(resourceName, "network_pool_name", testConfig.VCD.NsxtProviderVdc.NetworkPool),
					resource.TestCheckResourceAttr(resourceName, "provider_vdc_name", testConfig.VCD.NsxtProviderVdc.Name),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "storage_profile.0.name", testConfig.VCD.NsxtProviderVdc.StorageProfile),
					resource.TestCheckResourceAttrPair(resourceName, "default_compute_policy_id", "vcd_vm_placement_policy.placement1", "id"),
					resource.TestCheckResourceAttr(resourceName, "vm_placement_policy_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vm_sizing_policy_ids.#", "1"),
				),
			},
			{
				Config: configText2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "default_compute_policy_id", "vcd_vm_placement_policy.placement1", "id"),
					resource.TestCheckResourceAttr(resourceName, "vm_placement_policy_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vm_sizing_policy_ids.#", "2"),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccCheckVcdVdcAllComputePolicies = `
data "vcd_provider_vdc" "pvdc" {
  name = "{{.ProviderVdc}}"
}

data "vcd_vm_group" "vm-group" {
  name            = "{{.VmGroup}}"
  provider_vdc_id = data.vcd_provider_vdc.pvdc.id
}

resource "vcd_vm_placement_policy" "placement1" {
  name        = "placement1"
  provider_vdc_id = data.vcd_provider_vdc.pvdc.id
  description = "placement1 description"
  vm_group_ids = [ data.vcd_vm_group.vm-group.id ]
}

resource "vcd_vm_sizing_policy" "sizing1" {
  name        = "sizing1"
  description = "sizing1 description"
  cpu {
    shares                = "886"
    limit_in_mhz          = "12375"
    count                 = "9"
    speed_in_mhz          = "2500"
    cores_per_socket      = "3"
    reservation_guarantee = "0.55"
  }
}

resource "vcd_vm_sizing_policy" "sizing2" {
  name        = "sizing2"
  description = "sizing2 description"
  cpu {
    shares                = "886"
    limit_in_mhz          = "12375"
    count                 = "9"
    speed_in_mhz          = "2500"
    cores_per_socket      = "3"
    reservation_guarantee = "0.55"
  }
}

resource "vcd_org_vdc" "{{.VdcName}}" {
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

  default_compute_policy_id = vcd_vm_placement_policy.placement1.id
  vm_placement_policy_ids   = [vcd_vm_placement_policy.placement1.id]
  vm_sizing_policy_ids = [vcd_vm_sizing_policy.sizing1.id]
}
`

// Here we change assignments of the policies to test the Update operation
const testAccCheckVcdVdcAllComputePolicies_update = `
data "vcd_provider_vdc" "pvdc" {
  name = "{{.ProviderVdc}}"
}

data "vcd_vm_group" "vm-group" {
  name            = "{{.VmGroup}}"
  provider_vdc_id = data.vcd_provider_vdc.pvdc.id
}

resource "vcd_vm_placement_policy" "placement1" {
  name        = "placement1"
  provider_vdc_id = data.vcd_provider_vdc.pvdc.id
  description = "placement1 description"
  vm_group_ids = [ data.vcd_vm_group.vm-group.id ]
}

resource "vcd_vm_sizing_policy" "sizing1" {
  name        = "sizing1"
  description = "sizing1 description"
  cpu {
    shares                = "886"
    limit_in_mhz          = "12375"
    count                 = "9"
    speed_in_mhz          = "2500"
    cores_per_socket      = "3"
    reservation_guarantee = "0.55"
  }
}

resource "vcd_vm_sizing_policy" "sizing2" {
  name        = "sizing2"
  description = "sizing2 description"
  cpu {
    shares                = "886"
    limit_in_mhz          = "12375"
    count                 = "9"
    speed_in_mhz          = "2500"
    cores_per_socket      = "3"
    reservation_guarantee = "0.55"
  }
}

resource "vcd_org_vdc" "{{.VdcName}}" {
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

  default_compute_policy_id = vcd_vm_placement_policy.placement1.id
  vm_placement_policy_ids   = [vcd_vm_placement_policy.placement1.id]
  vm_sizing_policy_ids = [vcd_vm_sizing_policy.sizing1.id, vcd_vm_sizing_policy.sizing2.id]
}
`
