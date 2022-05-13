//go:build network || nsxt || ALL || functional
// +build network nsxt ALL functional

package vcd

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// TestAccVcdNsxtIpSetEmptyStart starts with an IP Set with no IP addresses defined, updates it and tries to add IP
// addresses
func TestAccVcdNsxtIpSetEmptyStart(t *testing.T) {
	preTestChecks(t)

	// String map to fill the template
	var params = StringMap{
		"Org":         testConfig.VCD.Org,
		"NsxtVdc":     testConfig.Nsxt.Vdc,
		"EdgeGw":      testConfig.Nsxt.EdgeGateway,
		"NetworkName": t.Name(),
		"Tags":        "network nsxt",
	}
	testParamsNotEmpty(t, params)

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
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckNsxtFirewallGroupDestroy(testConfig.Nsxt.Vdc, "test-ip-set", types.FirewallGroupTypeIpSet),
			testAccCheckNsxtFirewallGroupDestroy(testConfig.Nsxt.Vdc, "test-ip-set-changed", types.FirewallGroupTypeIpSet),
		),
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_ip_set.set1", "id", regexp.MustCompile(`^urn:vcloud:firewallGroup:.*$`)),
					resource.TestCheckResourceAttr("vcd_nsxt_ip_set.set1", "name", "test-ip-set"),
					resource.TestCheckResourceAttr("vcd_nsxt_ip_set.set1", "description", "test-ip-set-description"),
					resource.TestCheckResourceAttr("vcd_nsxt_ip_set.set1", "ip_addresses.#", "0"),
				),
			},
			{
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_ip_set.set1", "id", regexp.MustCompile(`^urn:vcloud:firewallGroup:.*$`)),
					resource.TestCheckResourceAttr("vcd_nsxt_ip_set.set1", "name", "test-ip-set-changed"),
					resource.TestCheckResourceAttr("vcd_nsxt_ip_set.set1", "description", ""),
					resource.TestCheckResourceAttr("vcd_nsxt_ip_set.set1", "ip_addresses.#", "0"),
				),
			},
			{
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
			{
				ResourceName:      "vcd_nsxt_ip_set.set1",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdNsxtEdgeGatewayObject(testConfig, testConfig.Nsxt.EdgeGateway, "test-ip-set-changed"),
			},
			{
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
			{
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

	// String map to fill the template
	var params = StringMap{
		"Org":         testConfig.VCD.Org,
		"NsxtVdc":     testConfig.Nsxt.Vdc,
		"EdgeGw":      testConfig.Nsxt.EdgeGateway,
		"NetworkName": t.Name(),
		"Tags":        "network nsxt",
	}
	testParamsNotEmpty(t, params)

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
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckNsxtFirewallGroupDestroy(testConfig.Nsxt.Vdc, "test-ip-set", types.FirewallGroupTypeIpSet),
			testAccCheckNsxtFirewallGroupDestroy(testConfig.Nsxt.Vdc, "test-ip-set-changed", types.FirewallGroupTypeIpSet),
		),
		Steps: []resource.TestStep{
			{
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
					resource.TestMatchResourceAttr("vcd_nsxt_ip_set.set1", "owner_id", regexp.MustCompile(`^urn:vcloud:vdc:.*$`)),
				),
			},
			{
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_ip_set.set1", "id", regexp.MustCompile(`^urn:vcloud:firewallGroup:.*$`)),
					resource.TestCheckResourceAttr("vcd_nsxt_ip_set.set1", "name", "test-ip-set-changed"),
					resource.TestCheckResourceAttr("vcd_nsxt_ip_set.set1", "description", ""),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ip_set.set1", "ip_addresses.*", "12.12.12.1"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ip_set.set1", "ip_addresses.*", "2001:db6:0:0:0:0:0:0-2001:db6:0:ffff:ffff:ffff:ffff:ffff"),
					resource.TestCheckResourceAttr("vcd_nsxt_ip_set.set1", "ip_addresses.#", "2"),
					resource.TestMatchResourceAttr("vcd_nsxt_ip_set.set1", "owner_id", regexp.MustCompile(`^urn:vcloud:vdc:.*$`)),
				),
			},
			{
				Config: configText11,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_ip_set.set1", "id", regexp.MustCompile(`^urn:vcloud:firewallGroup:.*$`)),
					resource.TestCheckResourceAttr("vcd_nsxt_ip_set.set1", "name", "test-ip-set-changed"),
					resource.TestCheckResourceAttr("vcd_nsxt_ip_set.set1", "description", ""),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ip_set.set1", "ip_addresses.*", "12.12.12.1"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ip_set.set1", "ip_addresses.*", "2001:db6:0:0:0:0:0:0-2001:db6:0:ffff:ffff:ffff:ffff:ffff"),
					resource.TestCheckResourceAttr("vcd_nsxt_ip_set.set1", "ip_addresses.#", "2"),
					resource.TestMatchResourceAttr("vcd_nsxt_ip_set.set1", "owner_id", regexp.MustCompile(`^urn:vcloud:vdc:.*$`)),

					resourceFieldsEqual("vcd_nsxt_ip_set.set1", "data.vcd_nsxt_ip_set.ds", []string{}),
				),
			},
			// Test import with IP addresses
			{
				ResourceName:      "vcd_nsxt_ip_set.set1",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdNsxtEdgeGatewayObject(testConfig, testConfig.Nsxt.EdgeGateway, "test-ip-set-changed"),
			},
			{
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_ip_set.set1", "id", regexp.MustCompile(`^urn:vcloud:firewallGroup:.*$`)),
					resource.TestCheckResourceAttr("vcd_nsxt_ip_set.set1", "name", "test-ip-set"),
					resource.TestCheckResourceAttr("vcd_nsxt_ip_set.set1", "description", "test-ip-set-description"),
					resource.TestCheckResourceAttr("vcd_nsxt_ip_set.set1", "ip_addresses.#", "0"),
					resource.TestMatchResourceAttr("vcd_nsxt_ip_set.set1", "owner_id", regexp.MustCompile(`^urn:vcloud:vdc:.*$`)),
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

  edge_gateway_id = data.vcd_nsxt_edgegateway.existing_gw.id

  name        = "test-ip-set"
  description = "test-ip-set-description"
}
`

const testAccNsxtIpSetEmpty2 = testAccNsxtIpSetPrereqs + `
resource "vcd_nsxt_ip_set" "set1" {
  org  = "{{.Org}}"

  edge_gateway_id = data.vcd_nsxt_edgegateway.existing_gw.id

  name = "test-ip-set-changed"
}
`

const testAccNsxtIpSetIpRanges = testAccNsxtIpSetPrereqs + `
resource "vcd_nsxt_ip_set" "set1" {
  org = "{{.Org}}"

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

  edge_gateway_id = data.vcd_nsxt_edgegateway.existing_gw.id
  name            = "test-ip-set-changed"
}
`

// TestAccVcdNsxtIpSetOwnerVdcGroup starts with creating an IP Set with IP addresses defined in VDC Group and later on removes them all
func TestAccVcdNsxtIpSetOwnerVdcGroup(t *testing.T) {
	preTestChecks(t)
	if !usingSysAdmin() {
		t.Skipf("this test requires Sysadmin user to create prerequisites")
	}

	// String map to fill the template
	var params = StringMap{
		"Org":         testConfig.VCD.Org,
		"NsxtVdc":     testConfig.Nsxt.Vdc,
		"EdgeGw":      testConfig.Nsxt.EdgeGateway,
		"NetworkName": t.Name(),
		"Tags":        "network nsxt",

		"Name":                      t.Name(),
		"Dfw":                       "false",
		"DefaultPolicy":             "false",
		"ProviderVdc":               testConfig.VCD.NsxtProviderVdc.Name,
		"NetworkPool":               testConfig.VCD.NsxtProviderVdc.NetworkPool,
		"ProviderVdcStorageProfile": testConfig.VCD.ProviderVdc.StorageProfile,
		"ExternalNetwork":           testConfig.Nsxt.ExternalNetwork,
		"NsxtEdgeGatewayVcd":        t.Name() + "-edge",
		"TestName":                  t.Name(),
	}
	testParamsNotEmpty(t, params)

	configText := templateFill(testAccNsxtIpSetOwnByVdcGroup, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText)
	params["FuncName"] = t.Name() + "-step2"
	configText2 := templateFill(testAccNsxtIpSetOwnByVdcGroupUpdate, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 4: %s", configText2)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckNsxtFirewallGroupDestroy(testConfig.Nsxt.Vdc, "test-ip-set", types.FirewallGroupTypeIpSet),
		),
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_ip_set.set1", "id", regexp.MustCompile(`^urn:vcloud:firewallGroup:.*$`)),
					resource.TestCheckResourceAttr("vcd_nsxt_ip_set.set1", "name", "test-ip-set"),
					resource.TestCheckResourceAttr("vcd_nsxt_ip_set.set1", "description", ""),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ip_set.set1", "ip_addresses.*", "12.12.12.1"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ip_set.set1", "ip_addresses.*", "10.10.10.0/24"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ip_set.set1", "ip_addresses.*", "11.11.11.1-11.11.11.2"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ip_set.set1", "ip_addresses.*", "2001:db8::/48"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ip_set.set1", "ip_addresses.*", "2001:db6:0:0:0:0:0:0-2001:db6:0:ffff:ffff:ffff:ffff:ffff"),
					resource.TestCheckResourceAttr("vcd_nsxt_ip_set.set1", "ip_addresses.#", "5"),
					resource.TestMatchResourceAttr("vcd_nsxt_ip_set.set1", "owner_id", regexp.MustCompile(`^urn:vcloud:vdcGroup:.*$`)),
				),
			},
			{
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_ip_set.set1", "id", regexp.MustCompile(`^urn:vcloud:firewallGroup:.*$`)),
					resource.TestCheckResourceAttr("vcd_nsxt_ip_set.set1", "name", "test-ip-set-changed"),
					resource.TestCheckResourceAttr("vcd_nsxt_ip_set.set1", "description", ""),
					resource.TestCheckResourceAttr("vcd_nsxt_ip_set.set1", "ip_addresses.#", "0"),
					resource.TestMatchResourceAttr("vcd_nsxt_ip_set.set1", "owner_id", regexp.MustCompile(`^urn:vcloud:vdcGroup:.*$`)),
				),
			},
			// Test import with IP addresses
			{
				ResourceName:      "vcd_nsxt_ip_set.set1",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdNsxtEdgeGatewayObjectUsingVdcGroup(params["Name"].(string), params["NsxtEdgeGatewayVcd"].(string), "test-ip-set-changed"),
			},
		},
	})
	postTestChecks(t)
}

const testAccNsxtIpSetOwnByVdcGroupPrereqs = testAccVcdVdcGroupNew + `
data "vcd_external_network_v2" "existing-extnet" {
	name = "{{.ExternalNetwork}}"
}

