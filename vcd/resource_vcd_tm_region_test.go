//go:build tm || ALL || functional

package vcd

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// TODO: TM: adjust the test for testing update operation once available
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
		"VcenterSupervisorZone": testConfig.Tm.VcenterSupervisorZone,

		"Tags": "tm",
	}
	testParamsNotEmpty(t, params)

	configText1 := templateFill(testAccVcdRegionStep1, params)
	params["FuncName"] = t.Name() + "-step1"
	configText2 := templateFill(testAccVcdRegionStep2DS, params)
	params["FuncName"] = t.Name() + "-step2"

	debugPrintf("#[DEBUG] CONFIGURATION step1: %s\n", configText1)
	debugPrintf("#[DEBUG] CONFIGURATION step2: %s\n", configText2)

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
					resource.TestCheckResourceAttrSet("vcd_tm_region.test", "id"),

					resource.TestCheckResourceAttrSet("data.vcd_tm_supervisor.test", "id"),
					resource.TestCheckResourceAttrSet("data.vcd_tm_supervisor_zone.test", "id"),
				),
			},
			{
				Config: configText2,
				Check: resource.ComposeTestCheckFunc(
					resourceFieldsEqual("vcd_tm_region.test", "data.vcd_tm_region.test", []string{
						"storage_policy_names.#", // TODO: TM: field is not populated on read
						"storage_policy_names.0", // TODO: TM: field is not populated on read
						"is_enabled",             // TODO: TM: field is not populated on read
					}),
				),
			},
			{
				ResourceName:      "vcd_tm_region.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     params["Testname"].(string),
				ImportStateVerifyIgnore: []string{
					"storage_policy_names.#", // TODO: TM: field is not populated on read
					"storage_policy_names.0", // TODO: TM: field is not populated on read
					"is_enabled",             // TODO: TM: field is not populated on read
				},
			},
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
  name                     = "{{.Testname}}"
  url                      = "{{.VcenterUrl}}"
  auto_trust_certificate   = true
  refresh_vcenter_on_read  = true
  refresh_policies_on_read = true
  username                 = "{{.VcenterUsername}}"
  password                 = "{{.VcenterPassword}}"
  is_enabled               = true
}

data "vcd_tm_supervisor" "test" {
  name = "{{.VcenterSupervisor}}"

  depends_on = [vcd_vcenter.test]
}

data "vcd_tm_supervisor_zone" "test" {
  supervisor_id = data.vcd_tm_supervisor.test.id
  name          = "{{.VcenterSupervisorZone}}"
}

resource "vcd_tm_region" "test" {
  name                 = "{{.Testname}}"
  is_enabled           = true
  nsx_manager_id       = vcd_nsxt_manager.test.id
  supervisor_ids       = [data.vcd_tm_supervisor.test.id]
  storage_policy_names = ["{{.VcenterStorageProfile}}"]
}
`

const testAccVcdRegionStep2DS = testAccVcdRegionStep1 + `
data "vcd_tm_region" "test" {
  name = vcd_tm_region.test.name
}
`
