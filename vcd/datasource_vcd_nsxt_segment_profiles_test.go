//go:build network || nsxt || ALL || functional

package vcd

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcdDataSourceNsxtSegmentProfiles(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	// String map to fill the template
	var params = StringMap{
		"TestName":     t.Name(),
		"OrgName":      testConfig.VCD.Org,
		"VdcName":      testConfig.Nsxt.Vdc,
		"VdcGroupName": testConfig.Nsxt.VdcGroup,
		"NsxtManager":  testConfig.Nsxt.Manager,

		"IpDiscoveryProfileName":     testConfig.Nsxt.IpDiscoveryProfile,
		"MacDiscoveryProfileName":    testConfig.Nsxt.MacDiscoveryProfile,
		"QosProfileName":             testConfig.Nsxt.QosProfile,
		"SpoofGuardProfileName":      testConfig.Nsxt.SpoofGuardProfile,
		"SegmentSecurityProfileName": testConfig.Nsxt.SegmentSecurityProfile,

		"Tags": "nsxt",
	}
	testParamsNotEmpty(t, params)

	configText1 := templateFill(testAccVcdDataSourceNsxtSegmentProfilesByNsxtManager, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)

	params["FuncName"] = t.Name() + "step2"
	configText2 := templateFill(testAccVcdDataSourceNsxtSegmentProfilesByVdcId, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText2)

	params["FuncName"] = t.Name() + "step3"
	configText3 := templateFill(testAccVcdDataSourceNsxtSegmentProfilesByVdcGroupId, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 3: %s", configText3)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_ip_discovery_profile.first", "id"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_ip_discovery_profile.first", "description"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_ip_discovery_profile.first", "arp_binding_limit"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_ip_discovery_profile.first", "arp_binding_timeout"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_ip_discovery_profile.first", "is_arp_snooping_enabled"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_ip_discovery_profile.first", "is_dhcp_snooping_v4_enabled"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_ip_discovery_profile.first", "is_dhcp_snooping_v6_enabled"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_ip_discovery_profile.first", "is_duplicate_ip_detection_enabled"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_ip_discovery_profile.first", "is_nd_snooping_enabled"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_ip_discovery_profile.first", "is_tofu_enabled"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_ip_discovery_profile.first", "is_vmtools_v4_enabled"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_ip_discovery_profile.first", "is_vmtools_v6_enabled"),

					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_mac_discovery_profile.first", "id"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_mac_discovery_profile.first", "description"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_mac_discovery_profile.first", "is_mac_change_enabled"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_mac_discovery_profile.first", "is_mac_learning_enabled"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_mac_discovery_profile.first", "is_unknown_unicast_flooding_enabled"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_mac_discovery_profile.first", "mac_learning_aging_time"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_mac_discovery_profile.first", "mac_limit"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_mac_discovery_profile.first", "mac_policy"),

					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_spoof_guard_profile.first", "id"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_spoof_guard_profile.first", "description"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_spoof_guard_profile.first", "is_address_binding_whitelist_enabled"),

					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_qos_profile.first", "id"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_qos_profile.first", "description"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_qos_profile.first", "class_of_service"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_qos_profile.first", "dscp_priority"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_qos_profile.first", "dscp_trust_mode"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_qos_profile.first", "egress_rate_limiter_avg_bandwidth"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_qos_profile.first", "egress_rate_limiter_burst_size"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_qos_profile.first", "egress_rate_limiter_peak_bandwidth"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_qos_profile.first", "ingress_broadcast_rate_limiter_avg_bandwidth"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_qos_profile.first", "ingress_broadcast_rate_limiter_burst_size"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_qos_profile.first", "ingress_broadcast_rate_limiter_peak_bandwidth"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_qos_profile.first", "ingress_rate_limiter_avg_bandwidth"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_qos_profile.first", "ingress_rate_limiter_burst_size"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_qos_profile.first", "ingress_rate_limiter_peak_bandwidth"),

					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_security_profile.first", "id"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_security_profile.first", "description"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_security_profile.first", "bpdu_filter_allow_list.#"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_security_profile.first", "is_bpdu_filter_enabled"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_security_profile.first", "is_dhcp_v4_client_block_enabled"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_security_profile.first", "is_dhcp_v6_client_block_enabled"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_security_profile.first", "is_dhcp_v4_server_block_enabled"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_security_profile.first", "is_dhcp_v6_server_block_enabled"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_security_profile.first", "is_non_ip_traffic_block_enabled"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_security_profile.first", "is_ra_guard_enabled"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_security_profile.first", "is_rate_limitting_enabled"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_security_profile.first", "rx_broadcast_limit"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_security_profile.first", "rx_multicast_limit"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_security_profile.first", "tx_broadcast_limit"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_security_profile.first", "tx_multicast_limit"),
				),
			},
			{
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_ip_discovery_profile.first", "id"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_ip_discovery_profile.first", "description"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_ip_discovery_profile.first", "arp_binding_limit"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_ip_discovery_profile.first", "arp_binding_timeout"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_ip_discovery_profile.first", "is_arp_snooping_enabled"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_ip_discovery_profile.first", "is_dhcp_snooping_v4_enabled"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_ip_discovery_profile.first", "is_dhcp_snooping_v6_enabled"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_ip_discovery_profile.first", "is_duplicate_ip_detection_enabled"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_ip_discovery_profile.first", "is_nd_snooping_enabled"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_ip_discovery_profile.first", "is_tofu_enabled"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_ip_discovery_profile.first", "is_vmtools_v4_enabled"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_ip_discovery_profile.first", "is_vmtools_v6_enabled"),

					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_mac_discovery_profile.first", "id"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_mac_discovery_profile.first", "description"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_mac_discovery_profile.first", "is_mac_change_enabled"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_mac_discovery_profile.first", "is_mac_learning_enabled"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_mac_discovery_profile.first", "is_unknown_unicast_flooding_enabled"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_mac_discovery_profile.first", "mac_learning_aging_time"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_mac_discovery_profile.first", "mac_limit"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_mac_discovery_profile.first", "mac_policy"),

					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_spoof_guard_profile.first", "id"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_spoof_guard_profile.first", "description"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_spoof_guard_profile.first", "is_address_binding_whitelist_enabled"),

					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_qos_profile.first", "id"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_qos_profile.first", "description"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_qos_profile.first", "class_of_service"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_qos_profile.first", "dscp_priority"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_qos_profile.first", "dscp_trust_mode"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_qos_profile.first", "egress_rate_limiter_avg_bandwidth"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_qos_profile.first", "egress_rate_limiter_burst_size"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_qos_profile.first", "egress_rate_limiter_peak_bandwidth"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_qos_profile.first", "ingress_broadcast_rate_limiter_avg_bandwidth"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_qos_profile.first", "ingress_broadcast_rate_limiter_burst_size"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_qos_profile.first", "ingress_broadcast_rate_limiter_peak_bandwidth"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_qos_profile.first", "ingress_rate_limiter_avg_bandwidth"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_qos_profile.first", "ingress_rate_limiter_burst_size"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_qos_profile.first", "ingress_rate_limiter_peak_bandwidth"),

					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_security_profile.first", "id"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_security_profile.first", "description"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_security_profile.first", "bpdu_filter_allow_list.#"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_security_profile.first", "is_bpdu_filter_enabled"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_security_profile.first", "is_dhcp_v4_client_block_enabled"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_security_profile.first", "is_dhcp_v6_client_block_enabled"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_security_profile.first", "is_dhcp_v4_server_block_enabled"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_security_profile.first", "is_dhcp_v6_server_block_enabled"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_security_profile.first", "is_non_ip_traffic_block_enabled"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_security_profile.first", "is_ra_guard_enabled"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_security_profile.first", "is_rate_limitting_enabled"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_security_profile.first", "rx_broadcast_limit"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_security_profile.first", "rx_multicast_limit"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_security_profile.first", "tx_broadcast_limit"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_security_profile.first", "tx_multicast_limit"),
				),
			},
			{
				Config: configText3,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_ip_discovery_profile.first", "id"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_ip_discovery_profile.first", "description"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_ip_discovery_profile.first", "arp_binding_limit"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_ip_discovery_profile.first", "arp_binding_timeout"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_ip_discovery_profile.first", "is_arp_snooping_enabled"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_ip_discovery_profile.first", "is_dhcp_snooping_v4_enabled"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_ip_discovery_profile.first", "is_dhcp_snooping_v6_enabled"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_ip_discovery_profile.first", "is_duplicate_ip_detection_enabled"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_ip_discovery_profile.first", "is_nd_snooping_enabled"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_ip_discovery_profile.first", "is_tofu_enabled"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_ip_discovery_profile.first", "is_vmtools_v4_enabled"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_ip_discovery_profile.first", "is_vmtools_v6_enabled"),

					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_mac_discovery_profile.first", "id"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_mac_discovery_profile.first", "description"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_mac_discovery_profile.first", "is_mac_change_enabled"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_mac_discovery_profile.first", "is_mac_learning_enabled"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_mac_discovery_profile.first", "is_unknown_unicast_flooding_enabled"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_mac_discovery_profile.first", "mac_learning_aging_time"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_mac_discovery_profile.first", "mac_limit"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_mac_discovery_profile.first", "mac_policy"),

					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_spoof_guard_profile.first", "id"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_spoof_guard_profile.first", "description"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_spoof_guard_profile.first", "is_address_binding_whitelist_enabled"),

					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_qos_profile.first", "id"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_qos_profile.first", "description"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_qos_profile.first", "class_of_service"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_qos_profile.first", "dscp_priority"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_qos_profile.first", "dscp_trust_mode"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_qos_profile.first", "egress_rate_limiter_avg_bandwidth"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_qos_profile.first", "egress_rate_limiter_burst_size"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_qos_profile.first", "egress_rate_limiter_peak_bandwidth"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_qos_profile.first", "ingress_broadcast_rate_limiter_avg_bandwidth"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_qos_profile.first", "ingress_broadcast_rate_limiter_burst_size"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_qos_profile.first", "ingress_broadcast_rate_limiter_peak_bandwidth"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_qos_profile.first", "ingress_rate_limiter_avg_bandwidth"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_qos_profile.first", "ingress_rate_limiter_burst_size"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_qos_profile.first", "ingress_rate_limiter_peak_bandwidth"),

					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_security_profile.first", "id"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_security_profile.first", "description"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_security_profile.first", "bpdu_filter_allow_list.#"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_security_profile.first", "is_bpdu_filter_enabled"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_security_profile.first", "is_dhcp_v4_client_block_enabled"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_security_profile.first", "is_dhcp_v6_client_block_enabled"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_security_profile.first", "is_dhcp_v4_server_block_enabled"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_security_profile.first", "is_dhcp_v6_server_block_enabled"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_security_profile.first", "is_non_ip_traffic_block_enabled"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_security_profile.first", "is_ra_guard_enabled"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_security_profile.first", "is_rate_limitting_enabled"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_security_profile.first", "rx_broadcast_limit"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_security_profile.first", "rx_multicast_limit"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_security_profile.first", "tx_broadcast_limit"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_segment_security_profile.first", "tx_multicast_limit"),
				),
			},
		},
	})
}

