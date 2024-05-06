//go:build org || ALL || functional

package vcd

import (
	_ "embed"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"net/url"
	"testing"
)

func TestAccVcdOrgOidc(t *testing.T) {
	preTestChecks(t)

	if testConfig.Networking.OidcServer.Url == "" ||
		testConfig.Networking.OidcServer.WellKnownEndpoint == "" {
		t.Skip(t.Name(), "requires OIDC Server URL and its well-known endpoint")
	}

	oidcServer, err := url.Parse(testConfig.Networking.OidcServer.Url)
	if err != nil {
		t.Skip(t.Name(), "requires a valid OIDC Server URL but got ", err)
	}
	oidcServer = oidcServer.JoinPath(testConfig.Networking.OidcServer.WellKnownEndpoint)

	orgName1 := t.Name() + "1"
	orgName2 := t.Name() + "2"
	oidcResource1 := "vcd_org_oidc.oidc1"
	//oidcResource2 := "vcd_org_oidc.oidc2"

	var params = StringMap{
		"OrgName1":          orgName1,
		"OrgName2":          orgName2,
		"WellKnownEndpoint": oidcServer.String(),
	}
	testParamsNotEmpty(t, params)

	skipIfNotSysAdmin(t)

	configText := templateFill(testAccCheckVcdOrgOidc, params)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckVcdOrgExists(orgName1),
			testAccCheckVcdOrgExists(orgName2),
		),
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdOrgExists("vcd_org.org1"),
					testAccCheckVcdOrgExists("vcd_org.org2"),
				),
			},
			{
				ResourceName:            oidcResource1,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdFunc:       importStateIdTopHierarchy(orgName1),
				ImportStateVerifyIgnore: []string{},
			},
		},
	})

	postTestChecks(t)
}

const testAccCheckVcdOrgOidc = `
resource "vcd_org" "org1" {
  name              = "{{.OrgName1}}"
  full_name         = "{{.OrgName1}}"
  description       = "{{.OrgName1}}"
  delete_force      = true
  delete_recursive  = true
}

resource "vcd_org" "org2" {
  name              = "{{.OrgName2}}"
  full_name         = "{{.OrgName2}}"
  description       = "{{.OrgName2}}"
  delete_force      = true
  delete_recursive  = true
}

resource "vcd_org_oidc" "oidc1" {
  org_id                      = vcd_org.org1.id
  enabled                     = true
  prefer_id_token             = false
  client_id                   = "clientId"
  client_secret               = "clientSecret"
  max_clock_skew_seconds      = 60
  wellknown_endpoint          = "{{.WellKnownEndpoint}}"
}
`
