//go:build (standaloneVm || vm || ALL || functional) && !skipStandaloneVm
// +build standaloneVm vm ALL functional
// +build !skipStandaloneVm

package vcd

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccVcdStandaloneVmWithVmSizing(t *testing.T) {
	preTestChecks(t)
	var (
		standaloneVmName        = fmt.Sprintf("%s-%d", t.Name(), os.Getpid())
		netVmName1              = standaloneVmName + "-1"
		netVmName2              = standaloneVmName + "-2"
		netVmName3              = standaloneVmName + "-3"
		netVmName4              = standaloneVmName + "-4"
		netVmName5              = standaloneVmName + "-5"
		testAccVcdVdc    string = "test_VmSizing"
	)

	var params = StringMap{
		"Org":         testConfig.VCD.Org,
		"EdgeGateway": testConfig.Networking.EdgeGateway,
		"Catalog":     testSuiteCatalogName,
		"CatalogItem": testSuiteCatalogOVAItem,
		"VMName1":     netVmName1,
		"VMName2":     netVmName2,
		"VMName3":     netVmName3,
		"VMName4":     netVmName4,
		"VMName5":     netVmName5,
		"Tags":        "standaloneVm vm",

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

	if testConfig.VCD.ProviderVdc.StorageProfile == "" || testConfig.VCD.ProviderVdc.StorageProfile2 == "" {
		t.Skip("Both variables testConfig.VCD.ProviderVdc.StorageProfile and testConfig.VCD.ProviderVdc.StorageProfile2 must be set")
	}

	vcdClient, err := getTestVCDFromJson(testConfig)
	if err != nil {
		t.Skip("unable to validate vCD version - skipping test")
	}

	err = ProviderAuthenticate(vcdClient, testConfig.Provider.User, testConfig.Provider.Password, testConfig.Provider.Token, testConfig.Provider.SysOrg, testConfig.Provider.ApiToken)
	if err != nil {
		t.Skip(fmt.Sprintf("authentication error: %s", err))
	}
	if !vcdClient.Client.IsSysAdmin {
		t.Skip("Test can only run as System admin")
	}

	if vcdClient.Client.APIVCDMaxVersionIs("< 33.0") {
		t.Skip("TestAccVcdOrgVdcWithVmSizingPolicy requires VCD 10.0+")
	}

	configTextVM := templateFill(testAccCheckVcdEmptyVmWithSizing, params)

	params["FuncName"] = t.Name() + "-step1"
	configTextVMUpdateStep1 := templateFill(testAccCheckVcdEmptyVmWithSizingUpdate, params)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	debugPrintf("#[DEBUG] CONFIGURATION: %s\n", configTextVM)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVcdStandaloneVmDestroyByVdc(testAccVcdVdc),
		Steps: []resource.TestStep{
			// Step 0 - Create using sizing policy
			resource.TestStep{
				Config: configTextVM,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVcdStandaloneVmExistsByVdc(testAccVcdVdc, netVmName1, "vcd_vm."+netVmName1),
					resource.TestCheckResourceAttr("vcd_vm."+netVmName1, "name", netVmName1),

					resource.TestCheckResourceAttr("vcd_vm."+netVmName1, "os_type", "sles11_64Guest"),
					resource.TestCheckResourceAttr("vcd_vm."+netVmName1, "hardware_version", "vmx-13"),
					resource.TestCheckResourceAttr("vcd_vm."+netVmName1, "expose_hardware_virtualization", "true"),
					resource.TestCheckResourceAttr("vcd_vm."+netVmName1, "computer_name", "compName"),
					resource.TestCheckResourceAttr("vcd_vm."+netVmName1, "description", "test empty VM"),

					resource.TestCheckResourceAttrPair("vcd_vm."+netVmName1, "sizing_policy_id",
						"vcd_vm_sizing_policy.size_cpu", "id"),
					resource.TestCheckResourceAttr("vcd_vm."+netVmName1, "cpus", "3"),
					resource.TestCheckResourceAttr("vcd_vm."+netVmName1, "cpu_cores", "3"),
					resource.TestCheckResourceAttr("vcd_vm."+netVmName1, "memory", "1024"),

					testAccCheckVcdStandaloneVmExistsByVdc(testAccVcdVdc, netVmName2, "vcd_vm."+netVmName2),
					resource.TestCheckResourceAttr("vcd_vm."+netVmName2, "name", netVmName2),

					resource.TestCheckResourceAttr("vcd_vm."+netVmName2, "os_type", "sles11_64Guest"),
					resource.TestCheckResourceAttr("vcd_vm."+netVmName2, "hardware_version", "vmx-13"),
					resource.TestCheckResourceAttr("vcd_vm."+netVmName2, "expose_hardware_virtualization", "true"),
					resource.TestCheckResourceAttr("vcd_vm."+netVmName2, "computer_name", "compName"),
					resource.TestCheckResourceAttr("vcd_vm."+netVmName2, "description", "test empty VM2"),

					resource.TestCheckResourceAttrPair("vcd_vm."+netVmName2, "sizing_policy_id",
						"vcd_vm_sizing_policy.size_full", "id"),
					resource.TestCheckResourceAttr("vcd_vm."+netVmName2, "cpus", "3"),
					resource.TestCheckResourceAttr("vcd_vm."+netVmName2, "cpu_cores", "3"),
					resource.TestCheckResourceAttr("vcd_vm."+netVmName2, "memory", "2048"),

					testAccCheckVcdStandaloneVmExistsByVdc(testAccVcdVdc, netVmName3, "vcd_vm."+netVmName3),
					resource.TestCheckResourceAttr("vcd_vm."+netVmName3, "name", netVmName3),

					resource.TestCheckResourceAttrPair("vcd_vm."+netVmName3, "sizing_policy_id",
						"vcd_vm_sizing_policy.size_cpu", "id"),
					resource.TestCheckResourceAttr("vcd_vm."+netVmName3, "cpus", "3"),
					resource.TestCheckResourceAttr("vcd_vm."+netVmName3, "cpu_cores", "3"),
					resource.TestCheckResourceAttr("vcd_vm."+netVmName3, "memory", "2048"),

					testAccCheckVcdStandaloneVmExistsByVdc(testAccVcdVdc, netVmName4, "vcd_vm."+netVmName4),
					resource.TestCheckResourceAttr("vcd_vm."+netVmName4, "name", netVmName4),

					resource.TestCheckResourceAttrPair("vcd_vm."+netVmName4, "sizing_policy_id",
						"vcd_vm_sizing_policy.size_full", "id"),
					resource.TestCheckResourceAttr("vcd_vm."+netVmName4, "cpus", "3"),
					resource.TestCheckResourceAttr("vcd_vm."+netVmName4, "cpu_cores", "3"),
					resource.TestCheckResourceAttr("vcd_vm."+netVmName4, "memory", "2048"),
				),
			},
			// Step 1 - update
			resource.TestStep{
				Config: configTextVMUpdateStep1,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVcdStandaloneVmExistsByVdc(testAccVcdVdc, netVmName1, "vcd_vm."+netVmName1),
					resource.TestCheckResourceAttr("vcd_vm."+netVmName1, "name", netVmName1),

					resource.TestCheckResourceAttr("vcd_vm."+netVmName1, "os_type", "sles11_64Guest"),
					resource.TestCheckResourceAttr("vcd_vm."+netVmName1, "hardware_version", "vmx-13"),
					resource.TestCheckResourceAttr("vcd_vm."+netVmName1, "expose_hardware_virtualization", "false"),
					resource.TestCheckResourceAttr("vcd_vm."+netVmName1, "computer_name", "compNameUp"),
					resource.TestCheckResourceAttr("vcd_vm."+netVmName1, "description", "test empty VM updated"),

					resource.TestCheckResourceAttrPair("vcd_vm."+netVmName1, "sizing_policy_id",
						"vcd_vm_sizing_policy.size_full", "id"),
					resource.TestCheckResourceAttr("vcd_vm."+netVmName1, "cpus", "3"),
					resource.TestCheckResourceAttr("vcd_vm."+netVmName1, "cpu_cores", "3"),
					resource.TestCheckResourceAttr("vcd_vm."+netVmName1, "memory", "2048"),

					testAccCheckVcdStandaloneVmExistsByVdc(testAccVcdVdc, netVmName2, "vcd_vm."+netVmName2),
					resource.TestCheckResourceAttr("vcd_vm."+netVmName2, "name", netVmName2),

					resource.TestCheckResourceAttr("vcd_vm."+netVmName2, "os_type", "sles11_64Guest"),
					resource.TestCheckResourceAttr("vcd_vm."+netVmName2, "hardware_version", "vmx-13"),
					resource.TestCheckResourceAttr("vcd_vm."+netVmName2, "expose_hardware_virtualization", "true"),
					resource.TestCheckResourceAttr("vcd_vm."+netVmName2, "computer_name", "compName"),
					resource.TestCheckResourceAttr("vcd_vm."+netVmName2, "description", "test empty VM2"),

					resource.TestCheckResourceAttrPair("vcd_vm."+netVmName2, "sizing_policy_id",
						"vcd_vm_sizing_policy.size_cpu", "id"),
					resource.TestCheckResourceAttr("vcd_vm."+netVmName2, "cpus", "3"),
					resource.TestCheckResourceAttr("vcd_vm."+netVmName2, "cpu_cores", "3"),
					resource.TestCheckResourceAttr("vcd_vm."+netVmName2, "memory", "2048"),

					testAccCheckVcdStandaloneVmExistsByVdc(testAccVcdVdc, netVmName3, "vcd_vm."+netVmName3),
					resource.TestCheckResourceAttr("vcd_vm."+netVmName3, "name", netVmName3),

					resource.TestCheckResourceAttrPair("vcd_vm."+netVmName3, "sizing_policy_id",
						"vcd_vm_sizing_policy.size_cpu", "id"),
					resource.TestCheckResourceAttr("vcd_vm."+netVmName3, "cpus", "3"),
					resource.TestCheckResourceAttr("vcd_vm."+netVmName3, "cpu_cores", "3"),
					resource.TestCheckResourceAttr("vcd_vm."+netVmName3, "memory", "3072"),

					testAccCheckVcdStandaloneVmExistsByVdc(testAccVcdVdc, netVmName5, "vcd_vm."+netVmName5),
					resource.TestCheckResourceAttr("vcd_vm."+netVmName5, "name", netVmName5),

					resource.TestCheckResourceAttrPair("vcd_vm."+netVmName5, "sizing_policy_id",
						"vcd_vm_sizing_policy.minSize", "id"),
					resource.TestCheckResourceAttr("vcd_vm."+netVmName5, "cpus", "4"),
					resource.TestCheckResourceAttr("vcd_vm."+netVmName5, "cpu_cores", "2"),
					resource.TestCheckResourceAttr("vcd_vm."+netVmName5, "memory", "1536"),
				),
			},
		},
	})
	postTestChecks(t)
}

