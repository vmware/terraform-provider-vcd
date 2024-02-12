//go:build ALL || functional

package vcd

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcdVersion(t *testing.T) {
	//preTestChecks(t)
	//skipIfNotSysAdmin(t)

	vcdClient := createSystemTemporaryVCDConnection()
	currentVersion, err := vcdClient.Client.GetVcdShortVersion()
	if err != nil {
		t.Fatalf("could not get VCD version: %s", err)
	}

	apiVersion, err := vcdClient.VCDClient.Client.MaxSupportedVersion()
	if err != nil {
		t.Fatalf("could not get VCD API version: %s", err)
	}

	var params = StringMap{
		"Condition":      ">= 99.99.99",
		"FailIfNotMatch": "false",
	}
	testParamsNotEmpty(t, params)

	step1 := templateFill(testAccVcdVersion, params)

	params["FuncName"] = params["FuncName"].(string) + "-step2"
	params["FailIfNotMatch"] = "true"
	step2 := templateFill(testAccVcdVersion, params)

	params["FuncName"] = params["FuncName"].(string) + "-step3"
	params["Condition"] = "= " + currentVersion
	step3 := templateFill(testAccVcdVersion, params)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	debugPrintf("#[DEBUG] CONFIGURATION step1: %s", step1)

	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: step1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.vcd_version.version", "id", fmt.Sprintf("vcd_version=%s,condition=%s,fail_if_not_match=false", currentVersion, params["Condition"])),
					resource.TestCheckResourceAttr("data.vcd_version.version", "vcd_version", currentVersion),
					resource.TestCheckResourceAttr("data.vcd_version.version", "api_version", apiVersion),
					resource.TestCheckResourceAttr("data.vcd_version.version", "matches_condition", "false"),
				),
			},
			{
				Config:      step2,
				ExpectError: regexp.MustCompile(`the VCD version doesn't match the version constraint '>= 99.99.99'`),
			},
			{
				Config: step3,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.vcd_version.version", "id", fmt.Sprintf("vcd_version=%s,condition=%s,fail_if_not_match=true", currentVersion, params["Condition"])),
					resource.TestCheckResourceAttr("data.vcd_version.version", "vcd_version", currentVersion),
					resource.TestCheckResourceAttr("data.vcd_version.version", "api_version", apiVersion),
					resource.TestCheckResourceAttr("data.vcd_version.version", "matches_condition", "true"),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdVersion = `
data "vcd_version" "version" {
	name = "{{.Vcenter}}"
  }
`
