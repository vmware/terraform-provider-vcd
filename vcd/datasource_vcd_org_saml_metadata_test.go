//go:build org || ALL || functional

package vcd

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/util"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcdDatasourceOrgSamlMetadata(t *testing.T) {
	preTestChecks(t)

	skipIfNotSysAdmin(t)

	orgName1 := testConfig.VCD.Org
	metadataFileName := t.Name() + ".xml"
	var params = StringMap{
		"FuncName":         "TestAccVcdDatasourceOrg",
		"OrgName1":         orgName1,
		"Tags":             "org",
		"MetadataFileName": metadataFileName,
	}
	testParamsNotEmpty(t, params)

	configText := templateFill(testAccCheckVcdDatasourceOrgSamlMetadata, params)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

	vcdUrl := testConfig.Provider.Url
	vcdUrl = strings.Replace(vcdUrl, "/api", "", 1)
	datasource1 := "data.vcd_org." + orgName1

	defer func() {
		if fileExists(metadataFileName) {
			err := os.Remove(metadataFileName)
			if err != nil {
				util.Logger.Printf("[ERROR - TestAccVcdDatasourceOrg] error removing metadata file %s: %s\n", metadataFileName, err)
			}
		}
	}()
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				PreConfig: func() {
					if fileExists(metadataFileName) {
						_ = os.Remove(metadataFileName)
					}
				},
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdOrgExists(datasource1),
					// The retrieved metadata should contain both the VCD URL and the name of the target organization`
					metadataFileCheck(metadataFileName, []string{vcdUrl, orgName1}),
				),
			},
		},
	})
	postTestChecks(t)
}

// metadataFileCheck makes sure that:
// * the metadata file exists
// * the contents are validated
// * the metadata text contains all elements listed in 'wanted'
func metadataFileCheck(fileName string, wanted []string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		if !fileExists(fileName) {
			return fmt.Errorf("metadata file %s not found", fileName)
		}

		metadataText, err := os.ReadFile(fileName)
		if err != nil {
			return err
		}

		errors := govcd.ValidateSamlServiceProviderMetadata(string(metadataText))
		if errors != nil {
			return fmt.Errorf("error validating metadata file: %s", govcd.GetErrorMessageFromErrorSlice(errors))
		}
		var missing []string
		for _, w := range wanted {
			if !strings.Contains(string(metadataText), w) {
				missing = append(missing, w)
			}
		}
		if len(missing) > 0 {
			return fmt.Errorf("wanted pieces in metadata not found: %v", missing)
		}
		return nil
	}
}

const testAccCheckVcdDatasourceOrgSamlMetadata = `
data "vcd_org" "{{.OrgName1}}" {
  name = "{{.OrgName1}}"
}

data "vcd_org_saml_metadata" "{{.OrgName1}}" {
  org_id    = data.vcd_org.{{.OrgName1}}.id
  file_name = "{{.MetadataFileName}}"
}
`
