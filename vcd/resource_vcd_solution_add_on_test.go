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
		t.Fatalf("Add-On config value not specified: %s", err)
	}

	catalog, err := org.GetCatalogByName(testConfig.VCD.Catalog.NsxtBackedCatalogName, false)
	if err != nil {
		t.Fatalf("Add-On config value not specified: %s", err)
	}

	localAddOnPath, err := fetchCacheFile(catalog, testConfig.VCD.Catalog.NsxtCatalogAddonDse, t)
	if err != nil {
		t.Fatalf("Add-On config value not specified: %s", err)
	}

	// cacheDir :=
	// pwd, err := os.Getwd()
	// if err != nil {
	// 	t.Fatalf("error retrieving current directory")
	// }
	// fileName := testConfig.VCD.Catalog.NsxtCatalogAddonDse
	// cacheDirPath := pwd + "/../test-resources/cache"
	// cacheFilePath := cacheDirPath + "/" + fileName

	// While the image should already be present in the catalog
	// 'testConfig.VCD.Catalog.NsxtBackedCatalogName' with name
	// 'testConfig.VCD.Catalog.NsxtCatalogAddonDse', it must also be present locally as it must be
	// extracted and some data of it should be used to create Solution Add-On itself.

	params := StringMap{
		"Org":     testConfig.VCD.Org,
		"VdcName": testConfig.Nsxt.Vdc,

		"TestName":            t.Name(),
		"CatalogName":         testConfig.VCD.Catalog.NsxtBackedCatalogName,
		"RoutedNetworkName":   testConfig.Nsxt.RoutedNetwork,
		"IsolatedNetworkName": testConfig.Nsxt.IsolatedNetwork,

		"AddonIsoPath": localAddOnPath,
		// "AddonIsoPath2": "/Users/dainius/Downloads/vmware-vcd-ds-1.3.0-22829404.iso",
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
			/* {
				Config: configText2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vcd_solution_add_on.dse14", "state", "RESOLVED"),
				),
			}, */
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

  name              = basename("{{.AddonIsoPath}}")
}

resource "vcd_solution_add_on" "dse14" {
  catalog_item_id   = data.vcd_catalog_media.dse14.catalog_item_id
  addon_path        = "{{.AddonIsoPath}}"
  trust_certificate = true
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
			// printVerbose("# Creating directory '%s'\n", cacheDirPath)
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
