package vcd

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

func datasourceVcdTmVdc() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdTmVdcRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the VDC",
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Description of the VDC",
			},
			"is_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Defines if the VDC is enabled",
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
				Description: "A set of Supervisor IDs that back this VDC",
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
				Description: "VDC status",
			},
		},
	}
}

var tmVdcDsZoneResourceAllocation = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"zone_name": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Supervisor Zone Name",
		},
		"zone_id": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Supervisor Zone ID",
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

	v, err := vcdClient.GetTmVdcByName(d.Get("name").(string))
	if err != nil {
		return diag.Errorf("error getting VDC: %s", err)
	}

	err = setTmVdcData(d, v.TmVdc)
	if err != nil {
		return diag.Errorf("error storing Org data: %s", err)
	}

	d.SetId(v.TmVdc.ID)

	return nil
}

func setTmVdcData(d *schema.ResourceData, vdc *types.TmVdc) error {
	dSet(d, "name", vdc.Name)
	dSet(d, "description", vdc.Description)
	dSet(d, "is_enabled", vdc.IsEnabled)
	dSet(d, "status", vdc.Status)

	orgId := ""
	if vdc.Org != nil {
		orgId = vdc.Org.ID
	}
	dSet(d, "org_id", orgId)

	regionId := ""
	if vdc.Region != nil {
		regionId = vdc.Region.ID
	}
	dSet(d, "region_id", regionId)

	supervisors := extractIdsFromOpenApiReferences(vdc.Supervisors)
	err := d.Set("supervisor_ids", supervisors)
	if err != nil {
		return fmt.Errorf("error storing 'supervisor_ids': %s", err)
	}

	zoneCompute := make([]interface{}, 1)
	for _, zone := range vdc.ZoneResourceAllocation {
		oneZone := make(map[string]interface{})

		oneZone["zone_name"] = zone.Zone.Name
		oneZone["zone_id"] = zone.Zone.ID

		oneZone["memory_limit_mib"] = zone.ResourceAllocation.MemoryLimitMiB
		oneZone["memory_reservation_mib"] = zone.ResourceAllocation.MemoryReservationMiB
		oneZone["cpu_limit_mhz"] = zone.ResourceAllocation.CPULimitMHz
		oneZone["cpu_reservation_mhz"] = zone.ResourceAllocation.CPUReservationMHz

		zoneCompute = append(zoneCompute, oneZone)
	}

	autoAllocatedSubnetSet := schema.NewSet(schema.HashResource(tmVdcDsZoneResourceAllocation), zoneCompute)
	err = d.Set("zone_resource_allocations", autoAllocatedSubnetSet)
	if err != nil {
		return fmt.Errorf("error setting 'zone_resource_allocations' after read: %s", err)
	}

	return nil
}
