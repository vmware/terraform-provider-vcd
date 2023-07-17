//go:build functional || network || extnetwork || nsxt || ALL

package vcd

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/vmware/go-vcloud-director/v2/govcd"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcdExternalNetworkV2NsxtVrf(t *testing.T) {
	preTestChecks(t)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	testAccVcdExternalNetworkV2Nsxt(t, testConfig.Nsxt.Tier0routerVrf)
	postTestChecks(t)
}

func TestAccVcdExternalNetworkV2NsxtT0Router(t *testing.T) {
	preTestChecks(t)
	testAccVcdExternalNetworkV2Nsxt(t, testConfig.Nsxt.Tier0router)
	postTestChecks(t)
}

func testAccVcdExternalNetworkV2Nsxt(t *testing.T, nsxtTier0Router string) {

	skipIfNotSysAdmin(t)

	startAddress := "192.168.30.51"
	endAddress := "192.168.30.62"
	description := "Test External Network"
	var params = StringMap{
		"NsxtManager":         testConfig.Nsxt.Manager,
		"NsxtTier0Router":     nsxtTier0Router,
		"ExternalNetworkName": t.Name(),
		"StartAddress":        startAddress,
		"EndAddress":          endAddress,
		"Description":         description,
		"Gateway":             "192.168.30.49",
		"Netmask":             "24",
		"Tags":                "network extnetwork nsxt",
	}
	testParamsNotEmpty(t, params)

	params["FuncName"] = t.Name() + "step"
	configText := templateFill(testAccCheckVcdExternalNetworkV2NsxtStep1, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

	params["FuncName"] = t.Name() + "step2"
	configText2 := templateFill(testAccCheckVcdExternalNetworkV2NsxtStep2, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText2)

	params["FuncName"] = t.Name() + "step3"
	configText3 := templateFill(testAccCheckVcdExternalNetworkV2NsxtStep3, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText3)

	params["FuncName"] = t.Name() + "step4"
	configText4 := templateFill(testAccCheckVcdExternalNetworkV2NsxtStep4Ipv6, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText4)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	resourceName := "vcd_external_network_v2.ext-net-nsxt"
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckExternalNetworkDestroyV2(t.Name()),
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", t.Name()),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					resource.TestCheckResourceAttr(resourceName, "vsphere_network.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "nsxt_network.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "ip_scope.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ip_scope.*", map[string]string{
						"dns1":          "",
						"dns2":          "",
						"dns_suffix":    "",
						"enabled":       "false",
						"gateway":       "192.168.30.49",
						"prefix_length": "24",
					}),

					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ip_scope.*.static_ip_pool.*", map[string]string{
						"start_address": "192.168.30.51",
						"end_address":   "192.168.30.62",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ip_scope.*", map[string]string{
						"dns1":          "",
						"dns2":          "",
						"dns_suffix":    "",
						"enabled":       "true",
						"gateway":       "14.14.14.1",
						"prefix_length": "24",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ip_scope.*.static_ip_pool.*", map[string]string{
						"start_address": "14.14.14.20",
						"end_address":   "14.14.14.25",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ip_scope.*.static_ip_pool.*", map[string]string{
						"start_address": "14.14.14.10",
						"end_address":   "14.14.14.15",
					}),
					resource.TestCheckResourceAttr(resourceName, "nsxt_network.#", "1"),
					testCheckMatchOutput("nsxt-manager", regexp.MustCompile("^urn:vcloud:nsxtmanager:.*")),
					testCheckOutputNonEmpty("nsxt-tier0-router"), // Match any non empty string
				),
			},
			{
				Config: configText2,
				Taint:  []string{"vcd_external_network_v2.ext-net-nsxt"},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", t.Name()),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					resource.TestCheckResourceAttr(resourceName, "vsphere_network.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "nsxt_network.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "ip_scope.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ip_scope.*", map[string]string{
						"dns1": "8.8.8.8",
						"dns2": "8.8.4.4",
						// dns_suffix has a bug in VCD (<= 10.4.1) which does not return it after
						// setting
						// "dns_suffix":    "host.test",
						"enabled":       "false",
						"gateway":       "192.168.30.49",
						"prefix_length": "24",
					}),

					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ip_scope.*.static_ip_pool.*", map[string]string{
						"start_address": "192.168.30.51",
						"end_address":   "192.168.30.62",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ip_scope.*", map[string]string{
						"dns1":          "",
						"dns2":          "",
						"dns_suffix":    "",
						"enabled":       "true",
						"gateway":       "14.14.14.1",
						"prefix_length": "24",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ip_scope.*.static_ip_pool.*", map[string]string{
						"start_address": "14.14.14.20",
						"end_address":   "14.14.14.25",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ip_scope.*.static_ip_pool.*", map[string]string{
						"start_address": "14.14.14.10",
						"end_address":   "14.14.14.15",
					}),
					resource.TestCheckResourceAttr(resourceName, "nsxt_network.#", "1"),
					testCheckMatchOutput("nsxt-manager", regexp.MustCompile("^urn:vcloud:nsxtmanager:.*")),
					testCheckOutputNonEmpty("nsxt-tier0-router"), // Match any non empty string
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdTopHierarchy(t.Name()),
			},
			{
				Config: configText3,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", t.Name()),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					resource.TestCheckResourceAttr(resourceName, "vsphere_network.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "nsxt_network.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "ip_scope.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ip_scope.*", map[string]string{
						"dns1":          "",
						"dns2":          "",
						"dns_suffix":    "",
						"enabled":       "true",
						"gateway":       "192.168.30.49",
						"prefix_length": "24",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ip_scope.*.static_ip_pool.*", map[string]string{
						"start_address": "192.168.30.51",
						"end_address":   "192.168.30.62",
					}),
					resource.TestCheckResourceAttr(resourceName, "nsxt_network.#", "1"),
					testCheckMatchOutput("nsxt-manager", regexp.MustCompile("^urn:vcloud:nsxtmanager:.*")),
					testCheckOutputNonEmpty("nsxt-tier0-router"), // Match any non empty string
				),
			},
			{
				Config: configText4,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", t.Name()),
					resource.TestCheckResourceAttr(resourceName, "description", "IPv6"),
					resource.TestCheckResourceAttr(resourceName, "vsphere_network.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "nsxt_network.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "ip_scope.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ip_scope.*", map[string]string{
						"dns1":          "",
						"dns2":          "",
						"dns_suffix":    "",
						"enabled":       "true",
						"gateway":       "2002:0:0:1234:abcd:ffff:c0a8:101",
						"prefix_length": "124",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ip_scope.*.static_ip_pool.*", map[string]string{
						"start_address": "2002:0:0:1234:abcd:ffff:c0a8:103",
						"end_address":   "2002:0:0:1234:abcd:ffff:c0a8:104",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ip_scope.*.static_ip_pool.*", map[string]string{
						"start_address": "2002:0:0:1234:abcd:ffff:c0a8:107",
						"end_address":   "2002:0:0:1234:abcd:ffff:c0a8:109",
					}),
					resource.TestCheckResourceAttr(resourceName, "nsxt_network.#", "1"),
					testCheckMatchOutput("nsxt-manager", regexp.MustCompile("^urn:vcloud:nsxtmanager:.*")),
					testCheckOutputNonEmpty("nsxt-tier0-router"), // Match any non empty string
				),
			},
		},
	})
}

