//go:build catalog || disk || network || nsxt || vdc || org || vapp || vm || functional || ALL

package vcd

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"github.com/vmware/go-vcloud-director/v2/util"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// testMetadataEntryCRUD executes a test that asserts CRUD operation behaviours of "metadata_entry" attribute in the given HCL
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
func testMetadataEntryCRUD(t *testing.T, resourceTemplate, resourceAddress, datasourceTemplate, datasourceAddress string, extraParams StringMap) {
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
	params["Metadata"] = getMetadataTestingHcl(1, 1, 1, 1, 0, 0)
	createHcl := templateFill(resourceTemplate, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", createHcl)

	params["FuncName"] = t.Name() + "WithDatasource"
	withDatasourceHcl := templateFill(datasourceTemplate+"\n# skip-binary-test\n"+resourceTemplate, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", withDatasourceHcl)

	params["FuncName"] = t.Name() + "DeleteOneKey"
	params["Metadata"] = getMetadataTestingHcl(1, 1, 0, 1, 0, 0)
	deleteOneKeyHcl := templateFill(resourceTemplate, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", deleteOneKeyHcl)

	params["FuncName"] = t.Name() + "Update"
	params["Metadata"] = strings.NewReplacer("stringValue", "stringValueUpdated", "2022-10-", "2021-10-").Replace(params["Metadata"].(string))
	updateHcl := templateFill(resourceTemplate, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", updateHcl)

	params["FuncName"] = t.Name() + "Delete"
	params["Metadata"] = "metadata_entry {}"
	deleteHcl := templateFill(resourceTemplate, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", deleteHcl)

	params["FuncName"] = t.Name() + "CreateWithSystem"
	params["Metadata"] = getMetadataTestingHcl(0, 0, 0, 0, 1, 1)
	createWithSystemHcl := templateFill(resourceTemplate, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", createWithSystemHcl)

	params["FuncName"] = t.Name() + "UpdateWithSystem"
	params["Metadata"] = strings.NewReplacer("privateValue", "privateValueUpdated").Replace(params["Metadata"].(string))
	updateWithSystemHcl := templateFill(resourceTemplate, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", updateWithSystemHcl)

	params["FuncName"] = t.Name() + "DeleteWithSystem"
	params["Metadata"] = "metadata_entry {}"
	deleteWithSystemHcl := templateFill(resourceTemplate, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", deleteWithSystemHcl)

	params["FuncName"] = t.Name() + "WrongMetadataEntry"
	params["Metadata"] = "metadata_entry {\n\tkey = \"foo\"\n}"
	wrongMetadataEntryHcl := templateFill("# skip-binary-test\n"+resourceTemplate, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", wrongMetadataEntryHcl)

	params["FuncName"] = t.Name() + "WrongDomain"
	params["Metadata"] = getMetadataTestingHcl(0, 0, 0, 0, 1, 0)
	wrongDomainHcl := templateFill("# skip-binary-test\n"+resourceTemplate, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", wrongDomainHcl)

	// These are for the deprecated `metadata` value, to minimize possible regressions
	params["FuncName"] = t.Name() + "DeprecatedCreate"
	params["Metadata"] = "metadata = {\n\tfoo = \"bar\"\n}"
	deprecatedCreateHcl := templateFill(resourceTemplate, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", deprecatedCreateHcl)

	params["FuncName"] = t.Name() + "DeprecatedUpdate"
	params["Metadata"] = "metadata = {\n\tfoo = \"bar2\"\n}"
	deprecatedUpdateHcl := templateFill(resourceTemplate, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", deprecatedCreateHcl)

	params["FuncName"] = t.Name() + "DeprecatedDelete"
	params["Metadata"] = "metadata = {}"
	deprecatedDeleteHcl := templateFill(resourceTemplate, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", deprecatedCreateHcl)

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
					resource.TestCheckResourceAttr(resourceAddress, "metadata.%", "0"), // Deprecated
				),
			},
			{
				Config: createHcl,
				Taint:  []string{resourceAddress}, // Forces re-creation to test Create with metadata.
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddress, "name", t.Name()),
					resource.TestCheckResourceAttr(resourceAddress, "metadata_entry.#", "4"),
					testCheckMetadataEntrySetElemNestedAttrs(resourceAddress, "stringKey1", "stringValue1", types.MetadataStringValue, types.MetadataReadWriteVisibility, "false"),
					testCheckMetadataEntrySetElemNestedAttrs(resourceAddress, "numberKey1", "1", types.MetadataNumberValue, types.MetadataReadWriteVisibility, "false"),
					testCheckMetadataEntrySetElemNestedAttrs(resourceAddress, "boolKey1", "false", types.MetadataBooleanValue, types.MetadataReadWriteVisibility, "false"),
					testCheckMetadataEntrySetElemNestedAttrs(resourceAddress, "dateKey1", "2022-10-01T12:00:00.000Z", types.MetadataDateTimeValue, types.MetadataReadWriteVisibility, "false"),
				),
			},
			{
				Config: withDatasourceHcl,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceAddress, "id", resourceAddress, "id"),
					resource.TestCheckResourceAttr(resourceAddress, "metadata_entry.#", "4"),
					resource.TestCheckResourceAttr(datasourceAddress, "metadata_entry.#", "4"),
					testCheckMetadataEntrySetElemNestedAttrs(datasourceAddress, "stringKey1", "stringValue1", types.MetadataStringValue, types.MetadataReadWriteVisibility, "false"),
					testCheckMetadataEntrySetElemNestedAttrs(datasourceAddress, "numberKey1", "1", types.MetadataNumberValue, types.MetadataReadWriteVisibility, "false"),
					testCheckMetadataEntrySetElemNestedAttrs(datasourceAddress, "boolKey1", "false", types.MetadataBooleanValue, types.MetadataReadWriteVisibility, "false"),
					testCheckMetadataEntrySetElemNestedAttrs(datasourceAddress, "dateKey1", "2022-10-01T12:00:00.000Z", types.MetadataDateTimeValue, types.MetadataReadWriteVisibility, "false"),
				),
			},
			{
				Config: deleteOneKeyHcl,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddress, "name", t.Name()),
					resource.TestCheckResourceAttr(resourceAddress, "metadata_entry.#", "3"),
					// The bool is deleted
					testCheckMetadataEntrySetElemNestedAttrs(resourceAddress, "stringKey1", "stringValue1", types.MetadataStringValue, types.MetadataReadWriteVisibility, "false"),
					testCheckMetadataEntrySetElemNestedAttrs(resourceAddress, "numberKey1", "1", types.MetadataNumberValue, types.MetadataReadWriteVisibility, "false"),
					testCheckMetadataEntrySetElemNestedAttrs(resourceAddress, "dateKey1", "2022-10-01T12:00:00.000Z", types.MetadataDateTimeValue, types.MetadataReadWriteVisibility, "false"),
				),
			},
			{
				Config: updateHcl,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddress, "name", t.Name()),
					resource.TestCheckResourceAttr(resourceAddress, "metadata_entry.#", "3"),
					// Updated values:
					testCheckMetadataEntrySetElemNestedAttrs(resourceAddress, "stringKey1", "stringValueUpdated1", types.MetadataStringValue, types.MetadataReadWriteVisibility, "false"),
					testCheckMetadataEntrySetElemNestedAttrs(resourceAddress, "dateKey1", "2021-10-01T12:00:00.000Z", types.MetadataDateTimeValue, types.MetadataReadWriteVisibility, "false"),
					// Not updated values:
					testCheckMetadataEntrySetElemNestedAttrs(resourceAddress, "numberKey1", "1", types.MetadataNumberValue, types.MetadataReadWriteVisibility, "false"),
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
				Config: createWithSystemHcl,
				SkipFunc: func() (bool, error) {
					return !usingSysAdmin(), nil
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddress, "name", t.Name()),
					resource.TestCheckResourceAttr(resourceAddress, "metadata_entry.#", "2"),
					testCheckMetadataEntrySetElemNestedAttrs(resourceAddress, "readOnlyKey1", "readOnlyValue1", types.MetadataStringValue, types.MetadataReadOnlyVisibility, "true"),
					testCheckMetadataEntrySetElemNestedAttrs(resourceAddress, "privateKey1", "privateValue1", types.MetadataStringValue, types.MetadataHiddenVisibility, "true"),
				),
			},
			{
				Config: updateWithSystemHcl,
				SkipFunc: func() (bool, error) {
					return !usingSysAdmin(), nil
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddress, "name", t.Name()),
					resource.TestCheckResourceAttr(resourceAddress, "metadata_entry.#", "2"),
					testCheckMetadataEntrySetElemNestedAttrs(resourceAddress, "readOnlyKey1", "readOnlyValue1", types.MetadataStringValue, types.MetadataReadOnlyVisibility, "true"),
					testCheckMetadataEntrySetElemNestedAttrs(resourceAddress, "privateKey1", "privateValueUpdated1", types.MetadataStringValue, types.MetadataHiddenVisibility, "true"),
				),
			},
			{
				PreConfig: func() {
					util.Logger.Printf("a")
				},
				Config: deleteWithSystemHcl,
				SkipFunc: func() (bool, error) {
					return !usingSysAdmin(), nil
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddress, "name", t.Name()),
					// This is a side effect of having `metadata_entry` as Computed to be able to delete metadata.
					resource.TestCheckResourceAttr(resourceAddress, "metadata_entry.#", "1"),
					testCheckMetadataEntrySetElemNestedAttrs(resourceAddress, "", "", "", "", "false"),
				),
			},
			{
				Config:      wrongMetadataEntryHcl,
				ExpectError: regexp.MustCompile(".*all fields in a metadata_entry are required, but got some empty.*"),
			},
			{
				Config: wrongDomainHcl,
				SkipFunc: func() (bool, error) {
					return usingSysAdmin(), nil
				},
				ExpectError: regexp.MustCompile(".*This operation is denied*"),
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
					resource.TestCheckResourceAttr(resourceAddress, "metadata.%", "0"),
				),
			},
		},
	})
	postTestChecks(t)
}

