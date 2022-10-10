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
// The HCL template requires a {{.Name}} and {{.Metadata}} fields, as the usual {{.Org}} and {{.Vdc}}.
func testMetadataEntry(t *testing.T, hclTemplate string, resourceAddress string) {
	preTestChecks(t)
	var params = StringMap{
		"Org":      testConfig.VCD.Org,
		"Vdc":      testConfig.VCD.NsxtProviderVdc.Name,
		"Name":     t.Name(),
		"Metadata": getMetadataTestingHcl(),
	}
	testParamsNotEmpty(t, params)

	createHcl := templateFill(hclTemplate, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", createHcl)

	params["FuncName"] = t.Name() + "Update"
	params["Metadata"] = getMetadataTestingHclForUpdate()
	updateHcl := templateFill(hclTemplate, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", updateHcl)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: createHcl,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddress, "name", t.Name()),
					assertMetadata(resourceAddress),
				),
			},
			{
				Config: updateHcl,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddress, "name", t.Name()),
					assertUpdatedMetadata(resourceAddress),
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
		key         = "defaultKey"
		value       = "I'm a test for default values"
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
		value       = true
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
        user_access = "HIDDEN"
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
	return strings.ReplaceAll(
		strings.ReplaceAll(
			strings.ReplaceAll(getMetadataTestingHcl(),
				"I'm a test for", "I'm a test to update"),
			"1234", "9999"),
		"2022-10-05", "2022-10-06")
}

// assertMetadata checks that the state is updated after applying the HCL returned by getMetadataTestingHcl
func assertMetadata(resourceName string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttr(resourceName, "metadata_entry.#", "6"),

		// Tests default metadata values
		resource.TestCheckResourceAttr(resourceName, "metadata_entry.0.key", "defaultKey"),
		resource.TestCheckResourceAttr(resourceName, "metadata_entry.0.value", "I'm a test for default values"),
		resource.TestCheckResourceAttr(resourceName, "metadata_entry.0.type", types.MetadataStringValue),
		resource.TestCheckResourceAttr(resourceName, "metadata_entry.0.user_access", types.MetadataReadWriteVisibility),
		resource.TestCheckResourceAttr(resourceName, "metadata_entry.0.is_system", "false"),
		// Tests string metadata values
		resource.TestCheckResourceAttr(resourceName, "metadata_entry.1.key", "stringKey"),
		resource.TestCheckResourceAttr(resourceName, "metadata_entry.1.value", "I'm a test for string values"),
		resource.TestCheckResourceAttr(resourceName, "metadata_entry.1.type", types.MetadataStringValue),
		resource.TestCheckResourceAttr(resourceName, "metadata_entry.1.user_access", types.MetadataReadWriteVisibility),
		resource.TestCheckResourceAttr(resourceName, "metadata_entry.1.is_system", "false"),
		// Tests numeric metadata values
		resource.TestCheckResourceAttr(resourceName, "metadata_entry.2.key", "numberKey"),
		resource.TestCheckResourceAttr(resourceName, "metadata_entry.2.value", "1234"),
		resource.TestCheckResourceAttr(resourceName, "metadata_entry.2.type", types.MetadataNumberValue),
		resource.TestCheckResourceAttr(resourceName, "metadata_entry.2.user_access", types.MetadataReadWriteVisibility),
		resource.TestCheckResourceAttr(resourceName, "metadata_entry.2.is_system", "false"),
		// Tests date metadata values
		resource.TestCheckResourceAttr(resourceName, "metadata_entry.3.key", "dateKey"),
		resource.TestCheckResourceAttr(resourceName, "metadata_entry.3.value", "2022-10-05T13:44:00.000Z"),
		resource.TestCheckResourceAttr(resourceName, "metadata_entry.3.type", types.MetadataDateTimeValue),
		resource.TestCheckResourceAttr(resourceName, "metadata_entry.3.user_access", types.MetadataReadWriteVisibility),
		resource.TestCheckResourceAttr(resourceName, "metadata_entry.3.is_system", "false"),
		// Tests hidden metadata values in SYSTEM
		resource.TestCheckResourceAttr(resourceName, "metadata_entry.4.key", "hiddenKey"),
		resource.TestCheckResourceAttr(resourceName, "metadata_entry.4.value", "I'm a test for hidden values"),
		resource.TestCheckResourceAttr(resourceName, "metadata_entry.4.type", types.MetadataStringValue),
		resource.TestCheckResourceAttr(resourceName, "metadata_entry.4.user_access", types.MetadataHiddenVisibility),
		resource.TestCheckResourceAttr(resourceName, "metadata_entry.4.is_system", "true"),
		// Tests read only metadata values in SYSTEM
		resource.TestCheckResourceAttr(resourceName, "metadata_entry.5.key", "readOnlyKey"),
		resource.TestCheckResourceAttr(resourceName, "metadata_entry.5.value", "I'm a test for read only values"),
		resource.TestCheckResourceAttr(resourceName, "metadata_entry.5.type", types.MetadataStringValue),
		resource.TestCheckResourceAttr(resourceName, "metadata_entry.5.user_access", types.MetadataReadOnlyVisibility),
		resource.TestCheckResourceAttr(resourceName, "metadata_entry.5.is_system", "true"),
	)
}

