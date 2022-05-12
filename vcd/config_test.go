//go:build api || functional || catalog || vapp || network || extnetwork || org || query || vm || vdc || gateway || disk || binary || lb || lbServiceMonitor || lbServerPool || lbAppProfile || lbAppRule || lbVirtualServer || access_control || user || standaloneVm || search || auth || nsxt || role || alb || certificate || vdcGroup || ALL
// +build api functional catalog vapp network extnetwork org query vm vdc gateway disk binary lb lbServiceMonitor lbServerPool lbAppProfile lbAppRule lbVirtualServer access_control user standaloneVm search auth nsxt role alb certificate vdcGroup ALL

package vcd

// This module provides initialization routines for the whole suite

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
	"text/template"
	"time"

	"github.com/hashicorp/go-version"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/util"
)

func init() {

	// To list the flags when we run "go test -tags functional -vcd-help", the flag name must start with "vcd"
	// They will all appear alongside the native flags when we use an invalid one
	setBoolFlag(&vcdHelp, "vcd-help", "VCD_HELP", "Show vcd flags")
	setBoolFlag(&vcdRemoveTestList, "vcd-remove-test-list", "VCD_REMOVE_TEST_LIST", "Remove list of test runs")
	setBoolFlag(&vcdPrePostChecks, "vcd-pre-post-checks", "VCD_PRE_POST_CHECKS", "Perform checks before and after tests")
	setBoolFlag(&vcdShowTimestamp, "vcd-show-timestamp", "VCD_SHOW_TIMESTAMP", "Show timestamp in pre and post checks")
	setBoolFlag(&vcdShowElapsedTime, "vcd-show-elapsed-time", "VCD_SHOW_ELAPSED_TIME", "Show elapsed time since the start of the suite in pre and post checks")
	setBoolFlag(&vcdShowCount, "vcd-show-count", "VCD_SHOW_COUNT", "Show number of pass/fail tests")
	setBoolFlag(&vcdReRunFailed, "vcd-re-run-failed", "VCD_RE_RUN_FAILED", "Run only tests that failed in a previous run")
	setBoolFlag(&testDistributedNetworks, "vcd-test-distributed", "", "enables testing of distributed network")
	setBoolFlag(&enableDebug, "vcd-debug", "GOVCD_DEBUG", "enables debug output")
	setBoolFlag(&vcdTestVerbose, "vcd-verbose", "TEST_VERBOSE", "enables verbose output")
	setBoolFlag(&enableTrace, "vcd-trace", "GOVCD_TRACE", "enables function calls tracing")
	setBoolFlag(&vcdShortTest, "vcd-short", "VCD_SHORT_TEST", "runs short test")
	setBoolFlag(&vcdAddProvider, "vcd-add-provider", envVcdAddProvider, "add provider to test scripts")
	setBoolFlag(&vcdSkipTemplateWriting, "vcd-skip-template-write", envVcdSkipTemplateWriting, "Skip writing templates to file")
	setBoolFlag(&vcdRemoveOrgVdcFromTemplate, "vcd-remove-org-vdc-from-template", envVcdRemoveOrgVdcFromTemplate, "Remove org and VDC from template")
	setBoolFlag(&vcdTestOrgUser, "vcd-test-org-user", envVcdTestOrgUser, "Run tests with org user")
	setStringFlag(&vcdSkipPattern, "vcd-skip-pattern", "VCD_SKIP_PATTERN", "Skip tests that match the pattern (implies vcd-pre-post-checks")

}

// Structure to get info from a config json file that the user specifies
type TestConfig struct {
	Provider struct {
		User     string `json:"user"`
		Password string `json:"password"`
		Token    string `json:"token,omitempty"`
		ApiToken string `json:"api_token,omitempty"`

		// UseSamlAdfs specifies if SAML auth is used for authenticating vCD instead of local login.
		// The above `User` and `Password` will be used to authenticate against ADFS IdP when true.
		UseSamlAdfs bool `json:"useSamlAdfs"`

		// CustomAdfsRptId allows to set custom Relaying Party Trust identifier if needed. Only has
		// effect if `UseSamlAdfs` is true.
		CustomAdfsRptId string `json:"customAdfsRptId,omitempty"`

		// The variables `SamlUser`, `SamlPassword` and `SamlCustomRptId` are optional and are
		// related to additional test run specifically with SAML user/password. It can be useful in
		// case local user is used for test run (defined by above 'User', 'Password' variables).
		// SamlUser takes ADFS friendly format ('contoso.com\username' or 'username@contoso.com')
		SamlUser        string `json:"samlUser,omitempty"`
		SamlPassword    string `json:"samlPassword,omitempty"`
		SamlCustomRptId string `json:"samlCustomRptId,omitempty"`

		Url                      string `json:"url"`
		SysOrg                   string `json:"sysOrg"`
		AllowInsecure            bool   `json:"allowInsecure"`
		TerraformAcceptanceTests bool   `json:"tfAcceptanceTests"`
		UseVcdConnectionCache    bool   `json:"useVcdConnectionCache"`
		MaxRetryTimeout          int    `json:"maxRetryTimeout"`
	} `json:"provider"`
	VCD struct {
		Org         string `json:"org"`
		Vdc         string `json:"vdc"`
		ProviderVdc struct {
			Name            string `json:"name"`
			NetworkPool     string `json:"networkPool"`
			StorageProfile  string `json:"storageProfile"`
			StorageProfile2 string `json:"storageProfile2"`
		} `json:"providerVdc"`
		NsxtProviderVdc struct {
			Name           string `json:"name"`
			StorageProfile string `json:"storageProfile"`
			NetworkPool    string `json:"networkPool"`
		} `json:"nsxtProviderVdc"`
		Catalog struct {
			Name                    string `json:"name,omitempty"`
			CatalogItem             string `json:"catalogItem,omitempty"`
			CatalogItemWithMultiVms string `json:"catalogItemWithMultiVms,omitempty"`
			VmName1InMultiVmItem    string `json:"vmName1InMultiVmItem,omitempty"`
			VmName2InMultiVmItem    string `json:"VmName2InMultiVmItem,omitempty"`
		} `json:"catalog"`
	} `json:"vcd"`
	Networking struct {
		ExternalIp                   string `json:"externalIp,omitempty"`
		InternalIp                   string `json:"internalIp,omitempty"`
		EdgeGateway                  string `json:"edgeGateway,omitempty"`
		SharedSecret                 string `json:"sharedSecret"`
		Vcenter                      string `json:"vcenter,omitempty"`
		ExternalNetwork              string `json:"externalNetwork,omitempty"`
		ExternalNetworkPortGroup     string `json:"externalNetworkPortGroup,omitempty"`
		ExternalNetworkPortGroupType string `json:"externalNetworkPortGroupType,omitempty"`
		Local                        struct {
			LocalIp            string `json:"localIp"`
			LocalSubnetGateway string `json:"localSubnetGw"`
		} `json:"local"`
		Peer struct {
			PeerIp            string `json:"peerIp"`
			PeerSubnetGateway string `json:"peerSubnetGw"`
		} `json:"peer"`
	} `json:"networking"`
	Nsxt struct {
		Manager           string `json:"manager"`
		Tier0router       string `json:"tier0router"`
		Tier0routerVrf    string `json:"tier0routervrf"`
		Vdc               string `json:"vdc"`
		ExternalNetwork   string `json:"externalNetwork"`
		EdgeGateway       string `json:"edgeGateway"`
		NsxtImportSegment string `json:"nsxtImportSegment"`

		NsxtAlbControllerUrl      string `json:"nsxtAlbControllerUrl"`
		NsxtAlbControllerUser     string `json:"nsxtAlbControllerUser"`
		NsxtAlbControllerPassword string `json:"nsxtAlbControllerPassword"`
		NsxtAlbImportableCloud    string `json:"nsxtAlbImportableCloud"`
	} `json:"nsxt"`
	Logging struct {
		Enabled         bool   `json:"enabled,omitempty"`
		LogFileName     string `json:"logFileName,omitempty"`
		LogHttpRequest  bool   `json:"logHttpRequest,omitempty"`
		LogHttpResponse bool   `json:"logHttpResponse,omitempty"`
	} `json:"logging"`
	Ova struct {
		OvaPath             string `json:"ovaPath,omitempty"`
		OvfUrl              string `json:"ovfUrl,omitempty"`
		UploadPieceSize     int64  `json:"uploadPieceSize,omitempty"`
		UploadProgress      bool   `json:"uploadProgress,omitempty"`
		OvaTestFileName     string `json:"ovaTestFileName,omitempty"`
		OvaDownloadUrl      string `json:"ovaDownloadUrl,omitempty"`
		Preserve            bool   `json:"preserve,omitempty"`
		OvaVappMultiVmsPath string `json:"ovaVappMultiVmsPath,omitempty"`
	} `json:"ova"`
	Media struct {
		MediaPath       string `json:"mediaPath,omitempty"`
		UploadPieceSize int64  `json:"uploadPieceSize,omitempty"`
		UploadProgress  bool   `json:"uploadProgress,omitempty"`
		MediaName       string `json:"mediaName,omitempty"`
	} `json:"media"`
	Misc struct {
		LdapContainer string `json:"ldapContainer,omitempty"`
	} `json:"misc"`
	Certificates struct {
		Certificate1Path           string `json:"certificate1Path,omitempty"`           // absolute path to pem file
		Certificate1PrivateKeyPath string `json:"certificate1PrivateKeyPath,omitempty"` // absolute path to private key pem file
		Certificate1Pass           string `json:"certificate1Pass,omitempty"`           // pass phrase for private key
		Certificate2Path           string `json:"certificate2Path,omitempty"`           // absolute path to pem file
		Certificate2PrivateKeyPath string `json:"certificate2PrivateKeyPath,omitempty"` // absolute path to private key pem file
		Certificate2Pass           string `json:"certificate2Pass,omitempty"`           // absolute path to pem file
	} `json:"certificates"`
	// Data used to create a new environment, in addition to the regular test configuration file
	TestEnvBuild struct {
		Gateway                      string `json:"gateway"`                      // Gateway for external network
		Netmask                      string `json:"netmask"`                      // Netmask for external network
		ExternalNetworkStartIp       string `json:"externalNetworkStartIp"`       // Start IP for external network
		ExternalNetworkEndIp         string `json:"externalNetworkEndIp"`         // End IP for external network
		Dns1                         string `json:"dns1"`                         // DNS 1 for external network
		Dns2                         string `json:"dns2"`                         // DNS 2 for external network
		ExternalNetworkPortGroup     string `json:"externalNetworkPortGroup"`     // port group, if different from Networking.ExternalNetworkPortGroup
		ExternalNetworkPortGroupType string `json:"externalNetworkPortGroupType"` // port group type, if different from Networking.ExternalNetworkPortGroupType
		RoutedNetwork                string `json:"routedNetwork"`                // optional routed network name to create
		IsolatedNetwork              string `json:"isolatedNetwork"`              // optional isolated network name to create
		DirectNetwork                string `json:"directNetwork"`                // optional direct network name to create
		MediaPath                    string `json:"mediaPath"`                    // Media path, if different from Media.MediaPath
		MediaName                    string `json:"mediaName"`                    // Media name to create
		OvaPath                      string `json:"ovaPath"`                      // Ova Path, if different from Ova.OvaPath
		OrgUser                      string `json:"orgUser"`                      // Org User to be created within the organization
		OrgUserPassword              string `json:"orgUserPassword"`              // Password for the Org User to be created within the organization
	} `json:"testEnvBuild"`
	EnvVariables map[string]string `json:"envVariables,omitempty"`
}

