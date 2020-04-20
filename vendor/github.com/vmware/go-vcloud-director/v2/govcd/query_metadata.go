/*
 * Copyright 2020 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"strings"
)

/*
This file contains functions that allows an extended query including metadata fields.

The metadata fields need to be requested explicitly (we can't just ask for generic metadata to be included
in the query result. Due to the query system implementation, when we request metadata fields, we must also
list the regular fields that we want in the result.
For this reason, we need to have the list of fields supported by the query for each query type. Not all the
fields can be used in the "fields" parameter of the query.

The function queryFieldsOnDemand provides the fields for the supported types.

*/

// MetadataFilter is a definition of a value used to filter metadata.
// It is made of a Type (such as 'STRING', 'INT', 'BOOL") and a Value, which is the value we want to search for.
type MetadataFilter struct {
	Type  string
	Value string
}

const (
	// The QT* constants are the names used with Query requests to retrieve the corresponding entities
	QtVappTemplate      = "vAppTemplate"      // vApp template
	QtAdminVappTemplate = "adminVAppTemplate" // vApp template as admin
	QtEdgeGateway       = "edgeGateway"       // edge gateway
	QtOrgVdcNetwork     = "orgVdcNetwork"     // Org VDC network
	QtCatalog           = "catalog"           // catalog
	QtAdminCatalog      = "adminCatalog"      // catalog as admin
	QtCatalogItem       = "catalogItem"       // catalog item
	QtAdminCatalogItem  = "adminCatalogItem"  // catalog item as admin
	QtAdminMedia        = "adminMedia"        // media item as admin
	QtMedia             = "media"             // media item
)

// queryFieldsOnDemand returns the list of fields that can be requested in the option "fields" of a query
// Note that an alternative approach using `reflect` would require several exceptions to list all the
// fields that are not supported.
func queryFieldsOnDemand(queryType string) ([]string, error) {
	var (
		vappTemplatefields = []string{"ownerName", "catalogName", "isPublished", "name", "vdc", "vdcName",
			"org", "creationDate", "isBusy", "isGoldMaster", "isEnabled", "status", "isDeployed", "isExpired",
			"storageProfileName"}
		edgeGatewayFields = []string{"name", "vdc", "numberOfExtNetworks", "numberOfOrgNetworks", "isBusy",
			"gatewayStatus", "haStatus"}
		orgVdcNetworkFields = []string{"name", "defaultGateway", "netmask", "dns1", "dns2", "dnsSuffix", "linkType",
			"connectedTo", "vdc", "isBusy", "isShared", "vdcName", "isIpScopeInherited"}
		catalogFields = []string{"name", "isPublished", "isShared", "creationDate", "orgName", "ownerName",
			"numberOfMedia", "owner"}
		mediaFields = []string{"ownerName", "catalogName", "isPublished", "name", "vdc", "vdcName", "org",
			"creationDate", "isBusy", "storageB", "owner", "catalog", "catalogItem", "status",
			"storageProfileName", "taskStatusName", "isInCatalog", "task",
			"isIso", "isVdcEnabled", "taskStatus", "taskDetails"}
		// entities for which the fields on demand are supported
		catalogItemFields = []string{"entity", "entityName", "entityType", "catalog", "catalogName", "ownerName",
			"owner", "isPublished", "vdc", "vdcName", "isVdcEnabled", "creationDate", "isExpired", "status"}
		fieldsOnDemand = map[string][]string{
			QtVappTemplate:      vappTemplatefields,
			QtAdminVappTemplate: vappTemplatefields,
			QtEdgeGateway:       edgeGatewayFields,
			QtOrgVdcNetwork:     orgVdcNetworkFields,
			QtCatalog:           catalogFields,
			QtAdminCatalog:      catalogFields,
			QtMedia:             mediaFields,
			QtAdminMedia:        mediaFields,
			QtCatalogItem:       catalogItemFields,
			QtAdminCatalogItem:  catalogItemFields,
		}
	)

	fields, ok := fieldsOnDemand[queryType]
	if !ok {
		return nil, fmt.Errorf("query type '%s' not supported", queryType)
	}
	return fields, nil
}

// QueryWithMetadataFields is a wrapper around QueryWithNotEncodedParams with additional metadata fields
// being returned.
//
// * queryType is the type of the query. Only the ones listed within queryFieldsOnDemand are supported
// * params and notEncodedParams are the same ones passed to QueryWithNotEncodedParams
// * metadataFields is the list of fields to be included in the query results
// * if isSystem is true, metadata fields are requested as 'metadata@SYSTEM:fieldName'
func (client *Client) QueryWithMetadataFields(queryType string, params, notEncodedParams map[string]string,
	metadataFields []string, isSystem bool) (Results, error) {
	if notEncodedParams == nil {
		notEncodedParams = make(map[string]string)
	}
	notEncodedParams["type"] = queryType
	notEncodedParams["pageSize"] = "128" // Temporary workaround. TODO: loop until rows fetched == total

	if len(metadataFields) == 0 {
		return client.QueryWithNotEncodedParams(params, notEncodedParams)
	}

	fields, err := queryFieldsOnDemand(queryType)
	if err != nil {
		return Results{}, fmt.Errorf("[QueryWithMetadataFields] %s", err)
	}

	if len(fields) == 0 {
		return Results{}, fmt.Errorf("[QueryWithMetadataFields] no fields found for type '%s'", queryType)
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

	return client.QueryWithNotEncodedParams(params, notEncodedParams)
}

// QueryByMetadataFilter is a wrapper around QueryWithNotEncodedParams with additional filtering
// on metadata fields
// Unlike QueryWithMetadataFields, this function does not return the metadata fields, but only uses
// them to perform the filter.
//
// * params and notEncodedParams are the same ones passed to QueryWithNotEncodedParams
// * metadataFilter is is a map of conditions to use for filtering
// * if isSystem is true, metadata fields are requested as 'metadata@SYSTEM:fieldName'
func (client *Client) QueryByMetadataFilter(params, notEncodedParams map[string]string,
	metadataFilters map[string]MetadataFilter, isSystem bool) (Results, error) {

	if len(metadataFilters) == 0 {
		return Results{}, fmt.Errorf("[QueryByMetadataFilter] no metadata fields provided")
	}

	metadataFilterText := ""
	prefix := "metadata"
	if isSystem {
		prefix = "metadata@SYSTEM"
	}
	count := 0
	for key, value := range metadataFilters {
		metadataFilterText += fmt.Sprintf("%s:%s==%s:%s", prefix, key, value.Type, value.Value)
		if count < len(metadataFilters)-1 {
			metadataFilterText += ","
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
	notEncodedParams["pageSize"] = "128" // Temporary workaround. TODO: loop until rows fetched == total

	return client.QueryWithNotEncodedParams(params, notEncodedParams)
}
