// +build functional gateway ALL

package vcd

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccVcdVappFirewallRules(t *testing.T) {
	if testConfig.Networking.EdgeGateway == "" {
		t.Skip("Variable testConfig.Networking.EdgeGateway must be configured")
		return
	}

	var (
		vmName1  = "TestAccVcdVappFirewallRulesVm1"
		vmName2  = "TestAccVcdVappFirewallRulesVm2"
		vmName3  = "TestAccVcdVappFirewallRulesVm3"
		vappName = t.Name() + "_vapp"
	)

	var params = StringMap{
		"Org":           testConfig.VCD.Org,
		"Vdc":           testConfig.VCD.Vdc,
		"EdgeGateway":   testConfig.Networking.EdgeGateway,
		"DefaultAction": "drop",
		"ResourceName":  t.Name(),
		"FuncName":      t.Name(),
		"Description1":  "description1",
		"Description2":  "description2",
		"Description3":  "description3",
		"Description4":  "description4",
		"NetworkName":   "TestAccVcdVAppVmNet",
		"VappName":      vappName,
		"VmName1":       vmName1,
		"VmName2":       vmName2,
		"VmName3":       vmName3,
	}
	configText := templateFill(testAccVcdVappFirewallRules_rules, params)
	params["FuncName"] = t.Name() + "-step2"
	configTextForUpdate := templateFill(testAccVcdVappFirewallRules_rules_forUpdate, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdVappFirewallRulesExists("vcd_vapp_firewall_rules."+t.Name(), params["Description1"].(string)),
					resource.TestCheckResourceAttr("vcd_vapp_firewall_rules."+t.Name(), "rule.0.enabled", "true"),
					resource.TestCheckResourceAttr("vcd_vapp_firewall_rules."+t.Name(), "rule.0.policy", "drop"),
					resource.TestCheckResourceAttr("vcd_vapp_firewall_rules."+t.Name(), "rule.0.protocol", "udp"),
					resource.TestCheckResourceAttr("vcd_vapp_firewall_rules."+t.Name(), "rule.0.destination_port", "21"),
					resource.TestCheckResourceAttr("vcd_vapp_firewall_rules."+t.Name(), "rule.0.destination_ip", "any"),
					resource.TestCheckResourceAttr("vcd_vapp_firewall_rules."+t.Name(), "rule.0.source_port", "any"),
					resource.TestCheckResourceAttr("vcd_vapp_firewall_rules."+t.Name(), "rule.0.source_ip", "10.10.0.0/24"),
					resource.TestCheckResourceAttr("vcd_vapp_firewall_rules."+t.Name(), "rule.0.enable_logging", "false"),

					testAccCheckVcdVappFirewallRulesExists("vcd_vapp_firewall_rules."+t.Name(), params["Description2"].(string)),
					resource.TestCheckResourceAttr("vcd_vapp_firewall_rules."+t.Name(), "rule.1.enabled", "true"),
					resource.TestCheckResourceAttr("vcd_vapp_firewall_rules."+t.Name(), "rule.1.policy", "allow"),
					resource.TestCheckResourceAttr("vcd_vapp_firewall_rules."+t.Name(), "rule.1.protocol", "any"),
					resource.TestCheckResourceAttr("vcd_vapp_firewall_rules."+t.Name(), "rule.1.destination_port", "any"),
					resource.TestCheckResourceAttr("vcd_vapp_firewall_rules."+t.Name(), "rule.1.destination_ip", "any"),
					resource.TestCheckResourceAttr("vcd_vapp_firewall_rules."+t.Name(), "rule.1.source_port", "any"),
					resource.TestCheckResourceAttr("vcd_vapp_firewall_rules."+t.Name(), "rule.1.source_ip", "any"),
					resource.TestCheckResourceAttr("vcd_vapp_firewall_rules."+t.Name(), "rule.1.enable_logging", "false"),

					testAccCheckVcdVappFirewallRulesExists("vcd_vapp_firewall_rules."+t.Name(), params["Description3"].(string)),
					resource.TestCheckResourceAttr("vcd_vapp_firewall_rules."+t.Name(), "rule.2.enabled", "true"),
					resource.TestCheckResourceAttr("vcd_vapp_firewall_rules."+t.Name(), "rule.2.policy", "allow"),
					resource.TestCheckResourceAttr("vcd_vapp_firewall_rules."+t.Name(), "rule.2.protocol", "any"),
					resource.TestCheckResourceAttr("vcd_vapp_firewall_rules."+t.Name(), "rule.2.destination_port", "any"),
					resource.TestCheckResourceAttr("vcd_vapp_firewall_rules."+t.Name(), "rule.2.destination_vm_ip_type", "assigned"),
					resource.TestCheckResourceAttr("vcd_vapp_firewall_rules."+t.Name(), "rule.2.destination_vm_nic_id", "0"),
					resource.TestCheckResourceAttr("vcd_vapp_firewall_rules."+t.Name(), "rule.2.source_port", "any"),
					resource.TestCheckResourceAttr("vcd_vapp_firewall_rules."+t.Name(), "rule.2.source_vm_nic_id", "0"),
					resource.TestCheckResourceAttr("vcd_vapp_firewall_rules."+t.Name(), "rule.2.source_vm_ip_type", "NAT"),
					resource.TestCheckResourceAttr("vcd_vapp_firewall_rules."+t.Name(), "rule.2.enable_logging", "false"),

					testAccCheckVcdVappFirewallRulesExists("vcd_vapp_firewall_rules."+t.Name(), params["Description4"].(string)),
					resource.TestCheckResourceAttr("vcd_vapp_firewall_rules."+t.Name(), "rule.3.enabled", "false"),
					resource.TestCheckResourceAttr("vcd_vapp_firewall_rules."+t.Name(), "rule.3.policy", "drop"),
					resource.TestCheckResourceAttr("vcd_vapp_firewall_rules."+t.Name(), "rule.3.protocol", "any"),
					resource.TestCheckResourceAttr("vcd_vapp_firewall_rules."+t.Name(), "rule.3.destination_port", "any"),
					resource.TestCheckResourceAttr("vcd_vapp_firewall_rules."+t.Name(), "rule.3.destination_ip", "external"),
					resource.TestCheckResourceAttr("vcd_vapp_firewall_rules."+t.Name(), "rule.3.source_port", "any"),
					resource.TestCheckResourceAttr("vcd_vapp_firewall_rules."+t.Name(), "rule.3.source_ip", "internal"),
					resource.TestCheckResourceAttr("vcd_vapp_firewall_rules."+t.Name(), "rule.3.enable_logging", "true"),

					testAccCheckVcdVappFirewallRulesExists("vcd_vapp_firewall_rules."+t.Name()+"2", params["Description1"].(string)),
					resource.TestCheckResourceAttr("vcd_vapp_firewall_rules."+t.Name()+"2", "rule.0.enabled", "true"),
					resource.TestCheckResourceAttr("vcd_vapp_firewall_rules."+t.Name()+"2", "rule.0.policy", "drop"),
					resource.TestCheckResourceAttr("vcd_vapp_firewall_rules."+t.Name()+"2", "rule.0.protocol", "udp"),
					resource.TestCheckResourceAttr("vcd_vapp_firewall_rules."+t.Name()+"2", "rule.0.destination_port", "221"),
					resource.TestCheckResourceAttr("vcd_vapp_firewall_rules."+t.Name()+"2", "rule.0.destination_ip", "any"),
					resource.TestCheckResourceAttr("vcd_vapp_firewall_rules."+t.Name()+"2", "rule.0.source_port", "any"),
					resource.TestCheckResourceAttr("vcd_vapp_firewall_rules."+t.Name()+"2", "rule.0.source_ip", "10.10.0.10/24"),
					resource.TestCheckResourceAttr("vcd_vapp_firewall_rules."+t.Name()+"2", "rule.0.enable_logging", "false"),
				),
			},
			resource.TestStep{ // Step 2 - resource import
				ResourceName:            "vcd_vapp_firewall_rules.imported",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdFunc:       importStateVappFirewallRuleObject(testConfig, vappName, "vapp-routed-net"),
				ImportStateVerifyIgnore: []string{"network_id", "org", "vdc"},
			},
			resource.TestStep{ // Step r - update
				Config: configTextForUpdate,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdVappFirewallRulesExists("vcd_vapp_firewall_rules."+t.Name(), params["Description1"].(string)),
					resource.TestCheckResourceAttr("vcd_vapp_firewall_rules."+t.Name(), "rule.0.enabled", "true"),
					resource.TestCheckResourceAttr("vcd_vapp_firewall_rules."+t.Name(), "rule.0.policy", "drop"),
					resource.TestCheckResourceAttr("vcd_vapp_firewall_rules."+t.Name(), "rule.0.protocol", "udp"),
					resource.TestCheckResourceAttr("vcd_vapp_firewall_rules."+t.Name(), "rule.0.destination_port", "21"),
					resource.TestCheckResourceAttr("vcd_vapp_firewall_rules."+t.Name(), "rule.0.destination_ip", "any"),
					resource.TestCheckResourceAttr("vcd_vapp_firewall_rules."+t.Name(), "rule.0.source_port", "any"),
					resource.TestCheckResourceAttr("vcd_vapp_firewall_rules."+t.Name(), "rule.0.source_ip", "10.10.0.0/24"),
					resource.TestCheckResourceAttr("vcd_vapp_firewall_rules."+t.Name(), "rule.0.enable_logging", "false"),

					testAccCheckVcdVappFirewallRulesExists("vcd_vapp_firewall_rules."+t.Name(), params["Description2"].(string)),
					resource.TestCheckResourceAttr("vcd_vapp_firewall_rules."+t.Name(), "rule.1.enabled", "true"),
					resource.TestCheckResourceAttr("vcd_vapp_firewall_rules."+t.Name(), "rule.1.policy", "allow"),
					resource.TestCheckResourceAttr("vcd_vapp_firewall_rules."+t.Name(), "rule.1.protocol", "any"),
					resource.TestCheckResourceAttr("vcd_vapp_firewall_rules."+t.Name(), "rule.1.destination_port", "any"),
					resource.TestCheckResourceAttr("vcd_vapp_firewall_rules."+t.Name(), "rule.1.destination_ip", "any"),
					resource.TestCheckResourceAttr("vcd_vapp_firewall_rules."+t.Name(), "rule.1.source_port", "any"),
					resource.TestCheckResourceAttr("vcd_vapp_firewall_rules."+t.Name(), "rule.1.source_ip", "any"),
					resource.TestCheckResourceAttr("vcd_vapp_firewall_rules."+t.Name(), "rule.1.enable_logging", "false"),

					testAccCheckVcdVappFirewallRulesExists("vcd_vapp_firewall_rules."+t.Name(), params["Description4"].(string)),
					resource.TestCheckResourceAttr("vcd_vapp_firewall_rules."+t.Name(), "rule.2.enabled", "false"),
					resource.TestCheckResourceAttr("vcd_vapp_firewall_rules."+t.Name(), "rule.2.policy", "drop"),
					resource.TestCheckResourceAttr("vcd_vapp_firewall_rules."+t.Name(), "rule.2.protocol", "any"),
					resource.TestCheckResourceAttr("vcd_vapp_firewall_rules."+t.Name(), "rule.2.destination_port", "any"),
					resource.TestCheckResourceAttr("vcd_vapp_firewall_rules."+t.Name(), "rule.2.destination_ip", "internal"),
					resource.TestCheckResourceAttr("vcd_vapp_firewall_rules."+t.Name(), "rule.2.source_port", "any"),
					resource.TestCheckResourceAttr("vcd_vapp_firewall_rules."+t.Name(), "rule.2.source_ip", "external"),
					resource.TestCheckResourceAttr("vcd_vapp_firewall_rules."+t.Name(), "rule.2.enable_logging", "true"),

					testAccCheckVcdVappFirewallRulesExists("vcd_vapp_firewall_rules."+t.Name(), params["Description3"].(string)),
					resource.TestCheckResourceAttr("vcd_vapp_firewall_rules."+t.Name(), "rule.3.enabled", "true"),
					resource.TestCheckResourceAttr("vcd_vapp_firewall_rules."+t.Name(), "rule.3.policy", "allow"),
					resource.TestCheckResourceAttr("vcd_vapp_firewall_rules."+t.Name(), "rule.3.protocol", "any"),
					resource.TestCheckResourceAttr("vcd_vapp_firewall_rules."+t.Name(), "rule.3.destination_port", "any"),
					resource.TestCheckResourceAttr("vcd_vapp_firewall_rules."+t.Name(), "rule.3.destination_vm_ip_type", "assigned"),
					resource.TestCheckResourceAttr("vcd_vapp_firewall_rules."+t.Name(), "rule.3.destination_vm_nic_id", "0"),
					resource.TestCheckResourceAttr("vcd_vapp_firewall_rules."+t.Name(), "rule.3.source_port", "any"),
					resource.TestCheckResourceAttr("vcd_vapp_firewall_rules."+t.Name(), "rule.3.source_vm_nic_id", "0"),
					resource.TestCheckResourceAttr("vcd_vapp_firewall_rules."+t.Name(), "rule.3.source_vm_ip_type", "NAT"),
					resource.TestCheckResourceAttr("vcd_vapp_firewall_rules."+t.Name(), "rule.3.enable_logging", "false"),

					testAccCheckVcdVappFirewallRulesExists("vcd_vapp_firewall_rules."+t.Name()+"2", params["Description1"].(string)),
					resource.TestCheckResourceAttr("vcd_vapp_firewall_rules."+t.Name()+"2", "rule.0.enabled", "true"),
					resource.TestCheckResourceAttr("vcd_vapp_firewall_rules."+t.Name()+"2", "rule.0.policy", "drop"),
					resource.TestCheckResourceAttr("vcd_vapp_firewall_rules."+t.Name()+"2", "rule.0.protocol", "udp"),
					resource.TestCheckResourceAttr("vcd_vapp_firewall_rules."+t.Name()+"2", "rule.0.destination_port", "221"),
					resource.TestCheckResourceAttr("vcd_vapp_firewall_rules."+t.Name()+"2", "rule.0.destination_ip", "any"),
					resource.TestCheckResourceAttr("vcd_vapp_firewall_rules."+t.Name()+"2", "rule.0.source_port", "any"),
					resource.TestCheckResourceAttr("vcd_vapp_firewall_rules."+t.Name()+"2", "rule.0.source_ip", "10.10.0.10/24"),
					resource.TestCheckResourceAttr("vcd_vapp_firewall_rules."+t.Name()+"2", "rule.0.enable_logging", "false"),
				),
			},
		},
	})

}

