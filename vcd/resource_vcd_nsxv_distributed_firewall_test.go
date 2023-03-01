//go:build gateway || nsxt || ALL || functional || vdcGroup

package vcd

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcdNsxvDistributedFirewall(t *testing.T) {
	preTestChecks(t)

	orgName := testConfig.VCD.Org

	// TODO: get the network and vApp/VM names from configuration file (requires provisioner update)
	routedNetworkName := fmt.Sprintf("net-%s-r", orgName)
	directNetworkName := fmt.Sprintf("net-%s-d", orgName)
	isolatedNetworkName := fmt.Sprintf("net-%s-i", orgName)
	vappName := "TestVapp"
	testVmName := "TestVm"
	// String map to fill the template
	var params = StringMap{
		"Org":             orgName,
		"Name":            t.Name(),
		"Vdc":             testConfig.VCD.Vdc,
		"RoutedNetwork":   routedNetworkName,
		"DirectNetwork":   directNetworkName,
		"IsolatedNetwork": isolatedNetworkName,
		"IpSetName":       "TestIpSet",
		"VappName":        vappName,
		"VmName":          testVmName,
		"EdgeName":        testConfig.Networking.EdgeGateway,
		"TestName":        t.Name(),

		"Tags": "vdc network",
	}
	testParamsNotEmpty(t, params)

	configText := templateFill(testAccNsxvDistributedFirewall, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	dfwResource := "vcd_nsxv_distributed_firewall.dfw1"
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(dfwResource, "id"),
					resource.TestCheckResourceAttr(dfwResource, "rule.#", "3"),
					// TODO: add checks on rules components
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccNsxvDistributedFirewall = `
resource "vcd_nsxv_ip_set" "test-ipset" {
  org          = "{{.Org}}"
  vdc          = "{{.Vdc}}"

  name         = "{{.IpSetName}}"
  description  = "test-ip-set-description"
  ip_addresses = ["192.168.1.1","192.168.2.1"]
}

data "vcd_org_vdc" "my-vdc" {
  org  = "{{.Org}}"
  name = "{{.Vdc}}"
}

data "vcd_vapp_vm" "vm1" {
    vdc       = data.vcd_org_vdc.my-vdc.name
	vapp_name = "{{.VappName}}"
    name      = "{{.VmName}}"
}

data "vcd_network_routed" "net-r" {
    vdc  = data.vcd_org_vdc.my-vdc.name
 	name = "{{.RoutedNetwork}}"
}

data "vcd_network_direct" "net-d" {
    vdc  = data.vcd_org_vdc.my-vdc.name
 	name = "{{.DirectNetwork}}"
}

data "vcd_network_isolated" "net-i" {
    vdc  = data.vcd_org_vdc.my-vdc.name
 	name = "{{.IsolatedNetwork}}"
}

data "vcd_edgegateway" "edge" {
  vdc  = data.vcd_org_vdc.my-vdc.name
  name = "{{.EdgeName}}"
}

data "vcd_nsxv_application_finder" "found_applications" {
  vdc_id            = data.vcd_org_vdc.my-vdc.id
  search_expression = "^POP3$"
  case_sensitive    = true
  type              = "application"
}

data "vcd_nsxv_application" "application1" {
  vdc_id = data.vcd_org_vdc.my-vdc.id
  name   = tolist(data.vcd_nsxv_application_finder.found_applications.objects)[0].name
}

data "vcd_nsxv_application_group" "application_group1" {
  vdc_id = data.vcd_org_vdc.my-vdc.id
  name   = "MS Exchange 2010 Mailbox Servers"
}

resource "vcd_nsxv_distributed_firewall" "dfw1" {
  vdc_id  = data.vcd_org_vdc.my-vdc.id
  enabled = true

  rule {
	name      = "third"
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
  rule {
    name      = "second"
    direction = "inout"
    action    = "allow"
    logged    = true
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
  rule {
    name      = "first"
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
