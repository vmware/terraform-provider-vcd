//go:build org || multisite || ALL || functional

package vcd

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

func TestVcdMultisiteOrgAssociation(t *testing.T) {
	preTestChecks(t)
	org1Name := testConfig.VCD.Org
	org2Name := testConfig.VCD.Org + "-1"
	user1Name := testConfig.TestEnvBuild.OrgUser
	user1Password := testConfig.TestEnvBuild.OrgUser
	user2Name := testConfig.TestEnvBuild.OrgUser
	user2Password := testConfig.TestEnvBuild.OrgUser

	params1 := StringMap{
		"Alias":    "org1",
		"OrgDef":   "org1",
		"OrgName":  org1Name,
		"SysOrg":   org1Name,
		"UserName": user1Name,
		"VcdUrl":   testConfig.Provider.Url,
		"Password": user1Password,
	}
	params2 := StringMap{
		"Alias":    "org2",
		"OrgDef":   "org2",
		"OrgName":  org2Name,
		"SysOrg":   org2Name,
		"UserName": user2Name,
		"VcdUrl":   testConfig.Provider.Url,
		"Password": user2Password,
	}

	params1["FuncName"] = t.Name() + "-data1"
	configTextData1 := templateFill(testAccMultisiteOrgData, params1)
	params2["FuncName"] = t.Name() + "-data2"
	configTextData2 := templateFill(testAccMultisiteOrgData, params2)

	params1["FuncName"] = t.Name() + "-association1"
	params1["OtherOrgXml"] = "org2"
	params1["OrgAssociation"] = "org1-org2"
	configTextAssociation1 := templateFill(testAccMultisiteOrgAssociation, params1)
	params2["FuncName"] = t.Name() + "-association2"
	params2["OtherOrgXml"] = "org1"
	params1["OrgAssociation"] = "org2-org1"
	configTextAssociation2 := templateFill(testAccMultisiteOrgAssociation, params2)

	debugPrintf("#[DEBUG] CONFIGURATION DATA 1: %s", configTextData1)
	debugPrintf("#[DEBUG] CONFIGURATION DATA 2: %s", configTextData2)
	debugPrintf("#[DEBUG] CONFIGURATION Association 1: %s", configTextAssociation1)
	debugPrintf("#[DEBUG] CONFIGURATION Association 2: %s", configTextAssociation2)
	resource.Test(t, resource.TestCase{
		ProviderFactories: buildMultipleProviders(),
		CheckDestroy:      resource.ComposeTestCheckFunc(),
		Steps: []resource.TestStep{
			// extracting data org 1
			{
				Config: configTextData1,
			},
			// extracting data org 2
			{
				Config: configTextData2,
			},
			// associating org 1 with org 2
			{
				Config: configTextAssociation1,
			},
			// associating org 2 with org 1
			{
				Config: configTextAssociation2,
			},
		},
	})

	postTestChecks(t)
}

const testAccMultisiteOrgData = `
{{.SkipNotice}}

//provider "vcd" {
//  alias                = "{{.Alias}}"
//  user                 = "{{.UserName}}"
//  password             = "{{.Password}}"
//  token                = ""
//  api_token            = ""
//  auth_type            = "integrated"
//  saml_adfs_rpt_id     = ""
//  url                  = "{{.VcdUrl}}"
//  sysorg               = "{{.SysOrg}}"
//  org                  = "{{.OrgName}}"
//  allow_unverified_ssl = "true"
//  max_retry_timeout    = 600
//  logging              = true
//  logging_file         = "go-vcloud-director-{{.Alias}}.log"
//}

data "vcd_org" "{{.OrgDef}}" {
  name = "{{.OrgName}}"
}

data "vcd_multisite_org_data" "{{.OrgDef}}-data" {
  provider         = vcd.{{.Alias}}
  org_id           = data.vcd_org.{{.OrgDef}}.id
  download_to_file = "{{.Alias}}.xml"
}
`

const testAccMultisiteOrgAssociation = `
//provider "vcd" {
//  alias                = "{{.Alias}}"
//  user                 = "{{.UserName}}"
//  password             = "{{.Password}}"
//  token                = ""
//  api_token            = ""
//  auth_type            = "integrated"
//  saml_adfs_rpt_id     = ""
//  url                  = "{{.VcdUrl}}"
//  sysorg               = "{{.SysOrg}}"
//  org                  = "{{.OrgName}}"
//  allow_unverified_ssl = "true"
//  max_retry_timeout    = 600
//  logging              = true
//  logging_file         = "go-vcloud-director-{{.Alias}}.log"
//}

resource "vcd_multisite_org_association" "{{.OrgAssociation}}" {
  provider                = vcd.{{.Alias}}
  association_data_file   = "{{.OtherOrgXml}}.xml"
  connection_timeout_mins = 2
}
`
