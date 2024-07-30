//go:build rde || functional || ALL

package vcd

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// testOpenApiMetadataEntryCRUD executes a test that asserts CRUD operation behaviours of "metadata_entry" attribute in the given HCL
// templates, that must correspond to a resource and a data source referencing this resource.
// The HCL template requires {{.Name}} and {{.Metadata}} fields, and the usual {{.Org}} and {{.Vdc}}.
// You can add extra parameters as well to inject in the given HCL template, or override these mentioned ones.
// The data source HCL is always concatenated to the resource after creation, and it's skipped on binary tests.
func testOpenApiMetadataEntryCRUD(t *testing.T, resourceTemplate, resourceAddress, datasourceTemplate, datasourceAddress string, extraParams StringMap) {
	preTestChecks(t)
	var params = StringMap{
		"Org":  testConfig.VCD.Org,
		"Vdc":  testConfig.Nsxt.Vdc,
		"Name": t.Name(),
	}

	// This output allows to perform some assertions on the ID inside a TypeSet,
	// which is impossible to obtain otherwise.
	outputHcl := `
output "metadata_id" {
  value = tolist(` + resourceAddress + `.metadata_entry)[0].id
}
`
	templateWithOutput := resourceTemplate + outputHcl

	for extraParam, extraParamValue := range extraParams {
		params[extraParam] = extraParamValue
	}
	testParamsNotEmpty(t, params)

	params["FuncName"] = t.Name() + "NoMetadata"
	params["Metadata"] = " "
	noMetadataHcl := templateFill(resourceTemplate, params)
	debugPrintf("#[DEBUG] CONFIGURATION NoMetadata: %s", noMetadataHcl)

	params["FuncName"] = t.Name() + "Create"
	params["Metadata"] = getOpenApiMetadataTestingHcl(1, 1, 1, 1, 2, 1, 1)
	createHcl := templateFill(templateWithOutput, params)
	debugPrintf("#[DEBUG] CONFIGURATION Create: %s", createHcl)

	params["FuncName"] = t.Name() + "WithDatasource"
	withDatasourceHcl := templateFill(datasourceTemplate+"\n# skip-binary-test\n"+templateWithOutput, params)
	debugPrintf("#[DEBUG] CONFIGURATION WithDatasource: %s", withDatasourceHcl)

	params["FuncName"] = t.Name() + "DeleteOneKey"
	params["Metadata"] = getOpenApiMetadataTestingHcl(1, 1, 0, 1, 2, 1, 1)
	deleteOneKeyHcl := templateFill(templateWithOutput, params)
	debugPrintf("#[DEBUG] CONFIGURATION DeleteOneKey: %s", deleteOneKeyHcl)

	params["FuncName"] = t.Name() + "Update"
	params["Metadata"] = strings.NewReplacer("stringValue", "stringValueUpdated").Replace(params["Metadata"].(string))
	updateHcl := templateFill(templateWithOutput, params)
	debugPrintf("#[DEBUG] CONFIGURATION Update: %s", updateHcl)

	params["FuncName"] = t.Name() + "UpdateForceRecreate"
	params["Metadata"] = strings.NewReplacer("NumberEntry", "StringEntry").Replace(params["Metadata"].(string))
	updateForceRecreateHcl := templateFill(templateWithOutput, params)
	debugPrintf("#[DEBUG] CONFIGURATION UpdateForceRecreate: %s", updateForceRecreateHcl)

	params["FuncName"] = t.Name() + "Delete"
	params["Metadata"] = " "
	deleteHcl := templateFill(resourceTemplate, params)
	debugPrintf("#[DEBUG] CONFIGURATION Delete: %s", deleteHcl)

	params["FuncName"] = t.Name() + "WithDefaults"
	params["Metadata"] = "metadata_entry {\n\tkey = \"defaultKey\"\nvalue = \"defaultValue\"\n}"
	withDefaults := templateFill(resourceTemplate, params)
	debugPrintf("#[DEBUG] CONFIGURATION WithDefaults: %s", withDefaults)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	// This is used to validate that metadata IDs don't change despite of an update/delete.
	cachedId := testCachedFieldValue{}
	testIdsDontChange := func() func(s *terraform.State) error {
		return func(s *terraform.State) error {
			rs, ok := s.RootModule().Outputs["metadata_id"]
			if !ok {
				return fmt.Errorf("output 'metadata_id' not found")
			}

			value := rs.String()
			if cachedId.fieldValue != value || value == "" {
				return fmt.Errorf("expected metadata_id to be '%s' but it changed to '%s'", cachedId.fieldValue, value)
			}
			return nil
		}
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
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Outputs["metadata_id"]
						if !ok {
							return fmt.Errorf("output 'metadata_id' not found")
						}
						cachedId.fieldValue = rs.String()
						return nil
					},
					resource.TestCheckResourceAttr(resourceAddress, "name", t.Name()),
					resource.TestCheckResourceAttr(resourceAddress, "metadata_entry.#", "8"),
					testCheckOpenApiMetadataEntrySetElemNestedAttrs(resourceAddress, "stringKey1", "stringValue1", types.OpenApiMetadataStringEntry, "TENANT", "", "false", "false"),
					testCheckOpenApiMetadataEntrySetElemNestedAttrs(resourceAddress, "numberKey1", "1", types.OpenApiMetadataNumberEntry, "TENANT", "", "false", "false"),
					testCheckOpenApiMetadataEntrySetElemNestedAttrs(resourceAddress, "boolKey1", "false", types.OpenApiMetadataStringEntry, "TENANT", "", "false", "false"),
					testCheckOpenApiMetadataEntrySetElemNestedAttrs(resourceAddress, "readOnly1", "readOnly1", types.OpenApiMetadataStringEntry, "TENANT", "", "true", "false"),
					testCheckOpenApiMetadataEntrySetElemNestedAttrs(resourceAddress, "namespace", "namespace1", types.OpenApiMetadataStringEntry, "TENANT", "namespace1", "false", "false"),
					testCheckOpenApiMetadataEntrySetElemNestedAttrs(resourceAddress, "namespace", "namespace2", types.OpenApiMetadataStringEntry, "TENANT", "namespace2", "false", "false"),
					testCheckOpenApiMetadataEntrySetElemNestedAttrs(resourceAddress, "provider1", "provider1", types.OpenApiMetadataStringEntry, "PROVIDER", "", "false", "false"),
					testCheckOpenApiMetadataEntrySetElemNestedAttrs(resourceAddress, "persistent1", "persistent1", types.OpenApiMetadataStringEntry, "TENANT", "", "false", "true"),
				),
			},
			{
				Config: withDatasourceHcl,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceAddress, "id", resourceAddress, "id"),
					resource.TestCheckResourceAttr(resourceAddress, "metadata_entry.#", "8"),
					resource.TestCheckResourceAttr(datasourceAddress, "metadata_entry.#", "8"),
					testCheckOpenApiMetadataEntrySetElemNestedAttrs(resourceAddress, "stringKey1", "stringValue1", types.OpenApiMetadataStringEntry, "TENANT", "", "false", "false"),
					testCheckOpenApiMetadataEntrySetElemNestedAttrs(resourceAddress, "numberKey1", "1", types.OpenApiMetadataNumberEntry, "TENANT", "", "false", "false"),
					testCheckOpenApiMetadataEntrySetElemNestedAttrs(resourceAddress, "boolKey1", "false", types.OpenApiMetadataBooleanEntry, "TENANT", "", "false", "false"),
					testCheckOpenApiMetadataEntrySetElemNestedAttrs(resourceAddress, "readOnly1", "readOnly1", types.OpenApiMetadataStringEntry, "TENANT", "", "true", "false"),
					testCheckOpenApiMetadataEntrySetElemNestedAttrs(resourceAddress, "namespace", "namespace1", types.OpenApiMetadataStringEntry, "TENANT", "namespace1", "false", "false"),
					testCheckOpenApiMetadataEntrySetElemNestedAttrs(resourceAddress, "namespace", "namespace2", types.OpenApiMetadataStringEntry, "TENANT", "namespace2", "false", "false"),
					testCheckOpenApiMetadataEntrySetElemNestedAttrs(resourceAddress, "provider1", "provider1", types.OpenApiMetadataStringEntry, "PROVIDER", "", "false", "false"),
					testCheckOpenApiMetadataEntrySetElemNestedAttrs(resourceAddress, "persistent1", "persistent1", types.OpenApiMetadataStringEntry, "TENANT", "", "false", "true"),
				),
			},
			{
				Config: deleteOneKeyHcl,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddress, "name", t.Name()),
					resource.TestCheckResourceAttr(resourceAddress, "metadata_entry.#", "7"),
					// The bool is deleted
					testCheckOpenApiMetadataEntrySetElemNestedAttrs(resourceAddress, "stringKey1", "stringValue1", types.OpenApiMetadataStringEntry, "TENANT", "", "false", "false"),
					testCheckOpenApiMetadataEntrySetElemNestedAttrs(resourceAddress, "numberKey1", "1", types.OpenApiMetadataNumberEntry, "TENANT", "", "false", "false"),
					testCheckOpenApiMetadataEntrySetElemNestedAttrs(resourceAddress, "readOnly1", "readOnly1", types.OpenApiMetadataStringEntry, "TENANT", "", "true", "false"),
					testCheckOpenApiMetadataEntrySetElemNestedAttrs(resourceAddress, "namespace", "namespace1", types.OpenApiMetadataStringEntry, "TENANT", "namespace1", "false", "false"),
					testCheckOpenApiMetadataEntrySetElemNestedAttrs(resourceAddress, "namespace", "namespace2", types.OpenApiMetadataStringEntry, "TENANT", "namespace2", "false", "false"),
					testCheckOpenApiMetadataEntrySetElemNestedAttrs(resourceAddress, "provider1", "provider1", types.OpenApiMetadataStringEntry, "PROVIDER", "", "false", "false"),
					testCheckOpenApiMetadataEntrySetElemNestedAttrs(resourceAddress, "persistent1", "persistent1", types.OpenApiMetadataStringEntry, "TENANT", "", "false", "true"),
					testIdsDontChange(),
				),
			},
			{
				Config: updateHcl,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddress, "name", t.Name()),
					resource.TestCheckResourceAttr(resourceAddress, "metadata_entry.#", "7"),
					// Updated value:
					testCheckOpenApiMetadataEntrySetElemNestedAttrs(resourceAddress, "stringKey1", "stringValueUpdated1", types.OpenApiMetadataStringEntry, "TENANT", "", "false", "false"),
					// Not updated values:
					testCheckOpenApiMetadataEntrySetElemNestedAttrs(resourceAddress, "numberKey1", "1", types.OpenApiMetadataNumberEntry, "TENANT", "", "false", "false"),
					testCheckOpenApiMetadataEntrySetElemNestedAttrs(resourceAddress, "readOnly1", "readOnly1", types.OpenApiMetadataStringEntry, "TENANT", "", "true", "false"),
					testCheckOpenApiMetadataEntrySetElemNestedAttrs(resourceAddress, "namespace", "namespace1", types.OpenApiMetadataStringEntry, "TENANT", "namespace1", "false", "false"),
					testCheckOpenApiMetadataEntrySetElemNestedAttrs(resourceAddress, "namespace", "namespace2", types.OpenApiMetadataStringEntry, "TENANT", "namespace2", "false", "false"),
					testCheckOpenApiMetadataEntrySetElemNestedAttrs(resourceAddress, "provider1", "provider1", types.OpenApiMetadataStringEntry, "PROVIDER", "", "false", "false"),
					testCheckOpenApiMetadataEntrySetElemNestedAttrs(resourceAddress, "persistent1", "persistent1", types.OpenApiMetadataStringEntry, "TENANT", "", "false", "true"),
					testIdsDontChange(),
				),
			},
			{
				Config: updateForceRecreateHcl,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddress, "name", t.Name()),
					resource.TestCheckResourceAttr(resourceAddress, "metadata_entry.#", "7"),
					// Updated value, from number to string, should force a recreation of the entry:
					testCheckOpenApiMetadataEntrySetElemNestedAttrs(resourceAddress, "numberKey1", "1", types.OpenApiMetadataStringEntry, "TENANT", "", "false", "false"),
					// Not updated values:
					testCheckOpenApiMetadataEntrySetElemNestedAttrs(resourceAddress, "stringKey1", "stringValueUpdated1", types.OpenApiMetadataStringEntry, "TENANT", "", "false", "false"),
					testCheckOpenApiMetadataEntrySetElemNestedAttrs(resourceAddress, "readOnly1", "readOnly1", types.OpenApiMetadataStringEntry, "TENANT", "", "true", "false"),
					testCheckOpenApiMetadataEntrySetElemNestedAttrs(resourceAddress, "namespace", "namespace1", types.OpenApiMetadataStringEntry, "TENANT", "namespace1", "false", "false"),
					testCheckOpenApiMetadataEntrySetElemNestedAttrs(resourceAddress, "namespace", "namespace2", types.OpenApiMetadataStringEntry, "TENANT", "namespace2", "false", "false"),
					testCheckOpenApiMetadataEntrySetElemNestedAttrs(resourceAddress, "provider1", "provider1", types.OpenApiMetadataStringEntry, "PROVIDER", "", "false", "false"),
					testCheckOpenApiMetadataEntrySetElemNestedAttrs(resourceAddress, "persistent1", "persistent1", types.OpenApiMetadataStringEntry, "TENANT", "", "false", "true"),
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
					testCheckOpenApiMetadataEntrySetElemNestedAttrs(resourceAddress, "defaultKey", "defaultValue", types.OpenApiMetadataStringEntry, "TENANT", "", "false", "false"),
				),
			},
		},
	})
	postTestChecks(t)
}

