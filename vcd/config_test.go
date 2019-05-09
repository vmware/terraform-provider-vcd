// +build api functional catalog vapp network extnetwork org query vm vdc gateway disk ALL

package vcd

// This module provides initialization routines for the whole suite

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
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

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/util"
)

type StringMap map[string]interface{}

// Structure to get info from a config json file that the user specifies
type TestConfig struct {
	Provider struct {
		User                     string `json:"user"`
		Password                 string `json:"password"`
		Url                      string `json:"url"`
		SysOrg                   string `json:"sysOrg"`
		AllowInsecure            bool   `json:"allowInsecure"`
		TerraformAcceptanceTests bool   `json:"tfAcceptanceTests"`
		UseVcdConnectionCache    bool   `json:"useVcdConnectionCache"`
		MaxRetryTimeout          int    `json:"maxRetryTimeout"`
	} `json:"provider"`
	VCD struct {
		Org     string `json:"org"`
		Vdc     string `json:"vdc"`
		Catalog struct {
			Name        string `json:"name,omitempty"`
			CatalogItem string `json:"catalogItem,omitempty"`
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
	Logging struct {
		Enabled         bool   `json:"enabled,omitempty"`
		LogFileName     string `json:"logFileName,omitempty"`
		LogHttpRequest  bool   `json:"logHttpRequest,omitempty"`
		LogHttpResponse bool   `json:"logHttpResponse,omitempty"`
	} `json:"logging"`
	Ova struct {
		OvaPath         string `json:"ovaPath,omitempty"`
		UploadPieceSize int64  `json:"uploadPieceSize,omitempty"`
		UploadProgress  bool   `json:"uploadProgress,omitempty"`
		OvaTestFileName string `json:"ovaTestFileName,omitempty"`
		OvaDownloadUrl  string `json:"ovaDownloadUrl,omitempty"`
		Preserve        bool   `json:"preserve,omitempty"`
	} `json:"ova"`
	Media struct {
		MediaPath       string `json:"mediaPath,omitempty"`
		UploadPieceSize int64  `json:"uploadPieceSize,omitempty"`
		UploadProgress  bool   `json:"uploadProgress,omitempty"`
	} `json:"media"`
	EnvVariables map[string]string `json:"envVariables,omitempty"`
}

// names for created resources for all the tests
var (
	testSuiteCatalogName    = "TestSuiteCatalog"
	testSuiteCatalogOVAItem = "TestSuiteOVA"
)

const (
	// Warning message used for all tests
	acceptanceTestsSkipped = "Acceptance tests skipped unless env 'TF_ACC' set"
	// This template will be added to test resource snippets on demand
	providerTemplate = `
provider "vcd" {
  user                 = "{{.User}}"
  password             = "{{.Password}}"
  url                  = "{{.Url}}"
  sysorg               = "{{.SysOrg}}"
  org                  = "{{.Org}}"
  allow_unverified_ssl = "{{.AllowInsecure}}"
  version              = "~> {{.VersionRequired}}"
}

`
)

var (
	// This library major version
	currentProviderVersion string = getMajorVersion()
	// This is a global variable shared across all tests. It contains
	// the information from the configuration file.
	testConfig TestConfig

	// Enables the short test (used by "make test")
	vcdShortTest bool = os.Getenv("VCD_SHORT_TEST") != ""
)

// Checks if a directory exists
func dirExists(filename string) bool {
	f, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	fileMode := f.Mode()
	return fileMode.IsDir()
}

// Returns true if the current configuration uses a system administrator for connections
func usingSysAdmin() bool {
	return strings.ToLower(testConfig.Provider.SysOrg) == "system"
}

// Fills a template with data provided as a StringMap
// Returns the text of a ready-to-use Terraform directive.
// It also saves the filled template to a file, for further troubleshooting.
func templateFill(tmpl string, data StringMap) string {

	// Gets the name of the function containing the template
	caller := callFuncName()
	// Removes the full path to the function, leaving only package + function name
	caller = filepath.Base(caller)

	// If the call comes from a function that does not have a good descriptive name,
	// (for example when it's an auxiliary function that builds the template but does not
	// run the test) users can add the function name in the data, and it will be used instead of
	// the caller name.
	funcName, ok := data["FuncName"]
	if ok {
		caller = "vcd." + funcName.(string)
	}

	// If requested, the provider defined in testConfig will be added to test snippets.
	if os.Getenv("VCD_ADD_PROVIDER") != "" {
		// the original template is prefixed with the provider template
		tmpl = providerTemplate + tmpl

		// The data structure used to fill the template is integrated with
		// provider data
		data["User"] = testConfig.Provider.User
		data["Password"] = testConfig.Provider.Password
		data["Url"] = testConfig.Provider.Url
		data["SysOrg"] = testConfig.Provider.SysOrg
		data["Org"] = testConfig.VCD.Org
		data["AllowInsecure"] = testConfig.Provider.AllowInsecure
		data["VersionRequired"] = currentProviderVersion
	}

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
	if os.Getenv("VCD_SKIP_TEMPLATE_WRITING") != "" {
		TemplateWriting = false
	}
	var writeStr []byte = buf.Bytes()

	// This is a quick way of enabling an alternate testing mode:
	// When REMOVE_ORG_VDC_FROM_TEMPLATE is set, the terraform
	// templates will be changed on-the-fly, to comment out the
	// definitions of org and vdc. This will force the test to
	// borrow org and vcd from the provider.
	if os.Getenv("REMOVE_ORG_VDC_FROM_TEMPLATE") != "" {
		reOrg := regexp.MustCompile(`\sorg\s*=`)
		buf2 := reOrg.ReplaceAll(buf.Bytes(), []byte("# org = "))
		reVdc := regexp.MustCompile(`\svdc\s*=`)
		buf2 = reVdc.ReplaceAll(buf2, []byte("# vdc = "))
		writeStr = buf2
	}
	if TemplateWriting {
		testArtifacts := "test-artifacts"
		if !dirExists(testArtifacts) {
			err := os.Mkdir(testArtifacts, 0755)
			if err != nil {
				panic(fmt.Errorf("error creating directory %s: %s", testArtifacts, err))
			}
		}
		resourceFile := path.Join(testArtifacts, caller) + ".tf"
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
			panic(fmt.Errorf("error writing to file %s. Reported %d bytes written. %s", resourceFile, count, err))
		}
		_ = file.Close()
	}
	// Returns the populated template
	return string(writeStr)
}

// Reads the configuration file and returns its contents as a TestConfig structure
// The default file is called vcd_test_config.json in the same directory where
// the test files are.
// Users may define a file in a different location using the environment variable
// VCD_CONFIG
// This function doesn't return an error. It panics immediately because its failure
// will prevent the whole test suite from running
func getConfigStruct() TestConfig {
	// First, we see whether the user has indicated a custom configuration file
	// from a non-standard location
	config := os.Getenv("VCD_CONFIG")
	var configStruct TestConfig

	// If there was no custom file, we look for the default one
	if config == "" {
		config = getCurrentDir() + "/vcd_test_config.json"
	}
	// Looks if the configuration file exists before attempting to read it
	_, err := os.Stat(config)
	if os.IsNotExist(err) {
		panic(fmt.Errorf("configuration file %s not found: %s", config, err))
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
	_ = os.Setenv("VCD_USER", configStruct.Provider.User)
	_ = os.Setenv("VCD_PASSWORD", configStruct.Provider.Password)
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
	return configStruct
}

// Finds the current directory, through the path of this running test
func getCurrentDir() string {
	_, currentFilename, _, _ := runtime.Caller(0)
	return filepath.Dir(currentFilename)
}

// This function is called before any other test
func TestMain(m *testing.M) {
	// Fills the configuration variable: it will be available to all tests,
	// or the whole suite will fail if it is not found.
	// If VCD_SHORT_TEST is defined, it means that "make test" is called,
	// and we won't really run any tests involving vcd connections.
	if !vcdShortTest {
		testConfig = getConfigStruct()

		// Provider initialization moved here from provider_test.init
		testAccProvider = Provider().(*schema.Provider)
		testAccProviders = map[string]terraform.ResourceProvider{
			"vcd": testAccProvider,
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

	// TODO: cleanup leftovers
	os.Exit(exitCode)
}

//Creates catalog and/or catalog item if they are not preconfigured.
func createSuiteCatalogAndItem(config TestConfig) {
	fmt.Printf("Checking resources to create for test suite...\n")

	ovaFilePath := getCurrentDir() + "/../test-resources/" + config.Ova.OvaTestFileName

	if config.Ova.OvaTestFileName == "" && testConfig.VCD.Catalog.CatalogItem == "" {
		panic(fmt.Errorf("ovaTestFileName isn't configured. Tests aborted\n"))
	}

	if config.Ova.OvaDownloadUrl == "" && testConfig.VCD.Catalog.CatalogItem == "" {
		panic(fmt.Errorf("ovaDownloadUrl isn't configured. Tests aborted\n"))
	} else if testConfig.VCD.Catalog.CatalogItem == "" {
		fmt.Printf("Downloading OVA. File will be saved as: %s\n", ovaFilePath)

		if _, err := os.Stat(ovaFilePath); err == nil {
			fmt.Printf("File already exists. Skipping downloading\n")
		} else if os.IsNotExist(err) {
			err := downloadFile(ovaFilePath, testConfig.Ova.OvaDownloadUrl)
			if err != nil {
				panic(err)
			}
			fmt.Printf("OVA downloaded\n")
		} else {
			panic(err)
		}
	}

	vcdClient, err := getTestVCDFromJson(config)
	if vcdClient == nil || err != nil {
		panic(err)
	}
	err = vcdClient.Authenticate(config.Provider.User, config.Provider.Password, config.Provider.SysOrg)
	if err != nil {
		panic(err)
	}

	org, err := govcd.GetOrgByName(vcdClient, config.VCD.Org)
	if err != nil || org == (govcd.Org{}) {
		panic(err)
	}

	var catalog govcd.Catalog

	catalogPreserved := true
	catalog, err = org.FindCatalog(testSuiteCatalogName)
	if err != nil || catalog == (govcd.Catalog{}) {
		catalogPreserved = false
	}

	if testConfig.VCD.Catalog.Name == "" && !catalogPreserved {
		fmt.Printf("Creating catalog for test suite...\n")
		catalog, err = org.CreateCatalog(testSuiteCatalogName, "Test suite purpose")
		if err != nil || catalog == (govcd.Catalog{}) {
			panic(err)
		}
		fmt.Printf("Catalog created successfully\n")

	} else if testConfig.VCD.Catalog.Name != "" {
		fmt.Printf("Skipping catalog creation - found preconfigured one: %s \n", testConfig.VCD.Catalog.Name)

		catalog, err = org.FindCatalog(testConfig.VCD.Catalog.Name)
		if err != nil || catalog == (govcd.Catalog{}) {
			fmt.Printf("Preconfigured catalog wasn't found \n")
			panic(err)
		}

		fmt.Printf("Catalog found successfully\n")
		testSuiteCatalogName = testConfig.VCD.Catalog.Name
	} else {
		fmt.Printf("Skipping catalog creation - catalog was preserved from previous creation \n")
	}

	catalogItemPreserved := true
	catalogItem, err := catalog.FindCatalogItem(testSuiteCatalogOVAItem)
	if err != nil || catalogItem == (govcd.CatalogItem{}) {
		catalogItemPreserved = false
	}

	if testConfig.VCD.Catalog.CatalogItem == "" && !catalogItemPreserved {
		fmt.Printf("Creating catalog item for test suite...\n")
		task, err := catalog.UploadOvf(ovaFilePath, testSuiteCatalogOVAItem, "Test suite purpose", 20*1024*1024)
		if err != nil {
			fmt.Printf("error uploading new catalog item: %#v", err)
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

		item, err := catalog.FindCatalogItem(testConfig.VCD.Catalog.CatalogItem)
		if err != nil && item != (govcd.CatalogItem{}) {
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
	vcdClient := govcd.NewVCDClient(*configUrl, true)
	return vcdClient, nil
}

func destroySuiteCatalogAndItem(config TestConfig) {
	fmt.Printf("Looking for resources to delete from test suite...\n")
	vcdClient, err := getTestVCDFromJson(config)
	if vcdClient == nil || err != nil {
		panic(err)
	}

	err = vcdClient.Authenticate(config.Provider.User, config.Provider.Password, config.Provider.SysOrg)
	if err != nil {
		panic(err)
	}

	org, err := govcd.GetOrgByName(vcdClient, config.VCD.Org)
	if err != nil || org == (govcd.Org{}) {
		panic(err)
	}

	catalog, err := org.FindCatalog(testSuiteCatalogName)
	if err != nil || catalog == (govcd.Catalog{}) {
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
		catalogItem, err := catalog.FindCatalogItem(testSuiteCatalogOVAItem)
		if err != nil || catalogItem == (govcd.CatalogItem{}) {
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

// Reads the version from the VERSION file in the root directory
func getMajorVersion() string {

	versionFile := path.Join(getCurrentDir(), "..", "VERSION")

	// Checks whether the VERSION file exists
	_, err := os.Stat(versionFile)
	if os.IsNotExist(err) {
		panic("Could not find VERSION file")
	}

	// Reads the version from the file
	versionText, err := ioutil.ReadFile(versionFile)
	if err != nil {
		panic(fmt.Errorf("could not read VERSION file %s: %v", versionFile, err))
	}

	// The version is expected to be in the format v#.#.#
	// We only need the first two numbers
	reVersion := regexp.MustCompile(`v(\d+\.\d+)\.\d+`)
	versionList := reVersion.FindAllStringSubmatch(string(versionText), -1)
	if versionList == nil || len(versionList) == 0 {
		panic("empty or non-formatted version found in VERSION file")
	}
	if versionList[0] == nil || len(versionList[0]) < 2 {
		panic("unable to extract major version from VERSION file")
	}
	// A successful match will look like
	// [][]string{[]string{"v2.0.0", "2.0"}}
	// Where the first element is the full text matched, and the second one is the first captured text
	return versionList[0][1]
}
