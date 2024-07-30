package vcd

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

func resourceVcdNsxtOrgVdcNetworkSegmentProfileTemplate() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceVcdNsxtOrgVdcNetworkSegmentProfileCreateUpdate,
		ReadContext:   resourceVcdNsxtOrgVdcNetworkSegmentProfileRead,
		UpdateContext: resourceVcdNsxtOrgVdcNetworkSegmentProfileCreateUpdate,
		DeleteContext: resourceVcdNsxtOrgVdcNetworkSegmentProfileDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVcdNsxtOrgVdcNetworkSegmentProfileImport,
		},

		Schema: map[string]*schema.Schema{
			"org": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Description: "The name of organization to use, optional if defined at provider " +
					"level. Useful when connected as sysadmin working across different organizations",
			},
			"org_network_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "ID of the Organization Network that will have the segment profile",
			},
			// One can set either Segment Profile Template (which is composed of multiple Segment Profiles), or individual Segment Profiles
			"segment_profile_template_id": {
				Type:          schema.TypeString,
				Optional:      true,
				Description:   "Segment Profile Template ID",
				ConflictsWith: []string{"ip_discovery_profile_id", "mac_discovery_profile_id", "spoof_guard_profile_id", "qos_profile_id", "segment_security_profile_id"},
			},
			"segment_profile_template_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Segment Profile Template Name",
			},
			"ip_discovery_profile_id": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				Description:   "NSX-T IP Discovery Profile",
				ConflictsWith: []string{"segment_profile_template_id"},
			},
			"mac_discovery_profile_id": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				Description:   "NSX-T Mac Discovery Profile",
				ConflictsWith: []string{"segment_profile_template_id"},
			},
			"spoof_guard_profile_id": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				Description:   "NSX-T Spoof Guard Profile",
				ConflictsWith: []string{"segment_profile_template_id"},
			},
			"qos_profile_id": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				Description:   "NSX-T QoS Profile",
				ConflictsWith: []string{"segment_profile_template_id"},
			},
			"segment_security_profile_id": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				Description:   "NSX-T Segment Security Profile",
				ConflictsWith: []string{"segment_profile_template_id"},
			},
		},
	}
}

func resourceVcdNsxtOrgVdcNetworkSegmentProfileCreateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	vcdClient.lockParentOrgNetwork(d)
	defer vcdClient.unLockParentOrgNetwork(d)

	org, err := vcdClient.GetOrgFromResource(d)
	if err != nil {
		return diag.Errorf("[Org VDC Network Segment Profile configuration] error retrieving Org: %s", err)
	}

	orgNetworkId := d.Get("org_network_id").(string)
	orgVdcNet, err := org.GetOpenApiOrgVdcNetworkById(orgNetworkId)
	if err != nil {
		return diag.Errorf("[Org VDC Network Segment Profile configuration] error retrieving Org VDC network with ID '%s': %s", orgNetworkId, err)
	}

	if !orgVdcNet.IsNsxt() {
		return diag.Errorf("[Org VDC Network Segment Profile configuration] only NSX-T Org VDC networks support Segment Profiles")
	}

	ipDiscoveryProfileId := d.Get("ip_discovery_profile_id").(string)
	macDiscoveryProfileId := d.Get("mac_discovery_profile_id").(string)
	spoofGuardProfileId := d.Get("spoof_guard_profile_id").(string)
	qosProfileId := d.Get("qos_profile_id").(string)
	segmentSecurityProfileId := d.Get("segment_security_profile_id").(string)

	segmentProfileTemplateId := d.Get("segment_profile_template_id").(string)

	switch {
	// Setting `segment_profile_template_id` requires modifying Org VDC Network structure.
	// It can only be set (PUT/POST) using Org VDC network structure, but cannot be read (GET).
	// To read its value one must use orgVdcNet.GetSegmentProfile() function.
	case segmentProfileTemplateId != "":
		orgVdcNet.OpenApiOrgVdcNetwork.SegmentProfileTemplate = &types.OpenApiReference{ID: segmentProfileTemplateId}
		_, err = orgVdcNet.Update(orgVdcNet.OpenApiOrgVdcNetwork)
		if err != nil {
			return diag.Errorf("[Org VDC Network Segment Profile configuration] error setting Segment Profile Template for Org VDC Network: %s", err)
		}
	case ipDiscoveryProfileId != "" || macDiscoveryProfileId != "" || spoofGuardProfileId != "" || qosProfileId != "" || segmentSecurityProfileId != "":
		// Individual segment profiles should be applied using a dedicated Segment Profile orgVdcNet.UpdateSegmentProfile
		segmentProfileConfig := &types.OrgVdcNetworkSegmentProfiles{
			IPDiscoveryProfile:     &types.Reference{ID: ipDiscoveryProfileId},
			MacDiscoveryProfile:    &types.Reference{ID: macDiscoveryProfileId},
			SpoofGuardProfile:      &types.Reference{ID: spoofGuardProfileId},
			QosProfile:             &types.Reference{ID: qosProfileId},
			SegmentSecurityProfile: &types.Reference{ID: segmentSecurityProfileId},
		}
		_, err = orgVdcNet.UpdateSegmentProfile(segmentProfileConfig)
		if err != nil {
			return diag.Errorf("[Org VDC Network Segment Profile configuration] error configuring Segment Profile for Org VDC Network: %s", err)
		}
	default:
		return diag.Errorf("[Org VDC Network Segment Profile configuration] invalid configuration provided")
	}

	d.SetId(orgVdcNet.OpenApiOrgVdcNetwork.ID)

	return resourceVcdNsxtOrgVdcNetworkSegmentProfileRead(ctx, d, meta)
}

