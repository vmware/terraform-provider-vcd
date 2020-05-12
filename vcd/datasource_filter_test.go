// +build search functional ALL

package vcd

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

type filterCollection struct {
	client        *govcd.VCDClient
	org           *govcd.AdminOrg
	vdc           *govcd.Vdc
	catalog       *govcd.Catalog
	vAppTemplates []govcd.FilterMatch
	networks      []govcd.FilterMatch
	mediaItems    []govcd.FilterMatch
	catalogs      []govcd.FilterMatch
	edgeGateways  []govcd.FilterMatch
}

// filtersByType is a cache of datasource information.
// It allows using the same entities several times without repeating the queries.
var filtersByType filterCollection

// elements to create data source HCL scripts
const (
	onlyOrg = `
  org     = "{{.Org}}"
`
	orgAndCatalog = `
  org     = "{{.Org}}"
  catalog = "{{.Catalog}}"
`

	orgAndVdc = `
  org = "{{.Org}}"
  vdc = "{{.VDC}}"
`
)

var ancestors = map[string]string{
	"vcd_catalog":          onlyOrg,
	"vcd_catalog_item":     orgAndCatalog,
	"vcd_catalog_media":    orgAndCatalog,
	"vcd_network_routed":   orgAndVdc,
	"vcd_network_direct":   orgAndVdc,
	"vcd_network_isolated": orgAndVdc,
}

// Data needed to create test vApp templates
var vAppTemplateBaseName = "catItemQuery"
var vAppTemplateRequestData = []govcd.VappTemplateData{
	{vAppTemplateBaseName + "1", "", "", govcd.StringMap{"one": "first", "two": "second"}, false},
	{vAppTemplateBaseName + "2", "", "", govcd.StringMap{"abc": "first", "def": "dummy"}, false},
	{vAppTemplateBaseName + "3", "", "", govcd.StringMap{"one": "first", "two": "second"}, false},
	{vAppTemplateBaseName + "4", "", "", govcd.StringMap{"abc": "first", "def": "second", "xyz": "final"}, false},
	{vAppTemplateBaseName + "5", "", "", govcd.StringMap{"ghj": "first", "klm": "second"}, false},
}

// getFiltersForAvailableEntities collects data from existing resources and creates filters for each of them
func getFiltersForAvailableEntities(entityTYpe string, dataGeneration bool) ([]govcd.FilterMatch, error) {

	var (
		err       error
		vcdClient *govcd.VCDClient
		org       *govcd.AdminOrg
		vdc       *govcd.Vdc
		catalog   *govcd.Catalog
	)

	if filtersByType.client != nil {
		vcdClient = filtersByType.client
	} else {
		vcdClient, err = getTestVCDFromJson(testConfig)
		if err != nil {
			return nil, fmt.Errorf("error getting client configuration: %s", err)
		}
		err = ProviderAuthenticate(vcdClient, testConfig.Provider.User, testConfig.Provider.Password, testConfig.Provider.Token, testConfig.Provider.SysOrg)
		if err != nil {
			return nil, fmt.Errorf("authentication error: %s", err)
		}
		filtersByType.client = vcdClient
	}
	if filtersByType.org != nil {
		org = filtersByType.org
	} else {
		org, err = vcdClient.GetAdminOrgByName(testConfig.VCD.Org)
		if err != nil {
			return nil, fmt.Errorf("org not found : %s", err)
		}
		filtersByType.org = org
	}
	if filtersByType.catalog != nil {
		catalog = filtersByType.catalog
	} else {
		catalog, err = org.GetCatalogByName(testConfig.VCD.Catalog.Name, false)
		if err != nil {
			return nil, fmt.Errorf("catalog not found : %s", err)
		}
		filtersByType.catalog = catalog
	}
	if filtersByType.vdc != nil {
		vdc = filtersByType.vdc
	} else {
		vdc, err = org.GetVDCByName(testConfig.VCD.Vdc, false)
		if err != nil {
			return nil, fmt.Errorf("vdc not found : %s", err)
		}
	}

	if dataGeneration {
		vAppTemplateRequestData, err = govcd.HelperCreateMultipleCatalogItems(catalog, vAppTemplateRequestData, os.Getenv("GOVCD_DEBUG") != "")
		if err != nil {
			return nil, fmt.Errorf("error generating data: %s", err)
		}
	}
	var results []govcd.FilterMatch
	switch entityTYpe {
	case types.QtAdminVappTemplate, types.QtVappTemplate:
		if filtersByType.vAppTemplates != nil {
			return filtersByType.vAppTemplates, nil
		}
		vappTemplateFilters, err := govcd.HelperMakeFiltersFromVappTemplate(catalog)
		if err != nil {
			return nil, fmt.Errorf("error collecting vApp templates for catalog %s: %s", catalog.Catalog.Name, err)
		}
		filtersByType.vAppTemplates = vappTemplateFilters
		results = vappTemplateFilters

	case types.QtEdgeGateway:
		if filtersByType.edgeGateways != nil {
			return filtersByType.edgeGateways, nil
		}
		egwFilters, err := govcd.HelperMakeFiltersFromEdgeGateways(vdc)
		if err != nil {
			return nil, fmt.Errorf("error collecting edge gateways for VDC %s: %s", vdc.Vdc.Name, err)
		}
		filtersByType.edgeGateways = egwFilters
		results = egwFilters
	case types.QtMedia, types.QtAdminMedia:
		if filtersByType.mediaItems != nil {
			return filtersByType.mediaItems, nil
		}
		mediaFilters, err := govcd.HelperMakeFiltersFromMedia(vdc, catalog.Catalog.Name)
		if err != nil {
			return nil, fmt.Errorf("error collecting media items for VDC %s: %s", vdc.Vdc.Name, err)
		}
		filtersByType.mediaItems = mediaFilters
		results = mediaFilters

	case types.QtCatalog, types.QtAdminCatalog:
		if filtersByType.catalogs != nil {
			return filtersByType.catalogs, nil
		}
		catalogFilters, err := govcd.HelperMakeFiltersFromCatalogs(org)
		if err != nil {
			return nil, fmt.Errorf("error collecting catalogs for org %s: %s", org.AdminOrg.Name, err)
		}
		filtersByType.catalogs = catalogFilters
		results = catalogFilters
	case types.QtOrgVdcNetwork:
		if filtersByType.networks != nil {
			return filtersByType.networks, nil
		}
		networkFilters, err := govcd.HelperMakeFiltersFromNetworks(vdc)
		if err != nil {
			return nil, fmt.Errorf("error collecting networks for VDC %s: %s", vdc.Vdc.Name, err)
		}
		filtersByType.networks = networkFilters
		results = networkFilters
	}

	return results, nil
}

