//go:build rde || ALL || functional

package vcd

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccVcdRdeInterfaceBehavior(t *testing.T) {
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

	configText1 := templateFill(testAccVcdRdeInterfaceBehavior, params)
	debugPrintf("#[DEBUG] CONFIGURATION 1: %s\n", configText1)
	params["FuncName"] = t.Name() + "-Step2"
	params["BehaviorDescription"] = t.Name() + "Updated"
	params["ExecutionId"] = "MyActivityUpdated"
	configText2 := templateFill(testAccVcdRdeInterfaceBehavior, params)
	debugPrintf("#[DEBUG] CONFIGURATION 2: %s\n", configText2)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	interfaceName := "vcd_rde_interface.interface1"
	behaviorName := "vcd_rde_interface_behavior.behavior1"
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckRdeInterfaceDestroy(interfaceName), // If the RDE Interface is destroyed, the Behavior is also destroyed.
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

const testAccVcdRdeInterfaceBehavior = `
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
`

func importStateIdInterfaceBehavior(vendor, nss, version, name string) resource.ImportStateIdFunc {
	return func(*terraform.State) (string, error) {
		return vendor +
			ImportSeparator +
			nss +
			ImportSeparator +
			version +
			ImportSeparator +
			name, nil
	}
}
