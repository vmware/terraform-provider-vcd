package govcd

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// This file contains functions that help create tests for filtering.
// It is not in the '*_test.go' namespace because we want to use these functions from tests in other packages.
// All exported functions from this file have the prefix "Helper"
//
// Moreover, this file is not in a separate package for the following reasons:
//     * getExistingMedia is private
//     * getMetadata is private
//     * the 'client' component in all entity objects is private
//     * the tests that are now in filter_engine_test.go would need to go in a separate package, with consequent
//       need for configuration file parser duplication.

type StringMap map[string]string

type DateItem struct {
	Name       string
	Date       string
	Entity     interface{}
	EntityType string
}

// FilterMatch contains a filter, the name of the item that is expected to match, and the item itself
type FilterMatch struct {
	Criteria     *FilterDef
	ExpectedName string
	Entity       interface{}
	EntityType   string
}

type VappTemplateData struct {
	Name                     string
	ItemCreationDate         string
	VappTemplateCreationDate string
	Metadata                 StringMap
	Created                  bool
}

// retrievedMetadataTypes maps the internal value of metadata type with the
// string needed when searching for a metadata field in the API
var retrievedMetadataTypes = map[string]string{
	"MetadataBooleanValue":  "BOOLEAN",
	"MetadataStringValue":   "STRING",
	"MetadataNumberValue":   "NUMBER",
	"MetadataDateTimeValue": "STRING", // values for DATETIME can't be passed as such in a query when the date contains colons.
}

// HelperMakeFiltersFromEdgeGateways looks at the existing edge gateways and creates a set of criteria to retrieve each of them
func HelperMakeFiltersFromEdgeGateways(vdc *Vdc) ([]FilterMatch, error) {
	egwResult, err := vdc.GetEdgeGatewayRecordsType(false)
	if err != nil {
		return nil, err
	}

	if egwResult.EdgeGatewayRecord == nil || len(egwResult.EdgeGatewayRecord) == 0 {
		return []FilterMatch{}, nil
	}
	var filters = make([]FilterMatch, len(egwResult.EdgeGatewayRecord))
	for i, egw := range egwResult.EdgeGatewayRecord {

		filter := NewFilterDef()
		err = filter.AddFilter(types.FilterNameRegex, strToRegex(egw.Name))
		if err != nil {
			return nil, err
		}
		filters[i] = FilterMatch{filter, egw.Name, QueryEdgeGateway(*egw), "QueryEdgeGateway"}
	}
	return filters, nil
}

// HelperMakeFiltersFromNetworks looks at the existing networks and creates a set of criteria to retrieve each of them
func HelperMakeFiltersFromNetworks(vdc *Vdc) ([]FilterMatch, error) {
	netList, err := vdc.GetNetworkList()
	if err != nil {
		return nil, err
	}
	var filters = make([]FilterMatch, len(netList))
	for i, net := range netList {

		localizedItem := QueryOrgVdcNetwork(*net)
		qItem := QueryItem(localizedItem)
		filter, _, err := queryItemToFilter(qItem, "QueryOrgVdcNetwork")
		if err != nil {
			return nil, err
		}

		filter, err = vdc.client.metadataToFilter(net.HREF, filter)
		if err != nil {
			return nil, err
		}
		filters[i] = FilterMatch{filter, net.Name, localizedItem, "QueryOrgVdcNetwork"}
	}
	return filters, nil
}

