package govcd

import (
	"fmt"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// QueryItem is an entity that is used to evaluate a Condition
type QueryItem interface {
	GetDate() string
	GetName() string
	GetType() string
	GetIp() string
	GetMetadataValue(key string) string
	GetParentName() string
	GetParentId() string
	GetHref() string
}

type (
	// All the Query* types are localizations of Query records that can be returned from a query.
	// Each one of these implements the QueryItem interface
	QueryVAppTemplate  types.QueryResultVappTemplateType
	QueryCatalogItem   types.QueryResultCatalogItemType
	QueryEdgeGateway   types.QueryResultEdgeGatewayRecordType
	QueryAdminCatalog  types.AdminCatalogRecord
	QueryCatalog       types.CatalogRecord
	QueryOrgVdcNetwork types.QueryResultOrgVdcNetworkRecordType
	QueryMedia         types.MediaRecordType
)

// getMetadataValue is a generic metadata lookup for all query items
func getMetadataValue(metadata *types.Metadata, key string) string {
	if metadata == nil || len(metadata.MetadataEntry) == 0 {
		return ""
	}
	for _, x := range metadata.MetadataEntry {
		if key == x.Key {
			return x.TypedValue.Value
		}
	}
	return ""
}

// --------------------------------------------------------------
// vApp template
// --------------------------------------------------------------
func (vappTemplate QueryVAppTemplate) GetHref() string       { return vappTemplate.HREF }
func (vappTemplate QueryVAppTemplate) GetName() string       { return vappTemplate.Name }
func (vappTemplate QueryVAppTemplate) GetType() string       { return "vapp_template" }
func (vappTemplate QueryVAppTemplate) GetIp() string         { return "" }
func (vappTemplate QueryVAppTemplate) GetDate() string       { return vappTemplate.CreationDate }
func (vappTemplate QueryVAppTemplate) GetParentName() string { return vappTemplate.CatalogName }
func (vappTemplate QueryVAppTemplate) GetParentId() string   { return vappTemplate.Vdc }
func (vappTemplate QueryVAppTemplate) GetMetadataValue(key string) string {
	return getMetadataValue(vappTemplate.Metadata, key)
}

// --------------------------------------------------------------
// media item
// --------------------------------------------------------------
func (media QueryMedia) GetHref() string       { return media.HREF }
func (media QueryMedia) GetName() string       { return media.Name }
func (media QueryMedia) GetType() string       { return "catalog_media" }
func (media QueryMedia) GetIp() string         { return "" }
func (media QueryMedia) GetDate() string       { return media.CreationDate }
func (media QueryMedia) GetParentName() string { return media.CatalogName }
func (media QueryMedia) GetParentId() string   { return media.Catalog }
func (media QueryMedia) GetMetadataValue(key string) string {
	return getMetadataValue(media.Metadata, key)
}

// --------------------------------------------------------------
// catalog item
// --------------------------------------------------------------
func (catItem QueryCatalogItem) GetHref() string       { return catItem.HREF }
func (catItem QueryCatalogItem) GetName() string       { return catItem.Name }
func (catItem QueryCatalogItem) GetIp() string         { return "" }
func (catItem QueryCatalogItem) GetType() string       { return "catalog_item" }
func (catItem QueryCatalogItem) GetDate() string       { return catItem.CreationDate }
func (catItem QueryCatalogItem) GetParentName() string { return catItem.CatalogName }
func (catItem QueryCatalogItem) GetParentId() string   { return catItem.Catalog }
func (catItem QueryCatalogItem) GetMetadataValue(key string) string {
	return getMetadataValue(catItem.Metadata, key)
}

// --------------------------------------------------------------
// catalog
// --------------------------------------------------------------
func (catalog QueryCatalog) GetHref() string       { return catalog.HREF }
func (catalog QueryCatalog) GetName() string       { return catalog.Name }
func (catalog QueryCatalog) GetIp() string         { return "" }
func (catalog QueryCatalog) GetType() string       { return "catalog" }
func (catalog QueryCatalog) GetDate() string       { return catalog.CreationDate }
func (catalog QueryCatalog) GetParentName() string { return catalog.OrgName }
func (catalog QueryCatalog) GetParentId() string   { return "" }
func (catalog QueryCatalog) GetMetadataValue(key string) string {
	return getMetadataValue(catalog.Metadata, key)
}

func (catalog QueryAdminCatalog) GetHref() string       { return catalog.HREF }
func (catalog QueryAdminCatalog) GetName() string       { return catalog.Name }
func (catalog QueryAdminCatalog) GetIp() string         { return "" }
func (catalog QueryAdminCatalog) GetType() string       { return "catalog" }
func (catalog QueryAdminCatalog) GetDate() string       { return catalog.CreationDate }
func (catalog QueryAdminCatalog) GetParentName() string { return catalog.OrgName }
func (catalog QueryAdminCatalog) GetParentId() string   { return "" }
func (catalog QueryAdminCatalog) GetMetadataValue(key string) string {
	return getMetadataValue(catalog.Metadata, key)
}

// --------------------------------------------------------------
// edge gateway
// --------------------------------------------------------------
func (egw QueryEdgeGateway) GetHref() string       { return egw.HREF }
func (egw QueryEdgeGateway) GetName() string       { return egw.Name }
func (egw QueryEdgeGateway) GetIp() string         { return "" }
func (egw QueryEdgeGateway) GetType() string       { return "edge_gateway" }
func (egw QueryEdgeGateway) GetDate() string       { return "" }
func (egw QueryEdgeGateway) GetParentName() string { return egw.OrgVdcName }
func (egw QueryEdgeGateway) GetParentId() string   { return egw.Vdc }
func (egw QueryEdgeGateway) GetMetadataValue(key string) string {
	// Edge Gateway doesn't support metadata
	return ""
}

// --------------------------------------------------------------
// Org VDC network
// --------------------------------------------------------------
func (network QueryOrgVdcNetwork) GetHref() string { return network.HREF }
func (network QueryOrgVdcNetwork) GetName() string { return network.Name }
func (network QueryOrgVdcNetwork) GetIp() string   { return network.DefaultGateway }
func (network QueryOrgVdcNetwork) GetType() string {
	switch network.LinkType {
	case 0:
		return "network_direct"
	case 1:
		return "network_routed"
	case 2:
		return "network_isolated"
	default:
		// There are only three types so far, but just to make it future proof
		return "network"
	}
}
func (network QueryOrgVdcNetwork) GetDate() string       { return "" }
func (network QueryOrgVdcNetwork) GetParentName() string { return network.VdcName }
func (network QueryOrgVdcNetwork) GetParentId() string   { return network.Vdc }
func (network QueryOrgVdcNetwork) GetMetadataValue(key string) string {
	return getMetadataValue(network.Metadata, key)
}

// --------------------------------------------------------------
// result conversion
// --------------------------------------------------------------
// resultToQueryItems converts a set of query results into a list of query items
func resultToQueryItems(queryType string, results Results) ([]QueryItem, error) {
	resultSize := int64(results.Results.Total)
	if resultSize < 1 {
		return nil, nil
	}
	var items = make([]QueryItem, resultSize)
	switch queryType {
	case types.QtAdminCatalogItem:
		for i, item := range results.Results.AdminCatalogItemRecord {
			items[i] = QueryCatalogItem(*item)
		}
	case types.QtCatalogItem:
		for i, item := range results.Results.CatalogItemRecord {
			items[i] = QueryCatalogItem(*item)
		}
	case types.QtMedia:
		for i, item := range results.Results.MediaRecord {
			items[i] = QueryMedia(*item)
		}
	case types.QtAdminMedia:
		for i, item := range results.Results.AdminMediaRecord {
			items[i] = QueryMedia(*item)
		}
	case types.QtVappTemplate:
		for i, item := range results.Results.VappTemplateRecord {
			items[i] = QueryVAppTemplate(*item)
		}
	case types.QtAdminVappTemplate:
		for i, item := range results.Results.AdminVappTemplateRecord {
			items[i] = QueryVAppTemplate(*item)
		}
	case types.QtEdgeGateway:
		for i, item := range results.Results.EdgeGatewayRecord {
			items[i] = QueryEdgeGateway(*item)
		}
	case types.QtOrgVdcNetwork:
		for i, item := range results.Results.OrgVdcNetworkRecord {
			items[i] = QueryOrgVdcNetwork(*item)
		}
	case types.QtCatalog:
		for i, item := range results.Results.CatalogRecord {
			items[i] = QueryCatalog(*item)
		}
	case types.QtAdminCatalog:
		for i, item := range results.Results.AdminCatalogRecord {
			items[i] = QueryAdminCatalog(*item)
		}
	}
	if len(items) > 0 {
		return items, nil
	}
	return nil, fmt.Errorf("unsupported query type %s", queryType)
}
