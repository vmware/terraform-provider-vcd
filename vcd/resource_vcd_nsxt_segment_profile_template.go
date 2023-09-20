package vcd

import (
	"context"

	"github.com/vmware/go-vcloud-director/v2/govcd"

	"github.com/vmware/go-vcloud-director/v2/types/v56"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceVcdSegmentProfileTemplate() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdSegmentProfileTemplateCreate,
		ReadContext:   resourceVcdSegmentProfileTemplateRead,
		UpdateContext: resourceVcdSegmentProfileTemplateUpdate,
		DeleteContext: resourceVcdSegmentProfileTemplateDelete,
		// Importer: &schema.ResourceImporter{
		// 	StateContext: resourceVcdSegmentProfileTemplateImport,
		// },

		Schema: map[string]*schema.Schema{
			"nsxt_manager_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "NSX-T Segment Profile Template name",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "NSX-T Segment Profile Template name",
			},
			"description": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "NSX-T Segment Profile Template name",
			},

			"ip_discovery_profile_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "NSX-T Segment Profile Template name",
			},
			"mac_discovery_profile_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "NSX-T Segment Profile Template name",
			},
			"spoof_guard_profile_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "NSX-T Segment Profile Template name",
			},
			"qos_profile_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "NSX-T Segment Profile Template name",
			},
			"segment_security_profile_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "NSX-T Segment Profile Template name",
			},
		},
	}
}

func resourceVcdSegmentProfileTemplateCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	if !vcdClient.Client.IsSysAdmin {
		return diag.Errorf("this resource is only supported for Providers")
	}

	// nsxtManager, err := vcd.client.GetNsxtManagerByName(vcd.config.VCD.Nsxt.Manager)
	// check.Assert(err, IsNil)
	// check.Assert(nsxtManager, NotNil)

	segmentProfileTemplateCfg := getNsxtSegmentProfileTemplateType(d)
	createdSegmentProfileTemplate, err := vcdClient.CreateSegmentProfileTemplate(segmentProfileTemplateCfg)
	if err != nil {
		return diag.Errorf("error creating NSX-T Segment Profile Template '%s': %s", segmentProfileTemplateCfg.Name, err)
	}

	d.SetId(createdSegmentProfileTemplate.NsxtSegmentProfileTemplate.ID)

	return resourceVcdSegmentProfileTemplateRead(ctx, d, meta)
}

func resourceVcdSegmentProfileTemplateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	if !vcdClient.Client.IsSysAdmin {
		return diag.Errorf("this resource is only supported for Providers")
	}

	albController, err := vcdClient.GetSegmentProfileTemplateById(d.Id())
	if err != nil {
		return diag.Errorf("unable to find NSX-T Segment Profile Template: %s", err)
	}

	updateSegmentProfileTemplateConfig := getNsxtSegmentProfileTemplateType(d)
	updateSegmentProfileTemplateConfig.ID = d.Id()
	_, err = albController.Update(updateSegmentProfileTemplateConfig)
	if err != nil {
		return diag.Errorf("error updating NSX-T Segment Profile Template: %s", err)
	}

	return resourceVcdSegmentProfileTemplateRead(ctx, d, meta)
}

func resourceVcdSegmentProfileTemplateRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	if !vcdClient.Client.IsSysAdmin {
		return diag.Errorf("this resource is only supported for Providers")
	}

	albController, err := vcdClient.GetSegmentProfileTemplateById(d.Id())
	if err != nil {
		if govcd.ContainsNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.Errorf("unable to find NSX-T Segment Profile Template: %s", err)
	}

	setNsxtSegmentProfileTemplateData(d, albController.NsxtSegmentProfileTemplate)

	return nil
}

func resourceVcdSegmentProfileTemplateDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	if !vcdClient.Client.IsSysAdmin {
		return diag.Errorf("this resource is only supported for Providers")
	}

	albController, err := vcdClient.GetSegmentProfileTemplateById(d.Id())
	if err != nil {
		return diag.Errorf("unable to find NSX-T Segment Profile Template: %s", err)
	}

	err = albController.Delete()
	if err != nil {
		return diag.Errorf("error deleting NSX-T Segment Profile Template: %s", err)
	}

	return nil
}

// func resourceVcdSegmentProfileTemplateImport(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
// 	vcdClient := meta.(*VCDClient)
// 	if !vcdClient.Client.IsSysAdmin {
// 		return nil, fmt.Errorf("this resource is only supported for Providers")
// 	}

// 	resourceURI := d.Id()
// 	albController, err := vcdClient.GetSegmentProfileTemplateByName(resourceURI)
// 	if err != nil {
// 		return nil, fmt.Errorf("error finding NSX-T Segment Profile Template with Name '%s': %s", d.Id(), err)
// 	}

// 	d.SetId(albController.NsxtSegmentProfileTemplate.ID)
// 	return []*schema.ResourceData{d}, nil
// }

func getNsxtSegmentProfileTemplateType(d *schema.ResourceData) *types.NsxtSegmentProfileTemplate {

	config := &types.NsxtSegmentProfileTemplate{
		Name:                   d.Get("name").(string),
		Description:            d.Get("description").(string),
		IPDiscoveryProfile:     &types.NsxtSegmentProfileTemplateReference{ID: d.Get("ip_discovery_profile_id").(string)},
		MacDiscoveryProfile:    &types.NsxtSegmentProfileTemplateReference{ID: d.Get("mac_discovery_profile_id").(string)},
		QosProfile:             &types.NsxtSegmentProfileTemplateReference{ID: d.Get("qos_profile_id").(string)},
		SegmentSecurityProfile: &types.NsxtSegmentProfileTemplateReference{ID: d.Get("segment_security_profile_id").(string)},
		SpoofGuardProfile:      &types.NsxtSegmentProfileTemplateReference{ID: d.Get("spoof_guard_profile_id").(string)},
		SourceNsxTManagerRef:   &types.OpenApiReference{ID: d.Get("nsxt_manager_id").(string)},
	}

	return config
}

func setNsxtSegmentProfileTemplateData(d *schema.ResourceData, config *types.NsxtSegmentProfileTemplate) {
	dSet(d, "name", config.Name)
	dSet(d, "description", config.Description)
	// dSet(d, "url", albController.Url)
	// dSet(d, "username", albController.Username)
	// dSet(d, "license_type", albController.LicenseType)
	// dSet(d, "version", albController.Version)
}