func testMetadataEntryIgnore(t *testing.T, resourceTemplate, resourceAddress, datasourceTemplate, datasourceAddress, objectType string, retrieveObjectById func(string) (metadataCompatible, error), extraParams StringMap) {
	preTestChecks(t)
	var params = StringMap{
		"FuncName": t.Name() + "-Step1",
		"Org":      testConfig.VCD.Org,
		"Vdc":      testConfig.Nsxt.Vdc,
		"Name":     t.Name(),
		"Metadata": " ",
	}

	for extraParam, extraParamValue := range extraParams {
		params[extraParam] = extraParamValue
	}
	testParamsNotEmpty(t, params)

	step1 := templateFill(resourceTemplate, params)
	debugPrintf("#[DEBUG] CONFIGURATION 1: %s", step1)
	params["FuncName"] = t.Name() + "-Step2"
	params["Metadata"] = getMetadataTestingHcl(1, 0, 0, 0, 0, 0)
	step2 := templateFill(resourceTemplate, params)
	debugPrintf("#[DEBUG] CONFIGURATION 2: %s", step2)
	params["FuncName"] = t.Name() + "-Step3"
	step3 := templateFill("# skip-binary-test\n"+resourceTemplate+datasourceTemplate, params)
	debugPrintf("#[DEBUG] CONFIGURATION 3: %s", step3)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	cachedId := testCachedFieldValue{}
	// We need to create metadata with the `retrieveObjectById` func, which will be called
	// after the testing SDK creates the provider config with metadata to ignore, hence we wouldn't be able to do it,
	// unless we disable the cache to have separated clients.
	backupEnableConnectionCache := enableConnectionCache
	enableConnectionCache = false
	cachedVCDClients.reset()
	defer func() {
		enableConnectionCache = backupEnableConnectionCache
		cachedVCDClients.reset()
	}()

	testFunc := func(ignoredMetadata []map[string]string, expectedMetadata, expectedMetadataInVcd int) {
		var object metadataCompatible
		resource.Test(t, resource.TestCase{
			ProviderFactories: map[string]func() (*schema.Provider, error){
				providerVcdSystem: func() (*schema.Provider, error) {
					newProvider := Provider()
					newProvider.ConfigureContextFunc = func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
						// We configure the provider this way to be able to test that providerConfigure
						// retrieves and parses the 'ignore_metadata' argument correctly
						err := d.Set("ignore_metadata", ignoredMetadata)
						if err != nil {
							return nil, diag.FromErr(err)
						}
						return providerConfigure(ctx, d)
					}
					return newProvider, nil
				},
			},
			Steps: []resource.TestStep{
				// Create a resource without metadata
				{
					Config: step1,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrSet(resourceAddress, "id"),
						cachedId.cacheTestResourceFieldValue(resourceAddress, "id"),
						resource.TestCheckResourceAttr(resourceAddress, "metadata_entry.#", "0"),
					),
				},
				// In this step, an external actor (simulated in PreConfig by using the Go SDK) adds a metadata entry to the resource.
				// The provider is configured to ignore it. This entry is always added as the client cache is deactivated.
				{
					PreConfig: func() {
						var err error
						object, err = retrieveObjectById(cachedId.fieldValue)
						if err != nil {
							t.Errorf("could not add metadata to object with ID '%s': %s", cachedId.fieldValue, err)
						}
						err = object.AddMetadataEntryWithVisibility("foo", "bar", types.MetadataStringValue, types.MetadataReadWriteVisibility, false)
						if err != nil {
							t.Errorf("could not add metadata to object with ID '%s': %s", cachedId.fieldValue, err)
						}
					},
					Config: step2,
					Check: resource.ComposeAggregateTestCheckFunc(
						// We need to check both metadata in state and metadata in VCD
						func(state *terraform.State) error {
							realMetadata, err := object.GetMetadata()
							if err != nil {
								return err
							}
							if len(realMetadata.MetadataEntry) != expectedMetadataInVcd {
								return fmt.Errorf("expected %d metadata entries in VCD but got %d", expectedMetadataInVcd, len(realMetadata.MetadataEntry))
							}
							return nil
						},
						resource.TestCheckResourceAttr(resourceAddress, "metadata_entry.#", fmt.Sprintf("%d", expectedMetadata)),
						testCheckMetadataEntrySetElemNestedAttrs(resourceAddress, "stringKey1", "stringValue1", types.MetadataStringValue, types.MetadataReadWriteVisibility, "false"),
					),
				},
				// Test data source metadata
				{
					Config: step3,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(datasourceAddress, "metadata_entry.#", fmt.Sprintf("%d", expectedMetadata)),
						testCheckMetadataEntrySetElemNestedAttrs(datasourceAddress, "stringKey1", "stringValue1", types.MetadataStringValue, types.MetadataReadWriteVisibility, "false"),
						testCheckMetadataEntrySetElemNestedAttrs(datasourceAddress, "stringKey1", "stringValue1", types.MetadataStringValue, types.MetadataReadWriteVisibility, "false"),
						resource.TestCheckResourceAttrPair(datasourceAddress, "id", resourceAddress, "id"),
						resource.TestCheckResourceAttrPair(datasourceAddress, "metadata_entry.#", resourceAddress, "metadata_entry.#"),
					),
				},
			},
		})
	}

	t.Run("filter by object type and specific key", func(t *testing.T) {
		testFunc([]map[string]string{
			{
				"object_type": objectType,
				"key_regex":   "foo",
			},
		}, 1, 2) // We expect 1 in Terraform but foo is in VCD, so it has 2
	})
	t.Run("filter by object type and specific value", func(t *testing.T) {
		testFunc([]map[string]string{
			{
				"object_type": objectType,
				"value_regex": "bar",
			},
		}, 1, 2) // We expect 1 in Terraform but foo is in VCD, so it has 2
	})
	t.Run("filter by object type and key that doesn't match", func(t *testing.T) {
		testFunc([]map[string]string{
			{
				"object_type": objectType,
				"key_regex":   "notmatch",
			},
		}, 1, 1) // We expect 1 because foo has been deleted by Terraform as it was not ignored
	})
	t.Run("filter by object type and value that doesn't match", func(t *testing.T) {
		testFunc([]map[string]string{
			{
				"object_type": objectType,
				"value_regex": "notmatch",
			},
		}, 1, 1) // We expect 1 because foo has been deleted by Terraform as it was not ignored
	})
	t.Run("filter by object name and specific key", func(t *testing.T) {
		testFunc([]map[string]string{
			{
				"object_name": t.Name(),
				"key_regex":   "foo",
			},
		}, 1, 2) // We expect 1 in Terraform but foo is in VCD, so it has 2
	})
	t.Run("filter by object name and specific value", func(t *testing.T) {
		testFunc([]map[string]string{
			{
				"object_name": t.Name(),
				"value_regex": "bar",
			},
		}, 1, 2) // We expect 1 in Terraform but foo is in VCD, so it has 2
	})
	t.Run("filter by object name and key that doesn't match", func(t *testing.T) {
		testFunc([]map[string]string{
			{
				"object_name": t.Name(),
				"key_regex":   "notmatch",
			},
		}, 1, 1) // We expect 1 because foo has been deleted by Terraform as it was not ignored
	})
	t.Run("filter by object name and value that doesn't match", func(t *testing.T) {
		testFunc([]map[string]string{
			{
				"object_name": t.Name(),
				"value_regex": "notmatch",
			},
		}, 1, 1) // We expect 1 because foo has been deleted by Terraform as it was not ignored
	})
	t.Run("filter by key and value that match", func(t *testing.T) {
		testFunc([]map[string]string{
			{
				"key_regex":   "foo",
				"value_regex": "bar",
			},
		}, 1, 2) // We expect 1 in Terraform but foo is in VCD, so it has 2
	})
	t.Run("filter by key and value that don't match", func(t *testing.T) {
		testFunc([]map[string]string{
			{
				"key_regex":   "foo",
				"value_regex": "barz",
			},
		}, 1, 1) // We expect 1 because foo has been deleted by Terraform as it was not ignored
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
}
`
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
