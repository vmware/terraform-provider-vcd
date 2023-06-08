//go:build network || nsxt || ALL || functional

package vcd

import (
	"fmt"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccVcdIpSpaceIpAllocation(t *testing.T) {
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

	params["FuncName"] = t.Name() + "step2"
	configText2 := templateFill(testAccVcdIpSpaceIntegrationStep2, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText2)

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
				Check: resource.ComposeAggregateTestCheckFunc(
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

					resource.TestCheckResourceAttrSet("vcd_ip_space_ip_allocation.public-floating-ip", "id"),
					resource.TestCheckResourceAttrSet("vcd_ip_space_ip_allocation.public-floating-ip", "allocation_date"),
					resource.TestCheckResourceAttr("vcd_ip_space_ip_allocation.public-floating-ip", "type", "FLOATING_IP"),
					// usage_state is UNUSED because the state is updated during creation of this
					// resource and it is consumed in next dependent resource
					resource.TestCheckResourceAttr("vcd_ip_space_ip_allocation.public-floating-ip", "usage_state", "UNUSED"),
					// resource.TestCheckResourceAttr("vcd_ip_space_ip_allocation.public-floating-ip", "quantity", "1"),
					resource.TestCheckResourceAttrSet("vcd_ip_space_ip_allocation.public-floating-ip", "ip_address"),

					resource.TestCheckResourceAttrSet("vcd_ip_space_ip_allocation.public-floating-ip-manual", "id"),
					resource.TestCheckResourceAttr("vcd_ip_space_ip_allocation.public-floating-ip-manual", "usage_state", "USED_MANUAL"),
					resource.TestCheckResourceAttr("vcd_ip_space_ip_allocation.public-floating-ip-manual", "type", "FLOATING_IP"),
					resource.TestCheckResourceAttr("vcd_ip_space_ip_allocation.public-floating-ip-manual", "description", "manually used floating IP"),
					resource.TestCheckResourceAttrSet("vcd_ip_space_ip_allocation.public-floating-ip-manual", "ip_address"),

					// public-ip-prefix
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
					// sleepTester(5*time.Minute),
					// stateDumper(),
				),
			},
			{
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
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

					resource.TestCheckResourceAttrSet("vcd_ip_space_ip_allocation.public-floating-ip", "id"),
					resource.TestCheckResourceAttrSet("vcd_ip_space_ip_allocation.public-floating-ip", "allocation_date"),
					resource.TestCheckResourceAttr("vcd_ip_space_ip_allocation.public-floating-ip", "type", "FLOATING_IP"),
					// usage_state is UNUSED because the state is updated during creation of this
					// resource and it is consumed in next dependent resource
					resource.TestCheckResourceAttr("vcd_ip_space_ip_allocation.public-floating-ip", "usage_state", "USED"),
					resource.TestCheckResourceAttrSet("vcd_ip_space_ip_allocation.public-floating-ip", "ip_address"),

					resource.TestCheckResourceAttrSet("vcd_ip_space_ip_allocation.public-floating-ip-manual", "id"),
					resource.TestCheckResourceAttr("vcd_ip_space_ip_allocation.public-floating-ip-manual", "usage_state", "UNUSED"),
					resource.TestCheckResourceAttr("vcd_ip_space_ip_allocation.public-floating-ip-manual", "type", "FLOATING_IP"),
					resource.TestCheckResourceAttr("vcd_ip_space_ip_allocation.public-floating-ip-manual", "description", "manually used floating IP"),
					resource.TestCheckResourceAttrSet("vcd_ip_space_ip_allocation.public-floating-ip-manual", "ip_address"),

					// public-ip-prefix
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
					// sleepTester(5*time.Minute),
					// stateDumper(),
				),
			},
			{
				Config: configText3DS,
				Check: resource.ComposeAggregateTestCheckFunc(
					resourceFieldsEqual("data.vcd_ip_space_uplink.u1", "vcd_ip_space_uplink.u1", nil),
					resourceFieldsEqual("data.vcd_nsxt_edgegateway.ip-space", "vcd_nsxt_edgegateway.ip-space", []string{"%"}),

					resourceFieldsEqual("data.vcd_ip_space_ip_allocation.public-floating-ip", "vcd_ip_space_ip_allocation.public-floating-ip", nil),
					resourceFieldsEqual("data.vcd_ip_space_ip_allocation.public-floating-ip-manual", "vcd_ip_space_ip_allocation.public-floating-ip-manual", nil),
					resourceFieldsEqual("data.vcd_ip_space_ip_allocation.public-ip-prefix", "vcd_ip_space_ip_allocation.public-ip-prefix", nil),
					resourceFieldsEqual("data.vcd_ip_space_ip_allocation.public-ip-prefix-manual", "vcd_ip_space_ip_allocation.public-ip-prefix-manual", nil),
					resourceFieldsEqual("data.vcd_network_routed_v2.using-public-prefix", "vcd_network_routed_v2.using-public-prefix", []string{"%"}),
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

  internal_scope = ["192.168.1.0/24","10.10.10.0/24", "11.11.11.0/24"]
  external_scope = "0.0.0.0/24"

  route_advertisement_enabled = false

  ip_prefix {
	default_quota = 2

	prefix {
		first_ip = "192.168.1.100"
		prefix_length = 30
		prefix_count = 4
	}

	prefix {
		first_ip = "192.168.1.200"
		prefix_length = 30
		prefix_count = 4
	}
  }

  ip_prefix {
	default_quota = -1

	prefix {
		first_ip = "10.10.10.96"
		prefix_length = 29
		prefix_count = 4
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

  depends_on = [vcd_ip_space_uplink.u1]
}

resource "vcd_ip_space_ip_allocation" "public-floating-ip" {
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

  name        = "{{.TestName}}"
  rule_type   = "DNAT"

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
  usage_state = "USED_MANUAL"
  description = "manually used IP Prefix"

  depends_on = [vcd_nsxt_edgegateway.ip-space]
}
`

const testAccVcdIpSpaceIntegrationStep2 = testAccVcdIpSpaceIntegrationPrereqs + `
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

  name        = "{{.TestName}}"
  rule_type   = "DNAT"

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
  usage_state = "UNUSED"

  depends_on = [vcd_nsxt_edgegateway.ip-space]
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

data "vcd_ip_space_ip_allocation" "public-floating-ip" {
  org_id      = data.vcd_org.org1.id
  ip_space_id = vcd_ip_space.space1.id
  type        = "FLOATING_IP"
  ip_address = vcd_ip_space_ip_allocation.public-floating-ip.ip_address

  depends_on = [vcd_nsxt_edgegateway.ip-space]
}

data "vcd_ip_space_ip_allocation" "public-floating-ip-manual" {
  org_id      = data.vcd_org.org1.id
  ip_space_id = vcd_ip_space.space1.id
  type        = "FLOATING_IP"
  #usage_state = "USED_MANUAL"
  #description = "manually used floating IP"
  ip_address = vcd_ip_space_ip_allocation.public-floating-ip-manual.ip_address

  depends_on = [vcd_nsxt_edgegateway.ip-space]
}

data "vcd_ip_space_ip_allocation" "public-ip-prefix" {
  org_id        = data.vcd_org.org1.id
  ip_space_id   = vcd_ip_space.space1.id
  type          = "IP_PREFIX"
  #prefix_length = 30

  ip_address = vcd_ip_space_ip_allocation.public-ip-prefix.ip_address

  depends_on = [vcd_nsxt_edgegateway.ip-space]
}

data "vcd_network_routed_v2" "using-public-prefix" {
  org             = "{{.Org}}"
  name            = "{{.TestName}}"
  edge_gateway_id = vcd_nsxt_edgegateway.ip-space.id
}

data "vcd_ip_space_ip_allocation" "public-ip-prefix-manual" {
  org_id        = data.vcd_org.org1.id
  ip_space_id   = vcd_ip_space.space1.id
  type          = "IP_PREFIX"
  #prefix_length = 30
  #usage_state = "USED_MANUAL"
  #description = "manually used IP Prefix"

  ip_address = vcd_ip_space_ip_allocation.public-ip-prefix-manual.ip_address

  depends_on = [vcd_nsxt_edgegateway.ip-space]
}
`

func sleepTester(d time.Duration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		fmt.Printf("sleeping %s\n", d.String())
		time.Sleep(d)
		fmt.Println("finished sleeping")
		return nil
	}
}

func stateDumper() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		spew.Dump(s)
		return nil
	}
}

/*
func TestAccVcdIpSpaceIntegrationPrivate(t *testing.T) {
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
				Check: resource.ComposeAggregateTestCheckFunc(
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

					resource.TestCheckResourceAttrSet("vcd_ip_space_ip_allocation.public-floating-ip", "id"),
					resource.TestCheckResourceAttrSet("vcd_ip_space_ip_allocation.public-floating-ip", "allocation_date"),
					resource.TestCheckResourceAttr("vcd_ip_space_ip_allocation.public-floating-ip", "type", "FLOATING_IP"),
					// usage_state is UNUSED because the state is updated during creation of this
					// resource and it is consumed in next dependent resource
					resource.TestCheckResourceAttr("vcd_ip_space_ip_allocation.public-floating-ip", "usage_state", "UNUSED"),
					// resource.TestCheckResourceAttr("vcd_ip_space_ip_allocation.public-floating-ip", "quantity", "1"),
					resource.TestCheckResourceAttrSet("vcd_ip_space_ip_allocation.public-floating-ip", "ip_address"),

					resource.TestCheckResourceAttrSet("vcd_ip_space_ip_allocation.public-floating-ip-manual", "id"),
					resource.TestCheckResourceAttr("vcd_ip_space_ip_allocation.public-floating-ip-manual", "usage_state", "USED_MANUAL"),
					resource.TestCheckResourceAttr("vcd_ip_space_ip_allocation.public-floating-ip-manual", "type", "FLOATING_IP"),
					resource.TestCheckResourceAttr("vcd_ip_space_ip_allocation.public-floating-ip-manual", "description", "manually used floating IP"),
					resource.TestCheckResourceAttrSet("vcd_ip_space_ip_allocation.public-floating-ip-manual", "ip_address"),

					// public-ip-prefix
					resource.TestCheckResourceAttrSet("vcd_ip_space_ip_allocation.public-ip-prefix", "id"),
					resource.TestCheckResourceAttr("vcd_ip_space_ip_allocation.public-ip-prefix", "usage_state", "UNUSED"),
					resource.TestCheckResourceAttr("vcd_ip_space_ip_allocation.public-ip-prefix", "type", "IP_PREFIX"),
					resource.TestCheckResourceAttrSet("vcd_ip_space_ip_allocation.public-ip-prefix", "ip_address"),

					resource.TestCheckResourceAttrSet("vcd_ip_space_ip_allocation.public-ip-prefix-manual", "id"),
					resource.TestCheckResourceAttr("vcd_ip_space_ip_allocation.public-ip-prefix-manual", "usage_state", "USED_MANUAL"),
					resource.TestCheckResourceAttr("vcd_ip_space_ip_allocation.public-ip-prefix-manual", "type", "IP_PREFIX"),
					resource.TestCheckResourceAttr("vcd_ip_space_ip_allocation.public-ip-prefix-manual", "description", "manually used IP Prefix"),
					resource.TestCheckResourceAttrSet("vcd_ip_space_ip_allocation.public-ip-prefix-manual", "ip_address"),
				),
			},
			{
				Config: configText3DS,
				Check: resource.ComposeAggregateTestCheckFunc(
					resourceFieldsEqual("data.vcd_ip_space_uplink.u1", "vcd_ip_space_uplink.u1", nil),
					resourceFieldsEqual("data.vcd_nsxt_edgegateway.ip-space", "vcd_nsxt_edgegateway.ip-space", []string{"%"}),

					resourceFieldsEqual("data.vcd_ip_space_ip_allocation.public-floating-ip", "vcd_ip_space_ip_allocation.public-floating-ip", nil),
					resourceFieldsEqual("data.vcd_ip_space_ip_allocation.public-floating-ip-manual", "vcd_ip_space_ip_allocation.public-floating-ip-manual", nil),
					resourceFieldsEqual("data.vcd_ip_space_ip_allocation.public-ip-prefix", "vcd_ip_space_ip_allocation.public-ip-prefix", nil),
					resourceFieldsEqual("data.vcd_ip_space_ip_allocation.public-ip-prefix-manual", "vcd_ip_space_ip_allocation.public-ip-prefix-manual", nil),
				),
			},
		},
	})
	postTestChecks(t)
}
*/
