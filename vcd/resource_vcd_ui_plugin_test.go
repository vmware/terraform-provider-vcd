//go:build uiPlugin || ALL || functional

package vcd

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"regexp"
	"testing"
)

func init() {
	testingTags["uiPlugin"] = "resource_vcd_ui_plugin_test.go"
}

// This object is equivalent to the manifest.json that is inside the ../test-resources/ui_plugin.zip file
var testUIPluginMetadata = &types.UIPluginMetadata{
	Vendor:         "VMware",
	License:        "BSD-2-Clause",
	Link:           "http://www.vmware.com",
	PluginName:     "Test Plugin",
	Version:        "1.2.3",
	Description:    "Test Plugin description",
	ProviderScoped: true,
	TenantScoped:   true,
}

func TestAccVcdUiPlugin(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	var params = StringMap{
		"Org1":           testConfig.VCD.Org,
		"Org2":           testConfig.Provider.SysOrg,
		"Enabled":        "true",
		"PluginPath":     testConfig.Media.UiPluginPath,
		"TenantIds":      "tenant_ids = [data.vcd_org.org1.id, data.vcd_org.org2.id]",
		"ProviderScoped": " ",
		"TenantScoped":   " ",
		"FuncName":       t.Name(),
	}
	testParamsNotEmpty(t, params)

	step1Config := templateFill(testAccVcdUiPlugin, params)
	params["FuncName"] = t.Name() + "Step2"
	params["TenantIds"] = "tenant_ids = [data.vcd_org.org1.id]"
	step2Config := templateFill(testAccVcdUiPlugin, params)
	params["FuncName"] = t.Name() + "Step3"
	params["Enabled"] = "false"
	params["TenantIds"] = " "
	step3Config := templateFill(testAccVcdUiPlugin, params)
	params["FuncName"] = t.Name() + "Step4"
	params["Enabled"] = "true"
	params["TenantIds"] = "tenant_ids = [data.vcd_org.org1.id, data.vcd_org.org2.id]"
	params["ProviderScoped"] = "provider_scoped = false"
	params["TenantScoped"] = "tenant_scoped = false"
	step4Config := templateFill(testAccVcdUiPlugin, params)
	params["FuncName"] = t.Name() + "Step7"
	params["SkipBinary"] = "# skip-binary-test"
	step5Config := templateFill(testAccVcdUiPluginDS+testAccVcdUiPlugin, params)

	resourceName := "vcd_ui_plugin.plugin"
	dsName := "data.vcd_ui_plugin.pluginDS"

	debugPrintf("#[DEBUG] CONFIGURATION Step 1: %s\n", step1Config)
	debugPrintf("#[DEBUG] CONFIGURATION Step 2: %s\n", step2Config)
	debugPrintf("#[DEBUG] CONFIGURATION Step 3: %s\n", step3Config)
	debugPrintf("#[DEBUG] CONFIGURATION Step 4: %s\n", step4Config)
	debugPrintf("#[DEBUG] CONFIGURATION Step 5: %s\n", step5Config)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	cachedId := &testCachedFieldValue{}

	testCheckResourceCommonUIPluginAsserts := func(resourcePath string) resource.TestCheckFunc {
		return resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttr(resourcePath, "vendor", testUIPluginMetadata.Vendor),
			resource.TestCheckResourceAttr(resourcePath, "name", testUIPluginMetadata.PluginName),
			resource.TestCheckResourceAttr(resourcePath, "version", testUIPluginMetadata.Version),
			resource.TestCheckResourceAttr(resourcePath, "license", testUIPluginMetadata.License),
			resource.TestCheckResourceAttr(resourcePath, "description", testUIPluginMetadata.Description),
			resource.TestCheckResourceAttr(resourcePath, "link", testUIPluginMetadata.Link),
			resource.TestMatchResourceAttr(resourcePath, "status", regexp.MustCompile("^ready|unavailable$")),
		)
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckUIPluginDestroy(cachedId.fieldValue),
		Steps: []resource.TestStep{
			// Test UI Plugin creation with 2 tenants and enabled
			{
				Config: step1Config,
				Check: resource.ComposeAggregateTestCheckFunc(
					testCheckResourceCommonUIPluginAsserts(resourceName),
					resource.TestCheckResourceAttr(resourceName, "provider_scoped", "true"),
					resource.TestCheckResourceAttr(resourceName, "tenant_scoped", "true"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "tenant_ids.#", "2"),
				),
			},
			// Test UI Plugin update to unpublish one tenant
			{
				Config: step2Config,
				Check: resource.ComposeAggregateTestCheckFunc(
					testCheckResourceCommonUIPluginAsserts(resourceName),
					resource.TestCheckResourceAttr(resourceName, "provider_scoped", "true"),
					resource.TestCheckResourceAttr(resourceName, "tenant_scoped", "true"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "tenant_ids.#", "1"),
				),
			},
			// Test UI Plugin update to unpublish all tenants
			{
				Config: step3Config,
				Check: resource.ComposeAggregateTestCheckFunc(
					testCheckResourceCommonUIPluginAsserts(resourceName),
					resource.TestCheckResourceAttr(resourceName, "provider_scoped", "true"),
					resource.TestCheckResourceAttr(resourceName, "tenant_scoped", "true"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "tenant_ids.#", "0"),
				),
			},
			// Test UI Plugin update
			{
				Config: step4Config,
				Check: resource.ComposeAggregateTestCheckFunc(
					testCheckResourceCommonUIPluginAsserts(resourceName),
					resource.TestCheckResourceAttr(resourceName, "provider_scoped", "false"),
					resource.TestCheckResourceAttr(resourceName, "tenant_scoped", "false"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "tenant_ids.#", "2"),
				),
			},
			// Test UI Plugin data source
			{
				Config: step5Config,
				Check: resource.ComposeAggregateTestCheckFunc(
					testCheckResourceCommonUIPluginAsserts(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "vendor", dsName, "vendor"),
					resource.TestCheckResourceAttrPair(resourceName, "name", dsName, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "version", dsName, "version"),
					resource.TestCheckResourceAttrPair(resourceName, "license", dsName, "license"),
					resource.TestCheckResourceAttrPair(resourceName, "description", dsName, "description"),
					resource.TestCheckResourceAttrPair(resourceName, "link", dsName, "link"),
					resource.TestCheckResourceAttrPair(resourceName, "provider_scoped", dsName, "provider_scoped"),
					resource.TestCheckResourceAttrPair(resourceName, "tenant_scoped", dsName, "tenant_scoped"),
					resource.TestCheckResourceAttrPair(resourceName, "enabled", dsName, "enabled"),
					resource.TestCheckResourceAttrPair(resourceName, "link", dsName, "link"),
					resource.TestCheckResourceAttrPair(resourceName, "tenant_ids.#", dsName, "tenant_ids.#"),
					func(state *terraform.State) error {
						return nil
					},
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdFunc:       importStateIdUIPlugin(testUIPluginMetadata.Vendor, testUIPluginMetadata.PluginName, testUIPluginMetadata.Version),
				ImportStateVerifyIgnore: []string{"plugin_path"},
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdUiPlugin = `
data "vcd_org" "org1" {
  name = "{{.Org1}}"
}

data "vcd_org" "org2" {
  name = "{{.Org2}}"
}

resource "vcd_ui_plugin" "plugin" {
  plugin_path = "{{.PluginPath}}"
  enabled     = {{.Enabled}}

  {{.TenantIds}}
  {{.ProviderScoped}}
  {{.TenantScoped}}
}
`

const testAccVcdUiPluginDS = `
# skip-binary-test - Data source referencing the same resource
data "vcd_ui_plugin" "pluginDS" {
  vendor  = "VMware"
  name    = "Test Plugin"
  version = "1.2.3"
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

func importStateIdUIPlugin(vendor, name, version string) resource.ImportStateIdFunc {
	return func(*terraform.State) (string, error) {
		return vendor +
			ImportSeparator +
			name +
			ImportSeparator +
			version, nil
	}
}