const testAccCheckVcdExternalNetworkV2NsxtDS = `
data "vcd_nsxt_manager" "main" {
  name = "{{.NsxtManager}}"
}

data "vcd_nsxt_tier0_router" "router" {
  name            = "{{.NsxtTier0Router}}"
  nsxt_manager_id = data.vcd_nsxt_manager.main.id
}
`

const testAccCheckVcdExternalNetworkV2NsxtStep1 = testAccCheckVcdExternalNetworkV2NsxtDS + `
resource "vcd_external_network_v2" "ext-net-nsxt" {
  name        = "{{.ExternalNetworkName}}"
  description = "{{.Description}}"

  nsxt_network {
    nsxt_manager_id      = data.vcd_nsxt_manager.main.id
    nsxt_tier0_router_id = data.vcd_nsxt_tier0_router.router.id
  }

  ip_scope {
    enabled       = false
    gateway       = "{{.Gateway}}"
    prefix_length = "{{.Netmask}}"

    static_ip_pool {
      start_address = "{{.StartAddress}}"
      end_address   = "{{.EndAddress}}"
    }
  }

  ip_scope {
    gateway       = "14.14.14.1"
    prefix_length = "24"

    static_ip_pool {
      start_address = "14.14.14.10"
      end_address   = "14.14.14.15"
    }
    
    static_ip_pool {
      start_address = "14.14.14.20"
      end_address   = "14.14.14.25"
    }
  }
}

output "nsxt-manager" {
  value = tolist(vcd_external_network_v2.ext-net-nsxt.nsxt_network)[0].nsxt_manager_id
}

output "nsxt-tier0-router" {
  value = tolist(vcd_external_network_v2.ext-net-nsxt.nsxt_network)[0].nsxt_tier0_router_id
}
`

