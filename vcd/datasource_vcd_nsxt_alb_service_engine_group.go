package vcd

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdAlbServiceEngineGroup() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdAlbServiceEngineGroupRead,

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "NSX-T ALB Service Engine Group name",
			},
			"sync_on_refresh": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Boolean value that shows if sync should be performed on every refresh",
			},
			"description": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "NSX-T ALB Service Engine Group description",
			},
			"alb_cloud_id": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "NSX-T ALB backing Cloud ID",
			},
			"reservation_model": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "NSX-T ALB Service Engine Group reservation model. One of 'DEDICATED', 'SHARED'",
			},
			"max_virtual_services": &schema.Schema{
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "NSX-T ALB Service Engine Group maximum virtual services",
			},
			"reserved_virtual_services": &schema.Schema{
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "NSX-T ALB Service Engine Group reserved virtual services",
			},
			"deployed_virtual_services": &schema.Schema{
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "NSX-T ALB Service Engine Group deployed virtual services",
			},
			"ha_mode": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "NSX-T ALB Service Engine Group HA mode",
			},
			"overallocated": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Boolean value that shows if ",
			},
		},
	}
}

func datasourceVcdAlbServiceEngineGroupRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	if !vcdClient.Client.IsSysAdmin {
		return diag.Errorf("this resource is only supported for Providers")
	}

	name := d.Get("name").(string)
	albSeGroup, err := vcdClient.GetAlbServiceEngineGroupByName("", name)
	if err != nil {
		return diag.Errorf("unable to find NSX-T ALB Service Engine Group '%s': %s", name, err)
	}

	// If "Sync" is requested for read operations - perform Sync operation and re-read the entity to get latest data
	if d.Get("sync_on_refresh").(bool) {
		err = albSeGroup.Sync()
		if err != nil {
			return diag.Errorf("error executing Sync operation for NSX-T ALB Service Engine Group '%s': %s",
				albSeGroup.NsxtAlbServiceEngineGroup.Name, err)
		}

		// re-read new values post sync
		albSeGroup, err = vcdClient.GetAlbServiceEngineGroupById(albSeGroup.NsxtAlbServiceEngineGroup.ID)
		if err != nil {
			return diag.Errorf("error re-reading NSX-T ALB Service Engine Group after Sync: %s", err)
		}
	}

	err = setNsxtAlbServiceEngineGroupData(d, albSeGroup.NsxtAlbServiceEngineGroup)
	if err != nil {
		return diag.Errorf("error setting NSX-T ALB Service Engine Group data: %s", err)
	}

	d.SetId(albSeGroup.NsxtAlbServiceEngineGroup.ID)

	return nil
}
