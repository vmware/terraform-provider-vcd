package vcd

// This module provides initialization routines for the whole suite

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/vmware/go-vcloud-director/util"
	"html/template"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"testing"
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
	} `json:"provider"`
	VCD struct {
		Org     string `json:"org"`
		Vdc     string `json:"vdc"`
		Catalog struct {
			Name        string `json:"name,omitempty"`
			Catalogitem string `json:"catalogItem,omitempty"`
		} `json:"catalog"`
	} `json:"vcd"`
	Networking struct {
		ExternalIp   string `json:"externalIp,omitempty"`
		InternalIp   string `json:"internalIp,omitempty"`
		EdgeGateway  string `json:"edgeGateway,omitempty"`
		SharedSecret string `json:"sharedSecret"`
		Local        struct {
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
		UploadChunkSize int64  `json:"uploadChunkSize,omitempty"`
		UploadProgress  bool   `json:"uploadProgress,omitempty"`
	} `json:"ova"`
}

const (
	// Warning message used for all tests
	acceptanceTestsSkipped = "Acceptance tests skipped unless env 'TF_ACC' set"
)

var (
	// This is a global variable shared across all tests. It contains
	// the information from the configuration file.
	testConfig   TestConfig

	// Enables the short test (used by "make test")
	vcdShortTest bool = os.Getenv("VCD_SHORT_TEST") != ""
)

// Checks if a directory exists
func dirExists(filename string) bool {
	f, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	filemode := f.Mode()
	return filemode.IsDir()
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
		re_org := regexp.MustCompile(`\sorg\s*=`)
		buf2 := re_org.ReplaceAll(buf.Bytes(), []byte("# org = "))
		re_vdc := regexp.MustCompile(`\svdc\s*=`)
		buf2 = re_vdc.ReplaceAll(buf2, []byte("# vdc = "))
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
		writer.Flush()
		file.Close()
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
	var config_struct TestConfig

	// If there was no custom file, we look for the default one
	if config == "" {
		// Finds the current directory, through the path of this running test
		_, current_filename, _, _ := runtime.Caller(0)
		current_directory := filepath.Dir(current_filename)
		config = current_directory + "/vcd_test_config.json"
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
	err = json.Unmarshal(jsonFile, &config_struct)
	if err != nil {
		panic(fmt.Errorf("could not unmarshal json file: %v", err))
	}

	// Reading the configuration file was successful.
	// Now we fill the environment variables that the library is using for its own initialization.
	if config_struct.Provider.TerraformAcceptanceTests {
		// defined in vendor/github.com/hashicorp/terraform/helper/resource/testing.go
		os.Setenv("TF_ACC", "1")
	}
	// The following variables are used in ./provider.go
	if config_struct.Provider.SysOrg == "" {
		config_struct.Provider.SysOrg = config_struct.VCD.Org
	}
	os.Setenv("VCD_USER", config_struct.Provider.User)
	os.Setenv("VCD_PASSWORD", config_struct.Provider.Password)
	os.Setenv("VCD_URL", config_struct.Provider.Url)
	os.Setenv("VCD_SYS_ORG", config_struct.Provider.SysOrg)
	os.Setenv("VCD_ORG", config_struct.VCD.Org)
	os.Setenv("VCD_VDC", config_struct.VCD.Vdc)
	if config_struct.Provider.AllowInsecure {
		os.Setenv("VCD_ALLOW_UNVERIFIED_SSL", "1")
	}

	// Define logging parameters if enabled
	if config_struct.Logging.Enabled {
		util.EnableLogging = true
		if config_struct.Logging.LogFileName != "" {
			util.ApiLogFileName = config_struct.Logging.LogFileName
		}
		if config_struct.Logging.LogHttpResponse {
			util.LogHttpResponse = true
		}
		if config_struct.Logging.LogHttpRequest {
			util.LogHttpRequest = true
		}
		util.InitLogging()
	}
	return config_struct
}

// This function is called before any other test
func TestMain(m *testing.M) {
	// Fills the configuration variable: it will be available to all tests,
	// or the whole suite will fail if it is not found.
	// If VCD_SHORT_TEST is defined, it means that "make test" is called,
	// and we won't really run any tests involving vcd connections.
	if !vcdShortTest {
		testConfig = getConfigStruct()
	}

	// Runs all test functions
	exitCode := m.Run()

	// TODO: cleanup leftovers
	os.Exit(exitCode)
}
