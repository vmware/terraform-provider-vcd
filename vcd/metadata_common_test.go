//go:build catalog || disk || network || nsxt || vdc || org || vapp || vm || functional || ALL
// +build catalog disk network nsxt vdc org vapp vm functional ALL

package vcd

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

func stateDumper() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		spew.Dump(s)
		return nil
	}
}

// testMetadataEntryCRUD executes a test that asserts CRUD operation behaviours of "metadata_entry" attribute in the given HCL
// templates, that must correspond to a resource and a data source referencing this resource.
// The HCL template requires {{.Name}} and {{.Metadata}} fields, and the usual {{.Org}} and {{.Vdc}}.
// You can add extra parameters as well to inject in the given HCL template, or override these mentioned ones.
// The data source HCL is always concatenated to the resource after creation, and it's skipped on binary tests.
//
// Test scenario:
// - Create 7 metadata entries, 2 for string and 1 for number, bool, date, readOnly and hidden
// - Delete 1 of string value (so it remains one of each type)
// - Add a data source
// - Update 2 entries
// - Delete all of them except the string one
// - Delete all of them
// - Check a malformed metadata entry
func testMetadataEntryCRUD(t *testing.T, resourceTemplate, resourceAddress, datasourceTemplate, datasourceAddress string, extraParams StringMap) {
	preTestChecks(t)
	metadataHcl := getMetadataTestingHcl(2, 1, 1, 1, 1, 1)
	var params = StringMap{
		"Org":      testConfig.VCD.Org,
		"Vdc":      testConfig.Nsxt.Vdc,
		"Name":     t.Name(),
		"Metadata": metadataHcl,
	}

	for extraParam, extraParamValue := range extraParams {
		params[extraParam] = extraParamValue
	}
	testParamsNotEmpty(t, params)

	params["FuncName"] = t.Name() + "Create"
	createHcl := templateFill(resourceTemplate, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", createHcl)

	params["FuncName"] = t.Name() + "DeleteSomeKeys"
	metadataHcl = getMetadataTestingHcl(1, 1, 1, 1, 1, 1)
	params["Metadata"] = metadataHcl
	deleteSomeKeysHcl := templateFill(resourceTemplate, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", createHcl)

	params["FuncName"] = t.Name() + "Datasource"
	withDatasourceHcl := templateFill(datasourceTemplate+"\n# skip-binary-test\n"+resourceTemplate, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", withDatasourceHcl)

	params["FuncName"] = t.Name() + "Update"
	params["Metadata"] = strings.NewReplacer("stringValue", "stringValueUpdated", "2022-10-", "2021-10-").Replace(metadataHcl)
	updateHcl := templateFill(resourceTemplate, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", updateHcl)

	params["FuncName"] = t.Name() + "Update2"
	metadataHcl = getMetadataTestingHcl(1, 0, 0, 0, 0, 0)
	params["Metadata"] = metadataHcl
	update2Hcl := templateFill(resourceTemplate, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", updateHcl)

	params["FuncName"] = t.Name() + "Delete"
	params["Metadata"] = "metadata_entry {}"
	deleteHcl := templateFill(resourceTemplate, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", deleteHcl)

	params["FuncName"] = t.Name() + "Wrong1"
	params["Metadata"] = "metadata_entry {\n\tkey = \"foo\"\n}"
	wrongHcl := templateFill("# skip-binary-test\n"+resourceTemplate, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", deleteHcl)

	// These are for the deprecated `metadata` value, to minimize possible regressions
	params["FuncName"] = t.Name() + "DeprecatedCreate"
	params["Metadata"] = "metadata = {\n\tfoo = \"bar\"\n}"
	deprecatedCreateHcl := templateFill(resourceTemplate, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", createHcl)

	params["FuncName"] = t.Name() + "DeprecatedUpdate"
	params["Metadata"] = "metadata = {\n\tfoo = \"bar2\"\n}"
	deprecatedUpdateHcl := templateFill(resourceTemplate, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", createHcl)

	params["FuncName"] = t.Name() + "DeprecatedDelete"
	params["Metadata"] = "metadata = {}"
	deprecatedDeleteHcl := templateFill(resourceTemplate, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", createHcl)

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
					resource.TestCheckResourceAttr(resourceAddress, "metadata_entry.#", "7"),

					testCheckMetadataEntrySetElemNestedAttrs(resourceAddress, "stringKey1", "stringValue1", types.MetadataStringValue, types.MetadataReadWriteVisibility, "false"),
					testCheckMetadataEntrySetElemNestedAttrs(resourceAddress, "stringKey2", "stringValue2", types.MetadataStringValue, types.MetadataReadWriteVisibility, "false"),
					testCheckMetadataEntrySetElemNestedAttrs(resourceAddress, "numberKey1", "1", types.MetadataNumberValue, types.MetadataReadWriteVisibility, "false"),
					testCheckMetadataEntrySetElemNestedAttrs(resourceAddress, "boolKey1", "false", types.MetadataBooleanValue, types.MetadataReadWriteVisibility, "false"),
					testCheckMetadataEntrySetElemNestedAttrs(resourceAddress, "dateKey1", "2022-10-01T12:00:00.000Z", types.MetadataDateTimeValue, types.MetadataReadWriteVisibility, "false"),
					testCheckMetadataEntrySetElemNestedAttrs(resourceAddress, "readOnlyKey1", "readOnlyValue1", types.MetadataStringValue, types.MetadataReadOnlyVisibility, "true"),
					testCheckMetadataEntrySetElemNestedAttrs(resourceAddress, "privateKey1", "privateValue1", types.MetadataStringValue, types.MetadataHiddenVisibility, "true"),
				),
			},
			{
				Config: deleteSomeKeysHcl,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddress, "name", t.Name()),
					resource.TestCheckResourceAttr(resourceAddress, "metadata_entry.#", "6"),
					testCheckMetadataEntrySetElemNestedAttrs(resourceAddress, "stringKey1", "stringValue1", types.MetadataStringValue, types.MetadataReadWriteVisibility, "false"),
					testCheckMetadataEntrySetElemNestedAttrs(resourceAddress, "numberKey1", "1", types.MetadataNumberValue, types.MetadataReadWriteVisibility, "false"),
					testCheckMetadataEntrySetElemNestedAttrs(resourceAddress, "boolKey1", "false", types.MetadataBooleanValue, types.MetadataReadWriteVisibility, "false"),
					testCheckMetadataEntrySetElemNestedAttrs(resourceAddress, "dateKey1", "2022-10-01T12:00:00.000Z", types.MetadataDateTimeValue, types.MetadataReadWriteVisibility, "false"),
					testCheckMetadataEntrySetElemNestedAttrs(resourceAddress, "readOnlyKey1", "readOnlyValue1", types.MetadataStringValue, types.MetadataReadOnlyVisibility, "true"),
					testCheckMetadataEntrySetElemNestedAttrs(resourceAddress, "privateKey1", "privateValue1", types.MetadataStringValue, types.MetadataHiddenVisibility, "true"),
				),
			},
			{
				Config: withDatasourceHcl,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceAddress, "id", resourceAddress, "id"),
					resource.TestCheckResourceAttr(resourceAddress, "metadata_entry.#", "6"),
					resource.TestCheckResourceAttr(datasourceAddress, "metadata_entry.#", "6"),
					testCheckMetadataEntrySetElemNestedAttrs(datasourceAddress, "stringKey1", "stringValue1", types.MetadataStringValue, types.MetadataReadWriteVisibility, "false"),
					testCheckMetadataEntrySetElemNestedAttrs(datasourceAddress, "numberKey1", "1", types.MetadataNumberValue, types.MetadataReadWriteVisibility, "false"),
					testCheckMetadataEntrySetElemNestedAttrs(datasourceAddress, "boolKey1", "false", types.MetadataBooleanValue, types.MetadataReadWriteVisibility, "false"),
					testCheckMetadataEntrySetElemNestedAttrs(datasourceAddress, "dateKey1", "2022-10-01T12:00:00.000Z", types.MetadataDateTimeValue, types.MetadataReadWriteVisibility, "false"),
					testCheckMetadataEntrySetElemNestedAttrs(datasourceAddress, "readOnlyKey1", "readOnlyValue1", types.MetadataStringValue, types.MetadataReadOnlyVisibility, "true"),
					testCheckMetadataEntrySetElemNestedAttrs(datasourceAddress, "privateKey1", "privateValue1", types.MetadataStringValue, types.MetadataHiddenVisibility, "true"),
				),
			},
			{
				Config: updateHcl,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddress, "name", t.Name()),
					resource.TestCheckResourceAttr(resourceAddress, "metadata_entry.#", "6"),
					// Updated values:
					testCheckMetadataEntrySetElemNestedAttrs(resourceAddress, "stringKey1", "stringValueUpdated1", types.MetadataStringValue, types.MetadataReadWriteVisibility, "false"),
					testCheckMetadataEntrySetElemNestedAttrs(resourceAddress, "dateKey1", "2021-10-01T12:00:00.000Z", types.MetadataDateTimeValue, types.MetadataReadWriteVisibility, "false"),
					// Not updated values:
					testCheckMetadataEntrySetElemNestedAttrs(resourceAddress, "numberKey1", "1", types.MetadataNumberValue, types.MetadataReadWriteVisibility, "false"),
					testCheckMetadataEntrySetElemNestedAttrs(resourceAddress, "boolKey1", "false", types.MetadataBooleanValue, types.MetadataReadWriteVisibility, "false"),
					testCheckMetadataEntrySetElemNestedAttrs(resourceAddress, "readOnlyKey1", "readOnlyValue1", types.MetadataStringValue, types.MetadataReadOnlyVisibility, "true"),
					testCheckMetadataEntrySetElemNestedAttrs(resourceAddress, "privateKey1", "privateValue1", types.MetadataStringValue, types.MetadataHiddenVisibility, "true"),
				),
			},
			{
				Config: update2Hcl,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddress, "name", t.Name()),
					resource.TestCheckResourceAttr(resourceAddress, "metadata_entry.#", "1"),
					testCheckMetadataEntrySetElemNestedAttrs(resourceAddress, "stringKey1", "stringValue1", types.MetadataStringValue, types.MetadataReadWriteVisibility, "false"),
				),
			},
			{
				Config: deleteHcl,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddress, "name", t.Name()),
					// This is a side effect of having `metadata_entry` as Computed to be able to delete metadata.
					resource.TestCheckResourceAttr(resourceAddress, "metadata_entry.#", "1"),
					testCheckMetadataEntrySetElemNestedAttrs(resourceAddress, "", "", "", "", "false"),
				),
			},
			{
				Config: deprecatedCreateHcl,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddress, "name", t.Name()),
					resource.TestCheckResourceAttr(resourceAddress, "metadata.foo", "bar"),
				),
			},
			{
				Config: deprecatedUpdateHcl,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddress, "name", t.Name()),
					resource.TestCheckResourceAttr(resourceAddress, "metadata.foo", "bar2"),
				),
			},
			{
				Config: deprecatedDeleteHcl,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddress, "name", t.Name()),
					stateDumper(),
					resource.TestCheckResourceAttr(resourceAddress, "metadata.%", "0"),
				),
			},
			{
				Config:      wrongHcl,
				ExpectError: regexp.MustCompile(".*all fields in a metadata_entry are required, but got some empty.*"),
			},
		},
	})
	postTestChecks(t)
}