func resourceVcdNsxtOrgVdcNetworkSegmentProfileRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return resourceDataSourceVcdNsxtOrgVdcNetworkSegmentProfileRead(ctx, d, meta, "resource")
}

func resourceDataSourceVcdNsxtOrgVdcNetworkSegmentProfileRead(ctx context.Context, d *schema.ResourceData, meta interface{}, origin string) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	org, err := vcdClient.GetOrgFromResource(d)
	if err != nil {
		return diag.Errorf("[Org VDC Network Segment Profile configuration read] error retrieving Org: %s", err)
	}

	orgNetworkId := d.Get("org_network_id").(string)
	orgVdcNet, err := org.GetOpenApiOrgVdcNetworkById(orgNetworkId)
	if err != nil {
		if origin == "resource" && govcd.ContainsNotFound(err) {
			d.SetId("")
			return nil
		}
		return diag.Errorf("[Org VDC Network Segment Profile configuration read] error retrieving Org VDC network with ID '%s': %s", orgNetworkId, err)
	}

	segmentProfileConfig, err := orgVdcNet.GetSegmentProfile()
	if err != nil {
		return diag.Errorf("[Org VDC Network Segment Profile configuration read] error retrieving Segment Profile configuration for Org VDC Network: %s", err)
	}

	dSet(d, "segment_profile_template_name", "")
	dSet(d, "segment_profile_template_id", "")
	if segmentProfileConfig.SegmentProfileTemplate != nil && segmentProfileConfig.SegmentProfileTemplate.TemplateRef != nil {
		dSet(d, "segment_profile_template_id", segmentProfileConfig.SegmentProfileTemplate.TemplateRef.ID)
		dSet(d, "segment_profile_template_name", segmentProfileConfig.SegmentProfileTemplate.TemplateRef.Name)
	}

	dSet(d, "ip_discovery_profile_id", "")
	if segmentProfileConfig.IPDiscoveryProfile != nil {
		dSet(d, "ip_discovery_profile_id", segmentProfileConfig.IPDiscoveryProfile.ID)
	}

	dSet(d, "mac_discovery_profile_id", "")
	if segmentProfileConfig.MacDiscoveryProfile != nil {
		dSet(d, "mac_discovery_profile_id", segmentProfileConfig.MacDiscoveryProfile.ID)
	}

	dSet(d, "spoof_guard_profile_id", "")
	if segmentProfileConfig.SpoofGuardProfile != nil {
		dSet(d, "spoof_guard_profile_id", segmentProfileConfig.SpoofGuardProfile.ID)
	}

	dSet(d, "qos_profile_id", "")
	if segmentProfileConfig.QosProfile != nil {
		dSet(d, "qos_profile_id", segmentProfileConfig.QosProfile.ID)
	}

	dSet(d, "segment_security_profile_id", "")
	if segmentProfileConfig.SegmentSecurityProfile != nil {
		dSet(d, "segment_security_profile_id", segmentProfileConfig.SegmentSecurityProfile.ID)
	}

	d.SetId(orgVdcNet.OpenApiOrgVdcNetwork.ID)

	return nil
}

