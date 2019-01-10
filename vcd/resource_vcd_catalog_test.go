package vcd

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/vmware/go-vcloud-director/govcd"
)

var TestAccVcdCatalog = "TestAccVcdCatalogBasic"
var TestAccVcdCatalogDescription = "TestAccVcdCatalogBasicDescription"

func TestAccVcdCatalogBasic(t *testing.T) {

	var catalog govcd.Catalog
	var params = StringMap{
		"Org":         testConfig.VCD.Org,
		"CatalogName": TestAccVcdCatalog,
		"Description": TestAccVcdCatalogDescription,
	}

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	configText := templateFill(testAccCheckVcdCatalogBasic, params)
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
		},
	})
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

		newCatalog, err := adminOrg.FindCatalog(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("catalog %s does not exist (%#v)", rs.Primary.ID, newCatalog)
		}

		catalog = &newCatalog
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

		catalog, err := adminOrg.FindCatalog(rs.Primary.ID)

		if catalog != (govcd.Catalog{}) {
			return fmt.Errorf("catalog %s still exists", rs.Primary.ID)
		}
		if err != nil {
			return fmt.Errorf("catalog %s still exists or other error: %#v", rs.Primary.ID, err)
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
