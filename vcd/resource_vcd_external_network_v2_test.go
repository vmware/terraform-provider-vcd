// +build functional network extnetwork nsxt ALL

package vcd

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/vmware/go-vcloud-director/v2/govcd"

	"github.com/hashicorp/terraform-plugin-sdk/terraform"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccVcdExternalNetworkV2NsxtVrf(t *testing.T) {
	testAccVcdExternalNetworkV2Nsxt(t, testConfig.Nsxt.Tier0routerVrf)
}

func TestAccVcdExternalNetworkV2Nsxt(t *testing.T) {
	testAccVcdExternalNetworkV2Nsxt(t, testConfig.Nsxt.Tier0router)
}

func testAccVcdExternalNetworkV2Nsxt(t *testing.T, nsxtTier0Router string) {

	if !usingSysAdmin() {
		t.Skip(t.Name() + " requires system admin privileges")
		return
	}

	skipNoNsxtConfiguration(t)
	vcdClient := createTemporaryVCDConnection()
	if vcdClient.Client.APIVCDMaxVersionIs("< 33.0") {
		t.Skip(t.Name() + " requires at least API v33.0 (vCD 10+)")
	}

	startAddress := "192.168.30.51"
	endAddress := "192.168.30.62"
	description := "Test External Network"
	var params = StringMap{
		"NsxtManager":         testConfig.Nsxt.Manager,
		"NsxtTier0Router":     nsxtTier0Router,
		"ExternalNetworkName": t.Name(),
		"Type":                testConfig.Networking.ExternalNetworkPortGroupType,
		"PortGroup":           testConfig.Networking.ExternalNetworkPortGroup,
		"Vcenter":             testConfig.Networking.Vcenter,
		"StartAddress":        startAddress,
		"EndAddress":          endAddress,
		"Description":         description,
		"Gateway":             "192.168.30.49",
		"Netmask":             "24",
		"Tags":                "network extnetwork nsxt",
	}

	params["FuncName"] = t.Name()
	configText := templateFill(testAccCheckVcdExternalNetworkV2Nsxt, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

	params["FuncName"] = t.Name() + "step1"
	configText1 := templateFill(testAccCheckVcdExternalNetworkV2NsxtStep1, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText1)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	resourceName := "vcd_external_network_v2.ext-net-nsxt"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckExternalNetworkDestroyV2(t.Name()),
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", t.Name()),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					resource.TestCheckResourceAttr(resourceName, "vsphere_network.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "nsxt_network.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "ip_scope.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "ip_scope.1420917927.dns1", ""),
					resource.TestCheckResourceAttr(resourceName, "ip_scope.1420917927.dns2", ""),
					resource.TestCheckResourceAttr(resourceName, "ip_scope.1420917927.dns_suffix", ""),
					resource.TestCheckResourceAttr(resourceName, "ip_scope.1420917927.enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "ip_scope.1420917927.gateway", "192.168.30.49"),
					resource.TestCheckResourceAttr(resourceName, "ip_scope.1420917927.prefix_length", "24"),
					resource.TestCheckResourceAttr(resourceName, "ip_scope.1420917927.static_ip_pool.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "ip_scope.1420917927.static_ip_pool.1203345861.end_address", "192.168.30.62"),
					resource.TestCheckResourceAttr(resourceName, "ip_scope.1420917927.static_ip_pool.1203345861.start_address", "192.168.30.51"),
					resource.TestCheckResourceAttr(resourceName, "ip_scope.3421983869.dns1", ""),
					resource.TestCheckResourceAttr(resourceName, "ip_scope.3421983869.dns2", ""),
					resource.TestCheckResourceAttr(resourceName, "ip_scope.3421983869.dns_suffix", ""),
					resource.TestCheckResourceAttr(resourceName, "ip_scope.3421983869.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "ip_scope.3421983869.gateway", "14.14.14.1"),
					resource.TestCheckResourceAttr(resourceName, "ip_scope.3421983869.prefix_length", "24"),
					resource.TestCheckResourceAttr(resourceName, "ip_scope.3421983869.static_ip_pool.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "ip_scope.3421983869.static_ip_pool.2275320158.end_address", "14.14.14.25"),
					resource.TestCheckResourceAttr(resourceName, "ip_scope.3421983869.static_ip_pool.2275320158.start_address", "14.14.14.20"),
					resource.TestCheckResourceAttr(resourceName, "ip_scope.3421983869.static_ip_pool.550532203.end_address", "14.14.14.15"),
					resource.TestCheckResourceAttr(resourceName, "ip_scope.3421983869.static_ip_pool.550532203.start_address", "14.14.14.10"),
					resource.TestCheckResourceAttr(resourceName, "nsxt_network.#", "1"),
					testCheckMatchOutput("nsxt-manager", regexp.MustCompile("^urn:vcloud:nsxtmanager:.*")),
					testCheckOutputNonEmpty("nsxt-tier0-router"), // Match any non empty string
				),
			},
			resource.TestStep{
				ResourceName:      resourceName + "-import",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdTopHierarchy(t.Name()),
			},
			resource.TestStep{
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", t.Name()),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					resource.TestCheckResourceAttr(resourceName, "vsphere_network.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "nsxt_network.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "ip_scope.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "ip_scope.1428757071.dns1", ""),
					resource.TestCheckResourceAttr(resourceName, "ip_scope.1428757071.dns2", ""),
					resource.TestCheckResourceAttr(resourceName, "ip_scope.1428757071.dns_suffix", ""),
					resource.TestCheckResourceAttr(resourceName, "ip_scope.1428757071.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "ip_scope.1428757071.gateway", "192.168.30.49"),
					resource.TestCheckResourceAttr(resourceName, "ip_scope.1428757071.prefix_length", "24"),
					resource.TestCheckResourceAttr(resourceName, "ip_scope.1428757071.static_ip_pool.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "ip_scope.1428757071.static_ip_pool.1203345861.end_address", "192.168.30.62"),
					resource.TestCheckResourceAttr(resourceName, "ip_scope.1428757071.static_ip_pool.1203345861.start_address", "192.168.30.51"),
					resource.TestCheckResourceAttr(resourceName, "nsxt_network.#", "1"),
					testCheckMatchOutput("nsxt-manager", regexp.MustCompile("^urn:vcloud:nsxtmanager:.*")),
					testCheckOutputNonEmpty("nsxt-tier0-router"), // Match any non empty string

					// Data source
					resource.TestCheckResourceAttrPair(resourceName, "name", "data."+resourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "description", "data."+resourceName, "description"),
					resource.TestCheckResourceAttrPair(resourceName, "ip_scope.#", "data."+resourceName, "ip_scope.#"),
					resource.TestCheckResourceAttrPair(resourceName, "ip_scope.1428757071.dns1", "data."+resourceName, "ip_scope.1428757071.dns1"),
					resource.TestCheckResourceAttrPair(resourceName, "ip_scope.1428757071.dns2", "data."+resourceName, "ip_scope.1428757071.dns2"),
					resource.TestCheckResourceAttrPair(resourceName, "ip_scope.1428757071.dns_suffix", "data."+resourceName, "ip_scope.1428757071.dns_suffix"),
					resource.TestCheckResourceAttrPair(resourceName, "ip_scope.1428757071.enabled", "data."+resourceName, "ip_scope.1428757071.enabled"),
					resource.TestCheckResourceAttrPair(resourceName, "ip_scope.1428757071.gateway", "data."+resourceName, "ip_scope.1428757071.gateway"),
					resource.TestCheckResourceAttrPair(resourceName, "ip_scope.1428757071.prefix_length", "data."+resourceName, "ip_scope.1428757071.prefix_length"),
					resource.TestCheckResourceAttrPair(resourceName, "ip_scope.1428757071.static_ip_pool.#", "data."+resourceName, "ip_scope.1428757071.static_ip_pool.#"),
					resource.TestCheckResourceAttrPair(resourceName, "ip_scope.1428757071.static_ip_pool.1203345861.end_address", "data."+resourceName, "ip_scope.1428757071.static_ip_pool.1203345861.end_address"),
					resource.TestCheckResourceAttrPair(resourceName, "ip_scope.1428757071.static_ip_pool.1203345861.start_address", "data."+resourceName, "ip_scope.1428757071.static_ip_pool.1203345861.start_address"),
					resource.TestCheckResourceAttrPair(resourceName, "vsphere_network.#", "data."+resourceName, "vsphere_network.#"),
					resource.TestCheckResourceAttrPair(resourceName, "nsxt_network.#", "data."+resourceName, "nsxt_network.#"),
					resource.TestMatchResourceAttr("data."+resourceName, "nsxt_network.0.nsxt_manager_id", regexp.MustCompile("^urn:vcloud:nsxtmanager:.*")),
					resource.TestCheckResourceAttrSet("data."+resourceName, "nsxt_network.0.nsxt_tier0_router_id"),
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

const testAccCheckVcdExternalNetworkV2Nsxt = testAccCheckVcdExternalNetworkV2NsxtDS + `
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
const testAccCheckVcdExternalNetworkV2NsxtStep1 = testAccCheckVcdExternalNetworkV2NsxtDS + `
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

data "vcd_external_network_v2" "ext-net-nsxt" {
	name = vcd_external_network_v2.ext-net-nsxt.name
}

output "nsxt-manager" {
  value = tolist(vcd_external_network_v2.ext-net-nsxt.nsxt_network)[0].nsxt_manager_id
}

output "nsxt-tier0-router" {
  value = tolist(vcd_external_network_v2.ext-net-nsxt.nsxt_network)[0].nsxt_tier0_router_id
}
`

func TestAccVcdExternalNetworkV2Nsxv(t *testing.T) {
	if !usingSysAdmin() {
		t.Skip(t.Name() + " requires system admin privileges")
		return
	}

	vcdClient := createTemporaryVCDConnection()
	if vcdClient.Client.APIVCDMaxVersionIs("< 33.0") {
		t.Skip(t.Name() + " requires at least API v33.0 (vCD 10+)")
	}

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

	configText := templateFill(testAccCheckVcdExternalNetworkV2Nsxv, params)
	params["FuncName"] = t.Name() + "step1"
	configText1 := templateFill(testAccCheckVcdExternalNetworkV2NsxvUpdate, params)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText1)

	resourceName := "vcd_external_network_v2.ext-net-nsxv"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckExternalNetworkDestroyV2(t.Name()),
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", t.Name()),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					resource.TestCheckResourceAttr(resourceName, "ip_scope.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "ip_scope.2118535427.dns1", "192.168.0.164"),
					resource.TestCheckResourceAttr(resourceName, "ip_scope.2118535427.dns2", "192.168.0.196"),
					resource.TestCheckResourceAttr(resourceName, "ip_scope.2118535427.dns_suffix", "company.biz"),
					resource.TestCheckResourceAttr(resourceName, "ip_scope.2118535427.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "ip_scope.2118535427.gateway", "192.168.30.49"),
					resource.TestCheckResourceAttr(resourceName, "ip_scope.2118535427.prefix_length", "24"),
					resource.TestCheckResourceAttr(resourceName, "ip_scope.2118535427.static_ip_pool.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "ip_scope.2118535427.static_ip_pool.1203345861.end_address", "192.168.30.62"),
					resource.TestCheckResourceAttr(resourceName, "ip_scope.2118535427.static_ip_pool.1203345861.start_address", "192.168.30.51"),
					resource.TestCheckResourceAttr(resourceName, "nsxt_network.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "vsphere_network.#", "1"),
					testCheckOutputNonEmpty("vcenter-id"),   // Match any non empty string
					testCheckOutputNonEmpty("portgroup-id"), // Match any non empty string
				),
			},
			resource.TestStep{
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", t.Name()),
					resource.TestCheckResourceAttr(resourceName, "description", description),
					resource.TestCheckResourceAttr(resourceName, "ip_scope.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "ip_scope.2145267691.dns1", "192.168.0.164"),
					resource.TestCheckResourceAttr(resourceName, "ip_scope.2145267691.dns2", "192.168.0.196"),
					resource.TestCheckResourceAttr(resourceName, "ip_scope.2145267691.dns_suffix", "company.biz"),
					resource.TestCheckResourceAttr(resourceName, "ip_scope.2145267691.enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "ip_scope.2145267691.gateway", "192.168.30.49"),
					resource.TestCheckResourceAttr(resourceName, "ip_scope.2145267691.prefix_length", "24"),
					resource.TestCheckResourceAttr(resourceName, "ip_scope.2145267691.static_ip_pool.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "ip_scope.2145267691.static_ip_pool.1203345861.end_address", "192.168.30.62"),
					resource.TestCheckResourceAttr(resourceName, "ip_scope.2145267691.static_ip_pool.1203345861.start_address", "192.168.30.51"),
					resource.TestCheckResourceAttr(resourceName, "ip_scope.801323554.dns1", "8.8.8.8"),
					resource.TestCheckResourceAttr(resourceName, "ip_scope.801323554.dns2", "8.8.4.4"),
					resource.TestCheckResourceAttr(resourceName, "ip_scope.801323554.dns_suffix", "asd.biz"),
					resource.TestCheckResourceAttr(resourceName, "ip_scope.801323554.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "ip_scope.801323554.gateway", "88.88.88.1"),
					resource.TestCheckResourceAttr(resourceName, "ip_scope.801323554.prefix_length", "24"),
					resource.TestCheckResourceAttr(resourceName, "ip_scope.801323554.static_ip_pool.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "ip_scope.801323554.static_ip_pool.2396875145.end_address", "88.88.88.100"),
					resource.TestCheckResourceAttr(resourceName, "ip_scope.801323554.static_ip_pool.2396875145.start_address", "88.88.88.10"),
					resource.TestCheckResourceAttr(resourceName, "nsxt_network.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "vsphere_network.#", "1"),
					testCheckMatchOutput("vcenter-id", regexp.MustCompile("^urn:vcloud:vimserver:.*")),
					testCheckOutputNonEmpty("portgroup-id"), // Match any non empty string because IDs may differ
				),
			},
			resource.TestStep{
				ResourceName:      resourceName + "-import",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdTopHierarchy(t.Name()),
			},
		},
	})
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
