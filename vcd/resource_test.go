// +build ALL

package vcd

import (
	"bytes"
	"fmt"
	"html/template"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/terraform"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

var testNotFoundMap map[string]*testNotFoundDataSet

// testDeleteFunc is a
type testDeleteFunc func(t *testing.T, itemName string) func()

type testNotFoundDataSet struct {
	// deleteFunc matches signature of type testDeleteFunc and will handle the deletion of item defined in `deleteItem`
	// field
	deleteFunc testDeleteFunc

	// Two below parameters are then ones which are used throughout all tests
	//
	// params holds the variables which are to be injected into template
	params StringMap
	// config holds the .tf configuration with replaceable variables defined
	config string
}

func init() {
	// configFile := getConfigFileName()
	// if configFile != "" {
	// 	testConfig = getConfigStruct(configFile)
	// }

}

// Validate that all resources have functions defined
// func init() {
// 	for definedResourceName := range globalResourceMap {
// 		functionSet, resourceExists := testNotFoundMap[definedResourceName]
// 		if !resourceExists {
// 			panic(fmt.Errorf("no function defined for resource '%s'", definedResourceName))
// 		}
//
// 		if functionSet == nil {
// 			panic(fmt.Errorf("empty function data set for resource '%s'", definedResourceName))
// 		}
// 	}
// }

func setupList() {

	mainItemName := "ReadTest"

	// validate if all resources have functions mapped
	testNotFoundMap = make(map[string]*testNotFoundDataSet)
	testNotFoundMap["vcd_catalog"] = &testNotFoundDataSet{
		deleteFunc: testDeleteExistingCatalog,
		config:     testAccCheckVcdCatalogBasic,
		params: StringMap{
			"CatalogName": mainItemName,
			"Org":         testConfig.VCD.Org,
		},
	}

	testNotFoundMap["vcd_catalog_item"] = &testNotFoundDataSet{
		deleteFunc: testDeleteExistingCatalogItem,
		config:     testAccCheckVcdCatalogItemBasic,
		params: StringMap{
			"CatalogItemName": mainItemName,
			"Org":             testConfig.VCD.Org,
			"Catalog":         testConfig.VCD.Catalog.Name,
			"Description":     TestAccVcdCatalogItemDescription,
			"OvaPath":         testConfig.Ova.OvaPath,
			"UploadPieceSize": testConfig.Ova.UploadPieceSize,
			"UploadProgress":  testConfig.Ova.UploadProgress,
			"Tags":            "catalog",
		},
	}

	testNotFoundMap["vcd_catalog_media"] = &testNotFoundDataSet{
		deleteFunc: testDeleteExistingCatalogMedia,
		config:     testAccCheckVcdCatalogMediaBasic,
		params: StringMap{
			"CatalogMediaName": mainItemName,
			"Org":              testConfig.VCD.Org,
			"Catalog":          testConfig.VCD.Catalog.Name,
			"Description":      TestAccVcdCatalogMediaDescription,
			"MediaPath":        testConfig.Media.MediaPath,
			"UploadPieceSize":  testConfig.Media.UploadPieceSize,
			"UploadProgress":   testConfig.Media.UploadProgress,
			"Tags":             "catalog",
		},
	}

	testNotFoundMap["vcd_dnat"] = &testNotFoundDataSet{
		deleteFunc: testDeleteExistingDnatRule,
		config:     testAccCheckVcdDnatWithOrgNetw,
		params: StringMap{
			"DnatName":          mainItemName,
			"Org":               testConfig.VCD.Org,
			"Vdc":               testConfig.VCD.Vdc,
			"EdgeGateway":       testConfig.Networking.EdgeGateway,
			"ExternalIp":        testConfig.Networking.ExternalIp,
			"OrgVdcNetworkName": orgVdcNetworkName,
			"Gateway":           "10.10.102.1",
			"StartIpAddress":    "10.10.102.51",
			"EndIpAddress":      "10.10.102.100",
			"Tags":              "gateway",
		},
	}

}

func TestAccVcdResourceNotFound(t *testing.T) {
	// No point for running these tests when we don't have connection
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	setupList()

	// for providerResourceName := range globalResourceMRap { // With validation
	for providerResourceName := range testNotFoundMap {
		functionSet, resourceExists := testNotFoundMap[providerResourceName]
		if !resourceExists {
			t.Errorf("resource '%s' does not have tests", providerResourceName)
		}

		// Run subtests
		t.Run(providerResourceName, oneResourceTestRunner(t, providerResourceName, functionSet))

	}

}

// oneResourceTestRunner is meant to only check if plan\apply deleted outside of Terraform does not fail, but rather
// proposes to recreate the item.
func oneResourceTestRunner(t *testing.T, subTestName string, notFoundData *testNotFoundDataSet) func(t *testing.T) {
	return func(t *testing.T) {
		// No point for running test when we don't have connection
		if vcdShortTest {
			t.Skip(acceptanceTestsSkipped)
			return
		}

		var configTextBytes bytes.Buffer
		configTemplate := template.Must(template.New("letter").Parse(notFoundData.config))
		err := configTemplate.Execute(&configTextBytes, notFoundData.params)
		if err != nil {
			t.Errorf("could not generate template for '%s' resource", subTestName)
		}

		configText := configTextBytes.String()

		// Extract _resource_address_ from definition like resource 'vcd_xxx" "_resource_address_" {'
		resourceAddress := func(vcdName string) string {
			var rgx = regexp.MustCompile(`resource\s+"` + vcdName + `"\s+"(\w*)"`)
			rs := rgx.FindStringSubmatch(configText)
			return vcdName + "." + rs[1]
		}(subTestName)

		debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

		idStore := testCachedFieldValue{}

		wrappedDelete := func() {
			notFoundData.deleteFunc(t, idStore.fieldValue)()
			return
		}

		resource.Test(t, resource.TestCase{
			PreCheck:  func() { testAccPreCheck(t) },
			Providers: testAccProviders,
			Steps: []resource.TestStep{
				resource.TestStep{
					Config: configText,
					// Ensure the ID was set to make sure object was created
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrSet(resourceAddress, "id"),
						idStore.cacheTestResourceFieldValue(resourceAddress, "id"),
					),
				},
				resource.TestStep{
					Config:    configText,
					PlanOnly:  true,
					PreConfig: wrappedDelete,
					// It should offer to recreate during refresh when an object does not exist on apply
					ExpectNonEmptyPlan: true,
				},
			},
		})
	}
}

