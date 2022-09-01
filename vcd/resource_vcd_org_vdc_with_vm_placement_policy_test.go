//go:build vdc || ALL || functional
// +build vdc ALL functional

package vcd

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcdOrgVdcWithVmPlacementPolicy(t *testing.T) {
	preTestChecks(t)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	if !usingSysAdmin() {
		t.Skip(t.Name() + " requires system admin privileges")
	}
	if testConfig.VCD.ProviderVdc.Name == "" {
		t.Skip("Variable providerVdc.Name must be set to run VDC tests")
	}

	vmGroupUrn := getVmGroupUrn()
	if vmGroupUrn == "" {
		t.Skip(t.Name() + " could not find VM Group in testEnvBuild.placementPolicyVmGroup required to test VM Placement Policies")
	}

	var params = StringMap{
		"VmGroup":                   vmGroupUrn,
		"VdcName":                   t.Name(),
		"OrgName":                   testConfig.VCD.Org,
		"ProviderVdc":               testConfig.VCD.NsxtProviderVdc.Name,
		"NetworkPool":               testConfig.VCD.NsxtProviderVdc.NetworkPool,
		"ProviderVdcStorageProfile": testConfig.VCD.NsxtProviderVdc.StorageProfile,
		"FuncName":                  t.Name(),
	}
	testParamsNotEmpty(t, params)

	resourceName := "vcd_org_vdc."+t.Name()
	configText := templateFill(testAccCheckVcdVdcVmPlacementPolicies_basic, params)
	params["FuncName"] = t.Name() + "-Update"
	updateText := templateFill(testAccCheckVcdVdcVmPlacementPolicies_update, params)

	debugPrintf("#[DEBUG] CONFIGURATION - creation: %s", configText)
	debugPrintf("#[DEBUG] CONFIGURATION - update: %s", updateText)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVdcDestroy,
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
					resource.TestCheckResourceAttrPair(resourceName, "default_compute_policy_id","vcd_vm_placement_policy.placement1", "id"),
					resource.TestCheckResourceAttr(resourceName, "vm_placement_policy_ids.#","3"),
				),
			},
			{
				Config: updateText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdVdcExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", t.Name()),
					resource.TestCheckResourceAttr(resourceName, "org", testConfig.VCD.Org),
					resource.TestCheckResourceAttr(resourceName, "network_pool_name", testConfig.VCD.NsxtProviderVdc.NetworkPool),
					resource.TestCheckResourceAttr(resourceName, "provider_vdc_name", testConfig.VCD.NsxtProviderVdc.Name),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "storage_profile.0.name", testConfig.VCD.NsxtProviderVdc.StorageProfile),
					resource.TestCheckResourceAttrPair(resourceName, "default_compute_policy_id","vcd_vm_placement_policy.placement3", "id"),
					resource.TestCheckResourceAttr(resourceName, "vm_placement_policy_ids.#","2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdOrgObject(t.Name()),
				// These fields can't be retrieved
				ImportStateVerifyIgnore: []string{"delete_force", "delete_recursive"},
			},
		},
	})
	postTestChecks(t)
}

const testAccCheckVcdVdcVmPlacementPolicies_basic = `
data "vcd_provider_vdc" "pvdc" {
  name = "{{.ProviderVdc}}"
}

resource "vcd_vm_placement_policy" "placement1" {
  name        = "placement1"
  provider_vdc_id = data.vcd_provider_vdc.pvdc.id
  description = "placement1 description"
  vm_group_ids = ["{{.VmGroup}}"]
}

resource "vcd_vm_placement_policy" "placement2" {
  name        = "placement2"
  provider_vdc_id = data.vcd_provider_vdc.pvdc.id
  description = "placement2 description"
  vm_group_ids = ["{{.VmGroup}}"]
}

resource "vcd_vm_placement_policy" "placement3" {
  name        = "placement3"
  provider_vdc_id = data.vcd_provider_vdc.pvdc.id
  description = "placement3 description"
  vm_group_ids = ["{{.VmGroup}}"]
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
  vm_placement_policy_ids   = [vcd_vm_placement_policy.placement1.id, vcd_vm_placement_policy.placement2.id, vcd_vm_placement_policy.placement3.id]
}
`

const testAccCheckVcdVdcVmPlacementPolicies_update = `
# skip-binary-test: only for updates
data "vcd_provider_vdc" "pvdc" {
  name = "{{.ProviderVdc}}"
}

resource "vcd_vm_placement_policy" "placement1" {
  name        = "placement1"
  provider_vdc_id = data.vcd_provider_vdc.pvdc.id
  description = "placement1 description"
  vm_group_ids = ["{{.VmGroup}}"]
}

resource "vcd_vm_placement_policy" "placement2" {
  name        = "placement2"
  provider_vdc_id = data.vcd_provider_vdc.pvdc.id
  description = "placement2 description"
  vm_group_ids = ["{{.VmGroup}}"]
}

resource "vcd_vm_placement_policy" "placement3" {
  name        = "placement3"
  provider_vdc_id = data.vcd_provider_vdc.pvdc.id
  description = "placement3 description"
  vm_group_ids = ["{{.VmGroup}}"]
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

  default_compute_policy_id = vcd_vm_placement_policy.placement3.id
  vm_placement_policy_ids   = [vcd_vm_placement_policy.placement1.id, vcd_vm_placement_policy.placement3.id]
}
`