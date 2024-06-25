package vcd

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdSolutionAddonInstancePublish() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdSolutionAddonInstancePublishRead,

		Schema: map[string]*schema.Schema{
			"add_on_instance_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the Solution Add-On Instance",
			},
			"add_on_instance_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Solution Add-On Instance ID",
			},
			"org_ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"publish_to_all_tenants": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Publish Solution Add-On Instance to all tenants",
			},
			"rde_state": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Parent RDE state",
			},
		},
	}
}

func datasourceVcdSolutionAddonInstancePublishRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	addOnInstance, err := vcdClient.GetSolutionAddonInstanceByName(d.Get("add_on_instance_name").(string))
	if err != nil {
		return diag.Errorf("error retrieving Solution Add-On Instance: %s", err)
	}

	if addOnInstance.SolutionAddOnInstance != nil {
		dSet(d, "publish_to_all_tenants", addOnInstance.SolutionAddOnInstance.Scope.AllTenants)
	} else {
		dSet(d, "publish_to_all_tenants", false)
	}

	orgNames := addOnInstance.SolutionAddOnInstance.Scope.Tenants
	orgIds, err := orgNamesToIds(vcdClient, orgNames)
	if err != nil {
		return diag.Errorf("error converting Org IDs to Names: %s", err)
	}

	orgIdsSet := convertStringsToTypeSet(orgIds)
	err = d.Set("org_ids", orgIdsSet)
	if err != nil {
		return diag.Errorf("error storing Org IDs: %s", err)
	}

	dSet(d, "add_on_instance_id", addOnInstance.RdeId())
	dSet(d, "rde_state", addOnInstance.DefinedEntity.State())
	d.SetId(addOnInstance.RdeId())

	return nil
}
