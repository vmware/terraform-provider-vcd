//go:build functional || openapi || ALL

package vcd

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccVcdExternalEndpoint(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	vendor := "vmware"
	name := t.Name()
	version := "1.0.0"

	var params = StringMap{
		"Vendor":      vendor,
		"Name":        name,
		"Version":     version,
		"Description": "Description of" + name,
		"Enabled":     true,
		"RootUrl":     "https://www.broadcom.com",
		"FuncName":    t.Name() + "Step1",
	}
	testParamsNotEmpty(t, params)

	configText1 := templateFill(testAccCheckVcdExternalEndpoint, params)
	debugPrintf("#[DEBUG] CONFIGURATION 1: %s", configText1)

	params["FuncName"] = t.Name() + "Step2"
	params["RootUrl"] = "https://www.vmware.com"
	params["Enabled"] = false // Endpoint needs to be Disabled before destroying it
	configText2 := templateFill(testAccCheckVcdExternalEndpoint, params)
	debugPrintf("#[DEBUG] CONFIGURATION 2: %s", configText1)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resourceName := "vcd_external_endpoint.ep"
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckExternalEndpointDestroy(vendor, name, version),
		Steps: []resource.TestStep{
			{
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "id", fmt.Sprintf("urn:vcloud:extensionEndpoint:%s:%s:%s", vendor, name, version)),
					resource.TestCheckResourceAttr(resourceName, "vendor", vendor),
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "version", version),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "description", params["Description"].(string)),
					resource.TestCheckResourceAttr(resourceName, "root_url", "https://www.broadcom.com"),
				),
			},
			{
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "id", fmt.Sprintf("urn:vcloud:extensionEndpoint:%s:%s:%s", vendor, name, version)),
					resource.TestCheckResourceAttr(resourceName, "vendor", vendor),
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "version", version),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "description", params["Description"].(string)),
					resource.TestCheckResourceAttr(resourceName, "root_url", "https://www.vmware.com"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdTopHierarchy(fmt.Sprintf("%s%s%s%s%s", vendor, ImportSeparator, name, ImportSeparator, version)),
			},
		},
	})
	postTestChecks(t)
}

func testAccCheckExternalEndpointDestroy(vendor, name, version string) func(s *terraform.State) error {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "vcd_external_endpoint" &&
				rs.Primary.Attributes["vendor"] != vendor &&
				rs.Primary.Attributes["name"] != name &&
				rs.Primary.Attributes["version"] != version {
				continue
			}

			conn := testAccProvider.Meta().(*VCDClient)
			_, err := conn.GetExternalEndpointById(rs.Primary.ID)
			if err == nil {
				return fmt.Errorf("external endpoint '%s' still exists", rs.Primary.ID)
			}
		}

		return nil
	}
}

const testAccCheckVcdExternalEndpoint = `
resource "vcd_external_endpoint" "ep" {
  vendor      = "{{.Vendor}}"
  name        = "{{.Name}}"
  version     = "{{.Version}}"
  enabled     = {{.Enabled}}
  description = "{{.Description}}"
  root_url    = "{{.RootUrl}}"
}
`
