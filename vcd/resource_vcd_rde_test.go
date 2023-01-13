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

// TODO Create RDE with wrong JSON
//
//	Delete it, should be ok
//	Create again with wrong JSON
//	Update with correct JSON
//	Delete it and create a new with good JSON
//	Import
func TestAccVcdRde(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	var params = StringMap{
		"Namespace":  "namespace",
		"Version":    "1.0.0",
		"Vendor":     "vendor",
		"Name":       t.Name(),
		"SchemaPath": getCurrentDir() + "/../test-resources/rde_type.json",
		"EntityPath": getCurrentDir() + "/../test-resources/rde_instance.json",
		"EntityUrl":  "https://raw.githubusercontent.com/adambarreiro/terraform-provider-vcd/add-rde-support-3/test-resources/rde_instance.json", // FIXME
	}
	testParamsNotEmpty(t, params)

	configTextCreate := templateFill(testAccVcdRde, params)
	params["FuncName"] = t.Name() + "-Update"
	params["Name"] = params["FuncName"]
	configTextUpdate := templateFill(testAccVcdRde, params)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	debugPrintf("#[DEBUG] CONFIGURATION create: %s\n", configTextCreate)
	debugPrintf("#[DEBUG] CONFIGURATION update: %s\n", configTextUpdate)

	rdeType := "vcd_rde_type.rde-type"
	rdeFromFile := "vcd_rde.rde-file"
	rdeFromUrl := "vcd_rde.rde-url"
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckRdeDestroy(rdeType, rdeFromFile, rdeFromUrl),
		Steps: []resource.TestStep{
			{
				Config: configTextCreate,
				Check:  resource.ComposeTestCheckFunc(),
			},
			{
				ResourceName:      rdeFromFile,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdRde(params["Name"].(string)+"file", params["Vendor"].(string), params["Namespace"].(string), params["Version"].(string)),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdRde = `
data "vcd_rde_interface" "existing-interface" {
  namespace = "k8s"
  version   = "1.0.0"
  vendor    = "vmware"
}

resource "vcd_rde_type" "rde-type" {
  namespace     = "{{.Namespace}}"
  version       = "{{.Version}}"
  vendor        = "{{.Vendor}}"
  name          = "{{.Name}}-type"
  schema        = file("{{.SchemaPath}}")
}

resource "vcd_rde" "rde-file" {
  rde_type_id   = vcd_rde_type.rde-type.id
  name          = "{{.Name}}file"
  entity        = file("{{.EntityPath}}")
}

resource "vcd_rde" "rde-url" {
  rde_type_id   = vcd_rde_type.rde-type.id
  name          = "{{.Name}}url"
  entity_url    = "{{.EntityUrl}}"
}
`

// testAccCheckRdeDestroy checks that the RDE instances defined by their identifiers no longer
// exist in VCD.
func testAccCheckRdeDestroy(rdeTypeId string, identifiers ...string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, identifier := range identifiers {
			rdeTypeRes, ok := s.RootModule().Resources[rdeTypeId]
			if !ok {
				return fmt.Errorf("not found: %s", identifier)
			}

			if rdeTypeRes.Primary.ID == "" {
				return fmt.Errorf("no RDE type ID is set")
			}

			conn := testAccProvider.Meta().(*VCDClient)

			rdeType, err := conn.VCDClient.GetRdeTypeById(rdeTypeRes.Primary.ID)

			if err != nil {
				return fmt.Errorf("could not retrieve RDE type %s to destroy its instances: %s", rdeTypeRes.Primary.ID, err)
			}

			_, err = rdeType.GetRdeById(identifier)

			if err == nil || !govcd.ContainsNotFound(err) {
				return fmt.Errorf("RDE %s not deleted yet", identifier)
			}
		}
		return nil
	}
}

func importStateIdRde(name, vendor, namespace, version string) resource.ImportStateIdFunc {
	return func(*terraform.State) (string, error) {
		return name +
			ImportSeparator +
			vendor +
			ImportSeparator +
			namespace +
			ImportSeparator +
			version, nil
	}
}
