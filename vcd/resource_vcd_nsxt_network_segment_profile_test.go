//go:build nsxt || ALL || functional

package vcd

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcdNsxtNetworkSegmentProfileCustom(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	// String map to fill the template
	var params = StringMap{
		"TestName":                   t.Name(),
		"Org":                        testConfig.VCD.Org,
		"NsxtVdc":                    testConfig.Nsxt.Vdc,
		"EdgeGw":                     testConfig.Nsxt.EdgeGateway,
		"NsxtManager":                testConfig.Nsxt.Manager,
		"IpDiscoveryProfileName":     testConfig.Nsxt.IpDiscoveryProfile,
		"MacDiscoveryProfileName":    testConfig.Nsxt.MacDiscoveryProfile,
		"QosProfileName":             testConfig.Nsxt.QosProfile,
		"SpoofGuardProfileName":      testConfig.Nsxt.SpoofGuardProfile,
		"SegmentSecurityProfileName": testConfig.Nsxt.SegmentSecurityProfile,

		"Tags": "nsxt ",
	}

	configText1 := templateFill(testAccVcdNsxtNetworkSegmentProfileCustom, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)

	params["FuncName"] = t.Name() + "step2"
	configText2DS := templateFill(testAccVcdNsxtNetworkSegmentProfileCustomDS, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText2DS)

	params["FuncName"] = t.Name() + "step4"
	configText4 := templateFill(testAccVcdNsxtNetworkSegmentProfileCustomUpdate, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 4: %s", configText4)

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
					resource.TestCheckResourceAttrSet("vcd_nsxt_network_segment_profile.custom-prof", "id"),
					resource.TestCheckResourceAttrPair("data.vcd_nsxt_segment_ip_discovery_profile.first", "id", "vcd_nsxt_network_segment_profile.custom-prof", "ip_discovery_profile_id"),
					resource.TestCheckResourceAttrPair("data.vcd_nsxt_segment_mac_discovery_profile.first", "id", "vcd_nsxt_network_segment_profile.custom-prof", "mac_discovery_profile_id"),
					resource.TestCheckResourceAttrPair("data.vcd_nsxt_segment_spoof_guard_profile.first", "id", "vcd_nsxt_network_segment_profile.custom-prof", "spoof_guard_profile_id"),
					resource.TestCheckResourceAttrPair("data.vcd_nsxt_segment_qos_profile.first", "id", "vcd_nsxt_network_segment_profile.custom-prof", "qos_profile_id"),
					resource.TestCheckResourceAttrPair("data.vcd_nsxt_segment_security_profile.first", "id", "vcd_nsxt_network_segment_profile.custom-prof", "segment_security_profile_id"),
				),
			},
			{
				ResourceName:      "vcd_nsxt_network_segment_profile.custom-prof",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdOrgNsxtVdcObject(t.Name() + "-routed"),
			},
			{
				Config: configText2DS,
				Check: resource.ComposeAggregateTestCheckFunc(
					resourceFieldsEqual("data.vcd_nsxt_network_segment_profile.custom-prof", "vcd_nsxt_network_segment_profile.custom-prof", nil),
				),
			},
			{
				Config: configText4,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_nsxt_network_segment_profile.custom-prof", "id"),
					resource.TestCheckResourceAttrPair("data.vcd_nsxt_segment_ip_discovery_profile.first", "id", "vcd_nsxt_network_segment_profile.custom-prof", "ip_discovery_profile_id"),
					resource.TestCheckResourceAttrPair("data.vcd_nsxt_segment_mac_discovery_profile.first", "id", "vcd_nsxt_network_segment_profile.custom-prof", "mac_discovery_profile_id"),
					resource.TestCheckResourceAttrPair("data.vcd_nsxt_segment_spoof_guard_profile.first", "id", "vcd_nsxt_network_segment_profile.custom-prof", "spoof_guard_profile_id"),
					resource.TestCheckResourceAttrPair("data.vcd_nsxt_segment_qos_profile.first", "id", "vcd_nsxt_network_segment_profile.custom-prof", "qos_profile_id"),
					resource.TestCheckResourceAttrPair("data.vcd_nsxt_segment_security_profile.first", "id", "vcd_nsxt_network_segment_profile.custom-prof", "segment_security_profile_id"),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdNsxtNetworkSegmentProfileCustom = `
data "vcd_nsxt_manager" "nsxt" {
  name = "{{.NsxtManager}}"
}

data "vcd_nsxt_segment_ip_discovery_profile" "first" {
  name            = "{{.IpDiscoveryProfileName}}"
  nsxt_manager_id = data.vcd_nsxt_manager.nsxt.id
}

data "vcd_nsxt_segment_mac_discovery_profile" "first" {
  name            = "{{.MacDiscoveryProfileName}}"
  nsxt_manager_id = data.vcd_nsxt_manager.nsxt.id
}

data "vcd_nsxt_segment_spoof_guard_profile" "first" {
  name            = "{{.SpoofGuardProfileName}}"
  nsxt_manager_id = data.vcd_nsxt_manager.nsxt.id
}

data "vcd_nsxt_segment_qos_profile" "first" {
  name            = "{{.QosProfileName}}"
  nsxt_manager_id = data.vcd_nsxt_manager.nsxt.id
}

data "vcd_nsxt_segment_security_profile" "first" {
  name            = "{{.SegmentSecurityProfileName}}"
  nsxt_manager_id = data.vcd_nsxt_manager.nsxt.id
}

data "vcd_nsxt_edgegateway" "existing" {
  org  = "{{.Org}}"
  name = "{{.EdgeGw}}"
}

resource "vcd_network_routed_v2" "net1" {
  org  = "{{.Org}}"
  name = "{{.TestName}}-routed"

  edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id

  gateway = "1.1.1.1"
  prefix_length = 24

  static_ip_pool {
	start_address = "1.1.1.10"
    end_address = "1.1.1.20"
  }
}

resource "vcd_nsxt_network_segment_profile" "custom-prof" {
  org            = "{{.Org}}"
  org_network_id = vcd_network_routed_v2.net1.id

  ip_discovery_profile_id     = data.vcd_nsxt_segment_ip_discovery_profile.first.id
  mac_discovery_profile_id    = data.vcd_nsxt_segment_mac_discovery_profile.first.id
  spoof_guard_profile_id      = data.vcd_nsxt_segment_spoof_guard_profile.first.id
  qos_profile_id              = data.vcd_nsxt_segment_qos_profile.first.id
  segment_security_profile_id = data.vcd_nsxt_segment_security_profile.first.id
}
`

const testAccVcdNsxtNetworkSegmentProfileCustomDS = testAccVcdNsxtNetworkSegmentProfileCustom + `
data "vcd_nsxt_network_segment_profile" "custom-prof" {
  org            = "{{.Org}}"
  org_network_id = vcd_network_routed_v2.net1.id

  depends_on = [vcd_nsxt_network_segment_profile.custom-prof]
}
`

const testAccVcdNsxtNetworkSegmentProfileCustomUpdate = `
data "vcd_nsxt_manager" "nsxt" {
  name = "{{.NsxtManager}}"
}

data "vcd_nsxt_segment_ip_discovery_profile" "first" {
  name            = "{{.IpDiscoveryProfileName}}"
  nsxt_manager_id = data.vcd_nsxt_manager.nsxt.id
}

data "vcd_nsxt_segment_mac_discovery_profile" "first" {
  name            = "{{.MacDiscoveryProfileName}}"
  nsxt_manager_id = data.vcd_nsxt_manager.nsxt.id
}

data "vcd_nsxt_segment_spoof_guard_profile" "first" {
  name            = "{{.SpoofGuardProfileName}}"
  nsxt_manager_id = data.vcd_nsxt_manager.nsxt.id
}

data "vcd_nsxt_segment_qos_profile" "first" {
  name            = "{{.QosProfileName}}"
  nsxt_manager_id = data.vcd_nsxt_manager.nsxt.id
}

data "vcd_nsxt_segment_security_profile" "first" {
  name            = "{{.SegmentSecurityProfileName}}"
  nsxt_manager_id = data.vcd_nsxt_manager.nsxt.id
}

data "vcd_nsxt_edgegateway" "existing" {
  org  = "{{.Org}}"
  name = "{{.EdgeGw}}"
}

resource "vcd_network_routed_v2" "net1" {
  org         = "{{.Org}}"
  name        = "{{.TestName}}-routed"
  description = "{{.TestName}}-description"

  edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id

  gateway = "1.1.1.1"
  prefix_length = 24

  static_ip_pool {
	start_address = "1.1.1.10"
    end_address = "1.1.1.20"
  }
}

resource "vcd_nsxt_network_segment_profile" "custom-prof" {
  org            = "{{.Org}}"
  org_network_id = vcd_network_routed_v2.net1.id

  ip_discovery_profile_id     = data.vcd_nsxt_segment_ip_discovery_profile.first.id
  mac_discovery_profile_id    = data.vcd_nsxt_segment_mac_discovery_profile.first.id
  spoof_guard_profile_id      = data.vcd_nsxt_segment_spoof_guard_profile.first.id
  qos_profile_id              = data.vcd_nsxt_segment_qos_profile.first.id
  segment_security_profile_id = data.vcd_nsxt_segment_security_profile.first.id
}
`

func TestAccVcdNsxtNetworkSegmentProfileTemplate(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	// String map to fill the template
	var params = StringMap{
		"TestName":                   t.Name(),
		"Org":                        testConfig.VCD.Org,
		"NsxtVdc":                    testConfig.Nsxt.Vdc,
		"EdgeGw":                     testConfig.Nsxt.EdgeGateway,
		"NsxtImportSegment":          testConfig.Nsxt.NsxtImportSegment,
		"NsxtManager":                testConfig.Nsxt.Manager,
		"IpDiscoveryProfileName":     testConfig.Nsxt.IpDiscoveryProfile,
		"MacDiscoveryProfileName":    testConfig.Nsxt.MacDiscoveryProfile,
		"QosProfileName":             testConfig.Nsxt.QosProfile,
		"SpoofGuardProfileName":      testConfig.Nsxt.SpoofGuardProfile,
		"SegmentSecurityProfileName": testConfig.Nsxt.SegmentSecurityProfile,

		"Tags": "nsxt ",
	}

	configText1 := templateFill(testAccVcdNsxtNetworkSegmentProfileTemplateStep1, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)

	params["FuncName"] = t.Name() + "step2"
	configText2DS := templateFill(testAccVcdNsxtNetworkSegmentProfileTemplateDS, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText2DS)

	params["FuncName"] = t.Name() + "step4"
	configText4 := templateFill(testAccVcdNsxtNetworkSegmentProfileTemplateStep2, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 4: %s", configText4)

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
					resource.TestCheckResourceAttrSet("vcd_nsxt_segment_profile_template.complete", "id"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_network_segment_profile.custom-prof-routed", "id"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_network_segment_profile.custom-prof-isolated", "id"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_network_segment_profile.custom-prof-imported", "id"),
				),
			},
			{
				ResourceName:      "vcd_nsxt_network_segment_profile.custom-prof-routed",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdOrgNsxtVdcObject(t.Name() + "-routed"),
			},
			{
				Config: configText2DS,
				Check: resource.ComposeAggregateTestCheckFunc(
					resourceFieldsEqual("data.vcd_nsxt_network_segment_profile.custom-prof-routed", "vcd_nsxt_network_segment_profile.custom-prof-routed", nil),
				),
			},
			{
				// This step checks that updating Org VDC network does not compromise its Segment Profile configuration
				// after updating Org VDC networks
				Config: configText4,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_nsxt_segment_profile_template.complete", "id"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_network_segment_profile.custom-prof-routed", "id"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_network_segment_profile.custom-prof-isolated", "id"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_network_segment_profile.custom-prof-imported", "id"),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdNsxtNetworkSegmentProfileTemplateStep1 = `
data "vcd_org_vdc" "nsxt" {
  org  = "{{.Org}}"
  name = "{{.NsxtVdc}}"
}

data "vcd_nsxt_manager" "nsxt" {
  name = "{{.NsxtManager}}"
}

data "vcd_nsxt_segment_ip_discovery_profile" "first" {
  name            = "{{.IpDiscoveryProfileName}}"
  nsxt_manager_id = data.vcd_nsxt_manager.nsxt.id
}

data "vcd_nsxt_segment_mac_discovery_profile" "first" {
  name            = "{{.MacDiscoveryProfileName}}"
  nsxt_manager_id = data.vcd_nsxt_manager.nsxt.id
}

data "vcd_nsxt_segment_spoof_guard_profile" "first" {
  name            = "{{.SpoofGuardProfileName}}"
  nsxt_manager_id = data.vcd_nsxt_manager.nsxt.id
}

data "vcd_nsxt_segment_qos_profile" "first" {
  name            = "{{.QosProfileName}}"
  nsxt_manager_id = data.vcd_nsxt_manager.nsxt.id
}

data "vcd_nsxt_segment_security_profile" "first" {
  name            = "{{.SegmentSecurityProfileName}}"
  nsxt_manager_id = data.vcd_nsxt_manager.nsxt.id
}

resource "vcd_nsxt_segment_profile_template" "complete" {
  name        = "{{.TestName}}-complete"
  description = "description"

  nsxt_manager_id             = data.vcd_nsxt_manager.nsxt.id
  ip_discovery_profile_id     = data.vcd_nsxt_segment_ip_discovery_profile.first.id
  mac_discovery_profile_id    = data.vcd_nsxt_segment_mac_discovery_profile.first.id
  spoof_guard_profile_id      = data.vcd_nsxt_segment_spoof_guard_profile.first.id
  qos_profile_id              = data.vcd_nsxt_segment_qos_profile.first.id
  segment_security_profile_id = data.vcd_nsxt_segment_security_profile.first.id
}

data "vcd_nsxt_edgegateway" "existing" {
  org  = "{{.Org}}"
  name = "{{.EdgeGw}}"
}

resource "vcd_network_routed_v2" "net1" {
  org  = "{{.Org}}"
  name = "{{.TestName}}-routed"

  edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id

  gateway = "1.1.1.1"
  prefix_length = 24

  static_ip_pool {
	start_address = "1.1.1.10"
    end_address   = "1.1.1.20"
  }
}

resource "vcd_nsxt_network_segment_profile" "custom-prof-routed" {
  org            = "{{.Org}}"
  org_network_id = vcd_network_routed_v2.net1.id

  segment_profile_template_id = vcd_nsxt_segment_profile_template.complete.id
}

resource "vcd_network_isolated_v2" "nsxt-backed" {
  org      = "{{.Org}}"
  owner_id = data.vcd_org_vdc.nsxt.id

  name = "{{.TestName}}-isolated"

  gateway       = "1.1.1.1"
  prefix_length = 24

  static_ip_pool {
    start_address = "1.1.1.10"
    end_address   = "1.1.1.20"
  }

  static_ip_pool {
    start_address = "1.1.1.100"
    end_address   = "1.1.1.103"
  }
}

resource "vcd_nsxt_network_segment_profile" "custom-prof-isolated" {
  org            = "{{.Org}}"
  org_network_id = vcd_network_isolated_v2.nsxt-backed.id

  segment_profile_template_id = vcd_nsxt_segment_profile_template.complete.id
}

resource "vcd_nsxt_network_imported" "net1" {
  org      = "{{.Org}}"
  owner_id = data.vcd_org_vdc.nsxt.id
  name     = "{{.TestName}}-imported"

  nsxt_logical_switch_name = "{{.NsxtImportSegment}}"

  gateway       = "8.1.1.1"
  prefix_length = 24

  static_ip_pool {
	start_address = "8.1.1.10"
	end_address   = "8.1.1.20"
  }
}

resource "vcd_nsxt_network_segment_profile" "custom-prof-imported" {
  org            = "{{.Org}}"
  org_network_id = vcd_nsxt_network_imported.net1.id

  segment_profile_template_id = vcd_nsxt_segment_profile_template.complete.id
}
`

const testAccVcdNsxtNetworkSegmentProfileTemplateDS = testAccVcdNsxtNetworkSegmentProfileTemplateStep1 + `
data "vcd_nsxt_network_segment_profile" "custom-prof-routed" {
  org            = "{{.Org}}"
  org_network_id = vcd_network_routed_v2.net1.id

  depends_on = [vcd_nsxt_network_segment_profile.custom-prof-routed]
}
`

const testAccVcdNsxtNetworkSegmentProfileTemplateStep2 = `
data "vcd_org_vdc" "nsxt" {
  org  = "{{.Org}}"
  name = "{{.NsxtVdc}}"
}

data "vcd_nsxt_manager" "nsxt" {
  name = "{{.NsxtManager}}"
}

data "vcd_nsxt_segment_ip_discovery_profile" "first" {
  name            = "{{.IpDiscoveryProfileName}}"
  nsxt_manager_id = data.vcd_nsxt_manager.nsxt.id
}

data "vcd_nsxt_segment_mac_discovery_profile" "first" {
  name            = "{{.MacDiscoveryProfileName}}"
  nsxt_manager_id = data.vcd_nsxt_manager.nsxt.id
}

data "vcd_nsxt_segment_spoof_guard_profile" "first" {
  name            = "{{.SpoofGuardProfileName}}"
  nsxt_manager_id = data.vcd_nsxt_manager.nsxt.id
}

data "vcd_nsxt_segment_qos_profile" "first" {
  name            = "{{.QosProfileName}}"
  nsxt_manager_id = data.vcd_nsxt_manager.nsxt.id
}

data "vcd_nsxt_segment_security_profile" "first" {
  name            = "{{.SegmentSecurityProfileName}}"
  nsxt_manager_id = data.vcd_nsxt_manager.nsxt.id
}

resource "vcd_nsxt_segment_profile_template" "complete" {
  name        = "{{.TestName}}-complete"
  description = "description"

  nsxt_manager_id             = data.vcd_nsxt_manager.nsxt.id
  ip_discovery_profile_id     = data.vcd_nsxt_segment_ip_discovery_profile.first.id
  mac_discovery_profile_id    = data.vcd_nsxt_segment_mac_discovery_profile.first.id
  spoof_guard_profile_id      = data.vcd_nsxt_segment_spoof_guard_profile.first.id
  qos_profile_id              = data.vcd_nsxt_segment_qos_profile.first.id
  segment_security_profile_id = data.vcd_nsxt_segment_security_profile.first.id
}

data "vcd_nsxt_edgegateway" "existing" {
  org  = "{{.Org}}"
  name = "{{.EdgeGw}}"
}

resource "vcd_network_routed_v2" "net1" {
  org         = "{{.Org}}"
  name        = "{{.TestName}}-routed"
  description = "{{.TestName}}-added-description"

  edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id

  gateway = "1.1.1.1"
  prefix_length = 24

  static_ip_pool {
	start_address = "1.1.1.10"
    end_address   = "1.1.1.20"
  }

  static_ip_pool {
	start_address = "1.1.1.40"
    end_address   = "1.1.1.50"
  }
}

resource "vcd_nsxt_network_segment_profile" "custom-prof-routed" {
  org            = "{{.Org}}"
  org_network_id = vcd_network_routed_v2.net1.id

  segment_profile_template_id = vcd_nsxt_segment_profile_template.complete.id
}

resource "vcd_network_isolated_v2" "nsxt-backed" {
  org      = "{{.Org}}"
  owner_id = data.vcd_org_vdc.nsxt.id

  name        = "{{.TestName}}-isolated"
  description = "My isolated Org VDC network backed by NSX-T"

  gateway       = "1.1.1.1"
  prefix_length = 24

  static_ip_pool {
    start_address = "1.1.1.10"
    end_address   = "1.1.1.20"
  }

}

resource "vcd_nsxt_network_segment_profile" "custom-prof-isolated" {
  org            = "{{.Org}}"
  org_network_id = vcd_network_isolated_v2.nsxt-backed.id

  segment_profile_template_id = vcd_nsxt_segment_profile_template.complete.id
}

resource "vcd_nsxt_network_imported" "net1" {
  org         = "{{.Org}}"
  owner_id    = data.vcd_org_vdc.nsxt.id
  name        = "{{.TestName}}-imported"
  description = "{{.TestName}}-imported"

  nsxt_logical_switch_name = "{{.NsxtImportSegment}}"

  gateway       = "8.1.1.1"
  prefix_length = 24

  static_ip_pool {
	start_address = "8.1.1.10"
	end_address   = "8.1.1.20"
  }
}

resource "vcd_nsxt_network_segment_profile" "custom-prof-imported" {
  org            = "{{.Org}}"
  org_network_id = vcd_nsxt_network_imported.net1.id

  segment_profile_template_id = vcd_nsxt_segment_profile_template.complete.id
}
`
