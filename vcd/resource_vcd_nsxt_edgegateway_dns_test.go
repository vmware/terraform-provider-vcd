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
	datasourceName := "data.vcd_nsxt_edgegateway_dns." + "data_" + params["DnsConfig"].(string)
	datasourceNameVdcGroup := "data.vcd_nsxt_edgegateway_dns." + "data_" + params["VdcGroupDnsConfig"].(string)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckNsxtEdgegatewayDnsDestroy(testConfig.Nsxt.Vdc, params["EdgeGw"].(string)),
			testAccCheckNsxtEdgegatewayDnsDestroy(testConfig.Nsxt.VdcGroup, params["VdcGroupEdgeGw"].(string)),
		),
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

					resource.TestCheckResourceAttr(datasourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(datasourceName, "default_forwarder_zone.0.name", params["DefaultForwarderName"].(string)),
					resource.TestCheckTypeSetElemAttr(datasourceName, "default_forwarder_zone.0.upstream_servers.*", params["ServerIp1"].(string)),
					resource.TestCheckTypeSetElemAttr(datasourceName, "default_forwarder_zone.0.upstream_servers.*", params["ServerIp2"].(string)),
					resource.TestCheckResourceAttr(datasourceName, "conditional_forwarder_zone.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(datasourceName, "conditional_forwarder_zone.*", map[string]string{
						"name": params["ConditionalForwardZone1"].(string),
					}),
					resource.TestCheckTypeSetElemNestedAttrs(datasourceName, "conditional_forwarder_zone.*", map[string]string{
						"name": params["ConditionalForwardZone2"].(string),
					}),
					resource.TestCheckTypeSetElemAttr(datasourceName, "conditional_forwarder_zone.*.upstream_servers.*", params["ServerIp5"].(string)),
					resource.TestCheckTypeSetElemAttr(datasourceName, "conditional_forwarder_zone.*.domain_names.*", params["DomainName2"].(string)),
					resource.TestCheckTypeSetElemAttr(datasourceName, "conditional_forwarder_zone.*.domain_names.*", params["DomainName3"].(string)),

					resource.TestCheckTypeSetElemAttr(datasourceName, "conditional_forwarder_zone.*.upstream_servers.*", params["ServerIp3"].(string)),
					resource.TestCheckTypeSetElemAttr(datasourceName, "conditional_forwarder_zone.*.domain_names.*", params["DomainName1"].(string)),

					resource.TestCheckResourceAttr(datasourceNameVdcGroup, "enabled", "true"),

					resource.TestCheckResourceAttr(datasourceNameVdcGroup, "default_forwarder_zone.0.name", params["DefaultForwarderName"].(string)),
					resource.TestCheckTypeSetElemAttr(datasourceNameVdcGroup, "default_forwarder_zone.0.upstream_servers.*", params["ServerIp1"].(string)),
					resource.TestCheckTypeSetElemAttr(datasourceNameVdcGroup, "default_forwarder_zone.0.upstream_servers.*", params["ServerIp2"].(string)),
					resource.TestCheckResourceAttr(datasourceNameVdcGroup, "conditional_forwarder_zone.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(datasourceNameVdcGroup, "conditional_forwarder_zone.*", map[string]string{
						"name": params["ConditionalForwardZone1"].(string),
					}),
					resource.TestCheckTypeSetElemNestedAttrs(datasourceNameVdcGroup, "conditional_forwarder_zone.*", map[string]string{
						"name": params["ConditionalForwardZone2"].(string),
					}),
					resource.TestCheckTypeSetElemAttr(datasourceNameVdcGroup, "conditional_forwarder_zone.*.upstream_servers.*", params["ServerIp5"].(string)),
					resource.TestCheckTypeSetElemAttr(datasourceNameVdcGroup, "conditional_forwarder_zone.*.domain_names.*", params["DomainName2"].(string)),
					resource.TestCheckTypeSetElemAttr(datasourceNameVdcGroup, "conditional_forwarder_zone.*.domain_names.*", params["DomainName3"].(string)),

					resource.TestCheckTypeSetElemAttr(datasourceNameVdcGroup, "conditional_forwarder_zone.*.upstream_servers.*", params["ServerIp3"].(string)),
					resource.TestCheckTypeSetElemAttr(datasourceNameVdcGroup, "conditional_forwarder_zone.*.domain_names.*", params["DomainName1"].(string)),
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

