//go:build network || nsxt || ALL || functional

package vcd

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcdOpenApiDhcpNsxtRouted(t *testing.T) {
	preTestChecks(t)

	// String map to fill the template
	var params = StringMap{
		"Org":          testConfig.VCD.Org,
		"NsxtVdc":      testConfig.Nsxt.Vdc,
		"EdgeGw":       testConfig.Nsxt.EdgeGateway,
		"NetworkName":  t.Name(),
		"TestName":     t.Name(),
		"Binding1Name": t.Name() + "-dhcp-binding-1",
		"Tags":         "network nsxt",
	}
	testParamsNotEmpty(t, params)

	configText1 := templateFill(testAccRoutedNetDhcpStep1, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 0: %s", configText1)

	params["FuncName"] = t.Name() + "-step1"
	configText2 := templateFill(testAccRoutedNetDhcpStep2, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText2)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	vcdClient := createTemporaryVCDConnection(true)
	vcdVersionIsLowerThan1031 := func() (bool, error) {
		if vcdClient != nil && vcdClient.Client.APIVCDMaxVersionIs(">= 36.1") {
			return false, nil
		}
		return true, nil
	}

	// This case is specific for VCD 10.3.1 onwards since dns servers are not present in previous versions
	var configText3 string
	var configText4 string
	if vcdClient != nil && vcdClient.Client.APIVCDMaxVersionIs(">= 36.1") {
		params["SkipTest"] = "# skip-binary-test: VCD 10.3.1 onwards dns servers are not present in previous versions"
	}
	params["FuncName"] = t.Name() + "-step2"
	configText3 = templateFill(testAccRoutedNetDhcpStep3, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText3)

	params["FuncName"] = t.Name() + "-step3"
	configText4 = templateFill(testAccRoutedNetDhcpStep4, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText4)

	cacheDhcpBinding1dId := &testCachedFieldValue{}
	cacheDhcpBinding2dId := &testCachedFieldValue{}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckOpenApiVcdNetworkDestroy(testConfig.Nsxt.Vdc, "nsxt-routed-dhcp"),
		Steps: []resource.TestStep{
			{
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_network_dhcp.pools", "id", regexp.MustCompile(`^urn:vcloud:network:.*$`)),
					resource.TestCheckResourceAttr("vcd_nsxt_network_dhcp.pools", "lease_time", "86400"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_dhcp.pools", "mode", "EDGE"),
					resource.TestCheckNoResourceAttr("vcd_nsxt_network_dhcp.pools", "listener_ip_address"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_network_dhcp.pools", "pool.*", map[string]string{
						"start_address": "7.1.1.100",
						"end_address":   "7.1.1.110",
					}),
				),
			},
			{
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_network_dhcp.pools", "id", regexp.MustCompile(`^urn:vcloud:network:.*$`)),
					resource.TestCheckResourceAttr("vcd_nsxt_network_dhcp.pools", "lease_time", "4294967295"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_dhcp.pools", "mode", "EDGE"),
					resource.TestCheckNoResourceAttr("vcd_nsxt_network_dhcp.pools", "listener_ip_address"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_network_dhcp.pools", "pool.*", map[string]string{
						"start_address": "7.1.1.100",
						"end_address":   "7.1.1.110",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_network_dhcp.pools", "pool.*", map[string]string{
						"start_address": "7.1.1.130",
						"end_address":   "7.1.1.140",
					}),
				),
			},
			{
				ResourceName:            "vcd_nsxt_network_dhcp.pools",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdFunc:       importStateIdOrgNsxtVdcObject(params["NetworkName"].(string)),
				ImportStateVerifyIgnore: []string{"vdc"},
			},
			{
				Config:   configText3,
				SkipFunc: vcdVersionIsLowerThan1031,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_network_dhcp.pools", "id", regexp.MustCompile(`^urn:vcloud:network:.*$`)),
					resource.TestCheckResourceAttr("vcd_nsxt_network_dhcp.pools", "lease_time", "4294967295"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_dhcp.pools", "mode", "EDGE"),
					resource.TestCheckNoResourceAttr("vcd_nsxt_network_dhcp.pools", "listener_ip_address"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_network_dhcp.pools", "pool.*", map[string]string{
						"start_address": "7.1.1.100",
						"end_address":   "7.1.1.110",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_network_dhcp.pools", "pool.*", map[string]string{
						"start_address": "7.1.1.130",
						"end_address":   "7.1.1.140",
					}),
					resource.TestCheckResourceAttr("vcd_nsxt_network_dhcp.pools", "dns_servers.#", "2"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_dhcp.pools", "dns_servers.0", "1.1.1.1"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_dhcp.pools", "dns_servers.1", "1.0.0.1"),
					// DHCP binding checks
					cacheDhcpBinding1dId.cacheTestResourceFieldValue("vcd_nsxt_network_dhcp_binding.binding1", "id"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_network_dhcp_binding.binding1", "id"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_dhcp_binding.binding1", "name", params["Binding1Name"].(string)),
					resource.TestCheckResourceAttr("vcd_nsxt_network_dhcp_binding.binding1", "description", ""),
					resource.TestCheckResourceAttr("vcd_nsxt_network_dhcp_binding.binding1", "lease_time", "60"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_dhcp_binding.binding1", "binding_type", "IPV4"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_dhcp_binding.binding1", "dns_servers.#", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_dhcp_binding.binding1", "dhcp_v4_config.#", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_dhcp_binding.binding1", "ip_address", "7.1.1.189"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_dhcp_binding.binding1", "mac_address", "00:11:22:33:44:55"),

					cacheDhcpBinding2dId.cacheTestResourceFieldValue("vcd_nsxt_network_dhcp_binding.binding2", "id"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_network_dhcp_binding.binding2", "id"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_dhcp_binding.binding2", "name", t.Name()+"-dhcp-binding-2"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_dhcp_binding.binding2", "description", "DHCP binding description"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_dhcp_binding.binding2", "binding_type", "IPV4"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_dhcp_binding.binding2", "lease_time", "3600"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_dhcp_binding.binding2", "dns_servers.#", "2"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_dhcp_binding.binding2", "dns_servers.0", "7.1.1.242"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_dhcp_binding.binding2", "dns_servers.1", "7.1.1.243"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_dhcp_binding.binding2", "dhcp_v4_config.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_network_dhcp_binding.binding2", "dhcp_v4_config.*", map[string]string{
						"gateway_ip_address": "7.1.1.233",
						"hostname":           "non-existent",
					}),
					resource.TestCheckResourceAttr("vcd_nsxt_network_dhcp_binding.binding2", "ip_address", "7.1.1.190"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_dhcp_binding.binding2", "mac_address", "00:11:22:33:44:66"),
				),
			},
			{
				Config:   configText4,
				SkipFunc: vcdVersionIsLowerThan1031,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_network_dhcp.pools", "id", regexp.MustCompile(`^urn:vcloud:network:.*$`)),
					resource.TestCheckResourceAttr("vcd_nsxt_network_dhcp.pools", "lease_time", "4294967295"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_dhcp.pools", "mode", "EDGE"),
					resource.TestCheckNoResourceAttr("vcd_nsxt_network_dhcp.pools", "listener_ip_address"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_network_dhcp.pools", "pool.*", map[string]string{
						"start_address": "7.1.1.100",
						"end_address":   "7.1.1.110",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_network_dhcp.pools", "pool.*", map[string]string{
						"start_address": "7.1.1.130",
						"end_address":   "7.1.1.140",
					}),
					resource.TestCheckResourceAttr("vcd_nsxt_network_dhcp.pools", "dns_servers.#", "2"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_dhcp.pools", "dns_servers.0", "1.1.1.1"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_dhcp.pools", "dns_servers.1", "1.0.0.1"),

					// DHCP binding checks
					cacheDhcpBinding1dId.cacheTestResourceFieldValue("vcd_nsxt_network_dhcp_binding.binding1", "id"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_dhcp_binding.binding1", "name", params["Binding1Name"].(string)),
					resource.TestCheckResourceAttr("vcd_nsxt_network_dhcp_binding.binding1", "description", ""),
					resource.TestCheckResourceAttr("vcd_nsxt_network_dhcp_binding.binding1", "lease_time", "60"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_dhcp_binding.binding1", "binding_type", "IPV4"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_dhcp_binding.binding1", "dns_servers.#", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_dhcp_binding.binding1", "dhcp_v4_config.#", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_dhcp_binding.binding1", "ip_address", "7.1.1.167"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_dhcp_binding.binding1", "mac_address", "00:11:22:33:33:55"),

					cacheDhcpBinding2dId.cacheTestResourceFieldValue("vcd_nsxt_network_dhcp_binding.binding2", "id"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_network_dhcp_binding.binding2", "id"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_dhcp_binding.binding2", "name", t.Name()+"-dhcp-binding-2"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_dhcp_binding.binding2", "description", ""),
					resource.TestCheckResourceAttr("vcd_nsxt_network_dhcp_binding.binding2", "binding_type", "IPV4"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_dhcp_binding.binding2", "lease_time", "3600"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_dhcp_binding.binding2", "dns_servers.#", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_dhcp_binding.binding2", "dhcp_v4_config.#", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_dhcp_binding.binding2", "ip_address", "7.1.1.190"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_dhcp_binding.binding2", "mac_address", "00:11:22:33:44:66"),
				),
			},
			{
				SkipFunc:          vcdVersionIsLowerThan1031,
				ResourceName:      "vcd_nsxt_network_dhcp_binding.binding1",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importCustomObject([]string{params["Org"].(string), params["NsxtVdc"].(string), params["NetworkName"].(string), params["Binding1Name"].(string)}),
			},
		},
	})
	postTestChecks(t)
}

