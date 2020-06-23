// +build functional vapp ALL

package vcd

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccVcdVappNatRules(t *testing.T) {
	if testConfig.Networking.EdgeGateway == "" {
		t.Skip("Variable testConfig.Networking.EdgeGateway must be configured")
		return
	}
	if testConfig.Networking.ExternalIp == "" {
		t.Skip("Variable networking.externalIp must be set to run DNAT tests")
		return
	}

	var (
		vmName1         = "TestAccVcdVappNatRulesVm1"
		vmName2         = "TestAccVcdVappNatRulesVm2"
		vmName3         = "TestAccVcdVappNatRulesVm3"
		vappName        = t.Name() + "_vapp"
		vappNetworkName = "vapp-routed-net"
	)

	var params = StringMap{
		"Org":           testConfig.VCD.Org,
		"Vdc":           testConfig.VCD.Vdc,
		"EdgeGateway":   testConfig.Networking.EdgeGateway,
		"DefaultAction": "drop",
		"ResourceName":  t.Name(),
		"FuncName":      t.Name(),
		"NetworkName":   "TestAccVcdVAppVmNet",
		"VappName":      vappName,
		"VmName1":       vmName1,
		"VmName2":       vmName2,
		"VmName3":       vmName3,
		"ExternalIp":    testConfig.Networking.ExternalIp,
		"Tags":          "vapp",
	}
	configText := templateFill(testAccVcdVappNatRules_rules, params)
	params["FuncName"] = t.Name() + "-step2"
	configTextForUpdate := templateFill(testAccVcdVappNatRules_rules_forUpdate, params)
	params["FuncName"] = t.Name() + "-step3"
	configTextForDelete := templateFill(testAccVcdVappNatRules_rules_forDelete, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configTextForUpdate)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configTextForDelete)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	resourceName := "vcd_vapp_nat_rules." + t.Name()
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdVappNatRulesExists(resourceName, 2),

					resource.TestCheckResourceAttr(resourceName, "rule.0.mapping_mode", "automatic"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.vm_nic_id", "0"),

					resource.TestCheckResourceAttr(resourceName, "rule.1.mapping_mode", "manual"),
					resource.TestCheckResourceAttr(resourceName, "rule.1.external_ip", "10.10.102.13"),
					resource.TestCheckResourceAttr(resourceName, "rule.1.vm_nic_id", "0"),

					testAccCheckVcdVappNatRulesExists(resourceName+"2", 2),
					resource.TestCheckResourceAttr(resourceName+"2", "enable_ip_masquerade", "true"),

					resource.TestCheckResourceAttr(resourceName+"2", "rule.0.external_port", "22"),
					resource.TestCheckResourceAttr(resourceName+"2", "rule.0.vm_nic_id", "0"),
					resource.TestCheckResourceAttr(resourceName+"2", "rule.0.protocol", "TCP_UDP"),
					resource.TestCheckResourceAttr(resourceName+"2", "rule.0.forward_to_port", "80"),

					resource.TestCheckResourceAttr(resourceName+"2", "rule.1.external_port", "11"),
					resource.TestCheckResourceAttr(resourceName+"2", "rule.1.vm_nic_id", "0"),
					resource.TestCheckResourceAttr(resourceName+"2", "rule.1.forward_to_port", "1112"),
					resource.TestCheckResourceAttr(resourceName+"2", "rule.1.protocol", "TCP"),
				),
			},
			// we can reuse importStateVappFirewallRuleObject as import is the same
			resource.TestStep{ // Step 1 - resource import
				ResourceName:            "vcd_vapp_nat_rules.imported",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdFunc:       importStateVappFirewallRuleObject(testConfig, vappName, vappNetworkName),
				ImportStateVerifyIgnore: []string{"enable_ip_masquerade", "network_id", "org", "vdc"},
			},
			resource.TestStep{ // Step 2 - resource import by ID
				ResourceName:            "vcd_vapp_nat_rules.imported2",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdFunc:       importStateVappFirewallRuleById(testConfig, resourceName),
				ImportStateVerifyIgnore: []string{"enable_ip_masquerade", "network_id", "org", "vdc"},
			},
			resource.TestStep{ // Step 3 - update
				Config: configTextForUpdate,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdVappNatRulesExists(resourceName, 2),
					resource.TestCheckResourceAttr(resourceName, "rule.0.mapping_mode", "manual"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.external_ip", "10.10.102.14"),
					resource.TestCheckResourceAttr(resourceName, "rule.0.vm_nic_id", "0"),

					resource.TestCheckResourceAttr(resourceName, "rule.1.mapping_mode", "automatic"),
					resource.TestCheckResourceAttr(resourceName, "rule.1.vm_nic_id", "0"),

					testAccCheckVcdVappNatRulesExists(resourceName+"2", 2),
					resource.TestCheckResourceAttr(resourceName+"2", "enable_ip_masquerade", "false"),

					resource.TestCheckResourceAttr(resourceName+"2", "rule.0.external_port", "11"),
					resource.TestCheckResourceAttr(resourceName+"2", "rule.0.vm_nic_id", "0"),
					resource.TestCheckResourceAttr(resourceName+"2", "rule.0.forward_to_port", "1112"),
					resource.TestCheckResourceAttr(resourceName+"2", "rule.0.protocol", "TCP"),

					resource.TestCheckResourceAttr(resourceName+"2", "rule.1.external_port", "222"),
					resource.TestCheckResourceAttr(resourceName+"2", "rule.1.vm_nic_id", "0"),
					resource.TestCheckResourceAttr(resourceName+"2", "rule.1.protocol", "UDP"),
					resource.TestCheckResourceAttr(resourceName+"2", "rule.1.forward_to_port", "800"),
				),
			},
			resource.TestStep{ // Step 3 - delete
				Config: configTextForDelete,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdVappNatRulesDeleted("vcd_vapp_org_network.vappAttachedNet"),
					testAccCheckVcdVappNatRulesDeleted("vcd_vapp_network.vappRoutedNet"),
				),
			},
		},
	})

}

