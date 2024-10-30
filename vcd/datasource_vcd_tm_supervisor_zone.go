package vcd

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v3/govcd"
)

const labelSupervisorZone = "Supervisor Zone"

func datasourceVcdTmSupervisorZone() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdTmSupervisorZoneRead,

		Schema: map[string]*schema.Schema{
			"supervisor_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "ID of the Supervisor",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the Supervisor Zone",
			},
			"vcenter_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Parent vCenter ID",
			},
			"region_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Parent Region ID",
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
		},
	}
}

func datasourceVcdTmSupervisorZoneRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	s, err := vcdClient.GetSupervisorById(d.Get("supervisor_id").(string))
	if err != nil {
		return diag.Errorf("error getting Supervisor: %s", err)
	}

	sz, err := s.GetSupervisorZoneByName(d.Get("name").(string))
	if err != nil {
		return diag.Errorf("error getting Supervisor Zone '%s': %s", d.Get("name").(string), err)
	}

	err = setSupervisorZoneData(d, sz)
	if err != nil {
		return diag.Errorf("error storing Supervisor data: %s", err)
	}

	d.SetId(sz.SupervisorZone.ID)

	return nil
}

func setSupervisorZoneData(d *schema.ResourceData, s *govcd.SupervisorZone) error {
	if s == nil {
		return fmt.Errorf("error")
	}
	vCenterId := ""
	if s.SupervisorZone.VirtualCenter != nil {
		vCenterId = s.SupervisorZone.VirtualCenter.ID
	}
	dSet(d, "vcenter_id", vCenterId)
	dSet(d, "name", s.SupervisorZone.Name)
	supervisorId := ""
	if s.SupervisorZone.Supervisor != nil {
		supervisorId = s.SupervisorZone.Supervisor.ID
	}
	dSet(d, "supervisor_id", supervisorId)

	regionId := ""
	if s.SupervisorZone.Region != nil {
		supervisorId = s.SupervisorZone.Region.ID
	}
	dSet(d, "region_id", regionId)
	dSet(d, "cpu_capacity_mhz", s.SupervisorZone.TotalCPUCapacityMHz)
	dSet(d, "cpu_reservation_capacity_mhz", s.SupervisorZone.CpuUsedMHz)
	dSet(d, "memory_capacity_mib", s.SupervisorZone.TotalMemoryCapacityMiB)
	dSet(d, "memory_reservation_capacity_mib", s.SupervisorZone.MemoryUsedMiB)

	return nil
}
