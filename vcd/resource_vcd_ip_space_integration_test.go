//go:build network || nsxt || ALL || functional

package vcd

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcdIpSpaceIntegration(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	// String map to fill the template
	var params = StringMap{
		"TestName":            t.Name(),
		"NsxtManager":         testConfig.Nsxt.Manager,
		"NsxtTier0Router":     testConfig.Nsxt.Tier0router,
		"ExternalNetworkName": t.Name(),
		"Org":                 testConfig.VCD.Org,
		"VDC":                 testConfig.Nsxt.Vdc,

		"Tags": "network nsxt",
	}
	testParamsNotEmpty(t, params)

	params["FuncName"] = t.Name() + "step1"
	configText1 := templateFill(testAccVcdIpSpaceIntegrationStep1, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)

	// params["FuncName"] = t.Name() + "step2"
	// configText2 := templateFill(testAccVcdIpSpaceIntegrationStep2, params)
	// debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText2)

	params["FuncName"] = t.Name() + "step3DS"
	configText3DS := templateFill(testAccVcdIpSpaceIntegrationStep3DS, params)
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
					resource.TestCheckResourceAttrSet("vcd_nsxt_edgegateway.ip-space", "id"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.ip-space", "uses_ip_spaces", "true"),

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
				Config: configText3DS,
				Check: resource.ComposeTestCheckFunc(
					resourceFieldsEqual("data.vcd_ip_space_uplink.u1", "vcd_ip_space_uplink.u1", nil),
					resourceFieldsEqual("data.vcd_nsxt_edgegateway.ip-space", "vcd_nsxt_edgegateway.ip-space", []string{"%"}),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdIpSpaceIntegrationPrereqs = `
data "vcd_nsxt_manager" "main" {
  name = "{{.NsxtManager}}"
}

data "vcd_nsxt_tier0_router" "router" {
  name            = "{{.NsxtTier0Router}}"
  nsxt_manager_id = data.vcd_nsxt_manager.main.id
}

data "vcd_org_vdc" "vdc1" {
  org  = "{{.Org}}"
  name = "{{.VDC}}"
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

const testAccVcdIpSpaceIntegrationStep1 = testAccVcdIpSpaceIntegrationPrereqs + `
resource "vcd_ip_space_uplink" "u1" {
  name                = "{{.TestName}}"
  external_network_id = vcd_external_network_v2.provider-gateway.id
  ip_space_id         = vcd_ip_space.space1.id
}

resource "vcd_nsxt_edgegateway" "ip-space" {
  org                 = "{{.Org}}"
  name                = "{{.TestName}}"
  owner_id            = data.vcd_org_vdc.vdc1.id
  external_network_id = vcd_external_network_v2.provider-gateway.id
}
`

const testAccVcdIpSpaceIntegrationStep2 = testAccVcdIpSpaceIntegrationPrereqs + `
resource "vcd_ip_space_uplink" "u1" {
  name                = "{{.TestName}}-updated"
  description         = "description"
  external_network_id = vcd_external_network_v2.provider-gateway.id
  ip_space_id         = vcd_ip_space.space1.id
}

resource "vcd_nsxt_edgegateway" "ip-space" {
  org                 = "{{.Org}}"
  name                = "{{.TestName}}"
  owner_id            = data.vcd_org_vdc.vdc1.id
  external_network_id = vcd_external_network_v2.provider-gateway.id
}
`

const testAccVcdIpSpaceIntegrationStep3DS = testAccVcdIpSpaceIntegrationStep1 + `
# skip-binary-test: Data Source test
data "vcd_ip_space_uplink" "u1" {
  name                = "{{.TestName}}"
  external_network_id = vcd_external_network_v2.provider-gateway.id
}

data "vcd_nsxt_edgegateway" "ip-space" {
  org      = "{{.Org}}"
  name     = vcd_nsxt_edgegateway.ip-space.name
  owner_id = data.vcd_org_vdc.vdc1.id
}
`
