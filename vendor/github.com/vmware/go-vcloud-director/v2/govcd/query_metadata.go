/*
 * Copyright 2020 VMware, Inc.  All rights reserved.  Licensed under the Apache v2 License.
 */

package govcd

import (
	"fmt"
	"strings"
)

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
	var (
		vappTemplatefields = []string{"ownerName", "catalogName", "isPublished", "name", "vdc", "vdcName",
			"org", "creationDate", "isBusy", "isGoldMaster", "isEnabled", "status", "isDeployed", "isExpired",
			"storageProfileName"}
		edgeGatewayFields = []string{"name", "vdc", "numberOfExtNetworks", "numberOfOrgNetworks", "isBusy",
			"gatewayStatus", "haStatus"}
		orgVdcNetworkFields = []string{"name", "defaultGateway", "netmask", "dns1", "dns2", "dnsSuffix", "linkType",
			"connectedTo", "vdc", "isBusy", "isShared", "vdcName", "isIpScopeInherited"}
		adminCatalogFields = []string{"name", "isPublished", "isShared", "creationDate", "orgName", "ownerName",
			"numberOfVAppTemplates", "numberOfMedia", "owner", "publishSubscriptionType", "status"}
		fieldsOnDemand = map[string][]string{
			"vappTemplate":      vappTemplatefields,
			"adminVAppTemplate": vappTemplatefields,
			"edgeGateway":       edgeGatewayFields,
			"orgNetwork":        orgVdcNetworkFields,
			"adminOrgNetwork":   orgVdcNetworkFields,
			"adminCatalog":      adminCatalogFields,
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
// The metadata fields need to be requested explicitly (we can't just ask for generic metadata to be included
// in the query result. When we request metadata fields, we must also list the regular fields that we want
// in the result. For this reason, we neeed to have the list of fields supported by the query for each
// query type. Not all the fields can be used in the "fields" parameter of the query.
//
// * queryType is the type of the query. Only the ones listed within queryFieldsOnDemand are supported
// * params and notEncodedParams are the same ones passed to QueryWithNotEncodedParams
// * metadataFields is the list of fields to be included in the query results
// * if isSystem is true, metadata fields are requested as 'metadata@SYSTEM:fieldName'
func (catalog *Catalog) QueryWithMetadataFields(queryType string, params, notEncodedParams map[string]string,
	metadataFields []string, isSystem bool) (Results, error) {
	if notEncodedParams == nil {
		notEncodedParams = make(map[string]string)
	}
	notEncodedParams["type"] = queryType

	if len(metadataFields) == 0 {
		return catalog.client.QueryWithNotEncodedParams(params, notEncodedParams)
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

	return catalog.client.QueryWithNotEncodedParams(params, notEncodedParams)
}

// QueryWithMetadataFilter is a wrapper around QueryWithNotEncodedParams with additional filtering
// on metadata fields
// Unlike QueryWithMetadataFields, this function does not return the metadata fields, but only uses
// them to perform the filter.
// *
// * params and notEncodedParams are the same ones passed to QueryWithNotEncodedParams
// * metadataFilteris is a map of conditions to use for filtering
// * if isSystem is true, metadata fields are requested as 'metadata@SYSTEM:fieldName'
func (catalog *Catalog) QueryWithMetadataFilter(params, notEncodedParams map[string]string,
	metadataFilters map[string]MetadataFilter, isSystem bool) (Results, error) {

	if len(metadataFilters) == 0 {
		return Results{}, fmt.Errorf("[QueryWithMetadataFilter] no metadata fields provided")
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

	return catalog.client.QueryWithNotEncodedParams(params, notEncodedParams)
}
