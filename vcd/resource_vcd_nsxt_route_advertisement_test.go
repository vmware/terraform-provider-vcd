//go:build network || nsxt || ALL || functional
// +build network nsxt ALL functional

package vcd

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"regexp"
	"strconv"
	"testing"
)

func TestAccVcdNsxtRouteAdvertisement(t *testing.T) {
	preTestChecks(t)
	skipNoNsxtConfiguration(t)

	isRouteAdvertisementEnable := true
	subnet1 := "192.168.1.0/24"
	subnet2 := "192.168.2.0/24"

	// String map to fill the template
	var params = StringMap{
		"Name":              t.Name(),
		"Org":               testConfig.VCD.Org,
		"NsxtVdc":           testConfig.Nsxt.Vdc,
		"NsxtVdcGroup":      testConfig.Nsxt.VdcGroup,
		"EdgeGw":            testConfig.Nsxt.EdgeGateway,
		"EdgeGwVdcGroup":    testConfig.Nsxt.VdcGroupEdgeGateway,
		"Enabled":           strconv.FormatBool(isRouteAdvertisementEnable),
		"Subnet1Cidr":       subnet1,
		"Subnet2Cidr":       subnet2,
		"NsxtImportSegment": testConfig.Nsxt.NsxtImportSegment,
	}

	configText1 := templateFill(testAccNsxtRouteAdvertisementCreation, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)

	params["FuncName"] = t.Name() + "-step2"
	configText2 := templateFill(testAccNsxtRouteAdvertisementUpdate, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText2)

	params["FuncName"] = t.Name() + "-step3"
	configText3 := templateFill(testAccNsxtRouteAdvertisementDisabled, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 3: %s", configText3)

	params["FuncName"] = t.Name() + "-step4"
	configText4 := templateFill(testAccNsxtRouteAdvertisementCreationVDCGroup, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 4: %s", configText4)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	// Ensure Edge Gateway has a dedicated Tier 0 gateway (External network) as BGP and Route
	// Advertisement configuration requires it. Restore it right after the test so that other
	// tests are not impacted.
	updateEdgeGatewayTier0Dedication(t, true)
	defer updateEdgeGatewayTier0Dedication(t, false)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy:      testAccCheckNsxtRouteAdvertisement(testConfig.Nsxt.Vdc, testConfig.Nsxt.EdgeGateway),
		Steps: []resource.TestStep{
			{
				Config: configText4,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_route_advertisement.testing", "id", regexp.MustCompile(`^urn:vcloud:gateway:.*$`)),
					resource.TestCheckResourceAttr("vcd_nsxt_route_advertisement.testing", "enabled", strconv.FormatBool(isRouteAdvertisementEnable)),
					resource.TestCheckResourceAttr("vcd_nsxt_route_advertisement.testing", "subnets.#", "1"),
					resource.TestCheckResourceAttr("vcd_nsxt_route_advertisement.testing", "subnets.0", subnet1),
				),
			},
			{
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_route_advertisement.testing", "id", regexp.MustCompile(`^urn:vcloud:gateway:.*$`)),
					resource.TestCheckResourceAttr("vcd_nsxt_route_advertisement.testing", "enabled", strconv.FormatBool(isRouteAdvertisementEnable)),
					resource.TestCheckResourceAttr("vcd_nsxt_route_advertisement.testing", "subnets.#", "1"),
					resource.TestCheckResourceAttr("vcd_nsxt_route_advertisement.testing", "subnets.0", subnet1),
				),
			},
			{
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_route_advertisement.testing", "id", regexp.MustCompile(`^urn:vcloud:gateway:.*$`)),
					resource.TestCheckResourceAttr("vcd_nsxt_route_advertisement.testing", "enabled", strconv.FormatBool(isRouteAdvertisementEnable)),
					resource.TestCheckResourceAttr("vcd_nsxt_route_advertisement.testing", "subnets.#", "2"),
					resource.TestMatchResourceAttr("vcd_nsxt_route_advertisement.testing", "subnets.0", regexp.MustCompile(`^192.168.[1-2].0/24$`)),
					resource.TestMatchResourceAttr("vcd_nsxt_route_advertisement.testing", "subnets.1", regexp.MustCompile(`^192.168.[1-2].0/24$`)),
				),
			},
			{
				Config: configText3,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_nsxt_route_advertisement.testing", "id", regexp.MustCompile(`^urn:vcloud:gateway:.*$`)),
					resource.TestCheckResourceAttr("vcd_nsxt_route_advertisement.testing", "enabled", strconv.FormatBool(false)),
					resource.TestCheckResourceAttr("vcd_nsxt_route_advertisement.testing", "subnets.#", "0"),
				),
			},
			{
				ResourceName:            "vcd_nsxt_route_advertisement.testing",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdFunc:       importStateIdOrgNsxtVdcObject(testConfig, testConfig.Nsxt.EdgeGateway),
				ImportStateVerifyIgnore: []string{"vdc"},
			},
		},
	})
}

