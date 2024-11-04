package vcd

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v3/govcd"
	"github.com/vmware/go-vcloud-director/v3/types/v56"
)

// This is a template of how a "standard" resource can look using generic CRUD functions. It might
// not cover all scenarios, but is a skeleton for quicker bootstraping of a new entity.
//
// "Search and replace the following entries"
//
// TmVdc - constant name for entity label (the lower case prefix 'label' prefix is hardcoded)
// The 'label' prefix is hardcoded in the example so that we have autocompletion working for all labelXXXX. (e.g. TmOrg)
//
// TM Vdc - text for entity label (e.g. TM Organization)
// This will be the entity label (used for logging purposes in generic functions)
//
// TmVdc - outer type (e.g. TmOrg)
// This should be a non existing new type to create in 'govcd' package
//
// types.TmVdc - inner type (e.g. types.TmOrg)
// This should be an already existing inner type in `types` package
//
// TmVdc (e.g. VcdTmOrg)

const labelTmVdc = "TM Vdc"

func resourceTmVdc() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceTmVdcCreate,
		ReadContext:   resourceTmVdcRead,
		UpdateContext: resourceTmVdcUpdate,
		DeleteContext: resourceTmVdcDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceTmVdcImport,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf(" %s", labelTmVdc),
			},
		},
	}
}

func resourceTmVdcCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	c := crudConfig[*govcd.TmVdc, types.TmVdc]{
		entityLabel:      labelTmVdc,
		getTypeFunc:      getTmVdcType,
		stateStoreFunc:   setTmVdcData,
		createFunc:       vcdClient.CreateTmVdc,
		resourceReadFunc: resourceTmVdcRead,
	}
	return createResource(ctx, d, meta, c)
}

func resourceTmVdcUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	c := crudConfig[*govcd.TmVdc, types.TmVdc]{
		entityLabel:      labelTmVdc,
		getTypeFunc:      getTmVdcType,
		getEntityFunc:    vcdClient.GetTmVdcById,
		resourceReadFunc: resourceTmVdcRead,
		// preUpdateHooks: []outerEntityHookInnerEntityType[*govcd.TmVdc, *types.TmVdc]{},
	}

	return updateResource(ctx, d, meta, c)
}

func resourceTmVdcRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	c := crudConfig[*govcd.TmVdc, types.TmVdc]{
		entityLabel:    labelTmVdc,
		getEntityFunc:  vcdClient.GetTmVdcById,
		stateStoreFunc: setTmVdcData,
	}
	return readResource(ctx, d, meta, c)
}

func resourceTmVdcDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	c := crudConfig[*govcd.TmVdc, types.TmVdc]{
		entityLabel:   labelTmVdc,
		getEntityFunc: vcdClient.GetTmVdcById,
		// preDeleteHooks: []outerEntityHook[*govcd.TmVdc]{},
	}

	return deleteResource(ctx, d, meta, c)
}

func resourceTmVdcImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	vcdClient := meta.(*VCDClient)

	d.SetId("???")
	return []*schema.ResourceData{d}, nil
}

func getTmVdcType(_ *VCDClient, d *schema.ResourceData) (*types.TmVdc, error) {
	t := &types.TmVdc{}

	return t, nil
}

func setTmVdcData2(d *schema.ResourceData, org *govcd.TmVdc) error {
	// IMPLEMENT
	return nil
}
