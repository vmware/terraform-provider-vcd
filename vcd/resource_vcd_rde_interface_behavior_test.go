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
		"BehaviorName1":       t.Name(),
		"BehaviorName2":       t.Name() + "json",
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
	behaviorName2 := "vcd_rde_interface_behavior.behavior2"
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckRdeInterfaceDestroy(interfaceName), // If the RDE Interface is destroyed, the Behavior is also destroyed.
		Steps: []resource.TestStep{
			{
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(interfaceName, "id", behaviorName, "rde_interface_id"),
					resource.TestCheckResourceAttr(behaviorName, "name", params["BehaviorName1"].(string)),
					resource.TestCheckResourceAttr(behaviorName, "description", t.Name()),
					resource.TestCheckResourceAttr(behaviorName, "execution.id", "MyActivity"),
					resource.TestCheckResourceAttr(behaviorName, "execution.type", "Activity"),
					resource.TestCheckResourceAttrPair(behaviorName, "id", behaviorName, "ref"),
					// Compare JSON and map values of executions
					resource.TestCheckResourceAttrPair(behaviorName, "execution.id", behaviorName2, "execution.id"),
					resource.TestCheckResourceAttrPair(behaviorName, "execution.type", behaviorName2, "execution.type"),
					resource.TestCheckResourceAttrPair(behaviorName, "execution_json", behaviorName2, "execution_json"),
				),
			},
			{
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(interfaceName, "id", behaviorName, "rde_interface_id"),
					resource.TestCheckResourceAttr(behaviorName, "name", params["BehaviorName1"].(string)),
					resource.TestCheckResourceAttr(behaviorName, "description", t.Name()+"Updated"),
					resource.TestCheckResourceAttr(behaviorName, "execution.id", "MyActivityUpdated"),
					resource.TestCheckResourceAttr(behaviorName, "execution.type", "Activity"),
					resource.TestCheckResourceAttrPair(behaviorName, "id", behaviorName, "ref"),
					// Compare JSON and map values of executions
					resource.TestCheckResourceAttrPair(behaviorName, "execution.id", behaviorName2, "execution.id"),
					resource.TestCheckResourceAttrPair(behaviorName, "execution.type", behaviorName2, "execution.type"),
					resource.TestCheckResourceAttrPair(behaviorName, "execution_json", behaviorName2, "execution_json"),
				),
			},
			{
				ResourceName:      behaviorName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdInterfaceBehavior(params["Vendor"].(string), params["Nss"].(string), params["Version"].(string), params["BehaviorName1"].(string)),
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
  name             = "{{.BehaviorName1}}"
  description      = "{{.BehaviorDescription}}"
  execution = {
    "id":   "{{.ExecutionId}}"
    "type": "{{.ExecutionType}}"
  }
}

resource "vcd_rde_interface_behavior" "behavior2" {
  rde_interface_id = vcd_rde_interface.interface1.id
  name             = "{{.BehaviorName2}}"
  description      = "{{.BehaviorDescription}}"
  execution_json   = jsonencode({
    "id":   "{{.ExecutionId}}"
    "type": "{{.ExecutionType}}"
  })
}
`

func TestAccVcdRdeInterfaceBehaviorComplexExecution(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	var params = StringMap{
		"Nss":                 "nss1",
		"Version":             "1.0.0",
		"Vendor":              "vendor1",
		"InterfaceName":       t.Name(),
		"BehaviorName":        t.Name(),
		"BehaviorDescription": t.Name(),
		"ExecutionId":         "MyWebhook",
	}
	testParamsNotEmpty(t, params)

	configText1 := templateFill(testAccVcdRdeInterfaceBehaviorComplexExecution, params)
	debugPrintf("#[DEBUG] CONFIGURATION 1: %s\n", configText1)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	interfaceName := "vcd_rde_interface.interface1"
	behaviorName := "vcd_rde_interface_behavior.behavior"
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
					resource.TestCheckResourceAttrPair(behaviorName, "id", behaviorName, "ref"),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdRdeInterfaceBehaviorComplexExecution = `
resource "vcd_rde_interface" "interface1" {
  nss     = "{{.Nss}}"
  version = "{{.Version}}"
  vendor  = "{{.Vendor}}"
  name    = "{{.InterfaceName}}"
}

resource "vcd_rde_interface_behavior" "behavior" {
  rde_interface_id = vcd_rde_interface.interface1.id
  name             = "{{.BehaviorName}}"
  description      = "{{.BehaviorDescription}}"
  execution_json   = jsonencode({
    "id": "{{.ExecutionId}}"
	"type": "WebHook",
  	"href": "https://hooks.slack.com:443/services/T07UZFN0N/B01EW5NC42D/rfjhHCGIwzuzQFrpPZiuLkIX",
  	"_internal_key": "secretKey",
	"execution_properties": {
	  "template": {
	    "content": "test"
       },
       "_secure_token": "secureToken",
       "invocation_timeout": 7
    }
  })
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