resource "vcd_nsxt_edgegateway" "nsxt-edge" {
  org         = "{{.Org}}"
  owner_id    = vcd_vdc_group.test1.id
  name        = "{{.NsxtEdgeGatewayVcd}}"
  description = "Description"

  external_network_id = data.vcd_external_network_v2.existing-extnet.id

  subnet {
     gateway       = tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].gateway
     prefix_length = tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].prefix_length

     primary_ip = tolist(tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].static_ip_pool)[0].end_address
     allocated_ips {
       start_address = tolist(tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].static_ip_pool)[0].end_address
       end_address   = tolist(tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].static_ip_pool)[0].end_address
     }
  }
}


`
const testAccNsxtIpSetOwnByVdcGroup = testAccNsxtIpSetOwnByVdcGroupPrereqs + `
resource "vcd_nsxt_ip_set" "set1" {
  org = "{{.Org}}"

  edge_gateway_id = vcd_nsxt_edgegateway.nsxt-edge.id

  name = "test-ip-set"

  ip_addresses = [
    "12.12.12.1",
    "10.10.10.0/24",
    "11.11.11.1-11.11.11.2",
    "2001:db8::/48",
    "2001:db6:0:0:0:0:0:0-2001:db6:0:ffff:ffff:ffff:ffff:ffff",
  ]
}
`
const testAccNsxtIpSetOwnByVdcGroupUpdate = testAccNsxtIpSetOwnByVdcGroupPrereqs + `
resource "vcd_nsxt_ip_set" "set1" {
  org = "{{.Org}}"

  edge_gateway_id = vcd_nsxt_edgegateway.nsxt-edge.id

  name = "test-ip-set-changed"
}
`

// TestAccVcdNsxtIpSetMigration attempts to check migration path from legacy VDC
// configuration to new configuration which makes the NSX-T Edge Gateway follow membership of parent
// NSX-T Edge Gateway
// * Step 1 - creates prerequisites - VDC Group and 2 VDCs
// * Step 2 - creates an Edge Gateway and a IP Set attached to it
// * Step 3 - leaves the Edge Gateway as it is, but removed `vdc` field
// * Step 4 - migrates the Edge Gateway to VDC Group and observes that IP Set moves
// together and reflects it
func TestAccVcdNsxtIpSetMigration(t *testing.T) {
	preTestChecks(t)
	if !usingSysAdmin() {
		t.Skip(t.Name() + " requires system admin privileges to create VDCs")
		return
	}

	// String map to fill the template
	var params = StringMap{
		"Org":         testConfig.VCD.Org,
		"NsxtVdc":     testConfig.Nsxt.Vdc,
		"EdgeGw":      testConfig.Nsxt.EdgeGateway,
		"NetworkName": t.Name(),
		"Tags":        "network nsxt",

		"Name":                      t.Name(),
		"Dfw":                       "false",
		"DefaultPolicy":             "false",
		"ProviderVdc":               testConfig.VCD.NsxtProviderVdc.Name,
		"NetworkPool":               testConfig.VCD.NsxtProviderVdc.NetworkPool,
		"ProviderVdcStorageProfile": testConfig.VCD.ProviderVdc.StorageProfile,
		"ExternalNetwork":           testConfig.Nsxt.ExternalNetwork,
		"TestName":                  t.Name(),
		"NsxtEdgeGatewayVcd":        t.Name() + "-edge",
	}
	testParamsNotEmpty(t, params)

	params["FuncName"] = t.Name() + "-newVdc"
	configTextPre := templateFill(testAccVcdVdcGroupNew, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configTextPre)

	params["FuncName"] = t.Name() + "-step2"
	configText2 := templateFill(testAccVcdNsxtIpSetMigrationStep2, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText2)

	params["FuncName"] = t.Name() + "-step3"
	configText3 := templateFill(testAccVcdNsxtIpSetMigrationStep3, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 3: %s", configText3)

	params["FuncName"] = t.Name() + "-step4"
	configText4 := templateFill(testAccVcdNsxtIpSetMigrationStep4, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 4: %s", configText4)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckOpenApiVcdNetworkDestroy(testConfig.Nsxt.Vdc, t.Name()),
		Steps: []resource.TestStep{
			{ // step 1 - setup prerequisites
				Config: configTextPre,
			},
			{
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_nsxt_ip_set.set1", "id"),
					resource.TestMatchResourceAttr("vcd_nsxt_ip_set.set1", "id", regexp.MustCompile(`^urn:vcloud:firewallGroup:.*$`)),
					resource.TestCheckResourceAttr("vcd_nsxt_ip_set.set1", "name", "test-ip-set"),
					resource.TestCheckResourceAttr("vcd_nsxt_ip_set.set1", "description", ""),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ip_set.set1", "ip_addresses.*", "12.12.12.1"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ip_set.set1", "ip_addresses.*", "10.10.10.0/24"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ip_set.set1", "ip_addresses.*", "11.11.11.1-11.11.11.2"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ip_set.set1", "ip_addresses.*", "2001:db8::/48"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ip_set.set1", "ip_addresses.*", "2001:db6:0:0:0:0:0:0-2001:db6:0:ffff:ffff:ffff:ffff:ffff"),
					resource.TestCheckResourceAttr("vcd_nsxt_ip_set.set1", "ip_addresses.#", "5"),
					resource.TestMatchResourceAttr("vcd_nsxt_ip_set.set1", "owner_id", regexp.MustCompile(`^urn:vcloud:vdc:.*$`)),
					resource.TestCheckResourceAttrPair("vcd_nsxt_edgegateway.nsxt-edge", "owner_id", "vcd_nsxt_ip_set.set1", "owner_id"),
				),
			},
			{
				Config: configText3,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_nsxt_ip_set.set1", "id"),
					resource.TestMatchResourceAttr("vcd_nsxt_ip_set.set1", "id", regexp.MustCompile(`^urn:vcloud:firewallGroup:.*$`)),
					resource.TestCheckResourceAttr("vcd_nsxt_ip_set.set1", "name", "test-ip-set"),
					resource.TestCheckResourceAttr("vcd_nsxt_ip_set.set1", "description", ""),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ip_set.set1", "ip_addresses.*", "12.12.12.1"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ip_set.set1", "ip_addresses.*", "10.10.10.0/24"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ip_set.set1", "ip_addresses.*", "11.11.11.1-11.11.11.2"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ip_set.set1", "ip_addresses.*", "2001:db8::/48"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ip_set.set1", "ip_addresses.*", "2001:db6:0:0:0:0:0:0-2001:db6:0:ffff:ffff:ffff:ffff:ffff"),
					resource.TestCheckResourceAttr("vcd_nsxt_ip_set.set1", "ip_addresses.#", "5"),
					resource.TestMatchResourceAttr("vcd_nsxt_ip_set.set1", "owner_id", regexp.MustCompile(`^urn:vcloud:vdc:.*$`)),
					resource.TestCheckResourceAttrPair("vcd_nsxt_edgegateway.nsxt-edge", "owner_id", "vcd_nsxt_ip_set.set1", "owner_id"),
				),
			},
			{
				Config: configText4,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_nsxt_ip_set.set1", "id"),
					resource.TestMatchResourceAttr("vcd_nsxt_ip_set.set1", "id", regexp.MustCompile(`^urn:vcloud:firewallGroup:.*$`)),
					resource.TestCheckResourceAttr("vcd_nsxt_ip_set.set1", "name", "test-ip-set"),
					resource.TestCheckResourceAttr("vcd_nsxt_ip_set.set1", "description", ""),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ip_set.set1", "ip_addresses.*", "12.12.12.1"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ip_set.set1", "ip_addresses.*", "10.10.10.0/24"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ip_set.set1", "ip_addresses.*", "11.11.11.1-11.11.11.2"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ip_set.set1", "ip_addresses.*", "2001:db8::/48"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ip_set.set1", "ip_addresses.*", "2001:db6:0:0:0:0:0:0-2001:db6:0:ffff:ffff:ffff:ffff:ffff"),
					resource.TestCheckResourceAttr("vcd_nsxt_ip_set.set1", "ip_addresses.#", "5"),
				),
			},
			{ // Applying the same step once more to be sure that set1 has refreshed its fields after edge gateway was moved
				Config: configText4,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("vcd_nsxt_edgegateway.nsxt-edge", "owner_id", "vcd_nsxt_ip_set.set1", "owner_id"),
					resource.TestMatchResourceAttr("vcd_nsxt_ip_set.set1", "owner_id", regexp.MustCompile(`^urn:vcloud:vdcGroup:`)),
				),
			},

			// Check that import works
			{
				ResourceName:      "vcd_nsxt_ip_set.set1",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdNsxtEdgeGatewayObjectUsingVdcGroup(params["Name"].(string), params["NsxtEdgeGatewayVcd"].(string), "test-ip-set"),
				// field vdc during import isn't set
				ImportStateVerifyIgnore: []string{"vdc"},
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdNsxtIpSetMigrationStep2 = testAccVcdVdcGroupNew + `
data "vcd_external_network_v2" "existing-extnet" {
	name = "{{.ExternalNetwork}}"
}

