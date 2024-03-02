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
	preTestChecks(t)
	vAppTemplateName := t.Name()
	vAppTemplateDescription := vAppTemplateName + "Description"
	vAppTemplateFromUrlName := t.Name() + "FromUrl"

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
					resource.TestCheckResourceAttr(resourceVAppTemplate, "lease.0.storage_lease_in_sec", fmt.Sprintf("%d", 3600*24*3)),
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
					resource.TestCheckResourceAttr(resourceVAppTemplateFromUrl, "lease.0.storage_lease_in_sec", fmt.Sprintf("%d", 0)),
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
					resource.TestCheckResourceAttr(resourceVAppTemplateFromUrl, "lease.0.storage_lease_in_sec", fmt.Sprintf("%d", 3600*18)),
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
	return testAccCheckVcdVAppTemplateExistsInCatalog(testSuiteCatalogName, vAppTemplateName)
}

func testAccCheckVcdVAppTemplateExistsInCatalog(catalogName, vAppTemplateName string) resource.TestCheckFunc {
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

		catalog, err := org.GetCatalogByName(catalogName, false)
		if err != nil {
			return fmt.Errorf("catalog %s does not exist: %s", catalogName, err)
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

  name              = "{{.VAppTemplateName}}"
  description       = "{{.Description}}"
  ova_path          = "{{.OvaPath}}"
  upload_piece_size = {{.UploadPieceSize}}

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

  name              = "{{.VAppTemplateName}}Updated"
  description       = "{{.Description}}Updated"
  ova_path          = "{{.OvaPath}}"
  upload_piece_size = {{.UploadPieceSize}}

  lease {
	storage_lease_in_sec = 3600*24*3
  }

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

  lease {
	storage_lease_in_sec = 0
  }

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

  lease {
	storage_lease_in_sec = 3600*18
  }

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
		}, true)
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