// makeDateFilter creates date filters from a set of date records
// If there is more than one item, it creates an 'earliest' and 'latest' filter
func makeDateFilter(items []DateItem) ([]FilterMatch, error) {
	var filters []FilterMatch

	if len(items) == 0 {
		return filters, nil
	}
	entityType := items[0].EntityType
	if len(items) == 1 {
		filter := NewFilterDef()
		err := filter.AddFilter(types.FilterDate, "=="+items[0].Date)
		filters = append(filters, FilterMatch{filter, items[0].Name, items[0].Entity, entityType})
		return filters, err
	}
	earliestDate := time.Now().AddDate(100, 0, 0).String()
	latestDate := "1970-01-01 00:00:00"
	earliestName := ""
	latestName := ""
	var earliestEntity interface{}
	var latestEntity interface{}
	earliestFound := false
	latestFound := false
	for _, item := range items {
		greater, err := compareDate(">"+latestDate, item.Date)
		if err != nil {
			return nil, err
		}
		if greater {
			latestDate = item.Date
			latestName = item.Name
			latestEntity = item.Entity
			latestFound = true
		}
		greater, err = compareDate("<"+earliestDate, item.Date)
		if err != nil {
			return nil, err
		}
		if greater {
			earliestDate = item.Date
			earliestName = item.Name
			earliestEntity = item.Entity
			earliestFound = true
		}
		exactFilter := NewFilterDef()
		_ = exactFilter.AddFilter(types.FilterDate, "=="+item.Date)
		filters = append(filters, FilterMatch{exactFilter, item.Name, item.Entity, item.EntityType})
	}

	if earliestFound && latestFound && earliestDate != latestDate {
		earlyFilter := NewFilterDef()
		_ = earlyFilter.AddFilter(types.FilterDate, "<"+latestDate)
		_ = earlyFilter.AddFilter(types.FilterEarliest, "true")

		lateFilter := NewFilterDef()
		_ = lateFilter.AddFilter(types.FilterDate, ">"+earliestDate)
		_ = lateFilter.AddFilter(types.FilterLatest, "true")

		filters = append(filters, FilterMatch{earlyFilter, earliestName, earliestEntity, entityType})
		filters = append(filters, FilterMatch{lateFilter, latestName, latestEntity, entityType})
	}

	return filters, nil
}

func HelperMakeFiltersFromCatalogs(org *AdminOrg) ([]FilterMatch, error) {
	catalogs, err := org.QueryCatalogList()
	if err != nil {
		return []FilterMatch{}, err
	}

	var filters []FilterMatch

	var dateInfo []DateItem
	for _, cat := range catalogs {
		localizedItem := QueryCatalog(*cat)
		qItem := QueryItem(localizedItem)
		filter, dInfo, err := queryItemToFilter(qItem, "QueryCatalog")
		if err != nil {
			return nil, err
		}

		dateInfo = append(dateInfo, dInfo...)

		filter, err = org.client.metadataToFilter(cat.HREF, filter)
		if err != nil {
			return nil, err
		}

		filters = append(filters, FilterMatch{filter, cat.Name, localizedItem, "QueryCatalog"})
	}
	dateFilter, err := makeDateFilter(dateInfo)
	if err != nil {
		return []FilterMatch{}, err
	}
	if len(dateFilter) > 0 {
		filters = append(filters, dateFilter...)
	}
	return filters, nil
}

func HelperMakeFiltersFromMedia(vdc *Vdc, catalogName string) ([]FilterMatch, error) {
	var filters []FilterMatch
	items, err := getExistingMedia(vdc)
	if err != nil {
		return filters, err
	}
	var dateInfo []DateItem
	for _, item := range items {

		if item.CatalogName != catalogName {
			continue
		}
		localizedItem := QueryMedia(*item)
		qItem := QueryItem(localizedItem)
		filter, dInfo, err := queryItemToFilter(qItem, "QueryMedia")
		if err != nil {
			return nil, err
		}

		dateInfo = append(dateInfo, dInfo...)

		filter, err = vdc.client.metadataToFilter(item.HREF, filter)
		if err != nil {
			return nil, err
		}

		filters = append(filters, FilterMatch{filter, item.Name, localizedItem, "QueryMedia"})
	}
	dateFilter, err := makeDateFilter(dateInfo)
	if err != nil {
		return nil, err
	}
	if len(dateFilter) > 0 {
		filters = append(filters, dateFilter...)
	}
	return filters, nil
}

