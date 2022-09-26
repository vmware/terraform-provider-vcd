//go:build catalog || ALL || functional
// +build catalog ALL functional

package vcd

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// Test catalog and vApp Template data sources.
// Using a catalog data source we reference a vApp Template data source.
// Using a vApp Template data source we create another vApp Template.
// where the description is the first data source ID.
func TestAccVcdCatalogAndVappTemplateDatasource(t *testing.T) {
	preTestChecks(t)
	createdVAppTemplateName := t.Name()

	var params = StringMap{
		"Org":          testConfig.VCD.Org,
		"Vdc":          testConfig.Nsxt.Vdc,
		"Catalog":      testConfig.VCD.Catalog.NsxtBackedCatalogName,
		"VAppTemplate": testConfig.VCD.Catalog.NsxtCatalogItem,
	}
	testParamsNotEmpty(t, params)

	configText := templateFill(testAccCheckVcdCatalogVAppTemplateDS, params)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

	datasourceCatalog := "data.vcd_catalog." + params["Catalog"].(string)
	datasourceVdc := "data.vcd_org_vdc." + params["Vdc"].(string)
	datasourceCatalogVappTemplate1 := "data.vcd_catalog_vapp_template." + params["VAppTemplate"].(string) + "_1"
	datasourceCatalogVappTemplate2 := "data.vcd_catalog_vapp_template." + params["VAppTemplate"].(string) + "_2"
	datasourceCatalogVappTemplate3 := "data.vcd_catalog_vapp_template." + params["VAppTemplate"].(string) + "_3"
	datasourceCatalogVappTemplate4 := "data.vcd_catalog_vapp_template." + params["VAppTemplate"].(string) + "_4"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { preRunChecks(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      catalogVAppTemplateDestroyed(testSuiteCatalogName, createdVAppTemplateName),
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(datasourceCatalogVappTemplate1, "id", regexp.MustCompile(`urn:vcloud:vapptemplate:[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$`)),
					// Check that the attributes from the retrieved vApp Template match the related elements
					resource.TestCheckResourceAttrPair(datasourceCatalog, "id", datasourceCatalogVappTemplate1, "catalog_id"),
					resource.TestCheckResourceAttrPair(datasourceVdc, "id", datasourceCatalogVappTemplate1, "vdc_id"),

					// Check both data sources fetched by VDC and Catalog ID are equal
					resource.TestCheckResourceAttrPair(datasourceCatalogVappTemplate1, "id", datasourceCatalogVappTemplate2, "id"),
					resource.TestCheckResourceAttrPair(datasourceCatalogVappTemplate1, "catalog_id", datasourceCatalogVappTemplate2, "catalog_id"),
					resource.TestCheckResourceAttrPair(datasourceCatalogVappTemplate1, "vdc_id", datasourceCatalogVappTemplate2, "vdc_id"),

					// Check data sources with filter. Not using resourceFieldsEqual here as we'd need to exclude all filtering options by hardcoding the combinations.
					resource.TestCheckResourceAttrPair(datasourceCatalogVappTemplate3, "id", datasourceCatalogVappTemplate1, "id"),
					resource.TestCheckResourceAttrPair(datasourceCatalogVappTemplate3, "catalog_id", datasourceCatalogVappTemplate1, "catalog_id"),
					resource.TestCheckResourceAttrPair(datasourceCatalogVappTemplate3, "vdc_id", datasourceCatalogVappTemplate1, "vdc_id"),
					resource.TestCheckResourceAttrPair(datasourceCatalogVappTemplate4, "id", datasourceCatalogVappTemplate1, "id"),
					resource.TestCheckResourceAttrPair(datasourceCatalogVappTemplate4, "catalog_id", datasourceCatalogVappTemplate1, "catalog_id"),
					resource.TestCheckResourceAttrPair(datasourceCatalogVappTemplate4, "vdc_id", datasourceCatalogVappTemplate1, "vdc_id"),
				),
			},
		},
	})
	postTestChecks(t)
}

func catalogVAppTemplateDestroyed(catalog, itemName string) resource.TestCheckFunc {
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
		_, err = cat.GetVAppTemplateByName(itemName)
		if err == nil {
			return fmt.Errorf("vApp Template %s not deleted", itemName)
		}
		return nil
	}
}

const testAccCheckVcdCatalogVAppTemplateDS = `
data "vcd_catalog" "{{.Catalog}}" {
  org  = "{{.Org}}"
  name = "{{.Catalog}}"
}

data "vcd_org_vdc" "{{.Vdc}}" {
  org  = "{{.Org}}"
  name = "{{.Vdc}}"
}

data "vcd_catalog_vapp_template" "{{.VAppTemplate}}_1" {
  org        = "{{.Org}}"
  catalog_id = data.vcd_catalog.{{.Catalog}}.id
  name       = "{{.VAppTemplate}}"
}

data "vcd_catalog_vapp_template" "{{.VAppTemplate}}_2" {
  org    = "{{.Org}}"
  vdc_id = data.vcd_org_vdc.{{.Vdc}}.id
  name   = "{{.VAppTemplate}}"
}

data "vcd_catalog_vapp_template" "{{.VAppTemplate}}_3" {
  org        = "{{.Org}}"
  catalog_id = data.vcd_catalog.{{.Catalog}}.id
  filter {
    name_regex = "{{.VAppTemplate}}"
  }
}

data "vcd_catalog_vapp_template" "{{.VAppTemplate}}_4" {
  org    = "{{.Org}}"
  vdc_id = data.vcd_org_vdc.{{.Vdc}}.id
  filter {
    name_regex = "{{.VAppTemplate}}"
  }
}
`