const testAccNsxtRouteAdvertisementCreation = `
data "vcd_org_vdc" "{{.NsxtVdc}}" {
  org = "{{.Org}}"
  name = "{{.NsxtVdc}}"
}

data "vcd_nsxt_edgegateway" "{{.EdgeGw}}" {
  owner_id = data.vcd_org_vdc.{{.NsxtVdc}}.id
  name     = "{{.EdgeGw}}"
}

resource "vcd_nsxt_route_advertisement" "testing" {
  org = "{{.Org}}"
  edge_gateway_id = data.vcd_nsxt_edgegateway.{{.EdgeGw}}.id
  enabled = {{.Enabled}}
  subnets = ["{{.Subnet1Cidr}}"]
}
`

const testAccNsxtRouteAdvertisementUpdate = `
data "vcd_org_vdc" "{{.NsxtVdc}}" {
  org = "{{.Org}}"
  name = "{{.NsxtVdc}}"
}

data "vcd_nsxt_edgegateway" "{{.EdgeGw}}" {
  owner_id = data.vcd_org_vdc.{{.NsxtVdc}}.id
  name     = "{{.EdgeGw}}"
}

resource "vcd_nsxt_route_advertisement" "testing" {
  org = "{{.Org}}"
  edge_gateway_id = data.vcd_nsxt_edgegateway.{{.EdgeGw}}.id
  enabled = {{.Enabled}}
  subnets = ["{{.Subnet1Cidr}}", "{{.Subnet2Cidr}}"]
}
`

const testAccNsxtRouteAdvertisementDisabled = `
data "vcd_org_vdc" "{{.NsxtVdc}}" {
  org = "{{.Org}}"
  name = "{{.NsxtVdc}}"
}

data "vcd_nsxt_edgegateway" "{{.EdgeGw}}" {
  owner_id = data.vcd_org_vdc.{{.NsxtVdc}}.id
  name     = "{{.EdgeGw}}"
}

resource "vcd_nsxt_route_advertisement" "testing" {
  org = "{{.Org}}"
  edge_gateway_id = data.vcd_nsxt_edgegateway.{{.EdgeGw}}.id
  enabled = false
}
`