// getOpenApiMetadataTestingHcl gets valid metadata entries to inject them into an HCL for testing
func getOpenApiMetadataTestingHcl(stringEntries, numberEntries, boolEntries, readonlyEntries, namespacedEntries, providerEntries, persistentEntries int) string {
	hcl := ""
	for i := 1; i <= stringEntries; i++ {
		hcl += getOpenApiMetadataEntryHcl(fmt.Sprintf("stringKey%d", i), fmt.Sprintf("stringValue%d", i), types.OpenApiMetadataStringEntry, "TENANT", "", "false", "false")
	}
	for i := 1; i <= numberEntries; i++ {
		hcl += getOpenApiMetadataEntryHcl(fmt.Sprintf("numberKey%d", i), fmt.Sprintf("%d", i), types.OpenApiMetadataNumberEntry, "TENANT", "", "false", "false")
	}
	for i := 1; i <= boolEntries; i++ {
		hcl += getOpenApiMetadataEntryHcl(fmt.Sprintf("boolKey%d", i), strconv.FormatBool(i%2 == 0), types.OpenApiMetadataBooleanEntry, "TENANT", "", "false", "false")
	}
	for i := 1; i <= readonlyEntries; i++ {
		hcl += getOpenApiMetadataEntryHcl(fmt.Sprintf("readOnly%d", i), fmt.Sprintf("readOnly%d", i), types.OpenApiMetadataStringEntry, "TENANT", "", "true", "false")
	}
	for i := 1; i <= namespacedEntries; i++ {
		// This one shares the same key, as it is the goal of namespaces
		hcl += getOpenApiMetadataEntryHcl("namespace", fmt.Sprintf("namespace%d", i), types.OpenApiMetadataStringEntry, "TENANT", fmt.Sprintf("namespace%d", i), "false", "false")
	}
	for i := 1; i <= providerEntries; i++ {
		hcl += getOpenApiMetadataEntryHcl(fmt.Sprintf("provider%d", i), fmt.Sprintf("provider%d", i), types.OpenApiMetadataStringEntry, "PROVIDER", "", "false", "false")
	}
	for i := 1; i <= persistentEntries; i++ {
		hcl += getOpenApiMetadataEntryHcl(fmt.Sprintf("persistent%d", i), fmt.Sprintf("persistent%d", i), types.OpenApiMetadataStringEntry, "TENANT", "", "false", "true")
	}
	return hcl

}

