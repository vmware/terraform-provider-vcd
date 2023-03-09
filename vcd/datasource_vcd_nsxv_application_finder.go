package vcd

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func datasourceVcdNsxvApplicationFinder() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdNsxvApplicationFinderRead,

		Schema: map[string]*schema.Schema{
			"vdc_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of VDC",
			},
			"search_expression": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Regular expression used to search applications or groups",
			},
			// Note: the search is case-insensitive by default, to mimic the behavior of the UI
			"case_sensitive": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Convert the search to case sensitive",
			},
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "Type of object. One of 'application', 'application_group'",
				ValidateFunc: validation.StringInSlice([]string{"application", "application_group"}, false),
			},
			"objects": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "Objects found with the combination of 'search_expression' + 'type'",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Name of the object",
						},
						"value": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Identifier of the object",
						},
						"type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Type of the object",
						},
					},
				},
			},
		},
	}
}

func datasourceVcdNsxvApplicationFinderRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	vdcId := d.Get("vdc_id").(string)
	dfw := govcd.NewNsxvDistributedFirewall(&vcdClient.Client, vdcId)

	wantedType := d.Get("type").(string)
	rawRegexp := d.Get("search_expression").(string)
	rawCaseSensitive := d.Get("case_sensitive")
	caseSensitive := false
	if rawCaseSensitive != nil {
		caseSensitive = rawCaseSensitive.(bool)
	}
	if !caseSensitive {
		rawRegexp = `(?i)` + rawRegexp
	}

	var result []map[string]string

	var applications []types.Application
	var applicationGroups []types.ApplicationGroup
	var err error

	switch wantedType {
	case "application":
		applications, err = dfw.GetServicesByRegex(rawRegexp)
	case "application_group":
		applicationGroups, err = dfw.GetServiceGroupsByRegex(rawRegexp)
	}

	if err != nil {
		return diag.Errorf("error retrieving %s list: %s - %s", wantedType, govcd.ErrorEntityNotFound, err)
	}

	for _, application := range applications {
		item := map[string]string{
			"name":  application.Name,
			"type":  application.Type.TypeName,
			"value": application.ObjectID,
		}
		result = append(result, item)
	}

	for _, ag := range applicationGroups {
		item := map[string]string{
			"name":  ag.Name,
			"type":  ag.Type.TypeName,
			"value": ag.ObjectID,
		}
		result = append(result, item)
	}
	err = d.Set("objects", result)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(vdcId)

	return nil

}
