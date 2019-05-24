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
		"Tags":                         "custom",
		"Prefix":                       "cust",
		"CallerFileName":               "",
	}
	for _, fileName := range binaryTestList {

		baseName := strings.Replace(path.Base(fileName), ".tf", "", -1)

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
