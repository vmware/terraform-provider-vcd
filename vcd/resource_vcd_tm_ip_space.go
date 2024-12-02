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
// TmIpSpace - constant name for entity label (the lower case prefix 'label' prefix is hardcoded)
// The 'label' prefix is hardcoded in the example so that we have autocompletion working for all labelXXXX. (e.g. TmOrg)
//
// TM IP Space - text for entity label (e.g. TM Organization)
// This will be the entity label (used for logging purposes in generic functions)
//
// TmIpSpace - outer type (e.g. TmOrg)
// This should be a non existing new type to create in 'govcd' package
//
// types.TmIpSpace - inner type without the 'types.' prefix (e.g. types.TmOrg)
// This should be an already existing inner type in `types` package
//
// VcdTmIpSpace (e.g. VcdTmOrg)

const labelTmIpSpace = "TM IP Space"

func resourceVcdTmIpSpace() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdTmIpSpaceCreate,
		ReadContext:   resourceVcdTmIpSpaceRead,
		UpdateContext: resourceVcdTmIpSpaceUpdate,
		DeleteContext: resourceVcdTmIpSpaceDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdTmIpSpaceImport,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf(" %s", labelTmIpSpace),
			},
		},
	}
}

func resourceVcdTmIpSpaceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	c := crudConfig[*govcd.TmIpSpace, types.TmIpSpace]{
		entityLabel:      labelTmIpSpace,
		getTypeFunc:      getTmIpSpaceType,
		stateStoreFunc:   setTmIpSpaceData,
		createFunc:       vcdClient.CreateTmIpSpace,
		resourceReadFunc: resourceVcdTmIpSpaceRead,
	}
	return createResource(ctx, d, meta, c)
}

func resourceVcdTmIpSpaceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	c := crudConfig[*govcd.TmIpSpace, types.TmIpSpace]{
		entityLabel:      labelTmIpSpace,
		getTypeFunc:      getTmIpSpaceType,
		getEntityFunc:    vcdClient.GetTmIpSpaceById,
		resourceReadFunc: resourceVcdTmIpSpaceRead,
		// preUpdateHooks: []outerEntityHookInnerEntityType[*govcd.TmIpSpace, *types.TmIpSpace]{},
	}

	return updateResource(ctx, d, meta, c)
}

func resourceVcdTmIpSpaceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	c := crudConfig[*govcd.TmIpSpace, types.TmIpSpace]{
		entityLabel:    labelTmIpSpace,
		getEntityFunc:  vcdClient.GetTmIpSpaceById,
		stateStoreFunc: setTmIpSpaceData,
	}
	return readResource(ctx, d, meta, c)
}

func resourceVcdTmIpSpaceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	c := crudConfig[*govcd.TmIpSpace, types.TmIpSpace]{
		entityLabel:   labelTmIpSpace,
		getEntityFunc: vcdClient.GetTmIpSpaceById,
		// preDeleteHooks: []outerEntityHook[*govcd.TmIpSpace]{},
	}

	return deleteResource(ctx, d, meta, c)
}

func resourceVcdTmIpSpaceImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	vcdClient := meta.(*VCDClient)

	d.SetId("???")
	return []*schema.ResourceData{d}, nil
}

func getTmIpSpaceType(d *schema.ResourceData) (*types.TmIpSpace, error) {
	t := &types.TmIpSpace{}

	return t, nil
}

func setTmIpSpaceData(d *schema.ResourceData, org *govcd.TmIpSpace) error {
	// IMPLEMENT
	return nil
}
