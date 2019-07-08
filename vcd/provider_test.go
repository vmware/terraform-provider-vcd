// +build api functional catalog vapp network extnetwork org query vm vdc gateway disk binary lb lbAppProfile lbServiceMonitor lbServerPool ALL

package vcd

import (
	"testing"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

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
	if testConfig.Provider.User == "" {
		t.Fatal("provider.user must be set for acceptance tests")
	}
	if testConfig.Provider.Password == "" {
		t.Fatal("provider.password must be set for acceptance tests")
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

func init() {
	testingTags["api"] = "provider_test.go"
}

// createTemporaryVCDConnection is meant to create a VCDClient to check environment before executing specific acceptance
// tests and before VCDClient is accessible.
func createTemporaryVCDConnection() *VCDClient {
	config := Config{
		User:            testConfig.Provider.User,
		Password:        testConfig.Provider.Password,
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
