// +build ALL functional

package vcd

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

type listDef struct {
	name         string
	resourceType string
	parent       string
}

func TestAccVcdDatasourceResourceList(t *testing.T) {

	var lists = []listDef{
		{"resources", "resources", ""},

		// entities belonging to an Org don't require an explicit parent, as it is given from the Org passed in the provider
		// For each resource, we test with and without and explicit parent
		{"orgs", "vcd_org", ""},
		{"extnet", "vcd_external_network", ""},
		{"extnet-parent", "vcd_external_network", testConfig.VCD.Org},
		{"user", "vcd_org_user", ""},
		{"user-parent", "vcd_org_user", testConfig.VCD.Org},
		{"catalog", "vcd_catalog", ""},
		{"catalog-parent", "vcd_catalog", testConfig.VCD.Org},
		{"VDC", "vcd_org_vdc", ""},
		{"VDC-parent", "vcd_org_vdc", testConfig.VCD.Org},

		// entities belonging to a VDC don't require an explicit parent, as it is given from the VDC passed in the provider
		// For each resource, we test with and without and explicit parent
		{"edge_gateway", "vcd_edgegateway", ""},
		{"edge_gateway-parent", "vcd_edgegateway", testConfig.VCD.Vdc},
		{"network", "network", ""},
		{"network-parent", "network", testConfig.VCD.Vdc},
		{"network_isolated", "vcd_network_isolated", ""},
		{"network_isolated-parent", "vcd_network_isolated", testConfig.VCD.Vdc},
		{"network_routed", "vcd_network_routed", ""},
		{"network_routed-parent", "vcd_network_routed", testConfig.VCD.Vdc},
		{"network_direct", "vcd_network_direct", ""},
		{"network_direct-parent", "vcd_network_direct", testConfig.VCD.Vdc},
		{"vapp", "vcd_vapp", ""},
		{"vapp-parent", "vcd_vapp", testConfig.VCD.Vdc},

		// test for VM requires a VApp as parent, which may not be guaranteed, as there is none in the config file
		//{"vapp_vm", "vcd_vapp_vm", "TestVapp"},

		{"catalog_item", "vcd_catalog_item", testConfig.VCD.Catalog.Name},
		{"nsxv_dnat", "vcd_nsxv_dnat", testConfig.Networking.EdgeGateway},
		{"nsxv_snat", "vcd_nsxv_snat", testConfig.Networking.EdgeGateway},
		{"nsxv_firewall_rule", "vcd_nsxv_firewall_rule", testConfig.Networking.EdgeGateway},
		{"lb_server_pool", "vcd_lb_server_pool", testConfig.Networking.EdgeGateway},
		{"lb_service_monitor", "vcd_lb_service_monitor", testConfig.Networking.EdgeGateway},
		{"lb_virtual_server", "vcd_lb_virtual_server", testConfig.Networking.EdgeGateway},
		{"lb_app_profile", "vcd_lb_app_profile", testConfig.Networking.EdgeGateway},
		{"lb_app_rule", "vcd_lb_app_rule", testConfig.Networking.EdgeGateway},
	}
	for _, def := range lists {
		t.Run(def.name+"-"+def.resourceType, func(t *testing.T) { runResourceInfoTest(def, t) })
	}

}

func runResourceInfoTest(def listDef, t *testing.T) {

	var data = StringMap{
		"ResName":   def.name,
		"ResType":   def.resourceType,
		"ResParent": def.parent,
		"FuncName":  fmt.Sprintf("ResourceList-%s", def.name+"-"+def.resourceType),
	}
	var configText string
	if def.parent == "" {
		configText = templateFill(testAccCheckVcdDatasourceInfoSimple, data)
	} else {
		configText = templateFill(testAccCheckVcdDatasourceInfoWithParent, data)
	}

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"data.vcd_resource_list."+def.name, "name", def.name),
					//resource.TestCheckResourceAttr( "data.vcd_resource_list.orgs", "list.0", "System"),
				),
			},
		},
	})
}

const testAccCheckVcdDatasourceInfoSimple = `
data "vcd_resource_list" "{{.ResName}}" {
  name          = "{{.ResName}}"
  resource_type = "{{.ResType}}"
}

output "resources" {
  value = data.vcd_resource_list.{{.ResName}}
}
`
const testAccCheckVcdDatasourceInfoWithParent = `
data "vcd_resource_list" "{{.ResName}}" {
  name          = "{{.ResName}}"
  resource_type = "{{.ResType}}"
  parent        = "{{.ResParent}}"
}

output "resources" {
  value = data.vcd_resource_list.{{.ResName}}
}
`
