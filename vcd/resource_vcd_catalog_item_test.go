//go:build catalog || ALL || functional

package vcd

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// Deprecated
var TestAccVcdCatalogItem = "TestAccVcdCatalogItemBasic"

// Deprecated
var TestAccVcdCatalogItemDescription = "TestAccVcdCatalogItemBasicDescription"

// Deprecated
var TestAccVcdCatalogItemFromUrl = "TestAccVcdCatalogItemBasicFromUrl"

// Deprecated
var TestAccVcdCatalogItemDescriptionFromUrl = "TestAccVcdCatalogItemBasicDescriptionFromUrl"

// Deprecated
var TestAccVcdCatalogItemFromUrlUpdated = "TestAccVcdCatalogItemBasicFromUrlUpdated"

// Deprecated
var TestAccVcdCatalogItemDescriptionFromUrlUpdated = "TestAccVcdCatalogItemBasicDescriptionFromUrlUpdated"

// Deprecated
func TestAccVcdCatalogItemBasic(t *testing.T) {
	preTestChecks(t)

	if testConfig.Ova.OvfUrl == "" {
		t.Skip("Variables Ova.OvfUrl must be set")
	}

	var params = StringMap{
		"Org":                           testConfig.VCD.Org,
		"Catalog":                       testSuiteCatalogName,
		"CatalogItemName":               TestAccVcdCatalogItem,
		"CatalogItemNameFromUrl":        TestAccVcdCatalogItemFromUrl,
		"CatalogItemNameFromUrlUpdated": TestAccVcdCatalogItemFromUrlUpdated,
		"Description":                   TestAccVcdCatalogItemDescription,
		"DescriptionFromUrl":            TestAccVcdCatalogItemDescriptionFromUrl,
		"DescriptionFromUrlUpdated":     TestAccVcdCatalogItemDescriptionFromUrlUpdated,
		"OvaPath":                       testConfig.Ova.OvaPath,
		"OvfUrl":                        testConfig.Ova.OvfUrl,
		"UploadPieceSize":               testConfig.Ova.UploadPieceSize,
		"UploadProgress":                testConfig.Ova.UploadProgress,
		"UploadProgressFromUrl":         testConfig.Ova.UploadProgress,
		"Tags":                          "catalog",
	}

	configText := templateFill(testAccCheckVcdCatalogItemBasic, params)
	params["FuncName"] = t.Name() + "-Update"
	updateConfigText := templateFill(testAccCheckVcdCatalogItemUpdate, params)

	params["FuncName"] = t.Name() + "-FromUrl"
	fromUrlConfigText := templateFill(testAccCheckVcdCatalogItemFromUrl, params)

	params["FuncName"] = t.Name() + "-FromUrlUpdate"
	fromUrlConfigTextUpdate := templateFill(testAccCheckVcdCatalogItemFromUrlUpdated, params)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

	resourceCatalogItem := "vcd_catalog_item." + TestAccVcdCatalogItem
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { preRunChecks(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckCatalogItemDestroy,
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdCatalogItemExists("vcd_catalog_item."+TestAccVcdCatalogItem),
					resource.TestCheckResourceAttr(
						resourceCatalogItem, "name", TestAccVcdCatalogItem),
					resource.TestCheckResourceAttr(
						resourceCatalogItem, "description", TestAccVcdCatalogItemDescription),
					resource.TestCheckResourceAttr(
						resourceCatalogItem, "metadata.vapp_template_metadata", "vApp Template Metadata"),
					resource.TestCheckResourceAttr(
						resourceCatalogItem, "metadata.vapp_template_metadata2", "vApp Template Metadata2"),
					resource.TestCheckResourceAttr(
						resourceCatalogItem, "catalog_item_metadata.catalogItem_metadata", "catalogItem Metadata"),
					resource.TestCheckResourceAttr(
						resourceCatalogItem, "catalog_item_metadata.catalogItem_metadata2", "catalogItem Metadata2"),
				),
			},
			{
				Config: updateConfigText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdCatalogItemExists("vcd_catalog_item."+TestAccVcdCatalogItem),
					resource.TestCheckResourceAttr(
						"vcd_catalog_item."+TestAccVcdCatalogItem, "name", TestAccVcdCatalogItem),
					resource.TestCheckResourceAttr(
						"vcd_catalog_item."+TestAccVcdCatalogItem, "description", TestAccVcdCatalogItemDescription),
					resource.TestCheckResourceAttr(
						"vcd_catalog_item."+TestAccVcdCatalogItem, "metadata.vapp_template_metadata", "vApp Template Metadata v2"),
					resource.TestCheckResourceAttr(
						"vcd_catalog_item."+TestAccVcdCatalogItem, "metadata.vapp_template_metadata2", "vApp Template Metadata2 v2"),
					resource.TestCheckResourceAttr(
						"vcd_catalog_item."+TestAccVcdCatalogItem, "metadata.vapp_template_metadata3", "vApp Template Metadata3"),
					resource.TestCheckResourceAttr(
						"vcd_catalog_item."+TestAccVcdCatalogItem, "catalog_item_metadata.catalogItem_metadata", "catalogItem Metadata v2"),
					resource.TestCheckResourceAttr(
						"vcd_catalog_item."+TestAccVcdCatalogItem, "catalog_item_metadata.catalogItem_metadata2", "catalogItem Metadata2 v2"),
					resource.TestCheckResourceAttr(
						"vcd_catalog_item."+TestAccVcdCatalogItem, "catalog_item_metadata.catalogItem_metadata3", "catalogItem Metadata3"),
				),
			},
			{
				Config: fromUrlConfigText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdCatalogItemExists("vcd_catalog_item."+TestAccVcdCatalogItemFromUrl),
					resource.TestCheckResourceAttr(
						"vcd_catalog_item."+TestAccVcdCatalogItemFromUrl, "name", TestAccVcdCatalogItemFromUrl),
					resource.TestCheckResourceAttr(
						"vcd_catalog_item."+TestAccVcdCatalogItemFromUrl, "description", TestAccVcdCatalogItemDescriptionFromUrl),
					resource.TestCheckResourceAttr(
						"vcd_catalog_item."+TestAccVcdCatalogItemFromUrl, "metadata.vapp_template_metadata", "vApp Template Metadata"),
					resource.TestCheckResourceAttr(
						"vcd_catalog_item."+TestAccVcdCatalogItemFromUrl, "metadata.vapp_template_metadata2", "vApp Template Metadata2"),
					resource.TestCheckResourceAttr(
						"vcd_catalog_item."+TestAccVcdCatalogItemFromUrl, "metadata.vapp_template_metadata3", "vApp Template Metadata3"),
					resource.TestCheckResourceAttr(
						"vcd_catalog_item."+TestAccVcdCatalogItemFromUrl, "catalog_item_metadata.catalogItem_metadata", "catalogItem Metadata"),
					resource.TestCheckResourceAttr(
						"vcd_catalog_item."+TestAccVcdCatalogItemFromUrl, "catalog_item_metadata.catalogItem_metadata2", "catalogItem Metadata2"),
					resource.TestCheckResourceAttr(
						"vcd_catalog_item."+TestAccVcdCatalogItemFromUrl, "catalog_item_metadata.catalogItem_metadata3", "catalogItem Metadata3"),
				),
			},
			{
				Config: fromUrlConfigTextUpdate,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdCatalogItemExists("vcd_catalog_item."+TestAccVcdCatalogItemFromUrl),
					resource.TestCheckResourceAttr(
						"vcd_catalog_item."+TestAccVcdCatalogItemFromUrl, "name", TestAccVcdCatalogItemFromUrlUpdated),
					resource.TestCheckResourceAttr(
						"vcd_catalog_item."+TestAccVcdCatalogItemFromUrl, "description", TestAccVcdCatalogItemDescriptionFromUrlUpdated),
					resource.TestCheckResourceAttr(
						"vcd_catalog_item."+TestAccVcdCatalogItemFromUrl, "metadata.vapp_template_metadata", "vApp Template Metadata"),
					resource.TestCheckResourceAttr(
						"vcd_catalog_item."+TestAccVcdCatalogItemFromUrl, "metadata.vapp_template_metadata2", "vApp Template Metadata2_2"),
					resource.TestCheckResourceAttr(
						"vcd_catalog_item."+TestAccVcdCatalogItemFromUrl, "catalog_item_metadata.catalogItem_metadata", "catalogItem Metadata"),
					resource.TestCheckResourceAttr(
						"vcd_catalog_item."+TestAccVcdCatalogItemFromUrl, "catalog_item_metadata.catalogItem_metadata2", "catalogItem Metadata2_2"),
				),
			},
		},
	})
	postTestChecks(t)
}

