//go:build nsxt || alb || ALL || functional

package vcd

import (
	"fmt"
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
					resource.TestCheckResourceAttr("vcd_nsxt_segment_profile_template.complete", "name", t.Name()+"-complete"),

					sleepTester(1*time.Minute),
				),
			},
			// {
			// 	Config: configText2,
			// 	Check: resource.ComposeAggregateTestCheckFunc(
			// 		resource.TestMatchResourceAttr("vcd_nsxt_alb_controller.first", "id", regexp.MustCompile(`\d*`)),
			// 		resource.TestCheckResourceAttr("vcd_nsxt_alb_controller.first", "name", t.Name()+"-renamed"),
			// 		resource.TestCheckResourceAttr("vcd_nsxt_alb_controller.first", "description", ""),
			// 		resource.TestCheckResourceAttr("vcd_nsxt_alb_controller.first", "username", testConfig.Nsxt.NsxtSegmentProfileTemplateUser),
			// 		resource.TestCheckResourceAttr("vcd_nsxt_alb_controller.first", "password", testConfig.Nsxt.NsxtSegmentProfileTemplatePassword),
			// 		checkLicenseType("vcd_nsxt_alb_controller.first", true, isVersionLessThan37),
			// 	),
			// },
			// {
			// 	ResourceName:            "vcd_nsxt_alb_controller.first",
			// 	ImportState:             true,
			// 	ImportStateVerify:       true,
			// 	ImportStateId:           t.Name() + "-renamed",
			// 	ImportStateVerifyIgnore: []string{"password"},
			// },
			// {
			// 	Config: configText4,
			// 	Check: resource.ComposeAggregateTestCheckFunc(
			// 		resource.TestMatchResourceAttr("vcd_nsxt_alb_controller.first", "id", regexp.MustCompile(`\d*`)),
			// 		// Comparing all data source fields against resource. Field '%' (total attribute number) is skipped
			// 		// because data source does not have password field
			// 		resourceFieldsEqual("data.vcd_nsxt_alb_controller.first", "vcd_nsxt_alb_controller.first", []string{"%"}),
			// 	),
			// },
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

func sleepTester(d time.Duration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		fmt.Printf("sleeping %s\n", d.String())
		time.Sleep(d)
		fmt.Println("finished sleeping")
		return nil
	}
}
