// +build vapp vm ALL functional

package vcd

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/vmware/go-vcloud-director/v2/govcd"
)

func TestAccVcdVAppVmWithVmSizing(t *testing.T) {
	var (
		vm            govcd.VM
		netVappName   string = t.Name()
		netVmName1    string = t.Name() + "VM"
		netVmName2    string = t.Name() + "VM2"
		netVmName3    string = t.Name() + "VM3"
		netVmName4    string = t.Name() + "VM4"
		netVmName5    string = t.Name() + "VM5"
		testAccVcdVdc string = "test_VmSizing"
	)

	var params = StringMap{
		"Org":         testConfig.VCD.Org,
		"EdgeGateway": testConfig.Networking.EdgeGateway,
		"Catalog":     testSuiteCatalogName,
		"CatalogItem": testSuiteCatalogOVAItem,
		"VAppName":    netVappName,
		"VMName":      netVmName1,
		"Tags":        "vapp vm",

		"VdcName":                   testAccVcdVdc,
		"AllocationModel":           "Flex",
		"ProviderVdc":               testConfig.VCD.ProviderVdc.Name,
		"NetworkPool":               testConfig.VCD.ProviderVdc.NetworkPool,
		"Allocated":                 "10240",
		"Reserved":                  "0",
		"Limit":                     "10240",
		"ProviderVdcStorageProfile": testConfig.VCD.ProviderVdc.StorageProfile,
		"FuncName":                  t.Name(),
		"MemoryGuaranteed":          "0.5",
		"CpuGuaranteed":             "0.6",
		// The parameters below are for Flex allocation model
		// Part of HCL is created dynamically and these parameters with values result in the Flex part of the template being filled:
		"equalsChar":                   "=",
		"FlexElasticKey":               "elasticity",
		"FlexElasticValue":             "false",
		"ElasticityValueForAssert":     "false",
		"FlexMemoryOverheadKey":        "include_vm_memory_overhead",
		"FlexMemoryOverheadValue":      "false",
		"MemoryOverheadValueForAssert": "false",
	}

	vcdClient, err := getTestVCDFromJson(testConfig)
	if err != nil {
		t.Skip("unable to validate vCD version - skipping test")
	}

	if !vcdClient.Client.IsSysAdmin {
		t.Skip("Test can only run as System admin")
	}

	if vcdClient.Client.APIVCDMaxVersionIs("< 33.0") {
		t.Skip("TestAccVcdOrgVdcWithVmSizingPolicy requires VCD 10.0+")
	}

	configTextVM := templateFill(testAccCheckVcdVAppEmptyVmWithSizing, params)

	params["FuncName"] = t.Name() + "-step1"
	configTextVMUpdateStep1 := templateFill(testAccCheckVcdVAppEmptyVmWithSizingUpdate, params)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	debugPrintf("#[DEBUG] CONFIGURATION: %s\n", configTextVM)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVcdVAppVmDestroyByVdc(testAccVcdVdc, netVappName),
		Steps: []resource.TestStep{
			// Step 0 - Create using sizing policy
			resource.TestStep{
				Config: configTextVM,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVcdVAppVmExistsByVdc(testAccVcdVdc, netVappName, netVmName1, "vcd_vapp_vm."+netVmName1, &vm),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "name", netVmName1),

					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "os_type", "sles11_64Guest"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "hardware_version", "vmx-13"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "expose_hardware_virtualization", "true"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "computer_name", "compName"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "description", "test empty VM"),

					resource.TestCheckResourceAttrPair("vcd_vapp_vm."+netVmName1, "sizing_policy_id",
						"vcd_vm_sizing_policy.size_cpu", "id"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "cpus", "3"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "cpu_cores", "3"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "memory", "1024"),

					testAccCheckVcdVAppVmExistsByVdc(testAccVcdVdc, netVappName, netVmName2, "vcd_vapp_vm."+netVmName2, &vm),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName2, "name", netVmName2),

					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName2, "os_type", "sles11_64Guest"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName2, "hardware_version", "vmx-13"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName2, "expose_hardware_virtualization", "true"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName2, "computer_name", "compName"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName2, "description", "test empty VM2"),

					resource.TestCheckResourceAttrPair("vcd_vapp_vm."+netVmName2, "sizing_policy_id",
						"vcd_vm_sizing_policy.size_full", "id"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName2, "cpus", "3"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName2, "cpu_cores", "3"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName2, "memory", "2048"),

					testAccCheckVcdVAppVmExistsByVdc(testAccVcdVdc, netVappName, netVmName3, "vcd_vapp_vm."+netVmName3, &vm),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName3, "name", netVmName3),

					resource.TestCheckResourceAttrPair("vcd_vapp_vm."+netVmName3, "sizing_policy_id",
						"vcd_vm_sizing_policy.size_cpu", "id"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName3, "cpus", "3"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName3, "cpu_cores", "3"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName3, "memory", "2048"),

					testAccCheckVcdVAppVmExistsByVdc(testAccVcdVdc, netVappName, netVmName4, "vcd_vapp_vm."+netVmName4, &vm),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName4, "name", netVmName4),

					resource.TestCheckResourceAttrPair("vcd_vapp_vm."+netVmName4, "sizing_policy_id",
						"vcd_vm_sizing_policy.size_full", "id"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName4, "cpus", "3"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName4, "cpu_cores", "3"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName4, "memory", "2048"),
				),
			},
			// Step 1 - update
			resource.TestStep{
				Config: configTextVMUpdateStep1,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVcdVAppVmExistsByVdc(testAccVcdVdc, netVappName, netVmName1, "vcd_vapp_vm."+netVmName1, &vm),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "name", netVmName1),

					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "os_type", "sles11_64Guest"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "hardware_version", "vmx-13"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "expose_hardware_virtualization", "false"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "computer_name", "compNameUp"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "description", "test empty VM updated"),

					resource.TestCheckResourceAttrPair("vcd_vapp_vm."+netVmName1, "sizing_policy_id",
						"vcd_vm_sizing_policy.size_full", "id"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "cpus", "3"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "cpu_cores", "3"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName1, "memory", "2048"),

					testAccCheckVcdVAppVmExistsByVdc(testAccVcdVdc, netVappName, netVmName2, "vcd_vapp_vm."+netVmName2, &vm),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName2, "name", netVmName2),

					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName2, "os_type", "sles11_64Guest"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName2, "hardware_version", "vmx-13"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName2, "expose_hardware_virtualization", "true"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName2, "computer_name", "compName"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName2, "description", "test empty VM2"),

					resource.TestCheckResourceAttrPair("vcd_vapp_vm."+netVmName2, "sizing_policy_id",
						"vcd_vm_sizing_policy.size_cpu", "id"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName2, "cpus", "3"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName2, "cpu_cores", "3"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName2, "memory", "2048"),

					testAccCheckVcdVAppVmExistsByVdc(testAccVcdVdc, netVappName, netVmName3, "vcd_vapp_vm."+netVmName3, &vm),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName3, "name", netVmName3),

					resource.TestCheckResourceAttrPair("vcd_vapp_vm."+netVmName3, "sizing_policy_id",
						"vcd_vm_sizing_policy.size_cpu", "id"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName3, "cpus", "3"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName3, "cpu_cores", "3"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName3, "memory", "3072"),

					testAccCheckVcdVAppVmExistsByVdc(testAccVcdVdc, netVappName, netVmName5, "vcd_vapp_vm."+netVmName5, &vm),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName5, "name", netVmName5),

					resource.TestCheckResourceAttrPair("vcd_vapp_vm."+netVmName5, "sizing_policy_id",
						"vcd_vm_sizing_policy.minSize", "id"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName5, "cpus", "4"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName5, "cpu_cores", "2"),
					resource.TestCheckResourceAttr("vcd_vapp_vm."+netVmName5, "memory", "1536"),
				),
			},
		},
	})
}