const testAccCheckVcdExternalNetworkV2NsxtStep2 = testAccCheckVcdExternalNetworkV2NsxtDS + `
resource "vcd_external_network_v2" "ext-net-nsxt" {
  name        = "{{.ExternalNetworkName}}"
  description = "{{.Description}}"

  nsxt_network {
    nsxt_manager_id      = data.vcd_nsxt_manager.main.id
    nsxt_tier0_router_id = data.vcd_nsxt_tier0_router.router.id
  }

  ip_scope {
    enabled       = false
    gateway       = "{{.Gateway}}"
    prefix_length = "{{.Netmask}}"

	dns1       = "8.8.8.8"
	dns2       = "8.8.4.4"
	# VCD bug does not return the value after it is set
	# dns_suffix = "host.test"

    static_ip_pool {
      start_address = "{{.StartAddress}}"
      end_address   = "{{.EndAddress}}"
    }
  }

  ip_scope {
    gateway       = "14.14.14.1"
    prefix_length = "24"

    static_ip_pool {
      start_address = "14.14.14.10"
      end_address   = "14.14.14.15"
    }
    
    static_ip_pool {
      start_address = "14.14.14.20"
      end_address   = "14.14.14.25"
    }
  }
}

output "nsxt-manager" {
  value = tolist(vcd_external_network_v2.ext-net-nsxt.nsxt_network)[0].nsxt_manager_id
}

output "nsxt-tier0-router" {
  value = tolist(vcd_external_network_v2.ext-net-nsxt.nsxt_network)[0].nsxt_tier0_router_id
}
`

const testAccCheckVcdExternalNetworkV2NsxtStep3 = testAccCheckVcdExternalNetworkV2NsxtDS + `
# skip-binary-test: only for updates
resource "vcd_external_network_v2" "ext-net-nsxt" {
  name        = "{{.ExternalNetworkName}}"
  description = "{{.Description}}"

  nsxt_network {
    nsxt_manager_id      = data.vcd_nsxt_manager.main.id
    nsxt_tier0_router_id = data.vcd_nsxt_tier0_router.router.id
  }

  ip_scope {
    enabled       = true
    gateway       = "{{.Gateway}}"
    prefix_length = "{{.Netmask}}"

    static_ip_pool {
      start_address = "{{.StartAddress}}"
      end_address   = "{{.EndAddress}}"
    }
  }
}

output "nsxt-manager" {
  value = tolist(vcd_external_network_v2.ext-net-nsxt.nsxt_network)[0].nsxt_manager_id
}

output "nsxt-tier0-router" {
  value = tolist(vcd_external_network_v2.ext-net-nsxt.nsxt_network)[0].nsxt_tier0_router_id
}
`

const testAccCheckVcdExternalNetworkV2NsxtStep4Ipv6 = testAccCheckVcdExternalNetworkV2NsxtDS + `
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

output "nsxt-manager" {
  value = tolist(vcd_external_network_v2.ext-net-nsxt.nsxt_network)[0].nsxt_manager_id
}

output "nsxt-tier0-router" {
  value = tolist(vcd_external_network_v2.ext-net-nsxt.nsxt_network)[0].nsxt_tier0_router_id
}
`

func TestAccVcdExternalNetworkV2Nsxv(t *testing.T) {
	preTestChecks(t)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	skipIfNotSysAdmin(t)

	description := "Test External Network"
	var params = StringMap{
		"ExternalNetworkName": t.Name(),
		"Type":                testConfig.Networking.ExternalNetworkPortGroupType,
		"PortGroup":           testConfig.Networking.ExternalNetworkPortGroup,
		"Vcenter":             testConfig.Networking.Vcenter,
		"StartAddress":        "192.168.30.51",
		"EndAddress":          "192.168.30.62",
		"Description":         description,
		"Gateway":             "192.168.30.49",
		"Netmask":             "24",
		"Dns1":                "192.168.0.164",
		"Dns2":                "192.168.0.196",
		"Tags":                "network extnetwork nsxt",
	}
	testParamsNotEmpty(t, params)

	configText := templateFill(testAccCheckVcdExternalNetworkV2Nsxv, params)
	params["FuncName"] = t.Name() + "step1"
	configText1 := templateFill(testAccCheckVcdExternalNetworkV2NsxvUpdate, params)

	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText1)

	resourceName := "vcd_external_network_v2.ext-net-nsxv"
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckExternalNetworkDestroyV2(t.Name()),
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", t.Name()),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					resource.TestCheckResourceAttr(resourceName, "ip_scope.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ip_scope.*", map[string]string{
						"dns1":          "192.168.0.164",
						"dns2":          "192.168.0.196",
						"dns_suffix":    "company.biz",
						"enabled":       "true",
						"gateway":       "192.168.30.49",
						"prefix_length": "24",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ip_scope.0.static_ip_pool.*", map[string]string{
						"start_address": "192.168.30.51",
						"end_address":   "192.168.30.62",
					}),
					resource.TestCheckResourceAttr(resourceName, "ip_scope.0.static_ip_pool.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "nsxt_network.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "vsphere_network.#", "1"),
					testCheckOutputNonEmpty("vcenter-id"),   // Match any non empty string
					testCheckOutputNonEmpty("portgroup-id"), // Match any non empty string
				),
			},
			{
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", t.Name()),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					resource.TestCheckResourceAttr(resourceName, "ip_scope.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ip_scope.*", map[string]string{
						"dns1":          "192.168.0.164",
						"dns2":          "192.168.0.196",
						"dns_suffix":    "company.biz",
						"enabled":       "false",
						"gateway":       "192.168.30.49",
						"prefix_length": "24",
					}),
					resource.TestCheckResourceAttr(resourceName, "ip_scope.0.static_ip_pool.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ip_scope.*.static_ip_pool.*", map[string]string{
						"start_address": "192.168.30.51",
						"end_address":   "192.168.30.62",
					}),

					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ip_scope.*", map[string]string{
						"dns1":          "8.8.8.8",
						"dns2":          "8.8.4.4",
						"dns_suffix":    "asd.biz",
						"enabled":       "true",
						"gateway":       "88.88.88.1",
						"prefix_length": "24",
					}),
					resource.TestCheckResourceAttr(resourceName, "ip_scope.0.static_ip_pool.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ip_scope.*.static_ip_pool.*", map[string]string{
						"start_address": "88.88.88.10",
						"end_address":   "88.88.88.100",
					}),
					resource.TestCheckResourceAttr(resourceName, "nsxt_network.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "vsphere_network.#", "1"),
					testCheckMatchOutput("vcenter-id", regexp.MustCompile("^urn:vcloud:vimserver:.*")),
					testCheckOutputNonEmpty("portgroup-id"), // Match any non empty string because IDs may differ
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdTopHierarchy(t.Name()),
			},
		},
	})
	postTestChecks(t)
}

