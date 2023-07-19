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
	"os"
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

// testMetadataEntryIgnore executes a test that asserts that the "ignore_metadata_changes" Provider argument allows to ignore
// metadata entries in all cases.
//
// Tests:
// - Step 1: Create the resource with no metadata
// - Pre-Step 2: SDK creates a metadata entry to simulate an external actor adding metadata to the resource.
// - Step 2: Add a metadata entry to the resource
// - Step 3: Add a data source that fetches the created resource.
//
// The different ignore_metadata_changes sub-tests check what happens if the filter matches or doesn't match the metadata entry
// added in Pre-Step 2. If it doesn't match, Terraform will delete it from VCD. If it match, it gets ignored as it doesn't exist.
func testMetadataEntryIgnore(t *testing.T, resourceTemplate, resourceAddress, datasourceTemplate, datasourceAddress string, retrieveObjectById func(*VCDClient, string) (metadataCompatible, error), extraParams StringMap) {
	preTestChecks(t)
	resourceType := strings.Split(resourceAddress, ".")[0]
	var params = StringMap{
		"FuncName": t.Name() + "-Step1",
		"Org":      testConfig.VCD.Org,
		"Vdc":      testConfig.Nsxt.Vdc,
		"Name":     t.Name(),
		"Metadata": " ",
		// The IgnoreMetadataBlock entry below is for binary tests
		"IgnoreMetadataBlock": "ignore_metadata_changes {\n\tresource_type = \"" + resourceType + "\"\n\tresource_name   = \"" + t.Name() + "\"\n\tkey_regex     = \".*\"\n\tvalue_regex   = \".*\"\n\tconflict_action = \"warn\"\n}",
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
	step3 := templateFill(resourceTemplate+datasourceTemplate, params)
	debugPrintf("#[DEBUG] CONFIGURATION 3: %s", step3)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	// We will cache the ID of the created resource after Step 1, so it can be used afterward.
	cachedId := testCachedFieldValue{}

	// We need to disable the Client cache because we need different instances on each subtest, and we need a "special"
	// client for Step 2 PreConfig which doesn't have IgnoredMetadata configured.
	// If we don't disable the cache, the same clients would be reused for all subtests and
	// the Step 2 PreConfig would fail to create metadata as it would be ignored.
	backupEnableConnectionCache := enableConnectionCache
	enableConnectionCache = false
	cachedVCDClients.reset()
	vcdClient := createSystemTemporaryVCDConnection()
	defer func() {
		enableConnectionCache = backupEnableConnectionCache
		cachedVCDClients.reset()
	}()

	testFunc := func(t *testing.T, vcdClient *VCDClient, ignoredMetadata []map[string]string, expectedMetadataInVcd int) {
		var object metadataCompatible
		resource.Test(t, resource.TestCase{
			ProviderFactories: map[string]func() (*schema.Provider, error){
				providerVcdSystem: func() (*schema.Provider, error) {
					newProvider := Provider()
					newProvider.ConfigureContextFunc = func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
						// We configure the provider this way to be able to test that providerConfigure
						// retrieves and parses the 'ignore_metadata_changes' argument correctly.
						err := d.Set("ignore_metadata_changes", ignoredMetadata)
						if err != nil {
							return nil, diag.FromErr(err)
						}
						return providerConfigure(ctx, d)
					}
					return newProvider, nil
				},
			},
			Steps: []resource.TestStep{
				// Create a resource without metadata.
				{
					Config: step1,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttrSet(resourceAddress, "id"),
						cachedId.cacheTestResourceFieldValue(resourceAddress, "id"),
						resource.TestCheckResourceAttr(resourceAddress, "metadata_entry.#", "0"),
					),
				},
				// In this step, an external actor (simulated in PreConfig by using the Go SDK) adds a metadata entry to the resource.
				// The provider is configured to ignore it.
				{
					PreConfig: func() {
						var err error
						object, err = retrieveObjectById(vcdClient, cachedId.fieldValue)
						if err != nil {
							t.Fatalf("could not add metadata to object with ID '%s': %s", cachedId.fieldValue, err)
						}
						err = object.AddMetadataEntryWithVisibility("foo", "bar", types.MetadataStringValue, types.MetadataReadWriteVisibility, false)
						if err != nil {
							t.Fatalf("could not add metadata to object with ID '%s': %s", cachedId.fieldValue, err)
						}
						// Check that the metadata was added and not ignored
						_, err = object.GetMetadataByKey("foo", false)
						if err != nil {
							t.Fatalf("should have retrieved metadata with key 'foo' from object with ID '%s': %s", cachedId.fieldValue, err)
						}
					},
					Config: step2,
					Check: resource.ComposeAggregateTestCheckFunc(
						// We need to check both metadata in state and metadata in VCD to assert that the filter works.
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
						resource.TestCheckResourceAttr(resourceAddress, "metadata_entry.#", "1"),
						testCheckMetadataEntrySetElemNestedAttrs(resourceAddress, "stringKey1", "stringValue1", types.MetadataStringValue, types.MetadataReadWriteVisibility, "false"),
					),
				},
				// Test data source metadata.
				{
					Config: step3,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(datasourceAddress, "metadata_entry.#", "1"),
						testCheckMetadataEntrySetElemNestedAttrs(datasourceAddress, "stringKey1", "stringValue1", types.MetadataStringValue, types.MetadataReadWriteVisibility, "false"),
						testCheckMetadataEntrySetElemNestedAttrs(datasourceAddress, "stringKey1", "stringValue1", types.MetadataStringValue, types.MetadataReadWriteVisibility, "false"),
						resource.TestCheckResourceAttrPair(datasourceAddress, "id", resourceAddress, "id"),
						resource.TestCheckResourceAttrPair(datasourceAddress, "metadata_entry.#", resourceAddress, "metadata_entry.#"),
					),
				},
			},
		})
	}

	testName := t.Name()
	t.Run("filter by all options that match", func(t *testing.T) {
		testFunc(t, vcdClient, []map[string]string{
			{
				"resource_type": resourceType,
				"resource_name": testName,
				"key_regex":     "foo",
				"value_regex":   "bar",
			},
		}, 2) // As 'foo' is correctly ignored, VCD should always have 2 entries, one created by the test and 'foo'.
	})

	// This environment variable controls the execution of all remaining test cases.
	// As client cache is disabled for the whole test, it can take a long time to run, specially if it also involves
	// VM creation.
	testAll := os.Getenv("TEST_VCD_METADATA_IGNORE")
	if testAll != "" {
		t.Run("filter by object type and specific key", func(t *testing.T) {
			testFunc(t, vcdClient, []map[string]string{
				{
					"resource_type": resourceType,
					"key_regex":     "foo",
				},
			}, 2) // As 'foo' is correctly ignored, VCD should always have 2 entries, one created by the test and 'foo'.
		})
		t.Run("filter by object type and specific value", func(t *testing.T) {
			testFunc(t, vcdClient, []map[string]string{
				{
					"resource_type": resourceType,
					"value_regex":   "bar",
				},
			}, 2) // As 'foo' (with value 'bar') is correctly ignored, VCD should always have 2 entries, one created by the test and 'foo'.
		})
		t.Run("filter by object type and key that doesn't match", func(t *testing.T) {
			testFunc(t, vcdClient, []map[string]string{
				{
					"resource_type": resourceType,
					"key_regex":     "notmatch",
				},
			}, 1) // We expect 1 because 'foo' has been deleted by Terraform as it was not ignored
		})
		t.Run("filter by object type and value that doesn't match", func(t *testing.T) {
			testFunc(t, vcdClient, []map[string]string{
				{
					"resource_type": resourceType,
					"value_regex":   "notmatch",
				},
			}, 1) // We expect 1 because 'foo' 'has been deleted by Terraform as it was not ignored
		})
		t.Run("filter by object name and specific key", func(t *testing.T) {
			testFunc(t, vcdClient, []map[string]string{
				{
					"resource_name": testName,
					"key_regex":     "foo",
				},
			}, 2) // As 'foo' is correctly ignored, VCD should always have 2 entries, one created by the test and 'foo'.
		})
		t.Run("filter by object name and specific value", func(t *testing.T) {
			testFunc(t, vcdClient, []map[string]string{
				{
					"resource_name": testName,
					"value_regex":   "bar",
				},
			}, 2) // As 'foo' (with value 'bar') is correctly ignored, VCD should always have 2 entries, one created by the test and 'foo'.
		})
		t.Run("filter by object name and key that doesn't match", func(t *testing.T) {
			testFunc(t, vcdClient, []map[string]string{
				{
					"resource_name": testName,
					"key_regex":     "notmatch",
				},
			}, 1) // We expect 1 because 'foo' has been deleted by Terraform as it was not ignored
		})
		t.Run("filter by object name and value that doesn't match", func(t *testing.T) {
			testFunc(t, vcdClient, []map[string]string{
				{
					"resource_name": testName,
					"value_regex":   "notmatch",
				},
			}, 1) // We expect 1 because 'foo' has been deleted by Terraform as it was not ignored
		})
		t.Run("filter by key and value that don't match", func(t *testing.T) {
			testFunc(t, vcdClient, []map[string]string{
				{
					"key_regex":   "foo",
					"value_regex": "barz",
				},
			}, 1) // We expect 1 because 'foo' has been deleted by Terraform as it was not ignored
		})
	}

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
