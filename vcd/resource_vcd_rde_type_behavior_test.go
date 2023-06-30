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
		"Nss":                 "nss1",
		"Version":             "1.0.0",
		"Vendor":              "vendor1",
		"InterfaceName":       t.Name(),
		"BehaviorName":        t.Name(),
		"BehaviorDescription": t.Name(),
		"ExecutionId":         "MyActivity",
		"ExecutionType":       "Activity",
	}
	testParamsNotEmpty(t, params)

	configText1 := templateFill(testAccVcdRdeTypeBehavior, params)
	debugPrintf("#[DEBUG] CONFIGURATION 1: %s\n", configText1)
	params["FuncName"] = t.Name() + "-Step2"
	params["BehaviorDescription"] = t.Name() + "Updated"
	params["ExecutionId"] = "MyActivityUpdated"
	configText2 := templateFill(testAccVcdRdeTypeBehavior, params)
	debugPrintf("#[DEBUG] CONFIGURATION 2: %s\n", configText2)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	interfaceName := "vcd_rde_interface.interface1"
	behaviorName := "vcd_rde_interface_behavior.behavior1"
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckRdeTypesDestroy(interfaceName), // If the RDE Type is destroyed, the Behavior is also destroyed.
		Steps: []resource.TestStep{
			{
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(interfaceName, "id", behaviorName, "rde_interface_id"),
					resource.TestCheckResourceAttr(behaviorName, "name", params["BehaviorName"].(string)),
					resource.TestCheckResourceAttr(behaviorName, "description", t.Name()),
					resource.TestCheckResourceAttr(behaviorName, "execution.id", "MyActivity"),
					resource.TestCheckResourceAttr(behaviorName, "execution.type", "Activity"),
					resource.TestCheckResourceAttrPair(behaviorName, "id", behaviorName, "ref"),
				),
			},
			{
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(interfaceName, "id", behaviorName, "rde_interface_id"),
					resource.TestCheckResourceAttr(behaviorName, "name", params["BehaviorName"].(string)),
					resource.TestCheckResourceAttr(behaviorName, "description", t.Name()+"Updated"),
					resource.TestCheckResourceAttr(behaviorName, "execution.id", "MyActivityUpdated"),
					resource.TestCheckResourceAttr(behaviorName, "execution.type", "Activity"),
					resource.TestCheckResourceAttrPair(behaviorName, "id", behaviorName, "ref"),
				),
			},
			{
				ResourceName:      behaviorName,
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
  interface_ids = [{{.InterfaceReferences}}]
  schema        = file("{{.SchemaPath}}")

  # Behaviors can't be created after the RDE Interface is used by a RDE Type
  # so we need to depend on the Behavior to wait for it to be created first.
  depends_on = [ vcd_rde_interface_behavior.behavior1 ]
}

resource "vcd_rde_type_behavior" "behavior2" {
  rde_type_id               = vcd_rde_type.type1.id
  rde_interface_behavior_id = vcd_rde_interface_behavior.behavior1.id
  description               = "{{.BehaviorDescription}}"
  execution = {
    "id":   "{{.ExecutionId}}"
    "type": "{{.ExecutionType}}"
  }
}
`
