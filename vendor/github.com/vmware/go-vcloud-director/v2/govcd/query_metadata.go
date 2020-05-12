/*
 * Copyright 2020 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

/*
This file contains functions that allow an extended query including metadata fields.

The metadata fields need to be requested explicitly (we can't just ask for generic metadata to be included
in the query result. Due to the query system implementation, when we request metadata fields, we must also
list the regular fields that we want in the result.
For this reason, we need to have the list of fields supported by the query for each query type. Not all the
fields can be used in the "fields" parameter of the query.

The function queryFieldsOnDemand provides the fields for the supported types.

Example: we have type "X" with fields "a", "b", "c", "d". It supports metadata.
If we want to query X without metadata, we run a simple query?type=X;[...]

If we also want metadata, we need to know which keys we want to fetch, and run
query?type=X;fields=a,b,c,d,metadata:key1,metadata:key2

*/

// MetadataFilter is a definition of a value used to filter metadata.
// It is made of a Type (such as 'STRING', 'INT', 'BOOL") and a Value, which is the value we want to search for.
type MetadataFilter struct {
	Type  string
	Value string
}

// queryFieldsOnDemand returns the list of fields that can be requested in the option "fields" of a query
// Note that an alternative approach using `reflect` would require several exceptions to list all the
// fields that are not supported.
func queryFieldsOnDemand(queryType string) ([]string, error) {
	// entities for which the fields on demand are supported
	var (
		vappTemplatefields = []string{"ownerName", "catalogName", "isPublished", "name", "vdc", "vdcName",
			"org", "creationDate", "isBusy", "isGoldMaster", "isEnabled", "status", "isDeployed", "isExpired",
			"storageProfileName"}
		edgeGatewayFields = []string{"name", "vdc", "orgVdcName", "numberOfExtNetworks", "numberOfOrgNetworks", "isBusy",
			"gatewayStatus", "haStatus"}
		orgVdcNetworkFields = []string{"name", "defaultGateway", "netmask", "dns1", "dns2", "dnsSuffix", "linkType",
			"connectedTo", "vdc", "isBusy", "isShared", "vdcName", "isIpScopeInherited"}
		catalogFields = []string{"name", "isPublished", "isShared", "creationDate", "orgName", "ownerName",
			"numberOfMedia", "owner"}
		mediaFields = []string{"ownerName", "catalogName", "isPublished", "name", "vdc", "vdcName", "org",
			"creationDate", "isBusy", "storageB", "owner", "catalog", "catalogItem", "status",
			"storageProfileName", "taskStatusName", "isInCatalog", "task",
			"isIso", "isVdcEnabled", "taskStatus", "taskDetails"}
		catalogItemFields = []string{"entity", "entityName", "entityType", "catalog", "catalogName", "ownerName",
			"owner", "isPublished", "vdc", "vdcName", "isVdcEnabled", "creationDate", "isExpired", "status"}
		fieldsOnDemand = map[string][]string{
			types.QtVappTemplate:      vappTemplatefields,
			types.QtAdminVappTemplate: vappTemplatefields,
			types.QtEdgeGateway:       edgeGatewayFields,
			types.QtOrgVdcNetwork:     orgVdcNetworkFields,
			types.QtCatalog:           catalogFields,
			types.QtAdminCatalog:      catalogFields,
			types.QtMedia:             mediaFields,
			types.QtAdminMedia:        mediaFields,
			types.QtCatalogItem:       catalogItemFields,
			types.QtAdminCatalogItem:  catalogItemFields,
		}
	)

	fields, ok := fieldsOnDemand[queryType]
	if !ok {
		return nil, fmt.Errorf("query type '%s' not supported", queryType)
	}
	return fields, nil
}

// addResults takes records from the appropriate field in the latest results and adds them to the cumulative results
func addResults(queryType string, cumulativeResults, newResults Results) (Results, int, error) {

	var size int
	switch queryType {
	case types.QtAdminVappTemplate:
		cumulativeResults.Results.AdminVappTemplateRecord = append(cumulativeResults.Results.AdminVappTemplateRecord, newResults.Results.AdminVappTemplateRecord...)
		size = len(newResults.Results.AdminVappTemplateRecord)
	case types.QtVappTemplate:
		size = len(newResults.Results.VappTemplateRecord)
		cumulativeResults.Results.VappTemplateRecord = append(cumulativeResults.Results.VappTemplateRecord, newResults.Results.VappTemplateRecord...)
	case types.QtCatalogItem:
		cumulativeResults.Results.CatalogItemRecord = append(cumulativeResults.Results.CatalogItemRecord, newResults.Results.CatalogItemRecord...)
		size = len(newResults.Results.CatalogItemRecord)
	case types.QtAdminCatalogItem:
		cumulativeResults.Results.AdminCatalogItemRecord = append(cumulativeResults.Results.AdminCatalogItemRecord, newResults.Results.AdminCatalogItemRecord...)
		size = len(newResults.Results.AdminCatalogItemRecord)
	case types.QtMedia:
		cumulativeResults.Results.MediaRecord = append(cumulativeResults.Results.MediaRecord, newResults.Results.MediaRecord...)
		size = len(newResults.Results.MediaRecord)
	case types.QtAdminMedia:
		cumulativeResults.Results.AdminMediaRecord = append(cumulativeResults.Results.AdminMediaRecord, newResults.Results.AdminMediaRecord...)
		size = len(newResults.Results.AdminMediaRecord)
	case types.QtCatalog:
		cumulativeResults.Results.CatalogRecord = append(cumulativeResults.Results.CatalogRecord, newResults.Results.CatalogRecord...)
		size = len(newResults.Results.CatalogRecord)
	case types.QtAdminCatalog:
		cumulativeResults.Results.AdminCatalogRecord = append(cumulativeResults.Results.AdminCatalogRecord, newResults.Results.AdminCatalogRecord...)
		size = len(newResults.Results.AdminCatalogRecord)
	case types.QtOrgVdcNetwork:
		cumulativeResults.Results.OrgVdcNetworkRecord = append(cumulativeResults.Results.OrgVdcNetworkRecord, newResults.Results.OrgVdcNetworkRecord...)
		size = len(newResults.Results.OrgVdcNetworkRecord)
	case types.QtEdgeGateway:
		cumulativeResults.Results.EdgeGatewayRecord = append(cumulativeResults.Results.EdgeGatewayRecord, newResults.Results.EdgeGatewayRecord...)
		size = len(newResults.Results.EdgeGatewayRecord)

	default:
		return Results{}, 0, fmt.Errorf("query type %s not supported", queryType)
	}

	return cumulativeResults, size, nil
}

