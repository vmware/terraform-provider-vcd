package vcd

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/vmware/go-vcloud-director/v2/govcd"
)

// These elements are used to compose a filter block for data sources
var (

	// elementNameRegex should be available for most data sources.
	elementNameRegex = &schema.Schema{
		Type:        schema.TypeString,
		Optional:    true,
		Description: "Search by name with a regular expression",
	}

	// elementDate applies to those data sources that have a creation date
	elementDate = &schema.Schema{
		Type:        schema.TypeString,
		Optional:    true,
		Description: "Search by date comparison ({>|>=|<|<=|==} yyyy-mm-dd[ hh[:mm[:ss]]])",
	}

	// elementLatest applies to the same data sources where elementDate is applicable
	elementLatest = &schema.Schema{
		Type:        schema.TypeBool,
		Optional:    true,
		Description: "Retrieves the newest item",
	}

	// elementEarliest applies to the same data sources where elementDate is applicable
	elementEarliest = &schema.Schema{
		Type:        schema.TypeBool,
		Optional:    true,
		Description: "Retrieves the oldest item",
	}

	// elementIp applies to those data sources that expose an IP address
	elementIp = &schema.Schema{
		Type:        schema.TypeString,
		Optional:    true,
		Description: "Search by IP. The value can be a regular expression",
	}

	// elementMetadata applies to most data sources. It can be used even if the corresponding resource interface
	// does not handle metadata
	elementMetadata = &schema.Schema{
		Type:        schema.TypeList,
		Optional:    true,
		Description: "metadata filter",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"key": &schema.Schema{
					Type:        schema.TypeString,
					Required:    true,
					Description: "Metadata key (field name)",
				},
				"is_system": &schema.Schema{
					Type:        schema.TypeBool,
					Optional:    true,
					Description: "True if is a metadata@SYSTEM key",
				},
				"value": &schema.Schema{
					Type:        schema.TypeString,
					Required:    true,
					Description: `Metadata value (can be a regular expression if "use_api_search" is false)`,
				},
				"type": &schema.Schema{
					Type:         schema.TypeString,
					Optional:     true,
					Default:      "STRING",
					ValidateFunc: validation.StringInSlice(govcd.SupportedMetadataTypes, true),
					Description:  `Type of metadata value (needed only if "use_api_search" is true)`,
				},
				// API search means that metadata is used to filter items within the query.
				// The default behavior is to fetch all items, including the metadata info, and filter it
				// via regular expressions. The search by API is faster, although more strict: field types need
				// to be provided, searches only for exact matches.
				"use_api_search": &schema.Schema{
					Type:        schema.TypeBool,
					Optional:    true,
					Description: "If true, will search the vCD using native metadata query (without regular expressions)",
				},
			},
		},
	}
)

// To see a full filter block, look at datasource_vcd_catalog.go or datasource_vcd_catalog_item.go

// buildMetadataCriteria expands the values from a metadata block
// and returns a set of formal metadata filter definitions
func buildMetadataCriteria(metadataBlock interface{}) ([]govcd.MetadataDef, bool, error) {
	var definitions []govcd.MetadataDef
	var useApiSearch bool
	filterList, ok := metadataBlock.([]interface{})
	if !ok {
		return nil, useApiSearch, fmt.Errorf("metadata block is not a list")
	}
	for _, raw := range filterList {
		metadataMap, ok := raw.(map[string]interface{})
		if !ok {
			return nil, useApiSearch, fmt.Errorf("metadata internal block is not a map")
		}
		var def govcd.MetadataDef
		for key, value := range metadataMap {
			switch key {
			case "key":
				def.Key = value.(string)
			case "value":
				def.Value = value
			case "type":
				def.Type = value.(string)
			case "is_system":
				def.IsSystem = value.(bool)
			case "use_api_search":
				useApiSearch = value.(bool)
			}
		}
		definitions = append(definitions, def)
	}
	return definitions, useApiSearch, nil
}

// buildCriteria expands a filter block into a formal filter definition
func buildCriteria(filterBlock interface{}) (*govcd.FilterDef, error) {
	var criteria = govcd.NewFilterDef()

	filterList, ok := filterBlock.([]interface{})
	if !ok {
		return nil, fmt.Errorf("[buildCriteria] filter is not a list")
	}

	filterMap, ok := filterList[0].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("[buildCriteria] filter is not a map")
	}
	errorAddingFilter := "[buildCriteria] error adding filter '%s': %s"
	for key, value := range filterMap {
		switch key {

		case govcd.FilterNameRegex, govcd.FilterIp, govcd.FilterDate:
			err := criteria.AddFilter(key, value.(string))
			if err != nil {
				return nil, fmt.Errorf(errorAddingFilter, key, err)
			}
		case govcd.FilterLatest, govcd.FilterEarliest:
			strValue := fmt.Sprintf("%v", value.(bool))
			err := criteria.AddFilter(key, strValue)
			if err != nil {
				return nil, fmt.Errorf(errorAddingFilter, key, err)
			}
		case "metadata":
			definitions, useApiSearch, err := buildMetadataCriteria(value)
			if err != nil {
				return nil, fmt.Errorf(errorAddingFilter, key, err)
			}
			criteria.UseMetadataApiFilter = useApiSearch
			criteria.Metadata = definitions
		default:
			return nil, fmt.Errorf("unsupported filter key '%s'", key)
		}
	}
	return criteria, nil
}
