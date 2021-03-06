// +build gateway ALL functional

package vcd

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccVcdNsxvDhcpRelay(t *testing.T) {
	preTestChecks(t)

	// String map to fill the template
	var params = StringMap{
		"Org":         testConfig.VCD.Org,
		"Vdc":         testConfig.VCD.Vdc,
		"EdgeGateway": testConfig.Networking.EdgeGateway,
		"Tags":        "gateway",
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
		ProviderFactories: testAccProviders,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckVcdDhcpRelaySettingsEmpty(),
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxv_dhcp_relay.relay_config", "id", regexp.MustCompile(`^.*:dhcpRelay$`)),
					resource.TestCheckResourceAttr("vcd_nsxv_dhcp_relay.relay_config", "domain_names.#", "2"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxv_dhcp_relay.relay_config", "domain_names.*", "servergroups.domainname.com"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxv_dhcp_relay.relay_config", "domain_names.*", "other.domain.com"),

					resource.TestCheckResourceAttr("vcd_nsxv_dhcp_relay.relay_config", "ip_addresses.#", "2"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxv_dhcp_relay.relay_config", "ip_addresses.*", "2.2.2.2"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxv_dhcp_relay.relay_config", "ip_addresses.*", "1.1.1.1"),

					resource.TestCheckResourceAttr("vcd_nsxv_dhcp_relay.relay_config", "ip_sets.#", "2"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxv_dhcp_relay.relay_config", "ip_sets.*", "test-set1"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxv_dhcp_relay.relay_config", "ip_sets.*", "test-set2"),

					resource.TestCheckResourceAttr("vcd_nsxv_dhcp_relay.relay_config", "relay_agent.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxv_dhcp_relay.relay_config", "relay_agent.*", map[string]string{
						"network_name":       "dhcp-relay-0",
						"gateway_ip_address": "210.201.0.1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxv_dhcp_relay.relay_config", "relay_agent.*", map[string]string{
						"network_name":       "dhcp-relay-1",
						"gateway_ip_address": "210.201.1.1",
					}),
				),
			},
			resource.TestStep{
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxv_dhcp_relay.relay_config", "id", regexp.MustCompile(`^urn:vcloud:gateway:.*:dhcpRelay$`)),
					resource.TestCheckResourceAttr("vcd_nsxv_dhcp_relay.relay_config", "domain_names.#", "0"),
					resource.TestCheckResourceAttr("vcd_nsxv_dhcp_relay.relay_config", "ip_addresses.#", "0"),

					resource.TestCheckResourceAttr("vcd_nsxv_dhcp_relay.relay_config", "ip_sets.#", "2"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxv_dhcp_relay.relay_config", "ip_sets.*", "test-set1"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxv_dhcp_relay.relay_config", "ip_sets.*", "test-set2"),

					resource.TestCheckResourceAttr("vcd_nsxv_dhcp_relay.relay_config", "relay_agent.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxv_dhcp_relay.relay_config", "relay_agent.*", map[string]string{
						"network_name":       "dhcp-relay-0",
						"gateway_ip_address": "210.201.0.1",
					}),
				),
			},
			resource.TestStep{
				ResourceName:      "vcd_nsxv_dhcp_relay.relay_config",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     testConfig.VCD.Org + "." + testConfig.VCD.Vdc + "." + testConfig.Networking.EdgeGateway,
			},
		},
	})
	postTestChecks(t)
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
  type    = list(string)
  default = ["internal", "subinterface"]
}

resource "vcd_network_routed" "test-routed" {
  count          = 2
  name           = "dhcp-relay-${count.index}"
  org            = "{{.Org}}"
  vdc            = "{{.Vdc}}"
  edge_gateway   = "{{.EdgeGateway}}"
  gateway        = "210.201.${count.index}.1"
  netmask        = "255.255.255.0"
  interface_type = var.network_types[count.index]

  static_ip_pool {
    start_address = "210.201.${count.index}.10"
    end_address   = "210.201.${count.index}.20"
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
  ip_sets      = [vcd_nsxv_ip_set.myset1.name, vcd_nsxv_ip_set.myset2.name]

  relay_agent {
    network_name = vcd_network_routed.test-routed[0].name
  }

  relay_agent {
    network_name        = vcd_network_routed.test-routed[1].name
    gateway_ip_address = "210.201.1.1"
  }
}

resource "vcd_nsxv_ip_set" "myset1" {
  name         = "test-set1"
  ip_addresses = ["192.168.1.1"]
}

resource "vcd_nsxv_ip_set" "myset2" {
  name         = "test-set2"
  ip_addresses = ["192.168.1.1"]
}
`

const testAccVcdNsxvDhcpRelayUpdate = testAccRoutedNet + `
resource "vcd_nsxv_dhcp_relay" "relay_config" {
  org          = "{{.Org}}"
  vdc          = "{{.Vdc}}"
  edge_gateway = "{{.EdgeGateway}}"

  ip_sets = [vcd_nsxv_ip_set.myset1.name, vcd_nsxv_ip_set.myset2.name]

  relay_agent {
    network_name = vcd_network_routed.test-routed[0].name
  }
}

resource "vcd_nsxv_ip_set" "myset1" {
  name         = "test-set1"
  ip_addresses = ["192.168.1.1"]
}

resource "vcd_nsxv_ip_set" "myset2" {
  name         = "test-set2"
  ip_addresses = ["192.168.1.1"]
}
`
