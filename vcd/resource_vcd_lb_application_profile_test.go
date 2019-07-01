// +build gateway lb lbAppProfile ALL functional

package vcd

import (
	"fmt"
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
		// "Interval":           5,
		// "Timeout":            10,
		// "MaxRetries":         3,
		// "Method":             "POST",
		// "EnableTransparency": false,
		"Tags": "lb lbAppProfile",
	}

	configText := templateFill(testAccVcdLBAppProfile_Basic, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 0: %s", configText)

	// params["FuncName"] = t.Name() + "-step1"
	// params["EnableTransparency"] = true
	// configTextStep1 := templateFill(testAccVcdLbServerPool_Algorithm, params)
	// debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configTextStep1)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		Providers:    testAccProviders,
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccCheckVcdLBAppProfileDestroy(params["AppProfileName"].(string)),
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("vcd_lb_application_profile.http-profile", "name", params["AppProfileName"].(string)),
					// resource.TestMatchResourceAttr("vcd_lb_server_pool.server-pool", "id", regexp.MustCompile(`^pool-\d*$`)),
					// resource.TestCheckResourceAttr("vcd_lb_server_pool.server-pool", "algorithm", "round-robin"),
					// resource.TestCheckResourceAttr("vcd_lb_server_pool.server-pool", "enable_transparency", "false"),
				),
			},
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
			// resource.TestStep{
			// 	ResourceName:      "vcd_lb_server_pool.server-pool-import",
			// 	ImportState:       true,
			// 	ImportStateVerify: true,
			// 	ImportStateIdFunc: importStateIdByOrgVdcEdge(testConfig, params["ServerPoolName"].(string)),
			// },
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

const testAccVcdLBAppProfile_Basic = `
resource "vcd_lb_application_profile" "http-profile" {
	org          = "{{.Org}}"
	vdc          = "{{.Vdc}}"
	edge_gateway = "{{.EdgeGateway}}"
  
	name           = "{{.AppProfileName}}"
	type           = "TCP"
  
}
`
