// +build gateway firewall ALL functional

package vcd

import (
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/vmware/go-vcloud-director/v2/govcd"
)

func TestAccVcdNsxvEdgeFirewall(t *testing.T) {

	// String map to fill the template
	var params = StringMap{
		"Org":              testConfig.VCD.Org,
		"Vdc":              testConfig.VCD.Vdc,
		"EdgeGateway":      testConfig.Networking.EdgeGateway,
		"ExternalIp":       testConfig.Networking.ExternalIp,
		"InternalIp":       testConfig.Networking.InternalIp,
		"NetworkName":      testConfig.Networking.ExternalNetwork,
		"RouteNetworkName": "TestAccVcdVAppVmNet",
		"Catalog":          testSuiteCatalogName,
		"CatalogItem":      testSuiteCatalogOVAItem,
		"VappName":         vappName2,
		"VmName":           vmName,
		"Tags":             "gateway firewall",
	}

	configText := templateFill(testAccVcdEdgeFirewallRule0, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 0: %s", configText)

	params["FuncName"] = t.Name() + "-step1"
	configText1 := templateFill(testAccVcdEdgeFirewallRule1, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)

	params["FuncName"] = t.Name() + "-step2"
	configText2 := templateFill(testAccVcdEdgeFirewallRule2, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText2)

	params["FuncName"] = t.Name() + "-step3"
	configText3 := templateFill(testAccVcdEdgeFirewallRule3, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 3: %s", configText3)

	params["FuncName"] = t.Name() + "-step4"
	configText4 := templateFill(testAccVcdEdgeFirewallRule4, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 4: %s", configText4)

	params["FuncName"] = t.Name() + "-step5"
	configText5 := templateFill(testAccVcdEdgeFirewallRule5, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 5: %s", configText5)

	// -step6 is "import"

	params["FuncName"] = t.Name() + "-step7"
	configText7 := templateFill(testAccVcdEdgeFirewallRule6, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 7: %s", configText7)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.ParallelTest(t, resource.TestCase{
		Providers:    testAccProviders,
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccCheckVcdFirewallRuleDestroy("vcd_nsxv_firewall_rule.rule6"),
		Steps: []resource.TestStep{
			resource.TestStep{ // Step 0 - configuration only with ip_addresses
				Config: configText,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxv_firewall_rule.rule0", "id", regexp.MustCompile(`\d*`)),
					resource.TestMatchResourceAttr("vcd_nsxv_firewall_rule.rule0-2", "id", regexp.MustCompile(`\d*`)),
					// These two rules should go one after another because an explicit depends_on case is used
					// and above_rule_id field is not used
					firewallRuleOrderTest("vcd_nsxv_firewall_rule.rule0", "vcd_nsxv_firewall_rule.rule0-2"),
					// sleepTester(),
				),
			},
			resource.TestStep{ // Step 1 - configuration only with gateway_interfaces (internal, external)
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxv_firewall_rule.rule1", "id", regexp.MustCompile(`\d*`)),
					// sleepTester(),
				),
			},
			resource.TestStep{ // Step 2 - configuration only with gateway_interfaces (lookup)
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxv_firewall_rule.rule2", "id", regexp.MustCompile(`\d*`)),
					// sleepTester(),
				),
			},
			resource.TestStep{ // Step 3 - only org networks
				Config: configText3,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxv_firewall_rule.rule3", "id", regexp.MustCompile(`\d*`)),
					// sleepTester(),
				),
			},
			resource.TestStep{ // Step 4
				Config: configText4,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxv_firewall_rule.rule4", "id", regexp.MustCompile(`\d*`)),
					// sleepTester(),
				),
			},
			resource.TestStep{ // Step 5 -
				Config: configText5,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxv_firewall_rule.rule5", "id", regexp.MustCompile(`\d*`)),
					// sleepTester(),
				),
			},
			resource.TestStep{ // Step 6 - resource import
				ResourceName:      "vcd_nsxv_firewall_rule.imported",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdByResourceName("vcd_nsxv_firewall_rule.rule5"),
			},
			resource.TestStep{ // Step 7 - two rules - one above another
				Config: configText7,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxv_firewall_rule.rule6", "id", regexp.MustCompile(`\d*`)),
					resource.TestMatchResourceAttr("vcd_nsxv_firewall_rule.rule6-6", "id", regexp.MustCompile(`\d*`)),
					// vcd_nsxv_firewall_rule.rule6-6 should be above vcd_nsxv_firewall_rule.rule6
					// although it has depends_on = ["vcd_nsxv_firewall_rule.rule6"] which puts its
					// provisioning on later stage, but it uses the explicit positioning field
					// "above_rule_id =  vcd_nsxv_firewall_rule.rule6.id"
					firewallRuleOrderTest("vcd_nsxv_firewall_rule.rule6-6", "vcd_nsxv_firewall_rule.rule6"),
				),
			},
		},
	})
}

func sleepTester() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		fmt.Println("sleeping")
		time.Sleep(1 * time.Minute)
		return nil
	}
}

