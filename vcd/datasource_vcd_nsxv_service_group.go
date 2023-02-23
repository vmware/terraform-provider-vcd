package vcd

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
)

func datasourceVcdNsxvServiceGroup() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdNsxvServiceGroupRead,

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
				Description: "Identifier of the service group",
			},
			"services": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "Services belonging to this service group",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Name of the service",
						},
						"value": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Identifier of the service",
						},
					},
				},
			},
		},
	}
}

func datasourceVcdNsxvServiceGroupRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	vdcId := d.Get("vdc_id").(string)
	serviceGroupName := d.Get("name").(string)

	dfw := govcd.NewNsxvDistributedFirewall(&vcdClient.Client, vdcId)

	serviceGroup, err := dfw.GetServiceGroupByName(serviceGroupName)
	if err != nil {
		return diag.Errorf("error retrieving service groups: %s - %s", govcd.ErrorEntityNotFound, err)
	}

	var services []map[string]string

	for _, s := range serviceGroup.Member {
		item := map[string]string{
			"name":  s.Name,
			"value": s.ObjectID,
		}
		services = append(services, item)
	}
	err = d.Set("services", services)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(serviceGroup.ObjectID)

	return nil

}
