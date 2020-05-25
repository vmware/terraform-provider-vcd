// +build binary

package vcd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strings"
	"testing"
)

var saveEnvValues = make(map[string]string)

// Sets the environment variables needed for producing the Terraform files
func setEnvValues() {
	for _, v := range []string{envVcdAddProvider, envVcdRemoveOrgVdcFromTemplate, envVcdSkipTemplateWriting} {
		saveEnvValues[v] = os.Getenv(v)
	}
	_ = os.Setenv(envVcdAddProvider, "1")
	_ = os.Setenv(envVcdRemoveOrgVdcFromTemplate, "")
	_ = os.Setenv(envVcdSkipTemplateWriting, "")
}

// Restore the values that were saved by setEnvValues()
func restoreEnvValues() {
	for _, v := range []string{envVcdAddProvider, envVcdRemoveOrgVdcFromTemplate, envVcdSkipTemplateWriting} {
		_ = os.Setenv(v, saveEnvValues[v])
	}
}

// Reads a file and returns its contents as a string
func readFile(filename string) (string, error) {
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", err
	}
	return string(buf), nil
}

// Fills custom templates with data from configuration file
func TestCustomTemplates(t *testing.T) {
	var binaryTestList []string

	fileList, err := ioutil.ReadDir(customTemplatesDirectory)
	if err != nil {
		t.Skip("could not read files from " + customTemplatesDirectory)
	}
	for _, fileInfo := range fileList {
		extension := path.Ext(fileInfo.Name())
		if extension == ".tf" {
			binaryTestList = append(binaryTestList, fileInfo.Name())
		}
	}

	buildEnvFile := "full-env"
	var params = StringMap{
		"Org":                          testConfig.VCD.Org,
		"Vdc":                          testConfig.VCD.Vdc,
		"ProviderVdc":                  testConfig.VCD.ProviderVdc.Name,
		"NetworkPool":                  testConfig.VCD.ProviderVdc.NetworkPool,
		"StorageProfile":               testConfig.VCD.ProviderVdc.StorageProfile,
		"Catalog":                      testConfig.VCD.Catalog.Name,
		"CatalogItem":                  testConfig.VCD.Catalog.CatalogItem,
		"OvaPath":                      testConfig.Ova.OvaPath,
		"MediaPath":                    testConfig.Media.MediaPath,
		"MediaUploadPieceSize":         testConfig.Media.UploadPieceSize,
		"MediaUploadProgress":          testConfig.Media.UploadProgress,
		"OvaDownloadUrl":               testConfig.Ova.OvaDownloadUrl,
		"OvaTestFileName":              testConfig.Ova.OvaTestFileName,
		"OvaUploadProgress":            testConfig.Ova.UploadProgress,
		"OvaUploadPieceSize":           testConfig.Ova.UploadPieceSize,
		"OvaPreserve":                  testConfig.Ova.Preserve,
		"LoggingEnabled":               testConfig.Logging.Enabled,
		"LoggingFileName":              testConfig.Logging.LogFileName,
		"EdgeGateway":                  testConfig.Networking.EdgeGateway,
		"SharedSecret":                 testConfig.Networking.SharedSecret,
		"ExternalNetwork":              testConfig.Networking.ExternalNetwork,
		"ExternalNetworkPortGroup":     testConfig.Networking.ExternalNetworkPortGroup,
		"ExternalNetworkPortGroupType": testConfig.Networking.ExternalNetworkPortGroupType,
		"ExternalIp":                   testConfig.Networking.ExternalIp,
		"InternalIp":                   testConfig.Networking.InternalIp,
		"Vcenter":                      testConfig.Networking.Vcenter,
		"LocalIp":                      testConfig.Networking.Local.LocalIp,
		"LocalGateway":                 testConfig.Networking.Local.LocalSubnetGateway,
		"PeerIp":                       testConfig.Networking.Peer.PeerIp,
		"PeerGateway":                  testConfig.Networking.Peer.PeerSubnetGateway,
		"MaxRetryTimeout":              testConfig.Provider.MaxRetryTimeout,
		"AllowInsecure":                testConfig.Provider.AllowInsecure,
		"ProviderSysOrg":               testConfig.Provider.SysOrg,
		"ProviderUrl":                  testConfig.Provider.Url,
		"ProviderUser":                 testConfig.Provider.User,
		"ProviderPassword":             testConfig.Provider.Password,
		"ProviderSamlUser":             testConfig.Provider.SamlUser,
		"ProviderSamlPassword":         testConfig.Provider.SamlPassword,
		"ProviderSamlRptId":            testConfig.Provider.SamlCustomRptId,
		"Tags":                         "custom",
		"Prefix":                       "cust",
		"CallerFileName":               "",
		// The following properties are used to create a full environment
		"MainGateway":            testConfig.TestEnvBuild.Gateway,
		"MainNetmask":            testConfig.TestEnvBuild.Netmask,
		"MainDns1":               testConfig.TestEnvBuild.Dns1,
		"MainDns2":               testConfig.TestEnvBuild.Dns2,
		"MediaTestName":          testConfig.TestEnvBuild.MediaName,
		"StorageProfile2":        testConfig.TestEnvBuild.StorageProfile2,
		"ExternalNetworkStartIp": testConfig.TestEnvBuild.ExternalNetworkStartIp,
		"ExternalNetworkEndIp":   testConfig.TestEnvBuild.ExternalNetworkEndIp,
		"RoutedNetwork":          testConfig.TestEnvBuild.RoutedNetwork,
		"IsolatedNetwork":        testConfig.TestEnvBuild.IsolatedNetwork,
		"DirectNetwork":          testConfig.TestEnvBuild.DirectNetwork,
		"OrgUser":                testConfig.TestEnvBuild.OrgUser,
		"OrgUserPassword":        testConfig.TestEnvBuild.OrgUserPassword,
	}

	// optional fields
	if testConfig.TestEnvBuild.MediaName == "" {
		delete(params, "MediaTestName")
	}
	if testConfig.TestEnvBuild.StorageProfile2 == "" {
		delete(params, "StorageProfile2")
	}
	if testConfig.TestEnvBuild.RoutedNetwork == "" {
		delete(params, "RoutedNetwork")
	}
	if testConfig.TestEnvBuild.IsolatedNetwork == "" {
		delete(params, "IsolatedNetwork")
	}
	if testConfig.TestEnvBuild.DirectNetwork == "" {
		delete(params, "DirectNetwork")
	}

	// If either the org user or the password fields are blank, we remove both
	if testConfig.TestEnvBuild.OrgUser == "" || testConfig.TestEnvBuild.OrgUserPassword == "" {
		delete(params, "OrgUser")
		delete(params, "OrgUserPassword")
	}

	for _, fileName := range binaryTestList {

		baseName := strings.Replace(path.Base(fileName), ".tf", "", -1)

		usingBuildFile := baseName == buildEnvFile

		targetScript := path.Join(testArtifactsDirectory, params["Prefix"].(string)+"."+baseName+".tf")
		// It the target file exists, we remove it, as we need a fresh one to be generated
		if fileExists(targetScript) {
			err := os.Remove(targetScript)
			if err != nil {
				t.Logf("error removing %s\n", targetScript)
				panic("can't remove essential file")
			}
		}

		sourceFile := path.Join(getCurrentDir(), customTemplatesDirectory, fileName)
		templateText, err := readFile(sourceFile)
		if err != nil {

			t.Logf("error reading from %s: %s", fileName, err)
		}
		params["FuncName"] = baseName

		reHasProvider := regexp.MustCompile(`(?m)^\s*provider\s+"`)

		// If there is already a provider in the template, we abort
		if reHasProvider.MatchString(templateText) {
			fmt.Printf("File %s has already a provider: remove it and try again\n", sourceFile)
			fmt.Println("The provider will be generated using data from configuration file")
			continue
		}

		setEnvValues()
		params["CallerFileName"] = sourceFile

		if usingBuildFile {
			extraPieces := map[string]string{
				"Routed":   buildEnvRoutedNetwork,
				"Isolated": buildEnvIsolatedNetwork,
				"Direct":   buildEnvDirectNetwork,
			}
			// If any of the optional network names are filled,
			// the corresponding creation template is added
			// to the main one
			for key, value := range extraPieces {
				_, ok := params[key+"Network"]
				if ok {
					templateText = fmt.Sprintf("%s\n%s", templateText, value)
				}
			}

			// If a second storage profile was defined, we add the corresponding
			// text inside the VDC definition
			secondStorageProfileText := ""
			secondStorageParam, ok := params["StorageProfile2"]
			if ok && secondStorageParam != "" {
				secondStorageProfileText = secondStorageProfile
			}
			reSecondStorage := regexp.MustCompile(`#_SECOND_STORAGE_PROFILE_`)
			templateText = reSecondStorage.ReplaceAllString(templateText, secondStorageProfileText)

			// The media item will only be created if its name was defined in the configuration file
			mediaTestText := ""
			mediaTestParam, ok := params["MediaTestName"]
			if ok && mediaTestParam != "" {
				mediaTestText = mediaTest
			}
			reMediaTest := regexp.MustCompile(`#_MEDIA_TEST_`)
			templateText = reMediaTest.ReplaceAllString(templateText, mediaTestText)

			// The Org user will be created only if both user name and password were defined
			orgUserText := ""
			orgUserParam, ok := params["OrgUser"]
			if ok && orgUserParam != "" {
				orgUserText = buildEnvOrgUser
			}
			reOrgUserTest := regexp.MustCompile(`#_ORG_USER_`)
			templateText = reOrgUserTest.ReplaceAllString(templateText, orgUserText)

			// For some items, we want a different value for testing and for building
			// For example, the Ova for testing might be a tiny one, while the one for
			// building the environment would be a beefier one, which can also run the
			// VMware tools.
			if testConfig.TestEnvBuild.ExternalNetworkStartIp == "" {
				params["ExternalNetworkStaticStartIp"] = testConfig.Networking.ExternalIp
				if testConfig.TestEnvBuild.ExternalNetworkEndIp == "" {
					params["ExternalNetworkStaticEndIp"] = testConfig.Networking.ExternalIp
				}
			}
			if testConfig.TestEnvBuild.MediaPath != "" {
				params["MediaPath"] = testConfig.TestEnvBuild.MediaPath
			}
			if testConfig.TestEnvBuild.OvaPath != "" {
				params["OvaPath"] = testConfig.TestEnvBuild.OvaPath
			}
			if testConfig.TestEnvBuild.ExternalNetworkPortGroupType != "" {
				params["ExternalNetworkPortGroupType"] = testConfig.TestEnvBuild.ExternalNetworkPortGroupType
			}
			if testConfig.TestEnvBuild.ExternalNetworkPortGroup != "" {
				params["ExternalNetworkPortGroup"] = testConfig.TestEnvBuild.ExternalNetworkPortGroup
			}
			essentialData := []string{
				"Org", "Vdc", "Catalog", "CatalogItem", "ExternalNetwork", "EdgeGateway",
				"ProviderVdc", "NetworkPool", "StorageProfile", "Vcenter",
				"MainGateway", "MainNetmask", "MainDns1", "ExternalNetworkStartIp", "ExternalNetworkEndIp"}
			for _, essentialItem := range essentialData {
				_, ok := params[essentialItem]
				if ok {
					// By deleting the empty item, we will make sure that the
					// filled script will contain a warning
					if params[essentialItem] == "" {
						delete(params, essentialItem)
					}
				}
			}
		} else {
			params["MediaPath"] = testConfig.Media.MediaPath
			params["OvaPath"] = testConfig.Ova.OvaPath
			params["ExternalNetworkPortGroupType"] = testConfig.Networking.ExternalNetworkPortGroupType
			params["ExternalNetworkPortGroup"] = testConfig.Networking.ExternalNetworkPortGroup
		}

		// We create the configuration text only for the side effect of it being
		// written to the test-artifacts folder
		configText := templateFill(templateText, params)

		// Restore original values in env variables
		restoreEnvValues()
		debugPrintf("%s\n", configText)
		if !fileExists(targetScript) {
			panic(fmt.Sprintf("error: file %s was not generated\n", targetScript))
		}
		fmt.Printf("File: %s\n", targetScript)
	}
}

