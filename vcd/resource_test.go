// +build ALL

package vcd

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/terraform"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// testResourceNotFoundMap holds a map of definitions for all resources defined in provider
var testResourceNotFoundMap map[string]*testResourceNotFound

// testResourceDeleteFunc is a function which accepts an itemIdentifier and returns a function which is capable of
// deleting resource by that identifier upon call
type testResourceDeleteFunc func(t *testing.T, itemIdentifier string) func()

// testResourceNotFound is a structure which consists of all possible
type testResourceNotFound struct {
	// deleteFunc matches signature of type testResourceDeleteFunc and will handle the deletion of specific
	deleteFunc testResourceDeleteFunc

	// Two below parameters are then ones which are used throughout all tests
	//
	// params holds the variables which are to be injected into template
	params StringMap
	// config holds the .tf configuration with replaceable variables defined
	config string
}

// defineNotFoundTests must contain definitions for all resources which are present in 'globalResourceMap'
func defineNotFoundTests() {

	mainItemName := "ReadTest"

	// validate if all resources have functions mapped
	testResourceNotFoundMap = make(map[string]*testResourceNotFound)
	testResourceNotFoundMap["vcd_catalog"] = &testResourceNotFound{
		deleteFunc: testDeleteExistingCatalog,
		config:     testAccCheckVcdCatalogBasic,
		params: StringMap{
			"CatalogName": mainItemName,
			"Org":         testConfig.VCD.Org,
		},
	}

	testResourceNotFoundMap["vcd_catalog_item"] = &testResourceNotFound{
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

	testResourceNotFoundMap["vcd_catalog_media"] = &testResourceNotFound{
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

	testResourceNotFoundMap["vcd_dnat"] = &testResourceNotFound{
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
		},
	}

}

// TestAccVcdResourceNotFound loops over all resources defined in provider `globalResourceMap` and checks that there is
// a corresponding ResourceNotFound test defined in `testResourceNotFoundMap`. If not - it fails the test. Then for each
// definition it runs a 'singleResourceNotFoundTest' test.
func TestAccVcdResourceNotFound(t *testing.T) {
	// No point for running these tests when we don't have connection
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	defineNotFoundTests()

	// for providerResourceName := range globalResourceMap { // PRODUCTION. validate against all resources defined in provider
	for providerResourceName := range testResourceNotFoundMap { // DEVELOPMENT ONLY - run tests only for resources which have test structure defined
		functionSet, resourceExists := testResourceNotFoundMap[providerResourceName]
		if !resourceExists {
			t.Errorf("resource '%s' does not have tests", providerResourceName)
		}

		// Run a sub-test for a specific resource
		t.Run(providerResourceName, singleResourceNotFoundTest(t, providerResourceName, functionSet))

	}

}

// extractResourceAddress extracts resource address in format _resource_type_._resource_name_ (e.g. vcd_catalog.my-catalog)
// which is useful in acceptance test addressing
func extractResourceAddress(resourceType, configText string) string {
	var rgx = regexp.MustCompile(`resource\s+"` + resourceType + `"\s+"(\w*)"`)
	rs := rgx.FindStringSubmatch(configText)
	return resourceType + "." + rs[1]
}

// singleResourceNotFoundTest runs a NotFound test by specified data. It has the following workflow:
// 1. Creates a resource as defined in  testResourceNotFound.config parameter filled by testResourceNotFound.params
// parameters. Captures its ID as well.
// 2. Uses a supplied 'testResourceNotFound.deleteFunc' to delete the resource without using Terraform by the captured
// ID in step 1.
// 3. Runs apply (in acceptance test step 1) and expects a non empty plan which means that a resource must be recreated
// because it was not found
func singleResourceNotFoundTest(t *testing.T, subTestName string, notFoundData *testResourceNotFound) func(t *testing.T) {
	return func(t *testing.T) {
		params := notFoundData.params
		// Setting unique name to have a binary test file created for debugging if needed
		params["FuncName"] = "NotFound-" + subTestName
		// Adding skip directive as running these tests in binary test mode add no value
		binaryTestSkipText := "# skip-binary-test: resource not found test only works in acceptance tests\n"
		configText := templateFill(binaryTestSkipText+notFoundData.config, params)

		// Extract _resource_address_ from .tf definition like resource 'vcd_xxx" "_resource_address_" {'
		resourceAddress := extractResourceAddress(subTestName, configText)
		debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

		// Initialize a cached field to capture provisioned "ID" field of resource which is being tested
		idStore := testCachedFieldValue{}

		// Make a closure of 'deleteFunc' so that idStore.fieldValue can be evaluate at step1 (when it is already filled)
		deleteResourceWithId := func() {
			notFoundData.deleteFunc(t, idStore.fieldValue)()
			return
		}

		resource.Test(t, resource.TestCase{
			PreCheck:  func() { testAccPreCheck(t) },
			Providers: testAccProviders,
			Steps: []resource.TestStep{
				// Step 0 - create a resource to be deleted and capture its ID
				resource.TestStep{
					Config: configText,
					// in Step 1
					Check: resource.ComposeAggregateTestCheckFunc(
						// Ensure the ID is set
						resource.TestCheckResourceAttrSet(resourceAddress, "id"),
						// Store the ID for usage in step 1
						idStore.cacheTestResourceFieldValue(resourceAddress, "id"),
					),
				},
				// Step 1 - use 'PreConfig' function to 'delete' the resource created in Step 0 and expect and non empty
				// plan
				resource.TestStep{
					Config:    configText,
					PlanOnly:  true,
					PreConfig: deleteResourceWithId,
					// It should offer to recreate during refresh when an object does not exist on apply
					ExpectNonEmptyPlan: true,
				},
			},
		})
	}
}

///////////
/////////// NOT FOR REVIEW BEYOND THIS LINE. THIS WILL COME IN HERE WITH ANOTHER PR
///////////
///////////
///////////
///////////
///////////

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
