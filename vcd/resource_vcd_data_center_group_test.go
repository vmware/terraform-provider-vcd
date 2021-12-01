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

	if testConfig.Nsxt.Vdc == "" || testConfig.VCD.NsxtProviderVdc.Name == "" ||
		testConfig.VCD.NsxtProviderVdc.NetworkPool == "" || testConfig.VCD.ProviderVdc.StorageProfile == "" {
		t.Skip("Variables Nsxt.Vdc, VCD.NsxtProviderVdc.NetworkPool, VCD.NsxtProviderVdc.Name," +
			" VCD.ProviderVdc.StorageProfile  must be set")
	}

	// String map to fill the template
	var params = StringMap{
		"Org":                       testConfig.VCD.Org,
		"Name":                      "TestAccVcdDataCenterGroupResource",
		"NameUpdated":               "TestAccVcdDataCenterGroupResourceUpdated",
		"Description":               "myDescription",
		"DescriptionUpdate":         "myDescription updated",
		"StartingVdcName":           testConfig.Nsxt.Vdc,
		"ProviderVdc":               testConfig.VCD.NsxtProviderVdc.Name,
		"NetworkPool":               testConfig.VCD.NsxtProviderVdc.NetworkPool,
		"Allocated":                 "1024",
		"Limit":                     "1024",
		"ProviderVdcStorageProfile": testConfig.VCD.ProviderVdc.StorageProfile,
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
					resource.TestCheckOutput("participatingVdcCount", "2"),
				),
			},
			resource.TestStep{
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddressDataCenterGroup, "name", params["NameUpdated"].(string)),
					resource.TestMatchResourceAttr(resourceAddressDataCenterGroup, "id", regexp.MustCompile(`^\S+`)),
					resource.TestCheckResourceAttr(resourceAddressDataCenterGroup, "description", params["DescriptionUpdate"].(string)),
					resource.TestMatchResourceAttr(resourceAddressDataCenterGroup, "starting_vdc_id", regexp.MustCompile(`^\S+`)),
					resource.TestCheckOutput("participatingVdcCount", "1"),
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

const testAccVcdDataCenterGroupNewVdc = `
resource "vcd_org_vdc" "newVdc" {
  name = "newVdc"
  org  = "{{.Org}}"

  allocation_model  = "Flex"
  network_pool_name = "{{.NetworkPool}}"
  provider_vdc_name = "{{.ProviderVdc}}"

  compute_capacity {
    cpu {
      allocated = "{{.Allocated}}"
      limit     = "{{.Limit}}"
    }

    memory {
      allocated = "{{.Allocated}}"
      limit     = "{{.Limit}}"
    }
  }

  storage_profile {
    name    = "{{.ProviderVdcStorageProfile}}"
    enabled = true
    limit   = 10240
    default = true
  }

  metadata = {
    vdc_metadata = "VDC Metadata"
  }

  enabled                    = true
  enable_thin_provisioning   = true
  enable_fast_provisioning   = true
  delete_force               = true
  delete_recursive           = true
  elasticity      			 = true
  include_vm_memory_overhead = true
}
`

const testAccVcdDataCenterGroupResource = testAccVcdDataCenterGroupNewVdc + `
data "vcd_org_vdc" "startVdc"{
  name = "{{.StartingVdcName}}"
}

resource "vcd_data_center_group" "fromUnitTest" {
  org                   = "{{.Org}}"
  name                  = "{{.Name}}"
  description           = "{{.Description}}"
  starting_vdc_id       = data.vcd_org_vdc.startVdc.id
  participating_vdc_ids = [data.vcd_org_vdc.startVdc.id, vcd_org_vdc.newVdc.id]
}

output "participatingVdcCount" {
  value = length(vcd_data_center_group.fromUnitTest.participating_vdc_ids)
}

`

const testAccVcdDataCenterGroupResourceUpdate = testAccVcdDataCenterGroupNewVdc + `
data "vcd_org_vdc" "startVdc"{
  name = "{{.StartingVdcName}}"
}

resource "vcd_data_center_group" "fromUnitTest" {
  org                   = "{{.Org}}"
  name                  = "{{.NameUpdated}}"
  description           = "{{.DescriptionUpdate}}"
  starting_vdc_id       = data.vcd_org_vdc.startVdc.id
  participating_vdc_ids = [data.vcd_org_vdc.startVdc.id]
}

output "participatingVdcCount" {
  value = length(vcd_data_center_group.fromUnitTest.participating_vdc_ids)
}

`
const testAccVcdDataCenterGroupDatasource = testAccVcdDataCenterGroupResourceUpdate + `
# skip-binary-test: data source test only works in acceptance tests
data "vcd_data_center_group" "fetchCreated" {
  org  = "{{.Org}}"
  name = vcd_data_center_group.fromUnitTest.name
}
`
