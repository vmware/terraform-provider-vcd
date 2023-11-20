//go:build rde || functional || ALL

package vcd

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"strconv"
	"strings"
	"testing"
)

// testOpenApiMetadataEntryCRUD executes a test that asserts CRUD operation behaviours of "metadata_entry" attribute in the given HCL
// templates, that must correspond to a resource and a data source referencing this resource.
// The HCL template requires {{.Name}} and {{.Metadata}} fields, and the usual {{.Org}} and {{.Vdc}}.
// You can add extra parameters as well to inject in the given HCL template, or override these mentioned ones.
// The data source HCL is always concatenated to the resource after creation, and it's skipped on binary tests.
//
// Tests:
// - Step 1:  Create the resource with no metadata
// - Step 2:  Taint and re-create with 4 metadata entries, 1 for string, number, bool, date with GENERAL domain (is_system = false)
// - Step 3:  Add a data source
// - Step 4:  Delete 1 metadata entry, the bool one
// - Step 5:  Update the string and date metadata values
// - Step 6:  Delete all of them
// - Step 7:  (Sysadmin only) Create 2 entries with is_system=true (readonly and private user_access)
// - Step 8:  (Sysadmin only) Update the hidden one
// - Step 9:  (Sysadmin only) Delete all of them
// - Step 10:  Check a malformed metadata entry
// - Step 11: (Org user only) Check that specifying an is_system metadata entry with a tenant user gives an error
// - Step 12+: Some extra tests for deprecated `metadata` attribute
func testOpenApiMetadataEntryCRUD(t *testing.T, resourceTemplate, resourceAddress, datasourceTemplate, datasourceAddress string, extraParams StringMap) {
	preTestChecks(t)
	var params = StringMap{
		"Org":  testConfig.VCD.Org,
		"Vdc":  testConfig.Nsxt.Vdc,
		"Name": t.Name(),
	}

	for extraParam, extraParamValue := range extraParams {
		params[extraParam] = extraParamValue
	}
	testParamsNotEmpty(t, params)

	params["FuncName"] = t.Name() + "NoMetadata"
	params["Metadata"] = " "
	noMetadataHcl := templateFill(resourceTemplate, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", noMetadataHcl)

	params["FuncName"] = t.Name() + "Create"
	params["Metadata"] = getOpenApiMetadataTestingHcl(1, 1, 1, 1, 1, 1)
	createHcl := templateFill(resourceTemplate, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", createHcl)

	withDatasourceHcl := ""
	if datasourceTemplate != "" {
		params["FuncName"] = t.Name() + "WithDatasource"
		withDatasourceHcl = templateFill(datasourceTemplate+"\n# skip-binary-test\n"+resourceTemplate, params)
		debugPrintf("#[DEBUG] CONFIGURATION: %s", withDatasourceHcl)
	}

	params["FuncName"] = t.Name() + "DeleteOneKey"
	params["Metadata"] = getOpenApiMetadataTestingHcl(1, 1, 0, 1, 1, 1)
	deleteOneKeyHcl := templateFill(resourceTemplate, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", deleteOneKeyHcl)

	params["FuncName"] = t.Name() + "Update"
	params["Metadata"] = strings.NewReplacer("stringValue", "stringValueUpdated").Replace(params["Metadata"].(string))
	updateHcl := templateFill(resourceTemplate, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", updateHcl)

	params["FuncName"] = t.Name() + "Delete"
	params["Metadata"] = " "
	deleteHcl := templateFill(resourceTemplate, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", deleteHcl)

	params["FuncName"] = t.Name() + "MetadataEntryWithDefaults"
	params["Metadata"] = "metadata_entry {\n\tkey = \"defaultKey\"\nvalue = \"defaultValue\"\n}"
	withDefaults := templateFill(resourceTemplate, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", withDefaults)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: noMetadataHcl,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddress, "name", t.Name()),
					resource.TestCheckResourceAttr(resourceAddress, "metadata_entry.#", "0"),
				),
			},
			{
				Config: createHcl,
				Taint:  []string{resourceAddress}, // Forces re-creation to test Create with metadata.
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddress, "name", t.Name()),
					resource.TestCheckResourceAttr(resourceAddress, "metadata_entry.#", "6"),
					testCheckOpenApiMetadataEntrySetElemNestedAttrs(resourceAddress, "stringKey1", "stringValue1", types.OpenApiMetadataStringEntry, "TENANT", "", "false"),
					testCheckOpenApiMetadataEntrySetElemNestedAttrs(resourceAddress, "numberKey1", "1", types.OpenApiMetadataNumberEntry, "TENANT", "", "false"),
					testCheckOpenApiMetadataEntrySetElemNestedAttrs(resourceAddress, "boolKey1", "false", types.OpenApiMetadataStringEntry, "TENANT", "", "false"),
					testCheckOpenApiMetadataEntrySetElemNestedAttrs(resourceAddress, "readOnly1", "readOnly1", types.OpenApiMetadataStringEntry, "TENANT", "", "true"),
					testCheckOpenApiMetadataEntrySetElemNestedAttrs(resourceAddress, "namespaced1", "namespaced1", types.OpenApiMetadataStringEntry, "TENANT", "namespace", "false"),
					testCheckOpenApiMetadataEntrySetElemNestedAttrs(resourceAddress, "provider1", "provider1", types.OpenApiMetadataStringEntry, "PROVIDER", "", "false"),
				),
			},
			{
				SkipFunc: func() (bool, error) {
					return withDatasourceHcl == "", nil
				},
				Config: withDatasourceHcl,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceAddress, "id", resourceAddress, "id"),
					resource.TestCheckResourceAttr(resourceAddress, "metadata_entry.#", "6"),
					resource.TestCheckResourceAttr(datasourceAddress, "metadata_entry.#", "6"),
					testCheckOpenApiMetadataEntrySetElemNestedAttrs(resourceAddress, "stringKey1", "stringValue1", types.OpenApiMetadataStringEntry, "TENANT", "", "false"),
					testCheckOpenApiMetadataEntrySetElemNestedAttrs(resourceAddress, "numberKey1", "1", types.OpenApiMetadataNumberEntry, "TENANT", "", "false"),
					testCheckOpenApiMetadataEntrySetElemNestedAttrs(resourceAddress, "boolKey1", "false", types.OpenApiMetadataBooleanEntry, "TENANT", "", "false"),
					testCheckOpenApiMetadataEntrySetElemNestedAttrs(resourceAddress, "readOnly1", "readOnly1", types.OpenApiMetadataStringEntry, "TENANT", "", "true"),
					testCheckOpenApiMetadataEntrySetElemNestedAttrs(resourceAddress, "namespaced1", "namespaced1", types.OpenApiMetadataStringEntry, "TENANT", "namespace", "false"),
					testCheckOpenApiMetadataEntrySetElemNestedAttrs(resourceAddress, "provider1", "provider1", types.OpenApiMetadataStringEntry, "PROVIDER", "", "false"),
				),
			},
			{
				Config: deleteOneKeyHcl,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddress, "name", t.Name()),
					resource.TestCheckResourceAttr(resourceAddress, "metadata_entry.#", "5"),
					// The bool is deleted
					testCheckOpenApiMetadataEntrySetElemNestedAttrs(resourceAddress, "stringKey1", "stringValue1", types.OpenApiMetadataStringEntry, "TENANT", "", "false"),
					testCheckOpenApiMetadataEntrySetElemNestedAttrs(resourceAddress, "numberKey1", "1", types.OpenApiMetadataNumberEntry, "TENANT", "", "false"),
					testCheckOpenApiMetadataEntrySetElemNestedAttrs(resourceAddress, "readOnly1", "readOnly1", types.OpenApiMetadataStringEntry, "TENANT", "", "true"),
					testCheckOpenApiMetadataEntrySetElemNestedAttrs(resourceAddress, "namespaced1", "namespaced1", types.OpenApiMetadataStringEntry, "TENANT", "namespace", "false"),
					testCheckOpenApiMetadataEntrySetElemNestedAttrs(resourceAddress, "provider1", "provider1", types.OpenApiMetadataStringEntry, "PROVIDER", "", "false"),
				),
			},
			{
				Config: updateHcl,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddress, "name", t.Name()),
					resource.TestCheckResourceAttr(resourceAddress, "metadata_entry.#", "5"),
					// Updated value:
					testCheckOpenApiMetadataEntrySetElemNestedAttrs(resourceAddress, "stringKey1", "stringValueUpdated1", types.OpenApiMetadataStringEntry, "TENANT", "", "false"),
					// Not updated values:
					testCheckOpenApiMetadataEntrySetElemNestedAttrs(resourceAddress, "numberKey1", "1", types.OpenApiMetadataNumberEntry, "TENANT", "", "false"),
					testCheckOpenApiMetadataEntrySetElemNestedAttrs(resourceAddress, "readOnly1", "readOnly1", types.OpenApiMetadataStringEntry, "TENANT", "", "true"),
					testCheckOpenApiMetadataEntrySetElemNestedAttrs(resourceAddress, "namespaced1", "namespaced1", types.OpenApiMetadataStringEntry, "TENANT", "namespace", "false"),
					testCheckOpenApiMetadataEntrySetElemNestedAttrs(resourceAddress, "provider1", "provider1", types.OpenApiMetadataStringEntry, "PROVIDER", "", "false"),
				),
			},
			{
				Config: deleteHcl,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddress, "name", t.Name()),
					resource.TestCheckResourceAttr(resourceAddress, "metadata_entry.#", "0"),
				),
			},
			{
				Config: withDefaults,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddress, "name", t.Name()),
					resource.TestCheckResourceAttr(resourceAddress, "metadata_entry.#", "1"),
					testCheckOpenApiMetadataEntrySetElemNestedAttrs(resourceAddress, "defaultKey", "defaultValue", types.OpenApiMetadataStringEntry, "TENANT", "", "false"),
				),
			},
		},
	})
	postTestChecks(t)
}

