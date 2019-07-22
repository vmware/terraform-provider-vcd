// +build gateway lb lbVirtualServer ALL functional

package vcd

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/vmware/go-vcloud-director/v2/govcd"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccVcdLbVirtualServer(t *testing.T) {
	// String map to fill the template
	var params = StringMap{
		"Org":               testConfig.VCD.Org,
		"Vdc":               testConfig.VCD.Vdc,
		"EdgeGateway":       testConfig.Networking.EdgeGateway,
		"EdgeGatewayIp":     testConfig.Networking.ExternalIp,
		"VirtualServerName": t.Name(),
		"Tags":              "lb lbVirtualServer",
	}

	configText := templateFill(testAccVcdLbVirtualServer_Basic, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 0: %s", configText)

	params["FuncName"] = t.Name() + "-step1"
	configText1 := templateFill(testAccVcdLbVirtualServer_Empty, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		Providers:    testAccProviders,
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccCheckVcdLbVirtualServerDestroy(params["VirtualServerName"].(string)),
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_lb_virtual_server.http", "id", regexp.MustCompile(`^virtualServer-\d*$`)),
					resource.TestCheckResourceAttr("vcd_lb_virtual_server.http", "name", params["VirtualServerName"].(string)),
					resource.TestCheckResourceAttr("vcd_lb_virtual_server.http", "ip_address", params["EdgeGatewayIp"].(string)),
					resource.TestCheckResourceAttr("vcd_lb_virtual_server.http", "protocol", "http"),
					resource.TestCheckResourceAttr("vcd_lb_virtual_server.http", "port", "8888"),
					resource.TestMatchResourceAttr("vcd_lb_virtual_server.http", "app_profile_id", regexp.MustCompile(`^applicationProfile-\d*$`)),
					resource.TestMatchResourceAttr("vcd_lb_virtual_server.http", "server_pool_id", regexp.MustCompile(`^pool-\d*$`)),
					resource.TestCheckResourceAttr("vcd_lb_virtual_server.http", "app_rule_ids.#", "2"),
					resource.TestMatchResourceAttr("vcd_lb_virtual_server.http", "app_rule_ids.0", regexp.MustCompile(`^applicationRule-\d*$`)),
					resource.TestMatchResourceAttr("vcd_lb_virtual_server.http", "app_rule_ids.1", regexp.MustCompile(`^applicationRule-\d*$`)),

					// Data source
					resource.TestMatchResourceAttr("data.vcd_lb_virtual_server.http", "id", regexp.MustCompile(`^virtualServer-\d*$`)),
					resource.TestCheckResourceAttr("data.vcd_lb_virtual_server.http", "name", params["VirtualServerName"].(string)),
					resource.TestCheckResourceAttr("data.vcd_lb_virtual_server.http", "ip_address", params["EdgeGatewayIp"].(string)),
					resource.TestCheckResourceAttr("data.vcd_lb_virtual_server.http", "protocol", "http"),
					resource.TestCheckResourceAttr("data.vcd_lb_virtual_server.http", "port", "8888"),
					resource.TestMatchResourceAttr("data.vcd_lb_virtual_server.http", "app_profile_id", regexp.MustCompile(`^applicationProfile-\d*$`)),
					resource.TestMatchResourceAttr("data.vcd_lb_virtual_server.http", "server_pool_id", regexp.MustCompile(`^pool-\d*$`)),
					resource.TestCheckResourceAttr("data.vcd_lb_virtual_server.http", "app_rule_ids.#", "2"),
					resource.TestMatchResourceAttr("data.vcd_lb_virtual_server.http", "app_rule_ids.0", regexp.MustCompile(`^applicationRule-\d*$`)),
					resource.TestMatchResourceAttr("data.vcd_lb_virtual_server.http", "app_rule_ids.1", regexp.MustCompile(`^applicationRule-\d*$`)),
				),
			},

			// Check that import works
			resource.TestStep{
				ResourceName:      "vcd_lb_virtual_server.virtual-server-import",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdByOrgVdcEdge(testConfig, params["VirtualServerName"].(string)),
			},
			// Simple check - without app profile, rule ids or server pool
			resource.TestStep{
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_lb_virtual_server.http", "id", regexp.MustCompile(`^virtualServer-\d*$`)),
					resource.TestCheckResourceAttr("vcd_lb_virtual_server.http", "name", params["VirtualServerName"].(string)),
					resource.TestCheckResourceAttr("vcd_lb_virtual_server.http", "ip_address", params["EdgeGatewayIp"].(string)),
					resource.TestCheckResourceAttr("vcd_lb_virtual_server.http", "protocol", "http"),
					resource.TestCheckResourceAttr("vcd_lb_virtual_server.http", "port", "8888"),
					resource.TestCheckResourceAttr("vcd_lb_virtual_server.http", "app_rule_ids.#", "0"),
					resource.TestCheckNoResourceAttr("vcd_lb_virtual_server.http", "app_profile_id"),
					resource.TestCheckNoResourceAttr("vcd_lb_virtual_server.http", "server_pool_id"),

					// Data source
					resource.TestMatchResourceAttr("data.vcd_lb_virtual_server.http", "id", regexp.MustCompile(`^virtualServer-\d*$`)),
					resource.TestCheckResourceAttr("data.vcd_lb_virtual_server.http", "name", params["VirtualServerName"].(string)),
					resource.TestCheckResourceAttr("data.vcd_lb_virtual_server.http", "ip_address", params["EdgeGatewayIp"].(string)),
					resource.TestCheckResourceAttr("data.vcd_lb_virtual_server.http", "protocol", "http"),
					resource.TestCheckResourceAttr("data.vcd_lb_virtual_server.http", "port", "8888"),
					resource.TestCheckResourceAttr("data.vcd_lb_virtual_server.http", "app_rule_ids.#", "0"),
					resource.TestCheckNoResourceAttr("data.vcd_lb_virtual_server.http", "app_profile_id"),
					resource.TestCheckNoResourceAttr("data.vcd_lb_virtual_server.http", "server_pool_id"),
				),
			},
		},
	})
}

