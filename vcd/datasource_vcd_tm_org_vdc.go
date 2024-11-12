package vcd

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v3/govcd"
	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

func datasourceVcdTmVdc() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdTmVdcRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("Name of the %s", labelTmVdc),
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Description of the %s", labelTmVdc),
			},
			"is_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: fmt.Sprintf("Defines if the %s is enabled", labelTmVdc),
			},
			"org_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Parent Organization ID",
			},
			"region_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Parent Region ID",
			},
			"supervisor_ids": {
				Type:        schema.TypeSet,
				Computed:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: fmt.Sprintf("A set of Supervisor IDs that back this %s", labelTmVdc),
			},
			"zone_resource_allocations": {
				Type:        schema.TypeSet,
				Computed:    true,
				Elem:        tmVdcDsZoneResourceAllocation,
				Description: "A set of Supervisor Zones and their resource allocations",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("%s status", labelTmVdc),
			},
		},
	}
}

var tmVdcDsZoneResourceAllocation = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"zone_name": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Region Zone Name",
		},
		"zone_id": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Region Zone ID",
		},
		"memory_limit_mib": {
			Type:        schema.TypeInt,
			Computed:    true,
			Description: "Memory limit in MiB",
		},
		"memory_reservation_mib": {
			Type:        schema.TypeInt,
			Computed:    true,
			Description: "Memory reservation in MiB",
		},
		"cpu_limit_mhz": {
			Type:        schema.TypeInt,
			Computed:    true,
			Description: "CPU limit in MHz",
		},
		"cpu_reservation_mhz": {
			Type:        schema.TypeInt,
			Computed:    true,
			Description: "CPU reservation in MHz",
		},
	},
}

func datasourceVcdTmVdcRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	c := crudConfig[*govcd.TmVdc, types.TmVdc]{
		entityLabel:    labelTmVdc,
		getEntityFunc:  vcdClient.GetTmVdcByName,
		stateStoreFunc: setTmVdcData,
	}
	return readDatasource(ctx, d, meta, c)
}
