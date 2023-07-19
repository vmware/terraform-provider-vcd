//go:build catalog || functional || access_control || ALL

package vcd

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// TestAccVcdCatalogOrgUser tests the following:
//   - Accessing a catalog that was shared from another Org
//   - Accessing a catalog with duplicate name (due to sharing)
//   - Accessing catalog items belonging to the shared catalog
func TestAccVcdCatalogOrgUser(t *testing.T) {
	preTestChecks(t)

	// NOTE: this test uses multiple providers to test org user access,
	// and system administrator is one of the required providers
	skipIfNotSysAdmin(t)

	catalogName := t.Name() + "-cat"
	catalogMediaName := t.Name() + "-media"
	vappTemplateName := t.Name() + "-templ"
	localVmName := t.Name() + "-vm"
	org1Name := testConfig.VCD.Org
	org2Name := testConfig.VCD.Org + "-1"
	vdc2Name := testConfig.Nsxt.Vdc + "-1"
	descriptionOrg1 := "Belongs to " + org1Name
	descriptionOrg2 := "Belongs to " + org2Name
	var params = StringMap{
		"Org1":               org1Name,
		"Org2":               org2Name,
		"Vdc1":               testConfig.Nsxt.Vdc,
		"Vdc2":               vdc2Name,
		"SharedToEveryone":   "true",
		"NsxtStorageProfile": testConfig.VCD.NsxtProviderVdc.StorageProfile2,
		"CatalogName":        catalogName,
		"DescriptionOrg1":    descriptionOrg1,
		"DescriptionOrg2":    descriptionOrg2,
		"FuncName":           t.Name() + "-creation",
		"CatalogMediaName":   catalogMediaName,
		"VappTemplateName":   vappTemplateName,
		"MediaPath":          testConfig.Media.MediaPath,
		"OvaPath":            testConfig.Ova.OvaPath,
		"UploadPieceSize":    testConfig.Media.UploadPieceSize,
		"VmName":             localVmName,
		"SkipNotice":         "# skip-binary-test: temporary phase",
		"ProviderSystem":     providerVcdSystem,
		"ProviderOrg1":       providerVcdOrg1,
		"ProviderOrg2":       providerVcdOrg2,
		"Tags":               "catalog",
	}
	testParamsNotEmpty(t, params)

	configText := templateFill(testAccCatalogCreation, params)
	params["FuncName"] = t.Name() + "-access"
	// Remove skip: the full script will run fine in binary tests
	params["SkipNotice"] = " "
	accessText := templateFill(testAccCatalogCreation+testAccCatalogAccessOrgUser, params)

	params["SkipNotice"] = "# skip-binary-test: timing problems"
	params["FuncName"] = t.Name() + "-vm-creation"
	vmCreationText := templateFill(testAccCatalogCreation+testAccCatalogAccessOrgUser+testAccVMsFromSharedCatalogs, params)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	debugPrintf("#[DEBUG] CREATION CONFIGURATION: %s", configText)
	debugPrintf("#[DEBUG] ACCESS CONFIGURATION: %s", accessText)
	debugPrintf("#[DEBUG] VM CREATION CONFIGURATION: %s", vmCreationText)

	resourceCatalogOrg1 := "vcd_catalog.catalog_org1"
	resourceCatalogOrg2 := "vcd_catalog.catalog_org2"
	resourceMedia1 := "vcd_catalog_media.test_media1"
	resourcevAppTemplate1 := "vcd_catalog_vapp_template.test_vapp_template1"
	dataSourceCatalogOrg1 := "data.vcd_catalog.catalog_org1"
	dataSourceCatalogOrg2 := "data.vcd_catalog.catalog_org2"
	resource.Test(t, resource.TestCase{
		ProviderFactories: buildMultipleProviders(),
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckCatalogEntityState("vcd_catalog", org1Name, catalogName, false),
			testAccCheckCatalogEntityState("vcd_catalog", org2Name, catalogName, false),
		),
		Steps: []resource.TestStep{
			// Test creation
			{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCatalogEntityState("vcd_catalog", org1Name, catalogName, true),
					testAccCheckCatalogEntityState("vcd_catalog", org2Name, catalogName, true),
					testAccCheckCatalogEntityState("vcd_catalog_media", org2Name, catalogMediaName, true),
					testAccCheckCatalogEntityState("vcd_catalog_vapp_template", org2Name, vappTemplateName, true),
					resource.TestCheckResourceAttr(resourceCatalogOrg1, "name", catalogName),
					resource.TestCheckResourceAttr(resourceCatalogOrg1, "org", org1Name),
					resource.TestCheckResourceAttr(resourceCatalogOrg1, "description", descriptionOrg1),
					resource.TestCheckResourceAttr(resourceCatalogOrg2, "name", catalogName),
					resource.TestCheckResourceAttr(resourceCatalogOrg2, "org", org2Name),
					resource.TestCheckResourceAttr(resourceCatalogOrg2, "description", descriptionOrg2),
					resource.TestCheckResourceAttr(resourceMedia1, "org", org1Name),
					resource.TestCheckResourceAttr(resourceMedia1, "catalog", catalogName),
					resource.TestCheckResourceAttr(resourceMedia1, "name", catalogMediaName),
					resource.TestCheckResourceAttr(resourceMedia1, "description", descriptionOrg1),
					resource.TestCheckResourceAttr(resourcevAppTemplate1, "org", org1Name),
					resource.TestCheckResourceAttr(resourcevAppTemplate1, "name", vappTemplateName),
					resource.TestCheckResourceAttrPair(resourcevAppTemplate1, "catalog_id", resourceCatalogOrg1, "id"),
					resource.TestCheckResourceAttrPair(resourceMedia1, "catalog_id", resourceCatalogOrg1, "id"),
				),
			},
			{
				Config: accessText,
				Check: resource.ComposeTestCheckFunc(
					logState("catalog-related-data-sources"),
					resource.TestCheckResourceAttr(dataSourceCatalogOrg1, "name", catalogName),
					resource.TestCheckResourceAttr(dataSourceCatalogOrg1, "org", org1Name),
					resource.TestCheckResourceAttr(dataSourceCatalogOrg1, "description", descriptionOrg1),
					resource.TestCheckResourceAttr(dataSourceCatalogOrg2, "name", catalogName),
					resource.TestCheckResourceAttr(dataSourceCatalogOrg2, "org", org2Name),
					resource.TestCheckResourceAttr(dataSourceCatalogOrg2, "description", descriptionOrg2),

					resource.TestCheckResourceAttr("data.vcd_catalog.catalog_org1_from_org2", "name", catalogName),
					resource.TestCheckResourceAttr("data.vcd_catalog.catalog_org1_from_org2", "org", org1Name),
					resource.TestCheckResourceAttr("data.vcd_catalog.catalog_org1_from_org2", "description", descriptionOrg1),
					resource.TestCheckResourceAttr("data.vcd_catalog_media.test_media_by_catalog_name", "name", catalogMediaName),
					resource.TestCheckResourceAttr("data.vcd_catalog_media.test_media_by_catalog_name", "org", org1Name),
					resource.TestCheckResourceAttr("data.vcd_catalog_media.test_media_by_catalog_name", "catalog", catalogName),
					resource.TestCheckResourceAttr("data.vcd_catalog_media.test_media_by_catalog_id", "catalog", catalogName),
					resource.TestCheckResourceAttr("data.vcd_catalog_media.test_media_by_catalog_id", "description", descriptionOrg1),
					resource.TestCheckResourceAttr("data.vcd_catalog_vapp_template.test_vapp_template1", "name", vappTemplateName),
					resource.TestCheckResourceAttr("data.vcd_catalog_vapp_template.test_vapp_template1", "org", org1Name),
					resource.TestCheckResourceAttr("data.vcd_catalog_vapp_template.test_vapp_template1", "description", descriptionOrg1),
					resource.TestCheckResourceAttrPair("data.vcd_catalog_vapp_template.test_vapp_template1", "catalog_id", dataSourceCatalogOrg1, "id"),
				)},
			{
				Config: vmCreationText,
				Check: resource.ComposeTestCheckFunc(
					logState("vm-creation"),
					testAccCheckVcdStandaloneVmExists(localVmName+"-1", "vcd_vm."+localVmName+"-1", org2Name, vdc2Name),
					testAccCheckVcdStandaloneVmExists(localVmName+"-2", "vcd_vm."+localVmName+"-2", org2Name, vdc2Name),
					resource.TestCheckResourceAttr("vcd_vm."+localVmName+"-1", "name", localVmName+"-1"),
					resource.TestCheckResourceAttr("vcd_vm."+localVmName+"-2", "name", localVmName+"-2"),
				),
			},
		},
	})
	postTestChecks(t)
}

