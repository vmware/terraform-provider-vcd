//go:build network || nsxt || ALL || functional
// +build network nsxt ALL functional

package vcd

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcdOpenApiDhcpNsxtRouted(t *testing.T) {
	preTestChecks(t)
	skipNoNsxtConfiguration(t)

	// String map to fill the template
	var params = StringMap{
		"Org":         testConfig.VCD.Org,
		"NsxtVdc":     testConfig.Nsxt.Vdc,
		"EdgeGw":      testConfig.Nsxt.EdgeGateway,
		"NetworkName": t.Name(),
		"Tags":        "network nsxt",
	}

	configText := templateFill(testAccRoutedNetDhcpStep1, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 0: %s", configText)

	params["FuncName"] = t.Name() + "-step1"
	configText1 := templateFill(testAccRoutedNetDhcpStep2, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckOpenApiVcdNetworkDestroy(testConfig.Nsxt.Vdc, "nsxt-routed-dhcp"),
		Steps: []resource.TestStep{
			resource.TestStep{
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
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_network_dhcp.pools", "pool.*", map[string]string{
						"start_address": "7.1.1.130",
						"end_address":   "7.1.1.140",
					}),
				),
			},
			resource.TestStep{
				ResourceName:      "vcd_nsxt_network_dhcp.pools",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdOrgNsxtVdcObject(testConfig, "nsxt-routed-dhcp"),
			},
		},
	})
	postTestChecks(t)
}

const testAccRoutedNetDhcpConfig = `
data "vcd_nsxt_edgegateway" "existing" {
  org  = "{{.Org}}"
  vdc  = "{{.NsxtVdc}}"
  name = "{{.EdgeGw}}"
}

resource "vcd_network_routed_v2" "net1" {
  org         = "{{.Org}}"
  vdc         = "{{.NsxtVdc}}"
  name        = "nsxt-routed-dhcp"
  description = "NSX-T routed network for DHCP testing"

  edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id

  gateway       = "7.1.1.1"
  prefix_length = 24

  static_ip_pool {
    start_address = "7.1.1.10"
    end_address   = "7.1.1.20"
  }
}
`

const testAccRoutedNetDhcpStep1 = testAccRoutedNetDhcpConfig + `
resource "vcd_nsxt_network_dhcp" "pools" {
  org = "{{.Org}}"
  vdc = "{{.NsxtVdc}}"

  org_network_id = vcd_network_routed_v2.net1.id

  pool {
    start_address = "7.1.1.100"
    end_address   = "7.1.1.110"
  }
}
`

const testAccRoutedNetDhcpStep2 = testAccRoutedNetDhcpConfig + `
resource "vcd_nsxt_network_dhcp" "pools" {
  org = "{{.Org}}"
  vdc = "{{.NsxtVdc}}"

  org_network_id = vcd_network_routed_v2.net1.id

  pool {
    start_address = "7.1.1.100"
    end_address   = "7.1.1.110"
  }

  pool {
    start_address = "7.1.1.130"
    end_address   = "7.1.1.140"
  }
}
`
