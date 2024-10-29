//go:build tm || ALL || functional

package vcd

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcdNsxtManager(t *testing.T) {
	preTestChecks(t)

	skipIfNotSysAdmin(t)
	skipIfNotTm(t)

	if !testConfig.Tm.CreateNsxtManager {
		t.Skipf("Skipping NSX-T Manager creation")
	}

	var params = StringMap{
		"Testname": t.Name(),
		"Username": testConfig.Tm.NsxtManagerUsername,
		"Password": testConfig.Tm.NsxtManagerPassword,
		"Url":      testConfig.Tm.NsxtManagerUrl,

		"Tags": "tm",
	}
	testParamsNotEmpty(t, params)

	configText1 := templateFill(testAccVcdNsxtManagerStep1, params)
	params["FuncName"] = t.Name() + "-step2"
	configText2 := templateFill(testAccVcdNsxtManagerStep2, params)
	params["FuncName"] = t.Name() + "-step3"
	configText3 := templateFill(testAccVcdNsxtManagerStep3DS, params)

	debugPrintf("#[DEBUG] CONFIGURATION step1: %s\n", configText1)
	debugPrintf("#[DEBUG] CONFIGURATION step2: %s\n", configText2)
	debugPrintf("#[DEBUG] CONFIGURATION step3: %s\n", configText3)
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
					resource.TestMatchResourceAttr("vcd_nsxt_manager.test", "id", regexp.MustCompile(`^urn:vcloud:nsxtmanager:`)),
					resource.TestMatchResourceAttr("vcd_nsxt_manager.test", "href", regexp.MustCompile(`api/admin/extension/nsxtManagers/`)),
					resource.TestCheckResourceAttr("vcd_nsxt_manager.test", "name", params["Testname"].(string)),
					resource.TestCheckResourceAttr("vcd_nsxt_manager.test", "description", "terraform test"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_manager.test", "status"),
					resource.TestCheckResourceAttr("vcd_nsxt_manager.test", "url", params["Url"].(string)),
				),
			},
			{
				Config: configText2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_manager.test", "id", regexp.MustCompile(`^urn:vcloud:nsxtmanager:`)),
					resource.TestMatchResourceAttr("vcd_nsxt_manager.test", "href", regexp.MustCompile(`api/admin/extension/nsxtManagers/`)),
					resource.TestCheckResourceAttr("vcd_nsxt_manager.test", "name", params["Testname"].(string)),
					resource.TestCheckResourceAttr("vcd_nsxt_manager.test", "description", ""),
				),
			},
			{
				Config: configText3,
				Check: resource.ComposeTestCheckFunc(
					resourceFieldsEqual("vcd_nsxt_manager.test", "data.vcd_nsxt_manager.test", []string{"%", "auto_trust_certificate", "password"}),
				),
			},
			{
				ResourceName:            "vcd_nsxt_manager.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateId:           params["Testname"].(string),
				ImportStateVerifyIgnore: []string{"auto_trust_certificate", "password"},
			},
		},
	})

	postTestChecks(t)
}

const testAccVcdNsxtManagerStep1 = `
resource "vcd_nsxt_manager" "test" {
  name                   = "{{.Testname}}"
  description            = "terraform test"
  username               = "{{.Username}}"
  password               = "{{.Password}}"
  url                    = "{{.Url}}"
  network_provider_scope = ""
  auto_trust_certificate = true
}
`
const testAccVcdNsxtManagerStep2 = `
resource "vcd_nsxt_manager" "test" {
  name                   = "{{.Testname}}"
  description            = ""
  username               = "{{.Username}}"
  password               = "{{.Password}}"
  url                    = "{{.Url}}"
  auto_trust_certificate = true
}
`

const testAccVcdNsxtManagerStep3DS = testAccVcdNsxtManagerStep1 + `
data "vcd_nsxt_manager" "test" {
  name = vcd_nsxt_manager.test.name
}
`
