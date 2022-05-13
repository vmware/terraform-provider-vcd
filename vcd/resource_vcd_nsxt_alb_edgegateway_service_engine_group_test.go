//go:build nsxt || alb || ALL || functional
// +build nsxt alb ALL functional

package vcd

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcdNsxtEdgeGatewayServiceEngineGroupDedicated(t *testing.T) {
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
		"ReservationModel":   "DEDICATED",
		"ImportableCloud":    testConfig.Nsxt.NsxtAlbImportableCloud,
		"Org":                testConfig.VCD.Org,
		"NsxtVdc":            testConfig.Nsxt.Vdc,
		"EdgeGw":             testConfig.Nsxt.EdgeGateway,
		"Tags":               "nsxt alb",
	}
	testParamsNotEmpty(t, params)

	params["FuncName"] = t.Name() + "step1"
	params["IsActive"] = "true"
	configText1 := templateFill(testAccVcdNsxtAlbEdgeGatewayServiceEngineGroupDedicated, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)

	params["FuncName"] = t.Name() + "step2"
	configText2 := templateFill(testAccVcdNsxtAlbEdgeGatewayServiceEngineGroupDedicatedDS, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText2)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckVcdAlbControllerDestroy("vcd_nsxt_alb_controller.first"),
			testAccCheckVcdAlbServiceEngineGroupDestroy("vcd_nsxt_alb_cloud.first"),
			testAccCheckVcdAlbCloudDestroy("vcd_nsxt_alb_cloud.first"),
			testAccCheckVcdNsxtEdgeGatewayAlbSettingsDestroy(params["EdgeGw"].(string)),
		),

		Steps: []resource.TestStep{
			{
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_alb_edgegateway_service_engine_group.test", "id", regexp.MustCompile(`\d*`)),
					resource.TestMatchResourceAttr("vcd_nsxt_alb_edgegateway_service_engine_group.test", "deployed_virtual_services", regexp.MustCompile(`\d*`)),
					resource.TestMatchResourceAttr("vcd_nsxt_alb_edgegateway_service_engine_group.test", "reserved_virtual_services", regexp.MustCompile(`\d*`)),
				),
			},
			{
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_alb_edgegateway_service_engine_group.test", "id", regexp.MustCompile(`\d*`)),
					resource.TestMatchResourceAttr("vcd_nsxt_alb_edgegateway_service_engine_group.test", "deployed_virtual_services", regexp.MustCompile(`\d*`)),
					resourceFieldsEqual("data.vcd_nsxt_alb_edgegateway_service_engine_group.test", "vcd_nsxt_alb_edgegateway_service_engine_group.test", nil),
				),
			},
			{
				ResourceName:      "vcd_nsxt_alb_edgegateway_service_engine_group.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdNsxtEdgeGatewayObject(testConfig, params["EdgeGw"].(string), "first-se"),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdNsxtAlbEdgeGatewayServiceEngineGroupDedicated = testAccVcdNsxtAlbGeneralSettings + `
resource "vcd_nsxt_alb_edgegateway_service_engine_group" "test" {
  org  = "{{.Org}}"
  vdc  = "{{.NsxtVdc}}"

  edge_gateway_id         = vcd_nsxt_alb_settings.test.edge_gateway_id
  service_engine_group_id = vcd_nsxt_alb_service_engine_group.first.id
}
`

const testAccVcdNsxtAlbEdgeGatewayServiceEngineGroupDedicatedDS = testAccVcdNsxtAlbEdgeGatewayServiceEngineGroupDedicated + `
data "vcd_nsxt_alb_edgegateway_service_engine_group" "test" {
# skip-binary-test: Terraform resource cannot have resource and datasource in the same file

  org  = "{{.Org}}"
  vdc  = "{{.NsxtVdc}}"

  edge_gateway_id         = vcd_nsxt_alb_settings.test.edge_gateway_id
  service_engine_group_id = vcd_nsxt_alb_service_engine_group.first.id
}
`

func TestAccVcdNsxtEdgeGatewayServiceEngineGroupShared(t *testing.T) {
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
		"ReservationModel":   "SHARED",
		"ImportableCloud":    testConfig.Nsxt.NsxtAlbImportableCloud,
		"Org":                testConfig.VCD.Org,
		"NsxtVdc":            testConfig.Nsxt.Vdc,
		"EdgeGw":             testConfig.Nsxt.EdgeGateway,
		"Tags":               "nsxt alb",
	}
	testParamsNotEmpty(t, params)

	params["FuncName"] = t.Name() + "step1"
	params["IsActive"] = "true"
	configText1 := templateFill(testAccVcdNsxtAlbEdgeGatewayServiceEngineGroupShared, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)

	params["FuncName"] = t.Name() + "step2"
	configText2 := templateFill(testAccVcdNsxtAlbEdgeServiceEngineGroupSharedDS, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText2)

	params["FuncName"] = t.Name() + "step3"
	configText3 := templateFill(testAccVcdNsxtAlbEdgeGatewayServiceEngineGroupSharedStep3, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 3: %s", configText3)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckVcdAlbControllerDestroy("vcd_nsxt_alb_controller.first"),
			testAccCheckVcdAlbServiceEngineGroupDestroy("vcd_nsxt_alb_cloud.first"),
			testAccCheckVcdAlbCloudDestroy("vcd_nsxt_alb_cloud.first"),
			testAccCheckVcdNsxtEdgeGatewayAlbSettingsDestroy(params["EdgeGw"].(string)),
		),

		Steps: []resource.TestStep{
			{
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_alb_edgegateway_service_engine_group.test", "id", regexp.MustCompile(`\d*`)),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_edgegateway_service_engine_group.test", "max_virtual_services", "100"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_edgegateway_service_engine_group.test", "reserved_virtual_services", "30"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_edgegateway_service_engine_group.test", "deployed_virtual_services", "0"),
				),
			},
			{
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_alb_edgegateway_service_engine_group.test", "id", regexp.MustCompile(`\d*`)),
					resource.TestMatchResourceAttr("vcd_nsxt_alb_edgegateway_service_engine_group.test", "deployed_virtual_services", regexp.MustCompile(`\d*`)),
					resourceFieldsEqual("data.vcd_nsxt_alb_edgegateway_service_engine_group.test", "vcd_nsxt_alb_edgegateway_service_engine_group.test", nil),
				),
			},
			{
				Config: configText3,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_alb_edgegateway_service_engine_group.test", "id", regexp.MustCompile(`\d*`)),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_edgegateway_service_engine_group.test", "max_virtual_services", "70"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_edgegateway_service_engine_group.test", "reserved_virtual_services", "35"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_edgegateway_service_engine_group.test", "deployed_virtual_services", "0"),
				),
			},
			{
				ResourceName:      "vcd_nsxt_alb_edgegateway_service_engine_group.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdNsxtEdgeGatewayObject(testConfig, params["EdgeGw"].(string), "first-se"),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdNsxtAlbEdgeGatewayServiceEngineGroupShared = testAccVcdNsxtAlbGeneralSettings + `
resource "vcd_nsxt_alb_edgegateway_service_engine_group" "test" {
  org  = "{{.Org}}"
  vdc  = "{{.NsxtVdc}}"

  edge_gateway_id         = vcd_nsxt_alb_settings.test.edge_gateway_id
  service_engine_group_id = vcd_nsxt_alb_service_engine_group.first.id

  max_virtual_services      = 100
  reserved_virtual_services = 30
}
`

const testAccVcdNsxtAlbEdgeServiceEngineGroupSharedDS = testAccVcdNsxtAlbEdgeGatewayServiceEngineGroupShared + `
# skip-binary-test: Terraform resource cannot have resource and datasource in the same file
data "vcd_nsxt_alb_edgegateway_service_engine_group" "test" {
  org  = "{{.Org}}"
  vdc  = "{{.NsxtVdc}}"

  edge_gateway_id         = vcd_nsxt_alb_settings.test.edge_gateway_id
  service_engine_group_id = vcd_nsxt_alb_service_engine_group.first.id
}
`

const testAccVcdNsxtAlbEdgeGatewayServiceEngineGroupSharedStep3 = testAccVcdNsxtAlbGeneralSettings + `
resource "vcd_nsxt_alb_edgegateway_service_engine_group" "test" {
  org  = "{{.Org}}"
  vdc  = "{{.NsxtVdc}}"

  edge_gateway_id         = vcd_nsxt_alb_settings.test.edge_gateway_id
  service_engine_group_id = vcd_nsxt_alb_service_engine_group.first.id

  max_virtual_services      = 70
  reserved_virtual_services = 35
}
`
