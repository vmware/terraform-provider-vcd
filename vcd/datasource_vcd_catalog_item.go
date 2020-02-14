package vcd

import "github.com/hashicorp/terraform-plugin-sdk/helper/schema"

func datasourceVcdCatalogItem() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceVcdCatalogItemRead,
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
				Description: "catalog containing the item",
			},
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"description": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"metadata": {
				Type:        schema.TypeMap,
				Computed:    true,
				Description: "Key and value pairs for catalog item metadata",
			},
		},
	}
}

func dataSourceVcdCatalogItemRead(d *schema.ResourceData, meta interface{}) error {
	return genericVcdCatalogItemRead(d, meta, "datasource")
}
