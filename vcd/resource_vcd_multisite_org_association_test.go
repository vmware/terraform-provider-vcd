//go:build org || multisite || ALL || functional

package vcd

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"os"
	"testing"
)

func TestVcdMultisiteOrgAssociation(t *testing.T) {
	preTestChecks(t)
	org1Name := testConfig.VCD.Org
	org2Name := testConfig.VCD.Org + "-1"
	org1Xml := org1Name + ".xml"
	org2Xml := org2Name + ".xml"

	params := StringMap{
		"Org1Def":           "org1",
		"Org2Def":           "org2",
		"Org1Name":          org1Name,
		"Org2Name":          org2Name,
		"ProviderVcdSystem": providerVcdSystem,
		"ProviderVcdOrg1":   providerVcdOrg1,
		"ProviderVcdOrg2":   providerVcdOrg2,
		"Org1Association":   "org1-org2",
		"Org2Association":   "org2-org1",
		"SkipNotice":        " ",
	}

	params["FuncName"] = t.Name() + "-data"
	configTextData := templateFill(testAccMultisiteOrgCommon+testAccMultisiteOrgData, params)
	params["FuncName"] = t.Name() + "-association"
	configTextAssociation := templateFill(testAccMultisiteOrgCommon+testAccMultisiteOrgAssociation, params)

	debugPrintf("#[DEBUG] CONFIGURATION DATA: %s", configTextData)
	debugPrintf("#[DEBUG] CONFIGURATION Association: %s", configTextAssociation)
	defer func() {
		for _, fName := range []string{org1Xml, org2Xml} {
			if fileExists(fName) {
				if err := os.Remove(fName); err != nil {
					fmt.Printf("error removing file %s: %s\n", fName, err)
				}
			}
		}
	}()
	resource.Test(t, resource.TestCase{
		ProviderFactories: buildMultipleProviders(),
		CheckDestroy:      resource.ComposeTestCheckFunc(),
		Steps: []resource.TestStep{
			// extracting data from org1 and org2
			{
				Config: configTextData,
			},

			// associating org1 with org2 and org2 with org1
			{
				Config: configTextAssociation,
			},
		},
	})

	postTestChecks(t)
}

const testAccMultisiteOrgCommon = `
data "vcd_resource_list" "orgs" {
  provider      = {{.ProviderVcdSystem}}
  name          = "orgs"
  resource_type = "vcd_org"
}

data "vcd_org" "{{.Org1Def}}" {
  provider = {{.ProviderVcdOrg1}}
  name     = "{{.Org1Name}}"
}

data "vcd_org" "{{.Org2Def}}" {
  provider = {{.ProviderVcdOrg2}}
  name     = "{{.Org2Name}}"
}
`

const testAccMultisiteOrgData = `
{{.SkipNotice}}

data "vcd_multisite_org_data" "{{.Org1Def}}-data" {
  provider         = {{.ProviderVcdOrg1}}
  org_id           = data.vcd_org.{{.Org1Def}}.id
  download_to_file = "{{.Org1Name}}.xml"
}


data "vcd_multisite_org_data" "{{.Org2Def}}-data" {
  provider         = {{.ProviderVcdOrg2}}
  org_id           = data.vcd_org.{{.Org2Def}}.id
  download_to_file = "{{.Org2Name}}.xml"
}
`

const testAccMultisiteOrgAssociation = `
{{.SkipNotice}}

resource "vcd_multisite_org_association" "{{.Org1Association}}" {
  provider                = {{.ProviderVcdOrg1}}
  org_id                  = data.vcd_org.{{.Org1Def}}.id
  association_data_file   = "{{.Org2Name}}.xml"
  connection_timeout_mins = 2
}

resource "vcd_multisite_org_association" "{{.Org2Association}}" {
  provider                = {{.ProviderVcdOrg2}}
  org_id                  = data.vcd_org.{{.Org2Def}}.id
  association_data_file   = "{{.Org1Name}}.xml"
  connection_timeout_mins = 2
}
`
