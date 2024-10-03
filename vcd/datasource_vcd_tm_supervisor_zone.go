package vcd

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

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

	err = setSupervisorZoneData(d, sz.SupervisorZone)
	if err != nil {
		return diag.Errorf("error storing Supervisor data: %s", err)
	}

	d.SetId(sz.SupervisorZone.ID)

	return nil
}

func setSupervisorZoneData(d *schema.ResourceData, s *types.SupervisorZone) error {
	vCenterId := ""
	if s.VirtualCenter != nil {
		vCenterId = s.VirtualCenter.ID
	}
	dSet(d, "vcenter_id", vCenterId)
	dSet(d, "name", s.Name)
	supervisorId := ""
	if s.Supervisor != nil {
		supervisorId = s.Supervisor.ID
	}
	dSet(d, "supervisor_id", supervisorId)

	return nil
}
