package vcd

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdNsxtIpDiscoveryProfile() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceNsxtIpDiscoveryProfileRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Description of Segment IP Discovery Profile",
			},
			"context_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "ID of VDC, VDC Group, or NSX-T Manager. Required if the VCD instance has more than one NSX-T manager",
			},

			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Description of Segment IP Discovery Profile",
			},
		},
	}
}

func datasourceNsxtIpDiscoveryProfileRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	profileName := d.Get("name").(string)
	contextUrn := d.Get("context_id").(string)

	contextFilterField, err := getContextFilterField(contextUrn)
	if err != nil {
		return diag.FromErr(err)
	}

	queryFilter := url.Values{}
	queryFilter.Add("filter", fmt.Sprintf("%s==%s", contextFilterField, contextUrn))

	ipDiscoveryProfile, err := vcdClient.GetIpDiscoveryProfileByName(profileName, queryFilter)
	if err != nil {
		return diag.Errorf("could not find IP Discovery Profile by name '%s': %s", profileName, err)
	}

	d.SetId(ipDiscoveryProfile.ID)

	return nil
}

// getContextFilterField determines which field should be used for filtering
func getContextFilterField(urn string) (string, error) {
	contextFilterField := ""
	switch {
	case strings.Contains(urn, "urn:vcloud:nsxtmanager:"):
		contextFilterField = "nsxTManagerRef.id"
	case strings.Contains(urn, "urn:vcloud:vdcGroup:"):
		contextFilterField = "vdcGroupId"
	case strings.Contains(urn, "urn:vcloud:vdc:"):
		contextFilterField = "orgVdcId"
	default:
		return "", fmt.Errorf("unrecognized 'context_id', was expecting to get NSX-T Manager, VDC or VDC Group, got '%s'", urn)
	}

	return contextFilterField, nil

}
