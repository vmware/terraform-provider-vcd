// +build ALL functional

package vcd

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccVcdVcenter(t *testing.T) {
	if !usingSysAdmin() {
		t.Skip(t.Name() + "  requires system admin privileges")
	}

	var params = StringMap{
		"Vcenter": testConfig.Networking.Vcenter,
	}
	configText := templateFill(datasourceTestVcenter, params)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("data.vcd_vcenter.vc", "id", regexp.MustCompile("^urn:vcloud:vimserver:.*")),
					resource.TestCheckResourceAttrSet("data.vcd_vcenter.vc", "vcenter_version"),
					resource.TestCheckResourceAttrSet("data.vcd_vcenter.vc", "vcenter_host"),
					resource.TestCheckResourceAttrSet("data.vcd_vcenter.vc", "status"),
					resource.TestCheckResourceAttrSet("data.vcd_vcenter.vc", "is_enabled"),
					resource.TestCheckResourceAttrSet("data.vcd_vcenter.vc", "connection_status"),
				),
			},
		},
	})
}

const datasourceTestVcenter = `
data "vcd_vcenter" "vc" {
	name = "{{.Vcenter}}"
  }
`
