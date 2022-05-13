//go:build vdc || ALL || functional
// +build vdc ALL functional

package vcd

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcdOrgVdcWithVmSizingPolicy(t *testing.T) {
	preTestChecks(t)
	if !usingSysAdmin() {
		t.Skip("TestAccVcdOrgVdcWithVmSizingPolicy requires system admin privileges")
	}

	vcdClient, err := getTestVCDFromJson(testConfig)
	if err != nil {
		t.Skip("unable to validate vCD version - skipping test")
	}

	if vcdClient.Client.APIVCDMaxVersionIs("< 33.0") {
		t.Skip("TestAccVcdOrgVdcWithVmSizingPolicy requires VCD 10.0+")
	}

	validateConfiguration(t)

	allocationModel := "Flex"

	var params = StringMap{
		"VdcName":                   TestAccVcdVdc,
		"OrgName":                   testConfig.VCD.Org,
		"AllocationModel":           allocationModel,
		"ProviderVdc":               testConfig.VCD.ProviderVdc.Name,
		"NetworkPool":               testConfig.VCD.ProviderVdc.NetworkPool,
		"Allocated":                 "1024",
		"Reserved":                  "0",
		"Limit":                     "1024",
		"LimitIncreased":            "1124",
		"AllocatedIncreased":        "1124",
		"ProviderVdcStorageProfile": testConfig.VCD.ProviderVdc.StorageProfile,
		"Tags":                      "vdc",
		"FuncName":                  t.Name(),
		"MemoryGuaranteed":          "0.5",
		"CpuGuaranteed":             "0.6",
		// The parameters below are for Flex allocation model
		// Part of HCL is created dynamically and these parameters with values result in the Flex part of the template being filled:
		"equalsChar":                   "=",
		"FlexElasticKey":               "elasticity",
		"FlexElasticValue":             "false",
		"ElasticityValueForAssert":     "false",
		"FlexMemoryOverheadKey":        "include_vm_memory_overhead",
		"FlexMemoryOverheadValue":      "false",
		"MemoryOverheadValueForAssert": "false",
	}
	testParamsNotEmpty(t, params)

	configText := templateFill(testAccCheckVcdVdcVmSizingPolicies_basic, params)
	params["FuncName"] = t.Name() + "-Update"
	updateText := templateFill(testAccCheckVcdVdcVmSizingPolicies_update, params)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	debugPrintf("#[DEBUG] CONFIGURATION - creation: %s", configText)
	debugPrintf("#[DEBUG] CONFIGURATION - update: %s", updateText)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckVdcDestroy,
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdVdcExists("vcd_org_vdc."+TestAccVcdVdc),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+TestAccVcdVdc, "name", TestAccVcdVdc),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+TestAccVcdVdc, "org", testConfig.VCD.Org),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+TestAccVcdVdc, "allocation_model", allocationModel),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+TestAccVcdVdc, "network_pool_name", testConfig.VCD.ProviderVdc.NetworkPool),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+TestAccVcdVdc, "provider_vdc_name", testConfig.VCD.ProviderVdc.Name),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+TestAccVcdVdc, "enabled", "true"),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+TestAccVcdVdc, "enable_thin_provisioning", "true"),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+TestAccVcdVdc, "enable_fast_provisioning", "true"),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+TestAccVcdVdc, "delete_force", "true"),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+TestAccVcdVdc, "delete_recursive", "true"),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+TestAccVcdVdc, "metadata.vdc_metadata", "VDC Metadata"),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+TestAccVcdVdc, "storage_profile.0.name", testConfig.VCD.ProviderVdc.StorageProfile),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+TestAccVcdVdc, "storage_profile.0.enabled", "true"),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+TestAccVcdVdc, "storage_profile.0.limit", "10240"),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+TestAccVcdVdc, "storage_profile.0.default", "true"),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+TestAccVcdVdc, "compute_capacity.0.cpu.0.allocated", params["Allocated"].(string)),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+TestAccVcdVdc, "compute_capacity.0.cpu.0.limit", params["Limit"].(string)),
					resource.TestMatchResourceAttr(
						"vcd_org_vdc."+TestAccVcdVdc, "compute_capacity.0.cpu.0.reserved", regexp.MustCompile(`^\d+$`)),
					resource.TestMatchResourceAttr(
						"vcd_org_vdc."+TestAccVcdVdc, "compute_capacity.0.cpu.0.used", regexp.MustCompile(`^\d+$`)),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+TestAccVcdVdc, "compute_capacity.0.memory.0.allocated", params["Allocated"].(string)),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+TestAccVcdVdc, "compute_capacity.0.memory.0.limit", params["Limit"].(string)),
					resource.TestMatchResourceAttr(
						"vcd_org_vdc."+TestAccVcdVdc, "compute_capacity.0.memory.0.reserved", regexp.MustCompile(`^\d+$`)),
					resource.TestMatchResourceAttr(
						"vcd_org_vdc."+TestAccVcdVdc, "compute_capacity.0.memory.0.used", regexp.MustCompile(`^\d+$`)),
					resource.TestMatchResourceAttr(
						"vcd_org_vdc."+TestAccVcdVdc, "elasticity", regexp.MustCompile(`^`+params["ElasticityValueForAssert"].(string)+`$`)),
					resource.TestMatchResourceAttr(
						"vcd_org_vdc."+TestAccVcdVdc, "include_vm_memory_overhead", regexp.MustCompile(`^`+params["MemoryOverheadValueForAssert"].(string)+`$`)),

					resource.TestCheckResourceAttrPair("vcd_org_vdc."+TestAccVcdVdc, "default_vm_sizing_policy_id",
						"vcd_vm_sizing_policy.minSize3", "id"),
					resource.TestCheckResourceAttr("vcd_org_vdc."+TestAccVcdVdc, "vm_sizing_policy_ids.#",
						"3"),
				),
			},
			{
				Config: updateText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdVdcExists("vcd_org_vdc."+TestAccVcdVdc),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+TestAccVcdVdc, "name", TestAccVcdVdc),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+TestAccVcdVdc, "org", testConfig.VCD.Org),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+TestAccVcdVdc, "allocation_model", allocationModel),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+TestAccVcdVdc, "network_pool_name", testConfig.VCD.ProviderVdc.NetworkPool),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+TestAccVcdVdc, "provider_vdc_name", testConfig.VCD.ProviderVdc.Name),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+TestAccVcdVdc, "enabled", "true"),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+TestAccVcdVdc, "enable_thin_provisioning", "true"),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+TestAccVcdVdc, "enable_fast_provisioning", "true"),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+TestAccVcdVdc, "delete_force", "true"),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+TestAccVcdVdc, "delete_recursive", "true"),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+TestAccVcdVdc, "metadata.vdc_metadata", "VDC Metadata"),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+TestAccVcdVdc, "storage_profile.0.name", testConfig.VCD.ProviderVdc.StorageProfile),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+TestAccVcdVdc, "storage_profile.0.enabled", "true"),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+TestAccVcdVdc, "storage_profile.0.limit", "10240"),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+TestAccVcdVdc, "storage_profile.0.default", "true"),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+TestAccVcdVdc, "compute_capacity.0.cpu.0.allocated", params["Allocated"].(string)),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+TestAccVcdVdc, "compute_capacity.0.cpu.0.limit", params["Limit"].(string)),
					resource.TestMatchResourceAttr(
						"vcd_org_vdc."+TestAccVcdVdc, "compute_capacity.0.cpu.0.reserved", regexp.MustCompile(`^\d+$`)),
					resource.TestMatchResourceAttr(
						"vcd_org_vdc."+TestAccVcdVdc, "compute_capacity.0.cpu.0.used", regexp.MustCompile(`^\d+$`)),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+TestAccVcdVdc, "compute_capacity.0.memory.0.allocated", params["Allocated"].(string)),
					resource.TestCheckResourceAttr(
						"vcd_org_vdc."+TestAccVcdVdc, "compute_capacity.0.memory.0.limit", params["Limit"].(string)),
					resource.TestMatchResourceAttr(
						"vcd_org_vdc."+TestAccVcdVdc, "compute_capacity.0.memory.0.reserved", regexp.MustCompile(`^\d+$`)),
					resource.TestMatchResourceAttr(
						"vcd_org_vdc."+TestAccVcdVdc, "compute_capacity.0.memory.0.used", regexp.MustCompile(`^\d+$`)),
					resource.TestMatchResourceAttr(
						"vcd_org_vdc."+TestAccVcdVdc, "elasticity", regexp.MustCompile(`^`+params["ElasticityValueForAssert"].(string)+`$`)),
					resource.TestMatchResourceAttr(
						"vcd_org_vdc."+TestAccVcdVdc, "include_vm_memory_overhead", regexp.MustCompile(`^`+params["MemoryOverheadValueForAssert"].(string)+`$`)),

					resource.TestCheckResourceAttrPair("vcd_org_vdc."+TestAccVcdVdc, "default_vm_sizing_policy_id",
						"vcd_vm_sizing_policy.minSize2", "id"),
					resource.TestCheckResourceAttr("vcd_org_vdc."+TestAccVcdVdc, "vm_sizing_policy_ids.#",
						"1"),
				),
			},
			{
				ResourceName:      "vcd_org_vdc." + TestAccVcdVdc,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdOrgObject(testConfig, TestAccVcdVdc),
				// These fields can't be retrieved
				ImportStateVerifyIgnore: []string{"delete_force", "delete_recursive"},
			},
		},
	})
	postTestChecks(t)
}

