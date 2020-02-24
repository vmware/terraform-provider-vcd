// +build network vapp ALL functional

package vcd

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccVcdVappOrgNetwork_NotFenced(t *testing.T) {
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resourceName = "TestAccVcdVappOrgNetwork_NotFenced"

	var params = StringMap{
		"Org":                         testConfig.VCD.Org,
		"Vdc":                         testConfig.VCD.Vdc,
		"resourceName":                resourceName,
		"vappName":                    vappNameForNetworkTest,
		"EdgeGateway":                 testConfig.Networking.EdgeGateway,
		"NetworkName":                 "TestAccVcdVAppNetNotFenced",
		"orgNetwork":                  "TestAccVcdVAppNetNotFenced",
		"firewallEnabled":             "false",
		"firewallEnabledForUpdate":    "false",
		"natEnabled":                  "false",
		"natEnabledForUpdate":         "false",
		"retainIpMacEnabled":          "false",
		"retainIpMacEnabledForUpdate": "true",
		"isFenced":                    "false",
		"isFencedForUpdate":           "true",
		"FuncName":                    "TestAccVcdVappOrgNetwork_NotFenced",
	}

	rungVappOrgNetworkTest(t, params)
}

func TestAccVcdVappOrgNetwork_Fenced(t *testing.T) {
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resourceName = "TestAccVcdVappOrgNetwork_Fenced"

	var params = StringMap{
		"Org":                         testConfig.VCD.Org,
		"Vdc":                         testConfig.VCD.Vdc,
		"resourceName":                resourceName,
		"vappName":                    vappNameForNetworkTest,
		"EdgeGateway":                 testConfig.Networking.EdgeGateway,
		"NetworkName":                 "TestAccVcdVAppNetFenced",
		"orgNetwork":                  "TestAccVcdVAppNetFenced",
		"firewallEnabled":             "false",
		"firewallEnabledForUpdate":    "true",
		"natEnabled":                  "false",
		"natEnabledForUpdate":         "true",
		"retainIpMacEnabled":          "true",
		"retainIpMacEnabledForUpdate": "false",
		"isFenced":                    "true",
		"isFencedForUpdate":           "true",
		"FuncName":                    "TestAccVcdVappOrgNetwork_Fenced",
	}

	rungVappOrgNetworkTest(t, params)
}

func rungVappOrgNetworkTest(t *testing.T, params StringMap) {
	configText := templateFill(testAccCheckOrgVappNetwork_basic, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)
	params["FuncName"] = t.Name() + "-Update"
	updateConfigText := templateFill(testAccCheckOrgVappNetwork_update, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", updateConfigText)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVappNetworkDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVappNetworkExists("vcd_vapp_org_network."+params["resourceName"].(string)),
					resource.TestCheckResourceAttr(
						"vcd_vapp_org_network."+params["resourceName"].(string), "vapp_name", params["vappName"].(string)),
					resource.TestCheckResourceAttr(
						"vcd_vapp_org_network."+params["resourceName"].(string), "org_network_name", params["orgNetwork"].(string)),
					resource.TestCheckResourceAttr(
						"vcd_vapp_org_network."+params["resourceName"].(string), "retain_ip_mac_enabled", params["retainIpMacEnabled"].(string)),
					resource.TestCheckResourceAttr(
						"vcd_vapp_org_network."+params["resourceName"].(string), "is_fenced", params["isFenced"].(string)),
					resource.TestCheckResourceAttr(
						"vcd_vapp_org_network."+params["resourceName"].(string), "firewall_enabled", params["firewallEnabled"].(string)),
					resource.TestCheckResourceAttr(
						"vcd_vapp_org_network."+params["resourceName"].(string), "nat_enabled", params["natEnabled"].(string)),
				),
			},
			resource.TestStep{
				Config: updateConfigText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVappNetworkExists("vcd_vapp_org_network."+params["resourceName"].(string)),
					resource.TestCheckResourceAttr(
						"vcd_vapp_org_network."+params["resourceName"].(string), "vapp_name", params["vappName"].(string)),
					resource.TestCheckResourceAttr(
						"vcd_vapp_org_network."+params["resourceName"].(string), "org_network_name", params["orgNetwork"].(string)),
					resource.TestCheckResourceAttr(
						"vcd_vapp_org_network."+params["resourceName"].(string), "retain_ip_mac_enabled", params["retainIpMacEnabledForUpdate"].(string)),
					resource.TestCheckResourceAttr(
						"vcd_vapp_org_network."+params["resourceName"].(string), "is_fenced", params["isFencedForUpdate"].(string)),
					resource.TestCheckResourceAttr(
						"vcd_vapp_org_network."+params["resourceName"].(string), "firewall_enabled", params["firewallEnabledForUpdate"].(string)),
					resource.TestCheckResourceAttr(
						"vcd_vapp_org_network."+params["resourceName"].(string), "nat_enabled", params["natEnabledForUpdate"].(string)),
				),
			},
			resource.TestStep{
				ResourceName:      "vcd_vapp_org_network." + params["resourceName"].(string) + "-import",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdVappObject(testConfig, params["vappName"].(string), params["orgNetwork"].(string)),
				// These fields can't be retrieved from user data. firewall_enabled is null and as so we need ignore in some cases
				ImportStateVerifyIgnore: []string{"org", "vdc"},
			},
		},
	})
}