func testAccCheckCatalogEntityState(entityType, orgName, entityName string, wantExist bool) func(*terraform.State) error {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*VCDClient)
		for _, rs := range s.RootModule().Resources {
			if rs.Type != entityType {
				continue
			}
			if rs.Primary.Attributes["org"] != orgName {
				continue
			}
			if rs.Primary.Attributes["name"] != entityName {
				continue
			}
			var err error
			switch entityType {
			case "vcd_catalog":
				_, err = conn.Client.GetCatalogById(rs.Primary.ID)
			case "vcd_catalog_media":
				_, err = conn.QueryMediaById(rs.Primary.ID)
			case "vcd_catalog_vapp_template":
				_, err = conn.GetVAppTemplateById(rs.Primary.ID)
			}
			if wantExist && err != nil {
				return fmt.Errorf("%s %s/%s not found", entityType, orgName, entityName)
			}
			if !wantExist && err == nil {
				return fmt.Errorf("%s %s/%s still exists", entityType, orgName, entityName)
			}
		}
		return nil
	}
}

// testAccCatalogCreation contains the catalog and item creation
// This script is skipped in binary tests when used alone, and enabled
// when merged with testAccCatalogAccessOrgUser
const testAccCatalogCreation = `
{{.SkipNotice}}

data "vcd_org" "org2" {
	name = "{{.Org2}}"
}

data "vcd_storage_profile" "sp1" {
  provider = {{.ProviderSystem}}
  org      = "{{.Org1}}"
  vdc      = "{{.Vdc1}}"
  name     = "{{.NsxtStorageProfile}}"
}

resource "vcd_catalog" "catalog_org1" {
  provider           = {{.ProviderSystem}}
  org                = "{{.Org1}}"
  name               = "{{.CatalogName}}"
  description        = "{{.DescriptionOrg1}}"
  storage_profile_id = data.vcd_storage_profile.sp1.id
  delete_force     = true
  delete_recursive = true
}

resource "vcd_catalog_access_control" "catalog_org1" {
  provider    = {{.ProviderSystem}}
  org         = "{{.Org1}}"
  catalog_id  = vcd_catalog.catalog_org1.id

  shared_with_everyone = false

  shared_with {
    org_id       = data.vcd_org.org2.id
    access_level = "ReadOnly"
  }
}

resource "vcd_catalog_media"  "test_media1" {
  provider   = {{.ProviderOrg1}}
  org        = "{{.Org1}}"
  catalog_id = vcd_catalog.catalog_org1.id

  name                 = "{{.CatalogMediaName}}"
  description          = "{{.DescriptionOrg1}}"
  media_path           = "{{.MediaPath}}"
  upload_piece_size    = {{.UploadPieceSize}}
}

resource "vcd_catalog_vapp_template" "test_vapp_template1" {
  provider    = {{.ProviderOrg1}}
  org         = "{{.Org1}}"
  catalog_id  = vcd_catalog.catalog_org1.id
  description = "{{.DescriptionOrg1}}"

  name              = "{{.VappTemplateName}}"
  ova_path          = "{{.OvaPath}}"
  upload_piece_size = {{.UploadPieceSize}}
}

data "vcd_storage_profile" "sp2" {
  provider = {{.ProviderOrg2}}
  org      = "{{.Org2}}"
  vdc      = "{{.Vdc2}}"
  name     = "{{.NsxtStorageProfile}}"
}

# create a catalog with the same name in Org2
resource "vcd_catalog" "catalog_org2" {
  provider           = {{.ProviderOrg2}}
  org                = "{{.Org2}}"
  name               = "{{.CatalogName}}"
  description        = "{{.DescriptionOrg2}}"
  storage_profile_id = data.vcd_storage_profile.sp2.id
  delete_force     = true
  delete_recursive = true
}
`