// names for created resources for all the tests
var (
	testSuiteCatalogName    = "TestSuiteCatalog"
	testSuiteCatalogOVAItem = "TestSuiteOVA"

	// vcdAddProvider will add the provide section to the template
	vcdAddProvider = os.Getenv(envVcdAddProvider) != ""

	// vcdSkipTemplateWriting disable the writing of the template to a .tf file
	vcdSkipTemplateWriting = false

	// vcdRemoveOrgVdcFromTemplate removes org and vdc from template, thus tetsing with the
	// variables in provider section
	vcdRemoveOrgVdcFromTemplate = false

	// vcdTestOrgUser enables testing with the Org User
	vcdTestOrgUser = false

	// vcdRemoveTestList triggers the removal of the test run list if present
	vcdRemoveTestList = false

	// vcdPrePostChecks enables pre and post checks for all tests
	vcdPrePostChecks = false

	// vcdReRunFailed will run only tests that failed in a previous run
	vcdReRunFailed = false

	// vcdShowTimestamp shows a time stamp at the start of each test
	vcdShowTimestamp = false

	// vcdShowElapsedTime shows the elapsed time since the start od the suite
	vcdShowElapsedTime = false

	// vcdShowCount shows the count of pass/skip/fail at the end of the suite
	vcdShowCount = false

	// vcdSkipPattern will skip all tests with a name that matches a given pattern
	vcdSkipPattern string

	// vcdSkipAllFile is the name of the file that will skip all the tests if found during a pre-test check
	vcdSkipAllFile = "skip_vcd_tests"

	// vcdStartTime is he time when the tests started
	vcdStartTime = time.Now()

	// vcdPassCount, vcdFailCount, vcdSkipCount are the global counters for the tests result
	vcdPassCount = 0
	vcdFailCount = 0
	vcdSkipCount = 0

	// vcdHelp displays the vcd-* flags
	vcdHelp = false

	// Distributed networks require an edge gateway with distributed routing enabled,
	// which in turn requires a NSX controller. To run the distributed test, users
	// need to set the environment variable VCD_TEST_DISTRIBUTED_NETWORK
	testDistributedNetworks = false

	// runTestRunListFileLock regulates access to the list of run tests
	runTestRunListFileLock = newMutexKVSilent()
)

const (
	customTemplatesDirectory       = "test-templates"
	testArtifactsDirectory         = "test-artifacts"
	envVcdAddProvider              = "VCD_ADD_PROVIDER"
	envVcdSkipTemplateWriting      = "VCD_SKIP_TEMPLATE_WRITING"
	envVcdRemoveOrgVdcFromTemplate = "REMOVE_ORG_VDC_FROM_TEMPLATE"
	envVcdTestOrgUser              = "VCD_TEST_ORG_USER"

	// Warning message used for all tests
	acceptanceTestsSkipped = "Acceptance tests skipped unless env 'TF_ACC' set"
	// This template will be added to test resource snippets on demand
	providerTemplate = `
# tags {{.Tags}}
# dirname {{.DirName}}
# comment {{.Comment}}
# date {{.Timestamp}}
# file {{.CallerFileName}}
#

provider "vcd" {
  user                 = "{{.PrUser}}"
  password             = "{{.PrPassword}}"
  token                = "{{.Token}}"
  api_token            = "{{.ApiToken}}"
  auth_type            = "{{.AuthType}}"
  saml_adfs_rpt_id     = "{{.SamlAdfsCustomRptId}}"
  url                  = "{{.PrUrl}}"
  sysorg               = "{{.PrSysOrg}}"
  org                  = "{{.PrOrg}}"
  vdc                  = "{{.PrVdc}}"
  allow_unverified_ssl = "{{.AllowInsecure}}"
  max_retry_timeout    = {{.MaxRetryTimeout}}
  #version             = "~> {{.VersionRequired}}"
  logging              = {{.Logging}}
  logging_file         = "{{.LoggingFile}}"
}
`
)

var (

	// This is a global variable shared across all tests. It contains
	// the information from the configuration file.
	testConfig TestConfig

	// Enables the short test (used by "make test")
	vcdShortTest bool = os.Getenv("VCD_SHORT_TEST") != ""

	// Keeps track of test artifact names, to avoid duplicates
	testArtifactNames = make(map[string]string)
)

func testDistributedNetworksEnabled() bool {
	return testDistributedNetworks || os.Getenv("VCD_TEST_DISTRIBUTED_NETWORK") != ""
}

// Returns true if the current configuration uses a system administrator for connections
func usingSysAdmin() bool {
	return strings.ToLower(testConfig.Provider.SysOrg) == "system"
}

