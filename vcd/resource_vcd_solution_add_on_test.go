//go:build api || functional || ALL

package vcd

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/vmware/go-vcloud-director/v2/govcd"
)

func TestAccSolutionAddon(t *testing.T) {
	preTestChecks(t)

	if testConfig.VCD.Catalog.NsxtCatalogAddonDse == "" {
		t.Skipf("Add-On config value not specified ")
	}

	vcdClient := createTemporaryVCDConnection(true)
	org, err := vcdClient.GetOrgByName(testConfig.VCD.Org)
	if err != nil {
		t.Fatalf("rrror creating temporary VCD connection: %s", err)
	}

	catalog, err := org.GetCatalogByName(testConfig.VCD.Catalog.NsxtBackedCatalogName, false)
	if err != nil {
		t.Fatalf("error retrieving catalog: %s", err)
	}

	localAddOnPath, err := fetchCacheFile(catalog, testConfig.VCD.Catalog.NsxtCatalogAddonDse, t)
	if err != nil {
		t.Fatalf("error finding Solution Add-On cache file: %s", err)
	}

	params := StringMap{
		"Org":     testConfig.VCD.Org,
		"VdcName": testConfig.Nsxt.Vdc,

		"TestName":            t.Name(),
		"CatalogName":         testConfig.VCD.Catalog.NsxtBackedCatalogName,
		"RoutedNetworkName":   testConfig.Nsxt.RoutedNetwork,
		"IsolatedNetworkName": testConfig.Nsxt.IsolatedNetwork,

		"AddonIsoPath": localAddOnPath,
	}
	testParamsNotEmpty(t, params)

	params["FuncName"] = t.Name() + "step1"
	configText1 := templateFill(testAccSolutionAddonStep1, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

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
}
`

func fetchCacheFile(catalog *govcd.Catalog, fileName string, t *testing.T) (string, error) {
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("error getting current working directory: %s", err)
	}

	cacheDirPath := pwd + "/.." + "/test-resources/cache"
	cacheFilePath := cacheDirPath + "/" + fileName

	if _, err := os.Stat(cacheFilePath); errors.Is(err, os.ErrNotExist) {
		// Create cache directory if it doesn't exist
		if _, err := os.Stat(cacheDirPath); os.IsNotExist(err) {
			err := os.Mkdir(cacheDirPath, 0750)
			if err != nil {
				t.Fatalf("error creating cache directory: %s", err)
			}
		}

		fmt.Printf("# Solution Add-On image is not in cache, downloading  '%s' from VCD...", fileName)
		addOnMediaItem, err := catalog.GetMediaByName(fileName, false)
		if err != nil {
			t.Fatalf("error getting catalog media item: %s", err)
		}

		addOn, err := addOnMediaItem.Download()
		if err != nil {
			t.Fatalf("error getting download link: %s", err)
		}

		err = os.WriteFile(cacheFilePath, addOn, 0600)
		if err != nil {
			t.Fatalf("error writing file: %s", err)
		}

		addOn = nil // free memory
		fmt.Println("Done")
	}

	return cacheFilePath, nil
}
