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
			"catalog_item_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "absolute or relative path to Solution Add-on ISO file",
			},
			"addon_path": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "absolute or relative path to Solution Add-on ISO file",
			},
			// Trust certificate - should we untrust (remove the certificate) in "update"?
			"trust_certificate": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "",
			},
			"accept_eula": {
				Type:        schema.TypeBool,
				Computed:    true,
				ForceNew:    true,
				Description: "",
			},
			"state": {
				Type:        schema.TypeString,
				Description: "State reports RDE state",
				Computed:    true,
			},
		},
	}
}

func datasourceVcdSolutionAddonRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	slz, err := vcdClient.GetSolutionAddonById(d.Id())
	if err != nil {
		return diag.Errorf("error retrieving Solution Add-on: %s", err)
	}

	// dSet(d, "user", slz.SolutionEntity.Origin.AcceptedBy)
	dSet(d, "state", slz.DefinedEntity.DefinedEntity.State)
	dSet(d, "catalog_item_id", slz.SolutionEntity.Origin.CatalogItemId)

	return nil
}