resource "vcd_nsxt_edgegateway" "nsxt-edge" {
  org         = "{{.Org}}"
  owner_id    = vcd_org_vdc.newVdc.0.id
  name        = "{{.NsxtEdgeGatewayVcd}}"
  description = "Description"

  external_network_id = data.vcd_external_network_v2.existing-extnet.id

  subnet {
     gateway       = tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].gateway
     prefix_length = tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].prefix_length

     primary_ip = tolist(tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].static_ip_pool)[0].end_address
     allocated_ips {
       start_address = tolist(tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].static_ip_pool)[0].end_address
       end_address   = tolist(tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].static_ip_pool)[0].end_address
     }
  }
}

resource "vcd_nsxt_ip_set" "set1" {
  org  = "{{.Org}}"
  vdc  = vcd_org_vdc.newVdc.0.name

  edge_gateway_id = vcd_nsxt_edgegateway.nsxt-edge.id

  name = "test-ip-set"

  ip_addresses = [
    "12.12.12.1",
    "10.10.10.0/24",
    "11.11.11.1-11.11.11.2",
    "2001:db8::/48",
    "2001:db6:0:0:0:0:0:0-2001:db6:0:ffff:ffff:ffff:ffff:ffff",
  ]
}
`

const testAccVcdNsxtIpSetMigrationStep3 = testAccVcdVdcGroupNew + `
data "vcd_external_network_v2" "existing-extnet" {
	name = "{{.ExternalNetwork}}"
}

