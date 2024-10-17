package vcd

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v3/govcd"
)

func datasourceVcdNsxvApplication() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdNsxvApplicationRead,

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
				Description: "Identifier of the application",
			},
			"protocol": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Protocol used by the application",
			},
			"ports": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Ports used by the application",
			},
			"source_port": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Source port used by the application",
			},
			"app_guid": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Application GUID name",
			},
		},
	}
}

func datasourceVcdNsxvApplicationRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	vdcId := d.Get("vdc_id").(string)
	applicationName := d.Get("name").(string)

	dfw := govcd.NewNsxvDistributedFirewall(&vcdClient.Client, vdcId)

	application, err := dfw.GetServiceByName(applicationName)

	if err != nil {
		return diag.Errorf("error retrieving application: %s - %s", govcd.ErrorEntityNotFound, err)
	}

	dSet(d, "name", application.Name)
	dSet(d, "id", application.ObjectID)
	if application.Element.ApplicationProtocol != nil {
		dSet(d, "protocol", *application.Element.ApplicationProtocol)
	}
	if application.Element.Value != nil {
		dSet(d, "ports", *application.Element.Value)
	}
	if application.Element.SourcePort != nil {
		dSet(d, "source_port", *application.Element.SourcePort)
	}
	if application.Element.AppGuidName != nil {
		dSet(d, "app_guid", *application.Element.AppGuidName)
	}
	d.SetId(application.ObjectID)

	return nil
}
