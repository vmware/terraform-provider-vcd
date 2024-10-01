package vcd

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func datasourceVcdTmSupervisor() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdTmSupervisorRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the Supervisor",
			},
			"vcenter_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "",
			},
			"region_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "",
			},
		},
	}
}

func datasourceVcdTmSupervisorRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	s, err := vcdClient.GetSupervisorByName(d.Get("name").(string))
	if err != nil {
		return diag.Errorf("error getting Org: %s", err)
	}

	err = setSupervisorData(d, s.Supervisor)
	if err != nil {
		return diag.Errorf("error storing Supervisor data: %s", err)
	}

	d.SetId(s.Supervisor.SupervisorID)

	return nil
}

func setSupervisorData(d *schema.ResourceData, s *types.Supervisor) error {
	vCenterId := ""
	regionId := ""

	if s.VirtualCenter != nil {
		vCenterId = s.VirtualCenter.ID
	}

	if s.Region != nil {
		regionId = s.Region.ID
	}

	dSet(d, "vcenter_id", vCenterId)
	dSet(d, "region_id", regionId)

	return nil
}
