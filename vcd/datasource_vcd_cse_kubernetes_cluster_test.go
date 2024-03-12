//go:build ALL || cse || functional

package vcd

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"regexp"
	"testing"
)

// TestAccCseKubernetesClusterDataSourceNotFound checks that the CSE Kubernetes Cluster data source always returns an error
// and substring 'govcd.ErrorEntityNotFound' in it when a cluster is not found.
func TestAccCseKubernetesClusterDataSourceNotFound(t *testing.T) {
	preTestChecks(t)
	requireCseConfig(t, testConfig)

	var params = StringMap{
		"TenantOrg":  testConfig.Cse.TenantOrg,
		"CseVersion": testConfig.Cse.Version,
	}
	testParamsNotEmpty(t, params)

	params["FuncName"] = t.Name() + "1"
	configText1 := templateFill(testAccNotExistingCluster1, params)
	debugPrintf("#[DEBUG] CONFIGURATION 1: %s", configText1)

	params["FuncName"] = t.Name() + "2"
	configText2 := templateFill(testAccNotExistingCluster2, params)
	debugPrintf("#[DEBUG] CONFIGURATION 2: %s", configText2)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      configText1,
				ExpectError: regexp.MustCompile(`.*` + regexp.QuoteMeta(govcd.ErrorEntityNotFound.Error()) + `.*`),
			},
			{
				Config:      configText2,
				ExpectError: regexp.MustCompile(`.*` + regexp.QuoteMeta(govcd.ErrorEntityNotFound.Error()) + `.*`),
			},
		},
	})
	postTestChecks(t)
}

const testAccNotExistingCluster1 = `
# skip-binary-test: data-source-not-found test only works in acceptance tests

data vcd_cse_kubernetes_cluster "not-existing1" {
  cluster_id = "urn:vcloud:entity:vmware:capvcdCluster:00000000-0000-0000-0000-000000000000"
}
`

const testAccNotExistingCluster2 = `
# skip-binary-test: data-source-not-found test only works in acceptance tests

data "vcd_org" "tenant_org" {
  name = "{{.TenantOrg}}"
}

data vcd_cse_kubernetes_cluster "not-existing2" {
  org_id      = data.vcd_org.tenant_org.id
  cse_version = "{{.CseVersion}}"
  name        = "i-dont-exist"
}
`
