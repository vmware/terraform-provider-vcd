//go:build alb || ALL || functional
// +build alb ALL functional

package vcd

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/vmware/go-vcloud-director/v2/govcd"
)

func TestAccVcdNsxtAlbController(t *testing.T) {
	preTestChecks(t)
	if !usingSysAdmin() {
		t.Skip(t.Name() + " requires system admin privileges")
		return
	}

	vcdClient := createTemporaryVCDConnection()
	if vcdClient.Client.APIVCDMaxVersionIs("< 35.0") {
		t.Skip(t.Name() + " requires at least API v35.0 (vCD 10.2+)")
	}
	skipNoNsxtAlbConfiguration(t)

	// String map to fill the template
	var params = StringMap{
		"ControllerUrl":      testConfig.Nsxt.NsxtAlbControllerUrl,
		"ControllerUsername": testConfig.Nsxt.NsxtAlbControllerUser,
		"ControllerPassword": testConfig.Nsxt.NsxtAlbControllerPassword,
		"Tags":               "alb",
	}

	configText1 := templateFill(testAccVcdNsxtAlbController, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)

	params["FuncName"] = t.Name() + "-step2"
	configText2 := templateFill(testAccVcdNsxtAlbControllerStep2, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText2)

	params["FuncName"] = t.Name() + "-step3"
	configText4 := templateFill(testAccVcdNsxtAlbControllerStep3DS, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 3: %s", configText4)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckVcdAlbControllerDestroy("vcd_nsxt_alb_controller.first"),
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_alb_controller.first", "id", regexp.MustCompile(`\d*`)),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_controller.first", "name", "aviController1"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_controller.first", "description", "first alb controller"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_controller.first", "username", testConfig.Nsxt.NsxtAlbControllerUser),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_controller.first", "password", testConfig.Nsxt.NsxtAlbControllerPassword),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_controller.first", "license_type", "ENTERPRISE"),
				),
			},
			resource.TestStep{
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_alb_controller.first", "id", regexp.MustCompile(`\d*`)),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_controller.first", "name", "aviController1-renamed"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_controller.first", "description", ""),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_controller.first", "username", testConfig.Nsxt.NsxtAlbControllerUser),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_controller.first", "password", testConfig.Nsxt.NsxtAlbControllerPassword),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_controller.first", "license_type", "BASIC"),
				),
			},
			resource.TestStep{
				ResourceName:            "vcd_nsxt_alb_controller.first",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateId:           "aviController1-renamed",
				ImportStateVerifyIgnore: []string{"password"},
			},
			resource.TestStep{
				Config: configText4,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_alb_controller.first", "id", regexp.MustCompile(`\d*`)),
					// Comparing all data source fields against resource. Field '%' (total attribute number) is skipped
					// because data source does not have password field
					resourceFieldsEqual("data.vcd_nsxt_alb_controller.first", "vcd_nsxt_alb_controller.first", []string{"%"}),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdNsxtAlbController = `
resource "vcd_nsxt_alb_controller" "first" {
  name         = "aviController1"
  description  = "first alb controller"
  url          = "{{.ControllerUrl}}"
  username     = "{{.ControllerUsername}}"
  password     = "{{.ControllerPassword}}"
  license_type = "ENTERPRISE"
}
`

const testAccVcdNsxtAlbControllerStep2 = `
resource "vcd_nsxt_alb_controller" "first" {
  name         = "aviController1-renamed"
  url          = "{{.ControllerUrl}}"
  username     = "{{.ControllerUsername}}"
  password     = "{{.ControllerPassword}}"
  license_type = "BASIC"
}
`

const testAccVcdNsxtAlbControllerStep3DS = `
# skip-binary-test: Data Source test
resource "vcd_nsxt_alb_controller" "first" {
  name         = "aviController1-renamed"
  url          = "{{.ControllerUrl}}"
  username     = "{{.ControllerUsername}}"
  password     = "{{.ControllerPassword}}"
  license_type = "BASIC"
}

data "vcd_nsxt_alb_controller" "first" {
  name = "aviController1-renamed"
}
`

func testAccCheckVcdAlbControllerDestroy(resource string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("not found resource: %s", resource)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set for %s resource", resource)
		}

		client := testAccProvider.Meta().(*VCDClient)
		albController, err := client.GetAlbControllerById(rs.Primary.ID)

		if !govcd.IsNotFound(err) && albController != nil {
			return fmt.Errorf("ALB Controller (ID: %s) was not deleted: %s", rs.Primary.ID, err)
		}
		return nil
	}
}