func importStateVappFirewallRuleObject(testConfig TestConfig, vappName, objectName string) resource.ImportStateIdFunc {
	return func(*terraform.State) (string, error) {
		if testConfig.VCD.Org == "" || testConfig.VCD.Vdc == "" || vappName == "" || objectName == "" {
			return "", fmt.Errorf("missing information to generate import path")
		}
		return testConfig.VCD.Org +
			ImportSeparator +
			testConfig.VCD.Vdc +
			ImportSeparator +
			vappName +
			ImportSeparator +
			objectName, nil
	}
}

func testAccCheckVcdVappFirewallRulesExists(n, description string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]

		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no record ID is set")
		}

		conn := testAccProvider.Meta().(*VCDClient)

		_, vdc, err := conn.GetOrgAndVdc(testConfig.VCD.Org, testConfig.VCD.Vdc)
		if err != nil {
			return fmt.Errorf(errorRetrievingVdcFromOrg, testConfig.VCD.Vdc, testConfig.VCD.Org, err)
		}

		vapp, err := vdc.GetVAppById(rs.Primary.Attributes["vapp_id"], false)
		if err != nil {
			return err
		}

		vapp_network, err := vapp.GetVappNetworkById(rs.Primary.Attributes["network_id"], false)
		if err != nil {
			return err
		}

		for _, rule := range vapp_network.Configuration.Features.FirewallService.FirewallRule {
			if rule.Description == description {
				return nil
			}
		}

		return fmt.Errorf("no rule with provided description is found")
	}
}