func testAccCheckVcdLbVirtualServerDestroy(virtualServerName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*VCDClient)

		edgeGateway, err := conn.GetEdgeGateway(testConfig.VCD.Org, testConfig.VCD.Vdc, testConfig.Networking.EdgeGateway)
		if err != nil {
			return fmt.Errorf(errorUnableToFindEdgeGateway, err)
		}

		virtualServer, err := edgeGateway.ReadLBVirtualServerByName(virtualServerName)

		if !strings.Contains(err.Error(), govcd.ErrorEntityNotFound.Error()) || virtualServer != nil {
			return fmt.Errorf("load balancer virtual server was not deleted: %s", err)
		}
		return nil
	}
}

const testAccVcdLbVirtualServer_Basic = `
resource "vcd_lb_virtual_server" "http" {
  org          = "{{.Org}}"
  vdc          = "{{.Vdc}}"
  edge_gateway = "{{.EdgeGateway}}"

  name       = "{{.VirtualServerName}}"
  ip_address = "{{.EdgeGatewayIp}}"
  protocol   = "http"
  port       = 8888

  app_profile_id = "${vcd_lb_app_profile.http.id}"
  server_pool_id = "${vcd_lb_server_pool.web-servers.id}"
  app_rule_ids   = ["${vcd_lb_app_rule.redirect.id}", "${vcd_lb_app_rule.language.id}"]
}

data "vcd_lb_virtual_server" "http" {
  org          = "{{.Org}}"
  vdc          = "{{.Vdc}}"
  edge_gateway = "{{.EdgeGateway}}"
  name         = "${vcd_lb_virtual_server.http.name}"
}

# Prerequisites to make a working load balancer
resource "vcd_lb_service_monitor" "monitor" {
  org          = "{{.Org}}"
  vdc          = "{{.Vdc}}"
  edge_gateway = "{{.EdgeGateway}}"

  name        = "http-monitor"
  interval    = "5"
  timeout     = "20"
  max_retries = "3"
  type        = "http"
  method      = "GET"
  url         = "/health"
  send        = "{\"key\": \"value\"}"
  extension = {
    content-type = "application/json"
    linespan     = ""
  }
}

resource "vcd_lb_server_pool" "web-servers" {
  org          = "{{.Org}}"
  vdc          = "{{.Vdc}}"
  edge_gateway = "{{.EdgeGateway}}"

  name                 = "web-servers"
  description          = "description"
  algorithm            = "httpheader"
  algorithm_parameters = "headerName=host"
  enable_transparency  = "true"

  monitor_id = "${vcd_lb_service_monitor.monitor.id}"

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
}

resource "vcd_lb_app_profile" "http" {
  org          = "{{.Org}}"
  vdc          = "{{.Vdc}}"
  edge_gateway = "{{.EdgeGateway}}"

  name = "http-app-profile"
  type = "http"
}

resource "vcd_lb_app_rule" "redirect" {
  org          = "{{.Org}}"
  vdc          = "{{.Vdc}}"
  edge_gateway = "{{.EdgeGateway}}"

  name   = "redirect"
  script = "acl vmware_page url_beg / vmware redirect location https://www.vmware.com/ ifvmware_page"
}

resource "vcd_lb_app_rule" "language" {
  org          = "{{.Org}}"
  vdc          = "{{.Vdc}}"
  edge_gateway = "{{.EdgeGateway}}"

  name   = "language"
  script = "acl hello payload(0,6) -m bin 48656c6c6f0a"
}
`

const testAccVcdLbVirtualServer_Empty = `
resource "vcd_lb_virtual_server" "http" {
  org          = "{{.Org}}"
  vdc          = "{{.Vdc}}"
  edge_gateway = "{{.EdgeGateway}}"

  name       = "{{.VirtualServerName}}"
  ip_address = "{{.EdgeGatewayIp}}"
  protocol   = "http"
  port       = 8888
}

data "vcd_lb_virtual_server" "http" {
  org          = "{{.Org}}"
  vdc          = "{{.Vdc}}"
  edge_gateway = "{{.EdgeGateway}}"
  name         = "${vcd_lb_virtual_server.http.name}"
}
`
