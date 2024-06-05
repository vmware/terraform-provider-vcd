package vcd

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/go-vcloud-director/v2/govcd"
)

func resourceVcdVdcTemplate() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdVdcTemplateCreate,
		UpdateContext: resourceVcdVdcTemplateUpdate,
		ReadContext:   resourceVcdVdcTemplateRead,
		DeleteContext: resourceVcdVdcTemplateDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdVdcTemplateImport,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the VDC Template",
			},
			"network_provider_type": {
				Type:             schema.TypeString,
				Required:         true,
				Description:      "Type of network provider. One of: 'NSX_V' or 'NSX_T'",
				ValidateDiagFunc: validation.ToDiagFunc(validation.StringInSlice([]string{"NSX_V", "NSX_T"}, false)),
			},
		},
	}
}

func resourceVcdVdcTemplateCreate(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

func resourceVcdVdcTemplateRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return genericVcdVdcTemplateRead(ctx, d, meta)
}

func resourceVcdVdcTemplateUpdate(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return nil
}

func resourceVcdVdcTemplateDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vdcTemplate, err := getVdcTemplate(d, meta.(*VCDClient))
	if err != nil {
		if govcd.ContainsNotFound(err) {
			return nil
		}
		return diag.FromErr(err)
	}
	err = vdcTemplate.Delete()
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceVcdVdcTemplateImport(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	return nil, nil
}

func genericVcdVdcTemplateRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vdcTemplate, err := getVdcTemplate(d, meta.(*VCDClient))
	if err != nil {
		return diag.FromErr(err)
	}

	dSet(d, "name", vdcTemplate.VdcTemplate.Name)
	d.SetId(vdcTemplate.VdcTemplate.ID)

	return nil
}

// getVdcTemplate retrieves a VDC Template with the available information in the configuration.
func getVdcTemplate(d *schema.ResourceData, vcdClient *VCDClient) (*govcd.VdcTemplate, error) {
	var vdcTemplate *govcd.VdcTemplate
	var err error
	if d.Id() == "" {
		name := d.Get("name").(string)
		vdcTemplate, err = vcdClient.GetVdcTemplateByName(name)
		if err != nil {
			return nil, fmt.Errorf("could not read VDC Template with name %s: %s", name, err)
		}
	} else {
		vdcTemplate, err = vcdClient.GetVdcTemplateById(d.Id())
		if err != nil {
			return nil, fmt.Errorf("could not read VDC Template with ID %s: %s", d.Id(), err)
		}
	}
	return vdcTemplate, nil
}
