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
	params["Description"] = "TestAccVcdCatalogBasicDescription-description"
	configText1 := templateFill(testAccCheckVcdCatalogStep1, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText1)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resourceAddress := "vcd_catalog.test-catalog"
	// Use field value caching function across multiple test steps to ensure object wasn't recreated (ID did not change)
	cachedId := &testCachedFieldValue{}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckCatalogDestroy,
		Steps: []resource.TestStep{
			// Provision catalog without storage profile
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					cachedId.cacheTestResourceFieldValue(resourceAddress, "id"),
					resource.TestCheckResourceAttr(resourceAddress, "name", TestAccVcdCatalogName),
					resource.TestCheckResourceAttr(resourceAddress, "description", TestAccVcdCatalogDescription),
					resource.TestCheckResourceAttr(resourceAddress, "storage_profile_id", ""),
					testAccCheckVcdCatalogExists(resourceAddress),
				),
			},
			// Set storage profile for existing catalog
			resource.TestStep{
				Config: configText1,
				Check: resource.ComposeTestCheckFunc(
					cachedId.testCheckCachedResourceFieldValue(resourceAddress, "id"),
					resource.TestCheckResourceAttr(resourceAddress, "name", TestAccVcdCatalogName),
					resource.TestCheckResourceAttr(resourceAddress, "description", "TestAccVcdCatalogBasicDescription-description"),
					resource.TestMatchResourceAttr(resourceAddress, "storage_profile_id",
						regexp.MustCompile(`^urn:vcloud:vdcstorageProfile:`)),
					testAccCheckVcdCatalogExists(resourceAddress),
				),
			},
			// Remove storage profile just like it was provisioned in step 0
			resource.TestStep{

				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					cachedId.testCheckCachedResourceFieldValue(resourceAddress, "id"),
					resource.TestCheckResourceAttr(resourceAddress, "name", TestAccVcdCatalogName),
					resource.TestCheckResourceAttr(resourceAddress, "description", TestAccVcdCatalogDescription),
					resource.TestCheckResourceAttr(resourceAddress, "storage_profile_id", ""),
					testAccCheckVcdCatalogExists(resourceAddress),
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

// TestAccVcdCatalogWithStorageProfile is very similar to TestAccVcdCatalog, but it ensure that a catalog can be created
// using specific storage profile
func TestAccVcdCatalogWithStorageProfile(t *testing.T) {
	var params = StringMap{
		"Org":            testConfig.VCD.Org,
		"Vdc":            testConfig.VCD.Vdc,
		"CatalogName":    TestAccVcdCatalogName,
		"Description":    TestAccVcdCatalogDescription,
		"StorageProfile": testConfig.VCD.ProviderVdc.StorageProfile,
		"Tags":           "catalog",
	}

	configText := templateFill(testAccCheckVcdCatalogStep1, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resourceAddress := "vcd_catalog.test-catalog"
	dataSourceAddress := "data.vcd_storage_profile.sp"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckCatalogDestroy,
		Steps: []resource.TestStep{
			// Provision with storage profile
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdCatalogExists(resourceAddress),
					resource.TestCheckResourceAttr(resourceAddress, "name", TestAccVcdCatalogName),
					resource.TestCheckResourceAttr(resourceAddress, "description", TestAccVcdCatalogDescription),
					resource.TestMatchResourceAttr(resourceAddress, "storage_profile_id",
						regexp.MustCompile(`^urn:vcloud:vdcstorageProfile:`)),
					resource.TestCheckResourceAttrPair(resourceAddress, "storage_profile_id", dataSourceAddress, "id"),
					checkStorageProfileOriginatesInParentVdc(dataSourceAddress,
						params["StorageProfile"].(string),
						params["Org"].(string),
						params["Vdc"].(string)),
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
resource "vcd_catalog" "test-catalog" {
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

resource "vcd_catalog" "test-catalog" {
  org = "{{.Org}}" 
  
  name               = "{{.CatalogName}}"
  description        = "{{.Description}}"
  storage_profile_id = data.vcd_storage_profile.sp.id

  delete_force      = "true"
  delete_recursive  = "true"
}
`
