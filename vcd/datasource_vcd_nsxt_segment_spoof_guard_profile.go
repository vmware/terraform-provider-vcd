package vcd

import (
	"context"
	"fmt"
	"net/url"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdNsxtSegmentSpoofGuardProfile() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceNsxtSegmentSpoofGuardProfileRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of Segment Spoof Guard Profile",
			},
			"nsxt_manager_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"nsxt_manager_id", "vdc_id", "vdc_group_id"},
				Description:  "ID of NSX-T Manager",
			},
			"vdc_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"nsxt_manager_id", "vdc_id", "vdc_group_id"},
				Description:  "ID of VDC",
			},
			"vdc_group_id": {
				Type:         schema.TypeString,
				Optional:     true,
				ExactlyOneOf: []string{"nsxt_manager_id", "vdc_id", "vdc_group_id"},
				Description:  "ID of VDC Group",
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Description of Segment Spoof Guard Profile",
			},
			"is_address_binding_whitelist_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Indicates whether Spoof Guard is enabled",
			},
		},
	}
}

func datasourceNsxtSegmentSpoofGuardProfileRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	profileName := d.Get("name").(string)

	contextFilterField, contextUrn, err := getContextFilterField(d)
	if err != nil {
		return diag.FromErr(err)
	}

	queryFilter := url.Values{}
	queryFilter.Add("filter", fmt.Sprintf("%s==%s", contextFilterField, contextUrn))

	spoofGuardProfile, err := vcdClient.GetSpoofGuardProfileByName(profileName, queryFilter)
	if err != nil {
		return diag.Errorf("could not find Spoof Guard Profile by name '%s': %s", profileName, err)
	}

	dSet(d, "description", spoofGuardProfile.Description)
	dSet(d, "is_address_binding_whitelist_enabled", spoofGuardProfile.IsAddressBindingWhitelistEnabled)

	d.SetId(spoofGuardProfile.ID)

	return nil
}
