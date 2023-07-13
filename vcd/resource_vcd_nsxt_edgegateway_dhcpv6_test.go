//go:build gateway || network || nsxt || ALL || functional

package vcd

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccVcdNsxtEdgeDhcpV6(t *testing.T) {
	preTestChecks(t)

	vcdClient := createTemporaryVCDConnection(true)
	if vcdClient == nil {
		t.Skip(acceptanceTestsSkipped)
	}
	if vcdClient.Client.APIVCDMaxVersionIs("< 36.2") {
		t.Skipf("This test tests VCD 10.3.2+ (API V36.2+) features. Skipping.")
	}

	// String map to fill the template
	var params = StringMap{
		"Org":                  testConfig.VCD.Org,
		"NsxtVdc":              testConfig.Nsxt.Vdc,
		"NsxtVdcGroup":         testConfig.Nsxt.VdcGroup,
		"NsxtEdgeGwInVdcGroup": testConfig.Nsxt.VdcGroupEdgeGateway,
		"NsxtEdgeGw":           testConfig.Nsxt.EdgeGateway,
		"TestName":             t.Name(),
		"NsxtManager":          testConfig.Nsxt.Manager,
		"NsxtQosProfileName":   testConfig.Nsxt.GatewayQosProfile,

		"Tags": "network nsxt",
	}
	testParamsNotEmpty(t, params)

	configText1 := templateFill(testAccVcdNsxtEdgeDhcpV6Step1, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)

	params["FuncName"] = t.Name() + "-step2"
	configText2DS := templateFill(testAccVcdNsxtEdgeDhcpV6Step2DS, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText2DS)

	params["FuncName"] = t.Name() + "-step3"
	configText3 := templateFill(testAccVcdNsxtEdgeDhcpV6Step3, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 3: %s", configText3)

	params["FuncName"] = t.Name() + "-step4"
	configText4 := templateFill(testAccVcdNsxtEdgeDhcpV6Step4, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 4: %s", configText4)

	params["FuncName"] = t.Name() + "-step5"
	configText5 := templateFill(testAccVcdNsxtEdgeDhcpV6Step5, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 5: %s", configText5)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckNsxtEdgeDhcpv6Destroy(params["NsxtVdc"].(string), params["NsxtEdgeGw"].(string)),
			testAccCheckNsxtEdgeDhcpv6Destroy(params["NsxtVdcGroup"].(string), params["NsxtEdgeGwInVdcGroup"].(string)),
		),
		Steps: []resource.TestStep{
			{
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_nsxt_edgegateway_dhcpv6.testing-in-vdc", "id"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_dhcpv6.testing-in-vdc", "mode", "DHCPv6"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_edgegateway_dhcpv6.testing-in-vdc-group", "id"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_dhcpv6.testing-in-vdc-group", "mode", "DHCPv6"),
				),
			},
			{
				ResourceName:      "vcd_nsxt_edgegateway_dhcpv6.testing-in-vdc",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdOrgNsxtVdcObject(params["NsxtEdgeGw"].(string)),
			},
			{
				ResourceName:      "vcd_nsxt_edgegateway_dhcpv6.testing-in-vdc-group",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdOrgNsxtVdcGroupObject(params["NsxtVdcGroup"].(string), params["NsxtEdgeGwInVdcGroup"].(string)),
			},
			{
				Config: configText2DS,
				Check: resource.ComposeAggregateTestCheckFunc(

					resource.TestCheckResourceAttrSet("vcd_nsxt_edgegateway_dhcpv6.testing-in-vdc", "id"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_edgegateway_dhcpv6.testing-in-vdc-group", "id"),
					resourceFieldsEqual("data.vcd_nsxt_edgegateway_dhcpv6.testing-in-vdc", "vcd_nsxt_edgegateway_dhcpv6.testing-in-vdc", nil),
					resourceFieldsEqual("data.vcd_nsxt_edgegateway_dhcpv6.testing-in-vdc-group", "vcd_nsxt_edgegateway_dhcpv6.testing-in-vdc-group", nil),
				),
			},
			{
				Config: configText3,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_nsxt_edgegateway_dhcpv6.testing-in-vdc", "id"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_dhcpv6.testing-in-vdc", "mode", "SLAAC"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_dhcpv6.testing-in-vdc", "domain_names.#", "2"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_dhcpv6.testing-in-vdc", "dns_servers.#", "2"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_edgegateway_dhcpv6.testing-in-vdc", "domain_names.*", "non-existing.org.tld"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_edgegateway_dhcpv6.testing-in-vdc", "domain_names.*", "fake.org.tld"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_edgegateway_dhcpv6.testing-in-vdc", "dns_servers.*", "2001:4860:4860::8888"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_edgegateway_dhcpv6.testing-in-vdc", "dns_servers.*", "2001:4860:4860::8844"),

					resource.TestCheckResourceAttrSet("vcd_nsxt_edgegateway_dhcpv6.testing-in-vdc-group", "id"),
					// resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_dhcpv6.testing-in-vdc-group", "enabled", "true"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_dhcpv6.testing-in-vdc-group", "mode", "SLAAC"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_dhcpv6.testing-in-vdc-group", "domain_names.#", "2"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_dhcpv6.testing-in-vdc-group", "dns_servers.#", "2"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_edgegateway_dhcpv6.testing-in-vdc-group", "domain_names.*", "non-existing.org.tld"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_edgegateway_dhcpv6.testing-in-vdc-group", "domain_names.*", "fake.org.tld"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_edgegateway_dhcpv6.testing-in-vdc-group", "dns_servers.*", "2001:4860:4860::8888"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_edgegateway_dhcpv6.testing-in-vdc-group", "dns_servers.*", "2001:4860:4860::8844"),
				),
			},
			{
				Config: configText4,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_nsxt_edgegateway_dhcpv6.testing-in-vdc", "id"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_dhcpv6.testing-in-vdc", "mode", "SLAAC"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_dhcpv6.testing-in-vdc", "domain_names.#", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_dhcpv6.testing-in-vdc", "dns_servers.#", "0"),

					resource.TestCheckResourceAttrSet("vcd_nsxt_edgegateway_dhcpv6.testing-in-vdc-group", "id"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_dhcpv6.testing-in-vdc-group", "mode", "SLAAC"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_dhcpv6.testing-in-vdc-group", "domain_names.#", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_dhcpv6.testing-in-vdc-group", "dns_servers.#", "0"),
				),
			},
			{
				Config: configText5,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_nsxt_edgegateway_dhcpv6.testing-in-vdc", "id"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_dhcpv6.testing-in-vdc", "mode", "DISABLED"),

					resource.TestCheckResourceAttrSet("vcd_nsxt_edgegateway_dhcpv6.testing-in-vdc-group", "id"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway_dhcpv6.testing-in-vdc-group", "mode", "DISABLED"),
				),
			},
		},
	})
}

