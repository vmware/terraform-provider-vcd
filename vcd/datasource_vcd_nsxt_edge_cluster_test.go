//go:build functional || nsxt || ALL
// +build functional nsxt ALL

package vcd

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcdNsxtEdgeCluster(t *testing.T) {
	preTestChecks(t)
	skipNoNsxtConfiguration(t)

	if !usingSysAdmin() {
		t.Skip(t.Name() + " requires system admin privileges")
		return
	}

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	vcdClient := createTemporaryVCDConnection(false)

	_, vdc, err := vcdClient.GetOrgAndVdc(testConfig.VCD.Org, testConfig.Nsxt.Vdc)
	if err != nil {
		t.Errorf("error retrieving Org And Vdc: %s", err)
	}

	edgeClusters, err := vdc.GetAllNsxtEdgeClusters(nil)
	if err != nil {
		t.Errorf("got error retrieving Edge Clusters: %s", err)
		t.FailNow()
	}
	if len(edgeClusters) < 1 {
		t.Errorf("no edge clusters found")
		t.FailNow()
	}

	var params = StringMap{
		"NsxtVdc":             testConfig.Nsxt.Vdc,
		"ExistingEdgeCluster": edgeClusters[0].NsxtEdgeCluster.Name,
		"Tags":                "network",
	}

	configText := templateFill(nsxtEdgeClusterDatasource, params)

	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

	datasourceName := "data.vcd_nsxt_edge_cluster.ec"
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testParamsNotEmpty(t, params) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Check if result is UUID (e.g. 6c188839-ba06-4ceb-8255-2622fe69ce7c)
					resource.TestMatchResourceAttr(datasourceName, "id", regexp.MustCompile(`^[a-zA-Z0-9-_]{36}$`)),
					resource.TestCheckResourceAttr(datasourceName, "name", edgeClusters[0].NsxtEdgeCluster.Name),
					resource.TestCheckResourceAttrSet(datasourceName, "node_count"),
					resource.TestCheckResourceAttrSet(datasourceName, "node_type"),
					resource.TestCheckResourceAttrSet(datasourceName, "deployment_type"),
				),
			},
		},
	})
	postTestChecks(t)
}

const nsxtEdgeClusterDatasource = `
data "vcd_nsxt_edge_cluster" "ec" {
	vdc  = "{{.NsxtVdc}}"
	name = "{{.ExistingEdgeCluster}}"
}
`
