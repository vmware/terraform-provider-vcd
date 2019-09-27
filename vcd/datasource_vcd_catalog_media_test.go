// +build catalog ALL functional

package vcd

import (
	"fmt"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

// Test catalog and catalog media data sources
// Using a catalog data source we reference a catalog media data source
// Using a catalog media data source we create another catalog media
// where the description is the first data source ID
func TestAccVcdCatalogAndMediaDatasource(t *testing.T) {
	var TestCatalogMediaDS = "TestCatalogMediaDS"
	var TestAccVcdDataSourceMedia = "TestAccVcdCatalogMediaBasic"
	var TestAccVcdDataSourceMediaDescription = "TestAccVcdCatalogMediaBasicDescription"

	var catalogItem govcd.CatalogItem

	var params = StringMap{
		"Org":              testConfig.VCD.Org,
		"Catalog":          testSuiteCatalogName,
		"NewCatalogMedia":  TestCatalogMediaDS,
		"OvaPath":          testConfig.Ova.OvaPath,
		"UploadPieceSize":  testConfig.Ova.UploadPieceSize,
		"UploadProgress":   testConfig.Ova.UploadProgress,
		"Tags":             "catalog",
		"CatalogMediaName": TestAccVcdDataSourceMedia,
		"Description":      TestAccVcdDataSourceMediaDescription,
		"MediaPath":        testConfig.Media.MediaPath,
	}

	configText := templateFill(testAccCheckVcdCatalogMediaDS, params)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { preRunChecks(t) },
		Providers:    testAccProviders,
		CheckDestroy: catalogMediaDestroyed(testSuiteCatalogName, TestCatalogMediaDS),
		Steps: []resource.TestStep{
			resource.TestStep{
				Config:             configText,
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdCatalogMediaExists("vcd_catalog_media."+TestAccVcdDataSourceMedia, &catalogItem),
					resource.TestMatchOutput("media_size", regexp.MustCompile(`^\d*$`)),
					resource.TestCheckOutput("is_iso", "true"),
					resource.TestMatchOutput("owner_name", regexp.MustCompile(`^\w*$`)),
					resource.TestCheckOutput("is_published", "false"),
					resource.TestMatchOutput("creation_date", regexp.MustCompile(`^(2019|2020)-`)),
					resource.TestCheckOutput("status", "RESOLVED"),
					resource.TestMatchOutput("storage_profile_name", regexp.MustCompile(`(.|\s)*\S(.|\s)*`)),
				),
			},
			resource.TestStep{
				ResourceName:      "vcd_catalog_media." + TestAccVcdDataSourceMedia + "-import",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdByCatalogMedia(TestAccVcdDataSourceMedia),
				// These fields can't be retrieved from catalog media data
				ImportStateVerifyIgnore: []string{"media_path", "upload_piece_size", "show_upload_progress"},
			},
		},
	})
}

func catalogMediaDestroyed(catalog, mediaName string) resource.TestCheckFunc {
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
		_, err = cat.GetCatalogItemByName(mediaName, false)
		if err == nil {
			return fmt.Errorf("catalog media %s not deleted", mediaName)
		}
		return nil
	}
}

func importStateIdByCatalogMedia(objectName string) resource.ImportStateIdFunc {
	return func(*terraform.State) (string, error) {
		importId := testConfig.VCD.Org + "." + testSuiteCatalogName + "." + objectName
		if testConfig.VCD.Org == "" || testSuiteCatalogName == "" || objectName == "" {
			return "", fmt.Errorf("missing information to generate import path: %s", importId)
		}
		return importId, nil
	}
}

const testAccCheckVcdCatalogMediaDS = `
resource "vcd_catalog_media"  "{{.CatalogMediaName}}" {
  org     = "{{.Org}}"
  catalog = "{{.Catalog}}"

  name                 = "{{.CatalogMediaName}}"
  description          = "{{.Description}}"
  media_path           = "{{.MediaPath}}"
  upload_piece_size    = {{.UploadPieceSize}}
  show_upload_progress = "{{.UploadProgress}}"

  metadata = {
    catalogMedia_metadata = "catalogMedia Metadata"
    catalogMedia_metadata2 = "catalogMedia Metadata2"
  }
}

data "vcd_catalog_media" "{{.NewCatalogMedia}}" {
  org     = "{{.Org}}"
  catalog = "{{.Catalog}}"
  name    = "{{.CatalogMediaName}}"
  depends_on = ["vcd_catalog_media.{{.CatalogMediaName}}"]
}

output "media_size" {
  value = data.vcd_catalog_media.{{.NewCatalogMedia}}.size
  depends_on = ["data.vcd_catalog_media.{{.NewCatalogMedia}}"]
}
output "creation_date" {
  value = data.vcd_catalog_media.{{.NewCatalogMedia}}.creation_date
  depends_on = [data.vcd_catalog_media.{{.NewCatalogMedia}}]
}
output "is_iso" {
  value = data.vcd_catalog_media.{{.NewCatalogMedia}}.is_iso
  depends_on = [data.vcd_catalog_media.{{.NewCatalogMedia}}]
}
output "owner_name" {
  value = data.vcd_catalog_media.{{.NewCatalogMedia}}.owner_name
  depends_on = [data.vcd_catalog_media.{{.NewCatalogMedia}}]
}
output "is_published" {
  value = data.vcd_catalog_media.{{.NewCatalogMedia}}.is_published
  depends_on = [data.vcd_catalog_media.{{.NewCatalogMedia}}]
}
output "status" {
  value = data.vcd_catalog_media.{{.NewCatalogMedia}}.status
  depends_on = [data.vcd_catalog_media.{{.NewCatalogMedia}}]
}
output "storage_profile_name" {
  value = data.vcd_catalog_media.{{.NewCatalogMedia}}.storage_profile_name
  depends_on = [data.vcd_catalog_media.{{.NewCatalogMedia}}]
}
`