const testAccNsxtRouteAdvertisementCreationVDCGroup = `
data "vcd_vdc_group" "{{.NsxtVdcGroup}}" {
  org = "{{.Org}}"
  name = "{{.NsxtVdcGroup}}"
}

data "vcd_nsxt_edgegateway" "{{.EdgeGwVdcGroup}}" {
  owner_id = data.vcd_vdc_group.{{.NsxtVdcGroup}}.id
  name     = "{{.EdgeGwVdcGroup}}"
}

resource "vcd_nsxt_route_advertisement" "testing" {
  org = "{{.Org}}"
  edge_gateway_id = data.vcd_nsxt_edgegateway.{{.EdgeGwVdcGroup}}.id
  enabled = {{.Enabled}}
  subnets = ["{{.Subnet1Cidr}}"]
}

resource "vcd_nsxt_firewall" "testing" {
  org = "{{.Org}}"

  edge_gateway_id = data.vcd_nsxt_edgegateway.{{.EdgeGwVdcGroup}}.id

  rule {
    action      = "ALLOW"
    name        = "allow all IPv4 traffic"
    direction   = "IN_OUT"
    ip_protocol = "IPV4"
  }
}

resource "vcd_nsxt_nat_rule" "snat" {
  org = "{{.Org}}"

  edge_gateway_id = data.vcd_nsxt_edgegateway.{{.EdgeGwVdcGroup}}.id

  name        = "SNAT rule"
  rule_type   = "SNAT"
  description = "description"

  # Using primary_ip from edge gateway
  external_address         = tolist(data.vcd_nsxt_edgegateway.{{.EdgeGwVdcGroup}}.subnet)[0].primary_ip
  internal_address         = "11.11.11.0/24"
  snat_destination_address = "8.8.8.8"
  logging                  = true
}

resource "vcd_nsxt_nat_rule" "no-snat" {
  org = "{{.Org}}"

  edge_gateway_id = data.vcd_nsxt_edgegateway.{{.EdgeGwVdcGroup}}.id

  name        = "test-no-snat-rule"
  rule_type   = "NO_SNAT"
  description = "description"

  internal_address = "11.11.11.0/24"
}

resource "vcd_nsxt_ipsec_vpn_tunnel" "tunnel1" {
  org = "{{.Org}}"

  edge_gateway_id = data.vcd_nsxt_edgegateway.{{.EdgeGwVdcGroup}}.id

  name        = "First"
  description = "testing tunnel"

  pre_shared_key = "my-presharaed-key"
  # Primary IP address of Edge Gateway pulled from data source
  local_ip_address = tolist(data.vcd_nsxt_edgegateway.{{.EdgeGwVdcGroup}}.subnet)[0].primary_ip
  local_networks   = ["10.10.10.0/24", "30.30.30.0/28", "40.40.40.1/32"]
  # That is a fake remote IP address
  remote_ip_address = "1.2.3.4"
  remote_networks   = ["192.168.1.0/24", "192.168.10.0/24", "192.168.20.0/28"]
}

resource "vcd_network_routed_v2" "nsxt-backed" {
  org             = "{{.Org}}"
  name            = "{{.Name}}-routed"
  edge_gateway_id = data.vcd_nsxt_edgegateway.{{.EdgeGwVdcGroup}}.id

  gateway       = "1.1.1.1"
  prefix_length = 24

  static_ip_pool {
    start_address = "1.1.1.10"
    end_address   = "1.1.1.20"
  }
}

resource "vcd_nsxt_network_dhcp" "pools" {
  org = "{{.Org}}"

  org_network_id = vcd_network_routed_v2.nsxt-backed.id

  pool {
    start_address = "1.1.1.100"
    end_address   = "1.1.1.110"
  }

  pool {
    start_address = "1.1.1.111"
    end_address   = "1.1.1.112"
  }
}

resource "vcd_network_isolated_v2" "nsxt-backed" {
  org      = "{{.Org}}"
  owner_id = data.vcd_vdc_group.{{.NsxtVdcGroup}}.id

  name = "{{.Name}}-isolated"

  gateway       = "2.1.1.1"
  prefix_length = 24

  static_ip_pool {
    start_address = "2.1.1.10"
    end_address   = "2.1.1.20"
  }
}

resource "vcd_nsxt_network_imported" "nsxt-backed" {
  org      = "{{.Org}}"
  owner_id = data.vcd_vdc_group.{{.NsxtVdcGroup}}.id

  name = "{{.Name}}-imported"

  nsxt_logical_switch_name = "{{.NsxtImportSegment}}"

  gateway       = "4.1.1.1"
  prefix_length = 24

  static_ip_pool {
    start_address = "4.1.1.10"
    end_address   = "4.1.1.20"
  }
}
`

func testAccCheckNsxtRouteAdvertisement(vdcName, edgeGatewayName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*VCDClient)

		_, vdc, err := conn.GetOrgAndVdc(testConfig.VCD.Org, vdcName)
		if err != nil {
			return fmt.Errorf(errorRetrievingVdcFromOrg, vdcName, testConfig.VCD.Org, err)
		}

		edge, err := vdc.GetNsxtEdgeGatewayByName(edgeGatewayName)
		if err != nil {
			return fmt.Errorf(errorUnableToFindEdgeGateway, edgeGatewayName)
		}

		routeAdvertisement, err := edge.GetNsxtRouteAdvertisement(true)
		if err != nil {
			return fmt.Errorf("error trying to retrieve route advertisement - %s", err)
		}

		if routeAdvertisement.Enable {
			return fmt.Errorf("error destroying route advertisement. Wanted routeAdvertisement.Enable false, Got %t", routeAdvertisement.Enable)
		}

		if routeAdvertisement.Subnets != nil && len(routeAdvertisement.Subnets) > 0 {
			return fmt.Errorf("error destroying route advertisement. Wanted 0 routeAdvertisement.Subnets, got %d", len(routeAdvertisement.Subnets))
		}

		return nil
	}
}

func updateEdgeGatewayTier0Dedication(t *testing.T, dedicatedTier0 bool) {
	vcdClient := createSystemTemporaryVCDConnection()
	org, err := vcdClient.GetOrgByName(testConfig.VCD.Org)
	if err != nil {
		t.Fatalf("error retrieving Org '%s': %s", testConfig.VCD.Org, err)
	}
	edge, err := org.GetNsxtEdgeGatewayByName(testConfig.Nsxt.EdgeGateway)
	if err != nil {
		t.Fatalf("error retrieving NSX-T Edge Gateway '%s': %s", testConfig.Nsxt.EdgeGateway, err)
	}

	edge.EdgeGateway.EdgeGatewayUplinks[0].Dedicated = dedicatedTier0
	_, err = edge.Update(edge.EdgeGateway)
	if err != nil {
		t.Fatalf("error updating NSX-T Edge Gateway dedicated Tier 0 gateway usage to '%t': %s", dedicatedTier0, err)
	}
}
