//go:build gateway || network || nsxt || ALL || functional
// +build gateway network nsxt ALL functional

package vcd

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// TestAccVcdNsxtEdgeBgpConfigTier0 tests out NSX-T Edge Gateway BGP Configuration using dedicated
// Tier-0 gateway
func TestAccVcdNsxtEdgeBgpConfigTier0(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	// Binary tests cannot be run for this test because it requires dedicated Tier-0 gateway which
	// is enabled using custom SDK functions
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	// Ensure Edge Gateway has a dedicated Tier 0 gateway (External network) as BGP and Route
	// Advertisement configuration requires it. Restore it right after the test so that other
	// tests are not impacted.
	updateEdgeGatewayTier0Dedication(t, true)
	defer updateEdgeGatewayTier0Dedication(t, false)

	// String map to fill the template
	var params = StringMap{
		"Org":        testConfig.VCD.Org,
		"NsxtVdc":    testConfig.Nsxt.Vdc,
		"NsxtEdgeGw": testConfig.Nsxt.EdgeGateway,
		"Tags":       "network nsxt",
	}
	testParamsNotEmpty(t, params)

	// First step of test is going to alter some settings but not enable BGP because changing some of the fields
	configText1 := templateFill(testAccVcdNsxtBgpConfig1, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)

	params["FuncName"] = t.Name() + "-step2"
	configText2 := templateFill(testAccVcdNsxtBgpConfig2, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText2)

	params["FuncName"] = t.Name() + "-step3"
	configText3 := templateFill(testAccVcdNsxtBgpConfig3DS, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 3: %s", configText3)

	params["FuncName"] = t.Name() + "-step4"
	configText4 := templateFill(testAccVcdNsxtBgpConfig4, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 4: %s", configText4)

	params["FuncName"] = t.Name() + "-step5"
	configText5 := templateFill(testAccVcdNsxtBgpConfig5, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 5: %s", configText5)

	params["FuncName"] = t.Name() + "-step6"
	configText6 := templateFill(testAccVcdNsxtBgpConfig6DS, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 6: %s", configText6)

	params["FuncName"] = t.Name() + "-step8"
	configText8 := templateFill(testAccVcdNsxtBgpConfig8, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 8: %s", configText8)

	params["FuncName"] = t.Name() + "-step9"
	configText9 := templateFill(testAccVcdNsxtBgpConfig9, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 9: %s", configText9)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNsxtBgpConfigurationDisabled(testConfig.Nsxt.EdgeGateway),
		Steps: []resource.TestStep{
			{
				// Setting time values with BGP being disabled
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_edgegateway_bgp_configuration.testing", "id", regexp.MustCompile(`^urn:vcloud:gateway:.*$`)),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_bgp_configuration.testing", "enabled", "false"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_bgp_configuration.testing", "local_as_number", "65420"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_bgp_configuration.testing", "graceful_restart_mode", "GRACEFUL_AND_HELPER"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_bgp_configuration.testing", "graceful_restart_timer", "190"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_bgp_configuration.testing", "stale_route_timer", "500"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_bgp_configuration.testing", "ecmp_enabled", "false"),
				),
			},
			{
				// Enabling BGP configuration, using decimal format for ASN ('local_as_number' field)
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_edgegateway_bgp_configuration.testing", "id", regexp.MustCompile(`^urn:vcloud:gateway:.*$`)),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_bgp_configuration.testing", "enabled", "true"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_bgp_configuration.testing", "local_as_number", "0.65420"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_bgp_configuration.testing", "graceful_restart_mode", "DISABLE"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_bgp_configuration.testing", "graceful_restart_timer", "190"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_bgp_configuration.testing", "stale_route_timer", "600"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_bgp_configuration.testing", "ecmp_enabled", "true"),
				),
			},
			{
				// Performing data source test
				Config: configText3,
				Check: resource.ComposeAggregateTestCheckFunc(
					resourceFieldsEqual("vcd_nsxt_edgegateway_bgp_configuration.testing", "data.vcd_nsxt_edgegateway_bgp_configuration.testing", nil),
				),
			},
			{
				// Disabling BGP and changing `graceful_restart_mode` with some other fields
				Config: configText4,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_edgegateway_bgp_configuration.testing", "id", regexp.MustCompile(`^urn:vcloud:gateway:.*$`)),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_bgp_configuration.testing", "enabled", "false"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_bgp_configuration.testing", "local_as_number", "65430"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_bgp_configuration.testing", "graceful_restart_mode", "HELPER_ONLY"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_bgp_configuration.testing", "graceful_restart_timer", "190"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_bgp_configuration.testing", "stale_route_timer", "600"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_bgp_configuration.testing", "ecmp_enabled", "true"),
				),
			},
			{
				// Enabling BGP again with different configuration
				Config: configText5,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_edgegateway_bgp_configuration.testing", "id", regexp.MustCompile(`^urn:vcloud:gateway:.*$`)),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_bgp_configuration.testing", "enabled", "true"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_bgp_configuration.testing", "local_as_number", "65420"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_bgp_configuration.testing", "graceful_restart_mode", "GRACEFUL_AND_HELPER"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_bgp_configuration.testing", "graceful_restart_timer", "190"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_bgp_configuration.testing", "stale_route_timer", "600"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_bgp_configuration.testing", "ecmp_enabled", "true"),
				),
			},
			{
				// Testing data source
				Config: configText6,
				Check: resource.ComposeAggregateTestCheckFunc(
					resourceFieldsEqual("vcd_nsxt_edgegateway_bgp_configuration.testing", "data.vcd_nsxt_edgegateway_bgp_configuration.testing", nil),
				),
			},
			{
				// Testing Import functionality
				ResourceName:      "vcd_nsxt_edgegateway_bgp_configuration.testing",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdOrgNsxtVdcObject(testConfig.Nsxt.EdgeGateway),
			},
			{
				// Changing some timer values
				Config: configText8,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_edgegateway_bgp_configuration.testing", "id", regexp.MustCompile(`^urn:vcloud:gateway:.*$`)),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_bgp_configuration.testing", "enabled", "false"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_bgp_configuration.testing", "local_as_number", "65420"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_bgp_configuration.testing", "graceful_restart_mode", "GRACEFUL_AND_HELPER"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_bgp_configuration.testing", "graceful_restart_timer", "150"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_bgp_configuration.testing", "stale_route_timer", "200"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_bgp_configuration.testing", "ecmp_enabled", "true"),
				),
			},
			{
				Config: configText9,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_edgegateway_bgp_configuration.testing", "id", regexp.MustCompile(`^urn:vcloud:gateway:.*$`)),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_bgp_configuration.testing", "enabled", "false"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_bgp_configuration.testing", "local_as_number", "65420"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_bgp_configuration.testing", "graceful_restart_mode", "GRACEFUL_AND_HELPER"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_bgp_configuration.testing", "ecmp_enabled", "false"),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdNsxtBgpConfigPrereqs = `
data "vcd_org_vdc" "nsxt-vdc" {
  org  = "{{.Org}}"
  name = "{{.NsxtVdc}}"
}

data "vcd_nsxt_edgegateway" "testing" {
  org      = "{{.Org}}"
  owner_id = data.vcd_org_vdc.nsxt-vdc.id

  name = "{{.NsxtEdgeGw}}"
}
`

const testAccVcdNsxtBgpConfig1 = testAccVcdNsxtBgpConfigPrereqs + `
resource "vcd_nsxt_edgegateway_bgp_configuration" "testing" {
  org = "{{.Org}}"

  edge_gateway_id = data.vcd_nsxt_edgegateway.testing.id

  enabled                = false
  local_as_number        = "65420"
  graceful_restart_mode  = "GRACEFUL_AND_HELPER"
  graceful_restart_timer = 190
  stale_route_timer      = 500
  ecmp_enabled           = false
}
`

const testAccVcdNsxtBgpConfig2 = testAccVcdNsxtBgpConfigPrereqs + `
resource "vcd_nsxt_edgegateway_bgp_configuration" "testing" {
  org = "{{.Org}}"

  edge_gateway_id = data.vcd_nsxt_edgegateway.testing.id

  enabled                = true
  local_as_number        = "0.65420"
  graceful_restart_mode  = "DISABLE"
  graceful_restart_timer = 190
  stale_route_timer      = 600
  ecmp_enabled           = true
}
`

const testAccVcdNsxtBgpConfig3DS = testAccVcdNsxtBgpConfig2 + `
# skip-binary-test: Data Source test
data "vcd_nsxt_edgegateway_bgp_configuration" "testing" {
  org = "{{.Org}}"
  
  edge_gateway_id = data.vcd_nsxt_edgegateway.testing.id
}
`

const testAccVcdNsxtBgpConfig4 = testAccVcdNsxtBgpConfigPrereqs + `
resource "vcd_nsxt_edgegateway_bgp_configuration" "testing" {
  org = "{{.Org}}"

  edge_gateway_id = data.vcd_nsxt_edgegateway.testing.id

  enabled                = false
  local_as_number        = "65430"
  graceful_restart_mode  = "HELPER_ONLY"
  graceful_restart_timer = 190
  stale_route_timer      = 600
  ecmp_enabled           = true
}
`

const testAccVcdNsxtBgpConfig5 = testAccVcdNsxtBgpConfigPrereqs + `
resource "vcd_nsxt_edgegateway_bgp_configuration" "testing" {
  org = "{{.Org}}"

  edge_gateway_id = data.vcd_nsxt_edgegateway.testing.id

  enabled                = true
  local_as_number        = "65420"
  graceful_restart_mode  = "GRACEFUL_AND_HELPER"
  graceful_restart_timer = 190
  stale_route_timer      = 600
  ecmp_enabled           = true
}
`

const testAccVcdNsxtBgpConfig6DS = testAccVcdNsxtBgpConfig5 + `
# skip-binary-test: Data Source test
data "vcd_nsxt_edgegateway_bgp_configuration" "testing" {
  org = "{{.Org}}"
  
  edge_gateway_id = data.vcd_nsxt_edgegateway.testing.id
}
`

const testAccVcdNsxtBgpConfig8 = testAccVcdNsxtBgpConfigPrereqs + `
resource "vcd_nsxt_edgegateway_bgp_configuration" "testing" {
  org = "{{.Org}}"

  edge_gateway_id = data.vcd_nsxt_edgegateway.testing.id

  enabled                = false
  local_as_number        = "65420"
  graceful_restart_mode  = "GRACEFUL_AND_HELPER"
  graceful_restart_timer = 150
  stale_route_timer      = 200
  ecmp_enabled           = true
}
`

const testAccVcdNsxtBgpConfig9 = testAccVcdNsxtBgpConfigPrereqs + `
resource "vcd_nsxt_edgegateway_bgp_configuration" "testing" {
  org = "{{.Org}}"

  edge_gateway_id = data.vcd_nsxt_edgegateway.testing.id

  enabled                = false
  local_as_number        = "65420"
  graceful_restart_mode  = "GRACEFUL_AND_HELPER"
  ecmp_enabled           = false
}
`

// TestAccVcdNsxtEdgeBgpConfigTier0 tests out NSX-T Edge Gateway BGP Configuration using VRF In VRF
// mode only `enabled` and `ecmp_enabled` fields can be changed - everything else is inherited. This
// test shuffles the values of the `enabled` and `ecmp_enabled` fields and verifies that the other
// fields are filled in with some values.
func TestAccVcdNsxtEdgeBgpConfigVrf(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	// String map to fill the template
	var params = StringMap{
		"Org":             testConfig.VCD.Org,
		"NsxtVdc":         testConfig.Nsxt.Vdc,
		"NsxtManager":     testConfig.Nsxt.Manager,
		"NsxtTier0Router": testConfig.Nsxt.Tier0routerVrf,
		"TestName":        t.Name(),

		"Tags": "network nsxt",
	}
	testParamsNotEmpty(t, params)

	configText1 := templateFill(testAccVcdNsxtEdgeBgpConfigVrfStep1, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)

	params["FuncName"] = t.Name() + "-step2"
	configText2 := templateFill(testAccVcdNsxtEdgeBgpConfigVrfStep2, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText2)

	params["FuncName"] = t.Name() + "-step3"
	configText3DS := templateFill(testAccVcdNsxtEdgeBgpConfigVrfStep3DS, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 3: %s", configText3DS)

	params["FuncName"] = t.Name() + "-step5"
	configText5 := templateFill(testAccVcdNsxtEdgeBgpConfigVrfStep5, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 5: %s", configText5)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVcdNsxtEdgeGatewayDestroy(t.Name()),
		Steps: []resource.TestStep{
			{
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_edgegateway_bgp_configuration.testing", "id", regexp.MustCompile(`^urn:vcloud:gateway:.*$`)),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_bgp_configuration.testing", "enabled", "true"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_bgp_configuration.testing", "ecmp_enabled", "false"),
					resource.TestMatchResourceAttr("vcd_nsxt_edgegateway_bgp_configuration.testing", "local_as_number", regexp.MustCompile(`\d+`)),
					resource.TestMatchResourceAttr("vcd_nsxt_edgegateway_bgp_configuration.testing", "graceful_restart_mode", regexp.MustCompile(`\S+`)),
					resource.TestMatchResourceAttr("vcd_nsxt_edgegateway_bgp_configuration.testing", "graceful_restart_timer", regexp.MustCompile(`\d+`)),
					resource.TestMatchResourceAttr("vcd_nsxt_edgegateway_bgp_configuration.testing", "stale_route_timer", regexp.MustCompile(`\d+`)),
				),
			},
			{
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_edgegateway_bgp_configuration.testing", "id", regexp.MustCompile(`^urn:vcloud:gateway:.*$`)),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_bgp_configuration.testing", "enabled", "true"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_bgp_configuration.testing", "ecmp_enabled", "true"),
					resource.TestMatchResourceAttr("vcd_nsxt_edgegateway_bgp_configuration.testing", "local_as_number", regexp.MustCompile(`\d+`)),
					resource.TestMatchResourceAttr("vcd_nsxt_edgegateway_bgp_configuration.testing", "graceful_restart_mode", regexp.MustCompile(`\S+`)),
					resource.TestMatchResourceAttr("vcd_nsxt_edgegateway_bgp_configuration.testing", "graceful_restart_timer", regexp.MustCompile(`\d+`)),
					resource.TestMatchResourceAttr("vcd_nsxt_edgegateway_bgp_configuration.testing", "stale_route_timer", regexp.MustCompile(`\d+`)),
				),
			},
			{
				Config: configText3DS,
				Check: resource.ComposeAggregateTestCheckFunc(
					resourceFieldsEqual("data.vcd_nsxt_edgegateway_bgp_configuration.testing", "vcd_nsxt_edgegateway_bgp_configuration.testing", nil),
				),
			},
			{
				ResourceName:      "vcd_nsxt_edgegateway_bgp_configuration.testing",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdOrgNsxtVdcObject(t.Name()),
			},
			{
				Config: configText5,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_edgegateway_bgp_configuration.testing", "id", regexp.MustCompile(`^urn:vcloud:gateway:.*$`)),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_bgp_configuration.testing", "enabled", "false"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_bgp_configuration.testing", "ecmp_enabled", "false"),
					resource.TestMatchResourceAttr("vcd_nsxt_edgegateway_bgp_configuration.testing", "local_as_number", regexp.MustCompile(`\d+`)),
					resource.TestMatchResourceAttr("vcd_nsxt_edgegateway_bgp_configuration.testing", "graceful_restart_mode", regexp.MustCompile(`\S+`)),
					resource.TestMatchResourceAttr("vcd_nsxt_edgegateway_bgp_configuration.testing", "graceful_restart_timer", regexp.MustCompile(`\d+`)),
					resource.TestMatchResourceAttr("vcd_nsxt_edgegateway_bgp_configuration.testing", "stale_route_timer", regexp.MustCompile(`\d+`)),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdNsxtEdgeBgpConfigVrfConfig = `
data "vcd_nsxt_manager" "main" {
  name = "{{.NsxtManager}}"
}

data "vcd_nsxt_tier0_router" "router" {
  name            = "{{.NsxtTier0Router}}"
  nsxt_manager_id = data.vcd_nsxt_manager.main.id
}

data "vcd_org_vdc" "nsxt-vdc" {
  org  = "{{.Org}}"
  name = "{{.NsxtVdc}}"
}

resource "vcd_external_network_v2" "vrf-backed" {
  name = "{{.TestName}}"

  nsxt_network {
    nsxt_manager_id      = data.vcd_nsxt_manager.main.id
    nsxt_tier0_router_id = data.vcd_nsxt_tier0_router.router.id
  }

  ip_scope {
    gateway       = "14.14.14.1"
    prefix_length = "24"

    static_ip_pool {
      start_address = "14.14.14.10"
      end_address   = "14.14.14.50"
    }
  }
}

resource "vcd_nsxt_edgegateway" "vrf-backed" {
  org         = "{{.Org}}"
  owner_id    = data.vcd_org_vdc.nsxt-vdc.id
  name        = "{{.TestName}}"

  external_network_id = vcd_external_network_v2.vrf-backed.id


  subnet {
    gateway       = "14.14.14.1"
    prefix_length = "24"
    primary_ip    = "14.14.14.10"
    allocated_ips {
      start_address = "14.14.14.10"
      end_address   = "14.14.14.30"
    }
  }

  dedicate_external_network = true
}
`

const testAccVcdNsxtEdgeBgpConfigVrfStep1 = testAccVcdNsxtEdgeBgpConfigVrfConfig + `
resource "vcd_nsxt_edgegateway_bgp_configuration" "testing" {
  org = "{{.Org}}"

  edge_gateway_id = vcd_nsxt_edgegateway.vrf-backed.id

  enabled         = true
  ecmp_enabled    = false
}
`

const testAccVcdNsxtEdgeBgpConfigVrfStep2 = testAccVcdNsxtEdgeBgpConfigVrfConfig + `
resource "vcd_nsxt_edgegateway_bgp_configuration" "testing" {
  org = "{{.Org}}"

  edge_gateway_id = vcd_nsxt_edgegateway.vrf-backed.id

  enabled         = true
  ecmp_enabled    = true
}
`

const testAccVcdNsxtEdgeBgpConfigVrfStep3DS = testAccVcdNsxtEdgeBgpConfigVrfStep2 + `
# skip-binary-test: Data Source test
data "vcd_nsxt_edgegateway_bgp_configuration" "testing" {
  org = "{{.Org}}"
  
  edge_gateway_id = vcd_nsxt_edgegateway.vrf-backed.id
}
`

const testAccVcdNsxtEdgeBgpConfigVrfStep5 = testAccVcdNsxtEdgeBgpConfigVrfConfig + `
resource "vcd_nsxt_edgegateway_bgp_configuration" "testing" {
  org = "{{.Org}}"

  edge_gateway_id = vcd_nsxt_edgegateway.vrf-backed.id

  enabled         = false
  ecmp_enabled    = false
}
`

func testAccCheckNsxtBgpConfigurationDisabled(edgeGatewayName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*VCDClient)

		org, err := conn.GetOrg(testConfig.VCD.Org)
		if err != nil {
			return fmt.Errorf(errorRetrievingVdcFromOrg, testConfig.Nsxt.Vdc, testConfig.VCD.Org, err)
		}

		edge, err := org.GetNsxtEdgeGatewayByName(edgeGatewayName)
		if err != nil {
			return fmt.Errorf(errorUnableToFindEdgeGateway, edgeGatewayName)
		}

		bgpConfig, err := edge.GetBgpConfiguration()
		if err != nil {
			return fmt.Errorf("error retrieving NSX-T BGP Configuration for Edge Gateway '%s': %s", edgeGatewayName, err)
		}

		if bgpConfig.Enabled {
			return fmt.Errorf("BGP on Edge Gateway is not disabled")
		}

		return nil
	}
}

// TestAccVcdNsxtEdgeBgpConfigVdcGroup tests that BGP configuration works well when Edge Gateway is
// part of VDC Group
func TestAccVcdNsxtEdgeBgpConfigVdcGroup(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	// String map to fill the template
	var params = StringMap{
		"Org":                  testConfig.VCD.Org,
		"NsxtVdc":              testConfig.Nsxt.Vdc,
		"NsxtVdcGroup":         testConfig.Nsxt.VdcGroup,
		"NsxtEdgeGwInVdcGroup": testConfig.Nsxt.VdcGroupEdgeGateway,
		"Tags":                 "network nsxt",
	}
	testParamsNotEmpty(t, params)

	// First step of test is going to alter some settings but not enable BGP because changing some of the fields
	configText1 := templateFill(testAccVcdNsxtBgpVdcGroupConfig1, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)

	params["FuncName"] = t.Name() + "-step2"
	configText2 := templateFill(testAccVcdNsxtBgpVdcGroupConfig2, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText2)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNsxtBgpConfigurationDisabled(testConfig.Nsxt.VdcGroupEdgeGateway),
		Steps: []resource.TestStep{
			{
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_edgegateway_bgp_configuration.testing", "id", regexp.MustCompile(`^urn:vcloud:gateway:.*$`)),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_bgp_configuration.testing", "enabled", "false"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_bgp_configuration.testing", "local_as_number", "65420"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_bgp_configuration.testing", "graceful_restart_mode", "GRACEFUL_AND_HELPER"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_bgp_configuration.testing", "graceful_restart_timer", "190"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_bgp_configuration.testing", "stale_route_timer", "500"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_bgp_configuration.testing", "ecmp_enabled", "false"),
				),
			},
			{
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_edgegateway_bgp_configuration.testing", "id", regexp.MustCompile(`^urn:vcloud:gateway:.*$`)),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_bgp_configuration.testing", "enabled", "true"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_bgp_configuration.testing", "local_as_number", "65420"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_bgp_configuration.testing", "graceful_restart_mode", "GRACEFUL_AND_HELPER"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_bgp_configuration.testing", "graceful_restart_timer", "190"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_bgp_configuration.testing", "stale_route_timer", "500"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_bgp_configuration.testing", "ecmp_enabled", "true"),
				),
			},
		},
	})
}

const testAccVcdNsxtBgpConfigVdcGroupPrereqs = `
data "vcd_vdc_group" "g1" {
  org  = "{{.Org}}"
  name = "{{.NsxtVdcGroup}}"
}

data "vcd_nsxt_edgegateway" "testing" {
  org      = "{{.Org}}"
  owner_id = data.vcd_vdc_group.g1.id

  name = "{{.NsxtEdgeGwInVdcGroup}}"
}
`

const testAccVcdNsxtBgpVdcGroupConfig1 = testAccVcdNsxtBgpConfigVdcGroupPrereqs + `
resource "vcd_nsxt_edgegateway_bgp_configuration" "testing" {
  org = "{{.Org}}"

  edge_gateway_id = data.vcd_nsxt_edgegateway.testing.id

  enabled                = false
  local_as_number        = "65420"
  graceful_restart_mode  = "GRACEFUL_AND_HELPER"
  graceful_restart_timer = 190
  stale_route_timer      = 500
  ecmp_enabled           = false
}
`

const testAccVcdNsxtBgpVdcGroupConfig2 = testAccVcdNsxtBgpConfigVdcGroupPrereqs + `
resource "vcd_nsxt_edgegateway_bgp_configuration" "testing" {
  org = "{{.Org}}"

  edge_gateway_id = data.vcd_nsxt_edgegateway.testing.id

  enabled                = true
  local_as_number        = "65420"
  graceful_restart_mode  = "GRACEFUL_AND_HELPER"
  graceful_restart_timer = 190
  stale_route_timer      = 500
  ecmp_enabled           = true
}
`

// TestAccVcdNsxtEdgeBgpConfigIntegrationVdc tests integration of all 3 BGP configuration resources
// when an Edge Gateway is a member of single VDC:
// * vcd_nsxt_edgegateway_bgp_configuration
// * vcd_nsxt_edgegateway_bgp_neighbor
// * vcd_nsxt_edgegateway_bgp_ip_prefix_list
//
// Each of these resources are tested in their own respective test, but this test aims to test that
// there are no oustanding issues when all 3 resources are used together - this is what a normal user
// would do to achieve complete BGP configuration.
func TestAccVcdNsxtEdgeBgpConfigIntegrationVdc(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	// Binary tests cannot be run for this test because it requires dedicated Tier-0 gateway which
	// is enabled using custom SDK functions
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	// Ensure Edge Gateway has a dedicated Tier 0 gateway (External network) as BGP and Route
	// Advertisement configuration requires it. Restore it right after the test so that other
	// tests are not impacted.
	updateEdgeGatewayTier0Dedication(t, true)
	defer updateEdgeGatewayTier0Dedication(t, false)

	// String map to fill the template
	var params = StringMap{
		"Org":        testConfig.VCD.Org,
		"NsxtVdc":    testConfig.Nsxt.Vdc,
		"NsxtEdgeGw": testConfig.Nsxt.EdgeGateway,

		"Tags": "network nsxt",
	}
	testParamsNotEmpty(t, params)

	configText1 := templateFill(testAccVcdNsxtEdgeBgpConfigIntegration+testAccVcdNsxtBgpConfigPrereqs, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNsxtBgpConfigurationDisabled(testConfig.Nsxt.EdgeGateway),
		Steps: []resource.TestStep{
			{
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_edgegateway_bgp_configuration.testing", "id", regexp.MustCompile(`^urn:vcloud:gateway:.*$`)),
					resource.TestCheckResourceAttrSet("vcd_nsxt_edgegateway_bgp_ip_prefix_list.testing1", "id"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_edgegateway_bgp_ip_prefix_list.testing2", "id"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_edgegateway_bgp_neighbor.testing", "id"),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdNsxtEdgeBgpConfigIntegration = `
resource "vcd_nsxt_edgegateway_bgp_configuration" "testing" {
  org = "{{.Org}}"

  edge_gateway_id = data.vcd_nsxt_edgegateway.testing.id
  
  enabled         = true
  local_as_number = "65420"
  ecmp_enabled    = true
}

resource "vcd_nsxt_edgegateway_bgp_ip_prefix_list" "testing1" {
  org = "{{.Org}}"

  edge_gateway_id = data.vcd_nsxt_edgegateway.testing.id

  name = "{{.TestName}}-1"

  ip_prefix {
	network                  = "30.10.10.0/24"
	action                   = "DENY"
	greater_than_or_equal_to = "25"
	less_than_or_equal_to    = "27"
  }

  ip_prefix {
	network                  = "40.0.0.0/8"
	action                   = "PERMIT"
	greater_than_or_equal_to = "16"
	less_than_or_equal_to    = "24"
  }
}

resource "vcd_nsxt_edgegateway_bgp_ip_prefix_list" "testing2" {
  org = "{{.Org}}"

  edge_gateway_id = data.vcd_nsxt_edgegateway.testing.id

  name = "{{.TestName}}-2"

  ip_prefix {
	network = "10.10.10.0/24"
	action  = "PERMIT"
  }

  ip_prefix {
	network = "20.10.10.0/24"
	action  = "DENY"
  }

  ip_prefix {
	network = "2001:db8::/48"
	action  = "DENY"
  }
}

resource "vcd_nsxt_edgegateway_bgp_neighbor" "testing" {
  org = "{{.Org}}"

  edge_gateway_id = data.vcd_nsxt_edgegateway.testing.id

  ip_address       = "1.1.1.1"
  remote_as_number = "62513"

  keep_alive_timer      = 78
  hold_down_timer       = 400
  graceful_restart_mode = "GRACEFUL_AND_HELPER"
  allow_as_in           = false
  bfd_enabled           = false
  bfd_interval          = 800
  bfd_dead_multiple     = 5
  route_filtering       = "IPV4"

  in_filter_ip_prefix_list_id  = vcd_nsxt_edgegateway_bgp_ip_prefix_list.testing1.id
  out_filter_ip_prefix_list_id = vcd_nsxt_edgegateway_bgp_ip_prefix_list.testing2.id
}
`

// TestAccVcdNsxtEdgeBgpConfigIntegrationVdc tests integration of all 3 BGP configuration resources
// when an Edge Gateway is a member of a VDC Group:
// * vcd_nsxt_edgegateway_bgp_configuration
// * vcd_nsxt_edgegateway_bgp_neighbor
// * vcd_nsxt_edgegateway_bgp_ip_prefix_list
//
// Each of these resources are tested in their own respective test, but this test aims to test that
// there are no oustanding issues when all 3 resources are used together - this is what a normal user
// would do to achieve complete BGP configuration.
func TestAccVcdNsxtEdgeBgpConfigIntegrationVdcGroup(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	// String map to fill the template
	var params = StringMap{
		"Org":                  testConfig.VCD.Org,
		"NsxtVdcGroup":         testConfig.Nsxt.VdcGroup,
		"NsxtEdgeGwInVdcGroup": testConfig.Nsxt.VdcGroupEdgeGateway,

		"Tags": "network nsxt",
	}
	testParamsNotEmpty(t, params)

	configText1 := templateFill(testAccVcdNsxtEdgeBgpConfigIntegration+testAccVcdNsxtBgpConfigVdcGroupPrereqs, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNsxtBgpConfigurationDisabled(testConfig.Nsxt.VdcGroupEdgeGateway),
		Steps: []resource.TestStep{
			{
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_edgegateway_bgp_configuration.testing", "id", regexp.MustCompile(`^urn:vcloud:gateway:.*$`)),
					resource.TestCheckResourceAttrSet("vcd_nsxt_edgegateway_bgp_ip_prefix_list.testing1", "id"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_edgegateway_bgp_ip_prefix_list.testing2", "id"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_edgegateway_bgp_neighbor.testing", "id"),
				),
			},
		},
	})
	postTestChecks(t)
}
