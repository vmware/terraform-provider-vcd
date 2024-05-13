//go:build network || nsxt || vm || ALL || functional

package vcd

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccVcdSolutionLandingZone(t *testing.T) {
	preTestChecks(t)
	if checkVersion(testConfig.Provider.ApiVersion, "< 37.1") {
		t.Skipf("Solution Landing Zones are supported in VCD 10.4.1+. Skipping")
	}

	var params = StringMap{
		"Org":     testConfig.VCD.Org,
		"VdcName": testConfig.Nsxt.Vdc,

		"TestName":            t.Name(),
		"CatalogName":         testConfig.VCD.Catalog.NsxtBackedCatalogName,
		"RoutedNetworkName":   testConfig.Nsxt.RoutedNetwork,
		"IsolatedNetworkName": testConfig.Nsxt.IsolatedNetwork,
	}
	testParamsNotEmpty(t, params)

	configTextStep1 := templateFill(testAccSolutionLandingZoneStep1, params)

	params["FuncName"] = t.Name() + "-step2"
	configTextStep2 := templateFill(testAccSolutionLandingZoneStep2, params)

	params["FuncName"] = t.Name() + "-step3DS"
	configTextStep3DS := templateFill(testAccSolutionLandingZoneStep3DS, params)

	debugPrintf("#[DEBUG] CONFIGURATION step 1: %s\n", configTextStep1)
	debugPrintf("#[DEBUG] CONFIGURATION step 2: %s\n", configTextStep2)
	debugPrintf("#[DEBUG] CONFIGURATION step 3: %s\n", configTextStep3DS)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckSlzDestroy(),
		Steps: []resource.TestStep{
			{
				Config: configTextStep1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vcd_solution_landing_zone.slz", "org", testConfig.VCD.Org),
					resource.TestCheckResourceAttrSet("vcd_solution_landing_zone.slz", "id"),
					resource.TestCheckResourceAttr("vcd_solution_landing_zone.slz", "state", "RESOLVED"),
					resource.TestCheckResourceAttr("vcd_solution_landing_zone.slz", "catalog.#", "1"),

					resource.TestCheckTypeSetElemNestedAttrs("vcd_solution_landing_zone.slz", "catalog.*", map[string]string{"name": testConfig.VCD.Catalog.NsxtBackedCatalogName}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_solution_landing_zone.slz", "vdc.*.compute_policy.*", map[string]string{"name": "System Default", "is_default": "true"}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_solution_landing_zone.slz", "vdc.*.org_vdc_network.*", map[string]string{"name": testConfig.Nsxt.RoutedNetwork, "is_default": "true"}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_solution_landing_zone.slz", "vdc.*.storage_policy.*", map[string]string{"name": "*", "is_default": "true"}),
					resource.TestCheckResourceAttr("vcd_solution_landing_zone.slz", "vdc.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_solution_landing_zone.slz", "vdc.*", map[string]string{"is_default": "true", "name": testConfig.Nsxt.Vdc}),
				),
			},
			{
				Config: configTextStep2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vcd_solution_landing_zone.slz", "org", testConfig.VCD.Org),
					resource.TestCheckResourceAttrSet("vcd_solution_landing_zone.slz", "id"),
					resource.TestCheckResourceAttr("vcd_solution_landing_zone.slz", "state", "RESOLVED"),
					resource.TestCheckResourceAttr("vcd_solution_landing_zone.slz", "catalog.#", "1"),

					resource.TestCheckTypeSetElemNestedAttrs("vcd_solution_landing_zone.slz", "catalog.*", map[string]string{"name": testConfig.VCD.Catalog.NsxtBackedCatalogName}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_solution_landing_zone.slz", "vdc.*.compute_policy.*", map[string]string{"name": "System Default", "is_default": "false"}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_solution_landing_zone.slz", "vdc.*.org_vdc_network.*", map[string]string{"name": testConfig.Nsxt.IsolatedNetwork, "is_default": "false"}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_solution_landing_zone.slz", "vdc.*.storage_policy.*", map[string]string{"name": "*", "is_default": "false"}),
					resource.TestCheckResourceAttr("vcd_solution_landing_zone.slz", "vdc.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_solution_landing_zone.slz", "vdc.*", map[string]string{"is_default": "true", "name": testConfig.Nsxt.Vdc}),
				),
			},
			{
				Config: configTextStep3DS,
				Check: resource.ComposeTestCheckFunc(
					resourceFieldsEqual("data.vcd_solution_landing_zone.slz", "vcd_solution_landing_zone.slz", nil),
				),
			},
			{
				ResourceName:      "vcd_solution_landing_zone.slz",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
	postTestChecks(t)
}

const testAccSolutionLandingZoneStep1 = `
data "vcd_catalog" "nsxt" {
  org  = "{{.Org}}"
  name = "{{.CatalogName}}"
}

data "vcd_org_vdc" "vdc1" {
  org  = "{{.Org}}"
  name = "{{.VdcName}}"
}

data "vcd_network_routed_v2" "r1" {
  org  = "{{.Org}}"
  vdc  = "{{.VdcName}}"
  name = "{{.RoutedNetworkName}}"
}

data "vcd_storage_profile" "sp" {
  org  = "{{.Org}}"
  vdc  = "{{.VdcName}}"
  name = "*"
}

resource "vcd_solution_landing_zone" "slz" {
  org = "{{.Org}}"

  catalog {
	id           = data.vcd_catalog.nsxt.id
  }

  vdc {
	id         = data.vcd_org_vdc.vdc1.id
	is_default = true

	org_vdc_network {
	  id         = data.vcd_network_routed_v2.r1.id
	  is_default = true
	}

	compute_policy {
	  id         = data.vcd_org_vdc.vdc1.default_compute_policy_id
	  is_default = true
	}

	storage_policy {
	  id         = data.vcd_storage_profile.sp.id
	  is_default = true
	}
  }
}
`

const testAccSolutionLandingZoneStep2 = `
data "vcd_catalog" "nsxt" {
  org  = "{{.Org}}"
  name = "{{.CatalogName}}"
}

data "vcd_org_vdc" "vdc1" {
  org  = "{{.Org}}"
  name = "{{.VdcName}}"
}

data "vcd_network_routed_v2" "r1" {
  org  = "{{.Org}}"
  vdc  = "{{.VdcName}}"
  name = "{{.RoutedNetworkName}}"
}

data "vcd_network_isolated_v2" "i1" {
  org  = "{{.Org}}"
  vdc  = "{{.VdcName}}"
  name = "{{.IsolatedNetworkName}}"
}

data "vcd_storage_profile" "sp" {
  org  = "{{.Org}}"
  vdc  = "{{.VdcName}}"
  name = "*"
}

resource "vcd_solution_landing_zone" "slz" {
  org = "{{.Org}}"

  catalog {
	id = data.vcd_catalog.nsxt.id
  }

  vdc {
	id         = data.vcd_org_vdc.vdc1.id
	is_default = true

	org_vdc_network {
	  id = data.vcd_network_isolated_v2.i1.id
	}

	compute_policy {
	  id = data.vcd_org_vdc.vdc1.default_compute_policy_id
	}

	storage_policy {
	  id = data.vcd_storage_profile.sp.id
	}
  }
}
`

const testAccSolutionLandingZoneStep3DS = testAccSolutionLandingZoneStep2 + `
data "vcd_solution_landing_zone" "slz" {
  depends_on = [vcd_solution_landing_zone.slz]
}
`

func testAccCheckSlzDestroy() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*VCDClient)
		slz, err := conn.GetExactlyOneSolutionLandingZone()
		if err == nil {
			return fmt.Errorf("there is still an RDE for Solution Landing Zone: %s", slz.RdeId())
		}

		return nil
	}
}