resource "vcd_nsxt_edgegateway" "nsxt-edge" {
  org         = "{{.Org}}"
  owner_id    = vcd_org_vdc.newVdc.0.id
  name        = "{{.NsxtEdgeGatewayVcd}}"
  description = "Description"

  external_network_id = data.vcd_external_network_v2.existing-extnet.id

  subnet {
     gateway       = tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].gateway
     prefix_length = tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].prefix_length

     primary_ip = tolist(tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].static_ip_pool)[0].end_address
     allocated_ips {
       start_address = tolist(tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].static_ip_pool)[0].end_address
       end_address   = tolist(tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].static_ip_pool)[0].end_address
     }
  }
}

resource "vcd_nsxt_ip_set" "set1" {
  org  = "{{.Org}}"

  edge_gateway_id = vcd_nsxt_edgegateway.nsxt-edge.id

  name = "test-ip-set"

  ip_addresses = [
    "12.12.12.1",
    "10.10.10.0/24",
    "11.11.11.1-11.11.11.2",
    "2001:db8::/48",
    "2001:db6:0:0:0:0:0:0-2001:db6:0:ffff:ffff:ffff:ffff:ffff",
  ]
}
`

const testAccVcdNsxtIpSetMigrationStep4 = testAccVcdVdcGroupNew + `
data "vcd_external_network_v2" "existing-extnet" {
	name = "{{.ExternalNetwork}}"
}

