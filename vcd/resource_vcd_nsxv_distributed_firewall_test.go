//go:build gateway || nsxt || ALL || functional || vdcGroup

package vcd

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

type distributedFirewallEntities struct {
	org             string
	vdc             string
	routedNetwork   string
	isolatedNetwork string
	directNetwork   string
	vApp            string
	vm              string
	edge            string
	ipSet           string
}

func TestAccVcdNsxvDistributedFirewall(t *testing.T) {
	preTestChecks(t)

	entities := distributedFirewallEntities{
		org:             testConfig.VCD.Org,
		vdc:             testConfig.VCD.Vdc,
		routedNetwork:   fmt.Sprintf("net-%s-r", testConfig.VCD.Org),
		isolatedNetwork: fmt.Sprintf("net-%s-i", testConfig.VCD.Org),
		directNetwork:   fmt.Sprintf("net-%s-d", testConfig.VCD.Org),
		vApp:            "TestVapp",
		vm:              "TestVm",
		ipSet:           "TestIpSet",
		edge:            testConfig.Networking.EdgeGateway,
	}

	var params = StringMap{
		"Org":                   entities.org,
		"Name":                  t.Name(),
		"Vdc":                   entities.vdc,
		"RoutedNetwork":         entities.routedNetwork,
		"IsolatedNetwork":       entities.isolatedNetwork,
		"DirectNetwork":         entities.directNetwork,
		"IpSetName":             entities.ipSet,
		"VappName":              entities.vApp,
		"VmName":                entities.vm,
		"EdgeName":              entities.edge,
		"TestName":              t.Name(),
		"ProviderOrgDef":        " ",
		"ProviderAdminDef":      fmt.Sprintf(`provider = %s`, providerVcdSystem),
		"DistributedFirewallId": "data.vcd_org_vdc.my-vdc.id",
		"FuncName":              t.Name(),
		"Tags":                  "vdc network",
	}
	testParamsNotEmpty(t, params)

	configText := templateFill(testAccNsxvDistributedFirewall, params)
	debugPrintf("#[DEBUG] %s CONFIGURATION: %s", t.Name(), configText)

	params["ProviderOrgDef"] = fmt.Sprintf(`provider = %s`, providerVcdOrg1)
	params["FuncName"] = t.Name() + "-multi"
	params["DistributedFirewallId"] = "vcd_nsxv_distributed_firewall.dfw1Init.vdc_id"
	configTextMulti := templateFill(testAccNsxvDistributedFirewallInit+testAccNsxvDistributedFirewall, params)
	debugPrintf("#[DEBUG] %s CONFIGURATION - multi: %s", t.Name(), configTextMulti)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	dfwResource := "vcd_nsxv_distributed_firewall.dfw1"
	dfwDataSource := "data.vcd_nsxv_distributed_firewall.dfw1-ds"
	t.Run("one-provider", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			ProviderFactories: testAccProviders,
			CheckDestroy:      testCheckDistributedFirewallExistDestroy(entities, false),
			Steps:             distributedFirewallTestSteps(configText, dfwResource, dfwDataSource, entities),
		})
	})
	t.Run("multi-provider", func(t *testing.T) {
		resource.Test(t, resource.TestCase{
			ProviderFactories: buildMultipleProviders(),
			CheckDestroy:      testCheckDistributedFirewallExistDestroy(entities, false),
			Steps:             distributedFirewallTestSteps(configTextMulti, dfwResource, dfwDataSource, entities),
		})
	})
	postTestChecks(t)
}

