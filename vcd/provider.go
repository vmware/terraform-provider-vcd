package vcd

import (
	"context"
	"fmt"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"os"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/vmware/go-vcloud-director/v2/util"
)

// BuildVersion holds version which is meant to be injected at build time using ldflags
// (e.g. 'go build -ldflags="-X 'github.com/vmware/terraform-provider-vcd/v3/vcd.BuildVersion=v1.0.0'"')
var BuildVersion = "unset"

// DataSources is a public function which allows filtering and access all defined data sources
// When 'nameRegexp' is not empty - it will return only those matching the regexp
// When 'includeDeprecated' is false - it will skip out the resources which have a DeprecationMessage set
func DataSources(nameRegexp string, includeDeprecated bool) (map[string]*schema.Resource, error) {
	return vcdSchemaFilter(globalDataSourceMap, nameRegexp, includeDeprecated)
}

// Resources is a public function which allows filtering and access all defined resources
// When 'nameRegexp' is not empty - it will return only those matching the regexp
// When 'includeDeprecated' is false - it will skip out the resources which have a DeprecationMessage set
func Resources(nameRegexp string, includeDeprecated bool) (map[string]*schema.Resource, error) {
	return vcdSchemaFilter(globalResourceMap, nameRegexp, includeDeprecated)
}

