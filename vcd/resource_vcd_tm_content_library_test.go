//go:build tm || ALL || functional

package vcd

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcdTmContentLibrary(t *testing.T) {
	preTestChecks(t)

	skipIfNotSysAdmin(t)
	skipIfNotTm(t)

	var params = StringMap{
		"Name":                t.Name(),
		"RegionStoragePolicy": testConfig.Tm.RegionStoragePolicy,
		"Tags":                "tm",
	}
	testParamsNotEmpty(t, params)

	configText1 := templateFill(testAccVcdTmContentLibraryStep1, params)
	params["FuncName"] = t.Name() + "-step2"
	configText2 := templateFill(testAccVcdTmContentLibraryStep2, params)

	debugPrintf("#[DEBUG] CONFIGURATION step1: %s\n", configText1)
	debugPrintf("#[DEBUG] CONFIGURATION step2: %s\n", configText2)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resourceName := "vcd_tm_content_library.cl"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: configText1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", t.Name()),
					resource.TestCheckResourceAttr(resourceName, "description", t.Name()),
				),
			},
			{
				Config: configText2,
				Check: resource.ComposeTestCheckFunc(
					resourceFieldsEqual(resourceName, "data.vcd_tm_content_library.cl_ds", nil),
				),
			},
			{
				ResourceName:      "vcd_tm_content_library.cl",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     params["Name"].(string),
			},
		},
	})

	postTestChecks(t)
}

const testAccVcdTmContentLibraryStep1 = `
data "vcd_tm_region_storage_policy" "sp" {
  name = "{{.RegionStoragePolicy}}"
}

resource "vcd_tm_content_library" "cl" {
  name = "{{.Name}}"
  description = "{{.Name}}"
  storage_policy_ids = [
    data.vcd_tm_region_storage_policy.sp.id
  ]
}
`

const testAccVcdTmContentLibraryStep2 = testAccVcdTmContentLibraryStep1 + `
data "vcd_tm_content_library" "cl_ds" {
    name = vcd_tm_content_library.cl.name
}
`