// Deprecated
func testAccCheckVcdCatalogItemExists(itemName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		catalogItemRs, ok := s.RootModule().Resources[itemName]
		if !ok {
			return fmt.Errorf("not found: %s", itemName)
		}

		if catalogItemRs.Primary.ID == "" {
			return fmt.Errorf("no catalog item ID is set")
		}

		conn := testAccProvider.Meta().(*VCDClient)

		org, err := conn.GetOrgByName(testConfig.VCD.Org)
		if err != nil {
			return fmt.Errorf(errorRetrievingOrg, testConfig.VCD.Org+" and error: "+err.Error())
		}

		catalog, err := org.GetCatalogByName(testSuiteCatalogName, false)
		if err != nil {
			return fmt.Errorf("catalog %s does not exist: %s", testSuiteCatalogName, err)
		}

		_, err = catalog.GetCatalogItemByName(catalogItemRs.Primary.Attributes["name"], false)
		if err != nil {
			return fmt.Errorf("catalog item %s does not exist (%s)", catalogItemRs.Primary.ID, err)
		}

		return nil
	}
}

// Deprecated
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

		catalog, err := org.GetCatalogByName(testSuiteCatalogName, false)
		if err != nil {
			return fmt.Errorf("catalog query %s ended with error: %s", rs.Primary.ID, err)
		}

		itemName := rs.Primary.Attributes["name"]
		_, err = catalog.GetCatalogItemByName(itemName, false)

		if err == nil {
			return fmt.Errorf("catalog item %s still exists", itemName)
		}
	}

	return nil
}