const testAccRoutedNetDhcpConfig = `
data "vcd_nsxt_edgegateway" "existing" {
  org  = "{{.Org}}"
  vdc  = "{{.NsxtVdc}}"
  name = "{{.EdgeGw}}"
}

resource "vcd_network_routed_v2" "net1" {
  org  = "{{.Org}}"
  vdc  = "{{.NsxtVdc}}"
  name = "{{.NetworkName}}"
  description = "NSX-T routed network for DHCP testing"

  edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id

  gateway       = "7.1.1.1"
  prefix_length = 24

  static_ip_pool {
    start_address = "7.1.1.10"
    end_address   = "7.1.1.20"
  }
}
`

const testAccRoutedNetDhcpStep1 = testAccRoutedNetDhcpConfig + `
resource "vcd_nsxt_network_dhcp" "pools" {
  org  = "{{.Org}}"
  vdc  = "{{.NsxtVdc}}"

  org_network_id = vcd_network_routed_v2.net1.id
  
  pool {
    start_address = "7.1.1.100"
    end_address   = "7.1.1.110"
  }
}
`

const testAccRoutedNetDhcpStep2 = testAccRoutedNetDhcpConfig + `
resource "vcd_nsxt_network_dhcp" "pools" {
  org  = "{{.Org}}"
  vdc  = "{{.NsxtVdc}}"

  org_network_id = vcd_network_routed_v2.net1.id
  mode           = "EDGE"
  lease_time     = 4294967295 # maximum allowed lease time in seconds (~49 days)
  
  pool {
    start_address = "7.1.1.100"
    end_address   = "7.1.1.110"
  }

  pool {
    start_address = "7.1.1.130"
    end_address   = "7.1.1.140"
  }
}
`

