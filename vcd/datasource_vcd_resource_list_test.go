//go:build ALL || functional
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
		{"global_role", "vcd_global_role", "", "vApp Author"},
		{"rights_bundle", "vcd_rights_bundle", "", "Default Rights Bundle"},
		{"right", "vcd_right", "", "Catalog: Change Owner"},

		// entities belonging to an Org don't require an explicit parent, as it is given from the Org passed in the provider
		// For each resource, we test with and without and explicit parent
		{"user", "vcd_org_user", "", ""},

		// test for VM requires a VApp as parent, which may not be guaranteed, as there is none in the config file
		//{"vapp_vm", "vcd_vapp_vm", "TestVapp", ""},
	}

	if testConfig.VCD.Org != "" {
		lists = append(lists, listDef{"orgs", "vcd_org", "", testConfig.VCD.Org})

		// entities belonging to an Org don't require an explicit parent, as it is given from the Org passed in the provider
		// For each resource, we test with and without and explicit parent
		lists = append(lists, listDef{"user-parent", "vcd_org_user", testConfig.VCD.Org, ""})
		lists = append(lists, listDef{"role-parent", "vcd_role", testConfig.VCD.Org, "vApp Author"})
		if testConfig.Networking.ExternalNetwork != "" {
			lists = append(lists, listDef{"extnet-parent", "vcd_external_network", testConfig.VCD.Org, testConfig.Networking.ExternalNetwork})
		} else {
			fmt.Print("`testConfig.Networking.ExternalNetwork` value isn't configured, datasource test will be skipped\n")
		}
		if testConfig.VCD.Catalog.Name != "" {
			// entities belonging to an Org don't require an explicit parent, as it is given from the Org passed in the provider
			// For each resource, we test with and without and explicit parent
			lists = append(lists, listDef{"catalog-parent", "vcd_catalog", testConfig.VCD.Org, testConfig.VCD.Catalog.Name})
		} else {
			fmt.Print("`testConfig.VCD.Catalog.Name` value isn't configured, datasource test using this will be skipped\n")
		}
		if testConfig.VCD.Vdc != "" {
			// entities belonging to a VDC don't require an explicit parent, as it is given from the VDC passed in the provider
			// For each resource, we test with and without and explicit parent
			lists = append(lists, listDef{"VDC-parent", "vcd_org_vdc", testConfig.VCD.Org, testConfig.VCD.Vdc})
		} else {
			fmt.Print("`testConfig.VCD.Vdc` value isn't configured, datasource test using this will be skipped\n")
		}
	} else {
		fmt.Print("`testConfig.VCD.Org` value isn't configured, datasource test will be skipped\n")
	}

	if testConfig.Networking.ExternalNetwork != "" {
		// entities belonging to an Org don't require an explicit parent, as it is given from the Org passed in the provider
		// For each resource, we test with and without and explicit parent
		lists = append(lists, listDef{"extnet", "vcd_external_network", "", testConfig.Networking.ExternalNetwork})
	} else {
		fmt.Print("`testConfig.Networking.ExternalNetwork` value isn't configured, datasource test using this will be skipped\n")
	}

	if testConfig.VCD.Catalog.Name != "" {
		// entities belonging to an Org don't require an explicit parent, as it is given from the Org passed in the provider
		// For each resource, we test with and without and explicit parent
		lists = append(lists, listDef{"catalog", "vcd_catalog", "", testConfig.VCD.Catalog.Name})

		if testConfig.VCD.Catalog.CatalogItem != "" {
			// tests in this last group always require an explicit parent
			lists = append(lists, listDef{"catalog_item", "vcd_catalog_item", testConfig.VCD.Catalog.Name, testConfig.VCD.Catalog.CatalogItem})
		} else {
			fmt.Print("`testConfig.VCD.CatalogItem` value isn't configured, datasource test using this will be skipped\n")
		}
		if testConfig.Media.MediaName != "" {
			// tests in this last group always require an explicit parent
			lists = append(lists, listDef{"catalog_media", "vcd_catalog_media", testConfig.VCD.Catalog.Name, testConfig.Media.MediaName})
		} else {
			fmt.Print("`testConfig.Media.MediaName` value isn't configured, datasource test using this will be skipped\n")
		}
	} else {
		fmt.Print("`testConfig.VCD.Catalog.Name` value isn't configured, datasource test using this will be skipped\n")
	}

	if testConfig.VCD.Vdc != "" {
		// entities belonging to an Org don't require an explicit parent, as it is given from the Org passed in the provider
		// For each resource, we test with and without and explicit parent
		lists = append(lists, listDef{"VDC", "vcd_org_vdc", "", testConfig.VCD.Vdc})
		// entities belonging to a VDC don't require an explicit parent, as it is given from the VDC passed in the provider
		// For each resource, we test with and without and explicit parent
		lists = append(lists, listDef{"network-parent", "network", testConfig.VCD.Vdc, ""})
		lists = append(lists, listDef{"network_isolated-parent", "vcd_network_isolated", testConfig.VCD.Vdc, ""})
		lists = append(lists, listDef{"vcd_network_routed_v2", "vcd_network_routed_v2", testConfig.Nsxt.Vdc, ""})
		lists = append(lists, listDef{"vcd_network_isolated_v2", "vcd_network_isolated_v2", testConfig.Nsxt.Vdc, ""})
		lists = append(lists, listDef{"vcd_nsxt_network_imported", "vcd_nsxt_network_imported", testConfig.Nsxt.Vdc, ""})
		lists = append(lists, listDef{"network_routed-parent", "vcd_network_routed", testConfig.VCD.Vdc, ""})
		lists = append(lists, listDef{"vapp-parent", "vcd_vapp", testConfig.VCD.Vdc, ""})
		lists = append(lists, listDef{"network_direct-parent", "vcd_network_direct", testConfig.VCD.Vdc, ""})

		lists = append(lists, listDef{"network", "network", "", ""})
		lists = append(lists, listDef{"network_isolated", "vcd_network_isolated", "", ""})
		lists = append(lists, listDef{"network_routed", "vcd_network_routed", "", ""})
		lists = append(lists, listDef{"network_direct", "vcd_network_direct", "", ""})
		lists = append(lists, listDef{"ipset", "vcd_ipset", "", ""})
		lists = append(lists, listDef{"vapp", "vcd_vapp", "", ""})

		if testConfig.Networking.EdgeGateway != "" {
			// entities belonging to a VDC don't require an explicit parent, as it is given from the VDC passed in the provider
			// For each resource, we test with and without and explicit parent
			lists = append(lists, listDef{"edge_gateway-parent", "vcd_edgegateway", testConfig.VCD.Vdc, testConfig.Networking.EdgeGateway})
		} else {
			fmt.Print("`testConfig.Networking.EdgeGateway` value isn't configured, datasource test using this will be skipped\n")
		}
	} else {
		fmt.Print("`testConfig.VCD.Vdc` value isn't configured, datasource test using this will be skipped\n")
	}

	if testConfig.Networking.EdgeGateway != "" {
		// entities belonging to an Org don't require an explicit parent, as it is given from the Org passed in the provider
		// For each resource, we test with and without and explicit parent
		lists = append(lists, listDef{"edge_gateway", "vcd_edgegateway", "", testConfig.Networking.EdgeGateway})

		// tests in this last group always require an explicit parent
		lists = append(lists, listDef{"nsxv_dnat", "vcd_nsxv_dnat", testConfig.Networking.EdgeGateway, ""})
		lists = append(lists, listDef{"nsxv_snat", "vcd_nsxv_snat", testConfig.Networking.EdgeGateway, ""})
		lists = append(lists, listDef{"nsxv_firewall_rule", "vcd_nsxv_firewall_rule", testConfig.Networking.EdgeGateway, ""})
		lists = append(lists, listDef{"lb_server_pool", "vcd_lb_server_pool", testConfig.Networking.EdgeGateway, ""})
		lists = append(lists, listDef{"lb_service_monitor", "vcd_lb_service_monitor", testConfig.Networking.EdgeGateway, ""})
		lists = append(lists, listDef{"lb_virtual_server", "vcd_lb_virtual_server", testConfig.Networking.EdgeGateway, ""})
		lists = append(lists, listDef{"lb_app_profile", "vcd_lb_app_profile", testConfig.Networking.EdgeGateway, ""})
		lists = append(lists, listDef{"lb_app_rule", "vcd_lb_app_rule", testConfig.Networking.EdgeGateway, ""})

	} else {
		fmt.Print("`testConfig.Networking.EdgeGateway` value isn't configured, datasource test using this will be skipped\n")
	}

	vcdClient := createTemporaryVCDConnection(true)
	if vcdClient != nil && vcdClient.Client.APIVCDMaxVersionIs(">= 35") {
		lists = append(lists,
			listDef{"library_certificate", "vcd_library_certificate", "", ""},
		)
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

	if !usingSysAdmin() && (def.resourceType == "vcd_external_network" ||
		def.resourceType == "vcd_global_role" ||
		def.resourceType == "vcd_rights_bundle") {
		t.Skipf("test with %s requires system administrator privileges", def.resourceType)
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
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
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
