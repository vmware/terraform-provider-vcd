//go:build network || nsxt || vm || ALL || functional

package vcd

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
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

	params["FuncName"] = t.Name() + "-update"
	configTextUpdate := templateFill(testAccSolutionLandingZoneStep2, params)

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
				Config: configTextUpdate,
				Check: resource.ComposeTestCheckFunc(
					sleepTester(2 * time.Minute),
				// testAccCheckSecurityTagCreated(tag1, tag2),
				// testAccCheckSecurityTagOnVMCreated(tag1, vAppName, firstVMName),
				// testAccCheckSecurityTagOnVMCreated(tag1, vAppName, secondVMName),
				),
			},
			// {
			// 	ResourceName:      fmt.Sprintf("vcd_security_tag.%s", tag1),
			// 	ImportState:       true,
			// 	ImportStateVerify: true,
			// 	ImportStateId:     testConfig.VCD.Org + "." + tag1,
			// },
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
	id = data.vcd_catalog.nsxt.id
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

func sleepTester(d time.Duration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		fmt.Printf("sleeping %s\n", d.String())
		time.Sleep(d)
		fmt.Println("finished sleeping")
		return nil
	}
}
