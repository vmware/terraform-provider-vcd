package vcd

import (
	"context"
	"fmt"

	"github.com/vmware/go-vcloud-director/v3/govcd"

	"github.com/vmware/go-vcloud-director/v3/types/v56"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceVcdSegmentProfileTemplate() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdSegmentProfileTemplateCreate,
		ReadContext:   resourceVcdSegmentProfileTemplateRead,
		UpdateContext: resourceVcdSegmentProfileTemplateUpdate,
		DeleteContext: resourceVcdSegmentProfileTemplateDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdSegmentProfileTemplateImport,
		},

		Schema: map[string]*schema.Schema{
			"nsxt_manager_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "NSX-T Manager ID",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of Segment Profile Template",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of Segment Profile Template",
			},
			"ip_discovery_profile_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Segment IP Discovery Profile ID",
			},
			"mac_discovery_profile_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Segment MAC Discovery Profile ID",
			},
			"spoof_guard_profile_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Segment Spoof Guard Profile ID",
			},
			"qos_profile_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Segment QoS Profile ID",
			},
			"segment_security_profile_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Segment Security Profile ID",
			},
		},
	}
}

func resourceVcdSegmentProfileTemplateCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

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

	spt, err := vcdClient.GetSegmentProfileTemplateById(d.Id())
	if err != nil {
		return diag.Errorf("unable to find NSX-T Segment Profile Template: %s", err)
	}

	updateSegmentProfileTemplateConfig := getNsxtSegmentProfileTemplateType(d)
	updateSegmentProfileTemplateConfig.ID = d.Id()
	_, err = spt.Update(updateSegmentProfileTemplateConfig)
	if err != nil {
		return diag.Errorf("error updating NSX-T Segment Profile Template: %s", err)
	}

	return resourceVcdSegmentProfileTemplateRead(ctx, d, meta)
}

func resourceVcdSegmentProfileTemplateRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	spt, err := vcdClient.GetSegmentProfileTemplateById(d.Id())
	if err != nil {
		if govcd.ContainsNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.Errorf("unable to find NSX-T Segment Profile Template: %s", err)
	}

	setNsxtSegmentProfileTemplateData(d, spt.NsxtSegmentProfileTemplate)

	return nil
}

func resourceVcdSegmentProfileTemplateDelete(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)

	spt, err := vcdClient.GetSegmentProfileTemplateById(d.Id())
	if err != nil {
		return diag.Errorf("unable to find NSX-T Segment Profile Template: %s", err)
	}

	err = spt.Delete()
	if err != nil {
		return diag.Errorf("error deleting NSX-T Segment Profile Template: %s", err)
	}

	return nil
}

func resourceVcdSegmentProfileTemplateImport(_ context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	vcdClient := meta.(*VCDClient)

	resourceURI := d.Id()
	spt, err := vcdClient.GetSegmentProfileTemplateByName(resourceURI)
	if err != nil {
		return nil, fmt.Errorf("error finding NSX-T Segment Profile Template with Name '%s': %s", d.Id(), err)
	}

	d.SetId(spt.NsxtSegmentProfileTemplate.ID)
	return []*schema.ResourceData{d}, nil
}

func getNsxtSegmentProfileTemplateType(d *schema.ResourceData) *types.NsxtSegmentProfileTemplate {

	config := &types.NsxtSegmentProfileTemplate{
		Name:                   d.Get("name").(string),
		Description:            d.Get("description").(string),
		IPDiscoveryProfile:     &types.Reference{ID: d.Get("ip_discovery_profile_id").(string)},
		MacDiscoveryProfile:    &types.Reference{ID: d.Get("mac_discovery_profile_id").(string)},
		QosProfile:             &types.Reference{ID: d.Get("qos_profile_id").(string)},
		SegmentSecurityProfile: &types.Reference{ID: d.Get("segment_security_profile_id").(string)},
		SpoofGuardProfile:      &types.Reference{ID: d.Get("spoof_guard_profile_id").(string)},
		SourceNsxTManagerRef:   &types.OpenApiReference{ID: d.Get("nsxt_manager_id").(string)},
	}

	return config
}

func setNsxtSegmentProfileTemplateData(d *schema.ResourceData, config *types.NsxtSegmentProfileTemplate) {
	dSet(d, "name", config.Name)
	dSet(d, "description", config.Description)

	dSet(d, "nsxt_manager_id", "")
	if config.SourceNsxTManagerRef != nil {
		dSet(d, "nsxt_manager_id", config.SourceNsxTManagerRef.ID)
	}

	dSet(d, "ip_discovery_profile_id", "")
	if config.IPDiscoveryProfile != nil {
		dSet(d, "ip_discovery_profile_id", config.IPDiscoveryProfile.ID)
	}

	dSet(d, "mac_discovery_profile_id", "")
	if config.MacDiscoveryProfile != nil {
		dSet(d, "mac_discovery_profile_id", config.MacDiscoveryProfile.ID)
	}

	dSet(d, "qos_profile_id", "")
	if config.QosProfile != nil {
		dSet(d, "qos_profile_id", config.QosProfile.ID)
	}

	dSet(d, "segment_security_profile_id", "")
	if config.SegmentSecurityProfile != nil {
		dSet(d, "segment_security_profile_id", config.SegmentSecurityProfile.ID)
	}

	dSet(d, "spoof_guard_profile_id", "")
	if config.SpoofGuardProfile != nil {
		dSet(d, "spoof_guard_profile_id", config.SpoofGuardProfile.ID)
	}

}
