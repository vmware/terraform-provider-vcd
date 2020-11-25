// +build gateway nsxt ALL functional

package vcd

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/vmware/go-vcloud-director/v2/govcd"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcdNsxtEdgeGateway(t *testing.T) {

	if !usingSysAdmin() {
		t.Skip(t.Name() + " requires system admin privileges")
		return
	}

	skipNoNsxtConfiguration(t)
	vcdClient := createTemporaryVCDConnection()
	if vcdClient.Client.APIVCDMaxVersionIs("< 34.0") {
		t.Skip(t.Name() + " requires at least API v34.0 (vCD 10.1+)")
	}

	nsxtExtNet, err := govcd.GetExternalNetworkV2ByName(vcdClient.VCDClient, testConfig.Nsxt.ExternalNetwork)
	if err != nil {
		t.Skipf("%s - could not retrieve external network", t.Name())
	}

	// String map to fill the template
	var params = StringMap{
		"Org":                testConfig.VCD.Org,
		"NsxtVdc":            testConfig.Nsxt.Vdc,
		"NsxtEdgeGatewayVcd": "nsxt-edge",
		"ExternalNetwork":    testConfig.Networking.ExternalNetwork,
		"Tags":               "gateway nsxt",
	}
	configText := templateFill(testAccNsxtEdgeGateway, params)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	params["FuncName"] = t.Name() + "step1"
	configText1 := templateFill(testAccNsxtEdgeGatewayUpdate, params)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	if !usingSysAdmin() {
		t.Skip("Edge Gateway tests require system admin privileges")
		return
	}
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVcdNsxtEdgeGatewayDestroy(params["NsxtEdgeGatewayVcd"].(string)),
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "name", "nsxt-edge"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_edgegateway.nsxt-edge", "subnet.*", map[string]string{
						"enabled":       "true",
						"gateway":       nsxtExtNet.ExternalNetwork.Subnets.Values[0].Gateway,
						"prefix_length": strconv.Itoa(nsxtExtNet.ExternalNetwork.Subnets.Values[0].PrefixLength),
						"primary_ip":    nsxtExtNet.ExternalNetwork.Subnets.Values[0].IPRanges.Values[0].StartAddress,
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_edgegateway.nsxt-edge", "subnet.*.allocated_ips.*", map[string]string{
						"start_address": nsxtExtNet.ExternalNetwork.Subnets.Values[0].IPRanges.Values[0].StartAddress,
						"end_address":   nsxtExtNet.ExternalNetwork.Subnets.Values[0].IPRanges.Values[0].StartAddress,
					}),
				),
			},
			resource.TestStep{
				Config: configText1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "name", "nsxt-edge"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "description", "Updated-Description"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_edgegateway.nsxt-edge", "subnet.*", map[string]string{
						"enabled":       "true",
						"gateway":       nsxtExtNet.ExternalNetwork.Subnets.Values[0].Gateway,
						"prefix_length": strconv.Itoa(nsxtExtNet.ExternalNetwork.Subnets.Values[0].PrefixLength),
						"primary_ip":    nsxtExtNet.ExternalNetwork.Subnets.Values[0].IPRanges.Values[0].EndAddress,
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_edgegateway.nsxt-edge", "subnet.*.allocated_ips.*", map[string]string{
						"start_address": nsxtExtNet.ExternalNetwork.Subnets.Values[0].IPRanges.Values[0].StartAddress,
						"end_address":   nsxtExtNet.ExternalNetwork.Subnets.Values[0].IPRanges.Values[0].EndAddress,
					}),
				),
			},
			resource.TestStep{
				ResourceName:      "vcd_nsxt_edgegateway.nsxt-edge",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdOrgNsxtVdcObject(testConfig, params["NsxtEdgeGatewayVcd"].(string)),
			},
		},
	})
}

func testAccCheckVcdNsxtEdgeGatewayDestroy(edgeName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		for _, rs := range s.RootModule().Resources {
			edgeGatewayName := rs.Primary.Attributes["name"]
			if rs.Type != "vcd_edgegateway" {
				continue
			}
			if edgeGatewayName != edgeName {
				continue
			}
			conn := testAccProvider.Meta().(*VCDClient)
			orgName := rs.Primary.Attributes["org"]
			vdcName := rs.Primary.Attributes["vdc"]

			org, _, err := conn.GetOrgAndVdc(orgName, vdcName)
			if err != nil {
				return fmt.Errorf("error retrieving org %s and vdc %s : %s ", orgName, vdcName, err)
			}

			_, err = org.GetNsxtEdgeGatewayByName(edgeName)
			if err == nil {
				return fmt.Errorf("NSX-T edge gateway %s was not removed", edgeName)
			}
		}

		return nil
	}
}

const testAccNsxtEdgeGatewayDataSources = `
#data "vcd_nsxt_edge_cluster" "ec" {
#	vdc  = "{{.NsxtVdc}}"
#	name = "{{.ExistingEdgeCluster}}"
#}

data "vcd_external_network_v2" "existing-extnet" {
	name = "nsxt-extnet-dainius"
}

data "vcd_nsxt_manager" "main" {
  name = "nsxManager1"
}
`

const testAccNsxtEdgeGateway = testAccNsxtEdgeGatewayDataSources + `
resource "vcd_nsxt_edgegateway" "nsxt-edge" {
  org                     = "{{.Org}}"
  vdc                     = "{{.NsxtVdc}}"
  name                    = "{{.NsxtEdgeGatewayVcd}}"
  description             = "Description"
#  edge_cluster_id         = data.vcd_nsxt_edge_cluster.ec.id

  #nsxt_manager_id     = data.vcd_nsxt_manager.main.id
  external_network_id = data.vcd_external_network_v2.existing-extnet.id

  subnet {
     gateway               = tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].gateway
     prefix_length         = tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].prefix_length

     primary_ip            = tolist(tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].static_ip_pool)[0].start_address
     allocated_ips {
       start_address = tolist(tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].static_ip_pool)[0].start_address
       end_address   = tolist(tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].static_ip_pool)[0].start_address
     }
  }
}
`

const testAccNsxtEdgeGatewayUpdate = testAccNsxtEdgeGatewayDataSources + `
resource "vcd_nsxt_edgegateway" "nsxt-edge" {
  org                     = "{{.Org}}"
  vdc                     = "{{.NsxtVdc}}"
  name                    = "{{.NsxtEdgeGatewayVcd}}"
  description             = "Updated-Description"
#  edge_cluster_id         = data.vcd_nsxt_edge_cluster.ec.id

  #nsxt_manager_id     = data.vcd_nsxt_manager.main.id
  external_network_id = data.vcd_external_network_v2.existing-extnet.id

  subnet {
     gateway               = tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].gateway
     prefix_length         = tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].prefix_length
     primary_ip            = tolist(tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].static_ip_pool)[0].end_address
     allocated_ips {
       start_address = tolist(tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].static_ip_pool)[0].start_address
       end_address   = tolist(tolist(data.vcd_external_network_v2.existing-extnet.ip_scope)[0].static_ip_pool)[0].end_address
     }
  }
}
`
