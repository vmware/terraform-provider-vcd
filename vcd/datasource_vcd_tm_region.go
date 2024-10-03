package vcd

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdTmRegion() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdTmRegionRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Region name",
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Region description",
			},
			"is_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Defines whether the Region is enabled or not",
			},
			"nsx_manager_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "NSX Manager ID",
			},
			"cpu_capacity_mhz": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "CPU Capacity in MHz",
			},
			"cpu_reservation_capacity_mhz": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "CPU reservation in MHz",
			},
			"memory_capacity_mib": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Memory capacity in MiB",
			},
			"memory_reservation_capacity_mib": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Memory reservation in MiB",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Status of the region",
			},
			"supervisors": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "A set of supervisor IDs used in this Region",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"storage_policies": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "A set of storage policies",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func datasourceVcdTmRegionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	region, err := vcdClient.GetRegionByName(d.Get("name").(string))
	if err != nil {
		return diag.Errorf("error retrieving Region: %s", err)
	}

	err = setRegionData(d, region.Region)
	if err != nil {
		return diag.Errorf("error storing Region data: %s", err)
	}

	d.SetId(region.Region.ID)

	return nil
}