const testAccVcdNsxtEdgeDhcpV6Shared = `
data "vcd_vdc_group" "g1" {
  org  = "{{.Org}}"
  name = "{{.NsxtVdcGroup}}"
}

data "vcd_nsxt_edgegateway" "testing-in-vdc-group" {
  org      = "{{.Org}}"
  owner_id = data.vcd_vdc_group.g1.id

  name = "{{.NsxtEdgeGwInVdcGroup}}"
}

data "vcd_org_vdc" "v1" {
  org  = "{{.Org}}"
  name = "{{.NsxtVdc}}"
}

data "vcd_nsxt_edgegateway" "testing-in-vdc" {
  org      = "{{.Org}}"
  owner_id = data.vcd_org_vdc.v1.id

  name = "{{.NsxtEdgeGw}}"
}
`

const testAccVcdNsxtEdgeDhcpV6Step1 = testAccVcdNsxtEdgeDhcpV6Shared + `
resource "vcd_nsxt_edgegateway_dhcpv6" "testing-in-vdc" {
  org             = "{{.Org}}"
  edge_gateway_id = data.vcd_nsxt_edgegateway.testing-in-vdc.id

  mode = "DHCPv6"
}

resource "vcd_nsxt_edgegateway_dhcpv6" "testing-in-vdc-group" {
  org             = "{{.Org}}"
  edge_gateway_id = data.vcd_nsxt_edgegateway.testing-in-vdc-group.id

  mode = "DHCPv6"
}
`

