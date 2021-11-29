//go:build vdcGroup || ALL || functional
// +build vdcGroup ALL functional

package vcd

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/vmware/go-vcloud-director/v2/govcd"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// TestAccVcdDataCenterGroupDS tests that existing VDC group can be fetched
func TestAccVcdDataCenterGroupDS(t *testing.T) {
	preTestChecks(t)

	// This test requires access to the vCD before filling templates
	// Thus it won't run in the short test
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	vcdClient := createTemporaryVCDConnection()
	if vcdClient.Client.APIVCDMaxVersionIs("< 35.0") {
		t.Skip(t.Name() + " requires at least API v35.0 (vCD 10.2+)")
	}

	vdcGroup, err := getAvailableVdcGroup(vcdClient)
	if err != nil {
		t.Skip("No suitable VDC group found for this test")
		return
	}
	// String map to fill the template
	var params = StringMap{
		"Org":  testConfig.VCD.Org,
		"Name": vdcGroup.VdcGroup.Name,
		"Id":   vdcGroup.VdcGroup.Id,
	}

	configText1 := templateFill(testAccVcdDataCenterGroupDS, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)

	dataSourceName := "data.vcd_data_center_group.existing"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		PreCheck:          func() { testAccPreCheck(t) },

		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText1,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "name", vdcGroup.VdcGroup.Name),
					resource.TestCheckResourceAttr(dataSourceName, "id", vdcGroup.VdcGroup.Id),
					resource.TestCheckResourceAttr(dataSourceName, "description", vdcGroup.VdcGroup.Description),
					resource.TestCheckResourceAttr(dataSourceName, "type", vdcGroup.VdcGroup.Type),
					resource.TestCheckResourceAttr(dataSourceName, "dfw_enabled", strconv.FormatBool(vdcGroup.VdcGroup.DfwEnabled)),
					resource.TestCheckResourceAttr(dataSourceName, "error_message", vdcGroup.VdcGroup.ErrorMessage),
					resource.TestCheckResourceAttr(dataSourceName, "local_egress", strconv.FormatBool(vdcGroup.VdcGroup.LocalEgress)),
					resource.TestCheckResourceAttr(dataSourceName, "network_pool_id", vdcGroup.VdcGroup.NetworkPoolId),
					resource.TestCheckResourceAttr(dataSourceName, "network_pool_universal_id", vdcGroup.VdcGroup.NetworkPoolUniversalId),
					resource.TestCheckResourceAttr(dataSourceName, "network_provider_type", vdcGroup.VdcGroup.NetworkProviderType),
					resource.TestCheckResourceAttr(dataSourceName, "status", vdcGroup.VdcGroup.Status),
					resource.TestCheckResourceAttr(dataSourceName, "type", vdcGroup.VdcGroup.Type),
					resource.TestCheckResourceAttr(dataSourceName, "universal_networking_enabled", strconv.FormatBool(vdcGroup.VdcGroup.UniversalNetworkingEnabled)),
					resource.TestCheckTypeSetElemNestedAttrs(dataSourceName, "participating_org_vdcs.*", map[string]string{
						"fault_domain_tag":       vdcGroup.VdcGroup.ParticipatingOrgVdcs[0].FaultDomainTag,
						"network_provider_scope": vdcGroup.VdcGroup.ParticipatingOrgVdcs[0].NetworkProviderScope,
						"remote_org":             strconv.FormatBool(vdcGroup.VdcGroup.ParticipatingOrgVdcs[0].RemoteOrg),
						"status":                 vdcGroup.VdcGroup.ParticipatingOrgVdcs[0].Status,
					}),
				),
			},
		},
	})
	postTestChecks(t)
}

// getAvailableVdcGroup fetches one available VDC group to use in data source tests
func getAvailableVdcGroup(vcdClient *VCDClient) (*govcd.VdcGroup, error) {
	err := ProviderAuthenticate(vcdClient.VCDClient, testConfig.Provider.User, testConfig.Provider.Password, testConfig.Provider.Token, testConfig.Provider.SysOrg, testConfig.Provider.ApiToken)
	if err != nil {
		return nil, fmt.Errorf("authentication error: %v", err)
	}

	adminOrg, err := vcdClient.GetAdminOrgByName(testConfig.VCD.Org)
	if err != nil {
		return nil, fmt.Errorf("org not found : %s", err)
	}

	vdcGroups, err := adminOrg.GetAllVdcGroups(nil)
	if err != nil {
		return nil, fmt.Errorf("get all VDC groups failed : %s", err)
	}
	if len(vdcGroups) == 0 {
		return nil, fmt.Errorf("no VDC group found in org %v", testConfig.VCD.Org)
	}

	return vdcGroups[0], nil
}

const testAccVcdDataCenterGroupDS = `
data "vcd_data_center_group" "existing" {
  org    = "{{.Org}}"
  name   = "{{.Name}}"
}

data "vcd_data_center_group" "existingById" {
  org = "{{.Org}}"
  id  = "{{.Id}}"
}
`
