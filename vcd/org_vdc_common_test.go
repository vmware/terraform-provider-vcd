// +build vdc nsxt ALL functional

package vcd

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"testing"
	"text/template"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var TestAccVcdVdc = "TestAccVcdVdcBasic"

func runOrgVdcTest(t *testing.T, params StringMap, allocationModel string) {

	if !usingSysAdmin() {
		t.Skip(t.Name() + " requires system admin privileges")
		return
	}

	secondStorageProfile := params["ProviderVdcStorageProfile2"].(string)
	configText := templateFill(testAccCheckVcdVdc_basic, params)
	params["SecondStorageProfile"] = ""

	// If a second storage profile is defined in the configuration, we add its parameters in the update
	if secondStorageProfile != "" {
		unfilledTemplate := template.Must(template.New("").Parse(additionalStorageProfile))
		buf := &bytes.Buffer{}
		err := unfilledTemplate.Execute(buf, map[string]interface{}{
			"StorageProfileName":    secondStorageProfile,
			"StorageProfileDefault": false,
		})
		if err == nil {
			params["SecondStorageProfile"] = buf.String()
		} else {
			fmt.Printf("[WARNING] error reported while filling second storage profile details: %s\n", err)
		}
	}

	params["FuncName"] = t.Name() + "-Update"
	updateText := templateFill(testAccCheckVcdVdc_update, params)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", updateText)
	secondUpdateText := strings.Replace(updateText, "#START_STORAGE_PROFILE", "/*", 1)
	secondUpdateText = strings.Replace(secondUpdateText, "#END_STORAGE_PROFILE", "*/", 1)

	resourceDef := "vcd_org_vdc." + params["VdcName"].(string)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVdcDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdVdcExists("vcd_org_vdc."+params["VdcName"].(string)),
					resource.TestCheckResourceAttr(
						resourceDef, "name", params["VdcName"].(string)),
					resource.TestCheckResourceAttr(
						resourceDef, "org", testConfig.VCD.Org),
					resource.TestCheckResourceAttr(
						resourceDef, "allocation_model", allocationModel),
					resource.TestCheckResourceAttr(
						resourceDef, "network_pool_name", params["NetworkPool"].(string)),
					resource.TestCheckResourceAttr(
						resourceDef, "provider_vdc_name", params["ProviderVdc"].(string)),
					resource.TestCheckResourceAttr(
						resourceDef, "enabled", "true"),
					resource.TestCheckResourceAttr(
						resourceDef, "enable_thin_provisioning", "true"),
					resource.TestCheckResourceAttr(
						resourceDef, "enable_fast_provisioning", "true"),
					resource.TestCheckResourceAttr(
						resourceDef, "delete_force", "true"),
					resource.TestCheckResourceAttr(
						resourceDef, "delete_recursive", "true"),
					resource.TestCheckResourceAttr(
						resourceDef, "metadata.vdc_metadata", "VDC Metadata"),
					resource.TestCheckResourceAttr(
						resourceDef, "storage_profile.0.name", params["ProviderVdcStorageProfile"].(string)),
					resource.TestCheckResourceAttr(
						resourceDef, "storage_profile.0.enabled", "true"),
					resource.TestCheckResourceAttr(
						resourceDef, "storage_profile.0.limit", "10240"),
					resource.TestCheckResourceAttr(
						resourceDef, "storage_profile.0.default", "true"),
					resource.TestCheckResourceAttr(
						resourceDef, "compute_capacity.0.cpu.0.allocated", params["Allocated"].(string)),
					resource.TestCheckResourceAttr(
						resourceDef, "compute_capacity.0.cpu.0.limit", params["Limit"].(string)),
					resource.TestMatchResourceAttr(
						resourceDef, "compute_capacity.0.cpu.0.reserved", regexp.MustCompile(`^\d+$`)),
					resource.TestMatchResourceAttr(
						resourceDef, "compute_capacity.0.cpu.0.used", regexp.MustCompile(`^\d+$`)),
					resource.TestCheckResourceAttr(
						resourceDef, "compute_capacity.0.memory.0.allocated", params["Allocated"].(string)),
					resource.TestCheckResourceAttr(
						resourceDef, "compute_capacity.0.memory.0.limit", params["Limit"].(string)),
					resource.TestMatchResourceAttr(
						resourceDef, "compute_capacity.0.memory.0.reserved", regexp.MustCompile(`^\d+$`)),
					resource.TestMatchResourceAttr(
						resourceDef, "compute_capacity.0.memory.0.used", regexp.MustCompile(`^\d+$`)),
					resource.TestMatchResourceAttr(
						resourceDef, "elasticity", regexp.MustCompile(`^`+params["ElasticityValueForAssert"].(string)+`$`)),
					resource.TestMatchResourceAttr(
						resourceDef, "include_vm_memory_overhead", regexp.MustCompile(`^`+params["MemoryOverheadValueForAssert"].(string)+`$`)),
				),
			},
			resource.TestStep{
				Config: updateText,
				Check: resource.ComposeTestCheckFunc(
					testVcdVdcUpdated("vcd_org_vdc."+params["VdcName"].(string)),
					resource.TestCheckResourceAttr(
						resourceDef, "name", params["VdcName"].(string)),
					resource.TestCheckResourceAttr(
						resourceDef, "org", testConfig.VCD.Org),
					resource.TestCheckResourceAttr(
						resourceDef, "allocation_model", allocationModel),
					resource.TestCheckResourceAttr(
						resourceDef, "network_pool_name", params["NetworkPool"].(string)),
					resource.TestCheckResourceAttr(
						resourceDef, "provider_vdc_name", params["ProviderVdc"].(string)),
					resource.TestCheckResourceAttr(
						resourceDef, "enabled", "false"),
					resource.TestCheckResourceAttr(
						resourceDef, "enable_thin_provisioning", "false"),
					resource.TestCheckResourceAttr(
						resourceDef, "enable_fast_provisioning", "false"),
					resource.TestCheckResourceAttr(
						resourceDef, "delete_force", "false"),
					resource.TestCheckResourceAttr(
						resourceDef, "delete_recursive", "false"),
					resource.TestCheckResourceAttr(
						resourceDef, "memory_guaranteed", params["MemoryGuaranteed"].(string)),
					resource.TestCheckResourceAttr(
						resourceDef, "cpu_guaranteed", params["CpuGuaranteed"].(string)),
					resource.TestCheckResourceAttr(
						resourceDef, "metadata.vdc_metadata", "VDC Metadata"),
					resource.TestCheckResourceAttr(
						resourceDef, "metadata.vdc_metadata2", "VDC Metadata2"),
					resource.TestCheckResourceAttr(
						resourceDef, "metadata.vdc_metadata2", "VDC Metadata2"),
					testAccFindValuesInSet(resourceDef, "storage_profile", map[string]string{
						"name":    params["ProviderVdcStorageProfile"].(string),
						"enabled": "true",
						"default": "true",
						"limit":   "20480",
					}),
					// This test runs only if we have a second storage profile
					testConditionalCheck(secondStorageProfile != "",
						testAccFindValuesInSet(resourceDef, "storage_profile", map[string]string{
							"name":    secondStorageProfile,
							"enabled": "true",
							"default": "false",
							"limit":   "20480",
						})),
					resource.TestCheckResourceAttr(
						resourceDef, "compute_capacity.0.cpu.0.allocated", params["AllocatedIncreased"].(string)),
					resource.TestCheckResourceAttr(
						resourceDef, "compute_capacity.0.cpu.0.limit", params["LimitIncreased"].(string)),
					resource.TestMatchResourceAttr(
						resourceDef, "compute_capacity.0.cpu.0.reserved", regexp.MustCompile(`^\d+$`)),
					resource.TestMatchResourceAttr(
						resourceDef, "compute_capacity.0.cpu.0.used", regexp.MustCompile(`^\d+$`)),
					resource.TestCheckResourceAttr(
						resourceDef, "compute_capacity.0.memory.0.allocated", params["AllocatedIncreased"].(string)),
					resource.TestCheckResourceAttr(
						resourceDef, "compute_capacity.0.memory.0.limit", params["LimitIncreased"].(string)),
					resource.TestMatchResourceAttr(
						resourceDef, "compute_capacity.0.memory.0.reserved", regexp.MustCompile(`^\d+$`)),
					resource.TestMatchResourceAttr(
						resourceDef, "compute_capacity.0.memory.0.used", regexp.MustCompile(`^\d+$`)),
					resource.TestMatchResourceAttr(
						resourceDef, "elasticity", regexp.MustCompile(`^`+params["ElasticityUpdateValueForAssert"].(string)+`$`)),
					resource.TestMatchResourceAttr(
						resourceDef, "include_vm_memory_overhead", regexp.MustCompile(`^`+params["MemoryOverheadUpdateValueForAssert"].(string)+`$`)),
					// This test runs only if we have a second storage profile
					testConditionalCheck(secondStorageProfile != "",
						resource.TestCheckResourceAttr(resourceDef, "storage_profile.#", "2")),
				),
			},
			// Test removal of second storage profile
			resource.TestStep{
				Config: secondUpdateText,
				// This test runs only if we have a second storage profile
				Check: testConditionalCheck(secondStorageProfile != "", resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceDef, "storage_profile.#", "1"),
					testAccFindValuesInSet(resourceDef, "storage_profile", map[string]string{
						"name":    params["ProviderVdcStorageProfile"].(string),
						"enabled": "true",
						"default": "true",
						"limit":   "20480",
					}),
				)),
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

// testConditionalCheck runs the wanted check only if the preliminary condition is true
func testConditionalCheck(condition bool, f resource.TestCheckFunc) resource.TestCheckFunc {
	if condition {
		return f
	}
	return func(s *terraform.State) error { return nil }
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

  {{.SecondStorageProfile}}

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

// additionalStorageProfile is a component that allows the insertion of a second storage profile
// when one was defined in the configuration file.
// The start/end labels will be replaced by comment markers, thus eliminating the
// second storage profile from the script, so that we can test the removal of the storage profile.
const additionalStorageProfile = `
  #START_STORAGE_PROFILE
  storage_profile {
    name    = "{{.StorageProfileName}}"
    enabled = true
    limit   = 20480
    default = {{.StorageProfileDefault}}
  }
  #END_STORAGE_PROFILE
`
