//go:build tm || ALL || functional

package vcd

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcdTmProviderGateway(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)
	skipIfNotTm(t)

	vCenterHcl, vCenterHclRef := getVCenterHcl(t)
	nsxManagerHcl, nsxManagerHclRef := getNsxManagerHcl(t)
	regionHcl, regionHclRef := getRegionHcl(t, vCenterHclRef, nsxManagerHclRef)
	ipSpace1Hcl, ipSpace1HclRef := getIpSpaceHcl(t, regionHclRef, "1", "1")
	ipSpace2Hcl, ipSpace2HclRef := getIpSpaceHcl(t, regionHclRef, "2", "2")

	var params = StringMap{
		"Testname":     t.Name(),
		"VcenterRef":   vCenterHclRef,
		"RegionId":     fmt.Sprintf("%s.id", regionHclRef),
		"RegionName":   t.Name(),
		"IpSpace1Id":   fmt.Sprintf("%s.id", ipSpace1HclRef),
		"IpSpace2Id":   fmt.Sprintf("%s.id", ipSpace2HclRef),
		"Tier0Gateway": testConfig.Tm.NsxtTier0Gateway,

		"Tags": "tm",
	}
	testParamsNotEmpty(t, params)

	// TODO: TM: There shouldn't be a need to create `preRequisites` separately, but region
	// creation fails if it is spawned instantly after adding vCenter, therefore this extra step
	// give time (with additional 'refresh' and 'refresh storage policies' operations on vCenter)
	skipBinaryTest := "# skip-binary-test: prerequisite buildup for acceptance tests"
	configText0 := templateFill(vCenterHcl+nsxManagerHcl+skipBinaryTest, params)
	params["FuncName"] = t.Name() + "-step0"

	preRequisites := vCenterHcl + nsxManagerHcl + regionHcl + ipSpace1Hcl + ipSpace2Hcl
	configText1 := templateFill(preRequisites+testAccVcdTmProviderGatewayStep1, params)
	params["FuncName"] = t.Name() + "-step2"
	configText2 := templateFill(preRequisites+testAccVcdTmProviderGatewayStep2, params)
	params["FuncName"] = t.Name() + "-step3"
	configText3 := templateFill(preRequisites+testAccVcdTmProviderGatewayStep3, params)
	params["FuncName"] = t.Name() + "-step4"
	configText4 := templateFill(preRequisites+testAccVcdTmProviderGatewayStep4DS, params)

	debugPrintf("#[DEBUG] CONFIGURATION step1: %s\n", configText1)
	debugPrintf("#[DEBUG] CONFIGURATION step2: %s\n", configText2)
	debugPrintf("#[DEBUG] CONFIGURATION step3: %s\n", configText3)
	debugPrintf("#[DEBUG] CONFIGURATION step4: %s\n", configText4)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	cachedProviderGateway := &testCachedFieldValue{}
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: configText0,
			},
			{
				Config: configText1,
				Check: resource.ComposeTestCheckFunc(
					cachedProviderGateway.cacheTestResourceFieldValue("vcd_tm_provider_gateway.test", "id"),
					resource.TestCheckResourceAttrSet("vcd_tm_provider_gateway.test", "id"),
					resource.TestCheckResourceAttrSet("vcd_tm_provider_gateway.test", "region_id"),
					resource.TestCheckResourceAttrSet("vcd_tm_provider_gateway.test", "nsxt_tier0_gateway_id"),
					resource.TestCheckResourceAttr("vcd_tm_provider_gateway.test", "name", t.Name()),
					resource.TestCheckResourceAttr("vcd_tm_provider_gateway.test", "description", "Made using Terraform"),
					resource.TestCheckResourceAttr("vcd_tm_provider_gateway.test", "ip_space_ids.#", "1"),
				),
			},
			{
				Config: configText2,
				Check: resource.ComposeTestCheckFunc(
					cachedProviderGateway.testCheckCachedResourceFieldValue("vcd_tm_provider_gateway.test", "id"),
					resource.TestCheckResourceAttrSet("vcd_tm_provider_gateway.test", "id"),
					resource.TestCheckResourceAttrSet("vcd_tm_provider_gateway.test", "region_id"),
					resource.TestCheckResourceAttrSet("vcd_tm_provider_gateway.test", "nsxt_tier0_gateway_id"),
					resource.TestCheckResourceAttr("vcd_tm_provider_gateway.test", "name", t.Name()),
					resource.TestCheckResourceAttr("vcd_tm_provider_gateway.test", "description", "Made using Terraform updated"),
					resource.TestCheckResourceAttr("vcd_tm_provider_gateway.test", "ip_space_ids.#", "1"),
				),
			},
			{
				Config: configText3,
				Check: resource.ComposeTestCheckFunc(
					cachedProviderGateway.testCheckCachedResourceFieldValue("vcd_tm_provider_gateway.test", "id"),
					resource.TestCheckResourceAttrSet("vcd_tm_provider_gateway.test", "id"),
					resource.TestCheckResourceAttrSet("vcd_tm_provider_gateway.test", "region_id"),
					resource.TestCheckResourceAttrSet("vcd_tm_provider_gateway.test", "nsxt_tier0_gateway_id"),
					resource.TestCheckResourceAttr("vcd_tm_provider_gateway.test", "name", t.Name()+"-updated"),
					resource.TestCheckResourceAttr("vcd_tm_provider_gateway.test", "description", ""),
					resource.TestCheckResourceAttr("vcd_tm_provider_gateway.test", "ip_space_ids.#", "3"),
				),
			},
			{
				Config: configText4,
				Check: resource.ComposeTestCheckFunc(
					resourceFieldsEqual("vcd_tm_provider_gateway.test", "data.vcd_tm_provider_gateway.test", nil),
				),
			},
			{
				ResourceName:      "vcd_tm_provider_gateway.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     testConfig.Tm.Region + ImportSeparator + params["Testname"].(string) + "-updated",
			},
		},
	})

	postTestChecks(t)
}