const testAccCheckOrgVappNetwork_basic = `
resource "vcd_vapp" "{{.vappName}}" {
  name = "{{.vappName}}"
  org  = "{{.Org}}"
  vdc  = "{{.Vdc}}"
}

resource "vcd_network_routed" "{{.NetworkName}}" {
  name         = "{{.NetworkName}}"
  org          = "{{.Org}}"
  vdc          = "{{.Vdc}}"
  edge_gateway = "{{.EdgeGateway}}"
  gateway      = "10.10.102.1"

  static_ip_pool {
    start_address = "10.10.102.2"
    end_address   = "10.10.102.254"
  }
}

resource "vcd_vapp_org_network" "{{.resourceName}}" {
  org                = "{{.Org}}"
  vdc                = "{{.Vdc}}"
  vapp_name          = "{{.vappName}}"
  org_network_name   = "{{.orgNetwork}}"
  
  is_fenced = "{{.isFenced}}"

  firewall_enabled      = "{{.firewallEnabled}}"
  nat_enabled           = "{{.natEnabled}}"
  retain_ip_mac_enabled = "{{.retainIpMacEnabled}}"

  depends_on = ["vcd_vapp.{{.vappName}}", "vcd_network_routed.{{.NetworkName}}"]
}
`

const testAccCheckOrgVappNetwork_update = `
# skip-binary-test: only for updates
resource "vcd_vapp" "{{.vappName}}" {
  name = "{{.vappName}}"
  org  = "{{.Org}}"
  vdc  = "{{.Vdc}}"
}

resource "vcd_network_routed" "{{.NetworkName}}" {
  name         = "{{.NetworkName}}"
  org          = "{{.Org}}"
  vdc          = "{{.Vdc}}"
  edge_gateway = "{{.EdgeGateway}}"
  gateway      = "10.10.102.1"

  static_ip_pool {
    start_address = "10.10.102.2"
    end_address   = "10.10.102.254"
  }
}

resource "vcd_vapp_org_network" "{{.resourceName}}" {
  org                = "{{.Org}}"
  vdc                = "{{.Vdc}}"
  vapp_name          = "{{.vappName}}"
  org_network_name   = "{{.orgNetwork}}"
  
  is_fenced = "{{.isFencedForUpdate}}"

  firewall_enabled      = "{{.firewallEnabledForUpdate}}"
  nat_enabled           = "{{.natEnabledForUpdate}}"
  retain_ip_mac_enabled = "{{.retainIpMacEnabledForUpdate}}"

  depends_on = ["vcd_vapp.{{.vappName}}", "vcd_network_routed.{{.NetworkName}}"]
}
`
