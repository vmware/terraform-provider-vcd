//go:build tm || ALL || functional

package vcd

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcdTmContentLibraryItem(t *testing.T) {
	preTestChecks(t)

	skipIfNotSysAdmin(t)
	skipIfNotTm(t)

	vCenterHcl, vCenterHclRef := getVCenterHcl(t)
	nsxManagerHcl, nsxManagerHclRef := getNsxManagerHcl(t)
	regionHcl, regionHclRef := getRegionHcl(t, vCenterHclRef, nsxManagerHclRef)
	contentLibraryHcl, contentLibraryHclRef := getContentLibraryHcl(t, regionHclRef)

	var params = StringMap{
		"Name":              t.Name(),
		"ContentLibraryRef": fmt.Sprintf("%s.id", contentLibraryHclRef),
		"OvaPath":           "",
		"Tags":              "tm",
	}
	testParamsNotEmpty(t, params)

	preRequisites := vCenterHcl + nsxManagerHcl + regionHcl + contentLibraryHcl

	configText1 := templateFill(preRequisites+testAccVcdTmContentLibraryItemStep1, params)
	params["FuncName"] = t.Name() + "-step2"
	configText2 := templateFill(preRequisites+testAccVcdTmContentLibraryItemStep2, params)

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
					resource.TestCheckResourceAttrPair(resourceName, "content_library_id", contentLibraryHclRef, "id"),
				),
			},
			{
				Config: configText2,
				Check: resource.ComposeTestCheckFunc(
					resourceFieldsEqual(resourceName, "data.vcd_tm_content_library_item.cli_ds", []string{"file_path"}),
				),
			},
			{
				ResourceName:      "vcd_tm_content_library_item.cli",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     params["Name"].(string),
			},
		},
	})

	postTestChecks(t)
}

const testAccVcdTmContentLibraryItemStep1 = `
resource "vcd_tm_content_library_item" "cli" {
  name               = "{{.Name}}"
  description        = "{{.Name}}"
  content_library_id = "{{.ContentLibraryRef}}"
  file_path          = "{{.OvaPath}}"
}
`

const testAccVcdTmContentLibraryItemStep2 = testAccVcdTmContentLibraryStep1 + `
data "vcd_tm_content_library_item" "cli_ds" {
  name = vcd_tm_content_library_item.cli.name
}
`
