// +build auth ALL functional

package vcd

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// TestAccVcdSamlAuth tests a simple operation using SAML auth when explicit SAML testing user and
// password are specified.
// Note. The test cannot be run in parallel because it temporarily overrides authentication cache
// and credentials for the purpose of its run. It restores them at the end.
func TestAccVcdSamlAuth(t *testing.T) {

	// Skip test if explicit SAML credentials are not specified
	if testConfig.Provider.SamlUser == "" || testConfig.Provider.SamlPassword == "" {
		t.Skip(t.Name() + " requires 'samlUser' and 'samlPassword' to be specified in configuration")
		return
	}
	// Backup default authentication configuration
	backupUser := testConfig.Provider.User
	backupPassword := testConfig.Provider.Password
	backupToken := testConfig.Provider.Token
	backupUseSamlAdfs := testConfig.Provider.UseSamlAdfs
	backupCustomRptId := testConfig.Provider.CustomAdfsRptId
	backupSysOrg := testConfig.Provider.SysOrg
	backupEnableConnectionCache := enableConnectionCache
	enableConnectionCache = false
	cachedVCDClients.reset()

	t.Logf("Clearing connection cache and switching credentials to: user=%s, use_saml_adfs=%t, rpt_id=%s, org=%s\n",
		testConfig.Provider.SamlUser, true, testConfig.Provider.SamlCustomRptId, testConfig.VCD.Org)

	// Temporarily override user and password variables
	_ = os.Setenv("VCD_USER", testConfig.Provider.SamlUser)
	_ = os.Setenv("VCD_PASSWORD", testConfig.Provider.SamlPassword)
	_ = os.Unsetenv("VCD_TOKEN")
	_ = os.Setenv("VCD_AUTH_TYPE", "saml_adfs")
	_ = os.Setenv("VCD_SAML_ADFS_RPT_ID", testConfig.Provider.SamlCustomRptId)
	_ = os.Setenv("VCD_SYS_ORG", testConfig.VCD.Org)

	testConfig.Provider.User = testConfig.Provider.SamlUser
	testConfig.Provider.Password = testConfig.Provider.SamlPassword
	testConfig.Provider.Token = ""
	testConfig.Provider.UseSamlAdfs = true
	testConfig.Provider.CustomAdfsRptId = testConfig.Provider.SamlCustomRptId
	testConfig.Provider.SysOrg = testConfig.VCD.Org

	// Defer restore of configuration variables and reset connection cache to force new
	// authentication after this test run
	defer func() {
		t.Logf("Clearing connection cache and restoring credentials: user=%s, use_saml_adfs=%t, rpt_id=%s, org=%s\n",
			backupUser, backupUseSamlAdfs, backupCustomRptId, testConfig.Provider.SysOrg)
		testConfig.Provider.User = backupUser
		testConfig.Provider.Password = backupPassword
		testConfig.Provider.Token = backupToken
		testConfig.Provider.UseSamlAdfs = backupUseSamlAdfs
		testConfig.Provider.CustomAdfsRptId = backupCustomRptId
		testConfig.Provider.SysOrg = backupSysOrg
		_ = os.Setenv("VCD_USER", testConfig.Provider.User)
		_ = os.Setenv("VCD_PASSWORD", testConfig.Provider.Password)
		_ = os.Setenv("VCD_TOKEN", testConfig.Provider.Token)

		if testConfig.Provider.UseSamlAdfs {
			_ = os.Setenv("VCD_AUTH_TYPE", "saml_adfs")
		} else {
			_ = os.Unsetenv("VCD_AUTH_TYPE")
		}

		_ = os.Setenv("VCD_SAML_ADFS_RPT_ID", testConfig.Provider.CustomAdfsRptId)
		_ = os.Setenv("VCD_SYS_ORG", testConfig.Provider.SysOrg)
		enableConnectionCache = backupEnableConnectionCache
		cachedVCDClients.reset()
	}()

	var params = StringMap{
		"OrgName": testConfig.VCD.Org,
		"Tags":    "auth",
	}

	configText := templateFill(testAccCheckVcdOrg, params)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.vcd_org.auth", "id"),
				),
			},
		},
	})
}

const testAccCheckVcdOrg = `
data "vcd_org" "auth" {
  name = "{{.OrgName}}"
}
`
