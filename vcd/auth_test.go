// +build functional auth ALL

package vcd

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// TestAccAuth aims to test out all possible ways of `provider` section configuration and allow
// authentication. It tests:
// * local system username and password auth with default org and vdc
// * local system username and password auth with default org
// * local system username and password auth
// * Saml username and password auth (if testConfig.Provider.UseSamlAdfs is true)
// * token based authentication
// * token based authentication priority over username and password
// * token based authentication with auth_type=token
// * auth_type=saml_adfs,EmptySysOrg (if testConfig.Provider.SamlUser and
// testConfig.Provider.SamlPassword are provided)
// Note. Because this test does not use regular templateFill function - it will not generate binary
// tests, but there should be no need for them as well.
func TestAccAuth(t *testing.T) {
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	// Reset connection cache just to be sure that we are not reusing any
	cachedVCDClients.reset()

	// All other acceptance tests work by relying on environment variables being set in function
	// `getConfigStruct` to configure authentication method. However because this test wants to test
	// combinations of accepted `provider` block auth configurations we are setting it as string
	// directly in `provider` and any environment variables set need to be unset during the run of
	// this test and restored afterwards.
	envVars := newEnvVarHelper()
	envVars.saveVcdVars()
	t.Logf("Clearing all VCD env variables")
	envVars.unsetVcdVars()
	defer func() {
		t.Logf("Restoring all VCD env variables")
		envVars.restoreVcdVars()
	}()

	type authTestCase struct {
		name       string
		configText string
		skip       bool // To make subtests always show names
		skipReason string
	}
	type authTests []authTestCase

	testCases := authTests{}

	testCases = append(testCases, authTestCase{
		name:       "SystemUserAndPasswordWithDefaultOrgAndVdc",
		skip:       testConfig.Provider.UseSamlAdfs,
		skipReason: "testConfig.Provider.UseSamlAdfs must be false",
		configText: `
			provider "vcd" {
				user     = "` + testConfig.Provider.User + `"
				password = "` + testConfig.Provider.Password + `"
				sysorg   = "` + testConfig.Provider.SysOrg + `" 
				org      = "` + testConfig.VCD.Org + `"
				vdc      = "` + testConfig.VCD.Vdc + `"
				url      = "` + testConfig.Provider.Url + `"
				allow_unverified_ssl = true
			}
	  `,
	})

	testCases = append(testCases, authTestCase{
		name:       "SystemUserAndPasswordWithDefaultOrg",
		skip:       testConfig.Provider.UseSamlAdfs,
		skipReason: "testConfig.Provider.UseSamlAdfs must be false",
		configText: `
			provider "vcd" {
				user     = "` + testConfig.Provider.User + `"
				password = "` + testConfig.Provider.Password + `"
				sysorg   = "` + testConfig.Provider.SysOrg + `" 
				org      = "` + testConfig.VCD.Org + `"
				url      = "` + testConfig.Provider.Url + `"
				allow_unverified_ssl = true
			}
	  `,
	})

	testCases = append(testCases, authTestCase{
		name:       "SystemUserAndPassword,AuthType=integrated",
		skip:       testConfig.Provider.UseSamlAdfs,
		skipReason: "testConfig.Provider.UseSamlAdfs must be false",
		configText: `
			provider "vcd" {
				user      = "` + testConfig.Provider.User + `"
				password  = "` + testConfig.Provider.Password + `"
				auth_type = "integrated"
				sysorg    = "` + testConfig.Provider.SysOrg + `" 
				org       = "` + testConfig.VCD.Org + `"
				url       = "` + testConfig.Provider.Url + `"
				allow_unverified_ssl = true
			}
	  `,
	})

	testCases = append(testCases, authTestCase{
		name:       "SamlSystemUserAndPassword,AuthType=saml_adfs",
		skip:       !testConfig.Provider.UseSamlAdfs,
		skipReason: "testConfig.Provider.UseSamlAdfs must be true",
		configText: `
			provider "vcd" {
				user             = "` + testConfig.Provider.User + `"
				password         = "` + testConfig.Provider.Password + `"
				auth_type        = "saml_adfs"
				saml_adfs_rpt_id = "` + testConfig.Provider.CustomAdfsRptId + `"
				sysorg           = "` + testConfig.Provider.SysOrg + `" 
				org              = "` + testConfig.VCD.Org + `"
				vdc              = "` + testConfig.VCD.Vdc + `"
				url              = "` + testConfig.Provider.Url + `"
				allow_unverified_ssl = true
			}
	  `,
	})

	testCases = append(testCases, authTestCase{
		name: "SystemUserAndPasswordWithoutSysOrg",
		configText: `
		provider "vcd" {
		  user     = "` + testConfig.Provider.User + `"
		  password = "` + testConfig.Provider.Password + `"
		  org      = "` + testConfig.Provider.SysOrg + `" 
		  url      = "` + testConfig.Provider.Url + `"
		  allow_unverified_ssl = true
		}
	  `,
	})

	// To test token auth we must gain it at first
	tempConn := createTemporaryVCDConnection()

	testCases = append(testCases, authTestCase{
		name: "TokenAuth",
		configText: `
		provider "vcd" {
			token    = "` + tempConn.Client.VCDToken + `"
			sysorg   = "` + testConfig.Provider.SysOrg + `" 
			org      = "` + testConfig.VCD.Org + `"
			vdc      = "` + testConfig.VCD.Vdc + `"
			url      = "` + testConfig.Provider.Url + `"
			allow_unverified_ssl = true
		  }
	  `,
	})

	testCases = append(testCases, authTestCase{
		name: "TokenAuthOnly,AuthType=token",
		configText: `
		provider "vcd" {
			token    = "` + tempConn.Client.VCDToken + `"
			sysorg   = "` + testConfig.Provider.SysOrg + `" 
			org      = "` + testConfig.VCD.Org + `"
			vdc      = "` + testConfig.VCD.Vdc + `"
			url      = "` + testConfig.Provider.Url + `"
			allow_unverified_ssl = true
		  }
	  `,
	})

	testCases = append(testCases, authTestCase{
		name: "TokenPriorityOverUserAndPassword",
		configText: `
		provider "vcd" {
		  user     = "invalidUser"
		  password = "invalidPassword"
		  token    = "` + tempConn.Client.VCDToken + `"
		  sysorg   = "` + testConfig.Provider.SysOrg + `" 
		  org      = "` + testConfig.VCD.Org + `"
		  vdc      = "` + testConfig.VCD.Vdc + `"
		  url      = "` + testConfig.Provider.Url + `"
		  allow_unverified_ssl = true
		}
	  `,
	})

	testCases = append(testCases, authTestCase{
		name: "TokenWithUserAndPassword,AuthType=token",
		configText: `
		provider "vcd" {
		  auth_type = "token"
		  user      = "invalidUser"
		  password  = "invalidPassword"
		  token     = "` + tempConn.Client.VCDToken + `"
		  sysorg    = "` + testConfig.Provider.SysOrg + `" 
		  org       = "` + testConfig.VCD.Org + `"
		  vdc       = "` + testConfig.VCD.Vdc + `"
		  url       = "` + testConfig.Provider.Url + `"
		  allow_unverified_ssl = true
		}
	  `,
	})

	// auth_type=saml_adfs is only run if credentials were provided
	testCases = append(testCases, authTestCase{
		name:       "EmptySysOrg,AuthType=saml_adfs",
		skip:       (testConfig.Provider.SamlUser == "" || testConfig.Provider.SamlPassword == ""),
		skipReason: "testConfig.Provider.SamlUser and testConfig.Provider.SamlPassword must be set",
		configText: `
			provider "vcd" {
			  auth_type        = "saml_adfs"
			  user             = "` + testConfig.Provider.SamlUser + `"
			  password         = "` + testConfig.Provider.SamlPassword + `"
			  saml_adfs_rpt_id = "` + testConfig.Provider.SamlCustomRptId + `"
			  org              = "` + testConfig.VCD.Org + `"
			  vdc              = "` + testConfig.VCD.Vdc + `"
			  url              = "` + testConfig.Provider.Url + `"
			  allow_unverified_ssl = true
			}
		  `,
	})

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			if test.skip {
				t.Skip("Skipping: " + test.skipReason)
			}
			runAuthTest(t, test.configText)
		})
	}

	// Clear connection cache to force other tests use their own mechanism
	cachedVCDClients.reset()
}

func runAuthTest(t *testing.T, configText string) {

	dataSource := `
	data "vcd_org" "auth" {
		name = "` + testConfig.VCD.Org + `"
	}
	`

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText + dataSource,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.vcd_org.auth", "id"),
				),
			},
		},
	})
}
