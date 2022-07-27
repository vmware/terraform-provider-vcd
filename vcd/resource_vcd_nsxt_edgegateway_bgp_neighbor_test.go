//go:build gateway || network || nsxt || ALL || functional
// +build gateway network nsxt ALL functional

package vcd

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcdNsxtEdgeBgpNeighbor(t *testing.T) {
	preTestChecks(t)
	if !usingSysAdmin() {
		t.Skip(t.Name() + " requires system admin privileges")
		return
	}

	// String map to fill the template
	var params = StringMap{
		"Org":                  testConfig.VCD.Org,
		"NsxtVdc":              testConfig.Nsxt.Vdc,
		"NsxtVdcGroup":         testConfig.Nsxt.VdcGroup,
		"NsxtEdgeGwInVdcGroup": testConfig.Nsxt.VdcGroupEdgeGateway,
		"TestName":             t.Name(),

		"Tags": "network nsxt",
	}
	testParamsNotEmpty(t, params)

	configText1 := templateFill(testAccVcdNsxtBgpNeighborVdcGroupConfig1, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)

	params["SkipTest"] = "# skip-binary-test: datasource test will fail when run together with resource"
	params["FuncName"] = t.Name() + "-step2"
	configText2DS := templateFill(testAccVcdNsxtBgpNeighborVdcGroupConfig2DS, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText2DS)

	delete(params, "SkipTest")
	params["FuncName"] = t.Name() + "-step3"
	configText3 := templateFill(testAccVcdNsxtBgpNeighborVdcGroupConfig3, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 3: %s", configText3)

	params["SkipTest"] = "# skip-binary-test: datasource test will fail when run together with resource"
	params["FuncName"] = t.Name() + "-step4"
	configText4DS := templateFill(testAccVcdNsxtBgpNeighborVdcGroupConfig4DS, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 4: %s", configText4DS)

	delete(params, "SkipTest")
	params["FuncName"] = t.Name() + "-step6"
	configText6 := templateFill(testAccVcdNsxtBgpNeighborVdcGroupConfig6, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 6: %s", configText6)

	params["SkipTest"] = "# skip-binary-test: datasource test will fail when run together with resource"
	params["FuncName"] = t.Name() + "-step7"
	configText7DS := templateFill(testAccVcdNsxtBgpNeighborVdcGroupConfig7DS, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 7: %s", configText7DS)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_nsxt_edgegateway_bgp_neighbor.testing", "id"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_bgp_neighbor.testing", "ip_address", "1.1.1.1"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_bgp_neighbor.testing", "remote_as_number", "65211"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_bgp_neighbor.testing", "keep_alive_timer", "80"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_bgp_neighbor.testing", "hold_down_timer", "321"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_bgp_neighbor.testing", "password", "phinai0ca,iS"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_bgp_neighbor.testing", "graceful_restart_mode", "HELPER_ONLY"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_bgp_neighbor.testing", "allow_as_in", "true"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_bgp_neighbor.testing", "bfd_enabled", "true"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_bgp_neighbor.testing", "route_filtering", "DISABLED"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_edgegateway_bgp_neighbor.testing", "bfd_interval"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_edgegateway_bgp_neighbor.testing", "bfd_dead_multiple"),
				),
			},
			{
				Config: configText2DS,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Check that all fields are the same in resource and data source except `password` as `password` cannot be read at all
					resourceFieldsEqual("data.vcd_nsxt_edgegateway_bgp_neighbor.testing", "vcd_nsxt_edgegateway_bgp_neighbor.testing", []string{"password"}),
				),
			},
			{
				Config: configText3,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_nsxt_edgegateway_bgp_neighbor.testing", "id"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_bgp_neighbor.testing", "ip_address", "1.1.1.1"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_bgp_neighbor.testing", "remote_as_number", "62513"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_bgp_neighbor.testing", "password", ""),

					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_bgp_neighbor.testing", "keep_alive_timer", "78"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_bgp_neighbor.testing", "hold_down_timer", "235"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_bgp_neighbor.testing", "graceful_restart_mode", "HELPER_ONLY"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_bgp_neighbor.testing", "allow_as_in", "true"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_bgp_neighbor.testing", "bfd_enabled", "true"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_bgp_neighbor.testing", "route_filtering", "DISABLED"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_bgp_neighbor.testing", "bfd_interval", "800"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_bgp_neighbor.testing", "bfd_dead_multiple", "5"),
				),
			},
			{
				Config: configText4DS,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Check that all fields are the same in resource and data source except `password` as `password` cannot be read at all
					resourceFieldsEqual("data.vcd_nsxt_edgegateway_bgp_neighbor.testing", "vcd_nsxt_edgegateway_bgp_neighbor.testing", []string{"password"}),
				),
			},
			{
				ResourceName:      "vcd_nsxt_edgegateway_bgp_neighbor.testing",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdNsxtEdgeGatewayObjectUsingVdcGroup(testConfig.Nsxt.VdcGroup, testConfig.Nsxt.VdcGroupEdgeGateway, `1.1.1.1`),
				// 'password' field cannot be retrieved once it is set
				ImportStateVerifyIgnore: []string{"password"},
			},
			{
				Config: configText6,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_nsxt_edgegateway_bgp_neighbor.testing", "id"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_bgp_neighbor.testing", "ip_address", "1.1.1.1"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_bgp_neighbor.testing", "remote_as_number", "62513"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_bgp_neighbor.testing", "keep_alive_timer", "78"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_bgp_neighbor.testing", "hold_down_timer", "400"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_bgp_neighbor.testing", "graceful_restart_mode", "GRACEFUL_AND_HELPER"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_bgp_neighbor.testing", "allow_as_in", "false"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_bgp_neighbor.testing", "bfd_enabled", "false"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_bgp_neighbor.testing", "bfd_interval", "800"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_bgp_neighbor.testing", "bfd_dead_multiple", "5"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_bgp_neighbor.testing", "route_filtering", "IPV4"),

					// TODO - integrate tests when BGP IP Prefix List PRs are merged
					//resource.TestCheckResourceAttrSet("vcd_nsxt_edgegateway_bgp_neighbor.testing", "in_filter_ip_prefix_list_id"),
					//resource.TestCheckResourceAttrSet("vcd_nsxt_edgegateway_bgp_neighbor.testing", "out_filter_ip_prefix_list_id"),
				),
			},
			{
				Config: configText7DS,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Check that all fields are the same in resource and data source except `password` as `password` cannot be read at all
					resourceFieldsEqual("data.vcd_nsxt_edgegateway_bgp_neighbor.testing", "vcd_nsxt_edgegateway_bgp_neighbor.testing", []string{"password"}),
				),
			},
		},
	})
}