func init() {
	testingTags["vdc"] = "resource_vcd_org_vdc_with_vm_sizing_policy_test.go"
}

const testAccCheckVcdVdcVmSizingPolicies_basic = `
resource "vcd_vm_sizing_policy" "minSize" {
  name        = "min-size"
  description = "smallest size"
}

resource "vcd_vm_sizing_policy" "minSize2" {
  name        = "min-size2"
  description = "smallest size2"

  cpu {
    shares                = "886"
    limit_in_mhz          = "12375"
    count                 = "9"
    speed_in_mhz          = "2500"
    cores_per_socket      = "3"
    reservation_guarantee = "0.55"
  }

}

resource "vcd_vm_sizing_policy" "minSize3" {
  name        = "min-size3"
  description = "smallest size2"

  cpu {
    shares                = "886"
    limit_in_mhz          = "12375"
    count                 = "9"
    speed_in_mhz          = "2500"
    cores_per_socket      = "3"
    reservation_guarantee = "0.55"
  }

  memory {
    shares                = "1580"
    size_in_mb            = "3200"
    limit_in_mb           = "2800"
    reservation_guarantee = "0.3"
  }
}

resource "vcd_org_vdc" "{{.VdcName}}" {
  name = "{{.VdcName}}"
  org  = "{{.OrgName}}"

  allocation_model  = "{{.AllocationModel}}"
  network_pool_name = "{{.NetworkPool}}"
  provider_vdc_name = "{{.ProviderVdc}}"

  compute_capacity {
    cpu {
      allocated = "{{.Allocated}}"
      limit     = "{{.Limit}}"
    }

    memory {
      allocated = "{{.Allocated}}"
      limit     = "{{.Limit}}"
    }
  }

  storage_profile {
    name     = "{{.ProviderVdcStorageProfile}}"
    enabled  = true
    limit    = 10240
    default  = true
  }

  metadata = {
    vdc_metadata = "VDC Metadata"
  }

  enabled                    = true
  enable_thin_provisioning   = true
  enable_fast_provisioning   = true
  delete_force               = true
  delete_recursive           = true
  {{.FlexElasticKey}}                 {{.equalsChar}} {{.FlexElasticValue}}
  {{.FlexMemoryOverheadKey}} {{.equalsChar}} {{.FlexMemoryOverheadValue}}

  default_vm_sizing_policy_id = vcd_vm_sizing_policy.minSize3.id
  vm_sizing_policy_ids        = [vcd_vm_sizing_policy.minSize.id, vcd_vm_sizing_policy.minSize2.id,vcd_vm_sizing_policy.minSize3.id]
}
`