// firewallRuleOrderTest function accepts firewall rule HCL address (in format 'vcd_nsxv_firewall_rule.rule-name')
// and checks that its order is as specified (firstRule goes above secondRule)
func firewallRuleOrderTest(firstRule, secondRule string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*VCDClient)
		edgeGateway, err := conn.GetEdgeGateway(testConfig.VCD.Org, testConfig.VCD.Vdc, testConfig.Networking.EdgeGateway)
		if err != nil {
			return fmt.Errorf(errorUnableToFindEdgeGateway, err)
		}

		rule1, ok := s.RootModule().Resources[firstRule]
		if !ok {
			return fmt.Errorf("not found resource: %s", firstRule)
		}

		rule2, ok := s.RootModule().Resources[secondRule]
		if !ok {
			return fmt.Errorf("not found resource: %s", secondRule)
		}

		rule1Id := rule1.Primary.ID
		rule2Id := rule2.Primary.ID

		allFirewallRules, err := edgeGateway.GetAllNsxvFirewallRules()
		if err != nil {
			return fmt.Errorf("could not get all firewall rules: %s", err)
		}

		// This loop checks that at first rule1Id is found and then rule2Id is found in the ordered
		// list of firewall rules
		var foundSecondRule, foundFirstRule bool
		for _, rule := range allFirewallRules {
			if rule.ID == rule1Id {
				foundFirstRule = true
			}
			if foundFirstRule && rule.ID == rule2Id {
				foundSecondRule = true
			}
		}

		if !foundSecondRule || !foundFirstRule {
			return fmt.Errorf("incorrect rule order. %s is above %s. Should be reverse.", firstRule, secondRule)
		}

		return nil
	}
}

func testAccCheckVcdFirewallRuleDestroy(resource string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("not found resource: %s", resource)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set for %s resource", resource)
		}

		conn := testAccProvider.Meta().(*VCDClient)

		edgeGateway, err := conn.GetEdgeGateway(testConfig.VCD.Org, testConfig.VCD.Vdc, testConfig.Networking.EdgeGateway)
		if err != nil {
			return fmt.Errorf(errorUnableToFindEdgeGateway, err)
		}

		rule, err := edgeGateway.GetNsxvFirewallById(rs.Primary.ID)

		if !govcd.IsNotFound(err) || rule != nil {
			return fmt.Errorf("firewall rule (ID: %s) was not deleted: %s", rs.Primary.ID, err)
		}
		return nil
	}
}

const testAccVcdEdgeFirewallRule0 = `
resource "vcd_nsxv_firewall_rule" "rule0" {
	org          = "{{.Org}}"
	vdc          = "{{.Vdc}}"
	edge_gateway = "{{.EdgeGateway}}"
	name = "rule 30000"
	rule_tag = "30000"
	action = "deny"

	source {
		ip_addresses = ["any"]
	}
  
	destination {
		ip_addresses = ["192.168.1.110"]
	}
  
	#service {
	#	protocol = "any"
	#}
}

resource "vcd_nsxv_firewall_rule" "rule0-2" {
	org          = "{{.Org}}"
	vdc          = "{{.Vdc}}"
	edge_gateway = "{{.EdgeGateway}}"
	name = "rule 123123"
	action = "deny"

	source {
		ip_addresses = ["4.4.4.4"]
	}
  
	destination {
		ip_addresses = ["5.5.5.5"]
	}
  
	service {
		protocol = "any"
	}
	depends_on = ["vcd_nsxv_firewall_rule.rule0"]
}
`

const testAccVcdEdgeFirewallRule1 = `
resource "vcd_nsxv_firewall_rule" "rule1" {
	org          = "{{.Org}}"
	vdc          = "{{.Vdc}}"
	edge_gateway = "{{.EdgeGateway}}"
	name = "rule 30000"
	action = "deny"

	source {
		gateway_interfaces = ["internal"]
	}
  
	destination {
		gateway_interfaces = ["external"]
	}
	service {
		protocol = "tcp"
		port     = "443"
	}
  }
`

const testAccVcdEdgeFirewallRule2 = `
resource "vcd_nsxv_firewall_rule" "rule2" {
	org          = "{{.Org}}"
	vdc          = "{{.Vdc}}"
	edge_gateway = "{{.EdgeGateway}}"
	name = "rule 30000"
	action = "deny"

	source {
		gateway_interfaces = ["vse"]
	}
  
	destination {
		gateway_interfaces = ["{{.NetworkName}}"]
	}
	service {
		protocol    = "tcp"
		port        = "443-543"
		source_port = "2000-4000"
	}
  }
`

const testAccVcdEdgeFirewallRule3 = `
resource "vcd_nsxv_firewall_rule" "rule3" {
	org          = "{{.Org}}"
	vdc          = "{{.Vdc}}"
	edge_gateway = "{{.EdgeGateway}}"
	action = "deny"

	source {
		org_networks = ["${vcd_network_routed.test-routed[0].name}"]
	}
  
	destination {
		org_networks = ["${vcd_network_routed.test-routed[1].name}"]
	}
	service {
		protocol = "tcp"
		port     = "443"
	}
}
resource "vcd_network_routed" "test-routed" {
  count        = 2
  name         = "firewall-test-${count.index}"
  org          = "{{.Org}}"
  vdc          = "{{.Vdc}}"
  edge_gateway = "{{.EdgeGateway}}"
  gateway      = "10.201.${count.index}.1"
  netmask      = "255.255.255.0"

  static_ip_pool {
    start_address = "10.201.${count.index}.10"
    end_address   = "10.201.${count.index}.20"
  }
}
`

