// +build catalog vdc ALL functional

package vcd

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcdStorageProfileDS(t *testing.T) {
	var params = StringMap{
		"Org":                testConfig.VCD.Org,
		"Vdc":                testConfig.VCD.Vdc,
		"StorageProfileName": testConfig.VCD.ProviderVdc.StorageProfile,
		"Tags":               "catalog",
	}

	configText := templateFill(testAccVcdStorageProfile, params)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { preRunChecks(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("data.vcd_storage_profile.sp", "id", regexp.MustCompile(`^urn:vcloud:vdcstorageProfile:.*`)),
				),
			},
		},
	})
}

const testAccVcdStorageProfile = `
data "vcd_storage_profile" "sp" {
  org  = "{{.Org}}"
  vdc  = "{{.Vdc}}"
  name = "{{.StorageProfileName}}"
}
`
