//go:build network || nsxt || ALL || functional

package vcd

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcdIpSpaceCustomQuota(t *testing.T) {
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
		"Org":                 testConfig.VCD.Org,
		"VDC":                 testConfig.Nsxt.Vdc,

		"Tags": "network nsxt",
	}
	testParamsNotEmpty(t, params)

	params["FuncName"] = t.Name() + "step1"
	configText1 := templateFill(testAccVcdIpSpaceCustomQuotaStep1, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)

	params["FuncName"] = t.Name() + "step2"
	configText2 := templateFill(testAccVcdIpSpaceCustomQuotaStep2, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText2)

	params["FuncName"] = t.Name() + "step3"
	configText3 := templateFill(testAccVcdIpSpaceCustomQuotaStep3, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 3: %s", configText3)

	params["FuncName"] = t.Name() + "step4"
	configText4 := templateFill(testAccVcdIpSpaceCustomQuotaStep4, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 4: %s", configText4)

	params["FuncName"] = t.Name() + "step5DS"
	configText5DS := templateFill(testAccVcdIpSpaceCustomQuotaStep5DS, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 5: %s", configText5DS)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_ip_space.space1", "id"),
					resource.TestCheckResourceAttrSet("vcd_external_network_v2.provider-gateway", "id"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_edgegateway.ip-space", "id"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.ip-space", "use_ip_spaces", "true"),
					resource.TestCheckResourceAttrSet("vcd_ip_space_uplink.u1", "id"),
					resource.TestCheckResourceAttr("vcd_ip_space_uplink.u1", "name", t.Name()),
					resource.TestCheckResourceAttr("vcd_ip_space_uplink.u1", "description", ""),
					resource.TestCheckResourceAttrSet("vcd_ip_space_uplink.u1", "external_network_id"),
					resource.TestCheckResourceAttrSet("vcd_ip_space_uplink.u1", "ip_space_id"),
					resource.TestCheckResourceAttr("vcd_ip_space_uplink.u1", "ip_space_type", "PUBLIC"),
					resource.TestCheckResourceAttrSet("vcd_ip_space_uplink.u1", "status"),
					resource.TestCheckResourceAttrSet("vcd_ip_space_custom_quota.test", "id"),
					resource.TestCheckResourceAttr("vcd_ip_space_custom_quota.test", "ip_range_quota", "6"),
					resource.TestCheckResourceAttr("vcd_ip_space_custom_quota.test", "ip_prefix_quota.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_ip_space_custom_quota.test", "ip_prefix_quota.*", map[string]string{
						"prefix_length": "29",
						"quota":         "9",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_ip_space_custom_quota.test", "ip_prefix_quota.*", map[string]string{
						"prefix_length": "30",
						"quota":         "8",
					}),
				),
			},
			{
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_ip_space.space1", "id"),
					resource.TestCheckResourceAttrSet("vcd_external_network_v2.provider-gateway", "id"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_edgegateway.ip-space", "id"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.ip-space", "use_ip_spaces", "true"),
					resource.TestCheckResourceAttrSet("vcd_ip_space_uplink.u1", "id"),
					resource.TestCheckResourceAttr("vcd_ip_space_uplink.u1", "name", t.Name()),
					resource.TestCheckResourceAttr("vcd_ip_space_uplink.u1", "description", ""),
					resource.TestCheckResourceAttrSet("vcd_ip_space_uplink.u1", "external_network_id"),
					resource.TestCheckResourceAttrSet("vcd_ip_space_uplink.u1", "ip_space_id"),
					resource.TestCheckResourceAttr("vcd_ip_space_uplink.u1", "ip_space_type", "PUBLIC"),
					resource.TestCheckResourceAttrSet("vcd_ip_space_uplink.u1", "status"),
					resource.TestCheckResourceAttrSet("vcd_ip_space_custom_quota.test", "id"),
					resource.TestCheckResourceAttr("vcd_ip_space_custom_quota.test", "ip_range_quota", "4"),
					resource.TestCheckResourceAttr("vcd_ip_space_custom_quota.test", "ip_prefix_quota.#", "0"),
				),
			},
			{
				Config: configText3,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_ip_space.space1", "id"),
					resource.TestCheckResourceAttrSet("vcd_external_network_v2.provider-gateway", "id"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_edgegateway.ip-space", "id"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.ip-space", "use_ip_spaces", "true"),
					resource.TestCheckResourceAttrSet("vcd_ip_space_uplink.u1", "id"),
					resource.TestCheckResourceAttr("vcd_ip_space_uplink.u1", "name", t.Name()),
					resource.TestCheckResourceAttr("vcd_ip_space_uplink.u1", "description", ""),
					resource.TestCheckResourceAttrSet("vcd_ip_space_uplink.u1", "external_network_id"),
					resource.TestCheckResourceAttrSet("vcd_ip_space_uplink.u1", "ip_space_id"),
					resource.TestCheckResourceAttr("vcd_ip_space_uplink.u1", "ip_space_type", "PUBLIC"),
					resource.TestCheckResourceAttrSet("vcd_ip_space_uplink.u1", "status"),
					resource.TestCheckResourceAttrSet("vcd_ip_space_custom_quota.test", "id"),
					resource.TestCheckResourceAttr("vcd_ip_space_custom_quota.test", "ip_range_quota", ""),
					resource.TestCheckResourceAttr("vcd_ip_space_custom_quota.test", "ip_prefix_quota.#", "0"),
				),
			},
			{
				Config: configText4,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_ip_space.space1", "id"),
					resource.TestCheckResourceAttrSet("vcd_external_network_v2.provider-gateway", "id"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_edgegateway.ip-space", "id"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.ip-space", "use_ip_spaces", "true"),
					resource.TestCheckResourceAttrSet("vcd_ip_space_uplink.u1", "id"),
					resource.TestCheckResourceAttr("vcd_ip_space_uplink.u1", "name", t.Name()),
					resource.TestCheckResourceAttr("vcd_ip_space_uplink.u1", "description", ""),
					resource.TestCheckResourceAttrSet("vcd_ip_space_uplink.u1", "external_network_id"),
					resource.TestCheckResourceAttrSet("vcd_ip_space_uplink.u1", "ip_space_id"),
					resource.TestCheckResourceAttr("vcd_ip_space_uplink.u1", "ip_space_type", "PUBLIC"),
					resource.TestCheckResourceAttrSet("vcd_ip_space_uplink.u1", "status"),
					resource.TestCheckResourceAttrSet("vcd_ip_space_custom_quota.test", "id"),
					resource.TestCheckResourceAttr("vcd_ip_space_custom_quota.test", "ip_range_quota", "23"),
					resource.TestCheckResourceAttr("vcd_ip_space_custom_quota.test", "ip_prefix_quota.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_ip_space_custom_quota.test", "ip_prefix_quota.*", map[string]string{
						"prefix_length": "29",
						"quota":         "17",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_ip_space_custom_quota.test", "ip_prefix_quota.*", map[string]string{
						"prefix_length": "30",
						"quota":         "15",
					}),
				),
			},
			{
				ResourceName:      "vcd_ip_space_custom_quota.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     strings.Join([]string{t.Name(), testConfig.VCD.Org}, ImportSeparator),
			},
			{
				Config: configText5DS,
				Check: resource.ComposeAggregateTestCheckFunc(
					resourceFieldsEqual("data.vcd_ip_space_custom_quota.test", "vcd_ip_space_custom_quota.test", nil),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdIpSpaceCustomQuotaPrereqs = `
data "vcd_nsxt_manager" "main" {
  name = "{{.NsxtManager}}"
}

data "vcd_nsxt_tier0_router" "router" {
  name            = "{{.NsxtTier0Router}}"
  nsxt_manager_id = data.vcd_nsxt_manager.main.id
}

data "vcd_org" "org1" {
  name = "{{.Org}}"
}

data "vcd_org_vdc" "vdc1" {
  org  = "{{.Org}}"
  name = "{{.VDC}}"
}

resource "vcd_ip_space" "space1" {
  name = "{{.TestName}}"
  type = "PUBLIC"

  internal_scope = ["192.168.1.0/24", "10.10.10.0/24", "11.11.11.0/24"]
  external_scope = "0.0.0.0/24"

  route_advertisement_enabled = false

  ip_prefix {
    default_quota = 2

    prefix {
      first_ip      = "192.168.1.100"
      prefix_length = 30
      prefix_count  = 4
    }

    prefix {
      first_ip      = "192.168.1.200"
      prefix_length = 30
      prefix_count  = 4
    }
  }

  ip_prefix {
    default_quota = -1

    prefix {
      first_ip      = "10.10.10.96"
      prefix_length = 29
      prefix_count  = 4
    }
  }

  ip_range {
    start_address = "11.11.11.100"
    end_address   = "11.11.11.110"
  }
}

resource "vcd_external_network_v2" "provider-gateway" {
  name = "{{.ExternalNetworkName}}"

  nsxt_network {
    nsxt_manager_id      = data.vcd_nsxt_manager.main.id
    nsxt_tier0_router_id = data.vcd_nsxt_tier0_router.router.id
  }

  use_ip_spaces = true
}
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

  depends_on = [vcd_ip_space_uplink.u1]
}
`

const testAccVcdIpSpaceCustomQuotaStep1 = testAccVcdIpSpaceCustomQuotaPrereqs + `
resource "vcd_ip_space_custom_quota" "test" {
  org_id         = data.vcd_org.org1.id
  ip_space_id    = vcd_ip_space.space1.id
  ip_range_quota = 6

  ip_prefix_quota {
    prefix_length = 29
    quota         = 9
  }

  ip_prefix_quota {
    prefix_length = 30
    quota         = 8
  }

  # Custom Quota can only be configured once Edge Gateway is created
  depends_on = [vcd_nsxt_edgegateway.ip-space]
}
`

const testAccVcdIpSpaceCustomQuotaStep2 = testAccVcdIpSpaceCustomQuotaPrereqs + `
resource "vcd_ip_space_custom_quota" "test" {
  org_id         = data.vcd_org.org1.id
  ip_space_id    = vcd_ip_space.space1.id
  ip_range_quota = 4

  # Custom Quota can only be configured once Edge Gateway is created
  depends_on = [vcd_nsxt_edgegateway.ip-space]
}
`

const testAccVcdIpSpaceCustomQuotaStep3 = testAccVcdIpSpaceCustomQuotaPrereqs + `
resource "vcd_ip_space_custom_quota" "test" {
  org_id      = data.vcd_org.org1.id
  ip_space_id = vcd_ip_space.space1.id

  # Custom Quota can only be configured once Edge Gateway is created
  depends_on = [vcd_nsxt_edgegateway.ip-space]
}
`

const testAccVcdIpSpaceCustomQuotaStep4 = testAccVcdIpSpaceCustomQuotaPrereqs + `
resource "vcd_ip_space_custom_quota" "test" {
  org_id      = data.vcd_org.org1.id
  ip_space_id = vcd_ip_space.space1.id

  ip_range_quota = 23

  ip_prefix_quota {
    prefix_length = 29
    quota         = 17
  }

  ip_prefix_quota {
    prefix_length = 30
    quota         = 15
  }

  # Custom Quota can only be configured once Edge Gateway is created
  depends_on = [vcd_nsxt_edgegateway.ip-space]
}
`

const testAccVcdIpSpaceCustomQuotaStep5DS = testAccVcdIpSpaceCustomQuotaStep4 + `
# skip-binary-test: Data Source test
data "vcd_ip_space_custom_quota" "test" {
  org_id      = data.vcd_org.org1.id
  ip_space_id = vcd_ip_space.space1.id
}
`
