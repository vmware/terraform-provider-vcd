// +build catalog ALL functional

package vcd

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func init() {
	testingTags["catalog"] = "resource_vcd_catalog_test.go"
}

var TestAccVcdCatalog = "TestAccVcdCatalogBasic"
var TestAccVcdCatalogDescription = "TestAccVcdCatalogBasicDescription"

func TestAccVcdCatalogBasic(t *testing.T) {

	var params = StringMap{
		"Org":         testConfig.VCD.Org,
		"CatalogName": TestAccVcdCatalog,
		"Description": TestAccVcdCatalogDescription,
		"Tags":        "catalog",
	}

	configText := templateFill(testAccCheckVcdCatalogBasic, params)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCatalogDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdCatalogExists("vcd_catalog."+TestAccVcdCatalog),
					resource.TestCheckResourceAttr(
						"vcd_catalog."+TestAccVcdCatalog, "name", TestAccVcdCatalog),
					resource.TestCheckResourceAttr(
						"vcd_catalog."+TestAccVcdCatalog, "description", TestAccVcdCatalogDescription),
				),
			},
			// Step to ensure that "READ" works properly and proposes to create new item when an object does not exist
			// PlanOnly must true because otherwise it does not complain on the plan
			resource.TestStep{
				Config:             configText,
				PreConfig:          testDeleteExistingCatalog(t, TestAccVcdCatalog),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
				Check:              testAccCheckVcdCatalogExists("vcd_catalog." + TestAccVcdCatalog),
			},
			// Recreate the resource after detecting deletion
			resource.TestStep{
				Config: configText,
				Check:  testAccCheckVcdCatalogExists("vcd_catalog." + TestAccVcdCatalog),
			},
			resource.TestStep{
				ResourceName:      "vcd_catalog." + TestAccVcdCatalog + "-import",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdOrgObject(testConfig, TestAccVcdCatalog),
				// These fields can't be retrieved from catalog data
				ImportStateVerifyIgnore: []string{"delete_force", "delete_recursive"},
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
		if rs.Type != "vcd_catalog" && rs.Primary.Attributes["name"] != TestAccVcdCatalog {
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

const testAccCheckVcdCatalogBasic = `
resource "vcd_catalog" "{{.CatalogName}}" {
  org = "{{.Org}}" 
  
  name = "{{.CatalogName}}"
  description = "{{.Description}}"

  delete_force      = "true"
  delete_recursive  = "true"
}
`

// testDeleteExistingCatalog deletes catalog with name from test or returns a failure
func testDeleteExistingCatalog(t *testing.T, catalogName string) func() {
	return func() {
		vcdClient := createTemporaryVCDConnection()

		org, _, err := vcdClient.GetOrgAndVdc(testConfig.VCD.Org, testConfig.VCD.Vdc)
		if err != nil {
			t.Errorf(err.Error())
		}

		catalog, err := org.GetCatalogByName(catalogName, false)
		if err != nil {
			t.Errorf(err.Error())
		}

		err = catalog.Delete(false, false)
		if err != nil {
			t.Errorf(err.Error())
		}
		return
	}
}