const testAccRoutedNetDhcpStep3 = testAccRoutedNetDhcpConfig + `
resource "vcd_nsxt_network_dhcp" "pools" {
  org  = "{{.Org}}"
  vdc  = "{{.NsxtVdc}}"

  org_network_id = vcd_network_routed_v2.net1.id
  
  pool {
    start_address = "7.1.1.100"
    end_address   = "7.1.1.110"
  }

  pool {
    start_address = "7.1.1.130"
    end_address   = "7.1.1.140"
  }

  dns_servers = ["1.1.1.1", "1.0.0.1"]
}

resource "vcd_nsxt_network_dhcp_binding" "binding1" {
  org  = "{{.Org}}"

  # org_network_id = vcd_network_routed_v2.net1.id
  org_network_id = vcd_nsxt_network_dhcp.pools.id
  
  name         = "{{.TestName}}-dhcp-binding-1"
  binding_type = "IPV4"
  ip_address   = "7.1.1.189"
  lease_time   = 60
  mac_address  = "00:11:22:33:44:55"
}

resource "vcd_nsxt_network_dhcp_binding" "binding2" {
  org  = "{{.Org}}"

  # referencing vcd_nsxt_network_dhcp.pools.id instead of vcd_network_routed_v2.net1.id because
  # DHCP service must be enabled on the network before DHCP bindings can be created
  org_network_id = vcd_nsxt_network_dhcp.pools.id
  
  name         = "{{.TestName}}-dhcp-binding-2"
  description  = "DHCP binding description"
  binding_type = "IPV4"
  ip_address   = "7.1.1.190"
  lease_time   = 3600
  mac_address  = "00:11:22:33:44:66"
  dns_servers  = ["7.1.1.242", "7.1.1.243"]

  dhcp_v4_config {
	gateway_ip_address = "7.1.1.233"
	hostname           = "non-existent"
  }
}
`