func testAccCheckVcdVAppVmExistsByVdc(vdcName, vappName, vmName, node string, vm *govcd.VM) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[node]
		if !ok {
			return fmt.Errorf("not found: %s", node)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no vApp ID is set")
		}

		conn := testAccProvider.Meta().(*VCDClient)
		_, vdc, err := conn.GetOrgAndVdc(testConfig.VCD.Org, vdcName)
		if err != nil {
			return fmt.Errorf(errorRetrievingVdcFromOrg, vdcName, testConfig.VCD.Org, err)
		}

		vapp, err := vdc.GetVAppByName(vappName, false)
		if err != nil {
			return err
		}

		newVm, err := vapp.GetVMByName(vmName, false)

		if err != nil {
			return err
		}

		*vm = *newVm

		return nil
	}
}

func testAccCheckVcdVAppVmDestroyByVdc(vdcName, vappName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*VCDClient)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "vcd_vapp" {
				continue
			}
			org, err := conn.GetOrgByName(testConfig.VCD.Org)
			if err != nil {
				return fmt.Errorf(errorRetrievingOrg, testConfig.VCD.Org)
			}

			_, err = org.GetVDCByName(vdcName, false)
			if err == nil {
				return fmt.Errorf("VDC still exist")
			}

			return nil
		}

		return nil
	}
}

