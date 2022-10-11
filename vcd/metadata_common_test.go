//go:build ALL || functional
// +build ALL functional

package vcd

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"strings"
	"testing"
)

// testMetadataEntry executes a test to check CRUD operations on "metadata_entry" attribute for the given HCL
// template and the given resource.
// The HCL template requires a {{.Name}} and {{.Metadata}} fields, and the usual {{.Org}} and {{.Vdc}}.
// You can add extra parameters as well to inject in the given HCL template, or override existent ones.
func testMetadataEntry(t *testing.T, hclTemplate string, resourceAddress string, extraParams StringMap) {
	preTestChecks(t)
	var params = StringMap{
		"Org":      testConfig.VCD.Org,
		"Vdc":      testConfig.Nsxt.Vdc,
		"Name":     t.Name(),
		"Metadata": getMetadataTestingHcl(),
	}

	for extraParam, extraParamValue := range extraParams {
		params[extraParam] = extraParamValue
	}
	testParamsNotEmpty(t, params)

	params["FuncName"] = t.Name() + "Create"
	createHcl := templateFill(hclTemplate, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", createHcl)

	params["FuncName"] = t.Name() + "Update"
	params["Metadata"] = getMetadataTestingHclForUpdate()
	updateHcl := templateFill(hclTemplate, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", updateHcl)

	params["FuncName"] = t.Name() + "Delete"
	params["Metadata"] = " "
	deleteHcl := templateFill(hclTemplate, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", deleteHcl)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: createHcl,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddress, "name", t.Name()),
					assertMetadata(resourceAddress),
				),
			},
			{
				Config: updateHcl,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddress, "name", t.Name()),
					assertUpdatedMetadata(resourceAddress),
				),
			},
			{
				Config: deleteHcl,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddress, "name", t.Name()),
					resource.TestCheckResourceAttr(resourceAddress, "metadata_entry.#", "0"),
				),
			},
		},
	})
	postTestChecks(t)
}

// getMetadataTestingHcl gets valid metadata entries to inject them into an HCL for testing
func getMetadataTestingHcl() string {
	return `
	  metadata_entry {
		key   = "defaultKey"
		value = "I'm a test for default values"
	  }
	  metadata_entry {
		key         = "stringKey"
		value       = "I'm a test for string values"
		type        = "MetadataStringValue"
		user_access = "READWRITE"
		is_system   = false
	  }
	  metadata_entry {
		key         = "numberKey"
		value       = "1234"
		type        = "MetadataNumberValue"
		user_access = "READWRITE"
		is_system   = false
	  }
	  metadata_entry {
		key         = "boolKey"
		value       = "true"
		type        = "MetadataBooleanValue"
		user_access = "READWRITE"
		is_system   = false
	  }
	  metadata_entry {
		key         = "dateKey"
		value       = "2022-10-05T13:44:00.000Z"
		type        = "MetadataDateTimeValue"
		user_access = "READWRITE"
		is_system   = false
	  }
	  metadata_entry {
		key         = "hiddenKey"
		value       = "I'm a test for hidden values"
		type        = "MetadataStringValue"
		user_access = "PRIVATE"
		is_system   = true
	  }
	  metadata_entry {
		key         = "readOnlyKey"
		value       = "I'm a test for read only values"
		type        = "MetadataStringValue"
		user_access = "READONLY"
		is_system   = true
	  }
`
}

func getMetadataTestingHclForUpdate() string {
	replacer := strings.NewReplacer(
		"I'm a test for", "I'm a test to update",
		"1234", "9999",
		"2022-10-05", "2022-10-06",
		"\"true\"", "\"false\"",
	)
	return replacer.Replace(getMetadataTestingHcl())
}

