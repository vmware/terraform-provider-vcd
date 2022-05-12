//go:build network || nsxt || ALL || functional
// +build network nsxt ALL functional

package vcd

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcdNetworkRoutedV2NsxtDS(t *testing.T) {
	preTestChecks(t)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	// String map to fill the template
	var params = StringMap{
		"Org":                  testConfig.VCD.Org,
		"NsxtVdc":              testConfig.Nsxt.Vdc,
		"EdgeGw":               testConfig.Nsxt.EdgeGateway,
		"NetworkName":          t.Name(),
		"Tags":                 "network nsxt",
		"MetadataKey":          "key1",
		"MetadataValue":        "value1",
		"MetadataKeyUpdated":   "key2",
		"MetadataValueUpdated": "value2",
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
		PreCheck:          func() { testParamsNotEmpty(t, params) },
		CheckDestroy:      testAccCheckOpenApiVcdNetworkDestroy(testConfig.Nsxt.Vdc, "nsxt-routed-test-initial"),
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_network_routed_v2.net1", "id")),
			},

			{
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_network_routed_v2.net1", "id"),
					// Ensure that all fields are the same except field count '%' (because datasource has `filter` field)
					resourceFieldsEqual("vcd_network_routed_v2.net1", "data.vcd_network_routed_v2.ds", []string{"%"}),
				),
			},
			{
				Config: configText3,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_network_routed_v2.net1", "id"),
					// Ensure that all fields are the same except field count '%' (because datasource has `filter` field)
					resourceFieldsEqual("vcd_network_routed_v2.net1", "data.vcd_network_routed_v2.ds", []string{"%"}),
				),
			},
			{
				Config: configText4,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_network_routed_v2.net1", "id"),
					// Ensure that all fields are the same except field count '%' (because datasource has `filter` field)
					resourceFieldsEqual("vcd_network_routed_v2.net1", "data.vcd_network_routed_v2.ds", []string{"%"}),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdNetworkRoutedV2NsxtDS = TestAccVcdNetworkRoutedV2NsxtStep1 + `
data "vcd_network_routed_v2" "ds" {
  org  = "{{.Org}}"
  vdc  = "{{.NsxtVdc}}"
  name = "nsxt-routed-test-initial"
}
`

const testAccVcdNetworkRoutedV2NsxtDSStep3 = TestAccVcdNetworkRoutedV2NsxtStep1 + `
data "vcd_network_routed_v2" "ds" {
  org  = "{{.Org}}"
  vdc  = "{{.NsxtVdc}}"

  filter {
	name_regex = "^nsxt-routed"
  }
}
`

const testAccVcdNetworkRoutedV2NsxtDSStep4 = TestAccVcdNetworkRoutedV2NsxtStep1 + `
data "vcd_network_routed_v2" "ds" {
  org  = "{{.Org}}"
  vdc  = "{{.NsxtVdc}}"

  filter {
	ip = "1.1.1"
  }
}
`
