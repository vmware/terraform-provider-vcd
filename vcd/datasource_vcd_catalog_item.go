package vcd

import "github.com/hashicorp/terraform/helper/schema"

func datasourceVcdCatalogItem() *schema.Resource {
	return &schema.Resource{
		Read: resourceVcdCatalogItemRead,
		Schema: map[string]*schema.Schema{
			"org": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Org to which the catalog belongs",
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
