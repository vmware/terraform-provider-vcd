package vcd

import "github.com/hashicorp/terraform/helper/schema"

func datasourceVcdVApp() *schema.Resource {
	return &schema.Resource{
		Read: datasourceVcdVAppRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "A unique name for the vApp",
			},
			"org": {
				Type:     schema.TypeString,
				Required: false,
				Optional: true,
				ForceNew: true,
				Description: "The name of organization to use, optional if defined at provider " +
					"level. Useful when connected as sysadmin working across different organizations",
			},
			"vdc": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "The name of VDC to use, optional if defined at provider level",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Optional description of the vApp",
			},
			"metadata": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "Key value map of metadata to assign to this vApp. Key and value can be any string.",
			},
			"href": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"power_on": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "A boolean value stating if this vApp should be powered on",
			},
			"guest_properties": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "Key/value settings for guest properties",
			},
			"status": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Shows the status code of the vApp",
			},
			"status_text": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Shows the status of the vApp",
			},
		},
	}
}

func datasourceVcdVAppRead(d *schema.ResourceData, meta interface{}) error {
	return genericVcdVAppRead(d, meta, "datasource")
}
