package vcd

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v3/govcd"
	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

func datasourceVcdTmRegion() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdTmRegionRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("%s name", labelTmRegion),
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("%s description", labelTmRegion),
			},
			"is_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: fmt.Sprintf("Defines whether the %s is enabled or not", labelTmRegion),
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
				Description: fmt.Sprintf("Status of the %s", labelTmRegion),
			},
			"supervisor_ids": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: fmt.Sprintf("A set of supervisor IDs used in this %s", labelTmRegion),
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"storage_policy_names": {
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
	c := crudConfig[*govcd.Region, types.Region]{
		entityLabel:    labelTmRegion,
		getEntityFunc:  vcdClient.GetRegionByName,
		stateStoreFunc: setRegionData,
	}
	return readDatasource(ctx, d, meta, c)
}
