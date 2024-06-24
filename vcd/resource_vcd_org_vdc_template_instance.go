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
		UpdateContext: resourceVcdVdcTemplateInstantiateUpdate,
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
			"delete_instantiated_vdc_on_removal": {
				Type:        schema.TypeBool,
				Required:    true,
				Description: "If this flag is set to 'true', removing this resource will attempt to delete the instantiated VDC",
			},
			"delete_force": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "If this flag is set to 'true', it forcefully deletes the VDC, only when delete_instantiated_vdc_on_removal=true",
			},
			"delete_recursive": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "If this flag is set to 'true', it recursively deletes the VDC, only when delete_instantiated_vdc_on_removal=true",
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
		return diag.Errorf("failed instantiating the VDC Template: %s", err)
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

func resourceVcdVdcTemplateInstantiateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// No-op. This is needed as "delete_instantiated_vdc_on_removal", "delete_force" and "delete_recursive"
	// are not marked as "ForceNew: true" (they can be modified after creation), but they are just flags, not obtained from
	// VCD.
	return resourceVcdVdcTemplateRead(ctx, d, meta)
}

func resourceVcdVdcTemplateInstantiateDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	if !d.Get("delete_instantiated_vdc_on_removal").(bool) {
		return nil
	}

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

	err = vdc.DeleteWait(d.Get("delete_force").(bool), d.Get("delete_recursive").(bool))
	if err != nil {
		return diag.Errorf("failed deleting instantiated VDC '%s': %s", vdc.Vdc.ID, err)
	}

	return nil
}
