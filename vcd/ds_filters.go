package vcd

import (
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