// Gets a list of all variables mentioned in a template
func GetVarsFromTemplate(tmpl string) []string {
	var varList []string

	// Regular expression to match a template variable
	// Two opening braces       {{
	// one dot                  \.
	// non-closing-brace chars  [^}]+
	// Two closing braces       }}
	reTemplateVar := regexp.MustCompile(`{{\.([^{]+)}}`)
	captureList := reTemplateVar.FindAllStringSubmatch(tmpl, -1)
	if len(captureList) > 0 {
		for _, capture := range captureList {
			varList = append(varList, capture[1])
		}
	}
	return varList
}

// templateFill fills a template with data provided as a StringMap and adds `provider`
// configuration.
// Returns the text of a ready-to-use Terraform directive. It also saves the filled
// template to a file, for further troubleshooting.
func templateFill(tmpl string, data StringMap) string {

	// Gets the name of the function containing the template
	caller := callFuncName()
	realCaller := caller
	// Removes the full path to the function, leaving only package + function name
	caller = filepath.Base(caller)

	_, callerFileName, _, _ := runtime.Caller(1)
	// First, we get all variables in the pattern {{.VarName}}
	varList := GetVarsFromTemplate(tmpl)
	if len(varList) > 0 {
		for _, capture := range varList {
			// For each variable in the template text, we look whether it is
			// in the map
			_, ok := data[capture]
			if !ok {
				data[capture] = fmt.Sprintf("*** MISSING FIELD [%s] from func %s", capture, caller)
			}
		}
	}
	prefix := "vcd"
	_, ok := data["Prefix"]

	if ok {
		prefix = data["Prefix"].(string)
	}

	// If the call comes from a function that does not have a good descriptive name,
	// (for example when it's an auxiliary function that builds the template but does not
	// run the test) users can add the function name in the data, and it will be used instead of
	// the caller name.
	funcName, ok := data["FuncName"]
	if ok {
		caller = prefix + "." + funcName.(string)
	}

	// If requested, the provider defined in testConfig will be added to test snippets.
	if vcdAddProvider {
		// the original template is prefixed with the provider template
		tmpl = providerTemplate + tmpl

		// The data structure used to fill the template is integrated with
		// provider data
		data["PrUser"] = testConfig.Provider.User
		data["PrPassword"] = testConfig.Provider.Password
		data["SamlAdfsCustomRptId"] = testConfig.Provider.CustomAdfsRptId
		data["Token"] = testConfig.Provider.Token
		data["ApiToken"] = testConfig.Provider.ApiToken
		data["PrUrl"] = testConfig.Provider.Url
		data["PrSysOrg"] = testConfig.Provider.SysOrg
		data["PrOrg"] = testConfig.VCD.Org
		vdcName, found := data["PrVdc"]
		if !found || vdcName == "" {
			data["PrVdc"] = testConfig.VCD.Vdc
		}
		data["AllowInsecure"] = testConfig.Provider.AllowInsecure
		data["MaxRetryTimeout"] = testConfig.Provider.MaxRetryTimeout
		data["VersionRequired"] = currentProviderVersion
		data["Logging"] = testConfig.Logging.Enabled
		if testConfig.Logging.LogFileName != "" {
			data["LoggingFile"] = testConfig.Logging.LogFileName
		} else {
			data["LoggingFile"] = util.ApiLogFileName
		}

		// Pick correct auth_type
		switch {
		case testConfig.Provider.Token != "":
			data["AuthType"] = "token"
		case testConfig.Provider.ApiToken != "":
			data["AuthType"] = "api_token"
		case testConfig.Provider.UseSamlAdfs:
			data["AuthType"] = "saml_adfs"
		default:
			data["AuthType"] = "integrated" // default AuthType for local and LDAP users
		}
	}
	if _, ok := data["Tags"]; !ok {
		data["Tags"] = "ALL"
	}
	nullableItems := []string{"Comment", "DirName"}
	for _, item := range nullableItems {
		if _, ok := data[item]; !ok {
			data[item] = " "
		}
	}
	if _, ok := data["CallerFileName"]; !ok {
		data["CallerFileName"] = callerFileName
	}
	data["Timestamp"] = time.Now().Format("2006-01-02 15:04")

	// Creates a template. The template gets the same name of the calling function, to generate a better
	// error message in case of failure
	unfilledTemplate := template.Must(template.New(caller).Parse(tmpl))
	buf := &bytes.Buffer{}

	// If an error occurs, returns an empty string
	if err := unfilledTemplate.Execute(buf, data); err != nil {
		return ""
	}
	// Writes the populated template into a directory named "test-artifacts"
	// These templates will help investigate failed tests using Terraform
	// Writing is enabled by default. It can be skipped using an environment variable.
	TemplateWriting := true
	if vcdSkipTemplateWriting {
		TemplateWriting = false
	}
	var writeStr []byte = buf.Bytes()

	// This is a quick way of enabling an alternate testing mode:
	// When REMOVE_ORG_VDC_FROM_TEMPLATE is set, the terraform
	// templates will be changed on-the-fly, to comment out the
	// definitions of org and vdc. This will force the test to
	// borrow org and vcd from the provider.
	if vcdRemoveOrgVdcFromTemplate {
		reOrg := regexp.MustCompile(`\sorg\s*=`)
		buf2 := reOrg.ReplaceAll(buf.Bytes(), []byte("# org = "))
		reVdc := regexp.MustCompile(`\svdc\s*=`)
		buf2 = reVdc.ReplaceAll(buf2, []byte("# vdc = "))
		writeStr = buf2
	}
	if TemplateWriting {
		if !dirExists(testArtifactsDirectory) {
			err := os.Mkdir(testArtifactsDirectory, 0755)
			if err != nil {
				panic(fmt.Errorf("error creating directory %s: %s", testArtifactsDirectory, err))
			}
		}
		resourceFile := path.Join(testArtifactsDirectory, caller) + ".tf"
		storedFunc, alreadyWritten := testArtifactNames[resourceFile]
		if alreadyWritten {
			panic(fmt.Sprintf("File %s was already used from function %s", resourceFile, storedFunc))
		}
		testArtifactNames[resourceFile] = realCaller

		file, err := os.Create(resourceFile)
		if err != nil {
			panic(fmt.Errorf("error creating file %s: %s", resourceFile, err))
		}
		writer := bufio.NewWriter(file)
		count, err := writer.Write(writeStr)
		if err != nil || count == 0 {
			panic(fmt.Errorf("error writing to file %s. Reported %d bytes written. %s", resourceFile, count, err))
		}
		err = writer.Flush()
		if err != nil {
			panic(fmt.Errorf("error flushing file %s. %s", resourceFile, err))
		}
		_ = file.Close()
	}
	// Returns the populated template
	return string(writeStr)
}

func getConfigFileName() string {
	// First, we see whether the user has indicated a custom configuration file
	// from a non-standard location
	config := os.Getenv("VCD_CONFIG")

	// If there was no custom file, we look for the default one
	if config == "" {
		config = getCurrentDir() + "/vcd_test_config.json"
	}
	// Looks if the configuration file exists before attempting to read it
	if fileExists(config) {
		return config
	}
	return ""
}