// testAccCatalogAccessOrgUser shows the org user accessing
// catalog items from a shared catalog.
// This script runs in conjunction with testAccCatalogCreation.
// NOTE: the "depends_on" clause are there to ensure that we
// don't query the shared entities before the sharing operation is complete.
const testAccCatalogAccessOrgUser = `

# retrieve the catalog created in Org1
data "vcd_catalog" "catalog_org1" {
  provider = {{.ProviderOrg1}}
  org      = "{{.Org1}}"
  name     = vcd_catalog.catalog_org1.name

  depends_on = [vcd_catalog.catalog_org1]
}

# retrieve the catalog created in Org2
data "vcd_catalog" "catalog_org2" {
  provider = {{.ProviderOrg2}}
  org      = "{{.Org2}}"
  name     = "{{.CatalogName}}"

  depends_on = [vcd_catalog.catalog_org2]
}

# retrieve the catalog created in the other Org
data "vcd_catalog" "catalog_org1_from_org2" {
  provider = {{.ProviderOrg2}}
  org      = "{{.Org1}}"
  name     = "{{.CatalogName}}"

  depends_on = [vcd_catalog_access_control.catalog_org1]
}

# retrieve the media item (by catalog name) from the shared catalog
data "vcd_catalog_media" "test_media_by_catalog_name" {
  provider = {{.ProviderOrg2}}
  org      = "{{.Org1}}"
  catalog  = data.vcd_catalog.catalog_org1_from_org2.name
  name     = "{{.CatalogMediaName}}"

  depends_on = [vcd_catalog_access_control.catalog_org1]
}  

# retrieve the media item (by catalog ID) from the shared catalog
data "vcd_catalog_media" "test_media_by_catalog_id" {
  provider   = {{.ProviderOrg2}}
  catalog_id = data.vcd_catalog.catalog_org1.id
  name       = "{{.CatalogMediaName}}"

  depends_on = [vcd_catalog_access_control.catalog_org1]
}

# retrieve the vApp template from the shared catalog
data "vcd_catalog_vapp_template" "test_vapp_template1" {
  provider   = {{.ProviderOrg2}}
  org        = "{{.Org1}}"
  catalog_id = data.vcd_catalog.catalog_org1_from_org2.id
  name       = "{{.VappTemplateName}}"

  depends_on = [vcd_catalog_access_control.catalog_org1]
}

# retrieve the vApp template from the shared catalog as catalog item
data "vcd_catalog_item" "test_vapp_template1" {
  provider   = {{.ProviderOrg2}}
  org        = "{{.Org1}}"
  catalog    = data.vcd_catalog.catalog_org1_from_org2.name
  name       = "{{.VappTemplateName}}"

  depends_on = [vcd_catalog_access_control.catalog_org1]
}
`

// testAccVMsFromSharedCatalogs shows how to build VMs with shared resources
// NOTE: the depends_on clause is needed to guarantee the order of deletion
const testAccVMsFromSharedCatalogs = `

resource "vcd_vm" "{{.VmName}}-1" {
  provider = {{.ProviderOrg2}}

  org              = "{{.Org2}}"
  vdc              = "{{.Vdc2}}"
  name             = "{{.VmName}}-1"
  vapp_template_id = data.vcd_catalog_vapp_template.test_vapp_template1.id
  description      = "test standalone VM 1"
  power_on         = false
}

resource "vcd_vm" "{{.VmName}}-2" {
  provider = {{.ProviderOrg2}}

  org              = "{{.Org2}}"
  vdc              = "{{.Vdc2}}"
  name             = "{{.VmName}}-2"
  boot_image_id    = data.vcd_catalog_media.test_media_by_catalog_id.id
  description      = "test standalone VM 2"
  computer_name    = "standalone"
  cpus             = 1
  memory           = 1024
  os_type          = "sles10_64Guest"
  hardware_version = "vmx-14"
  power_on         = false

  depends_on = [vcd_catalog_media.test_media1]
}
`
