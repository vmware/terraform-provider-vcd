//go:build org || ALL || functional

package vcd

import (
	_ "embed"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"path/filepath"
	"testing"
)

func TestAccVcdOrgSaml(t *testing.T) {
	preTestChecks(t)

	var orgName = t.Name()
	var metadataFile = "../test-resources/saml-test-idp.xml"
	if !fileExists(metadataFile) {
		t.Skipf("metadata file %s not found\n", metadataFile)
	}
	metadataFullName, err := filepath.Abs(metadataFile)
	if err != nil {
		t.Skipf("could not achieve full file name for metadata file %s: %s", metadataFile, err)
	}
	var params = StringMap{
		"OrgName":      orgName,
		"FuncName":     orgName,
		"FullName":     "Full " + orgName,
		"Description":  "Organization " + orgName,
		"EntityId":     orgName,
		"MetadataFile": metadataFullName,
		"Tags":         "org",
	}
	testParamsNotEmpty(t, params)

	skipIfNotSysAdmin(t)

	configText := templateFill(testAccCheckVcdOrgSaml, params)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

	resourceOrgName := "vcd_org." + orgName
	resourceOrgSamlName := "vcd_org_saml." + orgName
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckOrgDestroy(orgName),
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdOrgExists(resourceOrgName),
					resource.TestCheckResourceAttr(resourceOrgName, "name", orgName),
					resource.TestCheckResourceAttr(resourceOrgName, "full_name", params["FullName"].(string)),
					resource.TestCheckResourceAttr(resourceOrgName, "description", params["Description"].(string)),
					resource.TestCheckResourceAttr(resourceOrgName, "is_enabled", "true"),
					resource.TestCheckResourceAttr(resourceOrgSamlName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceOrgSamlName, "entity_id", orgName),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccCheckVcdOrgSaml = `
resource "vcd_org" "{{.OrgName}}" {
  name              = "{{.OrgName}}"
  full_name         = "{{.FullName}}"
  description       = "{{.Description}}"
  delete_force      = true
  delete_recursive  = true
}

resource "vcd_org_saml" "{{.OrgName}}" {
    org_id                          = vcd_org.{{.OrgName}}.id
    enabled                         = true
	entity_id                       = "{{.EntityId}}"
    identity_provider_metadata_file = "{{.MetadataFile}}"
}
`