// Reads the configuration file and returns its contents as a TestConfig structure
// The default file is called vcd_test_config.json in the same directory where
// the test files are.
// Users may define a file in a different location using the environment variable
// VCD_CONFIG
// This function doesn't return an error. It panics immediately because its failure
// will prevent the whole test suite from running
func getConfigStruct(config string) TestConfig {
	var configStruct TestConfig

	// Looks if the configuration file exists before attempting to read it
	if config == "" {
		panic(fmt.Errorf("configuration file %s not found", config))
	}
	jsonFile, err := ioutil.ReadFile(config)
	if err != nil {
		panic(fmt.Errorf("could not read config file %s: %v", config, err))
	}
	err = json.Unmarshal(jsonFile, &configStruct)
	if err != nil {
		panic(fmt.Errorf("could not unmarshal json file: %v", err))
	}

	// Sets (or clears) environment variables defined in the configuration file
	if configStruct.EnvVariables != nil {
		for key, value := range configStruct.EnvVariables {
			currentEnvValue := os.Getenv(key)
			debugPrintf("# Setting environment variable '%s' from '%s' to '%s'\n", key, currentEnvValue, value)
			_ = os.Setenv(key, value)
		}
	}
	// Reading the configuration file was successful.
	// Now we fill the environment variables that the library is using for its own initialization.
	if configStruct.Provider.TerraformAcceptanceTests {
		// defined in vendor/github.com/hashicorp/terraform/helper/resource/testing.go
		_ = os.Setenv("TF_ACC", "1")
	}
	// The following variables are used in ./provider.go
	if configStruct.Provider.MaxRetryTimeout == 0 {
		// If there is no retry timeout in the configuration, and there is no env variable for it, we set a new one
		if os.Getenv("VCD_MAX_RETRY_TIMEOUT") == "" {
			// Setting a default value that should be reasonable for these tests, as we run many heavy operations
			_ = os.Setenv("VCD_MAX_RETRY_TIMEOUT", "300")
		}
	} else {
		newRetryTimeout := fmt.Sprintf("%d", configStruct.Provider.MaxRetryTimeout)
		_ = os.Setenv("VCD_MAX_RETRY_TIMEOUT", newRetryTimeout)
	}
	if configStruct.Provider.SysOrg == "" {
		configStruct.Provider.SysOrg = configStruct.VCD.Org
	}

	if vcdTestOrgUser {
		user := configStruct.TestEnvBuild.OrgUser
		password := configStruct.TestEnvBuild.OrgUserPassword
		if user == "" || password == "" {
			panic(fmt.Sprintf("%s was enabled, but org user credentials were not found in the configuration file", envVcdTestOrgUser))
		}
		configStruct.Provider.User = user
		configStruct.Provider.Password = password
		configStruct.Provider.SysOrg = configStruct.VCD.Org
		fmt.Println("VCD_TEST_ORG_USER was enabled. Using Org User credentials from configuration file")
	}
	if configStruct.Provider.Token != "" && configStruct.Provider.Password == "" {
		configStruct.Provider.Password = "TOKEN"
	}
	_ = os.Setenv("VCD_USER", configStruct.Provider.User)
	_ = os.Setenv("VCD_PASSWORD", configStruct.Provider.Password)
	// VCD_TOKEN supplied via CLI has bigger priority than configured one
	if os.Getenv("VCD_TOKEN") == "" {
		_ = os.Setenv("VCD_TOKEN", configStruct.Provider.Token)
	} else {
		configStruct.Provider.Token = os.Getenv("VCD_TOKEN")
	}

	if configStruct.Provider.UseSamlAdfs {
		_ = os.Setenv("VCD_AUTH_TYPE", "saml_adfs")
		_ = os.Setenv("VCD_SAML_ADFS_RPT_ID", configStruct.Provider.CustomAdfsRptId)
	}

	_ = os.Setenv("VCD_URL", configStruct.Provider.Url)
	_ = os.Setenv("VCD_SYS_ORG", configStruct.Provider.SysOrg)
	_ = os.Setenv("VCD_ORG", configStruct.VCD.Org)
	_ = os.Setenv("VCD_VDC", configStruct.VCD.Vdc)
	if configStruct.Provider.UseVcdConnectionCache {
		enableConnectionCache = true
	}
	if configStruct.Provider.AllowInsecure {
		_ = os.Setenv("VCD_ALLOW_UNVERIFIED_SSL", "1")
	}

	// Define logging parameters if enabled
	if configStruct.Logging.Enabled {
		util.EnableLogging = true
		if configStruct.Logging.LogFileName != "" {
			util.ApiLogFileName = configStruct.Logging.LogFileName
		}
		if configStruct.Logging.LogHttpResponse {
			util.LogHttpResponse = true
		}
		if configStruct.Logging.LogHttpRequest {
			util.LogHttpRequest = true
		}
		util.InitLogging()
	}

	if configStruct.Ova.OvaPath != "" {
		ovaPath, err := filepath.Abs(configStruct.Ova.OvaPath)
		if err != nil {
			panic("error retrieving absolute path for OVA path " + configStruct.Ova.OvaPath)
		}
		configStruct.Ova.OvaPath = ovaPath
	}
	if configStruct.Media.MediaPath != "" {
		mediaPath, err := filepath.Abs(configStruct.Media.MediaPath)
		if err != nil {
			panic("error retrieving absolute path for Media path " + configStruct.Media.MediaPath)
		}
		configStruct.Media.MediaPath = mediaPath
	}
	if configStruct.Ova.OvaVappMultiVmsPath != "" {
		multiVmOvaPath, err := filepath.Abs(configStruct.Ova.OvaVappMultiVmsPath)
		if err != nil {
			panic("error retrieving absolute path for multi OVA path " + configStruct.Ova.OvaVappMultiVmsPath)
		}
		configStruct.Ova.OvaVappMultiVmsPath = multiVmOvaPath
	}
	if configStruct.Certificates.Certificate1Path != "" {
		certificatePath1Path, err := filepath.Abs(configStruct.Certificates.Certificate1Path)
		if err != nil {
			panic("error retrieving absolute path for certificate 1 path " + configStruct.Certificates.Certificate1Path)
		}
		configStruct.Certificates.Certificate1Path = certificatePath1Path
	}
	if configStruct.Certificates.Certificate2Path != "" {
		certificatePath2Path, err := filepath.Abs(configStruct.Certificates.Certificate2Path)
		if err != nil {
			panic("error retrieving absolute path for certificate 2 path " + configStruct.Certificates.Certificate2Path)
		}
		configStruct.Certificates.Certificate2Path = certificatePath2Path
	}
	if configStruct.Certificates.Certificate1PrivateKeyPath != "" {
		certificatePrivatePath1Path, err := filepath.Abs(configStruct.Certificates.Certificate1PrivateKeyPath)
		if err != nil {
			panic("error retrieving absolute path for private certificate 1 path " + configStruct.Certificates.Certificate1PrivateKeyPath)
		}
		configStruct.Certificates.Certificate1PrivateKeyPath = certificatePrivatePath1Path
	}
	if configStruct.Certificates.Certificate2PrivateKeyPath != "" {
		certificatePrivatePath2Path, err := filepath.Abs(configStruct.Certificates.Certificate2PrivateKeyPath)
		if err != nil {
			panic("error retrieving absolute path for private certificate 2 path " + configStruct.Certificates.Certificate2PrivateKeyPath)
		}
		configStruct.Certificates.Certificate2PrivateKeyPath = certificatePrivatePath2Path
	}

	// Partial duplication of actions performed in createSuiteCatalogAndItem
	// It is needed when we run the binary tests without TEST_ACC
	// TODO: convert the actions from createSuiteCatalogAndItem into a terraform config file
	if configStruct.VCD.Catalog.Name != "" {
		testSuiteCatalogName = configStruct.VCD.Catalog.Name
	}
	if configStruct.VCD.Catalog.CatalogItem != "" {
		testSuiteCatalogOVAItem = configStruct.VCD.Catalog.CatalogItem
	}
	return configStruct
}

// setTestEnv enables environment variables that are also used in non-test code
func setTestEnv() {
	if enableDebug {
		_ = os.Setenv("GOVCD_DEBUG", "1")
	}
}

// getVcdVersion returns the VCD version (three digits + build number)
// To get the version, we establish a new connection with the credentials
// chosen for the current test.
func getVcdVersion(config TestConfig) (string, error) {
	vcdClient, err := getTestVCDFromJson(config)
	if vcdClient == nil || err != nil {
		return "", err
	}
	err = ProviderAuthenticate(vcdClient, config.Provider.User, config.Provider.Password, config.Provider.Token, config.Provider.SysOrg, config.Provider.ApiToken)
	if err != nil {
		return "", err
	}
	version, _, err := vcdClient.Client.GetVcdVersion()
	if err != nil {
		return "", err
	}
	return version, nil
}