func resourceVcdNsxtOrgVdcNetworkSegmentProfileDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	vcdClient.lockParentOrgNetwork(d)
	defer vcdClient.unLockParentOrgNetwork(d)

	org, err := vcdClient.GetOrgFromResource(d)
	if err != nil {
		return diag.Errorf("[Org VDC Network Segment Profile configuration delete] error retrieving Org: %s", err)
	}

	orgNetworkId := d.Get("org_network_id").(string)

	orgVdcNet, err := org.GetOpenApiOrgVdcNetworkById(orgNetworkId)
	if err != nil {
		return diag.Errorf("[Org VDC Network Segment Profile configuration delete] error retrieving Org VDC network with ID '%s': %s", orgNetworkId, err)
	}

	// Attempt to remove Segment Profile Template using main network structure (it is the only way, if it is set)
	if orgVdcNet.OpenApiOrgVdcNetwork != nil && orgVdcNet.OpenApiOrgVdcNetwork.SegmentProfileTemplate != nil {
		orgVdcNet.OpenApiOrgVdcNetwork.SegmentProfileTemplate = &types.OpenApiReference{}
		_, err := orgVdcNet.Update(orgVdcNet.OpenApiOrgVdcNetwork)
		if err != nil {
			return diag.Errorf("[Org VDC Network Segment Profile configuration delete] error reseting Segment Profile Template ID for Org VDC Network: %s", err)
		}
	}

	// Attempt to cleanup any custom segment profiles
	_, err = orgVdcNet.UpdateSegmentProfile(&types.OrgVdcNetworkSegmentProfiles{})
	if err != nil {
		return diag.Errorf("[Org VDC Network Segment Profile configuration delete] error reseting Segment Profile: %s", err)
	}

	return nil
}

func resourceVcdNsxtOrgVdcNetworkSegmentProfileImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	resourceURI := strings.Split(d.Id(), ImportSeparator)
	if len(resourceURI) != 3 {
		return nil, fmt.Errorf("resource name must be specified as org-name.vdc-org-vdc-group-name.org_network_name")
	}
	orgName, vdcOrVdcGroupName, orgVdcNetworkName := resourceURI[0], resourceURI[1], resourceURI[2]

	vcdClient := meta.(*VCDClient)
	vdcOrVdcGroup, err := lookupVdcOrVdcGroup(vcdClient, orgName, vdcOrVdcGroupName)
	if err != nil {
		return nil, err
	}

	if !vdcOrVdcGroup.IsNsxt() {
		return nil, fmt.Errorf("[Org VDC Network Segment Profile configuration import] Segment Profile configuration is only supported for NSX-T networks: %s", err)
	}

	orgVdcNet, err := vdcOrVdcGroup.GetOpenApiOrgVdcNetworkByName(orgVdcNetworkName)
	if err != nil {
		return nil, fmt.Errorf("[Org VDC Network Segment Profile configuration import] error retrieving Org VDC network with name '%s': %s", orgVdcNetworkName, err)
	}

	dSet(d, "org", orgName)
	dSet(d, "org_network_id", orgVdcNet.OpenApiOrgVdcNetwork.ID)
	d.SetId(orgVdcNet.OpenApiOrgVdcNetwork.ID)

	return []*schema.ResourceData{d}, nil
}
