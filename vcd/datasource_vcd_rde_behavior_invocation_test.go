//go:build rde || ALL || functional

package vcd

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcdRdeBehaviorInvocation(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	var params = StringMap{
		"Nss":           "nss",
		"Version":       "1.0.0",
		"Vendor":        "vendor",
		"Name":          t.Name(),
		"Description":   t.Name(),
		"SchemaPath":    getCurrentDir() + "/../test-resources/rde_type.json",
		"EntityPath":    getCurrentDir() + "/../test-resources/rde_instance.json",
		"ExecutionId":   "MyActivity",
		"ExecutionType": "noop",
		"AccessLevels":  "\"urn:vcloud:accessLevel:FullControl\"",
	}
	testParamsNotEmpty(t, params)

	configText1 := templateFill(testAccVcdRdeBehaviorInvocation, params)
	debugPrintf("#[DEBUG] CONFIGURATION 1: %s\n", configText1)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckRdeInterfaceDestroy("vcd_rde_interface.interface"), // If the interface is destroyed, everything is
		Steps: []resource.TestStep{
			{
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("vcd_rde.rde", "name", t.Name()),
					// No-op operations return the original entity data, hence the RDE Type should appear:
					resource.TestMatchResourceAttr("data.vcd_rde_behavior_invocation.invoke", "result",
						regexp.MustCompile(fmt.Sprintf("\"urn:vcloud:type:%s:%s:%s\"", params["Vendor"], params["Nss"], params["Version"]))),
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

resource "vcd_rde_interface_behavior" "behavior" {
  rde_interface_id = vcd_rde_interface.interface.id
  name             = "{{.Name}}"
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
  depends_on = [vcd_rde_interface_behavior.behavior]
}

resource "vcd_rde" "rde" {
  org          = "System" # We use System org to avoid using right bundles
  rde_type_id  = vcd_rde_type.type.id
  name         = "{{.Name}}"
  resolve      = true
  input_entity = file("{{.EntityPath}}")
}

# Required Access Levels to invoke Behaviors
resource "vcd_rde_type_behavior_acl" "interface_acl" {
  rde_type_id = vcd_rde_type.type.id
  behavior_id = vcd_rde_interface_behavior.behavior.id
  access_level_ids = [{{.AccessLevels}}]
}

data "vcd_rde_behavior_invocation" "invoke" {
  rde_id                  = vcd_rde.rde.id
  behavior_id             = vcd_rde_interface_behavior.behavior.id
}
`
