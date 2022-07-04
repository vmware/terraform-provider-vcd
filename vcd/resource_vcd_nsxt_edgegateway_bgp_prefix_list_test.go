//go:build gateway || network || nsxt || ALL || functional
// +build gateway network nsxt ALL functional

package vcd

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcdNsxtEdgeBgpIpPrefixListVdcGroup(t *testing.T) {
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

	// First step of test is going to alter some settings but not enable BGP because changing some of the fields
	configText1 := templateFill(testAccVcdNsxtBgpIpPrefixVdcGroupConfig1, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)

	params["SkipTest"] = "# skip-binary-test: datasource test will fail when run together with resource"
	params["FuncName"] = t.Name() + "-step2"
	configText2DS := templateFill(testAccVcdNsxtBgpIpPrefixVdcGroupConfig2DS, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText2DS)

	delete(params, "SkipTest")
	params["FuncName"] = t.Name() + "-step3"
	configText3 := templateFill(testAccVcdNsxtBgpIpPrefixVdcGroupConfig3, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 3: %s", configText3)

	params["SkipTest"] = "# skip-binary-test: datasource test will fail when run together with resource"
	params["FuncName"] = t.Name() + "-step4"
	configText4DS := templateFill(testAccVcdNsxtBgpIpPrefixVdcGroupConfig4DS, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 4: %s", configText4DS)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_nsxt_edgegateway_bgp_ip_prefix_list.testing", "id"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_bgp_ip_prefix_list.testing", "name", t.Name()),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_bgp_ip_prefix_list.testing", "description", "description"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_bgp_ip_prefix_list.testing", "ip_prefix.#", "5"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_edgegateway_bgp_ip_prefix_list.testing", "ip_prefix.*",
						map[string]string{
							"network": "10.10.10.0/24",
							"action":  "PERMIT",
						},
					),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_edgegateway_bgp_ip_prefix_list.testing", "ip_prefix.*",
						map[string]string{
							"network": "20.10.10.0/24",
							"action":  "DENY",
						},
					),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_edgegateway_bgp_ip_prefix_list.testing", "ip_prefix.*",
						map[string]string{
							"network": "2001:db8::/48",
							"action":  "DENY",
						},
					),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_edgegateway_bgp_ip_prefix_list.testing", "ip_prefix.*",
						map[string]string{
							"network":                  "30.10.10.0/24",
							"action":                   "DENY",
							"greater_than_or_equal_to": "25",
							"less_than_or_equal_to":    "27",
						},
					),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_edgegateway_bgp_ip_prefix_list.testing", "ip_prefix.*",
						map[string]string{
							"network":                  "40.0.0.0/8",
							"action":                   "PERMIT",
							"greater_than_or_equal_to": "16",
							"less_than_or_equal_to":    "24",
						},
					),
				),
			},
			{
				Config: configText2DS,
				Check: resource.ComposeAggregateTestCheckFunc(
					resourceFieldsEqual("data.vcd_nsxt_edgegateway_bgp_ip_prefix_list.testing", "vcd_nsxt_edgegateway_bgp_ip_prefix_list.testing", nil),
				),
			},
			{
				Config: configText3,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_nsxt_edgegateway_bgp_ip_prefix_list.testing", "id"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_bgp_ip_prefix_list.testing", "name", t.Name()+"-updated"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_bgp_ip_prefix_list.testing", "description", ""),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_bgp_ip_prefix_list.testing", "ip_prefix.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_edgegateway_bgp_ip_prefix_list.testing", "ip_prefix.*",
						map[string]string{
							"network":                  "30.10.10.0/24",
							"action":                   "DENY",
							"greater_than_or_equal_to": "25",
							"less_than_or_equal_to":    "27",
						},
					),
				),
			},
			{
				Config: configText4DS,
				Check: resource.ComposeAggregateTestCheckFunc(
					resourceFieldsEqual("data.vcd_nsxt_edgegateway_bgp_ip_prefix_list.testing", "vcd_nsxt_edgegateway_bgp_ip_prefix_list.testing", nil),
				),
			},
			{
				ResourceName:      "vcd_nsxt_edgegateway_bgp_ip_prefix_list.testing",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdNsxtEdgeGatewayObjectUsingVdcGroup(testConfig.Nsxt.VdcGroup, testConfig.Nsxt.VdcGroupEdgeGateway, t.Name()+"-updated"),
			},
		},
	})
}

const testAccVcdNsxtBgpIpPrefixConfigVdcGroupPrereqs = `
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

const testAccVcdNsxtBgpIpPrefixVdcGroupConfig1 = testAccVcdNsxtBgpIpPrefixConfigVdcGroupPrereqs + `
resource "vcd_nsxt_edgegateway_bgp_ip_prefix_list" "testing" {
  org = "{{.Org}}"

  edge_gateway_id = data.vcd_nsxt_edgegateway.testing.id

  name        = "{{.TestName}}"
  description = "description"

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
`

const testAccVcdNsxtBgpIpPrefixVdcGroupConfig2DS = testAccVcdNsxtBgpIpPrefixVdcGroupConfig1 + `
data "vcd_nsxt_edgegateway_bgp_ip_prefix_list" "testing" {
  org = "{{.Org}}"
  
  edge_gateway_id = data.vcd_nsxt_edgegateway.testing.id

  name = "{{.TestName}}"

  depends_on = [vcd_nsxt_edgegateway_bgp_ip_prefix_list.testing]
}
`

const testAccVcdNsxtBgpIpPrefixVdcGroupConfig3 = testAccVcdNsxtBgpIpPrefixConfigVdcGroupPrereqs + `
resource "vcd_nsxt_edgegateway_bgp_ip_prefix_list" "testing" {
  org = "{{.Org}}"

  edge_gateway_id = data.vcd_nsxt_edgegateway.testing.id

  name = "{{.TestName}}-updated"

  ip_prefix {
	network                  = "30.10.10.0/24"
	action                   = "DENY"
	greater_than_or_equal_to = "25"
	less_than_or_equal_to    = "27"
  }
}
`

const testAccVcdNsxtBgpIpPrefixVdcGroupConfig4DS = testAccVcdNsxtBgpIpPrefixVdcGroupConfig3 + `
data "vcd_nsxt_edgegateway_bgp_ip_prefix_list" "testing" {
  org = "{{.Org}}"

  edge_gateway_id = data.vcd_nsxt_edgegateway.testing.id

  name = "{{.TestName}}-updated"

  depends_on = [vcd_nsxt_edgegateway_bgp_ip_prefix_list.testing]
}
`
