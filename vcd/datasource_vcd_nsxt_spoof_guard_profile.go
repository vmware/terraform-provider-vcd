package vcd

import (
	"context"
	"fmt"
	"net/url"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdNsxtSpoofGuardProfile() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceNsxtSpoofGuardProfileRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Description of Segment Spoof Guard Profile",
			},
			"context_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "ID of VDC, VDC Group, or NSX-T Manager. Required if the VCD instance has more than one NSX-T manager",
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

func datasourceNsxtSpoofGuardProfileRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	profileName := d.Get("name").(string)
	contextUrn := d.Get("context_id").(string)

	contextFilterField, err := getContextFilterField(contextUrn)
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
