// +build gateway ALL functional

package vcd

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
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

	params["FuncName"] = t.Name() + "-step1"
	configText1 := templateFill(testAccVcdNsxvDhcpRelayUpdate, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	if !edgeGatewayIsAdvanced() {
		t.Skip(t.Name() + "requires advanced edge gateway to work")
	}

	resource.ParallelTest(t, resource.TestCase{
		Providers:    testAccProviders,
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccCheckVcdDhcpRelaySettingsEmpty(),
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxv_dhcp_relay.relay_config", "id", regexp.MustCompile(`^.*:dhcpRelay$`)),
					resource.TestCheckResourceAttr("vcd_nsxv_dhcp_relay.relay_config", "domain_names.#", "2"),
					resource.TestCheckResourceAttr("vcd_nsxv_dhcp_relay.relay_config", "domain_names.2956203856", "servergroups.domainname.com"),
					resource.TestCheckResourceAttr("vcd_nsxv_dhcp_relay.relay_config", "domain_names.4048773415", "other.domain.com"),

					resource.TestCheckResourceAttr("vcd_nsxv_dhcp_relay.relay_config", "ip_addresses.#", "2"),
					resource.TestCheckResourceAttr("vcd_nsxv_dhcp_relay.relay_config", "ip_addresses.1048647934", "2.2.2.2"),
					resource.TestCheckResourceAttr("vcd_nsxv_dhcp_relay.relay_config", "ip_addresses.251826590", "1.1.1.1"),

					resource.TestCheckResourceAttr("vcd_nsxv_dhcp_relay.relay_config", "ip_sets.#", "2"),
					resource.TestCheckResourceAttr("vcd_nsxv_dhcp_relay.relay_config", "ip_sets.489836264", "test-set1"),
					resource.TestCheckResourceAttr("vcd_nsxv_dhcp_relay.relay_config", "ip_sets.908008747", "test-set2"),

					resource.TestCheckResourceAttr("vcd_nsxv_dhcp_relay.relay_config", "relay_agent.#", "2"),
					resource.TestCheckResourceAttr("vcd_nsxv_dhcp_relay.relay_config", "relay_agent.3348209499.org_network", "dhcp-relay-0"),
					resource.TestCheckResourceAttr("vcd_nsxv_dhcp_relay.relay_config", "relay_agent.3348209499.gateway_ip_address", "10.201.0.1"),
					resource.TestCheckResourceAttr("vcd_nsxv_dhcp_relay.relay_config", "relay_agent.3180164926.org_network", "dhcp-relay-1"),
					resource.TestCheckResourceAttr("vcd_nsxv_dhcp_relay.relay_config", "relay_agent.3180164926.gateway_ip_address", "10.201.1.1"),

					// Validate that data source has all fields except the hashed IP set because it is turned into slice in data source
					// and only one due to outstanding problem in Terraform plugin SDK - https://github.com/hashicorp/terraform-plugin-sdk/pull/197
					resourceFieldsEqual("vcd_nsxv_dhcp_relay.relay_config", "data.vcd_nsxv_dhcp_relay.relay",
						[]string{"relay_agent.3348209499.gateway_ip_address", "relay_agent.3348209499.org_network", "relay_agent.#",
							"relay_agent.3180164926.org_network", "relay_agent.3180164926.gateway_ip_address"}),
				),
			},
			resource.TestStep{
				ResourceName:      "vcd_nsxv_dhcp_relay.imported",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     testConfig.VCD.Org + "." + testConfig.VCD.Vdc + "." + testConfig.Networking.EdgeGateway,
			},
			resource.TestStep{
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxv_dhcp_relay.relay_config", "id", regexp.MustCompile(`^urn:vcloud:gateway:.*:dhcpRelay$`)),
					resource.TestCheckResourceAttr("vcd_nsxv_dhcp_relay.relay_config", "domain_names.#", "0"),
					resource.TestCheckResourceAttr("vcd_nsxv_dhcp_relay.relay_config", "ip_addresses.#", "0"),

					resource.TestCheckResourceAttr("vcd_nsxv_dhcp_relay.relay_config", "ip_sets.#", "2"),
					resource.TestCheckResourceAttr("vcd_nsxv_dhcp_relay.relay_config", "ip_sets.489836264", "test-set1"),
					resource.TestCheckResourceAttr("vcd_nsxv_dhcp_relay.relay_config", "ip_sets.908008747", "test-set2"),

					resource.TestCheckResourceAttr("vcd_nsxv_dhcp_relay.relay_config", "relay_agent.#", "1"),
					resource.TestCheckResourceAttr("vcd_nsxv_dhcp_relay.relay_config", "relay_agent.3348209499.org_network", "dhcp-relay-0"),
					resource.TestCheckResourceAttr("vcd_nsxv_dhcp_relay.relay_config", "relay_agent.3348209499.gateway_ip_address", "10.201.0.1"),

					// Validate that data source has all fields except the hashed IP set because it is turned into slice in data source
					// and only one due to outstanding problem in Terraform plugin SDK - https://github.com/hashicorp/terraform-plugin-sdk/pull/197
					resourceFieldsEqual("vcd_nsxv_dhcp_relay.relay_config", "data.vcd_nsxv_dhcp_relay.relay",
						[]string{"relay_agent.3348209499.gateway_ip_address", "relay_agent.3348209499.org_network", "relay_agent.#"}),
				),
			},
		},
	})
}