const testAccCheckVcdVAppEmptyWithSizing = `
resource "vcd_vm_sizing_policy" "minSize" {
  name        = "min-size"
  description = "smallest size"
}

resource "vcd_vm_sizing_policy" "size_cpu" {
  name        = "min-size2"
  description = "smallest size2"

  cpu {
    shares                = "886"
    limit_in_mhz          = "2400"
    count                 = "3"
    speed_in_mhz          = "1500"
    cores_per_socket      = "3"
    reservation_guarantee = "0.45"
  }

}

resource "vcd_vm_sizing_policy" "size_full" {
  name        = "min-size3"
  description = "smallest size2"

  cpu {
    shares                = "886"
    limit_in_mhz          = "2400"
    count                 = "3"
    speed_in_mhz          = "1500"
    cores_per_socket      = "3"
    reservation_guarantee = "0.45"
  }

  memory {
    shares                = "1580"
    size_in_mb            = "2048"
    limit_in_mb           = "4800"
    reservation_guarantee = "0.5"
  }
}

resource "vcd_org_vdc" "{{.VdcName}}" {
  name = "{{.VdcName}}"
  org  = "{{.Org}}"

  allocation_model  = "{{.AllocationModel}}"
  network_pool_name = "{{.NetworkPool}}"
  provider_vdc_name = "{{.ProviderVdc}}"

  compute_capacity {
    cpu {
      allocated = "{{.Allocated}}"
      limit     = "{{.Limit}}"
    }

    memory {
      allocated = "{{.Allocated}}"
      limit     = "{{.Limit}}"
    }
  }

  storage_profile {
    name     = "{{.ProviderVdcStorageProfile}}"
    enabled  = true
    limit    = 90240
    default  = true
  }

  metadata = {
    vdc_metadata = "VDC Metadata"
  }

  enabled                    = true
  enable_thin_provisioning   = true
  enable_fast_provisioning   = true
  delete_force               = true
  delete_recursive           = true
  {{.FlexElasticKey}}                 {{.equalsChar}} {{.FlexElasticValue}}
  {{.FlexMemoryOverheadKey}} {{.equalsChar}} {{.FlexMemoryOverheadValue}}

  default_vm_sizing_policy_id = vcd_vm_sizing_policy.size_full.id
  vm_sizing_policy_ids        = [vcd_vm_sizing_policy.minSize.id, vcd_vm_sizing_policy.size_cpu.id,vcd_vm_sizing_policy.size_full.id]
}

resource "vcd_vapp" "{{.VAppName}}" {
  org = "{{.Org}}"
  vdc = vcd_org_vdc.{{.VdcName}}.name

  name       = "{{.VAppName}}"
}
`
const testAccCheckVcdVAppEmptyVmWithSizing = testAccCheckVcdVAppEmptyWithSizing + `
resource "vcd_vapp_vm" "{{.VMName}}" {
  org = "{{.Org}}"
  vdc = vcd_org_vdc.{{.VdcName}}.name

  power_on = true

  vapp_name     = vcd_vapp.{{.VAppName}}.name
  description   = "test empty VM"
  name          = "{{.VMName}}"
  
  os_type                        = "sles11_64Guest"
  hardware_version               = "vmx-13"
  catalog_name                   = "{{.Catalog}}"
  expose_hardware_virtualization = true
  computer_name                  = "compName"

  memory_hot_add_enabled = true

  sizing_policy_id = vcd_vm_sizing_policy.size_cpu.id
  memory           = 1024
 }

resource "vcd_vapp_vm" "{{.VMName}}2" {
  org = "{{.Org}}"
  vdc = vcd_org_vdc.{{.VdcName}}.name

  power_on = true

  vapp_name     = vcd_vapp.{{.VAppName}}.name
  description   = "test empty VM2"
  name          = "{{.VMName}}2"
  
  os_type                        = "sles11_64Guest"
  hardware_version               = "vmx-13"
  catalog_name                   = "{{.Catalog}}"
  expose_hardware_virtualization = true
  computer_name                  = "compName"

  memory_hot_add_enabled = true

  sizing_policy_id = vcd_vm_sizing_policy.size_full.id
 }

resource "vcd_vapp_vm" "{{.VMName}}3" {
  org = "{{.Org}}"
  vdc = vcd_org_vdc.{{.VdcName}}.name

  vapp_name     = vcd_vapp.{{.VAppName}}.name
  name          = "{{.VMName}}3"
  catalog_name  = "{{.Catalog}}"
  template_name = "{{.CatalogItem}}"
  power_on      = "false"

  sizing_policy_id = vcd_vm_sizing_policy.size_cpu.id
  memory           = 2048
}

resource "vcd_vapp_vm" "{{.VMName}}4" {
  org = "{{.Org}}"
  vdc = vcd_org_vdc.{{.VdcName}}.name

  vapp_name     = vcd_vapp.{{.VAppName}}.name
  name          = "{{.VMName}}4"
  catalog_name  = "{{.Catalog}}"
  template_name = "{{.CatalogItem}}"
  power_on      = "false"

  sizing_policy_id = vcd_vm_sizing_policy.size_full.id
 }
`

