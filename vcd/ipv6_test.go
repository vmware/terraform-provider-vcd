//go:build network || ALL || functional

package vcd

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

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
		"NsxtImportSegment2":  testConfig.Nsxt.NsxtImportSegment2,

		"VdcName": t.Name(),

		"ProviderVdc":               testConfig.VCD.NsxtProviderVdc.Name,
		"NetworkPool":               testConfig.VCD.NsxtProviderVdc.NetworkPool,
		"ProviderVdcStorageProfile": testConfig.VCD.NsxtProviderVdc.StorageProfile,
		"EdgeCluster":               testConfig.Nsxt.NsxtEdgeCluster,

		"Tags": "network extnetwork nsxt",
	}
	testParamsNotEmpty(t, params)

	configTex1 := templateFill(testAccVcdIpv6Step1, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configTex1)

	// Step 2 will test explicit 10.4.1+ feature - external network attachment for NSX-T Edge Gateway

	var skipVersionLessThan string
	if checkVersion(testConfig.Provider.ApiVersion, "< 37.1") {
		skipVersionLessThan = "This test step requires VCD 10.4.1+ (API V37.1+) Skipping."
		if vcdShortTest {
			t.Skip(acceptanceTestsSkipped)
			return
		}
	}

	params["FuncName"] = t.Name() + "-step2"
	configTex2 := templateFill(testAccVcdIpv6Step2, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configTex2)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVcdDestroy(t.Name()),
		Steps: []resource.TestStep{
			{
				Config: configTex1,
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
			{
				Config: configTex2,
				SkipFunc: func() (bool, error) {
					if skipVersionLessThan != "" {
						fmt.Println(skipVersionLessThan)
						return true, nil
					}
					return false, nil
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_edgegateway.nsxt-edge", "external_network.*", map[string]string{
						"gateway":            "2002:0:0:1234:abbd:ffff:c0a6:121",
						"prefix_length":      "124",
						"primary_ip":         "2002:0:0:1234:abbd:ffff:c0a6:122",
						"allocated_ip_count": "2",
					}),
				),
			},
		},
	})
	postTestChecks(t)
}

func testAccCheckVcdDestroy(vdcName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*VCDClient)
		org, err := conn.GetOrg(testConfig.VCD.Org)
		if err != nil {
			return fmt.Errorf(errorRetrievingOrg, err)
		}
		_, err = org.GetVDCByName(vdcName, true)
		if err == nil {
			return fmt.Errorf("VDC %s still exists", vdcName)
		}
		return nil
	}
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

const testAccVcdIpv6EdgeGateway = `
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
`

const testAccVcdIpv6networks = `
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

const testAccVcdIpv6Step1 = testAccVcdIpv6Prerequisites + testAccVcdIpv6networks + testAccVcdIpv6EdgeGateway
const testAccVcdIpv6Step2 = testAccVcdIpv6Prerequisites + testAccVcdIpv6networks + testAccVcdIpv6EdgeGatewayWithExternalNetworks

const testAccVcdIpv6EdgeGatewayWithExternalNetworks = `
resource "vcd_external_network_v2" "segment-backed" {
  name = "{{.TestName}}-segment"

  nsxt_network {
    nsxt_manager_id   = data.vcd_nsxt_manager.main.id
    nsxt_segment_name = "{{.NsxtImportSegment2}}"
  }

  ip_scope {
    gateway       = "2002:0:0:1234:abbd:ffff:c0a6:121"
    prefix_length = "124"

    static_ip_pool {
      start_address = "2002:0:0:1234:abbd:ffff:c0a6:122"
      end_address   = "2002:0:0:1234:abbd:ffff:c0a6:124"
    }
  }
}

