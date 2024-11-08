//go:build api || functional || catalog || vapp || network || extnetwork || org || query || vm || vdc || gateway || disk || binary || lb || lbServiceMonitor || lbServerPool || lbAppProfile || lbAppRule || lbVirtualServer || access_control || user || standaloneVm || search || auth || nsxt || role || alb || certificate || vdcGroup || ldap || rde || uiPlugin || providerVdc || cse || slz || multisite || tm || ALL

package vcd

import (
	"github.com/vmware/go-vcloud-director/v3/govcd"
	"testing"
)

func getVCenterHcl(t *testing.T) (string, string) {
	vcdClient := createTemporaryVCDConnection(false)
	vc, err := vcdClient.GetVCenterByUrl(testConfig.Tm.VcenterUrl)
	if err == nil {
		return `
data "vcd_vcenter" "vc" {
  name = "` + vc.VSphereVCenter.Name + `"
}
`, "data.vcd_vcenter.vc"
	}
	if !govcd.ContainsNotFound(err) {
		t.Fatal(err)
		return "", ""
	}
	if !testConfig.Tm.CreateVcenter {
		t.Skip("vCenter is not configured and configuration is not allowed in config file")
		return "", ""
	}
	return `
resource "vcd_vcenter" "vc" {
  name                     = "` + t.Name() + `"
  url                      = "` + testConfig.Tm.VcenterUrl + `"
  auto_trust_certificate   = true
  refresh_vcenter_on_read  = true
  refresh_policies_on_read = true
  username                 = "` + testConfig.Tm.VcenterUsername + `"
  password                 = "` + testConfig.Tm.VcenterPassword + `"
  is_enabled               = true
}
`, "vcd_vcenter.vc"
}

func getNsxManagerHcl(t *testing.T) (string, string) {
	vcdClient := createTemporaryVCDConnection(false)
	nsxtManager, err := vcdClient.GetNsxtManagerOpenApiByUrl(testConfig.Tm.NsxtManagerUrl)
	if err == nil {
		return `
data "vcd_nsxt_manager" "nsx_manager" {
  name = "` + nsxtManager.NsxtManagerOpenApi.Name + `"
}
`, "data.vcd_nsxt_manager.nsx_manager"
	}
	if !govcd.ContainsNotFound(err) {
		t.Fatal(err)
		return "", ""
	}
	if !testConfig.Tm.CreateNsxtManager {
		t.Skip("NSX Manager is not configured and configuration is not allowed in config file")
		return "", ""
	}
	return `
resource "vcd_nsxt_manager" "nsx_manager" {
  name                   = "` + t.Name() + `"
  description            = "` + t.Name() + `"
  username               = "` + testConfig.Tm.NsxtManagerUsername + `"
  password               = "` + testConfig.Tm.NsxtManagerPassword + `"
  url                    = "` + testConfig.Tm.NsxtManagerUrl + `"
  network_provider_scope = ""
  auto_trust_certificate = true
}

`, "vcd_nsxt_manager.nsx_manager"
}

func getRegionHcl(t *testing.T, vCenterName, nsxManagerName string) (string, string) {
	vcdClient := createTemporaryVCDConnection(false)
	region, err := vcdClient.GetRegionByName(testConfig.Tm.Region)
	if err == nil {
		return `
data "vcd_tm_region" "region" {
  name = "` + region.Region.Name + `"
}
`, "data.vcd_tm_region.region"
	}
	if !govcd.ContainsNotFound(err) {
		t.Fatal(err)
		return "", ""
	}
	if !testConfig.Tm.CreateRegion {
		t.Skip("Region is not configured and configuration is not allowed in config file")
		return "", ""
	}
	return `
data "vcd_tm_supervisor" "supervisor" {
  name       = "` + testConfig.Tm.VcenterSupervisor + `"
  vcenter_id = ` + vCenterName + `.id
  depends_on = [` + vCenterName + `]
}

resource "vcd_tm_region" "region" {
  name                 = "` + t.Name() + `"
  description          = "` + t.Name() + `"
  is_enabled           = true
  nsx_manager_id       = ` + nsxManagerName + `.id
  supervisor_ids       = [data.vcd_tm_supervisor.supervisor.id]
  storage_policy_names = ["` + testConfig.Tm.VcenterStorageProfile + `"]
}
`, "vcd_tm_region.region"
}
