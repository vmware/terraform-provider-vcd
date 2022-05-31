//go:build network || nsxt || ALL || functional
// +build network nsxt ALL functional

package vcd

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func updateEdgeGatewayTier0Dedication(t *testing.T, dedicatedTier0 bool) {
	vcdClient := createSystemTemporaryVCDConnection()
	org, err := vcdClient.GetOrgByName(testConfig.VCD.Org)
	if err != nil {
		t.Fatalf("error retrieving Org '%s': %s", testConfig.VCD.Org, err)
	}
	edge, err := org.GetNsxtEdgeGatewayByName(testConfig.Nsxt.EdgeGateway)
	if err != nil {
		t.Fatalf("error retrieving NSX-T Edge Gateway '%s': %s", testConfig.Nsxt.EdgeGateway, err)
	}

	edge.EdgeGateway.EdgeGatewayUplinks[0].Dedicated = dedicatedTier0
	_, err = edge.Update(edge.EdgeGateway)
	if err != nil {
		t.Fatalf("error updating NSX-T Edge Gateway dedicated Tier 0 gateway usage to '%t': %s", dedicatedTier0, err)
	}
}

func TestAccVcdNsxtEdgeBgpConfig(t *testing.T) {
	preTestChecks(t)
	if noTestCredentials() {
		t.Skip("Skipping test run as no credentials are provided and this test needs to lookup VCD version")
		return
	}
	skipNoNsxtConfiguration(t)

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

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckNsxtFirewallRulesDestroy(testConfig.Nsxt.Vdc, testConfig.Nsxt.EdgeGateway),
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
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_bgp_configuration.testing", "local_as_number", "0.65420"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_bgp_configuration.testing", "graceful_restart_mode", "DISABLE"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_bgp_configuration.testing", "graceful_restart_timer", "190"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_bgp_configuration.testing", "stale_route_timer", "600"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_bgp_configuration.testing", "ecmp_enabled", "true"),
				),
			},
			{
				Config: configText3,
				Check: resource.ComposeAggregateTestCheckFunc(
					resourceFieldsEqual("vcd_nsxt_edgegateway_bgp_configuration.testing", "data.vcd_nsxt_edgegateway_bgp_configuration.testing", nil),
				),
			},
			{
				Config: configText4,
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
				Config: configText6,
				Check: resource.ComposeAggregateTestCheckFunc(
					resourceFieldsEqual("vcd_nsxt_edgegateway_bgp_configuration.testing", "data.vcd_nsxt_edgegateway_bgp_configuration.testing", nil),
				),
			},
			{
				ResourceName:      "vcd_nsxt_edgegateway_bgp_configuration.testing",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdOrgNsxtVdcObject(testConfig, testConfig.Nsxt.EdgeGateway),
			},
			{
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
		},
	})
	postTestChecks(t)
}

const testAccVcdNsxtBgpConfigPrereqs = `
data "vcd_nsxt_edgegateway" "testing" {
  org = "{{.Org}}"
  vdc = "{{.NsxtVdc}}"

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

  enabled             = true
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
