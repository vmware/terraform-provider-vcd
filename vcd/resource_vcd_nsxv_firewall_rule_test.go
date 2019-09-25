// +build gateway nat ALL functional

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
		"Org":         testConfig.VCD.Org,
		"Vdc":         testConfig.VCD.Vdc,
		"EdgeGateway": testConfig.Networking.EdgeGateway,
		"ExternalIp":  testConfig.Networking.ExternalIp,
		"InternalIp":  testConfig.Networking.InternalIp,
		"NetworkName": testConfig.Networking.ExternalNetwork,
		"Tags":        "egatewaydge nat",
	}

	configText := templateFill(testAccVcdEdgeFirewallRule, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 0: %s", configText)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.ParallelTest(t, resource.TestCase{
		Providers: testAccProviders,
		PreCheck:  func() { testAccPreCheck(t) },
		// CheckDestroy: testAccCheckVcdNatRuleDestroy("vcd_nsxv_dnat.test2"),
		Steps: []resource.TestStep{
			resource.TestStep{ // Step 0 - minimal configuration
				Config: configText,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxv_firewall_rule.rule1", "id", regexp.MustCompile(`\d*`)),
					sleepTester(),
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

const testAccVcdEdgeFirewallRule = `
resource "vcd_nsxv_firewall_rule" "rule1" {
  org          = "{{.Org}}"
  vdc          = "{{.Vdc}}"
  edge_gateway = "{{.EdgeGateway}}"
  name = "rule 40000"
  rule_tag = "40000"
  
  source {
	  exclude = false
	  network_ids = ["vse"]

  }

  destination {
	  
	  ips = ["any"]
  }

  service {
	  protocol = "tcp"
	  port     = "60"
  }
  depends_on = ["vcd_nsxv_firewall_rule.rule2"]

}

resource "vcd_nsxv_firewall_rule" "rule2" {
	org          = "{{.Org}}"
	vdc          = "{{.Vdc}}"
	edge_gateway = "{{.EdgeGateway}}"
	name = "rule 50000"
	rule_tag = "50000"

	source {
		exclude = false
		ips = ["192.168.1.1/24"]
	}
  
	destination {
		
		ips = ["any"]
	}
  
	service {
		protocol = "tcp"
		port     = "60"
	}
  }

`