const testAccVcdNsxtEdgeDhcpV6Step2DS = testAccVcdNsxtEdgeDhcpV6Step1 + `
# skip-binary-test: datasource test will fail when run together with resource
data "vcd_nsxt_edgegateway_dhcpv6" "testing-in-vdc" {
  org             = "{{.Org}}"
  edge_gateway_id = data.vcd_nsxt_edgegateway.testing-in-vdc.id
}

data "vcd_nsxt_edgegateway_dhcpv6" "testing-in-vdc-group" {
  org             = "{{.Org}}"
  edge_gateway_id = data.vcd_nsxt_edgegateway.testing-in-vdc-group.id
}
`

const testAccVcdNsxtEdgeDhcpV6Step3 = testAccVcdNsxtEdgeDhcpV6Shared + `
resource "vcd_nsxt_edgegateway_dhcpv6" "testing-in-vdc" {
  org             = "{{.Org}}"
  edge_gateway_id = data.vcd_nsxt_edgegateway.testing-in-vdc.id

  mode         = "SLAAC"
  domain_names = ["non-existing.org.tld","fake.org.tld"]
  dns_servers  = ["2001:4860:4860::8888","2001:4860:4860::8844"]
}

resource "vcd_nsxt_edgegateway_dhcpv6" "testing-in-vdc-group" {
  org             = "{{.Org}}"
  edge_gateway_id = data.vcd_nsxt_edgegateway.testing-in-vdc-group.id

  mode         = "SLAAC"
  domain_names = ["non-existing.org.tld","fake.org.tld"]
  dns_servers  = ["2001:4860:4860::8888","2001:4860:4860::8844"]
}
`

const testAccVcdNsxtEdgeDhcpV6Step4 = testAccVcdNsxtEdgeDhcpV6Shared + `
resource "vcd_nsxt_edgegateway_dhcpv6" "testing-in-vdc" {
  org             = "{{.Org}}"
  edge_gateway_id = data.vcd_nsxt_edgegateway.testing-in-vdc.id

  mode = "SLAAC"
}

resource "vcd_nsxt_edgegateway_dhcpv6" "testing-in-vdc-group" {
  org             = "{{.Org}}"
  edge_gateway_id = data.vcd_nsxt_edgegateway.testing-in-vdc-group.id

  mode = "SLAAC"
}
`

const testAccVcdNsxtEdgeDhcpV6Step5 = testAccVcdNsxtEdgeDhcpV6Shared + `
resource "vcd_nsxt_edgegateway_dhcpv6" "testing-in-vdc" {
  org             = "{{.Org}}"
  edge_gateway_id = data.vcd_nsxt_edgegateway.testing-in-vdc.id

  mode = "DISABLED"
}

resource "vcd_nsxt_edgegateway_dhcpv6" "testing-in-vdc-group" {
  org             = "{{.Org}}"
  edge_gateway_id = data.vcd_nsxt_edgegateway.testing-in-vdc-group.id

  mode = "DISABLED"
}
`

func testAccCheckNsxtEdgeDhcpv6Destroy(vdcOrVdcGroupName, edgeGatewayName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*VCDClient)

		vdcOrVdcGroup, err := lookupVdcOrVdcGroup(conn, testConfig.VCD.Org, vdcOrVdcGroupName)
		if err != nil {
			return fmt.Errorf("unable to find VDC or VDC group %s: %s", vdcOrVdcGroupName, err)
		}

		edge, err := vdcOrVdcGroup.GetNsxtEdgeGatewayByName(edgeGatewayName)
		if err != nil {
			return fmt.Errorf(errorUnableToFindEdgeGateway, edgeGatewayName)
		}

		dhcpv6Config, err := edge.GetSlaacProfile()
		if err != nil {
			return fmt.Errorf("unable to get DHCPv6 config: %s", err)
		}

		if dhcpv6Config.Enabled {
			return fmt.Errorf("DHCPv6 is still enabled")
		}

		return nil
	}
}
