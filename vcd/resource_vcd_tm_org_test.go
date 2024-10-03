//go:build tm || ALL || functional

package vcd

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcdTmOrg(t *testing.T) {
	preTestChecks(t)

	skipIfNotSysAdmin(t)
	skipIfNotTm(t)

	var params = StringMap{
		"Org": testConfig.Tm.Org,

		"Tags": "tm",
	}
	testParamsNotEmpty(t, params)

	configText1 := templateFill(testAccVcdTmOrgStep1, params)
	params["FuncName"] = t.Name() + "-step2"
	configText2 := templateFill(testAccVcdTmOrgStep2, params)

	debugPrintf("#[DEBUG] CONFIGURATION step2: %s\n", configText1)
	debugPrintf("#[DEBUG] CONFIGURATION step1: %s\n", configText2)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: configText1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vcd_tm_org.test", "name", "tf-test"),
					resource.TestCheckResourceAttr("vcd_tm_org.test", "display_name", "tf-test"),
					resource.TestCheckResourceAttr("vcd_tm_org.test", "description", "terraform test"),
					resource.TestCheckResourceAttr("vcd_tm_org.test", "is_enabled", "true"),
					resource.TestCheckResourceAttr("vcd_tm_org.test", "is_subprovider", "false"),
				),
			},
			{
				Config: configText2,
				Check: resource.ComposeTestCheckFunc(
					resourceFieldsEqual("vcd_tm_org.test", "data.vcd_tm_org.test", nil),
				),
			},
			{
				ResourceName:      "vcd_tm_org.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     params["Org"].(string),
			},
		},
	})

	postTestChecks(t)
}

const testAccVcdTmOrgStep1 = `
resource "vcd_tm_org" "test" {
  name         = "tf-test"
  display_name = "tf-test"
  description  = "terraform test"
  is_enabled   = true
}
`

const testAccVcdTmOrgStep2 = testAccVcdTmOrgStep1 + `
data "vcd_tm_org" "test" {
  name = vcd_tm_org.test.name
}
`

func TestAccVcdTmOrgSubProvider(t *testing.T) {
	preTestChecks(t)

	skipIfNotSysAdmin(t)
	skipIfNotTm(t)

	var params = StringMap{
		"Org": testConfig.Tm.Org,

		"Tags": "tm",
	}
	testParamsNotEmpty(t, params)

	configText1 := templateFill(testAccVcdTmOrgSubproviderStep1, params)
	params["FuncName"] = t.Name() + "-step2"
	configText2 := templateFill(testAccVcdTmOrgSubproviderStep2, params)

	debugPrintf("#[DEBUG] CONFIGURATION step2: %s\n", configText1)
	debugPrintf("#[DEBUG] CONFIGURATION step1: %s\n", configText2)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: configText1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vcd_tm_org.test", "name", "tf-test"),
					resource.TestCheckResourceAttr("vcd_tm_org.test", "display_name", "tf-test"),
					resource.TestCheckResourceAttr("vcd_tm_org.test", "description", "terraform test"),
					resource.TestCheckResourceAttr("vcd_tm_org.test", "is_enabled", "true"),
					resource.TestCheckResourceAttr("vcd_tm_org.test", "is_subprovider", "true"),
				),
			},
			{
				Config: configText2,
				Check: resource.ComposeTestCheckFunc(
					resourceFieldsEqual("vcd_tm_org.test", "data.vcd_tm_org.test", nil),
				),
			},
			{
				ResourceName:      "vcd_tm_org.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     params["Org"].(string),
			},
		},
	})

	postTestChecks(t)
}

const testAccVcdTmOrgSubproviderStep1 = `
resource "vcd_tm_org" "test" {
  name           = "tf-test"
  display_name   = "tf-test"
  description    = "terraform test"
  is_enabled     = true
  is_subprovider = true
}
`

const testAccVcdTmOrgSubproviderStep2 = testAccVcdTmOrgSubproviderStep1 + `
data "vcd_tm_org" "test" {
  name = vcd_tm_org.test.name
}
`
