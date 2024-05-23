package vcd

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdSolutionAddonInstance() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdSolutionAddonInstanceRead,

		Schema: map[string]*schema.Schema{
			"add_on_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Solution Add-On ID",
			},

			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Solution Add-On Name",
			},
			"input": {
				Type:        schema.TypeMap,
				Computed:    true,
				Description: "Key value map of Solution Add-On instance",
			},
			"delete_input": { // These will only be applicable to "delete" operation
				Type:        schema.TypeMap,
				Computed:    true,
				Description: "Key value map of Solution Add-On instance",
			},
			"rde_state": {
				Type:        schema.TypeString,
				Description: "Parent RDE state",
				Computed:    true,
			},
		},
	}
}

func datasourceVcdSolutionAddonInstanceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	slzAddOn, err := vcdClient.GetSolutionAddonById(d.Get("add_on_id").(string))
	if err != nil {
		return diag.Errorf("error retrieving Solution Add-On: %s", err)
	}

	slzAddOnInstance, err := slzAddOn.GetInstanceByName(d.Get("name").(string))
	if err != nil {
		return diag.Errorf("error retrieving Solution Add-On Instance: %s", err)
	}

	d.SetId(slzAddOnInstance.RdeId())

	// d.Set("publish_to_all_tenants", slzAddOnInstance.SolutionAddOnInstance.Scope.AllTenants)

	// orgNames := slzAddOnInstance.SolutionAddOnInstance.Scope.Tenants
	// orgIds, err := orgNamesToIds(vcdClient, orgNames)
	// if err != nil {
	// 	return diag.Errorf("error converting Org IDs to Names: %s", err)
	// }

	// orgIdsSet := convertStringsToTypeSet(orgIds)
	// err = d.Set("org_ids", orgIdsSet)
	// if err != nil {
	// 	return diag.Errorf("error storing Org IDs: %s", err)
	// }

	// dSet(d, "rde_state", slzAddOnInstance.DefinedEntity.DefinedEntity.State)
	// d.SetId(slzAddOnInstance.RdeId())

	return nil
}
