package vcd

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
)

func datasourceVcdNsxvApplicationGroup() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdNsxvApplicationGroupRead,

		Schema: map[string]*schema.Schema{
			"vdc_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of VDC",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the object",
			},
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Identifier of the application group",
			},
			"applications": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "Applications belonging to this application group",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Name of the application",
						},
						"value": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Identifier of the application",
						},
					},
				},
			},
		},
	}
}

func datasourceVcdNsxvApplicationGroupRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	vdcId := d.Get("vdc_id").(string)
	applicationGroupName := d.Get("name").(string)

	dfw := govcd.NewNsxvDistributedFirewall(&vcdClient.Client, vdcId)

	applicationGroup, err := dfw.GetServiceGroupByName(applicationGroupName)
	if err != nil {
		return diag.Errorf("error retrieving application groups: %s - %s", govcd.ErrorEntityNotFound, err)
	}

	var applications []map[string]string

	for _, s := range applicationGroup.Member {
		item := map[string]string{
			"name":  s.Name,
			"value": s.ObjectID,
		}
		applications = append(applications, item)
	}
	err = d.Set("applications", applications)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(applicationGroup.ObjectID)

	return nil

}
