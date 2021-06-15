// +build ALL functional

package vcd

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"testing"
)

type listDef struct {
	name         string
	resourceType string
	parent       string
	knownItem    string
}

func TestAccVcdDatasourceResourceList(t *testing.T) {
	preTestChecks(t)

	var lists = []listDef{
		{"resources", "resources", "", "vcd_org"},
		{"orgs", "vcd_org", "", testConfig.VCD.Org},
		{"global_role", "vcd_global_role", "", "vApp Author"},
		{"rights_bundle", "vcd_rights_bundle", "", "Default Rights Bundle"},
		{"right", "vcd_right", "", "Catalog: Change Owner"},

		// entities belonging to an Org don't require an explicit parent, as it is given from the Org passed in the provider
		// For each resource, we test with and without and explicit parent
		{"extnet", "vcd_external_network", "", testConfig.Networking.ExternalNetwork},
		{"extnet-parent", "vcd_external_network", testConfig.VCD.Org, testConfig.Networking.ExternalNetwork},
		{"user", "vcd_org_user", "", ""},
		{"user-parent", "vcd_org_user", testConfig.VCD.Org, ""},
		{"catalog", "vcd_catalog", "", testConfig.VCD.Catalog.Name},
		{"catalog-parent", "vcd_catalog", testConfig.VCD.Org, testConfig.VCD.Catalog.Name},
		{"VDC", "vcd_org_vdc", "", testConfig.VCD.Vdc},
		{"VDC-parent", "vcd_org_vdc", testConfig.VCD.Org, testConfig.VCD.Vdc},
		{"role-parent", "vcd_role", testConfig.VCD.Org, "vApp Author"},

		// entities belonging to a VDC don't require an explicit parent, as it is given from the VDC passed in the provider
		// For each resource, we test with and without and explicit parent
		{"edge_gateway", "vcd_edgegateway", "", testConfig.Networking.EdgeGateway},
		{"edge_gateway-parent", "vcd_edgegateway", testConfig.VCD.Vdc, testConfig.Networking.EdgeGateway},
		{"network", "network", "", ""},
		{"network-parent", "network", testConfig.VCD.Vdc, ""},
		{"network_isolated", "vcd_network_isolated", "", ""},
		{"network_isolated-parent", "vcd_network_isolated", testConfig.VCD.Vdc, ""},
		{"network_routed", "vcd_network_routed", "", ""},
		{"vcd_network_routed_v2", "vcd_network_routed_v2", testConfig.Nsxt.Vdc, ""},
		{"vcd_network_isolated_v2", "vcd_network_isolated_v2", testConfig.Nsxt.Vdc, ""},
		{"vcd_nsxt_network_imported", "vcd_nsxt_network_imported", testConfig.Nsxt.Vdc, ""},
		{"network_routed-parent", "vcd_network_routed", testConfig.VCD.Vdc, ""},
		{"network_direct", "vcd_network_direct", "", ""},
		{"network_direct-parent", "vcd_network_direct", testConfig.VCD.Vdc, ""},
		{"ipset", "vcd_ipset", "", ""},
		{"vapp", "vcd_vapp", "", ""},
		{"vapp-parent", "vcd_vapp", testConfig.VCD.Vdc, ""},

		// test for VM requires a VApp as parent, which may not be guaranteed, as there is none in the config file
		//{"vapp_vm", "vcd_vapp_vm", "TestVapp", ""},

		// tests in this last group always require an explicit parent
		{"catalog_item", "vcd_catalog_item", testConfig.VCD.Catalog.Name, testConfig.VCD.Catalog.CatalogItem},
		{"catalog_media", "vcd_catalog_media", testConfig.VCD.Catalog.Name, testConfig.Media.MediaName},
		{"nsxv_dnat", "vcd_nsxv_dnat", testConfig.Networking.EdgeGateway, ""},
		{"nsxv_snat", "vcd_nsxv_snat", testConfig.Networking.EdgeGateway, ""},
		{"nsxv_firewall_rule", "vcd_nsxv_firewall_rule", testConfig.Networking.EdgeGateway, ""},
		{"lb_server_pool", "vcd_lb_server_pool", testConfig.Networking.EdgeGateway, ""},
		{"lb_service_monitor", "vcd_lb_service_monitor", testConfig.Networking.EdgeGateway, ""},
		{"lb_virtual_server", "vcd_lb_virtual_server", testConfig.Networking.EdgeGateway, ""},
		{"lb_app_profile", "vcd_lb_app_profile", testConfig.Networking.EdgeGateway, ""},
		{"lb_app_rule", "vcd_lb_app_rule", testConfig.Networking.EdgeGateway, ""},
	}
	for _, def := range lists {
		t.Run(def.name+"-"+def.resourceType, func(t *testing.T) { runResourceInfoTest(def, t) })
	}
	postTestChecks(t)
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

	if !usingSysAdmin() && (def.resourceType == "vcd_external_network") {
		t.Skip("test with external network requires system administrator privileges")
	}
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

	// networks aren't in the configuration file, but we can easily search for existing ones
	if strings.HasPrefix(def.resourceType, "vcd_network") {
		err := getAvailableNetworks()
		if err == nil {
			network, ok := availableNetworks[def.resourceType]
			if ok {
				def.knownItem = network.network.Name
			}
		}
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.vcd_resource_list."+def.name, "name", def.name),
					checkListForKnownItem(def.name, def.knownItem),
				),
			},
		},
	})
}

func checkListForKnownItem(resName, wanted string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if wanted == "" {
			return nil
		}

		resourcePath := "data.vcd_resource_list." + resName

		resource, ok := s.RootModule().Resources[resourcePath]
		if !ok {
			return fmt.Errorf("resource %s not found", resName)
		}

		var list = make([]string, 0)

		for key, value := range resource.Primary.Attributes {

			if strings.HasPrefix(key, "list") {
				list = append(list, value)
			}

		}

		for _, item := range list {
			if item == wanted {
				return nil
			}
		}
		return fmt.Errorf("item '%s' not found in list %s", wanted, resourcePath)
	}
}

const testAccCheckVcdDatasourceInfoSimple = `
data "vcd_resource_list" "{{.ResName}}" {
  name          = "{{.ResName}}"
  resource_type = "{{.ResType}}"
}
`
const testAccCheckVcdDatasourceInfoWithParent = `
data "vcd_resource_list" "{{.ResName}}" {
  name          = "{{.ResName}}"
  resource_type = "{{.ResType}}"
  parent        = "{{.ResParent}}"
}
`
