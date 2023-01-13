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

// TestAccVcdRde tests the behaviour of RDE instances:
// - Step 1: Create 3 RDEs: One with file, other with URL, last one with wrong JSON.
// - Step 2: Taint to test delete on wrong RDEs and repeat step 1.
// - Step 3: Update one RDE name. Update wrong JSON in RDE.
// - Step 4: Import
func TestAccVcdRde(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	var params = StringMap{
		"FuncName":   t.Name() + "-Step1-and-2",
		"Namespace":  "namespace",
		"Version":    "1.0.0",
		"Vendor":     "vendor",
		"Name":       t.Name(),
		"SchemaPath": getCurrentDir() + "/../test-resources/rde_type.json",
		"EntityPath": getCurrentDir() + "/../test-resources/rde_instance.json",
		"EntityUrl":  "https://raw.githubusercontent.com/adambarreiro/terraform-provider-vcd/add-rde-support-3/test-resources/rde_instance.json", // FIXME
	}
	testParamsNotEmpty(t, params)

	step1and2 := templateFill(testAccVcdRdeStep1and2, params)
	params["FuncName"] = t.Name() + "-Step3"
	step3 := templateFill(testAccVcdRdeStep3, params)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	debugPrintf("#[DEBUG] CONFIGURATION step 1 and 2: %s\n", step1and2)
	debugPrintf("#[DEBUG] CONFIGURATION step 3: %s\n", step3)

	rdeUrnRegexp := fmt.Sprintf(`urn:vcloud:entity:%s:%s:[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$`, params["Vendor"].(string), params["Namespace"].(string))
	rdeType := "vcd_rde_type.rde-type"
	rdeFromFile := "vcd_rde.rde-file"
	rdeFromUrl := "vcd_rde.rde-url"
	rdeWrong := "vcd_rde.rde-naughty"
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckRdeDestroy(rdeType, rdeFromFile, rdeFromUrl),
		Steps: []resource.TestStep{
			{
				Config: step1and2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(rdeFromFile, "id", regexp.MustCompile(rdeUrnRegexp)),
					resource.TestCheckResourceAttr(rdeFromFile, "name", t.Name()+"file"),
					resource.TestCheckResourceAttrPair(rdeFromFile, "rde_type_id", rdeType, "id"),
					resource.TestMatchResourceAttr(rdeFromFile, "entity", regexp.MustCompile("{.*\"stringValue\".*}")),
					resource.TestCheckResourceAttr(rdeFromFile, "state", "RESOLVED"),
					resource.TestMatchResourceAttr(rdeFromFile, "org_id", regexp.MustCompile(`urn:vcloud:org:[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$`)),
					resource.TestMatchResourceAttr(rdeFromFile, "owner_id", regexp.MustCompile(`urn:vcloud:user:[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$`)),

					resource.TestMatchResourceAttr(rdeFromFile, "id", regexp.MustCompile(rdeUrnRegexp)),
					resource.TestCheckResourceAttr(rdeFromUrl, "name", t.Name()+"url"),
					resource.TestCheckResourceAttrPair(rdeFromUrl, "rde_type_id", rdeFromFile, "rde_type_id"),
					resource.TestCheckResourceAttrPair(rdeFromUrl, "entity", rdeFromFile, "entity"),
					resource.TestCheckResourceAttr(rdeFromUrl, "state", "RESOLVED"),
					resource.TestCheckResourceAttrPair(rdeFromUrl, "org_id", rdeFromFile, "org_id"),
					resource.TestCheckResourceAttrPair(rdeFromUrl, "owner_id", rdeFromFile, "owner_id"),

					resource.TestMatchResourceAttr(rdeFromFile, "id", regexp.MustCompile(rdeUrnRegexp)),
					resource.TestCheckResourceAttr(rdeWrong, "name", t.Name()+"naughty"),
					resource.TestCheckResourceAttrPair(rdeWrong, "rde_type_id", rdeFromFile, "rde_type_id"),
					resource.TestCheckResourceAttr(rdeWrong, "entity", "{\"this_json_is_bad\":\"yes\"}"),
					resource.TestCheckResourceAttr(rdeWrong, "state", "RESOLUTION_ERROR"),
					resource.TestCheckResourceAttrPair(rdeFromUrl, "org_id", rdeFromFile, "org_id"),
					resource.TestCheckResourceAttrPair(rdeFromUrl, "owner_id", rdeFromFile, "owner_id"),
				),
			},
			{
				Config: step1and2,
				Taint:  []string{rdeWrong}, // We force a deletion of a wrongly resolved RDE.
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(rdeWrong, "state", "RESOLUTION_ERROR"),
				),
			},
			{
				Config: step3,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(rdeFromFile, "name", t.Name()+"file-updated"),
					resource.TestCheckResourceAttr(rdeFromUrl, "name", t.Name()+"url-updated"),
					resource.TestCheckResourceAttr(rdeWrong, "state", "RESOLVED"),
				),
			},
			{
				ResourceName:      rdeFromFile,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdRde(params["Name"].(string)+"file-updated", params["Vendor"].(string), params["Namespace"].(string), params["Version"].(string)),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdRdePrerequisites = `
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
`

const testAccVcdRdeStep1and2 = testAccVcdRdePrerequisites + `
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

resource "vcd_rde" "rde-naughty" {
  rde_type_id   = vcd_rde_type.rde-type.id
  name          = "{{.Name}}naughty"
  entity        = "{ \"this_json_is_bad\": \"yes\"}"
}
`

const testAccVcdRdeStep3 = testAccVcdRdePrerequisites + `
resource "vcd_rde" "rde-file" {
  rde_type_id   = vcd_rde_type.rde-type.id
  name          = "{{.Name}}file-updated" # Updated name
  entity        = file("{{.EntityPath}}")
}

resource "vcd_rde" "rde-url" {
  rde_type_id   = vcd_rde_type.rde-type.id
  name          = "{{.Name}}url-updated" # Updated name
  entity_url    = "{{.EntityUrl}}"
}

resource "vcd_rde" "rde-naughty" {
  rde_type_id   = vcd_rde_type.rde-type.id
  name          = "{{.Name}}naughty"
  entity        = file("{{.EntityPath}}") # Updated to a correct JSON
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
				if govcd.ContainsNotFound(err) {
					continue
				}
				return fmt.Errorf("error getting the RDE type %s to be able to destroy its instances: %s", rdeTypeRes.Primary.ID, err)
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
