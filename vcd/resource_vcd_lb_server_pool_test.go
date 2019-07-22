// +build gateway lb lbServerPool ALL functional

package vcd

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccVcdLbServerPool(t *testing.T) {
	// String map to fill the template
	var params = StringMap{
		"Org":                testConfig.VCD.Org,
		"Vdc":                testConfig.VCD.Vdc,
		"EdgeGateway":        testConfig.Networking.EdgeGateway,
		"ServerPoolName":     t.Name(),
		"Interval":           5,
		"Timeout":            10,
		"MaxRetries":         3,
		"Method":             "POST",
		"EnableTransparency": false,
		"Tags":               "lb lbServerPool",
	}

	configText := templateFill(testAccVcdLbServerPool_Basic, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 0: %s", configText)

	params["FuncName"] = t.Name() + "-step1"
	params["EnableTransparency"] = true
	configTextStep1 := templateFill(testAccVcdLbServerPool_Algorithm, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configTextStep1)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		Providers:    testAccProviders,
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccCheckVcdLbServerPoolDestroy(params["ServerPoolName"].(string)),
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("vcd_lb_server_pool.server-pool", "name", params["ServerPoolName"].(string)),
					resource.TestMatchResourceAttr("vcd_lb_server_pool.server-pool", "id", regexp.MustCompile(`^pool-\d*$`)),
					resource.TestCheckResourceAttr("vcd_lb_server_pool.server-pool", "algorithm", "round-robin"),
					resource.TestCheckResourceAttr("vcd_lb_server_pool.server-pool", "enable_transparency", "false"),

					// Member 1
					resource.TestMatchResourceAttr("vcd_lb_server_pool.server-pool", "member.0.id", regexp.MustCompile(`^member-\d*$`)),
					resource.TestCheckResourceAttr("vcd_lb_server_pool.server-pool", "member.0.condition", "enabled"),
					resource.TestCheckResourceAttr("vcd_lb_server_pool.server-pool", "member.0.name", "member1"),
					resource.TestCheckResourceAttr("vcd_lb_server_pool.server-pool", "member.0.ip_address", "1.1.1.1"),
					resource.TestCheckResourceAttr("vcd_lb_server_pool.server-pool", "member.0.port", "8443"),
					resource.TestCheckResourceAttr("vcd_lb_server_pool.server-pool", "member.0.monitor_port", "9000"),
					resource.TestCheckResourceAttr("vcd_lb_server_pool.server-pool", "member.0.weight", "1"),
					resource.TestCheckResourceAttr("vcd_lb_server_pool.server-pool", "member.0.min_connections", "0"),
					resource.TestCheckResourceAttr("vcd_lb_server_pool.server-pool", "member.0.max_connections", "100"),

					// Member 2
					resource.TestMatchResourceAttr("vcd_lb_server_pool.server-pool", "member.1.id", regexp.MustCompile(`^member-\d*$`)),
					resource.TestCheckResourceAttr("vcd_lb_server_pool.server-pool", "member.1.condition", "drain"),
					resource.TestCheckResourceAttr("vcd_lb_server_pool.server-pool", "member.1.name", "member2"),
					resource.TestCheckResourceAttr("vcd_lb_server_pool.server-pool", "member.1.ip_address", "2.2.2.2"),
					resource.TestCheckResourceAttr("vcd_lb_server_pool.server-pool", "member.1.port", "7000"),
					resource.TestCheckResourceAttr("vcd_lb_server_pool.server-pool", "member.1.monitor_port", "4000"),
					resource.TestCheckResourceAttr("vcd_lb_server_pool.server-pool", "member.1.weight", "2"),
					resource.TestCheckResourceAttr("vcd_lb_server_pool.server-pool", "member.1.min_connections", "6"),
					resource.TestCheckResourceAttr("vcd_lb_server_pool.server-pool", "member.1.max_connections", "8"),

					// Member 3
					resource.TestMatchResourceAttr("vcd_lb_server_pool.server-pool", "member.2.id", regexp.MustCompile(`^member-\d*$`)),
					resource.TestCheckResourceAttr("vcd_lb_server_pool.server-pool", "member.2.condition", "disabled"),
					resource.TestCheckResourceAttr("vcd_lb_server_pool.server-pool", "member.2.name", "member3"),
					resource.TestCheckResourceAttr("vcd_lb_server_pool.server-pool", "member.2.ip_address", "3.3.3.3"),
					resource.TestCheckResourceAttr("vcd_lb_server_pool.server-pool", "member.2.port", "3333"),
					resource.TestCheckResourceAttr("vcd_lb_server_pool.server-pool", "member.2.monitor_port", "4444"),
					resource.TestCheckResourceAttr("vcd_lb_server_pool.server-pool", "member.2.weight", "6"),
					resource.TestCheckResourceAttr("vcd_lb_server_pool.server-pool", "member.2.min_connections", "3"),
					resource.TestCheckResourceAttr("vcd_lb_server_pool.server-pool", "member.2.max_connections", "3"),

					// Member 4
					resource.TestMatchResourceAttr("vcd_lb_server_pool.server-pool", "member.3.id", regexp.MustCompile(`^member-\d*$`)),
					resource.TestCheckResourceAttr("vcd_lb_server_pool.server-pool", "member.3.condition", "disabled"),
					resource.TestCheckResourceAttr("vcd_lb_server_pool.server-pool", "member.3.name", "member4"),
					resource.TestCheckResourceAttr("vcd_lb_server_pool.server-pool", "member.3.ip_address", "4.4.4.4"),
					resource.TestCheckResourceAttr("vcd_lb_server_pool.server-pool", "member.3.port", "3333"),
					resource.TestCheckResourceAttr("vcd_lb_server_pool.server-pool", "member.3.monitor_port", "4444"),
					resource.TestCheckResourceAttr("vcd_lb_server_pool.server-pool", "member.3.weight", "6"),
					resource.TestCheckResourceAttr("vcd_lb_server_pool.server-pool", "member.3.min_connections", "0"),
					resource.TestCheckResourceAttr("vcd_lb_server_pool.server-pool", "member.3.max_connections", "0"),
				),
			},
			// configTextStep1 attaches monitor_id, changes some member settings
			resource.TestStep{
				Config: configTextStep1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_lb_server_pool.server-pool", "id", regexp.MustCompile(`^pool-\d*$`)),
					resource.TestCheckResourceAttr("vcd_lb_server_pool.server-pool", "name", params["ServerPoolName"].(string)),
					resource.TestCheckResourceAttr("vcd_lb_server_pool.server-pool", "algorithm", "httpheader"),
					resource.TestCheckResourceAttr("vcd_lb_server_pool.server-pool", "algorithm_parameters", "headerName=host"),
					resource.TestCheckResourceAttr("vcd_lb_server_pool.server-pool", "enable_transparency", "true"),
					resource.TestMatchResourceAttr("vcd_lb_server_pool.server-pool", "monitor_id", regexp.MustCompile(`^monitor-\d*$`)),

					// Member 1
					resource.TestMatchResourceAttr("vcd_lb_server_pool.server-pool", "member.0.id", regexp.MustCompile(`^member-\d*$`)),
					resource.TestCheckResourceAttr("vcd_lb_server_pool.server-pool", "member.0.condition", "drain"),
					resource.TestCheckResourceAttr("vcd_lb_server_pool.server-pool", "member.0.name", "member1"),
					resource.TestCheckResourceAttr("vcd_lb_server_pool.server-pool", "member.0.ip_address", "1.1.1.1"),
					resource.TestCheckResourceAttr("vcd_lb_server_pool.server-pool", "member.0.port", "8443"),
					resource.TestCheckResourceAttr("vcd_lb_server_pool.server-pool", "member.0.monitor_port", "9000"),
					resource.TestCheckResourceAttr("vcd_lb_server_pool.server-pool", "member.0.weight", "1"),
					resource.TestCheckResourceAttr("vcd_lb_server_pool.server-pool", "member.0.min_connections", "0"),
					resource.TestCheckResourceAttr("vcd_lb_server_pool.server-pool", "member.0.max_connections", "100"),

					// Member 2
					resource.TestMatchResourceAttr("vcd_lb_server_pool.server-pool", "member.1.id", regexp.MustCompile(`^member-\d*$`)),
					resource.TestCheckResourceAttr("vcd_lb_server_pool.server-pool", "member.1.condition", "drain"),
					resource.TestCheckResourceAttr("vcd_lb_server_pool.server-pool", "member.1.name", "member2"),
					resource.TestCheckResourceAttr("vcd_lb_server_pool.server-pool", "member.1.ip_address", "2.2.2.2"),
					resource.TestCheckResourceAttr("vcd_lb_server_pool.server-pool", "member.1.port", "7000"),
					resource.TestCheckResourceAttr("vcd_lb_server_pool.server-pool", "member.1.monitor_port", "4444"),
					resource.TestCheckResourceAttr("vcd_lb_server_pool.server-pool", "member.1.weight", "2"),
					resource.TestCheckResourceAttr("vcd_lb_server_pool.server-pool", "member.1.min_connections", "6"),
					resource.TestCheckResourceAttr("vcd_lb_server_pool.server-pool", "member.1.max_connections", "8"),

					// Member 3
					resource.TestMatchResourceAttr("vcd_lb_server_pool.server-pool", "member.2.id", regexp.MustCompile(`^member-\d*$`)),
					resource.TestCheckResourceAttr("vcd_lb_server_pool.server-pool", "member.2.condition", "enabled"),
					resource.TestCheckResourceAttr("vcd_lb_server_pool.server-pool", "member.2.name", "member3"),
					resource.TestCheckResourceAttr("vcd_lb_server_pool.server-pool", "member.2.ip_address", "3.3.3.3"),
					resource.TestCheckResourceAttr("vcd_lb_server_pool.server-pool", "member.2.port", "3333"),
					resource.TestCheckResourceAttr("vcd_lb_server_pool.server-pool", "member.2.monitor_port", "4444"),
					resource.TestCheckResourceAttr("vcd_lb_server_pool.server-pool", "member.2.weight", "6"),
					resource.TestCheckResourceAttr("vcd_lb_server_pool.server-pool", "member.2.min_connections", "3"),
					resource.TestCheckResourceAttr("vcd_lb_server_pool.server-pool", "member.2.max_connections", "3"),

					// Member 4
					resource.TestMatchResourceAttr("vcd_lb_server_pool.server-pool", "member.3.id", regexp.MustCompile(`^member-\d*$`)),
					resource.TestCheckResourceAttr("vcd_lb_server_pool.server-pool", "member.3.condition", "enabled"),
					resource.TestCheckResourceAttr("vcd_lb_server_pool.server-pool", "member.3.name", "member44"),
					resource.TestCheckResourceAttr("vcd_lb_server_pool.server-pool", "member.3.ip_address", "6.6.6.6"),
					resource.TestCheckResourceAttr("vcd_lb_server_pool.server-pool", "member.3.port", "33333"),
					resource.TestCheckResourceAttr("vcd_lb_server_pool.server-pool", "member.3.monitor_port", "44444"),
					resource.TestCheckResourceAttr("vcd_lb_server_pool.server-pool", "member.3.weight", "1"),
					resource.TestCheckResourceAttr("vcd_lb_server_pool.server-pool", "member.3.min_connections", "0"),
					resource.TestCheckResourceAttr("vcd_lb_server_pool.server-pool", "member.3.max_connections", "0"),
					//resource.TestCheckResourceAttr("vcd_lb_server_pool.server-pool", "type", "http"),

					// Data source testing - it must expose all fields which resource has
					resource.TestMatchResourceAttr("data.vcd_lb_server_pool.ds-lb-server-pool", "id", regexp.MustCompile(`^pool-\d*$`)),
					resource.TestCheckResourceAttr("data.vcd_lb_server_pool.ds-lb-server-pool", "name", params["ServerPoolName"].(string)),
					resource.TestCheckResourceAttr("data.vcd_lb_server_pool.ds-lb-server-pool", "algorithm", "httpheader"),
					resource.TestCheckResourceAttr("data.vcd_lb_server_pool.ds-lb-server-pool", "algorithm_parameters", "headerName=host"),
					resource.TestCheckResourceAttr("data.vcd_lb_server_pool.ds-lb-server-pool", "enable_transparency", "true"),
					resource.TestMatchResourceAttr("data.vcd_lb_server_pool.ds-lb-server-pool", "monitor_id", regexp.MustCompile(`^monitor-\d*$`)),

					// Member 1
					resource.TestCheckResourceAttr("data.vcd_lb_server_pool.ds-lb-server-pool", "member.0.condition", "drain"),
					resource.TestCheckResourceAttr("data.vcd_lb_server_pool.ds-lb-server-pool", "member.0.name", "member1"),
					resource.TestCheckResourceAttr("data.vcd_lb_server_pool.ds-lb-server-pool", "member.0.ip_address", "1.1.1.1"),
					resource.TestCheckResourceAttr("data.vcd_lb_server_pool.ds-lb-server-pool", "member.0.port", "8443"),
					resource.TestCheckResourceAttr("data.vcd_lb_server_pool.ds-lb-server-pool", "member.0.monitor_port", "9000"),
					resource.TestCheckResourceAttr("data.vcd_lb_server_pool.ds-lb-server-pool", "member.0.weight", "1"),
					resource.TestCheckResourceAttr("data.vcd_lb_server_pool.ds-lb-server-pool", "member.0.min_connections", "0"),
					resource.TestCheckResourceAttr("data.vcd_lb_server_pool.ds-lb-server-pool", "member.0.max_connections", "100"),

					// Member 2
					resource.TestCheckResourceAttr("data.vcd_lb_server_pool.ds-lb-server-pool", "member.1.condition", "drain"),
					resource.TestCheckResourceAttr("data.vcd_lb_server_pool.ds-lb-server-pool", "member.1.name", "member2"),
					resource.TestCheckResourceAttr("data.vcd_lb_server_pool.ds-lb-server-pool", "member.1.ip_address", "2.2.2.2"),
					resource.TestCheckResourceAttr("data.vcd_lb_server_pool.ds-lb-server-pool", "member.1.port", "7000"),
					resource.TestCheckResourceAttr("data.vcd_lb_server_pool.ds-lb-server-pool", "member.1.monitor_port", "4444"),
					resource.TestCheckResourceAttr("data.vcd_lb_server_pool.ds-lb-server-pool", "member.1.weight", "2"),
					resource.TestCheckResourceAttr("data.vcd_lb_server_pool.ds-lb-server-pool", "member.1.min_connections", "6"),
					resource.TestCheckResourceAttr("data.vcd_lb_server_pool.ds-lb-server-pool", "member.1.max_connections", "8"),

					// Member 3
					resource.TestCheckResourceAttr("data.vcd_lb_server_pool.ds-lb-server-pool", "member.2.condition", "enabled"),
					resource.TestCheckResourceAttr("data.vcd_lb_server_pool.ds-lb-server-pool", "member.2.name", "member3"),
					resource.TestCheckResourceAttr("data.vcd_lb_server_pool.ds-lb-server-pool", "member.2.ip_address", "3.3.3.3"),
					resource.TestCheckResourceAttr("data.vcd_lb_server_pool.ds-lb-server-pool", "member.2.port", "3333"),
					resource.TestCheckResourceAttr("data.vcd_lb_server_pool.ds-lb-server-pool", "member.2.monitor_port", "4444"),
					resource.TestCheckResourceAttr("data.vcd_lb_server_pool.ds-lb-server-pool", "member.2.weight", "6"),
					resource.TestCheckResourceAttr("data.vcd_lb_server_pool.ds-lb-server-pool", "member.2.min_connections", "3"),
					resource.TestCheckResourceAttr("data.vcd_lb_server_pool.ds-lb-server-pool", "member.2.max_connections", "3"),

					// Member 4
					resource.TestCheckResourceAttr("data.vcd_lb_server_pool.ds-lb-server-pool", "member.3.condition", "enabled"),
					resource.TestCheckResourceAttr("data.vcd_lb_server_pool.ds-lb-server-pool", "member.3.name", "member44"),
					resource.TestCheckResourceAttr("data.vcd_lb_server_pool.ds-lb-server-pool", "member.3.ip_address", "6.6.6.6"),
					resource.TestCheckResourceAttr("data.vcd_lb_server_pool.ds-lb-server-pool", "member.3.port", "33333"),
					resource.TestCheckResourceAttr("data.vcd_lb_server_pool.ds-lb-server-pool", "member.3.monitor_port", "44444"),
					resource.TestCheckResourceAttr("data.vcd_lb_server_pool.ds-lb-server-pool", "member.3.weight", "1"),
					resource.TestCheckResourceAttr("data.vcd_lb_server_pool.ds-lb-server-pool", "member.3.min_connections", "0"),
					resource.TestCheckResourceAttr("data.vcd_lb_server_pool.ds-lb-server-pool", "member.3.max_connections", "0"),
				),
			},
			// Check that import works
			resource.TestStep{
				ResourceName:      "vcd_lb_server_pool.server-pool-import",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdByOrgVdcEdge(testConfig, params["ServerPoolName"].(string)),
			},
		},
	})
}

