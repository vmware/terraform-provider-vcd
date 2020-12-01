// +build catalog ALL functional

package vcd

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	testingTags["catalog"] = "resource_vcd_catalog_test.go"
}

var TestAccVcdCatalogName = "TestAccVcdCatalog"
var TestAccVcdCatalogDescription = "TestAccVcdCatalogBasicDescription"

func TestAccVcdCatalog(t *testing.T) {
	var params = StringMap{
		"Org":            testConfig.VCD.Org,
		"CatalogName":    TestAccVcdCatalogName,
		"Description":    TestAccVcdCatalogDescription,
		"StorageProfile": testConfig.VCD.ProviderVdc.StorageProfile,
		"Tags":           "catalog",
	}

	configText := templateFill(testAccCheckVcdCatalog, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

	params["FuncName"] = t.Name() + "step1"
	configText1 := templateFill(testAccCheckVcdCatalogStep1, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText1)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resourceAddress := "vcd_catalog." + TestAccVcdCatalogName

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckCatalogDestroy,
		Steps: []resource.TestStep{
			// Provision catalog without storage profile
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdCatalogExists(resourceAddress),
					resource.TestCheckResourceAttr(resourceAddress, "name", TestAccVcdCatalogName),
					resource.TestCheckResourceAttr(resourceAddress, "description", TestAccVcdCatalogDescription),
					resource.TestCheckResourceAttr(resourceAddress, "storage_profile_id", ""),
				),
			},
			// Set storage profile for existing catalog
			resource.TestStep{
				Config: configText1,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdCatalogExists(resourceAddress),
					resource.TestCheckResourceAttr(resourceAddress, "name", TestAccVcdCatalogName),
					resource.TestCheckResourceAttr(resourceAddress, "description", TestAccVcdCatalogDescription),
					resource.TestMatchResourceAttr(resourceAddress, "storage_profile_id",
						regexp.MustCompile(`^urn:vcloud:vdcstorageProfile:.*`)),
				),
			},
			// Remove storage profile just like it was provisioned in step 0
			resource.TestStep{

				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdCatalogExists(resourceAddress),
					resource.TestCheckResourceAttr(resourceAddress, "name", TestAccVcdCatalogName),
					resource.TestCheckResourceAttr(resourceAddress, "description", TestAccVcdCatalogDescription),
					resource.TestCheckResourceAttr(resourceAddress, "storage_profile_id", ""),
				),
			},
			resource.TestStep{
				ResourceName:      resourceAddress,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdOrgObject(testConfig, TestAccVcdCatalogName),
				// These fields can't be retrieved from catalog data
				ImportStateVerifyIgnore: []string{"delete_force", "delete_recursive"},
			},
		},
	})
}

func TestAccVcdCatalogWithStorageProfile(t *testing.T) {
	var params = StringMap{
		"Org":            testConfig.VCD.Org,
		"CatalogName":    TestAccVcdCatalogName,
		"Description":    TestAccVcdCatalogDescription,
		"StorageProfile": testConfig.VCD.ProviderVdc.StorageProfile,
		"Tags":           "catalog",
	}

	configText := templateFill(testAccCheckVcdCatalog, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resourceAddress := "vcd_catalog." + TestAccVcdCatalogName

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckCatalogDestroy,
		Steps: []resource.TestStep{
			// Provision catalog without storage profile
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdCatalogExists(resourceAddress),
					resource.TestCheckResourceAttr(resourceAddress, "name", TestAccVcdCatalogName),
					resource.TestCheckResourceAttr(resourceAddress, "description", TestAccVcdCatalogDescription),
					resource.TestMatchResourceAttr(resourceAddress, "storage_profile_id",
						regexp.MustCompile(`^urn:vcloud:vdcstorageProfile:.*`)),
				),
			},
		},
	})
}

func testAccCheckVcdCatalogExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no Org ID is set")
		}

		conn := testAccProvider.Meta().(*VCDClient)

		adminOrg, err := conn.GetAdminOrg(testConfig.VCD.Org)
		if err != nil {
			return fmt.Errorf(errorRetrievingOrg, testConfig.VCD.Org+" and error: "+err.Error())
		}

		_, err = adminOrg.GetCatalogByNameOrId(rs.Primary.ID, false)
		if err != nil {
			return fmt.Errorf("catalog %s does not exist (%s)", rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccCheckCatalogDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*VCDClient)
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "vcd_catalog" && rs.Primary.Attributes["name"] != TestAccVcdCatalogName {
			continue
		}

		adminOrg, err := conn.GetAdminOrg(testConfig.VCD.Org)
		if err != nil {
			return fmt.Errorf(errorRetrievingOrg, testConfig.VCD.Org+" and error: "+err.Error())
		}

		_, err = adminOrg.GetCatalogByName(rs.Primary.ID, false)

		if err == nil {
			return fmt.Errorf("catalog %s still exists", rs.Primary.ID)
		}

	}

	return nil
}

const testAccCheckVcdCatalog = `
resource "vcd_catalog" "{{.CatalogName}}" {
  org = "{{.Org}}" 
  
  name        = "{{.CatalogName}}"
  description = "{{.Description}}"

  delete_force      = "true"
  delete_recursive  = "true"
}
`

const testAccCheckVcdCatalogStep1 = `
data "vcd_storage_profile" "sp" {
	name = "{{.StorageProfile}}"
}

resource "vcd_catalog" "{{.CatalogName}}" {
  org = "{{.Org}}" 
  
  name               = "{{.CatalogName}}"
  description        = "{{.Description}}"
  storage_profile_id = data.vcd_storage_profile.sp.id

  delete_force      = "true"
  delete_recursive  = "true"
}
`
