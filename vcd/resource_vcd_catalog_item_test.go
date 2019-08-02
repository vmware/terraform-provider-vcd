// +build catalog ALL functional

package vcd

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/vmware/go-vcloud-director/v2/govcd"
)

var TestAccVcdCatalogItem = "TestAccVcdCatalogItemBasic"
var TestAccVcdCatalogItemDescription = "TestAccVcdCatalogItemBasicDescription"

func TestAccVcdCatalogItemBasic(t *testing.T) {

	var catalogItem govcd.CatalogItem
	var params = StringMap{
		"Org":             testConfig.VCD.Org,
		"Catalog":         testSuiteCatalogName,
		"CatalogItemName": TestAccVcdCatalogItem,
		"Description":     TestAccVcdCatalogItemDescription,
		"OvaPath":         testConfig.Ova.OvaPath,
		"UploadPieceSize": testConfig.Ova.UploadPieceSize,
		"UploadProgress":  testConfig.Ova.UploadProgress,
		"Tags":            "catalog",
	}

	configText := templateFill(testAccCheckVcdCatalogItemBasic, params)
	params["FuncName"] = t.Name() + "-Update"
	updateConfigText := templateFill(testAccCheckVcdCatalogItemUpdate, params)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { preRunChecks(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCatalogItemDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdCatalogItemExists("vcd_catalog_item."+TestAccVcdCatalogItem, &catalogItem),
					resource.TestCheckResourceAttr(
						"vcd_catalog_item."+TestAccVcdCatalogItem, "name", TestAccVcdCatalogItem),
					resource.TestCheckResourceAttr(
						"vcd_catalog_item."+TestAccVcdCatalogItem, "description", TestAccVcdCatalogItemDescription),
					resource.TestCheckResourceAttr(
						"vcd_catalog_item."+TestAccVcdCatalogItem, "metadata.catalogItem_metadata", "catalogItem Metadata"),
					resource.TestCheckResourceAttr(
						"vcd_catalog_item."+TestAccVcdCatalogItem, "metadata.catalogItem_metadata2", "catalogItem Metadata2"),
				),
			},
			resource.TestStep{
				Config: updateConfigText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdCatalogItemExists("vcd_catalog_item."+TestAccVcdCatalogItem, &catalogItem),
					resource.TestCheckResourceAttr(
						"vcd_catalog_item."+TestAccVcdCatalogItem, "name", TestAccVcdCatalogItem),
					resource.TestCheckResourceAttr(
						"vcd_catalog_item."+TestAccVcdCatalogItem, "description", TestAccVcdCatalogItemDescription),
					resource.TestCheckResourceAttr(
						"vcd_catalog_item."+TestAccVcdCatalogItem, "metadata.catalogItem_metadata", "catalogItem Metadata v2"),
					resource.TestCheckResourceAttr(
						"vcd_catalog_item."+TestAccVcdCatalogItem, "metadata.catalogItem_metadata2", "catalogItem Metadata2 v2"),
					resource.TestCheckResourceAttr(
						"vcd_catalog_item."+TestAccVcdCatalogItem, "metadata.catalogItem_metadata3", "catalogItem Metadata3"),
				),
			},
		},
	})
}

func preRunChecks(t *testing.T) {
	testAccPreCheck(t)
	checkOvaPath(t)
}

func checkOvaPath(t *testing.T) {
	file, err := os.Stat(testConfig.Ova.OvaPath)
	if err != nil {
		t.Fatal("configured catalog item issue. Configured: ", testConfig.Ova.OvaPath, err)
	}
	if os.IsNotExist(err) {
		t.Fatal("configured catalog item isn't found. Configured: ", testConfig.Ova.OvaPath)
	}
	if file.IsDir() {
		t.Fatal("configured catalog item is dir and not a file. Configured: ", testConfig.Ova.OvaPath)
	}
}

func testAccCheckVcdCatalogItemExists(itemName string, catalogItem *govcd.CatalogItem) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		catalogItemRs, ok := s.RootModule().Resources[itemName]
		if !ok {
			return fmt.Errorf("not found: %s", itemName)
		}

		if catalogItemRs.Primary.ID == "" {
			return fmt.Errorf("no catalog item ID is set")
		}

		conn := testAccProvider.Meta().(*VCDClient)

		org, _, err := conn.GetOrgAndVdc(testConfig.VCD.Org, testConfig.VCD.Vdc)
		if err != nil {
			return fmt.Errorf(errorRetrievingOrg, testConfig.VCD.Org+" and error: "+err.Error())
		}

		catalog, err := org.FindCatalog(testSuiteCatalogName)
		if err != nil {
			return fmt.Errorf("catalog %s does not exist (%#v)", testSuiteCatalogName, catalog.Catalog)
		}

		newCatalogItem, err := catalog.FindCatalogItem(catalogItemRs.Primary.Attributes["name"])
		if err != nil {
			return fmt.Errorf("catalog item %s does not exist (%#v)", catalogItemRs.Primary.ID, catalogItem.CatalogItem)
		}

		catalogItem = &newCatalogItem
		return nil
	}
}

func testAccCheckCatalogItemDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*VCDClient)
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "vcd_catalog_item" && rs.Primary.Attributes["name"] != TestAccVcdCatalogItem {
			continue
		}

		org, _, err := conn.GetOrgAndVdc(testConfig.VCD.Org, testConfig.VCD.Vdc)
		if err != nil {
			return fmt.Errorf(errorRetrievingOrg, testConfig.VCD.Org+" and error: "+err.Error())
		}

		catalog, err := org.FindCatalog(testSuiteCatalogName)
		if err != nil {
			return fmt.Errorf("catalog query %s ended with error: %#v", rs.Primary.ID, err)
		}

		itemName := rs.Primary.Attributes["name"]
		catalogItem, err := catalog.FindCatalogItem(itemName)

		if catalogItem != (govcd.CatalogItem{}) {
			return fmt.Errorf("catalog item %s still exists", itemName)
		}
		if err != nil {
			return fmt.Errorf("catalog item %s still exists or other error: %#v", itemName, err)
		}

	}

	return nil
}

const testAccCheckVcdCatalogItemBasic = `
  resource "vcd_catalog_item" "{{.CatalogItemName}}" {
  org     = "{{.Org}}"
  catalog = "{{.Catalog}}"

  name                 = "{{.CatalogItemName}}"
  description          = "{{.Description}}"
  ova_path             = "{{.OvaPath}}"
  upload_piece_size    = {{.UploadPieceSize}}
  show_upload_progress = "{{.UploadProgress}}"

  metadata = {
    catalogItem_metadata = "catalogItem Metadata"
    catalogItem_metadata2 = "catalogItem Metadata2"
  }
}
`

const testAccCheckVcdCatalogItemUpdate = `
  resource "vcd_catalog_item" "{{.CatalogItemName}}" {
  org     = "{{.Org}}"
  catalog = "{{.Catalog}}"

  name                 = "{{.CatalogItemName}}"
  description          = "{{.Description}}"
  ova_path             = "{{.OvaPath}}"
  upload_piece_size    = {{.UploadPieceSize}}
  show_upload_progress = "{{.UploadProgress}}"

  metadata = {
    catalogItem_metadata = "catalogItem Metadata v2"
    catalogItem_metadata2 = "catalogItem Metadata2 v2"
    catalogItem_metadata3 = "catalogItem Metadata3"
  }
}
`
