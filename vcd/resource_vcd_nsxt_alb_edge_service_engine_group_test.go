//go:build nsxt || alb || ALL || functional
// +build nsxt alb ALL functional

package vcd

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcdNsxtEdgeServiceEngineGroupAssignmentDedicated(t *testing.T) {
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
		"ReservationModel":   "DEDICATED",
		"ImportableCloud":    testConfig.Nsxt.NsxtAlbImportableCloud,
		"Org":                testConfig.VCD.Org,
		"NsxtVdc":            testConfig.Nsxt.Vdc,
		"EdgeGw":             testConfig.Nsxt.EdgeGateway,
		"Tags":               "nsxt alb",
	}

	params["FuncName"] = t.Name() + "step1"
	params["IsActive"] = "true"
	configText1 := templateFill(testAccVcdNsxtAlbEdgeServiceEngineGroupAssignmentDedicated, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)

	params["FuncName"] = t.Name() + "step2"
	configText2 := templateFill(testAccVcdNsxtAlbEdgeServiceEngineGroupAssignmentDedicatedDS, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText2)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckVcdAlbControllerDestroy("vcd_nsxt_alb_controller.first"),
			testAccCheckVcdAlbServiceEngineGroupDestroy("vcd_nsxt_alb_cloud.first"),
			testAccCheckVcdAlbCloudDestroy("vcd_nsxt_alb_cloud.first"),
			testAccCheckVcdNsxtEdgeGatewayAlbSettingsDestroy(params["EdgeGw"].(string)),
		),

		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_alb_edgegateway_service_engine_group.test", "id", regexp.MustCompile(`\d*`)),
					resource.TestMatchResourceAttr("vcd_nsxt_alb_edgegateway_service_engine_group.test", "deployed_virtual_services", regexp.MustCompile(`\d*`)),
					resource.TestMatchResourceAttr("vcd_nsxt_alb_edgegateway_service_engine_group.test", "reserved_virtual_services", regexp.MustCompile(`\d*`)),
				),
			},
			resource.TestStep{
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_alb_edgegateway_service_engine_group.test", "id", regexp.MustCompile(`\d*`)),
					resource.TestMatchResourceAttr("vcd_nsxt_alb_edgegateway_service_engine_group.test", "deployed_virtual_services", regexp.MustCompile(`\d*`)),
					resourceFieldsEqual("data.vcd_nsxt_alb_edgegateway_service_engine_group.test", "vcd_nsxt_alb_edgegateway_service_engine_group.test", nil),
				),
			},
			resource.TestStep{
				ResourceName:      "vcd_nsxt_alb_edgegateway_service_engine_group.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdNsxtEdgeGatewayObject(testConfig, params["EdgeGw"].(string), "first-se"),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdNsxtAlbEdgeServiceEngineGroupAssignmentDedicated = testAccVcdNsxtAlbGeneralSettings + `
resource "vcd_nsxt_alb_edgegateway_service_engine_group" "test" {
	edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id
	service_engine_group_id = vcd_nsxt_alb_service_engine_group.first.id
}
`

const testAccVcdNsxtAlbEdgeServiceEngineGroupAssignmentDedicatedDS = testAccVcdNsxtAlbEdgeServiceEngineGroupAssignmentDedicated + `
data "vcd_nsxt_alb_edgegateway_service_engine_group" "test" {
	edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id
	service_engine_group_id = vcd_nsxt_alb_service_engine_group.first.id
}
`

func TestAccVcdNsxtEdgeServiceEngineGroupAssignmentShared(t *testing.T) {
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
		"ReservationModel":   "SHARED",
		"ImportableCloud":    testConfig.Nsxt.NsxtAlbImportableCloud,
		"Org":                testConfig.VCD.Org,
		"NsxtVdc":            testConfig.Nsxt.Vdc,
		"EdgeGw":             testConfig.Nsxt.EdgeGateway,
		"Tags":               "nsxt alb",
	}

	params["FuncName"] = t.Name() + "step1"
	params["IsActive"] = "true"
	configText1 := templateFill(testAccVcdNsxtAlbEdgeServiceEngineGroupAssignmentShared, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)

	params["FuncName"] = t.Name() + "step2"
	configText2 := templateFill(testAccVcdNsxtAlbEdgeServiceEngineGroupAssignmentSharedDS, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText2)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckVcdAlbControllerDestroy("vcd_nsxt_alb_controller.first"),
			testAccCheckVcdAlbServiceEngineGroupDestroy("vcd_nsxt_alb_cloud.first"),
			testAccCheckVcdAlbCloudDestroy("vcd_nsxt_alb_cloud.first"),
			testAccCheckVcdNsxtEdgeGatewayAlbSettingsDestroy(params["EdgeGw"].(string)),
		),

		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_alb_edgegateway_service_engine_group.test", "id", regexp.MustCompile(`\d*`)),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_edgegateway_service_engine_group.test", "max_virtual_services", "100"),
					resource.TestCheckResourceAttr("vcd_nsxt_alb_edgegateway_service_engine_group.test", "reserved_virtual_services", "30"),
					resource.TestMatchResourceAttr("vcd_nsxt_alb_edgegateway_service_engine_group.test", "deployed_virtual_services", regexp.MustCompile(`\d*`)),
				),
			},
			resource.TestStep{
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_alb_edgegateway_service_engine_group.test", "id", regexp.MustCompile(`\d*`)),
					resource.TestMatchResourceAttr("vcd_nsxt_alb_edgegateway_service_engine_group.test", "deployed_virtual_services", regexp.MustCompile(`\d*`)),
					resourceFieldsEqual("data.vcd_nsxt_alb_edgegateway_service_engine_group.test", "vcd_nsxt_alb_edgegateway_service_engine_group.test", nil),
				),
			},
			resource.TestStep{
				ResourceName:      "vcd_nsxt_alb_edgegateway_service_engine_group.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdNsxtEdgeGatewayObject(testConfig, params["EdgeGw"].(string), "first-se"),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdNsxtAlbEdgeServiceEngineGroupAssignmentShared = testAccVcdNsxtAlbGeneralSettings + `
resource "vcd_nsxt_alb_edgegateway_service_engine_group" "test" {
  edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id
  service_engine_group_id = vcd_nsxt_alb_service_engine_group.first.id

  max_virtual_services      = 100
  reserved_virtual_services = 30
}
`

const testAccVcdNsxtAlbEdgeServiceEngineGroupAssignmentSharedDS = testAccVcdNsxtAlbEdgeServiceEngineGroupAssignmentDedicated + `
data "vcd_nsxt_alb_edgegateway_service_engine_group" "test" {
  edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id
  service_engine_group_id = vcd_nsxt_alb_service_engine_group.first.id
}
`
