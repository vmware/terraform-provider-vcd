//go:build nsxt || alb || ALL || functional

package vcd

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcdNsxtNetworkSegmentProfileCustom(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	skipNoNsxtAlbConfiguration(t)

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

	// params["FuncName"] = t.Name() + "step3"
	// configText3 := templateFill(testAccVcdNsxtNetworkSegmentProfileGlobalDefault, params)
	// debugPrintf("#[DEBUG] CONFIGURATION for step 3: %s", configText3)

	// params["FuncName"] = t.Name() + "step4"
	// configText4DS := templateFill(testAccVcdNsxtNetworkSegmentProfileGlobalDefaultDS, params)
	// debugPrintf("#[DEBUG] CONFIGURATION for step 4: %s", configText4DS)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		// CheckDestroy:      testAccCheckVcdSegmentProfileTemplateDestroy("vcd_nsxt_alb_controller.first"),
		Steps: []resource.TestStep{
			{
				Config: configText1,
				Check:  resource.ComposeAggregateTestCheckFunc(
				// resource.TestMatchResourceAttr("vcd_nsxt_segment_profile_template.first", "id", regexp.MustCompile(`\d*`)),
				// resource.TestCheckResourceAttr("vcd_nsxt_segment_profile_template.empty", "name", t.Name()+"-empty"),
				// resource.TestCheckResourceAttr("vcd_nsxt_segment_profile_template.empty", "description", "description"),
				// resource.TestCheckResourceAttr("vcd_nsxt_segment_profile_template.empty", "ip_discovery_profile_id", ""),
				// resource.TestCheckResourceAttr("vcd_nsxt_segment_profile_template.empty", "mac_discovery_profile_id", ""),
				// resource.TestCheckResourceAttr("vcd_nsxt_segment_profile_template.empty", "spoof_guard_profile_id", ""),
				// resource.TestCheckResourceAttr("vcd_nsxt_segment_profile_template.empty", "qos_profile_id", ""),
				// resource.TestCheckResourceAttr("vcd_nsxt_segment_profile_template.empty", "segment_security_profile_id", ""),

				// resource.TestCheckResourceAttr("vcd_nsxt_segment_profile_template.complete", "name", t.Name()+"-complete"),
				// resource.TestMatchResourceAttr("vcd_nsxt_segment_profile_template.complete", "ip_discovery_profile_id", regexp.MustCompile(`[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$`)),
				// resource.TestMatchResourceAttr("vcd_nsxt_segment_profile_template.complete", "mac_discovery_profile_id", regexp.MustCompile(`[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$`)),
				// resource.TestMatchResourceAttr("vcd_nsxt_segment_profile_template.complete", "spoof_guard_profile_id", regexp.MustCompile(`[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$`)),
				// resource.TestMatchResourceAttr("vcd_nsxt_segment_profile_template.complete", "qos_profile_id", regexp.MustCompile(`[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$`)),
				// resource.TestMatchResourceAttr("vcd_nsxt_segment_profile_template.complete", "segment_security_profile_id", regexp.MustCompile(`[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$`)),

				// resource.TestCheckResourceAttr("vcd_nsxt_segment_profile_template.half-complete", "name", t.Name()+"-half-complete"),
				// resource.TestCheckResourceAttr("vcd_nsxt_segment_profile_template.half-complete", "description", ""),
				// resource.TestMatchResourceAttr("vcd_nsxt_segment_profile_template.half-complete", "ip_discovery_profile_id", regexp.MustCompile(`[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$`)),
				// resource.TestMatchResourceAttr("vcd_nsxt_segment_profile_template.half-complete", "mac_discovery_profile_id", regexp.MustCompile(`[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$`)),
				// resource.TestMatchResourceAttr("vcd_nsxt_segment_profile_template.half-complete", "spoof_guard_profile_id", regexp.MustCompile(`[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$`)),
				// resource.TestCheckResourceAttr("vcd_nsxt_segment_profile_template.half-complete", "qos_profile_id", ""),
				// resource.TestCheckResourceAttr("vcd_nsxt_segment_profile_template.half-complete", "segment_security_profile_id", ""),
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
			// {
			// 	Config: configText2DS,
			// 	Check: resource.ComposeAggregateTestCheckFunc(
			// 		resourceFieldsEqual("data.vcd_nsxt_segment_profile_template.empty", "vcd_nsxt_segment_profile_template.empty", nil),
			// 		resourceFieldsEqual("data.vcd_nsxt_segment_profile_template.half-complete", "vcd_nsxt_segment_profile_template.half-complete", nil),
			// 		resourceFieldsEqual("data.vcd_nsxt_segment_profile_template.complete", "vcd_nsxt_segment_profile_template.complete", nil),
			// 	),
			// },
			// {
			// 	Config: configText3,
			// 	Check: resource.ComposeAggregateTestCheckFunc(
			// 		resource.TestCheckResourceAttr("vcd_nsxt_global_default_segment_profile_template.singleton", "id", "no-real-id"),
			// 	),
			// },
			// {
			// 	Config: configText4DS,
			// 	Check: resource.ComposeAggregateTestCheckFunc(
			// 		resourceFieldsEqual("data.vcd_nsxt_global_default_segment_profile_template.singleton", "vcd_nsxt_global_default_segment_profile_template.singleton", nil),
			// 	),
			// },
		},
	})
	postTestChecks(t)
}

