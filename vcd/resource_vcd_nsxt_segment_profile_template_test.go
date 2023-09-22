//go:build nsxt || alb || ALL || functional

package vcd

import (
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccVcdNsxtSegmentProfileTemplate(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	skipNoNsxtAlbConfiguration(t)

	// String map to fill the template
	var params = StringMap{
		"TestName":                   t.Name(),
		"NsxtManager":                testConfig.Nsxt.Manager,
		"IpDiscoveryProfileName":     testConfig.Nsxt.IpDiscoveryProfile,
		"MacDiscoveryProfileName":    testConfig.Nsxt.MacDiscoveryProfile,
		"QosProfileName":             testConfig.Nsxt.QosProfile,
		"SpoofGuardProfileName":      testConfig.Nsxt.SpoofGuardProfile,
		"SegmentSecurityProfileName": testConfig.Nsxt.SegmentSecurityProfile,
		// "ControllerUrl":      testConfig.Nsxt.NsxtSegmentProfileTemplateUrl,
		// "ControllerUsername": testConfig.Nsxt.NsxtSegmentProfileTemplateUser,
		// "ControllerPassword": testConfig.Nsxt.NsxtSegmentProfileTemplatePassword,
		"Tags": "nsxt ",
	}

	configText1 := templateFill(testAccVcdNsxtSegmentProfileTemplate, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)

	params["FuncName"] = t.Name() + "step2"
	configText2DS := templateFill(testAccVcdNsxtSegmentProfileTemplateDS, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText2DS)

	params["FuncName"] = t.Name() + "step3"
	configText3 := templateFill(testAccVcdNsxtSegmentProfileTemplateGlobalDefault, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 3: %s", configText3)

	params["FuncName"] = t.Name() + "step4"
	configText4DS := templateFill(testAccVcdNsxtSegmentProfileTemplateGlobalDefaultDS, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 4: %s", configText4DS)

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
				Check: resource.ComposeAggregateTestCheckFunc(
					// resource.TestMatchResourceAttr("vcd_nsxt_segment_profile_template.first", "id", regexp.MustCompile(`\d*`)),
					resource.TestCheckResourceAttr("vcd_nsxt_segment_profile_template.empty", "name", t.Name()+"-empty"),
					resource.TestCheckResourceAttr("vcd_nsxt_segment_profile_template.empty", "description", "description"),
					resource.TestCheckResourceAttr("vcd_nsxt_segment_profile_template.empty", "ip_discovery_profile_id", ""),
					resource.TestCheckResourceAttr("vcd_nsxt_segment_profile_template.empty", "mac_discovery_profile_id", ""),
					resource.TestCheckResourceAttr("vcd_nsxt_segment_profile_template.empty", "spoof_guard_profile_id", ""),
					resource.TestCheckResourceAttr("vcd_nsxt_segment_profile_template.empty", "qos_profile_id", ""),
					resource.TestCheckResourceAttr("vcd_nsxt_segment_profile_template.empty", "segment_security_profile_id", ""),

					resource.TestCheckResourceAttr("vcd_nsxt_segment_profile_template.complete", "name", t.Name()+"-complete"),
					resource.TestMatchResourceAttr("vcd_nsxt_segment_profile_template.complete", "ip_discovery_profile_id", regexp.MustCompile(`[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$`)),
					resource.TestMatchResourceAttr("vcd_nsxt_segment_profile_template.complete", "mac_discovery_profile_id", regexp.MustCompile(`[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$`)),
					resource.TestMatchResourceAttr("vcd_nsxt_segment_profile_template.complete", "spoof_guard_profile_id", regexp.MustCompile(`[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$`)),
					resource.TestMatchResourceAttr("vcd_nsxt_segment_profile_template.complete", "qos_profile_id", regexp.MustCompile(`[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$`)),
					resource.TestMatchResourceAttr("vcd_nsxt_segment_profile_template.complete", "segment_security_profile_id", regexp.MustCompile(`[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$`)),

					resource.TestCheckResourceAttr("vcd_nsxt_segment_profile_template.half-complete", "name", t.Name()+"-half-complete"),
					resource.TestCheckResourceAttr("vcd_nsxt_segment_profile_template.half-complete", "description", ""),
					resource.TestMatchResourceAttr("vcd_nsxt_segment_profile_template.half-complete", "ip_discovery_profile_id", regexp.MustCompile(`[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$`)),
					resource.TestMatchResourceAttr("vcd_nsxt_segment_profile_template.half-complete", "mac_discovery_profile_id", regexp.MustCompile(`[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$`)),
					resource.TestMatchResourceAttr("vcd_nsxt_segment_profile_template.half-complete", "spoof_guard_profile_id", regexp.MustCompile(`[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$`)),
					resource.TestCheckResourceAttr("vcd_nsxt_segment_profile_template.half-complete", "qos_profile_id", ""),
					resource.TestCheckResourceAttr("vcd_nsxt_segment_profile_template.half-complete", "segment_security_profile_id", ""),
				),
			},

			{
				ResourceName:      "vcd_nsxt_segment_profile_template.complete",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     t.Name() + "-complete",
			},
			{
				Config: configText2DS,
				Check: resource.ComposeAggregateTestCheckFunc(
					resourceFieldsEqual("data.vcd_nsxt_segment_profile_template.empty", "vcd_nsxt_segment_profile_template.empty", nil),
					resourceFieldsEqual("data.vcd_nsxt_segment_profile_template.half-complete", "vcd_nsxt_segment_profile_template.half-complete", nil),
					resourceFieldsEqual("data.vcd_nsxt_segment_profile_template.complete", "vcd_nsxt_segment_profile_template.complete", nil),
				),
			},
			{
				Config: configText3,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("vcd_nsxt_global_default_segment_profile_template.singleton", "id", "no-real-id"),
				),
			},
			{
				Config: configText4DS,
				Check: resource.ComposeAggregateTestCheckFunc(
					resourceFieldsEqual("data.vcd_nsxt_global_default_segment_profile_template.singleton", "vcd_nsxt_global_default_segment_profile_template.singleton", nil),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdNsxtSegmentProfileTemplate = `
data "vcd_nsxt_manager" "nsxt" {
  name = "{{.NsxtManager}}"
}

resource "vcd_nsxt_segment_profile_template" "empty" {
  name        = "{{.TestName}}-empty"
  description = "description"

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

resource "vcd_nsxt_segment_profile_template" "half-complete" {
	name        = "{{.TestName}}-half-complete"
  
	nsxt_manager_id             = data.vcd_nsxt_manager.nsxt.id
	ip_discovery_profile_id     = data.vcd_nsxt_segment_ip_discovery_profile.first.id
	mac_discovery_profile_id    = data.vcd_nsxt_segment_mac_discovery_profile.first.id
	spoof_guard_profile_id      = data.vcd_nsxt_segment_spoof_guard_profile.first.id
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
`

const testAccVcdNsxtSegmentProfileTemplateDS = testAccVcdNsxtSegmentProfileTemplate + `
data "vcd_nsxt_segment_profile_template" "empty" {
  name = vcd_nsxt_segment_profile_template.empty.name

  depends_on = [vcd_nsxt_segment_profile_template.empty]
}

data "vcd_nsxt_segment_profile_template" "half-complete" {
  name = vcd_nsxt_segment_profile_template.half-complete.name

  depends_on = [vcd_nsxt_segment_profile_template.half-complete]
}

data "vcd_nsxt_segment_profile_template" "complete" {
  name = vcd_nsxt_segment_profile_template.complete.name

  depends_on = [vcd_nsxt_segment_profile_template.complete]
}
`

const testAccVcdNsxtSegmentProfileTemplateGlobalDefault = testAccVcdNsxtSegmentProfileTemplate + `
resource "vcd_nsxt_global_default_segment_profile_template" "singleton" {
  vdc_networks_default_segment_profile_template_id = vcd_nsxt_segment_profile_template.complete.id
  vapp_networks_default_segment_profile_template_id    = vcd_nsxt_segment_profile_template.empty.id
}
`

const testAccVcdNsxtSegmentProfileTemplateGlobalDefaultDS = testAccVcdNsxtSegmentProfileTemplateGlobalDefault + `
data "vcd_nsxt_global_default_segment_profile_template" "singleton" {

  depends_on = [vcd_nsxt_global_default_segment_profile_template.singleton]
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
