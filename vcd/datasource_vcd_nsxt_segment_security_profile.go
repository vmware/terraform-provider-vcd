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
				Description: "Name of Segment Security Profile",
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
				Description: "Description of Segment Security Profile",
			},
			"bpdu_filter_allow_list": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "Indicates pre-defined list of allowed MAC addresses to be excluded from BPDU filtering",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"is_bpdu_filter_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Indicates whether BPDU filter is enabled",
			},
			"is_dhcp_v4_client_block_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Indicates whether DHCP Client block IPv4 is enabled",
			},
			"is_dhcp_v6_client_block_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Indicates whether DHCP Client block IPv6 is enabled",
			},
			"is_dhcp_v4_server_block_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Indicates whether DHCP Server block IPv4 is enabled",
			},
			"is_dhcp_v6_server_block_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Indicates whether DHCP Server block IPv6 is enabled",
			},
			"is_non_ip_traffic_block_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Indicates whether non IP traffic block is enabled",
			},
			"is_ra_guard_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Indicates whether Router Advertisement Guard is enabled",
			},
			"is_rate_limitting_enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Indicates whether Rate Limiting is enabled",
			},
			"rx_broadcast_limit": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Incoming broadcast traffic limit in packets per second",
			},
			"rx_multicast_limit": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Incoming multicast traffic limit in packets per second",
			},
			"tx_broadcast_limit": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Outgoing broadcast traffic limit in packets per second",
			},
			"tx_multicast_limit": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Outgoing multicast traffic limit in packets per second",
			},
		},
	}
}

func datasourceNsxtSegmentSecurityProfileRead(_ context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	vcdClient := meta.(*VCDClient)
	profileName := d.Get("name").(string)

	contextFilterField, contextUrn, err := getContextFilterField(d)
	if err != nil {
		return diag.FromErr(err)
	}

	queryFilter := url.Values{}
	queryFilter.Add("filter", fmt.Sprintf("%s==%s", contextFilterField, contextUrn))

	segmentSecurityProfile, err := vcdClient.GetSegmentSecurityProfileByName(profileName, queryFilter)
	if err != nil {
		return diag.Errorf("could not find Segment Security Profile by name '%s': %s", profileName, err)
	}

	dSet(d, "description", segmentSecurityProfile.Description)

	bpduAllowList := convertStringsToTypeSet(segmentSecurityProfile.BpduFilterAllowList)
	err = d.Set("bpdu_filter_allow_list", bpduAllowList)
	if err != nil {
		return diag.Errorf("error storing 'bpdu_filter_allow_list': %s", err)
	}

	dSet(d, "is_bpdu_filter_enabled", segmentSecurityProfile.IsBpduFilterEnabled)
	dSet(d, "is_dhcp_v4_client_block_enabled", segmentSecurityProfile.IsDhcpClientBlockV4Enabled)
	dSet(d, "is_dhcp_v6_client_block_enabled", segmentSecurityProfile.IsDhcpClientBlockV6Enabled)
	dSet(d, "is_dhcp_v4_server_block_enabled", segmentSecurityProfile.IsDhcpServerBlockV4Enabled)
	dSet(d, "is_dhcp_v6_server_block_enabled", segmentSecurityProfile.IsDhcpServerBlockV6Enabled)
	dSet(d, "is_non_ip_traffic_block_enabled", segmentSecurityProfile.IsNonIPTrafficBlockEnabled)
	dSet(d, "is_ra_guard_enabled", segmentSecurityProfile.IsRaGuardEnabled)
	dSet(d, "is_rate_limitting_enabled", segmentSecurityProfile.IsRateLimitingEnabled)
	dSet(d, "rx_broadcast_limit", segmentSecurityProfile.RateLimits.RxBroadcast)
	dSet(d, "rx_multicast_limit", segmentSecurityProfile.RateLimits.RxMulticast)
	dSet(d, "tx_broadcast_limit", segmentSecurityProfile.RateLimits.TxBroadcast)
	dSet(d, "tx_multicast_limit", segmentSecurityProfile.RateLimits.TxMulticast)

	d.SetId(segmentSecurityProfile.ID)

	return nil
}