// getMetadataTestingHcl gets valid metadata entries to inject them into an HCL for testing
func getMetadataTestingHcl(stringEntries, numberEntries, boolEntries, dateEntries, readOnlyEntries, privateEntries int) string {
	hcl := ""
	for i := 1; i <= stringEntries; i++ {
		hcl += getMetadataEntryHcl(fmt.Sprintf("stringKey%d", i), fmt.Sprintf("stringValue%d", i), types.MetadataStringValue, types.MetadataReadWriteVisibility, "false")
	}
	for i := 1; i <= numberEntries; i++ {
		hcl += getMetadataEntryHcl(fmt.Sprintf("numberKey%d", i), fmt.Sprintf("%d", i), types.MetadataNumberValue, types.MetadataReadWriteVisibility, "false")
	}
	for i := 1; i <= boolEntries; i++ {
		hcl += getMetadataEntryHcl(fmt.Sprintf("boolKey%d", i), strconv.FormatBool(i%2 == 0), types.MetadataBooleanValue, types.MetadataReadWriteVisibility, "false")
	}
	for i := 1; i <= dateEntries; i++ {
		hcl += getMetadataEntryHcl(fmt.Sprintf("dateKey%d", i), fmt.Sprintf("2022-10-%02dT12:00:00.000Z", i%30), types.MetadataDateTimeValue, types.MetadataReadWriteVisibility, "false")
	}
	for i := 1; i <= readOnlyEntries; i++ {
		hcl += getMetadataEntryHcl(fmt.Sprintf("readOnlyKey%d", i), fmt.Sprintf("readOnlyValue%d", i), types.MetadataStringValue, types.MetadataReadOnlyVisibility, "true")
	}
	for i := 1; i <= privateEntries; i++ {
		hcl += getMetadataEntryHcl(fmt.Sprintf("privateKey%d", i), fmt.Sprintf("privateValue%d", i), types.MetadataStringValue, types.MetadataHiddenVisibility, "true")
	}
	return hcl

}

func getMetadataEntryHcl(key, value, typedValue, userAccess, isSystem string) string {
	return `
		  metadata_entry {
			key         = "` + key + `"
			value       = "` + value + `"
			type        = "` + typedValue + `"
			user_access = "` + userAccess + `"
			is_system   = "` + isSystem + `"
		  }`
}

// testCheckMetadataEntrySetElemNestedAttrs asserts that a given metadata_entry has the expected input for the given resourceAddress.
func testCheckMetadataEntrySetElemNestedAttrs(resourceAddress, expectedKey, expectedValue, expectedType, expectedUserAccess, expectedIsSystem string) resource.TestCheckFunc {
	return resource.TestCheckTypeSetElemNestedAttrs(resourceAddress, "metadata_entry.*",
		map[string]string{
			"key":         expectedKey,
			"value":       expectedValue,
			"type":        expectedType,
			"user_access": expectedUserAccess,
			"is_system":   expectedIsSystem,
		},
	)
}
