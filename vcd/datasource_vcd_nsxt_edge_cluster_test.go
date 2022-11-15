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

	if !usingSysAdmin() {
		t.Skip(t.Name() + " requires system admin privileges")
		return
	}

	var params = StringMap{
		"NsxtVdc":             testConfig.Nsxt.Vdc,
		"ExistingEdgeCluster": testConfig.Nsxt.NsxtEdgeCluster,
		"Tags":                "network",
	}
	testParamsNotEmpty(t, params)

	configText := templateFill(nsxtEdgeClusterDatasource, params)

	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	datasourceName := "data.vcd_nsxt_edge_cluster.ec"
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Check if result is UUID (e.g. 6c188839-ba06-4ceb-8255-2622fe69ce7c)
					resource.TestMatchResourceAttr(datasourceName, "id", regexp.MustCompile(`^[a-zA-Z0-9-_]{36}$`)),
					resource.TestCheckResourceAttr(datasourceName, "name", testConfig.Nsxt.NsxtEdgeCluster),
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

func TestAccVcdNsxtEdgeClusterVdcId(t *testing.T) {
	preTestChecks(t)

	if !usingSysAdmin() {
		t.Skip(t.Name() + " requires system admin privileges")
		return
	}

	var params = StringMap{
		"Org":                 testConfig.VCD.Org,
		"NsxtVdc":             testConfig.Nsxt.Vdc,
		"ExistingEdgeCluster": testConfig.Nsxt.NsxtEdgeCluster,
		"Tags":                "network",
	}
	testParamsNotEmpty(t, params)

	configText := templateFill(nsxtEdgeClusterDatasourceVdcId, params)

	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	datasourceName := "data.vcd_nsxt_edge_cluster.ec"
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Check if result is UUID (e.g. 6c188839-ba06-4ceb-8255-2622fe69ce7c)
					resource.TestMatchResourceAttr(datasourceName, "id", regexp.MustCompile(`^[a-zA-Z0-9-_]{36}$`)),
					resource.TestCheckResourceAttr(datasourceName, "name", testConfig.Nsxt.NsxtEdgeCluster),
					resource.TestCheckResourceAttrSet(datasourceName, "node_count"),
					resource.TestCheckResourceAttrSet(datasourceName, "node_type"),
					resource.TestCheckResourceAttrSet(datasourceName, "deployment_type"),
				),
			},
		},
	})
	postTestChecks(t)
}

const nsxtEdgeClusterDatasourceVdcId = `
data "vcd_org_vdc" "existing" {
  org  = "{{.Org}}"
  name = "{{.NsxtVdc}}"
}

data "vcd_nsxt_edge_cluster" "ec" {
  vdc_id  = data.vcd_org_vdc.existing.id
  name    = "{{.ExistingEdgeCluster}}"
}
`

func TestAccVcdNsxtEdgeClusterVdcGroupId(t *testing.T) {
	preTestChecks(t)

	if !usingSysAdmin() {
		t.Skip(t.Name() + " requires system admin privileges")
		return
	}

	var params = StringMap{
		"Org":                 testConfig.VCD.Org,
		"NsxtVdcGroup":        testConfig.Nsxt.VdcGroup,
		"ExistingEdgeCluster": testConfig.Nsxt.NsxtEdgeCluster,
		"Tags":                "network",
	}
	testParamsNotEmpty(t, params)

	configText := templateFill(nsxtEdgeClusterDatasourceVdcGroupId, params)

	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	datasourceName := "data.vcd_nsxt_edge_cluster.ec"
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Check if result is UUID (e.g. 6c188839-ba06-4ceb-8255-2622fe69ce7c)
					resource.TestMatchResourceAttr(datasourceName, "id", regexp.MustCompile(`^[a-zA-Z0-9-_]{36}$`)),
					resource.TestCheckResourceAttr(datasourceName, "name", testConfig.Nsxt.NsxtEdgeCluster),
					resource.TestCheckResourceAttrSet(datasourceName, "node_count"),
					resource.TestCheckResourceAttrSet(datasourceName, "node_type"),
					resource.TestCheckResourceAttrSet(datasourceName, "deployment_type"),
				),
			},
		},
	})
	postTestChecks(t)
}

const nsxtEdgeClusterDatasourceVdcGroupId = `
data "vcd_vdc_group" "existing" {
  org  = "{{.Org}}"
  name = "{{.NsxtVdcGroup}}"
}

data "vcd_nsxt_edge_cluster" "ec" {
  vdc_group_id = data.vcd_vdc_group.existing.id
  name         = "{{.ExistingEdgeCluster}}"
}
`

func TestAccVcdNsxtEdgeClusterProviderVdcId(t *testing.T) {
	preTestChecks(t)

	if !usingSysAdmin() {
		t.Skip(t.Name() + " requires system admin privileges")
		return
	}

	var params = StringMap{
		"ProviderVdc":         testConfig.VCD.NsxtProviderVdc.Name,
		"ExistingEdgeCluster": testConfig.Nsxt.NsxtEdgeCluster,
		"Tags":                "network",
	}
	testParamsNotEmpty(t, params)

	configText := templateFill(nsxtEdgeClusterDatasourceProviderVdcId, params)

	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	datasourceName := "data.vcd_nsxt_edge_cluster.ec"
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Check if result is UUID (e.g. 6c188839-ba06-4ceb-8255-2622fe69ce7c)
					resource.TestMatchResourceAttr(datasourceName, "id", regexp.MustCompile(`^[a-zA-Z0-9-_]{36}$`)),
					resource.TestCheckResourceAttr(datasourceName, "name", testConfig.Nsxt.NsxtEdgeCluster),
					resource.TestCheckResourceAttrSet(datasourceName, "node_count"),
					resource.TestCheckResourceAttrSet(datasourceName, "node_type"),
					resource.TestCheckResourceAttrSet(datasourceName, "deployment_type"),
				),
			},
		},
	})
	postTestChecks(t)
}

const nsxtEdgeClusterDatasourceProviderVdcId = `
data "vcd_provider_vdc" "nsxt-pvdc" {
  name = "{{.ProviderVdc}}"
}

data "vcd_nsxt_edge_cluster" "ec" {
  provider_vdc_id  = data.vcd_provider_vdc.nsxt-pvdc.id
  name             = "{{.ExistingEdgeCluster}}"
}
`
