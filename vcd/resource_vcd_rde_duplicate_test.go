//go:build rde || ALL || functional

package vcd

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// TestAccVcdRdeDuplicate tests the different duplication cases of RDE
func TestAccVcdRdeDuplicate(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	var params = StringMap{
		"Nss":        "nss",
		"Version":    "1.0.0",
		"Vendor":     "vendor",
		"Name":       t.Name(),
		"SchemaPath": getCurrentDir() + "/../test-resources/rde_type.json",
		"EntityPath": getCurrentDir() + "/../test-resources/rde_instance.json",
	}
	testParamsNotEmpty(t, params)

	params["FuncName"] = t.Name() + "-Step1"
	step1 := templateFill(testAccVcdRdeDuplicateStep1, params)
	params["FuncName"] = t.Name() + "-Step2"
	step2 := templateFill(testAccVcdRdeDuplicateStep2, params)
	params["FuncName"] = t.Name() + "-Step3"
	step3 := templateFill(testAccVcdRdeDuplicateStep3, params)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	debugPrintf("#[DEBUG] CONFIGURATION step1: %s\n", step1)

	rde1System := "vcd_rde.rde1_system"
	rde2Tenant := "vcd_rde.rde2_tenant"
	rde3Tenant := "vcd_rde.rde3_tenant"
	fetchedSystem := "data.vcd_rde.fetch_rde_system"

	// We will cache some RDE identifiers, so we can use them later
	cachedIds := make([]testCachedFieldValue, 2)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckRdeDestroy(rde1System, rde2Tenant, rde3Tenant),
		Steps: []resource.TestStep{
			// Create three RDEs that are exactly the same, but one is created in System org and
			// the others in the default tenant.
			{
				Config: step1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resourceFieldsEqual(rde2Tenant, rde3Tenant, []string{"id"}),
					resourceFieldsEqual(rde1System, rde3Tenant, []string{"id", "org", "org_id"}),
					// We cache some IDs to use it on later steps
					cachedIds[0].cacheTestResourceFieldValue(rde2Tenant, "id"),
					cachedIds[1].cacheTestResourceFieldValue(rde3Tenant, "id"),
				),
			},
			// We use a data source to fetch the RDE that belongs to System.
			// Despite it is duplicated (it has same name, type, etc), it should work as we can unequivocally fetch it
			// thanks that is the only one present in the System organization.
			{
				Config: step2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(fetchedSystem, "id", rde1System, "id"),
					resource.TestCheckResourceAttrPair(fetchedSystem, "rde_type_id", rde1System, "rde_type_id"),
					resource.TestCheckResourceAttrPair(fetchedSystem, "external_id", rde1System, "external_id"),
					resource.TestCheckResourceAttrPair(fetchedSystem, "entity", rde1System, "computed_entity"),
					resource.TestCheckResourceAttrPair(fetchedSystem, "org_id", rde1System, "org_id"),
					resource.TestCheckResourceAttrPair(fetchedSystem, "owner_user_id", rde1System, "owner_user_id"),
				),
			},
			// We use a data source to fetch the RDE that belongs to the configured tenant.
			// This one is duplicated (it has same name, type, etc) as we have another on the same tenant, so fetching
			// it should fail and return the IDs of both RDEs to help the user.
			{
				Config:      step3,
				ExpectError: regexp.MustCompile(`there are 2 RDEs.*` + cachedIds[0].fieldValue + ` ` + cachedIds[1].fieldValue),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdRdeDuplicateStep1 = `
data "vcd_rde_interface" "existing_interface" {
  nss     = "k8s"
  version = "1.0.0"
  vendor  = "vmware"
}

resource "vcd_rde_type" "rde_type" {
  nss     = "{{.Nss}}"
  version = "{{.Version}}"
  vendor  = "{{.Vendor}}"
  name    = "{{.Name}}_type"
  schema  = file("{{.SchemaPath}}")
}

# Creating a RDE Type creates a bundle and some rights in the background, but the
# bundle is unpublished by default.
# We create another bundle so we can publish it to all tenants and they can
# use RDEs.
resource "vcd_rights_bundle" "rde_type_bundle" {
  name                   = "{{.Name}} bundle"
  description            = "{{.Name}} bundle"
  publish_to_all_tenants = true
  rights = [
    "{{.Vendor}}:{{.Nss}}: Administrator Full access",
    "{{.Vendor}}:{{.Nss}}: Full Access",
    "{{.Vendor}}:{{.Nss}}: Modify",
    "{{.Vendor}}:{{.Nss}}: View",
    "{{.Vendor}}:{{.Nss}}: Administrator View",
  ]
  depends_on = [vcd_rde_type.rde_type]
}

# This one does NOT depend on the rights bundle as it will be
# created in System org
resource "vcd_rde" "rde1_system" {
  org          = "System"
  rde_type_id  = vcd_rde_type.rde_type.id
  name         = "{{.Name}}"
  resolve      = true
  input_entity = file("{{.EntityPath}}")
}

# This one depends on the rights bundle as it will be
# created in the tenant that is configured in the provider block
resource "vcd_rde" "rde2_tenant" {
  rde_type_id  = vcd_rde_type.rde_type.id
  name         = "{{.Name}}"
  resolve      = true
  input_entity = file("{{.EntityPath}}")

  depends_on = [vcd_rights_bundle.rde_type_bundle]
}

resource "vcd_rde" "rde3_tenant" {
  rde_type_id  = vcd_rde.rde2_tenant.rde_type_id
  name         = vcd_rde.rde2_tenant.name
  resolve      = vcd_rde.rde2_tenant.resolve
  input_entity = vcd_rde.rde2_tenant.input_entity
}
`

const testAccVcdRdeDuplicateStep2 = testAccVcdRdeDuplicateStep1 + `
# skip-binary-test: Using a data source that references a resource created in same config
data "vcd_rde" "fetch_rde_system" {
  org         = "System"
  rde_type_id = vcd_rde.rde1_system.rde_type_id
  name        = vcd_rde.rde1_system.name
}
`

const testAccVcdRdeDuplicateStep3 = testAccVcdRdeDuplicateStep2 + `
# skip-binary-test: Using a data source that references a resource created in same config
data "vcd_rde" "fetch_rde_tenant" {
  rde_type_id = vcd_rde.rde2_tenant.rde_type_id
  name        = vcd_rde.rde2_tenant.name
}
`