data "vcd_nsxt_edgegateway_dns" "data_{{.DnsConfig}}" {
  edge_gateway_id = vcd_nsxt_edgegateway_dns.{{.DnsConfig}}.edge_gateway_id

  depends_on = [vcd_nsxt_edgegateway_dns.{{.DnsConfig}}]
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

data "vcd_nsxt_edgegateway_dns" "data_{{.VdcGroupDnsConfig}}" {
  edge_gateway_id = vcd_nsxt_edgegateway_dns.{{.VdcGroupDnsConfig}}.edge_gateway_id

  depends_on = [vcd_nsxt_edgegateway_dns.{{.VdcGroupDnsConfig}}]
}
`

func TestAccVcdNsxtEdgegatewayDnsIpSpaces(t *testing.T) {
	preTestChecks(t)

	if checkVersion(testConfig.Provider.ApiVersion, "<38.0") {
		t.Skip("This test is only supported since version 38.1 of the API")
	}

	// String map to fill the template
	var params = StringMap{
		"TestName":             t.Name(),
		"NsxtManager":          testConfig.Nsxt.Manager,
		"NsxtTier0Router":      testConfig.Nsxt.Tier0router,
		"Org":                  testConfig.VCD.Org,
		"NsxtVdc":              testConfig.Nsxt.Vdc,
		"EdgeGw":               "ip-spaces-egw",
		"ExtNet":               t.Name(),
		"DnsConfig":            t.Name(),
		"DefaultForwarderName": t.Name() + "default",
		"ServerIp1":            "1.1.1.1",
		"ServerIp2":            "2.2.2.2",
		"ServerIp3":            "3.3.3.3",
		"IpAddress1":           "11.11.11.100",
		"IpAddress2":           "11.11.11.101",
	}
	testParamsNotEmpty(t, params)

	params["FuncName"] = t.Name() + "step1"
	configText1 := templateFill(testAccVcdNsxtEdgegatewayDnsIpSpacesStep1, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s\n", configText1)

	params["FuncName"] = t.Name() + "step2"
	configText2 := templateFill(testAccVcdNsxtEdgegatewayDnsIpSpacesStep2, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s\n", configText2)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resourceName := "vcd_nsxt_edgegateway_dns." + params["DnsConfig"].(string)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVcdNsxtEdgeGatewayDestroy(params["EdgeGw"].(string)),
		Steps: []resource.TestStep{
			{
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "default_forwarder_zone.0.name", params["DefaultForwarderName"].(string)),
					resource.TestCheckTypeSetElemAttr(resourceName, "default_forwarder_zone.0.upstream_servers.*", params["ServerIp1"].(string)),
					resource.TestCheckTypeSetElemAttr(resourceName, "default_forwarder_zone.0.upstream_servers.*", params["ServerIp2"].(string)),
					resource.TestCheckTypeSetElemAttr(resourceName, "default_forwarder_zone.0.upstream_servers.*", params["ServerIp3"].(string)),
					resource.TestCheckResourceAttr(resourceName, "default_forwarder_zone.0.name", params["DefaultForwarderName"].(string)),
					resource.TestCheckResourceAttr(resourceName, "snat_rule_ip_address", params["IpAddress1"].(string)),
					resource.TestCheckResourceAttr(resourceName, "snat_rule_enabled", "true"),
				),
			},
			{
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "default_forwarder_zone.0.name", params["DefaultForwarderName"].(string)),
					resource.TestCheckTypeSetElemAttr(resourceName, "default_forwarder_zone.0.upstream_servers.*", params["ServerIp1"].(string)),
					resource.TestCheckTypeSetElemAttr(resourceName, "default_forwarder_zone.0.upstream_servers.*", params["ServerIp2"].(string)),
					resource.TestCheckResourceAttr(resourceName, "snat_rule_ip_address", params["IpAddress2"].(string)),
					resource.TestCheckResourceAttr(resourceName, "snat_rule_enabled", "true"),
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

const testAccVcdNsxtEdgegatewayDnsIpSpacesPrereqs = `
data "vcd_nsxt_manager" "main" {
  name = "{{.NsxtManager}}"
}

data "vcd_nsxt_tier0_router" "router" {
  name            = "{{.NsxtTier0Router}}"
  nsxt_manager_id = data.vcd_nsxt_manager.main.id
}

data "vcd_org" "org1" {
  name = "{{.Org}}"
}

data "vcd_org_vdc" "{{.NsxtVdc}}" {
  org  = "{{.Org}}"
  name = "{{.NsxtVdc}}"
}

resource "vcd_ip_space" "ipSpace1" {
  name = "{{.TestName}}"
  type = "PUBLIC"

  internal_scope = ["11.11.11.0/24"]
  external_scope = "0.0.0.0/24"

  route_advertisement_enabled = false

  ip_range {
    start_address = "11.11.11.100"
    end_address   = "11.11.11.110"
  }
}

resource "vcd_external_network_v2" "{{.ExtNet}}" {
  name = "{{.ExtNet}}"

  nsxt_network {
    nsxt_manager_id      = data.vcd_nsxt_manager.main.id
    nsxt_tier0_router_id = data.vcd_nsxt_tier0_router.router.id
  }

  use_ip_spaces = true
}

resource "vcd_ip_space_uplink" "u1" {
  name                = "{{.TestName}}"
  external_network_id = vcd_external_network_v2.{{.ExtNet}}.id
  ip_space_id         = vcd_ip_space.ipSpace1.id
}

resource "vcd_nsxt_edgegateway" "{{.EdgeGw}}" {
  org                 = "{{.Org}}"
  name                = "{{.EdgeGw}}"
  owner_id            = data.vcd_org_vdc.{{.NsxtVdc}}.id
  external_network_id = vcd_external_network_v2.{{.ExtNet}}.id

  depends_on = [vcd_ip_space_uplink.u1]
}

resource "vcd_ip_space_ip_allocation" "public-floating-ip" {
  org_id      = data.vcd_org.org1.id
  ip_space_id = vcd_ip_space.ipSpace1.id
  type        = "FLOATING_IP"

  value = "{{.IpAddress1}}"

  depends_on = [vcd_nsxt_edgegateway.{{.EdgeGw}}]
}

resource "vcd_ip_space_ip_allocation" "public-floating-ip-2" {
  org_id      = data.vcd_org.org1.id
  ip_space_id = vcd_ip_space.ipSpace1.id
  type        = "FLOATING_IP"

  value = "{{.IpAddress2}}"

  depends_on = [vcd_nsxt_edgegateway.{{.EdgeGw}}]
}
`

const testAccVcdNsxtEdgegatewayDnsIpSpacesStep1 = testAccVcdNsxtEdgegatewayDnsIpSpacesPrereqs + `
resource "vcd_nsxt_edgegateway_dns" "{{.DnsConfig}}" {
  edge_gateway_id      = vcd_nsxt_edgegateway.{{.EdgeGw}}.id
  enabled              = true
  snat_rule_ip_address = vcd_ip_space_ip_allocation.public-floating-ip.ip_address

  default_forwarder_zone {
    name = "{{.DefaultForwarderName}}"

    upstream_servers = [
      "{{.ServerIp1}}",
      "{{.ServerIp2}}",
      "{{.ServerIp3}}",
    ]
  }
}
`

const testAccVcdNsxtEdgegatewayDnsIpSpacesStep2 = testAccVcdNsxtEdgegatewayDnsIpSpacesPrereqs + `
resource "vcd_nsxt_edgegateway_dns" "{{.DnsConfig}}" {
  edge_gateway_id      = vcd_nsxt_edgegateway.{{.EdgeGw}}.id
  enabled              = true
  snat_rule_ip_address = vcd_ip_space_ip_allocation.public-floating-ip-2.ip_address

  default_forwarder_zone {
    name = "{{.DefaultForwarderName}}"

    upstream_servers = [
      "{{.ServerIp1}}",
      "{{.ServerIp2}}",
      "{{.ServerIp3}}",
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
