// +build network nsxt ALL functional

package vcd

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// TestAccVcdNsxtIpSetEmptyStart starts with an IP set with no IP addresses defined, updates it and tries to add IP
// addresses
func TestAccVcdNsxtIpSetEmptyStart(t *testing.T) {
	preTestChecks(t)
	skipNoNsxtConfiguration(t)

	// String map to fill the template
	var params = StringMap{
		"Org":         testConfig.VCD.Org,
		"NsxtVdc":     testConfig.Nsxt.Vdc,
		"EdgeGw":      testConfig.Nsxt.EdgeGateway,
		"NetworkName": t.Name(),
		"Tags":        "network nsxt",
	}

	configText := templateFill(testAccNsxtIpSetEmpty, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText)

	params["FuncName"] = t.Name() + "-step1"
	configText1 := templateFill(testAccNsxtIpSetEmpty2, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText1)

	params["FuncName"] = t.Name() + "-step11"
	configText11 := templateFill(testAccNsxtIpSetEmpty2+testAccNsxtIpSetDS, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 3: %s", configText11)

	params["FuncName"] = t.Name() + "-step2"
	configText2 := templateFill(testAccNsxtIpSetIpRanges, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 4: %s", configText2)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckNsxtFirewallGroupDestroy(testConfig.Nsxt.Vdc, "test-ip-set", types.FirewallGroupTypeIpSet),
			testAccCheckNsxtFirewallGroupDestroy(testConfig.Nsxt.Vdc, "test-ip-set-changed", types.FirewallGroupTypeIpSet),
		),
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_ip_set.set1", "id", regexp.MustCompile(`^urn:vcloud:firewallGroup:.*$`)),
					resource.TestCheckResourceAttr("vcd_nsxt_ip_set.set1", "name", "test-ip-set"),
					resource.TestCheckResourceAttr("vcd_nsxt_ip_set.set1", "description", "test-ip-set-description"),
					resource.TestCheckResourceAttr("vcd_nsxt_ip_set.set1", "ip_addresses.#", "0"),
				),
			},
			resource.TestStep{
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_ip_set.set1", "id", regexp.MustCompile(`^urn:vcloud:firewallGroup:.*$`)),
					resource.TestCheckResourceAttr("vcd_nsxt_ip_set.set1", "name", "test-ip-set-changed"),
					resource.TestCheckResourceAttr("vcd_nsxt_ip_set.set1", "description", ""),
					resource.TestCheckResourceAttr("vcd_nsxt_ip_set.set1", "ip_addresses.#", "0"),
				),
			},
			resource.TestStep{
				Config: configText11,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_ip_set.set1", "id", regexp.MustCompile(`^urn:vcloud:firewallGroup:.*$`)),
					resource.TestCheckResourceAttr("vcd_nsxt_ip_set.set1", "name", "test-ip-set-changed"),
					resource.TestCheckResourceAttr("vcd_nsxt_ip_set.set1", "description", ""),
					resource.TestCheckResourceAttr("vcd_nsxt_ip_set.set1", "ip_addresses.#", "0"),

					resourceFieldsEqual("vcd_nsxt_ip_set.set1", "data.vcd_nsxt_ip_set.ds", []string{}),
				),
			},
			// Test import with no IP addresses
			resource.TestStep{
				ResourceName:      "vcd_nsxt_ip_set.set1",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdNsxtEdgeGatewayObject(testConfig, testConfig.Nsxt.EdgeGateway, "test-ip-set-changed"),
			},
			resource.TestStep{
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_ip_set.set1", "id", regexp.MustCompile(`^urn:vcloud:firewallGroup:.*$`)),
					resource.TestCheckResourceAttr("vcd_nsxt_ip_set.set1", "name", "test-ip-set-changed"),
					resource.TestCheckResourceAttr("vcd_nsxt_ip_set.set1", "description", ""),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ip_set.set1", "ip_addresses.*", "12.12.12.1"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ip_set.set1", "ip_addresses.*", "10.10.10.0/24"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ip_set.set1", "ip_addresses.*", "11.11.11.1-11.11.11.2"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ip_set.set1", "ip_addresses.*", "2001:db8::/48"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ip_set.set1", "ip_addresses.*", "2001:db6:0:0:0:0:0:0-2001:db6:0:ffff:ffff:ffff:ffff:ffff"),
					resource.TestCheckResourceAttr("vcd_nsxt_ip_set.set1", "ip_addresses.#", "5"),
				),
			},
			// Test import with IP addresses
			resource.TestStep{
				ResourceName:      "vcd_nsxt_ip_set.set1",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdNsxtEdgeGatewayObject(testConfig, testConfig.Nsxt.EdgeGateway, "test-ip-set-changed"),
			},
		},
	})
	postTestChecks(t)
}

