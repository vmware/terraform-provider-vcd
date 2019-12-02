// +build gateway lb lbAppProfile ALL functional

package vcd

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/vmware/go-vcloud-director/v2/govcd"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccVcdLBAppProfile(t *testing.T) {
	// String map to fill the template
	var params = StringMap{
		"Org":            testConfig.VCD.Org,
		"Vdc":            testConfig.VCD.Vdc,
		"EdgeGateway":    testConfig.Networking.EdgeGateway,
		"AppProfileName": t.Name(),
		"Type":           "tcp",
		"Tags":           "lb lbAppProfile",
	}

	configText := templateFill(testAccVcdLBAppProfile_TCP, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 0: %s", configText)

	params["FuncName"] = t.Name() + "-step1"
	params["Type"] = "udp"
	params["AppProfileName"] = t.Name() + "-step1"
	configTextStep1 := templateFill(testAccVcdLBAppProfile_UDP, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configTextStep1)

	params["FuncName"] = t.Name() + "-step2"
	params["Type"] = "http"
	params["AppProfileName"] = t.Name() + "-step2"
	configTextStep2 := templateFill(testAccVcdLBAppProfile_HTTP_Cookie, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configTextStep2)

	params["FuncName"] = t.Name() + "-step3"
	params["AppProfileName"] = t.Name() + "-step3"
	configTextStep3 := templateFill(testAccVcdLBAppProfile_HTTP_SourceIP, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 3: %s", configTextStep3)

	params["FuncName"] = t.Name() + "-step4"
	params["Type"] = "https"
	params["AppProfileName"] = t.Name() + "-step4"
	configTextStep4 := templateFill(testAccVcdLBAppProfile_HTTPS, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 4: %s", configTextStep4)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.ParallelTest(t, resource.TestCase{
		Providers:    testAccProviders,
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccCheckVcdLBAppProfileDestroy(params["AppProfileName"].(string)),
		Steps: []resource.TestStep{
			resource.TestStep{ // TCP
				Config: configText,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_lb_app_profile.test", "id", regexp.MustCompile(`^applicationProfile-\d*$`)),
					resource.TestCheckResourceAttr("vcd_lb_app_profile.test", "name", t.Name()),
					resource.TestCheckResourceAttr("vcd_lb_app_profile.test", "type", "tcp"),
					resource.TestCheckResourceAttr("vcd_lb_app_profile.test", "enable_ssl_passthrough", "false"),
					resource.TestCheckResourceAttr("vcd_lb_app_profile.test", "insert_x_forwarded_http_header", "false"),
					resource.TestCheckResourceAttr("vcd_lb_app_profile.test", "enable_pool_side_ssl", "false"),

					// Data source testing - it must expose all fields which resource has
					resource.TestMatchResourceAttr("data.vcd_lb_app_profile.test", "id", regexp.MustCompile(`^applicationProfile-\d*$`)),
					resource.TestCheckResourceAttr("data.vcd_lb_app_profile.test", "name", t.Name()),
					resource.TestCheckResourceAttr("data.vcd_lb_app_profile.test", "type", "tcp"),
					resource.TestCheckResourceAttr("data.vcd_lb_app_profile.test", "enable_ssl_passthrough", "false"),
					resource.TestCheckResourceAttr("data.vcd_lb_app_profile.test", "insert_x_forwarded_http_header", "false"),
					resource.TestCheckResourceAttr("data.vcd_lb_app_profile.test", "enable_pool_side_ssl", "false"),
				),
			},
			resource.TestStep{ // UDP
				Config: configTextStep1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_lb_app_profile.test", "id", regexp.MustCompile(`^applicationProfile-\d*$`)),
					resource.TestCheckResourceAttr("vcd_lb_app_profile.test", "name", t.Name()+"-step1"),
					resource.TestCheckResourceAttr("vcd_lb_app_profile.test", "type", "udp"),
					resource.TestCheckResourceAttr("vcd_lb_app_profile.test", "enable_ssl_passthrough", "true"),
					resource.TestCheckResourceAttr("vcd_lb_app_profile.test", "insert_x_forwarded_http_header", "true"),
					resource.TestCheckResourceAttr("vcd_lb_app_profile.test", "enable_pool_side_ssl", "true"),

					// Data source testing - it must expose all fields which resource has
					resource.TestMatchResourceAttr("data.vcd_lb_app_profile.test", "id", regexp.MustCompile(`^applicationProfile-\d*$`)),
					resource.TestCheckResourceAttr("data.vcd_lb_app_profile.test", "name", t.Name()+"-step1"),
					resource.TestCheckResourceAttr("data.vcd_lb_app_profile.test", "type", "udp"),
					resource.TestCheckResourceAttr("data.vcd_lb_app_profile.test", "enable_ssl_passthrough", "true"),
					resource.TestCheckResourceAttr("data.vcd_lb_app_profile.test", "insert_x_forwarded_http_header", "true"),
					resource.TestCheckResourceAttr("data.vcd_lb_app_profile.test", "enable_pool_side_ssl", "true"),
				),
			},
			resource.TestStep{ // HTTP - Cookie
				Config: configTextStep2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_lb_app_profile.test", "id", regexp.MustCompile(`^applicationProfile-\d*$`)),
					resource.TestCheckResourceAttr("vcd_lb_app_profile.test", "name", t.Name()+"-step2"),
					resource.TestCheckResourceAttr("vcd_lb_app_profile.test", "type", "http"),
					resource.TestCheckResourceAttr("vcd_lb_app_profile.test", "http_redirect_url", "/service-one"),
					resource.TestCheckResourceAttr("vcd_lb_app_profile.test", "persistence_mechanism", "cookie"),
					resource.TestCheckResourceAttr("vcd_lb_app_profile.test", "cookie_name", "JSESSIONID"),
					resource.TestCheckResourceAttr("vcd_lb_app_profile.test", "cookie_mode", "insert"),
					resource.TestCheckResourceAttr("vcd_lb_app_profile.test", "insert_x_forwarded_http_header", "true"),
					resource.TestCheckResourceAttr("vcd_lb_app_profile.test", "enable_ssl_passthrough", "false"),
					resource.TestCheckResourceAttr("vcd_lb_app_profile.test", "enable_pool_side_ssl", "false"),

					// Data source testing - it must expose all fields which resource has
					resource.TestMatchResourceAttr("data.vcd_lb_app_profile.test", "id", regexp.MustCompile(`^applicationProfile-\d*$`)),
					resource.TestCheckResourceAttr("data.vcd_lb_app_profile.test", "name", t.Name()+"-step2"),
					resource.TestCheckResourceAttr("data.vcd_lb_app_profile.test", "type", "http"),
					resource.TestCheckResourceAttr("data.vcd_lb_app_profile.test", "http_redirect_url", "/service-one"),
					resource.TestCheckResourceAttr("data.vcd_lb_app_profile.test", "persistence_mechanism", "cookie"),
					resource.TestCheckResourceAttr("data.vcd_lb_app_profile.test", "cookie_name", "JSESSIONID"),
					resource.TestCheckResourceAttr("data.vcd_lb_app_profile.test", "cookie_mode", "insert"),
					resource.TestCheckResourceAttr("data.vcd_lb_app_profile.test", "insert_x_forwarded_http_header", "true"),
					resource.TestCheckResourceAttr("data.vcd_lb_app_profile.test", "enable_ssl_passthrough", "false"),
					resource.TestCheckResourceAttr("data.vcd_lb_app_profile.test", "enable_pool_side_ssl", "false"),
				),
			},

			resource.TestStep{ // HTTP - Source IP
				Config: configTextStep3,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_lb_app_profile.test", "id", regexp.MustCompile(`^applicationProfile-\d*$`)),
					resource.TestCheckResourceAttr("vcd_lb_app_profile.test", "name", t.Name()+"-step3"),
					resource.TestCheckResourceAttr("vcd_lb_app_profile.test", "type", "http"),
					resource.TestCheckResourceAttr("vcd_lb_app_profile.test", "persistence_mechanism", "sourceip"),
					resource.TestCheckResourceAttr("vcd_lb_app_profile.test", "http_redirect_url", ""),
					resource.TestCheckResourceAttr("vcd_lb_app_profile.test", "expiration", "17"),
					resource.TestCheckResourceAttr("vcd_lb_app_profile.test", "insert_x_forwarded_http_header", "false"),
					resource.TestCheckResourceAttr("vcd_lb_app_profile.test", "enable_ssl_passthrough", "false"),
					resource.TestCheckResourceAttr("vcd_lb_app_profile.test", "enable_pool_side_ssl", "false"),

					// Data source testing - it must expose all fields which resource has
					resource.TestMatchResourceAttr("data.vcd_lb_app_profile.test", "id", regexp.MustCompile(`^applicationProfile-\d*$`)),
					resource.TestCheckResourceAttr("data.vcd_lb_app_profile.test", "name", t.Name()+"-step3"),
					resource.TestCheckResourceAttr("data.vcd_lb_app_profile.test", "type", "http"),
					resource.TestCheckResourceAttr("data.vcd_lb_app_profile.test", "persistence_mechanism", "sourceip"),
					resource.TestCheckResourceAttr("data.vcd_lb_app_profile.test", "http_redirect_url", ""),
					resource.TestCheckResourceAttr("data.vcd_lb_app_profile.test", "expiration", "17"),
					resource.TestCheckResourceAttr("data.vcd_lb_app_profile.test", "insert_x_forwarded_http_header", "false"),
					resource.TestCheckResourceAttr("data.vcd_lb_app_profile.test", "enable_ssl_passthrough", "false"),
					resource.TestCheckResourceAttr("data.vcd_lb_app_profile.test", "enable_pool_side_ssl", "false"),
				),
			},

			resource.TestStep{ // HTTPS
				Config: configTextStep4,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_lb_app_profile.test", "id", regexp.MustCompile(`^applicationProfile-\d*$`)),
					resource.TestCheckResourceAttr("vcd_lb_app_profile.test", "name", t.Name()+"-step4"),
					resource.TestCheckResourceAttr("vcd_lb_app_profile.test", "type", "https"),
					resource.TestCheckResourceAttr("vcd_lb_app_profile.test", "http_redirect_url", ""),
					resource.TestCheckResourceAttr("vcd_lb_app_profile.test", "insert_x_forwarded_http_header", "true"),
					resource.TestCheckResourceAttr("vcd_lb_app_profile.test", "enable_ssl_passthrough", "true"),
					resource.TestCheckResourceAttr("vcd_lb_app_profile.test", "enable_pool_side_ssl", "true"),
					resource.TestCheckResourceAttr("vcd_lb_app_profile.test", "expiration", "0"),

					// Data source testing - it must expose all fields which resource has
					resource.TestMatchResourceAttr("data.vcd_lb_app_profile.test", "id", regexp.MustCompile(`^applicationProfile-\d*$`)),
					resource.TestCheckResourceAttr("data.vcd_lb_app_profile.test", "name", t.Name()+"-step4"),
					resource.TestCheckResourceAttr("data.vcd_lb_app_profile.test", "type", "https"),
					resource.TestCheckResourceAttr("data.vcd_lb_app_profile.test", "http_redirect_url", ""),
					resource.TestCheckResourceAttr("data.vcd_lb_app_profile.test", "insert_x_forwarded_http_header", "true"),
					resource.TestCheckResourceAttr("data.vcd_lb_app_profile.test", "persistence_mechanism", "sourceip"),
					resource.TestCheckResourceAttr("data.vcd_lb_app_profile.test", "enable_ssl_passthrough", "true"),
					resource.TestCheckResourceAttr("data.vcd_lb_app_profile.test", "enable_pool_side_ssl", "true"),
					resource.TestCheckResourceAttr("data.vcd_lb_app_profile.test", "expiration", "0"),
				),
			},
			resource.TestStep{
				ResourceName:      "vcd_lb_app_profile.imported",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdEdgeGatewayObject(testConfig, testConfig.Networking.EdgeGateway, params["AppProfileName"].(string)),
			},
		},
	})
}

