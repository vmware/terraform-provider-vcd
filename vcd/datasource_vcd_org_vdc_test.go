//go:build vdc || ALL || functional
// +build vdc ALL functional

package vcd

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

var vdcName = "TestAccVcdVdcDatasource"

func TestAccVcdVdcDatasource(t *testing.T) {
	preTestChecks(t)
	validateConfiguration(t)

	var params = StringMap{
		"ExistingVdcName": testConfig.VCD.Vdc,
		"VdcName":         vdcName,
		"OrgName":         testConfig.VCD.Org,
		"StorageProfile":  testConfig.VCD.ProviderVdc.StorageProfile,
		"FuncName":        vdcName,
	}
	testParamsNotEmpty(t, params)

	vcdClient, err := getTestVCDFromJson(testConfig)
	if err != nil {
		t.Skipf("unable to get vcdClient: %s", err)
	}
	err = ProviderAuthenticate(vcdClient, testConfig.Provider.User, testConfig.Provider.Password, testConfig.Provider.Token, testConfig.Provider.SysOrg, testConfig.Provider.ApiToken)
	if err != nil {
		t.Skipf("authentication error: %s", err)
	}
	org, err := vcdClient.GetAdminOrgByName(testConfig.VCD.Org)
	if err != nil {
		t.Skipf("unable to get Org: %s, err: %s", testConfig.VCD.Org, err)
	}
	vdc, err := org.GetVDCByName(testConfig.VCD.Vdc, false)
	if err != nil {
		t.Skipf("unable to get VDC: %s, err: %s", testConfig.VCD.Vdc, err)
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

		validateDataSource(t, configText, datasourceVdc, params)
	} else {
		if vdc.Vdc.AllocationModel == "Flex" {
			configText = templateFill(testAccCheckVcdVdcDatasource_basic_flex, params)
		} else {
			configText = templateFill(testAccCheckVcdVdcDatasource_basic, params)
		}

		debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

		if vcdShortTest {
			t.Skip(acceptanceTestsSkipped)
			return
		}

		validateResourceAndDataSource(t, configText, datasourceVdc, params)
	}
	postTestChecks(t)
}

func validateResourceAndDataSource(t *testing.T, configText string, datasourceVdc string, params StringMap) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVdcDestroy,
		Steps: []resource.TestStep{
			{
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
					resource.TestCheckResourceAttr("vcd_org_vdc."+vdcName, "storage_profile.0.enabled", "true"),
					resource.TestCheckResourceAttr("vcd_org_vdc."+vdcName, "storage_profile.0.default", "true"),
					resource.TestCheckResourceAttr("vcd_org_vdc."+vdcName, "storage_profile.0.limit", "0"),
					resource.TestCheckResourceAttr("vcd_org_vdc."+vdcName, "storage_profile.0.name", testConfig.VCD.ProviderVdc.StorageProfile),
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
						"data."+datasourceVdc, "compute_capacity.0.memory.0.used", regexp.MustCompile(`^\d+$`)),
					resource.TestMatchResourceAttr(
						"data."+datasourceVdc, "storage_profile.0.storage_used_in_mb", regexp.MustCompile(`^\d+$`))),
			},
		},
	})
}

func validateDataSource(t *testing.T, configText string, datasourceVdc string, params StringMap) {
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVdcDestroy,
		Steps: []resource.TestStep{
			{
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
					resource.TestMatchResourceAttr("data."+datasourceVdc, "compute_capacity.0.memory.0.reserved", regexp.MustCompile(`^\d+$`)),
					resource.TestMatchResourceAttr("data."+datasourceVdc, "storage_profile.0.storage_used_in_mb", regexp.MustCompile(`^\d+$`))),
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
    name    = "{{.StorageProfile}}"
    enabled = true
    limit   = 0
    default = true
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
const testAccCheckVcdVdcDatasource_basic_flex = `
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
    name    = "{{.StorageProfile}}"
    enabled = true
    limit   = 0
    default = true
  }

  metadata = {
    vdc_metadata = "VDC Metadata"
  }

  enabled                  = data.vcd_org_vdc.existingVdc.enabled
  enable_thin_provisioning = data.vcd_org_vdc.existingVdc.enable_thin_provisioning
  enable_fast_provisioning = data.vcd_org_vdc.existingVdc.enable_fast_provisioning
  delete_force             = true
  delete_recursive         = true

  elasticity                 = data.vcd_org_vdc.existingVdc.elasticity
  include_vm_memory_overhead = data.vcd_org_vdc.existingVdc.include_vm_memory_overhead
}
`

const testAccCheckVcdVdcDatasource_orgAdmin = `
data "vcd_org_vdc" "existingVdc" {
  org  = "{{.OrgName}}"
  name = "{{.ExistingVdcName}}"
}
`
