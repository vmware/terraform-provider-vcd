//go:build (standaloneVm || vm || ALL || functional) && !skipStandaloneVm
// +build standaloneVm vm ALL functional
// +build !skipStandaloneVm

package vcd

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcdVAppAndVmWithComputePolicies(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	var params = StringMap{
		"Name":                      t.Name(),
		"Org":                       testConfig.VCD.Org,
		"PvdcName":                  testConfig.VCD.NsxtProviderVdc.Name,
		"VmGroupName":               testConfig.VCD.NsxtProviderVdc.PlacementPolicyVmGroup,
		"Catalog":                   testConfig.VCD.Catalog.NsxtBackedCatalogName,
		"Media":                     testConfig.Media.NsxtBackedMediaName,
		"ProviderVdc":               testConfig.VCD.NsxtProviderVdc.Name,
		"NetworkPool":               testConfig.VCD.NsxtProviderVdc.NetworkPool,
		"ProviderVdcStorageProfile": testConfig.VCD.NsxtProviderVdc.StorageProfile,
	}
	testParamsNotEmpty(t, params)

	params["SizingPolicy"] = " "
	params["PlacementPolicy"] = " "
	params["FuncName"] = t.Name() + "1-CreateWithDefault"
	createWithDefaultHcl := templateFill(testAccCheckVcdVappVmAndVmWithComputePolicies, params)

	params["SizingPolicy"] = "sizing_policy_id = vcd_vm_sizing_policy.sizing1.id"
	params["PlacementPolicy"] = "placement_policy_id = \"\""
	params["FuncName"] = t.Name() + "2-RemovePlacementAndAssignSizing"
	removePlacementAndAssignSizingHcl := templateFill(testAccCheckVcdVappVmAndVmWithComputePolicies, params)

	params["SizingPolicy"] = "sizing_policy_id = vcd_vm_sizing_policy.sizing1.id"
	params["PlacementPolicy"] = " "
	params["FuncName"] = t.Name() + "3-CreateWithPolicies"
	createWithSizingPolicy := templateFill(testAccCheckVcdVappVmAndVmWithComputePolicies, params)

	params["SizingPolicy"] = "sizing_policy_id = \"\""
	params["PlacementPolicy"] = "placement_policy_id = vcd_vm_placement_policy.placement1.id"
	params["FuncName"] = t.Name() + "4-ChangeToPlacementPolicy"
	changeToPlacementPolicyHcl := templateFill(testAccCheckVcdVappVmAndVmWithComputePolicies, params)

	params["SizingPolicy"] = "sizing_policy_id = vcd_vm_sizing_policy.sizing1.id"
	params["PlacementPolicy"] = "placement_policy_id = vcd_vm_placement_policy.placement1.id"
	params["FuncName"] = t.Name() + "5-BothPolicies"
	bothPoliciesHcl := templateFill(testAccCheckVcdVappVmAndVmWithComputePolicies, params)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	debugPrintf("#[DEBUG] CONFIGURATION 1: %s\n", createWithDefaultHcl)
	debugPrintf("#[DEBUG] CONFIGURATION 2: %s\n", removePlacementAndAssignSizingHcl)
	debugPrintf("#[DEBUG] CONFIGURATION 3: %s\n", createWithSizingPolicy)
	debugPrintf("#[DEBUG] CONFIGURATION 4: %s\n", changeToPlacementPolicyHcl)
	debugPrintf("#[DEBUG] CONFIGURATION 5: %s\n", bothPoliciesHcl)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVcdStandaloneVmDestroyByVdc(t.Name()),
		Steps: []resource.TestStep{
			{
				Config: createWithDefaultHcl,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVcdStandaloneVmExistsByVdc(t.Name(), t.Name()+"_vm", "vcd_vm."+t.Name()),
					testAccCheckVcdStandaloneVmExistsByVdc(t.Name(), t.Name()+"_vapp_vm", "vcd_vapp_vm."+t.Name()),
					// Standalone VM
					resource.TestCheckResourceAttr("vcd_vm."+t.Name(), "name", t.Name()+"_vm"),
					resource.TestCheckResourceAttr("vcd_vm."+t.Name(), "sizing_policy_id", ""),
					resource.TestCheckResourceAttrPair("vcd_vm."+t.Name(), "placement_policy_id", "vcd_vm_placement_policy.placement1", "id"),
					// vApp VM
					resource.TestCheckResourceAttr("vcd_vapp_vm."+t.Name(), "name", t.Name()+"_vapp_vm"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+t.Name(), "sizing_policy_id", ""),
					resource.TestCheckResourceAttrPair("vcd_vapp_vm."+t.Name(), "placement_policy_id", "vcd_vm_placement_policy.placement1", "id"),
				),
			},
			{
				Config: removePlacementAndAssignSizingHcl,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Standalone VM
					resource.TestCheckResourceAttrPair("vcd_vm."+t.Name(), "sizing_policy_id", "vcd_vm_sizing_policy.sizing1", "id"),
					resource.TestCheckResourceAttr("vcd_vm."+t.Name(), "placement_policy_id", ""),
					// vApp VM
					resource.TestCheckResourceAttrPair("vcd_vapp_vm."+t.Name(), "sizing_policy_id", "vcd_vm_sizing_policy.sizing1", "id"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+t.Name(), "placement_policy_id", ""),
				),
			},
			{
				Config: createWithSizingPolicy,
				Taint:  []string{"vcd_vm." + t.Name(), "vcd_vapp_vm." + t.Name()},
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVcdStandaloneVmExistsByVdc(t.Name(), t.Name()+"_vm", "vcd_vm."+t.Name()),
					testAccCheckVcdStandaloneVmExistsByVdc(t.Name(), t.Name()+"_vapp_vm", "vcd_vapp_vm."+t.Name()),
					// Standalone VM
					resource.TestCheckResourceAttrPair("vcd_vm."+t.Name(), "sizing_policy_id", "vcd_vm_sizing_policy.sizing1", "id"),
					resource.TestCheckResourceAttr("vcd_vm."+t.Name(), "placement_policy_id", ""),
					// vApp VM
					resource.TestCheckResourceAttrPair("vcd_vapp_vm."+t.Name(), "sizing_policy_id", "vcd_vm_sizing_policy.sizing1", "id"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+t.Name(), "placement_policy_id", ""),
				),
			},
			{
				Config: changeToPlacementPolicyHcl,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Standalone VM
					resource.TestCheckResourceAttr("vcd_vm."+t.Name(), "sizing_policy_id", ""),
					resource.TestCheckResourceAttrPair("vcd_vm."+t.Name(), "placement_policy_id", "vcd_vm_placement_policy.placement1", "id"),
					// vApp VM
					resource.TestCheckResourceAttr("vcd_vapp_vm."+t.Name(), "sizing_policy_id", ""),
					resource.TestCheckResourceAttrPair("vcd_vapp_vm."+t.Name(), "placement_policy_id", "vcd_vm_placement_policy.placement1", "id"),
				),
			},
			{
				Config: bothPoliciesHcl,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Standalone VM
					resource.TestCheckResourceAttrPair("vcd_vm."+t.Name(), "sizing_policy_id", "vcd_vm_sizing_policy.sizing1", "id"),
					resource.TestCheckResourceAttrPair("vcd_vm."+t.Name(), "placement_policy_id", "vcd_vm_placement_policy.placement1", "id"),
					// vApp VM
					resource.TestCheckResourceAttrPair("vcd_vapp_vm."+t.Name(), "sizing_policy_id", "vcd_vm_sizing_policy.sizing1", "id"),
					resource.TestCheckResourceAttrPair("vcd_vapp_vm."+t.Name(), "placement_policy_id", "vcd_vm_placement_policy.placement1", "id"),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccCheckVcdVappVmAndVmWithComputePolicies = `
data "vcd_provider_vdc" "pvdc" {
  name = "{{.PvdcName}}"
}

data "vcd_vm_group" "vmgroup" {
  name            = "{{.VmGroupName}}"
  provider_vdc_id = data.vcd_provider_vdc.pvdc.id
}

resource "vcd_vm_placement_policy" "placement1" {
  name            = "placement1"
  provider_vdc_id = data.vcd_provider_vdc.pvdc.id
  vm_group_ids    = [data.vcd_vm_group.vmgroup.id]
}

resource "vcd_vm_placement_policy" "placement2" {
  name            = "placement2"
  provider_vdc_id = data.vcd_provider_vdc.pvdc.id
  vm_group_ids    = [data.vcd_vm_group.vmgroup.id]
}

resource "vcd_vm_sizing_policy" "sizing1" {
  name        = "sizing1"
}

resource "vcd_vm_sizing_policy" "sizing2" {
  name        = "sizing2"
}

resource "vcd_org_vdc" "{{.Name}}" {
  name = "{{.Name}}"
  org  = "{{.Org}}"

  allocation_model  = "AllocationVApp"
  network_pool_name = "{{.NetworkPool}}"
  provider_vdc_name = data.vcd_provider_vdc.pvdc.name

  compute_capacity {
    cpu {
      allocated = "0"
      limit     = "0"
    }

    memory {
      allocated = "0"
      limit     = "0"
    }
  }

  storage_profile {
    name     = "{{.ProviderVdcStorageProfile}}"
    enabled  = true
    limit    = 20240
    default  = true
  }
  enabled                    = true
  enable_thin_provisioning   = true
  enable_fast_provisioning   = true
  delete_force               = true
  delete_recursive           = true

  default_compute_policy_id   = vcd_vm_placement_policy.placement1.id
  vm_sizing_policy_ids        = [vcd_vm_sizing_policy.sizing1.id, vcd_vm_sizing_policy.sizing2.id]
  vm_placement_policy_ids     = [vcd_vm_placement_policy.placement1.id, vcd_vm_placement_policy.placement2.id]
}

resource "vcd_vapp" "{{.Name}}" {
  org         = "{{.Org}}"
  vdc         = vcd_org_vdc.{{.Name}}.name
  name        = "{{.Name}}"
  description = "{{.Name}}"
}

resource "vcd_vapp_vm" "{{.Name}}" {
  vdc              = vcd_vapp.{{.Name}}.vdc
  vapp_name        = vcd_vapp.{{.Name}}.name
  name             = "{{.Name}}_vapp_vm"
  memory           = 512
  cpus             = 1
  cpu_cores        = 1
  os_type          = "sles11_64Guest"
  hardware_version = "vmx-14"
  computer_name    = "foo"
  catalog_name     = "{{.Catalog}}"
  boot_image       = "{{.Media}}"
  power_on         = "true"

  {{.SizingPolicy}}
  {{.PlacementPolicy}}
}

resource "vcd_vm" "{{.Name}}" {
  name              = "{{.Name}}_vm"
  org               = "{{.Org}}"
  vdc               = vcd_org_vdc.{{.Name}}.name
  memory            = 512
  cpus              = 1
  cpu_cores         = 1
  os_type           = "sles11_64Guest"
  hardware_version  = "vmx-14"
  computer_name     = "foo"
  catalog_name      = "{{.Catalog}}"
  boot_image        = "{{.Media}}"
  power_on          = "true"

  {{.SizingPolicy}}
  {{.PlacementPolicy}}
}
`