// This function is called before any other test
func TestMain(m *testing.M) {

	// Set BuildVersion to have consistent User-Agent for tests:
	// [e.g. terraform-provider-vcd/test (darwin/amd64; isProvider:true)]
	BuildVersion = "test"

	// Enable custom flags
	flag.Parse()
	setTestEnv()
	flag.CommandLine.VisitAll(func(f *flag.Flag) {
		if f.Name == "test.v" {
			if f.Value.String() == "false" {
				fmt.Printf("Missing '-v' flag\n")
				os.Exit(1)
			}
		}
	})
	// If -vcd-help was in the command line
	if vcdHelp {
		fmt.Println("vcd flags:")
		fmt.Println()
		// Prints only the flags defined in this package
		flag.CommandLine.VisitAll(func(f *flag.Flag) {
			if strings.Contains(f.Name, "vcd-") {
				fmt.Printf("  -%-40s %s (%v)\n", f.Name, f.Usage, f.Value)
			}
		})
		fmt.Println()
		os.Exit(0)
	}
	// If any of the checks is enabled, we enable the pre and post test functions
	if vcdSkipPattern != "" || vcdShowElapsedTime || vcdShowTimestamp || vcdRemoveTestList ||
		vcdShowCount || vcdReRunFailed {
		vcdPrePostChecks = true
	}
	if vcdPrePostChecks {
		// remove the user-placed skip file
		_ = os.Remove(vcdSkipAllFile)
	}

	// Fills the configuration variable: it will be available to all tests,
	// or the whole suite will fail if it is not found.
	// If VCD_SHORT_TEST is defined, it means that "make test" is called,
	// and we won't really run any tests involving vcd connections.
	configFile := getConfigFileName()
	if configFile != "" {
		testConfig = getConfigStruct(configFile)
	}
	if vcdRemoveTestList {
		for _, ft := range []string{"pass", "fail"} {
			err := removeTestRunList(ft)
			if err != nil {
				fmt.Printf("Error removing testRunList: %s", err)
				fmt.Printf("You should remove the file %s manually before trying again", getTestListFile(ft))
				os.Exit(0)
			}
		}
	}
	if !vcdShortTest {

		if configFile == "" {
			fmt.Println("No configuration file found")
			os.Exit(1)
		}
		versionInfo, err := getVcdVersion(testConfig)
		if err == nil {
			versionInfo = fmt.Sprintf("(version %s)", versionInfo)
		}
		fmt.Printf("Connecting to %s %s\n", testConfig.Provider.Url, versionInfo)

		authentication := "password"
		if testConfig.Provider.UseSamlAdfs {
			authentication = "SAML password"
		}
		// Token based auth has priority over other types
		if testConfig.Provider.Token != "" {
			authentication = "token"
		}
		if testConfig.Provider.ApiToken != "" {
			authentication = "API-token"
		}

		fmt.Printf("as user %s@%s (using %s)\n", testConfig.Provider.User, testConfig.Provider.SysOrg, authentication)
		// Provider initialization moved here from provider_test.init
		testAccProvider = Provider()
		testAccProviders = map[string]func() (*schema.Provider, error){
			"vcd": func() (*schema.Provider, error) {
				return testAccProvider, nil
			},
		}

		// forcing item cleanup before test run
		if os.Getenv("VCD_TEST_SUITE_CLEANUP") != "" {
			fmt.Printf("VCD_TEST_SUITE_CLEANUP found and TestSuite resource cleanup initiated\n")
			destroySuiteCatalogAndItem(testConfig)
		}

		createSuiteCatalogAndItem(testConfig)
	}

	// Runs all test functions
	exitCode := m.Run()

	if !vcdShortTest {

		if !testConfig.Ova.Preserve {
			destroySuiteCatalogAndItem(testConfig)
		} else {
			fmt.Printf("TestSuite destroy skipped - preserve turned on \n")
		}
	}
	if vcdShowCount {
		fmt.Printf("Pass: %5d - Skip: %5d - Fail: %5d\n", vcdPassCount, vcdSkipCount, vcdFailCount)
	}

	// TODO: cleanup leftovers
	os.Exit(exitCode)
}

// createSuiteCatalogAndItem creates catalog and/or catalog item if they are not preconfigured.
func createSuiteCatalogAndItem(config TestConfig) {
	fmt.Printf("Checking resources to create for test suite...\n")

	ovaFilePath := getCurrentDir() + "/../test-resources/" + config.Ova.OvaTestFileName

	if config.Ova.OvaTestFileName == "" && testConfig.VCD.Catalog.CatalogItem == "" {
		panic(fmt.Errorf("ovaTestFileName isn't configured. Tests terminated\n"))
	}

	if config.Ova.OvaDownloadUrl == "" && testConfig.VCD.Catalog.CatalogItem == "" {
		panic(fmt.Errorf("ovaDownloadUrl isn't configured. Tests terminated\n"))
	} else if testConfig.VCD.Catalog.CatalogItem == "" {
		fmt.Printf("Downloading OVA. File will be saved as: %s\n", ovaFilePath)

		if fileExists(ovaFilePath) {
			fmt.Printf("File already exists. Skipping downloading\n")
		} else {
			err := downloadFile(ovaFilePath, testConfig.Ova.OvaDownloadUrl)
			if err != nil {
				panic(err)
			}
			fmt.Printf("OVA downloaded\n")
		}
	}

	vcdClient, err := getTestVCDFromJson(config)
	if vcdClient == nil || err != nil {
		panic(err)
	}
	err = ProviderAuthenticate(vcdClient, config.Provider.User, config.Provider.Password, config.Provider.Token, config.Provider.SysOrg, config.Provider.ApiToken)
	if err != nil {
		panic(err)
	}

	org, err := vcdClient.GetOrgByName(config.VCD.Org)
	if err != nil {
		panic(err)
	}

	var catalog *govcd.Catalog

	catalogPreserved := true
	catalog, err = org.GetCatalogByName(testSuiteCatalogName, false)
	if err != nil {
		catalogPreserved = false
	}

	if testConfig.VCD.Catalog.Name == "" && !catalogPreserved {
		fmt.Printf("Creating catalog for test suite...\n")
		*catalog, err = org.CreateCatalog(testSuiteCatalogName, "Test suite purpose")
		if err != nil || *catalog == (govcd.Catalog{}) {
			panic(err)
		}
		fmt.Printf("Catalog created successfully\n")

	} else if testConfig.VCD.Catalog.Name != "" {
		fmt.Printf("Skipping catalog creation - found preconfigured one: %s \n", testConfig.VCD.Catalog.Name)

		existingCatalog, err := org.GetCatalogByName(testConfig.VCD.Catalog.Name, false)
		if err != nil {
			fmt.Printf("Preconfigured catalog wasn't found \n")
			panic(err)
		}

		catalog = existingCatalog
		fmt.Printf("Catalog found successfully\n")
		testSuiteCatalogName = testConfig.VCD.Catalog.Name
	} else {
		fmt.Printf("Skipping catalog creation - catalog was preserved from previous creation \n")
	}

	catalogItemPreserved := true
	_, err = catalog.GetCatalogItemByName(testSuiteCatalogOVAItem, false)
	if err != nil {
		catalogItemPreserved = false
	}

	if testConfig.VCD.Catalog.CatalogItem == "" && !catalogItemPreserved {
		fmt.Printf("Creating catalog item for test suite...\n")
		task, err := catalog.UploadOvf(ovaFilePath, testSuiteCatalogOVAItem, "Test suite purpose", 20*1024*1024)
		if err != nil {
			fmt.Printf("error uploading new catalog item: %s", err)
			panic(err)
		}

		err = task.ShowUploadProgress()
		if err != nil {
			fmt.Printf("error waiting for task to complete: %+v", err)
			panic(err)
		}

		err = task.WaitTaskCompletion()
		if err != nil {
			fmt.Printf("error waiting for task to complete: %+v", err)
			panic(err)
		}

		fmt.Printf("Catalog item created successfully\n")

	} else if testConfig.VCD.Catalog.CatalogItem != "" {
		fmt.Printf("Skipping catalog item creation - found preconfigured one: %s \n", testConfig.VCD.Catalog.CatalogItem)

		item, err := catalog.GetCatalogItemByName(testConfig.VCD.Catalog.CatalogItem, false)
		if err != nil && item != nil {
			fmt.Printf("Preconfigured catalog item wasn't found \n")
			panic(err)
		}
		fmt.Printf("Catalog item found successfully\n")
		testSuiteCatalogOVAItem = testConfig.VCD.Catalog.CatalogItem
	} else {
		fmt.Printf("Skipping catalog item creation - catalog item was preserved from previous creation \n")
	}

}

