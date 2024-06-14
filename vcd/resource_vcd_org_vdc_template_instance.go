package vcd

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
)

func resourceVcdOrgVdcTemplateInstance() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdVdcTemplateInstantiateCreate,
		ReadContext:   resourceVcdVdcTemplateInstantiateRead,
		DeleteContext: resourceVcdVdcTemplateInstantiateDelete,
		Schema: map[string]*schema.Schema{
			"org_vdc_template_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the VDC template to instantiate",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Name of the VDC to be instantiated",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				ForceNew:    true,
				Description: "Description of the VDC to be instantiated",
			},
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the Organization where the VDC will be instantiated",
			},
		},
	}
}

func resourceVcdVdcTemplateInstantiateCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vdcTemplateId := d.Get("org_vdc_template_id").(string)
	vdcTemplate, err := meta.(*VCDClient).GetVdcTemplateById(vdcTemplateId)
	if err != nil {
		return diag.Errorf("could not instantiate the VDC Template: %s", err)
	}
	vdcId, err := vdcTemplate.Instantiate(d.Get("name").(string), d.Get("description").(string), d.Get("organization_id").(string))
	if err != nil {
		diag.Errorf("failed instantiating the VDC Template: %s", err)
	}
	d.SetId(vdcId)
	return resourceVcdVdcTemplateInstantiateRead(ctx, d, meta)
}

func resourceVcdVdcTemplateInstantiateRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vdc, err := getInstantiatedVdc(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(vdc.Vdc.ID)
	return nil
}

func resourceVcdVdcTemplateInstantiateDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vdc, err := getInstantiatedVdc(d, meta)
	if err != nil {
		return diag.FromErr(err)
	}
	err = vdc.DeleteWait(true, true)
	if err != nil {
		return diag.Errorf("failed deleting instantiated VDC '%s': %s", vdc.Vdc.ID, err)
	}
	return nil
}

func getInstantiatedVdc(d *schema.ResourceData, meta interface{}) (*govcd.Vdc, error) {
	vcdClient := meta.(*VCDClient)

	org, err := vcdClient.GetOrgById(d.Get("org_id").(string))
	if err != nil {
		return nil, fmt.Errorf("could not retrieve the instantiated VDC: %s", err)
	}

	var vdc *govcd.Vdc
	if d.Id() == "" {
		vdc, err = org.GetVDCById(d.Id(), false)
	} else {
		vdc, err = org.GetVDCByName(d.Get("name").(string), false)
	}
	if err != nil {
		return nil, fmt.Errorf("could not retrieve the instantiated VDC: %s", err)
	}
	return vdc, nil
}
