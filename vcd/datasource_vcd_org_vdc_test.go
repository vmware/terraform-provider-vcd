// +build vdc ALL functional

package vcd

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

var vdcName = "TestAccVcdVdcDatasource"

func TestAccVcdVdcDatasource(t *testing.T) {
	validateConfiguration(t)

	var params = StringMap{
		"ExistingVdcName": testConfig.VCD.Vdc,
		"VdcName":         vdcName,
		"OrgName":         testConfig.VCD.Org,
		"FuncName":        vdcName,
	}

	var configText string
	datasourceVdc := "vcd_org_vdc.existingVdc"
	if !usingSysAdmin() {
		params["FuncName"] = t.Name() + "-orgAdmin"
		configText = templateFill(testAccCheckVcdVdcDatasource_orgAdmin, params)

		debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

		if vcdShortTest {
			t.Skip(acceptanceTestsSkipped)
			return
		}

		validateDataSource(t, configText, datasourceVdc)
	} else {
		configText = templateFill(testAccCheckVcdVdcDatasource_basic, params)

		debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

		if vcdShortTest {
			t.Skip(acceptanceTestsSkipped)
			return
		}

		validateResourceAndDataSource(t, configText, datasourceVdc)
	}

}

func validateResourceAndDataSource(t *testing.T, configText string, datasourceVdc string) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { preRunChecks(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVdcDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdVdcExists("vcd_org_vdc."+vdcName),
					resource.TestCheckResourceAttrPair(
						"data."+datasourceVdc, "org", "vcd_org_vdc."+vdcName, "org"),
					resource.TestCheckResourceAttrPair(
						"data."+datasourceVdc, "allocation_model", "vcd_org_vdc."+vdcName, "allocation_model"),
					resource.TestCheckResourceAttrPair(
						"data."+datasourceVdc, "network_pool_name", "vcd_org_vdc."+vdcName, "network_pool_name"),
					resource.TestCheckResourceAttrPair(
						"data."+datasourceVdc, "provider_vdc_name", "vcd_org_vdc."+vdcName, "provider_vdc_name"),
					resource.TestCheckResourceAttrPair(
						"data."+datasourceVdc, "enabled", "vcd_org_vdc."+vdcName, "enabled"),
					resource.TestCheckResourceAttrPair(
						"data."+datasourceVdc, "enable_thin_provisioning", "vcd_org_vdc."+vdcName, "enable_thin_provisioning"),
					resource.TestCheckResourceAttrPair(
						"data."+datasourceVdc, "storage_profile.0.enabled", "vcd_org_vdc."+vdcName, "storage_profile.0.enabled"),
					resource.TestCheckResourceAttrPair(
						"data."+datasourceVdc, "storage_profile.0.default", "vcd_org_vdc."+vdcName, "storage_profile.0.default"),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+vdcName, "metadata.vdc_metadata", "VDC Metadata"),
					resource.TestCheckResourceAttrPair(
						"data."+datasourceVdc, "compute_capacity.0.cpu.0.allocated", "vcd_org_vdc."+vdcName, "compute_capacity.0.cpu.0.allocated"),
					resource.TestCheckResourceAttrPair(
						"data."+datasourceVdc, "compute_capacity.0.cpu.0.limit", "vcd_org_vdc."+vdcName, "compute_capacity.0.cpu.0.limit"),
					resource.TestMatchResourceAttr(
						"data."+datasourceVdc, "compute_capacity.0.cpu.0.reserved", regexp.MustCompile(`^\d+$`)),
					resource.TestMatchResourceAttr(
						"data."+datasourceVdc, "compute_capacity.0.cpu.0.used", regexp.MustCompile(`^\d+$`)),
					resource.TestCheckResourceAttrPair(
						"data."+datasourceVdc, "compute_capacity.0.memory.0.allocated", "vcd_org_vdc."+vdcName, "compute_capacity.0.memory.0.allocated"),
					resource.TestCheckResourceAttrPair(
						"data."+datasourceVdc, "compute_capacity.0.memory.0.limit", "vcd_org_vdc."+vdcName, "compute_capacity.0.memory.0.limit"),
					resource.TestMatchResourceAttr(
						"data."+datasourceVdc, "compute_capacity.0.memory.0.reserved", regexp.MustCompile(`^\d+$`)),
					resource.TestMatchResourceAttr(
						"data."+datasourceVdc, "compute_capacity.0.memory.0.used", regexp.MustCompile(`^\d+$`))),
			},
		},
	})
}

