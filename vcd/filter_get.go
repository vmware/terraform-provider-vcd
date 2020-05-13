package vcd

import (
	"fmt"

	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// Functions in this file perform search by filter for all the data sources that support filters

// searchByFilterFunc represents a function that retrieves query items using criteria
type searchByFilterFunc = func(queryType string, criteria *govcd.FilterDef) ([]govcd.QueryItem, string, error)

// getEntityByFilter is a low level search function that builds criteria from a filter block
// and then runs the search using a searchByFilterFunc
// Returns a single item, and fails for more than one item
func getEntityByFilter(search searchByFilterFunc, queryType, label string, filter interface{}) (govcd.QueryItem, error) {
	criteria, err := buildCriteria(filter)
	if err != nil {
		return nil, err
	}
	queryItems, explanation, err := search(queryType, criteria)
	if err != nil {
		return nil, err
	}
	if len(queryItems) == 0 {
		return nil, fmt.Errorf("no %s found with given criteria (%s)", label, explanation)
	}
	if len(queryItems) > 1 {
		var itemNames = make([]string, len(queryItems))
		for i, item := range queryItems {
			itemNames[i] = item.GetName()
		}
		return nil, fmt.Errorf("more than one %s found by given criteria: %v", label, itemNames)
	}
	return queryItems[0], nil
}

// getCatalogByFilter finds a catalog using a filter block
func getCatalogByFilter(org *govcd.AdminOrg, filter interface{}, isSysAdmin bool) (*govcd.Catalog, error) {
	queryType := types.QtCatalog
	if isSysAdmin {
		queryType = types.QtAdminCatalog
	}

	var searchFunc = func(queryType string, criteria *govcd.FilterDef) ([]govcd.QueryItem, string, error) {
		return org.SearchByFilter(queryType, criteria)
	}
	queryItem, err := getEntityByFilter(searchFunc, queryType, "catalog", filter)
	if err != nil {
		return nil, err
	}

	catalog, err := org.GetCatalogByHref(queryItem.GetHref())
	if err != nil {
		return nil, fmt.Errorf("[getCatalogByFilter] error retrieving catalog %s: %s", queryItem.GetName(), err)
	}
	return catalog, nil
}

// getCatalogItemByFilter finds a catalog item using a filter block
// TODO: This function should be updated in the context of Issue #502
func getCatalogItemByFilter(catalog *govcd.Catalog, filter interface{}, isSysAdmin bool) (*govcd.CatalogItem, error) {
	queryType := types.QtVappTemplate
	if isSysAdmin {
		queryType = types.QtAdminVappTemplate
	}
	var searchFunc = func(queryType string, criteria *govcd.FilterDef) ([]govcd.QueryItem, string, error) {
		return catalog.SearchByFilter(queryType, "catalogName", criteria)
	}
	queryItem, err := getEntityByFilter(searchFunc, queryType, "vApp template", filter)
	if err != nil {
		return nil, err
	}

	catalogItem, err := catalog.GetCatalogItemByName(queryItem.GetName(), false)
	if err != nil {
		return nil, fmt.Errorf("[getCatalogItemByFilter] error retrieving catalog item %s: %s", queryItem.GetName(), err)
	}
	return catalogItem, nil
}

// getMediaByFilter finds a media item using a filter block
func getMediaByFilter(catalog *govcd.Catalog, filter interface{}, isSysAdmin bool) (*govcd.Media, error) {
	queryType := types.QtMedia
	if isSysAdmin {
		queryType = types.QtAdminMedia
	}
	var searchFunc = func(queryType string, criteria *govcd.FilterDef) ([]govcd.QueryItem, string, error) {
		return catalog.SearchByFilter(queryType, "catalog", criteria)
	}

	queryItem, err := getEntityByFilter(searchFunc, queryType, "media item", filter)
	if err != nil {
		return nil, err
	}

	mediaItem, err := catalog.GetMediaByHref(queryItem.GetHref())
	if err != nil {
		return nil, fmt.Errorf("[getMediaByFilter] error retrieving media item %s: %s", queryItem.GetName(), err)
	}
	return mediaItem, nil
}

// getNetworkByFilter finds a network using a filter block
func getNetworkByFilter(vdc *govcd.Vdc, filter interface{}, wanted string) (*govcd.OrgVDCNetwork, error) {
	queryType := types.QtOrgVdcNetwork
	var searchFunc = func(queryType string, criteria *govcd.FilterDef) ([]govcd.QueryItem, string, error) {
		items, explanation, err := vdc.SearchByFilter(queryType, "vdc", criteria)
		var newItems []govcd.QueryItem
		for _, item := range items {
			if item.GetType() == "network_"+wanted {
				newItems = append(newItems, item)
			}
		}
		// If no items were found, we need to bail out here. If we don't,
		// we will get the standard error message from getEntityByFilter, which may contain
		// references to networks of different type
		if len(newItems) == 0 {
			return nil, "", fmt.Errorf("no network_%s found", wanted)
		}
		return newItems, explanation, err
	}

	queryItem, err := getEntityByFilter(searchFunc, queryType, "network_"+wanted, filter)
	if err != nil {
		return nil, err
	}

	network, err := vdc.GetOrgVdcNetworkByHref(queryItem.GetHref())
	if err != nil {
		return nil, fmt.Errorf("[getNetworkByFilter] error retrieving network %s: %s", queryItem.GetName(), err)
	}
	return network, nil
}

// getEdgeGatewayByFilter finds an edge gateway using a filter block
func getEdgeGatewayByFilter(vdc *govcd.Vdc, filter interface{}) (*govcd.EdgeGateway, error) {
	queryType := types.QtEdgeGateway
	var searchFunc = func(queryType string, criteria *govcd.FilterDef) ([]govcd.QueryItem, string, error) {
		return vdc.SearchByFilter(queryType, "vdc", criteria)
	}

	queryItem, err := getEntityByFilter(searchFunc, queryType, "edge gateway", filter)
	if err != nil {
		return nil, err
	}

	egw, err := vdc.GetEdgeGatewayByHref(queryItem.GetHref())
	if err != nil {
		return nil, fmt.Errorf("[getEdgeGatewayByFilter] error retrieving edge gateway %s: %s", queryItem.GetName(), err)
	}
	return egw, nil
}
