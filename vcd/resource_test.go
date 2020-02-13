// +build ALL functional

package vcd

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// testResourceNotFoundTestMap holds a map of definitions for all resources defined in provider
var testResourceNotFoundTestMap = make(map[string]*testResourceNotFound)

// testResourceNotFound is a structure which consists of all possible
type testResourceNotFound struct {
	// deleteFunc is a functions with schema.DeleteFunc behavior but having a signature of locally defined
	// interface 'vcdResourceDataInterface'
	deleteFunc func(d vcdResourceDataInterface, meta interface{}) error

	// Two below parameters are then ones which are used throughout all tests
	//
	// params holds the variables which are to be injected into template
	params StringMap
	// config holds the .tf configuration with replaceable variables defined
	config string
}

// registerFuncStack makes a slice of functions for delayed call. Main reason for this delayed call is to have late
// evaluation of variables which are populated later than init() functions occur
var registerFuncStack = []func(){}

// Push a function to registerFuncStack for late evaluation
func registerReadTest(f func()) {
	registerFuncStack = append(registerFuncStack, f)
}

// TestAccVcdResourceNotFound loops over all resources defined in provider `globalResourceMap` and checks that there is
// a corresponding ResourceNotFound test defined in `testResourceNotFoundTestMap`. If not - it fails the test. Then for each
// definition it runs a 'singleResourceNotFoundTest' test.
func TestAccVcdResourceNotFound(t *testing.T) {
	// No point for running these tests when we don't have connection
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	// Execute all funcs
	for _, f := range registerFuncStack {
		f()
	}

	// for providerResourceName := range globalResourceMap { // PRODUCTION. validate against all resources defined in provider
	for providerResourceName := range testResourceNotFoundTestMap { // DEVELOPMENT ONLY - run tests only for resources which have test structure defined
		functionSet, resourceExists := testResourceNotFoundTestMap[providerResourceName]
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
		params["FuncName"] = "NotFoundResource-" + subTestName
		// Adding skip directive as running these tests in binary test mode add no value
		binaryTestSkipText := "# skip-binary-test: resource not found test only works in acceptance tests\n"
		configText := templateFill(binaryTestSkipText+notFoundData.config, params)

		// Extract _resource_address_ from .tf definition like resource 'vcd_xxx" "_resource_address_" {'
		resourceAddress := extractResourceAddress(subTestName, configText)
		debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)
		// fmt.Println(configText)

		// Initialize a cached field to capture provisioned "ID" field of resource which is being tested
		idStore := testCachedFieldValue{}

		// Make a closure of 'deleteFunc' so that idStore.fieldValue can be evaluated at step1 (when it is already
		// filled in step 0)
		deleteResourceWithId := func() {
			vcdClient := createTemporaryVCDConnection()
			d := schemaResourceData{id: idStore.fieldValue, configText: configText, org: vcdClient.Org, vdc: vcdClient.Vdc}
			err := notFoundData.deleteFunc(d, vcdClient)
			if err != nil {
				panic(fmt.Sprintf("could not delete '%s': %s", resourceAddress, err))
			}
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