// DownloadFile will download a url to a local file. It's efficient because it will
// write as it downloads and not load the whole file into memory.
func downloadFile(filepath string, url string) error {

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

// Creates a VCDClient based on the endpoint given in the TestConfig argument.
// TestConfig struct can be obtained by calling GetConfigStruct. Throws an error
// if endpoint given is not a valid url.
func getTestVCDFromJson(testConfig TestConfig) (*govcd.VCDClient, error) {
	configUrl, err := url.ParseRequestURI(testConfig.Provider.Url)
	if err != nil {
		return &govcd.VCDClient{}, fmt.Errorf("could not parse Url: %s", err)
	}
	vcdClient := govcd.NewVCDClient(*configUrl, true,
		govcd.WithSamlAdfs(testConfig.Provider.UseSamlAdfs, testConfig.Provider.CustomAdfsRptId),
		govcd.WithHttpUserAgent(buildUserAgent("test", testConfig.Provider.SysOrg)))
	return vcdClient, nil
}

func destroySuiteCatalogAndItem(config TestConfig) {
	fmt.Printf("Looking for resources to delete from test suite...\n")
	vcdClient, err := getTestVCDFromJson(config)
	if vcdClient == nil || err != nil {
		panic(err)
	}

	err = ProviderAuthenticate(vcdClient, config.Provider.User, config.Provider.Password, config.Provider.Token, config.Provider.SysOrg, config.Provider.ApiToken)
	if err != nil {
		panic(err)
	}

	org, err := vcdClient.GetOrgByName(config.VCD.Org)
	if err != nil {
		panic(err)
	}

	catalog, err := org.GetCatalogByName(testSuiteCatalogName, false)
	if err != nil {
		fmt.Printf("catalog already removed %#v", err)
		return
	}

	isCatalogDeleted := false
	if testConfig.VCD.Catalog.Name == "" {
		fmt.Printf("Deleting catalog for test suite...\n")
		err = catalog.Delete(true, true)
		if err != nil {
			fmt.Printf("error removing catalog %#v", err)
			return
		}
		isCatalogDeleted = true
		fmt.Printf("Catalog %s removed successfully\n", catalog.Catalog.Name)
	} else {
		fmt.Printf("Catalog deletion skipped as user defined resource used \n")
	}

	if testConfig.VCD.Catalog.CatalogItem == "" && !isCatalogDeleted {
		catalogItem, err := catalog.GetCatalogItemByName(testSuiteCatalogOVAItem, false)
		if err != nil {
			fmt.Printf("error finding catalog item %#v", err)
			return
		}
		err = catalogItem.Delete()
		if err != nil {
			fmt.Printf("error removing catalog item %#v", err)
			return
		}
		fmt.Printf("Catalog %s item removed successfully\n", catalogItem.CatalogItem.Name)
	} else {
		fmt.Printf("Catalog item deletion skipped as user defined resource is used or removed with catalog\n")
	}

}

// Used by resources at the top of the hierarchy (such as Org, ExternalNetwork)
func importStateIdTopHierarchy(objectName string) resource.ImportStateIdFunc {
	return func(*terraform.State) (string, error) {
		return objectName, nil
	}
}

// Used by all entities that depend on Org (such as Catalog, OrgUser)
func importStateIdOrgObject(vcd TestConfig, objectName string) resource.ImportStateIdFunc {
	return func(*terraform.State) (string, error) {
		if testConfig.VCD.Org == "" || objectName == "" {
			return "", fmt.Errorf("missing information to generate import path")
		}
		return testConfig.VCD.Org +
			ImportSeparator +
			objectName, nil
	}
}

// Used by all entities that depend on Org + VDC (such as Vapp, networks, edge gateway)
func importStateIdOrgVdcObject(vcd TestConfig, objectName string) resource.ImportStateIdFunc {
	return func(*terraform.State) (string, error) {
		if testConfig.VCD.Org == "" || testConfig.VCD.Vdc == "" || objectName == "" {
			return "", fmt.Errorf("missing information to generate import path")
		}
		return testConfig.VCD.Org +
			ImportSeparator +
			testConfig.VCD.Vdc +
			ImportSeparator +
			objectName, nil
	}
}

// importStateIdOrgNsxtVdcObject can be used by all entities that depend on Org + NSX-T VDC (such as Vapp, networks,
// edge gateway) in NSX-T VDC
func importStateIdOrgNsxtVdcObject(vcd TestConfig, objectName string) resource.ImportStateIdFunc {
	return func(*terraform.State) (string, error) {
		if testConfig.VCD.Org == "" || testConfig.Nsxt.Vdc == "" || objectName == "" {
			return "", fmt.Errorf("missing information to generate import path")
		}
		return testConfig.VCD.Org +
			ImportSeparator +
			testConfig.Nsxt.Vdc +
			ImportSeparator +
			objectName, nil
	}
}

// importStateIdOrgNsxtVdcGroupObject can be used by all entities that depend on Org + NSX-T VDC
// Group (such as Vapp, networks, edge gateway) in NSX-T VDC
func importStateIdOrgNsxtVdcGroupObject(vcd TestConfig, vdcGroupName, objectName string) resource.ImportStateIdFunc {
	return func(*terraform.State) (string, error) {
		if testConfig.VCD.Org == "" || testConfig.Nsxt.Vdc == "" || objectName == "" {
			return "", fmt.Errorf("missing information to generate import path")
		}
		return testConfig.VCD.Org +
			ImportSeparator +
			vdcGroupName +
			ImportSeparator +
			objectName, nil
	}
}

// importStateIdNsxtManagerObject can be used by all entities that depend on NSX-T manager name + objectName
func importStateIdNsxtManagerObject(vcd TestConfig, objectName string) resource.ImportStateIdFunc {
	return func(*terraform.State) (string, error) {
		if testConfig.Nsxt.Manager == "" || objectName == "" {
			return "", fmt.Errorf("missing information to generate import path")
		}
		return testConfig.Nsxt.Manager + ImportSeparator + objectName, nil
	}
}

// Used by all entities that depend on Org + Catalog (such as catalog item, media item)
func importStateIdOrgCatalogObject(vcd TestConfig, objectName string) resource.ImportStateIdFunc {
	return func(*terraform.State) (string, error) {
		if testConfig.VCD.Org == "" || testConfig.VCD.Catalog.Name == "" || objectName == "" {
			return "", fmt.Errorf("missing information to generate import path")
		}
		return testConfig.VCD.Org +
			ImportSeparator +
			testConfig.VCD.Catalog.Name +
			ImportSeparator +
			objectName, nil
	}
}

// Used by all entities that depend on Org + VDC + vApp (such as VM, vapp networks)
func importStateIdVappObject(vcd TestConfig, vappName, objectName string) resource.ImportStateIdFunc {
	return func(*terraform.State) (string, error) {
		if testConfig.VCD.Org == "" || testConfig.VCD.Vdc == "" || vappName == "" || objectName == "" {
			return "", fmt.Errorf("missing information to generate import path")
		}
		return testConfig.VCD.Org +
			ImportSeparator +
			testConfig.VCD.Vdc +
			ImportSeparator +
			vappName +
			ImportSeparator +
			objectName, nil
	}
}

// Used by all entities that depend on Org + VDC + edge gateway (such as FW, LB, NAT)
func importStateIdEdgeGatewayObject(vcd TestConfig, edgeGatewayName, objectName string) resource.ImportStateIdFunc {
	return func(*terraform.State) (string, error) {
		if testConfig.VCD.Org == "" || testConfig.VCD.Vdc == "" || edgeGatewayName == "" || objectName == "" {
			return "", fmt.Errorf("missing information to generate import path")
		}
		return testConfig.VCD.Org +
			ImportSeparator +
			testConfig.VCD.Vdc +
			ImportSeparator +
			edgeGatewayName +
			ImportSeparator +
			objectName, nil
	}
}

// importStateIdNsxtEdgeGatewayObject used by all entities that depend on Org + NSX-T VDC + edge gateway (such as FW, NAT, Security Groups)
func importStateIdNsxtEdgeGatewayObject(vcd TestConfig, edgeGatewayName, objectName string) resource.ImportStateIdFunc {
	return func(*terraform.State) (string, error) {
		if testConfig.VCD.Org == "" || testConfig.VCD.Vdc == "" || edgeGatewayName == "" || objectName == "" {
			return "", fmt.Errorf("missing information to generate import path for object %s", objectName)
		}
		return testConfig.VCD.Org +
			ImportSeparator +
			testConfig.Nsxt.Vdc +
			ImportSeparator +
			edgeGatewayName +
			ImportSeparator +
			objectName, nil
	}
}

// Used by all entities that depend on Org + VDC + vApp VM (such as VM internal disks)
func importStateIdVmObject(orgName, vdcName, vappName, vmName, objectIdentifier string) resource.ImportStateIdFunc {
	return func(*terraform.State) (string, error) {
		if orgName == "" || vdcName == "" || vappName == "" || vmName == "" || objectIdentifier == "" {
			return "", fmt.Errorf("missing information to generate import path")
		}
		return orgName +
			ImportSeparator +
			vdcName +
			ImportSeparator +
			vappName +
			ImportSeparator +
			vmName +
			ImportSeparator +
			objectIdentifier, nil
	}
}

// importStateIdNsxtEdgeGatewayObjectUsingVdcGroup used by all entities that depend on Org + NSX-T edge gateway (such as IP Sets, Security Groups)
func importStateIdNsxtEdgeGatewayObjectUsingVdcGroup(vdcGroupName, edgeGatewayName, objectName string) resource.ImportStateIdFunc {
	return func(*terraform.State) (string, error) {
		if testConfig.VCD.Org == "" || vdcGroupName == "" || edgeGatewayName == "" || objectName == "" {
			return "", fmt.Errorf("missing information to generate import path for object %s", objectName)
		}
		return testConfig.VCD.Org +
			ImportSeparator +
			vdcGroupName +
			ImportSeparator +
			edgeGatewayName +
			ImportSeparator +
			objectName, nil
	}
}

// setBoolFlag binds a flag to a boolean variable (passed as pointer)
// it also uses an optional environment variable that, if set, will
// update the variable before binding it to the flag.
func setBoolFlag(varPointer *bool, name, envVar, help string) {
	if envVar != "" && os.Getenv(envVar) != "" {
		*varPointer = true
	}
	flag.BoolVar(varPointer, name, *varPointer, help)
}

// setStringFlag binds a flag to a string variable (passed as pointer)
// it also uses an optional environment variable that, if set, will
// update the variable before binding it to the flag.
func setStringFlag(varPointer *string, name, envVar, help string) {
	if envVar != "" && os.Getenv(envVar) != "" {
		*varPointer = os.Getenv(envVar)
	}
	flag.StringVar(varPointer, name, *varPointer, help)
}

type envHelper struct {
	vars map[string]string
}

// newEnvVarHelper helps to initialize
func newEnvVarHelper() *envHelper {
	return &envHelper{vars: make(map[string]string)}
}

// saveVcdVars checks all env vars with VCD prefix and saves them in a map
func (env *envHelper) saveVcdVars() {
	for _, envVar := range os.Environ() {
		if strings.HasPrefix(envVar, "VCD") {
			// os.Environ returns a slice of "key=value" strings. The first "=" separated "key" and
			// "value" therefore we split only first "=" match as env vars may have syntax of
			// "key=value=else"
			splitKeyValue := strings.SplitN(envVar, "=", 2)
			key := splitKeyValue[0]
			value := splitKeyValue[1]
			env.vars[key] = value
		}
	}

}

// unsetVcdVars unsets all environment variables with prefix "VCD"
func (env *envHelper) unsetVcdVars() {
	for keyName := range env.vars {
		os.Unsetenv(keyName)
	}
}

// restoreVcdVars restores all env variables with prefix "VCD" stored in parent struct
func (env *envHelper) restoreVcdVars() {
	for keyName, valueName := range env.vars {
		os.Setenv(keyName, valueName)
	}
}

// importStateIdViaResource runs the import of a VM affinity rule using the resource ID
func importStateIdViaResource(resource string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resource]
		if !ok {
			return "", fmt.Errorf("resource not found: %s", resource)
		}

		if rs.Primary.ID == "" {
			return "", fmt.Errorf("no ID is set for %s resource", resource)
		}

		importId := testConfig.VCD.Org + "." + testConfig.VCD.Vdc + "." + rs.Primary.ID
		if testConfig.VCD.Org == "" || testConfig.VCD.Vdc == "" || rs.Primary.ID == "" {
			return "", fmt.Errorf("missing information to generate import path: %s", importId)
		}
		return importId, nil
	}
}

