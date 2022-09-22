//go:build catalog || ALL || functional
// +build catalog ALL functional

package vcd

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var TestAccVcdVAppTemplate = "TestAccVcdVAppTemplateBasic"
var TestAccVcdVAppTemplateDescription = "TestAccVcdVAppTemplateBasicDescription"
var TestAccVcdVAppTemplateFromUrl = "TestAccVcdVAppTemplateBasicFromUrl"
var TestAccVcdVAppTemplateFromUrlUpdated = "TestAccVcdVAppTemplateBasicFromUrlUpdated"
var TestAccVcdVAppTemplateDescriptionFromUrlUpdated = "TestAccVcdVAppTemplateBasicDescriptionFromUrlUpdated"

func TestAccVcdCatalogVAppTemplateBasic(t *testing.T) {
	preTestChecks(t)

	if testConfig.Ova.OvfUrl == "" {
		t.Skip("Variables Ova.OvfUrl must be set")
	}

	var params = StringMap{
		"Org":                            testConfig.VCD.Org,
		"Catalog":                        testSuiteCatalogName,
		"VAppTemplateName":               TestAccVcdVAppTemplate,
		"VAppTemplateNameFromUrl":        TestAccVcdVAppTemplateFromUrl,
		"VAppTemplateNameFromUrlUpdated": TestAccVcdVAppTemplateFromUrlUpdated,
		"Description":                    TestAccVcdVAppTemplateDescription,
		"OvaPath":                        testConfig.Ova.OvaPath,
		"OvfUrl":                         testConfig.Ova.OvfUrl,
		"UploadPieceSize":                testConfig.Ova.UploadPieceSize,
		"Tags":                           "catalog",
	}

	configText := templateFill(testAccCheckVcdVAppTemplateBasic, params)
	params["FuncName"] = t.Name() + "-Update"
	updateConfigText := templateFill(testAccCheckVcdVAppTemplateUpdate, params)

	params["FuncName"] = t.Name() + "-FromUrl"
	fromUrlConfigText := templateFill(testAccCheckVcdVAppTemplateFromUrl, params)

	params["FuncName"] = t.Name() + "-FromUrlUpdate"
	fromUrlConfigTextUpdate := templateFill(testAccCheckVcdVAppTemplateFromUrlUpdated, params)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

	resourceVAppTemplate := "vcd_catalog_vapp_template." + TestAccVcdVAppTemplate
	resourceVAppTemplateFromUrl := "vcd_catalog_vapp_template." + TestAccVcdVAppTemplateFromUrl
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { preRunChecks(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVAppTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdVAppTemplateExists(resourceVAppTemplate),
					resource.TestCheckResourceAttr(
						resourceVAppTemplate, "name", TestAccVcdVAppTemplate),
					resource.TestCheckResourceAttr(
						resourceVAppTemplate, "description", TestAccVcdVAppTemplateDescription),
					resource.TestCheckResourceAttr(
						resourceVAppTemplate, "metadata.vapp_template_metadata", "vApp Template Metadata"),
					resource.TestCheckResourceAttr(
						resourceVAppTemplate, "metadata.vapp_template_metadata2", "vApp Template Metadata2"),
				),
			},
			{
				Config: updateConfigText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdVAppTemplateExists(resourceVAppTemplate),
					resource.TestCheckResourceAttr(
						resourceVAppTemplate, "name", TestAccVcdVAppTemplate),
					resource.TestCheckResourceAttr(
						resourceVAppTemplate, "description", TestAccVcdVAppTemplateDescription),
					resource.TestCheckResourceAttr(
						resourceVAppTemplate, "metadata.vapp_template_metadata", "vApp Template Metadata v2"),
					resource.TestCheckResourceAttr(
						resourceVAppTemplate, "metadata.vapp_template_metadata2", "vApp Template Metadata2 v2"),
					resource.TestCheckResourceAttr(
						resourceVAppTemplate, "metadata.vapp_template_metadata3", "vApp Template Metadata3"),
				),
			},
			{
				Config: fromUrlConfigText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdVAppTemplateExists(resourceVAppTemplateFromUrl),
					resource.TestCheckResourceAttr(
						resourceVAppTemplateFromUrl, "name", TestAccVcdVAppTemplateFromUrl),
					// FIXME: Due to a bug in VCD, description is overridden by the present in the OVA
					resource.TestMatchResourceAttr(resourceVAppTemplateFromUrl, "description", regexp.MustCompile(`^Name: yVM.*`)),
					resource.TestCheckResourceAttr(
						resourceVAppTemplateFromUrl, "metadata.vapp_template_metadata", "vApp Template Metadata"),
					resource.TestCheckResourceAttr(
						resourceVAppTemplateFromUrl, "metadata.vapp_template_metadata2", "vApp Template Metadata2"),
					resource.TestCheckResourceAttr(
						resourceVAppTemplateFromUrl, "metadata.vapp_template_metadata3", "vApp Template Metadata3"),
				),
			},
			{
				Config: fromUrlConfigTextUpdate,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdVAppTemplateExists(resourceVAppTemplateFromUrl),
					resource.TestCheckResourceAttr(
						resourceVAppTemplateFromUrl, "name", TestAccVcdVAppTemplateFromUrlUpdated),
					// FIXME: Due to a bug in VCD, description is overridden by the present in the OVA
					resource.TestMatchResourceAttr(resourceVAppTemplateFromUrl, "description", regexp.MustCompile(`^Name: yVM.*`)),
					resource.TestCheckResourceAttr(
						resourceVAppTemplateFromUrl, "metadata.vapp_template_metadata", "vApp Template Metadata"),
					resource.TestCheckResourceAttr(
						resourceVAppTemplateFromUrl, "metadata.vapp_template_metadata2", "vApp Template Metadata2_2"),
				),
			},
		},
	})
	postTestChecks(t)
}