func testAccCheckVcdLbServerPoolDestroy(serverPoolName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*VCDClient)

		edgeGateway, err := conn.GetEdgeGateway(testConfig.VCD.Org, testConfig.VCD.Vdc, testConfig.Networking.EdgeGateway)
		if err != nil {
			return fmt.Errorf(errorUnableToFindEdgeGateway, err)
		}

		monitor, err := edgeGateway.ReadLBServerPool(&types.LBPool{Name: serverPoolName})
		if !strings.Contains(err.Error(), govcd.ErrorEntityNotFound.Error()) || monitor != nil {
			return fmt.Errorf("load balancer server pool was not deleted: %s", err)
		}
		return nil
	}
}

const testAccVcdLbServerPool_Basic = `
resource "vcd_lb_server_pool" "server-pool" {
	org          = "{{.Org}}"
	vdc          = "{{.Vdc}}"
	edge_gateway = "{{.EdgeGateway}}"
  
	name                = "{{.ServerPoolName}}"
	algorithm           = "round-robin"
  
	member {
	  condition       = "enabled"
	  name            = "member1"
	  ip_address      = "1.1.1.1"
	  port            = 8443
	  monitor_port    = 9000
	  weight          = 1
	  min_connections = 0
	  max_connections = 100
	}
  
	member {
	  condition       = "drain"
	  name            = "member2"
	  ip_address      = "2.2.2.2"
	  port            = 7000
	  monitor_port    = 4000
	  weight          = 2
	  min_connections = 6
	  max_connections = 8
	}
  
	member {
	  condition       = "disabled"
	  name            = "member3"
	  ip_address      = "3.3.3.3"
	  port            = 3333
	  monitor_port    = 4444
	  weight          = 6
	  min_connections = 3
	  max_connections = 3
	}
  
	member {
	  condition    = "disabled"
	  name         = "member4"
	  ip_address   = "4.4.4.4"
	  port         = 3333
	  monitor_port = 4444
	  weight       = 6
	}
  }
  
  data "vcd_lb_server_pool" "ds-lb-server-pool" {
	org          = "{{.Org}}"
	vdc          = "{{.Vdc}}"
	edge_gateway = "{{.EdgeGateway}}"
	name         = "${vcd_lb_server_pool.server-pool.name}"
  }  
`

