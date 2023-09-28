//go:build vdc || ALL || functional

package vcd

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcdOrgVdcNsxtDefaultSegmentProfileTemplates(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	var params = StringMap{
		"VdcName":                   t.Name(),
		"OrgName":                   testConfig.VCD.Org,
		"ProviderVdc":               testConfig.VCD.NsxtProviderVdc.Name,
		"NetworkPool":               testConfig.VCD.NsxtProviderVdc.NetworkPool,
		"ProviderVdcStorageProfile": testConfig.VCD.NsxtProviderVdc.StorageProfile,
		"EdgeCluster":               testConfig.Nsxt.NsxtEdgeCluster,

		"TestName":                   t.Name(),
		"NsxtManager":                testConfig.Nsxt.Manager,
		"IpDiscoveryProfileName":     testConfig.Nsxt.IpDiscoveryProfile,
		"MacDiscoveryProfileName":    testConfig.Nsxt.MacDiscoveryProfile,
		"QosProfileName":             testConfig.Nsxt.QosProfile,
		"SpoofGuardProfileName":      testConfig.Nsxt.SpoofGuardProfile,
		"SegmentSecurityProfileName": testConfig.Nsxt.SegmentSecurityProfile,

		"Tags": "vdc",
	}
	testParamsNotEmpty(t, params)

	configText1 := templateFill(testAccVcdOrgVdcNsxtDefaultSegmentProfileTemplate, params)
	params["FuncName"] = t.Name() + "-step2DS"
	configText2 := templateFill(testAccVcdOrgVdcNsxtDefaultSegmentProfileTemplateDataSource, params)

	params["FuncName"] = t.Name() + "-Update"
	configText3 := templateFill(testAccVcdOrgVdcNsxtDefaultSegmentProfileTemplate_update, params)

	debugPrintf("#[DEBUG] CONFIGURATION - Step1: %s", configText1)
	debugPrintf("#[DEBUG] CONFIGURATION - Step2: %s", configText2)
	debugPrintf("#[DEBUG] CONFIGURATION - Step3: %s", configText3)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckVdcDestroy,
		),
		Steps: []resource.TestStep{
			{
				Config: configText1,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdVdcExists("vcd_org_vdc.with-spt"),
					resource.TestCheckResourceAttr("vcd_org_vdc.with-spt", "name", t.Name()),
					resource.TestCheckResourceAttr("vcd_org_vdc.with-spt", "org", testConfig.VCD.Org),
					resource.TestCheckResourceAttr("vcd_org_vdc.with-spt", "network_pool_name", testConfig.VCD.NsxtProviderVdc.NetworkPool),
					resource.TestCheckResourceAttr("vcd_org_vdc.with-spt", "provider_vdc_name", testConfig.VCD.NsxtProviderVdc.Name),
					resource.TestCheckResourceAttr("vcd_org_vdc.with-spt", "enabled", "true"),
					resource.TestCheckResourceAttrSet("vcd_org_vdc.with-spt", "vdc_networks_default_segment_profile_template_id"),
					resource.TestCheckResourceAttrSet("vcd_org_vdc.with-spt", "vapp_networks_default_segment_profile_template_id"),
				),
			},
			{
				Config: configText2,
				Check: resource.ComposeTestCheckFunc(
					resourceFieldsEqual("vcd_org_vdc.with-spt", "data.vcd_org_vdc.ds", []string{"delete_recursive", "delete_force", "%"}),
				),
			},
			{
				Config: configText3,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("vcd_org_vdc.with-spt", "edge_cluster_id", ""),
					resource.TestCheckResourceAttr("vcd_org_vdc.with-spt", "vdc_networks_default_segment_profile_template_id", ""),
					resource.TestCheckResourceAttr("vcd_org_vdc.with-spt", "vapp_networks_default_segment_profile_template_id", ""),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdOrgVdcNsxtDefaultSegmentProfileTemplateCommon = `
data "vcd_nsxt_manager" "nsxt" {
  name = "{{.NsxtManager}}"
}

data "vcd_nsxt_segment_ip_discovery_profile" "first" {
  name            = "{{.IpDiscoveryProfileName}}"
  nsxt_manager_id = data.vcd_nsxt_manager.nsxt.id
}

data "vcd_nsxt_segment_mac_discovery_profile" "first" {
  name            = "{{.MacDiscoveryProfileName}}"
  nsxt_manager_id = data.vcd_nsxt_manager.nsxt.id
}

data "vcd_nsxt_segment_spoof_guard_profile" "first" {
  name            = "{{.SpoofGuardProfileName}}"
  nsxt_manager_id = data.vcd_nsxt_manager.nsxt.id
}

data "vcd_nsxt_segment_qos_profile" "first" {
  name            = "{{.QosProfileName}}"
  nsxt_manager_id = data.vcd_nsxt_manager.nsxt.id
}

data "vcd_nsxt_segment_security_profile" "first" {
  name            = "{{.SegmentSecurityProfileName}}"
  nsxt_manager_id = data.vcd_nsxt_manager.nsxt.id
}

resource "vcd_nsxt_segment_profile_template" "complete" {
  name        = "{{.TestName}}-complete"
  description = "description"

  nsxt_manager_id             = data.vcd_nsxt_manager.nsxt.id
  ip_discovery_profile_id     = data.vcd_nsxt_segment_ip_discovery_profile.first.id
  mac_discovery_profile_id    = data.vcd_nsxt_segment_mac_discovery_profile.first.id
  spoof_guard_profile_id      = data.vcd_nsxt_segment_spoof_guard_profile.first.id
  qos_profile_id              = data.vcd_nsxt_segment_qos_profile.first.id
  segment_security_profile_id = data.vcd_nsxt_segment_security_profile.first.id
}

data "vcd_provider_vdc" "pvdc" {
  name = "{{.ProviderVdc}}"
}
`

const testAccVcdOrgVdcNsxtDefaultSegmentProfileTemplateDS = `
# skip-binary-test: resource and data source cannot refer itself in a single file
data "vcd_org_vdc" "ds" {
  org  = "{{.OrgName}}"
  name = "{{.VdcName}}"
}
`

const testAccVcdOrgVdcNsxtDefaultSegmentProfileTemplate = testAccVcdOrgVdcNsxtDefaultSegmentProfileTemplateCommon + `
resource "vcd_org_vdc" "with-spt" {
  name = "{{.VdcName}}"
  org  = "{{.OrgName}}"

  allocation_model  = "ReservationPool"
  network_pool_name = "{{.NetworkPool}}"
  provider_vdc_name = data.vcd_provider_vdc.pvdc.name

  compute_capacity {
    cpu {
      allocated = 1024
      limit     = 1024
    }

    memory {
      allocated = 1024
      limit     = 1024
    }
  }

  storage_profile {
    name     = "{{.ProviderVdcStorageProfile}}"
    enabled  = true
    limit    = 10240
    default  = true
  }

  vdc_networks_default_segment_profile_template_id  = vcd_nsxt_segment_profile_template.complete.id
  vapp_networks_default_segment_profile_template_id = vcd_nsxt_segment_profile_template.complete.id

  enabled                    = true
  enable_thin_provisioning   = true
  enable_fast_provisioning   = true
  delete_force               = true
  delete_recursive           = true
}
`

const testAccVcdOrgVdcNsxtDefaultSegmentProfileTemplateDataSource = testAccVcdOrgVdcNsxtDefaultSegmentProfileTemplate + testAccVcdOrgVdcNsxtDefaultSegmentProfileTemplateDS

const testAccVcdOrgVdcNsxtDefaultSegmentProfileTemplate_update = testAccVcdOrgVdcNsxtDefaultSegmentProfileTemplateCommon + `
resource "vcd_org_vdc" "with-spt" {
  name = "{{.VdcName}}"
  org  = "{{.OrgName}}"

  allocation_model  = "ReservationPool"
  network_pool_name = "{{.NetworkPool}}"
  provider_vdc_name = data.vcd_provider_vdc.pvdc.name

  compute_capacity {
    cpu {
      allocated = 1024
      limit     = 1024
    }

    memory {
      allocated = 1024
      limit     = 1024
    }
  }

  storage_profile {
    name     = "{{.ProviderVdcStorageProfile}}"
    enabled  = true
    limit    = 10240
    default  = true
  }

  enabled                    = true
  enable_thin_provisioning   = true
  enable_fast_provisioning   = true
  delete_force               = true
  delete_recursive           = true
}
`
