//go:build gateway || lb || lbAppRule || ALL || functional
// +build gateway lb lbAppRule ALL functional

package vcd

import (
	"fmt"
	"html/template"
	"regexp"
	"strings"
	"testing"

	"github.com/vmware/go-vcloud-director/v2/govcd"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccVcdLBAppRule(t *testing.T) {
	preTestChecks(t)
	// The Script parameter must be sent as multiline string separated by newline (\n) characters.
	// Terraform has a native HEREDOC format for sending raw strings (with newline characters).
	// This variable is established for easier test comparison and is wrapped into HEREDOC syntax
	// in the `params` map using type `template.HTML` so that template engine does not
	// escape HEREDOC syntax <<- characters.
	MultiLineScript := `acl vmware_page url_beg / vmware redirect location https://www.vmware.com/ ifvmware_page
acl other_page2 url_beg / other2 redirect location https://www.other2.com/ ifother_page2
`

	// String map to fill the template
	var params = StringMap{
		"Org":              testConfig.VCD.Org,
		"Vdc":              testConfig.VCD.Vdc,
		"EdgeGateway":      testConfig.Networking.EdgeGateway,
		"AppRuleName":      t.Name(),
		"SingleLineScript": "acl vmware_page url_beg / vmware redirect location https://www.vmware.com/ ifvmware_page",
		"MultilineScript": template.HTML(`<<-EOT
` + MultiLineScript + `EOT`),
		"MultilineFailScript": template.HTML(`<<-EOT
			acl vmware_page url_beg / vmware redirect location https://www.vmware.com/ ifvmware_page
			acl other_page2 url_beg / other2 redirect location https://www.other2.com/ ifother_page2
			acl en req.fhdr(accept-language),language(es;fr;en) -m str en
			use_backend english if en
		EOT`),
		"Tags":     "lb lbAppRule",
		"SkipTest": "",
	}

	configText := templateFill(testAccVcdLBAppRule_OneLine, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 0: %s", configText)

	params["FuncName"] = t.Name() + "-step1"
	params["AppRuleName"] = t.Name() + "-step1"
	configText1 := templateFill(testAccVcdLBAppRule_MultiLine, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)

	params["FuncName"] = t.Name() + "-step3"
	// This test must fail with invalid rule script so we avoid running it in `make test-binary`
	params["SkipTest"] = "# skip-binary-test: it will fail on purpose"
	configText3 := templateFill(testAccVcdLBAppRule_FailMultiLine, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 3: %s", configText3)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckVcdLBAppRuleDestroy(params["AppRuleName"].(string)),
		Steps: []resource.TestStep{
			resource.TestStep{ // Single Line Script
				Config: configText,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_lb_app_rule.test", "id", regexp.MustCompile(`^applicationRule-\d*$`)),
					resource.TestCheckResourceAttr("vcd_lb_app_rule.test", "name", t.Name()),
					resource.TestCheckResourceAttr("vcd_lb_app_rule.test", "script", params["SingleLineScript"].(string)),

					// Data source testing - it must expose all fields which resource has
					resource.TestMatchResourceAttr("data.vcd_lb_app_rule.test", "id", regexp.MustCompile(`^applicationRule-\d*$`)),
					resource.TestCheckResourceAttr("data.vcd_lb_app_rule.test", "name", t.Name()),
					resource.TestCheckResourceAttr("data.vcd_lb_app_rule.test", "script", params["SingleLineScript"].(string)),
				),
			},

			resource.TestStep{ // Multi Line Script
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_lb_app_rule.test", "id", regexp.MustCompile(`^applicationRule-\d*$`)),
					resource.TestCheckResourceAttr("vcd_lb_app_rule.test", "name", t.Name()+"-step1"),
					resource.TestCheckResourceAttr("vcd_lb_app_rule.test", "script", MultiLineScript),

					// Data source testing - it must expose all fields which resource has
					resource.TestMatchResourceAttr("data.vcd_lb_app_rule.test", "id", regexp.MustCompile(`^applicationRule-\d*$`)),
					resource.TestCheckResourceAttr("data.vcd_lb_app_rule.test", "name", t.Name()+"-step1"),
					resource.TestCheckResourceAttr("data.vcd_lb_app_rule.test", "script", MultiLineScript),
				),
			},

			resource.TestStep{
				ResourceName:      "vcd_lb_app_rule.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdEdgeGatewayObject(testConfig, testConfig.Networking.EdgeGateway, params["AppRuleName"].(string)),
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
	postTestChecks(t)
}

func testAccCheckVcdLBAppRuleDestroy(appRuleName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*VCDClient)

		edgeGateway, err := conn.GetEdgeGateway(testConfig.VCD.Org, testConfig.VCD.Vdc, testConfig.Networking.EdgeGateway)
		if err != nil {
			return fmt.Errorf(errorUnableToFindEdgeGateway, err)
		}

		monitor, err := edgeGateway.GetLbAppRuleByName(appRuleName)
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
  script = "{{.SingleLineScript}}"
}

data "vcd_lb_app_rule" "test" {
  org          = "{{.Org}}"
  vdc          = "{{.Vdc}}"
  edge_gateway = "{{.EdgeGateway}}"
  name         = vcd_lb_app_rule.test.name
  depends_on   = [vcd_lb_app_rule.test]
}  
`

const testAccVcdLBAppRule_MultiLine = `
resource "vcd_lb_app_rule" "test" {
  org          = "{{.Org}}"
  vdc          = "{{.Vdc}}"
  edge_gateway = "{{.EdgeGateway}}"

  name   = "{{.AppRuleName}}"
  script = {{.MultilineScript}}
}

data "vcd_lb_app_rule" "test" {
  org          = "{{.Org}}"
  vdc          = "{{.Vdc}}"
  edge_gateway = "{{.EdgeGateway}}"
  name         = vcd_lb_app_rule.test.name
  depends_on   = [vcd_lb_app_rule.test]
} 
`

const testAccVcdLBAppRule_FailMultiLine = `
{{.SkipTest}}

resource "vcd_lb_app_rule" "test" {
  org          = "{{.Org}}"
  vdc          = "{{.Vdc}}"
  edge_gateway = "{{.EdgeGateway}}"

  name   = "{{.AppRuleName}}"
  script = {{.MultilineFailScript}}
}

data "vcd_lb_app_rule" "test" {
  org          = "{{.Org}}"
  vdc          = "{{.Vdc}}"
  edge_gateway = "{{.EdgeGateway}}"
  name         = vcd_lb_app_rule.test.name
  depends_on   = [vcd_lb_app_rule.test]
}
`
