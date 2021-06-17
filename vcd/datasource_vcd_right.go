package vcd

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdRight() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceRightRead,
		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of Right.",
			},
			"description": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Right description",
			},
			"category_id": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Id of the category for this right",
			},
			"bundle_key": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Key used for internationalization",
			},
			"right_type": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Type of the right",
			},
			"implied_rights": &schema.Schema{
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "list of rights that are implied with this one",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Computed:    true,
							Type:        schema.TypeString,
							Description: "Name of the implied right",
						},
						"id": {
							Computed:    true,
							Type:        schema.TypeString,
							Description: "ID of the implied right",
						},
					},
				},
			},
		},
	}
}

func datasourceRightRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	rightName := d.Get("name").(string)

	right, err := vcdClient.Client.GetRightByName(rightName)
	if err != nil {
		return diag.Errorf("[right read] error searching for right %s: %s", rightName, err)
	}

	d.SetId(right.ID)
	_ = d.Set("description", right.Description)
	_ = d.Set("right_type", right.RightType)
	_ = d.Set("category_id", right.Category)
	_ = d.Set("bundle_key", right.BundleKey)
	var impliedRights []map[string]interface{}
	for _, ir := range right.ImpliedRights {
		impliedRights = append(impliedRights, map[string]interface{}{
			"name": ir.Name,
			"id":   ir.ID,
		})
	}
	err = d.Set("implied_rights", impliedRights)
	if err != nil {
		return diag.Errorf("[right read] error setting implied rights for right %s: %s", rightName, err)
	}
	return nil
}