const testAccVcdLbServerPool_Algorithm = `
resource "vcd_lb_service_monitor" "test-monitor" {
	org          = "{{.Org}}"
	vdc          = "{{.Vdc}}"
	edge_gateway = "{{.EdgeGateway}}"
  
	name        = "test-monitor"
	type        = "tcp"
	interval    = 10
	timeout     = 15
	max_retries = 3
  }
  
  resource "vcd_lb_server_pool" "server-pool" {
	org          = "{{.Org}}"
	vdc          = "{{.Vdc}}"
	edge_gateway = "{{.EdgeGateway}}"
  
	name                 = "{{.ServerPoolName}}"
	description          = "description"
	algorithm            = "httpheader"
	algorithm_parameters = "headerName=host"
	enable_transparency  = "{{.EnableTransparency}}"
  
	monitor_id = "${vcd_lb_service_monitor.test-monitor.id}"
  
	member {
	  condition       = "drain"
	  name            = "member1"
	  ip_address      = "1.1.1.1"
	  port            = 8443
	  monitor_port    = 9000
	  weight          = 1
	  min_connections = 0
	  max_connections = 100
	}
  
	member {
	  condition       = "drain"
	  name            = "member2"
	  ip_address      = "2.2.2.2"
	  port            = 7000
	  monitor_port    = 4444
	  weight          = 2
	  min_connections = 6
	  max_connections = 8
	}
  
	member {
	  condition       = "enabled"
	  name            = "member3"
	  ip_address      = "3.3.3.3"
	  port            = 3333
	  monitor_port    = 4444
	  weight          = 6
	  min_connections = 3
	  max_connections = 3
	}
  
	member {
	  condition    = "enabled"
	  name         = "member44"
	  ip_address   = "6.6.6.6"
	  port         = 33333
	  monitor_port = 44444
	  weight       = 1
	}
  }
  
  data "vcd_lb_server_pool" "ds-lb-server-pool" {
	org          = "{{.Org}}"
	vdc          = "{{.Vdc}}"
	edge_gateway = "{{.EdgeGateway}}"
	name         = "${vcd_lb_server_pool.server-pool.name}"
  }  
`
