package vcd

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func datasourceVcdTmVdc() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdTmVdcRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"is_enabled": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"org_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"region_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"supervisor_ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"zone_resource_allocations": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     tmVdcDsZoneResourceAllocation,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

var tmVdcDsZoneResourceAllocation = &schema.Resource{
	Schema: map[string]*schema.Schema{
		"zone_name": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Gateway address for a subnet",
		},
		"zone_id": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Prefix length for a subnet (e.g. 24)",
		},
		"memory_limit_mib": {
			Type:        schema.TypeInt,
			Computed:    true,
			Description: "IP address on the edge gateway",
		},
		"memory_reservation_mib": {
			Type:        schema.TypeInt,
			Computed:    true,
			Description: "IP address on the edge gateway",
		},
		"cpu_limit_mhz": {
			Type:        schema.TypeInt,
			Computed:    true,
			Description: "IP address on the edge gateway",
		},
		"cpu_reservation_mhz": {
			Type:        schema.TypeInt,
			Computed:    true,
			Description: "IP address on the edge gateway",
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
