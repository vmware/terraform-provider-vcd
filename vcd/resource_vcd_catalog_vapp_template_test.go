//go:build catalog || ALL || functional

package vcd

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccVcdCatalogVAppTemplateResource(t *testing.T) {
	vAppTemplateName := t.Name()
	vAppTemplateDescription := vAppTemplateName + "Description"
	vAppTemplateFromUrlName := t.Name() + "FromUrl"

	preTestChecks(t)

	if testConfig.Ova.OvfUrl == "" {
		t.Skip("Variable Ova.OvfUrl must be set in test configuration")
	}

	var params = StringMap{
		"Org":              testConfig.VCD.Org,
		"Vdc":              testConfig.VCD.Vdc, // TODO: Use NSX-T VDC by default
		"Catalog":          testSuiteCatalogName,
		"VAppTemplateName": vAppTemplateName,
		"Description":      vAppTemplateDescription,
		"OvaPath":          testConfig.Ova.OvaPath,
		"OvfUrl":           testConfig.Ova.OvfUrl,
		"UploadPieceSize":  testConfig.Ova.UploadPieceSize,
	}
	createConfigHcl := templateFill(testAccCheckVcdVAppTemplateCreate, params)

	params["FuncName"] = t.Name() + "-Update"
	updateConfigHcl := templateFill(testAccCheckVcdVAppTemplateUpdate, params)

	params["FuncName"] = t.Name() + "-FromUrl"
	params["VAppTemplateName"] = vAppTemplateFromUrlName
	createWithUrlConfigHcl := templateFill(testAccCheckVcdVAppTemplateFromUrlCreate, params)

	params["FuncName"] = t.Name() + "-FromUrlUpdate"
	updateWithUrlConfigHcl := templateFill(testAccCheckVcdVAppTemplateFromUrlUpdate, params)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	debugPrintf("#[DEBUG] CONFIGURATION: %s", createConfigHcl)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", updateConfigHcl)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", createWithUrlConfigHcl)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", updateWithUrlConfigHcl)

	resourceVAppTemplate := "vcd_catalog_vapp_template." + vAppTemplateName
	resourceVAppTemplateFromUrl := "vcd_catalog_vapp_template." + vAppTemplateFromUrlName
	datasourceVdc := "data.vcd_org_vdc." + params["Vdc"].(string)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { preRunChecks(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVAppTemplateDestroy(vAppTemplateFromUrlName + "Updated"),
		Steps: []resource.TestStep{
			{
				Config: createConfigHcl,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdVAppTemplateExists(resourceVAppTemplate),
					resource.TestCheckResourceAttr(resourceVAppTemplate, "name", vAppTemplateName),
					resource.TestCheckResourceAttr(resourceVAppTemplate, "description", vAppTemplateDescription),
					resource.TestCheckResourceAttrPair(resourceVAppTemplate, "vdc_id", datasourceVdc, "id"),
					resource.TestCheckResourceAttrSet(resourceVAppTemplate, "vm_names.0"),
					resource.TestCheckResourceAttr(resourceVAppTemplate, "metadata.vapp_template_metadata", "vApp Template Metadata"),
					resource.TestCheckResourceAttr(resourceVAppTemplate, "metadata.vapp_template_metadata2", "vApp Template Metadata2"),
				),
			},
			{
				Config: updateConfigHcl,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdVAppTemplateExists(resourceVAppTemplate),
					resource.TestCheckResourceAttr(resourceVAppTemplate, "name", vAppTemplateName+"Updated"),
					resource.TestCheckResourceAttr(resourceVAppTemplate, "description", vAppTemplateDescription+"Updated"),
					resource.TestCheckResourceAttrPair(resourceVAppTemplate, "vdc_id", datasourceVdc, "id"),
					resource.TestCheckResourceAttrSet(resourceVAppTemplate, "vm_names.0"),
					resource.TestCheckResourceAttr(resourceVAppTemplate, "metadata.vapp_template_metadata", "vApp Template Metadata v2"),
					resource.TestCheckResourceAttr(resourceVAppTemplate, "metadata.vapp_template_metadata2", "vApp Template Metadata2 v2"),
					resource.TestCheckResourceAttr(resourceVAppTemplate, "metadata.vapp_template_metadata3", "vApp Template Metadata3"),
				),
			},
			{
				Config: createWithUrlConfigHcl,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdVAppTemplateExists(resourceVAppTemplateFromUrl),
					resource.TestCheckResourceAttr(resourceVAppTemplateFromUrl, "name", vAppTemplateFromUrlName),
					// FIXME: Due to a bug in VCD, description is overridden by the present in the OVA
					resource.TestMatchResourceAttr(resourceVAppTemplateFromUrl, "description", regexp.MustCompile(`^Name: yVM.*`)),
					resource.TestCheckResourceAttrPair(resourceVAppTemplateFromUrl, "vdc_id", datasourceVdc, "id"),
					resource.TestCheckResourceAttrSet(resourceVAppTemplateFromUrl, "vm_names.0"),
					resource.TestCheckResourceAttr(resourceVAppTemplateFromUrl, "metadata.vapp_template_metadata", "vApp Template Metadata"),
					resource.TestCheckResourceAttr(resourceVAppTemplateFromUrl, "metadata.vapp_template_metadata2", "vApp Template Metadata2"),
					resource.TestCheckResourceAttr(resourceVAppTemplateFromUrl, "metadata.vapp_template_metadata3", "vApp Template Metadata3"),
				),
			},
			{
				Config: updateWithUrlConfigHcl,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdVAppTemplateExists(resourceVAppTemplateFromUrl),
					resource.TestCheckResourceAttr(resourceVAppTemplateFromUrl, "name", vAppTemplateFromUrlName+"Updated"),
					// FIXME: Due to a bug in VCD, description is overridden by the present in the OVA
					resource.TestMatchResourceAttr(resourceVAppTemplateFromUrl, "description", regexp.MustCompile(`^Name: yVM.*`)),
					resource.TestCheckResourceAttrPair(resourceVAppTemplateFromUrl, "vdc_id", datasourceVdc, "id"),
					resource.TestCheckResourceAttrSet(resourceVAppTemplateFromUrl, "vm_names.0"),
					resource.TestCheckResourceAttr(resourceVAppTemplateFromUrl, "metadata.vapp_template_metadata", "vApp Template Metadata"),
					resource.TestCheckResourceAttr(resourceVAppTemplateFromUrl, "metadata.vapp_template_metadata2", "vApp Template Metadata2_2"),
				),
			},
			{
				ResourceName:      resourceVAppTemplateFromUrl,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdOrgCatalogObject(vAppTemplateFromUrlName + "Updated"),
				// These fields can't be retrieved from vApp Template data
				ImportStateVerifyIgnore: []string{"ovf_url", "ova_path", "upload_piece_size"},
			},
			{
				ResourceName:      resourceVAppTemplateFromUrl,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdOrgVdcObject(vAppTemplateFromUrlName + "Updated"),
				// These fields can't be retrieved from vApp Template data
				ImportStateVerifyIgnore: []string{"ovf_url", "ova_path", "upload_piece_size"},
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

func testAccCheckVcdVAppTemplateExists(vAppTemplateName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		vAppTemplateRs, ok := s.RootModule().Resources[vAppTemplateName]
		if !ok {
			return fmt.Errorf("not found: %s", vAppTemplateName)
		}

		if vAppTemplateRs.Primary.ID == "" {
			return fmt.Errorf("no vApp Template ID is set")
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

		_, err = catalog.GetVAppTemplateByName(vAppTemplateRs.Primary.Attributes["name"])
		if err != nil {
			return fmt.Errorf("vApp Template %s does not exist (%s)", vAppTemplateRs.Primary.ID, err)
		}

		return nil
	}
}

func testAccCheckVAppTemplateDestroy(vAppTemplateName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*VCDClient)
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "vcd_catalog_vapp_template" && rs.Primary.Attributes["name"] != vAppTemplateName {
				continue
			}

			_, vdc, err := conn.GetOrgAndVdc(testConfig.VCD.Org, testConfig.VCD.Vdc)
			if err != nil {
				return fmt.Errorf(errorRetrievingOrg, testConfig.VCD.Org+" and error: "+err.Error())
			}

			itemName := rs.Primary.Attributes["name"]
			_, err = vdc.GetVAppTemplateByName(itemName)
			if err == nil {
				return fmt.Errorf("vApp Template %s still exists", itemName)
			}
		}

		return nil
	}

}

