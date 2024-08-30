//go:build rde || ALL || functional

package vcd

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// TestAccVcdRdeUpgrade tests the upgrade process of a RDE
func TestAccVcdRdeUpgrade(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	var params = StringMap{
		"Nss":        "nss",
		"Vendor":     "vendor",
		"Name":       t.Name(),
		"SchemaPath": getCurrentDir() + "/../test-resources/rde_type.json",
		"EntityPath": getCurrentDir() + "/../test-resources/rde_instance.json",
		"RdeTypeId":  "vcd_rde_type.rde_type1.id",
		"Tags":       "rde",
	}
	testParamsNotEmpty(t, params)

	params["FuncName"] = t.Name() + "-Step1"
	step1 := templateFill(testAccVcdRdeUpgrade, params)
	debugPrintf("#[DEBUG] CONFIGURATION step1: %s\n", step1)

	params["FuncName"] = t.Name() + "-Step2"
	params["RdeTypeId"] = "vcd_rde_type.rde_type2.id"
	step2 := templateFill(testAccVcdRdeUpgrade, params)
	debugPrintf("#[DEBUG] CONFIGURATION step2: %s\n", step2)

	params["FuncName"] = t.Name() + "-Step3"
	params["RdeTypeId"] = "vcd_rde_type.rde_type3.id"
	step3 := templateFill(testAccVcdRdeUpgrade, params)
	debugPrintf("#[DEBUG] CONFIGURATION step3: %s\n", step3)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckRdeTypesDestroy("vcd_rde_type.rde_type1", "vcd_rde_type.rde_type2", "vcd_rde_type.rde_type3"),
		Steps: []resource.TestStep{
			{
				Config: step1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("vcd_rde_type.rde_type1", "id", "vcd_rde.rde", "rde_type_id"),
				),
			},
			// We update to a new type that differs only in the Version field
			{
				Config: step2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair("vcd_rde_type.rde_type2", "id", "vcd_rde.rde", "rde_type_id"),
				),
			},
			// We update to a different type with a different "nss", this will fail
			{
				Config:      step3,
				ExpectError: regexp.MustCompile("RDE_INVALID_TYPE_UPDATE"),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdRdeUpgrade = `
resource "vcd_rde_type" "rde_type1" {
  nss     = "{{.Nss}}"
  version = "1.0.0"
  vendor  = "{{.Vendor}}"
  name    = "{{.Name}}"
  schema  = file("{{.SchemaPath}}")
}

resource "vcd_rde_type" "rde_type2" {
  nss     = "{{.Nss}}"
  version = "1.1.0"
  vendor  = "{{.Vendor}}"
  name    = "{{.Name}}"
  schema  = file("{{.SchemaPath}}")
}

resource "vcd_rde_type" "rde_type3" {
  nss     = "DifferentType"
  version = "1.2.0"
  vendor  = "{{.Vendor}}"
  name    = "{{.Name}}"
  schema  = file("{{.SchemaPath}}")
}

# We create all RDEs in System to avoid using Rights Bundles
resource "vcd_rde" "rde" {
  org          = "System"
  rde_type_id  = {{.RdeTypeId}}
  name         = "{{.Name}}"
  resolve      = true
  input_entity = file("{{.EntityPath}}")

  # This is a workaround for the post-test destroy to work. The step 3 tests that
  # this RDE cannot be updated to rde_type3, so it fails. As it fails, the RDE still uses
  # RDE Type 2 in VCD, but it lost its reference in this config, so we need to force it with a depends_on, so this RDE
  # gets deleted before RDE Type 2.
  depends_on = [ vcd_rde_type.rde_type2]
}
`