func TestAccVcdCatalogVAppTemplateCaptureEmptyPoweredOffVms(t *testing.T) {
	preTestChecks(t)
	vAppTemplateName := t.Name()
	vAppTemplateDescription := vAppTemplateName + "Description"

	if testConfig.Ova.OvfUrl == "" {
		t.Skip("Variable Ova.OvfUrl must be set in test configuration")
	}

	var params = StringMap{
		"Org":              testConfig.VCD.Org,
		"Vdc":              testConfig.Nsxt.Vdc,
		"Catalog":          testConfig.VCD.Catalog.NsxtBackedCatalogName,
		"CatalogItem":      testConfig.VCD.Catalog.CatalogItemWithMultiVms,
		"VAppTemplateName": vAppTemplateName,
		"Description":      vAppTemplateDescription,
		"TestName":         t.Name(),
	}
	config1 := templateFill(testAccVcdCatalogVAppTemplateCaptureEmptyVms, params)
	params["FuncName"] = t.Name() + "-step2"
	config2 := templateFill(testAccVcdCatalogVAppTemplateCaptureEmptyVmsDS, params)

	ignoreDatasourceCheckFields := []string{"%", "capture_vapp.0.copy_tpm_on_instantiate", "capture_vapp.0.source_id",
		"capture_vapp.0.customize_on_instantiate", "capture_vapp.0.%", "capture_vapp.#", "upload_piece_size"}

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	debugPrintf("#[DEBUG] CONFIGURATION: %s", config1)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { preRunChecks(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: config1,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdVAppTemplateExistsInCatalog(testConfig.VCD.Catalog.NsxtBackedCatalogName, "vcd_catalog_vapp_template.from-vapp-no-customization"),
					testAccCheckVcdVAppTemplateExistsInCatalog(testConfig.VCD.Catalog.NsxtBackedCatalogName, "vcd_catalog_vapp_template.from-standalone-vm-no-customization"),
					testAccCheckVcdVAppTemplateExistsInCatalog(testConfig.VCD.Catalog.NsxtBackedCatalogName, "vcd_catalog_vapp_template.from-vapp-customization"),
					testAccCheckVcdVAppTemplateExistsInCatalog(testConfig.VCD.Catalog.NsxtBackedCatalogName, "vcd_catalog_vapp_template.from-standalone-vm-customization"),

					resource.TestCheckResourceAttr("vcd_catalog_vapp_template.from-vapp-no-customization", "name", vAppTemplateName+"-from-vapp-no-cust"),
					resource.TestCheckResourceAttr("vcd_catalog_vapp_template.from-vapp-no-customization", "description", vAppTemplateDescription),
					resource.TestCheckResourceAttrSet("vcd_catalog_vapp_template.from-vapp-no-customization", "vm_names.0"),

					resource.TestCheckResourceAttr("vcd_catalog_vapp_template.from-standalone-vm-no-customization", "name", vAppTemplateName+"-from-standalone-no-cust"),
					resource.TestCheckResourceAttr("vcd_catalog_vapp_template.from-standalone-vm-no-customization", "description", vAppTemplateDescription),
					resource.TestCheckResourceAttrSet("vcd_catalog_vapp_template.from-standalone-vm-no-customization", "vm_names.0"),

					resource.TestCheckResourceAttr("vcd_catalog_vapp_template.from-vapp-customization", "name", vAppTemplateName+"-from-vapp-cust"),
					resource.TestCheckResourceAttr("vcd_catalog_vapp_template.from-vapp-customization", "description", vAppTemplateDescription),
					resource.TestCheckResourceAttrSet("vcd_catalog_vapp_template.from-vapp-customization", "vm_names.0"),

					resource.TestCheckResourceAttr("vcd_catalog_vapp_template.from-standalone-vm-customization", "name", vAppTemplateName+"-from-standalone-cust"),
					resource.TestCheckResourceAttr("vcd_catalog_vapp_template.from-standalone-vm-customization", "description", vAppTemplateDescription),
					resource.TestCheckResourceAttrSet("vcd_catalog_vapp_template.from-standalone-vm-customization", "vm_names.0"),
				),
			},
			{
				Config: config2,
				Check: resource.ComposeTestCheckFunc(
					resourceFieldsEqual("vcd_catalog_vapp_template.from-vapp-no-customization", "data.vcd_catalog_vapp_template.from-vapp-no-customization", ignoreDatasourceCheckFields),
					resourceFieldsEqual("vcd_catalog_vapp_template.from-standalone-vm-no-customization", "data.vcd_catalog_vapp_template.from-standalone-vm-no-customization", ignoreDatasourceCheckFields),
					resourceFieldsEqual("vcd_catalog_vapp_template.from-vapp-customization", "data.vcd_catalog_vapp_template.from-vapp-customization", ignoreDatasourceCheckFields),
					resourceFieldsEqual("vcd_catalog_vapp_template.from-standalone-vm-customization", "data.vcd_catalog_vapp_template.from-standalone-vm-customization", ignoreDatasourceCheckFields),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdCatalogVAppTemplateCaptureEmptyVms = `
data "vcd_catalog" "cat" {
  org  = "{{.Org}}"
  name = "{{.Catalog}}"
}

resource "vcd_vapp" "web" {
  org      = "{{.Org}}"
  name     = "{{.TestName}}-vapp"
  power_on = false
}

resource "vcd_vapp_vm" "emptyVM" {
  org  = "{{.Org}}"
  name = "{{.TestName}}-vm"
  
  power_on      = false
  vapp_name     = vcd_vapp.web.name
  computer_name = "emptyVM"
  memory        = 1024
  cpus          = 2
  cpu_cores     = 1

  os_type          = "sles10_64Guest"
  hardware_version = "vmx-14"
}

resource "vcd_vm" "standalone" {
  org      = "{{.Org}}"
  name     = "{{.TestName}}-vm"
  power_on = false

  computer_name = "emptyVM"
  memory        = 1024
  cpus          = 2
  cpu_cores     = 1

  os_type          = "sles10_64Guest"
  hardware_version = "vmx-14"
}

resource "vcd_catalog_vapp_template" "from-vapp-no-customization" {
  org        = "{{.Org}}"
  catalog_id = data.vcd_catalog.cat.id

  name              = "{{.VAppTemplateName}}-from-vapp-no-cust"
  description       = "{{.Description}}"

  capture_vapp {
	source_id                = vcd_vapp.web.id
	customize_on_instantiate = false
  }

  depends_on = [ vcd_vapp_vm.emptyVM ]
}

resource "vcd_catalog_vapp_template" "from-standalone-vm-no-customization" {
  org        = "{{.Org}}"
  catalog_id = data.vcd_catalog.cat.id

  name              = "{{.VAppTemplateName}}-from-standalone-no-cust"
  description       = "{{.Description}}"

  capture_vapp {
    source_id                = vcd_vm.standalone.vapp_id
    customize_on_instantiate = false
  }
}

resource "vcd_catalog_vapp_template" "from-vapp-customization" {
  org        = "{{.Org}}"
  catalog_id = data.vcd_catalog.cat.id
  
  name        = "{{.VAppTemplateName}}-from-vapp-cust"
  description = "{{.Description}}"
  
  capture_vapp {
    source_id                = vcd_vapp.web.id
    customize_on_instantiate = true
  }

  depends_on = [ vcd_vapp_vm.emptyVM ]
}
  
resource "vcd_catalog_vapp_template" "from-standalone-vm-customization" {
  org        = "{{.Org}}"
  catalog_id = data.vcd_catalog.cat.id

  name        = "{{.VAppTemplateName}}-from-standalone-cust"
  description = "{{.Description}}"
  
  capture_vapp {
    source_id                = vcd_vm.standalone.vapp_id
    customize_on_instantiate = true
  }
}
`

const testAccVcdCatalogVAppTemplateCaptureEmptyVmsDS = testAccVcdCatalogVAppTemplateCaptureEmptyVms + `
data "vcd_catalog_vapp_template" "from-vapp-no-customization" {
  org        = "{{.Org}}"
  catalog_id = data.vcd_catalog.cat.id
  name       = vcd_catalog_vapp_template.from-vapp-no-customization.name
}

data "vcd_catalog_vapp_template" "from-standalone-vm-no-customization" {
  org        = "{{.Org}}"
  catalog_id = data.vcd_catalog.cat.id
  name       = vcd_catalog_vapp_template.from-standalone-vm-no-customization.name
}

data "vcd_catalog_vapp_template" "from-vapp-customization" {
  org        = "{{.Org}}"
  catalog_id = data.vcd_catalog.cat.id
  name       = vcd_catalog_vapp_template.from-vapp-customization.name
}

data "vcd_catalog_vapp_template" "from-standalone-vm-customization" {
  org        = "{{.Org}}"
  catalog_id = data.vcd_catalog.cat.id
  name       = vcd_catalog_vapp_template.from-standalone-vm-customization.name
}
`

func TestAccVcdCatalogVAppTemplateCaptureTemplatePoweredOnVms(t *testing.T) {
	preTestChecks(t)
	vAppTemplateName := t.Name()
	vAppTemplateDescription := vAppTemplateName + "Description"

	if testConfig.Ova.OvfUrl == "" {
		t.Skip("Variable Ova.OvfUrl must be set in test configuration")
	}

	var params = StringMap{
		"Org":              testConfig.VCD.Org,
		"Vdc":              testConfig.Nsxt.Vdc,
		"Catalog":          testConfig.VCD.Catalog.NsxtBackedCatalogName,
		"CatalogItem":      testConfig.VCD.Catalog.CatalogItemWithMultiVms,
		"VAppTemplateName": vAppTemplateName,
		"Description":      vAppTemplateDescription,
		"TestName":         t.Name(),
	}
	config1 := templateFill(testAccVcdCatalogVAppTemplateCaptureTemplateVmsStep1, params)
	params["FuncName"] = t.Name() + "-step2"
	config2 := templateFill(testAccVcdCatalogVAppTemplateCaptureTemplateVmsStep2, params)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	debugPrintf("#[DEBUG] CONFIGURATION: %s", config1)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { preRunChecks(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: config1,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdVAppTemplateExistsInCatalog(testConfig.VCD.Catalog.NsxtBackedCatalogName, "vcd_catalog_vapp_template.from-vapp-no-customization"),
					testAccCheckVcdVAppTemplateExistsInCatalog(testConfig.VCD.Catalog.NsxtBackedCatalogName, "vcd_catalog_vapp_template.from-standalone-vm-no-customization"),

					resource.TestCheckResourceAttr("vcd_catalog_vapp_template.from-vapp-no-customization", "name", vAppTemplateName+"-from-vapp-no-cust"),
					resource.TestCheckResourceAttr("vcd_catalog_vapp_template.from-vapp-no-customization", "description", vAppTemplateDescription),
					resource.TestCheckResourceAttrSet("vcd_catalog_vapp_template.from-vapp-no-customization", "vm_names.0"),
					resource.TestCheckResourceAttr("vcd_catalog_vapp_template.from-vapp-no-customization", "lease.0.storage_lease_in_sec", fmt.Sprintf("%d", 3600*24*3)),
					resource.TestCheckResourceAttr("vcd_catalog_vapp_template.from-vapp-no-customization", "metadata.vapp_template_metadata", "vApp Template Metadata"),
					resource.TestCheckResourceAttr("vcd_catalog_vapp_template.from-vapp-no-customization", "metadata.vapp_template_metadata2", "vApp Template Metadata2"),
					resource.TestCheckResourceAttr("vcd_catalog_vapp_template.from-vapp-no-customization", "metadata.vapp_template_metadata3", "vApp Template Metadata3"),

					resource.TestCheckResourceAttr("vcd_catalog_vapp_template.from-standalone-vm-no-customization", "name", vAppTemplateName+"-from-standalone-no-cust"),
					resource.TestCheckResourceAttr("vcd_catalog_vapp_template.from-standalone-vm-no-customization", "description", vAppTemplateDescription),
					resource.TestCheckResourceAttrSet("vcd_catalog_vapp_template.from-standalone-vm-no-customization", "vm_names.0"),
					resource.TestCheckResourceAttr("vcd_catalog_vapp_template.from-standalone-vm-no-customization", "lease.0.storage_lease_in_sec", fmt.Sprintf("%d", 3600*24*3)),
					resource.TestCheckResourceAttr("vcd_catalog_vapp_template.from-standalone-vm-no-customization", "metadata.vapp_template_metadata", "vApp Template Metadata"),
					resource.TestCheckResourceAttr("vcd_catalog_vapp_template.from-standalone-vm-no-customization", "metadata.vapp_template_metadata2", "vApp Template Metadata2"),
					resource.TestCheckResourceAttr("vcd_catalog_vapp_template.from-standalone-vm-no-customization", "metadata.vapp_template_metadata3", "vApp Template Metadata3"),
				),
			},
			{
				Config: config2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vcd_vapp.captured", "name", t.Name()+"-vapp-captured"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.captured", "name", t.Name()+"-vm"),
					resource.TestCheckResourceAttr("vcd_vm.captured", "name", t.Name()+"-vm"),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdCatalogVAppTemplateCaptureTemplateVmsStep1 = `
data "vcd_catalog" "cat" {
  org  = "{{.Org}}"
  name = "{{.Catalog}}"
}

data "vcd_catalog_vapp_template" "three-vms" {
  org        = "{{.Org}}"
  catalog_id = data.vcd_catalog.cat.id
  name       = "{{.CatalogItem}}"
}

resource "vcd_vapp" "web" {
  org      = "{{.Org}}"
  name     = "{{.TestName}}-vapp"
  power_on = true
}

resource "vcd_vapp_vm" "emptyVM" {
  org  = "{{.Org}}"
  name = "{{.TestName}}-vm"

  vapp_template_id = data.vcd_catalog_vapp_template.three-vms.id
  
  power_on      = true
  vapp_name     = vcd_vapp.web.name
  computer_name = "emptyVM"
  memory        = 1024
  cpus          = 2
  cpu_cores     = 1 
}

resource "vcd_vm" "standalone" {
  org      = "{{.Org}}"
  name     = "{{.TestName}}-vm"
  power_on = true

  vapp_template_id = data.vcd_catalog_vapp_template.three-vms.id

  computer_name = "emptyVM"
  memory        = 1024
  cpus          = 2
  cpu_cores     = 1
}

resource "vcd_catalog_vapp_template" "from-vapp-no-customization" {
  org        = "{{.Org}}"
  catalog_id = data.vcd_catalog.cat.id

  name              = "{{.VAppTemplateName}}-from-vapp-no-cust"
  description       = "{{.Description}}"

  capture_vapp {
	source_id                = vcd_vapp.web.id
	customize_on_instantiate = false
  }

  lease {
	storage_lease_in_sec = 3600*24*3
  }

  metadata = {
    vapp_template_metadata  = "vApp Template Metadata"
    vapp_template_metadata2 = "vApp Template Metadata2"
    vapp_template_metadata3 = "vApp Template Metadata3"
  }

  depends_on = [ vcd_vapp_vm.emptyVM ]
}

resource "vcd_catalog_vapp_template" "from-standalone-vm-no-customization" {
  org        = "{{.Org}}"
  catalog_id = data.vcd_catalog.cat.id

  name              = "{{.VAppTemplateName}}-from-standalone-no-cust"
  description       = "{{.Description}}"

  capture_vapp {
    source_id                = vcd_vm.standalone.vapp_id
    customize_on_instantiate = false
  }

  lease {
	storage_lease_in_sec = 3600*24*3
  }

  metadata = {
    vapp_template_metadata  = "vApp Template Metadata"
    vapp_template_metadata2 = "vApp Template Metadata2"
    vapp_template_metadata3 = "vApp Template Metadata3"
  }
}
`

const testAccVcdCatalogVAppTemplateCaptureTemplateVmsStep2 = testAccVcdCatalogVAppTemplateCaptureTemplateVmsStep1 + `
resource "vcd_vapp" "captured" {
  org      = "{{.Org}}"
  name     = "{{.TestName}}-vapp-captured"
  power_on = true
}

resource "vcd_vapp_vm" "captured" {
  org  = "{{.Org}}"
  name = "{{.TestName}}-vm"

  vapp_template_id = vcd_catalog_vapp_template.from-vapp-no-customization.id
  
  power_on      = true
  vapp_name     = vcd_vapp.captured.name
  computer_name = "emptyVM"
  memory        = 1024
  cpus          = 2
  cpu_cores     = 1 
}

resource "vcd_vm" "captured" {
  org      = "{{.Org}}"
  name     = "{{.TestName}}-vm"
  power_on = true

  vapp_template_id = vcd_catalog_vapp_template.from-standalone-vm-no-customization.id

  computer_name = "emptyVM"
  memory        = 1024
  cpus          = 2
  cpu_cores     = 1
}
`

// func TestAccVcdCatalogVAppTemplateCaptureEmptyPoweredOffEfiTpm(t *testing.T) {
// 	preTestChecks(t)
// 	vAppTemplateName := t.Name()
// 	vAppTemplateDescription := vAppTemplateName + "Description"

// 	if testConfig.Ova.OvfUrl == "" {
// 		t.Skip("Variable Ova.OvfUrl must be set in test configuration")
// 	}

// 	var params = StringMap{
// 		"Org":              testConfig.VCD.Org,
// 		"Vdc":              testConfig.Nsxt.Vdc,
// 		"Catalog":          testConfig.VCD.Catalog.NsxtBackedCatalogName,
// 		"CatalogItem":      testConfig.VCD.Catalog.CatalogItemWithEfiSupport,
// 		"VAppTemplateName": vAppTemplateName,
// 		"Description":      vAppTemplateDescription,
// 		"TestName":         t.Name(),
// 	}
// 	config1 := templateFill(testAccVcdCatalogVAppTemplateCaptureEmptyPoweredOffEfiTpm, params)

// 	if vcdShortTest {
// 		t.Skip(acceptanceTestsSkipped)
// 		return
// 	}
// 	debugPrintf("#[DEBUG] CONFIGURATION: %s", config1)
// 	resource.Test(t, resource.TestCase{
// 		PreCheck:          func() { preRunChecks(t) },
// 		ProviderFactories: testAccProviders,
// 		// CheckDestroy:      testAccCheckVAppTemplateDestroy(vAppTemplateFromUrlName + "Updated"),
// 		Steps: []resource.TestStep{
// 			{
// 				Config: config1,
// 				Check:  resource.ComposeTestCheckFunc(
// 				// testAccCheckVcdVAppTemplateExists(resourceVAppTemplate),
// 				),
// 			},
// 		},
// 	})
// 	postTestChecks(t)
// }

const testAccVcdCatalogVAppTemplateCaptureEmptyPoweredOffEfiTpm = `
data "vcd_catalog" "cat" {
  org  = "{{.Org}}"
  name = "{{.Catalog}}"
}

data "vcd_catalog_vapp_template" "three-vms" {
  org        = "{{.Org}}"
  catalog_id = data.vcd_catalog.cat.id
  name       = "{{.CatalogItem}}"
}

resource "vcd_vapp" "web" {
  org      = "{{.Org}}"
  name     = "{{.TestName}}-vapp"
  power_on = true
}

resource "vcd_vapp_vm" "emptyVM" {
  org  = "{{.Org}}"
  name = "{{.TestName}}-vm"

  vapp_template_id = data.vcd_catalog_vapp_template.three-vms.id
  
  firmware = "efi"
  
  power_on      = true
  vapp_name     = vcd_vapp.web.name
  computer_name = "emptyVM"
  memory        = 1024
  cpus          = 2
  cpu_cores     = 1 
}

resource "vcd_vm" "standalone" {
  org      = "{{.Org}}"
  name     = "{{.TestName}}-vm"
  power_on = true

  vapp_template_id = data.vcd_catalog_vapp_template.three-vms.id

  firmware = "efi"

  computer_name = "emptyVM"
  memory        = 1024
  cpus          = 2
  cpu_cores     = 1
}

resource "vcd_catalog_vapp_template" "from-vapp-no-customization" {
  org        = "{{.Org}}"
  catalog_id = data.vcd_catalog.cat.id

  name                    = "{{.VAppTemplateName}}-from-vapp-no-cust"
  description             = "{{.Description}}"

  capture_vapp {
	source_id                = vcd_vapp.web.id
	customize_on_instantiate = false
	copy_tpm_on_instantiate  = true
  }

  lease {
	storage_lease_in_sec = 3600*24*3
  }

  depends_on = [ vcd_vapp_vm.emptyVM ]
}

resource "vcd_catalog_vapp_template" "from-standalone-vm-no-customization" {
  org        = "{{.Org}}"
  catalog_id = data.vcd_catalog.cat.id

  name                    = "{{.VAppTemplateName}}-from-standalone-no-cust"
  description             = "{{.Description}}"

  capture_vapp {
    source_id                = vcd_vm.standalone.vapp_id
    customize_on_instantiate = false
	copy_tpm_on_instantiate  = true
  }

  lease {
	storage_lease_in_sec = 3600*24*3
  }
}

resource "vcd_vapp" "captured" {
  org      = "{{.Org}}"
  name     = "{{.TestName}}-vapp-captured"
  power_on = true
}

resource "vcd_vapp_vm" "captured" {
  org  = "{{.Org}}"
  name = "{{.TestName}}-vm"

  vapp_template_id = vcd_catalog_vapp_template.from-vapp-no-customization.id
  
  firmware      = "efi"
  power_on      = true
  vapp_name     = vcd_vapp.captured.name
  computer_name = "emptyVM"
  memory        = 1024
  cpus          = 2
  cpu_cores     = 1 
}

resource "vcd_vm" "captured" {
  org      = "{{.Org}}"
  name     = "{{.TestName}}-vm"
  power_on = true

  vapp_template_id = vcd_catalog_vapp_template.from-standalone-vm-no-customization.id

  firmware      = "efi"
  computer_name = "emptyVM"
  memory        = 1024
  cpus          = 2
  cpu_cores     = 1
}
`

func TestAccVcdCatalogVAppTemplateOverwriteExistingItem(t *testing.T) {
	preTestChecks(t)

	// The test uses SDK to upload an item
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	// Get the data from configuration file. This client is still inactive at this point
	vcdClient, err := getTestVCDFromJson(testConfig)
	if err != nil {
		t.Fatalf(err.Error())
	}
	err = ProviderAuthenticate(vcdClient, testConfig.Provider.User, testConfig.Provider.Password, testConfig.Provider.Token, testConfig.Provider.SysOrg, testConfig.Provider.ApiToken, testConfig.Provider.ApiTokenFile, testConfig.Provider.ServiceAccountTokenFile)
	if err != nil {
		t.Fatalf("authentication error: %s", err)
	}
	org, err := vcdClient.GetAdminOrgByName(testConfig.VCD.Org)
	if err != nil {
		t.Fatalf("org not found : %s", err)
	}

	catalog, err := org.GetCatalogByName(testConfig.VCD.Catalog.NsxtBackedCatalogName, false)
	if err != nil {
		t.Fatalf("catalog not found : %s", err)
	}

	itemName1 := t.Name() + "-1"
	itemName2 := t.Name() + "-2"
	description := t.Name() + "-description"
	uploadTask1, err := catalog.UploadOvfByLink(testConfig.Ova.OvfUrl, itemName1, description)
	if err != nil {
		t.Fatalf("upload failed : %s", err)
	}
	uploadTask2, err := catalog.UploadOvfByLink(testConfig.Ova.OvfUrl, itemName2, description)
	if err != nil {
		t.Fatalf("upload failed : %s", err)
	}
	err = uploadTask1.WaitTaskCompletion()
	if err != nil {
		t.Fatalf("upload task failed : %s", err)
	}
	err = uploadTask2.WaitTaskCompletion()
	if err != nil {
		t.Fatalf("upload task failed : %s", err)
	}
	vAppTemplate, err := catalog.GetVAppTemplateByName(itemName1)
	if err != nil {
		t.Fatalf("error retrieving vApp template after upload : %s", err)
	}

	//
	vappTemplateId := vAppTemplate.VAppTemplate.ID
	catalogItemId, err := vAppTemplate.GetCatalogItemId()
	if err != nil {
		t.Fatalf("error catalog item ID for vApp : %s", err)
	}

	var params = StringMap{
		"Org":                    testConfig.VCD.Org,
		"Vdc":                    testConfig.Nsxt.Vdc,
		"Catalog":                testConfig.VCD.Catalog.NsxtBackedCatalogName,
		"CatalogItem":            testConfig.VCD.Catalog.CatalogItemWithMultiVms,
		"VAppTemplateName":       t.Name(),
		"Description":            t.Name() + "-description",
		"TestName":               t.Name(),
		"ExistingVappTemplateId": vappTemplateId,
		"ExistingCatalogItemId":  catalogItemId,
		"UploadedItemName1":      itemName1,
		"UploadedItemName2":      itemName2,
	}
	config1 := templateFill(testAccVcdCatalogVAppTemplateOverwriteItem, params)
	params["FuncName"] = t.Name() + "-step2"
	config2 := templateFill(testAccVcdCatalogVAppTemplateOverwriteItem2, params)

	cachePrecreatedTemplateId1 := &testCachedFieldValue{}
	cachePrecreatedTemplateId2 := &testCachedFieldValue{}

	debugPrintf("#[DEBUG] CONFIGURATION: %s", config1)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { preRunChecks(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: config1,
				Check: resource.ComposeTestCheckFunc(
					cachePrecreatedTemplateId1.cacheTestResourceFieldValue("data.vcd_catalog_item.existing", "id"),
					cachePrecreatedTemplateId2.cacheTestResourceFieldValue("data.vcd_catalog_vapp_template.existing", "catalog_item_id"),
				),
			},
			{
				Config: config2,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdVAppTemplateExistsInCatalog(testConfig.VCD.Catalog.NsxtBackedCatalogName, "vcd_catalog_vapp_template.catalog-item-overwrite"),
					testAccCheckVcdVAppTemplateExistsInCatalog(testConfig.VCD.Catalog.NsxtBackedCatalogName, "vcd_catalog_vapp_template.vapp-template-overwrite"),

					// Catalog item ID remains the same
					cachePrecreatedTemplateId1.testCheckCachedResourceFieldValue("vcd_catalog_vapp_template.catalog-item-overwrite", "catalog_item_id"),
					cachePrecreatedTemplateId2.testCheckCachedResourceFieldValue("vcd_catalog_vapp_template.vapp-template-overwrite", "catalog_item_id"),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdCatalogVAppTemplateOverwriteItem = `
data "vcd_catalog" "cat" {
  org  = "{{.Org}}"
  name = "{{.Catalog}}"
}

resource "vcd_vm" "standalone" {
  org      = "{{.Org}}"
  name     = "{{.TestName}}-vm"
  power_on = false

  computer_name = "emptyVM"
  memory        = 1024
  cpus          = 2
  cpu_cores     = 1

  os_type          = "sles10_64Guest"
  hardware_version = "vmx-14"
}

data "vcd_catalog_item" "existing" {
  org     = "{{.Org}}"
  catalog = "{{.Catalog}}"
  name    = "{{.UploadedItemName1}}"
}

data "vcd_catalog_vapp_template" "existing" {
  org        = "{{.Org}}"
  catalog_id = data.vcd_catalog.cat.id
  name       = "{{.UploadedItemName2}}"
}
`

const testAccVcdCatalogVAppTemplateOverwriteItem2 = testAccVcdCatalogVAppTemplateOverwriteItem + `
resource "vcd_catalog_vapp_template" "catalog-item-overwrite" {
  org        = "{{.Org}}"
  catalog_id = data.vcd_catalog.cat.id
  
  name        = "{{.UploadedItemName1}}"
  description = "{{.Description}}"
  
  capture_vapp {
    source_id                 = vcd_vm.standalone.vapp_id
    customize_on_instantiate  = true

	# check that 'vcd_catalog_item' id can be used
	overwrite_catalog_item_id = data.vcd_catalog_item.existing.id	
  }
}

resource "vcd_catalog_vapp_template" "vapp-template-overwrite" {
  org        = "{{.Org}}"
  catalog_id = data.vcd_catalog.cat.id
	
  name        = "{{.UploadedItemName2}}"
  description = "{{.Description}}"
	
  capture_vapp {
    source_id                 = vcd_vm.standalone.vapp_id
    customize_on_instantiate  = true

	# check that 'catalog_item_id' from 'vcd_catalog_vapp_template' can be used
    overwrite_catalog_item_id = data.vcd_catalog_vapp_template.existing.catalog_item_id
  }
}
`
