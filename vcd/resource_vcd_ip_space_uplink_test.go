//go:build network || nsxt || ALL || functional

package vcd

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcdIpSpaceUplink(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	vcdClient := createTemporaryVCDConnection(true)
	if vcdClient == nil {
		t.Skip(acceptanceTestsSkipped)
	}
	if vcdClient.Client.APIVCDMaxVersionIs("< 37.1") {
		t.Skipf("This test tests VCD 10.4.1+ (API V37.1+) features. Skipping.")
	}

	// String map to fill the template
	var params = StringMap{
		"TestName":            t.Name(),
		"NsxtManager":         testConfig.Nsxt.Manager,
		"NsxtTier0Router":     testConfig.Nsxt.Tier0router,
		"ExternalNetworkName": t.Name(),

		"Tags": "network nsxt",
	}
	testParamsNotEmpty(t, params)

	params["FuncName"] = t.Name() + "step1"
	configText1 := templateFill(testAccVcdIpSpaceUplinkStep1, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)

	params["FuncName"] = t.Name() + "step2"
	configText2 := templateFill(testAccVcdIpSpaceUplinkStep2, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText2)

	params["FuncName"] = t.Name() + "step3DS"
	configText3DS := templateFill(testAccVcdIpSpaceUplinkStep3DS, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 3: %s", configText3DS)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: configText1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_ip_space.space1", "id"),
					resource.TestCheckResourceAttrSet("vcd_external_network_v2.provider-gateway", "id"),
					resource.TestCheckResourceAttrSet("vcd_ip_space_uplink.u1", "id"),
					resource.TestCheckResourceAttr("vcd_ip_space_uplink.u1", "name", t.Name()),
					resource.TestCheckResourceAttr("vcd_ip_space_uplink.u1", "description", ""),
					resource.TestCheckResourceAttrSet("vcd_ip_space_uplink.u1", "external_network_id"),
					resource.TestCheckResourceAttrSet("vcd_ip_space_uplink.u1", "ip_space_id"),
					resource.TestCheckResourceAttr("vcd_ip_space_uplink.u1", "ip_space_type", "PUBLIC"),
					resource.TestCheckResourceAttrSet("vcd_ip_space_uplink.u1", "status"),
				),
			},
			{
				Config: configText2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_ip_space.space1", "id"),
					resource.TestCheckResourceAttrSet("vcd_external_network_v2.provider-gateway", "id"),
					resource.TestCheckResourceAttrSet("vcd_ip_space_uplink.u1", "id"),
					resource.TestCheckResourceAttr("vcd_ip_space_uplink.u1", "name", t.Name()+"-updated"),
					resource.TestCheckResourceAttr("vcd_ip_space_uplink.u1", "description", "description"),
					resource.TestCheckResourceAttrSet("vcd_ip_space_uplink.u1", "external_network_id"),
					resource.TestCheckResourceAttrSet("vcd_ip_space_uplink.u1", "ip_space_id"),
					resource.TestCheckResourceAttr("vcd_ip_space_uplink.u1", "ip_space_type", "PUBLIC"),
					resource.TestCheckResourceAttrSet("vcd_ip_space_uplink.u1", "status"),
				),
			},
			{
				ResourceName:      "vcd_ip_space_uplink.u1",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importCustomObject([]string{params["ExternalNetworkName"].(string), params["TestName"].(string) + "-updated"}),
			},
			{
				Config: configText3DS,
				Check: resource.ComposeTestCheckFunc(
					resourceFieldsEqual("data.vcd_ip_space_uplink.u1", "vcd_ip_space_uplink.u1", nil),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdIpSpaceUplinkPrereqs = `
data "vcd_nsxt_manager" "main" {
  name = "{{.NsxtManager}}"
}

data "vcd_nsxt_tier0_router" "router" {
  name            = "{{.NsxtTier0Router}}"
  nsxt_manager_id = data.vcd_nsxt_manager.main.id
}

resource "vcd_ip_space" "space1" {
  name = "{{.TestName}}"
  type = "PUBLIC"

  internal_scope = ["192.168.1.0/24"]

  route_advertisement_enabled = false
}

resource "vcd_external_network_v2" "provider-gateway" {
  name = "{{.ExternalNetworkName}}"

  nsxt_network {
    nsxt_manager_id      = data.vcd_nsxt_manager.main.id
    nsxt_tier0_router_id = data.vcd_nsxt_tier0_router.router.id
  }

  use_ip_spaces = true
}
`

const testAccVcdIpSpaceUplinkStep1 = testAccVcdIpSpaceUplinkPrereqs + `
resource "vcd_ip_space_uplink" "u1" {
  name                = "{{.TestName}}"
  external_network_id = vcd_external_network_v2.provider-gateway.id
  ip_space_id         = vcd_ip_space.space1.id
}
`

const testAccVcdIpSpaceUplinkStep2 = testAccVcdIpSpaceUplinkPrereqs + `
resource "vcd_ip_space_uplink" "u1" {
  name                = "{{.TestName}}-updated"
  description         = "description"
  external_network_id = vcd_external_network_v2.provider-gateway.id
  ip_space_id         = vcd_ip_space.space1.id
}
`

const testAccVcdIpSpaceUplinkStep3DS = testAccVcdIpSpaceUplinkPrereqs + `
data "vcd_ip_space_uplink" "u1" {
  name                = "{{.TestName}}-updated"
  external_network_id = vcd_external_network_v2.provider-gateway.id

  depends_on = [vcd_ip_space_uplink.u1]
}

resource "vcd_ip_space_uplink" "u1" {
  name                = "{{.TestName}}-updated"
  description         = "description"
  external_network_id = vcd_external_network_v2.provider-gateway.id
  ip_space_id         = vcd_ip_space.space1.id
}
`