// getOpenApiMetadataTestingHcl gets valid metadata entries to inject them into an HCL for testing
func getOpenApiMetadataTestingHcl(stringEntries, numberEntries, boolEntries, readonlyEntries, namespacedEntries, providerEntries int) string {
	hcl := ""
	for i := 1; i <= stringEntries; i++ {
		hcl += getOpenApiMetadataEntryHcl(fmt.Sprintf("stringKey%d", i), fmt.Sprintf("stringValue%d", i), types.OpenApiMetadataStringEntry, "TENANT", "", "false")
	}
	for i := 1; i <= numberEntries; i++ {
		hcl += getOpenApiMetadataEntryHcl(fmt.Sprintf("numberKey%d", i), fmt.Sprintf("%d", i), types.OpenApiMetadataNumberEntry, "TENANT", "", "false")
	}
	for i := 1; i <= boolEntries; i++ {
		hcl += getOpenApiMetadataEntryHcl(fmt.Sprintf("boolKey%d", i), strconv.FormatBool(i%2 == 0), types.OpenApiMetadataBooleanEntry, "TENANT", "", "false")
	}
	for i := 1; i <= readonlyEntries; i++ {
		hcl += getOpenApiMetadataEntryHcl(fmt.Sprintf("readOnly%d", i), fmt.Sprintf("readOnly%d", i), types.OpenApiMetadataStringEntry, "TENANT", "", "true")
	}
	for i := 1; i <= namespacedEntries; i++ {
		hcl += getOpenApiMetadataEntryHcl(fmt.Sprintf("namespaced%d", i), fmt.Sprintf("namespaced%d", i), types.OpenApiMetadataStringEntry, "TENANT", "namespace", "false")
	}
	for i := 1; i <= providerEntries; i++ {
		hcl += getOpenApiMetadataEntryHcl(fmt.Sprintf("provider%d", i), fmt.Sprintf("provider%d", i), types.OpenApiMetadataStringEntry, "PROVIDER", "", "false")
	}
	return hcl

}

func getOpenApiMetadataEntryHcl(key, value, typedValue, domain, namespace, readonly string) string {
	hclNamespace := ""
	if namespace != "" {
		hclNamespace = `namespace   = "` + namespace + `"`
	}

	return `
		  metadata_entry {
			key         = "` + key + `"
			value       = "` + value + `"
			type        = "` + typedValue + `"
			domain      = "` + domain + `"
            ` + hclNamespace + `
			readonly    = ` + readonly + `
		  }`
}

// testCheckOpenApiMetadataEntrySetElemNestedAttrs asserts that a given metadata_entry has the expected input for the given resourceAddress.
func testCheckOpenApiMetadataEntrySetElemNestedAttrs(resourceAddress, expectedKey, expectedValue, expectedType, expectedDomain, expectedNamespace, expectedReadonly string) resource.TestCheckFunc {
	return resource.TestCheckTypeSetElemNestedAttrs(resourceAddress, "metadata_entry.*",
		map[string]string{
			"key":       expectedKey,
			"value":     expectedValue,
			"type":      expectedType,
			"domain":    expectedDomain,
			"readonly":  expectedReadonly,
			"namespace": expectedNamespace,
		},
	)
}