// testAccFindValuesInSet finds several elements as belonging to the same item in a set
// * resourceName is the complete identifier of the resource (such as vcd_vapp_access_control.Name)
// * prefix is the name of the set (e.g. "shared" in vApp access control)
// * wanted is a map of values to check (such as {"subject_name" : "xxx", "access_level": "yyy"})
// The function returns successfully if all the wanted elements are found within the same set ID
// For example, given the following contents in the resource:
//
//  "shared.2503357709.access_level":"FullControl",
//  "shared.3479897784.user_id":"urn:vcloud:user:ec571e04-7e75-4dc5-8f53-c3ef63b9b414",
//  "shared.2503357709.user_id":"urn:vcloud:user:465308a5-7456-42c8-939c-bd971b0e0d3f",
//  "shared.2503357709.subject_name":"ac-user1",
//  "shared.3479897784.subject_name":"ac-user2",
//  "shared.3479897784.access_level":"Change"
//
// We pass "shared" as prefix, and map[string]string{"subject_name": "ac-user1", "access_level": "FullControl"} as wanted
// The function will match the elements belonging to set "2503357709", and return successfully, because both elements were found.
func testAccFindValuesInSet(resourceName string, prefix string, wanted map[string]string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		var matches = make(map[string]int)
		for key, value := range rs.Primary.Attributes {
			keyList := strings.Split(key, ".")
			if len(keyList) == 3 {
				foundPrefix := keyList[0]
				setID := keyList[1]
				foundKey := keyList[2]
				for wKey, wValue := range wanted {
					if foundPrefix == prefix && foundKey == wKey {
						if wValue == value {
							_, ok := matches[setID]
							if !ok {
								matches[setID] = 0
							}
							matches[setID]++
						}
					}
				}
			}
		}

		for _, value := range matches {
			if value == len(wanted) {
				return nil
			}
		}
		return fmt.Errorf("resource %s - %d matches found - wanted %d", resourceName, len(matches), len(wanted))
	}
}

// skipOnEnvVariable takes a TestCheckFunc and skips it if the given environment variable was set with
// an expected value
func skipOnEnvVariable(envVar, envValue, notes string, f resource.TestCheckFunc) resource.TestCheckFunc {
	if os.Getenv(envVar) == envValue {
		fmt.Printf("### Check skipped at user request - Variable %s - reason: %s\n", envVar, notes)
		return func(s *terraform.State) error {
			return nil
		}
	}
	return f
}

// skipNoConfiguration allows to skip a test if NSX-T configuration is missing
func skipNoConfiguration(t *testing.T, paramsMap StringMap) {
	for key, value := range paramsMap {
		if value == "" {
			t.Skipf("Missing test config: No %s specified", key)
		}
	}
}

func skipNoNsxtAlbConfiguration(t *testing.T) {
	generalMessage := "Missing NSX-T ALB config: "

	if testConfig.Nsxt.NsxtAlbControllerUrl == "" {
		t.Skip(generalMessage + "URL")
	}

	if testConfig.Nsxt.NsxtAlbControllerUser == "" {
		t.Skip(generalMessage + "User")
	}

	if testConfig.Nsxt.NsxtAlbControllerPassword == "" {
		t.Skip(generalMessage + "Password")
	}

	if testConfig.Nsxt.NsxtAlbImportableCloud == "" {
		t.Skip(generalMessage + "Importable Cloud")
	}
}

func testAccCheckVcdStandaloneVmExists(vmName, node, orgName, vdcName string) resource.TestCheckFunc {
	if orgName == "" {
		orgName = testConfig.VCD.Org
	}
	if vdcName == "" {
		vdcName = testConfig.VCD.Vdc
	}
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[node]
		if !ok {
			return fmt.Errorf("not found: %s", node)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no VM ID is set")
		}

		conn := testAccProvider.Meta().(*VCDClient)
		_, vdc, err := conn.GetOrgAndVdc(orgName, vdcName)
		if err != nil {
			return fmt.Errorf(errorRetrievingVdcFromOrg, vdcName, orgName, err)
		}

		_, err = vdc.QueryVmByName(vmName)
		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckVcdStandaloneVmDestroy(vmName string, orgName string, vdcName string) resource.TestCheckFunc {
	if orgName == "" {
		orgName = testConfig.VCD.Org
	}
	if vdcName == "" {
		vdcName = testConfig.VCD.Vdc
	}
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*VCDClient)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "vcd_vm" {
				continue
			}
			_, vdc, err := conn.GetOrgAndVdc(orgName, vdcName)
			if err != nil {
				return fmt.Errorf(errorRetrievingVdcFromOrg, vdcName, orgName, err)
			}

			_, err = vdc.QueryVmByName(vmName)

			if err == nil {
				return fmt.Errorf("VM still exist")
			}

			return nil
		}

		return nil
	}
}