func testAccCheckVcdVappNatRulesExists(n string, rulesCount int) resource.TestCheckFunc {
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

		if len(vapp_network.Configuration.Features.NatService.NatRule) == rulesCount {
			return nil
		}
		return fmt.Errorf("no rule with provided name is found")
	}
}

func testAccCheckVcdVappNatRulesDeleted(n string) resource.TestCheckFunc {
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

		vapp, err := vdc.GetVAppByName(rs.Primary.Attributes["vapp_name"], false)
		if err != nil {
			return err
		}

		vapp_network, err := vapp.GetVappNetworkById(rs.Primary.ID, false)
		if err != nil {
			return err
		}

		if len(vapp_network.Configuration.Features.NatService.NatRule) == 0 {
			return nil
		}
		return fmt.Errorf("no rule with provided network name is found")
	}
}

const testAccVcdVappNatRules_vappAndVm = `
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
  gateway          = "192.168.22.1"
  netmask          = "255.255.255.0"
  org_network_name = vcd_network_routed.network_routed.name

  nat_enabled = true

  static_ip_pool {
    start_address = "192.168.22.2"
    end_address   = "192.168.22.254"
  }
}

resource "vcd_vapp_org_network" "vappAttachedNet" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  vapp_name        = vcd_vapp.{{.VappName}}.name
  org_network_name = vcd_network_routed.network_routed.name
  is_fenced        = true

  nat_enabled = true
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

  power_on = false  

  os_type                        = "sles10_64Guest"
  hardware_version               = "vmx-11"
  expose_hardware_virtualization = true
  computer_name                  = "compName"
  network {
    type               = "vapp"
    name               = vcd_vapp_network.vappRoutedNet.name
    ip_allocation_mode = "MANUAL"
    ip                 = "192.168.22.11"
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

  power_on = false
  
  os_type          = "sles10_64Guest"
  hardware_version = "vmx-11"
  computer_name    = "compName"
  network {
    type               = "vapp"
    name               = vcd_vapp_network.vappRoutedNet.name
    ip_allocation_mode = "MANUAL"
    ip            = "192.168.22.12"
  }

  network {
    type               = "org"
    name               = vcd_vapp_org_network.vappAttachedNet.org_network_name
    ip_allocation_mode = "POOL"
  }
}
`