// assertUpdatedMetadata checks that the state is updated after applying the HCL returned by getMetadataTestingHclForUpdate
func assertUpdatedMetadata(resourceName string) resource.TestCheckFunc {
	return resource.ComposeTestCheckFunc(
		resource.TestCheckResourceAttr(resourceName, "metadata_entry.#", "6"),

		// Tests default metadata values
		resource.TestCheckResourceAttr(resourceName, "metadata_entry.0.key", "defaultKey"),
		resource.TestCheckResourceAttr(resourceName, "metadata_entry.0.value", "I'm a test to update default values"),
		resource.TestCheckResourceAttr(resourceName, "metadata_entry.0.type", types.MetadataStringValue),
		resource.TestCheckResourceAttr(resourceName, "metadata_entry.0.user_access", types.MetadataReadWriteVisibility),
		resource.TestCheckResourceAttr(resourceName, "metadata_entry.0.is_system", "false"),

		// Tests string metadata values
		resource.TestCheckResourceAttr(resourceName, "metadata_entry.1.key", "stringKey"),
		resource.TestCheckResourceAttr(resourceName, "metadata_entry.1.value", "I'm a test to update string values"),
		resource.TestCheckResourceAttr(resourceName, "metadata_entry.1.type", types.MetadataStringValue),
		resource.TestCheckResourceAttr(resourceName, "metadata_entry.1.user_access", types.MetadataReadWriteVisibility),
		resource.TestCheckResourceAttr(resourceName, "metadata_entry.1.is_system", "false"),

		// Tests numeric metadata values
		resource.TestCheckResourceAttr(resourceName, "metadata_entry.2.key", "numberKey"),
		resource.TestCheckResourceAttr(resourceName, "metadata_entry.2.value", "9999"),
		resource.TestCheckResourceAttr(resourceName, "metadata_entry.2.type", types.MetadataNumberValue),
		resource.TestCheckResourceAttr(resourceName, "metadata_entry.2.user_access", types.MetadataReadWriteVisibility),
		resource.TestCheckResourceAttr(resourceName, "metadata_entry.2.is_system", "false"),

		// Tests date metadata values
		resource.TestCheckResourceAttr(resourceName, "metadata_entry.3.key", "dateKey"),
		resource.TestCheckResourceAttr(resourceName, "metadata_entry.3.value", "2022-10-06T13:44:00.000Z"),
		resource.TestCheckResourceAttr(resourceName, "metadata_entry.3.type", types.MetadataDateTimeValue),
		resource.TestCheckResourceAttr(resourceName, "metadata_entry.3.user_access", types.MetadataReadWriteVisibility),
		resource.TestCheckResourceAttr(resourceName, "metadata_entry.3.is_system", "false"),

		// Tests hidden metadata values in SYSTEM
		resource.TestCheckResourceAttr(resourceName, "metadata_entry.4.key", "hiddenKey"),
		resource.TestCheckResourceAttr(resourceName, "metadata_entry.4.value", "I'm a test to update hidden values"),
		resource.TestCheckResourceAttr(resourceName, "metadata_entry.4.type", types.MetadataStringValue),
		resource.TestCheckResourceAttr(resourceName, "metadata_entry.4.user_access", types.MetadataHiddenVisibility),
		resource.TestCheckResourceAttr(resourceName, "metadata_entry.4.is_system", "true"),

		// Tests read only metadata values in SYSTEM
		resource.TestCheckResourceAttr(resourceName, "metadata_entry.5.key", "readOnlyKey"),
		resource.TestCheckResourceAttr(resourceName, "metadata_entry.5.value", "I'm a test to update read only values"),
		resource.TestCheckResourceAttr(resourceName, "metadata_entry.5.type", types.MetadataStringValue),
		resource.TestCheckResourceAttr(resourceName, "metadata_entry.5.user_access", types.MetadataReadOnlyVisibility),
		resource.TestCheckResourceAttr(resourceName, "metadata_entry.5.is_system", "true"),
	)
}