type testCachedFieldValue struct {
	fieldValue string
}

// cacheTestResourceFieldValue has the same signature as builtin Terraform Test functions, however
// it is attached to a struct which allows to store a field value and then check against this value
// with 'testCheckCachedResourceFieldValue'
func (c *testCachedFieldValue) cacheTestResourceFieldValue(resource, field string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("resource not found: %s", resource)
		}

		value, exists := rs.Primary.Attributes[field]
		if !exists {
			return fmt.Errorf("field %s in resource %s does not exist", field, resource)
		}
		// Store the value in cache
		c.fieldValue = value
		return nil
	}
}

// testCheckCachedResourceFieldValue has the default signature of Terraform acceptance test
// functions, but is able to verify if the value is equal to previously cached value using
// 'cacheTestResourceFieldValue'. This allows to check if a particular field value changed across
// multiple resource.TestSteps.
func (c *testCachedFieldValue) testCheckCachedResourceFieldValue(resource, field string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("resource not found: %s", resource)
		}

		value, exists := rs.Primary.Attributes[field]
		if !exists {
			return fmt.Errorf("field %s in resource %s does not exist", field, resource)
		}

		if value != c.fieldValue {
			return fmt.Errorf("got '%s - %s' field value %s, expected: %s",
				resource, field, value, c.fieldValue)
		}

		return nil
	}
}
