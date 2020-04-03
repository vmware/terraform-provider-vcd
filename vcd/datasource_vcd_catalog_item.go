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
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Name of the item. It is optional when a filter is provided",
			},
			"description": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"created": &schema.Schema{
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
				Description: "",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name_regex": &schema.Schema{
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Search by name with a regular expression",
						},
						"date": &schema.Schema{
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Search by date comparison ({>|>=|<|<=|==} yyyy-mm-dd[ hh[:mm[:ss]]])",
						},
						"latest": &schema.Schema{
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Retrieves the newest catalog item",
						},
						"metadata": &schema.Schema{
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

									//"key_type": &schema.Schema{
									//	Type:        schema.TypeString,
									//	Required:    true,
									//	Description: "Metadata key type (one of STRING, INT, BOOL)",
									//},
									"value": &schema.Schema{
										Type:        schema.TypeString,
										Required:    true,
										Description: "Metadata value (can be a regular expression)",
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func dataSourceVcdCatalogItemRead(d *schema.ResourceData, meta interface{}) error {
	return genericVcdCatalogItemRead(d, meta, "datasource")
}
