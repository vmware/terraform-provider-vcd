//go:build gateway || nsxt || ALL || functional
// +build gateway nsxt ALL functional

package vcd

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/vmware/go-vcloud-director/v2/govcd"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// TestAccVcdNsxtEdgeGateway tests out creating and updating edge gateway using existing external network
// testConfig.Nsxt.ExternalNetwork which is expected to be correctly configured.
func TestAccVcdNsxtEdgeGateway(t *testing.T) {
	preTestChecks(t)
	if !usingSysAdmin() {
		t.Skip(t.Name() + " requires system admin privileges")
		return
	}

	skipNoNsxtConfiguration(t)
	vcdClient := createTemporaryVCDConnection(false)

	nsxtExtNet, err := govcd.GetExternalNetworkV2ByName(vcdClient.VCDClient, testConfig.Nsxt.ExternalNetwork)
	if err != nil {
		t.Skipf("%s - could not retrieve external network", t.Name())
	}

	var params = StringMap{
		"Org":                testConfig.VCD.Org,
		"NsxtVdc":            testConfig.Nsxt.Vdc,
		"NsxtEdgeGatewayVcd": "nsxt-edge-test",
		"ExternalNetwork":    testConfig.Nsxt.ExternalNetwork,
		"EdgeClusterId":      lookupAvailableEdgeClusterId(t, vcdClient),
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
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "name", params["NsxtEdgeGatewayVcd"].(string)),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "dedicate_external_network", "false"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "primary_ip", nsxtExtNet.ExternalNetwork.Subnets.Values[0].IPRanges.Values[0].EndAddress),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_edgegateway.nsxt-edge", "subnet.*", map[string]string{
						"gateway":       nsxtExtNet.ExternalNetwork.Subnets.Values[0].Gateway,
						"prefix_length": strconv.Itoa(nsxtExtNet.ExternalNetwork.Subnets.Values[0].PrefixLength),
						"primary_ip":    nsxtExtNet.ExternalNetwork.Subnets.Values[0].IPRanges.Values[0].EndAddress,
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_edgegateway.nsxt-edge", "subnet.*.allocated_ips.*", map[string]string{
						"start_address": nsxtExtNet.ExternalNetwork.Subnets.Values[0].IPRanges.Values[0].EndAddress,
						"end_address":   nsxtExtNet.ExternalNetwork.Subnets.Values[0].IPRanges.Values[0].EndAddress,
					}),
				),
			},
			resource.TestStep{
				Config: configText1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "name", params["NsxtEdgeGatewayVcd"].(string)),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "dedicate_external_network", "false"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "primary_ip", nsxtExtNet.ExternalNetwork.Subnets.Values[0].IPRanges.Values[0].EndAddress),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "description", "Updated-Description"),
					resource.TestCheckResourceAttr("vcd_nsxt_edgegateway.nsxt-edge", "edge_cluster_id", params["EdgeClusterId"].(string)),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_edgegateway.nsxt-edge", "subnet.*", map[string]string{
						"gateway":       nsxtExtNet.ExternalNetwork.Subnets.Values[0].Gateway,
						"prefix_length": strconv.Itoa(nsxtExtNet.ExternalNetwork.Subnets.Values[0].PrefixLength),
						"primary_ip":    nsxtExtNet.ExternalNetwork.Subnets.Values[0].IPRanges.Values[0].EndAddress,
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_edgegateway.nsxt-edge", "subnet.*.allocated_ips.*", map[string]string{
						"start_address": nsxtExtNet.ExternalNetwork.Subnets.Values[0].IPRanges.Values[0].EndAddress,
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
	postTestChecks(t)
}

const testAccNsxtEdgeGatewayDataSources = `
data "vcd_external_network_v2" "existing-extnet" {
	name = "{{.ExternalNetwork}}"
}
`

const testAccNsxtEdgeGateway = testAccNsxtEdgeGatewayDataSources + `
resource "vcd_nsxt_edgegateway" "nsxt-edge" {
  org                     = "{{.Org}}"
  vdc                     = "{{.NsxtVdc}}"
  name                    = "{{.NsxtEdgeGatewayVcd}}"
  description             = "Description"

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
`

const testAccNsxtEdgeGatewayUpdate = testAccNsxtEdgeGatewayDataSources + `
resource "vcd_nsxt_edgegateway" "nsxt-edge" {
  org                     = "{{.Org}}"
  vdc                     = "{{.NsxtVdc}}"
  name                    = "{{.NsxtEdgeGatewayVcd}}"
  description             = "Updated-Description"
  edge_cluster_id         = "{{.EdgeClusterId}}"

  external_network_id = data.vcd_external_network_v2.existing-extnet.id
  dedicate_external_network = false

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
`

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

func lookupAvailableEdgeClusterId(t *testing.T, vcdClient *VCDClient) string {

	// Lookup available Edge Clusters to explicitly specify for edge gateway
	_, vdc, err := vcdClient.GetOrgAndVdc(testConfig.VCD.Org, testConfig.Nsxt.Vdc)
	if err != nil {
		t.Errorf("error retrieving vdc: %s", err)
		t.FailNow()
	}

	eClusters, err := vdc.GetAllNsxtEdgeClusters(nil)
	if len(eClusters) < 1 {
		t.Errorf("no edge clusters found: %s", err)
		t.FailNow()
	}

	return eClusters[0].NsxtEdgeCluster.ID
}
