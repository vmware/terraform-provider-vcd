//go:build vdc || ALL || functional
// +build vdc ALL functional

package vcd

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcdVmPlacementPolicy(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	var params = StringMap{
		"VdcName":     testConfig.Nsxt.Vdc,
		"PolicyName":  t.Name(),
		"VmGroup":     testConfig.VCD.NsxtProviderVdc.PlacementPolicyVmGroup,
		"Description": t.Name() + "_description",
	}
	testParamsNotEmpty(t, params)
	policyName := "vcd_vm_placement_policy." + params["PolicyName"].(string)
	datasourcePolicyName := "data.vcd_vm_placement_policy.data-" + params["PolicyName"].(string)
	configText := templateFill(testAccCheckVmPlacementPolicy_create, params)
	params["FuncName"] = t.Name() + "-Update"
	configTextUpdate := templateFill(testAccCheckVmPlacementPolicy_update, params)

	debugPrintf("#[DEBUG] CONFIGURATION - creation: %s", configText)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckComputePolicyDestroyed(t.Name()+"-update", "placement"),
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(policyName, "id", regexp.MustCompile(`urn:vcloud:vdcComputePolicy:[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$`)),
					resource.TestCheckResourceAttr(policyName, "name", params["PolicyName"].(string)),
					resource.TestCheckResourceAttr(policyName, "description", params["Description"].(string)),
					resource.TestMatchResourceAttr(policyName, "provider_vdc_id", regexp.MustCompile(`urn:vcloud:providervdc:[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$`)),
					resource.TestCheckResourceAttr(policyName, "vm_group_ids.#", "1"),
					resource.TestCheckResourceAttr(policyName, "logical_vm_group_ids.#", "0"),
					resource.TestMatchResourceAttr(policyName, "vm_group_ids.0", regexp.MustCompile(`^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$`)),
					resource.TestCheckNoResourceAttr(policyName, "vdc_id"),
					resourceFieldsEqual(policyName, datasourcePolicyName, []string{"%"}), // Data source has extra attribute `vdc_id`
				),
			},
			{
				Config: configTextUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(policyName, "id", regexp.MustCompile(`urn:vcloud:vdcComputePolicy:[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$`)),
					resource.TestCheckResourceAttr(policyName, "name", params["PolicyName"].(string)+"-update"),
					resource.TestCheckResourceAttr(policyName, "description", params["Description"].(string)+"-update"),
					resource.TestMatchResourceAttr(policyName, "provider_vdc_id", regexp.MustCompile(`urn:vcloud:providervdc:[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$`)),
					resource.TestCheckResourceAttr(policyName, "vm_group_ids.#", "1"),
					resource.TestCheckResourceAttr(policyName, "logical_vm_group_ids.#", "0"),
					resource.TestMatchResourceAttr(policyName, "vm_group_ids.0", regexp.MustCompile(`^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$`)),
					resource.TestCheckNoResourceAttr(policyName, "vdc_id"),
					resourceFieldsEqual(policyName, datasourcePolicyName, []string{"%"}), // Data source has extra attribute `vdc_id`
				),
			},
			// Tests import by id
			{
				ResourceName:      policyName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateComputePolicyByIdOrName(policyName, true),
			},
			// Tests import by name
			{
				ResourceName:      policyName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateComputePolicyByIdOrName(policyName, false),
			},
		},
	})
	postTestChecks(t)
}

const testAccCheckVmPlacementPolicy_create = `
data "vcd_org_vdc" "vdc" {
  name = "{{.VdcName}}"
}

data "vcd_provider_vdc" "pvdc" {
  name = data.vcd_org_vdc.vdc.provider_vdc_name
}

data "vcd_vm_group" "vm-group" {
  name            = "{{.VmGroup}}"
  provider_vdc_id = data.vcd_provider_vdc.pvdc.id
}

resource "vcd_vm_placement_policy" "{{.PolicyName}}" {
  name            = "{{.PolicyName}}"
  description     = "{{.Description}}"
  provider_vdc_id = data.vcd_provider_vdc.pvdc.id
  vm_group_ids    = [ data.vcd_vm_group.vm-group.id ]
}

data "vcd_vm_placement_policy" "data-{{.PolicyName}}" {
  name            = vcd_vm_placement_policy.{{.PolicyName}}.name
  provider_vdc_id = vcd_vm_placement_policy.{{.PolicyName}}.provider_vdc_id
}
`

const testAccCheckVmPlacementPolicy_update = `
data "vcd_org_vdc" "vdc" {
  name = "{{.VdcName}}"
}

data "vcd_provider_vdc" "pvdc" {
  name = data.vcd_org_vdc.vdc.provider_vdc_name
}

data "vcd_vm_group" "vm-group" {
  name            = "{{.VmGroup}}"
  provider_vdc_id = data.vcd_provider_vdc.pvdc.id
}

resource "vcd_vm_placement_policy" "{{.PolicyName}}" {
  name            = "{{.PolicyName}}-update"
  description     = "{{.Description}}-update"
  provider_vdc_id = data.vcd_provider_vdc.pvdc.id
  vm_group_ids    = [ data.vcd_vm_group.vm-group.id ]
}

data "vcd_vm_placement_policy" "data-{{.PolicyName}}" {
  name            = vcd_vm_placement_policy.{{.PolicyName}}.name
  provider_vdc_id = vcd_vm_placement_policy.{{.PolicyName}}.provider_vdc_id
}
`