const testAccVcdEdgeFirewallRule4 = `
resource "vcd_nsxv_firewall_rule" "rule4" {
	org          = "{{.Org}}"
	vdc          = "{{.Vdc}}"
	edge_gateway = "{{.EdgeGateway}}"
	name = "rule 30000"
	action = "deny"

	source {
		gateway_interfaces = ["internal"]
	}
  
	destination {
		gateway_interfaces = ["external"]
	}

	service {
		protocol = "tcp"
		port     = "443"
	}
	
	service {
		protocol = "tcp"
		port     = "8443"
		source_port = "20000-40000"
	}

	service {
		protocol = "udp"
		port     = "10000"
	}

	service {
		protocol    = "udp"
		port        = "10000"
		source_port = "20000"
	}

	service {
		protocol = "icmp"
	}
  }
`

const testAccVcdEdgeFirewallRule5 = `
resource "vcd_nsxv_firewall_rule" "rule5" {
	org          = "{{.Org}}"
	vdc          = "{{.Vdc}}"
	edge_gateway = "{{.EdgeGateway}}"
	name = "rule 30000"
	action = "deny"

	source {
		gateway_interfaces = ["internal"]
	}
  
	destination {
		gateway_interfaces = ["external"]
	}

	service {
		protocol = "tcp"
	}

	service {
		protocol = "udp"
	}

  }
`

const testAccVcdEdgeFirewallRule6 = `
resource "vcd_nsxv_firewall_rule" "rule6" {
	org          = "{{.Org}}"
	vdc          = "{{.Vdc}}"
	edge_gateway = "{{.EdgeGateway}}"
	name = "below-rule"
	action = "accept"

	source {
		ip_addresses = ["10.10.10.0/24", "11.10.10.0/24"]
	}
  
	destination {
		ip_addresses = ["20.10.10.0/24", "21.10.10.0/24"]
	}

	service {
		protocol = "any"
	}
}

resource "vcd_nsxv_firewall_rule" "rule6-6" {
	org          = "{{.Org}}"
	vdc          = "{{.Vdc}}"
	edge_gateway = "{{.EdgeGateway}}"
	name = "above-rule"
	action = "accept"
	above_rule_id = "${vcd_nsxv_firewall_rule.rule6.id}"


	source {
		ip_addresses = ["10.10.10.0/24", "11.10.10.0/24"]
	}
  
	destination {
		ip_addresses = ["20.10.10.0/24", "21.10.10.0/24"]
	}

	service {
		protocol = "any"
	}

	depends_on = ["vcd_nsxv_firewall_rule.rule6"]
}
`

const testAccVcdEdgeFirewallRuleX = `
resource "vcd_network_routed" "{{.RouteNetworkName}}" {
  name         = "{{.RouteNetworkName}}"
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
  depends_on = ["vcd_network_routed.{{.RouteNetworkName}}"]
}

resource "vcd_vapp_vm" "{{.VmName}}" {
  org           = "{{.Org}}"
  vdc           = "{{.Vdc}}"
  vapp_name     = "${vcd_vapp.{{.VappName}}.name}"
  network_name  = "${vcd_network_routed.{{.RouteNetworkName}}.name}"
  name          = "{{.VmName}}"
  catalog_name  = "{{.Catalog}}"
  template_name = "{{.CatalogItem}}"
  memory        = 1024
  cpus          = 2
  cpu_cores     = 1
  ip            = "10.10.102.161"
}

resource "vcd_vapp_vm" "{{.VmName}}-2" {
	org           = "{{.Org}}"
	vdc           = "{{.Vdc}}"
	vapp_name     = "${vcd_vapp.{{.VappName}}.name}"
	network_name  = "${vcd_network_routed.{{.RouteNetworkName}}.name}"
	name          = "{{.VmName}}-2"
	catalog_name  = "{{.Catalog}}"
	template_name = "{{.CatalogItem}}"
	memory        = 1024
	cpus          = 2
	cpu_cores     = 1
	ip            = "10.10.102.162"
  }

resource "vcd_nsxv_firewall_rule" "rule4" {
	org          = "{{.Org}}"
	vdc          = "{{.Vdc}}"
	edge_gateway = "{{.EdgeGateway}}"
	action = "deny"

	source {
		virtual_machines_ids = [${vcd_vapp_vm.{{.VmName}}.id}]
	}
  
	destination {
		virtual_machines_ids = [${vcd_vapp_vm.{{.VmName}}-2.id}]
	}
	service {
		protocol    = "udp"
		port        = "8080"
		source_port = "1000-4000"
	}
  }
`
