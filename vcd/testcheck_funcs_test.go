//go:build api || vapp || vm || user || nsxt || extnetwork || network || gateway || catalog || standaloneVm || alb || vdcGroup || ldap || vdc || access_control || rde || uiPlugin || org || disk || providerVdc || cse || ALL || slz || tm || functional

package vcd

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// testCachedFieldValue structure with attached functions is useful for testing specific field value
// across different `resource.TestStep` in Terraform acceptance tests. One particular use case is
// to check whether MAC address does not change when a `vcd_vapp_vm` resource's network stack is
// updated (between different TestSteps).
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
		if vcdTestVerbose {
			fmt.Printf("# stored field '%s' with value '%s'\n", field, c.fieldValue)
		}
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

		if vcdTestVerbose {
			fmt.Printf("# Comparing field '%s' '%s==%s' in resource '%s'\n", field, value, c.fieldValue, resource)
		}

		if value != c.fieldValue {
			return fmt.Errorf("got '%s - %s' field value %s, expected: %s",
				resource, field, value, c.fieldValue)
		}

		return nil
	}
}

// testCheckCachedResourceFieldValueChanged is the reverse of testCheckCachedResourceFieldValue and
// it will fail if the values remain the same
// It is useful for checking if the resource was recreated (id field values should differ)
func (c *testCachedFieldValue) testCheckCachedResourceFieldValueChanged(resource, field string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("resource not found: %s", resource)
		}

		value, exists := rs.Primary.Attributes[field]
		if !exists {
			return fmt.Errorf("field %s in resource %s does not exist", field, resource)
		}

		if vcdTestVerbose {
			fmt.Printf("# Comparing field '%s' '%s!=%s' in resource '%s'\n", field, value, c.fieldValue, resource)
		}

		if value == c.fieldValue {
			return fmt.Errorf("got '%s - %s' field value %s, expected a different one",
				resource, field, value)
		}

		return nil
	}
}

// String satisfies stringer interface (supports fmt.Printf...)
func (c *testCachedFieldValue) String() string {
	return c.fieldValue
}

// testCheckMatchOutput allows to match output field with regexp
func testCheckMatchOutput(name string, r *regexp.Regexp) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ms := s.RootModule()
		rs, ok := ms.Outputs[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if !r.MatchString(rs.Value.(string)) {
			return fmt.Errorf(
				"Output '%s': expected %#v, got %#v", name, rs.Value, rs)
		}

		return nil
	}
}

// testCheckOutputNonEmpty checks that output field is not empty
func testCheckOutputNonEmpty(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ms := s.RootModule()
		rs, ok := ms.Outputs[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Value.(string) == "" {
			return fmt.Errorf("Output '%s': expected '', got %#v", name, rs)
		}

		return nil
	}
}

// resourceFieldsEqual checks if secondObject has all the fields and their values set as the
// firstObject except `[]excludeFields`. This is very useful to check if data sources have all
// the same values as resources
func resourceFieldsEqual(firstObject, secondObject string, excludeFields []string) resource.TestCheckFunc {
	return resourceFieldsEqualCustom(firstObject, secondObject, excludeFields, stringInSlice)
}
func resourceFieldsEqualCustom(firstObject, secondObject string, excludeFields []string, exclusionChecker func(str string, list []string) bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		resource1, ok := s.RootModule().Resources[firstObject]
		if !ok {
			return fmt.Errorf("unable to find %s", firstObject)
		}

		resource2, ok := s.RootModule().Resources[secondObject]
		if !ok {
			return fmt.Errorf("unable to find %s", secondObject)
		}

		for fieldName := range resource1.Primary.Attributes {
			// Do not validate the fields marked for exclusion
			if excludeFields != nil && exclusionChecker(fieldName, excludeFields) {
				continue
			}

			if vcdTestVerbose {
				fmt.Printf("field %s %s (value %s) and %s (value %s))\n", fieldName, firstObject,
					resource1.Primary.Attributes[fieldName], secondObject, resource2.Primary.Attributes[fieldName])
			}
			if !reflect.DeepEqual(resource1.Primary.Attributes[fieldName], resource2.Primary.Attributes[fieldName]) {
				return fmt.Errorf("field %s differs in resources %s (value %s) and %s (value %s)",
					fieldName, firstObject, resource1.Primary.Attributes[fieldName], secondObject, resource2.Primary.Attributes[fieldName])
			}
		}
		return nil
	}
}

func resourceFieldIntNotEqual(object, field string, notEqualTo int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		resource, ok := s.RootModule().Resources[object]
		if !ok {
			return fmt.Errorf("unable to find %s", object)
		}

		fieldValue, fieldSet := resource.Primary.Attributes[field]
		if !fieldSet {
			return fmt.Errorf("expected field '%s' to be set in resource '%s'", field, object)
		}

		intFieldValue, err := strconv.Atoi(fieldValue)
		if err != nil {
			return fmt.Errorf("could not convert field '%s' to int: %s", fieldValue, err)
		}

		if intFieldValue == notEqualTo {
			return fmt.Errorf("expected field '%s' value to be '%d'!='%d'", object, intFieldValue, notEqualTo)
		}

		return nil
	}
}

// testMatchResourceAttrWhenVersionMatches is equivalent to resource.TestMatchResourceAttr but only runs this function
// if the VCD version matches the one from the input. Otherwise, this will be a no-op.
func testMatchResourceAttrWhenVersionMatches(name, key string, r *regexp.Regexp, versionConstraint string) resource.TestCheckFunc {
	return testCheckResourceGenericWhenVersionMatches(resource.TestMatchResourceAttr(name, key, r), versionConstraint)
}

// testCheckResourceAttrWhenVersionMatches is equivalent to resource.TestCheckResourceAttr but only runs this function
// if the VCD version matches the one from the input. Otherwise, this will be a no-op.
func testCheckResourceAttrWhenVersionMatches(name, key, value string, versionConstraint string) resource.TestCheckFunc {
	return testCheckResourceGenericWhenVersionMatches(resource.TestCheckResourceAttr(name, key, value), versionConstraint)
}

// testCheckResourceAttrSetWhenVersionMatches is equivalent to resource.TestCheckResourceAttrSet but only runs this function
// if the VCD version matches the one from the input. Otherwise, this will be a no-op.
func testCheckResourceAttrSetWhenVersionMatches(name, key, versionConstraint string) resource.TestCheckFunc {
	return testCheckResourceGenericWhenVersionMatches(resource.TestCheckResourceAttrSet(name, key), versionConstraint)
}

// testCheckResourceGenericWhenVersionMatches runs the given test checking function if the VCD version matches
// the one from the input. Otherwise, this will be a no-op.
func testCheckResourceGenericWhenVersionMatches(checker resource.TestCheckFunc, versionConstraint string) resource.TestCheckFunc {
	if !checkVersion(testConfig.Provider.ApiVersion, versionConstraint) {
		debugPrintf("This test check requires VCD version to be '%s'. Skipping", versionConstraint)
		// Returns a dummy checker that does nothing
		return func(state *terraform.State) error {
			return nil
		}
	}
	return checker
}
