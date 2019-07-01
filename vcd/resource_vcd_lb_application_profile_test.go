// +build gateway lb lbAppProfile ALL functional

package vcd

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func TestAccVcdLBAppProfile(t *testing.T) {
	// String map to fill the template
	var params = StringMap{
		"Org":            testConfig.VCD.Org,
		"Vdc":            testConfig.VCD.Vdc,
		"EdgeGateway":    testConfig.Networking.EdgeGateway,
		"AppProfileName": t.Name(),
		"Type":           "TCP",
		"Tags":           "lb lbAppProfile",
	}

	configText := templateFill(testAccVcdLBAppProfile_TCP, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 0: %s", configText)

	params["FuncName"] = t.Name() + "-step1"
	params["Type"] = "UDP"
	configTextStep1 := templateFill(testAccVcdLBAppProfile_UDP, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configTextStep1)

	params["FuncName"] = t.Name() + "-step2"
	params["Type"] = "HTTP"
	configTextStep2 := templateFill(testAccVcdLBAppProfile_HTTP_Cookie, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configTextStep2)

	params["FuncName"] = t.Name() + "-step3"
	configTextStep3 := templateFill(testAccVcdLBAppProfile_HTTP_SourceIP, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 3: %s", configTextStep3)

	params["FuncName"] = t.Name() + "-step4"
	params["Type"] = "HTTPS"
	configTextStep4 := templateFill(testAccVcdLBAppProfile_HTTPS, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 4: %s", configTextStep3)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		Providers:    testAccProviders,
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccCheckVcdLBAppProfileDestroy(params["AppProfileName"].(string)),
		Steps: []resource.TestStep{
			resource.TestStep{ // TCP
				Config: configText,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_lb_application_profile.http-profile", "id", regexp.MustCompile(`^applicationProfile-\d*$`)),
					resource.TestCheckResourceAttr("vcd_lb_application_profile.http-profile", "name", params["AppProfileName"].(string)),
					resource.TestCheckResourceAttr("vcd_lb_application_profile.http-profile", "type", "TCP"),
					// resource.TestCheckResourceAttr("vcd_lb_server_pool.server-pool", "algorithm", "round-robin"),
					// resource.TestCheckResourceAttr("vcd_lb_server_pool.server-pool", "enable_transparency", "false"),
				),
			},
			resource.TestStep{ // UDP
				Config: configTextStep1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_lb_application_profile.http-profile", "id", regexp.MustCompile(`^applicationProfile-\d*$`)),
					resource.TestCheckResourceAttr("vcd_lb_application_profile.http-profile", "name", params["AppProfileName"].(string)),
					resource.TestCheckResourceAttr("vcd_lb_application_profile.http-profile", "type", "UDP"),
					// resource.TestCheckResourceAttr("vcd_lb_server_pool.server-pool", "algorithm", "round-robin"),
					// resource.TestCheckResourceAttr("vcd_lb_server_pool.server-pool", "enable_transparency", "false"),
				),
			},
			resource.TestStep{ // HTTP - Cookie
				Config: configTextStep2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_lb_application_profile.http-profile", "id", regexp.MustCompile(`^applicationProfile-\d*$`)),
					resource.TestCheckResourceAttr("vcd_lb_application_profile.http-profile", "name", params["AppProfileName"].(string)),
					resource.TestCheckResourceAttr("vcd_lb_application_profile.http-profile", "type", "HTTP"),
					resource.TestCheckResourceAttr("vcd_lb_application_profile.http-profile", "http_redirect_url", "/service-one"),
					resource.TestCheckResourceAttr("vcd_lb_application_profile.http-profile", "persistence_mechanism", "cookie"),
					resource.TestCheckResourceAttr("vcd_lb_application_profile.http-profile", "cookie_name", "JSESSIONID"),
					resource.TestCheckResourceAttr("vcd_lb_application_profile.http-profile", "cookie_mode", "insert"),
					resource.TestCheckResourceAttr("vcd_lb_application_profile.http-profile", "insert_x_forwarded_http_header", "true"),
					// resource.TestCheckResourceAttr("vcd_lb_server_pool.server-pool", "algorithm", "round-robin"),
					// resource.TestCheckResourceAttr("vcd_lb_server_pool.server-pool", "enable_transparency", "false"),
				),
			},

			resource.TestStep{ // HTTP - Source IP
				Config: configTextStep3,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_lb_application_profile.http-profile", "id", regexp.MustCompile(`^applicationProfile-\d*$`)),
					resource.TestCheckResourceAttr("vcd_lb_application_profile.http-profile", "name", params["AppProfileName"].(string)),
					resource.TestCheckResourceAttr("vcd_lb_application_profile.http-profile", "type", "HTTP"),
					resource.TestCheckResourceAttr("vcd_lb_application_profile.http-profile", "persistence_mechanism", "sourceip"),
					// resource.TestCheckResourceAttr("vcd_lb_application_profile.http-profile", "expiration", "17"),
					resource.TestCheckResourceAttr("vcd_lb_application_profile.http-profile", "insert_x_forwarded_http_header", "false"),
					resource.TestCheckResourceAttr("vcd_lb_application_profile.http-profile", "http_redirect_url", ""),

					// resource.TestCheckResourceAttr("vcd_lb_application_profile.http-profile", "cookie_name", "persistence-cookie"),
					// resource.TestCheckResourceAttr("vcd_lb_application_profile.http-profile", "cookie_mode", "insert"),
					// resource.TestCheckResourceAttr("vcd_lb_server_pool.server-pool", "algorithm", "round-robin"),
					// resource.TestCheckResourceAttr("vcd_lb_server_pool.server-pool", "enable_transparency", "false"),
				),
			},

			resource.TestStep{ // HTTPS
				Config: configTextStep4,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_lb_application_profile.http-profile", "id", regexp.MustCompile(`^applicationProfile-\d*$`)),
					resource.TestCheckResourceAttr("vcd_lb_application_profile.http-profile", "name", params["AppProfileName"].(string)),
					resource.TestCheckResourceAttr("vcd_lb_application_profile.http-profile", "type", "HTTPS"),
					resource.TestCheckResourceAttr("vcd_lb_application_profile.http-profile", "persistence_mechanism", "sourceip"),
					// resource.TestCheckResourceAttr("vcd_lb_application_profile.http-profile", "expiration", "13"),
					resource.TestCheckResourceAttr("vcd_lb_application_profile.http-profile", "insert_x_forwarded_http_header", "true"),
					resource.TestCheckResourceAttr("vcd_lb_application_profile.http-profile", "http_redirect_url", ""),
					resource.TestCheckResourceAttr("vcd_lb_application_profile.http-profile", "enable_ssl_passthrough", "true"),
					resource.TestCheckResourceAttr("vcd_lb_application_profile.http-profile", "enable_pool_side_ssl", "true"),

					// resource.TestCheckResourceAttr("vcd_lb_application_profile.http-profile", "cookie_name", "persistence-cookie"),
					// resource.TestCheckResourceAttr("vcd_lb_application_profile.http-profile", "cookie_mode", "insert"),
					// resource.TestCheckResourceAttr("vcd_lb_server_pool.server-pool", "algorithm", "round-robin"),
					// resource.TestCheckResourceAttr("vcd_lb_server_pool.server-pool", "enable_transparency", "false"),
				),
			},

			// http_redirect_url = "/service-one"
			// persistence_mechanism = "sourceip"
			// expiration = "20"

			// configTextStep1 attaches monitor_id, changes some member settings
			// resource.TestStep{
			// 	Config: configTextStep1,
			// 	Check: resource.ComposeAggregateTestCheckFunc(
			// 		resource.TestMatchResourceAttr("vcd_lb_server_pool.server-pool", "id", regexp.MustCompile(`^pool-\d*$`)),
			// 		resource.TestCheckResourceAttr("vcd_lb_server_pool.server-pool", "name", params["ServerPoolName"].(string)),
			// 		resource.TestCheckResourceAttr("vcd_lb_server_pool.server-pool", "algorithm", "httpheader"),
			// 		resource.TestCheckResourceAttr("vcd_lb_server_pool.server-pool", "algorithm_parameters", "headerName=host"),
			// 		resource.TestCheckResourceAttr("vcd_lb_server_pool.server-pool", "enable_transparency", "true"),
			// 		resource.TestMatchResourceAttr("vcd_lb_server_pool.server-pool", "monitor_id", regexp.MustCompile(`^monitor-\d*$`)),

			// 		// Data source testing - it must expose all fields which resource has
			// 		resource.TestMatchResourceAttr("data.vcd_lb_server_pool.ds-lb-server-pool", "id", regexp.MustCompile(`^pool-\d*$`)),
			// 		resource.TestCheckResourceAttr("data.vcd_lb_server_pool.ds-lb-server-pool", "name", params["ServerPoolName"].(string)),
			// 		resource.TestCheckResourceAttr("data.vcd_lb_server_pool.ds-lb-server-pool", "algorithm", "httpheader"),
			// 		resource.TestCheckResourceAttr("data.vcd_lb_server_pool.ds-lb-server-pool", "algorithm_parameters", "headerName=host"),
			// 		resource.TestCheckResourceAttr("data.vcd_lb_server_pool.ds-lb-server-pool", "enable_transparency", "true"),
			// 		resource.TestMatchResourceAttr("data.vcd_lb_server_pool.ds-lb-server-pool", "monitor_id", regexp.MustCompile(`^monitor-\d*$`)),
			// 	),
			// },
			// Check that import works
			resource.TestStep{
				ResourceName:      "vcd_lb_application_profile.imported",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdByOrgVdcEdge(testConfig, params["AppProfileName"].(string)),
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

		monitor, err := edgeGateway.ReadLBAppProfile(&types.LBAppProfile{Name: appProfileName})
		if !strings.Contains(err.Error(), "could not find load balancer application profile") ||
			monitor != nil {
			return fmt.Errorf("load balancer application profile was not deleted: %s", err)
		}
		return nil
	}
}