func validateDataSource(t *testing.T, configText string, datasourceVdc string) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { preRunChecks(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVdcDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data."+datasourceVdc, "name", testConfig.VCD.Vdc),
					resource.TestMatchResourceAttr("data."+datasourceVdc, "allocation_model", regexp.MustCompile(`^\S+$`)),
					resource.TestMatchResourceAttr("data."+datasourceVdc, "enabled", regexp.MustCompile(`^\S+$`)),
					resource.TestMatchResourceAttr("data."+datasourceVdc, "storage_profile.0.enabled", regexp.MustCompile(`^\S+$`)),
					resource.TestMatchResourceAttr("data."+datasourceVdc, "storage_profile.0.default", regexp.MustCompile(`^\S+$`)),
					resource.TestMatchResourceAttr("data."+datasourceVdc, "nic_quota", regexp.MustCompile(`^\d+$`)),
					resource.TestMatchResourceAttr("data."+datasourceVdc, "network_quota", regexp.MustCompile(`^\d+$`)),
					resource.TestMatchResourceAttr("data."+datasourceVdc, "vm_quota", regexp.MustCompile(`^\d+$`)),
					resource.TestMatchResourceAttr("data."+datasourceVdc, "enable_vm_discovery", regexp.MustCompile(`^\S+$`)),
					resource.TestMatchResourceAttr("data."+datasourceVdc, "compute_capacity.0.cpu.0.allocated", regexp.MustCompile(`^\d+$`)),
					resource.TestMatchResourceAttr("data."+datasourceVdc, "compute_capacity.0.cpu.0.limit", regexp.MustCompile(`^\d+$`)),
					resource.TestMatchResourceAttr("data."+datasourceVdc, "compute_capacity.0.cpu.0.allocated", regexp.MustCompile(`^\d+$`)),
					resource.TestMatchResourceAttr("data."+datasourceVdc, "compute_capacity.0.cpu.0.reserved", regexp.MustCompile(`^\d+$`)),
					resource.TestMatchResourceAttr("data."+datasourceVdc, "compute_capacity.0.memory.0.allocated", regexp.MustCompile(`^\d+$`)),
					resource.TestMatchResourceAttr("data."+datasourceVdc, "compute_capacity.0.memory.0.limit", regexp.MustCompile(`^\d+$`)),
					resource.TestMatchResourceAttr("data."+datasourceVdc, "compute_capacity.0.memory.0.allocated", regexp.MustCompile(`^\d+$`)),
					resource.TestMatchResourceAttr("data."+datasourceVdc, "compute_capacity.0.memory.0.reserved", regexp.MustCompile(`^\d+$`))),
			},
		},
	})
}

const testAccCheckVcdVdcDatasource_basic = `
data "vcd_org_vdc" "existingVdc" {
  org  = "{{.OrgName}}"
  name = "{{.ExistingVdcName}}"
}

resource "vcd_org_vdc" "{{.VdcName}}" {
  name = "{{.VdcName}}"
  org  = "{{.OrgName}}"

  allocation_model  = data.vcd_org_vdc.existingVdc.allocation_model
  network_pool_name = data.vcd_org_vdc.existingVdc.network_pool_name
  provider_vdc_name = data.vcd_org_vdc.existingVdc.provider_vdc_name

  compute_capacity {
    cpu {
     allocated = data.vcd_org_vdc.existingVdc.compute_capacity[0].cpu[0].allocated
     limit     = data.vcd_org_vdc.existingVdc.compute_capacity[0].cpu[0].limit
    }

    memory {
     allocated = data.vcd_org_vdc.existingVdc.compute_capacity[0].memory[0].allocated
     limit     = data.vcd_org_vdc.existingVdc.compute_capacity[0].memory[0].limit
    }
  }

  storage_profile {
    name    = data.vcd_org_vdc.existingVdc.storage_profile[0].name
    enabled = data.vcd_org_vdc.existingVdc.storage_profile[0].enabled
    limit   = data.vcd_org_vdc.existingVdc.storage_profile[0].limit
    default = data.vcd_org_vdc.existingVdc.storage_profile[0].default
  }

  metadata = {
    vdc_metadata = "VDC Metadata"
  }

  enabled                  = data.vcd_org_vdc.existingVdc.enabled
  enable_thin_provisioning = data.vcd_org_vdc.existingVdc.enable_thin_provisioning
  enable_fast_provisioning = data.vcd_org_vdc.existingVdc.enable_fast_provisioning
  delete_force             = true
  delete_recursive         = true
}
`

const testAccCheckVcdVdcDatasource_orgAdmin = `
data "vcd_org_vdc" "existingVdc" {
  org  = "{{.OrgName}}"
  name = "{{.ExistingVdcName}}"
}
`
