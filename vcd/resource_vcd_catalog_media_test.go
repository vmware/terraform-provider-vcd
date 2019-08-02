// +build catalog ALL functional

package vcd

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/vmware/go-vcloud-director/v2/govcd"
)

var TestAccVcdCatalogMedia = "TestAccVcdCatalogMediaBasic"
var TestAccVcdCatalogMediaDescription = "TestAccVcdCatalogMediaBasicDescription"

func TestAccVcdCatalogMediaBasic(t *testing.T) {

	var catalogItem govcd.CatalogItem
	var params = StringMap{
		"Org":              testConfig.VCD.Org,
		"Catalog":          testSuiteCatalogName,
		"CatalogMediaName": TestAccVcdCatalogMedia,
		"Description":      TestAccVcdCatalogMediaDescription,
		"MediaPath":        testConfig.Media.MediaPath,
		"UploadPieceSize":  testConfig.Media.UploadPieceSize,
		"UploadProgress":   testConfig.Media.UploadProgress,
		"Tags":             "catalog",
	}

	configText := templateFill(testAccCheckVcdCatalogMediaBasic, params)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCatalogMediaDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdCatalogMediaExists("vcd_catalog_media."+TestAccVcdCatalogMedia, &catalogItem),
					resource.TestCheckResourceAttr(
						"vcd_catalog_media."+TestAccVcdCatalogMedia, "name", TestAccVcdCatalogMedia),
					resource.TestCheckResourceAttr(
						"vcd_catalog_media."+TestAccVcdCatalogMedia, "description", TestAccVcdCatalogMediaDescription),
					resource.TestCheckResourceAttr(
						"vcd_catalog_media."+TestAccVcdCatalogMedia, "metadata.mediaItem_metadata", "mediaItem Metadata"),
					resource.TestCheckResourceAttr(
						"vcd_catalog_media."+TestAccVcdCatalogMedia, "metadata.mediaItem_metadata2", "mediaItem Metadata2"),
				),
			},
		},
	})
}

func testAccCheckVcdCatalogMediaExists(mediaName string, catalogItem *govcd.CatalogItem) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		catalogMediaRs, ok := s.RootModule().Resources[mediaName]
		if !ok {
			return fmt.Errorf("not found: %s", mediaName)
		}

		if catalogMediaRs.Primary.ID == "" {
			return fmt.Errorf("catalog media ID is not set")
		}

		conn := testAccProvider.Meta().(*VCDClient)

		adminOrg, err := conn.GetAdminOrg(testConfig.VCD.Org)
		if err != nil {
			return fmt.Errorf(errorRetrievingOrg, testConfig.VCD.Org+" and error: "+err.Error())
		}

		catalog, err := adminOrg.FindCatalog(testSuiteCatalogName)
		if err != nil {
			return fmt.Errorf("catalog %s does not exist (%#v)", testSuiteCatalogName, catalog.Catalog)
		}

		newCatalogItem, err := catalog.FindCatalogItem(catalogMediaRs.Primary.Attributes["name"])
		if err != nil {
			return fmt.Errorf("catalog media %s does not exist (%#v)", catalogMediaRs.Primary.ID, catalogItem.CatalogItem)
		}

		catalogItem = &newCatalogItem
		return nil
	}
}

func testAccCheckCatalogMediaDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*VCDClient)
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "vcd_catalog_media" && rs.Primary.Attributes["name"] != TestAccVcdCatalogMedia {
			continue
		}

		adminOrg, err := conn.GetAdminOrg(testConfig.VCD.Org)
		if err != nil {
			return fmt.Errorf(errorRetrievingOrg, testConfig.VCD.Org+" and error: "+err.Error())
		}

		catalog, err := adminOrg.FindCatalog(testSuiteCatalogName)
		if err != nil {
			return fmt.Errorf("catalog query %s ended with error: %#v", rs.Primary.ID, err)
		}

		mediaName := rs.Primary.Attributes["name"]
		catalogItem, err := catalog.FindCatalogItem(mediaName)

		if catalogItem != (govcd.CatalogItem{}) {
			return fmt.Errorf("catalog media %s still exists", mediaName)
		}
		if err != nil {
			return fmt.Errorf("catalog media %s still exists or other error: %#v", mediaName, err)
		}

	}

	return nil
}

const testAccCheckVcdCatalogMediaBasic = `
  resource "vcd_catalog_media"  "{{.CatalogMediaName}}" {
  org     = "{{.Org}}"
  catalog = "{{.Catalog}}"

  name                 = "{{.CatalogMediaName}}"
  description          = "{{.Description}}"
  media_path           = "{{.MediaPath}}"
  upload_piece_size    = {{.UploadPieceSize}}
  show_upload_progress = "{{.UploadProgress}}"

  metadata = {
    mediaItem_metadata = "mediaItem Metadata"
    mediaItem_metadata2 = "mediaItem Metadata2"
  }
}
`
