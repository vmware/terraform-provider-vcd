package vcd

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func datasourceVcdNsxvServiceFinder() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdNsxvServiceFinderRead,

		Schema: map[string]*schema.Schema{
			"vdc_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of VDC",
			},
			"regexp": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Regular expression used to search services or groups",
			},
			"type": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "Type of object. One of 'service', 'service_group'",
				ValidateFunc: validation.StringInSlice([]string{"service", "service_group"}, false),
			},
			"objects": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "Objects found with the combination of 'regexp' + 'type'",
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

func datasourceVcdNsxvServiceFinderRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	vdcId := d.Get("vdc_id").(string)
	dfw := govcd.NewNsxvDistributedFirewall(&vcdClient.Client, vdcId)

	wantedType := d.Get("type").(string)
	rawRegexp := d.Get("regexp").(string)

	var result []map[string]string

	var services []types.Application
	var serviceGroups []types.ApplicationGroup
	var err error

	switch wantedType {
	case "service":
		services, err = dfw.GetServicesByRegex(rawRegexp)
	case "service_group":
		serviceGroups, err = dfw.GetServiceGroupsByRegex(rawRegexp)
	}

	if err != nil {
		return diag.FromErr(err)
	}

	for _, service := range services {
		item := map[string]string{
			"name":  service.Name,
			"type":  service.Type.TypeName,
			"value": service.ObjectID,
		}
		result = append(result, item)
	}

	for _, sg := range serviceGroups {
		item := map[string]string{
			"name":  sg.Name,
			"type":  sg.Type.TypeName,
			"value": sg.ObjectID,
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
