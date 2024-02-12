//go:build ALL || functional

package vcd

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcdVersion(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	vcdClient := createSystemTemporaryVCDConnection()
	currentVersion, err := vcdClient.Client.GetVcdShortVersion()
	if err != nil {
		t.Fatalf("could not get VCD version: %s", err)
	}

	var params = StringMap{
		"Condition": ">= " + currentVersion,
	}
	testParamsNotEmpty(t, params)

	configText1 := templateFill(testAccVcdVersion, params)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText1)

	resource.ParallelTest(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: configText1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.vcd_version.version", "id", fmt.Sprintf("vcd_version=%s,condition=%s,fail_if_not_match=false", currentVersion, params["Condition"])),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdVersion = `
data "vcd_version" "version" {
	name = "{{.Vcenter}}"
  }
`