const testAccCheckVcdExternalNetworkV2NsxvDs = `
data "vcd_vcenter" "vc" {
  name = "{{.Vcenter}}"
}

data "vcd_portgroup" "sw" {
  name = "{{.PortGroup}}"
  type = "{{.Type}}"
}

`

const testAccCheckVcdExternalNetworkV2Nsxv = testAccCheckVcdExternalNetworkV2NsxvDs + `
resource "vcd_external_network_v2" "ext-net-nsxv" {
  name        = "{{.ExternalNetworkName}}"
  description = "{{.Description}}"

  vsphere_network {
    vcenter_id     = data.vcd_vcenter.vc.id
    portgroup_id   = data.vcd_portgroup.sw.id
  }

  ip_scope {
    gateway       = "{{.Gateway}}"
    prefix_length = "{{.Netmask}}"
    dns1          = "{{.Dns1}}"
    dns2          = "{{.Dns2}}"
    dns_suffix    = "company.biz"

    static_ip_pool {
      start_address = "{{.StartAddress}}"
      end_address   = "{{.EndAddress}}"
    }
  }
}

output "vcenter-id" {
  value = tolist(vcd_external_network_v2.ext-net-nsxv.vsphere_network)[0].vcenter_id
}

output "portgroup-id" {
  value = tolist(vcd_external_network_v2.ext-net-nsxv.vsphere_network)[0].portgroup_id
}
`

const testAccCheckVcdExternalNetworkV2NsxvUpdate = testAccCheckVcdExternalNetworkV2NsxvDs + `
# skip-binary-test: only for updates
resource "vcd_external_network_v2" "ext-net-nsxv" {
  name        = "{{.ExternalNetworkName}}"
  description = "{{.Description}}"

  vsphere_network {
    vcenter_id     = data.vcd_vcenter.vc.id
    portgroup_id   = data.vcd_portgroup.sw.id
  }

  ip_scope {
    enabled       = false
    gateway       = "{{.Gateway}}"
    prefix_length = "{{.Netmask}}"
    dns1          = "{{.Dns1}}"
    dns2          = "{{.Dns2}}"
    dns_suffix    = "company.biz"

    static_ip_pool {
      start_address = "{{.StartAddress}}"
      end_address   = "{{.EndAddress}}"
    }
  }

  ip_scope {
    gateway       = "88.88.88.1"
    prefix_length = "24"
    dns1          = "8.8.8.8"
    dns2          = "8.8.4.4"
    dns_suffix    = "asd.biz"

    static_ip_pool {
      start_address = "88.88.88.10"
      end_address   = "88.88.88.100"
    }
  }
}

output "vcenter-id" {
  value = tolist(vcd_external_network_v2.ext-net-nsxv.vsphere_network)[0].vcenter_id
}

output "portgroup-id" {
  value = tolist(vcd_external_network_v2.ext-net-nsxv.vsphere_network)[0].portgroup_id
}
`