func testAccCheckVcdLBAppProfileDestroy(appProfileName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*VCDClient)

		edgeGateway, err := conn.GetEdgeGateway(testConfig.VCD.Org, testConfig.VCD.Vdc, testConfig.Networking.EdgeGateway)
		if err != nil {
			return fmt.Errorf(errorUnableToFindEdgeGateway, err)
		}

		monitor, err := edgeGateway.GetLbAppProfileByName(appProfileName)
		if !strings.Contains(err.Error(), govcd.ErrorEntityNotFound.Error()) ||
			monitor != nil {
			return fmt.Errorf("load balancer application profile was not deleted: %s", err)
		}
		return nil
	}
}

const testAccVcdLBAppProfile_TCP = `
resource "vcd_lb_app_profile" "test" {
	org          = "{{.Org}}"
	vdc          = "{{.Vdc}}"
	edge_gateway = "{{.EdgeGateway}}"
  
	name           = "{{.AppProfileName}}"
	type           = "{{.Type}}"

	enable_ssl_passthrough         = "false"
	insert_x_forwarded_http_header = "false"
	enable_pool_side_ssl           = "false"
}

data "vcd_lb_app_profile" "test" {
	org          = "{{.Org}}"
	vdc          = "{{.Vdc}}"
	edge_gateway = "{{.EdgeGateway}}"
	name         = vcd_lb_app_profile.test.name
}
`

