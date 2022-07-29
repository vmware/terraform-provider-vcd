//go:build vdc || ALL || functional
// +build vdc ALL functional

package vcd

import (
	"testing"
)

func init() {
	testingTags["vdc"] = "resource_vcd_org_vdc_test.go"
}

func TestAccVcdOrgVdcReservationPool(t *testing.T) {
	preTestChecks(t)
	if !usingSysAdmin() {
		t.Skip(t.Name() + " requires system admin privileges")
	}

	allocationModel := "ReservationPool"

	var params = StringMap{
		"VdcName":                    TestAccVcdVdc,
		"OrgName":                    testConfig.VCD.Org,
		"AllocationModel":            "ReservationPool",
		"ProviderVdc":                testConfig.VCD.NsxtProviderVdc.Name,
		"NetworkPool":                testConfig.VCD.NsxtProviderVdc.NetworkPool,
		"Allocated":                  "1024",
		"Reserved":                   "1024",
		"Limit":                      "1024",
		"LimitIncreased":             "1100",
		"AllocatedIncreased":         "1100",
		"ProviderVdcStorageProfile":  testConfig.VCD.NsxtProviderVdc.StorageProfile,
		"ProviderVdcStorageProfile2": testConfig.VCD.NsxtProviderVdc.StorageProfile2,
		"Tags":                       "vdc",
		"FuncName":                   t.Name(),
		// cause vDC ignores empty values and use default
		"MemoryGuaranteed": "1",
		"CpuGuaranteed":    "1",
		// The parameters below are for Flex allocation model
		// Part of HCL is created dynamically and with empty values we don't create the Flex part:
		"equalsChar":                         " ",
		"FlexElasticKey":                     " ",
		"FlexElasticValue":                   " ",
		"FlexElasticValueUpdate":             " ",
		"FlexMemoryOverheadKey":              " ",
		"FlexMemoryOverheadValue":            " ",
		"FlexMemoryOverheadValueUpdate":      " ",
		"MemoryOverheadValueForAssert":       "true",
		"MemoryOverheadUpdateValueForAssert": "true",
		"ElasticityValueForAssert":           "false",
		"ElasticityUpdateValueForAssert":     "false",
	}

	runOrgVdcTest(t, params, allocationModel)
	postTestChecks(t)
}

func TestAccVcdOrgVdcAllocationPool(t *testing.T) {
	preTestChecks(t)
	if !usingSysAdmin() {
		t.Skip(t.Name() + " requires system admin privileges")
	}

	allocationModel := "AllocationPool"

	var params = StringMap{
		"VdcName":                    TestAccVcdVdc,
		"OrgName":                    testConfig.VCD.Org,
		"AllocationModel":            "AllocationPool",
		"ProviderVdc":                testConfig.VCD.NsxtProviderVdc.Name,
		"NetworkPool":                testConfig.VCD.NsxtProviderVdc.NetworkPool,
		"Allocated":                  "2048",
		"Reserved":                   "1024",
		"Limit":                      "2048",
		"LimitIncreased":             "2148",
		"AllocatedIncreased":         "2148",
		"ProviderVdcStorageProfile":  testConfig.VCD.NsxtProviderVdc.StorageProfile,
		"ProviderVdcStorageProfile2": testConfig.VCD.NsxtProviderVdc.StorageProfile2,
		"Tags":                       "vdc",
		"FuncName":                   t.Name(),
		"MemoryGuaranteed":           "0.3",
		"CpuGuaranteed":              "0.45",
		// The parameters below are for Flex allocation model
		// Part of HCL is created dynamically and with empty values we don't create the Flex part:
		"equalsChar":                         " ",
		"FlexElasticKey":                     " ",
		"FlexElasticValue":                   " ",
		"FlexElasticValueUpdate":             " ",
		"ElasticityValueForAssert":           "false",
		"ElasticityUpdateValueForAssert":     "false",
		"FlexMemoryOverheadKey":              " ",
		"FlexMemoryOverheadValue":            " ",
		"FlexMemoryOverheadValueUpdate":      " ",
		"MemoryOverheadValueForAssert":       "true",
		"MemoryOverheadUpdateValueForAssert": "true",
	}

	runOrgVdcTest(t, params, allocationModel)
	postTestChecks(t)
}

