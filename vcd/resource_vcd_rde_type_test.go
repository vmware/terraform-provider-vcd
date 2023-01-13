//go:build rde || ALL || functional

package vcd

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"regexp"
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
		"Description":         "Created by " + t.Name(),
		"InterfaceReferences": "vcd_rde_interface.rde-interface1.id",
		"SchemaPath":          getCurrentDir() + "/../test-resources/rde_type.json",                                                                   // TODO: Parameterize this value???
		"SchemaUrl":           "https://raw.githubusercontent.com/adambarreiro/terraform-provider-vcd/add-rde-support-2/test-resources/rde_type.json", // FIXME
	}
	testParamsNotEmpty(t, params)

	configTextCreate := templateFill(testAccVcdRdeType, params)
	params["FuncName"] = t.Name() + "-Update"
	params["Name"] = params["FuncName"]
	params["Description"] = "Created by" + params["FuncName"].(string)
	params["InterfaceReferences"] = "vcd_rde_interface.rde-interface1.id, vcd_rde_interface.rde-interface2.id"
	configTextUpdate := templateFill(testAccVcdRdeType, params)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	debugPrintf("#[DEBUG] CONFIGURATION create: %s\n", configTextCreate)
	debugPrintf("#[DEBUG] CONFIGURATION update: %s\n", configTextUpdate)

	rdeTypeFromFile := "vcd_rde_type.rde-type-file"
	rdeTypeFromUrl := "vcd_rde_type.rde-type-url"
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckRdeTypesDestroy(rdeTypeFromFile, rdeTypeFromUrl),
		Steps: []resource.TestStep{
			{
				Config: configTextCreate,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(rdeTypeFromFile, "namespace", params["Namespace"].(string)+"file"),
					resource.TestCheckResourceAttr(rdeTypeFromFile, "version", params["Version"].(string)),
					resource.TestCheckResourceAttr(rdeTypeFromFile, "vendor", params["Vendor"].(string)+"file"),
					resource.TestCheckResourceAttr(rdeTypeFromFile, "name", t.Name()),
					resource.TestCheckResourceAttr(rdeTypeFromFile, "description", "Created by "+t.Name()),
					resource.TestCheckResourceAttrPair(rdeTypeFromFile, "interface_ids.0", "vcd_rde_interface.rde-interface1", "id"),
					resource.TestMatchResourceAttr(rdeTypeFromFile, "schema", regexp.MustCompile("{.*\"foo\".*\"bar\".*}")),

					resource.TestCheckResourceAttr(rdeTypeFromUrl, "namespace", params["Namespace"].(string)+"url"),
					resource.TestCheckResourceAttr(rdeTypeFromUrl, "vendor", params["Vendor"].(string)+"url"),

					resource.TestCheckResourceAttrPair(rdeTypeFromUrl, "version", rdeTypeFromFile, "version"),
					resource.TestCheckResourceAttrPair(rdeTypeFromUrl, "name", rdeTypeFromFile, "name"),
					resource.TestCheckResourceAttrPair(rdeTypeFromUrl, "description", rdeTypeFromFile, "description"),
					resource.TestCheckResourceAttrPair(rdeTypeFromUrl, "interface_ids.0", rdeTypeFromFile, "interface_ids.0"),
					resource.TestCheckResourceAttrPair(rdeTypeFromUrl, "schema", rdeTypeFromFile, "schema"),
				),
			},
			{
				Config: configTextUpdate,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(rdeTypeFromFile, "namespace", params["Namespace"].(string)+"file"),
					resource.TestCheckResourceAttr(rdeTypeFromFile, "version", params["Version"].(string)),
					resource.TestCheckResourceAttr(rdeTypeFromFile, "vendor", params["Vendor"].(string)+"file"),
					resource.TestCheckResourceAttr(rdeTypeFromFile, "name", t.Name()+"-Update"),
					resource.TestCheckResourceAttr(rdeTypeFromFile, "description", "Created by"+t.Name()+"-Update"),
					resource.TestCheckResourceAttr(rdeTypeFromFile, "interface_ids.#", "2"),

					resource.TestCheckResourceAttr(rdeTypeFromUrl, "namespace", params["Namespace"].(string)+"url"),
					resource.TestCheckResourceAttr(rdeTypeFromUrl, "vendor", params["Vendor"].(string)+"url"),

					resource.TestCheckResourceAttrPair(rdeTypeFromUrl, "version", rdeTypeFromFile, "version"),
					resource.TestCheckResourceAttrPair(rdeTypeFromUrl, "name", rdeTypeFromFile, "name"),
					resource.TestCheckResourceAttrPair(rdeTypeFromUrl, "description", rdeTypeFromFile, "description"),
					resource.TestCheckResourceAttrPair(rdeTypeFromUrl, "interface_ids.#", rdeTypeFromFile, "interface_ids.#"),
					resource.TestCheckResourceAttrPair(rdeTypeFromUrl, "schema", rdeTypeFromFile, "schema"),
				),
			},
			{
				ResourceName:      rdeTypeFromFile,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdDefinedInterface(params["Vendor"].(string)+"file", params["Namespace"].(string)+"file", params["Version"].(string)),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdRdeType = `
resource "vcd_rde_interface" "rde-interface1" {
  namespace = "namespace1"
  version   = "1.0.0"
  vendor    = "vendor1"
  name      = "name1"
}

resource "vcd_rde_interface" "rde-interface2" {
  namespace   = "namespace2"
  version     = "2.0.0"
  vendor      = "vendor2"
  name        = "name2"
}

resource "vcd_rde_type" "rde-type-file" {
  namespace     = "{{.Namespace}}file"
  version       = "{{.Version}}"
  vendor        = "{{.Vendor}}file"
  name          = "{{.Name}}"
  description   = "{{.Description}}"
  interface_ids = [{{.InterfaceReferences}}]
  schema        = file("{{.SchemaPath}}")
}

resource "vcd_rde_type" "rde-type-url" {
  namespace     = "{{.Namespace}}url"
  version       = "{{.Version}}"
  vendor        = "{{.Vendor}}url"
  name          = "{{.Name}}"
  description   = "{{.Description}}"
  interface_ids = [{{.InterfaceReferences}}]
  schema_url    = "{{.SchemaUrl}}"
}
`

// testAccCheckRdeTypeDestroy checks that the RDE type defined by its identifier no longer
// exists in VCD.
func testAccCheckRdeTypesDestroy(identifiers ...string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, identifier := range identifiers {
			rs, ok := s.RootModule().Resources[identifier]
			if !ok {
				return fmt.Errorf("not found: %s", identifier)
			}

			if rs.Primary.ID == "" {
				return fmt.Errorf("no RDE type ID is set")
			}

			conn := testAccProvider.Meta().(*VCDClient)

			_, err := conn.VCDClient.GetRdeTypeById(rs.Primary.ID)

			if err == nil || !govcd.ContainsNotFound(err) {
				return fmt.Errorf("%s not deleted yet", identifier)
			}
		}
		return nil
	}
}
