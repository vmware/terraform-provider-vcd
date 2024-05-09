//go:build org || ALL || functional

package vcd

import (
	_ "embed"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"net/url"
	"regexp"
	"testing"
)

func TestAccVcdOrgOidc(t *testing.T) {
	preTestChecks(t)
	oidcServerUrl := validateAndGetOidcServerUrl(t, testConfig)

	orgName1 := t.Name() + "1"
	orgName2 := t.Name() + "2"
	orgName3 := t.Name() + "3"
	oidcResource1 := "vcd_org_oidc.oidc1"
	oidcResource2 := "vcd_org_oidc.oidc2"
	oidcResource3 := "vcd_org_oidc.oidc3"
	oidcData := "data.vcd_org_oidc.oidc_data"

	var params = StringMap{
		"OrgName1":          orgName1,
		"OrgName2":          orgName2,
		"OrgName3":          orgName3,
		"WellKnownEndpoint": oidcServerUrl.String(),
		"FuncName":          t.Name() + "-Step1",
		"PreferIdToken":     " ",
		"UIButtonLabel":     " ",
	}
	client := createSystemTemporaryVCDConnection()
	if client.Client.APIVCDMaxVersionIs(">= 37.1") {
		params["PreferIdToken"] = "prefer_id_token             = true"
	}
	if client.Client.APIVCDMaxVersionIs(">= 38.1") {
		params["UIButtonLabel"] = "ui_button_label             = \"this is a test\""
	}
	testParamsNotEmpty(t, params)

	skipIfNotSysAdmin(t)

	step1 := templateFill(testAccCheckVcdOrgOidc, params)
	params["FuncName"] = t.Name() + "-Step2"
	step2 := templateFill(testAccCheckVcdOrgOidc2, params)
	params["FuncName"] = t.Name() + "-Step3"
	step3 := templateFill(testAccCheckVcdOrgOidc3, params)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	debugPrintf("#[DEBUG] Configuration Step 1: %s", step1)
	debugPrintf("#[DEBUG] Configuration Step 2: %s", step2)
	debugPrintf("#[DEBUG] Configuration Step 3: %s", step3)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckOrgDestroy(orgName1),
			testAccCheckOrgDestroy(orgName2),
			testAccCheckOrgDestroy(orgName3),
		),
		Steps: []resource.TestStep{
			{
				Config: step1,
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckVcdOrgExists("vcd_org.org1"),
					testAccCheckVcdOrgExists("vcd_org.org2"),
					testAccCheckVcdOrgExists("vcd_org.org3"),

					resource.TestMatchResourceAttr(oidcResource1, "redirect_uri", regexp.MustCompile(fmt.Sprintf(".*=tenant:%s", orgName1))),
					resource.TestCheckResourceAttr(oidcResource1, "client_id", "clientId"),
					resource.TestCheckResourceAttr(oidcResource1, "client_secret", "clientSecret"),
					resource.TestCheckResourceAttr(oidcResource1, "enabled", "true"),
					resource.TestCheckResourceAttr(oidcResource1, "wellknown_endpoint", params["WellKnownEndpoint"].(string)),
					resource.TestMatchResourceAttr(oidcResource1, "issuer_id", regexp.MustCompile(fmt.Sprintf("^%s://%s.*$", oidcServerUrl.Scheme, oidcServerUrl.Host))),
					resource.TestMatchResourceAttr(oidcResource1, "user_authorization_endpoint", regexp.MustCompile(fmt.Sprintf("^%s://%s.*$", oidcServerUrl.Scheme, oidcServerUrl.Host))),
					resource.TestMatchResourceAttr(oidcResource1, "access_token_endpoint", regexp.MustCompile(fmt.Sprintf("^%s://%s.*$", oidcServerUrl.Scheme, oidcServerUrl.Host))),
					resource.TestMatchResourceAttr(oidcResource1, "userinfo_endpoint", regexp.MustCompile(fmt.Sprintf("^%s://%s.*$", oidcServerUrl.Scheme, oidcServerUrl.Host))),
					testMatchResourceAttrWhenVersionMatches(oidcResource1, "prefer_id_token", regexp.MustCompile("^true$"), ">= 37.1"),
					resource.TestCheckResourceAttr(oidcResource1, "max_clock_skew_seconds", "60"),
					resource.TestMatchResourceAttr(oidcResource1, "scopes.#", regexp.MustCompile(`[1-9][0-9]*`)),
					resource.TestCheckResourceAttrSet(oidcResource1, "claims_mapping.0.email"),
					resource.TestCheckResourceAttrSet(oidcResource1, "claims_mapping.0.subject"),
					resource.TestCheckResourceAttrSet(oidcResource1, "claims_mapping.0.last_name"),
					resource.TestCheckResourceAttrSet(oidcResource1, "claims_mapping.0.first_name"),
					resource.TestMatchResourceAttr(oidcResource1, "key.#", regexp.MustCompile(`[1-9][0-9]*`)),
					testMatchResourceAttrWhenVersionMatches(oidcResource1, "ui_button_label", regexp.MustCompile("^this is a test$"), ">= 38.1"),
				),
			},
			{
				Config: step2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resourceFieldsEqual(oidcResource1, oidcResource2, []string{
						"id", "org_id", "redirect_uri", "wellknown_endpoint", "key_refresh_endpoint",
						"user_authorization_endpoint", "claims_mapping.0.subject", "ui_button_label", "prefer_id_token",
					}),
					resource.TestCheckResourceAttr(oidcResource2, "user_authorization_endpoint", "https://www.dummy.com"),
					resource.TestCheckResourceAttr(oidcResource2, "claims_mapping.0.subject", "foo"),
					resourceFieldsEqual(oidcResource1, oidcResource3, []string{
						"id", "org_id", "redirect_uri", "wellknown_endpoint", "key_refresh_endpoint",
					}),
				),
			},
			{
				Config: step3,
				Check: resource.ComposeAggregateTestCheckFunc(
					resourceFieldsEqual(oidcResource1, oidcData, nil),
				),
			},
			{
				ResourceName:      oidcResource1,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdTopHierarchy(orgName1),
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

resource "vcd_org" "org3" {
  name              = "{{.OrgName3}}"
  full_name         = "{{.OrgName3}}"
  description       = "{{.OrgName3}}"
  delete_force      = true
  delete_recursive  = true
}

resource "vcd_org_oidc" "oidc1" {
  org_id                      = vcd_org.org1.id
  enabled                     = true
  {{.PreferIdToken}}
  client_id                   = "clientId"
  client_secret               = "clientSecret"
  max_clock_skew_seconds      = 60
  wellknown_endpoint          = "{{.WellKnownEndpoint}}"
  {{.UIButtonLabel}}
}
`

const testAccCheckVcdOrgOidc2 = testAccCheckVcdOrgOidc + `
resource "vcd_org_oidc" "oidc2" {
  org_id                      = vcd_org.org2.id
  enabled                     = true
  client_id                   = "clientId"
  client_secret               = "clientSecret"
  max_clock_skew_seconds      = 60
  wellknown_endpoint          = "{{.WellKnownEndpoint}}"
  user_authorization_endpoint = "https://www.dummy.com"
  claims_mapping {
	subject = "foo"
  }
}

resource "vcd_org_oidc" "oidc3" {
  org_id                      = vcd_org.org3.id
  enabled                     = vcd_org_oidc.oidc1.enabled
  {{.PreferIdToken}}
  client_id                   = vcd_org_oidc.oidc1.client_id
  client_secret               = vcd_org_oidc.oidc1.client_secret
  max_clock_skew_seconds      = vcd_org_oidc.oidc1.max_clock_skew_seconds
  issuer_id                   = vcd_org_oidc.oidc1.issuer_id
  user_authorization_endpoint = vcd_org_oidc.oidc1.user_authorization_endpoint
  access_token_endpoint       = vcd_org_oidc.oidc1.access_token_endpoint
  userinfo_endpoint           = vcd_org_oidc.oidc1.userinfo_endpoint
  scopes                      = vcd_org_oidc.oidc1.scopes
  claims_mapping {
    email      = vcd_org_oidc.oidc1.claims_mapping[0].email
    subject    = vcd_org_oidc.oidc1.claims_mapping[0].subject
    last_name  = vcd_org_oidc.oidc1.claims_mapping[0].last_name
    first_name = vcd_org_oidc.oidc1.claims_mapping[0].first_name
    full_name  = vcd_org_oidc.oidc1.claims_mapping[0].full_name
    groups     = vcd_org_oidc.oidc1.claims_mapping[0].groups
    roles      = vcd_org_oidc.oidc1.claims_mapping[0].roles
  }
  key {
    id              = tolist(vcd_org_oidc.oidc1.key)[0].id
    algorithm       = tolist(vcd_org_oidc.oidc1.key)[0].algorithm
    certificate     = tolist(vcd_org_oidc.oidc1.key)[0].certificate
	expiration_date = tolist(vcd_org_oidc.oidc1.key)[0].expiration_date
  }
  {{.UIButtonLabel}}
}
`

const testAccCheckVcdOrgOidc3 = testAccCheckVcdOrgOidc2 + `
data "vcd_org_oidc" "oidc_data" {
  org_id = vcd_org.org1.id
}
`

func validateAndGetOidcServerUrl(t *testing.T, testConfig TestConfig) *url.URL {
	if testConfig.Networking.OidcServer.Url == "" || testConfig.Networking.OidcServer.WellKnownEndpoint == "" {
		t.Skip("test requires OIDC configuration")
	}

	oidcServer, err := url.Parse(testConfig.Networking.OidcServer.Url)
	if err != nil {
		t.Skip(t.Name() + " requires OIDC Server URL and its well-known endpoint")
	}
	return oidcServer.JoinPath(testConfig.Networking.OidcServer.WellKnownEndpoint)
}
