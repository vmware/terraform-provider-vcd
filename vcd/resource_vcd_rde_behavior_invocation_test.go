//go:build rde || ALL || functional

package vcd

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcdRdeBehaviorInvocation(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	var params = StringMap{
		"Nss":                   "nss1",
		"Version":               "1.0.0",
		"Vendor":                "vendor1",
		"Name":                  t.Name(),
		"Description":           t.Name(),
		"SchemaPath":            getCurrentDir() + "/../test-resources/rde_type.json",
		"ExecutionId":           "MyActivity",
		"ExecutionType":         "noop",
		"TypeAccessLevels":      "\"urn:vcloud:accessLevel:FullControl\"",
		"InterfaceAccessLevels": "\"urn:vcloud:accessLevel:ReadOnly\", \"urn:vcloud:accessLevel:FullControl\"",
	}
	testParamsNotEmpty(t, params)

	configText1 := templateFill(testAccVcdRdeBehaviorInvocation, params)
	debugPrintf("#[DEBUG] CONFIGURATION 1: %s\n", configText1)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	interfaceBehavior1 := "vcd_rde_interface_behavior.behavior1"
	rdeTypeBehavior := "vcd_rde_type_behavior.behavior_override"
	rdeType := "vcd_rde_type.type"
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckRdeTypesDestroy(rdeType), // If the RDE Type is destroyed, the Behavior is also destroyed.
		Steps: []resource.TestStep{
			{
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					// RDE Type Behavior
					resource.TestCheckResourceAttrPair(rdeTypeBehavior, "ref", interfaceBehavior1, "id"),
					resource.TestCheckResourceAttr(rdeTypeBehavior, "name", t.Name()+"1"),
					resource.TestCheckResourceAttr(rdeTypeBehavior, "description", t.Name()+"Override"),
					resource.TestCheckResourceAttr(rdeTypeBehavior, "execution.id", "MyActivityOverride"),
					resource.TestCheckResourceAttr(rdeTypeBehavior, "execution.type", "noop"),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdRdeBehaviorInvocation = `
resource "vcd_rde_interface" "interface" {
  nss     = "{{.Nss}}"
  version = "{{.Version}}"
  vendor  = "{{.Vendor}}"
  name    = "{{.Name}}"
}

resource "vcd_rde_interface_behavior" "behavior1" {
  rde_interface_id = vcd_rde_interface.interface.id
  name             = "{{.Name}}1"
  description      = "{{.Description}}"
  execution = {
    "id" : "{{.ExecutionId}}1"
    "type" : "{{.ExecutionType}}"
  }
}

resource "vcd_rde_type" "type" {
  nss           = "{{.Nss}}"
  version       = "{{.Version}}"
  vendor        = "{{.Vendor}}"
  name          = "{{.Name}}"
  description   = "{{.Description}}"
  interface_ids = [vcd_rde_interface.interface.id]
  schema        = file("{{.SchemaPath}}")

  # Behaviors can't be created after the RDE Interface is used by a RDE Type
  # so we need to depend on the Behaviors to wait for them to be created first.
  depends_on = [vcd_rde_interface_behavior.behavior1, vcd_rde_interface_behavior.behavior2]
}


`
