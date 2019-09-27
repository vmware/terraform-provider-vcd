// +build gateway firewall ALL functional

package vcd

import (
	"regexp"
	"testing"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
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

	configText := templateFill(testAccVcdEdgeFirewallRule1, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 0: %s", configText)

	params["FuncName"] = t.Name() + "-step1"
	configText1 := templateFill(testAccVcdEdgeFirewallRule2, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)

	params["FuncName"] = t.Name() + "-step2"
	configText2 := templateFill(testAccVcdEdgeFirewallRule3, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText2)

	params["FuncName"] = t.Name() + "-step3"
	configText3 := templateFill(testAccVcdEdgeFirewallRule4, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 3: %s", configText3)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.ParallelTest(t, resource.TestCase{
		Providers: testAccProviders,
		PreCheck:  func() { testAccPreCheck(t) },
		// CheckDestroy: testAccCheckVcdNatRuleDestroy("vcd_nsxv_dnat.test2"),
		Steps: []resource.TestStep{
			resource.TestStep{ // Step 0 - configuration only with ip_addresses
				Config: configText,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxv_firewall_rule.rule1", "id", regexp.MustCompile(`\d*`)),
				),
			},
			resource.TestStep{ // Step 1 - configuration only with gateway_interfaces (internal, external)
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxv_firewall_rule.rule2", "id", regexp.MustCompile(`\d*`)),
					// sleepTester(),
				),
			},
			resource.TestStep{ // Step 2 - configuration only with gateway_interfaces (lookup)
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxv_firewall_rule.rule3", "id", regexp.MustCompile(`\d*`)),
					// sleepTester(),
				),
			},
			resource.TestStep{ // Step 3 - only virtual machines
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxv_firewall_rule.rule4", "id", regexp.MustCompile(`\d*`)),
					// sleepTester(),
				),
			},
		},
	})
}

func sleepTester() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		time.Sleep(1 * time.Minute)
		return nil
	}
}

const testAccVcdEdgeFirewallRule1 = `
resource "vcd_nsxv_firewall_rule" "rule1" {
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
  
	service {
		protocol = "icmp"
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

const testAccVcdEdgeFirewallRule3 = `
resource "vcd_nsxv_firewall_rule" "rule3" {
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
		protocol = "tcp"
		port     = "443"
	}
  }
`

const testAccVcdEdgeFirewallRule4 = `
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
