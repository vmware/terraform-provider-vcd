// +build api functional catalog vapp network extnetwork org query vm vdc gateway disk binary lb lbAppProfile lbAppRule lbServiceMonitor lbServerPool lbVirtualServer user search auth ALL

package vcd

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func init() {
	testingTags["api"] = "provider_test.go"
}

var testAccProviders map[string]terraform.ResourceProvider
var testAccProvider *schema.Provider

func TestProvider(t *testing.T) {
	if err := Provider().(*schema.Provider).InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvider_impl(t *testing.T) {
	var _ terraform.ResourceProvider = Provider()
}

// When this function is called, the initialization in config_test.go has already happened.
// Therefore, we can safely require that testConfig fields have been filled.
func testAccPreCheck(t *testing.T) {
	if testConfig.Provider.User == "" && testConfig.Provider.Token == "" {
		t.Fatal("provider.user or provider.token must be set for acceptance tests")
	}
	if testConfig.Provider.Password == "" && testConfig.Provider.Token == "" {
		t.Fatal("provider.password or provider.token must be set for acceptance tests")
	}
	if testConfig.Provider.SysOrg == "" {
		t.Fatal("provider.sysOrg must be set for acceptance tests")
	}
	if testConfig.VCD.Org == "" {
		t.Fatal("vcd.org must be set for acceptance tests")
	}
	if testConfig.Provider.Url == "" {
		t.Fatal("provider.Url must be set for acceptance tests")
	}
	if testConfig.Networking.EdgeGateway == "" {
		t.Fatal("networking.edgeGateway must be set for acceptance tests")
	}
	if testConfig.VCD.Vdc == "" {
		t.Fatal("vcd.vdc must be set for acceptance tests")
	}
}

// createTemporaryVCDConnection is meant to create a VCDClient to check environment before executing specific acceptance
// tests and before VCDClient is accessible.
func createTemporaryVCDConnection() *VCDClient {
	config := Config{
		User:            testConfig.Provider.User,
		Password:        testConfig.Provider.Password,
		Token:           testConfig.Provider.Token,
		UseSamlAdfs:     testConfig.Provider.UseSamlAdfs,
		CustomAdfsRptId: testConfig.Provider.CustomAdfsRptId,
		SysOrg:          testConfig.Provider.SysOrg,
		Org:             testConfig.VCD.Org,
		Vdc:             testConfig.VCD.Vdc,
		Href:            testConfig.Provider.Url,
		InsecureFlag:    testConfig.Provider.AllowInsecure,
		MaxRetryTimeout: testConfig.Provider.MaxRetryTimeout,
	}
	conn, err := config.Client()
	if err != nil {
		panic("unable to initialize VCD connection :" + err.Error())
	}
	return conn
}

// minIfLess returns:
// `min` if `value` is less than min
// `value` if `value` > `min`
func minIfLess(min, value int) int {
	if value < min {
		return min
	}

	return value
}
