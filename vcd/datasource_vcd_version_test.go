//go:build ALL || functional

package vcd

import (
	"fmt"
	"regexp"
	"strings"
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
		"SkipBinaryTest": " ",
		"Condition":      ">= 99.99.99",
		"FailIfNotMatch": "false",
	}
	testParamsNotEmpty(t, params)

	step1 := templateFill(testAccVcdVersion, params)
	debugPrintf("#[DEBUG] CONFIGURATION step1: %s", step1)

	params["FuncName"] = t.Name() + "-step2"
	params["FailIfNotMatch"] = "true"
	params["SkipBinaryTest"] = "# skip-binary-test - This one triggers an error"
	step2 := templateFill(testAccVcdVersion, params)
	debugPrintf("#[DEBUG] CONFIGURATION step2: %s", step2)

	params["FuncName"] = t.Name() + "-step3"
	params["Condition"] = "= " + currentVersion
	params["SkipBinaryTest"] = " "
	step3 := templateFill(testAccVcdVersion, params)
	debugPrintf("#[DEBUG] CONFIGURATION step3: %s", step3)

	params["FuncName"] = t.Name() + "-step4"
	versionTokens := strings.Split(currentVersion, ".")
	params["Condition"] = fmt.Sprintf("~> %s.%s", versionTokens[0], versionTokens[1])
	step4 := templateFill(testAccVcdVersion, params)
	debugPrintf("#[DEBUG] CONFIGURATION step4: %s", step4)

	params["FuncName"] = t.Name() + "-step5"
	params["Condition"] = "!= 10.3.0"
	step5 := templateFill(testAccVcdVersion, params)
	debugPrintf("#[DEBUG] CONFIGURATION step5: %s", step5)

	params["FuncName"] = t.Name() + "-step6"
	params["Condition"] = " " // Not used, but illustrates the point of this check
	params["FailIfNotMatch"] = " "
	step6 := templateFill(testAccVcdVersionWithoutArguments, params)
	debugPrintf("#[DEBUG] CONFIGURATION step6: %s", step6)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

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
				ExpectError: regexp.MustCompile(fmt.Sprintf(`the VCD version '%s' doesn't match the version constraint '>= 99.99.99'`, currentVersion)),
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
					resource.TestCheckResourceAttr("data.vcd_version.version", "id", fmt.Sprintf("vcd_version='%s',condition='~> %s.%s',fail_if_not_match='true'", currentVersion, versionTokens[0], versionTokens[1])),
					resource.TestCheckResourceAttr("data.vcd_version.version", "vcd_version", currentVersion),
					resource.TestCheckResourceAttr("data.vcd_version.version", "api_version", apiVersion),
					resource.TestCheckResourceAttr("data.vcd_version.version", "matches_condition", "true"),
				),
			},
			{
				Config: step5,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.vcd_version.version", "id", fmt.Sprintf("vcd_version='%s',condition='!= 10.3.0',fail_if_not_match='true'", currentVersion)),
					resource.TestCheckResourceAttr("data.vcd_version.version", "vcd_version", currentVersion),
					resource.TestCheckResourceAttr("data.vcd_version.version", "api_version", apiVersion),
					resource.TestCheckResourceAttr("data.vcd_version.version", "matches_condition", "true"),
				),
			},
			{
				Config: step6,
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
{{.SkipBinaryTest}}
data "vcd_version" "version" {
  condition         = "{{.Condition}}"
  fail_if_not_match = {{.FailIfNotMatch}}
}
`

const testAccVcdVersionWithoutArguments = `
data "vcd_version" "version" {
}
`