const testAccVcdNsxtBgpNeighborConfigVdcGroupPrereqs = `
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

const testAccVcdNsxtBgpNeighborVdcGroupConfig1 = testAccVcdNsxtBgpNeighborConfigVdcGroupPrereqs + `
resource "vcd_nsxt_edgegateway_bgp_neighbor" "testing" {
  org = "{{.Org}}"

  edge_gateway_id = data.vcd_nsxt_edgegateway.testing.id

  ip_address       = "1.1.1.1"
  remote_as_number = "65211"
  
  password              = "phinai0ca,iS"
  keep_alive_timer      = 80
  hold_down_timer       = 321
  graceful_restart_mode = "HELPER_ONLY"
  allow_as_in           = true
  bfd_enabled           = true
  route_filtering       = "DISABLED"
}
`

const testAccVcdNsxtBgpNeighborVdcGroupConfig2DS = testAccVcdNsxtBgpNeighborVdcGroupConfig1 + `
data "vcd_nsxt_edgegateway_bgp_neighbor" "testing" {
  org = "{{.Org}}"
  
  edge_gateway_id = data.vcd_nsxt_edgegateway.testing.id

  ip_address = vcd_nsxt_edgegateway_bgp_neighbor.testing.ip_address

  depends_on = [vcd_nsxt_edgegateway_bgp_neighbor.testing]
}
`

const testAccVcdNsxtBgpNeighborVdcGroupConfig3 = testAccVcdNsxtBgpNeighborConfigVdcGroupPrereqs + `
resource "vcd_nsxt_edgegateway_bgp_neighbor" "testing" {
  org = "{{.Org}}"

  edge_gateway_id = data.vcd_nsxt_edgegateway.testing.id

  ip_address       = "1.1.1.1"
  remote_as_number = "62513"

  keep_alive_timer      = 78
  hold_down_timer       = 235
  graceful_restart_mode = "HELPER_ONLY"
  allow_as_in           = true
  bfd_enabled           = true
  bfd_interval          = 800
  bfd_dead_multiple     = 5
  route_filtering       = "DISABLED"
}
`

const testAccVcdNsxtBgpNeighborVdcGroupConfig4DS = testAccVcdNsxtBgpNeighborVdcGroupConfig3 + `
data "vcd_nsxt_edgegateway_bgp_neighbor" "testing" {
  org = "{{.Org}}"

  edge_gateway_id = data.vcd_nsxt_edgegateway.testing.id

  ip_address = vcd_nsxt_edgegateway_bgp_neighbor.testing.ip_address

  depends_on = [vcd_nsxt_edgegateway_bgp_neighbor.testing]
}
`

const testAccVcdNsxtBgpNeighborVdcGroupConfig6 = testAccVcdNsxtBgpNeighborConfigVdcGroupPrereqs + `
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
  
  # TODO - integrate tests when BGP IP Prefix List PRs are merged
  # in_filter_ip_prefix_list_id = "3a2021ed-eaf8-4ae6-a651-fb7870b2807e"
  # out_filter_ip_prefix_list_id = "84019fbc-e895-4247-be0e-a3e39ee4da81"
}
`

const testAccVcdNsxtBgpNeighborVdcGroupConfig7DS = testAccVcdNsxtBgpNeighborVdcGroupConfig6 + `
data "vcd_nsxt_edgegateway_bgp_neighbor" "testing" {
  org = "{{.Org}}"

  edge_gateway_id = data.vcd_nsxt_edgegateway.testing.id

  ip_address = vcd_nsxt_edgegateway_bgp_neighbor.testing.ip_address

  depends_on = [vcd_nsxt_edgegateway_bgp_neighbor.testing]
}
`
