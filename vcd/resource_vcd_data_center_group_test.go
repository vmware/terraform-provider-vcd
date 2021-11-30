//go:build vdcGroup || ALL || functional
// +build vdcGroup ALL functional

package vcd

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// TestAccVcdDataCenterGroupResource tests that data center group can be managed
func TestAccVcdDataCenterGroupResource(t *testing.T) {
	preTestChecks(t)

	// This test requires access to the vCD before filling templates
	// Thus it won't run in the short test
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	vcdClient := createTemporaryVCDConnection()
	if vcdClient.Client.APIVCDMaxVersionIs("< 35.0") {
		t.Skip(t.Name() + " requires at least API v35.0 (vCD 10.2+)")
	}

	// TODO !!! run tests not as Sys Admin
	if !vcdClient.Client.IsSysAdmin {
		t.Skip(t.Name() + " only System Administrator can create Data center group")
	}

	if testConfig.Nsxt.Vdc == "" {
		t.Skip("Variables Nsxt.Vdc must be set")
	}

	// String map to fill the template
	var params = StringMap{
		"Org":                          testConfig.VCD.Org,
		"Name":                         "TestAccVcdDataCenterGroupResource",
		"NameUpdated":                  "TestAccVcdDataCenterGroupResourceUpdated",
		"Description":                  "myDescription",
		"DescriptionUpdate":            "myDescription updated",
		"StartingVdcName":              testConfig.Nsxt.Vdc,
		"ParticipatingVdcName1":        testConfig.Nsxt.Vdc,
		"ParticipatingVdcName2Updated": testConfig.Nsxt.Vdc,
	}

	configText1 := templateFill(testAccVcdDataCenterGroupResource, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)

	params["FuncName"] = t.Name() + "-update"
	configText2 := templateFill(testAccVcdDataCenterGroupResourceUpdate, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText2)

	params["FuncName"] = t.Name() + "-datasource"
	configText3 := templateFill(testAccVcdDataCenterGroupDatasource, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText3)

	resourceAddressDataCenterGroup := "vcd_data_center_group.fromUnitTest"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		PreCheck:          func() { testAccPreCheck(t) },

		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddressDataCenterGroup, "name", params["Name"].(string)),
					resource.TestMatchResourceAttr(resourceAddressDataCenterGroup, "id", regexp.MustCompile(`^\S+`)),
					resource.TestCheckResourceAttr(resourceAddressDataCenterGroup, "description", params["Description"].(string)),
					resource.TestMatchResourceAttr(resourceAddressDataCenterGroup, "starting_vdc_id", regexp.MustCompile(`^\S+`)),
				),
			},
			resource.TestStep{
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddressDataCenterGroup, "name", params["NameUpdated"].(string)),
					resource.TestMatchResourceAttr(resourceAddressDataCenterGroup, "id", regexp.MustCompile(`^\S+`)),
					resource.TestCheckResourceAttr(resourceAddressDataCenterGroup, "description", params["DescriptionUpdate"].(string)),
					resource.TestMatchResourceAttr(resourceAddressDataCenterGroup, "starting_vdc_id", regexp.MustCompile(`^\S+`)),
					//resource.TestMatchResourceAttr(resourceAddressDataCenterGroup, "participating_vdc_ids[0]", regexp.MustCompile(`^\S+`)),
				),
			},
			resource.TestStep{
				Config: configText3,
				Check: resource.ComposeAggregateTestCheckFunc(
					resourceFieldsEqual(resourceAddressDataCenterGroup, "data.vcd_data_center_group.fetchCreated", []string{"participating_vdc_ids.#",
						"starting_vdc_id", "%", "participating_vdc_ids.0"}),
				),
			},
			resource.TestStep{
				ResourceName:            resourceAddressDataCenterGroup,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdFunc:       importStateIdOrgObject(testConfig, params["NameUpdated"].(string)),
				ImportStateVerifyIgnore: []string{"starting_vdc_id"},
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdDataCenterGroupResource = `
data "vcd_org_vdc" "startVdc"{
  name = "{{.StartingVdcName}}"
}

data "vcd_org_vdc" "participatingVdcId"{
  name = "{{.ParticipatingVdcName1}}"
}

resource "vcd_data_center_group" "fromUnitTest" {
  org                   = "{{.Org}}"
  name                  = "{{.Name}}"
  description           = "{{.Description}}"
  starting_vdc_id       = data.vcd_org_vdc.startVdc.id
  participating_vdc_ids = [data.vcd_org_vdc.participatingVdcId.id]
}
`

const testAccVcdDataCenterGroupResourceUpdate = `
data "vcd_org_vdc" "startVdc"{
  name = "{{.StartingVdcName}}"
}

data "vcd_org_vdc" "participatingVdcId"{
  name = "{{.ParticipatingVdcName1}}"
}

resource "vcd_data_center_group" "fromUnitTest" {
  org                   = "{{.Org}}"
  name                  = "{{.NameUpdated}}"
  description           = "{{.DescriptionUpdate}}"
  starting_vdc_id       = data.vcd_org_vdc.startVdc.id
  participating_vdc_ids = [data.vcd_org_vdc.participatingVdcId.id]
}

output "participatingVdcId1" {
  value = tolist(vcd_data_center_group.fromUnitTest.participating_vdc_ids)[0]
}

`
const testAccVcdDataCenterGroupDatasource = testAccVcdDataCenterGroupResourceUpdate + `
# skip-binary-test: data source test only works in acceptance tests
data "vcd_data_center_group" "fetchCreated" {
  org  = "{{.Org}}"
  name = vcd_data_center_group.fromUnitTest.name
}
`
