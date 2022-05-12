//go:build catalog || ALL || functional
// +build catalog ALL functional

package vcd

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var TestAccVcdCatalogItem = "TestAccVcdCatalogItemBasic"
var TestAccVcdCatalogItemDescription = "TestAccVcdCatalogItemBasicDescription"
var TestAccVcdCatalogItemFromUrl = "TestAccVcdCatalogItemBasicFromUrl"
var TestAccVcdCatalogItemDescriptionFromUrl = "TestAccVcdCatalogItemBasicDescriptionFromUrl"
var TestAccVcdCatalogItemFromUrlUpdated = "TestAccVcdCatalogItemBasicFromUrlUpdated"
var TestAccVcdCatalogItemDescriptionFromUrlUpdated = "TestAccVcdCatalogItemBasicDescriptionFromUrlUpdated"

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

	var skipOnVcd1020 bool
	vcdClient := createTemporaryVCDConnection(false)
	if vcdClient.Client.APIVCDMaxVersionIs("< 35.2") {
		skipOnVcd1020 = true
	}

	configText := templateFill(testAccCheckVcdCatalogItemBasic, params)
	params["FuncName"] = t.Name() + "-Update"
	updateConfigText := templateFill(testAccCheckVcdCatalogItemUpdate, params)

	var fromUrlConfigText string
	var fromUrlConfigTextUpdate string

	// Conditionally skipping `templateFill` for 10.2.0 to avoid creating failing binary tests
	if !skipOnVcd1020 {
		params["FuncName"] = t.Name() + "-FromUrl"
		fromUrlConfigText = templateFill(testAccCheckVcdCatalogItemFromUrl, params)
		params["FuncName"] = t.Name() + "-FromUrlUpdate"
		fromUrlConfigTextUpdate = templateFill(testAccCheckVcdCatalogItemFromUrlUpdated, params)
	}

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
				SkipFunc: func() (bool, error) { return skipOnVcd1020, nil },
				Config:   fromUrlConfigText,
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
				SkipFunc: func() (bool, error) { return skipOnVcd1020, nil },
				Config:   fromUrlConfigTextUpdate,
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

		org, _, err := conn.GetOrgAndVdc(testConfig.VCD.Org, testConfig.VCD.Vdc)
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
    vapp_template_metadata = "vApp Template Metadata"
    vapp_template_metadata2 = "vApp Template Metadata2"
  }

  catalog_item_metadata = {
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

const testAccCheckVcdCatalogItemFromUrl = `
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

const testAccCheckVcdCatalogItemFromUrlUpdated = `
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
