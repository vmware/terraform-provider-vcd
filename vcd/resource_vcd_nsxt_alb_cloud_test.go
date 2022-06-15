//go:build nsxt || alb || ALL || functional
// +build nsxt alb ALL functional

package vcd

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/vmware/go-vcloud-director/v2/govcd"
)

func TestAccVcdNsxtAlbCloud(t *testing.T) {
	preTestChecks(t)

	if !usingSysAdmin() {
		t.Skip(t.Name() + " requires system admin privileges")
		return
	}

	skipNoNsxtAlbConfiguration(t)

	// String map to fill the template
	var params = StringMap{
		"ControllerName":     t.Name(),
		"ControllerUrl":      testConfig.Nsxt.NsxtAlbControllerUrl,
		"ControllerUsername": testConfig.Nsxt.NsxtAlbControllerUser,
		"ControllerPassword": testConfig.Nsxt.NsxtAlbControllerPassword,
		"ImportableCloud":    testConfig.Nsxt.NsxtAlbImportableCloud,
		"Tags":               "nsxt alb",
	}
	testParamsNotEmpty(t, params)

	configText1 := templateFill(testAccVcdNsxtAlbCloud, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)

	params["FuncName"] = t.Name() + "-step2"
	configText2 := templateFill(testAccVcdNsxtAlbCloudStep2, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText2)

	params["FuncName"] = t.Name() + "-step4"
	configText4 := templateFill(testAccVcdNsxtAlbCloudStep4DS, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 4: %s", configText4)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVcdAlbCloudDestroy("vcd_nsxt_alb_cloud.first"),
		Steps: []resource.TestStep{
			{
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_alb_cloud.first", "id", regexp.MustCompile(`\d*`)),
					resource.TestMatchResourceAttr("vcd_nsxt_alb_cloud.first", "controller_id", regexp.MustCompile(`\d*`)),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_cloud.first", "name", "nsxt-cloud"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_cloud.first", "description", "first alb cloud"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_alb_cloud.first", "network_pool_name"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_alb_cloud.first", "network_pool_id"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_alb_cloud.first", "health_status"),
					// health_message might be set or not depending on timing therefore it is unreliable to check for it
					//resource.TestCheckResourceAttrSet("vcd_nsxt_alb_cloud.first", "health_message"),
				),
			},
			{
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_alb_cloud.first", "id", regexp.MustCompile(`\d*`)),
					resource.TestMatchResourceAttr("vcd_nsxt_alb_cloud.first", "controller_id", regexp.MustCompile(`\d*`)),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_cloud.first", "name", "nsxt-cloud-renamed"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_cloud.first", "description", ""),
					resource.TestCheckResourceAttrSet("vcd_nsxt_alb_cloud.first", "network_pool_name"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_alb_cloud.first", "network_pool_id"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_alb_cloud.first", "health_status"),
					// health_message might be set or not depending on timing therefore it is unreliable to check for it
					//resource.TestCheckResourceAttrSet("vcd_nsxt_alb_cloud.first", "health_message"),
				),
			},
			{
				ResourceName:      "vcd_nsxt_alb_cloud.first",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     "nsxt-cloud-renamed",
			},
			{
				Config: configText4,
				Check: resource.ComposeAggregateTestCheckFunc(
					// health_status and health_message might have different values purely because of time difference
					resourceFieldsEqual("data.vcd_nsxt_alb_cloud.first", "vcd_nsxt_alb_cloud.first", []string{"%", "health_status", "health_message"}),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdNsxtAlbCloudPrereqs = `
resource "vcd_nsxt_alb_controller" "first" {
  name         = "{{.ControllerName}}"
  description  = "first alb controller"
  url          = "{{.ControllerUrl}}"
  username     = "{{.ControllerUsername}}"
  password     = "{{.ControllerPassword}}"
  license_type = "ENTERPRISE"
}

data "vcd_nsxt_alb_importable_cloud" "cld" {
  name          = "{{.ImportableCloud}}"
  controller_id = vcd_nsxt_alb_controller.first.id
}
`

const testAccVcdNsxtAlbCloud = testAccVcdNsxtAlbCloudPrereqs + `
resource "vcd_nsxt_alb_cloud" "first" {
  name        = "nsxt-cloud"
  description = "first alb cloud"

  controller_id       = vcd_nsxt_alb_controller.first.id
  importable_cloud_id = data.vcd_nsxt_alb_importable_cloud.cld.id
  network_pool_id     = data.vcd_nsxt_alb_importable_cloud.cld.network_pool_id
}
`

const testAccVcdNsxtAlbCloudStep2 = testAccVcdNsxtAlbCloudPrereqs + `
resource "vcd_nsxt_alb_cloud" "first" {
  name        = "nsxt-cloud-renamed"

  controller_id       = vcd_nsxt_alb_controller.first.id
  importable_cloud_id = data.vcd_nsxt_alb_importable_cloud.cld.id
  network_pool_id     = data.vcd_nsxt_alb_importable_cloud.cld.network_pool_id
}
`

const testAccVcdNsxtAlbCloudStep4DS = testAccVcdNsxtAlbCloudStep2 + `
# skip-binary-test: Data Source test
data "vcd_nsxt_alb_cloud" "first" {
  name          = vcd_nsxt_alb_cloud.first.name
}
`

func testAccCheckVcdAlbCloudDestroy(resource string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("not found resource: %s", resource)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set for %s resource", resource)
		}

		client := testAccProvider.Meta().(*VCDClient)
		albCloud, err := client.GetAlbCloudById(rs.Primary.ID)

		if !govcd.IsNotFound(err) && albCloud != nil {
			return fmt.Errorf("ALB Cloud (ID: %s) was not deleted: %s", rs.Primary.ID, err)
		}
		return nil
	}
}
