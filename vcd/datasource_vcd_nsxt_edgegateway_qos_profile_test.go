//go:build ALL || nsxt || gateway || functional

package vcd

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcdDatasourceNsxtGatewayQosProfile(t *testing.T) {
	skipIfNotSysAdmin(t)

	vcdClient := createTemporaryVCDConnection(true)
	if vcdClient == nil {
		t.Skip(acceptanceTestsSkipped)
	}
	if vcdClient.Client.APIVCDMaxVersionIs("< 36.2") {
		t.Skipf("This test tests VCD 10.3.2+ (API V36.2+) features. Skipping.")
	}

	var params = StringMap{
		"FuncName":           t.Name(),
		"NsxtManager":        testConfig.Nsxt.Manager,
		"NsxtQosProfileName": testConfig.Nsxt.GatewayQosProfile,

		"Tags": "nsxt gateway",
	}
	testParamsNotEmpty(t, params)

	configText := templateFill(testAccVcdDatasourceNsxtGatewayQosProfile, params)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					// ID must match URN 'urn:vcloud:nsxtmanager:09722307-aee0-4623-af95-7f8e577c9ebc'
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_manager.nsxt", "id"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_edgegateway_qos_profile.qos-1", "id"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_edgegateway_qos_profile.qos-1", "committed_bandwidth"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_edgegateway_qos_profile.qos-1", "burst_size"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_edgegateway_qos_profile.qos-1", "excess_action"),
				),
			},
		},
	})
}

const testAccVcdDatasourceNsxtGatewayQosProfile = `
data "vcd_nsxt_manager" "nsxt" {
  name = "{{.NsxtManager}}"
}

data "vcd_nsxt_edgegateway_qos_profile" "qos-1" {
  nsxt_manager_id = data.vcd_nsxt_manager.nsxt.id
  name            = "{{.NsxtQosProfileName}}"
}
`
