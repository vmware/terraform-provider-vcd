//go:build api || functional || ALL

package vcd

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccSolutionAddon(t *testing.T) {
	preTestChecks(t)
	skipTestForServiceAccountAndApiToken(t)

	params := StringMap{
		"Org":     testConfig.VCD.Org,
		"VdcName": testConfig.Nsxt.Vdc,

		"TestName":            t.Name(),
		"CatalogName":         testConfig.VCD.Catalog.NsxtBackedCatalogName,
		"RoutedNetworkName":   testConfig.Nsxt.RoutedNetwork,
		"IsolatedNetworkName": testConfig.Nsxt.IsolatedNetwork,

		"AddonIsoPath":  "/Users/dainius/Downloads/vmware-vcd-ds-1.4.0-23376809.iso",
		"AddonIsoPath2": "/Users/dainius/Downloads/vmware-vcd-ds-1.3.0-22829404.iso",
	}
	testParamsNotEmpty(t, params)

	params["FuncName"] = t.Name() + "step1"
	configText1 := templateFill(testAccSolutionAddonStep1, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)

	params["FuncName"] = t.Name() + "step2"
	configText2 := templateFill(testAccSolutionAddonStep2, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText2)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		// CheckDestroy:      testAccCheckServiceAccountDestroy(params["Org"].(string), params["SaName"].(string)),
		Steps: []resource.TestStep{
			{
				Config: configText1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vcd_solution_add_on.dse14", "state", "RESOLVED"),
				),
			},
			{
				Config: configText2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vcd_solution_add_on.dse14", "state", "RESOLVED"),
				),
			},
		},
	})
}

const testAccSolutionAddonStep1 = `
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

data "vcd_catalog_media" "dse14" {
  org        = "{{.Org}}"
  catalog_id = data.vcd_catalog.nsxt.id

  name              = basename("{{.AddonIsoPath}}")
  #description       = "new os versions"
  #media_path        = "{{.AddonIsoPath}}"
  #upload_any_file   = false # Add-ons are packaged in '.iso' files
  #upload_piece_size = 10
}

#resource "vcd_catalog_media" "dse14" {
#  org        = "{{.Org}}"
#  catalog_id = data.vcd_catalog.nsxt.id
#
#  name              = basename("{{.AddonIsoPath}}")
#  description       = "new os versions"
#  media_path        = "{{.AddonIsoPath}}"
#  upload_any_file   = false # Add-ons are packaged in '.iso' files
#  upload_piece_size = 10
#}

resource "vcd_solution_add_on" "dse14" {
  org               = "{{.Org}}"
  catalog_item_id   = data.vcd_catalog_media.dse14.catalog_item_id
  addon_path        = "{{.AddonIsoPath}}"
  trust_certificate = true
  accept_eula       = true
}
`

const testAccSolutionAddonStep2 = `
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

data "vcd_catalog_media" "dse14" {
  org        = "{{.Org}}"
  catalog_id = data.vcd_catalog.nsxt.id

  name              = basename("{{.AddonIsoPath}}")
  #description       = "new os versions"
  #media_path        = "{{.AddonIsoPath}}"
  #upload_any_file   = false # Add-ons are packaged in '.iso' files
  #upload_piece_size = 10
}

resource "vcd_catalog_media" "dse13" {
  org        = "{{.Org}}"
  catalog_id = data.vcd_catalog.nsxt.id

  name              = basename("{{.AddonIsoPath}}")
  description       = "new os versions"
  media_path        = "{{.AddonIsoPath2}}"
  upload_any_file   = false # Add-ons are packaged in '.iso' files
  upload_piece_size = 10
}

resource "vcd_solution_add_on" "dse14" {
  org               = "{{.Org}}"
  catalog_item_id   = vcd_catalog_media.dse13.catalog_item_id
  addon_path        = "{{.AddonIsoPath2}}"
  trust_certificate = true
  accept_eula       = true
}
`
