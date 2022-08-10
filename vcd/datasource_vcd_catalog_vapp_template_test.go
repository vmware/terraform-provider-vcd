//go:build catalog || ALL || functional
// +build catalog ALL functional

package vcd

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// Test catalog and vApp Template data sources
// Using a catalog data source we reference a vApp Template data source
// Using a vApp Template data source we create another vApp Template
// where the description is the first data source ID
// It also includes deprecated catalog item data source and resource testing.
func TestAccVcdCatalogAndVappTemplateDatasource(t *testing.T) {
	preTestChecks(t)
	var TestCatalogVappTemplateDS = "TestCatalogVappTemplateDS"

	var params = StringMap{
		"Org":             testConfig.VCD.Org,
		"Catalog":         testSuiteCatalogName,
		"VAppTemplate":    testSuiteCatalogOVAItem,
		"NewVappTemplate": TestCatalogVappTemplateDS,
		"OvaPath":         testConfig.Ova.OvaPath,
		"UploadPieceSize": testConfig.Ova.UploadPieceSize,
		"UploadProgress":  testConfig.Ova.UploadProgress,
		"Tags":            "catalog",
	}
	testParamsNotEmpty(t, params)

	configText := templateFill(testAccCheckVcdCatalogItemDS, params)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

	datasourceCatalog := "data.vcd_catalog." + testSuiteCatalogName
	datasourceCatalogVappTemplate := "data.vcd_catalog_vapp_template." + testSuiteCatalogOVAItem
	resourceCatalogVappTemplate := "vcd_catalog_vapp_template." + TestCatalogVappTemplateDS
	datasourceCatalogItem := "data.vcd_catalog_item." + testSuiteCatalogOVAItem // Deprecated
	resourceCatalogItem := "vcd_catalog_item." + TestCatalogVappTemplateDS      // Deprecated

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { preRunChecks(t, params) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      catalogItemDestroyed(testSuiteCatalogName, TestCatalogVappTemplateDS),
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdCatalogItemExists("vcd_catalog_vapp_template."+TestCatalogVappTemplateDS),
					resource.TestCheckResourceAttr(resourceCatalogVappTemplate, "name", TestCatalogVappTemplateDS),
					resource.TestCheckResourceAttrPair(datasourceCatalog, "name", resourceCatalogItem, "catalog"),

					testAccCheckVcdCatalogItemExists("vcd_catalog_item."+TestCatalogVappTemplateDS),        // Deprecated
					resource.TestCheckResourceAttr(resourceCatalogItem, "name", TestCatalogVappTemplateDS), // Deprecated

					// The description of the new catalog item was created using
					// the ID of the catalog item data source
					resource.TestCheckResourceAttrPair(datasourceCatalogVappTemplate, "id", resourceCatalogVappTemplate, "description"),
					resource.TestCheckResourceAttrPair(datasourceCatalogItem, "id", resourceCatalogItem, "description"), // Deprecated

					resource.TestCheckResourceAttr(resourceCatalogItem, "metadata.key1", "value1"),
					resource.TestCheckResourceAttr(resourceCatalogItem, "metadata.key2", "value2"),
					resource.TestCheckResourceAttr(resourceCatalogItem, "metadata.catalogItem_metadata", "catalogItem Metadata"),   // Deprecated
					resource.TestCheckResourceAttr(resourceCatalogItem, "metadata.catalogItem_metadata2", "catalogItem Metadata2"), // Deprecated
				),
			},
			{ // Deprecated
				ResourceName:      "vcd_catalog_item." + TestCatalogVappTemplateDS,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdOrgCatalogObject(TestCatalogVappTemplateDS),
				// These fields can't be retrieved from catalog item data
				ImportStateVerifyIgnore: []string{"ova_path", "upload_piece_size", "show_upload_progress"},
			},
			{
				ResourceName:      "vcd_catalog_vapp_template." + TestCatalogVappTemplateDS,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdOrgCatalogObject(TestCatalogVappTemplateDS),
				// These fields can't be retrieved from catalog item data
				ImportStateVerifyIgnore: []string{"ova_path", "upload_piece_size", "show_upload_progress"},
			},
		},
	})
	postTestChecks(t)
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

const testAccCheckVcdCatalogItemDS = `
data "vcd_catalog" "{{.Catalog}}" {
  org  = "{{.Org}}"
  name = "{{.Catalog}}"
}

# Deprecated datasource. It's here to avoid regressions
data "vcd_catalog_item" "{{.VAppTemplate}}" {
  org     = "{{.Org}}"
  catalog = data.vcd_catalog.{{.Catalog}}.name
  name    = "{{.VAppTemplate}}"
}

data "vcd_catalog_vapp_template" "{{.VAppTemplate}}" {
  org     = "{{.Org}}"
  catalog = data.vcd_catalog.{{.Catalog}}.name
  name    = "{{.VAppTemplate}}"
}

# Deprecated datasource. It's here to avoid regressions
resource "vcd_catalog_item" "{{.NewVappTemplate}}" {
  org     = "{{.Org}}"
  catalog = data.vcd_catalog.{{.Catalog}}.name

  name                 = "{{.NewVappTemplate}}"
  description          = data.vcd_catalog_item.{{.VAppTemplate}}.id
  ova_path             = "{{.OvaPath}}"
  upload_piece_size    = {{.UploadPieceSize}}
  show_upload_progress = "{{.UploadProgress}}"

  metadata = {
    catalogItem_metadata = "catalogItem Metadata"
    catalogItem_metadata2 = "catalogItem Metadata2"
  }
}

resource "vcd_catalog_vapp_template" "{{.NewVappTemplate}}" {
  org     = "{{.Org}}"
  catalog = data.vcd_catalog.{{.Catalog}}.name

  name                 = "{{.NewVappTemplate}}"
  description          = data.vcd_catalog_item.{{.VAppTemplate}}.id
  ova_path             = "{{.OvaPath}}"
  upload_piece_size    = {{.UploadPieceSize}}
  show_upload_progress = "{{.UploadProgress}}"

  metadata = {
    key1 = "value1"
    key2 = "value2"
  }
}
`
