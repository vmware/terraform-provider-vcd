package vcd

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v3/govcd"
	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

const labelTmProviderGateway = "TM Provider Gateway"

func resourceTVcdTmProviderGateway() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceTVcdTmProviderGatewayCreate,
		ReadContext:   resourceTVcdTmProviderGatewayRead,
		UpdateContext: resourceTVcdTmProviderGatewayUpdate,
		DeleteContext: resourceTVcdTmProviderGatewayDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceTVcdTmProviderGatewayImport,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf(" %s", labelTmProviderGateway),
			},
		},
	}
}

func resourceTVcdTmProviderGatewayCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	c := crudConfig[*govcd.TmProviderGateway, types.TmProviderGateway]{
		entityLabel:      labelTmProviderGateway,
		getTypeFunc:      getTmProviderGatewayType,
		stateStoreFunc:   setTmProviderGatewayData,
		createFunc:       vcdClient.CreateTmProviderGateway,
		resourceReadFunc: resourceTVcdTmProviderGatewayRead,
	}
	return createResource(ctx, d, meta, c)
}

func resourceTVcdTmProviderGatewayUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	c := crudConfig[*govcd.TmProviderGateway, types.TmProviderGateway]{
		entityLabel:      labelTmProviderGateway,
		getTypeFunc:      getTmProviderGatewayType,
		getEntityFunc:    vcdClient.GetTmProviderGatewayById,
		resourceReadFunc: resourceTVcdTmProviderGatewayRead,
		// preUpdateHooks: []outerEntityHookInnerEntityType[*govcd.TmProviderGateway, *types.TmProviderGateway]{},
	}

	return updateResource(ctx, d, meta, c)
}

func resourceTVcdTmProviderGatewayRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	c := crudConfig[*govcd.TmProviderGateway, types.TmProviderGateway]{
		entityLabel:    labelTmProviderGateway,
		getEntityFunc:  vcdClient.GetTmProviderGatewayById,
		stateStoreFunc: setTmProviderGatewayData,
	}
	return readResource(ctx, d, meta, c)
}

func resourceTVcdTmProviderGatewayDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	c := crudConfig[*govcd.TmProviderGateway, types.TmProviderGateway]{
		entityLabel:   labelTmProviderGateway,
		getEntityFunc: vcdClient.GetTmProviderGatewayById,
		// preDeleteHooks: []outerEntityHook[*govcd.TmProviderGateway]{},
	}

	return deleteResource(ctx, d, meta, c)
}

func resourceTVcdTmProviderGatewayImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	vcdClient := meta.(*VCDClient)

	d.SetId(vcdClient.Org)
	return []*schema.ResourceData{d}, nil
}

func getTmProviderGatewayType(vcdClient *VCDClient, d *schema.ResourceData) (*types.TmProviderGateway, error) {
	t := &types.TmProviderGateway{}

	return t, nil
}

func setTmProviderGatewayData(d *schema.ResourceData, org *govcd.TmProviderGateway) error {
	// IMPLEMENT
	return nil
}
