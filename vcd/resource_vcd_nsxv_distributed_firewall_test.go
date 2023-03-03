//go:build gateway || nsxt || ALL || functional || vdcGroup

package vcd

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcdNsxvDistributedFirewall(t *testing.T) {
	preTestChecks(t)

	skipIfNotSysAdmin(t)
	orgName := testConfig.VCD.Org

	// TODO: get the network and vApp/VM names from configuration file (requires provisioner update)
	routedNetworkName := fmt.Sprintf("net-%s-r", orgName)
	directNetworkName := fmt.Sprintf("net-%s-d", orgName)
	isolatedNetworkName := fmt.Sprintf("net-%s-i", orgName)
	vappName := "TestVapp"
	testVmName := "TestVm"
	// String map to fill the template
	var params = StringMap{
		"Org":                   orgName,
		"Name":                  t.Name(),
		"Vdc":                   testConfig.VCD.Vdc,
		"RoutedNetwork":         routedNetworkName,
		"DirectNetwork":         directNetworkName,
		"IsolatedNetwork":       isolatedNetworkName,
		"IpSetName":             "TestIpSet",
		"VappName":              vappName,
		"VmName":                testVmName,
		"EdgeName":              testConfig.Networking.EdgeGateway,
		"TestName":              t.Name(),
		"ProviderOrgDef":        " ",
		"ProviderSystemDef":     fmt.Sprintf(`provider = "%s"`, providerVcdSystem),
		"DistributedFirewallId": "data.vcd_org_vdc.my-vdc.id",
		"FuncName":              t.Name(),
		"SkipMessage":           " ",
		"Tags":                  "vdc network",
	}
	testParamsNotEmpty(t, params)

	configText := templateFill(testAccNsxvDistributedFirewall, params)
	debugPrintf("#[DEBUG] %s CONFIGURATION: %s", t.Name(), configText)

	params["ProviderOrgDef"] = fmt.Sprintf(`provider = "%s"`, providerVcdOrg1)
	params["SkipMessage"] = "# skip-binary-test: not suitable for binary tests"
	params["FuncName"] = t.Name() + "-multi"
	params["DistributedFirewallId"] = "vcd_nsxv_distributed_firewall.dfw1Init.vdc_id"
	configTextMulti := templateFill(testAccNsxvDistributedFirewallInit+testAccNsxvDistributedFirewall, params)
	debugPrintf("#[DEBUG] %s CONFIGURATION: %s", t.Name(), configText)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	dfwResource := "vcd_nsxv_distributed_firewall.dfw1"
	t.Run("all-sysadmin", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			ProviderFactories: testAccProviders,
			Steps:             distributedFirewallTestSteps(configText, dfwResource, isolatedNetworkName),
		})
	})
	t.Run("multi-provider", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			ProviderFactories: buildMultipleProviders(),
			Steps:             distributedFirewallTestSteps(configTextMulti, dfwResource, isolatedNetworkName),
		})
	})
	postTestChecks(t)
}

