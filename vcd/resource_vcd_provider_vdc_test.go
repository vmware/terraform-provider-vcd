//go:build ALL || functional

package vcd

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcdResourceProviderVdc(t *testing.T) {
	// Pre-checks
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	providerVdcName := t.Name()
	providerVdcDescription := t.Name() + "description"
	// Test configuration
	var params = StringMap{
		"ProviderVdcName":        providerVdcName,
		"ProviderVdcDescription": providerVdcDescription,
		"ResourcePool1":          testConfig.VSphere.ResourcePoolForVcd1,
		"NsxtManager":            testConfig.Nsxt.Manager,
		"NsxtNetworkPool":        testConfig.VCD.NsxtProviderVdc.NetworkPool,
		"StorageProfile":         testConfig.VCD.ProviderVdc.StorageProfile,
		"Vcenter":                testConfig.Networking.Vcenter,
	}
	testParamsNotEmpty(t, params)
	configText := templateFill(testAccVcdResourceProviderVdc, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resourceDef := "vcd_provider_vdc.pvdc1"
	// Test cases
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceDef, "name", providerVdcName),
					resource.TestCheckResourceAttr(resourceDef, "description", providerVdcDescription),
					resource.TestMatchResourceAttr(resourceDef, "id", getProviderVdcDatasourceAttributeUrnRegex("providervdc")),
					resource.TestCheckResourceAttr(resourceDef, "is_enabled", "true"),
					resource.TestCheckResourceAttr(resourceDef, "status", "1"),
					resource.TestMatchResourceAttr(resourceDef, "nsxt_manager_id", getProviderVdcDatasourceAttributeUrnRegex("nsxtmanager")),
					resource.TestMatchResourceAttr(resourceDef, "highest_supported_hardware_version", regexp.MustCompile(`vmx-[\d]+`)),
					resource.TestCheckResourceAttr(resourceDef, "compute_provider_scope", testConfig.Networking.Vcenter),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdResourceProviderVdc = `
data "vcd_vcenter" "vcenter1" {
  name = "{{.Vcenter}}"
}

data "vcd_resource_pool" "rp1" {
  name       = "{{.ResourcePool1}}"
  vcenter_id = data.vcd_vcenter.vcenter1.id 
}

data "vcd_nsxt_manager" "mgr1" {
  name = "{{.NsxtManager}}"
}

data "vcd_network_pool" "np1" {
  name = "{{.NsxtNetworkPool}}"
}

resource "vcd_provider_vdc" "pvdc1" {
  name                               = "{{.ProviderVdcName}}"
  description                        = "{{.ProviderVdcDescription}}"
  is_enabled                         = true
  vcenter_id                         = data.vcd_vcenter.vcenter1.id
  nsxt_manager_id                    = data.vcd_nsxt_manager.mgr1.id
  network_pool_ids                   = [data.vcd_network_pool.np1.id]
  resource_pool_ids                  = [data.vcd_resource_pool.rp1.id]
  storage_profile_names              = ["{{.StorageProfile}}"]
  highest_supported_hardware_version = data.vcd_resource_pool.rp1.hardware_version
}
`
