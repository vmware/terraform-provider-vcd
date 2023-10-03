//go:build catalog || ALL || functional

package vcd

import (
	"fmt"
	"os"
	"regexp"
	"runtime"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// Test catalog and catalog media data sources
// Using a catalog data source we reference a catalog media data source
// Using a catalog media data source we create another catalog media
// where the description is the first data source ID
func TestAccVcdCatalogAndMediaDatasource(t *testing.T) {
	preTestChecks(t)
	var TestCatalogMediaDS = "TestCatalogMediaDS"
	var TestAccVcdDataSourceMedia = "TestAccVcdCatalogMediaBasic"
	var TestAccVcdDataSourceMediaDescription = "TestAccVcdCatalogMediaBasicDescription"

	_, sourceFile, _, _ := runtime.Caller(0)
	if !fileExists(sourceFile) {
		t.Skip("source file for this test was not found")
	}
	tempFile := "source_file.txt"
	var params = StringMap{
		"Org":              testConfig.VCD.Org,
		"Catalog":          testConfig.VCD.Catalog.Name,
		"NewCatalogMedia":  TestCatalogMediaDS,
		"OvaPath":          testConfig.Ova.OvaPath,
		"UploadPieceSize":  testConfig.Ova.UploadPieceSize,
		"UploadProgress":   testConfig.Ova.UploadProgress,
		"Tags":             "catalog",
		"CatalogMediaName": TestAccVcdDataSourceMedia,
		"MediaFileName":    "TestMediaFile",
		"MediaFilePath":    sourceFile,
		"DownloadToFile":   tempFile,
		"Description":      TestAccVcdDataSourceMediaDescription,
		"MediaPath":        testConfig.Media.MediaPath,
		"FuncName":         t.Name(),
	}
	testParamsNotEmpty(t, params)

	configText := templateFill(testAccCheckVcdCatalogMediaDS, params)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

	defer func() {
		if fileExists(tempFile) {
			err := os.Remove(tempFile)
			if err != nil {
				fmt.Printf("error deleting file '%s': %s", tempFile, err)
			}
		}
	}()
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { preRunChecks(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      catalogMediaDestroyed(testConfig.VCD.Catalog.Name, TestCatalogMediaDS),
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdCatalogMediaExists("vcd_catalog_media."+TestAccVcdDataSourceMedia),
					resource.TestMatchOutput("owner_name", regexp.MustCompile(`^\S+`)),
					resource.TestMatchOutput("creation_date", regexp.MustCompile(`^\d{4}-\d{2}-\d{2}.*`)),
					resource.TestCheckOutput("status", "RESOLVED"),
					resource.TestMatchOutput("storage_profile_name", regexp.MustCompile(`^\S+`)),
					resource.TestCheckResourceAttr("vcd_catalog_media.media_file", "name", "TestMediaFile"),
					testCheckMediaNonStringOutputs(),
				),
			},
			{
				Config: configText,
				Check:  checkFileContentsAreEqual(sourceFile, tempFile),
			},
		},
	})
	postTestChecks(t)
}

func checkFileContentsAreEqual(fileName1, fileName2 string) resource.TestCheckFunc {

	return func(s *terraform.State) error {
		contents1, err := os.ReadFile(fileName1)
		if err != nil {
			return err
		}
		contents2, err := os.ReadFile(fileName2)
		if err != nil {
			return err
		}
		if string(contents1) == string(contents2) {
			return nil
		}
		return fmt.Errorf("file %s and %s have different content", fileName1, fileName2)
	}
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

data "vcd_catalog" "mycat" {
  org  = "{{.Org}}"
  name = "{{.Catalog}}"
}

resource "vcd_catalog_media"  "{{.CatalogMediaName}}" {
  org        = "{{.Org}}"
  catalog_id = data.vcd_catalog.mycat.id

  name                 = "{{.CatalogMediaName}}"
  description          = "{{.Description}}"
  media_path           = "{{.MediaPath}}"
  upload_piece_size    = {{.UploadPieceSize}}
  show_upload_progress = {{.UploadProgress}}

  metadata = {
    catalogMedia_metadata = "catalogMedia Metadata"
    catalogMedia_metadata2 = "catalogMedia Metadata2"
  }
}

# this resource uploads the source file for the current test as a media item
resource "vcd_catalog_media"  "media_file" {
  org        = "{{.Org}}"
  catalog_id = data.vcd_catalog.mycat.id

  name                 = "{{.MediaFileName}}"
  description          = "{{.Description}}"
  media_path           = "{{.MediaFilePath}}"
  upload_piece_size    = {{.UploadPieceSize}}
  show_upload_progress = {{.UploadProgress}}
  upload_any_file      = true
}

data "vcd_catalog_media" "{{.NewCatalogMedia}}" {
  org        = "{{.Org}}"
  catalog_id = data.vcd_catalog.mycat.id
  name       = vcd_catalog_media.{{.CatalogMediaName}}.name
  depends_on = [vcd_catalog_media.{{.CatalogMediaName}}]
}

# This data source downloads the contents of the media item to a local file
data "vcd_catalog_media" "media_file_ds" {
  org        = "{{.Org}}"
  catalog_id = data.vcd_catalog.mycat.id
  name       = vcd_catalog_media.media_file.name

  download_to_file = "{{.DownloadToFile}}"
}

output "size" {
  value = data.vcd_catalog_media.{{.NewCatalogMedia}}.size
}
output "creation_date" {
  value = data.vcd_catalog_media.{{.NewCatalogMedia}}.creation_date
}
output "is_iso" {
  value = data.vcd_catalog_media.{{.NewCatalogMedia}}.is_iso
}
output "owner_name" {
  value = data.vcd_catalog_media.{{.NewCatalogMedia}}.owner_name
}
output "is_published" {
  value = data.vcd_catalog_media.{{.NewCatalogMedia}}.is_published
}
output "status" {
  value = data.vcd_catalog_media.{{.NewCatalogMedia}}.status
}
output "storage_profile_name" {
  value = data.vcd_catalog_media.{{.NewCatalogMedia}}.storage_profile_name
}
`
