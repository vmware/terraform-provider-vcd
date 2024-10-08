package vcd

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func resourceVcdVcenter() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdVcenterCreate,
		ReadContext:   resourceVcdVcenterRead,
		UpdateContext: resourceVcdVcenterUpdate,
		DeleteContext: resourceVcdVcenterDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdVcenterImport,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of vCenter.",
			},
			"vcenter_version": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "vCenter version",
			},
			"vcenter_host": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "vCenter hostname",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "vCenter status",
			},
			"is_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "vCenter version",
			},
			// In UI this field is called `connection`, but it is a reserved field in Terraform
			"connection_status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "vCenter connection state",
			},
		},
	}
}

func resourceVcdVcenterCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	t, err := getVcenterType(d)
	if err != nil {
		return diag.Errorf("error getting vCenter type: %s")
	}

	vcenter, err := vcdClient.CreateVcenter(t)
	if err != nil {
		return diag.Errorf("error creating vCenter: %s")
	}

	d.SetId(vcenter.VSphereVCenter.VcId)
	return resourceVcdVcenterRead(ctx, d, meta)
}

func resourceVcdVcenterUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	t, err := getVcenterType(d)
	if err != nil {
		return diag.Errorf("error getting vCenter type: %s")
	}

	vcenter, err := vcdClient.GetVCenterById(d.Id())
	if err != nil {
		return diag.Errorf("error retrieving vCenter: %s")
	}

	_, err = vcenter.Update(t)
	if err != nil {
		return diag.Errorf("error updating vCenter: %s", err)
	}

	return resourceVcdVcenterRead(ctx, d, meta)
}

func resourceVcdVcenterRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	t, err := getVcenterType(d)
	if err != nil {
		return diag.Errorf("error getting vCenter type: %s")
	}

	vcenter, err := vcdClient.GetVCenterById(d.Id())
	if err != nil {
		return diag.Errorf("error retrieving vCenter: %s")
	}

	setVCenterData(d, vcenter.VSphereVCenter)

	return nil
}

func resourceVcdVcenterDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	vcenter, err := vcdClient.GetVCenterById(d.Id())
	if err != nil {
		return diag.Errorf("error retrieving vCenter: %s", err)
	}

	err = vcenter.Delete()
	if err != nil {
		return diag.Errorf("error deleting vCenter: %s", err)
	}

	return nil
}

func resourceVcdVcenterImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	vcdClient := meta.(*VCDClient)

	vcenter, err := vcdClient.GetVCenterByName(d.Id())
	if err != nil {
		return nil, fmt.Errorf("error retrieving vCenter '%s': %s", d.Id(), err)
	}
	d.SetId(vcenter.VSphereVCenter.VcId)
	return []*schema.ResourceData{d}, nil
}

func getVcenterType(d *schema.ResourceData) (*types.VSphereVirtualCenter, error) {

	return nil, nil
}

func setVCenterData(d *schema.ResourceData, v *types.VSphereVirtualCenter) error {

	return nil
}
