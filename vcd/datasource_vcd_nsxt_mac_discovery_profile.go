package vcd

import (
	"context"
	"fmt"
	"net/url"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdNsxtMacDiscoveryProfile() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceNsxtMacDiscoveryProfileRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Description of Segment MAC Discovery Profile",
			},
			"context_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "ID of VDC, VDC Group, or NSX-T Manager. Required if the VCD instance has more than one NSX-T manager",
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Description of Segment MAC Discovery Profile",
			},
			"is_mac_change_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Indcates whether source MAC address change is enabled",
			},
			"is_mac_learning_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Indicates whether source MAC address learning is enabled",
			},
			"is_unknown_unicast_flooding_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Indicates whether unknown unicast flooding rule is enabled",
			},
			"mac_learning_aging_time": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Indicates aging time in seconds for learned MAC address",
			},
			"mac_limit": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Indicates the maximum number of MAC addresses that can be learned on this port",
			},
			"mac_policy": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Defines the policy after MAC Limit is exceeded. It can be either 'ALLOW' or 'DROP'",
			},
		},
	}
}

func datasourceNsxtMacDiscoveryProfileRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	profileName := d.Get("name").(string)
	contextUrn := d.Get("context_id").(string)

	contextFilterField, err := getContextFilterField(contextUrn)
	if err != nil {
		return diag.FromErr(err)
	}

	queryFilter := url.Values{}
	queryFilter.Add("filter", fmt.Sprintf("%s==%s", contextFilterField, contextUrn))

	macDiscoveryProfile, err := vcdClient.GetMacDiscoveryProfileByName(profileName, queryFilter)
	if err != nil {
		return diag.Errorf("could not find MAC Discovery Profile by name '%s': %s", profileName, err)
	}

	dSet(d, "description", macDiscoveryProfile.Description)
	dSet(d, "is_mac_change_enabled", macDiscoveryProfile.IsMacChangeEnabled)
	dSet(d, "is_mac_learning_enabled", macDiscoveryProfile.IsMacLearningEnabled)
	dSet(d, "is_unknown_unicast_flooding_enabled", macDiscoveryProfile.IsUnknownUnicastFloodingEnabled)
	dSet(d, "mac_learning_aging_time", macDiscoveryProfile.MacLearningAgingTime)
	dSet(d, "mac_limit", macDiscoveryProfile.MacLimit)
	dSet(d, "mac_policy", macDiscoveryProfile.MacPolicy)

	d.SetId(macDiscoveryProfile.ID)

	return nil
}
