//go:build network || nsxt || ALL || functional

package vcd

import (
	"fmt"
	"strings"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccVcdIpSpaceIpAllocation(t *testing.T) {
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
	configText1 := templateFill(testAccVcdIpSpaceIpAllocationStep1, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)

	params["FuncName"] = t.Name() + "step2"
	configText2 := templateFill(testAccVcdIpSpaceIpAllocationStep2, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText2)

	params["FuncName"] = t.Name() + "step3DS"
	configText3DS := templateFill(testAccVcdIpSpaceIpAllocationStep3DS, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 3: %s", configText3DS)

	params["FuncName"] = t.Name() + "step4"
	configText4 := templateFill(testAccVcdIpSpaceIpAllocationStep4, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 4: %s", configText4)

	params["FuncName"] = t.Name() + "step5"
	configText5 := templateFill(testAccVcdIpSpaceIpAllocationStep5, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 5: %s", configText5)

	params["FuncName"] = t.Name() + "step6"
	configText6 := templateFill(testAccVcdIpSpaceIpAllocationStep6, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 6: %s", configText6)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	// define structures for capturing IPs that are going to be used for import test
	floatingIpForImport := &testCachedFieldValue{}
	ipPrefixForImport := &testCachedFieldValue{}
	// capture IP Space ID field to ensure that it is not being recreated for the duration of this test
	ipSpaceId := &testCachedFieldValue{}
	ipSpaceUplinkId := &testCachedFieldValue{}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_ip_space.space1", "id"),
					ipSpaceId.cacheTestResourceFieldValue("vcd_ip_space.space1", "id"),
					resource.TestCheckResourceAttrSet("vcd_external_network_v2.provider-gateway", "id"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_edgegateway.ip-space", "id"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.ip-space", "uses_ip_spaces", "true"),
					resource.TestCheckResourceAttrSet("vcd_ip_space_uplink.u1", "id"),
					ipSpaceUplinkId.cacheTestResourceFieldValue("vcd_ip_space_uplink.u1", "id"),
					resource.TestCheckResourceAttr("vcd_ip_space_uplink.u1", "name", t.Name()),
					resource.TestCheckResourceAttr("vcd_ip_space_uplink.u1", "description", ""),
					resource.TestCheckResourceAttrSet("vcd_ip_space_uplink.u1", "external_network_id"),
					resource.TestCheckResourceAttrSet("vcd_ip_space_uplink.u1", "ip_space_id"),
					resource.TestCheckResourceAttr("vcd_ip_space_uplink.u1", "ip_space_type", "PUBLIC"),
					resource.TestCheckResourceAttrSet("vcd_ip_space_uplink.u1", "status"),
					resource.TestCheckResourceAttrSet("vcd_ip_space_ip_allocation.public-floating-ip", "id"),
					resource.TestCheckResourceAttrSet("vcd_ip_space_ip_allocation.public-floating-ip", "allocation_date"),
					resource.TestCheckResourceAttr("vcd_ip_space_ip_allocation.public-floating-ip", "type", "FLOATING_IP"),
					// usage_state is UNUSED because the state is updated during creation of this
					// resource and it is consumed in next dependent resource
					resource.TestCheckResourceAttr("vcd_ip_space_ip_allocation.public-floating-ip", "usage_state", "UNUSED"),
					resource.TestCheckResourceAttrSet("vcd_ip_space_ip_allocation.public-floating-ip", "ip_address"),
					resource.TestCheckResourceAttrSet("vcd_ip_space_ip_allocation.public-floating-ip-manual", "id"),
					resource.TestCheckResourceAttr("vcd_ip_space_ip_allocation.public-floating-ip-manual", "usage_state", "USED_MANUAL"),
					resource.TestCheckResourceAttr("vcd_ip_space_ip_allocation.public-floating-ip-manual", "type", "FLOATING_IP"),
					resource.TestCheckResourceAttr("vcd_ip_space_ip_allocation.public-floating-ip-manual", "description", "manually used floating IP"),
					resource.TestCheckResourceAttrSet("vcd_ip_space_ip_allocation.public-floating-ip-manual", "ip_address"),
					resource.TestCheckResourceAttrSet("vcd_ip_space_ip_allocation.public-ip-prefix", "id"),
					resource.TestCheckResourceAttr("vcd_ip_space_ip_allocation.public-ip-prefix", "usage_state", "UNUSED"),
					resource.TestCheckResourceAttr("vcd_ip_space_ip_allocation.public-ip-prefix", "type", "IP_PREFIX"),
					resource.TestCheckResourceAttrSet("vcd_ip_space_ip_allocation.public-ip-prefix", "ip_address"),
					resource.TestCheckResourceAttrSet("vcd_ip_space_ip_allocation.public-ip-prefix-manual", "id"),
					resource.TestCheckResourceAttr("vcd_ip_space_ip_allocation.public-ip-prefix-manual", "usage_state", "USED_MANUAL"),
					resource.TestCheckResourceAttr("vcd_ip_space_ip_allocation.public-ip-prefix-manual", "type", "IP_PREFIX"),
					resource.TestCheckResourceAttr("vcd_ip_space_ip_allocation.public-ip-prefix-manual", "description", "manually used IP Prefix"),
					resource.TestCheckResourceAttrSet("vcd_ip_space_ip_allocation.public-ip-prefix-manual", "ip_address"),
					resource.TestCheckResourceAttrSet("vcd_network_routed_v2.using-public-prefix", "id"),
				),
			},
			{
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_ip_space.space1", "id"),
					ipSpaceId.testCheckCachedResourceFieldValue("vcd_ip_space.space1", "id"),
					resource.TestCheckResourceAttrSet("vcd_external_network_v2.provider-gateway", "id"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_edgegateway.ip-space", "id"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.ip-space", "uses_ip_spaces", "true"),
					resource.TestCheckResourceAttrSet("vcd_ip_space_uplink.u1", "id"),
					ipSpaceUplinkId.testCheckCachedResourceFieldValue("vcd_ip_space_uplink.u1", "id"),
					resource.TestCheckResourceAttr("vcd_ip_space_uplink.u1", "name", t.Name()),
					resource.TestCheckResourceAttr("vcd_ip_space_uplink.u1", "description", ""),
					resource.TestCheckResourceAttrSet("vcd_ip_space_uplink.u1", "external_network_id"),
					resource.TestCheckResourceAttrSet("vcd_ip_space_uplink.u1", "ip_space_id"),
					resource.TestCheckResourceAttr("vcd_ip_space_uplink.u1", "ip_space_type", "PUBLIC"),
					resource.TestCheckResourceAttrSet("vcd_ip_space_uplink.u1", "status"),
					resource.TestCheckResourceAttrSet("vcd_ip_space_ip_allocation.public-floating-ip", "id"),
					resource.TestCheckResourceAttrSet("vcd_ip_space_ip_allocation.public-floating-ip", "allocation_date"),
					resource.TestCheckResourceAttr("vcd_ip_space_ip_allocation.public-floating-ip", "type", "FLOATING_IP"),
					resource.TestCheckResourceAttr("vcd_ip_space_ip_allocation.public-floating-ip", "usage_state", "USED"),
					resource.TestCheckResourceAttrSet("vcd_ip_space_ip_allocation.public-floating-ip", "ip_address"),
					resource.TestCheckResourceAttrSet("vcd_ip_space_ip_allocation.public-floating-ip-manual", "id"),
					resource.TestCheckResourceAttr("vcd_ip_space_ip_allocation.public-floating-ip-manual", "usage_state", "UNUSED"),
					resource.TestCheckResourceAttr("vcd_ip_space_ip_allocation.public-floating-ip-manual", "type", "FLOATING_IP"),
					resource.TestCheckResourceAttr("vcd_ip_space_ip_allocation.public-floating-ip-manual", "description", "manually used floating IP"),
					resource.TestCheckResourceAttrSet("vcd_ip_space_ip_allocation.public-floating-ip-manual", "ip_address"),
					resource.TestCheckResourceAttrSet("vcd_ip_space_ip_allocation.public-ip-prefix", "id"),
					resource.TestCheckResourceAttr("vcd_ip_space_ip_allocation.public-ip-prefix", "usage_state", "USED"),
					resource.TestCheckResourceAttr("vcd_ip_space_ip_allocation.public-ip-prefix", "type", "IP_PREFIX"),
					resource.TestCheckResourceAttrSet("vcd_ip_space_ip_allocation.public-ip-prefix", "ip_address"),
					resource.TestCheckResourceAttrSet("vcd_ip_space_ip_allocation.public-ip-prefix-manual", "id"),
					resource.TestCheckResourceAttr("vcd_ip_space_ip_allocation.public-ip-prefix-manual", "usage_state", "UNUSED"),
					resource.TestCheckResourceAttr("vcd_ip_space_ip_allocation.public-ip-prefix-manual", "type", "IP_PREFIX"),
					resource.TestCheckResourceAttr("vcd_ip_space_ip_allocation.public-ip-prefix-manual", "description", "manually used IP Prefix"),
					resource.TestCheckResourceAttrSet("vcd_ip_space_ip_allocation.public-ip-prefix-manual", "ip_address"),
					resource.TestCheckResourceAttrSet("vcd_network_routed_v2.using-public-prefix", "id"),

					// Capture IP addresses for import testing in next few steps
					floatingIpForImport.cacheTestResourceFieldValue("vcd_ip_space_ip_allocation.public-floating-ip", "ip_address"),
					ipPrefixForImport.cacheTestResourceFieldValue("vcd_ip_space_ip_allocation.public-ip-prefix", "ip_address"),
				),
			},
			// org-name.ip-space-name.ip-allocation-type.ip-allocation-ip

			{ // terraform import vcd_ip_space_ip_allocation.ip my-org.my-ip-space.FLOATING_IP.X.X.X.X
				ResourceName:      "vcd_ip_space_ip_allocation.public-floating-ip",
				ImportState:       true,
				ImportStateVerify: true,
				// this field is ignored because its value is compared against value of step1, when it was created
				// however at step 2 it is not being used anymore
				ImportStateVerifyIgnore: []string{"used_by_id"},
				ImportStateIdFunc:       importCustomIpAllocationFunc([]string{testConfig.VCD.Org, t.Name(), "FLOATING_IP"}, floatingIpForImport),
			},
			{ // terraform import vcd_ip_space_ip_allocation.ip my-org.my-ip-space.IP_PREFIX.Y.Y.Y.Y/ZZ
				ResourceName:      "vcd_ip_space_ip_allocation.public-ip-prefix",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importCustomIpAllocationFunc([]string{testConfig.VCD.Org, t.Name(), "IP_PREFIX"}, ipPrefixForImport),
			},
			{
				Config: configText3DS,
				Check: resource.ComposeAggregateTestCheckFunc(
					ipSpaceId.testCheckCachedResourceFieldValue("vcd_ip_space.space1", "id"),
					ipSpaceUplinkId.testCheckCachedResourceFieldValue("vcd_ip_space_uplink.u1", "id"),
					resourceFieldsEqual("data.vcd_ip_space_uplink.u1", "vcd_ip_space_uplink.u1", nil),
					resourceFieldsEqual("data.vcd_nsxt_edgegateway.ip-space", "vcd_nsxt_edgegateway.ip-space", []string{"%"}),
					resourceFieldsEqual("data.vcd_ip_space_ip_allocation.public-floating-ip", "vcd_ip_space_ip_allocation.public-floating-ip", nil),
					resourceFieldsEqual("data.vcd_ip_space_ip_allocation.public-floating-ip-manual", "vcd_ip_space_ip_allocation.public-floating-ip-manual", nil),
					resourceFieldsEqual("data.vcd_ip_space_ip_allocation.public-ip-prefix", "vcd_ip_space_ip_allocation.public-ip-prefix", nil),
					resourceFieldsEqual("data.vcd_ip_space_ip_allocation.public-ip-prefix-manual", "vcd_ip_space_ip_allocation.public-ip-prefix-manual", nil),
					resourceFieldsEqual("data.vcd_network_routed_v2.using-public-prefix", "vcd_network_routed_v2.using-public-prefix", []string{"%"}),
				),
			},
			{
				Config: configText4,
				Check: resource.ComposeAggregateTestCheckFunc(
					ipSpaceId.testCheckCachedResourceFieldValue("vcd_ip_space.space1", "id"),
					ipSpaceUplinkId.testCheckCachedResourceFieldValue("vcd_ip_space_uplink.u1", "id"),
					resource.TestCheckResourceAttr("vcd_ip_space.space1", "ip_range.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_ip_space.space1", "ip_range.*", map[string]string{
						"start_address": "11.11.11.100",
						"end_address":   "11.11.11.109",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_ip_space.space1", "ip_prefix.*.prefix.*", map[string]string{
						"first_ip":      "192.168.1.100",
						"prefix_length": "30",
						"prefix_count":  "4",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_ip_space.space1", "ip_prefix.*.prefix.*", map[string]string{
						"first_ip":      "10.10.10.96",
						"prefix_length": "29",
						"prefix_count":  "3",
					}),
				),
			},
			{
				Config: configText5,
				Check: resource.ComposeAggregateTestCheckFunc(
					ipSpaceId.testCheckCachedResourceFieldValue("vcd_ip_space.space1", "id"),
					ipSpaceUplinkId.testCheckCachedResourceFieldValue("vcd_ip_space_uplink.u1", "id"),
					resource.TestCheckResourceAttr("vcd_ip_space.space1", "ip_range.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_ip_space.space1", "ip_range.*", map[string]string{
						"start_address": "11.11.11.100",
						"end_address":   "11.11.11.111",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_ip_space.space1", "ip_range.*", map[string]string{
						"start_address": "10.10.10.13",
						"end_address":   "10.10.10.14",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_ip_space.space1", "ip_prefix.*.prefix.*", map[string]string{
						"first_ip":      "192.168.1.100",
						"prefix_length": "30",
						"prefix_count":  "3",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_ip_space.space1", "ip_prefix.*.prefix.*", map[string]string{
						"first_ip":      "10.10.10.96",
						"prefix_length": "29",
						"prefix_count":  "2",
					}),
				),
			},
			{
				Config: configText6,
				Check: resource.ComposeAggregateTestCheckFunc(
					ipSpaceId.testCheckCachedResourceFieldValue("vcd_ip_space.space1", "id"),
					ipSpaceUplinkId.testCheckCachedResourceFieldValue("vcd_ip_space_uplink.u1", "id"),
					resource.TestCheckResourceAttr("vcd_ip_space.space1", "ip_range.#", "0"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_ip_space.space1", "ip_prefix.*.prefix.*", map[string]string{
						"first_ip":      "192.168.1.104",
						"prefix_length": "29",
						"prefix_count":  "3",
					}),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdIpSpaceIpAllocationPrereqs = `
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

  ip_range {
    start_address = "11.11.11.120"
    end_address   = "11.11.11.123"
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
`

const testAccVcdIpSpaceIpAllocationStep1 = testAccVcdIpSpaceIpAllocationPrereqs + `
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

resource "vcd_ip_space_ip_allocation" "public-floating-ip" {
  org_id      = data.vcd_org.org1.id
  ip_space_id = vcd_ip_space.space1.id
  type        = "FLOATING_IP"

  depends_on = [vcd_nsxt_edgegateway.ip-space]
}

resource "vcd_ip_space_ip_allocation" "public-floating-ip-2" {
  org_id      = data.vcd_org.org1.id
  ip_space_id = vcd_ip_space.space1.id
  type        = "FLOATING_IP"

  depends_on = [vcd_nsxt_edgegateway.ip-space]
}

resource "vcd_ip_space_ip_allocation" "public-floating-ip-manual" {
  org_id      = data.vcd_org.org1.id
  ip_space_id = vcd_ip_space.space1.id
  type        = "FLOATING_IP"
  usage_state = "USED_MANUAL"
  description = "manually used floating IP"

  depends_on = [vcd_nsxt_edgegateway.ip-space]
}

resource "vcd_nsxt_nat_rule" "dnat-floating-ip" {
  org             = "{{.Org}}"
  edge_gateway_id = vcd_nsxt_edgegateway.ip-space.id

  name      = "{{.TestName}}"
  rule_type = "DNAT"

  # Using Floating IP From IP Space
  external_address = vcd_ip_space_ip_allocation.public-floating-ip.ip_address
  internal_address = "77.77.77.1"
  logging          = true
}

resource "vcd_ip_space_ip_allocation" "public-ip-prefix" {
  org_id        = data.vcd_org.org1.id
  ip_space_id   = vcd_ip_space.space1.id
  type          = "IP_PREFIX"
  prefix_length = 29

  depends_on = [vcd_nsxt_edgegateway.ip-space]
}

resource "vcd_network_routed_v2" "using-public-prefix" {
  org             = "{{.Org}}"
  name            = "{{.TestName}}"
  edge_gateway_id = vcd_nsxt_edgegateway.ip-space.id
  gateway         = cidrhost(vcd_ip_space_ip_allocation.public-ip-prefix.ip_address, 1)
  prefix_length   = split("/", vcd_ip_space_ip_allocation.public-ip-prefix.ip_address)[1]

  static_ip_pool {
    start_address = cidrhost(vcd_ip_space_ip_allocation.public-ip-prefix.ip_address, 2)
    end_address   = cidrhost(vcd_ip_space_ip_allocation.public-ip-prefix.ip_address, 4)
  }
}

resource "vcd_ip_space_ip_allocation" "public-ip-prefix-manual" {
  org_id        = data.vcd_org.org1.id
  ip_space_id   = vcd_ip_space.space1.id
  type          = "IP_PREFIX"
  prefix_length = 30
  usage_state   = "USED_MANUAL"
  description   = "manually used IP Prefix"

  depends_on = [vcd_nsxt_edgegateway.ip-space]
}
`

const testAccVcdIpSpaceIpAllocationStep2 = testAccVcdIpSpaceIpAllocationPrereqs + `
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

resource "vcd_ip_space_ip_allocation" "public-floating-ip" {
  org_id      = data.vcd_org.org1.id
  ip_space_id = vcd_ip_space.space1.id
  type        = "FLOATING_IP"

  depends_on = [vcd_nsxt_edgegateway.ip-space]
}

resource "vcd_ip_space_ip_allocation" "public-floating-ip-2" {
  org_id      = data.vcd_org.org1.id
  ip_space_id = vcd_ip_space.space1.id
  type        = "FLOATING_IP"

  depends_on = [vcd_nsxt_edgegateway.ip-space]
}

resource "vcd_ip_space_ip_allocation" "public-floating-ip-manual" {
  org_id      = data.vcd_org.org1.id
  ip_space_id = vcd_ip_space.space1.id
  type        = "FLOATING_IP"
  usage_state = "UNUSED"

  depends_on = [vcd_nsxt_edgegateway.ip-space]
}

resource "vcd_nsxt_nat_rule" "dnat-floating-ip" {
  org             = "{{.Org}}"
  edge_gateway_id = vcd_nsxt_edgegateway.ip-space.id

  name      = "{{.TestName}}"
  rule_type = "DNAT"

  # Using Floating IP From IP Space
  external_address = vcd_ip_space_ip_allocation.public-floating-ip-2.ip_address
  internal_address = "77.77.77.1"
  logging          = true
}

resource "vcd_ip_space_ip_allocation" "public-ip-prefix" {
  org_id        = data.vcd_org.org1.id
  ip_space_id   = vcd_ip_space.space1.id
  type          = "IP_PREFIX"
  prefix_length = 29

  depends_on = [vcd_nsxt_edgegateway.ip-space]
}

resource "vcd_network_routed_v2" "using-public-prefix" {
  org             = "{{.Org}}"
  name            = "{{.TestName}}"
  edge_gateway_id = vcd_nsxt_edgegateway.ip-space.id
  gateway         = cidrhost(vcd_ip_space_ip_allocation.public-ip-prefix.ip_address, 1)
  prefix_length   = split("/", vcd_ip_space_ip_allocation.public-ip-prefix.ip_address)[1]

  static_ip_pool {
    start_address = cidrhost(vcd_ip_space_ip_allocation.public-ip-prefix.ip_address, 2)
    end_address   = cidrhost(vcd_ip_space_ip_allocation.public-ip-prefix.ip_address, 4)
  }
}

resource "vcd_ip_space_ip_allocation" "public-ip-prefix-manual" {
  org_id        = data.vcd_org.org1.id
  ip_space_id   = vcd_ip_space.space1.id
  type          = "IP_PREFIX"
  prefix_length = 30
  usage_state   = "UNUSED"

  depends_on = [vcd_nsxt_edgegateway.ip-space]
}
`

const testAccVcdIpSpaceIpAllocationStep3DS = testAccVcdIpSpaceIpAllocationStep1 + `
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

data "vcd_ip_space_ip_allocation" "public-floating-ip" {
  org_id      = data.vcd_org.org1.id
  ip_space_id = vcd_ip_space.space1.id
  type        = "FLOATING_IP"
  ip_address  = vcd_ip_space_ip_allocation.public-floating-ip.ip_address

  depends_on = [vcd_nsxt_edgegateway.ip-space]
}

data "vcd_ip_space_ip_allocation" "public-floating-ip-manual" {
  org_id      = data.vcd_org.org1.id
  ip_space_id = vcd_ip_space.space1.id
  type        = "FLOATING_IP"
  ip_address  = vcd_ip_space_ip_allocation.public-floating-ip-manual.ip_address

  depends_on = [vcd_nsxt_edgegateway.ip-space]
}

data "vcd_ip_space_ip_allocation" "public-ip-prefix" {
  org_id      = data.vcd_org.org1.id
  ip_space_id = vcd_ip_space.space1.id
  type        = "IP_PREFIX"

  ip_address = vcd_ip_space_ip_allocation.public-ip-prefix.ip_address

  depends_on = [vcd_nsxt_edgegateway.ip-space]
}

data "vcd_network_routed_v2" "using-public-prefix" {
  org             = "{{.Org}}"
  name            = "{{.TestName}}"
  edge_gateway_id = vcd_nsxt_edgegateway.ip-space.id
}

data "vcd_ip_space_ip_allocation" "public-ip-prefix-manual" {
  org_id      = data.vcd_org.org1.id
  ip_space_id = vcd_ip_space.space1.id
  type        = "IP_PREFIX"
  ip_address  = vcd_ip_space_ip_allocation.public-ip-prefix-manual.ip_address

  depends_on = [vcd_nsxt_edgegateway.ip-space]
}
`

const step4plusPrerequisites = `
# skip-binary-test: Only required to test Update
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

resource "vcd_ip_space_ip_allocation" "public-floating-ip" {
  org_id      = data.vcd_org.org1.id
  ip_space_id = vcd_ip_space.space1.id
  type        = "FLOATING_IP"

  depends_on = [vcd_nsxt_edgegateway.ip-space]
}

resource "vcd_ip_space_ip_allocation" "public-floating-ip-2" {
  org_id      = data.vcd_org.org1.id
  ip_space_id = vcd_ip_space.space1.id
  type        = "FLOATING_IP"

  depends_on = [vcd_nsxt_edgegateway.ip-space]
}

resource "vcd_ip_space_ip_allocation" "public-floating-ip-manual" {
  org_id      = data.vcd_org.org1.id
  ip_space_id = vcd_ip_space.space1.id
  type        = "FLOATING_IP"
  usage_state = "USED_MANUAL"
  description = "manually used floating IP"

  depends_on = [vcd_nsxt_edgegateway.ip-space]
}

resource "vcd_nsxt_nat_rule" "dnat-floating-ip" {
  org             = "{{.Org}}"
  edge_gateway_id = vcd_nsxt_edgegateway.ip-space.id

  name      = "{{.TestName}}"
  rule_type = "DNAT"

  # Using Floating IP From IP Space
  external_address = vcd_ip_space_ip_allocation.public-floating-ip.ip_address
  internal_address = "77.77.77.1"
  logging          = true
}

resource "vcd_ip_space_ip_allocation" "public-ip-prefix" {
  org_id        = data.vcd_org.org1.id
  ip_space_id   = vcd_ip_space.space1.id
  type          = "IP_PREFIX"
  prefix_length = 29

  depends_on = [vcd_nsxt_edgegateway.ip-space]
}

resource "vcd_network_routed_v2" "using-public-prefix" {
  org             = "{{.Org}}"
  name            = "{{.TestName}}"
  edge_gateway_id = vcd_nsxt_edgegateway.ip-space.id
  gateway         = cidrhost(vcd_ip_space_ip_allocation.public-ip-prefix.ip_address, 1)
  prefix_length   = split("/", vcd_ip_space_ip_allocation.public-ip-prefix.ip_address)[1]

  static_ip_pool {
    start_address = cidrhost(vcd_ip_space_ip_allocation.public-ip-prefix.ip_address, 2)
    end_address   = cidrhost(vcd_ip_space_ip_allocation.public-ip-prefix.ip_address, 4)
  }
}

resource "vcd_ip_space_ip_allocation" "public-ip-prefix-manual" {
  org_id        = data.vcd_org.org1.id
  ip_space_id   = vcd_ip_space.space1.id
  type          = "IP_PREFIX"
  prefix_length = 30
  usage_state   = "USED_MANUAL"
  description   = "manually used IP Prefix"

  depends_on = [vcd_nsxt_edgegateway.ip-space]
}
`

// testAccVcdIpSpaceIpAllocationStep4 contains all data and its goal is to verify that IP Space
// definitions can be updated once there are allocated IPs
const testAccVcdIpSpaceIpAllocationStep4 = step4plusPrerequisites + `
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
  }

  ip_prefix {
    default_quota = -1

    prefix {
      first_ip      = "10.10.10.96"
      prefix_length = 29
      prefix_count  = 3
    }
  }

  ip_range_quota = 2

  ip_range {
    start_address = "11.11.11.100"
    end_address   = "11.11.11.109"
  }
}
`

const testAccVcdIpSpaceIpAllocationStep5 = step4plusPrerequisites + `
resource "vcd_ip_space" "space1" {
  name = "{{.TestName}}"
  type = "PUBLIC"

  internal_scope = ["192.168.1.0/24", "10.10.10.0/24", "11.11.11.0/24"]
  external_scope = "0.0.0.0/24"

  route_advertisement_enabled = false

  ip_prefix {
    default_quota = 5

    prefix {
      first_ip      = "192.168.1.100"
      prefix_length = 30
      prefix_count  = 3
    }
  }

  ip_prefix {
    default_quota = 0

    prefix {
      first_ip      = "10.10.10.96"
      prefix_length = 29
      prefix_count  = 2
    }
  }

  ip_range_quota = 2

  ip_range {
    start_address = "11.11.11.100"
    end_address   = "11.11.11.111"
  }

  ip_range {
    start_address = "10.10.10.13"
    end_address   = "10.10.10.14"
  }
}
`

const testAccVcdIpSpaceIpAllocationStep6 = `
# skip-binary-test: Only required to test Update
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

resource "vcd_ip_space" "space1" {
  name = "{{.TestName}}"
  type = "PUBLIC"

  internal_scope = ["192.168.1.0/24", "10.10.10.0/24", "11.11.11.0/24"]
  external_scope = "0.0.0.0/24"

  route_advertisement_enabled = false

  ip_prefix {
    default_quota = 5

    prefix {
      first_ip      = "192.168.1.104"
      prefix_length = 29
      prefix_count  = 3
    }
  }
}
`

// importCustomIpAllocationFunc has specific that it accepts cachedField address so that late
// evaluation of this field can be done. It is needed so that VCD Allocation IP address can be
// captured and then used in the import. It cannot be known in advance as it is up to VCD to
// allocate IP address from IP Space
func importCustomIpAllocationFunc(path []string, cachedField *testCachedFieldValue) resource.ImportStateIdFunc {
	return func(*terraform.State) (string, error) {
		completePath := strings.Join(path, ImportSeparator) + ImportSeparator + cachedField.String()
		if vcdTestVerbose {
			fmt.Printf("# Import path '%s'\n", completePath)
		}
		return completePath, nil
	}
}

func stateDumper() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		spew.Dump(s)
		return nil
	}
}
