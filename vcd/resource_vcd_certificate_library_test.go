//go:build certificate || ALL || functional
// +build certificate ALL functional

package vcd

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// TestAccVcdCertificateInLibraryResource tests that certificate can add to library
func TestAccVcdCertificateInLibraryResource(t *testing.T) {
	preTestChecks(t)

	// This test requires access to the vCD before filling templates
	// Thus it won't run in the short test
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	vcdClient := createTemporaryVCDConnection()
	if vcdClient.Client.APIVCDMaxVersionIs("< 35.0") {
		t.Skip(t.Name() + " requires at least API v35.0 (vCD 10.2+)")
	}

	// String map to fill the template
	var params = StringMap{
		"Org":                      testConfig.VCD.Org,
		"Alias":                    "TestAccVcdCertificateInLibraryResource",
		"AliasUpdate":              "TestAccVcdCertificateInLibraryResourceUpdated",
		"AliasSystem":              "TestAccVcdCertificateInLibraryResourceSys",
		"AliasPrivate":             "TestAccVcdCertificateInLibraryResourcePrivate",
		"AliasPrivateSystem":       "TestAccVcdCertificateInLibraryResourcePrivateSys",
		"AliasPrivateSystemUpdate": "TestAccVcdCertificateInLibraryResourcePrivateSysUpdated",
		"Certificate1Path":         testConfig.Certificates.Certificate1Path,
		"Certificate2Path":         testConfig.Certificates.Certificate2Path,
		"PrivateKey2":              testConfig.Certificates.Certificate2PrivateKeyPath,
		"PassPhrase":               testConfig.Certificates.Certificate2Pass,
		"Description1":             "myDescription 1",
		"Description1Update":       "myDescription 1 updated",
		"Description2":             "myDescription 2",
		"Description3":             "myDescription 3",
		"Description4":             "myDescription 4",
		"Description4Update":       "myDescription 4 updated",
	}

	if !vcdClient.Client.IsSysAdmin {
		t.Skip(t.Name() + " only System Administrator can add Certificates")
	}

	configText1 := templateFill(testAccVcdCertificateInLibraryResource, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)

	params["FuncName"] = t.Name() + "-update"
	configText2 := templateFill(testAccVcdCertificateInLibraryResourceUpdate, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText2)

	resourceAddressOrgCert := "vcd_certificate_in_library.orgCertificate"
	resourceAddressOrgPrivateCert := "vcd_certificate_in_library.OrgWithPrivateCertificate"
	resourceAddressSysCert := "vcd_certificate_in_library.sysCertificate"
	resourceAddressSysPrivateCert := "vcd_certificate_in_library.sysCertificateWithPrivate"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		PreCheck:          func() { testAccPreCheck(t) },

		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddressOrgCert, "alias", params["Alias"].(string)),
					resource.TestMatchResourceAttr(resourceAddressOrgCert, "id", regexp.MustCompile(`^\S+`)),
					resource.TestCheckResourceAttr(resourceAddressOrgCert, "description", params["Description1"].(string)),
					resource.TestMatchResourceAttr(resourceAddressOrgCert, "certificate", regexp.MustCompile(`^\S+`)),
					resource.TestCheckResourceAttr(resourceAddressOrgPrivateCert, "alias", params["AliasPrivate"].(string)),
					resource.TestMatchResourceAttr(resourceAddressOrgPrivateCert, "id", regexp.MustCompile(`^\S+`)),
					resource.TestCheckResourceAttr(resourceAddressOrgPrivateCert, "description", params["Description2"].(string)),
					resource.TestMatchResourceAttr(resourceAddressOrgPrivateCert, "certificate", regexp.MustCompile(`^\S+`)),
					resource.TestCheckResourceAttr(resourceAddressSysCert, "alias", params["AliasSystem"].(string)),
					resource.TestMatchResourceAttr(resourceAddressSysCert, "id", regexp.MustCompile(`^\S+`)),
					resource.TestCheckResourceAttr(resourceAddressSysCert, "description", params["Description3"].(string)),
					resource.TestMatchResourceAttr(resourceAddressSysCert, "certificate", regexp.MustCompile(`^\S+`)),
					resource.TestCheckResourceAttr(resourceAddressSysPrivateCert, "alias", params["AliasPrivateSystem"].(string)),
					resource.TestMatchResourceAttr(resourceAddressSysPrivateCert, "id", regexp.MustCompile(`^\S+`)),
					resource.TestCheckResourceAttr(resourceAddressSysPrivateCert, "description", params["Description4"].(string)),
					resource.TestMatchResourceAttr(resourceAddressSysPrivateCert, "certificate", regexp.MustCompile(`^\S+`)),
				),
			},
			resource.TestStep{
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddressOrgCert, "alias", params["AliasUpdate"].(string)),
					resource.TestMatchResourceAttr(resourceAddressOrgCert, "id", regexp.MustCompile(`^\S+`)),
					resource.TestCheckResourceAttr(resourceAddressOrgCert, "description", params["Description1Update"].(string)),
					resource.TestMatchResourceAttr(resourceAddressOrgCert, "certificate", regexp.MustCompile(`^\S+`)),
					resource.TestCheckResourceAttr(resourceAddressSysPrivateCert, "alias", params["AliasPrivateSystemUpdate"].(string)),
					resource.TestMatchResourceAttr(resourceAddressSysPrivateCert, "id", regexp.MustCompile(`^\S+`)),
					resource.TestCheckResourceAttr(resourceAddressSysPrivateCert, "description", params["Description4Update"].(string)),
					resource.TestMatchResourceAttr(resourceAddressSysPrivateCert, "certificate", regexp.MustCompile(`^\S+`)),
				),
			},
			resource.TestStep{
				ResourceName:      resourceAddressOrgCert,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdOrgObject(testConfig, params["AliasUpdate"].(string)),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdCertificateInLibraryResource = `
resource "vcd_certificate_in_library" "orgCertificate" {
  org         = "{{.Org}}"
  alias       = "{{.Alias}}"
  description = "{{.Description1}}"
  certificate = file("{{.Certificate1Path}}")
}

resource "vcd_certificate_in_library" "OrgWithPrivateCertificate" {
  org                    = "{{.Org}}"
  alias                  = "{{.AliasPrivate}}"
  description            = "{{.Description2}}"
  certificate            = file("{{.Certificate2Path}}")
  private_key            = file("{{.PrivateKey2}}")
  private_key_passphrase = "{{.PassPhrase}}"
}

resource "vcd_certificate_in_library" "sysCertificate" {
  org         = "System"
  alias       = "{{.AliasSystem}}"
  description = "{{.Description3}}"
  certificate = file("{{.Certificate1Path}}")
}

resource "vcd_certificate_in_library" "sysCertificateWithPrivate" {
  org                    = "System"
  alias                  = "{{.AliasPrivateSystem}}"
  description            = "{{.Description4}}"
  certificate            = file("{{.Certificate2Path}}")
  private_key            = file("{{.PrivateKey2}}")
  private_key_passphrase = "{{.PassPhrase}}"
}
`

const testAccVcdCertificateInLibraryResourceUpdate = `
resource "vcd_certificate_in_library" "orgCertificate" {
  org         = "{{.Org}}"
  alias       = "{{.AliasUpdate}}"
  description = "{{.Description1Update}}"
  certificate = file("{{.Certificate1Path}}")
}

resource "vcd_certificate_in_library" "sysCertificateWithPrivate" {
  org                    = "System"
  alias                  = "{{.AliasPrivateSystemUpdate}}"
  description            = "{{.Description4Update}}"
  certificate            = file("{{.Certificate2Path}}")
  private_key            = file("{{.PrivateKey2}}")
  private_key_passphrase = "{{.PassPhrase}}"
}
`
