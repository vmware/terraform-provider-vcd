//go:build network || nsxt || ALL || functional

package vcd

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccVcdNsxtEdgegatewayDns(t *testing.T) {
	preTestChecks(t)

	// String map to fill the template
	var params = StringMap{
		"Org":                     testConfig.VCD.Org,
		"NsxtVdc":                 testConfig.Nsxt.Vdc,
		"VdcGroup":                testConfig.Nsxt.VdcGroup,
		"VdcGroupEdgeGw":          testConfig.Nsxt.VdcGroupEdgeGateway,
		"EdgeGw":                  testConfig.Nsxt.EdgeGateway,
		"DnsConfig":               t.Name(),
		"VdcGroupDnsConfig":       t.Name() + "vdcgroup",
		"DefaultForwarderName":    t.Name() + "default",
		"ConditionalForwardZone1": t.Name() + "conditional1",
		"ConditionalForwardZone2": t.Name() + "conditional2",
		"ServerIp1":               "1.1.1.1",
		"ServerIp2":               "2.2.2.2",
		"ServerIp3":               "3.3.3.3",
		"ServerIp4":               "4.4.4.4",
		"ServerIp5":               "5.5.5.5",
		"DomainName1":             "example.org",
		"DomainName2":             "testwebsite.org",
		"DomainName3":             "nonexistent.nan",
	}
	testParamsNotEmpty(t, params)

	params["FuncName"] = t.Name() + "step1"
	configText1 := templateFill(testAccVcdNsxtEdgegatewayDnsStep1, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s\n", configText1)

	params["FuncName"] = t.Name() + "step2"
	configText2 := templateFill(testAccVcdNsxtEdgegatewayDnsStep2, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s\n", configText2)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resourceName := "vcd_nsxt_edgegateway_dns." + params["DnsConfig"].(string)
	resourceNameVdcGroup := "vcd_nsxt_edgegateway_dns." + params["VdcGroupDnsConfig"].(string)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckNsxtEdgegatewayDnsDestroy(testConfig.Nsxt.Vdc, testConfig.Nsxt.EdgeGateway),
		Steps: []resource.TestStep{
			{
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "default_forwarder_zone.0.name", params["DefaultForwarderName"].(string)),
					resource.TestCheckTypeSetElemAttr(resourceName, "default_forwarder_zone.0.upstream_servers.*", params["ServerIp1"].(string)),
					resource.TestCheckTypeSetElemAttr(resourceName, "default_forwarder_zone.0.upstream_servers.*", params["ServerIp2"].(string)),
					resource.TestCheckTypeSetElemAttr(resourceName, "default_forwarder_zone.0.upstream_servers.*", params["ServerIp3"].(string)),
					resource.TestCheckResourceAttr(resourceName, "conditional_forwarder_zone.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "conditional_forwarder_zone.*", map[string]string{
						"name": params["ConditionalForwardZone1"].(string),
					}),
					resource.TestCheckTypeSetElemAttr(resourceName, "conditional_forwarder_zone.*.upstream_servers.*", params["ServerIp4"].(string)),
					resource.TestCheckTypeSetElemAttr(resourceName, "conditional_forwarder_zone.*.upstream_servers.*", params["ServerIp5"].(string)),
					resource.TestCheckTypeSetElemAttr(resourceName, "conditional_forwarder_zone.*.domain_names.*", params["DomainName1"].(string)),
					resource.TestCheckTypeSetElemAttr(resourceName, "conditional_forwarder_zone.*.domain_names.*", params["DomainName2"].(string)),
					resource.TestCheckTypeSetElemAttr(resourceName, "conditional_forwarder_zone.*.domain_names.*", params["DomainName3"].(string)),

					resource.TestCheckResourceAttr(resourceNameVdcGroup, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceNameVdcGroup, "default_forwarder_zone.0.name", params["DefaultForwarderName"].(string)),
					resource.TestCheckTypeSetElemAttr(resourceNameVdcGroup, "default_forwarder_zone.0.upstream_servers.*", params["ServerIp1"].(string)),
					resource.TestCheckTypeSetElemAttr(resourceNameVdcGroup, "default_forwarder_zone.0.upstream_servers.*", params["ServerIp2"].(string)),
					resource.TestCheckTypeSetElemAttr(resourceNameVdcGroup, "default_forwarder_zone.0.upstream_servers.*", params["ServerIp3"].(string)),
					resource.TestCheckResourceAttr(resourceNameVdcGroup, "conditional_forwarder_zone.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceNameVdcGroup, "conditional_forwarder_zone.*", map[string]string{
						"name": params["ConditionalForwardZone1"].(string),
					}),
					resource.TestCheckTypeSetElemAttr(resourceNameVdcGroup, "conditional_forwarder_zone.*.upstream_servers.*", params["ServerIp4"].(string)),
					resource.TestCheckTypeSetElemAttr(resourceNameVdcGroup, "conditional_forwarder_zone.*.upstream_servers.*", params["ServerIp5"].(string)),
					resource.TestCheckTypeSetElemAttr(resourceNameVdcGroup, "conditional_forwarder_zone.*.domain_names.*", params["DomainName1"].(string)),
					resource.TestCheckTypeSetElemAttr(resourceNameVdcGroup, "conditional_forwarder_zone.*.domain_names.*", params["DomainName2"].(string)),
					resource.TestCheckTypeSetElemAttr(resourceNameVdcGroup, "conditional_forwarder_zone.*.domain_names.*", params["DomainName3"].(string)),
				),
			},
			{
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),

					resource.TestCheckResourceAttr(resourceName, "default_forwarder_zone.0.name", params["DefaultForwarderName"].(string)),
					resource.TestCheckTypeSetElemAttr(resourceName, "default_forwarder_zone.0.upstream_servers.*", params["ServerIp1"].(string)),
					resource.TestCheckTypeSetElemAttr(resourceName, "default_forwarder_zone.0.upstream_servers.*", params["ServerIp2"].(string)),
					resource.TestCheckResourceAttr(resourceName, "conditional_forwarder_zone.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "conditional_forwarder_zone.*", map[string]string{
						"name": params["ConditionalForwardZone1"].(string),
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "conditional_forwarder_zone.*", map[string]string{
						"name": params["ConditionalForwardZone2"].(string),
					}),
					resource.TestCheckTypeSetElemAttr(resourceName, "conditional_forwarder_zone.*.upstream_servers.*", params["ServerIp5"].(string)),
					resource.TestCheckTypeSetElemAttr(resourceName, "conditional_forwarder_zone.*.domain_names.*", params["DomainName2"].(string)),
					resource.TestCheckTypeSetElemAttr(resourceName, "conditional_forwarder_zone.*.domain_names.*", params["DomainName3"].(string)),

					resource.TestCheckTypeSetElemAttr(resourceName, "conditional_forwarder_zone.*.upstream_servers.*", params["ServerIp3"].(string)),
					resource.TestCheckTypeSetElemAttr(resourceName, "conditional_forwarder_zone.*.domain_names.*", params["DomainName1"].(string)),

					resource.TestCheckResourceAttr(resourceNameVdcGroup, "enabled", "true"),

					resource.TestCheckResourceAttr(resourceNameVdcGroup, "default_forwarder_zone.0.name", params["DefaultForwarderName"].(string)),
					resource.TestCheckTypeSetElemAttr(resourceNameVdcGroup, "default_forwarder_zone.0.upstream_servers.*", params["ServerIp1"].(string)),
					resource.TestCheckTypeSetElemAttr(resourceNameVdcGroup, "default_forwarder_zone.0.upstream_servers.*", params["ServerIp2"].(string)),
					resource.TestCheckResourceAttr(resourceNameVdcGroup, "conditional_forwarder_zone.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceNameVdcGroup, "conditional_forwarder_zone.*", map[string]string{
						"name": params["ConditionalForwardZone1"].(string),
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceNameVdcGroup, "conditional_forwarder_zone.*", map[string]string{
						"name": params["ConditionalForwardZone2"].(string),
					}),
					resource.TestCheckTypeSetElemAttr(resourceNameVdcGroup, "conditional_forwarder_zone.*.upstream_servers.*", params["ServerIp5"].(string)),
					resource.TestCheckTypeSetElemAttr(resourceNameVdcGroup, "conditional_forwarder_zone.*.domain_names.*", params["DomainName2"].(string)),
					resource.TestCheckTypeSetElemAttr(resourceNameVdcGroup, "conditional_forwarder_zone.*.domain_names.*", params["DomainName3"].(string)),

					resource.TestCheckTypeSetElemAttr(resourceNameVdcGroup, "conditional_forwarder_zone.*.upstream_servers.*", params["ServerIp3"].(string)),
					resource.TestCheckTypeSetElemAttr(resourceNameVdcGroup, "conditional_forwarder_zone.*.domain_names.*", params["DomainName1"].(string)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdFunc:       importStateIdOrgNsxtVdcObject(params["EdgeGw"].(string)),
				ImportStateVerifyIgnore: []string{"org"},
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdNsxtEdgegatewayDnsPrereqs = `
data "vcd_org_vdc" "{{.NsxtVdc}}" {
  name = "{{.NsxtVdc}}"		
}

data "vcd_vdc_group" "{{.VdcGroup}}" {
  org  = "{{.Org}}"
  name = "{{.VdcGroup}}"
}
	
data "vcd_nsxt_edgegateway" "{{.EdgeGw}}" {
  owner_id = data.vcd_org_vdc.{{.NsxtVdc}}.id
  name     = "{{.EdgeGw}}"
}

data "vcd_nsxt_edgegateway" "{{.VdcGroupEdgeGw}}" {
  owner_id = data.vcd_vdc_group.{{.VdcGroup}}.id
  name     = "{{.VdcGroupEdgeGw}}"
}
`

const testAccVcdNsxtEdgegatewayDnsStep1 = testAccVcdNsxtEdgegatewayDnsPrereqs + `
resource "vcd_nsxt_edgegateway_dns" "{{.DnsConfig}}" {
  edge_gateway_id = data.vcd_nsxt_edgegateway.{{.EdgeGw}}.id
  enabled         = true
  
  default_forwarder_zone {
    name = "{{.DefaultForwarderName}}"
    
    upstream_servers = [
      "{{.ServerIp1}}",
      "{{.ServerIp2}}",
      "{{.ServerIp3}}",
    ]
  }

  conditional_forwarder_zone {
    name = "{{.ConditionalForwardZone1}}"

    upstream_servers = [
      "{{.ServerIp4}}",
      "{{.ServerIp5}}",
    ]

    domain_names = [
      "{{.DomainName1}}",
      "{{.DomainName2}}",
      "{{.DomainName3}}",
    ]
  }
}

resource "vcd_nsxt_edgegateway_dns" "{{.VdcGroupDnsConfig}}" {
  edge_gateway_id = data.vcd_nsxt_edgegateway.{{.VdcGroupEdgeGw}}.id
  enabled         = true
  
  default_forwarder_zone {
    name = "{{.DefaultForwarderName}}"
    
    upstream_servers = [
      "{{.ServerIp1}}",
      "{{.ServerIp2}}",
      "{{.ServerIp3}}",
    ]
  }

  conditional_forwarder_zone {
    name = "{{.ConditionalForwardZone1}}"

    upstream_servers = [
      "{{.ServerIp4}}",
      "{{.ServerIp5}}",
    ]

    domain_names = [
      "{{.DomainName1}}",
      "{{.DomainName2}}",
      "{{.DomainName3}}",
    ]
  }
}
`

const testAccVcdNsxtEdgegatewayDnsStep2 = testAccVcdNsxtEdgegatewayDnsPrereqs + `
resource "vcd_nsxt_edgegateway_dns" "{{.DnsConfig}}" {
  edge_gateway_id = data.vcd_nsxt_edgegateway.{{.EdgeGw}}.id
  enabled         = true
  
  default_forwarder_zone {
    name = "{{.DefaultForwarderName}}"
    
    upstream_servers = [
      "{{.ServerIp1}}",
      "{{.ServerIp2}}",
    ]
  }

  conditional_forwarder_zone {
    name = "{{.ConditionalForwardZone1}}"

    upstream_servers = [
      "{{.ServerIp5}}",
    ]

    domain_names = [
      "{{.DomainName2}}",
      "{{.DomainName3}}",
    ]
  }

  conditional_forwarder_zone {
    name = "{{.ConditionalForwardZone2}}"

    upstream_servers = [
      "{{.ServerIp3}}",
    ]

    domain_names = [
      "{{.DomainName1}}",
    ]
  }
}

resource "vcd_nsxt_edgegateway_dns" "{{.VdcGroupDnsConfig}}" {
  edge_gateway_id = data.vcd_nsxt_edgegateway.{{.EdgeGw}}.id
  enabled         = true
  
  default_forwarder_zone {
    name = "{{.DefaultForwarderName}}"
    
    upstream_servers = [
      "{{.ServerIp1}}",
      "{{.ServerIp2}}",
    ]
  }

  conditional_forwarder_zone {
    name = "{{.ConditionalForwardZone1}}"

    upstream_servers = [
      "{{.ServerIp5}}",
    ]

    domain_names = [
      "{{.DomainName2}}",
      "{{.DomainName3}}",
    ]
  }

  conditional_forwarder_zone {
    name = "{{.ConditionalForwardZone2}}"

    upstream_servers = [
      "{{.ServerIp3}}",
    ]

    domain_names = [
      "{{.DomainName1}}",
    ]
  }
}
`

func testAccCheckNsxtEdgegatewayDnsDestroy(vdcOrVdcGroupName, edgeGatewayName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*VCDClient)

		vdcOrVdcGroup, err := lookupVdcOrVdcGroup(conn, testConfig.VCD.Org, vdcOrVdcGroupName)
		if err != nil {
			return fmt.Errorf("unable to find VDC or VDC group %s: %s", vdcOrVdcGroupName, err)
		}

		edge, err := vdcOrVdcGroup.GetNsxtEdgeGatewayByName(edgeGatewayName)
		if err != nil {
			return fmt.Errorf(errorUnableToFindEdgeGateway, edgeGatewayName)
		}

		dns, err := edge.GetDnsConfig()
		if err != nil {
			return fmt.Errorf("failed to get DNS configuration: %s", err)
		}

		if dns.NsxtEdgeGatewayDns.Enabled || dns.NsxtEdgeGatewayDns.ListenerIp != "" {
			return fmt.Errorf("dns configuration wasn't deleted")
		}

		return nil
	}
}
