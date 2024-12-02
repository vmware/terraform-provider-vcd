//go:build tm || ALL || functional

package vcd

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcdTmIpSpace(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)
	skipIfNotTm(t)

	vCenterHcl, vCenterHclRef := getVCenterHcl(t)
	nsxManagerHcl, nsxManagerHclRef := getNsxManagerHcl(t)
	regionHcl, regionHclRef := getRegionHcl(t, vCenterHclRef, nsxManagerHclRef)
	var params = StringMap{
		"Testname":   t.Name(),
		"VcenterRef": vCenterHclRef,
		"RegionId":   fmt.Sprintf("%s.id", regionHclRef),

		"Tags": "tm",
	}
	testParamsNotEmpty(t, params)

	// TODO: TM: There shouldn't be a need to create `preRequisites` separately, but region
	// creation fails if it is spawned instantly after adding vCenter, therefore this extra step
	// give time (with additional 'refresh' and 'refresh storage policies' operations on vCenter)
	skipBinaryTest := "# skip-binary-test: prerequisite buildup for acceptance tests"
	configText0 := templateFill(vCenterHcl+nsxManagerHcl+skipBinaryTest, params)
	params["FuncName"] = t.Name() + "-step0"

	preRequisites := vCenterHcl + nsxManagerHcl + regionHcl
	configText1 := templateFill(preRequisites+testAccVcdTmIpSpaceStep1, params)
	params["FuncName"] = t.Name() + "-step2"
	configText2 := templateFill(preRequisites+testAccVcdTmIpSpaceStep2, params)
	params["FuncName"] = t.Name() + "-step3"
	configText3 := templateFill(preRequisites+testAccVcdTmIpSpaceStep3DS, params)

	debugPrintf("#[DEBUG] CONFIGURATION step1: %s\n", configText1)
	debugPrintf("#[DEBUG] CONFIGURATION step2: %s\n", configText2)
	debugPrintf("#[DEBUG] CONFIGURATION step3: %s\n", configText3)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	cachedIpSpaceId := &testCachedFieldValue{}
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: configText0,
			},
			{
				Config: configText1,
				Check: resource.ComposeTestCheckFunc(
					cachedIpSpaceId.cacheTestResourceFieldValue("vcd_tm_ip_space.test", "id"),
					resource.TestCheckResourceAttrSet("vcd_tm_ip_space.test", "id"),
					resource.TestCheckResourceAttrSet("vcd_tm_ip_space.test", "status"),
					resource.TestCheckResourceAttr("vcd_tm_ip_space.test", "name", t.Name()),
					resource.TestCheckResourceAttr("vcd_tm_ip_space.test", "description", "description test"),
					resource.TestCheckResourceAttr("vcd_tm_ip_space.test", "external_scope", "12.12.0.0/16"),
					resource.TestCheckResourceAttr("vcd_tm_ip_space.test", "default_quota_max_subnet_size", "24"),
					resource.TestCheckResourceAttr("vcd_tm_ip_space.test", "default_quota_max_cidr_count", "1"),
					resource.TestCheckResourceAttr("vcd_tm_ip_space.test", "default_quota_max_ip_count", "1"),
					resource.TestCheckResourceAttr("vcd_tm_ip_space.test", "internal_scope.#", "3"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_tm_ip_space.test", "internal_scope.*", map[string]string{
						"name": "scope1",
						"cidr": "10.0.0.0/24",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_tm_ip_space.test", "internal_scope.*", map[string]string{
						"cidr": "11.0.0.0/26",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_tm_ip_space.test", "internal_scope.*", map[string]string{
						"name": "scope3",
						"cidr": "12.0.0.0/27",
					}),
				),
			},
			{
				Config: configText2,
				Check: resource.ComposeTestCheckFunc(
					cachedIpSpaceId.testCheckCachedResourceFieldValue("vcd_tm_ip_space.test", "id"),
					resource.TestCheckResourceAttrSet("vcd_tm_ip_space.test", "id"),
					resource.TestCheckResourceAttr("vcd_tm_ip_space.test", "status", "REALIZED"),
					resource.TestCheckResourceAttr("vcd_tm_ip_space.test", "name", t.Name()+"-updated"),
					resource.TestCheckResourceAttr("vcd_tm_ip_space.test", "description", "description test - update"),
					resource.TestCheckResourceAttr("vcd_tm_ip_space.test", "external_scope", "12.12.0.0/20"),
					resource.TestCheckResourceAttr("vcd_tm_ip_space.test", "default_quota_max_subnet_size", "25"),
					resource.TestCheckResourceAttr("vcd_tm_ip_space.test", "default_quota_max_cidr_count", "-1"),
					resource.TestCheckResourceAttr("vcd_tm_ip_space.test", "default_quota_max_ip_count", "-1"),
					resource.TestCheckResourceAttr("vcd_tm_ip_space.test", "internal_scope.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_tm_ip_space.test", "internal_scope.*", map[string]string{
						"name": "scope3",
						"cidr": "12.0.0.0/27",
					}),
				),
			},
			{
				Config: configText3,
				Check: resource.ComposeTestCheckFunc(
					resourceFieldsEqual("vcd_tm_ip_space.test", "data.vcd_tm_ip_space.test", nil),
				),
			},
			{
				ResourceName:      "vcd_tm_ip_space.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     params["Testname"].(string) + "-updated",
			},
		},
	})

	postTestChecks(t)
}

const testAccVcdTmIpSpaceStep1 = `
resource "vcd_tm_ip_space" "test" {
  name                          = "{{.Testname}}"
  description                   = "description test"
  region_id                     = {{.RegionId}}
  external_scope                = "12.12.0.0/16"
  default_quota_max_subnet_size = 24
  default_quota_max_cidr_count  = 1
  default_quota_max_ip_count    = 1

  internal_scope {
     name = "scope1"
	 cidr = "10.0.0.0/24"
  }

  internal_scope {
	 cidr = "11.0.0.0/26"
  }

  internal_scope {
     name = "scope3"
	 cidr = "12.0.0.0/27"
  }
}
`

const testAccVcdTmIpSpaceStep2 = `
resource "vcd_tm_ip_space" "test" {
  name                          = "{{.Testname}}-updated"
  description                   = "description test - update"
  region_id                     = {{.RegionId}}
  external_scope                = "12.12.0.0/20"
  default_quota_max_subnet_size = 25
  default_quota_max_cidr_count  = -1
  default_quota_max_ip_count    = -1

  internal_scope {
     name = "scope3"
	 cidr = "12.0.0.0/27"
  }
}
`

const testAccVcdTmIpSpaceStep3DS = testAccVcdTmIpSpaceStep2 + `
data "vcd_tm_ip_space" "test" {
  name      = vcd_tm_ip_space.test.name
  region_id = {{.RegionId}}

  depends_on = [ vcd_tm_ip_space.test ]
}
	`
