// +build vdc nsxt ALL functional

package vcd

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

var TestAccVcdVdc = "TestAccVcdVdcBasic"

func runOrgVdcTest(t *testing.T, params StringMap, allocationModel string) {

	if !usingSysAdmin() {
		t.Skip("TestAccVcdVdcBasic requires system admin privileges")
		return
	}

	configText := templateFill(testAccCheckVcdVdc_basic, params)
	params["FuncName"] = t.Name() + "-Update"
	updateText := templateFill(testAccCheckVcdVdc_update, params)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", updateText)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVdcDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdVdcExists("vcd_org_vdc."+params["VdcName"].(string)),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+params["VdcName"].(string), "name", params["VdcName"].(string)),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+params["VdcName"].(string), "org", testConfig.VCD.Org),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+params["VdcName"].(string), "allocation_model", allocationModel),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+params["VdcName"].(string), "network_pool_name", params["NetworkPool"].(string)),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+params["VdcName"].(string), "provider_vdc_name", params["ProviderVdc"].(string)),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+params["VdcName"].(string), "enabled", "true"),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+params["VdcName"].(string), "enable_thin_provisioning", "true"),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+params["VdcName"].(string), "enable_fast_provisioning", "true"),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+params["VdcName"].(string), "delete_force", "true"),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+params["VdcName"].(string), "delete_recursive", "true"),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+params["VdcName"].(string), "metadata.vdc_metadata", "VDC Metadata"),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+params["VdcName"].(string), "storage_profile.0.name", params["ProviderVdcStorageProfile"].(string)),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+params["VdcName"].(string), "storage_profile.0.enabled", "true"),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+params["VdcName"].(string), "storage_profile.0.limit", "10240"),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+params["VdcName"].(string), "storage_profile.0.default", "true"),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+params["VdcName"].(string), "compute_capacity.0.cpu.0.allocated", params["Allocated"].(string)),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+params["VdcName"].(string), "compute_capacity.0.cpu.0.limit", params["Limit"].(string)),
					resource.TestMatchResourceAttr(
						"vcd_org_vdc."+params["VdcName"].(string), "compute_capacity.0.cpu.0.reserved", regexp.MustCompile(`^\d+$`)),
					resource.TestMatchResourceAttr(
						"vcd_org_vdc."+params["VdcName"].(string), "compute_capacity.0.cpu.0.used", regexp.MustCompile(`^\d+$`)),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+params["VdcName"].(string), "compute_capacity.0.memory.0.allocated", params["Allocated"].(string)),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+params["VdcName"].(string), "compute_capacity.0.memory.0.limit", params["Limit"].(string)),
					resource.TestMatchResourceAttr(
						"vcd_org_vdc."+params["VdcName"].(string), "compute_capacity.0.memory.0.reserved", regexp.MustCompile(`^\d+$`)),
					resource.TestMatchResourceAttr(
						"vcd_org_vdc."+params["VdcName"].(string), "compute_capacity.0.memory.0.used", regexp.MustCompile(`^\d+$`)),
					resource.TestMatchResourceAttr(
						"vcd_org_vdc."+params["VdcName"].(string), "elasticity", regexp.MustCompile(`^`+params["ElasticityValueForAssert"].(string)+`$`)),
					resource.TestMatchResourceAttr(
						"vcd_org_vdc."+params["VdcName"].(string), "include_vm_memory_overhead", regexp.MustCompile(`^`+params["MemoryOverheadValueForAssert"].(string)+`$`)),
				),
			},
			resource.TestStep{
				Config: updateText,
				Check: resource.ComposeTestCheckFunc(
					testVcdVdcUpdated("vcd_org_vdc."+params["VdcName"].(string)),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+params["VdcName"].(string), "name", params["VdcName"].(string)),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+params["VdcName"].(string), "org", testConfig.VCD.Org),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+params["VdcName"].(string), "allocation_model", allocationModel),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+params["VdcName"].(string), "network_pool_name", params["NetworkPool"].(string)),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+params["VdcName"].(string), "provider_vdc_name", params["ProviderVdc"].(string)),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+params["VdcName"].(string), "enabled", "false"),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+params["VdcName"].(string), "enable_thin_provisioning", "false"),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+params["VdcName"].(string), "enable_fast_provisioning", "false"),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+params["VdcName"].(string), "delete_force", "false"),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+params["VdcName"].(string), "delete_recursive", "false"),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+params["VdcName"].(string), "memory_guaranteed", params["MemoryGuaranteed"].(string)),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+params["VdcName"].(string), "cpu_guaranteed", params["CpuGuaranteed"].(string)),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+params["VdcName"].(string), "metadata.vdc_metadata", "VDC Metadata"),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+params["VdcName"].(string), "metadata.vdc_metadata2", "VDC Metadata2"),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+params["VdcName"].(string), "metadata.vdc_metadata2", "VDC Metadata2"),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+params["VdcName"].(string), "storage_profile.0.name", params["ProviderVdcStorageProfile"].(string)),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+params["VdcName"].(string), "storage_profile.0.enabled", "true"),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+params["VdcName"].(string), "storage_profile.0.limit", "20480"),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+params["VdcName"].(string), "storage_profile.0.default", "true"),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+params["VdcName"].(string), "compute_capacity.0.cpu.0.allocated", params["AllocatedIncreased"].(string)),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+params["VdcName"].(string), "compute_capacity.0.cpu.0.limit", params["LimitIncreased"].(string)),
					resource.TestMatchResourceAttr(
						"vcd_org_vdc."+params["VdcName"].(string), "compute_capacity.0.cpu.0.reserved", regexp.MustCompile(`^\d+$`)),
					resource.TestMatchResourceAttr(
						"vcd_org_vdc."+params["VdcName"].(string), "compute_capacity.0.cpu.0.used", regexp.MustCompile(`^\d+$`)),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+params["VdcName"].(string), "compute_capacity.0.memory.0.allocated", params["AllocatedIncreased"].(string)),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+params["VdcName"].(string), "compute_capacity.0.memory.0.limit", params["LimitIncreased"].(string)),
					resource.TestMatchResourceAttr(
						"vcd_org_vdc."+params["VdcName"].(string), "compute_capacity.0.memory.0.reserved", regexp.MustCompile(`^\d+$`)),
					resource.TestMatchResourceAttr(
						"vcd_org_vdc."+params["VdcName"].(string), "compute_capacity.0.memory.0.used", regexp.MustCompile(`^\d+$`)),
					resource.TestMatchResourceAttr(
						"vcd_org_vdc."+params["VdcName"].(string), "elasticity", regexp.MustCompile(`^`+params["ElasticityUpdateValueForAssert"].(string)+`$`)),
					resource.TestMatchResourceAttr(
						"vcd_org_vdc."+params["VdcName"].(string), "include_vm_memory_overhead", regexp.MustCompile(`^`+params["MemoryOverheadUpdateValueForAssert"].(string)+`$`)),
				),
			},
			resource.TestStep{
				ResourceName:      "vcd_org_vdc." + params["VdcName"].(string),
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdOrgObject(testConfig, params["VdcName"].(string)),
				// These fields can't be retrieved
				ImportStateVerifyIgnore: []string{"delete_force", "delete_recursive"},
			},
		},
	})
}

func testAccCheckVcdVdcExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no VDC ID is set")
		}

		conn := testAccProvider.Meta().(*VCDClient)

		adminOrg, err := conn.GetAdminOrg(testConfig.VCD.Org)
		if err != nil {
			return fmt.Errorf(errorRetrievingOrg, testConfig.VCD.Org+" and error: "+err.Error())
		}

		_, err = adminOrg.GetVDCByName(rs.Primary.Attributes["name"], false)
		if err != nil {
			return fmt.Errorf("vdc %s does not exist (%s)", rs.Primary.Attributes["name"], err)
		}

		return nil
	}
}

func testVcdVdcUpdated(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no VDC ID is set")
		}

		conn := testAccProvider.Meta().(*VCDClient)

		adminOrg, err := conn.GetAdminOrg(testConfig.VCD.Org)
		if err != nil {
			return fmt.Errorf(errorRetrievingOrg, testConfig.VCD.Org+" and error: "+err.Error())
		}

		updateVdc, err := adminOrg.GetVDCByName(rs.Primary.Attributes["name"], false)
		if err != nil {
			return fmt.Errorf("vdc %s does not exist (%s)", rs.Primary.Attributes["name"], err)
		}

		if updateVdc.Vdc.IsEnabled != false {
			return fmt.Errorf("VDC update failed - VDC still enabled")
		}
		return nil
	}
}

func testAccCheckVdcDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*VCDClient)
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "vcd_org_vdc" && rs.Primary.Attributes["name"] != TestAccVcdVdc {
			continue
		}

		adminOrg, err := conn.GetAdminOrg(testConfig.VCD.Org)
		if err != nil {
			return fmt.Errorf(errorRetrievingOrg, testConfig.VCD.Org+" and error: "+err.Error())
		}

		_, err = adminOrg.GetVDCByName(rs.Primary.ID, false)

		if err == nil {
			return fmt.Errorf("vdc %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

const testAccCheckVcdVdc_basic = `
resource "vcd_org_vdc" "{{.VdcName}}" {
  name = "{{.VdcName}}"
  org  = "{{.OrgName}}"

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
    name    = "{{.ProviderVdcStorageProfile}}"
    enabled = true
    limit   = 10240
    default = true
  }

  metadata = {
    vdc_metadata = "VDC Metadata"
  }

  enabled                    = true
  enable_thin_provisioning   = true
  enable_fast_provisioning   = true
  delete_force               = true
  delete_recursive           = true
  {{.FlexElasticKey}}        {{.equalsChar}} {{.FlexElasticValue}}
  {{.FlexMemoryOverheadKey}} {{.equalsChar}} {{.FlexMemoryOverheadValue}}
}
`

const testAccCheckVcdVdc_update = `
# skip-binary-test: only for updates
resource "vcd_org_vdc" "{{.VdcName}}" {
  name = "{{.VdcName}}"
  org  = "{{.OrgName}}"

  allocation_model  = "{{.AllocationModel}}"
  network_pool_name = "{{.NetworkPool}}"
  provider_vdc_name = "{{.ProviderVdc}}"

  compute_capacity {
    cpu {
      allocated = "{{.AllocatedIncreased}}"
      limit     = "{{.LimitIncreased}}"
    }

    memory {
      allocated = "{{.AllocatedIncreased}}"
      limit     = "{{.LimitIncreased}}"
    }
  }

  storage_profile {
    name    = "{{.ProviderVdcStorageProfile}}"
    enabled = true
    limit   = 20480
    default = true
  }

  metadata = {
    vdc_metadata  = "VDC Metadata"
    vdc_metadata2 = "VDC Metadata2"
  }

  cpu_guaranteed             = {{.CpuGuaranteed}}
  memory_guaranteed          = {{.MemoryGuaranteed}}
  enabled                    = false
  enable_thin_provisioning   = false
  enable_fast_provisioning   = false
  delete_force               = false
  delete_recursive           = false
  {{.FlexElasticKey}}        {{.equalsChar}} {{.FlexElasticValueUpdate}}
  {{.FlexMemoryOverheadKey}} {{.equalsChar}} {{.FlexMemoryOverheadValueUpdate}}
}
`
