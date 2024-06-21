package vcd

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"log"
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
		return diag.Errorf("could not retrieve the VDC Template: %s", err)
	}
	vdc, err := vdcTemplate.InstantiateVdc(d.Get("name").(string), d.Get("description").(string), d.Get("org_id").(string))
	if err != nil {
		diag.Errorf("failed instantiating the VDC Template: %s", err)
	}
	d.SetId(vdc.Vdc.ID)
	return resourceVcdVdcTemplateInstantiateRead(ctx, d, meta)
}

func resourceVcdVdcTemplateInstantiateRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if d.Id() == "" {
		log.Printf("[INFO] unable to find instantiated VDC")
		return nil
	}
	vcdClient := meta.(*VCDClient)
	org, err := vcdClient.GetOrgById(d.Get("org_id").(string))
	if err != nil {
		return diag.Errorf("could not retrieve the Organization of the instantiated VDC: %s", err)
	}

	vdc, err := org.GetVDCById(d.Id(), false)
	if err != nil {
		return diag.Errorf("could not retrieve the instantiated VDC: %s", err)
	}
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(vdc.Vdc.ID)
	return nil
}

func resourceVcdVdcTemplateInstantiateDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	org, err := vcdClient.GetOrgById(d.Get("org_id").(string))
	if err != nil {
		return diag.Errorf("could not retrieve the Organization of the instantiated VDC: %s", err)
	}

	vdc, err := org.GetVDCById(d.Id(), false)
	if err != nil {
		if govcd.ContainsNotFound(err) {
			// The VDC is already gone
			return nil
		}
		return diag.Errorf("could not retrieve the instantiated VDC: %s", err)
	}

	err = vdc.DeleteWait(true, true)
	if err != nil {
		return diag.Errorf("failed deleting instantiated VDC '%s': %s", vdc.Vdc.ID, err)
	}
	return nil
}