var globalDataSourceMap = map[string]*schema.Resource{
	"vcd_org":                                       datasourceVcdOrg(),                              // 2.5
	"vcd_org_group":                                 datasourceVcdOrgGroup(),                         // 3.6
	"vcd_org_user":                                  datasourceVcdOrgUser(),                          // 3.0
	"vcd_org_vdc":                                   datasourceVcdOrgVdc(),                           // 2.5
	"vcd_catalog":                                   datasourceVcdCatalog(),                          // 2.5
	"vcd_catalog_media":                             datasourceVcdCatalogMedia(),                     // 2.5
	"vcd_catalog_item":                              datasourceVcdCatalogItem(),                      // 2.5
	"vcd_edgegateway":                               datasourceVcdEdgeGateway(),                      // 2.5
	"vcd_external_network":                          datasourceVcdExternalNetwork(),                  // 2.5
	"vcd_external_network_v2":                       datasourceVcdExternalNetworkV2(),                // 3.0
	"vcd_independent_disk":                          datasourceVcIndependentDisk(),                   // 2.5
	"vcd_network_routed":                            datasourceVcdNetworkRouted(),                    // 2.5
	"vcd_network_direct":                            datasourceVcdNetworkDirect(),                    // 2.5
	"vcd_network_isolated":                          datasourceVcdNetworkIsolated(),                  // 2.5
	"vcd_vapp":                                      datasourceVcdVApp(),                             // 2.5
	"vcd_vapp_vm":                                   datasourceVcdVAppVm(),                           // 2.6
	"vcd_lb_service_monitor":                        datasourceVcdLbServiceMonitor(),                 // 2.4
	"vcd_lb_server_pool":                            datasourceVcdLbServerPool(),                     // 2.4
	"vcd_lb_app_profile":                            datasourceVcdLBAppProfile(),                     // 2.4
	"vcd_lb_app_rule":                               datasourceVcdLBAppRule(),                        // 2.4
	"vcd_lb_virtual_server":                         datasourceVcdLbVirtualServer(),                  // 2.4
	"vcd_nsxv_dnat":                                 datasourceVcdNsxvDnat(),                         // 2.5
	"vcd_nsxv_snat":                                 datasourceVcdNsxvSnat(),                         // 2.5
	"vcd_nsxv_firewall_rule":                        datasourceVcdNsxvFirewallRule(),                 // 2.5
	"vcd_nsxv_dhcp_relay":                           datasourceVcdNsxvDhcpRelay(),                    // 2.6
	"vcd_nsxv_ip_set":                               datasourceVcdIpSet(),                            // 2.6
	"vcd_vapp_network":                              datasourceVcdVappNetwork(),                      // 2.7
	"vcd_vapp_org_network":                          datasourceVcdVappOrgNetwork(),                   // 2.7
	"vcd_vm_affinity_rule":                          datasourceVcdVmAffinityRule(),                   // 2.9
	"vcd_vm_sizing_policy":                          datasourceVcdVmSizingPolicy(),                   // 3.0
	"vcd_nsxt_manager":                              datasourceVcdNsxtManager(),                      // 3.0
	"vcd_nsxt_tier0_router":                         datasourceVcdNsxtTier0Router(),                  // 3.0
	"vcd_portgroup":                                 datasourceVcdPortgroup(),                        // 3.0
	"vcd_vcenter":                                   datasourceVcdVcenter(),                          // 3.0
	"vcd_resource_list":                             datasourceVcdResourceList(),                     // 3.1
	"vcd_resource_schema":                           datasourceVcdResourceSchema(),                   // 3.1
	"vcd_nsxt_edge_cluster":                         datasourceVcdNsxtEdgeCluster(),                  // 3.1
	"vcd_nsxt_edgegateway":                          datasourceVcdNsxtEdgeGateway(),                  // 3.1
	"vcd_storage_profile":                           datasourceVcdStorageProfile(),                   // 3.1
	"vcd_vm":                                        datasourceVcdStandaloneVm(),                     // 3.2
	"vcd_network_routed_v2":                         datasourceVcdNetworkRoutedV2(),                  // 3.2
	"vcd_network_isolated_v2":                       datasourceVcdNetworkIsolatedV2(),                // 3.2
	"vcd_nsxt_network_imported":                     datasourceVcdNsxtNetworkImported(),              // 3.2
	"vcd_nsxt_network_dhcp":                         datasourceVcdOpenApiDhcp(),                      // 3.2
	"vcd_right":                                     datasourceVcdRight(),                            // 3.3
	"vcd_role":                                      datasourceVcdRole(),                             // 3.3
	"vcd_global_role":                               datasourceVcdGlobalRole(),                       // 3.3
	"vcd_rights_bundle":                             datasourceVcdRightsBundle(),                     // 3.3
	"vcd_nsxt_ip_set":                               datasourceVcdNsxtIpSet(),                        // 3.3
	"vcd_nsxt_security_group":                       datasourceVcdNsxtSecurityGroup(),                // 3.3
	"vcd_nsxt_app_port_profile":                     datasourceVcdNsxtAppPortProfile(),               // 3.3
	"vcd_nsxt_nat_rule":                             datasourceVcdNsxtNatRule(),                      // 3.3
	"vcd_nsxt_firewall":                             datasourceVcdNsxtFirewall(),                     // 3.3
	"vcd_nsxt_ipsec_vpn_tunnel":                     datasourceVcdNsxtIpSecVpnTunnel(),               // 3.3
	"vcd_nsxt_alb_importable_cloud":                 datasourceVcdAlbImportableCloud(),               // 3.4
	"vcd_nsxt_alb_controller":                       datasourceVcdAlbController(),                    // 3.4
	"vcd_nsxt_alb_cloud":                            datasourceVcdAlbCloud(),                         // 3.4
	"vcd_nsxt_alb_service_engine_group":             datasourceVcdAlbServiceEngineGroup(),            // 3.4
	"vcd_nsxt_alb_settings":                         datasourceVcdAlbSettings(),                      // 3.5
	"vcd_nsxt_alb_edgegateway_service_engine_group": datasourceVcdAlbEdgeGatewayServiceEngineGroup(), // 3.5
	"vcd_library_certificate":                       datasourceLibraryCertificate(),                  // 3.5
	"vcd_nsxt_alb_pool":                             datasourceVcdAlbPool(),                          // 3.5
	"vcd_nsxt_alb_virtual_service":                  datasourceVcdAlbVirtualService(),                // 3.5
	"vcd_vdc_group":                                 datasourceVdcGroup(),                            // 3.5
	"vcd_nsxt_distributed_firewall":                 datasourceVcdNsxtDistributedFirewall(),          // 3.6
	"vcd_nsxt_network_context_profile":              datasourceVcdNsxtNetworkContextProfile(),        // 3.6
	"vcd_nsxt_route_advertisement":                  datasourceVcdNsxtRouteAdvertisement(),           // 3.7
	"vcd_nsxt_edgegateway_bgp_configuration":        datasourceVcdEdgeBgpConfig(),                    // 3.7
	"vcd_nsxt_edgegateway_bgp_neighbor":             datasourceVcdEdgeBgpNeighbor(),                  // 3.7
	"vcd_nsxt_edgegateway_bgp_ip_prefix_list":       datasourceVcdEdgeBgpIpPrefixList(),              // 3.7
	"vcd_nsxt_dynamic_security_group":               datasourceVcdDynamicSecurityGroup(),             // 3.7
	"vcd_org_ldap":                                  datasourceVcdOrgLdap(),                          // 3.8
	"vcd_vm_placement_policy":                       datasourceVcdVmPlacementPolicy(),                // 3.8
	"vcd_provider_vdc":                              datasourceVcdProviderVdc(),                      // 3.8
	"vcd_vm_group":                                  datasourceVcdVmGroup(),                          // 3.8
	"vcd_catalog_vapp_template":                     datasourceVcdCatalogVappTemplate(),              // 3.8
	"vcd_subscribed_catalog":                        datasourceVcdSubscribedCatalog(),                // 3.8
	"vcd_task":                                      datasourceVcdTask(),                             // 3.8
	"vcd_nsxv_distributed_firewall":                 datasourceVcdNsxvDistributedFirewall(),          // 3.9
	"vcd_nsxv_application_finder":                   datasourceVcdNsxvApplicationFinder(),            // 3.9
	"vcd_nsxv_application":                          datasourceVcdNsxvApplication(),                  // 3.9
	"vcd_nsxv_application_group":                    datasourceVcdNsxvApplicationGroup(),             // 3.9
	"vcd_rde_interface":                             datasourceVcdRdeInterface(),                     // 3.9
	"vcd_rde_type":                                  datasourceVcdRdeType(),                          // 3.9
	"vcd_rde":                                       datasourceVcdRde(),                              // 3.9
	"vcd_nsxt_edgegateway_qos_profile":              datasourceVcdNsxtEdgeGatewayQosProfile(),        // 3.9
	"vcd_nsxt_edgegateway_rate_limiting":            datasourceVcdNsxtEdgegatewayRateLimiting(),      // 3.9
	"vcd_nsxt_network_dhcp_binding":                 datasourceVcdNsxtDhcpBinding(),                  // 3.9
	"vcd_ip_space":                                  datasourceVcdIpSpace(),                          // 3.10
	"vcd_ip_space_uplink":                           datasourceVcdIpSpaceUplink(),                    // 3.10
	"vcd_ip_space_ip_allocation":                    datasourceVcdIpAllocation(),                     // 3.10
	"vcd_ip_space_custom_quota":                     datasourceVcdIpSpaceCustomQuota(),               // 3.10
	"vcd_nsxt_edgegateway_dhcp_forwarding":          datasourceVcdNsxtEdgegatewayDhcpForwarding(),    // 3.10
	"vcd_nsxt_edgegateway_dhcpv6":                   datasourceVcdNsxtEdgegatewayDhcpV6(),            // 3.10
	"vcd_org_saml":                                  datasourceVcdOrgSaml(),                          // 3.10
	"vcd_org_saml_metadata":                         datasourceVcdOrgSamlMetadata(),                  // 3.10
	"vcd_nsxt_distributed_firewall_rule":            datasourceVcdNsxtDistributedFirewallRule(),      // 3.10
	"vcd_nsxt_edgegateway_static_route":             datasourceVcdNsxtEdgeGatewayStaticRoute(),       // 3.10
	"vcd_resource_pool":                             datasourceVcdResourcePool(),                     // 3.10
	"vcd_network_pool":                              datasourceVcdNetworkPool(),                      // 3.10
	"vcd_ui_plugin":                                 datasourceVcdUIPlugin(),                         // 3.10
	"vcd_service_account":                           datasourceVcdServiceAccount(),                   // 3.10
	"vcd_rde_interface_behavior":                    datasourceVcdRdeInterfaceBehavior(),             // 3.10
	"vcd_rde_type_behavior":                         datasourceVcdRdeTypeBehavior(),                  // 3.10
	"vcd_rde_type_behavior_acl":                     datasourceVcdRdeTypeBehaviorAccessLevel(),       // 3.10
}

