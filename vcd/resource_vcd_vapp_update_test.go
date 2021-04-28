// +build vapp ALL functional

package vcd

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/vmware/go-vcloud-director/v2/govcd"
)

func TestAccVcdVApp_Update(t *testing.T) {
	preTestChecks(t)
	var vapp govcd.VApp
	var vappName = "TestAccVcdVAppUpdate"
	var vappDescription = "A long description going up to 256 characters. " + strings.Repeat("x", 209)
	var vappUpdateDescription = "A shorter description."
	var vappUpdateName = vappName + "_new"

	var params = StringMap{
		"Org":             testConfig.VCD.Org,
		"Vdc":             testConfig.VCD.Vdc,
		"VappDef":         vappName,
		"VappName":        vappName,
		"VappDescription": vappDescription,
		"FuncName":        "TestAccVcdVApp_Update",
		"Note":            "",
		"Tags":            "vapp",
	}
	configText := templateFill(testAccCheckVcdVAppUpdate, params)

	params["FuncName"] = "TestAccCheckVcdVApp_Update_update"
	params["VappDescription"] = vappUpdateDescription
	params["VappName"] = vappUpdateName
	params["Note"] = "# skip-binary-test: only for updates"
	configTextUpdate := templateFill(testAccCheckVcdVAppUpdate, params)

	params["FuncName"] = "TestAccCheckVcdVApp_Update_restore"
	params["VappDescription"] = vappDescription
	params["VappName"] = vappName
	configTextRestore := templateFill(testAccCheckVcdVAppUpdate, params)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	debugPrintf("#[DEBUG] CONFIGURATION basic: %s\n", configText)
	debugPrintf("#[DEBUG] CONFIGURATION update: %s\n", configTextUpdate)
	debugPrintf("#[DEBUG] CONFIGURATION restore: %s\n", configTextRestore)

	resourceName := "vcd_vapp." + vappName
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVcdVAppDestroy,
		Steps: []resource.TestStep{
			// Deploy vApp
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdVAppExists(resourceName, &vapp),
					resource.TestCheckResourceAttr(resourceName, "name", vappName),
					resource.TestCheckResourceAttr(resourceName, "description", vappDescription),
				),
			},
			// Rename vApp and update description
			resource.TestStep{
				Config: configTextUpdate,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdVAppExists(resourceName, &vapp),
					resource.TestCheckResourceAttr(resourceName, "name", vappUpdateName),
					resource.TestCheckResourceAttr(resourceName, "description", vappUpdateDescription),
				),
			},
			// Restore original values
			resource.TestStep{
				Config: configTextRestore,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdVAppExists(resourceName, &vapp),
					resource.TestCheckResourceAttr(resourceName, "name", vappName),
					resource.TestCheckResourceAttr(resourceName, "description", vappDescription),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccCheckVcdVAppUpdate = `
{{.Note}}
resource "vcd_vapp" "{{.VappDef}}" {
  org         = "{{.Org}}"
  vdc         = "{{.Vdc}}"
  name        = "{{.VappName}}"
  description = "{{.VappDescription}}"
}
`
