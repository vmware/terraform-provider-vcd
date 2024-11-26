package vcd

import (
	"context"
	"fmt"
	"net/url"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v3/govcd"
	"github.com/vmware/go-vcloud-director/v3/types/v56"
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
				Description: fmt.Sprintf("URL of %s", labelVirtualCenter),
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
			"connection_status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Listener state of %s", labelVirtualCenter),
			},
			"cluster_health_status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Mode of %s", labelVirtualCenter),
			},
			"vcenter_version": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Version of %s", labelVirtualCenter),
			},
			"uuid": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("%s UUID", labelVirtualCenter),
			},
			"vcenter_host": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("%s hostname", labelVirtualCenter),
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
	if err := classicVcdVcenterReadStatus(vcdClient, d); err != nil {
		return err
	}

	c := crudConfig[*govcd.VCenter, types.VSphereVirtualCenter]{
		entityLabel:    labelVirtualCenter,
		getEntityFunc:  vcdClient.GetVCenterByName,
		stateStoreFunc: setTmVcenterData,
	}
	return readDatasource(ctx, d, meta, c)
}

// classicVcdVcenterReadStatus
func classicVcdVcenterReadStatus(vcdClient *VCDClient, d *schema.ResourceData) diag.Diagnostics {
	if vcdClient.Client.IsTm() {
		return nil
	}
	vCenterName := d.Get("name").(string)

	vcs, err := govcd.QueryVirtualCenters(vcdClient.VCDClient, "name=="+url.QueryEscape(vCenterName))
	if err != nil {
		return diag.Errorf("error occurred while querying vCenters: %s", err)
	}

	if len(vcs) == 0 {
		return diag.Errorf("%s: could not identify single vCenter. Got %d with name '%s'",
			govcd.ErrorEntityNotFound, len(vcs), vCenterName)
	}

	if len(vcs) > 1 {
		return diag.Errorf("could not identify single vCenter. Got %d with name '%s'",
			len(vcs), vCenterName)
	}

	dSet(d, "status", vcs[0].Status)

	return nil
}
