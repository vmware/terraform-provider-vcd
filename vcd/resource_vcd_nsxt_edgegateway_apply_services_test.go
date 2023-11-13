//go:build network || nsxt || ALL || functional

package vcd

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func sleepTester(d time.Duration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		fmt.Printf("sleeping %s\n", d.String())
		time.Sleep(d)
		fmt.Println("finished sleeping")
		return nil
	}
}
func TestAccVcdIpSpacePublicEdgeGatewayDefaultServices(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	if checkVersion(testConfig.Provider.ApiVersion, "< 38.0") {
		t.Skipf("This test tests VCD 10.5.0+ (API V38.0+) features. Skipping.")
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
	configText1 := templateFill(testAccVcdIpSpacePublicEdgeGatewayDefaultServices, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)

	// params["FuncName"] = t.Name() + "step2"
	// configText2 := templateFill(testAccVcdIpSpacePublicStep2, params)
	// debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText2)

	// params["FuncName"] = t.Name() + "step3"
	// configText3 := templateFill(testAccVcdIpSpacePublicStep3, params)
	// debugPrintf("#[DEBUG] CONFIGURATION for step 3: %s", configText3)

	// params["FuncName"] = t.Name() + "step4"
	// configText4 := templateFill(testAccVcdIpSpacePublicStep4, params)
	// debugPrintf("#[DEBUG] CONFIGURATION for step 4: %s", configText4)

	// params["FuncName"] = t.Name() + "step5"
	// configText5 := templateFill(testAccVcdIpSpacePublicStep5, params)
	// debugPrintf("#[DEBUG] CONFIGURATION for step 5: %s", configText5)

	// params["FuncName"] = t.Name() + "step6"
	// configText6DS := templateFill(testAccVcdIpSpacePublicStep5DS, params)
	// debugPrintf("#[DEBUG] CONFIGURATION for step 6: %s", configText6DS)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVcdNsxtIpSpacesDestroy(params["TestName"].(string)),
		Steps: []resource.TestStep{
			{
				Config: configText1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_ip_space.space1", "id"),
					// sleepTester(5*time.Minute),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdIpSpacePublicEdgeGatewayDefaultServicesPrereqs = `
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
  external_scope = "55.0.0.0/24"

  route_advertisement_enabled            = false
  default_firewall_rule_creation_enabled = true
  default_no_snat_rule_creation_enabled  = true
  default_snat_rule_creation_enabled     = true

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

const testAccVcdIpSpacePublicEdgeGatewayDefaultServices = testAccVcdIpSpacePublicEdgeGatewayDefaultServicesPrereqs + `
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

resource "vcd_nsxt_edgegateway_apply_ip_space_services" "edge1" {
  org             = "{{.Org}}"
  edge_gateway_id = vcd_nsxt_edgegateway.ip-space.id
  
  trigger_on_create = true

  depends_on = [vcd_nsxt_edgegateway.ip-space]
}
`