const testAccCheckVcdVdcVmSizingPolicies_update = `
# skip-binary-test: only for updates
resource "vcd_vm_sizing_policy" "minSize" {
  name        = "min-size"
  description = "smallest size"
}

resource "vcd_vm_sizing_policy" "minSize2" {
  name        = "min-size2"
  description = "smallest size2"

  cpu {
    shares                = "886"
    limit_in_mhz          = "12375"
    count                 = "9"
    speed_in_mhz          = "2500"
    cores_per_socket      = "3"
    reservation_guarantee = "0.55"
  }

}

resource "vcd_vm_sizing_policy" "minSize3" {
  name        = "min-size3"
  description = "smallest size2"

  cpu {
    shares                = "886"
    limit_in_mhz          = "12375"
    count                 = "9"
    speed_in_mhz          = "2500"
    cores_per_socket      = "3"
    reservation_guarantee = "0.55"
  }

  memory {
    shares                = "1580"
    size_in_mb            = "3200"
    limit_in_mb           = "2800"
    reservation_guarantee = "0.3"
  }
}

resource "vcd_org_vdc" "{{.VdcName}}" {
  name = "{{.VdcName}}"
  org  = "{{.OrgName}}"

  allocation_model  = "{{.AllocationModel}}"
  network_pool_name = "{{.NetworkPool}}"
  provider_vdc_name = "{{.ProviderVdc}}"

  compute_capacity {
    cpu {
      allocated = "{{.Allocated}}"
      limit     = "{{.Limit}}"
    }

    memory {
      allocated = "{{.Allocated}}"
      limit     = "{{.Limit}}"
    }
  }

  storage_profile {
    name     = "{{.ProviderVdcStorageProfile}}"
    enabled  = true
    limit    = 10240
    default  = true
  }

  metadata = {
    vdc_metadata = "VDC Metadata"
  }

  enabled                    = true
  enable_thin_provisioning   = true
  enable_fast_provisioning   = true
  delete_force               = true
  delete_recursive           = true
  {{.FlexElasticKey}}                 {{.equalsChar}} {{.FlexElasticValue}}
  {{.FlexMemoryOverheadKey}} {{.equalsChar}} {{.FlexMemoryOverheadValue}}

  default_vm_sizing_policy_id = vcd_vm_sizing_policy.minSize2.id
  vm_sizing_policy_ids        = [vcd_vm_sizing_policy.minSize2.id]
}
`
