//go:build org || multisite || ALL || functional

package vcd

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"os"
	"regexp"
	"strings"
	"testing"
)

func checkClientConnectionFromEnv() error {
	vcdUrl := os.Getenv(envSecondVcdUrl)
	user := os.Getenv(envSecondVcdUser)
	password := os.Getenv(envSecondVcdPassword)
	orgName := os.Getenv(envSecondVcdSysOrg)

	var missing []string
	if vcdUrl == "" {
		missing = append(missing, envSecondVcdUrl)
	}
	if user == "" {
		missing = append(missing, envSecondVcdUser)
	}
	if password == "" {
		missing = append(missing, envSecondVcdPassword)
	}
	if orgName == "" {
		missing = append(missing, envSecondVcdSysOrg)
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing environment variables for connection: %v", missing)
	}
	if !strings.HasSuffix(vcdUrl, "/api") {
		return fmt.Errorf("the VCD URL must terminate with '/api'")
	}
	return nil
}

/*
	TestVcdMultisiteSiteAssociation will test the associations between two sites
	To run this test, make a shell script like the one below, filling the variables
	in addition to the VCD defined in vcd_test_config.json

$ cat connection.sh
export VCD_URL2=https://some-vcd-url.com/api
export VCD_USER2=administrator
export VCD_PASSWORD2='myPassword'
export VCD_SYSORG2=System
export VCD_ORG2=orgname2
export VCD_ORGUSER2=org-admin-name
export VCD_ORGUSER_PASSWORD2='myOrgAdminPassword'

$	source connection.sh
$ go test -tags functional -run TestVcdMultisiteSiteAssociation  -v -timeout 0
*/

func TestVcdMultisiteSiteAssociation(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)
	err := checkClientConnectionFromEnv()
	if err != nil {
		t.Skipf("second connection not available: %s", err)
	}
	secondVcdUrl := os.Getenv(envSecondVcdUrl)
	secondVcdSysorg := os.Getenv(envSecondVcdSysOrg)
	secondVcdUser := os.Getenv(envSecondVcdUser)
	secondVcdPassword := os.Getenv(envSecondVcdPassword)
	site1XmlName := "site1.xml"
	site2XmlName := "site2.xml"

	params := StringMap{
		"FuncName":         t.Name() + "-step1",
		"Site1XmlName":     site1XmlName,
		"Site2XmlName":     site2XmlName,
		"VcdSystem2":       providerVcdSystem2,
		"TimeoutMins":      "0",
		"SkipNotice":       " ",
		"Site1Association": "site1-site2",
		"Site2Association": "site2-site1",
	}

	configTextData := templateFill(testAccMultisiteSiteData, params)
	params["FuncName"] = t.Name() + "-step2"
	params["SkipNotice"] = "# skip-binary-test: can't persist XML files across different scripts"
	configTextAssociation := templateFill(testAccMultisiteSiteAssociation, params)
	params["FuncName"] = t.Name() + "-step3"
	params["TimeoutMins"] = "2"
	configTextAssociationUpdate := templateFill(testAccMultisiteSiteAssociation+testAccMultisiteSiteDS, params)

	debugPrintf("#[DEBUG] CONFIGURATION DATA: %s", configTextData)
	debugPrintf("#[DEBUG] CONFIGURATION association 1: %s", configTextAssociation)
	debugPrintf("#[DEBUG] CONFIGURATION association 2: %s", configTextAssociationUpdate)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	defer func() {
		// Remove XML files, if they were left behind
		for _, fName := range []string{site1XmlName, site2XmlName} {
			if fileExists(fName) {
				if err := os.Remove(fName); err != nil {
					fmt.Printf("error removing file %s: %s\n", fName, err)
				}
			}
		}
	}()
	resource.Test(t, resource.TestCase{
		ProviderFactories: buildMultipleSysProviders(secondVcdUrl, secondVcdUser, secondVcdPassword, secondVcdSysorg),
		CheckDestroy:      resource.ComposeTestCheckFunc(),
		Steps: []resource.TestStep{
			// extracting data
			{
				Config: configTextData,
			},
			{
				Config: configTextAssociation,
				Check: resource.ComposeAggregateTestCheckFunc(
					// The status, depending on the operation speed, could be either 'ACTIVE' or 'ASYMMETRIC'
					resource.TestMatchResourceAttr("vcd_multisite_site_association.site1-site2",
						"status", regexp.MustCompilePOSIX(string(types.StatusAsymmetric)+`|`+string(types.StatusActive))),
				),
			},
			{
				Config: configTextAssociationUpdate,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"vcd_multisite_site_association.site1-site2", "associated_site_id",
						"data.vcd_multisite_site.remote", "id",
					),
					resource.TestCheckResourceAttrPair(
						"vcd_multisite_site_association.site2-site1", "associated_site_id",
						"data.vcd_multisite_site.local", "id",
					),
					// After the mandatory check (connection_timeout_mins=2), the status must be 'ACTIVE'
					resource.TestCheckResourceAttr("vcd_multisite_site_association.site1-site2", "status", string(types.StatusActive)),
					resource.TestCheckResourceAttr("vcd_multisite_site_association.site2-site1", "status", string(types.StatusActive)),
				),
			},
			{
				ResourceName:            "vcd_multisite_site_association.site1-site2",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdFunc:       importSiteStateIdFromXmlFile(site2XmlName),
				ImportStateVerifyIgnore: []string{"association_data_file", "connection_timeout_mins"},
			},
			{
				ResourceName:            "vcd_multisite_site_association.site2-site1",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdFunc:       importSiteStateIdFromXmlFile(site1XmlName),
				ImportStateVerifyIgnore: []string{"association_data_file", "connection_timeout_mins"},
			},
		},
	})

	postTestChecks(t)
}

func importSiteStateIdFromXmlFile(fileName string) resource.ImportStateIdFunc {
	return func(*terraform.State) (string, error) {
		result, err := govcd.ReadXmlDataFromFile[types.SiteAssociationMember](fileName)
		if err != nil {
			return "", fmt.Errorf("error getting %T from file %s: %s", types.SiteAssociationMember{}, fileName, err)
		}
		return result.SiteID, nil
	}
}

const testAccMultisiteSiteData = `
{{.SkipNotice}}

data "vcd_multisite_site_data" "site1-data" {
  provider         = vcd
  download_to_file = "{{.Site1XmlName}}"
}

data "vcd_multisite_site_data" "site2-data" {
  provider         = {{.VcdSystem2}}
  download_to_file = "{{.Site2XmlName}}"
}

`
const testAccMultisiteSiteAssociation = `
{{.SkipNotice}}

resource "vcd_multisite_site_association" "{{.Site1Association}}" {
  provider                = vcd
  association_data_file   = "{{.Site2XmlName}}"
  connection_timeout_mins = {{.TimeoutMins}}
}

resource "vcd_multisite_site_association" "{{.Site2Association}}" {
  provider                = {{.VcdSystem2}}
  association_data_file   = "{{.Site1XmlName}}"
  connection_timeout_mins = {{.TimeoutMins}}
}
`

const testAccMultisiteSiteDS = `

data "vcd_multisite_site" "local" {
}

data "vcd_multisite_site" "remote" {
  provider = {{.VcdSystem2}}
}
`
