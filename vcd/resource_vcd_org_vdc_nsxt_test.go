// +build nsxt vdc ALL functional

package vcd

import "testing"

func init() {
	testingTags["vdc"] = "resource_vcd_org_vdc_nsxt_test.go"
}

func TestAccVcdOrgVdcNsxt(t *testing.T) {
	if !usingSysAdmin() {
		t.Skip(t.Name() + " requires system admin privileges")
	}
	skipNoNsxtConfiguration(t)
	validateConfiguration(t)

	allocationModel := "ReservationPool"

	var params = StringMap{
		"VdcName":                   TestAccVcdVdc,
		"OrgName":                   testConfig.VCD.Org,
		"AllocationModel":           "ReservationPool",
		"ProviderVdc":               testConfig.VCD.NsxtProviderVdc.Name,
		"NetworkPool":               testConfig.VCD.NsxtProviderVdc.NetworkPool,
		"Allocated":                 "1024",
		"Reserved":                  "1024",
		"Limit":                     "1024",
		"LimitIncreased":            "1100",
		"AllocatedIncreased":        "1100",
		"ProviderVdcStorageProfile": testConfig.VCD.NsxtProviderVdc.StorageProfile,
		"Tags":                      "vdc",
		"FuncName":                  t.Name(),
		// cause vDC ignores empty values and use default
		"MemoryGuaranteed": "1",
		"CpuGuaranteed":    "1",
		// The parameters below are for Flex allocation model
		// Part of HCL is created dynamically and with empty values we don't create the Flex part:
		"equalsChar":                         "",
		"FlexElasticKey":                     "",
		"FlexElasticValue":                   "",
		"FlexElasticValueUpdate":             "",
		"FlexMemoryOverheadKey":              "",
		"FlexMemoryOverheadValue":            "",
		"FlexMemoryOverheadValueUpdate":      "",
		"MemoryOverheadValueForAssert":       "true",
		"MemoryOverheadUpdateValueForAssert": "true",
		"ElasticityValueForAssert":           "false",
		"ElasticityUpdateValueForAssert":     "false",
	}

	runOrgVdcTest(t, params, allocationModel)
}