var globalResourceMap = map[string]*schema.Resource{
	"vcd_network_routed":                            resourceVcdNetworkRouted(),                    // 2.0
	"vcd_network_direct":                            resourceVcdNetworkDirect(),                    // 2.0
	"vcd_network_isolated":                          resourceVcdNetworkIsolated(),                  // 2.0
	"vcd_vapp_network":                              resourceVcdVappNetwork(),                      // 2.1
	"vcd_vapp":                                      resourceVcdVApp(),                             // 1.0
	"vcd_edgegateway":                               resourceVcdEdgeGateway(),                      // 2.4
	"vcd_edgegateway_vpn":                           resourceVcdEdgeGatewayVpn(),                   // 1.0
	"vcd_edgegateway_settings":                      resourceVcdEdgeGatewaySettings(),              // 3.0
	"vcd_vapp_vm":                                   resourceVcdVAppVm(),                           // 1.0
	"vcd_org":                                       resourceOrg(),                                 // 2.0
	"vcd_org_vdc":                                   resourceVcdOrgVdc(),                           // 2.2
	"vcd_org_user":                                  resourceVcdOrgUser(),                          // 2.4
	"vcd_catalog":                                   resourceVcdCatalog(),                          // 2.0
	"vcd_catalog_item":                              resourceVcdCatalogItem(),                      // 2.0
	"vcd_catalog_media":                             resourceVcdCatalogMedia(),                     // 2.0
	"vcd_inserted_media":                            resourceVcdInsertedMedia(),                    // 2.1
	"vcd_independent_disk":                          resourceVcdIndependentDisk(),                  // 2.1
	"vcd_external_network":                          resourceVcdExternalNetwork(),                  // 2.2
	"vcd_lb_service_monitor":                        resourceVcdLbServiceMonitor(),                 // 2.4
	"vcd_lb_server_pool":                            resourceVcdLBServerPool(),                     // 2.4
	"vcd_lb_app_profile":                            resourceVcdLBAppProfile(),                     // 2.4
	"vcd_lb_app_rule":                               resourceVcdLBAppRule(),                        // 2.4
	"vcd_lb_virtual_server":                         resourceVcdLBVirtualServer(),                  // 2.4
	"vcd_nsxv_dnat":                                 resourceVcdNsxvDnat(),                         // 2.5
	"vcd_nsxv_snat":                                 resourceVcdNsxvSnat(),                         // 2.5
	"vcd_nsxv_firewall_rule":                        resourceVcdNsxvFirewallRule(),                 // 2.5
	"vcd_nsxv_dhcp_relay":                           resourceVcdNsxvDhcpRelay(),                    // 2.6
	"vcd_nsxv_ip_set":                               resourceVcdIpSet(),                            // 2.6
	"vcd_vm_internal_disk":                          resourceVmInternalDisk(),                      // 2.7
	"vcd_vapp_org_network":                          resourceVcdVappOrgNetwork(),                   // 2.7
	"vcd_org_group":                                 resourceVcdOrgGroup(),                         // 2.9
	"vcd_vapp_firewall_rules":                       resourceVcdVappFirewallRules(),                // 2.9
	"vcd_vapp_nat_rules":                            resourceVcdVappNetworkNatRules(),              // 2.9
	"vcd_vapp_static_routing":                       resourceVcdVappNetworkStaticRouting(),         // 2.9
	"vcd_vm_affinity_rule":                          resourceVcdVmAffinityRule(),                   // 2.9
	"vcd_vapp_access_control":                       resourceVcdAccessControlVapp(),                // 3.0
	"vcd_external_network_v2":                       resourceVcdExternalNetworkV2(),                // 3.0
	"vcd_vm_sizing_policy":                          resourceVcdVmSizingPolicy(),                   // 3.0
	"vcd_nsxt_edgegateway":                          resourceVcdNsxtEdgeGateway(),                  // 3.1
	"vcd_vm":                                        resourceVcdStandaloneVm(),                     // 3.2
	"vcd_network_routed_v2":                         resourceVcdNetworkRoutedV2(),                  // 3.2
	"vcd_network_isolated_v2":                       resourceVcdNetworkIsolatedV2(),                // 3.2
	"vcd_nsxt_network_imported":                     resourceVcdNsxtNetworkImported(),              // 3.2
	"vcd_nsxt_network_dhcp":                         resourceVcdOpenApiDhcp(),                      // 3.2
	"vcd_role":                                      resourceVcdRole(),                             // 3.3
	"vcd_global_role":                               resourceVcdGlobalRole(),                       // 3.3
	"vcd_rights_bundle":                             resourceVcdRightsBundle(),                     // 3.3
	"vcd_nsxt_ip_set":                               resourceVcdNsxtIpSet(),                        // 3.3
	"vcd_nsxt_security_group":                       resourceVcdSecurityGroup(),                    // 3.3
	"vcd_nsxt_firewall":                             resourceVcdNsxtFirewall(),                     // 3.3
	"vcd_nsxt_app_port_profile":                     resourceVcdNsxtAppPortProfile(),               // 3.3
	"vcd_nsxt_nat_rule":                             resourceVcdNsxtNatRule(),                      // 3.3
	"vcd_nsxt_ipsec_vpn_tunnel":                     resourceVcdNsxtIpSecVpnTunnel(),               // 3.3
	"vcd_nsxt_alb_cloud":                            resourceVcdAlbCloud(),                         // 3.4
	"vcd_nsxt_alb_controller":                       resourceVcdAlbController(),                    // 3.4
	"vcd_nsxt_alb_service_engine_group":             resourceVcdAlbServiceEngineGroup(),            // 3.4
	"vcd_nsxt_alb_settings":                         resourceVcdAlbSettings(),                      // 3.5
	"vcd_nsxt_alb_edgegateway_service_engine_group": resourceVcdAlbEdgeGatewayServiceEngineGroup(), // 3.5
	"vcd_library_certificate":                       resourceLibraryCertificate(),                  // 3.5
	"vcd_nsxt_alb_pool":                             resourceVcdAlbPool(),                          // 3.5
	"vcd_nsxt_alb_virtual_service":                  resourceVcdAlbVirtualService(),                // 3.5
	"vcd_vdc_group":                                 resourceVdcGroup(),                            // 3.5
	"vcd_nsxt_distributed_firewall":                 resourceVcdNsxtDistributedFirewall(),          // 3.6
	"vcd_security_tag":                              resourceVcdSecurityTag(),                      // 3.7
	"vcd_nsxt_route_advertisement":                  resourceVcdNsxtRouteAdvertisement(),           // 3.7
	"vcd_org_vdc_access_control":                    resourceVcdOrgVdcAccessControl(),              // 3.7
	"vcd_nsxt_dynamic_security_group":               resourceVcdDynamicSecurityGroup(),             // 3.7
	"vcd_nsxt_edgegateway_bgp_neighbor":             resourceVcdEdgeBgpNeighbor(),                  // 3.7
	"vcd_nsxt_edgegateway_bgp_ip_prefix_list":       resourceVcdEdgeBgpIpPrefixList(),              // 3.7
	"vcd_nsxt_edgegateway_bgp_configuration":        resourceVcdEdgeBgpConfig(),                    // 3.7
	"vcd_org_ldap":                                  resourceVcdOrgLdap(),                          // 3.8
	"vcd_vm_placement_policy":                       resourceVcdVmPlacementPolicy(),                // 3.8
	"vcd_catalog_vapp_template":                     resourceVcdCatalogVappTemplate(),              // 3.8
	"vcd_catalog_access_control":                    resourceVcdCatalogAccessControl(),             // 3.8
	"vcd_subscribed_catalog":                        resourceVcdSubscribedCatalog(),                // 3.8
	"vcd_nsxv_distributed_firewall":                 resourceVcdNsxvDistributedFirewall(),          // 3.9
	"vcd_rde_interface":                             resourceVcdRdeInterface(),                     // 3.9
	"vcd_rde_type":                                  resourceVcdRdeType(),                          // 3.9
	"vcd_rde":                                       resourceVcdRde(),                              // 3.9
	"vcd_nsxt_edgegateway_rate_limiting":            resourceVcdNsxtEdgegatewayRateLimiting(),      // 3.9
	"vcd_nsxt_network_dhcp_binding":                 resourceVcdNsxtDhcpBinding(),                  // 3.9
	"vcd_ip_space":                                  resourceVcdIpSpace(),                          // 3.10
	"vcd_ip_space_uplink":                           resourceVcdIpSpaceUplink(),                    // 3.10
	"vcd_ip_space_ip_allocation":                    resourceVcdIpAllocation(),                     // 3.10
	"vcd_ip_space_custom_quota":                     resourceVcdIpSpaceCustomQuota(),               // 3.10
	"vcd_nsxt_edgegateway_dhcp_forwarding":          resourceVcdNsxtEdgegatewayDhcpForwarding(),    // 3.10
	"vcd_nsxt_edgegateway_dhcpv6":                   resourceVcdNsxtEdgegatewayDhcpV6(),            // 3.10
	"vcd_org_saml":                                  resourceVcdOrgSaml(),                          // 3.10
	"vcd_nsxt_distributed_firewall_rule":            resourceVcdNsxtDistributedFirewallRule(),      // 3.10
	"vcd_nsxt_edgegateway_static_route":             resourceVcdNsxtEdgeGatewayStaticRoute(),       // 3.10
	"vcd_provider_vdc":                              resourceVcdProviderVdc(),                      // 3.10
	"vcd_cloned_vapp":                               resourceVcdClonedVApp(),                       // 3.10
	"vcd_ui_plugin":                                 resourceVcdUIPlugin(),                         // 3.10
	"vcd_api_token":                                 resourceVcdApiToken(),                         // 3.10
	"vcd_service_account":                           resourceVcdServiceAccount(),                   // 3.10
	"vcd_rde_interface_behavior":                    resourceVcdRdeInterfaceBehavior(),             // 3.10
	"vcd_rde_type_behavior":                         resourceVcdRdeTypeBehavior(),                  // 3.10
	"vcd_rde_type_behavior_acl":                     resourceVcdRdeTypeBehaviorAccessLevel(),       // 3.10
}

