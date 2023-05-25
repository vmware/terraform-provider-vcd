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
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	var params = StringMap{
		"Org1":                  testConfig.VCD.Org,
		"Org2":                  testConfig.Provider.SysOrg,
		"Enabled":               "true",
		"PluginPath":            "../test-resources/ui_plugin.zip",
		"PublishedToAllTenants": "publish_to_all_tenants = true",
		"PublishedTenantIds":    " ",
		"ProviderScoped":        " ",
		"TenantScoped":          " ",
		"FuncName":              t.Name(),
	}
	testParamsNotEmpty(t, params)

	step1Config := templateFill(testAccVcdUiPlugin, params)
	params["FuncName"] = t.Name() + "Step2"
	params["PublishedToAllTenants"] = " "
	params["PublishedTenantIds"] = "published_tenant_ids = [data.vcd_org.org1.id, data.vcd_org.org2.id]"
	step2Config := templateFill(testAccVcdUiPlugin, params)
	params["FuncName"] = t.Name() + "Step3"
	params["Enabled"] = "false"
	params["PublishedToAllTenants"] = "publish_to_all_tenants = true"
	params["PublishedTenantIds"] = " "
	step3Config := templateFill(testAccVcdUiPlugin, params)
	params["FuncName"] = t.Name() + "Step4"
	params["Enabled"] = "false"
	params["PublishedToAllTenants"] = " "
	params["PublishedTenantIds"] = "published_tenant_ids = [data.vcd_org.org1.id, data.vcd_org.org2.id]"
	step4Config := templateFill(testAccVcdUiPlugin, params)
	params["FuncName"] = t.Name() + "Step5"
	params["Enabled"] = "true"
	params["PublishedToAllTenants"] = "publish_to_all_tenants = true"
	params["PublishedTenantIds"] = " "
	params["ProviderScoped"] = "provider_scoped = false"
	params["TenantScoped"] = "tenant_scoped = false"
	step5Config := templateFill(testAccVcdUiPlugin, params)
	params["FuncName"] = t.Name() + "Step6"
	params["Enabled"] = "false"
	params["PublishedToAllTenants"] = " "
	params["PublishedTenantIds"] = "published_tenant_ids = [data.vcd_org.org1.id, data.vcd_org.org2.id]"
	params["ProviderScoped"] = "provider_scoped = true"
	params["TenantScoped"] = "tenant_scoped = true"
	step6Config := templateFill(testAccVcdUiPlugin, params)
	params["FuncName"] = t.Name() + "Step7"
	params["PublishedToAllTenants"] = "publish_to_all_tenants = false"
	step7Config := templateFill(testAccVcdUiPlugin, params)
	params["FuncName"] = t.Name() + "Step8"
	params["SkipBinary"] = "# skip-binary-test"
	step8Config := templateFill(testAccVcdUiPluginDS+testAccVcdUiPlugin, params)

	resourceName := "vcd_ui_plugin.plugin"
	dsName := "data.vcd_ui_plugin.pluginDS"

	debugPrintf("#[DEBUG] CONFIGURATION Step 1: %s\n", step1Config)
	debugPrintf("#[DEBUG] CONFIGURATION Step 2: %s\n", step2Config)
	debugPrintf("#[DEBUG] CONFIGURATION Step 3: %s\n", step3Config)
	debugPrintf("#[DEBUG] CONFIGURATION Step 4: %s\n", step4Config)
	debugPrintf("#[DEBUG] CONFIGURATION Step 5: %s\n", step5Config)
	debugPrintf("#[DEBUG] CONFIGURATION Step 6: %s\n", step6Config)
	debugPrintf("#[DEBUG] CONFIGURATION Step 7: %s\n", step7Config)
	debugPrintf("#[DEBUG] CONFIGURATION Step 8: %s\n", step8Config)
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
					testCheckResourceCommonUIPluginAsserts(resourceName),
					resource.TestCheckResourceAttr(resourceName, "provider_scoped", "true"),
					resource.TestCheckResourceAttr(resourceName, "tenant_scoped", "true"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "publish_to_all_tenants", "true"),
					resource.TestMatchResourceAttr(resourceName, "published_tenant_ids.#", regexp.MustCompile(`^[1-9]+$`)),
				),
			},
			// Test UI Plugin creation (we taint it for that) with publish to only specific tenants and enabled
			{
				Config: step2Config,
				Taint:  []string{resourceName},
				Check: resource.ComposeAggregateTestCheckFunc(
					testCheckResourceCommonUIPluginAsserts(resourceName),
					resource.TestCheckResourceAttr(resourceName, "provider_scoped", "true"),
					resource.TestCheckResourceAttr(resourceName, "tenant_scoped", "true"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "publish_to_all_tenants", "false"),
					resource.TestMatchResourceAttr(resourceName, "published_tenant_ids.#", regexp.MustCompile("2")),
				),
			},
			// Test UI Plugin creation (we taint it for that) with publish to all tenants and disabled
			{
				Config: step3Config,
				Taint:  []string{resourceName},
				Check: resource.ComposeAggregateTestCheckFunc(
					testCheckResourceCommonUIPluginAsserts(resourceName),
					resource.TestCheckResourceAttr(resourceName, "provider_scoped", "true"),
					resource.TestCheckResourceAttr(resourceName, "tenant_scoped", "true"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "publish_to_all_tenants", "true"),
					resource.TestMatchResourceAttr(resourceName, "published_tenant_ids.#", regexp.MustCompile(`^[1-9]+$`)),
				),
			},
			// Test UI Plugin creation (we taint it for that) with publish only specific tenants and disabled
			{
				Config: step4Config,
				Taint:  []string{resourceName},
				Check: resource.ComposeAggregateTestCheckFunc(
					testCheckResourceCommonUIPluginAsserts(resourceName),
					resource.TestCheckResourceAttr(resourceName, "provider_scoped", "true"),
					resource.TestCheckResourceAttr(resourceName, "tenant_scoped", "true"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "publish_to_all_tenants", "false"),
					resource.TestMatchResourceAttr(resourceName, "published_tenant_ids.#", regexp.MustCompile("2")),
				),
			},
			// Test UI Plugin update
			{
				Config: step5Config,
				Check: resource.ComposeAggregateTestCheckFunc(
					testCheckResourceCommonUIPluginAsserts(resourceName),
					resource.TestCheckResourceAttr(resourceName, "provider_scoped", "false"),
					resource.TestCheckResourceAttr(resourceName, "tenant_scoped", "false"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "publish_to_all_tenants", "true"),
					resource.TestMatchResourceAttr(resourceName, "published_tenant_ids.#", regexp.MustCompile(`^[1-9]+$`)),
				),
			},
			// Test UI Plugin update
			{
				Config: step6Config,
				PreConfig: func() {
					fmt.Printf("a")
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testCheckResourceCommonUIPluginAsserts(resourceName),
					resource.TestCheckResourceAttr(resourceName, "provider_scoped", "true"),
					resource.TestCheckResourceAttr(resourceName, "tenant_scoped", "true"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "publish_to_all_tenants", "false"),
					resource.TestMatchResourceAttr(resourceName, "published_tenant_ids.#", regexp.MustCompile("2")),
				),
			},
			// Test UI Plugin update
			{
				Config: step6Config,
				Check: resource.ComposeAggregateTestCheckFunc(
					testCheckResourceCommonUIPluginAsserts(resourceName),
					resource.TestCheckResourceAttr(resourceName, "provider_scoped", "true"),
					resource.TestCheckResourceAttr(resourceName, "tenant_scoped", "true"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "publish_to_all_tenants", "false"),
					resource.TestCheckResourceAttr(resourceName, "published_tenant_ids.#", "0"),
				),
			},
			// Test UI Plugin data source
			{
				Config: step8Config,
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
					resource.TestCheckResourceAttrPair(resourceName, "published_tenant_ids.#", dsName, "published_tenant_ids.#"),
				),
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
  plugin_path            = "{{.PluginPath}}"
  enabled                = {{.Enabled}}
  {{.PublishedToAllTenants}}
  {{.ProviderScoped}}
  {{.TenantScoped}}
  {{.PublishedTenantIds}}
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
