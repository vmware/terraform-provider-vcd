package vcd

import (
	"context"
	"maps"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/util"
)

func resourceVcdSolutionAddonInstance() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdSolutionAddonInstanceCreate,
		ReadContext:   resourceVcdSolutionAddonInstanceRead,
		UpdateContext: resourceVcdSolutionAddonInstanceUpdate,
		DeleteContext: resourceVcdSolutionAddonInstanceDelete,
		/* Importer: &schema.ResourceImporter{
			StateContext: resourceVcdSolutionAddonInstanceImport,
		}, */

		Schema: map[string]*schema.Schema{
			"add_on_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Solution Add-On ID",
			},
			"accept_eula": {
				Type:        schema.TypeBool,
				Required:    true,
				Description: "Defines if the resource should automatically trust Solution Add-On certificate",
			},
			"instance_name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Solution Add-On Name",
			},
			"input": {
				Type:     schema.TypeMap,
				Optional: true,
				// Computed:      true, // To be compatible with `metadata_entry`
				Description: "Key value map of Solution Add-On instance",
			},
			"delete_input": { // These will only be applicable to "delete" operation
				Type:     schema.TypeMap,
				Optional: true,
				// Computed:      true, // To be compatible with `metadata_entry`
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

func resourceVcdSolutionAddonInstanceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	addOn, err := vcdClient.GetSolutionAddonById(d.Get("add_on_id").(string))
	if err != nil {
		return diag.Errorf("error retrieving Solution Add-On: %s", err)
	}

	if addOn.SolutionEntity.Eula != "" && !d.Get("accept_eula").(bool) {
		return diag.Errorf("cannot create Add-On instance without accepting EULA.\n\n%s\n\n: %s", addOn.SolutionEntity.Eula, err)
	}

	input := d.Get("input")
	inputMap := input.(map[string]interface{})
	inputCopy := make(map[string]interface{})

	maps.Copy(inputCopy, inputMap)

	inputCopy["name"] = d.Get("instance_name").(string)
	inputCopy["input-delete-previous-uiplugin-versions"] = false

	addOnInstance, _, err := addOn.CreateSolutionAddOnInstance(inputCopy)
	if err != nil {
		return diag.Errorf("error creating Solution Add-On ('%s') instance: %s",
			addOn.DefinedEntity.DefinedEntity.Name, err)
	}

	util.Logger.Println("[TRACE] DAINIUS DAINIUS DAINIUS")
	util.Logger.Printf("[TRACE] DAINIUS %#v\n", addOnInstance.DefinedEntity.DefinedEntity.ID)

	d.SetId(addOnInstance.DefinedEntity.DefinedEntity.ID)
	// d.SetId(addOnInstance.RdeId())

	return resourceVcdSolutionAddonInstanceRead(ctx, d, meta)
}

func resourceVcdSolutionAddonInstanceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// vcdClient := meta.(*VCDClient)

	return resourceVcdSolutionAddonInstanceRead(ctx, d, meta)
}

func resourceVcdSolutionAddonInstanceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// vcdClient := meta.(*VCDClient)

	// addOnInstance, err := vcdClient.GetSolutionAddOnInstanceById(d.Id())
	// if err != nil {
	// 	return diag.Errorf("error retrieving Solution Add-On instance by ID: %s", err)
	// }

	// dSet(d, "rde_state", addOnInstance.DefinedEntity.DefinedEntity.State)

	return nil
}

func resourceVcdSolutionAddonInstanceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	addOnInstance, err := vcdClient.GetSolutionAddOnInstanceById(d.Id())
	if err != nil {
		return diag.Errorf("error retrieving Solution Add-On instance by ID: %s", err)
	}

	deleteInput := d.Get("delete_input")
	// inputMap := convertToStringMap(input.(map[string]interface{}))
	deleteInputMap := deleteInput.(map[string]interface{})

	_, err = addOnInstance.RemoveSolutionAddOnInstance(deleteInputMap)
	if err != nil {
		return diag.Errorf("error removing Solution Add-On Instance: %s", err)
	}

	return nil

}

/* func resourceVcdSolutionAddonInstanceImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	vcdClient := meta.(*VCDClient)

	return []*schema.ResourceData{d}, nil

} */
