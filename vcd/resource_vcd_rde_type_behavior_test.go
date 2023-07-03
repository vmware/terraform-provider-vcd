//go:build rde || ALL || functional

package vcd

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcdRdeTypeBehavior(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	var params = StringMap{
		"Nss":                     "nss1",
		"Version":                 "1.0.0",
		"Vendor":                  "vendor1",
		"InterfaceName":           t.Name(),
		"SchemaPath":              getCurrentDir() + "/../test-resources/rde_type.json",
		"BehaviorName":            t.Name(),
		"BehaviorDescription":     t.Name(),
		"ExecutionId":             "MyActivity",
		"ExecutionType":           "Activity",
		"TypeBehaviorDescription": t.Name(),
		"TypeExecutionId":         "MyActivityOverride",
		"TypeExecutionType":       "noop",
		"AccessLevelIds":          " ",
	}
	testParamsNotEmpty(t, params)

	configText1 := templateFill(testAccVcdRdeTypeBehavior, params)
	debugPrintf("#[DEBUG] CONFIGURATION 1: %s\n", configText1)

	params["FuncName"] = t.Name() + "-Step2"
	params["TypeBehaviorDescription"] = t.Name() + "Updated"
	params["TypeExecutionId"] = "MyActivityOverrideUpdated"
	params["AccessLevelIds"] = "\"urn:vcloud:accessLevel:ReadOnly\""
	configText2 := templateFill(testAccVcdRdeTypeBehavior, params)
	debugPrintf("#[DEBUG] CONFIGURATION 2: %s\n", configText2)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	interfaceBehavior := "vcd_rde_interface_behavior.behavior1"
	rdeTypeBehavior := "vcd_rde_type_behavior.behavior2"
	rdeType := "vcd_rde_type.type1"
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckRdeTypesDestroy(rdeType), // If the RDE Type is destroyed, the Behavior is also destroyed.
		Steps: []resource.TestStep{
			{
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(rdeTypeBehavior, "ref", interfaceBehavior, "id"),
					resource.TestCheckResourceAttr(rdeTypeBehavior, "name", params["BehaviorName"].(string)),
					resource.TestCheckResourceAttr(rdeTypeBehavior, "description", t.Name()+"Override"),
					resource.TestCheckResourceAttr(rdeTypeBehavior, "execution.id", "MyActivityOverride"),
					resource.TestCheckResourceAttr(rdeTypeBehavior, "execution.type", "noop"),
					resource.TestCheckResourceAttr(rdeTypeBehavior, "access_level_ids.#", "0"),
				),
			},
			{
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(rdeTypeBehavior, "ref", interfaceBehavior, "id"),
					resource.TestCheckResourceAttr(rdeTypeBehavior, "description", t.Name()+"UpdatedOverride"),
					resource.TestCheckResourceAttr(rdeTypeBehavior, "execution.id", "MyActivityOverrideUpdated"),
					resource.TestCheckResourceAttr(rdeTypeBehavior, "execution.type", "noop"),
					resource.TestCheckResourceAttr(rdeTypeBehavior, "access_level_ids.#", "1"),
					resource.TestCheckTypeSetElemAttr(rdeTypeBehavior, "access_level_ids.*", "urn:vcloud:accessLevel:ReadOnly"),
				),
			},
			{
				ResourceName:      rdeTypeBehavior,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdInterfaceBehavior(params["Vendor"].(string), params["Nss"].(string), params["Version"].(string), params["BehaviorName"].(string)),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdRdeTypeBehavior = `
resource "vcd_rde_interface" "interface1" {
  nss     = "{{.Nss}}"
  version = "{{.Version}}"
  vendor  = "{{.Vendor}}"
  name    = "{{.InterfaceName}}"
}

resource "vcd_rde_interface_behavior" "behavior1" {
  rde_interface_id = vcd_rde_interface.interface1.id
  name             = "{{.BehaviorName}}"
  description      = "{{.BehaviorDescription}}"
  execution = {
    "id":   "{{.ExecutionId}}"
    "type": "{{.ExecutionType}}"
  }
}

resource "vcd_rde_type" "type1" {
  nss           = "{{.Nss}}"
  version       = "{{.Version}}"
  vendor        = "{{.Vendor}}"
  name          = "{{.Name}}"
  description   = "{{.Description}}"
  interface_ids = [ vcd_rde_interface.interface1.id ]
  schema        = file("{{.SchemaPath}}")

  # Behaviors can't be created after the RDE Interface is used by a RDE Type
  # so we need to depend on the Behavior to wait for it to be created first.
  depends_on = [ vcd_rde_interface_behavior.behavior1 ]
}

resource "vcd_rde_type_behavior" "behavior2" {
  rde_type_id               = vcd_rde_type.type1.id
  rde_interface_behavior_id = vcd_rde_interface_behavior.behavior1.id
  description               = "{{.TypeBehaviorDescription}}Override"
  execution = {
    "id":   "{{.TypeExecutionId}}"
    "type": "{{.TypeExecutionType}}"
  }

  access_level_ids = [ {{.AccessLevelIds}} ]
}
`
