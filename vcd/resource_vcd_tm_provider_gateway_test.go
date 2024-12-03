//go:build tm || ALL || functional

package vcd

// func TestAccVcdTmProviderGateway(t *testing.T) {
// 	preTestChecks(t)
// 	skipIfNotSysAdmin(t)
// 	skipIfNotTm(t)

// 	vCenterHcl, vCenterHclRef := getVCenterHcl(t)
// 	nsxManagerHcl, nsxManagerHclRef := getNsxManagerHcl(t)
// 	regionHcl, regionHclRef := getRegionHcl(t, vCenterHclRef, nsxManagerHclRef)
// 	var params = StringMap{
// 		"Testname":   t.Name(),
// 		"VcenterRef": vCenterHclRef,
// 		"RegionId":   fmt.Sprintf("%s.id", regionHclRef),
// 		"RegionName": t.Name(),

// 		"Tags": "tm",
// 	}
// 	testParamsNotEmpty(t, params)
// }