const testAccVcdVappFirewallRules_vappAndVm = `
resource "vcd_network_routed" "network_routed" {
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

resource "vcd_vapp" "{{.VappName}}" {
  name = "{{.VappName}}"
  org  = "{{.Org}}"
  vdc  = "{{.Vdc}}"
}

resource "vcd_vapp_network" "vappRoutedNet" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  name             = "vapp-routed-net"
  vapp_name        = vcd_vapp.{{.VappName}}.name
  gateway          = "192.168.2.1"
  netmask          = "255.255.255.0"
  org_network_name = vcd_network_routed.network_routed.name
}

resource "vcd_vapp_org_network" "vappAttachedNet" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  vapp_name        = vcd_vapp.{{.VappName}}.name
  org_network_name = vcd_network_routed.network_routed.name
  is_fenced        = true
}

resource "vcd_vapp_vm" "{{.VmName1}}" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  vapp_name   = vcd_vapp.{{.VappName}}.name
  description = "test empty VM"
  name        = "{{.VmName1}}"
  memory      = 512
  cpus        = 2
  cpu_cores   = 1 
  
  os_type                        = "sles10_64Guest"
  hardware_version               = "vmx-11"
  expose_hardware_virtualization = true
  computer_name                  = "compName"
  network {
    type               = "vapp"
    name               = vcd_vapp_network.vappRoutedNet.name
    ip_allocation_mode = "MANUAL"
    ip                 = "192.168.2.11"
  }

}

resource "vcd_vapp_vm" "{{.VmName2}}" {
  org           = "{{.Org}}"
  vdc           = "{{.Vdc}}"
  vapp_name     = vcd_vapp.{{.VappName}}.name

  description = "test empty VM"
  name        = "{{.VmName2}}"
  memory      = 512
  cpus        = 2
  cpu_cores   = 1 
  
  os_type          = "sles10_64Guest"
  hardware_version = "vmx-11"
  computer_name    = "compName"
  network {
    type               = "vapp"
    name               = vcd_vapp_network.vappRoutedNet.name
    ip_allocation_mode = "MANUAL"
    ip            = "192.168.2.12"
  }

  network {
    type               = "org"
    name               = vcd_vapp_org_network.vappAttachedNet.org_network_name
    ip_allocation_mode = "POOL"
  }
}
`

