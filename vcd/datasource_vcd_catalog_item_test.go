// +build catalog ALL functional

package vcd

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

// Test catalog and catalog item data sources
// Using a catalog data source we reference a catalog item data source
// Using a catalog item data source we create another catalog item
// where the description is the first data source ID
func TestAccVcdCatalogAndItemDatasource(t *testing.T) {
	var TestCatalogItemDS = "TestCatalogItemDS"

	var params = StringMap{
		"Org":             testConfig.VCD.Org,
		"Catalog":         testSuiteCatalogName,
		"CatalogItem":     testSuiteCatalogOVAItem,
		"NewCatalogItem":  TestCatalogItemDS,
		"OvaPath":         testConfig.Ova.OvaPath,
		"UploadPieceSize": testConfig.Ova.UploadPieceSize,
		"UploadProgress":  testConfig.Ova.UploadProgress,
		"Tags":            "catalog",
	}

	configText := templateFill(testAccCheckVcdCatalogItemDS, params)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

	datasourceCatalog := "data.vcd_catalog." + testSuiteCatalogName
	datasourceCatalogItem := "data.vcd_catalog_item." + testSuiteCatalogOVAItem
	resourceCatalogItem := "vcd_catalog_item." + TestCatalogItemDS
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { preRunChecks(t) },
		Providers:    testAccProviders,
		CheckDestroy: catalogItemDestroyed(testSuiteCatalogName, TestCatalogItemDS),
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdCatalogItemExists("vcd_catalog_item."+TestCatalogItemDS),
					resource.TestCheckResourceAttr(
						resourceCatalogItem, "name", TestCatalogItemDS),
					resource.TestCheckResourceAttrPair(
						datasourceCatalog, "name",
						resourceCatalogItem, "catalog"),
					// The description of the new catalog item was created using
					// the ID of the catalog item data source
					resource.TestCheckResourceAttrPair(
						datasourceCatalogItem, "id",
						resourceCatalogItem, "description"),
					resource.TestCheckResourceAttr(
						resourceCatalogItem, "metadata.catalogItem_metadata", "catalogItem Metadata"),
					resource.TestCheckResourceAttr(
						resourceCatalogItem, "metadata.catalogItem_metadata2", "catalogItem Metadata2"),
				),
			},
			resource.TestStep{
				ResourceName:      "vcd_catalog_item." + TestCatalogItemDS + "-import",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdOrgCatalogObject(testConfig, TestCatalogItemDS),
				// These fields can't be retrieved from catalog item data
				ImportStateVerifyIgnore: []string{"ova_path", "upload_piece_size", "show_upload_progress"},
			},
		},
	})
}

func TestAccVcdCatalogItemFilter(t *testing.T) {
	var params = StringMap{
		"Org":         testConfig.VCD.Org,
		"Catalog":     testConfig.VCD.Catalog.Name,
		"CatalogItem": "mystery",
		// This is a bad regular expression (full name)
		// but it's the only one that allows us to retrieve
		// the entity consistently
		// A better one would be testConfig.VCD.Catalog.CatalogItem[0:6],
		// but only if we don't have more recent items with similar name
		"NameRegex": testConfig.VCD.Catalog.CatalogItem,
		"Tags":      "catalog",
	}

	configText := templateFill(testAccCatalogItemFilter, params)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { preRunChecks(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckOutput("name", testConfig.VCD.Catalog.CatalogItem),
				),
			},
		},
	})
}

func catalogItemDestroyed(catalog, itemName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*VCDClient)
		org, err := conn.GetOrgByName(testConfig.VCD.Org)
		if err != nil {
			return err
		}
		cat, err := org.GetCatalogByName(catalog, false)
		if err != nil {
			return err
		}
		_, err = cat.GetCatalogItemByName(itemName, false)
		if err == nil {
			return fmt.Errorf("catalog item %s not deleted", itemName)
		}
		return nil
	}
}

const testAccCatalogItemFilter = `
data "vcd_catalog_item" "{{.CatalogItem}}" {
  org     = "{{.Org}}"
  catalog = "{{.Catalog}}"
  # name    = "{{.CatalogItem}}"
  filter {
    name_regex = "{{.NameRegex}}"
    latest     = false
    //metadata {
    // key       = "ONE"
    // value     = "\\w+"
    //}
    //metadata {
    // key       = "TWO"
    // value     = "\\w+ \\w+"
    //}
  }
}

output "name" {
  value = data.vcd_catalog_item.{{.CatalogItem}}.name
}
`

const testAccCheckVcdCatalogItemDS = `
data "vcd_catalog" "{{.Catalog}}" {
  org  = "{{.Org}}"
  name = "{{.Catalog}}"
}

data "vcd_catalog_item" "{{.CatalogItem}}" {
  org     = "{{.Org}}"
  catalog = data.vcd_catalog.{{.Catalog}}.name
  name    = "{{.CatalogItem}}"
}

resource "vcd_catalog_item" "{{.NewCatalogItem}}" {
  org     = "{{.Org}}"
  catalog = data.vcd_catalog.{{.Catalog}}.name

  name                 = "{{.NewCatalogItem}}"
  description          = data.vcd_catalog_item.{{.CatalogItem}}.id
  ova_path             = "{{.OvaPath}}"
  upload_piece_size    = {{.UploadPieceSize}}
  show_upload_progress = "{{.UploadProgress}}"

  metadata = {
    catalogItem_metadata = "catalogItem Metadata"
    catalogItem_metadata2 = "catalogItem Metadata2"
  }
}
`
