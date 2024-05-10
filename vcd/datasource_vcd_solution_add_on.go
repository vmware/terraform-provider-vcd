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
			"org": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Description: "The name of organization to use, optional if defined at provider " +
					"level. Useful when connected as sysadmin working across different organizations",
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				// Description: "absolute or relative path to Solution Add-On ISO file",
			},
			"catalog_item_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "absolute or relative path to Solution Add-On ISO file",
			},
			"rde_state": {
				Type:        schema.TypeString,
				Description: "State reports RDE state",
				Computed:    true,
			},
		},
	}
}

func datasourceVcdSolutionAddonRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	slz, err := vcdClient.GetSolutionAddonByName(d.Get("name").(string))
	if err != nil {
		return diag.Errorf("error retrieving Solution Add-On: %s", err)
	}

	dSet(d, "rde_state", slz.DefinedEntity.DefinedEntity.State)
	dSet(d, "catalog_item_id", slz.SolutionEntity.Origin.CatalogItemId)

	return nil
}