func TestAccVcdExternalNetworkV2NsxtSegment(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	startAddress := "192.168.30.51"
	endAddress := "192.168.30.62"
	description := "Test External Network"
	var params = StringMap{
		"NsxtManager":         testConfig.Nsxt.Manager,
		"NsxtSegment":         testConfig.Nsxt.NsxtImportSegment,
		"ExternalNetworkName": t.Name(),
		"StartAddress":        startAddress,
		"EndAddress":          endAddress,
		"Description":         description,
		"Gateway":             "192.168.30.49",
		"Netmask":             "24",
		"Tags":                "network extnetwork nsxt",
	}
	testParamsNotEmpty(t, params)

	params["FuncName"] = t.Name()
	configText := templateFill(testAccCheckVcdExternalNetworkV2NsxtSegment, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

	params["FuncName"] = t.Name() + "step1"
	configText1 := templateFill(testAccCheckVcdExternalNetworkV2NsxtSegmentStep2, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText1)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	resourceName := "vcd_external_network_v2.ext-net-nsxt"
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckExternalNetworkDestroyV2(t.Name()),
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", t.Name()),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					resource.TestCheckResourceAttr(resourceName, "vsphere_network.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "nsxt_network.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "ip_scope.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ip_scope.*", map[string]string{
						"dns1":          "",
						"dns2":          "",
						"dns_suffix":    "",
						"enabled":       "false",
						"gateway":       "192.168.30.49",
						"prefix_length": "24",
					}),

					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ip_scope.*.static_ip_pool.*", map[string]string{
						"start_address": "192.168.30.51",
						"end_address":   "192.168.30.62",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ip_scope.*", map[string]string{
						"dns1":          "",
						"dns2":          "",
						"dns_suffix":    "",
						"enabled":       "true",
						"gateway":       "14.14.14.1",
						"prefix_length": "24",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ip_scope.*.static_ip_pool.*", map[string]string{
						"start_address": "14.14.14.20",
						"end_address":   "14.14.14.25",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ip_scope.*.static_ip_pool.*", map[string]string{
						"start_address": "14.14.14.10",
						"end_address":   "14.14.14.15",
					}),
					resource.TestCheckResourceAttr(resourceName, "nsxt_network.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdTopHierarchy(t.Name()),
			},
			{
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", t.Name()),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					resource.TestCheckResourceAttr(resourceName, "vsphere_network.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "nsxt_network.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "ip_scope.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ip_scope.*", map[string]string{
						"dns1":          "",
						"dns2":          "",
						"dns_suffix":    "",
						"enabled":       "true",
						"gateway":       "192.168.30.49",
						"prefix_length": "24",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ip_scope.*.static_ip_pool.*", map[string]string{
						"start_address": "192.168.30.51",
						"end_address":   "192.168.30.62",
					}),
					resource.TestCheckResourceAttr(resourceName, "nsxt_network.#", "1"),
				),
			},
		},
	})

	postTestChecks(t)
}

const testAccCheckVcdExternalNetworkV2NsxtSegment = `
data "vcd_nsxt_manager" "main" {
  name = "{{.NsxtManager}}"
}

resource "vcd_external_network_v2" "ext-net-nsxt" {
  name        = "{{.ExternalNetworkName}}"
  description = "{{.Description}}"

  nsxt_network {
    nsxt_manager_id   = data.vcd_nsxt_manager.main.id
    nsxt_segment_name = "{{.NsxtSegment}}"
  }

  ip_scope {
    enabled       = false
    gateway       = "{{.Gateway}}"
    prefix_length = "{{.Netmask}}"

    static_ip_pool {
      start_address = "{{.StartAddress}}"
      end_address   = "{{.EndAddress}}"
    }
  }

  ip_scope {
    gateway       = "14.14.14.1"
    prefix_length = "24"

    static_ip_pool {
      start_address = "14.14.14.10"
      end_address   = "14.14.14.15"
    }
    
    static_ip_pool {
      start_address = "14.14.14.20"
      end_address   = "14.14.14.25"
    }
  }
}
`

const testAccCheckVcdExternalNetworkV2NsxtSegmentStep2 = `
# skip-binary-test: only for updates
data "vcd_nsxt_manager" "main" {
  name = "{{.NsxtManager}}"
}

resource "vcd_external_network_v2" "ext-net-nsxt" {
  name        = "{{.ExternalNetworkName}}"
  description = "{{.Description}}"

  nsxt_network {
    nsxt_manager_id   = data.vcd_nsxt_manager.main.id
    nsxt_segment_name = "{{.NsxtSegment}}"
  }

  ip_scope {
    enabled       = true
    gateway       = "{{.Gateway}}"
    prefix_length = "{{.Netmask}}"

    static_ip_pool {
      start_address = "{{.StartAddress}}"
      end_address   = "{{.EndAddress}}"
    }
  }
}
`