resource "vcd_nsxt_edgegateway" "nsxt-edge" {
  org      = "{{.Org}}"
  owner_id = vcd_org_vdc.with-edge-cluster.id
  name     = "{{.TestName}}-edge-gateway"

  external_network_id = vcd_external_network_v2.ext-net-nsxt.id

  subnet {
     gateway       = tolist(vcd_external_network_v2.ext-net-nsxt.ip_scope)[0].gateway
     prefix_length = tolist(vcd_external_network_v2.ext-net-nsxt.ip_scope)[0].prefix_length
     primary_ip    = tolist(tolist(vcd_external_network_v2.ext-net-nsxt.ip_scope)[0].static_ip_pool)[0].end_address

     allocated_ips {
       start_address = tolist(tolist(vcd_external_network_v2.ext-net-nsxt.ip_scope)[0].static_ip_pool)[0].end_address
       end_address   = tolist(tolist(vcd_external_network_v2.ext-net-nsxt.ip_scope)[0].static_ip_pool)[0].end_address
     }
  }

  external_network {
    external_network_id = vcd_external_network_v2.segment-backed.id
    gateway             = tolist(vcd_external_network_v2.segment-backed.ip_scope)[0].gateway
    prefix_length       = tolist(vcd_external_network_v2.segment-backed.ip_scope)[0].prefix_length
    allocated_ip_count  = 2
  }
}
`

func TestAccVcdIpv6SupportLargeSubnet(t *testing.T) {
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
		"NsxtImportSegment2":  testConfig.Nsxt.NsxtImportSegment2,

		"VdcName": t.Name(),

		"ProviderVdc":               testConfig.VCD.NsxtProviderVdc.Name,
		"NetworkPool":               testConfig.VCD.NsxtProviderVdc.NetworkPool,
		"ProviderVdcStorageProfile": testConfig.VCD.NsxtProviderVdc.StorageProfile,
		"EdgeCluster":               testConfig.Nsxt.NsxtEdgeCluster,

		"Tags": "network extnetwork nsxt",
	}
	testParamsNotEmpty(t, params)

	configTex1 := templateFill(testAccVcdIpv6LargeSubnetStep1, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configTex1)

	// Step 2 will test explicit 10.4.1+ feature - external network attachment for NSX-T Edge Gateway

	var skipVersionLessThan string
	if checkVersion(testConfig.Provider.ApiVersion, "< 37.1") {
		skipVersionLessThan = "This test step requires VCD 10.4.1+ (API V37.1+) Skipping."
		if vcdShortTest {
			t.Skip(acceptanceTestsSkipped)
			return
		}
	}

	params["FuncName"] = t.Name() + "-step2"
	configTex2 := templateFill(testAccVcdIpv6LargeSubnetStep2, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configTex2)

	params["FuncName"] = t.Name() + "-step3"
	configTex3DS := templateFill(testAccVcdIpv6LargeSubnetStep2DS, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configTex3DS)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVcdDestroy(t.Name()),
		Steps: []resource.TestStep{
			{
				Config: configTex1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_org_vdc.with-edge-cluster", "id"),
					resource.TestCheckResourceAttrSet("vcd_external_network_v2.ext-net-nsxt", "id"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_external_network_v2.ext-net-nsxt", "ip_scope.*", map[string]string{
						"enabled":       "true",
						"gateway":       "2a02:a404:11:0:ffff:ffff:ffff:ffff",
						"prefix_length": "64",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_external_network_v2.ext-net-nsxt", "ip_scope.*.static_ip_pool.*", map[string]string{
						"start_address": "2a02:a404:11:0:0:0:0:1",
						"end_address":   "2a02:a404:11:0:ffff:ffff:ffff:fffd",
					}),

					resource.TestCheckResourceAttrSet("vcd_nsxt_edgegateway.nsxt-edge", "id"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "ip_count_read_limit", strconv.Itoa(defaultReadLimitOfUnusedIps)),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "unused_ip_count", "999999"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "used_ip_count", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_edgegateway.nsxt-edge", "subnet.*", map[string]string{
						"gateway":       "2a02:a404:11:0:ffff:ffff:ffff:ffff",
						"prefix_length": "64",
					}),
					resource.TestCheckResourceAttrSet("vcd_network_routed_v2.ipv6", "id"),
					resource.TestCheckResourceAttr("vcd_network_routed_v2.ipv6", "gateway", "3a02:a404:11:0:ffff:ffff:ffff:ffff"),
					resource.TestCheckResourceAttr("vcd_network_routed_v2.ipv6", "prefix_length", "64"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_network_routed_v2.ipv6", "static_ip_pool.*", map[string]string{
						"start_address": "3a02:a404:11:0:0:0:0:1",
						"end_address":   "3a02:a404:11:0:ffff:ffff:ffff:fffd",
					}),

					resource.TestCheckResourceAttrSet("vcd_network_routed_v2.ipv6-dualstack", "id"),
					resource.TestCheckResourceAttr("vcd_network_routed_v2.ipv6-dualstack", "gateway", "192.168.1.1"),
					resource.TestCheckResourceAttr("vcd_network_routed_v2.ipv6-dualstack", "prefix_length", "24"),
					resource.TestCheckResourceAttr("vcd_network_routed_v2.ipv6-dualstack", "dual_stack_enabled", "true"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_network_routed_v2.ipv6-dualstack", "static_ip_pool.*", map[string]string{
						"start_address": "192.168.1.10",
						"end_address":   "192.168.1.20",
					}),
					resource.TestCheckResourceAttr("vcd_network_routed_v2.ipv6-dualstack", "secondary_gateway", "5a02:a404:11:0:ffff:ffff:ffff:ffff"),
					resource.TestCheckResourceAttr("vcd_network_routed_v2.ipv6-dualstack", "secondary_prefix_length", "64"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_network_routed_v2.ipv6-dualstack", "secondary_static_ip_pool.*", map[string]string{
						"start_address": "5a02:a404:11:0:0:0:0:130",
						"end_address":   "5a02:a404:11:0:ffff:ffff:ffff:fffd",
					}),
					resource.TestCheckResourceAttrSet("vcd_nsxt_network_dhcp.routed-ipv6-dual-stack", "id"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_dhcp.routed-ipv6-dual-stack", "listener_ip_address", "5a02:a404:11:0:0:0:0:129"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_network_dhcp.routed-ipv6-dual-stack", "pool.*", map[string]string{
						"start_address": "5a02:a404:11:0:0:0:0:125",
						"end_address":   "5a02:a404:11:0:0:0:0:126",
					}),

					resource.TestCheckResourceAttrSet("vcd_nsxt_network_dhcp_binding.ipv6-binding1", "id"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_dhcp_binding.ipv6-binding1", "ip_address", "5a02:a404:11:0:0:0:0:127"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_network_dhcp_binding.ipv6-binding1", "dhcp_v6_config.*.sntp_servers.*", "4b0d:74eb:ee01:0ff4:ab1b:f7cc:4d74:d2a3"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_network_dhcp_binding.ipv6-binding1", "dhcp_v6_config.*.sntp_servers.*", "cc80:5498:18da:0883:d78a:4e4b:754d:df47"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_network_dhcp_binding.ipv6-binding1", "dhcp_v6_config.*.domain_names.*", "non-existing.org.tld"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_network_dhcp_binding.ipv6-binding1", "dhcp_v6_config.*.domain_names.*", "fake.org.tld"),

					resource.TestCheckResourceAttrSet("vcd_nsxt_network_dhcp_binding.ipv6-binding2", "id"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_dhcp_binding.ipv6-binding2", "ip_address", "5a02:a404:11:0:0:0:0:128"),

					resource.TestCheckResourceAttrSet("vcd_network_isolated_v2.ipv6", "id"),
					resource.TestCheckResourceAttr("vcd_network_isolated_v2.ipv6", "gateway", "6a02:a404:11:0:ffff:ffff:ffff:ffff"),
					resource.TestCheckResourceAttr("vcd_network_isolated_v2.ipv6", "prefix_length", "64"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_network_isolated_v2.ipv6", "static_ip_pool.*", map[string]string{
						"start_address": "6a02:a404:11:0:0:0:0:1",
						"end_address":   "6a02:a404:11:0:ffff:ffff:ffff:fffd",
					}),

					resource.TestCheckResourceAttrSet("vcd_network_isolated_v2.ipv6-dualstack", "id"),
					resource.TestCheckResourceAttr("vcd_network_isolated_v2.ipv6-dualstack", "gateway", "192.168.1.1"),
					resource.TestCheckResourceAttr("vcd_network_isolated_v2.ipv6-dualstack", "prefix_length", "24"),
					resource.TestCheckResourceAttr("vcd_network_isolated_v2.ipv6-dualstack", "dual_stack_enabled", "true"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_network_isolated_v2.ipv6-dualstack", "static_ip_pool.*", map[string]string{
						"start_address": "192.168.1.10",
						"end_address":   "192.168.1.20",
					}),
					resource.TestCheckResourceAttr("vcd_network_isolated_v2.ipv6-dualstack", "secondary_gateway", "8a02:a404:11:0:ffff:ffff:ffff:ffff"),
					resource.TestCheckResourceAttr("vcd_network_isolated_v2.ipv6-dualstack", "secondary_prefix_length", "64"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_network_isolated_v2.ipv6-dualstack", "secondary_static_ip_pool.*", map[string]string{
						"start_address": "8a02:a404:11:0:0:0:0:122",
						"end_address":   "8a02:a404:11:0:ffff:ffff:ffff:124",
					}),

					resource.TestCheckResourceAttrSet("vcd_nsxt_network_imported.ipv6-dualstack", "id"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_imported.ipv6-dualstack", "gateway", "192.168.1.1"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_imported.ipv6-dualstack", "prefix_length", "24"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_imported.ipv6-dualstack", "dual_stack_enabled", "true"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_network_imported.ipv6-dualstack", "static_ip_pool.*", map[string]string{
						"start_address": "192.168.1.10",
						"end_address":   "192.168.1.20",
					}),
					resource.TestCheckResourceAttr("vcd_nsxt_network_imported.ipv6-dualstack", "secondary_gateway", "9a02:a404:11:0:ffff:ffff:ffff:ffff"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_imported.ipv6-dualstack", "secondary_prefix_length", "64"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_network_imported.ipv6-dualstack", "secondary_static_ip_pool.*", map[string]string{
						"start_address": "9a02:a404:11:0:0:0:0:122",
						"end_address":   "9a02:a404:11:0:0:0:0:124",
					}),

					resource.TestCheckResourceAttrSet("vcd_nsxt_ip_set.ipv6", "id"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ip_set.ipv6", "ip_addresses.*", "2001:db8::/48"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ip_set.ipv6", "ip_addresses.*", "2001:db6:0:0:0:0:0:0-2001:db6:0:ffff:ffff:ffff:ffff:ffff"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ip_set.ipv6", "ip_addresses.*", "2002:0:0:1234:abcd:ffff:c0a8:120/124"),
				),
			},
			{
				Config: configTex2,
				SkipFunc: func() (bool, error) {
					if skipVersionLessThan != "" {
						fmt.Println(skipVersionLessThan)
						return true, nil
					}
					return false, nil
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_edgegateway.nsxt-edge", "external_network.*", map[string]string{
						"gateway":            "2024:0:0:1234:abbd:ffff:c0a6:121",
						"prefix_length":      "96",
						"primary_ip":         "2024:0:0:1234:abbd:ffff:0:10",
						"allocated_ip_count": "2",
					}),
				),
			},
			{
				Config: configTex3DS,
				SkipFunc: func() (bool, error) {
					if skipVersionLessThan != "" {
						fmt.Println(skipVersionLessThan)
						return true, nil
					}
					return false, nil
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "unused_ip_count", "199998"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "ip_count_read_limit", "200000"), // 200k

					resource.TestCheckResourceAttr("data.vcd_nsxt_edgegateway.nsxt-edge", "ip_count_read_limit", strconv.Itoa(defaultReadLimitOfUnusedIps)),
					resource.TestCheckResourceAttr("data.vcd_nsxt_edgegateway.nsxt-edge", "unused_ip_count", "999998"), // data source does not have limit

					resourceFieldsEqual("vcd_nsxt_edgegateway.nsxt-edge", "data.vcd_nsxt_edgegateway.nsxt-edge", []string{"%", "unused_ip_count", "ip_count_read_limit"}),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdIpv6LargeSubnetStep1 = testAccVcdIpv6Prerequisites + testAccVcdIpv6networksLargeSubnets + testAccVcdIpv6EdgeGatewayLargeSubnet
const testAccVcdIpv6LargeSubnetStep2 = testAccVcdIpv6Prerequisites + testAccVcdIpv6networksLargeSubnets + testAccVcdIpv6EdgeGatewayLargeSubnetWithExternalNetworks
const testAccVcdIpv6LargeSubnetStep2DS = testAccVcdIpv6LargeSubnetStep2 + `
# skip-binary-test: Data source test
data "vcd_nsxt_edgegateway" "nsxt-edge" {
  org         = "{{.Org}}"
  owner_id    = vcd_org_vdc.with-edge-cluster.id
  name        = vcd_nsxt_edgegateway.nsxt-edge.name

  depends_on = [vcd_nsxt_edgegateway.nsxt-edge]
}
`

const testAccVcdIpv6networksLargeSubnets = `
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
    gateway       = "2a02:a404:11:0:ffff:ffff:ffff:ffff"
    prefix_length = "64"

    static_ip_pool {
      start_address = "2a02:a404:11:0:0:0:0:1"
      end_address   = "2a02:a404:11:0:ffff:ffff:ffff:fffd"
    }
  }
}