const testAccRoutedNetDhcpStep4 = testAccRoutedNetDhcpConfig + `
resource "vcd_nsxt_network_dhcp" "pools" {
  org  = "{{.Org}}"
  vdc  = "{{.NsxtVdc}}"

  org_network_id = vcd_network_routed_v2.net1.id
  
  pool {
    start_address = "7.1.1.100"
    end_address   = "7.1.1.110"
  }

  pool {
    start_address = "7.1.1.130"
    end_address   = "7.1.1.140"
  }

  dns_servers = ["1.1.1.1", "1.0.0.1"]
}

resource "vcd_nsxt_network_dhcp_binding" "binding1" {
  org  = "{{.Org}}"

  org_network_id = vcd_nsxt_network_dhcp.pools.id
  
  name         = "{{.TestName}}-dhcp-binding-1"
  binding_type = "IPV4"
  ip_address   = "7.1.1.167"
  lease_time   = 60
  mac_address  = "00:11:22:33:33:55"
}

resource "vcd_nsxt_network_dhcp_binding" "binding2" {
  org  = "{{.Org}}"

  # referencing vcd_nsxt_network_dhcp.pools.id instead of vcd_network_routed_v2.net1.id because
  # DHCP service must be enabled on the network before DHCP bindings can be created
  org_network_id = vcd_nsxt_network_dhcp.pools.id
  
  name         = "{{.TestName}}-dhcp-binding-2"
  binding_type = "IPV4"
  lease_time   = 3600
  ip_address   = "7.1.1.190"
  mac_address  = "00:11:22:33:44:66"
}
`