const testAccVcdLBAppProfile_UDP = `
resource "vcd_lb_app_profile" "test" {
	org          = "{{.Org}}"
	vdc          = "{{.Vdc}}"
	edge_gateway = "{{.EdgeGateway}}"
  
	name           = "{{.AppProfileName}}"
	type           = "{{.Type}}"

	enable_ssl_passthrough         = "true"
	insert_x_forwarded_http_header = "true"
	enable_pool_side_ssl           = "true"
}

data "vcd_lb_app_profile" "test" {
	org          = "{{.Org}}"
	vdc          = "{{.Vdc}}"
	edge_gateway = "{{.EdgeGateway}}"
	name         = vcd_lb_app_profile.test.name
}
`

const testAccVcdLBAppProfile_HTTP_Cookie = `
resource "vcd_lb_app_profile" "test" {
  org          = "{{.Org}}"
  vdc          = "{{.Vdc}}"
  edge_gateway = "{{.EdgeGateway}}"

  name = "{{.AppProfileName}}"
  type = "{{.Type}}"

  http_redirect_url              = "/service-one"
  persistence_mechanism          = "cookie"
  cookie_name                    = "JSESSIONID"
  cookie_mode                    = "insert"
  insert_x_forwarded_http_header = "true"
}

data "vcd_lb_app_profile" "test" {
  org          = "{{.Org}}"
  vdc          = "{{.Vdc}}"
  edge_gateway = "{{.EdgeGateway}}"
  name         = vcd_lb_app_profile.test.name
}  
`