// TestAccVcdNsxtIpSet starts with creating an IP Set with IP addresses defined and later on removes them all
func TestAccVcdNsxtIpSet(t *testing.T) {
	preTestChecks(t)
	skipNoNsxtConfiguration(t)

	// String map to fill the template
	var params = StringMap{
		"Org":         testConfig.VCD.Org,
		"NsxtVdc":     testConfig.Nsxt.Vdc,
		"EdgeGw":      testConfig.Nsxt.EdgeGateway,
		"NetworkName": t.Name(),
		"Tags":        "network nsxt",
	}

	configText := templateFill(testAccNsxtIpSetIpRanges, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText)

	params["FuncName"] = t.Name() + "-step1"
	configText1 := templateFill(testAccNsxtIpSetIpRangesRemoved, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText1)

	params["FuncName"] = t.Name() + "-step1DS"
	configText11 := templateFill(testAccNsxtIpSetIpRangesRemoved+testAccNsxtIpSetDS, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 3: %s", configText11)

	params["FuncName"] = t.Name() + "-step2"
	configText2 := templateFill(testAccNsxtIpSetEmpty, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 4: %s", configText2)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckNsxtFirewallGroupDestroy(testConfig.Nsxt.Vdc, "test-ip-set", types.FirewallGroupTypeIpSet),
			testAccCheckNsxtFirewallGroupDestroy(testConfig.Nsxt.Vdc, "test-ip-set-changed", types.FirewallGroupTypeIpSet),
		),
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_ip_set.set1", "id", regexp.MustCompile(`^urn:vcloud:firewallGroup:.*$`)),
					resource.TestCheckResourceAttr("vcd_nsxt_ip_set.set1", "name", "test-ip-set-changed"),
					resource.TestCheckResourceAttr("vcd_nsxt_ip_set.set1", "description", ""),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ip_set.set1", "ip_addresses.*", "12.12.12.1"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ip_set.set1", "ip_addresses.*", "10.10.10.0/24"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ip_set.set1", "ip_addresses.*", "11.11.11.1-11.11.11.2"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ip_set.set1", "ip_addresses.*", "2001:db8::/48"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ip_set.set1", "ip_addresses.*", "2001:db6:0:0:0:0:0:0-2001:db6:0:ffff:ffff:ffff:ffff:ffff"),
					resource.TestCheckResourceAttr("vcd_nsxt_ip_set.set1", "ip_addresses.#", "5"),
				),
			},
			resource.TestStep{
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_ip_set.set1", "id", regexp.MustCompile(`^urn:vcloud:firewallGroup:.*$`)),
					resource.TestCheckResourceAttr("vcd_nsxt_ip_set.set1", "name", "test-ip-set-changed"),
					resource.TestCheckResourceAttr("vcd_nsxt_ip_set.set1", "description", ""),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ip_set.set1", "ip_addresses.*", "12.12.12.1"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ip_set.set1", "ip_addresses.*", "2001:db6:0:0:0:0:0:0-2001:db6:0:ffff:ffff:ffff:ffff:ffff"),
					resource.TestCheckResourceAttr("vcd_nsxt_ip_set.set1", "ip_addresses.#", "2"),
				),
			},
			resource.TestStep{
				Config: configText11,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_ip_set.set1", "id", regexp.MustCompile(`^urn:vcloud:firewallGroup:.*$`)),
					resource.TestCheckResourceAttr("vcd_nsxt_ip_set.set1", "name", "test-ip-set-changed"),
					resource.TestCheckResourceAttr("vcd_nsxt_ip_set.set1", "description", ""),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ip_set.set1", "ip_addresses.*", "12.12.12.1"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ip_set.set1", "ip_addresses.*", "2001:db6:0:0:0:0:0:0-2001:db6:0:ffff:ffff:ffff:ffff:ffff"),
					resource.TestCheckResourceAttr("vcd_nsxt_ip_set.set1", "ip_addresses.#", "2"),

					resourceFieldsEqual("vcd_nsxt_ip_set.set1", "data.vcd_nsxt_ip_set.ds", []string{}),
				),
			},
			// Test import with IP addresses
			resource.TestStep{
				ResourceName:      "vcd_nsxt_ip_set.set1",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdNsxtEdgeGatewayObject(testConfig, testConfig.Nsxt.EdgeGateway, "test-ip-set-changed"),
			},
			resource.TestStep{
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_ip_set.set1", "id", regexp.MustCompile(`^urn:vcloud:firewallGroup:.*$`)),
					resource.TestCheckResourceAttr("vcd_nsxt_ip_set.set1", "name", "test-ip-set"),
					resource.TestCheckResourceAttr("vcd_nsxt_ip_set.set1", "description", "test-ip-set-description"),
					resource.TestCheckResourceAttr("vcd_nsxt_ip_set.set1", "ip_addresses.#", "0"),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccNsxtIpSetPrereqs = `
data "vcd_nsxt_edgegateway" "existing_gw" {
  org = "{{.Org}}"
  vdc = "{{.NsxtVdc}}"

  name = "{{.EdgeGw}}"
}
`

const testAccNsxtIpSetEmpty = testAccNsxtIpSetPrereqs + `
resource "vcd_nsxt_ip_set" "set1" {
  org = "{{.Org}}"
  vdc = "{{.NsxtVdc}}"

  edge_gateway_id = data.vcd_nsxt_edgegateway.existing_gw.id

  name        = "test-ip-set"
  description = "test-ip-set-description"
}
`

const testAccNsxtIpSetEmpty2 = testAccNsxtIpSetPrereqs + `
resource "vcd_nsxt_ip_set" "set1" {
  org  = "{{.Org}}"
  vdc  = "{{.NsxtVdc}}"

  edge_gateway_id = data.vcd_nsxt_edgegateway.existing_gw.id

  name = "test-ip-set-changed"
}
`

const testAccNsxtIpSetIpRanges = testAccNsxtIpSetPrereqs + `
resource "vcd_nsxt_ip_set" "set1" {
  org = "{{.Org}}"
  vdc = "{{.NsxtVdc}}"

  edge_gateway_id = data.vcd_nsxt_edgegateway.existing_gw.id

  name = "test-ip-set-changed"

  ip_addresses = [
    "12.12.12.1",
    "10.10.10.0/24",
    "11.11.11.1-11.11.11.2",
    "2001:db8::/48",
    "2001:db6:0:0:0:0:0:0-2001:db6:0:ffff:ffff:ffff:ffff:ffff",
  ]
}
`

const testAccNsxtIpSetIpRangesRemoved = testAccNsxtIpSetPrereqs + `
resource "vcd_nsxt_ip_set" "set1" {
  org = "{{.Org}}"
  vdc = "{{.NsxtVdc}}"

  edge_gateway_id = data.vcd_nsxt_edgegateway.existing_gw.id

  name = "test-ip-set-changed"

  ip_addresses = [
    "12.12.12.1",
    "2001:db6:0:0:0:0:0:0-2001:db6:0:ffff:ffff:ffff:ffff:ffff",
  ]
}
`

const testAccNsxtIpSetDS = `
# skip-binary-test: Terraform resource cannot have resource and datasource in the same file
data "vcd_nsxt_ip_set" "ds" {
  org = "{{.Org}}"
  vdc = "{{.NsxtVdc}}"

  edge_gateway_id = data.vcd_nsxt_edgegateway.existing_gw.id
  name            = "test-ip-set-changed"
}
`
