package vcd

import (
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func datasourceVcdCatalogMedia() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceVcdMediaRead,

		Schema: map[string]*schema.Schema{
			"org": {
				Type:     schema.TypeString,
				Optional: true,
				Description: "The name of organization to use, optional if defined at provider " +
					"level. Useful when connected as sysadmin working across different organizations",
			},
			"catalog": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "catalog name where upload the Media file",
			},
			"name": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"name", "filter"},
				Description:  "media name (Optional when 'filter' is used)",
			},
			"description": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"metadata": {
				Type:        schema.TypeMap,
				Computed:    true,
				Description: "Key and value pairs for catalog item metadata",
				// For now underlying go-vcloud-director repo only supports
				// a value of type String in this map.
			},
			"is_iso": &schema.Schema{
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "True if this media file is ISO",
			},
			"owner_name": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Owner name",
			},
			"is_published": &schema.Schema{
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "True if this media file is in a published catalog",
			},
			"creation_date": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Creation date",
			},
			"size": &schema.Schema{
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Media storage in Bytes",
			},
			"status": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Media status",
			},
			"storage_profile_name": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Storage profile name",
			},
			"filter": &schema.Schema{
				Type:        schema.TypeList,
				MaxItems:    1,
				MinItems:    1,
				Optional:    true,
				Description: "Criteria for retrieving a catalog media by various attributes",
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

func dataSourceVcdMediaRead(d *schema.ResourceData, meta interface{}) error {
	return genericVcdMediaRead(d, meta, "datasource")
}
