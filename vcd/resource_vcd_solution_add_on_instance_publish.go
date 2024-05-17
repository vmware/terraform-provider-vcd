package vcd

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceVcdSolutionAddonInstancePublish() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdSolutionAddonInstancePublishCreate,
		ReadContext:   resourceVcdSolutionAddonInstancePublishRead,
		UpdateContext: resourceVcdSolutionAddonInstancePublishUpdate,
		DeleteContext: resourceVcdSolutionAddonInstancePublishDelete,
		/* Importer: &schema.ResourceImporter{
			StateContext: resourceVcdSolutionAddonInstancePublishImport,
		}, */

		Schema: map[string]*schema.Schema{
			"add_on_instance_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Solution Add-On ID",
			},
			"orgs": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "A set of Org Names",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"publish_to_all_tenants": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Publish Solution Add-On Instance to all tenants",
			},
		},
	}
}

func resourceVcdSolutionAddonInstancePublishCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	addOnInstance, err := vcdClient.GetSolutionAddOnInstanceById(d.Get("add_on_instance_id").(string))
	if err != nil {
		return diag.Errorf("error retrieving Solution Add-On Instance: %s", err)
	}

	scopes := convertSchemaSetToSliceOfStrings(d.Get("orgs").(*schema.Set))

	_, err = addOnInstance.Publishing(scopes, d.Get("publish_to_all_tenants").(bool))
	if err != nil {
		return diag.Errorf("error publishing Solution Add-On Instance %s: %s", addOnInstance.SolutionEntity.Name, err)
	}

	d.SetId(addOnInstance.DefinedEntity.DefinedEntity.ID)

	return resourceVcdSolutionAddonInstancePublishRead(ctx, d, meta)
}

func resourceVcdSolutionAddonInstancePublishUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// vcdClient := meta.(*VCDClient)

	return resourceVcdSolutionAddonInstancePublishRead(ctx, d, meta)
}

func resourceVcdSolutionAddonInstancePublishRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// vcdClient := meta.(*VCDClient)

	// addOnInstance, err := vcdClient.GetSolutionAddOnInstanceById(d.Id())
	// if err != nil {
	// 	return diag.Errorf("error retrieving Solution Add-On instance by ID: %s", err)
	// }

	// dSet(d, "rde_state", addOnInstance.DefinedEntity.DefinedEntity.State)

	return nil
}

func resourceVcdSolutionAddonInstancePublishDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	addOnInstance, err := vcdClient.GetSolutionAddOnInstanceById(d.Get("add_on_instance_id").(string))
	if err != nil {
		return diag.Errorf("error retrieving Solution Add-On Instance: %s", err)
	}

	_, err = addOnInstance.Publishing(nil, false)
	if err != nil {
		return diag.Errorf("error unpublishing Solution Add-On Instance %s: %s", addOnInstance.SolutionEntity.Name, err)
	}

	return resourceVcdSolutionAddonInstancePublishRead(ctx, d, meta)
}

/* func resourceVcdSolutionAddonInstancePublishImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	vcdClient := meta.(*VCDClient)

	return []*schema.ResourceData{d}, nil

} */
