package vcd

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func datasourceVcdVcenter() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdVcenterRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("Name of %s", labelVirtualCenter),
			},
			"url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("URL including port of %s", labelVirtualCenter),
			},
			"username": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Username of %s", labelVirtualCenter),
			},
			"is_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: fmt.Sprintf("Should the %s be enabled", labelVirtualCenter),
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Description of %s", labelVirtualCenter),
			},
			"has_proxy": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: fmt.Sprintf("A flag that shows if %s has proxy defined", labelVirtualCenter),
			},
			"is_connected": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: fmt.Sprintf("A flag that shows if %s is connected", labelVirtualCenter),
			},
			"mode": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Mode of %s", labelVirtualCenter),
			},
			"listener_state": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Listener state of %s", labelVirtualCenter),
			},
			"cluster_health_status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Mode of %s", labelVirtualCenter),
			},
			"version": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Version of %s", labelVirtualCenter),
			},
			"uuid": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("%s UUID", labelVirtualCenter),
			},
		},
	}
}

func datasourceVcdVcenterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	c := crudConfig[*govcd.VCenter, types.VSphereVirtualCenter]{
		entityLabel:    labelVirtualCenter,
		getEntityFunc:  vcdClient.GetVCenterByName,
		stateStoreFunc: setTmVcenterData,
	}
	return readDatasource(ctx, d, meta, c)
}
