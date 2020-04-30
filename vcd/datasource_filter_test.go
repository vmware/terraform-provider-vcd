// +build search ALL

package vcd

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/vmware/go-vcloud-director/v2/govcd"
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
var vAppTemplateBaseName = "catItemQuery"
var vAppTemplateRequestData = []govcd.VappTemplateData{
	{vAppTemplateBaseName + "1", "", "", govcd.StringMap{"one": "first", "two": "second"}, false},
	{vAppTemplateBaseName + "2", "", "", govcd.StringMap{"abc": "first", "def": "dummy"}, false},
	{vAppTemplateBaseName + "3", "", "", govcd.StringMap{"one": "first", "two": "second"}, false},
	{vAppTemplateBaseName + "4", "", "", govcd.StringMap{"abc": "first", "def": "second", "xyz": "final"}, false},
	{vAppTemplateBaseName + "5", "", "", govcd.StringMap{"ghj": "first", "klm": "second"}, false},
}

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
	case govcd.QtAdminVappTemplate, govcd.QtVappTemplate:
		if filtersByType.vAppTemplates != nil {
			return filtersByType.vAppTemplates, nil
		}
		vappTemplateFilters, err := govcd.HelperMakeFiltersFromVappTemplate(catalog)
		if err != nil {
			return nil, fmt.Errorf("error collecting vApp templates for catalog %s: %s", catalog.Catalog.Name, err)
		}
		filtersByType.vAppTemplates = vappTemplateFilters
		results = vappTemplateFilters

	case govcd.QtEdgeGateway:
		if filtersByType.edgeGateways != nil {
			return filtersByType.edgeGateways, nil
		}
		egwFilters, err := govcd.HelperMakeFiltersFromEdgeGateways(vdc)
		if err != nil {
			return nil, fmt.Errorf("error collecting edge gateways for VDC %s: %s", vdc.Vdc.Name, err)
		}
		filtersByType.edgeGateways = egwFilters
		results = egwFilters
	case govcd.QtMedia, govcd.QtAdminMedia:
		if filtersByType.mediaItems != nil {
			return filtersByType.mediaItems, nil
		}
		mediaFilters, err := govcd.HelperMakeFiltersFromMedia(vdc, catalog.Catalog.Name)
		if err != nil {
			return nil, fmt.Errorf("error collecting media items for VDC %s: %s", vdc.Vdc.Name, err)
		}
		filtersByType.mediaItems = mediaFilters
		results = mediaFilters

	case govcd.QtCatalog, govcd.QtAdminCatalog:
		if filtersByType.catalogs != nil {
			return filtersByType.catalogs, nil
		}
		catalogFilters, err := govcd.HelperMakeFiltersFromCatalogs(org)
		if err != nil {
			return nil, fmt.Errorf("error collecting catalogs for org %s: %s", org.AdminOrg.Name, err)
		}
		filtersByType.catalogs = catalogFilters
		results = catalogFilters
	case govcd.QtOrgVdcNetwork:
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

func generateTemplates(matches []govcd.FilterMatch) (string, map[string]string, error) {

	var expectedResults = make(map[string]string)
	var templates string
	var err error
	maxItems := 5
	itemsNum := os.Getenv("VCD_MAX_ITEMS")
	if itemsNum != "" {
		maxItems, err = strconv.Atoi(itemsNum)
		if err != nil {
			maxItems = 5
		}
	}

	for i, match := range matches {
		if i > maxItems {
			break
		}
		match = updateMatchEntity(match)
		hasMetadata := false
		dsName := fmt.Sprintf("mystery%d", i)
		entityText := fmt.Sprintf("data \"%s\" \"%s\"{\n", match.EntityType, dsName)
		entityText += fmt.Sprintf("  # expected name: '%s'\n", match.ExpectedName)
		entityText += ancestors[match.EntityType]
		filterText := "  filter {\n"
		for k, v := range match.Criteria.Filters {
			filterText += fmt.Sprintf("    %s = \"%s\"\n", k, strings.ReplaceAll(v, `\`, `\\`))
		}
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

		entityText += fmt.Sprintf("output \"%s\" {\n", dsName)
		entityText += fmt.Sprintf("  value = data.%s.%s.name\n", match.EntityType, dsName)
		entityText += "}\n"

		templates += entityText + "\n"
		expectedResults[dsName] = match.ExpectedName
		if hasMetadata {
			newDsName := fmt.Sprintf("mystery%d", i+100)
			secondText := strings.ReplaceAll(entityText, dsName, newDsName)
			secondText = strings.ReplaceAll(secondText, "use_api_search = false", "use_api_search = true")
			templates += secondText + "\n"
			expectedResults[newDsName] = match.ExpectedName
		}
	}
	return templates, expectedResults, nil
}

func TestAccSearchEngine(t *testing.T) {
	// This test requires access to the vCD before filling templates
	// Thus it won't run in the short test
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	t.Run("networks", func(t *testing.T) { runSearchTest(govcd.QtOrgVdcNetwork, "networks", t) })
	t.Run("catalog_items", func(t *testing.T) { runSearchTest(govcd.QtVappTemplate, "catalog_items", t) })
	t.Run("catalog", func(t *testing.T) { runSearchTest(govcd.QtCatalog, "catalog", t) })
	t.Run("edge_gateway", func(t *testing.T) { runSearchTest(govcd.QtEdgeGateway, "edge_gateway", t) })
	t.Run("media", func(t *testing.T) { runSearchTest(govcd.QtMedia, "media", t) })
}

func runSearchTest(entityType, label string, t *testing.T) {

	generateData := false

	if entityType == govcd.QtAdminVappTemplate || entityType == govcd.QtVappTemplate {
		generateData = true
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

func makeCheckFuncsFromMap(m map[string]string) resource.TestCheckFunc {
	var checkFuncs []resource.TestCheckFunc
	for k, v := range m {
		checkFuncs = append(checkFuncs, resource.TestCheckOutput(k, v))
	}
	return resource.ComposeTestCheckFunc(checkFuncs...)
}