// Provider returns a terraform.ResourceProvider.
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"user": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("VCD_USER", nil),
				Description: "The user name for VCD API operations.",
			},

			"password": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("VCD_PASSWORD", nil),
				Description: "The user password for VCD API operations.",
			},

			"auth_type": {
				Type:         schema.TypeString,
				Optional:     true,
				DefaultFunc:  schema.EnvDefaultFunc("VCD_AUTH_TYPE", "integrated"),
				Description:  "'integrated', 'saml_adfs', 'token', 'api_token', 'api_token_file' and 'service_account_token_file' are supported. 'integrated' is default.",
				ValidateFunc: validation.StringInSlice([]string{"integrated", "saml_adfs", "token", "api_token", "api_token_file", "service_account_token_file"}, false),
			},

			"saml_adfs_rpt_id": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("VCD_SAML_ADFS_RPT_ID", nil),
				Description: "Allows to specify custom Relaying Party Trust Identifier for auth_type=saml_adfs",
			},

			"token": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("VCD_TOKEN", nil),
				Description: "The token used instead of username/password for VCD API operations.",
			},

			"api_token": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("VCD_API_TOKEN", nil),
				Description: "The API token used instead of username/password for VCD API operations. (Requires VCD 10.3.1+)",
			},

			"api_token_file": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("VCD_API_TOKEN_FILE", nil),
				Description: "The API token file instead of username/password for VCD API operations. (Requires VCD 10.3.1+)",
			},

			"allow_api_token_file": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Set this to true if you understand the security risks of using API token files and would like to suppress the warnings",
			},

			"service_account_token_file": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("VCD_SA_TOKEN_FILE", nil),
				Description: "The Service Account API token file instead of username/password for VCD API operations. (Requires VCD 10.4.0+)",
			},

			"allow_service_account_token_file": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Set this to true if you understand the security risks of using Service Account token files and would like to suppress the warnings",
			},

			"sysorg": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("VCD_SYS_ORG", nil),
				Description: "The VCD Org for user authentication",
			},

			"org": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("VCD_ORG", nil),
				Description: "The VCD Org for API operations",
			},

			"vdc": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("VCD_VDC", nil),
				Description: "The VDC for API operations",
			},

			"url": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("VCD_URL", nil),
				Description: "The VCD url for VCD API operations.",
			},

			"max_retry_timeout": {
				Type:        schema.TypeInt,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("VCD_MAX_RETRY_TIMEOUT", 60),
				Description: "Max num seconds to wait for successful response when operating on resources within vCloud (defaults to 60)",
			},

			"allow_unverified_ssl": {
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("VCD_ALLOW_UNVERIFIED_SSL", false),
				Description: "If set, VCDClient will permit unverifiable SSL certificates.",
			},

			"logging": {
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("VCD_API_LOGGING", false),
				Description: "If set, it will enable logging of API requests and responses",
			},

			"logging_file": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("VCD_API_LOGGING_FILE", "go-vcloud-director.log"),
				Description: "Defines the full name of the logging file for API calls (requires 'logging')",
			},
			"import_separator": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("VCD_IMPORT_SEPARATOR", "."),
				Description: "Defines the import separation string to be used with 'terraform import'",
			},
			"ignore_metadata_changes": ignoreMetadataSchema(),
		},
		ResourcesMap:         globalResourceMap,
		DataSourcesMap:       globalDataSourceMap,
		ConfigureContextFunc: providerConfigure,
	}
}

