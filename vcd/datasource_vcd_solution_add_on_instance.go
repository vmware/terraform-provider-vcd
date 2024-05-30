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
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Solution Add-On Name",
			},
			"add_on_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Parent Solution Add-On ID",
			},
			"input": {
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

	addOnInstance, err := vcdClient.GetSolutionAddonInstanceByName(d.Get("name").(string))
	if err != nil {
		return diag.Errorf("error retrieving Solution Add-On Instance: %s", err)
	}

	dSet(d, "add_on_id", addOnInstance.SolutionAddOnInstance.Prototype)
	dSet(d, "name", addOnInstance.SolutionAddOnInstance.Name)
	dSet(d, "rde_state", addOnInstance.DefinedEntity.DefinedEntity.State)

	// Retrieve creation input fields
	// 'delete_input' values cannot be read from Solution Add-On Instance as they are specified only
	// when deleting the Add-On Instance.
	inputValues, err := addOnInstance.ReadCreationInputValues(true)
	if err != nil {
		return diag.Errorf("error reading Input values from Solution Add-On instance: %s", err)
	}

	err = d.Set("input", inputValues)
	if err != nil {
		return diag.Errorf("error storing 'input' field: %s", err)
	}

	d.SetId(addOnInstance.RdeId())

	return nil
}