// TestAccVcdOpenApiDhcpNsxtIsolated checks that DHCP works in NSX-T Isolated networks.
// It requires a VDC with assigned Edge Cluster to work therefore it creates its own VDC
func TestAccVcdOpenApiDhcpNsxtIsolated(t *testing.T) {
	preTestChecks(t)

	skipIfNotSysAdmin(t) // creates its own VDC

	// Requires VCD 10.3.1+
	vcdClient := createTemporaryVCDConnection(true)
	if vcdClient == nil || vcdClient.Client.APIVCDMaxVersionIs("< 36.1") {
		t.Skipf("NSX-T Isolated network DHCP requires VCD 10.3.1+ (API v36.1+)")
	}

	// String map to fill the template
	var params = StringMap{
		"Org":                       testConfig.VCD.Org,
		"NetworkName":               t.Name(),
		"VdcName":                   t.Name(),
		"ProviderVdc":               testConfig.VCD.NsxtProviderVdc.Name,
		"NetworkPool":               testConfig.VCD.NsxtProviderVdc.NetworkPool,
		"ProviderVdcStorageProfile": testConfig.VCD.NsxtProviderVdc.StorageProfile,
		"EdgeCluster":               testConfig.Nsxt.NsxtEdgeCluster,
		"TestName":                  t.Name(),
		"Binding1Name":              t.Name() + "-dhcp-binding-1",

		"Tags": "network nsxt",
	}
	testParamsNotEmpty(t, params)

	configText1 := templateFill(testAccRoutedNetDhcpIsolatedStep1, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)

	params["FuncName"] = t.Name() + "-step2"
	configText2 := templateFill(testAccRoutedNetDhcpIsolatedStep2, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText2)

	params["FuncName"] = t.Name() + "-step3"
	configText3 := templateFill(testAccRoutedNetDhcpIsolatedStep3, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 3: %s", configText3)

	params["FuncName"] = t.Name() + "-step4DS"
	configText4DS := templateFill(testAccRoutedNetDhcpIsolatedStep4DS, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 4: %s", configText4DS)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	// This case is specific for VCD 10.3.1 onwards since dns servers are not present in previous versions
	// var configText2 string
	// if vcdClient != nil && vcdClient.Client.APIVCDMaxVersionIs(">= 36.1") {
	// 	params["SkipTest"] = "# skip-binary-test: VCD 10.3.1 onwards dns servers are not present in previous versions"
	// }
	// params["FuncName"] = t.Name() + "-step2"
	// configText2 = templateFill(testAccRoutedNetDhcpStep3, params)
	// debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText2)

	cacheDhcpBinding1dId := &testCachedFieldValue{}
	cacheDhcpBinding2dId := &testCachedFieldValue{}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckOpenApiVcdNetworkDestroy(testConfig.Nsxt.Vdc, "nsxt-routed-dhcp"),
		Steps: []resource.TestStep{
			{
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_network_dhcp.pools", "id", regexp.MustCompile(`^urn:vcloud:network:.*$`)),
					resource.TestCheckResourceAttr("vcd_nsxt_network_dhcp.pools", "lease_time", "86400"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_dhcp.pools", "mode", "NETWORK"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_dhcp.pools", "listener_ip_address", "7.1.1.254"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_dhcp.pools", "pool.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_network_dhcp.pools", "pool.*", map[string]string{
						"start_address": "7.1.1.100",
						"end_address":   "7.1.1.110",
					}),

					// DHCP binding checks
					cacheDhcpBinding1dId.cacheTestResourceFieldValue("vcd_nsxt_network_dhcp_binding.binding1", "id"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_network_dhcp_binding.binding1", "id"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_dhcp_binding.binding1", "name", t.Name()+"-dhcp-binding-1"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_dhcp_binding.binding1", "description", ""),
					resource.TestCheckResourceAttr("vcd_nsxt_network_dhcp_binding.binding1", "lease_time", "60"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_dhcp_binding.binding1", "binding_type", "IPV4"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_dhcp_binding.binding1", "dns_servers.#", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_dhcp_binding.binding1", "dhcp_v4_config.#", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_dhcp_binding.binding1", "ip_address", "7.1.1.189"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_dhcp_binding.binding1", "mac_address", "00:11:22:33:44:55"),

					cacheDhcpBinding2dId.cacheTestResourceFieldValue("vcd_nsxt_network_dhcp_binding.binding2", "id"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_network_dhcp_binding.binding2", "id"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_dhcp_binding.binding2", "name", t.Name()+"-dhcp-binding-2"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_dhcp_binding.binding2", "description", "DHCP binding description"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_dhcp_binding.binding2", "binding_type", "IPV4"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_dhcp_binding.binding2", "lease_time", "3600"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_dhcp_binding.binding2", "dns_servers.#", "2"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_dhcp_binding.binding2", "dns_servers.0", "7.1.1.242"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_dhcp_binding.binding2", "dns_servers.1", "7.1.1.243"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_dhcp_binding.binding2", "dhcp_v4_config.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_network_dhcp_binding.binding2", "dhcp_v4_config.*", map[string]string{
						"gateway_ip_address": "7.1.1.233",
						"hostname":           "non-existent",
					}),
					resource.TestCheckResourceAttr("vcd_nsxt_network_dhcp_binding.binding2", "ip_address", "7.1.1.190"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_dhcp_binding.binding2", "mac_address", "00:11:22:33:44:66"),
				),
			},
			{
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_network_dhcp.pools", "id", regexp.MustCompile(`^urn:vcloud:network:.*$`)),
					resource.TestCheckResourceAttr("vcd_nsxt_network_dhcp.pools", "lease_time", "60"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_dhcp.pools", "mode", "NETWORK"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_dhcp.pools", "pool.#", "2"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_dhcp.pools", "listener_ip_address", "7.1.1.254"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_network_dhcp.pools", "pool.*", map[string]string{
						"start_address": "7.1.1.100",
						"end_address":   "7.1.1.110",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_network_dhcp.pools", "pool.*", map[string]string{
						"start_address": "7.1.1.120",
						"end_address":   "7.1.1.140",
					}),

					// DHCP binding checks
					cacheDhcpBinding1dId.cacheTestResourceFieldValue("vcd_nsxt_network_dhcp_binding.binding1", "id"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_dhcp_binding.binding1", "name", t.Name()+"-dhcp-binding-1"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_dhcp_binding.binding1", "description", ""),
					resource.TestCheckResourceAttr("vcd_nsxt_network_dhcp_binding.binding1", "lease_time", "60"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_dhcp_binding.binding1", "binding_type", "IPV4"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_dhcp_binding.binding1", "dns_servers.#", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_dhcp_binding.binding1", "dhcp_v4_config.#", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_dhcp_binding.binding1", "ip_address", "7.1.1.167"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_dhcp_binding.binding1", "mac_address", "00:11:22:33:33:55"),

					cacheDhcpBinding2dId.cacheTestResourceFieldValue("vcd_nsxt_network_dhcp_binding.binding2", "id"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_network_dhcp_binding.binding2", "id"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_dhcp_binding.binding2", "name", t.Name()+"-dhcp-binding-2"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_dhcp_binding.binding2", "description", ""),
					resource.TestCheckResourceAttr("vcd_nsxt_network_dhcp_binding.binding2", "binding_type", "IPV4"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_dhcp_binding.binding2", "lease_time", "3600"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_dhcp_binding.binding2", "dns_servers.#", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_dhcp_binding.binding2", "dhcp_v4_config.#", "0"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_dhcp_binding.binding2", "ip_address", "7.1.1.190"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_dhcp_binding.binding2", "mac_address", "00:11:22:33:44:66"),
				),
			},
			{
				Config: configText3,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_network_dhcp.pools", "id", regexp.MustCompile(`^urn:vcloud:network:.*$`)),
					resource.TestCheckResourceAttr("vcd_nsxt_network_dhcp.pools", "lease_time", "60"),
					resource.TestCheckResourceAttr("vcd_nsxt_network_dhcp.pools", "pool.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_network_dhcp.pools", "pool.*", map[string]string{
						"start_address": "7.1.1.100",
						"end_address":   "7.1.1.110",
					}),
				),
			},
			{
				Config: configText4DS,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_network_dhcp.pools", "id", regexp.MustCompile(`^urn:vcloud:network:.*$`)),
					resourceFieldsEqual("data.vcd_nsxt_network_dhcp.pools", "vcd_nsxt_network_dhcp.pools", nil),

					resourceFieldsEqual("data.vcd_nsxt_network_dhcp_binding.binding1", "vcd_nsxt_network_dhcp_binding.binding1", nil),
					resourceFieldsEqual("data.vcd_nsxt_network_dhcp_binding.binding2", "vcd_nsxt_network_dhcp_binding.binding2", nil),
				),
			},
			{
				ResourceName:            "vcd_nsxt_network_dhcp.pools",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdFunc:       importStateIdOrgNsxtVdcGroupObject(params["VdcName"].(string), params["NetworkName"].(string)),
				ImportStateVerifyIgnore: []string{"vdc"},
			},
			{
				ResourceName:      "vcd_nsxt_network_dhcp_binding.binding1",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importCustomObject([]string{params["Org"].(string), params["VdcName"].(string), params["NetworkName"].(string), params["Binding1Name"].(string)}),
			},
		},
	})
	postTestChecks(t)
}

