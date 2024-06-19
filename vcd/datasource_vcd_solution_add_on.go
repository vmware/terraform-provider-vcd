package vcd

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdSolutionAddon() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdSolutionAddonRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the Solution Add-On Defined Entity (e.g. 'vmware.ds-1.4.0-23376809')",
			},
			"catalog_item_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Catalog item ID of the Solution Add-On",
			},
			"rde_state": {
				Type:        schema.TypeString,
				Description: "RDE State of the Solution Add-On",
				Computed:    true,
			},
		},
	}
}

func datasourceVcdSolutionAddonRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	slzAddOn, err := vcdClient.GetSolutionAddonByName(d.Get("name").(string))
	if err != nil {
		return diag.Errorf("error retrieving Solution Add-On: %s", err)
	}

	if slzAddOn == nil || slzAddOn.DefinedEntity == nil ||
		slzAddOn.DefinedEntity.DefinedEntity == nil || slzAddOn.DefinedEntity.DefinedEntity.State == nil ||
		*slzAddOn.DefinedEntity.DefinedEntity.State == "" ||
		slzAddOn.SolutionAddOnEntity.Origin.CatalogItemId == "" {
		return diag.Errorf("no values filled in for Solution Add-On '%s'", d.Get("name").(string))
	}

	dSet(d, "rde_state", slzAddOn.DefinedEntity.DefinedEntity.State)
	dSet(d, "catalog_item_id", slzAddOn.SolutionAddOnEntity.Origin.CatalogItemId)
	d.SetId(slzAddOn.RdeId())

	return nil
}