const testAccCheckVcdVAppTemplateCreate = `
data "vcd_catalog" "{{.Catalog}}" {
  org  = "{{.Org}}"
  name = "{{.Catalog}}"
}

data "vcd_org_vdc" "{{.Vdc}}" {
  org  = "{{.Org}}"
  name = "{{.Vdc}}"
}

resource "vcd_catalog_vapp_template" "{{.VAppTemplateName}}" {
  org        = "{{.Org}}"
  catalog_id = data.vcd_catalog.{{.Catalog}}.id

  name                 = "{{.VAppTemplateName}}"
  description          = "{{.Description}}"
  ova_path             = "{{.OvaPath}}"
  upload_piece_size    = {{.UploadPieceSize}}

  metadata = {
    vapp_template_metadata  = "vApp Template Metadata"
    vapp_template_metadata2 = "vApp Template Metadata2"
  }
}
`

const testAccCheckVcdVAppTemplateUpdate = `
data "vcd_catalog" "{{.Catalog}}" {
  org  = "{{.Org}}"
  name = "{{.Catalog}}"
}

data "vcd_org_vdc" "{{.Vdc}}" {
  org  = "{{.Org}}"
  name = "{{.Vdc}}"
}

resource "vcd_catalog_vapp_template" "{{.VAppTemplateName}}" {
  org        = "{{.Org}}"
  catalog_id = data.vcd_catalog.{{.Catalog}}.id

  name                 = "{{.VAppTemplateName}}Updated"
  description          = "{{.Description}}Updated"
  ova_path             = "{{.OvaPath}}"
  upload_piece_size    = {{.UploadPieceSize}}

  metadata = {
    vapp_template_metadata  = "vApp Template Metadata v2"
    vapp_template_metadata2 = "vApp Template Metadata2 v2"
    vapp_template_metadata3 = "vApp Template Metadata3"
  }
}
`

