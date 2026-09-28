//go:build network || nsxt || ALL || functional

package vcd

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/vmware/go-vcloud-director/v3/govcd"
)

// TestAccVcdNsxtNetworkContextProfile tests custom (TENANT scoped) Network Context Profile
// lifecycle in a VDC Group together with the pre-existing data source
func TestAccVcdNsxtNetworkContextProfile(t *testing.T) {
	preTestChecks(t)

	var params = StringMap{
		"Org":         testConfig.VCD.Org,
		"VdcGroup":    testConfig.Nsxt.VdcGroup,
		"ProfileName": t.Name(),
		"Tags":        "nsxt network",
	}
	testParamsNotEmpty(t, params)

	configText1 := templateFill(testAccVcdNsxtNetworkContextProfileStep1, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)

	params["FuncName"] = t.Name() + "-step2"
	configText2 := templateFill(testAccVcdNsxtNetworkContextProfileStep2Ds, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText2)

	params["FuncName"] = t.Name() + "-step3"
	configText3 := templateFill(testAccVcdNsxtNetworkContextProfileStep3, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 3: %s", configText3)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckNsxtNetworkContextProfileDestroy(t.Name()),
			testAccCheckNsxtNetworkContextProfileDestroy(t.Name()+"-updated"),
		),
		Steps: []resource.TestStep{
			{
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_nsxt_network_context_profile.custom", "id"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_context_profile.custom", "name", t.Name()),
					resource.TestCheckResourceAttr("vcd_nsxt_network_context_profile.custom", "scope", "TENANT"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_context_profile.custom", "attribute.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_network_context_profile.custom", "attribute.*", map[string]string{
						"type": "APP_ID",
					}),
				),
			},
			{
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_nsxt_network_context_profile.custom", "id"),
					resource.TestCheckResourceAttrPair("vcd_nsxt_network_context_profile.custom", "id",
						"data.vcd_nsxt_network_context_profile.lookup", "id"),
				),
			},
			{
				Config: configText3,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_nsxt_network_context_profile.custom", "id"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_context_profile.custom", "name", t.Name()+"-updated"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_context_profile.custom", "attribute.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_network_context_profile.custom", "attribute.*", map[string]string{
						"type": "APP_ID",
					}),
				),
			},
			{
				ResourceName:      "vcd_nsxt_network_context_profile.custom",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdOrgNsxtVdcGroupObject(testConfig.Nsxt.VdcGroup, t.Name()+"-updated"),
				// context_id is not returned by GET queries, therefore it cannot be set on import
				ImportStateVerifyIgnore: []string{"context_id"},
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdNsxtNetworkContextProfileData = `
data "vcd_vdc_group" "group1" {
  org  = "{{.Org}}"
  name = "{{.VdcGroup}}"
}
`

const testAccVcdNsxtNetworkContextProfileStep1 = testAccVcdNsxtNetworkContextProfileData + `
resource "vcd_nsxt_network_context_profile" "custom" {
  org        = "{{.Org}}"
  context_id = data.vcd_vdc_group.group1.id

  name        = "{{.ProfileName}}"
  description = "Custom Network Context Profile"
  scope       = "TENANT"

  attribute {
    type   = "APP_ID"
    values = ["HTTP", "SSL"]
  }
}
`

const testAccVcdNsxtNetworkContextProfileStep2Ds = testAccVcdNsxtNetworkContextProfileStep1 + `
# skip-binary-test: Data Source test
data "vcd_nsxt_network_context_profile" "lookup" {
  context_id = data.vcd_vdc_group.group1.id
  scope      = "TENANT"
  name       = vcd_nsxt_network_context_profile.custom.name

  depends_on = [vcd_nsxt_network_context_profile.custom]
}
`

const testAccVcdNsxtNetworkContextProfileStep3 = testAccVcdNsxtNetworkContextProfileData + `
resource "vcd_nsxt_network_context_profile" "custom" {
  org        = "{{.Org}}"
  context_id = data.vcd_vdc_group.group1.id

  name        = "{{.ProfileName}}-updated"
  description = "Custom Network Context Profile with sub-attributes"
  scope       = "TENANT"

  attribute {
    type   = "APP_ID"
    values = ["SSL"]

    sub_attribute {
      type   = "TLS_VERSION"
      values = ["TLS_V12", "TLS_V13"]
    }
  }
}
`

// testAccCheckNsxtNetworkContextProfileDestroy checks that a Network Context Profile with a given
// name no longer exists in the testing VDC Group
func testAccCheckNsxtNetworkContextProfileDestroy(profileName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*VCDClient)

		adminOrg, err := conn.GetAdminOrgByName(testConfig.VCD.Org)
		if err != nil {
			return fmt.Errorf("error retrieving Org '%s': %s", testConfig.VCD.Org, err)
		}

		vdcGroup, err := adminOrg.GetVdcGroupByName(testConfig.Nsxt.VdcGroup)
		if err != nil {
			return fmt.Errorf("error retrieving VDC Group '%s': %s", testConfig.Nsxt.VdcGroup, err)
		}

		_, err = govcd.GetNetworkContextProfilesByNameScopeAndContext(&conn.Client, profileName, "TENANT", vdcGroup.VdcGroup.Id)
		if !govcd.ContainsNotFound(err) {
			return fmt.Errorf("NSX-T Network Context Profile '%s' still exists or another error occurred: %s", profileName, err)
		}

		return nil
	}
}
