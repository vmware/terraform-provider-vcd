// +build network ALL functional

package vcd

import (
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

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		PreCheck:          func() { testAccPreCheck(t) },
		// CheckDestroy:      testAccCheckVcdLbVirtualServerDestroy(params["VirtualServerName"].(string)),
		Steps: []resource.TestStep{
			resource.TestStep{ // step 1 - create resources
				Config: configText,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_network_routed_v2.net1", "id")),
			},

			resource.TestStep{ // step 1 - use datasource to read the resource
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_network_routed_v2.net1", "id"),
					// Ensure that all fields are the same
					resourceFieldsEqual("vcd_network_routed_v2.net1", "data.vcd_network_routed_v2.ds", []string{}),
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