func distributedFirewallTestSteps(configText, dfwResource, networkName string) []resource.TestStep {

	return []resource.TestStep{
		{
			Config: configText,
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttrSet(dfwResource, "id"),
				resource.TestCheckResourceAttr(dfwResource, "rule.#", "4"),
				//logState("distributed-firewall-rules"),
				resource.TestCheckTypeSetElemNestedAttrs(dfwResource, "rule.1.source.*",
					map[string]string{
						"name":  "10.10.1.0-10.10.1.100",
						"value": "10.10.1.0-10.10.1.100",
						"type":  "Ipv4Address",
					}),
				resource.TestCheckResourceAttr(dfwResource, "rule.3.name", "fallback"),
				resource.TestCheckResourceAttr(dfwResource, "rule.3.action", "deny"),
				resource.TestCheckResourceAttr(dfwResource, "rule.3.direction", "inout"),

				resource.TestCheckResourceAttr(dfwResource, "rule.2.name", "negated-destination"),
				resource.TestCheckResourceAttr(dfwResource, "rule.2.action", "allow"),
				resource.TestCheckResourceAttr(dfwResource, "rule.2.direction", "in"),
				resource.TestCheckResourceAttr(dfwResource, "rule.2.exclude_destination", "true"),
				resource.TestCheckResourceAttr(dfwResource, "rule.2.exclude_source", "false"),

				resource.TestCheckResourceAttr(dfwResource, "rule.1.name", "negated-source"),
				resource.TestCheckResourceAttr(dfwResource, "rule.1.action", "allow"),
				resource.TestCheckResourceAttr(dfwResource, "rule.1.direction", "in"),
				resource.TestCheckResourceAttr(dfwResource, "rule.1.exclude_source", "true"),
				resource.TestCheckResourceAttr(dfwResource, "rule.1.exclude_destination", "false"),

				resource.TestCheckResourceAttr(dfwResource, "rule.0.name", "straight"),
				resource.TestCheckResourceAttr(dfwResource, "rule.0.action", "allow"),
				resource.TestCheckResourceAttr(dfwResource, "rule.0.direction", "inout"),
				resource.TestCheckTypeSetElemNestedAttrs(dfwResource, "rule.0.source.*",
					map[string]string{
						"name": "TestIpSet",
						"type": "IPSet",
					}),
				resource.TestCheckTypeSetElemNestedAttrs(dfwResource, "rule.0.application.*",
					map[string]string{
						"protocol":         "TCP",
						"source_port":      "20250",
						"destination_port": "20251",
					}),
				resource.TestCheckTypeSetElemNestedAttrs(dfwResource, "rule.0.application.*",
					map[string]string{
						"name": "POP3",
						"type": "Application",
					}),
				resource.TestCheckTypeSetElemNestedAttrs(dfwResource, "rule.0.application.*",
					map[string]string{
						"type": "ApplicationGroup",
					}),
				resource.TestCheckTypeSetElemNestedAttrs(dfwResource, "rule.0.applied_to.*",
					map[string]string{
						"name": testConfig.Networking.EdgeGateway,
						"type": "Edge",
					}),
				resource.TestCheckTypeSetElemNestedAttrs(dfwResource, "rule.2.destination.*",
					map[string]string{
						"name": networkName,
						"type": "Network",
					}),
			),
		},
		{
			ResourceName:      dfwResource,
			ImportState:       true,
			ImportStateVerify: true,
			ImportStateIdFunc: importStateIdOrgObject(testConfig.VCD.Org, testConfig.VCD.Vdc),
		},
	}
}

