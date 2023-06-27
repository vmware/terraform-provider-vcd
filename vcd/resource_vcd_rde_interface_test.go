//go:build rde || ALL || functional

package vcd

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/vmware/go-vcloud-director/v2/govcd"
)

func TestAccVcdRdeDefinedInterface(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	var params = StringMap{
		"Nss":     "nss1",
		"Version": "1.0.0",
		"Vendor":  "vendor1",
		"Name":    t.Name(),
	}
	testParamsNotEmpty(t, params)

	configTextCreate := templateFill(testAccVcdRdeDefinedInterface, params)
	params["FuncName"] = t.Name() + "-Update"
	params["Name"] = params["FuncName"]
	configTextUpdate := templateFill(testAccVcdRdeDefinedInterface, params)

	// We change the nss to force deletion and re-creation
	params["FuncName"] = t.Name() + "-ForceNew"
	params["Name"] = t.Name()
	params["Nss"] = "nss2"
	configTextForceNew := templateFill(testAccVcdRdeDefinedInterface, params)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	debugPrintf("#[DEBUG] CONFIGURATION create: %s\n", configTextCreate)
	debugPrintf("#[DEBUG] CONFIGURATION update: %s\n", configTextUpdate)
	debugPrintf("#[DEBUG] CONFIGURATION force new: %s\n", configTextForceNew)

	interfaceName := "vcd_rde_interface.interface1"
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckRdeInterfaceDestroy(interfaceName),
		Steps: []resource.TestStep{
			{
				Config: configTextCreate,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(interfaceName, "id", fmt.Sprintf("urn:vcloud:interface:%s:%s:%s", params["Vendor"].(string), "nss1", params["Version"].(string))),
					resource.TestCheckResourceAttr(interfaceName, "nss", "nss1"),
					resource.TestCheckResourceAttr(interfaceName, "version", params["Version"].(string)),
					resource.TestCheckResourceAttr(interfaceName, "vendor", params["Vendor"].(string)),
					resource.TestCheckResourceAttr(interfaceName, "name", t.Name()),
					resource.TestCheckResourceAttr(interfaceName, "readonly", "false"),
				),
			},
			{
				Config: configTextUpdate,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(interfaceName, "id", fmt.Sprintf("urn:vcloud:interface:%s:%s:%s", params["Vendor"].(string), "nss1", params["Version"].(string))),
					resource.TestCheckResourceAttr(interfaceName, "nss", "nss1"),
					resource.TestCheckResourceAttr(interfaceName, "version", params["Version"].(string)),
					resource.TestCheckResourceAttr(interfaceName, "vendor", params["Vendor"].(string)),
					resource.TestCheckResourceAttr(interfaceName, "name", t.Name()+"-Update"),
					resource.TestCheckResourceAttr(interfaceName, "readonly", "false"),
				),
			},
			{
				Config: configTextForceNew,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(interfaceName, "id", fmt.Sprintf("urn:vcloud:interface:%s:%s:%s", params["Vendor"].(string), "nss2", params["Version"].(string))),
					resource.TestCheckResourceAttr(interfaceName, "nss", "nss2"),
					resource.TestCheckResourceAttr(interfaceName, "version", params["Version"].(string)),
					resource.TestCheckResourceAttr(interfaceName, "vendor", params["Vendor"].(string)),
					resource.TestCheckResourceAttr(interfaceName, "name", t.Name()),
					resource.TestCheckResourceAttr(interfaceName, "readonly", "false"),
				),
			},
			{
				ResourceName:      interfaceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdDefinedInterface(params["Vendor"].(string), params["Nss"].(string), params["Version"].(string)),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdRdeDefinedInterface = `
resource "vcd_rde_interface" "interface1" {
  nss     = "{{.Nss}}"
  version = "{{.Version}}"
  vendor  = "{{.Vendor}}"
  name    = "{{.Name}}"
}
`

func TestAccVcdRdeDefinedInterfacBehavior(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	var params = StringMap{
		"Nss":                   "nss1",
		"Version":               "1.0.0",
		"Vendor":                "vendor1",
		"Name":                  t.Name(),
		"BehaviorName":          "TestBehavior",
		"BehaviorExecutionId":   "TestActivity",
		"BehaviorExecutionType": "Activity",
		"BehaviorDescription":   t.Name(),
		"ExtraBehavior":         " ",
	}
	testParamsNotEmpty(t, params)

	configText1 := templateFill(testAccVcdRdeDefinedInterfaceBehavior, params)
	debugPrintf("#[DEBUG] CONFIGURATION 1: %s\n", configText1)

	params["FuncName"] = t.Name() + "-Step2"
	params["ExtraBehavior"] = `
	behavior {
    	name = "TestBehavior2"
    	execution = {
      		"id":   "TestActivity2"
      		"type": "noop"
    	}
    	description = "` + t.Name() + `2"
  	}
	`
	params["BehaviorName"] = "TestBehaviorUpdate"
	params["BehaviorDescription"] = t.Name() + "Update"
	params["BehaviorExecutionId"] = "TestActivityUpdate"
	configText2 := templateFill(testAccVcdRdeDefinedInterfaceBehavior, params)
	debugPrintf("#[DEBUG] CONFIGURATION 2: %s\n", configText2)

	params["FuncName"] = t.Name() + "-Step3"
	configText3 := templateFill(testAccVcdRdeDefinedInterface, params)
	debugPrintf("#[DEBUG] CONFIGURATION 3: %s\n", configText3)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	interfaceName := "vcd_rde_interface.interface1"
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckRdeInterfaceDestroy(interfaceName),
		Steps: []resource.TestStep{
			{
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(interfaceName, "id", fmt.Sprintf("urn:vcloud:interface:%s:%s:%s", params["Vendor"].(string), "nss1", params["Version"].(string))),
					resource.TestCheckResourceAttr(interfaceName, "nss", "nss1"),
					resource.TestCheckResourceAttr(interfaceName, "version", params["Version"].(string)),
					resource.TestCheckResourceAttr(interfaceName, "vendor", params["Vendor"].(string)),
					resource.TestCheckResourceAttr(interfaceName, "name", t.Name()),
					resource.TestCheckResourceAttr(interfaceName, "readonly", "false"),
					resource.TestCheckTypeSetElemNestedAttrs(interfaceName, "behavior.*", map[string]string{
						"name":           "TestBehavior",
						"description":    t.Name(),
						"execution.%":    "2",
						"execution.id":   "TestActivity",
						"execution.type": "Activity",
					}),
				),
			},
			{
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(interfaceName, "id", fmt.Sprintf("urn:vcloud:interface:%s:%s:%s", params["Vendor"].(string), "nss1", params["Version"].(string))),
					resource.TestCheckResourceAttr(interfaceName, "nss", "nss1"),
					resource.TestCheckResourceAttr(interfaceName, "version", params["Version"].(string)),
					resource.TestCheckResourceAttr(interfaceName, "vendor", params["Vendor"].(string)),
					resource.TestCheckResourceAttr(interfaceName, "name", t.Name()),
					resource.TestCheckResourceAttr(interfaceName, "readonly", "false"),
					resource.TestCheckTypeSetElemNestedAttrs(interfaceName, "behavior.*", map[string]string{
						"name":           "TestBehaviorUpdate",
						"description":    t.Name() + "Update",
						"execution.%":    "2",
						"execution.id":   "TestActivityUpdate",
						"execution.type": "Activity",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(interfaceName, "behavior.*", map[string]string{
						"name":           "TestBehavior2",
						"description":    t.Name() + "2",
						"execution.%":    "2",
						"execution.id":   "TestActivity2",
						"execution.type": "noop",
					}),
				),
			},
			{
				Config: configText3,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(interfaceName, "id", fmt.Sprintf("urn:vcloud:interface:%s:%s:%s", params["Vendor"].(string), "nss1", params["Version"].(string))),
					resource.TestCheckResourceAttr(interfaceName, "nss", "nss1"),
					resource.TestCheckResourceAttr(interfaceName, "version", params["Version"].(string)),
					resource.TestCheckResourceAttr(interfaceName, "vendor", params["Vendor"].(string)),
					resource.TestCheckResourceAttr(interfaceName, "name", t.Name()),
					resource.TestCheckResourceAttr(interfaceName, "readonly", "false"),
					resource.TestCheckResourceAttr(interfaceName, "behavior.#", "0"),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdRdeDefinedInterfaceBehavior = `
resource "vcd_rde_interface" "interface1" {
  nss     = "{{.Nss}}"
  version = "{{.Version}}"
  vendor  = "{{.Vendor}}"
  name    = "{{.Name}}"

  behavior {
    name = "{{.BehaviorName}}"
    execution = {
      "id":   "{{.BehaviorExecutionId}}"
      "type": "{{.BehaviorExecutionType}}"
    }
    description = "{{.BehaviorDescription}}"
  }
  {{.ExtraBehavior}}
}`

// testAccCheckRdeInterfaceDestroy checks that the Defined Interface defined by its identifier no longer
// exists in VCD.
func testAccCheckRdeInterfaceDestroy(identifier string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[identifier]
		if !ok {
			return fmt.Errorf("not found: %s", identifier)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no Defined Interface ID is set")
		}

		conn := testAccProvider.Meta().(*VCDClient)

		_, err := conn.VCDClient.GetDefinedInterfaceById(rs.Primary.ID)

		if err == nil || !govcd.ContainsNotFound(err) {
			return fmt.Errorf("%s not deleted yet", identifier)
		}

		return nil

	}
}

func importStateIdDefinedInterface(vendor, nss, version string) resource.ImportStateIdFunc {
	return func(*terraform.State) (string, error) {
		return vendor +
			ImportSeparator +
			nss +
			ImportSeparator +
			version, nil
	}
}

// TestAccVcdRdeInterfaceValidation tests the validation rules for the RDE Interface resource
func TestAccVcdRdeInterfaceValidation(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	var params = StringMap{
		"Nss":     "wrong%%%%",
		"Version": "1.0.0",
		"Vendor":  "Vendor_0-9",
		"Name":    t.Name(),
	}
	testParamsNotEmpty(t, params)

	config1 := templateFill(testAccVcdRdeInterfaceWrongFields, params)
	params["FuncName"] = t.Name() + "2"
	params["Nss"] = "Nss_0-9"
	params["Vendor"] = "wrong%%%%"
	config2 := templateFill(testAccVcdRdeInterfaceWrongFields, params)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	debugPrintf("#[DEBUG] CONFIGURATION 1: %s\n", config1)
	debugPrintf("#[DEBUG] CONFIGURATION 2: %s\n", config2)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      config1,
				ExpectError: regexp.MustCompile(".*only alphanumeric characters, underscores and hyphens allowed.*"),
			},
			{
				Config:      config2,
				ExpectError: regexp.MustCompile(".*only alphanumeric characters, underscores and hyphens allowed.*"),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdRdeInterfaceWrongFields = `
# skip-binary-test - This test checks early failure validations
resource "vcd_rde_interface" "rde_interface_validation" {
  nss           = "{{.Nss}}"
  version       = "{{.Version}}"
  vendor        = "{{.Vendor}}"
  name          = "{{.Name}}"
}
`
