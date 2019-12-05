// +build gateway ALL functional

package vcd

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccVcdNsxvDhcpRelay(t *testing.T) {

	// String map to fill the template
	var params = StringMap{
		"Org":         testConfig.VCD.Org,
		"Vdc":         testConfig.VCD.Vdc,
		"EdgeGateway": testConfig.Networking.EdgeGateway,
		// "OrgNetwork": testConfig.Networking.,
		"Tags": "gateway",
	}

	configText := templateFill(testAccVcdNsxvDhcpRelay, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 0: %s", configText)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	if !edgeGatewayIsAdvanced() {
		t.Skip(t.Name() + "requires advanced edge gateway to work")
	}

	resource.ParallelTest(t, resource.TestCase{
		Providers: testAccProviders,
		PreCheck:  func() { testAccPreCheck(t) },
		// CheckDestroy: testAccCheckVcdLbServiceMonitorDestroy(params["ServiceMonitorName"].(string)),
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxv_dhcp_relay.relay_config", "id", regexp.MustCompile(`^.*:dhcpRelaySettings`)),
				),
			},
		},
	})
}

const testAccVcdNsxvDhcpRelay = `
resource "vcd_nsxv_dhcp_relay" "relay_config" {
	org          = "{{.Org}}"
	vdc          = "{{.Vdc}}"
	edge_gateway = "{{.EdgeGateway}}"
	
    ip_addresses = ["1.1.1.1", "2.2.2.2"]
    domain_names = ["servergroups.domainname.com", "other.domain.com"]
    ip_sets      = ["myset1", "myset2"]
    relay_agent {
        org_network        = "my-vdc-int-net"
        # gateway_ip_address  = "10.10.10.5"  # optional
    }

}
`
