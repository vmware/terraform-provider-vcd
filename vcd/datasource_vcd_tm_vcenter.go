package vcd

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v3/govcd"
	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

func datasourceVcdTmVcenter() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdVcenterRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("Name of %s", labelTmVirtualCenter),
			},
			"url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("URL of %s", labelTmVirtualCenter),
			},
			"username": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Username of %s", labelTmVirtualCenter),
			},
			"is_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: fmt.Sprintf("Should the %s be enabled", labelTmVirtualCenter),
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Description of %s", labelTmVirtualCenter),
			},
			"has_proxy": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: fmt.Sprintf("A flag that shows if %s has proxy defined", labelTmVirtualCenter),
			},
			"is_connected": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: fmt.Sprintf("A flag that shows if %s is connected", labelTmVirtualCenter),
			},
			"mode": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Mode of %s", labelTmVirtualCenter),
			},
			"connection_status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Listener state of %s", labelTmVirtualCenter),
			},
			"cluster_health_status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Mode of %s", labelTmVirtualCenter),
			},
			"vcenter_version": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Version of %s", labelTmVirtualCenter),
			},
			"uuid": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("%s UUID", labelTmVirtualCenter),
			},
			"vcenter_host": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("%s hostname", labelTmVirtualCenter),
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "vCenter status",
			},
		},
	}
}

func datasourceVcdVcenterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	c := dsReadConfig[*govcd.VCenter, types.VSphereVirtualCenter]{
		entityLabel:    labelTmVirtualCenter,
		getEntityFunc:  vcdClient.GetVCenterByName,
		stateStoreFunc: setTmVcenterData,
	}
	return readDatasource(ctx, d, meta, c)
}