func init() {
	testingTags["binary"] = "terraform_binary_test.go"
}

// Optional elements used for build environment

const buildEnvOrgUser = `
resource "vcd_org_user" "{{.OrgUser}}" {
  org               = vcd_org.{{.Org}}.name
  name              = "{{.OrgUser}}"
  password          = "{{.OrgUserPassword}}"
  role              = "Organization Administrator"
  enabled           = true
  take_ownership    = true
  provider_type     = "INTEGRATED"
  stored_vm_quota   = 50
  deployed_vm_quota = 50
}
`

const buildEnvRoutedNetwork = `
resource "vcd_network_routed" "{{.RoutedNetwork}}" {
  name         = "{{.RoutedNetwork}}"
  org          = vcd_org.{{.Org}}.name
  vdc          = vcd_org_vdc.{{.Vdc}}.name
  edge_gateway = vcd_edgegateway.{{.EdgeGateway}}.name
  gateway      = "192.168.2.1"

  static_ip_pool {
    start_address = "192.168.2.2"
    end_address   = "192.168.2.100"
  }
}
`

const buildEnvIsolatedNetwork = `
resource "vcd_network_isolated" "{{.IsolatedNetwork}}" {
  name    = "{{.IsolatedNetwork}}"
  org     = vcd_org.{{.Org}}.name
  vdc     = vcd_org_vdc.{{.Vdc}}.name
  gateway = "192.168.3.1"

  static_ip_pool {
    start_address = "192.168.3.2"
    end_address   = "192.168.3.100"
  }
}
`

const buildEnvDirectNetwork = `
resource "vcd_network_direct" "{{.DirectNetwork}}" {
  name             = "{{.DirectNetwork}}"
  org              = vcd_org.{{.Org}}.name
  vdc              = vcd_org_vdc.{{.Vdc}}.name
  external_network = "{{.ExternalNetwork}}"
}
`

const secondStorageProfile = `
  storage_profile {
    name    = "{{.StorageProfile2}}"
    enabled = true
    limit   = 0
    default = false
  }
`

const mediaTest = `
resource "vcd_catalog_media" "{{.MediaTestName}}" {
  org     = vcd_org.{{.Org}}.name
  catalog = vcd_catalog.{{.Catalog}}.name

  name                 = "{{.MediaTestName}}"
  description          = "{{.MediaTestName}}"
  media_path           = "{{.MediaPath}}"
  upload_piece_size    = 5
  show_upload_progress = "true"
}
`
