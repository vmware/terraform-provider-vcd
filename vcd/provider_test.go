//go:build api || functional || catalog || vapp || network || extnetwork || org || query || vm || vdc || gateway || disk || binary || lb || lbAppProfile || lbAppRule || lbServiceMonitor || lbServerPool || lbVirtualServer || user || access_control || standaloneVm || search || auth || nsxt || role || alb || certificate || ALL
// +build api functional catalog vapp network extnetwork org query vm vdc gateway disk binary lb lbAppProfile lbAppRule lbServiceMonitor lbServerPool lbVirtualServer user access_control standaloneVm search auth nsxt role alb certificate ALL

package vcd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func init() {
	testingTags["api"] = "provider_test.go"
}

// testAccProvider is a global provider used in tests
var testAccProvider *schema.Provider

// testAccProviders used in field ProviderFactories required for test runs in SDK 2.x
var testAccProviders map[string]func() (*schema.Provider, error)

func TestProvider(t *testing.T) {
	// Do not add pre and post checks
	if err := Provider().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvider_impl(t *testing.T) {
	// Do not add pre and post checks
	var _ *schema.Provider = Provider()
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
		ApiToken:        testConfig.Provider.ApiToken,
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

// createSystemTemporaryVCDConnection is like createTemporaryVCDConnection, but it will ignore all conditional
// configurations like `VCD_TEST_ORG_USER=1` and will still return a System client instead of user one. This allows to
// perform System actions (entities which require System rights - Org, Vdc, etc...)
func createSystemTemporaryVCDConnection() *VCDClient {
	var configStruct TestConfig
	configFileName := getConfigFileName()

	// Looks if the configuration file exists before attempting to read it
	if configFileName == "" {
		panic(fmt.Errorf("configuration file %s not found", configFileName))
	}
	jsonFile, err := ioutil.ReadFile(configFileName)
	if err != nil {
		panic(fmt.Errorf("could not read config file %s: %v", configFileName, err))
	}
	err = json.Unmarshal(jsonFile, &configStruct)
	if err != nil {
		panic(fmt.Errorf("could not unmarshal json file: %v", err))
	}

	config := Config{
		User:            configStruct.Provider.User,
		Password:        configStruct.Provider.Password,
		Token:           configStruct.Provider.Token,
		UseSamlAdfs:     configStruct.Provider.UseSamlAdfs,
		CustomAdfsRptId: configStruct.Provider.CustomAdfsRptId,
		SysOrg:          configStruct.Provider.SysOrg,
		Org:             configStruct.VCD.Org,
		Vdc:             configStruct.VCD.Vdc,
		Href:            configStruct.Provider.Url,
		InsecureFlag:    configStruct.Provider.AllowInsecure,
		MaxRetryTimeout: configStruct.Provider.MaxRetryTimeout,
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

// TestAccClientUserAgent ensures that client initialization config.Client() used by provider initializes
// go-vcloud-director client by having User-Agent set
func TestAccClientUserAgent(t *testing.T) {
	// Do not add pre and post checks
	// Exit the test early
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	clientConfig := Config{
		User:            testConfig.Provider.User,
		Password:        testConfig.Provider.Password,
		Token:           testConfig.Provider.Token,
		SysOrg:          testConfig.Provider.SysOrg,
		Org:             testConfig.VCD.Org,
		Vdc:             testConfig.VCD.Vdc,
		Href:            testConfig.Provider.Url,
		MaxRetryTimeout: testConfig.Provider.MaxRetryTimeout,
		InsecureFlag:    testConfig.Provider.AllowInsecure,
	}

	vcdClient, err := clientConfig.Client()
	if err != nil {
		t.Fatal("error initializing go-vcloud-director client: " + err.Error())
	}

	expectedHeaderPrefix := "terraform-provider-vcd/"
	if !strings.HasPrefix(vcdClient.VCDClient.Client.UserAgent, expectedHeaderPrefix) {
		t.Fatalf("Expected User-Agent header in go-vcloud-director to be '%s', got '%s'",
			expectedHeaderPrefix, vcdClient.VCDClient.Client.UserAgent)
	}
}