resource "vcd_nsxt_edgegateway" "nsxt-edge" {
  org         = "{{.Org}}"
  owner_id    = vcd_vdc_group.test1.id
  name        = "{{.NsxtEdgeGatewayVcd}}"
  description = "Description"

  external_network_id = data.vcd_external_network_v2.existing-extnet.id

  subnet {
     gateway       = tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].gateway
     prefix_length = tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].prefix_length

     primary_ip = tolist(tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].static_ip_pool)[0].end_address
     allocated_ips {
       start_address = tolist(tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].static_ip_pool)[0].end_address
       end_address   = tolist(tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].static_ip_pool)[0].end_address
     }
  }
}

resource "vcd_nsxt_ip_set" "set1" {
  org  = "{{.Org}}"

  edge_gateway_id = vcd_nsxt_edgegateway.nsxt-edge.id

  name = "test-ip-set"

  ip_addresses = [
    "12.12.12.1",
    "10.10.10.0/24",
    "11.11.11.1-11.11.11.2",
    "2001:db8::/48",
    "2001:db6:0:0:0:0:0:0-2001:db6:0:ffff:ffff:ffff:ffff:ffff",
  ]
}
`

// TestAccVcdNsxtIpSetInheritedVdc tests that NSX-T Edge Gateway IP Set can be created by
// using `vdc` field inherited from provider in NSX-T VDC
// * Step 1 - Rely on configuration coming from `provider` configuration for `vdc` value
// * Step 2 - Test that import works correctly
// * Step 3 - Test that data source works correctly
// * Step 4 - Start using `vdc` fields in resource and make sure it is not recreated
// * Step 5 - Test that import works correctly
// * Step 6 - Test data source
// Note. It does not test `org` field inheritance because our import sets it by default.
func TestAccVcdNsxtIpSetInheritedVdc(t *testing.T) {
	preTestChecks(t)
	if !usingSysAdmin() {
		t.Skip(t.Name() + " requires system admin privileges")
		return
	}

	// String map to fill the template
	var params = StringMap{
		"Org":                testConfig.VCD.Org,
		"NsxtVdc":            testConfig.Nsxt.Vdc,
		"IpSetName":          t.Name(),
		"NsxtEdgeGatewayVcd": "nsxt-edge-test",
		"ExternalNetwork":    testConfig.Nsxt.ExternalNetwork,

		// This particular field is consumed by `templateFill` to generate binary tests with correct
		// default VDC (NSX-T)
		"PrVdc": testConfig.Nsxt.Vdc,

		"Tags": "network",
	}
	testParamsNotEmpty(t, params)

	// This test explicitly tests that `vdc` field inherited from provider works correctly therefore
	// it must override default `vdc` field value at provider level to be NSX-T VDC and restore it
	// after this test.
	restoreDefaultVdcFunc := overrideDefaultVdcForTest(testConfig.Nsxt.Vdc)
	defer restoreDefaultVdcFunc()

	params["FuncName"] = t.Name() + "-step1"
	configText1 := templateFill(testAccVcdNsxtIpSetInheritedVdcStep1, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)

	params["FuncName"] = t.Name() + "-step3"
	configText3 := templateFill(testAccVcdNsxtIpSetInheritedVdcStep3, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 3: %s", configText1)

	params["FuncName"] = t.Name() + "-step4"
	configText4 := templateFill(testAccVcdNsxtIpSetInheritedVdcStep4, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 4: %s", configText4)

	params["FuncName"] = t.Name() + "-step6"
	configText6 := templateFill(testAccVcdNsxtIpSetInheritedVdcStep6, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 6: %s", configText6)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	cacheEdgeGatewaydId := &testCachedFieldValue{}
	cacheIpSetId := &testCachedFieldValue{}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckOpenApiVcdNetworkDestroy(testConfig.Nsxt.Vdc, t.Name()),
		Steps: []resource.TestStep{
			{
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					cacheEdgeGatewaydId.cacheTestResourceFieldValue("vcd_nsxt_edgegateway.nsxt-edge", "id"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_edgegateway.nsxt-edge", "id"),
					resource.TestMatchResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "owner_id", regexp.MustCompile(`^urn:vcloud:vdc:`)),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "vdc", testConfig.Nsxt.Vdc),

					cacheIpSetId.cacheTestResourceFieldValue("vcd_nsxt_ip_set.set1", "id"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_ip_set.set1", "id"),
					resource.TestMatchResourceAttr("vcd_nsxt_ip_set.set1", "id", regexp.MustCompile(`^urn:vcloud:firewallGroup:.*$`)),
					resource.TestCheckResourceAttr("vcd_nsxt_ip_set.set1", "name", params["IpSetName"].(string)),
					resource.TestCheckResourceAttr("vcd_nsxt_ip_set.set1", "description", ""),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ip_set.set1", "ip_addresses.*", "12.12.12.1"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ip_set.set1", "ip_addresses.*", "10.10.10.0/24"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ip_set.set1", "ip_addresses.*", "11.11.11.1-11.11.11.2"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ip_set.set1", "ip_addresses.*", "2001:db8::/48"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ip_set.set1", "ip_addresses.*", "2001:db6:0:0:0:0:0:0-2001:db6:0:ffff:ffff:ffff:ffff:ffff"),
					resource.TestCheckResourceAttr("vcd_nsxt_ip_set.set1", "ip_addresses.#", "5"),
				),
			},
			{
				ResourceName:      "vcd_nsxt_ip_set.set1",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdNsxtEdgeGatewayObject(testConfig, params["NsxtEdgeGatewayVcd"].(string), params["IpSetName"].(string)),
			},
			{
				Config: configText3,
				Check: resource.ComposeAggregateTestCheckFunc(
					cacheEdgeGatewaydId.testCheckCachedResourceFieldValue("vcd_nsxt_edgegateway.nsxt-edge", "id"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_edgegateway.nsxt-edge", "id"),
					resource.TestMatchResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "owner_id", regexp.MustCompile(`^urn:vcloud:vdc:`)),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "vdc", testConfig.Nsxt.Vdc),
					resourceFieldsEqual("data.vcd_nsxt_edgegateway.nsxt-edge", "vcd_nsxt_edgegateway.nsxt-edge", []string{"%"}),

					cacheIpSetId.testCheckCachedResourceFieldValue("vcd_nsxt_ip_set.set1", "id"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_ip_set.set1", "id"),
					resource.TestMatchResourceAttr("vcd_nsxt_ip_set.set1", "id", regexp.MustCompile(`^urn:vcloud:firewallGroup:.*$`)),
					resource.TestCheckResourceAttr("vcd_nsxt_ip_set.set1", "name", params["IpSetName"].(string)),
					resource.TestCheckResourceAttr("vcd_nsxt_ip_set.set1", "description", ""),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ip_set.set1", "ip_addresses.*", "12.12.12.1"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ip_set.set1", "ip_addresses.*", "10.10.10.0/24"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ip_set.set1", "ip_addresses.*", "11.11.11.1-11.11.11.2"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ip_set.set1", "ip_addresses.*", "2001:db8::/48"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ip_set.set1", "ip_addresses.*", "2001:db6:0:0:0:0:0:0-2001:db6:0:ffff:ffff:ffff:ffff:ffff"),
					resource.TestCheckResourceAttr("vcd_nsxt_ip_set.set1", "ip_addresses.#", "5"),
					resourceFieldsEqual("data.vcd_nsxt_ip_set.set1", "vcd_nsxt_ip_set.set1", []string{"%"}),
				),
			},
			{
				Config: configText4,
				Check: resource.ComposeAggregateTestCheckFunc(
					cacheEdgeGatewaydId.testCheckCachedResourceFieldValue("vcd_nsxt_edgegateway.nsxt-edge", "id"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_edgegateway.nsxt-edge", "id"),
					resource.TestMatchResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "owner_id", regexp.MustCompile(`^urn:vcloud:vdc:`)),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "vdc", testConfig.Nsxt.Vdc),

					resource.TestCheckResourceAttrSet("vcd_nsxt_ip_set.set1", "id"),
					resource.TestMatchResourceAttr("vcd_nsxt_ip_set.set1", "id", regexp.MustCompile(`^urn:vcloud:firewallGroup:.*$`)),
					resource.TestCheckResourceAttr("vcd_nsxt_ip_set.set1", "name", params["IpSetName"].(string)),
					resource.TestCheckResourceAttr("vcd_nsxt_ip_set.set1", "description", ""),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ip_set.set1", "ip_addresses.*", "12.12.12.1"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ip_set.set1", "ip_addresses.*", "10.10.10.0/24"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ip_set.set1", "ip_addresses.*", "11.11.11.1-11.11.11.2"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ip_set.set1", "ip_addresses.*", "2001:db8::/48"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ip_set.set1", "ip_addresses.*", "2001:db6:0:0:0:0:0:0-2001:db6:0:ffff:ffff:ffff:ffff:ffff"),
					resource.TestCheckResourceAttr("vcd_nsxt_ip_set.set1", "ip_addresses.#", "5"),
					resource.TestCheckResourceAttr("vcd_nsxt_ip_set.set1", "vdc", testConfig.Nsxt.Vdc),
				),
			},
			{
				ResourceName:      "vcd_nsxt_ip_set.set1",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdNsxtEdgeGatewayObject(testConfig, params["NsxtEdgeGatewayVcd"].(string), params["IpSetName"].(string)),
				// field vdc during import isn't set
				ImportStateVerifyIgnore: []string{"vdc"},
			},
			{
				Config: configText6,
				Check: resource.ComposeAggregateTestCheckFunc(
					cacheEdgeGatewaydId.testCheckCachedResourceFieldValue("vcd_nsxt_edgegateway.nsxt-edge", "id"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_edgegateway.nsxt-edge", "id"),
					resource.TestMatchResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "owner_id", regexp.MustCompile(`^urn:vcloud:vdc:`)),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "vdc", testConfig.Nsxt.Vdc),
					resourceFieldsEqual("data.vcd_nsxt_edgegateway.nsxt-edge", "vcd_nsxt_edgegateway.nsxt-edge", []string{"%"}),

					cacheIpSetId.testCheckCachedResourceFieldValue("vcd_nsxt_ip_set.set1", "id"),
					resource.TestCheckResourceAttrSet("vcd_nsxt_ip_set.set1", "id"),
					resource.TestMatchResourceAttr("vcd_nsxt_ip_set.set1", "id", regexp.MustCompile(`^urn:vcloud:firewallGroup:.*$`)),
					resource.TestCheckResourceAttr("vcd_nsxt_ip_set.set1", "name", params["IpSetName"].(string)),
					resource.TestCheckResourceAttr("vcd_nsxt_ip_set.set1", "description", ""),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ip_set.set1", "ip_addresses.*", "12.12.12.1"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ip_set.set1", "ip_addresses.*", "10.10.10.0/24"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ip_set.set1", "ip_addresses.*", "11.11.11.1-11.11.11.2"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ip_set.set1", "ip_addresses.*", "2001:db8::/48"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_ip_set.set1", "ip_addresses.*", "2001:db6:0:0:0:0:0:0-2001:db6:0:ffff:ffff:ffff:ffff:ffff"),
					resource.TestCheckResourceAttr("vcd_nsxt_ip_set.set1", "ip_addresses.#", "5"),
					resource.TestCheckResourceAttr("vcd_nsxt_ip_set.set1", "vdc", testConfig.Nsxt.Vdc),
					resourceFieldsEqual("data.vcd_nsxt_ip_set.set1", "vcd_nsxt_ip_set.set1", []string{"%"}),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdNsxtIpSetInheritedVdcStep1 = `
