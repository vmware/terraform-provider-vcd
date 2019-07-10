// +build gateway lb lbAppRule ALL functional

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

func TestAccVcdLBAppRule(t *testing.T) {
	// String map to fill the template
	var params = StringMap{
		"Org":         testConfig.VCD.Org,
		"Vdc":         testConfig.VCD.Vdc,
		"EdgeGateway": testConfig.Networking.EdgeGateway,
		"AppRuleName": t.Name(),
		"ScriptLine1": "acl vmware_page url_beg / vmware redirect location https://www.vmware.com/ ifvmware_page",
		"ScriptLine2": "acl other_page2 url_beg / other2 redirect location https://www.other2.com/ ifother_page2",
		"ScriptLine3": "acl en req.fhdr(accept-language),language(es;fr;en) -m str en",
		"ScriptLine4": "use_backend english if en",
		"Tags":        "lb lbAppRule",
		"SkipTest":    "",
	}

	configText := templateFill(testAccVcdLBAppRule_OneLine, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 0: %s", configText)

	params["FuncName"] = t.Name() + "-step1"
	configText1 := templateFill(testAccVcdLBAppRule_MultiLine, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)

	params["FuncName"] = t.Name() + "-step3"
	// This test must fail with invalid rule script so we avoid running it in `make test-binary`
	params["SkipTest"] = "# skip-test"
	configText3 := templateFill(testAccVcdLBAppRule_FailMultiLine, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 3: %s", configText3)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		Providers:    testAccProviders,
		PreCheck:     func() { testAccPreCheck(t) },
		CheckDestroy: testAccCheckVcdLBAppRuleDestroy(params["AppRuleName"].(string)),
		Steps: []resource.TestStep{
			resource.TestStep{ // Single Line Script
				Config: configText,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_lb_app_rule.test", "id", regexp.MustCompile(`^applicationRule-\d*$`)),
					resource.TestCheckResourceAttr("vcd_lb_app_rule.test", "name", params["AppRuleName"].(string)),
					resource.TestCheckResourceAttr("vcd_lb_app_rule.test", "script.0", params["ScriptLine1"].(string)),

					// Data source testing - it must expose all fields which resource has
					resource.TestMatchResourceAttr("data.vcd_lb_app_rule.test", "id", regexp.MustCompile(`^applicationRule-\d*$`)),
					resource.TestCheckResourceAttr("data.vcd_lb_app_rule.test", "name", params["AppRuleName"].(string)),
					resource.TestCheckResourceAttr("data.vcd_lb_app_rule.test", "script.0", params["ScriptLine1"].(string)),
				),
			},

			resource.TestStep{ // Multi Line Script
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_lb_app_rule.test", "id", regexp.MustCompile(`^applicationRule-\d*$`)),
					resource.TestCheckResourceAttr("vcd_lb_app_rule.test", "name", params["AppRuleName"].(string)),
					resource.TestCheckResourceAttr("vcd_lb_app_rule.test", "script.0", params["ScriptLine1"].(string)),
					resource.TestCheckResourceAttr("vcd_lb_app_rule.test", "script.1", params["ScriptLine2"].(string)),

					// Data source testing - it must expose all fields which resource has
					resource.TestMatchResourceAttr("data.vcd_lb_app_rule.test", "id", regexp.MustCompile(`^applicationRule-\d*$`)),
					resource.TestCheckResourceAttr("data.vcd_lb_app_rule.test", "name", params["AppRuleName"].(string)),
					resource.TestCheckResourceAttr("data.vcd_lb_app_rule.test", "script.0", params["ScriptLine1"].(string)),
					resource.TestCheckResourceAttr("data.vcd_lb_app_rule.test", "script.1", params["ScriptLine2"].(string)),
				),
			},

			resource.TestStep{
				ResourceName:      "vcd_lb_app_rule.imported",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdByOrgVdcEdge(testConfig, params["AppRuleName"].(string)),
			},

			resource.TestStep{ // Multi Line Script with invalid rule
				Config:      configText3,
				ExpectError: regexp.MustCompile(`.*vShield Edge .* Not found pool name .* in rules.*`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr("vcd_lb_app_rule.test", "id"),
					resource.TestCheckNoResourceAttr("vcd_lb_app_rule.test", "name"),

					// Data source testing - it must expose all fields which resource has
					resource.TestCheckNoResourceAttr("data.vcd_lb_app_rule.test", "id"),
					resource.TestCheckNoResourceAttr("data.vcd_lb_app_rule.test", "name"),
				),
			},
		},
	})
}

func testAccCheckVcdLBAppRuleDestroy(appRuleName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*VCDClient)

		edgeGateway, err := conn.GetEdgeGateway(testConfig.VCD.Org, testConfig.VCD.Vdc, testConfig.Networking.EdgeGateway)
		if err != nil {
			return fmt.Errorf(errorUnableToFindEdgeGateway, err)
		}

		monitor, err := edgeGateway.ReadLBAppRuleByName(appRuleName)
		if !strings.Contains(err.Error(), govcd.ErrorEntityNotFound.Error()) ||
			monitor != nil {
			return fmt.Errorf("load balancer application rule was not deleted: %s", err)
		}
		return nil
	}
}

const testAccVcdLBAppRule_OneLine = `
resource "vcd_lb_app_rule" "test" {
	org          = "{{.Org}}"
	vdc          = "{{.Vdc}}"
	edge_gateway = "{{.EdgeGateway}}"
  
	name   = "{{.AppRuleName}}"
	script = ["{{.ScriptLine1}}"]
  }
  
  data "vcd_lb_app_rule" "test" {
	org          = "{{.Org}}"
	vdc          = "{{.Vdc}}"
	edge_gateway = "{{.EdgeGateway}}"
	name         = "${vcd_lb_app_rule.test.name}"
  }  
`

const testAccVcdLBAppRule_MultiLine = `
resource "vcd_lb_app_rule" "test" {
	org          = "{{.Org}}"
	vdc          = "{{.Vdc}}"
	edge_gateway = "{{.EdgeGateway}}"
  
	name   = "{{.AppRuleName}}"
	script = ["{{.ScriptLine1}}", "{{.ScriptLine2}}"]
  }
  
  data "vcd_lb_app_rule" "test" {
	org          = "{{.Org}}"
	vdc          = "{{.Vdc}}"
	edge_gateway = "{{.EdgeGateway}}"
	name         = "${vcd_lb_app_rule.test.name}"
  }  
`

const testAccVcdLBAppRule_FailMultiLine = `
{{.SkipTest}}

resource "vcd_lb_app_rule" "test" {
	org          = "{{.Org}}"
	vdc          = "{{.Vdc}}"
	edge_gateway = "{{.EdgeGateway}}"
  
	name   = "{{.AppRuleName}}"
	script = ["{{.ScriptLine1}}", "{{.ScriptLine2}}", "{{.ScriptLine3}}", "{{.ScriptLine4}}"]
  }
  
  data "vcd_lb_app_rule" "test" {
	org          = "{{.Org}}"
	vdc          = "{{.Vdc}}"
	edge_gateway = "{{.EdgeGateway}}"
	name         = "${vcd_lb_app_rule.test.name}"
  }  
`
