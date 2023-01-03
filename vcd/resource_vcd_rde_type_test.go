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

func TestAccVcdRdeType(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	var params = StringMap{
		"Namespace":           "namespace",
		"Version":             "1.0.0",
		"Vendor":              "vendor",
		"Name":                t.Name(),
		"InterfaceReferences": "vcd_rde_interface.rde-interface1.id",
		"ExternalId":          "externalId",
		"SchemaPath":          getCurrentDir() + "/../test-resources/rde_type.json", // TODO: Parameterize this value???
	}
	testParamsNotEmpty(t, params)

	configTextCreate := templateFill(testAccVcdRdeType, params)
	params["FuncName"] = t.Name() + "-Update"
	params["Name"] = params["FuncName"]
	params["InterfaceReferences"] = "vcd_rde_interface.rde-interface1.id, vcd_rde_interface.rde-interface2.id"
	configTextUpdate := templateFill(testAccVcdRdeType, params)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	debugPrintf("#[DEBUG] CONFIGURATION create: %s\n", configTextCreate)
	debugPrintf("#[DEBUG] CONFIGURATION update: %s\n", configTextUpdate)

	rdeTypeName := "vcd_rde_type.rde-type"
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckRdeTypeDestroy(rdeTypeName),
		Steps: []resource.TestStep{
			{
				Config: configTextCreate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(rdeTypeName, "namespace", params["Namespace"].(string)),
					resource.TestCheckResourceAttr(rdeTypeName, "version", params["Version"].(string)),
					resource.TestCheckResourceAttr(rdeTypeName, "vendor", params["Vendor"].(string)),
					resource.TestCheckResourceAttr(rdeTypeName, "name", t.Name()),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdRdeType = `
resource "vcd_rde_interface" "rde-interface1" {
  namespace = "{{.Namespace}}1"
  version   = "{{.Version}}"
  vendor    = "{{.Vendor}}1"
  name      = "{{.Name}}1"
}

resource "vcd_rde_interface" "rde-interface2" {
  namespace   = "{{.Namespace}}2"
  version     = "{{.Version}}"
  vendor      = "{{.Vendor}}2"
  name        = "{{.Name}}2"
}

resource "vcd_rde_type" "rde-type" {
  namespace     = "{{.Namespace}}"
  version       = "{{.Version}}"
  vendor        = "{{.Vendor}}"
  name          = "{{.Name}}"
  interface_ids = [ {{ .InterfaceReferences }} ]
  external_id   = "{{.ExternalId}}"
  schema        = file("{{.SchemaPath}}")
}
`

// testAccCheckRdeTypeDestroy checks that the RDE type defined by its identifier no longer
// exists in VCD.
func testAccCheckRdeTypeDestroy(identifier string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[identifier]
		if !ok {
			return fmt.Errorf("not found: %s", identifier)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no RDE ID is set")
		}

		conn := testAccProvider.Meta().(*VCDClient)

		_, err := conn.VCDClient.GetRDETypeById(rs.Primary.ID)

		if err == nil || !govcd.ContainsNotFound(err) {
			return fmt.Errorf("%s not deleted yet", identifier)
		}
		return nil

	}
}
