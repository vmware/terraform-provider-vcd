// +build gateway ALL functional

package vcd

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccVcdVpn_Basic(t *testing.T) {
	var vpnName string = "TestAccVcdVpnVpn"

	// String map to fill the template
	var params = StringMap{
		"Org":              testConfig.VCD.Org,
		"Vdc":              testConfig.VCD.Vdc,
		"EdgeGateway":      testConfig.Networking.EdgeGateway,
		"PeerID":           testConfig.Networking.Peer.PeerIp,
		"PeerIP":           testConfig.Networking.Peer.PeerIp,
		"LocalID":          testConfig.Networking.Local.LocalIp,
		"LocalIP":          testConfig.Networking.Local.LocalIp,
		"SharedSecret":     testConfig.Networking.SharedSecret,
		"LocalSubnetName1": "10.150.192.0/24",
		"LocalSubnetGW1":   "10.150.192.0",
		"LocalSubnetName2": "10.150.193.0/24",
		"LocalSubnetGW2":   "10.150.193.0",
		"PeerSubnetName1":  "192.168.5.0/24",
		"PeerSubnetGW1":    "192.168.5.0",
		"PeerSubnetName2":  "192.168.6.0/24",
		"PeerSubnetGW2":    "192.168.6.0",
		"SiteName":         "TestAccVcdVpnSite",
		"VpnName":          vpnName,
		"Tags":             "gateway",
	}
	configText := templateFill(testAccCheckVcdVpn_basic, params)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVcdVpnDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
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
    peer_subnet_name    = "{{.PeerSubnetName1}}"
    peer_subnet_gateway = "{{.PeerSubnetGW1}}"
    peer_subnet_mask    = "255.255.255.0"
  }

  peer_subnets {
    peer_subnet_name    = "{{.PeerSubnetName2}}"
    peer_subnet_gateway = "{{.PeerSubnetGW2}}"
    peer_subnet_mask    = "255.255.255.0"
  }

  local_subnets {
    local_subnet_name    = "{{.LocalSubnetName1}}"
    local_subnet_gateway = "{{.LocalSubnetGW1}}"
    local_subnet_mask    = "255.255.255.0"
  }

  local_subnets {
    local_subnet_name    = "{{.LocalSubnetName2}}"
    local_subnet_gateway = "{{.LocalSubnetGW2}}"
    local_subnet_mask    = "255.255.255.0"
  }
}
`
