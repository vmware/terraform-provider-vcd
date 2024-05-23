//go:build api || functional || ALL

package vcd

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccSolutionAddonInstance(t *testing.T) {
	preTestChecks(t)

	// if testConfig.VCD.So == "" {
	// 	t.Skipf("Add-On config value not specified ")
	// }

	vcdClient := createTemporaryVCDConnection(true)
	org, err := vcdClient.GetOrgByName(testConfig.SolutionAddOn.Org)
	if err != nil {
		t.Fatalf("error creating temporary VCD connection: %s", err)
	}

	catalog, err := org.GetCatalogByName(testConfig.SolutionAddOn.Catalog, false)
	if err != nil {
		t.Fatalf("error retrieving catalog: %s", err)
	}

	localAddOnPath, err := fetchCacheFile(catalog, testConfig.SolutionAddOn.AddonImageDse, t)
	if err != nil {
		t.Fatalf("error finding Solution Add-On cache file: %s", err)
	}

	params := StringMap{
		"Org":     testConfig.SolutionAddOn.Org,
		"VdcName": testConfig.SolutionAddOn.Vdc,

		"TestName":          t.Name(),
		"CatalogName":       testConfig.SolutionAddOn.Catalog,
		"RoutedNetworkName": testConfig.SolutionAddOn.RoutedNetwork,

		"PublishToOrg": testConfig.Cse.TenantOrg,

		"AddonIsoPath": localAddOnPath,
	}
	testParamsNotEmpty(t, params)

	params["FuncName"] = t.Name() + "step1"
	configText1 := templateFill(testAccSolutionAddonInstanceStep1, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)

	params["FuncName"] = t.Name() + "step2"
	configText2 := templateFill(testAccSolutionAddonInstanceStep2, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText2)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	cacheAddOnId := &testCachedFieldValue{}
	cacheAddOnName := &testCachedFieldValue{}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: configText1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_solution_add_on.dse14", "id"),
					resource.TestCheckResourceAttrSet("vcd_solution_add_on.dse14", "catalog_item_id"),
					resource.TestCheckResourceAttr("vcd_solution_add_on.dse14", "rde_state", "RESOLVED"),
					resource.TestCheckResourceAttr("vcd_solution_add_on.dse14", "trust_certificate", "true"),
					cacheAddOnId.cacheTestResourceFieldValue("vcd_solution_add_on.dse14", "id"),
					cacheAddOnName.cacheTestResourceFieldValue("vcd_solution_add_on.dse14", "name"),
				),
			},
			{
				Config: configText2,
				Check:  resource.ComposeTestCheckFunc(
				// resourceFieldsEqual("vcd_solution_add_on.dse14", "data.vcd_solution_add_on.dse14", []string{"%", "trust_certificate", "addon_path"}),
				),
			},
		},
	})
}

const testAccSolutionAddonInstanceStep1 = `
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

  name = basename("{{.AddonIsoPath}}")
}

resource "vcd_solution_add_on" "dse14" {
  catalog_item_id   = data.vcd_catalog_media.dse14.catalog_item_id
  addon_path        = "{{.AddonIsoPath}}"
  trust_certificate = true

  depends_on = [ vcd_solution_landing_zone.slz ]
}


resource "vcd_solution_add_on_instance" "dse14" {
  add_on_id     = vcd_solution_add_on.dse14.id
  accept_eula   = true
  instance_name = "ds-asdER"
  input = {
    name = "ds-qwer"
    input-delete-previous-uiplugin-versions = false
  }
}

data "vcd_org" "recipient" {
  name = "{{.PublishToOrg}}"
}

resource "vcd_solution_add_on_instance_publish" "public" {
  add_on_instance_id = vcd_solution_add_on_instance.dse14.id
  org_ids = [data.vcd_org.recipient.id]
  publish_to_all_tenants = false
}
`

const testAccSolutionAddonInstanceStep2 = testAccSolutionAddonInstanceStep1 + `
data "vcd_solution_add_on" "dse14" {
  name = vcd_solution_add_on.dse14.name
}

data "vcd_solution_add_on_instance_publish" "published" {
  add_on_instance_id = vcd_solution_add_on_instance.dse14.id
}
`
