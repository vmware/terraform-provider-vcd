//go:build network || nsxt || ALL || functional
// +build network nsxt ALL functional

package vcd

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// TestAccVcdNsxtAppPortProfileDsSystem tests if a built-in SYSTEM scope application port profile can be read
func TestAccVcdNsxtAppPortProfileDsSystem(t *testing.T) {
	preTestChecks(t)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	vcdClient := createTemporaryVCDConnection()
	if vcdClient.Client.APIVCDMaxVersionIs("< 34.0") {
		t.Skip(t.Name() + " requires at least API v34.0 (vCD 10.1.1+)")
	}
	skipNoNsxtConfiguration(t)

	var params = StringMap{
		"Org":         testConfig.VCD.Org,
		"NsxtVdc":     testConfig.Nsxt.Vdc,
		"ProfileName": "Active Directory Server", // Existing System built-in Application Port Profile
		"Scope":       "SYSTEM",
		"Tags":        "nsxt network",
	}

	configText1 := templateFill(testAccVcdNsxtAppPortProfileSystemDSStep1, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		PreCheck:          func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_app_port_profile.custom", "id"),
					resource.TestCheckResourceAttr("data.vcd_nsxt_app_port_profile.custom", "name", "Active Directory Server"),
					resource.TestCheckResourceAttr("data.vcd_nsxt_app_port_profile.custom", "scope", "SYSTEM"),
					resource.TestCheckTypeSetElemAttr("data.vcd_nsxt_app_port_profile.custom", "app_port.*.port.*", "464"),
					resource.TestCheckTypeSetElemNestedAttrs("data.vcd_nsxt_app_port_profile.custom", "app_port.*", map[string]string{
						"protocol": "TCP",
					}),
				),
			},
		},
	})
	postTestChecks(t)
}

// TestAccVcdNsxtAppPortProfileDsSystem tests if "Active Directory Server" Application Port Profile is not found in
// PROVIDER context (because it is defined in SYSTEM context)
func TestAccVcdNsxtAppPortProfileDsProviderNotFound(t *testing.T) {
	preTestChecks(t)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	vcdClient := createTemporaryVCDConnection()
	if vcdClient.Client.APIVCDMaxVersionIs("< 34.0") {
		t.Skip(t.Name() + " requires at least API v34.0 (vCD 10.1.1+)")
	}
	skipNoNsxtConfiguration(t)

	var params = StringMap{
		"Org":         testConfig.VCD.Org,
		"NsxtVdc":     testConfig.Nsxt.Vdc,
		"ProfileName": "Active Directory Server",
		"Scope":       "PROVIDER",
		"Tags":        "nsxt network",
	}

	configText1 := templateFill(testAccVcdNsxtAppPortProfileSystemDSStep1, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		PreCheck:          func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			resource.TestStep{
				Config:      configText1,
				ExpectError: regexp.MustCompile(`.*ENF.*`),
			},
		},
	})
	postTestChecks(t)
}

// TestAccVcdNsxtAppPortProfileDsTenantNotFound tests if "Active Directory Server" Application Port Profile is not found in
// TENANT context (because it is defined in SYSTEM context)
func TestAccVcdNsxtAppPortProfileDsTenantNotFound(t *testing.T) {
	preTestChecks(t)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	vcdClient := createTemporaryVCDConnection()
	if vcdClient.Client.APIVCDMaxVersionIs("< 34.0") {
		t.Skip(t.Name() + " requires at least API v34.0 (vCD 10.1.1+)")
	}
	skipNoNsxtConfiguration(t)

	var params = StringMap{
		"Org":         testConfig.VCD.Org,
		"NsxtVdc":     testConfig.Nsxt.Vdc,
		"ProfileName": "Active Directory Server",
		"Scope":       "TENANT",
		"Tags":        "nsxt network",
	}

	configText1 := templateFill(testAccVcdNsxtAppPortProfileSystemDSStep1, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		PreCheck:          func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			resource.TestStep{
				Config:      configText1,
				ExpectError: regexp.MustCompile(`.*ENF.*`),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdNsxtAppPortProfileSystemDSStep1 = `
data "vcd_nsxt_app_port_profile" "custom" {
  org  = "{{.Org}}"
  vdc  = "{{.NsxtVdc}}"

  name  = "{{.ProfileName}}"
  scope = "{{.Scope}}"
}
`