const testAccCheckVcdVAppTemplateFromUrlCreate = `
data "vcd_catalog" "{{.Catalog}}" {
  org  = "{{.Org}}"
  name = "{{.Catalog}}"
}

data "vcd_org_vdc" "{{.Vdc}}" {
  org  = "{{.Org}}"
  name = "{{.Vdc}}"
}

resource "vcd_catalog_vapp_template" "{{.VAppTemplateName}}" {
  org        = "{{.Org}}"
  catalog_id = data.vcd_catalog.{{.Catalog}}.id

  name           = "{{.VAppTemplateName}}"
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

const testAccCheckVcdVAppTemplateFromUrlUpdate = `
data "vcd_catalog" "{{.Catalog}}" {
  org  = "{{.Org}}"
  name = "{{.Catalog}}"
}

data "vcd_org_vdc" "{{.Vdc}}" {
  org  = "{{.Org}}"
  name = "{{.Vdc}}"
}

resource "vcd_catalog_vapp_template" "{{.VAppTemplateName}}" {
  org        = "{{.Org}}"
  catalog_id = data.vcd_catalog.{{.Catalog}}.id

  name           = "{{.VAppTemplateName}}Updated"
  # Due to a bug in VCD we omit the description
  # description  = ""
  ovf_url        = "{{.OvfUrl}}"

  metadata = {
    vapp_template_metadata  = "vApp Template Metadata"
    vapp_template_metadata2 = "vApp Template Metadata2_2"
  }
}
`

// TestAccVcdCatalogVAppTemplateMetadata tests metadata CRUD on Catalog vApp Templates
func TestAccVcdCatalogVAppTemplateMetadata(t *testing.T) {
	testMetadataEntryCRUD(t,
		testAccCheckVcdCatalogVAppTemplateMetadata, "vcd_catalog_vapp_template.test-catalog-vapp-template",
		testAccCheckVcdCatalogVAppTemplateMetadataDatasource, "data.vcd_catalog_vapp_template.test-catalog-vapp-template-ds",
		StringMap{
			"Catalog": testConfig.VCD.Catalog.NsxtBackedCatalogName,
			"OvfUrl":  testConfig.Ova.OvfUrl,
		})
}

const testAccCheckVcdCatalogVAppTemplateMetadata = `
data "vcd_catalog" "test-catalog" {
  org  = "{{.Org}}"
  name = "{{.Catalog}}"
}

resource "vcd_catalog_vapp_template" "test-catalog-vapp-template" {
  org        = data.vcd_catalog.test-catalog.org
  catalog_id = data.vcd_catalog.test-catalog.id
  name       = "{{.Name}}"
  ovf_url    = "{{.OvfUrl}}"
  {{.Metadata}}
}
`

const testAccCheckVcdCatalogVAppTemplateMetadataDatasource = `
data "vcd_catalog_vapp_template" "test-catalog-vapp-template-ds" {
  org        = vcd_catalog_vapp_template.test-catalog-vapp-template.org
  catalog_id = vcd_catalog_vapp_template.test-catalog-vapp-template.catalog_id
  name       = vcd_catalog_vapp_template.test-catalog-vapp-template.name
}
`

func TestAccVcdCatalogVAppTemplateMetadataIgnore(t *testing.T) {
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
		media, err := catalog.GetVAppTemplateById(id)
		if err != nil {
			return nil, fmt.Errorf("could not retrieve vApp Template '%s': %s", id, err)
		}
		return media, nil
	}

	testMetadataEntryIgnore(t,
		testAccCheckVcdCatalogVAppTemplateMetadata, "vcd_catalog_vapp_template.test-catalog-vapp-template",
		testAccCheckVcdCatalogVAppTemplateMetadataDatasource, "data.vcd_catalog_vapp_template.test-catalog-vapp-template-ds",
		getObjectById, StringMap{
			"Catalog": testConfig.VCD.Catalog.NsxtBackedCatalogName,
			"OvfUrl":  testConfig.Ova.OvfUrl,
		})
}
