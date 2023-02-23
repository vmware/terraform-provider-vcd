package vcd

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
)

func datasourceVcdNsxvService() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceVcdNsxvServiceRead,

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
				Description: "Identifier of the service",
			},
			"protocol": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Protocol used by the service",
			},
			"ports": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Ports used by the service",
			},
			"source_port": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Source port used by the service",
			},
			"app_guid": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Application GUID name",
			},
		},
	}
}

func datasourceVcdNsxvServiceRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	vdcId := d.Get("vdc_id").(string)
	serviceName := d.Get("name").(string)

	dfw := govcd.NewNsxvDistributedFirewall(&vcdClient.Client, vdcId)

	service, err := dfw.GetServiceByName(serviceName)

	if err != nil {
		return diag.Errorf("error retrieving service: %s - %s", govcd.ErrorEntityNotFound, err)
	}

	dSet(d, "name", service.Name)
	dSet(d, "id", service.ObjectID)
	if service.Element.ApplicationProtocol != nil {
		dSet(d, "protocol", *service.Element.ApplicationProtocol)
	}
	if service.Element.Value != nil {
		dSet(d, "ports", *service.Element.Value)
	}
	if service.Element.SourcePort != nil {
		dSet(d, "source_port", *service.Element.SourcePort)
	}
	if service.Element.AppGuidName != nil {
		dSet(d, "app_guid", *service.Element.AppGuidName)
	}
	d.SetId(service.ObjectID)

	return nil
}