// Deprecated
const testAccCheckVcdCatalogItemBasic = `
  # Deprecated
  resource "vcd_catalog_item" "{{.CatalogItemName}}" {
  org     = "{{.Org}}"
  catalog = "{{.Catalog}}"

  name                 = "{{.CatalogItemName}}"
  description          = "{{.Description}}"
  ova_path             = "{{.OvaPath}}"
  upload_piece_size    = {{.UploadPieceSize}}
  show_upload_progress = "{{.UploadProgress}}"

  metadata = {
    vapp_template_metadata = "vApp Template Metadata"
    vapp_template_metadata2 = "vApp Template Metadata2"
  }

  catalog_item_metadata = {
    catalogItem_metadata = "catalogItem Metadata"
    catalogItem_metadata2 = "catalogItem Metadata2"
  }
}
`

// Deprecated
const testAccCheckVcdCatalogItemUpdate = `
  # Deprecated
  resource "vcd_catalog_item" "{{.CatalogItemName}}" {
  org     = "{{.Org}}"
  catalog = "{{.Catalog}}"

  name                 = "{{.CatalogItemName}}"
  description          = "{{.Description}}"
  ova_path             = "{{.OvaPath}}"
  upload_piece_size    = {{.UploadPieceSize}}
  show_upload_progress = "{{.UploadProgress}}"

  metadata = {
    vapp_template_metadata = "vApp Template Metadata v2"
    vapp_template_metadata2 = "vApp Template Metadata2 v2"
    vapp_template_metadata3 = "vApp Template Metadata3"
  }

  catalog_item_metadata = {
    catalogItem_metadata = "catalogItem Metadata v2"
    catalogItem_metadata2 = "catalogItem Metadata2 v2"
    catalogItem_metadata3 = "catalogItem Metadata3"
  }
}
`

