package vcd

import (
	"context"
	"fmt"
	"net/url"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdNsxtSegmentSecurityProfile() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceNsxtSegmentSecurityProfileRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Description of Segment Security Profile",
			},
			"context_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "ID of VDC, VDC Group, or NSX-T Manager. Required if the VCD instance has more than one NSX-T manager",
			},

			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Description of Segment Security Profile",
			},
		},
	}
}

func datasourceNsxtSegmentSecurityProfileRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	profileName := d.Get("name").(string)
	contextUrn := d.Get("context_id").(string)

	contextFilterField, err := getContextFilterField(contextUrn)
	if err != nil {
		return diag.FromErr(err)
	}

	queryFilter := url.Values{}
	queryFilter.Add("filter", fmt.Sprintf("%s==%s", contextFilterField, contextUrn))

	segmentSecurityProfile, err := vcdClient.GetSegmentSecurityProfileByName(profileName, queryFilter)
	if err != nil {
		return diag.Errorf("could not find Segment Security Profile by name '%s': %s", profileName, err)
	}

	d.SetId(segmentSecurityProfile.ID)

	return nil
}
