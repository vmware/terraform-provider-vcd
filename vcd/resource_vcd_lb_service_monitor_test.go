// +build gateway lb lbServiceMonitor ALL functional

package vcd

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/vmware/go-vcloud-director/v2/types/v56"

	"github.com/hashicorp/terraform/terraform"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccVcdLbServiceMonitor(t *testing.T) {

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	// String map to fill the template
	var params = StringMap{
		"Org":                testConfig.VCD.Org,
		"Vdc":                testConfig.VCD.Vdc,
		"EdgeGateway":        testConfig.Networking.EdgeGateway,
		"ServiceMonitorName": t.Name(),
		"Interval":           5,
		"Timeout":            10,
		"MaxRetries":         3,
		"Method":             "POST",
		"Tags":               "lb lbServiceMonitor",
	}

	configText := templateFill(testAccVcdLbServiceMonitor_Basic, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 0: %s", configText)

	params["FuncName"] = t.Name() + "-step1"
	configTextStep1 := templateFill(testAccVcdLbServiceMonitor_Basic2, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configTextStep1)

	resource.Test(t, resource.TestCase{
		Providers:    testAccProviders,
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccCheckVcdLbServiceMonitorDestroy(params["ServiceMonitorName"].(string)),
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("vcd_lb_service_monitor.lb-service-monitor", "name", params["ServiceMonitorName"].(string)),
					resource.TestMatchResourceAttr("vcd_lb_service_monitor.lb-service-monitor", "id", regexp.MustCompile(`^monitor-\d*$`)),
					resource.TestCheckResourceAttr("vcd_lb_service_monitor.lb-service-monitor", "method", params["Method"].(string)),
					resource.TestCheckResourceAttr("vcd_lb_service_monitor.lb-service-monitor", "type", "http"),
					resource.TestCheckResourceAttr("vcd_lb_service_monitor.lb-service-monitor", "url", "health"),
					// Data source testing - it must expose all fields which resource has
					resource.TestCheckResourceAttr("data.vcd_lb_service_monitor.ds-lb-service-monitor", "interval", strconv.Itoa(params["Interval"].(int))),
					resource.TestCheckResourceAttr("data.vcd_lb_service_monitor.ds-lb-service-monitor", "timeout", strconv.Itoa(params["Timeout"].(int))),
					resource.TestCheckResourceAttr("data.vcd_lb_service_monitor.ds-lb-service-monitor", "max_retries", strconv.Itoa(params["MaxRetries"].(int))),
					resource.TestMatchResourceAttr("data.vcd_lb_service_monitor.ds-lb-service-monitor", "id", regexp.MustCompile(`^monitor-\d*$`)),
					resource.TestCheckResourceAttr("data.vcd_lb_service_monitor.ds-lb-service-monitor", "extension.content-type", "application/json"),
					resource.TestCheckResourceAttr("data.vcd_lb_service_monitor.ds-lb-service-monitor", "extension.no-body", ""),
					resource.TestCheckResourceAttr("data.vcd_lb_service_monitor.ds-lb-service-monitor", "method", params["Method"].(string)),
					resource.TestCheckResourceAttr("data.vcd_lb_service_monitor.ds-lb-service-monitor", "type", "http"),
					resource.TestCheckResourceAttr("data.vcd_lb_service_monitor.ds-lb-service-monitor", "send", "{\"key\": \"value\"}"),
					resource.TestCheckResourceAttr("data.vcd_lb_service_monitor.ds-lb-service-monitor", "expected", "HTTP/1.1"),
					resource.TestCheckResourceAttr("data.vcd_lb_service_monitor.ds-lb-service-monitor", "receive", "OK"),
					resource.TestCheckResourceAttr("data.vcd_lb_service_monitor.ds-lb-service-monitor", "url", "health"),
				),
			},
			resource.TestStep{
				Config: configTextStep1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("vcd_lb_service_monitor.lb-service-monitor", "name", params["ServiceMonitorName"].(string)),
					resource.TestMatchResourceAttr("vcd_lb_service_monitor.lb-service-monitor", "id", regexp.MustCompile(`^monitor-\d*$`)),
					resource.TestCheckResourceAttr("vcd_lb_service_monitor.lb-service-monitor", "type", "tcp"),
				),
			},
			// Check that import works
			resource.TestStep{
				ResourceName:      "vcd_lb_service_monitor.service-monitor-import",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdByOrgVdcEdge(testConfig, params["ServiceMonitorName"].(string)),
			},
		},
	})
}

func testAccCheckVcdLbServiceMonitorDestroy(serviceMonitorName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*VCDClient)

		edgeGateway, err := conn.GetEdgeGateway(testConfig.VCD.Org, testConfig.VCD.Vdc, testConfig.Networking.EdgeGateway)
		if err != nil {
			return fmt.Errorf(errorUnableToFindEdgeGateway, err)
		}

		monitor, err := edgeGateway.ReadLBServiceMonitor(&types.LBMonitor{Name: serviceMonitorName})
		if !strings.Contains(err.Error(), "could not find load balancer") || monitor != nil {
			return fmt.Errorf("load balancer service monitor was not deleted: %s", err)
		}
		return nil
	}
}

// importStateIdByOrgVdcEdge constructs an import path (ID in Terraform import terms) in the format of:
// organization.vdc.edge-gateway-nane.import-object-name (i.e. my-org.my-vdc.my-edge-gw.objectName) from TestConfig and
// object state.
func importStateIdByOrgVdcEdge(vcd TestConfig, objectName string) resource.ImportStateIdFunc {
	return func(*terraform.State) (string, error) {
		importId := testConfig.VCD.Org + "." + testConfig.VCD.Vdc + "." + testConfig.Networking.EdgeGateway + "." + objectName
		if testConfig.VCD.Org == "" || testConfig.VCD.Vdc == "" || testConfig.Networking.EdgeGateway == "" || objectName == "" {
			return "", fmt.Errorf("missing information to generate import path: %s", importId)
		}

		return importId, nil
	}
}

const testAccVcdLbServiceMonitor_Basic = `
resource "vcd_lb_service_monitor" "lb-service-monitor" {
  org          = "{{.Org}}"
  vdc          = "{{.Vdc}}"
  edge_gateway = "{{.EdgeGateway}}"

  name        = "{{.ServiceMonitorName}}"
  interval    = {{.Interval}}
  timeout     = {{.Timeout}}
  max_retries = {{.MaxRetries}}
  type        = "http"
  method      = "{{.Method}}"
  send        = "{\"key\": \"value\"}"
  expected    = "HTTP/1.1"
  receive     = "OK"
  url         = "health"

  extension = {
    "content-type" = "application/json"
    "no-body"      = ""
  }
}


data "vcd_lb_service_monitor" "ds-lb-service-monitor" {
  org          = "{{.Org}}"
  vdc          = "{{.Vdc}}"
  edge_gateway = "{{.EdgeGateway}}"
  name         = "${vcd_lb_service_monitor.lb-service-monitor.name}"
}
`

const testAccVcdLbServiceMonitor_Basic2 = `
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