func queryItemToFilter(item QueryItem, entityType string) (*FilterDef, []DateItem, error) {

	var dateInfo []DateItem
	filter := NewFilterDef()
	err := filter.AddFilter(types.FilterNameRegex, strToRegex(item.GetName()))
	if err != nil {
		return nil, nil, err
	}

	if item.GetIp() != "" {
		err = filter.AddFilter(types.FilterIp, ipToRegex(item.GetIp()))
		if err != nil {
			return nil, nil, err
		}
	}
	if item.GetDate() != "" {
		dateInfo = append(dateInfo, DateItem{item.GetName(), item.GetDate(), item, entityType})
	}
	return filter, dateInfo, nil
}

func HelperMakeFiltersFromCatalogItem(catalog *Catalog) ([]FilterMatch, error) {
	var filters []FilterMatch
	items, err := catalog.QueryCatalogItemList()
	if err != nil {
		return filters, err
	}
	var dateInfo []DateItem
	for _, item := range items {

		localItem := QueryCatalogItem(*item)
		qItem := QueryItem(localItem)

		filter, dInfo, err := queryItemToFilter(qItem, "QueryCatalogItem")
		if err != nil {
			return nil, err
		}

		dateInfo = append(dateInfo, dInfo...)

		filter, err = catalog.client.metadataToFilter(item.HREF, filter)
		if err != nil {
			return nil, err
		}

		filters = append(filters, FilterMatch{filter, item.Name, localItem, "QueryCatalogItem"})
	}
	dateFilter, err := makeDateFilter(dateInfo)
	if err != nil {
		return nil, err
	}
	if len(dateFilter) > 0 {
		filters = append(filters, dateFilter...)
	}
	return filters, nil
}

func HelperMakeFiltersFromVappTemplate(catalog *Catalog) ([]FilterMatch, error) {
	var filters []FilterMatch
	items, err := catalog.QueryVappTemplateList()
	if err != nil {
		return filters, err
	}
	var dateInfo []DateItem
	for _, item := range items {

		localItem := QueryVAppTemplate(*item)
		qItem := QueryItem(localItem)

		filter, dInfo, err := queryItemToFilter(qItem, "QueryVAppTemplate")
		if err != nil {
			return nil, err
		}

		dateInfo = append(dateInfo, dInfo...)

		filter, err = catalog.client.metadataToFilter(item.HREF, filter)
		if err != nil {
			return nil, err
		}

		filters = append(filters, FilterMatch{filter, item.Name, localItem, "QueryVAppTemplate"})
	}
	dateFilter, err := makeDateFilter(dateInfo)
	if err != nil {
		return nil, err
	}
	if len(dateFilter) > 0 {
		filters = append(filters, dateFilter...)
	}
	return filters, nil
}

