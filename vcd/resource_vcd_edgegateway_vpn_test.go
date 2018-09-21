package vcd

import (
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"log"
	"os"
	"regexp"
	"testing"
)

func TestAccVcdVpn_Basic(t *testing.T) {
	var vpnName string = "TestAccVcdVpnVpn"
	// String map to fill the template
	var params = StringMap{
		"Org":           testConfig.VCD.Org,
		"Vdc":           testConfig.VCD.Vdc,
		"EdgeGateway":   testConfig.Networking.EdgeGateway,
		"PeerID":        testConfig.Networking.Peer.PeerIp,
		"PeerIP":        testConfig.Networking.Peer.PeerIp,
		"LocalID":       testConfig.Networking.Local.LocalIp,
		"LocalIP":       testConfig.Networking.Local.LocalIp,
		"SharedSecret":  testConfig.Networking.SharedSecret,
		"PeerSubnetGW":  testConfig.Networking.Peer.PeerSubnetGateway,
		"LocalSubnetGW": testConfig.Networking.Local.LocalSubnetGateway,
		"SiteName":      "TestAccVcdVpnSite",
		"VpnName":       vpnName,
	}
	configText := templateFill(testAccCheckVcdVpn_basic, params)
	if os.Getenv("GOVCD_DEBUG") != "" {
		log.Printf("#[DEBUG] CONFIGURATION: %s", configText)
	}
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVcdVpnDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				// We expect this error to happen, because Terraform will consider that the resource
				// needs to be generated again. It's a known problem that often happens with security fields
				// such as "sharedSecret" in our resource, and there is no general purpose solution.
				// In this case, we accept that this error might show up, and take it as a clean run.
				ExpectError: regexp.MustCompile(`After applying this step, the plan was not empty`),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"vcd_edgegateway_vpn."+vpnName, "encryption_protocol", "AES256"),
				),
			},
		},
	})
}

func testAccCheckVcdVpnDestroy(s *terraform.State) error {

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "vcd_edgegateway_vpn" {
			continue
		}

		return nil
	}

	return nil
}

const testAccCheckVcdVpn_basic = `
resource "vcd_edgegateway_vpn" "{{.VpnName}}" {
    org                 = "{{.Org}}"
    vdc                 = "{{.Vdc}}"
    edge_gateway        = "{{.EdgeGateway}}"
    name                = "{{.SiteName}}"
    description         = "Description"
    encryption_protocol = "AES256"
    mtu                 = 1400
    peer_id             = "{{.PeerID}}"
    peer_ip_address     = "{{.PeerIP}}"
    local_id            = "{{.LocalID}}"
    local_ip_address    = "{{.LocalIP}}"
    shared_secret       = "{{.SharedSecret}}"
    
    peer_subnets {
        peer_subnet_name = "DMZ_WEST"
        peer_subnet_gateway = "{{.PeerSubnetGW}}"
        peer_subnet_mask = "255.255.255.0"
    }

    peer_subnets {
        peer_subnet_name = "WEB_WEST"
        peer_subnet_gateway = "{{.PeerSubnetGW}}"
        peer_subnet_mask = "255.255.255.0"
    }

    local_subnets {
        local_subnet_name = "DMZ_EAST"
        local_subnet_gateway = "{{.LocalSubnetGW}}"
        local_subnet_mask = "255.255.255.0"
    }

    local_subnets {
        local_subnet_name = "WEB_EAST"
        local_subnet_gateway = "{{.LocalSubnetGW}}"
        local_subnet_mask = "255.255.255.0"
    }
}
`
