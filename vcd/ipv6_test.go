//go:build catalog || ALL || functional

package vcd

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcdIpv6Support(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	params := StringMap{
		"NsxtManager":         testConfig.Nsxt.Manager,
		"NsxtTier0Router":     testConfig.Nsxt.Tier0router,
		"Org":                 testConfig.VCD.Org,
		"Vdc":                 testConfig.Nsxt.Vdc,
		"ExternalNetworkName": t.Name(),
		"TestName":            t.Name(),
		"NsxtImportSegment":   testConfig.Nsxt.NsxtImportSegment,

		"NetworkName": t.Name(),
		"VdcName":     t.Name(),

		"ProviderVdc":               testConfig.VCD.NsxtProviderVdc.Name,
		"NetworkPool":               testConfig.VCD.NsxtProviderVdc.NetworkPool,
		"ProviderVdcStorageProfile": testConfig.VCD.NsxtProviderVdc.StorageProfile,
		"EdgeCluster":               testConfig.Nsxt.NsxtEdgeCluster,

		"Tags": "network extnetwork nsxt",
	}
	testParamsNotEmpty(t, params)

	configText := templateFill(testAccVcdIpv6, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckCatalogDestroy,
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_org_vdc.with-edge-cluster", "id"),
					resource.TestCheckResourceAttrSet("vcd_external_network_v2.ext-net-nsxt", "id"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_external_network_v2.ext-net-nsxt", "ip_scope.*", map[string]string{
						"enabled":       "true",
						"gateway":       "2002:0:0:1234:abcd:ffff:c0a8:101",
						"prefix_length": "124",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_external_network_v2.ext-net-nsxt", "ip_scope.*.static_ip_pool.*", map[string]string{
						"start_address": "2002:0:0:1234:abcd:ffff:c0a8:103",
						"end_address":   "2002:0:0:1234:abcd:ffff:c0a8:104",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_external_network_v2.ext-net-nsxt", "ip_scope.*.static_ip_pool.*", map[string]string{
						"start_address": "2002:0:0:1234:abcd:ffff:c0a8:107",
						"end_address":   "2002:0:0:1234:abcd:ffff:c0a8:109",
					}),
					resource.TestCheckResourceAttrSet("vcd_nsxt_edgegateway.nsxt-edge", "id"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_edgegateway.nsxt-edge", "subnet.*", map[string]string{
						"gateway":       "2002:0:0:1234:abcd:ffff:c0a8:101",
						"prefix_length": "124",
					}),
					resource.TestCheckResourceAttrSet("vcd_network_routed_v2.ipv6", "id"),
					resource.TestCheckResourceAttr("vcd_network_routed_v2.ipv6", "gateway", "2002:0:0:1234:abcd:ffff:c0a7:121"),
					resource.TestCheckResourceAttr("vcd_network_routed_v2.ipv6", "prefix_length", "124"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_network_routed_v2.ipv6", "static_ip_pool.*", map[string]string{
						"start_address": "2002:0:0:1234:abcd:ffff:c0a7:122",
						"end_address":   "2002:0:0:1234:abcd:ffff:c0a7:124",
					}),

					resource.TestCheckResourceAttrSet("vcd_network_routed_v2.ipv6-dualstack", "id"),
					resource.TestCheckResourceAttr("vcd_network_routed_v2.ipv6-dualstack", "gateway", "192.168.1.1"),
					resource.TestCheckResourceAttr("vcd_network_routed_v2.ipv6-dualstack", "prefix_length", "24"),
					resource.TestCheckResourceAttr("vcd_network_routed_v2.ipv6-dualstack", "dual_stack_enabled", "true"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_network_routed_v2.ipv6-dualstack", "static_ip_pool.*", map[string]string{
						"start_address": "192.168.1.10",
						"end_address":   "192.168.1.20",
					}),
					resource.TestCheckResourceAttr("vcd_network_routed_v2.ipv6-dualstack", "secondary_gateway", "2002:0:0:1234:abcd:ffff:c0a6:121"),
					resource.TestCheckResourceAttr("vcd_network_routed_v2.ipv6-dualstack", "secondary_prefix_length", "124"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_network_routed_v2.ipv6-dualstack", "secondary_static_ip_pool.*", map[string]string{
						"start_address": "2002:0:0:1234:abcd:ffff:c0a6:122",
						"end_address":   "2002:0:0:1234:abcd:ffff:c0a6:124",
					}),
					resource.TestCheckResourceAttrSet("vcd_nsxt_network_dhcp.routed-ipv6-dual-stack", "id"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_dhcp.routed-ipv6-dual-stack", "listener_ip_address", "2002:0:0:1234:abcd:ffff:c0a6:129"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_network_dhcp.routed-ipv6-dual-stack", "pool.*", map[string]string{
						"start_address": "2002:0:0:1234:abcd:ffff:c0a6:125",
						"end_address":   "2002:0:0:1234:abcd:ffff:c0a6:126",
					}),

					resource.TestCheckResourceAttrSet("vcd_nsxt_network_dhcp_binding.ipv6-binding1", "id"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_dhcp_binding.ipv6-binding1", "ip_address", "2002:0:0:1234:abcd:ffff:c0a6:127"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_network_dhcp_binding.ipv6-binding1", "dhcp_v6_config.*.sntp_servers.*", "4b0d:74eb:ee01:0ff4:ab1b:f7cc:4d74:d2a3"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_network_dhcp_binding.ipv6-binding1", "dhcp_v6_config.*.sntp_servers.*", "cc80:5498:18da:0883:d78a:4e4b:754d:df47"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_network_dhcp_binding.ipv6-binding1", "dhcp_v6_config.*.domain_names.*", "non-existing.org.tld"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_network_dhcp_binding.ipv6-binding1", "dhcp_v6_config.*.domain_names.*", "fake.org.tld"),

					resource.TestCheckResourceAttrSet("vcd_nsxt_network_dhcp_binding.ipv6-binding2", "id"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_dhcp_binding.ipv6-binding2", "ip_address", "2002:0:0:1234:abcd:ffff:c0a6:128"),

					resource.TestCheckResourceAttrSet("vcd_network_isolated_v2.ipv6", "id"),
					resource.TestCheckResourceAttr("vcd_network_isolated_v2.ipv6", "gateway", "2002:0:0:1234:abcd:ffff:c0a8:121"),
					resource.TestCheckResourceAttr("vcd_network_isolated_v2.ipv6", "prefix_length", "124"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_network_isolated_v2.ipv6", "static_ip_pool.*", map[string]string{
						"start_address": "2002:0:0:1234:abcd:ffff:c0a8:122",
						"end_address":   "2002:0:0:1234:abcd:ffff:c0a8:123",
					}),

					resource.TestCheckResourceAttrSet("vcd_network_isolated_v2.ipv6-dualstack", "id"),
					resource.TestCheckResourceAttr("vcd_network_isolated_v2.ipv6-dualstack", "gateway", "192.168.1.1"),
					resource.TestCheckResourceAttr("vcd_network_isolated_v2.ipv6-dualstack", "prefix_length", "24"),
					resource.TestCheckResourceAttr("vcd_network_isolated_v2.ipv6-dualstack", "dual_stack_enabled", "true"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_network_isolated_v2.ipv6-dualstack", "static_ip_pool.*", map[string]string{
						"start_address": "192.168.1.10",
						"end_address":   "192.168.1.20",
					}),
					resource.TestCheckResourceAttr("vcd_network_isolated_v2.ipv6-dualstack", "secondary_gateway", "2002:0:0:1234:abcd:ffff:c0a6:121"),
					resource.TestCheckResourceAttr("vcd_network_isolated_v2.ipv6-dualstack", "secondary_prefix_length", "124"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_network_isolated_v2.ipv6-dualstack", "secondary_static_ip_pool.*", map[string]string{
						"start_address": "2002:0:0:1234:abcd:ffff:c0a6:122",
						"end_address":   "2002:0:0:1234:abcd:ffff:c0a6:124",
					}),

					resource.TestCheckResourceAttrSet("vcd_nsxt_network_imported.ipv6-dualstack", "id"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_imported.ipv6-dualstack", "gateway", "192.168.1.1"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_imported.ipv6-dualstack", "prefix_length", "24"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_imported.ipv6-dualstack", "dual_stack_enabled", "true"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_network_imported.ipv6-dualstack", "static_ip_pool.*", map[string]string{
						"start_address": "192.168.1.10",
						"end_address":   "192.168.1.20",
					}),
					resource.TestCheckResourceAttr("vcd_nsxt_network_imported.ipv6-dualstack", "secondary_gateway", "2002:0:0:1234:abcd:ffff:c0a6:121"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_imported.ipv6-dualstack", "secondary_prefix_length", "124"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_network_imported.ipv6-dualstack", "secondary_static_ip_pool.*", map[string]string{
						"start_address": "2002:0:0:1234:abcd:ffff:c0a6:122",
						"end_address":   "2002:0:0:1234:abcd:ffff:c0a6:124",
					}),

					resource.TestCheckResourceAttrSet("vcd_nsxt_ip_set.ipv6", "id"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ip_set.ipv6", "ip_addresses.*", "2001:db8::/48"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ip_set.ipv6", "ip_addresses.*", "2001:db6:0:0:0:0:0:0-2001:db6:0:ffff:ffff:ffff:ffff:ffff"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ip_set.ipv6", "ip_addresses.*", "2002:0:0:1234:abcd:ffff:c0a8:120/124"),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdIpv6Prerequisites = `
data "vcd_provider_vdc" "pvdc" {
  name = "{{.ProviderVdc}}"
}

data "vcd_nsxt_edge_cluster" "ec" {
	provider_vdc_id = data.vcd_provider_vdc.pvdc.id
	name            = "{{.EdgeCluster}}"
}

resource "vcd_org_vdc" "with-edge-cluster" {
  name = "{{.VdcName}}"
  org  = "{{.Org}}"

  allocation_model  = "ReservationPool"
  network_pool_name = "{{.NetworkPool}}"
  provider_vdc_name = data.vcd_provider_vdc.pvdc.name
  network_quota     = 5

  edge_cluster_id = data.vcd_nsxt_edge_cluster.ec.id

  compute_capacity {
    cpu {
      allocated = 1024
      limit     = 1024
    }

    memory {
      allocated = 1024
      limit     = 1024
    }
  }

  storage_profile {
    name     = "{{.ProviderVdcStorageProfile}}"
    enabled  = true
    limit    = 10240
    default  = true
  }

  enabled                    = true
  enable_thin_provisioning   = true
  enable_fast_provisioning   = true
  delete_force               = true
  delete_recursive           = true
}
`

const testAccVcdIpv6 = testAccVcdIpv6Prerequisites + `
data "vcd_nsxt_manager" "main" {
  name = "{{.NsxtManager}}"
}

data "vcd_nsxt_tier0_router" "router" {
  name            = "{{.NsxtTier0Router}}"
  nsxt_manager_id = data.vcd_nsxt_manager.main.id
}

resource "vcd_external_network_v2" "ext-net-nsxt" {
  name        = "{{.ExternalNetworkName}}"
  description = "IPv6"

  nsxt_network {
    nsxt_manager_id      = data.vcd_nsxt_manager.main.id
    nsxt_tier0_router_id = data.vcd_nsxt_tier0_router.router.id
  }

  ip_scope {
    gateway       = "2002:0:0:1234:abcd:ffff:c0a8:101"
    prefix_length = "124"

    static_ip_pool {
      start_address = "2002:0:0:1234:abcd:ffff:c0a8:103"
      end_address   = "2002:0:0:1234:abcd:ffff:c0a8:104"
    }
    
    static_ip_pool {
      start_address = "2002:0:0:1234:abcd:ffff:c0a8:107"
      end_address   = "2002:0:0:1234:abcd:ffff:c0a8:109"
    }
  }
}

resource "vcd_nsxt_edgegateway" "nsxt-edge" {
  org         = "{{.Org}}"
  owner_id    = vcd_org_vdc.with-edge-cluster.id
  name        = "{{.TestName}}-edge-gateway"

  external_network_id = vcd_external_network_v2.ext-net-nsxt.id

  subnet {
     gateway       = tolist(vcd_external_network_v2.ext-net-nsxt.ip_scope)[0].gateway
     prefix_length = tolist(vcd_external_network_v2.ext-net-nsxt.ip_scope)[0].prefix_length

     primary_ip = tolist(tolist(vcd_external_network_v2.ext-net-nsxt.ip_scope)[0].static_ip_pool)[0].end_address
     allocated_ips {
       start_address = tolist(tolist(vcd_external_network_v2.ext-net-nsxt.ip_scope)[0].static_ip_pool)[0].end_address
       end_address   = tolist(tolist(vcd_external_network_v2.ext-net-nsxt.ip_scope)[0].static_ip_pool)[0].end_address
     }
  }
}

resource "vcd_network_routed_v2" "ipv6" {
  org  = "{{.Org}}"
  name = "{{.TestName}}-routed"

  edge_gateway_id = vcd_nsxt_edgegateway.nsxt-edge.id
  
  gateway       = "2002:0:0:1234:abcd:ffff:c0a7:121"
  prefix_length = 124

  static_ip_pool {
    start_address = "2002:0:0:1234:abcd:ffff:c0a7:122"
    end_address   = "2002:0:0:1234:abcd:ffff:c0a7:124"
  }
}

resource "vcd_network_routed_v2" "ipv6-dualstack" {
  org  = "{{.Org}}"
  name = "{{.TestName}}-routed-dualstack"

  edge_gateway_id = vcd_nsxt_edgegateway.nsxt-edge.id
  
  gateway       = "192.168.1.1"
  prefix_length = 24
  static_ip_pool {
    start_address = "192.168.1.10"
    end_address   = "192.168.1.20"
  }

  dual_stack_enabled      = true
  secondary_gateway       = "2002:0:0:1234:abcd:ffff:c0a6:121"
  secondary_prefix_length = 124

  secondary_static_ip_pool {
    start_address = "2002:0:0:1234:abcd:ffff:c0a6:122"
    end_address   = "2002:0:0:1234:abcd:ffff:c0a6:124"
  }
}

resource "vcd_nsxt_edgegateway_dhcpv6" "test" {
  org             = "{{.Org}}"
  edge_gateway_id = vcd_nsxt_edgegateway.nsxt-edge.id

  mode    = "DHCPv6"
}

resource "vcd_nsxt_network_dhcp" "routed-ipv6-dual-stack" {
  org = "{{.Org}}"

  org_network_id      = vcd_network_routed_v2.ipv6-dualstack.id
  mode                = "NETWORK"
  listener_ip_address = "2002:0:0:1234:abcd:ffff:c0a6:129"

  pool {
    start_address = "2002:0:0:1234:abcd:ffff:c0a6:125"
    end_address   = "2002:0:0:1234:abcd:ffff:c0a6:126"
  }

  depends_on = [vcd_nsxt_edgegateway_dhcpv6.test]
}


resource "vcd_nsxt_network_dhcp_binding" "ipv6-binding1" {
  org = "{{.Org}}"

  org_network_id = vcd_nsxt_network_dhcp.routed-ipv6-dual-stack.id

  name         = "{{.TestName}}-DHCP Binding-1"
  binding_type = "IPV6"
  ip_address   = "2002:0:0:1234:abcd:ffff:c0a6:127"
  lease_time   = 3600
  mac_address  = "00:11:22:33:44:66"

  dhcp_v6_config {
    sntp_servers = ["4b0d:74eb:ee01:0ff4:ab1b:f7cc:4d74:d2a3","cc80:5498:18da:0883:d78a:4e4b:754d:df47"]
    domain_names = ["non-existing.org.tld","fake.org.tld"]
  }
}

resource "vcd_nsxt_network_dhcp_binding" "ipv6-binding2" {
  org = "{{.Org}}"

  org_network_id = vcd_nsxt_network_dhcp.routed-ipv6-dual-stack.id

  name         = "{{.TestName}}-DHCP Binding-2"
  binding_type = "IPV6"
  ip_address   = "2002:0:0:1234:abcd:ffff:c0a6:128"
  lease_time   = 3600
  mac_address  = "00:11:22:33:33:66"

}

resource "vcd_network_isolated_v2" "ipv6" {
  org      = "{{.Org}}"
  owner_id = vcd_org_vdc.with-edge-cluster.id

  name = "{{.TestName}}-isolated"

  gateway       = "2002:0:0:1234:abcd:ffff:c0a8:121"
  prefix_length = 124

  static_ip_pool {
    start_address = "2002:0:0:1234:abcd:ffff:c0a8:122"
    end_address   = "2002:0:0:1234:abcd:ffff:c0a8:123"
  }
}

resource "vcd_network_isolated_v2" "ipv6-dualstack" {
  org      = "{{.Org}}"
  owner_id = vcd_org_vdc.with-edge-cluster.id

  name = "{{.TestName}}-isolated-dualstack"

  gateway       = "192.168.1.1"
  prefix_length = 24
  static_ip_pool {
    start_address = "192.168.1.10"
    end_address   = "192.168.1.20"
  }

  dual_stack_enabled      = true
  secondary_gateway       = "2002:0:0:1234:abcd:ffff:c0a6:121"
  secondary_prefix_length = 124

  secondary_static_ip_pool {
    start_address = "2002:0:0:1234:abcd:ffff:c0a6:122"
    end_address   = "2002:0:0:1234:abcd:ffff:c0a6:124"
  }
}

resource "vcd_nsxt_network_imported" "ipv6-dualstack" {
  org  = "{{.Org}}"
  vdc  = vcd_org_vdc.with-edge-cluster.name
  name = "{{.TestName}}-imported-dualstack"

  nsxt_logical_switch_name = "{{.NsxtImportSegment}}"

  gateway       = "192.168.1.1"
  prefix_length = 24
  static_ip_pool {
    start_address = "192.168.1.10"
    end_address   = "192.168.1.20"
  }

  dual_stack_enabled      = true
  secondary_gateway       = "2002:0:0:1234:abcd:ffff:c0a6:121"
  secondary_prefix_length = 124

  secondary_static_ip_pool {
    start_address = "2002:0:0:1234:abcd:ffff:c0a6:122"
    end_address   = "2002:0:0:1234:abcd:ffff:c0a6:124"
  }
}

resource "vcd_nsxt_ip_set" "ipv6" {
  org             = "{{.Org}}"
  edge_gateway_id = vcd_nsxt_edgegateway.nsxt-edge.id
  name            = "{{.TestName}}-ipset"

  ip_addresses = [
    "2001:db8::/48",
    "2001:db6:0:0:0:0:0:0-2001:db6:0:ffff:ffff:ffff:ffff:ffff",
    "2002:0:0:1234:abcd:ffff:c0a8:120/124",
  ]
}
`