const testAccVcdVappFirewallRules_rules = testAccVcdVappFirewallRules_vappAndVm + `
resource "vcd_vapp_firewall_rules" "{{.ResourceName}}" {
  org            = "{{.Org}}"
  vdc            = "{{.Vdc}}"
  vapp_id        = vcd_vapp.TestAccVcdVappFirewallRules_vapp.id
  default_action = "{{.DefaultAction}}"
  network_id     = vcd_vapp_network.vappRoutedNet.id

  rule {
    description      = "{{.Description1}}"
    policy           = "drop"
    protocol         = "udp"
    destination_port = "21"
    destination_ip   = "any"
    source_port      = "any"
    source_ip        = "10.10.0.0/24"
  }

  rule {
    description      = "{{.Description2}}"
    policy           = "allow"
    protocol         = "any"
    destination_port = "any"
    destination_ip   = "any"
    source_port      = "any"
    source_ip        = "any"
  }

  rule {
    description            = "{{.Description3}}"
    policy                 = "allow"
    protocol               = "any"
    destination_vm_id      = vcd_vapp_vm.{{.VmName2}}.id
    destination_vm_nic_id  = 0
    destination_vm_ip_type = "assigned"
    destination_port       = "any"
    source_vm_id           = vcd_vapp_vm.{{.VmName1}}.id
    source_vm_nic_id       = 0
    source_vm_ip_type      = "NAT"
    source_port            = "any"
  }

  rule {
    description      = "{{.Description4}}"
    enabled          = false
    policy           = "drop"
    protocol         = "any"
    destination_port = "any"
    destination_ip   = "external"
    source_port      = "any"
    source_ip        = "internal"
    enable_logging   = true
  }
}

resource "vcd_vapp_firewall_rules" "{{.ResourceName}}2" {
  org            = "{{.Org}}"
  vdc            = "{{.Vdc}}"
  vapp_id        = vcd_vapp.TestAccVcdVappFirewallRules_vapp.id
  default_action = "{{.DefaultAction}}"
  network_id     = vcd_vapp_org_network.vappAttachedNet.id

  rule {
    description      = "{{.Description1}}"
    policy           = "drop"
    protocol         = "udp"
    destination_port = "221"
    destination_ip   = "any"
    source_port      = "any"
    source_ip        = "10.10.0.10/24"
  }
}
`