// Deprecated
const testAccCheckVcdCatalogItemFromUrl = `
  # Deprecated
  resource "vcd_catalog_item" "{{.CatalogItemNameFromUrl}}" {
  org     = "{{.Org}}"
  catalog = "{{.Catalog}}"

  name                 = "{{.CatalogItemNameFromUrl}}"
  description          = "{{.DescriptionFromUrl}}"
  ovf_url              = "{{.OvfUrl}}"
  show_upload_progress = "{{.UploadProgressFromUrl}}"

  metadata = {
    vapp_template_metadata = "vApp Template Metadata"
    vapp_template_metadata2 = "vApp Template Metadata2"
    vapp_template_metadata3 = "vApp Template Metadata3"
  }

  catalog_item_metadata = {
    catalogItem_metadata = "catalogItem Metadata"
    catalogItem_metadata2 = "catalogItem Metadata2"
    catalogItem_metadata3 = "catalogItem Metadata3"
  }
}
`

// Deprecated
const testAccCheckVcdCatalogItemFromUrlUpdated = `
  # Deprecated
  resource "vcd_catalog_item" "{{.CatalogItemNameFromUrl}}" {
  org     = "{{.Org}}"
  catalog = "{{.Catalog}}"

  name                 = "{{.CatalogItemNameFromUrlUpdated}}"
  description          = "{{.DescriptionFromUrlUpdated}}"
  ovf_url              = "{{.OvfUrl}}"
  show_upload_progress = "{{.UploadProgressFromUrl}}"

  metadata = {
    vapp_template_metadata = "vApp Template Metadata"
    vapp_template_metadata2 = "vApp Template Metadata2_2"
  }

  catalog_item_metadata = {
    catalogItem_metadata = "catalogItem Metadata"
    catalogItem_metadata2 = "catalogItem Metadata2_2"
  }
}
`

// TestAccVcdCatalogItemMetadata tests metadata CRUD on catalog items
func TestAccVcdCatalogItemMetadata(t *testing.T) {
	testMetadataEntryCRUD(t,
		testAccCheckVcdCatalogItemMetadata, "vcd_catalog_item.test-catalog-item",
		testAccCheckVcdCatalogItemMetadataDatasource, "data.vcd_catalog_item.test-catalog-item-ds",
		StringMap{
			"Catalog": testConfig.VCD.Catalog.NsxtBackedCatalogName,
			"OvfUrl":  testConfig.Ova.OvfUrl,
		})
}

const testAccCheckVcdCatalogItemMetadata = `
resource "vcd_catalog_item" "test-catalog-item" {
  org     = "{{.Org}}"
  catalog = "{{.Catalog}}"
  name    = "{{.Name}}"
  ovf_url = "{{.OvfUrl}}"
  {{.Metadata}}
}
`

const testAccCheckVcdCatalogItemMetadataDatasource = `
data "vcd_catalog_item" "test-catalog-item-ds" {
  org     = vcd_catalog_item.test-catalog-item.org
  catalog = vcd_catalog_item.test-catalog-item.catalog
  name    = vcd_catalog_item.test-catalog-item.name
}
`

func TestAccVcdCatalogItemMetadataIgnore(t *testing.T) {
	skipIfNotSysAdmin(t)

	getObjectById := func(vcdClient *VCDClient, id string) (metadataCompatible, error) {
		adminOrg, err := vcdClient.GetAdminOrgByName(testConfig.VCD.Org)
		if err != nil {
			return nil, fmt.Errorf("could not retrieve Org '%s': %s", testConfig.VCD.Org, err)
		}
		catalog, err := adminOrg.GetCatalogByName(testConfig.VCD.Catalog.NsxtBackedCatalogName, true)
		if err != nil {
			return nil, fmt.Errorf("could not retrieve Catalog '%s': %s", testConfig.VCD.Catalog.NsxtBackedCatalogName, err)
		}
		catalogItem, err := catalog.GetCatalogItemById(id, true)
		if err != nil {
			return nil, fmt.Errorf("could not retrieve Catalog Item '%s': %s", id, err)
		}
		return catalogItem, nil
	}

	testMetadataEntryIgnore(t,
		testAccCheckVcdCatalogItemMetadata, "vcd_catalog_item.test-catalog-item",
		testAccCheckVcdCatalogItemMetadataDatasource, "data.vcd_catalog_item.test-catalog-item-ds",
		getObjectById, StringMap{
			"Catalog": testConfig.VCD.Catalog.NsxtBackedCatalogName,
			"OvfUrl":  testConfig.Ova.OvfUrl,
		})
}
