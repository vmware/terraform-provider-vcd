//go:build nsxt || alb || ALL || functional
// +build nsxt alb ALL functional

package vcd

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/vmware/go-vcloud-director/v2/govcd"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcdNsxtAlbServiceEngineGroup(t *testing.T) {
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
		"ControllerName":     t.Name(),
		"ControllerUrl":      testConfig.Nsxt.NsxtAlbControllerUrl,
		"ControllerUsername": testConfig.Nsxt.NsxtAlbControllerUser,
		"ControllerPassword": testConfig.Nsxt.NsxtAlbControllerPassword,
		"ImportableCloud":    testConfig.Nsxt.NsxtAlbImportableCloud,
		"Tags":               "nsxt alb",
	}

	configText1 := templateFill(testAccVcdNsxtAlbServiceEngineStep1, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)
	params["FuncName"] = t.Name() + "-step2"
	configText2 := templateFill(testAccVcdNsxtAlbServiceEngineStep2, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText2)

	params["FuncName"] = t.Name() + "-step3"
	configText3 := templateFill(testAccVcdNsxtAlbServiceEngineStep3DS, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 3: %s", configText3)

	params["FuncName"] = t.Name() + "-step4"
	configText4 := templateFill(testAccVcdNsxtAlbServiceEngineStep4, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 4: %s", configText4)

	params["FuncName"] = t.Name() + "-step5"
	configText5 := templateFill(testAccVcdNsxtAlbServiceEngineStep5DS, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 5: %s", configText5)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckVcdAlbServiceEngineGroupDestroy("vcd_nsxt_alb_cloud.first"),
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_alb_service_engine_group.first", "id", regexp.MustCompile(`\d*`)),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_service_engine_group.first", "name", "first-se"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_service_engine_group.first", "description", ""),
					resource.TestCheckResourceAttrSet("vcd_nsxt_alb_service_engine_group.first", "alb_cloud_id"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_service_engine_group.first", "reservation_model", "DEDICATED"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_service_engine_group.first", "service_engine_group_name", "Default-Group"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_alb_service_engine_group.first", "max_virtual_services"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_service_engine_group.first", "reserved_virtual_services", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_service_engine_group.first", "deployed_virtual_services", "0"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_alb_service_engine_group.first", "ha_mode"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_service_engine_group.first", "overallocated", "false"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_service_engine_group.first", "sync_on_refresh", "false"),
				),
			},
			resource.TestStep{
				ResourceName:      "vcd_nsxt_alb_service_engine_group.first",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     "first-se",
				// Because the Importable Service Engine Group API does not list objects once they are consumed
				// by Service Engine Group - it is impossible to lookup name when having only ID. Therefore, on import
				// this field remains empty.
				ImportStateVerifyIgnore: []string{"service_engine_group_name"},
			},
			resource.TestStep{
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_alb_service_engine_group.first", "id", regexp.MustCompile(`\d*`)),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_service_engine_group.first", "name", "first-se-updated"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_service_engine_group.first", "description", "test-description"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_alb_service_engine_group.first", "alb_cloud_id"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_service_engine_group.first", "reservation_model", "SHARED"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_service_engine_group.first", "service_engine_group_name", "Default-Group"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_alb_service_engine_group.first", "max_virtual_services"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_service_engine_group.first", "reserved_virtual_services", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_service_engine_group.first", "deployed_virtual_services", "0"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_alb_service_engine_group.first", "ha_mode"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_service_engine_group.first", "overallocated", "false"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_service_engine_group.first", "sync_on_refresh", "false"),
				),
			},
			resource.TestStep{
				Config: configText3,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_alb_service_engine_group.first", "id", regexp.MustCompile(`\d*`)),
					// Number of fields '%' differs because datasource does not have `service_engine_group_name` as it
					// is impossible to read it after it is consumed
					resourceFieldsEqual("data.vcd_nsxt_alb_service_engine_group.first", "vcd_nsxt_alb_service_engine_group.first", []string{"%"}),
				),
			},
			resource.TestStep{
				Config: configText4,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_alb_service_engine_group.first", "id", regexp.MustCompile(`\d*`)),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_service_engine_group.first", "name", "first-se-updated"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_service_engine_group.first", "description", "test-description"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_alb_service_engine_group.first", "alb_cloud_id"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_service_engine_group.first", "reservation_model", "SHARED"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_service_engine_group.first", "service_engine_group_name", "Default-Group"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_alb_service_engine_group.first", "max_virtual_services"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_service_engine_group.first", "reserved_virtual_services", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_service_engine_group.first", "deployed_virtual_services", "0"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_alb_service_engine_group.first", "ha_mode"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_service_engine_group.first", "overallocated", "false"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_service_engine_group.first", "sync_on_refresh", "true"),
				),
			},
			resource.TestStep{
				Config: configText5,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_alb_service_engine_group.first", "id", regexp.MustCompile(`\d*`)),
					// Number of fields '%' differs because datasource does not have `service_engine_group_name` as it
					// is impossible to read it after it is consumed
					resourceFieldsEqual("data.vcd_nsxt_alb_service_engine_group.first", "vcd_nsxt_alb_service_engine_group.first", []string{"%"}),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdNsxtAlbServiceEnginePrereqs = `
data "vcd_nsxt_alb_importable_cloud" "cld" {
  name          = "{{.ImportableCloud}}"
  controller_id = local.controller_id
}

# Local variable is used to avoid direct reference and cover Terraform core bug https://github.com/hashicorp/terraform/issues/29484
# Even changing NSX-T ALB Controller name in UI, plan will cause to recreate all resources depending 
# on vcd_nsxt_alb_importable_cloud data source if this indirect reference (via local) variable is not used.
locals {
  controller_id = vcd_nsxt_alb_controller.first.id
}


resource "vcd_nsxt_alb_controller" "first" {
  name         = "{{.ControllerName}}"
  description  = "first alb controller"
  url          = "{{.ControllerUrl}}"
  username     = "{{.ControllerUsername}}"
  password     = "{{.ControllerPassword}}"
  license_type = "ENTERPRISE"
}

resource "vcd_nsxt_alb_cloud" "first" {
  name        = "nsxt-cloud"
  description = "first alb cloud"

  controller_id       = vcd_nsxt_alb_controller.first.id
  importable_cloud_id = data.vcd_nsxt_alb_importable_cloud.cld.id
  network_pool_id     = data.vcd_nsxt_alb_importable_cloud.cld.network_pool_id
}
`

const testAccVcdNsxtAlbServiceEngineStep1 = testAccVcdNsxtAlbServiceEnginePrereqs + `
resource "vcd_nsxt_alb_service_engine_group" "first" {
  name                      = "first-se"
  alb_cloud_id              = vcd_nsxt_alb_cloud.first.id
  service_engine_group_name = "Default-Group"
  reservation_model         = "DEDICATED"
}
`

const testAccVcdNsxtAlbServiceEngineStep2 = testAccVcdNsxtAlbServiceEnginePrereqs + `
resource "vcd_nsxt_alb_service_engine_group" "first" {
  name                      = "first-se-updated"
  description               = "test-description"
  alb_cloud_id              = vcd_nsxt_alb_cloud.first.id
  service_engine_group_name = "Default-Group"
  reservation_model         = "SHARED"
}
`

const testAccVcdNsxtAlbServiceEngineStep3DS = testAccVcdNsxtAlbServiceEnginePrereqs + `
# skip-binary-test: Data Source test
resource "vcd_nsxt_alb_service_engine_group" "first" {
  name                      = "first-se-updated"
  description               = "test-description"
  alb_cloud_id              = vcd_nsxt_alb_cloud.first.id
  service_engine_group_name = "Default-Group"
  reservation_model         = "SHARED"
}

data "vcd_nsxt_alb_service_engine_group" "first" {
  name = vcd_nsxt_alb_service_engine_group.first.name
}
`

const testAccVcdNsxtAlbServiceEngineStep4 = testAccVcdNsxtAlbServiceEnginePrereqs + `
resource "vcd_nsxt_alb_service_engine_group" "first" {
  name                      = "first-se-updated"
  description               = "test-description"
  alb_cloud_id              = vcd_nsxt_alb_cloud.first.id
  service_engine_group_name = "Default-Group"
  reservation_model         = "SHARED"
  
  # This feature remains not fully tested as it will impact some of the attributes, but only when tenant operations
  # are available. It will be possible to explicitly check that Sync worked. Now this test ensures it does not break
  # code. 
  sync_on_refresh = true
}
`

const testAccVcdNsxtAlbServiceEngineStep5DS = testAccVcdNsxtAlbServiceEnginePrereqs + `
# skip-binary-test: Data Source test
resource "vcd_nsxt_alb_service_engine_group" "first" {
  name                      = "first-se-updated"
  description               = "test-description"
  alb_cloud_id              = vcd_nsxt_alb_cloud.first.id
  service_engine_group_name = "Default-Group"
  reservation_model         = "SHARED"
  
  # This feature remains not fully tested as it will impact some of the attributes, but only when tenant operations
  # are available. It will be possible to explicitly check that Sync worked. Now this test ensures it does not break
  # code. 
  sync_on_refresh = true
}

data "vcd_nsxt_alb_service_engine_group" "first" {
  name            = vcd_nsxt_alb_service_engine_group.first.name
  sync_on_refresh = true
}
`

func testAccCheckVcdAlbServiceEngineGroupDestroy(resource string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("not found resource: %s", resource)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no ID is set for %s resource", resource)
		}

		client := testAccProvider.Meta().(*VCDClient)
		albCloud, err := client.GetAlbServiceEngineGroupById(rs.Primary.ID)

		if !govcd.IsNotFound(err) && albCloud != nil {
			return fmt.Errorf("ALB Service Engine Group (ID: %s) was not deleted: %s", rs.Primary.ID, err)
		}
		return nil
	}
}
