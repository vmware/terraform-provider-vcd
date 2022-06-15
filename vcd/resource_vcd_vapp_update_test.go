//go:build vapp || ALL || functional
// +build vapp ALL functional

package vcd

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/vmware/go-vcloud-director/v2/govcd"
)

func TestAccVcdVAppUpdate(t *testing.T) {
	preTestChecks(t)
	var vapp govcd.VApp
	var vappName = t.Name()
	var vappDescription = "A long description going up to 256 characters. " + strings.Repeat("x", 209)
	var vappUpdateDescription = "A shorter description."
	var vappUpdateName = vappName + "_new"

	var params = StringMap{
		"Org":             testConfig.VCD.Org,
		"Vdc":             testConfig.VCD.Vdc,
		"VappDef":         vappName,
		"VappName":        vappName,
		"VappDescription": vappDescription,
		"FuncName":        t.Name(),
		"Note":            " ",
		"Tags":            "vapp",
	}
	testParamsNotEmpty(t, params)

	configText := templateFill(testAccVcdVAppUpdate, params)

	params["FuncName"] = t.Name() + "_update"
	params["VappDescription"] = vappUpdateDescription
	params["VappName"] = vappUpdateName
	params["Note"] = "# skip-binary-test: only for updates"
	configTextUpdate := templateFill(testAccVcdVAppUpdate, params)

	params["FuncName"] = t.Name() + "_removal"
	params["VappDescription"] = ""
	configRemoveDescription := templateFill(testAccVcdVAppUpdate, params)

	params["FuncName"] = t.Name() + "_restore"
	params["VappDescription"] = vappDescription
	params["VappName"] = vappName
	configTextRestore := templateFill(testAccVcdVAppUpdate, params)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	debugPrintf("#[DEBUG] CONFIGURATION basic: %s\n", configText)
	debugPrintf("#[DEBUG] CONFIGURATION update: %s\n", configTextUpdate)
	debugPrintf("#[DEBUG] CONFIGURATION removal: %s\n", configRemoveDescription)
	debugPrintf("#[DEBUG] CONFIGURATION restore: %s\n", configTextRestore)

	resourceName := "vcd_vapp." + vappName
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVcdVAppDestroy,
		Steps: []resource.TestStep{
			// Deploy vApp
			{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdVAppExists(resourceName, &vapp),
					resource.TestCheckResourceAttr(resourceName, "name", vappName),
					resource.TestCheckResourceAttr(resourceName, "description", vappDescription),
				),
			},
			// Rename vApp and update description
			{
				Config: configTextUpdate,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdVAppExists(resourceName, &vapp),
					resource.TestCheckResourceAttr(resourceName, "name", vappUpdateName),
					resource.TestCheckResourceAttr(resourceName, "description", vappUpdateDescription),
				),
			},
			// remove description
			{
				Config: configRemoveDescription,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdVAppExists(resourceName, &vapp),
					resource.TestCheckResourceAttr(resourceName, "name", vappUpdateName),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
				),
			},
			// Restore original values
			{
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

const testAccVcdVAppUpdate = `
{{.Note}}
resource "vcd_vapp" "{{.VappDef}}" {
  org         = "{{.Org}}"
  vdc         = "{{.Vdc}}"
  name        = "{{.VappName}}"
  description = "{{.VappDescription}}"
}
`
