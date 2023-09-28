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
			"arp_binding_limit": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Indicates the number of arp snooped IP addresses to be remembered per logical port",
			},
			"arp_binding_timeout": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Indicates ARP and ND cache timeout (in minutes)",
			},
			"is_arp_snooping_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Defines whether ARP snooping is enabled",
			},
			"is_dhcp_snooping_v4_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Defines whether DHCP snooping for IPv4 is enabled",
			},
			"is_dhcp_snooping_v6_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Defines whether DHCP snooping for IPv6 is enabled",
			},
			"is_duplicate_ip_detection_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Indicates whether duplicate IP detection is enabled",
			},
			"is_nd_snooping_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Indicates whether neighbor discovery (ND) snooping is enabled",
			},
			"is_tofu_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Defines whether 'Trust on First Use(TOFU)' paradigm is enabled",
			},
			"is_vmtools_v4_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Indicates whether fetching IPv4 address using vm-tools is enabled",
			},
			"is_vmtools_v6_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Indicates whether fetching IPv6 address using vm-tools is enabled",
			},
			"nd_snooping_limit": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Maximum number of Neighbor Discovery (ND) snooped IPv6 addresses",
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

	dSet(d, "description", ipDiscoveryProfile.Description)
	dSet(d, "arp_binding_limit", ipDiscoveryProfile.ArpBindingLimit)
	dSet(d, "arp_binding_timeout", ipDiscoveryProfile.ArpNdBindingTimeout)
	dSet(d, "is_arp_snooping_enabled", ipDiscoveryProfile.IsArpSnoopingEnabled)
	dSet(d, "is_dhcp_snooping_v4_enabled", ipDiscoveryProfile.IsDhcpSnoopingV4Enabled)
	dSet(d, "is_dhcp_snooping_v6_enabled", ipDiscoveryProfile.IsDhcpSnoopingV6Enabled)
	dSet(d, "is_duplicate_ip_detection_enabled", ipDiscoveryProfile.IsDuplicateIPDetectionEnabled)
	dSet(d, "is_nd_snooping_enabled", ipDiscoveryProfile.IsNdSnoopingEnabled)
	dSet(d, "is_tofu_enabled", ipDiscoveryProfile.IsTofuEnabled)
	dSet(d, "is_vmtools_v4_enabled", ipDiscoveryProfile.IsVMToolsV4Enabled)
	dSet(d, "is_vmtools_v6_enabled", ipDiscoveryProfile.IsVMToolsV6Enabled)
	dSet(d, "nd_snooping_limit", ipDiscoveryProfile.IsNdSnoopingEnabled)

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