const testAccVcdLBAppProfile_TCP = `
resource "vcd_lb_application_profile" "http-profile" {
	org          = "{{.Org}}"
	vdc          = "{{.Vdc}}"
	edge_gateway = "{{.EdgeGateway}}"
  
	name           = "{{.AppProfileName}}"
	type           = "{{.Type}}"
}
`

const testAccVcdLBAppProfile_UDP = `
resource "vcd_lb_application_profile" "http-profile" {
	org          = "{{.Org}}"
	vdc          = "{{.Vdc}}"
	edge_gateway = "{{.EdgeGateway}}"
  
	name           = "{{.AppProfileName}}"
	type           = "{{.Type}}"
}
`

const testAccVcdLBAppProfile_HTTP_Cookie = `
resource "vcd_lb_application_profile" "http-profile" {
	org          = "{{.Org}}"
	vdc          = "{{.Vdc}}"
	edge_gateway = "{{.EdgeGateway}}"
  
	name           = "{{.AppProfileName}}"
	type           = "{{.Type}}"
	
	http_redirect_url = "/service-one"
	persistence_mechanism = "cookie"
	cookie_name = "JSESSIONID"
	cookie_mode = "insert"
	insert_x_forwarded_http_header = "true"
}
`

const testAccVcdLBAppProfile_HTTP_SourceIP = `
resource "vcd_lb_application_profile" "http-profile" {
	org          = "{{.Org}}"
	vdc          = "{{.Vdc}}"
	edge_gateway = "{{.EdgeGateway}}"
  
	name           = "{{.AppProfileName}}"
	type           = "{{.Type}}"
	
	http_redirect_url = ""
	persistence_mechanism = "sourceip"
}
`

const testAccVcdLBAppProfile_HTTPS = `
resource "vcd_lb_application_profile" "http-profile" {
	org          = "{{.Org}}"
	vdc          = "{{.Vdc}}"
	edge_gateway = "{{.EdgeGateway}}"
  
	name           = "{{.AppProfileName}}"
	type           = "{{.Type}}"
	
	persistence_mechanism = "sourceip"
	enable_ssl_passthrough = "true"
	enable_pool_side_ssl = "true"
	insert_x_forwarded_http_header = "true"
}
`