const testAccVcdLBAppProfile_HTTP_SourceIP = `
resource "vcd_lb_app_profile" "test" {
  org          = "{{.Org}}"
  vdc          = "{{.Vdc}}"
  edge_gateway = "{{.EdgeGateway}}"

  name = "{{.AppProfileName}}"
  type = "{{.Type}}"

  http_redirect_url     = ""
  persistence_mechanism = "sourceip"
  expiration = "17"
}

data "vcd_lb_app_profile" "test" {
  org          = "{{.Org}}"
  vdc          = "{{.Vdc}}"
  edge_gateway = "{{.EdgeGateway}}"
  name         = vcd_lb_app_profile.test.name
}
`

const testAccVcdLBAppProfile_HTTPS = `
resource "vcd_lb_app_profile" "test" {
  org          = "{{.Org}}"
  vdc          = "{{.Vdc}}"
  edge_gateway = "{{.EdgeGateway}}"

  name = "{{.AppProfileName}}"
  type = "{{.Type}}"

  persistence_mechanism = "sourceip"
  expiration = 0
  enable_ssl_passthrough         = "true"
  enable_pool_side_ssl           = "true"
  insert_x_forwarded_http_header = "true"
}

data "vcd_lb_app_profile" "test" {
  org          = "{{.Org}}"
  vdc          = "{{.Vdc}}"
  edge_gateway = "{{.EdgeGateway}}"
  name         = vcd_lb_app_profile.test.name
}
`
