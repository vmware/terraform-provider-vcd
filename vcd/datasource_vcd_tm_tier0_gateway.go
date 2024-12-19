package vcd

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v3/govcd"
	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

const labelTmTier0Gateway = "TM Tier 0 Gateway"

func datasourceVcdTmTier0Gateway() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdTmTier0GatewayRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("Display Name of %s", labelTmTier0Gateway),
			},
			"region_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf("Parent %s ID", labelTmRegion),
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Description of %s", labelTmTier0Gateway),
			},
			"parent_tier_0_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: fmt.Sprintf("Parent Tier 0 Gateway of %s", labelTmTier0Gateway),
			},
			"already_imported": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: fmt.Sprintf("Defines if the T0 is already imported of %s", labelTmTier0Gateway),
			},
		},
	}
}

func datasourceVcdTmTier0GatewayRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	getT0ByName := func(name string) (*govcd.TmTier0Gateway, error) {
		return vcdClient.GetTmTier0GatewayWithContextByName(name, d.Get("region_id").(string), true)
	}

	c := dsReadConfig[*govcd.TmTier0Gateway, types.TmTier0Gateway]{
		entityLabel:    labelTmTier0Gateway,
		getEntityFunc:  getT0ByName,
		stateStoreFunc: setTmTier0GatewayData,
	}
	return readDatasource(ctx, d, meta, c)
}

func setTmTier0GatewayData(_ *VCDClient, d *schema.ResourceData, t *govcd.TmTier0Gateway) error {
	d.SetId(t.TmTier0Gateway.ID) // So far the API returns plain UUID (not URN)
	dSet(d, "name", t.TmTier0Gateway.DisplayName)
	dSet(d, "description", t.TmTier0Gateway.Description)
	dSet(d, "parent_tier_0_id", t.TmTier0Gateway.ParentTier0ID)
	dSet(d, "already_imported", t.TmTier0Gateway.AlreadyImported)

	return nil
}