data "vcd_external_network_v2" "existing-extnet" {
	name = "{{.ExternalNetwork}}"
}

resource "vcd_nsxt_edgegateway" "nsxt-edge" {
  org  = "{{.Org}}"
  name = "{{.NsxtEdgeGatewayVcd}}"

  external_network_id = data.vcd_external_network_v2.existing-extnet.id

  subnet {
     gateway               = tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].gateway
     prefix_length         = tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].prefix_length

     primary_ip            = tolist(tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].static_ip_pool)[0].end_address
     allocated_ips {
       start_address = tolist(tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].static_ip_pool)[0].end_address
       end_address   = tolist(tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].static_ip_pool)[0].end_address
     }
  }
}

resource "vcd_nsxt_ip_set" "set1" {
  org  = "{{.Org}}"

  edge_gateway_id = vcd_nsxt_edgegateway.nsxt-edge.id

  name = "{{.IpSetName}}"

  ip_addresses = [
    "12.12.12.1",
    "10.10.10.0/24",
    "11.11.11.1-11.11.11.2",
    "2001:db8::/48",
    "2001:db6:0:0:0:0:0:0-2001:db6:0:ffff:ffff:ffff:ffff:ffff",
  ]
}
`

const testAccVcdNsxtIpSetInheritedVdcStep3 = testAccVcdNsxtIpSetInheritedVdcStep1 + `
# skip-binary-test: Data Source test
data "vcd_nsxt_edgegateway" "nsxt-edge" {
  org  = "{{.Org}}"
  name = "{{.NsxtEdgeGatewayVcd}}"
}

