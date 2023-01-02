//go:build rde || ALL || functional
// +build rde ALL functional

package vcd

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"testing"
)

func TestAccVcdRdeDefinedInterface(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	var params = StringMap{
		"Namespace": "namespace1",
		"Version":   "1.0.0",
		"Vendor":    "vendor1",
		"Name":      t.Name(),
		"Readonly":  "false",
	}
	testParamsNotEmpty(t, params)

	configTextCreate := templateFill(testAccVcdRdeDefinedInterface, params)
	params["FuncName"] = t.Name() + "-Update"
	params["Name"] = params["FuncName"]
	configTextUpdate := templateFill(testAccVcdRdeDefinedInterface, params)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	debugPrintf("#[DEBUG] CONFIGURATION create: %s\n", configTextCreate)
	debugPrintf("#[DEBUG] CONFIGURATION update: %s\n", configTextUpdate)

	interfaceName := "vcd_rde_interface.interface1"
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckRdeInterfaceDestroy(interfaceName),
		Steps: []resource.TestStep{
			{
				Config: configTextCreate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(interfaceName, "namespace", params["Namespace"].(string)),
					resource.TestCheckResourceAttr(interfaceName, "version", params["Version"].(string)),
					resource.TestCheckResourceAttr(interfaceName, "vendor", params["Vendor"].(string)),
					resource.TestCheckResourceAttr(interfaceName, "name", t.Name()),
					resource.TestCheckResourceAttr(interfaceName, "readonly", params["Readonly"].(string)),
				),
			},
			{
				Config: configTextUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(interfaceName, "namespace", params["Namespace"].(string)),
					resource.TestCheckResourceAttr(interfaceName, "version", params["Version"].(string)),
					resource.TestCheckResourceAttr(interfaceName, "vendor", params["Vendor"].(string)),
					resource.TestCheckResourceAttr(interfaceName, "name", t.Name()+"-Update"),
					resource.TestCheckResourceAttr(interfaceName, "readonly", params["Readonly"].(string)),
				),
			},
			{
				ResourceName:      interfaceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdDefinedInterface(params["Vendor"].(string), params["Namespace"].(string), params["Version"].(string)),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdRdeDefinedInterface = `
resource "vcd_rde_interface" "interface1" {
  namespace = "{{.Namespace}}"
  version   = "{{.Version}}"
  vendor    = "{{.Vendor}}"
  name      = "{{.Name}}"
  readonly  = {{.Readonly}} 
}
`

// testAccCheckRdeInterfaceDestroy checks that the Defined Interface defined by its identifier no longer
// exists in VCD.
func testAccCheckRdeInterfaceDestroy(identifier string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[identifier]
		if !ok {
			return fmt.Errorf("not found: %s", identifier)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no Defined Interface ID is set")
		}

		conn := testAccProvider.Meta().(*VCDClient)

		_, err := conn.VCDClient.GetDefinedInterfaceById(rs.Primary.ID)

		if err == nil || !govcd.ContainsNotFound(err) {
			return fmt.Errorf("%s not deleted yet", identifier)
		}
		return nil

	}
}

func importStateIdDefinedInterface(vendor, namespace, version string) resource.ImportStateIdFunc {
	return func(*terraform.State) (string, error) {
		return vendor +
			ImportSeparator +
			namespace +
			ImportSeparator +
			version, nil
	}
}