// HelperCreateMultipleCatalogItems deploys several catalog items, as defined in requestData
// Returns a set of VappTemplateData with what was created.
// If the requested objects exist already, returns updated information about the existing items.
func HelperCreateMultipleCatalogItems(catalog *Catalog, requestData []VappTemplateData, verbose bool) ([]VappTemplateData, error) {
	var data []VappTemplateData
	ova := "../test-resources/test_vapp_template.ova"
	_, err := os.Stat(ova)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("[HelperCreateMultipleCatalogItems] sample OVA %s not found", ova)
	}
	overallStart := time.Now()
	for _, requested := range requestData {
		name := requested.Name

		var item *CatalogItem
		var vappTemplate VAppTemplate
		created := false
		item, err := catalog.GetCatalogItemByName(name, false)
		if err == nil {
			// If the item already exists, we skip the creation, and just retrieve the vapp template
			vappTemplate, err = item.GetVAppTemplate()
			if err != nil {
				return nil, fmt.Errorf("[HelperCreateMultipleCatalogItems] error retrieving vApp template from catalog item %s : %s", item.CatalogItem.Name, err)
			}
		} else {

			start := time.Now()
			if verbose {
				fmt.Printf("%-55s %s ", start, name)
			}
			task, err := catalog.UploadOvf(ova, name, "test "+name, 10)
			if err != nil {
				return nil, fmt.Errorf("[HelperCreateMultipleCatalogItems] error uploading OVA: %s", err)
			}
			err = task.WaitTaskCompletion()
			if err != nil {
				return nil, fmt.Errorf("[HelperCreateMultipleCatalogItems] error completing task :%s", err)
			}
			item, err = catalog.GetCatalogItemByName(name, true)
			if err != nil {
				return nil, fmt.Errorf("[HelperCreateMultipleCatalogItems] error retrieving item %s: %s", name, err)
			}
			vappTemplate, err = item.GetVAppTemplate()
			if err != nil {
				return nil, fmt.Errorf("[HelperCreateMultipleCatalogItems] error retrieving vApp template: %s", err)
			}

			for k, v := range requested.Metadata {
				_, err := vappTemplate.AddMetadata(k, v)
				if err != nil {
					return nil, fmt.Errorf("[HelperCreateMultipleCatalogItems], error adding metadata: %s", err)
				}
			}
			duration := time.Since(start)
			if verbose {
				fmt.Printf("- elapsed: %s\n", duration)
			}

			created = true
		}
		data = append(data, VappTemplateData{
			Name:                     name,
			ItemCreationDate:         item.CatalogItem.DateCreated,
			VappTemplateCreationDate: vappTemplate.VAppTemplate.DateCreated,
			Metadata:                 requested.Metadata,
			Created:                  created,
		})
	}
	overallDuration := time.Since(overallStart)
	if verbose {
		fmt.Printf("total elapsed: %s\n", overallDuration)
	}

	return data, nil
}

// ipToRegex creates a regular expression that matches an IP without the last element
func ipToRegex(ip string) string {
	elements := strings.Split(ip, ".")
	result := "^"
	for i := 0; i < len(elements)-1; i++ {
		result += elements[i] + `\.`
	}
	return result
}

// strToRegex creates a regular expression that matches perfectly with the input query
func strToRegex(s string) string {
	var result strings.Builder
	result.WriteString("^")
	for _, ch := range s {
		if ch == '.' {
			result.WriteString(fmt.Sprintf("\\%c", ch))
		} else {
			result.WriteString(fmt.Sprintf("[%c]", ch))
		}
	}
	result.WriteString("$")
	return result.String()
}

// guessMetadataType guesses the type of a metadata value from its contents
// If the value looks like a number, or a true/false value, the corresponding type is returned
// Otherwise, we assume it's a string.
// We do this because the API doesn't return the metadata type
// (it would if the field TypedValue.XsiType were defined as `xml:"type,attr"`, but then metadata updates would fail.)
func guessMetadataType(value string) string {
	fType := "STRING"
	reNumber := regexp.MustCompile(`^[0-9]+$`)
	reBool := regexp.MustCompile(`^(?:true|false)$`)
	if reNumber.MatchString(value) {
		fType = "NUMBER"
	}
	if reBool.MatchString(value) {
		fType = "BOOLEAN"
	}
	return fType
}

// metadataToFilter adds metadata elements to an existing filter
// href is the address of the entity for which we want to retrieve metadata
// filter is an existing filter to which we want to add metadata elements
func (client *Client) metadataToFilter(href string, filter *FilterDef) (*FilterDef, error) {
	if filter == nil {
		filter = &FilterDef{}
	}
	metadata, err := getMetadata(client, href)
	if err == nil && metadata != nil && len(metadata.MetadataEntry) > 0 {
		for _, md := range metadata.MetadataEntry {
			isSystem := md.Domain == "SYSTEM"
			var fType string
			var ok bool
			if md.TypedValue.XsiType == "" {
				fType = guessMetadataType(md.TypedValue.Value)
			} else {
				fType, ok = retrievedMetadataTypes[md.TypedValue.XsiType]
				if !ok {
					fType = "STRING"
				}
			}
			err = filter.AddMetadataFilter(md.Key, md.TypedValue.Value, fType, isSystem, false)
			if err != nil {
				return nil, err
			}
		}
	}
	return filter, nil
}