const testAccVcdVappNatRules_rules = testAccVcdVappNatRules_vappAndVm + `
resource "vcd_vapp_nat_rules" "{{.ResourceName}}" {
  org        = "{{.Org}}"
  vdc        = "{{.Vdc}}"
  vapp_id    = vcd_vapp.TestAccVcdVappNatRules_vapp.id
  network_id = vcd_vapp_network.vappRoutedNet.id
  nat_type   = "ipTranslation"

  rule {
    mapping_mode = "automatic" 
    vm_nic_id    = 0
    vm_id        = vcd_vapp_vm.{{.VmName1}}.id
  }

  rule {
    mapping_mode = "manual"
    external_ip  = "10.10.102.13"
    vm_nic_id    = 0
    vm_id        = vcd_vapp_vm.{{.VmName2}}.id
  }
}

resource "vcd_vapp_nat_rules" "{{.ResourceName}}2" {
  org        = "{{.Org}}"
  vdc        = "{{.Vdc}}"
  vapp_id    = vcd_vapp.TestAccVcdVappNatRules_vapp.id
  network_id = vcd_vapp_org_network.vappAttachedNet.id
  
  nat_type             = "portForwarding"
  enable_ip_masquerade = true

  rule {
    external_port        = 22
    vm_nic_id            = 0
    forward_to_port      = 80
    protocol             = "TCP_UDP"
    vm_id                = vcd_vapp_vm.{{.VmName1}}.id
  }

  rule {
    external_port   = 11
    vm_nic_id       = 0
    forward_to_port = 1112
    protocol        = "TCP"
    vm_id           = vcd_vapp_vm.{{.VmName2}}.id
  }
}
`

const testAccVcdVappNatRules_rules_forUpdate = testAccVcdVappNatRules_vappAndVm + `
resource "vcd_vapp_nat_rules" "{{.ResourceName}}" {
  org        = "{{.Org}}"
  vdc        = "{{.Vdc}}"
  vapp_id    = vcd_vapp.TestAccVcdVappNatRules_vapp.id
  network_id = vcd_vapp_network.vappRoutedNet.id
  nat_type   = "ipTranslation"

  rule {
    mapping_mode = "manual"
    external_ip  = "10.10.102.14"
    vm_nic_id    = 0
    vm_id        = vcd_vapp_vm.{{.VmName1}}.id
  }

  rule {
    mapping_mode = "automatic"
    vm_nic_id    = 0
    vm_id        = vcd_vapp_vm.{{.VmName2}}.id
  }
}

resource "vcd_vapp_nat_rules" "{{.ResourceName}}2" {
  org        = "{{.Org}}"
  vdc        = "{{.Vdc}}"
  vapp_id    = vcd_vapp.TestAccVcdVappNatRules_vapp.id
  network_id = vcd_vapp_org_network.vappAttachedNet.id
 
  nat_type             = "portForwarding"
  enable_ip_masquerade = false
 
  rule {
    external_port   = 11
    vm_nic_id       = 0
    forward_to_port = 1112
    protocol        = "TCP"
    vm_id           = vcd_vapp_vm.{{.VmName2}}.id
  }

  rule {
    external_port        = 222
    vm_nic_id            = 0
    forward_to_port      = 800
    protocol             = "UDP"
    vm_id                = vcd_vapp_vm.{{.VmName1}}.id
  }
}
`
const testAccVcdVappNatRules_rules_forDelete = testAccVcdVappNatRules_vappAndVm
