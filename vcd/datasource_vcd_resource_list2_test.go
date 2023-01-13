//go:build ALL || functional

package vcd

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

// TestResourceInfoProviders tests the feasibility of using multiple providers in the same script.
func TestResourceInfoProviders(t *testing.T) {
	preTestChecks(t)

	// This test requires system administrator because it runs the same data source
	// as sysAdmin and as Org user, using different provider definitions
	skipIfNotSysAdmin(t)

	org1 := testConfig.VCD.Org
	org2 := testConfig.VCD.Org + "-1"

	if org1 == "" {
		t.Skip("org name missing from configuration file")
	}

	// We should be careful to use the appropriate provider names for ProviderFactories
	var data = StringMap{
		"ProviderSystem": providerVcdSystem,
		"ProviderOrg1":   providerVcdOrg1,
		"ProviderOrg2":   providerVcdOrg2,
	}
	var configText string = templateFill(testAccCheckVcdDatasourceInfoProvider, data)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	// NOTE: the text displayed here and the text saved to test-artifacts file are different.
	// This is the text used by the text framework, with specific names for each provider,
	// while the text in the file contain properly aliased provider names
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

	resource.Test(t, resource.TestCase{
		// buildMultipleProviders creates three providers, which run in the test framework with separate
		// names (as ProviderFactories docs say, the framework doesn't support aliases)
		// https://pkg.go.dev/github.com/hashicorp/terraform-plugin-sdk/v2@v2.24.1/helper/resource#TestCase
		ProviderFactories: buildMultipleProviders(),
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					// logState prints the state in the provider's log.
					// Inspect the log to make sure that the state of the three data sources
					// are what we expect
					logState(t.Name()),
					// Running as system administrator, we find every Org
					checkListForKnownItem("orgs_system", "System", true),
					checkListForKnownItem("orgs_system", org1, true),
					checkListForKnownItem("orgs_system", org2, true),
					// Running as Org1 user, we only find the current org and fail to find any others
					checkListForKnownItem("orgs1", org1, true),
					checkListForKnownItem("orgs1", org2, false),
					checkListForKnownItem("orgs1", "System", false),
					// Running as Org2 user, we only find Org2, but not Org1, or System
					checkListForKnownItem("orgs2", org2, true),
					checkListForKnownItem("orgs2", org1, false),
					checkListForKnownItem("orgs2", "System", false),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccCheckVcdDatasourceInfoProvider = `
data "vcd_resource_list" "orgs_system" {
  provider      = {{.ProviderSystem}}
  name          = "orgs"
  resource_type = "vcd_org"
}

data "vcd_resource_list" "orgs1" {
  provider      = {{.ProviderOrg1}}
  name          = "orgs1"
  resource_type = "vcd_org"
}

data "vcd_resource_list" "orgs2" {
  provider      = {{.ProviderOrg2}}
  name          = "orgs2"
  resource_type = "vcd_org"
}
`