const testAccRoutedNetDhcpIsolated = `
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

resource "vcd_network_isolated_v2" "net1" {
  org      = "{{.Org}}"
  owner_id = vcd_org_vdc.with-edge-cluster.id
  name     = "{{.NetworkName}}"

  gateway       = "7.1.1.1"
  prefix_length = 24

  static_ip_pool {
    start_address = "7.1.1.10"
    end_address   = "7.1.1.20"
  }
}
`

const testAccRoutedNetDhcpIsolatedStep1 = testAccRoutedNetDhcpIsolated + `
resource "vcd_nsxt_network_dhcp" "pools" {
  org  = "{{.Org}}"
  vdc  = vcd_org_vdc.with-edge-cluster.name

  org_network_id      = vcd_network_isolated_v2.net1.id
  mode                = "NETWORK"
  listener_ip_address = "7.1.1.254"
  
  pool {
    start_address = "7.1.1.100"
    end_address   = "7.1.1.110"
  }
}

resource "vcd_nsxt_network_dhcp_binding" "binding1" {
  org  = "{{.Org}}"

  # org_network_id = vcd_network_routed_v2.net1.id
  org_network_id = vcd_nsxt_network_dhcp.pools.id
  
  name         = "{{.Binding1Name}}"
  binding_type = "IPV4"
  ip_address   = "7.1.1.189"
  lease_time   = 60
  mac_address  = "00:11:22:33:44:55"
}

resource "vcd_nsxt_network_dhcp_binding" "binding2" {
  org  = "{{.Org}}"

  # referencing vcd_nsxt_network_dhcp.pools.id instead of vcd_network_routed_v2.net1.id because
  # DHCP service must be enabled on the network before DHCP bindings can be created
  org_network_id = vcd_nsxt_network_dhcp.pools.id
  
  name         = "{{.TestName}}-dhcp-binding-2"
  description  = "DHCP binding description"
  binding_type = "IPV4"
  ip_address   = "7.1.1.190"
  lease_time   = 3600
  mac_address  = "00:11:22:33:44:66"
  dns_servers  = ["7.1.1.242", "7.1.1.243"]

  dhcp_v4_config {
	gateway_ip_address = "7.1.1.233"
	hostname           = "non-existent"
  }
}
`