const testAccVcdTmProviderGatewayPrereqs = `
resource "vcd_tm_ip_space" "test" {
  name                          = "{{.Testname}}"
  description                   = "description test"
  region_id                     = {{.RegionId}}
  external_scope                = "12.12.0.0/30"
  default_quota_max_subnet_size = 24
  default_quota_max_cidr_count  = 1
  default_quota_max_ip_count    = 1

  internal_scope {
    name = "scope1"
    cidr = "10.0.0.0/28"
  }
}

resource "vcd_tm_ip_space" "test2" {
  name                          = "{{.Testname}}-2"
  description                   = "description test"
  region_id                     = {{.RegionId}}
  external_scope                = "13.12.0.0/30"
  default_quota_max_subnet_size = 24
  default_quota_max_cidr_count  = 1
  default_quota_max_ip_count    = 1

  internal_scope {
    name = "scope1"
    cidr = "9.0.0.0/28"
  }
}

data "vcd_tm_tier0_gateway" "test" {
  name      = "{{.Tier0Gateway}}"
  region_id = {{.RegionId}}
}
`

const testAccVcdTmProviderGatewayStep1 = testAccVcdTmProviderGatewayPrereqs + `
resource "vcd_tm_provider_gateway" "test" {
  name                  = "{{.Testname}}"
  description           = "Made using Terraform"
  region_id             = {{.RegionId}}
  nsxt_tier0_gateway_id = data.vcd_tm_tier0_gateway.test.id
  ip_space_ids          = [ vcd_tm_ip_space.test.id ]
}
`

const testAccVcdTmProviderGatewayStep2 = testAccVcdTmProviderGatewayPrereqs + `
resource "vcd_tm_provider_gateway" "test" {
  name                  = "{{.Testname}}"
  description           = "Made using Terraform updated"
  region_id             = {{.RegionId}}
  nsxt_tier0_gateway_id = data.vcd_tm_tier0_gateway.test.id
  ip_space_ids          = [ vcd_tm_ip_space.test2.id ]
}
`

const testAccVcdTmProviderGatewayStep3 = testAccVcdTmProviderGatewayPrereqs + `
resource "vcd_tm_provider_gateway" "test" {
  name                  = "{{.Testname}}-updated"
  region_id             = {{.RegionId}}
  nsxt_tier0_gateway_id = data.vcd_tm_tier0_gateway.test.id
  ip_space_ids          = [ vcd_tm_ip_space.test2.id, vcd_tm_ip_space.test.id, {{.IpSpace1Id}} ]
}
`

const testAccVcdTmProviderGatewayStep4DS = testAccVcdTmProviderGatewayStep3 + `
data "vcd_tm_provider_gateway" "test" {
  name      = vcd_tm_provider_gateway.test.name
  region_id = {{.RegionId}}
}
`
