package vcd

import (
	"github.com/hashicorp/terraform/helper/schema"
)

func datasourceVcIndependentDisk() *schema.Resource {
	return &schema.Resource{
		Read: resourceVcdIndependentDiskRead,
		Schema: map[string]*schema.Schema{
			"org": {
				Type:     schema.TypeString,
				Optional: true,
				Description: "The name of organization to use, optional if defined at provider " +
					"level. Useful when connected as sysadmin working across different organizations",
			},
			"vdc": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The name of VDC to use, optional if defined at provider level",
			},
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"description": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "independent disk description",
			},
			"storage_profile": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"size_in_bytes": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "size in bytes",
			},
			"bus_type": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"bus_sub_type": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"iops": &schema.Schema{
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "IOPS request for the created disk",
			},
			"owner_name": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The owner name of the disk",
			},
			"datastore_name": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Datastore name",
			},
			"is_attached": &schema.Schema{
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "True if the disk is already attached",
			},
		},
	}
}