func TestAccVcdExternalNetworkV2NsxtConfigError(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	startAddress := "192.168.30.51"
	endAddress := "192.168.30.62"
	description := "Test External Network"
	var params = StringMap{
		"NsxtManager":         testConfig.Nsxt.Manager,
		"NsxtSegment":         testConfig.Nsxt.NsxtImportSegment,
		"NsxtTier0Router":     testConfig.Nsxt.Tier0router,
		"ExternalNetworkName": t.Name(),
		"StartAddress":        startAddress,
		"EndAddress":          endAddress,
		"Description":         description,
		"Gateway":             "192.168.30.49",
		"Netmask":             "24",
		"Tags":                "network extnetwork nsxt",
	}
	testParamsNotEmpty(t, params)

	params["FuncName"] = t.Name()
	configText := templateFill(testAccCheckVcdExternalNetworkV2NsxtConfigError, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      configText,
				ExpectError: regexp.MustCompile(`Invalid combination of arguments`),
			},
		},
	})

	postTestChecks(t)
}

const testAccCheckVcdExternalNetworkV2NsxtConfigError = `
# skip-binary-test: fails on purpose
data "vcd_nsxt_manager" "main" {
  name = "{{.NsxtManager}}"
}

data "vcd_nsxt_tier0_router" "router" {
  name            = "{{.NsxtTier0Router}}"
  nsxt_manager_id = data.vcd_nsxt_manager.main.id
}

resource "vcd_external_network_v2" "ext-net-nsxt" {
  name        = "{{.ExternalNetworkName}}"
  description = "{{.Description}}"

  nsxt_network {
    nsxt_manager_id   = data.vcd_nsxt_manager.main.id
    nsxt_segment_name = "{{.NsxtSegment}}"
    nsxt_tier0_router_id = data.vcd_nsxt_tier0_router.router.id
  }

  ip_scope {
    enabled       = false
    gateway       = "{{.Gateway}}"
    prefix_length = "{{.Netmask}}"

    static_ip_pool {
      start_address = "{{.StartAddress}}"
      end_address   = "{{.EndAddress}}"
    }
  }

  ip_scope {
    gateway       = "14.14.14.1"
    prefix_length = "24"

    static_ip_pool {
      start_address = "14.14.14.10"
      end_address   = "14.14.14.15"
    }
    
    static_ip_pool {
      start_address = "14.14.14.20"
      end_address   = "14.14.14.25"
    }
  }
}
`

func testAccCheckExternalNetworkDestroyV2(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "vcd_external_network_v2" && rs.Primary.Attributes["name"] != name {
				continue
			}

			conn := testAccProvider.Meta().(*VCDClient)
			_, err := govcd.GetExternalNetworkV2ByName(conn.VCDClient, rs.Primary.ID)
			if err == nil {
				return fmt.Errorf("external network v2 %s still exists", rs.Primary.ID)
			}
		}

		return nil
	}
}

// TestAccVcdExternalNetworkV2NsxtSegmentIntegration attempts to test creation of NSX-T backed segment and also an NSX-T
// Org direct network resource (the only possible while implementing this feature)
func TestAccVcdExternalNetworkV2NsxtSegmentIntegration(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	startAddress := "192.168.30.51"
	endAddress := "192.168.30.62"
	description := "Test External Network"
	var params = StringMap{
		"NsxtManager":         testConfig.Nsxt.Manager,
		"NsxtSegment":         testConfig.Nsxt.NsxtImportSegment,
		"NsxtVdc":             testConfig.Nsxt.Vdc,
		"ExternalNetworkName": t.Name(),
		"StartAddress":        startAddress,
		"EndAddress":          endAddress,
		"Description":         description,
		"Gateway":             "192.168.30.49",
		"Netmask":             "24",
		"Tags":                "network extnetwork nsxt",
	}
	testParamsNotEmpty(t, params)

	configText := templateFill(testAccCheckVcdExternalNetworkV2NsxtSegmentIntegration, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)
	params["FuncName"] = t.Name() + "-step2"
	configText1 := templateFill(testAccCheckVcdExternalNetworkV2NsxtSegmentIntegrationDS, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText1)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	resourceName := "vcd_external_network_v2.ext-net-nsxt"
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckExternalNetworkDestroyV2(t.Name()),
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", t.Name()),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					resource.TestCheckResourceAttr(resourceName, "vsphere_network.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "nsxt_network.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "ip_scope.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ip_scope.*", map[string]string{
						"dns1":          "",
						"dns2":          "",
						"dns_suffix":    "",
						"enabled":       "false",
						"gateway":       "192.168.30.49",
						"prefix_length": "24",
					}),

					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ip_scope.*.static_ip_pool.*", map[string]string{
						"start_address": "192.168.30.51",
						"end_address":   "192.168.30.62",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ip_scope.*", map[string]string{
						"dns1":          "",
						"dns2":          "",
						"dns_suffix":    "",
						"enabled":       "true",
						"gateway":       "14.14.14.1",
						"prefix_length": "24",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ip_scope.*.static_ip_pool.*", map[string]string{
						"start_address": "14.14.14.20",
						"end_address":   "14.14.14.25",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ip_scope.*.static_ip_pool.*", map[string]string{
						"start_address": "14.14.14.10",
						"end_address":   "14.14.14.15",
					}),
					resource.TestCheckResourceAttr(resourceName, "nsxt_network.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "nsxt_network.0.nsxt_manager_id"),
					resource.TestCheckResourceAttrSet(resourceName, "nsxt_network.0.nsxt_segment_name"),
				),
			},
			{
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resourceFieldsEqual("data.vcd_external_network_v2.ext-net-nsxt", "vcd_external_network_v2.ext-net-nsxt", nil),
					// Field count differs because data source has `filter` field
					resourceFieldsEqual("data.vcd_network_direct.net", "vcd_network_direct.net", []string{"%"}),
				),
			},
		},
	})

	postTestChecks(t)
}

