//go:build plugin || ALL || functional

package vcd

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"regexp"
	"testing"
)

func init() {
	testingTags["plugin"] = "resource_vcd_ui_plugin_test.go"
}

func TestAccVcdUiPlugin(t *testing.T) {

	var params = StringMap{
		"Org1":                testConfig.VCD.Org,
		"Org2":                testConfig.Provider.SysOrg,
		"Enabled":             "true",
		"PluginPath":          "../test-resources/ui_plugin.zip",
		"PublishToAllTenants": "true",
		"PublishedTenantIds":  " ",
		"FuncName":            t.Name() + "Step1",
	}
	testParamsNotEmpty(t, params)

	step1Config := templateFill(testAccVcdUiPluginStep1, params)
	params["FuncName"] = t.Name() + "Step2"
	params["PublishToAllTenants"] = "false"
	params["PublishedTenantIds"] = "published_tenant_ids = [data.vcd_org.org1.id, data.vcd_org.org2.id]"
	step2Config := templateFill(testAccVcdUiPluginStepOrgs+testAccVcdUiPluginStep1, params)

	plugin1 := "vcd_ui_plugin.plugin1"

	debugPrintf("#[DEBUG] CONFIGURATION Step 1: %s\n", step1Config)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	cachedId := &testCachedFieldValue{}

	testCheckResourceCommonUIPluginAsserts := func(resourcePath string) resource.TestCheckFunc {
		return resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttr(resourcePath, "vendor", "VMware"),
			resource.TestCheckResourceAttr(resourcePath, "name", "Test Plugin"),
			resource.TestCheckResourceAttr(resourcePath, "version", "1.2.3"),
			resource.TestCheckResourceAttr(resourcePath, "license", "BSD-2-Clause"),
			resource.TestCheckResourceAttr(resourcePath, "description", "Test Plugin description"),
			resource.TestCheckResourceAttr(resourcePath, "link", "http://www.vmware.com"),
		)
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckUIPluginDestroy(cachedId.fieldValue),
		Steps: []resource.TestStep{
			// Test UI Plugin creation with publish to all tenants and enabled
			{
				Config: step1Config,
				Check: resource.ComposeAggregateTestCheckFunc(
					testCheckResourceCommonUIPluginAsserts(plugin1),
					resource.TestCheckResourceAttr(plugin1, "enabled", "true"),
					resource.TestCheckResourceAttr(plugin1, "publish_to_all_tenants", "true"),
					resource.TestMatchResourceAttr(plugin1, "published_tenant_ids.#", regexp.MustCompile(`^[1-9]+$`)),
				),
			},
			// Test UI Plugin creation (we taint it for that) with publish to only specific tenants and enabled
			{
				Config: step2Config,
				Taint:  []string{plugin1},
				Check: resource.ComposeAggregateTestCheckFunc(
					testCheckResourceCommonUIPluginAsserts(plugin1),
					resource.TestCheckResourceAttr(plugin1, "enabled", "true"),
					resource.TestCheckResourceAttr(plugin1, "publish_to_all_tenants", "false"),
					resource.TestMatchResourceAttr(plugin1, "published_tenant_ids.#", regexp.MustCompile("2")),
				),
			},
		},
	})
	postTestChecks(t)
}

// Test UI Plugin creation with publish to all tenants
const testAccVcdUiPluginStep1 = `
resource "vcd_ui_plugin" "plugin1" {
  plugin_path            = "{{.PluginPath}}"
  enabled                = {{.Enabled}}
  publish_to_all_tenants = {{.PublishToAllTenants}}
  {{.PublishedTenantIds}}
}
`

const testAccVcdUiPluginStepOrgs = `
data "vcd_org" "org1" {
  name = "{{.Org1}}"
}

data "vcd_org" "org2" {
  name = "{{.Org2}}"
}
`

func testAccCheckUIPluginDestroy(id string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*VCDClient)
		_, err := conn.GetUIPluginById(id)
		if err == nil {
			return fmt.Errorf("UI Plugin %s still exists", id)
		}
		return nil
	}
}