const testAccVcdDataSourceNsxtSegmentProfilesByNsxtManager = `
data "vcd_nsxt_manager" "nsxt" {
  name = "{{.NsxtManager}}"
}

data "vcd_nsxt_segment_ip_discovery_profile" "first" {
  name       = "{{.IpDiscoveryProfileName}}"
  nsxt_manager_id = data.vcd_nsxt_manager.nsxt.id
}

data "vcd_nsxt_segment_mac_discovery_profile" "first" {
  name       = "{{.MacDiscoveryProfileName}}"
  nsxt_manager_id = data.vcd_nsxt_manager.nsxt.id
}

data "vcd_nsxt_segment_spoof_guard_profile" "first" {
  name       = "{{.SpoofGuardProfileName}}"
  nsxt_manager_id = data.vcd_nsxt_manager.nsxt.id
}

data "vcd_nsxt_segment_qos_profile" "first" {
  name       = "{{.QosProfileName}}"
  nsxt_manager_id = data.vcd_nsxt_manager.nsxt.id
}

data "vcd_nsxt_segment_security_profile" "first" {
  name       = "{{.SegmentSecurityProfileName}}"
  nsxt_manager_id = data.vcd_nsxt_manager.nsxt.id
}
`

const testAccVcdDataSourceNsxtSegmentProfilesByVdcId = `
data "vcd_org_vdc" "nsxt" {
  org  = "{{.OrgName}}"
  name = "{{.VdcName}}"
}

data "vcd_nsxt_segment_ip_discovery_profile" "first" {
  name   = "{{.IpDiscoveryProfileName}}"
  vdc_id = data.vcd_org_vdc.nsxt.id
}

data "vcd_nsxt_segment_mac_discovery_profile" "first" {
  name   = "{{.MacDiscoveryProfileName}}"
  vdc_id = data.vcd_org_vdc.nsxt.id
}

data "vcd_nsxt_segment_spoof_guard_profile" "first" {
  name   = "{{.SpoofGuardProfileName}}"
  vdc_id = data.vcd_org_vdc.nsxt.id
}

data "vcd_nsxt_segment_qos_profile" "first" {
  name   = "{{.QosProfileName}}"
  vdc_id = data.vcd_org_vdc.nsxt.id
}

data "vcd_nsxt_segment_security_profile" "first" {
  name   = "{{.SegmentSecurityProfileName}}"
  vdc_id = data.vcd_org_vdc.nsxt.id
}
`

const testAccVcdDataSourceNsxtSegmentProfilesByVdcGroupId = `
data "vcd_vdc_group" "nsxt" {
  org  = "{{.OrgName}}"
  name = "{{.VdcGroupName}}"
}

data "vcd_nsxt_segment_ip_discovery_profile" "first" {
  name         = "{{.IpDiscoveryProfileName}}"
  vdc_group_id = data.vcd_vdc_group.nsxt.id
}

data "vcd_nsxt_segment_mac_discovery_profile" "first" {
  name         = "{{.MacDiscoveryProfileName}}"
  vdc_group_id = data.vcd_vdc_group.nsxt.id
}

data "vcd_nsxt_segment_spoof_guard_profile" "first" {
  name         = "{{.SpoofGuardProfileName}}"
  vdc_group_id = data.vcd_vdc_group.nsxt.id
}

data "vcd_nsxt_segment_qos_profile" "first" {
  name         = "{{.QosProfileName}}"
  vdc_group_id = data.vcd_vdc_group.nsxt.id
}

data "vcd_nsxt_segment_security_profile" "first" {
  name         = "{{.SegmentSecurityProfileName}}"
  vdc_group_id = data.vcd_vdc_group.nsxt.id
}
`