func testAccCheckVcdStandaloneVmExistsByVdc(vdcName, vmName, node string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[node]
		if !ok {
			return fmt.Errorf("not found: %s", node)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no VM ID is set")
		}

		conn := testAccProvider.Meta().(*VCDClient)
		_, vdc, err := conn.GetOrgAndVdc(testConfig.VCD.Org, vdcName)
		if err != nil {
			return fmt.Errorf(errorRetrievingVdcFromOrg, vdcName, testConfig.VCD.Org, err)
		}

		_, err = vdc.QueryVmByName(vmName)

		return err
	}
}

func testAccCheckVcdStandaloneVmDestroyByVdc(vdcName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*VCDClient)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "vcd_vm" {
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

const testAccCheckVcdEmptyWithSizing = `
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
`
const testAccCheckVcdEmptyVmWithSizing = testAccCheckVcdEmptyWithSizing + `
resource "vcd_vm" "{{.VMName1}}" {
  org = "{{.Org}}"
  vdc = vcd_org_vdc.{{.VdcName}}.name

  power_on = true

  description   = "test empty VM"
  name          = "{{.VMName1}}"
  
  os_type                        = "sles11_64Guest"
  hardware_version               = "vmx-13"
  catalog_name                   = "{{.Catalog}}"
  expose_hardware_virtualization = true
  computer_name                  = "compName"

  memory_hot_add_enabled = true

  sizing_policy_id = vcd_vm_sizing_policy.size_cpu.id
  memory           = 1024
 }

resource "vcd_vm" "{{.VMName2}}" {
  org = "{{.Org}}"
  vdc = vcd_org_vdc.{{.VdcName}}.name

  power_on = true

  description   = "test empty VM2"
  name          = "{{.VMName2}}"
  
  os_type                        = "sles11_64Guest"
  hardware_version               = "vmx-13"
  catalog_name                   = "{{.Catalog}}"
  expose_hardware_virtualization = true
  computer_name                  = "compName"

  memory_hot_add_enabled = true

  sizing_policy_id = vcd_vm_sizing_policy.size_full.id
 }

resource "vcd_vm" "{{.VMName3}}" {
  org = "{{.Org}}"
  vdc = vcd_org_vdc.{{.VdcName}}.name

  name          = "{{.VMName3}}"
  catalog_name  = "{{.Catalog}}"
  template_name = "{{.CatalogItem}}"
  power_on      = "false"

  sizing_policy_id = vcd_vm_sizing_policy.size_cpu.id
  memory           = 2048
}

resource "vcd_vm" "{{.VMName4}}" {
  org = "{{.Org}}"
  vdc = vcd_org_vdc.{{.VdcName}}.name

  name          = "{{.VMName4}}"
  catalog_name  = "{{.Catalog}}"
  template_name = "{{.CatalogItem}}"
  power_on      = "false"

  sizing_policy_id = vcd_vm_sizing_policy.size_full.id
 }
`

const testAccCheckVcdEmptyVmWithSizingUpdate = "# skip-binary-test: only for updates " +
	"" + testAccCheckVcdEmptyWithSizing + `

resource "vcd_vm" "{{.VMName1}}" {
  org = "{{.Org}}"
  vdc = vcd_org_vdc.{{.VdcName}}.name

  name          = "{{.VMName1}}"
  description   = "test empty VM updated"

  os_type                        = "sles11_64Guest"
  hardware_version               = "vmx-13"
  catalog_name                   = ""
  expose_hardware_virtualization = false
  computer_name                  = "compNameUp"

  memory_hot_add_enabled = true

  sizing_policy_id = vcd_vm_sizing_policy.size_full.id
}

resource "vcd_vm" "{{.VMName2}}" {
  org = "{{.Org}}"
  vdc = vcd_org_vdc.{{.VdcName}}.name

  power_on = true

  description   = "test empty VM2"
  name          = "{{.VMName2}}"
  
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

resource "vcd_vm" "{{.VMName3}}" {
  org = "{{.Org}}"
  vdc = vcd_org_vdc.{{.VdcName}}.name

  name          = "{{.VMName3}}"
  catalog_name  = "{{.Catalog}}"
  template_name = "{{.CatalogItem}}"
  power_on      = "false"

  sizing_policy_id = vcd_vm_sizing_policy.size_cpu.id
  # allows to change only not defined in sizing policy
  memory           = 3072
}

resource "vcd_vm" "{{.VMName5}}" {
  org = "{{.Org}}"
  vdc = vcd_org_vdc.{{.VdcName}}.name

  name          = "{{.VMName5}}"
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
