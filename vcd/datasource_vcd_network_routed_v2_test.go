// +build network ALL functional

package vcd

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcdNetworkRoutedV2NsxtDS(t *testing.T) {
	vcdClient := createTemporaryVCDConnection()
	if vcdClient.Client.APIVCDMaxVersionIs("< 34.0") {
		t.Skip(t.Name() + " requires at least API v34.0 (vCD 10.1.1+)")
	}
	skipNoNsxtConfiguration(t)

	// String map to fill the template
	var params = StringMap{
		"Org":         testConfig.VCD.Org,
		"Vdc":         testConfig.Nsxt.Vdc,
		"EdgeGw":      testConfig.Nsxt.EdgeGateway,
		"NetworkName": t.Name(),
	}

	params["FuncName"] = t.Name() + "-DS"
	configText := templateFill(TestAccVcdNetworkRoutedV2NsxtStep1, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText)

	params["FuncName"] = t.Name() + "-DS-step2"
	configText2 := templateFill(testAccVcdNetworkRoutedV2NsxtDS, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText)

	params["FuncName"] = t.Name() + "-DS-step3"
	configText3 := templateFill(testAccVcdNetworkRoutedV2NsxtDSStep3, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 3: %s", configText3)

	params["FuncName"] = t.Name() + "-DS-step4"
	configText4 := templateFill(testAccVcdNetworkRoutedV2NsxtDSStep4, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 4: %s", configText4)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckOpenApiVcdNetworkDestroy(testConfig.Nsxt.Vdc, "nsxt-routed-test-initial"),
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_network_routed_v2.net1", "id")),
			},

			resource.TestStep{
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_network_routed_v2.net1", "id"),
					// Ensure that all fields are the same except field count '%' (because datasource has `filter` field)
					resourceFieldsEqual("vcd_network_routed_v2.net1", "data.vcd_network_routed_v2.ds", []string{"%"}),
				),
			},
			resource.TestStep{
				Config: configText3,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_network_routed_v2.net1", "id"),
					// Ensure that all fields are the same except field count '%' (because datasource has `filter` field)
					resourceFieldsEqual("vcd_network_routed_v2.net1", "data.vcd_network_routed_v2.ds", []string{"%"}),
				),
			},
			resource.TestStep{
				Config: configText4,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_network_routed_v2.net1", "id"),
					// Ensure that all fields are the same except field count '%' (because datasource has `filter` field)
					resourceFieldsEqual("vcd_network_routed_v2.net1", "data.vcd_network_routed_v2.ds", []string{"%"}),
				),
			},
		},
	})
}

const testAccVcdNetworkRoutedV2NsxtDS = TestAccVcdNetworkRoutedV2NsxtStep1 + `
data "vcd_network_routed_v2" "ds" {
  org  = "{{.Org}}"
  vdc  = "{{.Vdc}}"
  name = "nsxt-routed-test-initial"
}
`

const testAccVcdNetworkRoutedV2NsxtDSStep3 = TestAccVcdNetworkRoutedV2NsxtStep1 + `
data "vcd_network_routed_v2" "ds" {
  org  = "{{.Org}}"
  vdc  = "{{.Vdc}}"
  
  filter {
	name_regex = "^nsxt-routed"
  }
}
`

const testAccVcdNetworkRoutedV2NsxtDSStep4 = TestAccVcdNetworkRoutedV2NsxtStep1 + `
data "vcd_network_routed_v2" "ds" {
  org  = "{{.Org}}"
  vdc  = "{{.Vdc}}"
  
  filter {
	ip = "1.1.1"
  }
}
`

func TestAccVcdNetworkRoutedV2NsxvDS(t *testing.T) {
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
		t.Skip("no routed network found ")
		return
	}

	// String map to fill the template
	var params = StringMap{
		"Org":         testConfig.VCD.Org,
		"Vdc":         testConfig.VCD.Vdc,
		"EdgeGw":      testConfig.Nsxt.EdgeGateway,
		"NetworkName": data.network.Name,
	}

	configText := templateFill(testAccVcdNetworkRoutedV2NsxvDS, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText)

	params["FuncName"] = t.Name() + "-step2"
	configText2 := templateFill(testAccVcdNetworkRoutedV2NsxvDSNameFilter, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText2)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		PreCheck:          func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.vcd_network_routed_v2.ds", "id"),
					resource.TestCheckResourceAttr("data.vcd_network_routed_v2.ds", "id", data.network.ID),
					resource.TestCheckResourceAttr("data.vcd_network_routed_v2.ds", "name", data.network.Name)),
			},
			resource.TestStep{
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.vcd_network_routed_v2.ds", "id"),
					resource.TestCheckResourceAttr("data.vcd_network_routed_v2.ds", "id", data.network.ID),
					resource.TestCheckResourceAttr("data.vcd_network_routed_v2.ds", "name", data.network.Name)),
			},
		},
	})

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
