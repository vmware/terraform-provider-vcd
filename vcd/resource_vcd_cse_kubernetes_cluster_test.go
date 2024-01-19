//go:build cse || ALL

package vcd

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcdCseKubernetesCluster(t *testing.T) {
	preTestChecks(t)

	if cse := os.Getenv("TEST_VCD_CSE"); cse == "" {
		t.Skip("CSE tests deactivated, skipping " + t.Name())
	}

	tokenFilename := getCurrentDir() + t.Name() + ".json"
	defer func() {
		// Clean the API Token file
		if fileExists(tokenFilename) {
			err := os.Remove(tokenFilename)
			if err != nil {
				fmt.Printf("could not delete API token file '%s', please delete it manually", tokenFilename)
			}
		}
	}()

	now := time.Now()
	var params = StringMap{
		"Name":         strings.ToLower(t.Name()),
		"OvaCatalog":   testConfig.Cse.OvaCatalog,
		"OvaName":      testConfig.Cse.OvaName,
		"SolutionsOrg": testConfig.Cse.SolutionsOrg,
		"TenantOrg":    testConfig.Cse.TenantOrg,
		"Vdc":          testConfig.Cse.Vdc,
		"EdgeGateway":  testConfig.Cse.EdgeGateway,
		"Network":      testConfig.Cse.RoutedNetwork,
		"TokenName":    fmt.Sprintf("%s%d%d%d", strings.ToLower(t.Name()), now.Day(), now.Hour(), now.Minute()),
		"TokenFile":    tokenFilename,
	}
	testParamsNotEmpty(t, params)

	configText := templateFill(testAccVcdCseKubernetesCluster, params)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	cacheId := testCachedFieldValue{}
	clusterName := "vcd_cse_kubernetes_cluster.my_cluster"
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy: func(state *terraform.State) error {
			if cacheId.fieldValue == "" {
				return fmt.Errorf("cached ID '%s' is empty", cacheId.fieldValue)
			}
			conn := testAccProvider.Meta().(*VCDClient)
			_, err := conn.GetRdeById(cacheId.fieldValue)
			if err == nil {
				return fmt.Errorf("cluster with ID '%s' still exists", cacheId.fieldValue)
			}
			return nil
		},
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check: resource.ComposeAggregateTestCheckFunc(
					cacheId.cacheTestResourceFieldValue(clusterName, "id"),
					resource.TestCheckResourceAttrSet(clusterName, "id"),
					resource.TestCheckResourceAttr(clusterName, "name", strings.ToLower(t.Name())),
					resource.TestCheckResourceAttr(clusterName, "state", "provisioned"),
					resource.TestCheckResourceAttrSet(clusterName, "kubeconfig"),
					resource.TestCheckResourceAttrSet(clusterName, "raw_cluster_rde_json"),
				),
			},
		},
	})
	postTestChecks(t)
}

