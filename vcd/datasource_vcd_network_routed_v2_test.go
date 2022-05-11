//go:build network || ALL || functional
// +build network ALL functional

package vcd

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcdNetworkRoutedV2NsxvDS(t *testing.T) {
	preTestChecks(t)
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

	networkType := "vcd_network_routed"
	data, ok := availableNetworks[networkType]
	if !ok {
		t.Skip("no routed network found")
		return
	}

	// String map to fill the template
	var params = StringMap{
		"Org":         testConfig.VCD.Org,
		"Vdc":         testConfig.VCD.Vdc,
		"EdgeGw":      testConfig.Nsxt.EdgeGateway,
		"NetworkName": data.network.Name,
		"Tags":        "network",
	}

	configText := templateFill(testAccVcdNetworkRoutedV2NsxvDS, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText)

	params["FuncName"] = t.Name() + "-step2"
	configText2 := templateFill(testAccVcdNetworkRoutedV2NsxvDSNameFilter, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText2)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		PreCheck:          func() { testParamsNotEmpty(t, params) },
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.vcd_network_routed_v2.ds", "id"),
					resource.TestCheckResourceAttr("data.vcd_network_routed_v2.ds", "id", data.network.ID),
					resource.TestCheckResourceAttr("data.vcd_network_routed_v2.ds", "name", data.network.Name)),
			},
			{
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.vcd_network_routed_v2.ds", "id"),
					resource.TestCheckResourceAttr("data.vcd_network_routed_v2.ds", "id", data.network.ID),
					resource.TestCheckResourceAttr("data.vcd_network_routed_v2.ds", "name", data.network.Name)),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdNetworkRoutedV2NsxvDS = `
data "vcd_network_routed_v2" "ds" {
  org  = "{{.Org}}"
  vdc  = "{{.Vdc}}"
  name = "{{.NetworkName}}"
}
`

const testAccVcdNetworkRoutedV2NsxvDSNameFilter = `
data "vcd_network_routed_v2" "ds" {
  org  = "{{.Org}}"
  vdc  = "{{.Vdc}}"
  filter {
    name_regex = "{{.NetworkName}}"
  }
}
`
