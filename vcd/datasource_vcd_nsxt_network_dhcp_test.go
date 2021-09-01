//go:build network || nsxt || ALL || functional
// +build network nsxt ALL functional

package vcd

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcdOpenApiDhcpNsxtRoutedDS(t *testing.T) {
	preTestChecks(t)
	skipNoNsxtConfiguration(t)

	// This test creates a resource and uses datasource which is not possible in single file
	// therefore skipping binary tests
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	// String map to fill the template
	var params = StringMap{
		"Org":         testConfig.VCD.Org,
		"NsxtVdc":     testConfig.Nsxt.Vdc,
		"EdgeGw":      testConfig.Nsxt.EdgeGateway,
		"NetworkName": t.Name(),
		"Tags":        "network nsxt",
	}

	configText := templateFill(testAccRoutedNetDhcpStep1DS, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 0: %s", configText)

	params["FuncName"] = t.Name() + "-step1"
	configText1 := templateFill(testAccRoutedNetDhcpStep2DS, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckOpenApiVcdNetworkDestroy(testConfig.Nsxt.Vdc, "nsxt-routed-dhcp"),
		Steps: []resource.TestStep{
			resource.TestStep{ // Define network and DHCP pools
				Config: configText,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_network_dhcp.pools", "id", regexp.MustCompile(`^urn:vcloud:network:.*$`)),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_network_dhcp.pools", "pool.*", map[string]string{
						"start_address": "7.1.1.100",
						"end_address":   "7.1.1.110",
					}),
				),
			},
			resource.TestStep{
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_network_dhcp.pools", "id", regexp.MustCompile(`^urn:vcloud:network:.*$`)),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_network_dhcp.pools", "pool.*", map[string]string{
						"start_address": "7.1.1.100",
						"end_address":   "7.1.1.110",
					}),
					resourceFieldsEqual("vcd_nsxt_network_dhcp.pools", "data.vcd_nsxt_network_dhcp.pools", []string{}),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccRoutedNetDhcpStep1DS = testAccRoutedNetDhcpConfig + `
resource "vcd_nsxt_network_dhcp" "pools" {
  org  = "{{.Org}}"
  vdc  = "{{.NsxtVdc}}"

  org_network_id = vcd_network_routed_v2.net1.id
  
  pool {
    start_address = "7.1.1.100"
    end_address   = "7.1.1.110"
  }
}
`

const testAccRoutedNetDhcpStep2DS = testAccRoutedNetDhcpStep1DS + `
data "vcd_nsxt_network_dhcp" "pools" {
  org  = "{{.Org}}"
  vdc  = "{{.NsxtVdc}}"

  org_network_id = vcd_network_routed_v2.net1.id

  depends_on = [vcd_network_routed_v2.net1]
}
`
