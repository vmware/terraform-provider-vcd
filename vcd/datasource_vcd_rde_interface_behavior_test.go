//go:build rde || ALL || functional

package vcd

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcdRdeInterfaceBehaviorDS(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	var params = StringMap{
		// This is a Defined Interface that comes with VCD out of the box
		"InterfaceNss":     "k8s",
		"InterfaceVersion": "1.0.0",
		"InterfaceVendor":  "vmware",
		"BehaviorName":     "createKubeConfig", // This Behavior is also included
	}
	testParamsNotEmpty(t, params)

	configText := templateFill(testAccVcdRdeInterfaceBehaviorDS, params)
	debugPrintf("#[DEBUG] CONFIGURATION data source: %s\n", configText)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	rdeInterfaceBehavior := "data.vcd_rde_interface_behavior.behavior_ds"
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(rdeInterfaceBehavior, "id", rdeInterfaceBehavior, "ref"),
					resource.TestCheckResourceAttr(rdeInterfaceBehavior, "name", params["BehaviorName"].(string)),
					resource.TestCheckResourceAttr(rdeInterfaceBehavior, "description", "Creates and returns a kubeconfig"),
					resource.TestCheckResourceAttr(rdeInterfaceBehavior, "ref", fmt.Sprintf("urn:vcloud:behavior-interface:%s:%s:%s:%s", params["BehaviorName"].(string), params["InterfaceVendor"].(string), params["InterfaceNss"].(string), params["InterfaceVersion"].(string))),
					resource.TestCheckResourceAttr(rdeInterfaceBehavior, "execution.id", "CreateKubeConfigActivity"),
					resource.TestCheckResourceAttr(rdeInterfaceBehavior, "execution.type", "Activity"),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdRdeInterfaceBehaviorDS = `
data "vcd_rde_interface" "interface_ds" {
  nss     = "{{.InterfaceNss}}"
  version = "{{.InterfaceVersion}}"
  vendor  = "{{.InterfaceVendor}}"
}

data "vcd_rde_interface_behavior" "behavior_ds" {
  rde_interface_id = data.vcd_rde_interface.interface_ds.id
  name             = "{{.BehaviorName}}"
}
`
