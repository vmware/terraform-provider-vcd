package vcd

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dsSlcChildComponent(title string) *schema.Schema {
	return &schema.Schema{
		Type:        schema.TypeSet,
		Required:    true,
		Description: "",
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"id": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: fmt.Sprintf("ID of %s", title),
				},
				"name": {
					Type:        schema.TypeString,
					Computed:    true,
					Description: fmt.Sprintf("Name of %s", title),
				},
				"is_default": {
					Type:        schema.TypeBool,
					Computed:    true,
					Description: fmt.Sprintf("Boolean value that marks if this %s should be default", title),
				},
				"capabilities": {
					Type:        schema.TypeSet,
					Computed:    true,
					Description: fmt.Sprintf("Set of capabilities for %s", title),
					Elem:        &schema.Schema{Type: schema.TypeString},
				},
			},
		},
	}
}

func datasourceVcdSolutionLandingZone() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdSolutionLandingZoneRead,

		Schema: map[string]*schema.Schema{
			"org": {
				Type:     schema.TypeString,
				Optional: true,

				// Description: "The name of organization to use, optional if defined at provider " +
				// 	"level. Useful when connected as sysadmin working across different organizations",
			},

			"state": {
				Type:        schema.TypeString,
				Description: "State reports RDE state",
				Computed:    true,
			},
			"catalog": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "IP Address of pool member",
				// 	// Warning: This catalog stores all executable .ISO files for your solution add-ons.
				// 	//
				// 	// Selecting another catalog to use in the Solution Add-On Landing Zone does not affect the solution add-ons that you already installed, but prevents you from running day-2 operations on them. То ensure that you can run day-2 operations on the add-ons that are already installed, reupload their original add-on .ISO files.
				// 	// Capabilities???
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Shows is the member is enabled or not",
						},
						"capabilities": {
							Type:        schema.TypeSet,
							Computed:    true,
							Description: "",
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},

			"vdc": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "",
						},
						"is_default": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Shows is the member is enabled or not",
						},
						"capabilities": {
							Type:        schema.TypeSet,
							Computed:    true,
							Description: "",
							Elem:        &schema.Schema{Type: schema.TypeString},
						},
						"org_vdc_network": dsSlcChildComponent("Org VDC Network"),
						"storage_policy":  dsSlcChildComponent("Storage Policy"),
						"compute_policy":  dsSlcChildComponent("Compute Policy"),
					},
				},
			},
		},
	}
}

func datasourceVcdSolutionLandingZoneRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	slz, err := vcdClient.GetExactlyOneSolutionLandingZone()
	if err != nil {
		return diag.Errorf("error retrieving Solution Landing Zone: %s", err)
	}

	err = setSlzData(d, slz)
	if err != nil {
		return diag.Errorf("error storing data to schema: %s", err)
	}

	// The real ID of Solution Landing Zone is RDE ID
	d.SetId(slz.Id())

	return nil
}
