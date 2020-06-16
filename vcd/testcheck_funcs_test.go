// +build vapp vm user ALL functional

package vcd

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
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
