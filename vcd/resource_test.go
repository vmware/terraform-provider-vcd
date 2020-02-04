// +build ALL

package vcd

import (
	"bytes"
	"html/template"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

var testNotFoundMap map[string]*testNotFoundDataSet

type testNotFoundDataSet struct {
	config     string
	deleteFunc func(t *testing.T, itemName string) func()
}

func init() {
	// validate if all resources have functions mapped
	testNotFoundMap = make(map[string]*testNotFoundDataSet)
	testNotFoundMap["vcd_catalog"] = &testNotFoundDataSet{
		deleteFunc: testDeleteExistingCatalog,
		config:     testAccCheckVcdCatalogBasic,
	}

	testNotFoundMap["vcd_catalog_item"] = &testNotFoundDataSet{
		deleteFunc: testDeleteExistingCatalogItem,
		config:     testAccCheckVcdCatalogItemBasic,
	}
}

// Validate that all resources have functions defined
// func init() {
// 	for definedResourceName := range globalResourceMap {
// 		functionSet, resourceExists := testNotFoundMap[definedResourceName]
// 		if !resourceExists {
// 			panic(fmt.Errorf("no function defined for resource '%s'", definedResourceName))
// 		}
//
// 		if functionSet == nil {
// 			panic(fmt.Errorf("empty function data set for resource '%s'", definedResourceName))
// 		}
// 	}
// }

func TestAccVcdResourceNotFound(t *testing.T) {
	// No point for running these tests when we don't have connection
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	// for providerResourceName := range globalResourceMap { // With validation
	for providerResourceName := range testNotFoundMap {
		functionSet, resourceExists := testNotFoundMap[providerResourceName]
		if !resourceExists {
			t.Errorf("resource '%s' does not have tests", providerResourceName)
		}

		// Run subtests
		t.Run(providerResourceName, oneResourceTestRunner(t, providerResourceName, functionSet))

	}

}

// TestAccVcdCatalogNotFound is meant to only check if plan\apply deleted outside of Terraform does not fail, but rather
// proposes to recreate the item.
func oneResourceTestRunner(t *testing.T, subTestName string, notFoundData *testNotFoundDataSet) func(t *testing.T) {
	return func(t *testing.T) {
		// No point for running test when we don't have connection
		if vcdShortTest {
			t.Skip(acceptanceTestsSkipped)
			return
		}

		catalogName := subTestName

		// name                 = "{{.CatalogItemName}}"
		// description          = "{{.Description}}"
		// ova_path             = "{{.OvaPath}}"
		// upload_piece_size    = {{.UploadPieceSize}}
		// show_upload_progress = "{{.UploadProgress}}"
		//
		// metadata = {
		// 	catalogItem_metadata = "catalogItem Metadata"
		// 	catalogItem_metadata2 = "catalogItem Metadata2"
		// }

		// var params = StringMap{
		// 	"Org":             testConfig.VCD.Org,
		// 	"Catalog":         testSuiteCatalogName,
		// 	"CatalogItemName": TestAccVcdCatalogItem,
		// 	"Description":     TestAccVcdCatalogItemDescription,
		// 	"OvaPath":         testConfig.Ova.OvaPath,
		// 	"UploadPieceSize": testConfig.Ova.UploadPieceSize,
		// 	"UploadProgress":  testConfig.Ova.UploadProgress,
		// 	"Tags":            "catalog",
		// }

		var params = StringMap{
			"Org":             testConfig.VCD.Org,
			"CatalogName":     catalogName,
			"CatalogItemName": subTestName,
			"Description":     "NotFoundDescription",
			"OvaPath":         testConfig.Ova.OvaPath,
			"UploadPieceSize": testConfig.Ova.UploadPieceSize,
			"UploadProgress":  testConfig.Ova.UploadProgress,
		}

		// configText := templateFill(notFoundData.config, params)

		var configTextBytes bytes.Buffer
		configTemplate := template.Must(template.New("letter").Parse(notFoundData.config))
		err := configTemplate.Execute(&configTextBytes, params)
		if err != nil {
			t.Errorf("could not generate template for '%s' resource", subTestName)
		}

		configText := configTextBytes.String()

		debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

		resourceAddress := func(vcdName string) string {
			var rgx = regexp.MustCompile(`resource\s+"` + catalogName + `"\s+"(\w*)"`)
			rs := rgx.FindStringSubmatch(configText)
			// fmt.Println(rs[1])

			return vcdName + "." + rs[1]
		}(subTestName)

		debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

		resource.Test(t, resource.TestCase{
			PreCheck:  func() { testAccPreCheck(t) },
			Providers: testAccProviders,
			Steps: []resource.TestStep{
				resource.TestStep{
					Config: configText,
					// Ensure the ID was set to make sure object was created
					Check: resource.TestCheckResourceAttrSet(resourceAddress, "id"),
				},
				resource.TestStep{
					Config:   configText,
					PlanOnly: true,
					// PreConfig: testDeleteExistingCatalog(t, catalogName),
					PreConfig: notFoundData.deleteFunc(t, catalogName),
					// It should offer to recreate during refresh when an object does not exist on apply
					ExpectNonEmptyPlan: true,
				},
			},
		})
	}
}
