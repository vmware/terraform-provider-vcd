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
		"Nss":           "nss1",
		"Version":       "1.0.0",
		"Vendor":        "vendor1",
		"Name":          t.Name(),
		"Description":   t.Name(),
		"SchemaPath":    getCurrentDir() + "/../test-resources/rde_type.json",
		"ExecutionId":   "MyActivity",
		"ExecutionType": "noop",
	}
	testParamsNotEmpty(t, params)

	configText1 := templateFill(testAccVcdRdeTypeBehavior, params)
	debugPrintf("#[DEBUG] CONFIGURATION 1: %s\n", configText1)

	params["FuncName"] = t.Name() + "-Step2"
	params["Description"] = t.Name() + "Updated"
	params["ExecutionId"] = "MyActivityUpdated"
	configText2 := templateFill(testAccVcdRdeTypeBehavior, params)
	debugPrintf("#[DEBUG] CONFIGURATION 2: %s\n", configText2)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	interfaceBehavior := "vcd_rde_interface_behavior.behavior"
	rdeTypeBehavior := "vcd_rde_type_behavior.behavior_override"
	rdeType := "vcd_rde_type.type"
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckRdeTypesDestroy(rdeType), // If the RDE Type is destroyed, the Behavior is also destroyed.
		Steps: []resource.TestStep{
			{
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(rdeTypeBehavior, "ref", interfaceBehavior, "id"),
					resource.TestCheckResourceAttr(rdeTypeBehavior, "name", t.Name()),
					resource.TestCheckResourceAttr(rdeTypeBehavior, "description", t.Name()+"Override"),
					resource.TestCheckResourceAttr(rdeTypeBehavior, "execution.id", "MyActivityOverride"),
					resource.TestCheckResourceAttr(rdeTypeBehavior, "execution.type", "noop"),
				),
			},
			{
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Behavior with execution override
					resource.TestCheckResourceAttrPair(rdeTypeBehavior, "ref", interfaceBehavior, "id"),
					resource.TestCheckResourceAttr(rdeTypeBehavior, "description", t.Name()+"UpdatedOverride"),
					resource.TestCheckResourceAttr(rdeTypeBehavior, "execution.id", "MyActivityUpdatedOverride"),
					resource.TestCheckResourceAttr(rdeTypeBehavior, "execution.type", "noop"),
				),
			},
			{
				ResourceName:      rdeTypeBehavior,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdInterfaceBehavior(params["Vendor"].(string), params["Nss"].(string), params["Version"].(string), params["Name"].(string)),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdRdeTypeBehavior = `
resource "vcd_rde_interface" "interface" {
  nss     = "{{.Nss}}"
  version = "{{.Version}}"
  vendor  = "{{.Vendor}}"
  name    = "{{.Name}}"
}

resource "vcd_rde_interface_behavior" "behavior" {
  rde_interface_id = vcd_rde_interface.interface.id
  name             = "{{.Name}}"
  description      = "{{.Description}}"
  execution = {
    "id" : "{{.ExecutionId}}"
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
  # so we need to depend on the Behavior to wait for it to be created first.
  depends_on = [vcd_rde_interface_behavior.behavior]
}

resource "vcd_rde_type_behavior" "behavior_override" {
  rde_type_id               = vcd_rde_type.type.id
  rde_interface_behavior_id = vcd_rde_interface_behavior.behavior.id
  description               = "{{.Description}}Override"
  execution = {
    "id" : "{{.ExecutionId}}Override"
    "type" : "{{.ExecutionType}}"
  }
}
`
