//go:build nsxt || network || vdcGroup || functional || ALL
// +build nsxt network vdcGroup functional ALL

package vcd

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcdNsxtNetworkContextProfileInVdc(t *testing.T) {
	preTestChecks(t)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	skipNoNsxtConfiguration(t)

	var params = StringMap{
		"Org":     testConfig.VCD.Org,
		"NsxtVdc": testConfig.Nsxt.Vdc,
		"Tags":    "nsxt network vdcGroup",
	}

	configText := templateFill(testAccVcdNsxtNetworkContextProfileDS, params)

	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testParamsNotEmpty(t, params) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("data.vcd_nsxt_network_context_profile.p", "id", regexp.MustCompile("urn:vcloud:networkContextProfile:")),
					resource.TestCheckResourceAttr("data.vcd_nsxt_network_context_profile.p", "name", "CTRXICA"),
					resource.TestCheckResourceAttr("data.vcd_nsxt_network_context_profile.p", "scope", "SYSTEM"),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdNsxtNetworkContextProfileDS = `
data "vcd_org_vdc" "nsxt" {
  org  = "{{.Org}}"
  name = "{{.NsxtVdc}}"
}
data "vcd_nsxt_network_context_profile" "p" {
  context_id = data.vcd_org_vdc.nsxt.id
  name       = "CTRXICA"
}
`

func TestAccVcdNsxtNetworkContextProfileInNsxtManager(t *testing.T) {
	preTestChecks(t)
	if !usingSysAdmin() {
		t.Skipf("this test requires Sysadmin user to lookup NSX-T Manager")
	}
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	skipNoNsxtConfiguration(t)

	var params = StringMap{
		"Org":             testConfig.VCD.Org,
		"NsxtManagerName": testConfig.Nsxt.Manager,
		"Scope":           "SYSTEM",
		"Tags":            "nsxt network vdcGroup",
	}

	configText1 := templateFill(testAccVcdNsxtNetworkContextProfileNsxtManagerDS, params)
	debugPrintf("#[DEBUG] CONFIGURATION - step 1: %s", configText1)

	params["FuncName"] = t.Name() + "-step2"
	params["Scope"] = "PROVIDER"
	configText2 := templateFill(testAccVcdNsxtNetworkContextProfileNsxtManagerDS, params)
	debugPrintf("#[DEBUG] CONFIGURATION - step 2: %s", configText2)

	params["FuncName"] = t.Name() + "-step3"
	params["Scope"] = "TENANT"
	configText3 := templateFill(testAccVcdNsxtNetworkContextProfileNsxtManagerDS, params)
	debugPrintf("#[DEBUG] CONFIGURATION - step 3: %s", configText3)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testParamsNotEmpty(t, params) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: configText1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("data.vcd_nsxt_network_context_profile.p", "id", regexp.MustCompile("urn:vcloud:networkContextProfile:")),
					resource.TestCheckResourceAttr("data.vcd_nsxt_network_context_profile.p", "name", "CTRXICA"),
					resource.TestCheckResourceAttr("data.vcd_nsxt_network_context_profile.p", "scope", "SYSTEM"),
				),
			},
			{
				Config: configText2,
				// VCD has no capability to create items in PROVIDER context yet
				ExpectError: regexp.MustCompile("entity not found"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("data.vcd_nsxt_network_context_profile.p", "id", regexp.MustCompile("urn:vcloud:networkContextProfile:")),
					resource.TestCheckResourceAttr("data.vcd_nsxt_network_context_profile.p", "name", "CTRXICA"),
					resource.TestCheckResourceAttr("data.vcd_nsxt_network_context_profile.p", "scope", "PROVIDER"),
				),
			},
			{
				Config: configText3,
				// VCD has no capability to create items in TENANT context yet
				ExpectError: regexp.MustCompile("entity not found"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("data.vcd_nsxt_network_context_profile.p", "id", regexp.MustCompile("urn:vcloud:networkContextProfile:")),
					resource.TestCheckResourceAttr("data.vcd_nsxt_network_context_profile.p", "name", "CTRXICA"),
					resource.TestCheckResourceAttr("data.vcd_nsxt_network_context_profile.p", "scope", "TENANT"),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdNsxtNetworkContextProfileNsxtManagerDS = `
data "vcd_nsxt_manager" "main" {
  name = "{{.NsxtManagerName}}"
}

data "vcd_nsxt_network_context_profile" "p" {
  context_id = data.vcd_nsxt_manager.main.id
  name       = "CTRXICA"
  scope      = "{{.Scope}}"
}
`