const testAccNsxvDistributedFirewall = `
resource "vcd_nsxv_ip_set" "test-ipset" {
  {{.ProviderOrgDef}}
  org          = "{{.Org}}"
  vdc          = "{{.Vdc}}"

  name         = "{{.IpSetName}}"
  description  = "test-ip-set-description"
  ip_addresses = ["192.168.1.1","192.168.2.1"]
}

data "vcd_org_vdc" "my-vdc" {
  {{.ProviderOrgDef}}
  org  = "{{.Org}}"
  name = "{{.Vdc}}"
}

data "vcd_vapp_vm" "vm1" {
  {{.ProviderOrgDef}}
  vdc       = data.vcd_org_vdc.my-vdc.name
  vapp_name = "{{.VappName}}"
  name      = "{{.VmName}}"
}

data "vcd_network_routed" "net-r" {
  {{.ProviderOrgDef}}
  vdc  = data.vcd_org_vdc.my-vdc.name
  name = "{{.RoutedNetwork}}"
}

data "vcd_network_direct" "net-d" {
  {{.ProviderOrgDef}}
  vdc  = data.vcd_org_vdc.my-vdc.name
  name = "{{.DirectNetwork}}"
}

data "vcd_network_isolated" "net-i" {
  {{.ProviderOrgDef}}
  vdc  = data.vcd_org_vdc.my-vdc.name
  name = "{{.IsolatedNetwork}}"
}

data "vcd_edgegateway" "edge" {
  {{.ProviderOrgDef}}
  vdc  = data.vcd_org_vdc.my-vdc.name
  name = "{{.EdgeName}}"
}

data "vcd_nsxv_application_finder" "found_applications" {
  {{.ProviderOrgDef}}
  vdc_id            = data.vcd_org_vdc.my-vdc.id
  search_expression = "^POP3$"
  case_sensitive    = true
  type              = "application"
}

data "vcd_nsxv_application" "application1" {
  {{.ProviderOrgDef}}
  vdc_id = data.vcd_org_vdc.my-vdc.id
  name   = tolist(data.vcd_nsxv_application_finder.found_applications.objects)[0].name
}

data "vcd_nsxv_application_group" "application_group1" {
  {{.ProviderOrgDef}}
  vdc_id = data.vcd_org_vdc.my-vdc.id
  name   = "MS Exchange 2010 Mailbox Servers"
}

resource "vcd_nsxv_distributed_firewall" "dfw1" {
  {{.ProviderOrgDef}}
  # vdc_id  = data.vcd_org_vdc.my-vdc.id
  vdc_id  = {{.DistributedFirewallId}}
  enabled = true

  # rule n. 0 
  rule {
	name      = "straight"
    direction = "inout"
    action    = "allow"
    source {
      name  = vcd_nsxv_ip_set.test-ipset.name
      value = vcd_nsxv_ip_set.test-ipset.id
      type  = "IPSet"
    }

    # anonymous application
    application {
      protocol         = "TCP"
      source_port      = "20250"
      destination_port = "20251"
	}
    # named application
    application {
      name  = data.vcd_nsxv_application.application1.name
      value = data.vcd_nsxv_application.application1.id
      type  = "Application"
    }
    # named application group
    application {
      name  = data.vcd_nsxv_application_group.application_group1.name
      value = data.vcd_nsxv_application_group.application_group1.id
      type  = "ApplicationGroup"
    }
    applied_to {
        name  = data.vcd_edgegateway.edge.name
        type = "Edge"
        value = data.vcd_edgegateway.edge.id
    }
  }
  # rule n. 1 
  rule {
    name           = "negated-source"
    direction      = "in"
    action         = "allow"
    logged         = true
	exclude_source = true
    # literal source
    source {
      name  = "10.10.1.0-10.10.1.100"
      value = "10.10.1.0-10.10.1.100"
      type  = "Ipv4Address"
    }
    # VM source
    source {
      name  = data.vcd_vapp_vm.vm1.name
      value = data.vcd_vapp_vm.vm1.id
      type  = "VirtualMachine"
    }
    # routed network destination
    destination {
      name  = data.vcd_network_routed.net-r.name
      value = data.vcd_network_routed.net-r.id
      type  = "Network"
    }
    # direct network destination
    # (currently omitted due to VCD bug: the network comes back as "Unknown object(VM Network)")
    #destination {
    #  name  = data.vcd_network_direct.net-d.name
    #  value = data.vcd_network_direct.net-d.id
    #  type  = "Network"
    #}
    applied_to {
        name  = data.vcd_org_vdc.my-vdc.name
        type = "VDC"
        value = data.vcd_org_vdc.my-vdc.id
    }
  }
  # rule # 2
  rule {
    name                = "negated-destination"
    direction           = "in"
    action              = "allow"
    logged              = true
	exclude_destination = true
    # isolated network destination
    destination {
      name  = data.vcd_network_isolated.net-i.name
      value = data.vcd_network_isolated.net-i.id
      type  = "Network"
    }
    applied_to {
        name  = data.vcd_org_vdc.my-vdc.name
        type = "VDC"
        value = data.vcd_org_vdc.my-vdc.id
    }
  }
  # rule n. 3 
  rule {
    name      = "fallback"
    direction = "inout"
    action    = "deny"
    applied_to {
        name  = data.vcd_org_vdc.my-vdc.name
        type = "VDC"
        value = data.vcd_org_vdc.my-vdc.id
    }
  }
}
`

const testAccNsxvDistributedFirewallInit = `
{{.SkipMessage}}
data "vcd_org_vdc" "my-vdc-system" {
  {{.ProviderSystemDef}}
  org  = "{{.Org}}"
  name = "{{.Vdc}}"
}

resource "vcd_nsxv_distributed_firewall" "dfw1Init" {
  {{.ProviderSystemDef}}
  vdc_id  = data.vcd_org_vdc.my-vdc-system.id
  enabled = true
  lifecycle {
    ignore_changes = [ "rule" ]
  }
}

`
