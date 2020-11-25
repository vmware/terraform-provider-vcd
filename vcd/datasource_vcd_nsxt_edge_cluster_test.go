// +build functional  ALL

package vcd

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcdNsxtEdgeCluster(t *testing.T) {
	if !usingSysAdmin() {
		t.Skip(t.Name() + " requires system admin privileges")
		return
	}

	vcdClient := createTemporaryVCDConnection()
	if vcdClient.Client.APIVCDMaxVersionIs("< 33.0") {
		t.Skip(t.Name() + " requires at least API v33.0 (vCD 10+)")
	}

	_, vdc, err := vcdClient.GetOrgAndVdc(testConfig.VCD.Org, testConfig.Nsxt.Vdc)
	if err != nil {
		t.Errorf("error retrieving Org And Vdc: %s", err)
	}

	edgeClusters, err := vdc.GetAllNsxtEdgeClusters(nil)
	if len(edgeClusters) < 1 {
		t.Errorf("no edge clusters found")
	}

	var params = StringMap{
		"NsxtVdc":             testConfig.Nsxt.Vdc,
		"ExistingEdgeCluster": edgeClusters[0].NsxtEdgeCluster.Name,
		"Tags":                "network",
	}

	configText := templateFill(nsxtEdgeClusterDatasource, params)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

	datasourceName := "data.vcd_nsxt_edge_cluster.ec"
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Check if result is UUID (e.g. 6c188839-ba06-4ceb-8255-2622fe69ce7c)
					resource.TestMatchResourceAttr(datasourceName, "id", regexp.MustCompile(`^[a-zA-Z0-9-_]{36}$`)),
					resource.TestCheckResourceAttr(datasourceName, "name", edgeClusters[0].NsxtEdgeCluster.Name),
					// resource.TestCheckResourceAttrSet(datasourceName, "description"),
					resource.TestCheckResourceAttrSet(datasourceName, "node_count"),
					resource.TestCheckResourceAttrSet(datasourceName, "node_type"),
					resource.TestCheckResourceAttrSet(datasourceName, "deployment_type"),
				),
			},
		},
	})
}

const nsxtEdgeClusterDatasource = `
data "vcd_nsxt_edge_cluster" "ec" {
	vdc  = "{{.NsxtVdc}}"
	name = "{{.ExistingEdgeCluster}}"
}
`