const testAccVcdNsxtNetworkSegmentProfileCustom = `
data "vcd_nsxt_manager" "nsxt" {
  name = "{{.NsxtManager}}"
}

data "vcd_nsxt_segment_ip_discovery_profile" "first" {
  name       = "{{.IpDiscoveryProfileName}}"
  context_id = data.vcd_nsxt_manager.nsxt.id
}

data "vcd_nsxt_segment_mac_discovery_profile" "first" {
  name       = "{{.MacDiscoveryProfileName}}"
  context_id = data.vcd_nsxt_manager.nsxt.id
}

data "vcd_nsxt_segment_spoof_guard_profile" "first" {
  name       = "{{.SpoofGuardProfileName}}"
  context_id = data.vcd_nsxt_manager.nsxt.id
}

data "vcd_nsxt_segment_qos_profile" "first" {
  name       = "{{.QosProfileName}}"
  context_id = data.vcd_nsxt_manager.nsxt.id
}

data "vcd_nsxt_segment_security_profile" "first" {
  name       = "{{.SegmentSecurityProfileName}}"
  context_id = data.vcd_nsxt_manager.nsxt.id
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

func TestAccVcdNsxtNetworkSegmentProfileTemplate(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	skipNoNsxtAlbConfiguration(t)

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

	configText1 := templateFill(testAccVcdNsxtNetworkSegmentProfileTemplate, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)

	params["FuncName"] = t.Name() + "step2"
	configText2DS := templateFill(testAccVcdNsxtNetworkSegmentProfileTemplateDS, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText2DS)

	// params["FuncName"] = t.Name() + "step3"
	// configText3 := templateFill(testAccVcdNsxtNetworkSegmentProfileGlobalDefault, params)
	// debugPrintf("#[DEBUG] CONFIGURATION for step 3: %s", configText3)

	// params["FuncName"] = t.Name() + "step4"
	// configText4DS := templateFill(testAccVcdNsxtNetworkSegmentProfileGlobalDefaultDS, params)
	// debugPrintf("#[DEBUG] CONFIGURATION for step 4: %s", configText4DS)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		// CheckDestroy:      testAccCheckVcdSegmentProfileTemplateDestroy("vcd_nsxt_alb_controller.first"),
		Steps: []resource.TestStep{
			{
				Config: configText1,
				Check:  resource.ComposeAggregateTestCheckFunc(
				// resource.TestMatchResourceAttr("vcd_nsxt_segment_profile_template.first", "id", regexp.MustCompile(`\d*`)),
				// resource.TestCheckResourceAttr("vcd_nsxt_segment_profile_template.empty", "name", t.Name()+"-empty"),
				// resource.TestCheckResourceAttr("vcd_nsxt_segment_profile_template.empty", "description", "description"),
				// resource.TestCheckResourceAttr("vcd_nsxt_segment_profile_template.empty", "ip_discovery_profile_id", ""),
				// resource.TestCheckResourceAttr("vcd_nsxt_segment_profile_template.empty", "mac_discovery_profile_id", ""),
				// resource.TestCheckResourceAttr("vcd_nsxt_segment_profile_template.empty", "spoof_guard_profile_id", ""),
				// resource.TestCheckResourceAttr("vcd_nsxt_segment_profile_template.empty", "qos_profile_id", ""),
				// resource.TestCheckResourceAttr("vcd_nsxt_segment_profile_template.empty", "segment_security_profile_id", ""),

				// resource.TestCheckResourceAttr("vcd_nsxt_segment_profile_template.complete", "name", t.Name()+"-complete"),
				// resource.TestMatchResourceAttr("vcd_nsxt_segment_profile_template.complete", "ip_discovery_profile_id", regexp.MustCompile(`[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$`)),
				// resource.TestMatchResourceAttr("vcd_nsxt_segment_profile_template.complete", "mac_discovery_profile_id", regexp.MustCompile(`[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$`)),
				// resource.TestMatchResourceAttr("vcd_nsxt_segment_profile_template.complete", "spoof_guard_profile_id", regexp.MustCompile(`[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$`)),
				// resource.TestMatchResourceAttr("vcd_nsxt_segment_profile_template.complete", "qos_profile_id", regexp.MustCompile(`[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$`)),
				// resource.TestMatchResourceAttr("vcd_nsxt_segment_profile_template.complete", "segment_security_profile_id", regexp.MustCompile(`[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$`)),

				// resource.TestCheckResourceAttr("vcd_nsxt_segment_profile_template.half-complete", "name", t.Name()+"-half-complete"),
				// resource.TestCheckResourceAttr("vcd_nsxt_segment_profile_template.half-complete", "description", ""),
				// resource.TestMatchResourceAttr("vcd_nsxt_segment_profile_template.half-complete", "ip_discovery_profile_id", regexp.MustCompile(`[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$`)),
				// resource.TestMatchResourceAttr("vcd_nsxt_segment_profile_template.half-complete", "mac_discovery_profile_id", regexp.MustCompile(`[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$`)),
				// resource.TestMatchResourceAttr("vcd_nsxt_segment_profile_template.half-complete", "spoof_guard_profile_id", regexp.MustCompile(`[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$`)),
				// resource.TestCheckResourceAttr("vcd_nsxt_segment_profile_template.half-complete", "qos_profile_id", ""),
				// resource.TestCheckResourceAttr("vcd_nsxt_segment_profile_template.half-complete", "segment_security_profile_id", ""),
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
			// {
			// 	Config: configText3,
			// 	Check: resource.ComposeAggregateTestCheckFunc(
			// 		resource.TestCheckResourceAttr("vcd_nsxt_global_default_segment_profile_template.singleton", "id", "no-real-id"),
			// 	),
			// },
			// {
			// 	Config: configText4DS,
			// 	Check: resource.ComposeAggregateTestCheckFunc(
			// 		resourceFieldsEqual("data.vcd_nsxt_global_default_segment_profile_template.singleton", "vcd_nsxt_global_default_segment_profile_template.singleton", nil),
			// 	),
			// },
		},
	})
	postTestChecks(t)
}

const testAccVcdNsxtNetworkSegmentProfileTemplate = `
data "vcd_nsxt_manager" "nsxt" {
  name = "{{.NsxtManager}}"
}

data "vcd_nsxt_segment_ip_discovery_profile" "first" {
  name       = "{{.IpDiscoveryProfileName}}"
  context_id = data.vcd_nsxt_manager.nsxt.id
}

data "vcd_nsxt_segment_mac_discovery_profile" "first" {
  name       = "{{.MacDiscoveryProfileName}}"
  context_id = data.vcd_nsxt_manager.nsxt.id
}

data "vcd_nsxt_segment_spoof_guard_profile" "first" {
  name       = "{{.SpoofGuardProfileName}}"
  context_id = data.vcd_nsxt_manager.nsxt.id
}

data "vcd_nsxt_segment_qos_profile" "first" {
  name       = "{{.QosProfileName}}"
  context_id = data.vcd_nsxt_manager.nsxt.id
}

data "vcd_nsxt_segment_security_profile" "first" {
  name       = "{{.SegmentSecurityProfileName}}"
  context_id = data.vcd_nsxt_manager.nsxt.id
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
    end_address = "1.1.1.20"
  }
}

resource "vcd_nsxt_network_segment_profile" "custom-prof" {
  org            = "{{.Org}}"
  org_network_id = vcd_network_routed_v2.net1.id

  segment_profile_template_id = vcd_nsxt_segment_profile_template.complete.id
}
`

const testAccVcdNsxtNetworkSegmentProfileTemplateDS = testAccVcdNsxtNetworkSegmentProfileTemplate + `
data "vcd_nsxt_network_segment_profile" "custom-prof" {
  org            = "{{.Org}}"
  org_network_id = vcd_network_routed_v2.net1.id

  depends_on = [vcd_nsxt_network_segment_profile.custom-prof]
}
`
