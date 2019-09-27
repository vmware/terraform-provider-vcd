package vcd

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func datasourceVcdCatalogMedia() *schema.Resource {
	return &schema.Resource{
		Read: resourceVcdMediaRead,

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
				Type:        schema.TypeString,
				Required:    true,
				Description: "media name",
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
				Type:        schema.TypeString,
				Computed:    true,
				Description: "True if this media file is ISO",
			},
			"owner_name": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Owner name",
			},
			"is_published": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
				Description: " 	True if this media file is in a published catalog",
			},
			"creation_date": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Creation date ",
			},
			"size": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Media storage in Bytes",
			},
			"status": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Media status ",
			},
			"storage_profile_name": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Storage profile name ",
			},
		},
	}
}