const testAccCheckVcdExternalNetworkV2NsxtSegmentIntegration = `
data "vcd_nsxt_manager" "main" {
  name = "{{.NsxtManager}}"
}

resource "vcd_external_network_v2" "ext-net-nsxt" {
  name        = "{{.ExternalNetworkName}}"
  description = "{{.Description}}"

  nsxt_network {
    nsxt_manager_id   = data.vcd_nsxt_manager.main.id
    nsxt_segment_name = "{{.NsxtSegment}}"
  }

  ip_scope {
    enabled       = false
    gateway       = "{{.Gateway}}"
    prefix_length = "{{.Netmask}}"

    static_ip_pool {
      start_address = "{{.StartAddress}}"
      end_address   = "{{.EndAddress}}"
    }
  }

  ip_scope {
    gateway       = "14.14.14.1"
    prefix_length = "24"

    static_ip_pool {
      start_address = "14.14.14.10"
      end_address   = "14.14.14.15"
    }
    
    static_ip_pool {
      start_address = "14.14.14.20"
      end_address   = "14.14.14.25"
    }
  }
}

resource "vcd_network_direct" "net" {
  vdc = "{{.NsxtVdc}}"

  name             = "direct-net"
  external_network = vcd_external_network_v2.ext-net-nsxt.name

  depends_on = [vcd_external_network_v2.ext-net-nsxt]
}
`

const testAccCheckVcdExternalNetworkV2NsxtSegmentIntegrationDS = testAccCheckVcdExternalNetworkV2NsxtSegmentIntegration + `
# skip-binary-test: Data Source test 
data "vcd_external_network_v2" "ext-net-nsxt" {
  name = vcd_external_network_v2.ext-net-nsxt.name
}

data "vcd_network_direct" "net" {
  vdc = "{{.NsxtVdc}}"

  name = vcd_network_direct.net.name
}
`

