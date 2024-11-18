package vcd

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v3/govcd"
	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

func datasourceVcdTmOrgVdc() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdTmOrgVdcRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("Name of the %s", labelTmOrgVdc),
			},
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Parent Organization ID",
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Description of the %s", labelTmOrgVdc),
			},
			"is_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: fmt.Sprintf("Defines if the %s is enabled", labelTmOrgVdc),
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
				Description: fmt.Sprintf("A set of Supervisor IDs that back this %s", labelTmOrgVdc),
			},
			"zone_resource_allocations": {
				Type:        schema.TypeSet,
				Computed:    true,
				Elem:        tmOrgVdcDsZoneResourceAllocation,
				Description: "A set of Region Zones and their resource allocations",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("%s status", labelTmOrgVdc),
			},
		},
	}
}

var tmOrgVdcDsZoneResourceAllocation = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"region_zone_name": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: fmt.Sprintf("%s Name", labelTmRegionZone),
		},
		"region_zone_id": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: fmt.Sprintf("%s ID", labelTmRegionZone),
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

func datasourceVcdTmOrgVdcRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	getByNameAndOrgId := func(name string) (*govcd.TmVdc, error) {
		return vcdClient.GetTmVdcByNameAndOrgId(name, d.Get("org_id").(string))
	}

	c := dsCrudConfig[*govcd.TmVdc, types.TmVdc]{
		entityLabel:    labelTmOrgVdc,
		getEntityFunc:  getByNameAndOrgId,
		stateStoreFunc: setTmVdcData,
	}
	return readDatasource(ctx, d, meta, c)
}