const testAccRoutedNetDhcpIsolatedStep2 = testAccRoutedNetDhcpIsolated + `
resource "vcd_nsxt_network_dhcp" "pools" {
  org  = "{{.Org}}"
  vdc  = vcd_org_vdc.with-edge-cluster.name

  org_network_id      = vcd_network_isolated_v2.net1.id
  mode                = "NETWORK"
  listener_ip_address = "7.1.1.254"
  lease_time          = 60
  
  pool {
    start_address = "7.1.1.100"
    end_address   = "7.1.1.110"
  }

  pool {
    start_address = "7.1.1.120"
    end_address   = "7.1.1.140"
  }
}

resource "vcd_nsxt_network_dhcp_binding" "binding1" {
  org  = "{{.Org}}"

  org_network_id = vcd_nsxt_network_dhcp.pools.id
  
  name         = "{{.Binding1Name}}"
  binding_type = "IPV4"
  ip_address   = "7.1.1.167"
  lease_time   = 60
  mac_address  = "00:11:22:33:33:55"
}

resource "vcd_nsxt_network_dhcp_binding" "binding2" {
  org  = "{{.Org}}"

  # referencing vcd_nsxt_network_dhcp.pools.id instead of vcd_network_routed_v2.net1.id because
  # DHCP service must be enabled on the network before DHCP bindings can be created
  org_network_id = vcd_nsxt_network_dhcp.pools.id
  
  name         = "{{.TestName}}-dhcp-binding-2"
  binding_type = "IPV4"
  lease_time   = 3600
  ip_address   = "7.1.1.190"
  mac_address  = "00:11:22:33:44:66"
}
`

const testAccRoutedNetDhcpIsolatedStep3 = testAccRoutedNetDhcpIsolated + `
resource "vcd_nsxt_network_dhcp" "pools" {
  org  = "{{.Org}}"
  vdc  = vcd_org_vdc.with-edge-cluster.name

  org_network_id      = vcd_network_isolated_v2.net1.id
  mode                = "NETWORK"
  listener_ip_address = "7.1.1.254"
  
  pool {
    start_address = "7.1.1.100"
    end_address   = "7.1.1.110"
  }
}

resource "vcd_nsxt_network_dhcp_binding" "binding1" {
  org  = "{{.Org}}"

  org_network_id = vcd_nsxt_network_dhcp.pools.id
  
  name         = "{{.Binding1Name}}"
  binding_type = "IPV4"
  ip_address   = "7.1.1.167"
  lease_time   = 60
  mac_address  = "00:11:22:33:33:55"
}

resource "vcd_nsxt_network_dhcp_binding" "binding2" {
  org  = "{{.Org}}"

  # referencing vcd_nsxt_network_dhcp.pools.id instead of vcd_network_routed_v2.net1.id because
  # DHCP service must be enabled on the network before DHCP bindings can be created
  org_network_id = vcd_nsxt_network_dhcp.pools.id
  
  name         = "{{.TestName}}-dhcp-binding-2"
  binding_type = "IPV4"
  lease_time   = 3600
  ip_address   = "7.1.1.190"
  mac_address  = "00:11:22:33:44:66"
}
`

