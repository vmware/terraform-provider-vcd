// +build gateway lb lbServerPool ALL functional

package vcd

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/vmware/go-vcloud-director/v2/types/v56"

	"github.com/hashicorp/terraform/terraform"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccVcdLbServerPool_Basic(t *testing.T) {
	//var vpnName string = t.Name()

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	// String map to fill the template
	var params = StringMap{
		"Org":            testConfig.VCD.Org,
		"Vdc":            testConfig.VCD.Vdc,
		"EdgeGateway":    testConfig.Networking.EdgeGateway,
		"ServerPoolName": t.Name(),
		"Interval":       5,
		"Timeout":        10,
		"MaxRetries":     3,
		"Method":         "POST",
		"Tags":           "lb lbServerPool",
	}

	configText := templateFill(testAccVcdLbServerPool_Basic, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 0: %s", configText)
	//
	//params["FuncName"] = t.Name() + "-step1"
	//configTextStep1 := templateFill(testAccVcdLbServiceMonitor_Basic2, params)
	//debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configTextStep1)

	resource.Test(t, resource.TestCase{
		Providers:    testAccProviders,
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccCheckVcdLbServerPoolDestroy(params["ServerPoolName"].(string)),
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vcd_lb_server_pool.server-pool", "name", params["ServerPoolName"].(string)),
					resource.TestMatchResourceAttr("vcd_lb_server_pool.server-pool", "id", regexp.MustCompile(`^pool-\d*$`)),
					resource.TestCheckResourceAttr("vcd_lb_server_pool.server-pool", "algorithm", "round-robin"),
					//resource.TestCheckResourceAttr("vcd_lb_server_pool.server-pool", "type", "http"),
					// Data source testing - it must expose all fields which resource has
					//resource.TestCheckResourceAttr("data.vcd_lb_service_monitor.ds-lb-service-monitor", "interval", strconv.Itoa(params["Interval"].(int))),
					//resource.TestCheckResourceAttr("data.vcd_lb_service_monitor.ds-lb-service-monitor", "timeout", strconv.Itoa(params["Timeout"].(int))),
					//resource.TestCheckResourceAttr("data.vcd_lb_service_monitor.ds-lb-service-monitor", "max_retries", strconv.Itoa(params["MaxRetries"].(int))),
					//resource.TestMatchResourceAttr("data.vcd_lb_service_monitor.ds-lb-service-monitor", "id", regexp.MustCompile(`^monitor-\d*$`)),
					//resource.TestCheckResourceAttr("data.vcd_lb_service_monitor.ds-lb-service-monitor", "extension.content-type", "application/json"),
					//resource.TestCheckResourceAttr("data.vcd_lb_service_monitor.ds-lb-service-monitor", "extension.no-body", ""),
					//resource.TestCheckResourceAttr("data.vcd_lb_service_monitor.ds-lb-service-monitor", "method", params["Method"].(string)),
					//resource.TestCheckResourceAttr("data.vcd_lb_service_monitor.ds-lb-service-monitor", "type", "http"),
					//resource.TestCheckResourceAttr("data.vcd_lb_service_monitor.ds-lb-service-monitor", "send", "{\"key\": \"value\"}"),
					//resource.TestCheckResourceAttr("data.vcd_lb_service_monitor.ds-lb-service-monitor", "expected", "HTTP/1.1"),
					//resource.TestCheckResourceAttr("data.vcd_lb_service_monitor.ds-lb-service-monitor", "receive", "OK"),
				),
			},
			//resource.TestStep{
			//	Config: configTextStep1,
			//	Check: resource.ComposeTestCheckFunc(
			//		resource.TestCheckResourceAttr("vcd_lb_service_monitor.lb-service-monitor", "name", params["ServiceMonitorName"].(string)),
			//		resource.TestMatchResourceAttr("vcd_lb_service_monitor.lb-service-monitor", "id", regexp.MustCompile(`^monitor-\d*$`)),
			//		resource.TestCheckResourceAttr("vcd_lb_service_monitor.lb-service-monitor", "type", "tcp"),
			//	),
			//},
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
		if !strings.Contains(err.Error(), "could not find load balancer pool") || monitor != nil {
			return fmt.Errorf("load balancer server pool was not deleted: %s", err)
		}
		return nil
	}
}

const testAccVcdLbServerPool_Basic = `
resource "vcd_lb_server_pool" "server-pool" {
  org                 = "my-org"
  vdc                 = "my-org-vdc"
  edge_gateway        = "my-edge-gw"

  name = "{{.ServerPoolName}}"
  description = "description"
  algorithm = "round-robin"

  #monitor_id = vcd_lb_service_monitor.lb-service-monitor.id

  member {
    condition = "enabled"
    name = "member1"
    ip_address = "1.1.1.1"
    port = 8443
    monitor_port = 8443
    weight = 1
    min_connections = 0
    max_connections = 100
  }

  member {
    condition = "enabled"
    name = "member2"
    ip_address = "2.2.2.2"
    port = 8443
    monitor_port = 8443
    weight = 1
    min_connections = 0
    max_connections = 100
  }

  member {
    condition = "enabled"
    name = "member3"
    ip_address = "4.4.4.4"
    port = 8443
    monitor_port = 8443
    weight = 1
    min_connections = 0
    max_connections = 50
  }

  // Because of non existent locking
  # depends_on = ["vcd_lb_service_monitor.lb-service-monitor"]
}
`

const TestAccVcdLbServerPool_Basic2 = `
resource "vcd_lb_service_monitor" "lb-service-monitor" {
  org          = "{{.Org}}"
  vdc          = "{{.Vdc}}"
  edge_gateway = "{{.EdgeGateway}}"

  name        = "{{.ServiceMonitorName}}"
  type        = "tcp"
  interval    = {{.Interval}}
  timeout     = {{.Timeout}}
  max_retries = {{.MaxRetries}}
}
`
