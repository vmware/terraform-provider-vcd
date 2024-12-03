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
// TmTier0Gateway - constant name for entity label (the lower case prefix 'label' prefix is hardcoded)
// The 'label' prefix is hardcoded in the example so that we have autocompletion working for all labelXXXX. (e.g. TmOrg)
//
// ENTITY-LABELTEXT-PLACEHOLDER - text for entity label (e.g. TM Organization)
// Must already be defined in resource skeleton
// This will be the entity label (used for logging purposes in generic functions)
//
// TmTier0Gateway - outer type (e.g. TmOrg)
// This should be a non existing new type to create in 'govcd' package
//
// TmTier0Gateway - inner type without the 'types.' prefix (e.g. types.TmOrg)
// This should be an already existing inner type in `types` package
//
// VcdTmTier0Gateway (e.g. VcdTmOrg)

const labelTmTier0Gateway = "TM Tier 0 Gateway"

func resourceVcdTmTier0Gateway() *schema.Resource {
	return &schema.Resource{
		ReadContext: resourceVcdTmTier0GatewayRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: fmt.Sprintf(" %s", labelTmTier0Gateway),
			},
		},
	}
}

func resourceVcdTmTier0GatewayRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	c := dsReadConfig[*govcd.TmTier0Gateway, types.TmTier0Gateway]{
		entityLabel:    labelTmTier0Gateway,
		getEntityFunc:  vcdClient.GetTmTier0GatewayById,
		stateStoreFunc: setTmTier0GatewayData,
	}
	return readDatasource(ctx, d, meta, c)
}

func setTmTier0GatewayData(d *schema.ResourceData, org *govcd.TmTier0Gateway) error {
	// IMPLEMENT
	return nil
}
