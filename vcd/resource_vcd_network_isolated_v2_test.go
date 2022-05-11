//go:build network || ALL || functional
// +build network ALL functional

package vcd

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcdNetworkIsolatedV2Nsxv(t *testing.T) {
	preTestChecks(t)
	var params = StringMap{
		"Org":         testConfig.VCD.Org,
		"Vdc":         testConfig.VCD.Vdc,
		"NetworkName": t.Name(),
		"Tags":        "network",
	}

	params["FuncName"] = t.Name()
	configText := templateFill(testAccVcdNetworkIsolatedV2Nsxv, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 0: %s", configText)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		PreCheck:          func() { testParamsNotEmpty(t, params) },
		CheckDestroy:      testAccCheckOpenApiVcdNetworkDestroy(testConfig.VCD.Vdc, t.Name()),
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_network_isolated_v2.net1", "id"),
					resource.TestCheckResourceAttr("vcd_network_isolated_v2.net1", "name", t.Name()),
					resource.TestCheckResourceAttr("vcd_network_isolated_v2.net1", "description", "NSX-V isolated network test OpenAPI"),
					resource.TestCheckResourceAttr("vcd_network_isolated_v2.net1", "gateway", "1.1.1.1"),
					resource.TestCheckResourceAttr("vcd_network_isolated_v2.net1", "prefix_length", "24"),
					resource.TestCheckResourceAttr("vcd_network_isolated_v2.net1", "static_ip_pool.#", "1"),
					resource.TestCheckResourceAttr("vcd_network_isolated_v2.net1", "is_shared", "true"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_network_isolated_v2.net1", "static_ip_pool.*", map[string]string{
						"start_address": "1.1.1.10",
						"end_address":   "1.1.1.20",
					}),
				),
			},
			// Check that import works
			{
				ResourceName:      "vcd_network_isolated_v2.net1",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdOrgVdcObject(testConfig, t.Name()),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdNetworkIsolatedV2Nsxv = `
resource "vcd_network_isolated_v2" "net1" {
  org  = "{{.Org}}"
  vdc  = "{{.Vdc}}"
  name = "{{.NetworkName}}"
  
  description = "NSX-V isolated network test OpenAPI" 
  is_shared   = true

  gateway       = "1.1.1.1"
  prefix_length = 24

  static_ip_pool {
	start_address = "1.1.1.10"
	end_address   = "1.1.1.20"
  }
}
`
