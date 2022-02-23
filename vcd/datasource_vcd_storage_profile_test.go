//go:build catalog || ALL || functional
// +build catalog ALL functional

package vcd

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/vmware/go-vcloud-director/v2/govcd"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcdStorageProfileDS(t *testing.T) {
	preTestChecks(t)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	// Lookup
	storageProfileId := findStorageProfileIdInVdc(t, testConfig.VCD.ProviderVdc.StorageProfile)

	var params = StringMap{
		"Org":                testConfig.VCD.Org,
		"Vdc":                testConfig.VCD.Vdc,
		"StorageProfileName": testConfig.VCD.ProviderVdc.StorageProfile,
		"Tags":               "catalog",
	}

	configText := templateFill(testAccVcdStorageProfile, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { preRunChecks(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					// Ensure that ID matches storage profile URN format (e.g. urn:vcloud:vdcstorageProfile:db4aaa49-7593-4416-93df-37235a1a2c68)
					resource.TestMatchResourceAttr("data.vcd_storage_profile.sp", "id", regexp.MustCompile(`^urn:vcloud:vdcstorageProfile:`)),
					resource.TestCheckResourceAttr("data.vcd_storage_profile.sp", "id", storageProfileId),
					resource.TestMatchResourceAttr("data.vcd_storage_profile.sp", "limit", regexp.MustCompile(`\d*`)),
					resource.TestMatchResourceAttr("data.vcd_storage_profile.sp", "used_storage", regexp.MustCompile(`\d*`)),
					resource.TestMatchResourceAttr("data.vcd_storage_profile.sp", "default", regexp.MustCompile(`\s*`)),
					resource.TestCheckResourceAttr("data.vcd_storage_profile.sp", "enabled", "true"),
					resource.TestCheckResourceAttr("data.vcd_storage_profile.sp", "enabled", "true"),
					resource.TestMatchResourceAttr("data.vcd_storage_profile.sp", "iops_allocated", regexp.MustCompile(`\d*`)),
					resource.TestMatchResourceAttr("data.vcd_storage_profile.sp", "units", regexp.MustCompile(`\s*`)),
					resource.TestMatchResourceAttr("data.vcd_storage_profile.sp", "iops_settings.iops_limiting_enabled", regexp.MustCompile(`\s*`)),
					resource.TestMatchResourceAttr("data.vcd_storage_profile.sp", "iops_settings.maximum_disk_iops", regexp.MustCompile(`\d*`)),
					resource.TestMatchResourceAttr("data.vcd_storage_profile.sp", "iops_settings.default_disk_iops", regexp.MustCompile(`\d*`)),
					resource.TestMatchResourceAttr("data.vcd_storage_profile.sp", "iops_settings.disk_iops_per_gb_max", regexp.MustCompile(`\d*`)),
					resource.TestMatchResourceAttr("data.vcd_storage_profile.sp", "iops_settings.iops_limit", regexp.MustCompile(`\d*`)),
					checkStorageProfileOriginatesInParentVdc("data.vcd_storage_profile.sp",
						params["StorageProfileName"].(string),
						params["Org"].(string),
						params["Vdc"].(string)),
				),
			},
		},
	})
	postTestChecks(t)
}

// checkStorageProfileOriginatesInParentVdc tries to evaluate reverse order and ensure that the found storage profile ID
// does really belong to Org and Vdc specified in datasource
func checkStorageProfileOriginatesInParentVdc(resource, storageProfileName, orgName, vdcName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("resource not found: %s", resource)
		}
		// Lookup ID field of resource
		resourceId := rs.Primary.ID

		vcdClient := createTemporaryVCDConnection(false)
		adminOrg, err := vcdClient.GetAdminOrgByName(orgName)
		if err != nil {
			return fmt.Errorf("error getting adminOrg: %s", err)
		}

		// Retrieve VDCs
		allVdcs, err := adminOrg.GetAllVDCs(false)
		if err != nil {
			return fmt.Errorf("error getting adminOrg: %s", err)
		}

		// Check if in any of Orgs child VDCs there is a storage profile that would match ID, Name and Vdc name
		for _, vdc := range allVdcs {
			for _, storageProfile := range vdc.Vdc.VdcStorageProfiles.VdcStorageProfile {
				if storageProfile.ID == resourceId && storageProfile.Name == storageProfileName && vdc.Vdc.Name == vdcName {
					return nil
				}
			}
		}

		return fmt.Errorf("could not validate storage profile '%s' with ID '%s' belongs to VDC '%s",
			storageProfileName, resourceId, testConfig.VCD.Vdc)
	}
}

// findStorageProfileIdInVdc should find storage profile ID using the ID that comes from data source
func findStorageProfileIdInVdc(t *testing.T, storageProfileName string) string {
	vcdClient := createTemporaryVCDConnection(false)
	adminOrg, err := vcdClient.GetAdminOrgByName(testConfig.VCD.Org)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	// get all child VDCs
	childVdcs, err := adminOrg.GetAllVDCs(false)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	// Find VDC which should own storage profile defined in test configuration
	var foundVdc *govcd.Vdc
	for _, vdc := range childVdcs {
		// We are looking in correct VDC
		if vdc.Vdc.Name == testConfig.VCD.Vdc {
			foundVdc = vdc

		}
	}
	if foundVdc == nil {
		t.Errorf("unable to find VDC '%s'", testConfig.VCD.Vdc)
		t.FailNow()
	}

	// Search for storage profile in found VDC and return its ID
	for _, vdcStorageProfile := range foundVdc.Vdc.VdcStorageProfiles.VdcStorageProfile {
		if vdcStorageProfile.Name == storageProfileName {
			return vdcStorageProfile.ID
		}
	}

	// Did not find ID - return empty value
	return ""
}

const testAccVcdStorageProfile = `
data "vcd_storage_profile" "sp" {
  org  = "{{.Org}}"
  vdc  = "{{.Vdc}}"
  name = "{{.StorageProfileName}}"
}
`
