//go:build tm || ALL || functional

package vcd

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcdTmVcenter(t *testing.T) {
	preTestChecks(t)

	skipIfNotSysAdmin(t)
	skipIfNotTm(t)

	if !testConfig.Tm.CreateVcenter {
		t.Skipf("Skipping vCenter creation")
	}

	var params = StringMap{
		"Org":             testConfig.Tm.Org,
		"VcenterUsername": testConfig.Tm.VcenterUsername,
		"VcenterPassword": testConfig.Tm.VcenterPassword,
		"VcenterUrl":      testConfig.Tm.VcenterUrl,

		"Testname": t.Name(),

		"Tags": "tm",
	}
	testParamsNotEmpty(t, params)

	configText1 := templateFill(testAccVcdTmVcenterStep1, params)
	params["FuncName"] = t.Name() + "-step2"
	configText2 := templateFill(testAccVcdTmVcenterStep2, params)

	params["FuncName"] = t.Name() + "-step3"
	configText3 := templateFill(testAccVcdTmVcenterStep3, params)

	params["FuncName"] = t.Name() + "-step4"
	configText4 := templateFill(testAccVcdTmVcenterStep4DS, params)

	debugPrintf("#[DEBUG] CONFIGURATION step1: %s\n", configText1)
	debugPrintf("#[DEBUG] CONFIGURATION step2: %s\n", configText2)
	debugPrintf("#[DEBUG] CONFIGURATION step3: %s\n", configText3)
	debugPrintf("#[DEBUG] CONFIGURATION step4: %s\n", configText4)
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
					resource.TestCheckResourceAttrSet("vcd_tm_vcenter.test", "id"),
					resource.TestCheckResourceAttr("vcd_tm_vcenter.test", "name", t.Name()),
					resource.TestCheckResourceAttr("vcd_tm_vcenter.test", "description", ""),
					resource.TestCheckResourceAttr("vcd_tm_vcenter.test", "is_enabled", "true"),
					resource.TestCheckResourceAttr("vcd_tm_vcenter.test", "has_proxy", "false"),

					resource.TestCheckResourceAttrSet("vcd_tm_vcenter.test", "cluster_health_status"),
					resource.TestCheckResourceAttrSet("vcd_tm_vcenter.test", "is_connected"),
					resource.TestCheckResourceAttrSet("vcd_tm_vcenter.test", "listener_state"),
					resource.TestCheckResourceAttrSet("vcd_tm_vcenter.test", "mode"),
					resource.TestCheckResourceAttrSet("vcd_tm_vcenter.test", "uuid"),
					resource.TestCheckResourceAttrSet("vcd_tm_vcenter.test", "version"),
				),
			},
			{
				Config: configText2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_tm_vcenter.test", "id"),
					resource.TestCheckResourceAttr("vcd_tm_vcenter.test", "name", t.Name()+"-rename"),
					resource.TestCheckResourceAttr("vcd_tm_vcenter.test", "is_enabled", "false"),
					resource.TestCheckResourceAttr("vcd_tm_vcenter.test", "description", "description from Terraform"),
					resource.TestCheckResourceAttr("vcd_tm_vcenter.test", "has_proxy", "false"),

					resource.TestCheckResourceAttrSet("vcd_tm_vcenter.test", "cluster_health_status"),
					resource.TestCheckResourceAttrSet("vcd_tm_vcenter.test", "is_connected"),
					resource.TestCheckResourceAttrSet("vcd_tm_vcenter.test", "listener_state"),
					resource.TestCheckResourceAttrSet("vcd_tm_vcenter.test", "mode"),
					resource.TestCheckResourceAttrSet("vcd_tm_vcenter.test", "uuid"),
					resource.TestCheckResourceAttrSet("vcd_tm_vcenter.test", "version"),
				),
			},
			{
				Config: configText3,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_tm_vcenter.test", "id"),
					resource.TestCheckResourceAttr("vcd_tm_vcenter.test", "name", t.Name()),
					resource.TestCheckResourceAttr("vcd_tm_vcenter.test", "description", ""),
					resource.TestCheckResourceAttr("vcd_tm_vcenter.test", "is_enabled", "true"),
					resource.TestCheckResourceAttr("vcd_tm_vcenter.test", "has_proxy", "false"),

					resource.TestCheckResourceAttrSet("vcd_tm_vcenter.test", "cluster_health_status"),
					resource.TestCheckResourceAttrSet("vcd_tm_vcenter.test", "is_connected"),
					resource.TestCheckResourceAttrSet("vcd_tm_vcenter.test", "listener_state"),
					resource.TestCheckResourceAttrSet("vcd_tm_vcenter.test", "mode"),
					resource.TestCheckResourceAttrSet("vcd_tm_vcenter.test", "uuid"),
					resource.TestCheckResourceAttrSet("vcd_tm_vcenter.test", "version"),
				),
			},
			{
				ResourceName:            "vcd_tm_vcenter.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateId:           params["Testname"].(string),
				ImportStateVerifyIgnore: []string{"password", "auto_trust_certificate"},
			},
			{
				Config: configText4,
				Check: resource.ComposeTestCheckFunc(
					resourceFieldsEqual("vcd_tm_vcenter.test", "vcd_tm_vcenter.test", []string{"%"}),
				),
			},
		},
	})

	postTestChecks(t)
}

const testAccVcdTmVcenterStep1 = `
resource "vcd_tm_vcenter" "test" {
  name                   = "{{.Testname}}"
  url                    = "{{.VcenterUrl}}"
  auto_trust_certificate = true
  username               = "{{.VcenterUsername}}"
  password               = "{{.VcenterPassword}}"
  is_enabled             = true
}
`

const testAccVcdTmVcenterStep2 = `
resource "vcd_tm_vcenter" "test" {
  name                   = "{{.Testname}}-rename"
  description            = "description from Terraform"
  auto_trust_certificate = true
  url                    = "{{.VcenterUrl}}"
  username               = "{{.VcenterUsername}}"
  password               = "{{.VcenterPassword}}"
  is_enabled             = false
}
`

const testAccVcdTmVcenterStep3 = `
resource "vcd_tm_vcenter" "test" {
  name                   = "{{.Testname}}"
  url                    = "{{.VcenterUrl}}"
  auto_trust_certificate = true
  username               = "{{.VcenterUsername}}"
  password               = "{{.VcenterPassword}}"
  is_enabled             = true
}
`

const testAccVcdTmVcenterStep4DS = testAccVcdTmVcenterStep3 + `
data "vcd_tm_vcenter" "test" {
  name = "{{.Testname}}"
}
`
