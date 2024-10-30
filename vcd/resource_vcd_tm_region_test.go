//go:build tm || ALL || functional

package vcd

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcdTmRegion(t *testing.T) {
	preTestChecks(t)

	skipIfNotSysAdmin(t)
	skipIfNotTm(t)

	if !testConfig.Tm.CreateNsxtManager || !testConfig.Tm.CreateVcenter {
		t.Skipf("Skipping Region creation")
	}

	var params = StringMap{
		"Testname":            t.Name(),
		"NsxtManagerUsername": testConfig.Tm.NsxtManagerUsername,
		"NsxtManagerPassword": testConfig.Tm.NsxtManagerPassword,
		"NsxtManagerUrl":      testConfig.Tm.NsxtManagerUrl,

		"VcenterUsername":       testConfig.Tm.VcenterUsername,
		"VcenterPassword":       testConfig.Tm.VcenterPassword,
		"VcenterUrl":            testConfig.Tm.VcenterUrl,
		"VcenterStorageProfile": testConfig.Tm.VcenterStorageProfile,
		"VcenterSupervisor":     testConfig.Tm.VcenterSupervisor,

		"Tags": "tm",
	}
	testParamsNotEmpty(t, params)

	configText1 := templateFill(testAccVcdRegionStep1, params)
	params["FuncName"] = t.Name() + "-step2"
	// configText2 := templateFill(testAccVcdNsxtManagerStep2, params)
	// params["FuncName"] = t.Name() + "-step3"
	// configText3 := templateFill(testAccVcdNsxtManagerStep3DS, params)

	debugPrintf("#[DEBUG] CONFIGURATION step1: %s\n", configText1)
	// debugPrintf("#[DEBUG] CONFIGURATION step2: %s\n", configText2)
	// debugPrintf("#[DEBUG] CONFIGURATION step3: %s\n", configText3)
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
					resource.TestCheckResourceAttrSet("vcd_vcenter.test", "id"),
				),
			},
			// {
			// 	Config: configText2,
			// 	Check: resource.ComposeTestCheckFunc(
			// 		resource.TestMatchResourceAttr("vcd_nsxt_manager.test", "id", regexp.MustCompile(`^urn:vcloud:nsxtmanager:`)),
			// 		resource.TestMatchResourceAttr("vcd_nsxt_manager.test", "href", regexp.MustCompile(`api/admin/extension/nsxtManagers/`)),
			// 		resource.TestCheckResourceAttr("vcd_nsxt_manager.test", "name", params["Testname"].(string)),
			// 		resource.TestCheckResourceAttr("vcd_nsxt_manager.test", "description", ""),
			// 	),
			// },
			// {
			// 	Config: configText3,
			// 	Check: resource.ComposeTestCheckFunc(
			// 		resourceFieldsEqual("vcd_nsxt_manager.test", "data.vcd_nsxt_manager.test", []string{"%", "auto_trust_certificate", "password"}),
			// 	),
			// },
			// {
			// 	ResourceName:            "vcd_nsxt_manager.test",
			// 	ImportState:             true,
			// 	ImportStateVerify:       true,
			// 	ImportStateId:           params["Testname"].(string),
			// 	ImportStateVerifyIgnore: []string{"auto_trust_certificate", "password"},
			// },
		},
	})

	postTestChecks(t)
}

const testAccVcdRegionStep1 = `
resource "vcd_nsxt_manager" "test" {
  name                   = "{{.Testname}}"
  description            = "terraform test"
  username               = "{{.NsxtManagerUsername}}"
  password               = "{{.NsxtManagerPassword}}"
  url                    = "{{.NsxtManagerUrl}}"
  network_provider_scope = ""
  auto_trust_certificate = true
}

resource "vcd_vcenter" "test" {
  name                             = "{{.Testname}}"
  url                              = "{{.VcenterUrl}}"
  auto_trust_certificate           = true
  refresh_vcenter_on_read          = true
  refresh_policies_on_read = true
  username                         = "{{.VcenterUsername}}"
  password                         = "{{.VcenterPassword}}"
  is_enabled                       = true
}

data "vcd_tm_supervisor" "test" {
  name = "{{.VcenterSupervisor}}"

  depends_on = [vcd_vcenter.test]
}

resource "vcd_tm_region" "test" {
  name                 = "{{.Testname}}"
  is_enabled           = true
  nsx_manager_id       = vcd_nsxt_manager.test.id
  supervisor_ids       = [data.vcd_tm_supervisor.test.id]
  storage_policy_names = ["{{.VcenterStorageProfile}}"]
}
`
