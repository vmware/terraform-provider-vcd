//go:build tm || ALL || functional

package vcd

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcdTmOrgVdc(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)
	skipIfNotTm(t)

	vCenterHcl, vCenterHclRef := getVCenterHcl(t)
	nsxManagerHcl, nsxManagerHclRef := getNsxManagerHcl(t)
	regionHcl, regionHclRef := getRegionHcl(t, vCenterHclRef, nsxManagerHclRef)
	var params = StringMap{
		"Testname":           t.Name(),
		"SupervisorName":     testConfig.Tm.VcenterSupervisor,
		"SupervisorZoneName": testConfig.Tm.VcenterSupervisorZone,
		"VcenterRef":         vCenterHclRef,
		"RegionId":           fmt.Sprintf("%s.id", regionHclRef),
		// "RegionStoragePolicy": testConfig.Tm.RegionStoragePolicy,

		"Tags": "tm",
	}
	testParamsNotEmpty(t, params)

	preRequisites := vCenterHcl + nsxManagerHcl + regionHcl

	configText1 := templateFill(preRequisites+testAccVcdTmOrgVdcStep1, params)
	params["FuncName"] = t.Name() + "-step2"
	configText2 := templateFill(preRequisites+testAccVcdTmOrgVdcStep2, params)
	params["FuncName"] = t.Name() + "-step3"
	configText3 := templateFill(preRequisites+testAccVcdTmOrgVdcStep3DS, params)

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
					resource.TestCheckResourceAttrSet("vcd_tm_org_vdc.test", "id"),
				// resource.TestCheckResourceAttr("vcd_tm_org.test", "display_name", "terraform-test"),
				// resource.TestCheckResourceAttr("vcd_tm_org.test", "description", "terraform test"),
				// resource.TestCheckResourceAttr("vcd_tm_org.test", "is_enabled", "true"),
				// resource.TestCheckResourceAttr("vcd_tm_org.test", "is_subprovider", "false"),
				// resource.TestMatchResourceAttr("vcd_tm_org.test", "managed_by_id", regexp.MustCompile("^urn:vcloud:org:")),
				// resource.TestCheckResourceAttr("vcd_tm_org.test", "managed_by_name", "System"),
				),
			},
			{
				Config: configText2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_tm_org_vdc.test", "id"),
				// resource.TestCheckResourceAttr("vcd_tm_org.test", "display_name", "terraform-test"),
				// resource.TestCheckResourceAttr("vcd_tm_org.test", "description", "terraform test"),
				// resource.TestCheckResourceAttr("vcd_tm_org.test", "is_enabled", "true"),
				// resource.TestCheckResourceAttr("vcd_tm_org.test", "is_subprovider", "false"),
				// resource.TestMatchResourceAttr("vcd_tm_org.test", "managed_by_id", regexp.MustCompile("^urn:vcloud:org:")),
				// resource.TestCheckResourceAttr("vcd_tm_org.test", "managed_by_name", "System"),
				),
			},
			{
				Config: configText3,
				Check: resource.ComposeTestCheckFunc(
					resourceFieldsEqual("vcd_tm_org_vdc.test", "data.vcd_tm_org_vdc.test", []string{
						"is_enabled", // TODO: TM: is_enabled is always returned as false
					}),
				),
			},
			{
				ResourceName:      "vcd_tm_org_vdc.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     params["Testname"].(string) + "-updated",
				ImportStateVerifyIgnore: []string{
					"is_enabled", // TODO: TM: field is not populated on read
				},
			},
		},
	})

	postTestChecks(t)
}

const testAccVcdTmOrgVdcStep1 = `
data "vcd_tm_supervisor" "test" {
  name       = "{{.SupervisorName}}"
  vcenter_id = {{.VcenterRef}}.id
}

data "vcd_tm_region_zone" "test" {
  region_id = {{.RegionId}}
  name      = "{{.SupervisorZoneName}}"
}

resource "vcd_tm_org_vdc" "test" {
  name           = "{{.Testname}}"
  org_id         = vcd_tm_org.test.id
  region_id      = {{.RegionId}}
  supervisor_ids = [data.vcd_tm_supervisor.test.id]
  zone_resource_allocations {
    zone_id                = data.vcd_tm_region_zone.test.id
    cpu_limit_mhz          = 2000
    cpu_reservation_mhz    = 100
    memory_limit_mib       = 1024
    memory_reservation_mib = 512
  }
}

resource "vcd_tm_org" "test" {
  name         = "{{.Testname}}"
  display_name = "terraform-test"
  description  = "terraform test"
  is_enabled   = true
}
`

const testAccVcdTmOrgVdcStep2 = `
data "vcd_tm_supervisor" "test" {
  name       = "{{.SupervisorName}}"
  vcenter_id = {{.VcenterRef}}.id
  depends_on = [{{.VcenterRef}}]
}

data "vcd_tm_region_zone" "test" {
  region_id = {{.RegionId}}
  name      = "{{.SupervisorZoneName}}"
}

resource "vcd_tm_org_vdc" "test" {
  name           = "{{.Testname}}-updated"
  org_id         = vcd_tm_org.test.id
  region_id      = {{.RegionId}}
  supervisor_ids = [data.vcd_tm_supervisor.test.id]
  zone_resource_allocations {
    zone_id                = data.vcd_tm_region_zone.test.id
    cpu_limit_mhz          = 1900
    cpu_reservation_mhz    = 90
    memory_limit_mib       = 500
    memory_reservation_mib = 200
  }
}

resource "vcd_tm_org" "test" {
  name         = "{{.Testname}}"
  display_name = "terraform-test"
  description  = "terraform test"
  is_enabled   = true
}
`

const testAccVcdTmOrgVdcStep3DS = testAccVcdTmOrgVdcStep2 + `
data "vcd_tm_org_vdc" "test" {
  name = vcd_tm_org_vdc.test.name
}
`
