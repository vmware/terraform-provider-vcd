// +build catalog ALL functional

package vcd

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/vmware/go-vcloud-director/v2/govcd"
)

var TestAccVcdCatalog = "TestAccVcdCatalogBasic"
var TestAccVcdCatalogDescription = "TestAccVcdCatalogBasicDescription"

func TestAccVcdCatalogBasic(t *testing.T) {

	var catalog govcd.Catalog
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
					testAccCheckVcdCatalogExists("vcd_catalog."+TestAccVcdCatalog, &catalog),
					resource.TestCheckResourceAttr(
						"vcd_catalog."+TestAccVcdCatalog, "name", TestAccVcdCatalog),
					resource.TestCheckResourceAttr(
						"vcd_catalog."+TestAccVcdCatalog, "description", TestAccVcdCatalogDescription),
				),
			},
			resource.TestStep{
				ResourceName:      "vcd_catalog." + TestAccVcdCatalog + "-import",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdByCatalog(TestAccVcdCatalog),
				// These fields can't be retrieved from catalog data
				ImportStateVerifyIgnore: []string{"delete_force", "delete_recursive"},
			},
		},
	})
}

func importStateIdByCatalog(objectName string) resource.ImportStateIdFunc {
	return func(*terraform.State) (string, error) {
		importId := testConfig.VCD.Org + "." + objectName
		if testConfig.VCD.Org == "" || objectName == "" {
			return "", fmt.Errorf("missing information to generate import path: %s", importId)
		}
		return importId, nil
	}
}

func testAccCheckVcdCatalogExists(name string, catalog *govcd.Catalog) resource.TestCheckFunc {
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

		newCatalog, err := adminOrg.GetCatalogByNameOrId(rs.Primary.ID, false)
		if err != nil {
			return fmt.Errorf("catalog %s does not exist (%s)", rs.Primary.ID, err)
		}

		catalog = newCatalog
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

		catalog, err := adminOrg.GetCatalogByName(rs.Primary.ID, false)

		if catalog != nil || err == nil {
			return fmt.Errorf("catalog %s still exists", rs.Primary.ID)
		}

	}

	return nil
}

func init() {
	testingTags["catalog"] = "resource_vcd_catalog_test.go"
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
