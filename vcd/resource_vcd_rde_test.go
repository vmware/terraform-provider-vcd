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
		"Namespace":   "namespace",
		"Version":     "1.0.0",
		"Vendor":      "vendor",
		"Name":        t.Name(),
		"Description": "Created by " + t.Name(),
		"SchemaPath":  getCurrentDir() + "/../test-resources/rde_type.json", // TODO: Parameterize this value???
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
  description   = "{{.Description}}"
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
				return fmt.Errorf("RDE type %s is deleted before its RDE instances", rdeTypeRes.Primary.ID)
			}

			_, err = rdeType.GetRdeById(identifier)

			if err == nil || !govcd.ContainsNotFound(err) {
				return fmt.Errorf("RDE %s not deleted yet", identifier)
			}
		}
		return nil
	}
}
