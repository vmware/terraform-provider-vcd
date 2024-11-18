//go:build tm || ALL || functional

package vcd

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// TODO: TM: the test has an update, but it just recreates the resource behind the scenes now
// as the API does not support update yet
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
	configText2 := templateFill(testAccVcdRegionStep2, params)
	params["FuncName"] = t.Name() + "-step2"
	configText3 := templateFill(testAccVcdRegionStep3DS, params)
	params["FuncName"] = t.Name() + "-step3"

	debugPrintf("#[DEBUG] CONFIGURATION step1: %s\n", configText1)
	debugPrintf("#[DEBUG] CONFIGURATION step2: %s\n", configText2)
	debugPrintf("#[DEBUG] CONFIGURATION step3: %s\n", configText3)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	cachedRegionId := &testCachedFieldValue{}
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: configText1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_manager.test", "id", regexp.MustCompile(`^urn:vcloud:nsxtmanager:`)),
					resource.TestCheckResourceAttrSet("vcd_vcenter.test", "id"),
					resource.TestCheckResourceAttrSet("vcd_tm_region.test", "id"),
					cachedRegionId.cacheTestResourceFieldValue("vcd_tm_region.test", "id"),
					resource.TestCheckResourceAttr("vcd_tm_region.test", "is_enabled", "true"),
					resource.TestCheckResourceAttr("vcd_tm_region.test", "description", "Terraform description"),
					resource.TestCheckResourceAttrSet("vcd_tm_region.test", "cpu_capacity_mhz"),
					resource.TestCheckResourceAttrSet("vcd_tm_region.test", "cpu_reservation_capacity_mhz"),
					resource.TestCheckResourceAttrSet("vcd_tm_region.test", "memory_capacity_mib"),
					resource.TestCheckResourceAttrSet("vcd_tm_region.test", "memory_reservation_capacity_mib"),
					resource.TestCheckResourceAttr("vcd_tm_region.test", "status", "READY"),

					resource.TestCheckResourceAttrSet("data.vcd_tm_supervisor.test", "id"),
					resource.TestCheckResourceAttrPair("data.vcd_tm_supervisor.test", "vcenter_id", "vcd_vcenter.test", "id"),

					resource.TestCheckResourceAttrSet("data.vcd_tm_supervisor_zone.test", "id"),
					resource.TestCheckResourceAttrPair("data.vcd_tm_supervisor_zone.test", "vcenter_id", "vcd_vcenter.test", "id"),
					resource.TestCheckResourceAttrSet("data.vcd_tm_supervisor_zone.test", "cpu_capacity_mhz"),
					resource.TestCheckResourceAttrSet("data.vcd_tm_supervisor_zone.test", "cpu_used_mhz"),
					resource.TestCheckResourceAttrSet("data.vcd_tm_supervisor_zone.test", "memory_capacity_mib"),
					resource.TestCheckResourceAttrSet("data.vcd_tm_supervisor_zone.test", "memory_used_mib"),
				),
			},
			{
				Config: configText2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_manager.test", "id", regexp.MustCompile(`^urn:vcloud:nsxtmanager:`)),
					resource.TestCheckResourceAttrSet("vcd_vcenter.test", "id"),
					resource.TestCheckResourceAttrSet("vcd_tm_region.test", "id"),
					cachedRegionId.testCheckCachedResourceFieldValueChanged("vcd_tm_region.test", "id"),
					resource.TestCheckResourceAttr("vcd_tm_region.test", "is_enabled", "true"),
					resource.TestCheckResourceAttr("vcd_tm_region.test", "description", "Terraform description updated"),
					resource.TestCheckResourceAttrSet("vcd_tm_region.test", "cpu_capacity_mhz"),
					resource.TestCheckResourceAttrSet("vcd_tm_region.test", "cpu_reservation_capacity_mhz"),
					resource.TestCheckResourceAttrSet("vcd_tm_region.test", "memory_capacity_mib"),
					resource.TestCheckResourceAttrSet("vcd_tm_region.test", "memory_reservation_capacity_mib"),
					resource.TestCheckResourceAttr("vcd_tm_region.test", "status", "READY"),

					resource.TestCheckResourceAttrSet("data.vcd_tm_supervisor.test", "id"),
					resource.TestCheckResourceAttrPair("data.vcd_tm_supervisor.test", "vcenter_id", "vcd_vcenter.test", "id"),

					resource.TestCheckResourceAttrSet("data.vcd_tm_supervisor_zone.test", "id"),
					resource.TestCheckResourceAttrPair("data.vcd_tm_supervisor_zone.test", "vcenter_id", "vcd_vcenter.test", "id"),
					resource.TestCheckResourceAttrSet("data.vcd_tm_supervisor_zone.test", "cpu_capacity_mhz"),
					resource.TestCheckResourceAttrSet("data.vcd_tm_supervisor_zone.test", "cpu_used_mhz"),
					resource.TestCheckResourceAttrSet("data.vcd_tm_supervisor_zone.test", "memory_capacity_mib"),
					resource.TestCheckResourceAttrSet("data.vcd_tm_supervisor_zone.test", "memory_used_mib"),
				),
			},
			{
				Config: configText3,
				Check: resource.ComposeTestCheckFunc(
					resourceFieldsEqual("vcd_tm_region.test", "data.vcd_tm_region.test", []string{
						"is_enabled", // TODO: TM: field is not populated on read
					}),
				),
			},
			{
				ResourceName:      "vcd_tm_region.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     params["Testname"].(string),
				ImportStateVerifyIgnore: []string{
					"is_enabled", // TODO: TM: field is not populated on read
				},
			},
		},
	})

	postTestChecks(t)
}

const testAccVcdRegionPrerequisites = `
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
  name       = "{{.VcenterSupervisor}}"
  vcenter_id = vcd_vcenter.test.id

  depends_on = [vcd_vcenter.test]
}

data "vcd_tm_supervisor_zone" "test" {
  supervisor_id = data.vcd_tm_supervisor.test.id
  name          = "{{.VcenterSupervisorZone}}"
}
`

const testAccVcdRegionStep1 = testAccVcdRegionPrerequisites + `
resource "vcd_tm_region" "test" {
  name                 = "{{.Testname}}"
  description          = "Terraform description"
  is_enabled           = true
  nsx_manager_id       = vcd_nsxt_manager.test.id
  supervisor_ids       = [data.vcd_tm_supervisor.test.id]
  storage_policy_names = ["{{.VcenterStorageProfile}}"]
}
`

const testAccVcdRegionStep2 = testAccVcdRegionPrerequisites + `
# skip-binary-test: update test
resource "vcd_tm_region" "test" {
  name                 = "{{.Testname}}"
  description          = "Terraform description updated"
  is_enabled           = true
  nsx_manager_id       = vcd_nsxt_manager.test.id
  supervisor_ids       = [data.vcd_tm_supervisor.test.id]
  storage_policy_names = ["{{.VcenterStorageProfile}}"]
}
`

const testAccVcdRegionStep3DS = testAccVcdRegionStep2 + `
data "vcd_tm_region" "test" {
  name = vcd_tm_region.test.name
}
`
