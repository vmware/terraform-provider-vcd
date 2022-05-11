//go:build vapp || ALL || functional
// +build vapp ALL functional

package vcd

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/vmware/go-vcloud-director/v2/govcd"
)

func init() {
	testingTags["vapp"] = "resource_vcd_vapp_properties_test.go"
}

func TestAccVcdVAppProperties(t *testing.T) {
	preTestChecks(t)
	var vapp govcd.VApp

	var params = StringMap{
		"Org":         testConfig.VCD.Org,
		"Vdc":         testConfig.VCD.Vdc,
		"EdgeGateway": testConfig.Networking.EdgeGateway,
		"Catalog":     testSuiteCatalogName,
		"CatalogItem": testSuiteCatalogOVAItem,
		"VappName":    t.Name(),
		"Tags":        "vapp",
	}

	configText := templateFill(testAccCheckVcdVApp_properties, params)

	params["FuncName"] = t.Name() + "-step1"
	configText1 := templateFill(testAccCheckVcdVApp_propertiesUpdate, params)

	params["FuncName"] = t.Name() + "-step2"
	configText2 := templateFill(testAccCheckVcdVApp_propertiesRemove, params)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	debugPrintf("#[DEBUG] CONFIGURATION: %s\n", configText)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testParamsNotEmpty(t, params) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVcdVAppVmDestroy(vappName2),
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdVAppExists("vcd_vapp."+t.Name(), &vapp),
					resource.TestCheckResourceAttr("vcd_vapp."+t.Name(), "name", t.Name()),
					resource.TestCheckResourceAttr("vcd_vapp."+t.Name(), `guest_properties.guest.hostname`, "test-host"),
					resource.TestCheckResourceAttr("vcd_vapp."+t.Name(), `guest_properties.guest.another.subkey`, "another-value"),
				),
			},
			{
				Config: configText1,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdVAppExists("vcd_vapp."+t.Name(), &vapp),
					resource.TestCheckResourceAttr("vcd_vapp."+t.Name(), "name", t.Name()),
					resource.TestCheckNoResourceAttr("vcd_vapp."+t.Name(), `guest_properties.guest.hostname`),
					resource.TestCheckResourceAttr("vcd_vapp."+t.Name(), `guest_properties.guest.another.subkey`, "new-value"),
					resource.TestCheckResourceAttr("vcd_vapp."+t.Name(), `guest_properties.guest.third.subkey`, "third-value"),
				),
			},
			{
				Config: configText2,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdVAppExists("vcd_vapp."+t.Name(), &vapp),
					resource.TestCheckResourceAttr("vcd_vapp."+t.Name(), "name", t.Name()),
					resource.TestCheckTypeSetElemAttr("vcd_vapp."+t.Name(), "guest_properties.*", "0"),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccCheckVcdVApp_properties = `
resource "vcd_vapp" "{{.VappName}}" {
  name = "{{.VappName}}"
  org  = "{{.Org}}"
  vdc  = "{{.Vdc}}"

  guest_properties = {
	"guest.hostname"       = "test-host"
	"guest.another.subkey" = "another-value"
  }
}
`

const testAccCheckVcdVApp_propertiesUpdate = `
resource "vcd_vapp" "{{.VappName}}" {
  name = "{{.VappName}}"
  org  = "{{.Org}}"
  vdc  = "{{.Vdc}}"

  guest_properties = {
	"guest.another.subkey" = "new-value"
	"guest.third.subkey"   = "third-value"
  }
}
`

const testAccCheckVcdVApp_propertiesRemove = `
resource "vcd_vapp" "{{.VappName}}" {
  name = "{{.VappName}}"
  org  = "{{.Org}}"
  vdc  = "{{.Vdc}}"
}
`