// cumulativeQuery runs a paginated query and collects all elements until the total number of records is retrieved
func (client *Client) cumulativeQuery(queryType string, params, notEncodedParams map[string]string) (Results, error) {

	result, err := client.QueryWithNotEncodedParams(params, notEncodedParams)
	if err != nil {
		return Results{}, err
	}
	wanted := int(result.Results.Total)
	retrieved := int(wanted)
	if retrieved > result.Results.PageSize {
		retrieved = result.Results.PageSize
	}
	if retrieved == wanted {
		return result, nil
	}
	page := result.Results.Page

	var cumulativeResult = Results{
		Results: result.Results,
		client:  nil,
	}

	for retrieved != wanted {
		page++
		notEncodedParams["page"] = fmt.Sprintf("%d", page)
		var size int
		newResult, err := client.QueryWithNotEncodedParams(params, notEncodedParams)
		if err != nil {
			return Results{}, err
		}
		cumulativeResult, size, err = addResults(queryType, cumulativeResult, newResult)
		if err != nil {
			return Results{}, err
		}
		retrieved += size
	}

	return result, nil
}

// queryWithMetadataFields is a wrapper around QueryWithNotEncodedParams with additional metadata fields
// being returned.
//
// * queryType is the type of the query. Only the ones listed within queryFieldsOnDemand are supported
// * params and notEncodedParams are the same ones passed to QueryWithNotEncodedParams
// * metadataFields is the list of fields to be included in the query results
// * if isSystem is true, metadata fields are requested as 'metadata@SYSTEM:fieldName'
func (client *Client) queryWithMetadataFields(queryType string, params, notEncodedParams map[string]string,
	metadataFields []string, isSystem bool) (Results, error) {
	if notEncodedParams == nil {
		notEncodedParams = make(map[string]string)
	}
	notEncodedParams["type"] = queryType

	if len(metadataFields) == 0 {
		return client.cumulativeQuery(queryType, params, notEncodedParams)
	}

	fields, err := queryFieldsOnDemand(queryType)
	if err != nil {
		return Results{}, fmt.Errorf("[queryWithMetadataFields] %s", err)
	}

	if len(fields) == 0 {
		return Results{}, fmt.Errorf("[queryWithMetadataFields] no fields found for type '%s'", queryType)
	}
	metadataFieldText := ""
	prefix := "metadata"
	if isSystem {
		prefix = "metadata@SYSTEM"
	}
	for i, field := range metadataFields {
		metadataFieldText += fmt.Sprintf("%s:%s", prefix, field)
		if i != len(metadataFields) {
			metadataFieldText += ","
		}
	}

	notEncodedParams["fields"] = strings.Join(fields, ",") + "," + metadataFieldText

	return client.cumulativeQuery(queryType, params, notEncodedParams)
}

// queryByMetadataFilter is a wrapper around QueryWithNotEncodedParams with additional filtering
// on metadata fields
// Unlike queryWithMetadataFields, this function does not return the metadata fields, but only uses
// them to perform the filter.
//
// * params and notEncodedParams are the same ones passed to QueryWithNotEncodedParams
// * metadataFilter is is a map of conditions to use for filtering
// * if isSystem is true, metadata fields are requested as 'metadata@SYSTEM:fieldName'
func (client *Client) queryByMetadataFilter(queryType string, params, notEncodedParams map[string]string,
	metadataFilters map[string]MetadataFilter, isSystem bool) (Results, error) {

	if len(metadataFilters) == 0 {
		return Results{}, fmt.Errorf("[queryByMetadataFilter] no metadata fields provided")
	}
	if notEncodedParams == nil {
		notEncodedParams = make(map[string]string)
	}
	notEncodedParams["type"] = queryType

	metadataFilterText := ""
	prefix := "metadata"
	if isSystem {
		prefix = "metadata@SYSTEM"
	}
	count := 0
	for key, value := range metadataFilters {
		metadataFilterText += fmt.Sprintf("%s:%s==%s:%s", prefix, key, value.Type, url.QueryEscape(value.Value))
		if count < len(metadataFilters)-1 {
			metadataFilterText += ";"
		}
		count++
	}

	filter, ok := notEncodedParams["filter"]
	if ok {
		filter = "(" + filter + ";" + metadataFilterText + ")"
	} else {
		filter = metadataFilterText
	}
	notEncodedParams["filter"] = filter

	return client.cumulativeQuery(queryType, params, notEncodedParams)
}