func distributedFirewallTestSteps(configText, dfwResource, dfwDataSource string, entities distributedFirewallEntities) []resource.TestStep {
	checkInnerRules := func(resourceName string) resource.TestCheckFunc {
		return resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.0.source.*",
				map[string]string{
					"name": entities.ipSet,
					"type": "IPSet",
				}),
			resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.0.application.*",
				map[string]string{
					"protocol":         "TCP",
					"source_port":      "20250",
					"destination_port": "20251",
				}),
			resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.0.application.*",
				map[string]string{
					"name": "POP3",
					"type": "Application",
				}),
			resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.0.application.*",
				map[string]string{
					"type": "ApplicationGroup",
				}),
			resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.0.applied_to.*",
				map[string]string{
					"name": entities.edge,
					"type": "Edge",
				}),
			resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.1.source.*",
				map[string]string{
					"name":  "10.10.1.0-10.10.1.100",
					"value": "10.10.1.0-10.10.1.100",
					"type":  "Ipv4Address",
				}),
			resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.1.applied_to.*",
				map[string]string{
					"name": entities.vdc,
					"type": "VDC",
				}),
			resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.1.destination.*",
				map[string]string{
					"name": entities.routedNetwork,
					"type": "Network",
				}),
			resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.2.destination.*",
				map[string]string{
					"name": entities.isolatedNetwork,
					"type": "Network",
				}),
			resource.TestCheckTypeSetElemNestedAttrs(resourceName, "rule.3.applied_to.*",
				map[string]string{
					"name": entities.vdc,
					"type": "VDC",
				}),
		)
	}
	return []resource.TestStep{
		{
			Config: configText,
			Check: resource.ComposeAggregateTestCheckFunc(
				testCheckDistributedFirewallExistDestroy(entities, true),
				resource.TestCheckResourceAttrSet(dfwResource, "id"),
				resource.TestCheckResourceAttr(dfwResource, "rule.#", "4"),

				resource.TestCheckResourceAttr(dfwResource, "rule.3.name", "fallback"),
				resource.TestCheckResourceAttr(dfwResource, "rule.3.action", "deny"),
				resource.TestCheckResourceAttr(dfwResource, "rule.3.direction", "inout"),

				resource.TestCheckResourceAttr(dfwResource, "rule.2.name", "negated-destination"),
				resource.TestCheckResourceAttr(dfwResource, "rule.2.action", "allow"),
				resource.TestCheckResourceAttr(dfwResource, "rule.2.direction", "in"),
				resource.TestCheckResourceAttr(dfwResource, "rule.2.exclude_source", "false"),
				resource.TestCheckResourceAttr(dfwResource, "rule.2.exclude_destination", "true"),

				resource.TestCheckResourceAttr(dfwResource, "rule.1.name", "negated-source"),
				resource.TestCheckResourceAttr(dfwResource, "rule.1.action", "allow"),
				resource.TestCheckResourceAttr(dfwResource, "rule.1.direction", "in"),
				resource.TestCheckResourceAttr(dfwResource, "rule.1.exclude_source", "true"),
				resource.TestCheckResourceAttr(dfwResource, "rule.1.exclude_destination", "false"),

				resource.TestCheckResourceAttr(dfwResource, "rule.0.name", "straight"),
				resource.TestCheckResourceAttr(dfwResource, "rule.0.action", "allow"),
				resource.TestCheckResourceAttr(dfwResource, "rule.0.direction", "inout"),
				checkInnerRules(dfwResource),

				// Check data source correspondence
				resource.TestCheckResourceAttrPair(dfwResource, "rule.0.name", dfwDataSource, "rule.0.name"),
				resource.TestCheckResourceAttrPair(dfwResource, "rule.0.action", dfwDataSource, "rule.0.action"),
				resource.TestCheckResourceAttrPair(dfwResource, "rule.0.direction", dfwDataSource, "rule.0.direction"),
				resource.TestCheckResourceAttrPair(dfwResource, "rule.1.name", dfwDataSource, "rule.1.name"),
				resource.TestCheckResourceAttrPair(dfwResource, "rule.1.action", dfwDataSource, "rule.1.action"),
				resource.TestCheckResourceAttrPair(dfwResource, "rule.1.direction", dfwDataSource, "rule.1.direction"),
				resource.TestCheckResourceAttrPair(dfwResource, "rule.2.name", dfwDataSource, "rule.2.name"),
				resource.TestCheckResourceAttrPair(dfwResource, "rule.2.action", dfwDataSource, "rule.2.action"),
				resource.TestCheckResourceAttrPair(dfwResource, "rule.2.direction", dfwDataSource, "rule.2.direction"),
				resource.TestCheckResourceAttrPair(dfwResource, "rule.3.name", dfwDataSource, "rule.3.name"),
				resource.TestCheckResourceAttrPair(dfwResource, "rule.3.action", dfwDataSource, "rule.3.action"),
				resource.TestCheckResourceAttrPair(dfwResource, "rule.3.direction", dfwDataSource, "rule.3.direction"),
				checkInnerRules(dfwDataSource),
			),
		},
		{
			ResourceName:      dfwResource,
			ImportState:       true,
			ImportStateVerify: true,
			ImportStateIdFunc: importStateIdOrgObject(entities.org, entities.vdc),
		},
	}
}

func testCheckDistributedFirewallExistDestroy(entities distributedFirewallEntities, wantExist bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*VCDClient)
		org, err := conn.GetAdminOrgByName(entities.org)
		if err != nil {
			return err
		}
		vdc, err := org.GetVDCByName(entities.vdc, false)
		if err != nil {
			return err
		}
		dfw := govcd.NewNsxvDistributedFirewall(&conn.Client, vdc.Vdc.ID)
		configuration, err := dfw.GetConfiguration()
		if wantExist {
			if err != nil || (configuration == nil || configuration.Layer3Sections == nil || configuration.Layer3Sections.Section == nil ||
				len(configuration.Layer3Sections.Section.Rule) == 0) {
				return fmt.Errorf("distributed firewall for VDC %s does not exist", vdc.Vdc.Name)
			}
		} else {
			if err == nil || (configuration != nil && configuration.Layer3Sections != nil && configuration.Layer3Sections.Section != nil &&
				len(configuration.Layer3Sections.Section.Rule) > 0) {
				return fmt.Errorf("distributed firewall for VDC %s still exists", vdc.Vdc.Name)
			}
		}
		return nil
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

#data "vcd_network_direct" "net-d" {
#  {{.ProviderOrgDef}}
#  vdc  = data.vcd_org_vdc.my-vdc.name
#  name = "{{.DirectNetwork}}"
#}

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

data "vcd_nsxv_distributed_firewall" "dfw1-ds" {
  {{.ProviderOrgDef}}
  vdc_id  = vcd_nsxv_distributed_firewall.dfw1.vdc_id
}
`

const testAccNsxvDistributedFirewallInit = `
data "vcd_org_vdc" "my-vdc-admin" {
  {{.ProviderAdminDef}}
  org  = "{{.Org}}"
  name = "{{.Vdc}}"
}

resource "vcd_nsxv_distributed_firewall" "dfw1Init" {
  {{.ProviderAdminDef}}
  vdc_id  = data.vcd_org_vdc.my-vdc-admin.id
  lifecycle {
    ignore_changes = [ rule ]
  }
}
`