func providerConfigure(_ context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	maxRetryTimeout := d.Get("max_retry_timeout").(int)

	if err := validateProviderSchema(d); err != nil {
		return nil, diag.Errorf("[provider validation] :%s", err)
	}

	// If sysOrg is defined, we use it for authentication.
	// Otherwise, we use the default org defined for regular usage
	connectOrg := d.Get("sysorg").(string)
	if connectOrg == "" {
		connectOrg = d.Get("org").(string)
	}

	config := Config{
		User:                    d.Get("user").(string),
		Password:                d.Get("password").(string),
		Token:                   d.Get("token").(string),
		ApiToken:                d.Get("api_token").(string),
		ApiTokenFile:            d.Get("api_token_file").(string),
		AllowApiTokenFile:       d.Get("allow_api_token_file").(bool),
		ServiceAccountTokenFile: d.Get("service_account_token_file").(string),
		AllowSATokenFile:        d.Get("allow_service_account_token_file").(bool),
		SysOrg:                  connectOrg,            // Connection org
		Org:                     d.Get("org").(string), // Default org for operations
		Vdc:                     d.Get("vdc").(string), // Default vdc
		Href:                    d.Get("url").(string),
		MaxRetryTimeout:         maxRetryTimeout,
		InsecureFlag:            d.Get("allow_unverified_ssl").(bool),
	}

	// auth_type dependent configuration
	authType := d.Get("auth_type").(string)
	switch authType {
	case "saml_adfs":
		config.UseSamlAdfs = true
		config.CustomAdfsRptId = d.Get("saml_adfs_rpt_id").(string)
	case "token":
		if config.Token == "" {
			return nil, diag.Errorf("empty token detected with 'auth_type' == 'token'")
		}
	case "api_token":
		if config.ApiToken == "" {
			return nil, diag.Errorf("empty API token detected with 'auth_type' == 'api_token'")
		}
	case "service_account_token_file":
		if config.ServiceAccountTokenFile == "" {
			return nil, diag.Errorf("service account token file not provided with 'auth_type' == 'service_account_token_file'")
		}
	case "api_token_file":
		if config.ApiTokenFile == "" {
			return nil, diag.Errorf("api token file not provided with 'auth_type' == 'service_account_token_file'")
		}
	default:
		if config.ApiToken != "" || config.Token != "" {
			return nil, diag.Errorf("to use a token, the appropriate 'auth_type' (either 'token' or 'api_token') must be set")
		}
	}
	if config.ApiToken != "" && config.Token != "" {
		return nil, diag.Errorf("only one of 'token' or 'api_token' should be set")
	}

	var providerDiagnostics diag.Diagnostics
	if config.ServiceAccountTokenFile != "" && !config.AllowSATokenFile {
		providerDiagnostics = append(providerDiagnostics, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "The file " + config.ServiceAccountTokenFile + " should be considered sensitive information.",
			Detail: "The file " + config.ServiceAccountTokenFile + " containing the initial service account API " +
				"HAS BEEN UPDATED with a freshly generated token. The initial token was invalidated and the " +
				"token currently in the file will be invalidated at the next usage. In the meantime, it is " +
				"usable by anyone to run operations to the current VCD. As such, it should be considered SENSITIVE INFORMATION. " +
				"If you would like to remove this warning, add\n\n" + "	allow_service_account_token_file = true\n\nto the provider settings.",
		})
	}

	if config.ApiTokenFile != "" && !config.AllowApiTokenFile {
		providerDiagnostics = append(providerDiagnostics, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "The file " + config.ServiceAccountTokenFile + " should be considered sensitive information.",
			Detail: "The file " + config.ServiceAccountTokenFile + " contains the API token which can be used by anyone " +
				"to run operations to the current VCD. AS such, it should be considered SENSITIVE INFORMATION. " +
				"If you would like to remove this warning, add\n\n" + "	allow_api_token_file = true\n\nto the provider settings.",
		})
	}

	// If the provider includes logging directives,
	// it will activate logging from upstream go-vcloud-director
	logging := d.Get("logging").(bool)
	// Logging is disabled by default.
	// If enabled, we set the log file name and invoke the upstream logging set-up
	if logging {
		loggingFile := d.Get("logging_file").(string)
		if loggingFile != "" {
			util.EnableLogging = true
			util.ApiLogFileName = loggingFile
			util.InitLogging()
		}
	}

	separator := os.Getenv("VCD_IMPORT_SEPARATOR")
	if separator != "" {
		ImportSeparator = separator
	} else {
		ImportSeparator = d.Get("import_separator").(string)
	}

	ignoredMetadata, err := getIgnoredMetadata(d, "ignore_metadata_changes")
	if err != nil {
		return nil, diag.Errorf("could not process the metadata that needs to be ignored: %s", err)
	}
	config.IgnoredMetadata = make([]govcd.IgnoredMetadata, len(ignoredMetadata))
	IgnoreMetadataChangesConflictActions = map[string]string{}
	for i, im := range ignoredMetadata {
		config.IgnoredMetadata[i] = ignoredMetadata[i].IgnoredMetadata
		IgnoreMetadataChangesConflictActions[im.IgnoredMetadata.String()] = ignoredMetadata[i].ConflictAction
	}

	vcdClient, err := config.Client()
	if err != nil {
		return nil, diag.FromErr(err)
	}
	return vcdClient, providerDiagnostics
}

