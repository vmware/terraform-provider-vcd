package vcd

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdCatalogItem() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceVcdCatalogItemRead,
		Schema: map[string]*schema.Schema{
			"org": {
				Type:     schema.TypeString,
				Optional: true,
				Description: "The name of organization to use, optional if defined at provider " +
					"level. Useful when connected as sysadmin working across different organizations",
			},
			"catalog": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "catalog containing the item",
			},
			"name": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "Name of the item. It is optional when a filter is provided",
				ExactlyOneOf: []string{"name", "filter"},
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Time stamp of when the item was created",
			},
			"metadata": {
				Type:        schema.TypeMap,
				Computed:    true,
				Description: "Key and value pairs for catalog item metadata",
			},
			"filter": {
				Type:        schema.TypeList,
				MaxItems:    1,
				MinItems:    1,
				Optional:    true,
				Description: "Criteria for retrieving a catalog item by various attributes",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name_regex": elementNameRegex,
						"date":       elementDate,
						"earliest":   elementEarliest,
						"latest":     elementLatest,
						"metadata":   elementMetadata,
					},
				},
			},
		},
	}
}

func dataSourceVcdCatalogItemRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericVcdCatalogItemRead(d, meta, "datasource")
}