func preRunChecks(t *testing.T) {
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

func testAccCheckVcdVAppTemplateExists(itemName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		VAppTemplateRs, ok := s.RootModule().Resources[itemName]
		if !ok {
			return fmt.Errorf("not found: %s", itemName)
		}

		if VAppTemplateRs.Primary.ID == "" {
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

		_, err = catalog.GetVAppTemplateByName(VAppTemplateRs.Primary.Attributes["name"])
		if err != nil {
			return fmt.Errorf("vApp Template %s does not exist (%s)", VAppTemplateRs.Primary.ID, err)
		}

		return nil
	}
}

func testAccCheckVAppTemplateDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*VCDClient)
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "vcd_catalog_vapp_template" && rs.Primary.Attributes["name"] != TestAccVcdVAppTemplate {
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
		_, err = catalog.GetVAppTemplateByName(itemName)

		if err == nil {
			return fmt.Errorf("vApp Template %s still exists", itemName)
		}
	}

	return nil
}

const testAccCheckVcdVAppTemplateBasic = `
data "vcd_catalog" "{{.Catalog}}" {
  org  = "{{.Org}}"
  name = "{{.Catalog}}"
}

resource "vcd_catalog_vapp_template" "{{.VAppTemplateName}}" {
  org        = "{{.Org}}"
  catalog_id = data.vcd_catalog.{{.Catalog}}.id

  name                 = "{{.VAppTemplateName}}"
  description          = "{{.Description}}"
  ova_path             = "{{.OvaPath}}"
  upload_piece_size    = {{.UploadPieceSize}}

  metadata = {
    vapp_template_metadata = "vApp Template Metadata"
    vapp_template_metadata2 = "vApp Template Metadata2"
  }
}
`

const testAccCheckVcdVAppTemplateUpdate = `
data "vcd_catalog" "{{.Catalog}}" {
  org  = "{{.Org}}"
  name = "{{.Catalog}}"
}

resource "vcd_catalog_vapp_template" "{{.VAppTemplateName}}" {
  org        = "{{.Org}}"
  catalog_id = data.vcd_catalog.{{.Catalog}}.id

  name                 = "{{.VAppTemplateName}}"
  description          = "{{.Description}}"
  ova_path             = "{{.OvaPath}}"
  upload_piece_size    = {{.UploadPieceSize}}

  metadata = {
    vapp_template_metadata = "vApp Template Metadata v2"
    vapp_template_metadata2 = "vApp Template Metadata2 v2"
    vapp_template_metadata3 = "vApp Template Metadata3"
  }
}
`

const testAccCheckVcdVAppTemplateFromUrl = `
data "vcd_catalog" "{{.Catalog}}" {
  org  = "{{.Org}}"
  name = "{{.Catalog}}"
}

resource "vcd_catalog_vapp_template" "{{.VAppTemplateNameFromUrl}}" {
  org        = "{{.Org}}"
  catalog_id = data.vcd_catalog.{{.Catalog}}.id

  name           = "{{.VAppTemplateNameFromUrl}}"
  # Due to a bug in VCD we omit the description
  # description  = ""
  ovf_url        = "{{.OvfUrl}}"

  metadata = {
    vapp_template_metadata  = "vApp Template Metadata"
    vapp_template_metadata2 = "vApp Template Metadata2"
    vapp_template_metadata3 = "vApp Template Metadata3"
  }
}
`

const testAccCheckVcdVAppTemplateFromUrlUpdated = `
data "vcd_catalog" "{{.Catalog}}" {
  org  = "{{.Org}}"
  name = "{{.Catalog}}"
}

resource "vcd_catalog_vapp_template" "{{.VAppTemplateNameFromUrl}}" {
  org        = "{{.Org}}"
  catalog_id = data.vcd_catalog.{{.Catalog}}.id

  name           = "{{.VAppTemplateNameFromUrlUpdated}}"
  # Due to a bug in VCD we omit the description
  # description  = ""
  ovf_url        = "{{.OvfUrl}}"

  metadata = {
    vapp_template_metadata = "vApp Template Metadata"
    vapp_template_metadata2 = "vApp Template Metadata2_2"
  }
}
`