// vcdSchemaFilter is a function which allows to filters and export type 'map[string]*schema.Resource' which may hold
// Terraform's native resource or data source list
// When 'nameRegexp' is not empty - it will return only those matching the regexp
// When 'includeDeprecated' is false - it will skip out the resources which have a DeprecationMessage set
func vcdSchemaFilter(schemaMap map[string]*schema.Resource, nameRegexp string, includeDeprecated bool) (map[string]*schema.Resource, error) {
	var (
		err error
		re  *regexp.Regexp
	)
	filteredResources := make(map[string]*schema.Resource)

	// validate regex if it was provided
	if nameRegexp != "" {
		re, err = regexp.Compile(nameRegexp)
		if err != nil {
			return nil, fmt.Errorf("unable to compile regexp: %s", err)
		}
	}

	// copy the map with filtering out unwanted object
	for resourceName, schemaResource := range schemaMap {

		// Skip deprecated resources if it was requested so
		if !includeDeprecated && schemaResource.DeprecationMessage != "" {
			continue
		}
		// If regex was defined - try to filter based on it
		if re != nil {
			// if it does not match regex - skip it
			doesNotmatchRegex := !re.MatchString(resourceName)
			if doesNotmatchRegex {
				continue
			}

		}

		filteredResources[resourceName] = schemaResource
	}

	return filteredResources, nil
}

func validateProviderSchema(d *schema.ResourceData) error {

	// Validate org and sys org
	sysOrg := d.Get("sysorg").(string)
	org := d.Get("org").(string)
	if sysOrg == "" && org == "" {
		return fmt.Errorf(`both "org" and "sysorg" properties are empty`)
	}

	return nil
}