func getOpenApiMetadataEntryHcl(key, value, typedValue, domain, namespace, readonly, persistent string) string {
	hclNamespace := ""
	if namespace != "" {
		hclNamespace = `namespace  = "` + namespace + `"`
	}

	return `
  metadata_entry {
    key        = "` + key + `"
    value      = "` + value + `"
	type       = "` + typedValue + `"
	domain     = "` + domain + `"
    ` + hclNamespace + `
	readonly   = ` + readonly + `
    persistent = ` + persistent + `
  }`
}

// testCheckOpenApiMetadataEntrySetElemNestedAttrs asserts that a given metadata_entry has the expected input for the given resourceAddress.
func testCheckOpenApiMetadataEntrySetElemNestedAttrs(resourceAddress, expectedKey, expectedValue, expectedType, expectedDomain, expectedNamespace, expectedReadonly, expectedPersistent string) resource.TestCheckFunc {
	return resource.TestCheckTypeSetElemNestedAttrs(resourceAddress, "metadata_entry.*",
		map[string]string{
			"key":        expectedKey,
			"value":      expectedValue,
			"type":       expectedType,
			"domain":     expectedDomain,
			"readonly":   expectedReadonly,
			"namespace":  expectedNamespace,
			"persistent": expectedPersistent,
		},
	)
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
// added in Pre-Step 2. If it doesn't match, Terraform will delete it from VCD. If it matches, it gets ignored as it doesn't exist.
func testOpenApiMetadataEntryIgnore(t *testing.T, resourceTemplate, resourceAddress, datasourceTemplate, datasourceAddress string, retrieveObjectById func(*VCDClient, string) (openApiMetadataCompatible, error), extraParams StringMap) {
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

	// Steps 2 and 3 introduce conflicting configs (attempting to set metadata, that is specified in
	// `ignore_metadata_changes`). Binary tests cannot pass without the additional manipulation in
	// SDK that this test does
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	params["FuncName"] = t.Name() + "-Step2"
	params["Metadata"] = getOpenApiMetadataTestingHcl(1, 0, 0, 0, 0, 0, 0)
	step2 := templateFill(resourceTemplate, params)
	debugPrintf("#[DEBUG] CONFIGURATION 2: %s", step2)
	params["FuncName"] = t.Name() + "-Step3"
	step3 := templateFill(resourceTemplate+datasourceTemplate, params)
	debugPrintf("#[DEBUG] CONFIGURATION 3: %s", step3)

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
		var object openApiMetadataCompatible
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
						newEntry, err := object.AddMetadata(types.OpenApiMetadataEntry{
							IsPersistent: false,
							IsReadOnly:   false,
							KeyValue: types.OpenApiMetadataKeyValue{
								Domain: "TENANT",
								Key:    "foo",
								Value: types.OpenApiMetadataTypedValue{
									Value: "bar",
									Type:  types.OpenApiMetadataStringEntry,
								},
								Namespace: "",
							},
						})
						if err != nil {
							t.Fatalf("could not add metadata to object with ID '%s': %s", cachedId.fieldValue, err)
						}
						// Check that the metadata was added and not ignored
						_, err = object.GetMetadataById(newEntry.MetadataEntry.ID)
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
							if len(realMetadata) != expectedMetadataInVcd {
								return fmt.Errorf("expected %d metadata entries in VCD but got %d", expectedMetadataInVcd, len(realMetadata))
							}
							return nil
						},
						resource.TestCheckResourceAttr(resourceAddress, "metadata_entry.#", "1"),
						testCheckOpenApiMetadataEntrySetElemNestedAttrs(resourceAddress, "stringKey1", "stringValue1", types.OpenApiMetadataStringEntry, "TENANT", "", "false", "false"),
					),
				},
				// Test data source metadata.
				{
					Config: step3,
					Check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr(datasourceAddress, "metadata_entry.#", "1"),
						testCheckOpenApiMetadataEntrySetElemNestedAttrs(resourceAddress, "stringKey1", "stringValue1", types.OpenApiMetadataStringEntry, "TENANT", "", "false", "false"),
						testCheckOpenApiMetadataEntrySetElemNestedAttrs(datasourceAddress, "stringKey1", "stringValue1", types.OpenApiMetadataStringEntry, "TENANT", "", "false", "false"),
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