const testAccVcdVappFirewallRules_rules_forUpdate = testAccVcdVappFirewallRules_vappAndVm + `
resource "vcd_vapp_firewall_rules" "{{.ResourceName}}" {
  org            = "{{.Org}}"
  vdc            = "{{.Vdc}}"
  vapp_id        = vcd_vapp.TestAccVcdVappFirewallRules_vapp.id
  default_action = "{{.DefaultAction}}"
  network_id     = vcd_vapp_network.vappRoutedNet.id

  rule {
    description      = "{{.Description1}}"
    policy           = "drop"
    protocol         = "udp"
    destination_port = "21"
    destination_ip   = "any"
    source_port      = "any"
    source_ip        = "10.10.0.0/24"
  }

  rule {
    description      = "{{.Description2}}"
    policy           = "allow"
    protocol         = "any"
    destination_port = "any"
    destination_ip   = "any"
    source_port      = "any"
    source_ip        = "any"
  }

  rule {
    description      = "{{.Description4}}"
    enabled          = false
    policy           = "drop"
    protocol         = "any"
    destination_port = "any"
    destination_ip   = "internal"
    source_port      = "any"
    source_ip        = "external"
    enable_logging   = true
  }

  rule {
    description            = "{{.Description3}}"
    policy                 = "allow"
    protocol               = "any"
    destination_vm_id      = vcd_vapp_vm.{{.VmName2}}.id
    destination_vm_nic_id  = 0
    destination_vm_ip_type = "assigned"
    destination_port       = "any"
    source_vm_id           = vcd_vapp_vm.{{.VmName1}}.id
    source_vm_nic_id       = 0
    source_vm_ip_type      = "NAT"
    source_port            = "any"
  }
}

resource "vcd_vapp_firewall_rules" "{{.ResourceName}}2" {
  org            = "{{.Org}}"
  vdc            = "{{.Vdc}}"
  vapp_id        = vcd_vapp.TestAccVcdVappFirewallRules_vapp.id
  default_action = "{{.DefaultAction}}"
  network_id     = vcd_vapp_org_network.vappAttachedNet.id

  rule {
    description      = "{{.Description1}}"
    policy           = "drop"
    protocol         = "udp"
    destination_port = "221"
    destination_ip   = "any"
    source_port      = "any"
    source_ip        = "10.10.0.10/24"
  }
}
`