// testAccCheckVcdDhcpRelaySettingsEmpty reads DHCP relay configuration and ensure it has no
// settings set.
func testAccCheckVcdDhcpRelaySettingsEmpty() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*VCDClient)
		edgeGateway, err := conn.GetEdgeGateway(testConfig.VCD.Org, testConfig.VCD.Vdc, testConfig.Networking.EdgeGateway)
		if err != nil {
			return fmt.Errorf(errorUnableToFindEdgeGateway, err)
		}

		dhcpRelaySettings, err := edgeGateway.GetDhcpRelay()
		if err != nil {
			return fmt.Errorf("could not read DHCP relay settings: %s", err)
		}

		// Validate that DHCP relay settings are empty
		if dhcpRelaySettings.RelayServer != nil || dhcpRelaySettings.RelayAgents != nil {
			return fmt.Errorf("DHCP relay settings were not cleaned up")
		}

		return nil
	}
}

const testAccRoutedNet = `
variable "network_types" {
	type        = list(string)
	default     = ["internal", "subinterface"]
}

resource "vcd_network_routed" "test-routed" {
	count          = 2
	name           = "dhcp-relay-${count.index}"
	org            = "{{.Org}}"
	vdc            = "{{.Vdc}}"
	edge_gateway   = "{{.EdgeGateway}}"
	gateway        = "10.201.${count.index}.1"
	netmask        = "255.255.255.0"
	interface_type = var.network_types[count.index]

	static_ip_pool {
	  start_address = "10.201.${count.index}.10"
	  end_address   = "10.201.${count.index}.20"
	}
}
`

const testAccVcdNsxvDhcpRelay = testAccRoutedNet + `
resource "vcd_nsxv_dhcp_relay" "relay_config" {
	org          = "{{.Org}}"
	vdc          = "{{.Vdc}}"
	edge_gateway = "{{.EdgeGateway}}"
	
    ip_addresses = ["1.1.1.1", "2.2.2.2"]
    domain_names = ["servergroups.domainname.com", "other.domain.com"]
    ip_sets      = [vcd_ipset.myset1.name, vcd_ipset.myset2.name]
	
	relay_agent {
        org_network = vcd_network_routed.test-routed[0].name
	}
	
	relay_agent {
		org_network        = vcd_network_routed.test-routed[1].name
		gateway_ip_address = "10.201.1.1"
    }
}

data "vcd_nsxv_dhcp_relay" "relay" {
	org          = "{{.Org}}"
	vdc          = "{{.Vdc}}"
	edge_gateway = vcd_nsxv_dhcp_relay.relay_config.edge_gateway
}

resource "vcd_ipset" "myset1" {
	name                   = "test-set1"
	ip_addresses           = ["192.168.1.1"]
}

resource "vcd_ipset" "myset2" {
	name                   = "test-set2"
	ip_addresses           = ["192.168.1.1"]
}
`

const testAccVcdNsxvDhcpRelayUpdate = testAccRoutedNet + `
resource "vcd_nsxv_dhcp_relay" "relay_config" {
	org          = "{{.Org}}"
	vdc          = "{{.Vdc}}"
	edge_gateway = "{{.EdgeGateway}}"
	
    ip_sets      = [vcd_ipset.myset1.name, vcd_ipset.myset2.name]
	
	relay_agent {
        org_network = vcd_network_routed.test-routed[0].name
	}
}

data "vcd_nsxv_dhcp_relay" "relay" {
	org          = "{{.Org}}"
	vdc          = "{{.Vdc}}"
	edge_gateway = vcd_nsxv_dhcp_relay.relay_config.edge_gateway
}

resource "vcd_ipset" "myset1" {
	name                   = "test-set1"
	ip_addresses           = ["192.168.1.1"]
}

resource "vcd_ipset" "myset2" {
	name                   = "test-set2"
	ip_addresses           = ["192.168.1.1"]
}
`
