//go:build (standaloneVm || vm || ALL || functional) && !skipStandaloneVm
// +build standaloneVm vm ALL functional
// +build !skipStandaloneVm

package vcd

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcdStandaloneVmWithVmPlacement(t *testing.T) {
	preTestChecks(t)

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

	configTextVM := templateFill(testAccCheckVcdEmptyWithPlacement, params)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	debugPrintf("#[DEBUG] CONFIGURATION: %s\n", configTextVM)
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVcdStandaloneVmDestroyByVdc(t.Name()),
		Steps: []resource.TestStep{
			{
				Config: configTextVM,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVcdStandaloneVmExistsByVdc(t.Name(), t.Name()+"_vm", "vcd_vm."+t.Name()),
					testAccCheckVcdStandaloneVmExistsByVdc(t.Name(), t.Name()+"_vapp_vm", "vcd_vapp_vm."+t.Name()),
					// Standalone VM
					resource.TestCheckResourceAttr("vcd_vm."+t.Name(), "name", t.Name()+"_vm"),
					resource.TestCheckResourceAttrSet("vcd_vm."+t.Name(), "placement_policy_id"),
					resource.TestCheckResourceAttrPair("vcd_vm."+t.Name(), "placement_policy_id", "vcd_vm_placement_policy.placement", "id"),
					// vApp VM
					resource.TestCheckResourceAttr("vcd_vapp_vm."+t.Name(), "name", t.Name()+"_vapp_vm"),
					resource.TestCheckResourceAttrSet("vcd_vapp_vm."+t.Name(), "placement_policy_id"),
					resource.TestCheckResourceAttrPair("vcd_vapp_vm."+t.Name(), "placement_policy_id", "vcd_vm_placement_policy.placement", "id"),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccCheckVcdEmptyWithPlacement = `
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

resource "vcd_vm_sizing_policy" "sizing" {
  name        = "sizing"
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
  vm_sizing_policy_ids        = [vcd_vm_sizing_policy.sizing.id]
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

  placement_policy_id = vcd_vm_placement_policy.placement1.id
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

  placement_policy_id = vcd_vm_placement_policy.placement1.id
}
`