// updateMatchEntity returns the appropriate resource type for each entity in the filter
func updateMatchEntity(match govcd.FilterMatch) govcd.FilterMatch {
	switch match.EntityType {
	case "QueryVAppTemplate", "QueryCatalogItem":
		match.EntityType = "vcd_catalog_item"
	case "QueryMedia":
		match.EntityType = "vcd_catalog_media"
	case "QueryCatalog":
		match.EntityType = "vcd_catalog"
	case "QueryEdgeGateway":
		match.EntityType = "vcd_edgegateway"
	case "QueryOrgVdcNetwork":
		network := match.Entity.(govcd.QueryOrgVdcNetwork)
		switch network.LinkType {
		case 0:
			match.EntityType = "vcd_network_direct"
		case 1:
			match.EntityType = "vcd_network_routed"
		case 2:
			match.EntityType = "vcd_network_isolated"
		}
	}
	return match
}

// generateTemplates creates a template HCL script from a set of filters
// In addition to the script text, returns a map of expected values, to be evaluated in a Terraform test
func generateTemplates(matches []govcd.FilterMatch) (string, map[string]string, error) {

	const (
		itemDelta       = 200           // base for number generation when using two metadata methods for the same data
		maxAllowedItems = itemDelta / 2 // maximum number of items that will be collected
	)
	var (
		expectedResults = make(map[string]string)
		templates       string
		err             error
		maxItems        = 5
	)

	// Limits the number of items to poll for test generation.
	// If many items exist, it could lead to an expensive test
	itemsNum := os.Getenv("VCD_MAX_ITEMS")
	if itemsNum != "" {
		maxItems, err = strconv.Atoi(itemsNum)
		if err != nil {
			maxItems = 5
		}
		if maxItems <= 0 {
			maxItems = 1 // at least one item will be evaluated
		}
	}
	if maxItems > maxAllowedItems {
		maxItems = maxAllowedItems
	}

	for i, match := range matches {
		if i > maxItems {
			break
		}
		match = updateMatchEntity(match)
		hasMetadata := false
		dsName := fmt.Sprintf("unknown%d", i)
		// creates the header of the data source
		entityText := fmt.Sprintf("data \"%s\" \"%s\"{\n", match.EntityType, dsName)
		entityText += fmt.Sprintf("  # expected name: '%s'\n", match.ExpectedName)
		entityText += ancestors[match.EntityType]
		// Adds regular filters
		filterText := "  filter {\n"
		for k, v := range match.Criteria.Filters {
			filterText += fmt.Sprintf("    %s = \"%s\"\n", k, strings.ReplaceAll(v, `\`, `\\`))
		}
		// If there are metadata elements, adds filters for them
		for _, m := range match.Criteria.Metadata {
			hasMetadata = true
			filterText += "    metadata {\n"
			filterText += fmt.Sprintf("      key            = \"%s\"\n", m.Key)
			filterText += fmt.Sprintf("      value          = \"%s\"\n", m.Value)
			filterText += fmt.Sprintf("      type           = \"%s\"\n", m.Type)
			if m.IsSystem {
				filterText += "    is_system = true\n"
			}
			filterText += "      use_api_search = false\n"
			filterText += "    }\n"
		}
		filterText += "  }\n"
		entityText += filterText
		entityText += "}\n\n"

		// For each data source, adds an output element, to simplify the test checks
		entityText += fmt.Sprintf("output \"%s\" {\n", dsName)
		entityText += fmt.Sprintf("  value = data.%s.%s.name\n", match.EntityType, dsName)
		entityText += "}\n"

		templates += entityText + "\n"
		expectedResults[dsName] = match.ExpectedName
		// If there are metadata elements, generates a second data source, where the search is
		// performed in the query
		if hasMetadata {
			newDsName := fmt.Sprintf("unknown%d", i+itemDelta)
			secondText := strings.ReplaceAll(entityText, dsName, newDsName)
			secondText = strings.ReplaceAll(secondText, "use_api_search = false", "use_api_search = true")
			templates += secondText + "\n"
			expectedResults[newDsName] = match.ExpectedName
		}
	}
	return templates, expectedResults, nil
}

// TestAccSearchEngine generates a script with many data sources for each entity that
// supports the filter engine.
// The test triggers a search of the existing entities. For each type:
// 1. It will generate filters based on the data in the entity
// 2. It will then generate the HCL script for the data source
// 3. The test will check that each data source matches the expected entity name
func TestAccSearchEngine(t *testing.T) {
	// This test requires access to the vCD before filling templates
	// Thus it won't run in the short test
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	// "networks" includes vcd_network_isolated, vcd_network_routed, and vcd_network_direct
	t.Run("networks", func(t *testing.T) { runSearchTest(types.QtOrgVdcNetwork, "networks", t) })
	// The data used for catalog item filtering belongs to the inner vApp template
	t.Run("catalog_items", func(t *testing.T) { runSearchTest(types.QtVappTemplate, "catalog_items", t) })
	t.Run("media", func(t *testing.T) { runSearchTest(types.QtMedia, "media", t) })
	t.Run("catalog", func(t *testing.T) { runSearchTest(types.QtCatalog, "catalog", t) })
	t.Run("edge_gateway", func(t *testing.T) { runSearchTest(types.QtEdgeGateway, "edge_gateway", t) })
}

// runSearchTest builds the test elements for the given entityType and run the test itself
func runSearchTest(entityType, label string, t *testing.T) {

	generateData := false

	if entityType == types.QtAdminVappTemplate || entityType == types.QtVappTemplate {
		if os.Getenv("VCD_TEST_DATA_GENERATION") != "" {
			generateData = true
		}
	}
	filters, err := getFiltersForAvailableEntities(entityType, generateData)
	if err != nil {
		t.Skip(fmt.Sprintf("error getting available %s : %s", label, err))
		return
	}

	if len(filters) == 0 {
		t.Skip("No " + label + " found - data source test skipped")
		return
	}

	var params = StringMap{
		"Org":      testConfig.VCD.Org,
		"VDC":      testConfig.VCD.Vdc,
		"Catalog":  testConfig.VCD.Catalog.Name,
		"FuncName": "search_" + label,
		"Tags":     "search",
	}
	template, expectedResults, err := generateTemplates(filters)
	if err != nil {
		t.Skip("Error generating " + label + " templates - data source test skipped")
		return
	}

	configText := templateFill(template, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check:  makeCheckFuncsFromMap(expectedResults),
			},
		},
	})
	if generateData {
		// Remove items
		for _, item := range vAppTemplateRequestData {
			// If the item was already found in the server (item.Created = false)
			// we skip the deletion.
			// We also skip deletion if the variable GOVCD_KEEP_TEST_OBJECTS is set
			if !item.Created || os.Getenv("GOVCD_KEEP_TEST_OBJECTS") != "" {
				continue
			}

			catalogItem, err := filtersByType.catalog.GetCatalogItemByName(item.Name, true)
			if err == nil {
				err = catalogItem.Delete()
			}
			if err != nil {
				t.Errorf("### error deleting catalog item %s : %s\n", catalogItem.CatalogItem.Name, err)
			}
		}
	}
}

// makeCheckFuncsFromMap generates a container TestCheckFunc using a map containing the names of the
// output elements with the corresponding expected values
func makeCheckFuncsFromMap(m map[string]string) resource.TestCheckFunc {
	var checkFuncs []resource.TestCheckFunc
	for k, v := range m {
		checkFuncs = append(checkFuncs, resource.TestCheckOutput(k, v))
	}
	return resource.ComposeTestCheckFunc(checkFuncs...)
}
