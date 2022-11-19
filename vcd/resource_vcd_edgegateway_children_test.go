//go:build gateway || ALL || functional
// +build gateway ALL functional

package vcd

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// TestAccVcdEdgeGatewayChildrenResourceNotFound checks that deletion of parent Edge Gateway network
// is correctly handled when resource disappears (removes ID by using d.SetId("") instead of
// throwing error) outside of Terraform control. The following resources are verified here:
// * vcd_edgegateway_settings
// * vcd_nsxv_dhcp_relay
func TestAccVcdEdgeGatewayChildrenResourceNotFound(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)
	// This test invokes go-vcloud-director SDK directly
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	// String map to fill the template
	var params = StringMap{
		"Org":                   testConfig.VCD.Org,
		"Vdc":                   testConfig.VCD.Vdc,
		"EdgeGateway":           edgeGatewayNameBasic,
		"EdgeGatewayVcd":        "test_edge_gateway_basic",
		"ExternalNetwork":       testConfig.Networking.ExternalNetwork,
		"Advanced":              "true",
		"Tags":                  "gateway",
		"NewExternalNetwork":    "TestExternalNetwork",
		"NewExternalNetworkVcd": t.Name(),
		"Type":                  testConfig.Networking.ExternalNetworkPortGroupType,
		"PortGroup":             testConfig.Networking.ExternalNetworkPortGroup,
		"Vcenter":               testConfig.Networking.Vcenter,
	}
	testParamsNotEmpty(t, params)

	configText := templateFill(testAccEdgeGatewayResourceNotFound, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

	cachedEdgeGatewayId := &testCachedFieldValue{}
	cachedOrgVdcNetworkId := &testCachedFieldValue{}
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVcdEdgeGatewayDestroy(edgeGatewayNameBasic),
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_edgegateway.egw", "id"),
					cachedEdgeGatewayId.cacheTestResourceFieldValue("vcd_edgegateway.egw", "id"),
					cachedOrgVdcNetworkId.cacheTestResourceFieldValue("vcd_network_routed.test-routed", "id"),
				),
			},
			{
				// This function finds newly created resource and deletes it before
				// next plan check. In this case it will remove Org VDC Network and Edge Gateway
				PreConfig: func() {
					vcdClient := createSystemTemporaryVCDConnection()
					org, err := vcdClient.GetAdminOrgByName(params["Org"].(string))
					if err != nil {
						t.Errorf("error: could not find Org: %s", err)
					}
					vdc, err := org.GetVDCByName(params["Vdc"].(string), false)
					if err != nil {
						t.Errorf("error: could not find VDC: %s", err)
					}

					// Delete Routed Org VDC Network and Edge Gateway itself
					orgVdcNetwork, err := vdc.GetOrgVdcNetworkById(cachedOrgVdcNetworkId.fieldValue, false)
					if err != nil {
						t.Errorf("error: could not find Org VDC Network: %s", err)
					}

					task, err := orgVdcNetwork.Delete()
					if err != nil {
						t.Errorf("error initiating Org VDC network deletion task: %s", err)
					}

					err = task.WaitTaskCompletion()
					if err != nil {
						t.Errorf("error deleting Org VDC network: %s", err)
					}

					edgeGw, err := vdc.GetEdgeGatewayById(cachedEdgeGatewayId.fieldValue, false)
					if err != nil {
						t.Errorf("error: could not find EdgeGateway: %s", err)
					}

					err = edgeGw.Delete(true, true)
					if err != nil {
						t.Errorf("error deleting Edge Gateway: %s", err)
					}
				},
				// Expecting to get a non-empty plan because resource was removed using SDK in
				// PreConfig
				Config:             configText,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
	postTestChecks(t)
}

const testAccEdgeGatewayResourceNotFound = `
resource "vcd_external_network" "extnet" {
  name        = "{{.NewExternalNetworkVcd}}"
  description = "Test External Network"

  vsphere_network {
    vcenter = "{{.Vcenter}}"
    name    = "{{.PortGroup}}"
    type    = "{{.Type}}"
  }

  ip_scope {
    gateway      = "192.168.30.49"
    netmask      = "255.255.255.240"
    dns1         = "192.168.0.164"
    dns2         = "192.168.0.196"
    dns_suffix   = "company.biz"

    static_ip_pool {
      start_address = "192.168.30.51"
      end_address   = "192.168.30.62"
    }
  }


  retain_net_info_across_deployments = "false"
}
resource "vcd_edgegateway" "egw" {
  org                     = "{{.Org}}"
  vdc                     = "{{.Vdc}}"
  name                    = "{{.EdgeGatewayVcd}}"
  description             = "Description"
  configuration           = "compact"

  external_network {
     name = vcd_external_network.extnet.name
   
     subnet {
		ip_address = "192.168.30.51"
		gateway = "192.168.30.49"
		netmask = "255.255.255.240"
		use_for_default_route = true
	}
  }
}

resource "vcd_edgegateway_settings" "egw-settings" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  edge_gateway_id = vcd_edgegateway.egw.id
  fw_enabled      = true
}

resource "vcd_nsxv_dhcp_relay" "relay_config" {
  org          = "{{.Org}}"
  vdc          = "{{.Vdc}}"
  edge_gateway = vcd_edgegateway.egw.name


  relay_agent {
    network_name = vcd_network_routed.test-routed.name
  }
}

resource "vcd_network_routed" "test-routed" {
  name           = "dhcp-relay"
  org            = "{{.Org}}"
  vdc            = "{{.Vdc}}"
  edge_gateway   = vcd_edgegateway.egw.name
  gateway        = "210.201.11.1"
  netmask        = "255.255.255.0"

  static_ip_pool {
    start_address = "210.201.11.10"
    end_address   = "210.201.11.20"
  }
}
`