// TODO: Test:
// Basic (DONE)
// With machine health checks
// With machine health checks
// Without storage class
// With virtual IP and control plane IPs
// Nodes With vGPU policies
const testAccVcdCseKubernetesCluster = `
# skip-binary-test - This one requires a very special setup

data "vcd_catalog" "tkg_catalog" {
  org  = "{{.SolutionsOrg}}"
  name = "{{.OvaCatalog}}"
}

data "vcd_catalog_vapp_template" "tkg_ova" {
  org        = data.vcd_catalog.tkg_catalog.org
  catalog_id = data.vcd_catalog.tkg_catalog.id
  name       = "{{.OvaName}}"
}

data "vcd_org_vdc" "vdc" {
  org  = "{{.TenantOrg}}"
  name = "{{.Vdc}}"
}

data "vcd_nsxt_edgegateway" "egw" {
  org      = data.vcd_org_vdc.vdc.org
  owner_id = data.vcd_org_vdc.vdc.id
  name     = "{{.EdgeGateway}}"
}

data "vcd_network_routed_v2" "routed" {
  org             = data.vcd_nsxt_edgegateway.egw.org
  edge_gateway_id = data.vcd_nsxt_edgegateway.egw.id
  name            = "{{.Network}}"
}

data "vcd_vm_sizing_policy" "tkg_small" {
  name = "TKG small"
}

data "vcd_storage_profile" "sp" {
  org  = data.vcd_org_vdc.vdc.org
  vdc  = data.vcd_org_vdc.vdc.name
  name = "*"
}

resource "vcd_api_token" "token" {
  name             = "{{.TokenName}}"
  file_name        = "{{.TokenFile}}"
  allow_token_file = true
}

resource "vcd_cse_kubernetes_cluster" "my_cluster" {
  cse_version        = "4.2"
  runtime            = "tkg"
  name               = "{{.Name}}"
  ova_id             = data.vcd_catalog_vapp_template.tkg_ova.id
  org                = data.vcd_org_vdc.vdc.org
  vdc_id             = data.vcd_org_vdc.vdc.id
  network_id         = data.vcd_network_routed_v2.routed.id
  api_token_file	 = vcd_api_token.token.file_name

  control_plane {
    machine_count      = 1
    disk_size_gi       = 20
    sizing_policy_id   = data.vcd_vm_sizing_policy.tkg_small.id
    storage_profile_id = data.vcd_storage_profile.sp.id
  }

  node_pool {
    name               = "node-pool-1"
    machine_count      = 1
    disk_size_gi       = 20
    sizing_policy_id   = data.vcd_vm_sizing_policy.tkg_small.id
    storage_profile_id = data.vcd_storage_profile.sp.id
  }

  default_storage_class {
	name               = "sc-1"
	storage_profile_id = data.vcd_storage_profile.sp.id
    reclaim_policy     = "delete"
    filesystem         = "ext4"
  }

  auto_repair_on_errors = true
  node_health_check     = true

  operations_timeout_minutes = 0
}
`

// Test_getTkgVersionBundleFromVAppTemplateName requires connectivity with GitHub, as it fetches the 'tkg_versions.json' file.
// This tests asserts that getTkgVersionBundleFromVAppTemplateName works correctly, retrieving the correct TKG versions from that file.
func Test_getTkgVersionBundleFromVAppTemplateName(t *testing.T) {
	vcdClient := createSystemTemporaryVCDConnection()
	tests := []struct {
		name    string
		ovaName string
		want    tkgVersionBundle
		wantErr string
	}{
		{
			name:    "wrong ova name",
			ovaName: "randomOVA",
			want:    tkgVersionBundle{},
			wantErr: "the vApp Template 'randomOVA' is not a Kubernetes template OVA",
		},
		{
			name:    "not supported ova",
			ovaName: "ubuntu-2004-kube-v9.99.9+vmware.9-tkg.9-b8c57a6c8c98d227f74e7b1a9eef27st",
			want:    tkgVersionBundle{},
			wantErr: "the Kubernetes OVA 'v9.99.9+vmware.9-tkg.9-b8c57a6c8c98d227f74e7b1a9eef27st' is not supported",
		},
		{
			name:    "not supported photon ova",
			ovaName: "photon-3-kube-v1.27.5+vmware.1-tkg.1-cac282289bb29b217b808a2b9b0c0c46",
			want:    tkgVersionBundle{},
			wantErr: "the vApp Template 'photon-3-kube-v1.27.5+vmware.1-tkg.1-cac282289bb29b217b808a2b9b0c0c46' uses Photon, and it is not supported",
		},
		{
			name:    "supported ova",
			ovaName: "ubuntu-2004-kube-v1.26.8+vmware.1-tkg.1-0edd4dafbefbdb503f64d5472e500cf8",
			want: tkgVersionBundle{
				EtcdVersion:       "v3.5.6_vmware.20",
				CoreDnsVersion:    "v1.9.3_vmware.16",
				TkgVersion:        "v2.3.1",
				TkrVersion:        "v1.26.8---vmware.1-tkg.1",
				KubernetesVersion: "v1.26.8+vmware.1",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getTkgVersionBundleFromVAppTemplateName(vcdClient, tt.ovaName)
			if err != nil {
				if tt.wantErr == "" {
					t.Fatalf("getTkgVersionBundleFromVAppTemplateName() got error = %v, but should have not failed", err)
				}
				if err.Error() != tt.wantErr {
					t.Fatalf("getTkgVersionBundleFromVAppTemplateName() error = %v, wantErr = %v", err, tt.wantErr)
				}
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("getTkgVersionBundleFromVAppTemplateName() got = %v, want = %v", got, tt.want)
			}
		})
	}
}