func timeStamp() string {
	now := time.Now()
	return now.Format(time.RFC3339)
}

// preTestChecks is to be called at the beginning of a test function.
// It allows for several skipping mechanisms:
//
// 1) It will skip if the file 'skip_vcd_tests' is found.
//   This allows to interrupt the test suite in  a clean way, by creating the skipping trigger file
//   during the test run
//   When the user creates such file, the tests still running will continue until their natural end
//   and the other tests will skip
//
// 2) if the file 'skip_vcd_tests' contains a pattern, only the tests with a name that match such pattern will skip
//
// 3) It will skip if a test has already run successfully. This is useful when the suite was interrupted,
//   so that we can repeat the run without repeating the tests that have succeeded
//
// 4) It will skip the test if a given environment variable was set
//
// 5) It will skip the test if the option -vcd-skip-pattern or the environment variable 'VCD_SKIP_PATTERN'
//   contains a pattern that matches the test name.
// 6) If the flag -vcd-re-run-failed is true, it will only run the tests that failed in the previous run
func preTestChecks(t *testing.T) {
	// if the test runs without -vcd-pre-post-checks, all post-checks will be skipped
	if !vcdPrePostChecks {
		return
	}
	if vcdShowTimestamp {
		fmt.Printf("Test started at: %s\n", timeStamp())
	}
	if vcdShowElapsedTime {
		elapsed := time.Since(vcdStartTime)
		fmt.Printf("Elapsed: %s\n", elapsed.String())
	}
	if fileExists(vcdSkipAllFile) {
		vcdSkipCount += 1
		t.Skipf("File '%s' found at %s. Test %s skipped", vcdSkipAllFile, timeStamp(), t.Name())
	}
	if vcdSkipPattern != "" {
		re := regexp.MustCompile(vcdSkipPattern)
		if re.MatchString(t.Name()) {
			vcdSkipCount += 1
			t.Skipf("Skip pattern '%s' matches test name '%s'", vcdSkipPattern, t.Name())
		}
	}
	skipEnvVar := fmt.Sprintf("skip-%s", t.Name())

	if vcdTestVerbose {
		fmt.Printf("ENV VAR for %s: %s\n", t.Name(), skipEnvVar)
	}
	if os.Getenv(skipEnvVar) != "" {
		vcdSkipCount += 1
		t.Skipf("variable '%s' was set.", skipEnvVar)
	}
	// If this test has run already, we skip it
	if isTestInFile(t.Name(), "pass") {
		vcdSkipCount += 1
		t.Skipf("test '%s' found in '%s' ", t.Name(), getTestListFile("pass"))
	}
	if vcdReRunFailed {
		if !isTestInFile(t.Name(), "fail") {
			vcdSkipCount += 1
			t.Skip("only running tests that have failed at the previous run")
		}
	}
}

// postTestChecks runs checks after the test
// It performs the following:
// 1) shows a time stamp (if enabled by -vcd-show-timestamp
// 2) stores file name in the "pass" or "fail" list, depending on their outcome. The lists are distinct by VCD IP
// 3) increments the pass/fail counters
func postTestChecks(t *testing.T) {
	// if the test runs without -vcd-pre-post-checks, all post-checks will be skipped
	if !vcdPrePostChecks {
		return
	}
	if vcdShowTimestamp {
		fmt.Printf("Test ended at at: %s\n", timeStamp())
	}
	var err error
	var fileType = "pass"
	if t.Failed() {
		fileType = "fail"
		vcdFailCount += 1
	} else {
		vcdPassCount += 1
	}
	err = addToTestRunList(t.Name(), fileType)
	if err != nil {
		fmt.Printf("WARNING: error adding test name '%s' to file '%s'\n", t.Name(), getTestListFile(fileType))
	}
}

// getTestListFile returns the name of the file containing the wanted (pass/fail) list
// for the VCD being tested
func getTestListFile(fileType string) string {
	if testConfig.Provider.Url == "" {
		return ""
	}
	testingVcdIp := strings.Replace(testConfig.Provider.Url, "https://", "", -1)
	testingVcdIp = strings.Replace(testingVcdIp, "/api", "", -1)
	testingVcdIp = strings.Replace(testingVcdIp, "/", "", -1)
	testingVcdIp = strings.Replace(testingVcdIp, ".", "-", -1)
	return fmt.Sprintf("vcd_test_%s_list-%s.txt", fileType, testingVcdIp)
}

// isTestInFile returns true if a given test name is found in the wanted (pass/fail) list
func isTestInFile(testName, fileType string) bool {
	fileName := getTestListFile(fileType)
	if fileName == "" {
		return false
	}
	runTestRunListFileLock.kvLock(fileName)
	defer runTestRunListFileLock.kvUnlock(fileName)
	if !fileExists(fileName) {
		return false
	}
	f, err := os.Open(fileName) // #nosec G304
	if err != nil {
		return false
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == testName {
			return true
		}
	}
	return false
}

// removeTestRunList removes the wanted (pass/fail) list for the VCD being tested
// This operation is triggered by -vcd-remove-test-list, and it is needed to run
// a test again after running with -vcd-pre-post-checks
func removeTestRunList(fileType string) error {
	fileName := getTestListFile(fileType)
	runTestRunListFileLock.kvLock(fileName)
	defer runTestRunListFileLock.kvUnlock(fileName)
	if fileExists(vcdSkipAllFile) {
		err := os.Remove(vcdSkipAllFile)
		if err != nil {
			return err
		}
	}
	if !fileExists(fileName) {
		fmt.Printf("[removeTestRunList] '%s' not found\n", fileName)
		return nil
	}
	return os.Remove(fileName)
}

// addToTestRunList adds a given test name to a wanted (pass/fail) list
func addToTestRunList(testName, fileType string) error {
	fileName := getTestListFile(fileType)
	if fileName == "" {
		return nil
	}
	runTestRunListFileLock.kvLock(fileName)
	defer runTestRunListFileLock.kvUnlock(fileName)

	var file *os.File
	var err error
	if fileExists(fileName) {
		file, err = os.OpenFile(fileName, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	} else {
		file, err = os.Create(fileName)
	}
	if err != nil {
		return err
	}
	defer file.Close()

	w := bufio.NewWriter(file)
	_, err = fmt.Fprintf(w, "%s\n", testName)
	if err != nil {
		return fmt.Errorf("error writing to file %s: %s", fileName, err)
	}
	return w.Flush()
}

// noTestCredentials helps to check if a config file with credentials is actually provided. It helps to conditionally
// ignore tests in such case
func noTestCredentials() bool {
	return testConfig.Provider.User == ""
}

// skipTestForVcdExactVersion allows to skip tests for specific VCD version
// exactSkipVersion must match exact VCD version (e.g. 10.2.2.17855680)
func skipTestForVcdExactVersion(t *testing.T, exactSkipVersion, skipReason string) {
	vcdClient := createTemporaryVCDConnection(false)

	vcdVersion, err := vcdClient.Client.GetVcdFullVersion()
	if err != nil {
		t.Fatalf("Could not determine VCD version")
	}

	expectedVersion, err := version.NewVersion(exactSkipVersion)
	if err != nil {
		t.Fatalf("could not process versions")
	}
	if vcdVersion.Version.Equal(expectedVersion) {
		t.Skipf("skipping test on VCD version %s because %s", exactSkipVersion, skipReason)
	}
}

func skipTestForApiToken(t *testing.T) {
	if testConfig.Provider.ApiToken != "" {
		t.Skipf("skipping test %s because API token does not support this functionality", t.Name())
	}
}
