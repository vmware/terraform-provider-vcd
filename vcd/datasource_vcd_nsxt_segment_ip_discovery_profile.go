package vcd

import (
	"context"
	"fmt"
	"net/url"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func datasourceVcdNsxtSegmentIpDiscoveryProfile() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceNsxtSegmentIpDiscoveryProfileRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "name of Segment IP Discovery Profile",
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

func datasourceNsxtSegmentIpDiscoveryProfileRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	profileName := d.Get("name").(string)

	contextFilterField, contextUrn, err := getContextFilterField(d)
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
	dSet(d, "nd_snooping_limit", ipDiscoveryProfile.NdSnoopingLimit)

	d.SetId(ipDiscoveryProfile.ID)

	return nil
}

// getContextFilterField determines which field should be used for filtering
func getContextFilterField(d *schema.ResourceData) (string, string, error) {
	switch {
	case d.Get("nsxt_manager_id").(string) != "":
		return "nsxTManagerRef.id", d.Get("nsxt_manager_id").(string), nil
	case d.Get("vdc_id").(string) != "":
		return "orgVdcId", d.Get("vdc_id").(string), nil
	case d.Get("vdc_group_id").(string) != "":
		return "vdcGroupId", d.Get("vdc_group_id").(string), nil

	}

	return "", "", fmt.Errorf("unknown filtering field")
}
