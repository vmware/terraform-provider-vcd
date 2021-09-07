package vcd

import (
	"context"
	"fmt"

	"github.com/vmware/go-vcloud-director/v2/govcd"

	"github.com/vmware/go-vcloud-director/v2/types/v56"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceVcdAlbController() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdAlbControllerCreate,
		ReadContext:   resourceVcdAlbControllerRead,
		UpdateContext: resourceVcdAlbControllerUpdate,
		DeleteContext: resourceVcdAlbControllerDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdAlbControllerImport,
		},

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "NSX-T ALB Controller name",
			},
			"url": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "NSX-T ALB Controller URL",
			},
			"username": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "NSX-T ALB Controller Username",
			},
			"password": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				Description: "NSX-T ALB Controller Password",
			},
			"description": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "NSX-T ALB Controller description",
			},
			"license_type": &schema.Schema{
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice([]string{"BASIC", "ENTERPRISE"}, false),
				Description:  "NSX-T ALB License type. One of 'BASIC', 'ENTERPRISE'",
			},
			"version": &schema.Schema{
				Type:        schema.TypeString,
				Computed:    true,
				Description: "NSX-T ALB Controller version",
			},
		},
	}
}

func resourceVcdAlbControllerCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	if !vcdClient.Client.IsSysAdmin {
		return diag.Errorf("this resource is only supported for Providers")
	}

	albControllerConfig := getNsxtAlbControllerType(d)
	createdAlbController, err := vcdClient.CreateNsxtAlbController(albControllerConfig)
	if err != nil {
		return diag.Errorf("error creating NSX-T ALB Controller '%s': %s", albControllerConfig.Name, err)
	}

	d.SetId(createdAlbController.NsxtAlbController.ID)

	return resourceVcdAlbControllerRead(ctx, d, meta)
}

func resourceVcdAlbControllerUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	if !vcdClient.Client.IsSysAdmin {
		return diag.Errorf("this resource is only supported for Providers")
	}

	albController, err := vcdClient.GetAlbControllerById(d.Id())
	if err != nil {
		return diag.Errorf("unable to find NSX-T ALB Controller: %s", err)
	}

	updateAlbControllerConfig := getNsxtAlbControllerType(d)
	updateAlbControllerConfig.ID = d.Id()
	_, err = albController.Update(updateAlbControllerConfig)
	if err != nil {
		return diag.Errorf("error updating NSX-T ALB Controller: %s", err)
	}

	return resourceVcdAlbControllerRead(ctx, d, meta)
}

func resourceVcdAlbControllerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	if !vcdClient.Client.IsSysAdmin {
		return diag.Errorf("this resource is only supported for Providers")
	}

	albController, err := vcdClient.GetAlbControllerById(d.Id())
	if err != nil {
		if govcd.ContainsNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.Errorf("unable to find NSX-T ALB Controller: %s", err)
	}

	setNsxtAlbControllerData(d, albController.NsxtAlbController)

	return nil
}

func resourceVcdAlbControllerDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	if !vcdClient.Client.IsSysAdmin {
		return diag.Errorf("this resource is only supported for Providers")
	}

	albController, err := vcdClient.GetAlbControllerById(d.Id())
	if err != nil {
		return diag.Errorf("unable to find NSX-T ALB Controller: %s", err)
	}

	err = albController.Delete()
	if err != nil {
		return diag.Errorf("error deleting NSX-T ALB Controller: %s", err)
	}

	return nil
}

func resourceVcdAlbControllerImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	vcdClient := meta.(*VCDClient)
	if !vcdClient.Client.IsSysAdmin {
		return nil, fmt.Errorf("this resource is only supported for Providers")
	}

	resourceURI := d.Id()
	albController, err := vcdClient.GetAlbControllerByName(resourceURI)
	if err != nil {
		return nil, fmt.Errorf("error finding NSX-T ALB Controller with Name '%s': %s", d.Id(), err)
	}

	d.SetId(albController.NsxtAlbController.ID)
	return []*schema.ResourceData{d}, nil
}

func getNsxtAlbControllerType(d *schema.ResourceData) *types.NsxtAlbController {
	albControllerType := &types.NsxtAlbController{
		Name:        d.Get("name").(string),
		Description: d.Get("description").(string),
		Url:         d.Get("url").(string),
		Username:    d.Get("username").(string),
		Password:    d.Get("password").(string),
		LicenseType: d.Get("license_type").(string),
	}

	return albControllerType
}

func setNsxtAlbControllerData(d *schema.ResourceData, albController *types.NsxtAlbController) {
	_ = d.Set("name", albController.Name)
	_ = d.Set("description", albController.Description)
	_ = d.Set("url", albController.Url)
	_ = d.Set("username", albController.Username)
	_ = d.Set("license_type", albController.LicenseType)
	_ = d.Set("version", albController.Version)
}