const testAccCheckVcdVAppEmptyVmWithSizingUpdate = "# skip-binary-test: only for updates " +
	"" + testAccCheckVcdVAppEmptyWithSizing + `

resource "vcd_vapp_vm" "{{.VMName}}" {
  org = "{{.Org}}"
  vdc = vcd_org_vdc.{{.VdcName}}.name

  vapp_name     = vcd_vapp.{{.VAppName}}.name
  name          = "{{.VMName}}"
  description   = "test empty VM updated"

  os_type                        = "sles11_64Guest"
  hardware_version               = "vmx-13"
  catalog_name                   = ""
  expose_hardware_virtualization = false
  computer_name                  = "compNameUp"

  memory_hot_add_enabled = true

  sizing_policy_id = vcd_vm_sizing_policy.size_full.id
}

resource "vcd_vapp_vm" "{{.VMName}}2" {
  org = "{{.Org}}"
  vdc = vcd_org_vdc.{{.VdcName}}.name

  power_on = true

  vapp_name     = vcd_vapp.{{.VAppName}}.name
  description   = "test empty VM2"
  name          = "{{.VMName}}2"
  
  os_type                        = "sles11_64Guest"
  hardware_version               = "vmx-13"
  catalog_name                   = "{{.Catalog}}"
  expose_hardware_virtualization = true
  computer_name                  = "compName"

  memory_hot_add_enabled = true

  sizing_policy_id = vcd_vm_sizing_policy.size_cpu.id
  # allows to change only not defined in sizing policy
  memory           = 2048
}

resource "vcd_vapp_vm" "{{.VMName}}3" {
  org = "{{.Org}}"
  vdc = vcd_org_vdc.{{.VdcName}}.name

  vapp_name     = vcd_vapp.{{.VAppName}}.name
  name          = "{{.VMName}}3"
  catalog_name  = "{{.Catalog}}"
  template_name = "{{.CatalogItem}}"
  power_on      = "false"

  sizing_policy_id = vcd_vm_sizing_policy.size_cpu.id
  # allows to change only not defined in sizing policy
  memory           = 3072
}

resource "vcd_vapp_vm" "{{.VMName}}5" {
  org = "{{.Org}}"
  vdc = vcd_org_vdc.{{.VdcName}}.name

  vapp_name     = vcd_vapp.{{.VAppName}}.name
  name          = "{{.VMName}}5"
  catalog_name  = "{{.Catalog}}"
  template_name = "{{.CatalogItem}}"
  power_on      = "false"

  sizing_policy_id = vcd_vm_sizing_policy.minSize.id
  # allows to change only not defined in sizing policy
  cpus      = 4
  cpu_cores = 2
  memory    = 1536
}
`