func TestAccVcdOrgVdcAllocationVApp(t *testing.T) {
	preTestChecks(t)
	if !usingSysAdmin() {
		t.Skip(t.Name() + " requires system admin privileges")
	}

	allocationModel := "AllocationVApp"

	var params = StringMap{
		"VdcName":                    TestAccVcdVdc,
		"OrgName":                    testConfig.VCD.Org,
		"AllocationModel":            allocationModel,
		"ProviderVdc":                testConfig.VCD.NsxtProviderVdc.Name,
		"NetworkPool":                testConfig.VCD.NsxtProviderVdc.NetworkPool,
		"Allocated":                  "0",
		"Reserved":                   "0",
		"Limit":                      "2048",
		"LimitIncreased":             "2148",
		"AllocatedIncreased":         "0",
		"ProviderVdcStorageProfile":  testConfig.VCD.NsxtProviderVdc.StorageProfile,
		"ProviderVdcStorageProfile2": testConfig.VCD.NsxtProviderVdc.StorageProfile2,
		"Tags":                       "vdc",
		"FuncName":                   t.Name(),
		"MemoryGuaranteed":           "0.5",
		"CpuGuaranteed":              "0.6",
		// The parameters below are for Flex allocation model
		// Part of HCL is created dynamically and with empty values we don't create the Flex part:
		"equalsChar":                         " ",
		"FlexElasticKey":                     " ",
		"FlexElasticValue":                   " ",
		"FlexElasticValueUpdate":             " ",
		"ElasticityValueForAssert":           "true",
		"ElasticityUpdateValueForAssert":     "true",
		"FlexMemoryOverheadKey":              " ",
		"FlexMemoryOverheadValue":            " ",
		"FlexMemoryOverheadValueUpdate":      " ",
		"MemoryOverheadValueForAssert":       "false",
		"MemoryOverheadUpdateValueForAssert": "false",
	}

	runOrgVdcTest(t, params, allocationModel)
	postTestChecks(t)
}

func TestAccVcdOrgVdcAllocationFlex(t *testing.T) {
	preTestChecks(t)
	if !usingSysAdmin() {
		t.Skip(t.Name() + " requires system admin privileges")
	}

	allocationModel := "Flex"

	var params = StringMap{
		"VdcName":                    TestAccVcdVdc,
		"OrgName":                    testConfig.VCD.Org,
		"AllocationModel":            allocationModel,
		"ProviderVdc":                testConfig.VCD.NsxtProviderVdc.Name,
		"NetworkPool":                testConfig.VCD.NsxtProviderVdc.NetworkPool,
		"Allocated":                  "1024",
		"Reserved":                   "0",
		"Limit":                      "1024",
		"LimitIncreased":             "1124",
		"AllocatedIncreased":         "1124",
		"ProviderVdcStorageProfile":  testConfig.VCD.NsxtProviderVdc.StorageProfile,
		"ProviderVdcStorageProfile2": testConfig.VCD.NsxtProviderVdc.StorageProfile2,
		"Tags":                       "vdc",
		"FuncName":                   t.Name(),
		"MemoryGuaranteed":           "0.5",
		"CpuGuaranteed":              "0.6",
		// The parameters below are for Flex allocation model
		// Part of HCL is created dynamically and these parameters with values result in the Flex part of the template being filled:
		"equalsChar":                         "=",
		"FlexElasticKey":                     "elasticity",
		"FlexElasticValue":                   "false",
		"FlexElasticValueUpdate":             "true",
		"ElasticityValueForAssert":           "false",
		"ElasticityUpdateValueForAssert":     "true",
		"FlexMemoryOverheadKey":              "include_vm_memory_overhead",
		"FlexMemoryOverheadValue":            "false",
		"FlexMemoryOverheadValueUpdate":      "true",
		"MemoryOverheadValueForAssert":       "false",
		"MemoryOverheadUpdateValueForAssert": "true",
	}
	runOrgVdcTest(t, params, allocationModel)
	postTestChecks(t)
}
