//go:build ALL || functional

package vcd

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcdVersion(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

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

	params["FuncName"] = t.Name() + "-step2"
	params["FailIfNotMatch"] = "true"
	step2 := templateFill(testAccVcdVersion, params)

	params["FuncName"] = t.Name() + "-step3"
	params["Condition"] = "= " + currentVersion
	step3 := templateFill(testAccVcdVersion, params)

	params["FuncName"] = t.Name() + "-step4"
	params["Condition"] = " " // Not used, but illustrates the point of this check
	params["FailIfNotMatch"] = " "
	step4 := templateFill(testAccVcdVersionWithoutArguments, params)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	debugPrintf("#[DEBUG] CONFIGURATION step1: %s", step1)
	debugPrintf("#[DEBUG] CONFIGURATION step2: %s", step2)
	debugPrintf("#[DEBUG] CONFIGURATION step3: %s", step3)
	debugPrintf("#[DEBUG] CONFIGURATION step4: %s", step4)

	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: step1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.vcd_version.version", "id", fmt.Sprintf("vcd_version='%s',condition='>= 99.99.99',fail_if_not_match='false'", currentVersion)),
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
					resource.TestCheckResourceAttr("data.vcd_version.version", "id", fmt.Sprintf("vcd_version='%s',condition='= %s',fail_if_not_match='true'", currentVersion, currentVersion)),
					resource.TestCheckResourceAttr("data.vcd_version.version", "vcd_version", currentVersion),
					resource.TestCheckResourceAttr("data.vcd_version.version", "api_version", apiVersion),
					resource.TestCheckResourceAttr("data.vcd_version.version", "matches_condition", "true"),
				),
			},
			{
				Config: step4,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.vcd_version.version", "id", fmt.Sprintf("vcd_version='%s',condition='',fail_if_not_match='false'", currentVersion)),
					resource.TestCheckResourceAttr("data.vcd_version.version", "vcd_version", currentVersion),
					resource.TestCheckResourceAttr("data.vcd_version.version", "api_version", apiVersion),
					resource.TestCheckNoResourceAttr("data.vcd_version.version", "matches_condition"),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdVersion = `
data "vcd_version" "version" {
  condition         = "{{.Condition}}"
  fail_if_not_match = {{.FailIfNotMatch}}
}
`

const testAccVcdVersionWithoutArguments = `
data "vcd_version" "version" {
}
`
