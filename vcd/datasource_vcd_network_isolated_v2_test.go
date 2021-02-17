// +build network ALL functional

package vcd

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcdNetworkIsolatedV2NsxvDS(t *testing.T) {
	// This test requires access to the vCD before filling templates
	// Thus it won't run in the short test
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	err := getAvailableNetworks()

	if err != nil {
		fmt.Printf("%s\n", err)
		t.Skip("error getting available networks")
		return
	}
	if len(availableNetworks) == 0 {
		t.Skip("No networks found - data source test skipped")
		return
	}

	networkType := "vcd_network_isolated"
	data, ok := availableNetworks[networkType]
	if !ok {
		t.Skip("no routed network found ")
		return
	}

	// String map to fill the template
	var params = StringMap{
		"Org":         testConfig.VCD.Org,
		"Vdc":         testConfig.VCD.Vdc,
		"NetworkName": data.network.Name,
		"Tags":        "network",
	}

	configText := templateFill(testAccVcdNetworkIsolatedV2NsxvDS, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText)

	params["FuncName"] = t.Name() + "-step2"
	configText2 := templateFill(testAccVcdNetworkIsolatedV2NsxvDSNameFilter, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText2)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		PreCheck:          func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.vcd_network_isolated_v2.ds", "id"),
					resource.TestCheckResourceAttr("data.vcd_network_isolated_v2.ds", "id", data.network.ID),
					resource.TestCheckResourceAttr("data.vcd_network_isolated_v2.ds", "name", data.network.Name),
					resource.TestCheckResourceAttrSet("data.vcd_network_isolated_v2.ds", "gateway"),
					resource.TestCheckResourceAttrSet("data.vcd_network_isolated_v2.ds", "prefix_length"),
				),
			},
			resource.TestStep{
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.vcd_network_isolated_v2.ds", "id"),
					resource.TestCheckResourceAttr("data.vcd_network_isolated_v2.ds", "id", data.network.ID),
					resource.TestCheckResourceAttr("data.vcd_network_isolated_v2.ds", "name", data.network.Name),
					resource.TestCheckResourceAttrSet("data.vcd_network_isolated_v2.ds", "gateway"),
					resource.TestCheckResourceAttrSet("data.vcd_network_isolated_v2.ds", "prefix_length"),
				),
			},
		},
	})

}

const testAccVcdNetworkIsolatedV2NsxvDS = `
data "vcd_network_isolated_v2" "ds" {
  org  = "{{.Org}}"
  vdc  = "{{.Vdc}}"
  name = "{{.NetworkName}}"
}
`

const testAccVcdNetworkIsolatedV2NsxvDSNameFilter = `
data "vcd_network_isolated_v2" "ds" {
  org  = "{{.Org}}"
  vdc  = "{{.Vdc}}"
  filter {
    name_regex = "{{.NetworkName}}"
  }
}
`