resource "vcd_network_routed_v2" "ipv6" {
  org  = "{{.Org}}"
  name = "{{.TestName}}-routed"

  edge_gateway_id = vcd_nsxt_edgegateway.nsxt-edge.id
  
  gateway       = "3a02:a404:11:0:ffff:ffff:ffff:ffff"
  prefix_length = 64

  static_ip_pool {
    start_address = "3a02:a404:11:0:0:0:0:1"
    end_address   = "3a02:a404:11:0:ffff:ffff:ffff:fffd"
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
  secondary_gateway       = "5a02:a404:11:0:ffff:ffff:ffff:ffff"
  secondary_prefix_length = 64

  secondary_static_ip_pool {
    start_address = "5a02:a404:11:0:0:0:0:130"
    end_address   = "5a02:a404:11:0:ffff:ffff:ffff:fffd"
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
  listener_ip_address = "5a02:a404:11:0:0:0:0:129"

  pool {
    start_address = "5a02:a404:11:0:0:0:0:125"
    end_address   = "5a02:a404:11:0:0:0:0:126"
  }

  depends_on = [vcd_nsxt_edgegateway_dhcpv6.test]
}


resource "vcd_nsxt_network_dhcp_binding" "ipv6-binding1" {
  org = "{{.Org}}"

  org_network_id = vcd_nsxt_network_dhcp.routed-ipv6-dual-stack.id

  name         = "{{.TestName}}-DHCP Binding-1"
  binding_type = "IPV6"
  ip_address   = "5a02:a404:11:0:0:0:0:127"
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
  ip_address   = "5a02:a404:11:0:0:0:0:128"
  lease_time   = 3600
  mac_address  = "00:11:22:33:33:66"

}

resource "vcd_network_isolated_v2" "ipv6" {
  org      = "{{.Org}}"
  owner_id = vcd_org_vdc.with-edge-cluster.id

  name = "{{.TestName}}-isolated"

  gateway       = "6a02:a404:11:0:ffff:ffff:ffff:ffff"
  prefix_length = 64

  static_ip_pool {
    start_address = "6a02:a404:11:0:0:0:0:1"
    end_address   = "6a02:a404:11:0:ffff:ffff:ffff:fffd"
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
  secondary_gateway       = "8a02:a404:11:0:ffff:ffff:ffff:ffff"
  secondary_prefix_length = 64

  secondary_static_ip_pool {
    start_address = "8a02:a404:11:0:0:0:0:122"
    end_address   = "8a02:a404:11:0:ffff:ffff:ffff:124"
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
  secondary_gateway       = "9a02:a404:11:0:ffff:ffff:ffff:ffff"
  secondary_prefix_length = 64

  secondary_static_ip_pool {
    start_address = "9a02:a404:11:0:0:0:0:122"
    end_address   = "9a02:a404:11:0:0:0:0:124"
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

const testAccVcdIpv6EdgeGatewayLargeSubnet = `
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
       start_address = tolist(tolist(vcd_external_network_v2.ext-net-nsxt.ip_scope)[0].static_ip_pool)[0].start_address
       end_address   = tolist(tolist(vcd_external_network_v2.ext-net-nsxt.ip_scope)[0].static_ip_pool)[0].end_address
     }
  }
}
`

const testAccVcdIpv6EdgeGatewayLargeSubnetWithExternalNetworks = `
resource "vcd_external_network_v2" "segment-backed" {
  name = "{{.TestName}}-segment"

  nsxt_network {
    nsxt_manager_id   = data.vcd_nsxt_manager.main.id
    nsxt_segment_name = "{{.NsxtImportSegment2}}"
  }

  ip_scope {
    gateway       = "2024:0:0:1234:abbd:ffff:c0a6:121"
    prefix_length = "96"

    static_ip_pool {
      start_address = "2024:0:0:1234:abbd:ffff:0:10"
      end_address   = "2024:0:0:1234:abbd:ffff:0:1000"
    }
  }
}

resource "vcd_nsxt_edgegateway" "nsxt-edge" {
  org         = "{{.Org}}"
  owner_id    = vcd_org_vdc.with-edge-cluster.id
  name        = "{{.TestName}}-edge-gateway"

  external_network_id = vcd_external_network_v2.ext-net-nsxt.id
  
  ip_count_read_limit = 200000 # 200k

  subnet {
     gateway       = tolist(vcd_external_network_v2.ext-net-nsxt.ip_scope)[0].gateway
     prefix_length = tolist(vcd_external_network_v2.ext-net-nsxt.ip_scope)[0].prefix_length

     primary_ip = tolist(tolist(vcd_external_network_v2.ext-net-nsxt.ip_scope)[0].static_ip_pool)[0].end_address
     allocated_ips {
       start_address = tolist(tolist(vcd_external_network_v2.ext-net-nsxt.ip_scope)[0].static_ip_pool)[0].start_address
       end_address   = tolist(tolist(vcd_external_network_v2.ext-net-nsxt.ip_scope)[0].static_ip_pool)[0].end_address
     }
  }

  external_network {
    external_network_id = vcd_external_network_v2.segment-backed.id
    gateway             = tolist(vcd_external_network_v2.segment-backed.ip_scope)[0].gateway
    prefix_length       = tolist(vcd_external_network_v2.segment-backed.ip_scope)[0].prefix_length
    allocated_ip_count  = 2
  }
}
`