// TestAccVcdVmPlacementPolicyInVdc tests fetching a VM Placement Policy using the `vdc_id` instead of Provider VDC.
// This should be tested as tenant user, but due to limitations this must be done manually as we need Sysadmin to create
// the VDC and assign the policies first.
func TestAccVcdVmPlacementPolicyInVdc(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	var params = StringMap{
		"VdcName":     testConfig.Nsxt.Vdc,
		"PolicyName":  t.Name(),
		"ProviderVdc": testConfig.VCD.NsxtProviderVdc.Name,
		"NetworkPool": testConfig.VCD.NsxtProviderVdc.NetworkPool,
		"VmGroup":     testConfig.VCD.NsxtProviderVdc.PlacementPolicyVmGroup,
	}
	testParamsNotEmpty(t, params)
	policyName := "vcd_vm_placement_policy." + params["PolicyName"].(string)
	datasourcePolicyName := "data.vcd_vm_placement_policy.data-" + params["PolicyName"].(string)
	configText := templateFill(testAccCheckVmPlacementPolicyFromVdcId, params)

	debugPrintf("#[DEBUG] CONFIGURATION - creation: %s", configText)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckComputePolicyDestroyed(t.Name(), "placement"),
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(policyName, "id", regexp.MustCompile(`urn:vcloud:vdcComputePolicy:[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$`)),
					resource.TestCheckResourceAttr(policyName, "name", params["PolicyName"].(string)),
					resource.TestCheckResourceAttr(policyName, "description", "foo"),
					resource.TestMatchResourceAttr(policyName, "vdc_id", regexp.MustCompile(`urn:vcloud:vdc:[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$`)),
					resource.TestCheckResourceAttr(policyName, "vm_group_ids.#", "1"),
					resource.TestCheckResourceAttr(policyName, "logical_vm_group_ids.#", "0"),
					resource.TestMatchResourceAttr(policyName, "vm_group_ids.0", regexp.MustCompile(`^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$`)),
					resource.TestCheckNoResourceAttr(policyName, "provider_vdc_id"),
					resourceFieldsEqual(policyName, datasourcePolicyName, []string{"%"}), // Data source has extra attribute `vdc_id`
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccCheckVmPlacementPolicyFromVdcId_prereqs = `
data "vcd_provider_vdc" "pvdc" {
  name = "{{.ProviderVdc}}"
}

data "vcd_vm_group" "vm-group" {
  name            = "{{.VmGroup}}"
  provider_vdc_id = data.vcd_provider_vdc.pvdc.id
}

resource "vcd_vm_placement_policy" "{{.PolicyName}}" {
  name            = "{{.PolicyName}}"
  description     = "foo"
  provider_vdc_id = data.vcd_provider_vdc.pvdc.id
  vm_group_ids    = [ data.vcd_vm_group.vm-group.id ]
}

resource "vcd_org_vdc" "{{.VdcName}}" {
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

  default_compute_policy_id = vcd_vm_placement_policy.{{.PolicyName}}.id
  vm_placement_policy_ids   = [vcd_vm_placement_policy.{{.PolicyName}}.id]
}
`

const testAccCheckVmPlacementPolicyFromVdcId = testAccCheckVmPlacementPolicyFromVdcId_prereqs + `
data "vcd_vm_placement_policy" "data-{{.PolicyName}}" {
  name   = vcd_vm_placement_policy.{{.PolicyName}}.name
  vdc_id = vcd_org_vdc.{{.VdcName}}.id
}
`

// TestAccVcdVmPlacementPolicyWithoutDescription checks that a VM Placement Policy without description specified in the HCL
// corresponds to a VM Placement Policy with an empty description in VCD.
func TestAccVcdVmPlacementPolicyWithoutDescription(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)
	if testConfig.VCD.ProviderVdc.Name == "" {
		t.Skip("Variable providerVdc.Name must be set to run VDC tests")
	}

	var params = StringMap{
		"PvdcName":   testConfig.VCD.NsxtProviderVdc.Name,
		"PolicyName": t.Name(),
		"VmGroup":    testConfig.VCD.NsxtProviderVdc.PlacementPolicyVmGroup,
	}
	testParamsNotEmpty(t, params)
	policyName := "vcd_vm_placement_policy." + params["PolicyName"].(string)
	configText := templateFill(testAccCheckVmPlacementPolicyWithoutDescription, params)

	debugPrintf("#[DEBUG] CONFIGURATION - creation: %s", configText)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckComputePolicyDestroyed(t.Name(), "placement"),
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(policyName, "description", ""),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccCheckVmPlacementPolicyWithoutDescription = `
data "vcd_provider_vdc" "pvdc" {
  name = "{{.PvdcName}}"
}

data "vcd_vm_group" "vm-group" {
  name            = "{{.VmGroup}}"
  provider_vdc_id = data.vcd_provider_vdc.pvdc.id
}

resource "vcd_vm_placement_policy" "{{.PolicyName}}" {
  name            = "{{.PolicyName}}"
  provider_vdc_id = data.vcd_provider_vdc.pvdc.id
  vm_group_ids    = [ data.vcd_vm_group.vm-group.id ]
}
`