const testAccRoutedNetDhcpIsolatedStep4DS = testAccRoutedNetDhcpIsolatedStep3 + `
# skip-binary-test: cannot test resource and data source in binary test mode
data "vcd_nsxt_network_dhcp" "pools" {
  org = vcd_nsxt_network_dhcp.pools.org
  vdc = vcd_nsxt_network_dhcp.pools.vdc

  org_network_id = vcd_nsxt_network_dhcp.pools.org_network_id
}

data "vcd_nsxt_network_dhcp_binding" "binding1" {
  org            = "{{.Org}}"
  org_network_id = vcd_nsxt_network_dhcp.pools.id
  name           = "{{.TestName}}-dhcp-binding-1"
}

data "vcd_nsxt_network_dhcp_binding" "binding2" {
  org            = "{{.Org}}"
  org_network_id = vcd_nsxt_network_dhcp.pools.id
  name           = "{{.TestName}}-dhcp-binding-2"
}

`

// TestAccVcdOpenApiDhcpNsxtRoutedRelay tests RELAY mode for DHCP.
// TODO we do not yet have a DHCP Forwarding resource (configured in Edge Gateway) therefore this
// test was run with DHCP forwarding manually configured. Improve and and uncomment this test when
// DHCP Forwarding resource is created and can be used here
// func TestAccVcdOpenApiDhcpNsxtRoutedRelay(t *testing.T) {
// 	preTestChecks(t)

// 	// Requires VCD 10.3.1+
// 	vcdClient := createTemporaryVCDConnection(true)
// 	if vcdClient == nil && vcdClient.Client.APIVCDMaxVersionIs("< 36.1") {
// 		t.Skipf("NSX-T Isolated network DHCP requires VCD 10.3.1+ (API v36.1+)")
// 	}

// 	// String map to fill the template
// 	var params = StringMap{
// 		"Org":         testConfig.VCD.Org,
// 		"NsxtVdc":     testConfig.Nsxt.Vdc,
// 		"EdgeGw":      testConfig.Nsxt.EdgeGateway,
// 		"NetworkName": t.Name(),
// 		"Tags":        "network nsxt",
// 	}
// 	testParamsNotEmpty(t, params)

// 	configText1 := templateFill(testAccRoutedNetRelayDhcpStep1, params)
// 	debugPrintf("#[DEBUG] CONFIGURATION for step 0: %s", configText1)

// 	if vcdShortTest {
// 		t.Skip(acceptanceTestsSkipped)
// 		return
// 	}

// 	resource.Test(t, resource.TestCase{
// 		ProviderFactories: testAccProviders,
// 		CheckDestroy:      testAccCheckOpenApiVcdNetworkDestroy(testConfig.Nsxt.Vdc, "nsxt-routed-dhcp"),
// 		Steps: []resource.TestStep{
// 			{
// 				Config: configText1,
// 				Check: resource.ComposeAggregateTestCheckFunc(
// 					resource.TestMatchResourceAttr("vcd_nsxt_network_dhcp.pools", "id", regexp.MustCompile(`^urn:vcloud:network:.*$`)),
// 					resource.TestCheckResourceAttr("vcd_nsxt_network_dhcp.pools", "mode", "RELAY"),
// 					resource.TestCheckResourceAttr("vcd_nsxt_network_dhcp.pools", "pool.#", "0"),
// 				),
// 			},
// 			{
// 				ResourceName:            "vcd_nsxt_network_dhcp.pools",
// 				ImportState:             true,
// 				ImportStateVerify:       true,
// 				ImportStateIdFunc:       importStateIdOrgNsxtVdcObject("nsxt-routed-dhcp"),
// 				ImportStateVerifyIgnore: []string{"vdc"},
// 			},
// 		},
// 	})
// 	postTestChecks(t)
// }

// const testAccRoutedNetRelayDhcpStep1 = testAccRoutedNetDhcpConfig + `
// resource "vcd_nsxt_network_dhcp" "pools" {
//   org  = "{{.Org}}"
//   vdc  = "{{.NsxtVdc}}"

//   org_network_id = vcd_network_routed_v2.net1.id
//   mode           = "RELAY"
// }
// `
