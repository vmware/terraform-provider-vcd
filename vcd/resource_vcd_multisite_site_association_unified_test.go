//go:build org || multisite || ALL || functional

package vcd

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"os"
	"path"
	"regexp"
	"testing"
)

/*
	TestVcdMultisiteSiteAssociationUnified will test the associations between two sites
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

func TestVcdMultisiteSiteAssociationUnified(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)
	err := checkClientConnectionFromEnv()
	if err != nil {
		t.Skipf("second connection not available: %s", err)
	}

	tempDir := os.TempDir()
	site1XmlName := path.Join(tempDir, "site1.xml")
	site2XmlName := path.Join(tempDir, "site2.xml")

	params := StringMap{
		"Site1XmlName":     site1XmlName,
		"Site2XmlName":     site2XmlName,
		"VcdSystem2":       providerVcdSystem2,
		"Site1Association": "site1-site2",
		"Site2Association": "site2-site1",
	}

	configText := templateFill(testAccMultisiteSiteUnited, params)

	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

	secondVcdUrl := os.Getenv(envSecondVcdUrl)
	secondVcdSysorg := os.Getenv(envSecondVcdSysOrg)
	secondVcdUser := os.Getenv(envSecondVcdUser)
	secondVcdPassword := os.Getenv(envSecondVcdPassword)
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
		ExternalProviders: map[string]resource.ExternalProvider{
			"local": {
				Source:            "hashicorp/local",
				VersionConstraint: "2.4.0",
			},
		},
		CheckDestroy: resource.ComposeTestCheckFunc(),
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check: resource.ComposeAggregateTestCheckFunc(
					// The status, depending on the operation speed, could be either 'ACTIVE' or 'ASYMMETRIC'
					resource.TestMatchResourceAttr("vcd_multisite_site_association.site1-site2",
						"status", regexp.MustCompilePOSIX(string(types.StatusAsymmetric)+`|`+string(types.StatusActive))),
				),
			},
		},
	})

	postTestChecks(t)
}

const testAccMultisiteSiteUnited = `

data "vcd_multisite_site_data" "site1-data" {
  provider         = vcd
  download_to_file = "{{.Site1XmlName}}"
}

data "vcd_multisite_site_data" "site2-data" {
  provider         = {{.VcdSystem2}}
  download_to_file = "{{.Site2XmlName}}"
}

data "local_file" "site1_data" {
  filename =  "{{.Site1XmlName}}"
  depends_on = [data.vcd_multisite_site_data.site1-data]
}

data "local_file" "site2_data" {
  filename =  "{{.Site2XmlName}}"
  depends_on = [data.vcd_multisite_site_data.site2-data]
}

resource "vcd_multisite_site_association" "{{.Site1Association}}" {
  provider                = vcd
  association_data        = data.local_file.site2_data.content
}

resource "vcd_multisite_site_association" "{{.Site2Association}}" {
  provider                = {{.VcdSystem2}}
  association_data        = data.local_file.site1_data.content
}
`