// assertMetadata checks that the state is updated after applying the HCL returned by getMetadataTestingHclForUpdate
func assertMetadata(resourceName string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttr(resourceName, "metadata_entry.#", "7"),
		// Tests default metadata values
		resource.TestCheckTypeSetElemNestedAttrs(resourceName, "metadata_entry.*",
			map[string]string{
				"key":         "defaultKey",
				"value":       "I'm a test for default values",
				"type":        types.MetadataStringValue,
				"user_access": types.MetadataReadWriteVisibility,
				"is_system":   "false",
			},
		),
		// Tests string metadata values
		resource.TestCheckTypeSetElemNestedAttrs(resourceName, "metadata_entry.*",
			map[string]string{
				"key":         "stringKey",
				"value":       "I'm a test for string values",
				"type":        types.MetadataStringValue,
				"user_access": types.MetadataReadWriteVisibility,
				"is_system":   "false",
			},
		),
		// Tests numeric metadata values
		resource.TestCheckTypeSetElemNestedAttrs(resourceName, "metadata_entry.*",
			map[string]string{
				"key":         "numberKey",
				"value":       "1234",
				"type":        types.MetadataNumberValue,
				"user_access": types.MetadataReadWriteVisibility,
				"is_system":   "false",
			},
		),
		// Tests bool metadata values
		resource.TestCheckTypeSetElemNestedAttrs(resourceName, "metadata_entry.*",
			map[string]string{
				"key":         "boolKey",
				"value":       "true",
				"type":        types.MetadataBooleanValue,
				"user_access": types.MetadataReadWriteVisibility,
				"is_system":   "false",
			},
		),
		// Tests date metadata values
		resource.TestCheckTypeSetElemNestedAttrs(resourceName, "metadata_entry.*",
			map[string]string{
				"key":         "dateKey",
				"value":       "2022-10-05T13:44:00.000Z",
				"type":        types.MetadataDateTimeValue,
				"user_access": types.MetadataReadWriteVisibility,
				"is_system":   "false",
			},
		),
		// Tests hidden metadata values in SYSTEM
		resource.TestCheckTypeSetElemNestedAttrs(resourceName, "metadata_entry.*",
			map[string]string{
				"key":         "hiddenKey",
				"value":       "I'm a test for hidden values",
				"type":        types.MetadataStringValue,
				"user_access": types.MetadataHiddenVisibility,
				"is_system":   "true",
			},
		),
		// Tests read only metadata values in SYSTEM
		resource.TestCheckTypeSetElemNestedAttrs(resourceName, "metadata_entry.*",
			map[string]string{
				"key":         "readOnlyKey",
				"value":       "I'm a test for read only values",
				"type":        types.MetadataStringValue,
				"user_access": types.MetadataReadOnlyVisibility,
				"is_system":   "true",
			},
		),
	)
}

// assertUpdatedMetadata checks that the state is updated after applying the HCL returned by getMetadataTestingHcl
func assertUpdatedMetadata(resourceName string) resource.TestCheckFunc {
	return resource.ComposeAggregateTestCheckFunc(
		resource.TestCheckResourceAttr(resourceName, "metadata_entry.#", "7"),
		// Tests default metadata values
		resource.TestCheckTypeSetElemNestedAttrs(resourceName, "metadata_entry.*",
			map[string]string{
				"key":         "defaultKey",
				"value":       "I'm a test to update default values",
				"type":        types.MetadataStringValue,
				"user_access": types.MetadataReadWriteVisibility,
				"is_system":   "false",
			},
		),
		// Tests string metadata values
		resource.TestCheckTypeSetElemNestedAttrs(resourceName, "metadata_entry.*",
			map[string]string{
				"key":         "stringKey",
				"value":       "I'm a test to update string values",
				"type":        types.MetadataStringValue,
				"user_access": types.MetadataReadWriteVisibility,
				"is_system":   "false",
			},
		),
		// Tests numeric metadata values
		resource.TestCheckTypeSetElemNestedAttrs(resourceName, "metadata_entry.*",
			map[string]string{
				"key":         "numberKey",
				"value":       "9999",
				"type":        types.MetadataNumberValue,
				"user_access": types.MetadataReadWriteVisibility,
				"is_system":   "false",
			},
		),
		// Tests bool metadata values
		resource.TestCheckTypeSetElemNestedAttrs(resourceName, "metadata_entry.*",
			map[string]string{
				"key":         "boolKey",
				"value":       "false",
				"type":        types.MetadataBooleanValue,
				"user_access": types.MetadataReadWriteVisibility,
				"is_system":   "false",
			},
		),
		// Tests date metadata values
		resource.TestCheckTypeSetElemNestedAttrs(resourceName, "metadata_entry.*",
			map[string]string{
				"key":         "dateKey",
				"value":       "2022-10-06T13:44:00.000Z",
				"type":        types.MetadataDateTimeValue,
				"user_access": types.MetadataReadWriteVisibility,
				"is_system":   "false",
			},
		),
		// Tests hidden metadata values in SYSTEM
		resource.TestCheckTypeSetElemNestedAttrs(resourceName, "metadata_entry.*",
			map[string]string{
				"key":         "hiddenKey",
				"value":       "I'm a test to update hidden values",
				"type":        types.MetadataStringValue,
				"user_access": types.MetadataHiddenVisibility,
				"is_system":   "true",
			},
		),
		// Tests read only metadata values in SYSTEM
		resource.TestCheckTypeSetElemNestedAttrs(resourceName, "metadata_entry.*",
			map[string]string{
				"key":         "readOnlyKey",
				"value":       "I'm a test to update read only values",
				"type":        types.MetadataStringValue,
				"user_access": types.MetadataReadOnlyVisibility,
				"is_system":   "true",
			},
		),
	)
}
