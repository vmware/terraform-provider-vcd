//go:build functional || network || extnetwork || ALL
// +build functional network extnetwork ALL

package vcd

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcdExternalNetworkV2Datasource(t *testing.T) {
	preTestChecks(t)
	if !usingSysAdmin() {
		t.Skip(t.Name() + " requires system admin privileges")
		return
	}

	var params = StringMap{
		"ExistingExternalNetwork": testConfig.Nsxt.ExternalNetwork,
		"Tags":                    "network extnetwork",
	}
	testParamsNotEmpty(t, params)

	configText := templateFill(externalNetworkV2Datasource, params)

	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	datasourceName := "data.vcd_external_network_v2.ext-net-nsxt"
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(datasourceName, "id", regexp.MustCompile(`^urn:vcloud:network:.*`)),
					resource.TestCheckResourceAttrSet(datasourceName, "nsxt_network.0.nsxt_manager_id"),
					resource.TestCheckResourceAttrSet(datasourceName, "nsxt_network.0.nsxt_tier0_router_id"),
					resource.TestCheckResourceAttr(datasourceName, "nsxt_network.#", "1"),
					resourceFieldIntNotEqual(datasourceName, "ip_scope.#", 0),
					resource.TestMatchResourceAttr(datasourceName, "nsxt_network.0.nsxt_manager_id", regexp.MustCompile(`^urn:vcloud:nsxtmanager:.*`)),
				),
			},
		},
	})
	postTestChecks(t)
}

const externalNetworkV2Datasource = `
data "vcd_external_network_v2" "ext-net-nsxt" {
	name = "{{.ExistingExternalNetwork}}"
}
`