func TestAccVcdExternalNetworkV2NsxtIpSpace(t *testing.T) {
	skipIfNotSysAdmin(t)

	vcdClient := createTemporaryVCDConnection(true)
	if vcdClient == nil {
		t.Skip(acceptanceTestsSkipped)
	}
	if vcdClient.Client.APIVCDMaxVersionIs("< 37.1") {
		t.Skipf("This test tests VCD 10.4.1+ (API V37.1+) features. Skipping.")
	}

	var params = StringMap{
		"Org":                 testConfig.VCD.Org,
		"NsxtManager":         testConfig.Nsxt.Manager,
		"NsxtTier0Router":     testConfig.Nsxt.Tier0router,
		"ExternalNetworkName": t.Name(),

		"Tags": "network extnetwork nsxt",
	}
	testParamsNotEmpty(t, params)

	params["FuncName"] = t.Name() + "step1"
	configText1 := templateFill(testAccVcdExternalNetworkV2NsxtIpSpaceStep1, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText1)

	params["FuncName"] = t.Name() + "step2"
	configText2 := templateFill(testAccVcdExternalNetworkV2NsxtIpSpaceStep2, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText2)

	params["FuncName"] = t.Name() + "step2DS"
	configText3DS := templateFill(testAccVcdExternalNetworkV2NsxtIpSpaceStep2DS, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText3DS)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckExternalNetworkDestroyV2(t.Name()),
		Steps: []resource.TestStep{
			{
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("vcd_external_network_v2.ext-net-nsxt", "name", t.Name()),
					resource.TestCheckResourceAttr("vcd_external_network_v2.ext-net-nsxt", "vsphere_network.#", "0"),
					resource.TestCheckResourceAttr("vcd_external_network_v2.ext-net-nsxt", "nsxt_network.#", "1"),
					resource.TestCheckResourceAttr("vcd_external_network_v2.ext-net-nsxt", "ip_scope.#", "0"),
					resource.TestCheckResourceAttr("vcd_external_network_v2.ext-net-nsxt", "nsxt_network.#", "1"),
					resource.TestCheckResourceAttr("vcd_external_network_v2.ext-net-nsxt", "use_ip_spaces", "true"),
					resource.TestCheckNoResourceAttr("vcd_external_network_v2.ext-net-nsxt", "dedicated_org_id"),
				),
			},
			{
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("vcd_external_network_v2.ext-net-nsxt", "name", t.Name()),
					resource.TestCheckResourceAttr("vcd_external_network_v2.ext-net-nsxt", "vsphere_network.#", "0"),
					resource.TestCheckResourceAttr("vcd_external_network_v2.ext-net-nsxt", "nsxt_network.#", "1"),
					resource.TestCheckResourceAttr("vcd_external_network_v2.ext-net-nsxt", "ip_scope.#", "0"),
					resource.TestCheckResourceAttr("vcd_external_network_v2.ext-net-nsxt", "nsxt_network.#", "1"),
					resource.TestCheckResourceAttr("vcd_external_network_v2.ext-net-nsxt", "use_ip_spaces", "true"),
					resource.TestCheckResourceAttrSet("vcd_external_network_v2.ext-net-nsxt", "dedicated_org_id"),
				),
			},
			{
				ResourceName:      "vcd_external_network_v2.ext-net-nsxt",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdTopHierarchy(t.Name()),
			},
			{
				Taint:  []string{"vcd_external_network_v2.ext-net-nsxt"},
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("vcd_external_network_v2.ext-net-nsxt", "name", t.Name()),
					resource.TestCheckResourceAttr("vcd_external_network_v2.ext-net-nsxt", "vsphere_network.#", "0"),
					resource.TestCheckResourceAttr("vcd_external_network_v2.ext-net-nsxt", "nsxt_network.#", "1"),
					resource.TestCheckResourceAttr("vcd_external_network_v2.ext-net-nsxt", "ip_scope.#", "0"),
					resource.TestCheckResourceAttr("vcd_external_network_v2.ext-net-nsxt", "nsxt_network.#", "1"),
					resource.TestCheckResourceAttr("vcd_external_network_v2.ext-net-nsxt", "use_ip_spaces", "true"),
					resource.TestCheckResourceAttrSet("vcd_external_network_v2.ext-net-nsxt", "dedicated_org_id"),
				),
			},
			{
				Config: configText3DS,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("vcd_external_network_v2.ext-net-nsxt", "name", t.Name()),
					resourceFieldsEqual("data.vcd_external_network_v2.ext-net-nsxt", "vcd_external_network_v2.ext-net-nsxt", nil),
				),
			},
		},
	})
}

const testAccVcdExternalNetworkV2NsxtIpSpaceStep1 = testAccCheckVcdExternalNetworkV2NsxtDS + `
resource "vcd_external_network_v2" "ext-net-nsxt" {
  name = "{{.ExternalNetworkName}}"

  nsxt_network {
    nsxt_manager_id      = data.vcd_nsxt_manager.main.id
    nsxt_tier0_router_id = data.vcd_nsxt_tier0_router.router.id
  }

  use_ip_spaces = true
}
`

const testAccVcdExternalNetworkV2NsxtIpSpaceStep2 = testAccCheckVcdExternalNetworkV2NsxtDS + `
data "vcd_org" "org1" {
  name = "{{.Org}}"
}

resource "vcd_external_network_v2" "ext-net-nsxt" {
  name = "{{.ExternalNetworkName}}"

  nsxt_network {
    nsxt_manager_id      = data.vcd_nsxt_manager.main.id
    nsxt_tier0_router_id = data.vcd_nsxt_tier0_router.router.id
  }

  use_ip_spaces    = true
  dedicated_org_id = data.vcd_org.org1.id
}
`

const testAccVcdExternalNetworkV2NsxtIpSpaceStep2DS = testAccCheckVcdExternalNetworkV2NsxtDS + `
data "vcd_org" "org1" {
  name = "{{.Org}}"
}

data "vcd_external_network_v2" "ext-net-nsxt" {
  name = vcd_external_network_v2.ext-net-nsxt.name
  
  depends_on = [vcd_external_network_v2.ext-net-nsxt]
}

resource "vcd_external_network_v2" "ext-net-nsxt" {
  name = "{{.ExternalNetworkName}}"

  nsxt_network {
    nsxt_manager_id      = data.vcd_nsxt_manager.main.id
    nsxt_tier0_router_id = data.vcd_nsxt_tier0_router.router.id
  }

  use_ip_spaces    = true
  dedicated_org_id = data.vcd_org.org1.id
}
`
