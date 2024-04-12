//go:build network || nsxt || vm || ALL || functional

package vcd

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcdSolutionLandingZone(t *testing.T) {
	preTestChecks(t)

	var params = StringMap{
		"Org":     testConfig.VCD.Org,
		"VdcName": testConfig.Nsxt.Vdc,

		"CatalogName":         testConfig.VCD.Catalog.NsxtBackedCatalogName,
		"RoutedNetworkName":   testConfig.Nsxt.RoutedNetwork,
		"IsolatedNetworkName": testConfig.Nsxt.IsolatedNetwork,

		"TestName": t.Name(),
	}
	testParamsNotEmpty(t, params)

	configText := templateFill(testAccSolutionLandingZoneStep1, params)

	params["FuncName"] = t.Name() + "-step2"
	configTextStep2 := templateFill(testAccSolutionLandingZoneStep2, params)

	params["FuncName"] = t.Name() + "-step3DS"
	configTextStep3DS := templateFill(testAccSolutionLandingZoneStep3DS, params)

	debugPrintf("#[DEBUG] CONFIGURATION: %s\n", configText)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		// CheckDestroy:      testAccCheckSecurityTagDestroy(tag1),
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check:  resource.ComposeTestCheckFunc(
				// testAccCheckSecurityTagCreated(tag1),
				// testAccCheckSecurityTagOnVMCreated(tag1, vAppName, firstVMName),
				// testAccCheckSecurityTagOnVMCreated(tag1, vAppName, secondVMName),
				),
			},
			{
				Config: configTextStep2,
				Check:  resource.ComposeTestCheckFunc(
				// sleepTester(2 * time.Minute),
				// testAccCheckSecurityTagCreated(tag1, tag2),
				// testAccCheckSecurityTagOnVMCreated(tag1, vAppName, firstVMName),
				// testAccCheckSecurityTagOnVMCreated(tag1, vAppName, secondVMName),
				),
			},
			{
				Config: configTextStep3DS,
				Check: resource.ComposeTestCheckFunc(
					resourceFieldsEqual("data.vcd_solution_landing_zone.slz", "vcd_solution_landing_zone.slz", nil),
				// sleepTester(2 * time.Minute),
				// testAccCheckSecurityTagCreated(tag1, tag2),
				// testAccCheckSecurityTagOnVMCreated(tag1, vAppName, firstVMName),
				// testAccCheckSecurityTagOnVMCreated(tag1, vAppName, secondVMName),
				),
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
	capabilities = ["one", "two", "three"]
  }

  vdc {
	id         = data.vcd_org_vdc.vdc1.id
	is_default = true

	org_vdc_network {
	  id = data.vcd_network_routed_v2.r1.id
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
	  id = data.vcd_network_routed_v2.r1.id
	}

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
data "vcd_solution_landing_zone" "slz" {}
`
