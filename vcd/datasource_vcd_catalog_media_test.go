// +build catalog ALL functional

package vcd

import (
	"fmt"
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

	var params = StringMap{
		"Org":              testConfig.VCD.Org,
		"Catalog":          testConfig.VCD.Catalog.Name,
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
		CheckDestroy: catalogMediaDestroyed(testConfig.VCD.Catalog.Name, TestCatalogMediaDS),
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdCatalogMediaExists("vcd_catalog_media."+TestAccVcdDataSourceMedia),
					resource.TestMatchOutput("owner_name", regexp.MustCompile(`^\S+`)),
					resource.TestMatchOutput("creation_date", regexp.MustCompile(`^^\d{4}-\d{2}-\d{2}.*`)),
					resource.TestCheckOutput("status", "RESOLVED"),
					resource.TestMatchOutput("storage_profile_name", regexp.MustCompile(`^\S+`)),
					testCheckMediaNonStringOutputs(),
				),
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
		_, err = cat.GetMediaByName(mediaName, false)
		if err == nil {
			return fmt.Errorf("catalog media %s not deleted", mediaName)
		}
		return nil
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
  name    = "${vcd_catalog_media.{{.CatalogMediaName}}.name}"
}

output "size" {
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
