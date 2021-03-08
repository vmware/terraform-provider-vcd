// +build functional network extnetwork  ALL

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

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	vcdClient := createTemporaryVCDConnection()
	if vcdClient.Client.APIVCDMaxVersionIs("< 33.0") {
		t.Skip(t.Name() + " requires at least API v33.0 (vCD 10+)")
	}

	var params = StringMap{
		"ExistingExternalNetwork": testConfig.Networking.ExternalNetwork,
		"Tags":                    "network extnetwork",
	}

	configText := templateFill(externalNetworkV2Datasource, params)

	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

	datasourceName := "data.vcd_external_network_v2.ext-net-nsxv"
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(datasourceName, "id", regexp.MustCompile(`^urn:vcloud:network:.*`)),
					resource.TestCheckResourceAttrSet(datasourceName, "vsphere_network.0.portgroup_id"),
					resource.TestCheckResourceAttrSet(datasourceName, "vsphere_network.0.vcenter_id"),
					resource.TestCheckResourceAttr(datasourceName, "nsxt_network.#", "0"),
					resource.TestCheckResourceAttr(datasourceName, "vsphere_network.#", "1"),
					// Cannot be too explicit because this test depends on existing external network and it may have
					// wide configuration.
					resourceFieldIntNotEqual(datasourceName, "ip_scope.#", 0),
					resource.TestCheckResourceAttrSet(datasourceName, "vsphere_network.0.portgroup_id"),
					resource.TestMatchResourceAttr(datasourceName, "vsphere_network.0.vcenter_id", regexp.MustCompile(`^urn:vcloud:vimserver:.*`)),
					resource.TestCheckResourceAttr(datasourceName, "nsxt_network.#", "0"),
				),
			},
		},
	})
	postTestChecks(t)
}

const externalNetworkV2Datasource = `
data "vcd_external_network_v2" "ext-net-nsxv" {
	name = "{{.ExistingExternalNetwork}}"
}
`