data "vcd_nsxt_ip_set" "set1" {
  org  = "{{.Org}}"
  name = "{{.IpSetName}}"

  edge_gateway_id = data.vcd_nsxt_edgegateway.nsxt-edge.id
}
`

const testAccVcdNsxtIpSetInheritedVdcStep4 = `
data "vcd_external_network_v2" "existing-extnet" {
	name = "{{.ExternalNetwork}}"
}

resource "vcd_nsxt_edgegateway" "nsxt-edge" {
  org  = "{{.Org}}"
  vdc  = "{{.NsxtVdc}}"
  name = "{{.NsxtEdgeGatewayVcd}}"

  external_network_id = data.vcd_external_network_v2.existing-extnet.id

  subnet {
     gateway               = tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].gateway
     prefix_length         = tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].prefix_length

     primary_ip            = tolist(tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].static_ip_pool)[0].end_address
     allocated_ips {
       start_address = tolist(tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].static_ip_pool)[0].end_address
       end_address   = tolist(tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].static_ip_pool)[0].end_address
     }
  }
}

resource "vcd_nsxt_ip_set" "set1" {
  org  = "{{.Org}}"
  vdc  = "{{.NsxtVdc}}"
  name = "{{.IpSetName}}"

  edge_gateway_id = vcd_nsxt_edgegateway.nsxt-edge.id

  ip_addresses = [
    "12.12.12.1",
    "10.10.10.0/24",
    "11.11.11.1-11.11.11.2",
    "2001:db8::/48",
    "2001:db6:0:0:0:0:0:0-2001:db6:0:ffff:ffff:ffff:ffff:ffff",
  ]
}
`

const testAccVcdNsxtIpSetInheritedVdcStep6 = testAccVcdNsxtIpSetInheritedVdcStep4 + `
# skip-binary-test: Data Source test
data "vcd_nsxt_edgegateway" "nsxt-edge" {
  org  = "{{.Org}}"
  vdc  = "{{.NsxtVdc}}"
  name = "{{.NsxtEdgeGatewayVcd}}"
}

data "vcd_nsxt_ip_set" "set1" {
  org  = "{{.Org}}"
  vdc  = "{{.NsxtVdc}}"
  name = "{{.IpSetName}}"

  edge_gateway_id = data.vcd_nsxt_edgegateway.nsxt-edge.id
}
`
