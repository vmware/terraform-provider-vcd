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

// TODO: Test resolve = false
// Create non-resolved RDE
// Delete should fail
// Update schema of resolved entity?
// Update to resolve = false when resolved??

// TestAccVcdRde tests the behaviour of RDE instances:
// - Step 1: Create 3 RDEs: One with file, other with URL, last one with wrong JSON.
// - Step 2: Taint to test delete on wrong RDEs and repeat step 1.
// - Step 3: Update one RDE name. Update wrong JSON in RDE.
// - Step 4: Attempt to create a clone of an RDE with same name and type ID. It should fail.
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
		"Resolve":    true,
		"SchemaPath": getCurrentDir() + "/../test-resources/rde_type.json",
		"EntityPath": getCurrentDir() + "/../test-resources/rde_instance.json",
		"EntityUrl":  "https://raw.githubusercontent.com/adambarreiro/terraform-provider-vcd/add-rde-support-3/test-resources/rde_instance.json", // FIXME
	}
	testParamsNotEmpty(t, params)

	step1and2 := templateFill(testAccVcdRdeStep1and2, params)
	params["FuncName"] = t.Name() + "-Step3"
	step3 := templateFill(testAccVcdRdeStep3, params)
	params["FuncName"] = t.Name() + "-Step4"
	step4 := templateFill(testAccVcdRdeStep4, params)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	debugPrintf("#[DEBUG] CONFIGURATION step 1 and 2: %s\n", step1and2)
	debugPrintf("#[DEBUG] CONFIGURATION step 3: %s\n", step3)
	debugPrintf("#[DEBUG] CONFIGURATION step 4: %s\n", step4)

	rdeUrnRegexp := fmt.Sprintf(`urn:vcloud:entity:%s:%s:[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$`, params["Vendor"].(string), params["Namespace"].(string))
	rdeType := "vcd_rde_type.rde_type"
	rdeFromFile := "vcd_rde.rde_file"
	rdeFromUrl := "vcd_rde.rde_url"
	rdeWrong := "vcd_rde.rde_naughty"

	// We will cache an RDE identifier, so we can use it later for importing
	cachedId := &testCachedFieldValue{}

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckRdeDestroy(rdeType, rdeFromFile, rdeFromUrl),
		Steps: []resource.TestStep{
			{
				Config: step1and2,
				Check: resource.ComposeAggregateTestCheckFunc(
					cachedId.cacheTestResourceFieldValue(rdeFromFile, "id"),
					resource.TestMatchResourceAttr(rdeFromFile, "id", regexp.MustCompile(rdeUrnRegexp)),
					resource.TestCheckResourceAttr(rdeFromFile, "name", t.Name()+"file"),
					resource.TestCheckResourceAttrPair(rdeFromFile, "rde_type_vendor", rdeType, "vendor"),
					resource.TestCheckResourceAttrPair(rdeFromFile, "rde_type_namespace", rdeType, "namespace"),
					resource.TestCheckResourceAttrPair(rdeFromFile, "rde_type_version", rdeType, "version"),
					resource.TestMatchResourceAttr(rdeFromFile, "entity", regexp.MustCompile("{.*\"stringValue\".*}")),
					resource.TestCheckResourceAttr(rdeFromFile, "state", "RESOLVED"),
					resource.TestMatchResourceAttr(rdeFromFile, "org_id", regexp.MustCompile(`urn:vcloud:org:[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$`)),
					resource.TestMatchResourceAttr(rdeFromFile, "owner_id", regexp.MustCompile(`urn:vcloud:user:[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$`)),

					resource.TestMatchResourceAttr(rdeFromFile, "id", regexp.MustCompile(rdeUrnRegexp)),
					resource.TestCheckResourceAttr(rdeFromUrl, "name", t.Name()+"url"),
					resource.TestCheckResourceAttrPair(rdeFromUrl, "rde_type_vendor", rdeFromFile, "rde_type_vendor"),
					resource.TestCheckResourceAttrPair(rdeFromUrl, "rde_type_namespace", rdeFromFile, "rde_type_namespace"),
					resource.TestCheckResourceAttrPair(rdeFromUrl, "rde_type_version", rdeFromFile, "rde_type_version"),
					resource.TestCheckResourceAttrPair(rdeFromUrl, "entity", rdeFromFile, "entity"),
					resource.TestCheckResourceAttr(rdeFromUrl, "state", "RESOLVED"),
					resource.TestCheckResourceAttrPair(rdeFromUrl, "org_id", rdeFromFile, "org_id"),
					resource.TestCheckResourceAttrPair(rdeFromUrl, "owner_id", rdeFromFile, "owner_id"),

					resource.TestMatchResourceAttr(rdeFromFile, "id", regexp.MustCompile(rdeUrnRegexp)),
					resource.TestCheckResourceAttr(rdeWrong, "name", t.Name()+"naughty"),
					resource.TestCheckResourceAttrPair(rdeWrong, "rde_type_vendor", rdeFromFile, "rde_type_vendor"),
					resource.TestCheckResourceAttrPair(rdeWrong, "rde_type_namespace", rdeFromFile, "rde_type_namespace"),
					resource.TestCheckResourceAttrPair(rdeWrong, "rde_type_version", rdeFromFile, "rde_type_version"),
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
				Config:      step4,
				ExpectError: regexp.MustCompile(".*found other Runtime Defined Entities with same name.*"),
			},
			{
				// Import by vendor + namespace + version + name + position
				ResourceName:            rdeFromFile,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdFunc:       importStateIdRde(params["Vendor"].(string), params["Namespace"].(string), params["Version"].(string), t.Name()+"file-updated", "1", false),
				ImportStateVerifyIgnore: []string{"resolve"},
			},
			{
				// Import by ID
				ResourceName:      rdeFromFile,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					return cachedId.fieldValue, nil
				},
				ImportStateVerifyIgnore: []string{"resolve"},
			},
			{
				// Test list
				ResourceName:      rdeFromFile,
				ImportState:       true,
				ImportStateIdFunc: importStateIdRde(params["Vendor"].(string), params["Namespace"].(string), params["Version"].(string), t.Name()+"file-updated", "1", true),
				ExpectError:       regexp.MustCompile(".*"),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdRdePrerequisites = `
data "vcd_rde_interface" "existing_interface" {
  namespace = "k8s"
  version   = "1.0.0"
  vendor    = "vmware"
}

resource "vcd_rde_type" "rde_type" {
  namespace     = "{{.Namespace}}"
  version       = "{{.Version}}"
  vendor        = "{{.Vendor}}"
  name          = "{{.Name}}_type"
  schema        = file("{{.SchemaPath}}")
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
    "{{.Vendor}}:{{.Namespace}}: Administrator Full access",
    "{{.Vendor}}:{{.Namespace}}: Full Access",
    "{{.Vendor}}:{{.Namespace}}: Modify",
    "{{.Vendor}}:{{.Namespace}}: View",
    "{{.Vendor}}:{{.Namespace}}: Administrator View",
  ]
  depends_on = [vcd_rde_type.rde_type]
}
`

const testAccVcdRdeStep1and2 = testAccVcdRdePrerequisites + `
resource "vcd_rde" "rde_file" {
  rde_type_vendor    = vcd_rde_type.rde_type.vendor
  rde_type_namespace = vcd_rde_type.rde_type.namespace
  rde_type_version   = vcd_rde_type.rde_type.version
  name               = "{{.Name}}file"
  resolve            = {{.Resolve}}
  entity             = file("{{.EntityPath}}")

  depends_on = [vcd_rights_bundle.rde_type_bundle]
}

resource "vcd_rde" "rde_url" {
  rde_type_vendor    = vcd_rde_type.rde_type.vendor
  rde_type_namespace = vcd_rde_type.rde_type.namespace
  rde_type_version   = vcd_rde_type.rde_type.version
  name               = "{{.Name}}url"
  resolve            = {{.Resolve}}
  entity_url         = "{{.EntityUrl}}"

  depends_on = [vcd_rights_bundle.rde_type_bundle]
}

resource "vcd_rde" "rde_naughty" {
  rde_type_vendor    = vcd_rde_type.rde_type.vendor
  rde_type_namespace = vcd_rde_type.rde_type.namespace
  rde_type_version   = vcd_rde_type.rde_type.version
  name               = "{{.Name}}naughty"
  resolve            = {{.Resolve}}
  entity             = "{ \"this_json_is_bad\": \"yes\"}"

  depends_on = [vcd_rights_bundle.rde_type_bundle]
}
`

const testAccVcdRdeStep3 = testAccVcdRdePrerequisites + `
resource "vcd_rde" "rde_file" {
  rde_type_vendor    = vcd_rde_type.rde_type.vendor
  rde_type_namespace = vcd_rde_type.rde_type.namespace
  rde_type_version   = vcd_rde_type.rde_type.version
  name               = "{{.Name}}file-updated" # Updated name
  resolve            = {{.Resolve}}
  entity             = file("{{.EntityPath}}")

  depends_on = [vcd_rights_bundle.rde_type_bundle]
}

resource "vcd_rde" "rde_url" {
  rde_type_vendor    = vcd_rde_type.rde_type.vendor
  rde_type_namespace = vcd_rde_type.rde_type.namespace
  rde_type_version   = vcd_rde_type.rde_type.version
  name               = "{{.Name}}url-updated" # Updated name
  resolve            = {{.Resolve}}
  entity_url         = "{{.EntityUrl}}"

  depends_on = [vcd_rights_bundle.rde_type_bundle]
}

resource "vcd_rde" "rde_naughty" {
  rde_type_vendor    = vcd_rde_type.rde_type.vendor
  rde_type_namespace = vcd_rde_type.rde_type.namespace
  rde_type_version   = vcd_rde_type.rde_type.version
  name               = "{{.Name}}naughty"
  resolve            = {{.Resolve}}
  entity             = file("{{.EntityPath}}") # Updated to a correct JSON

  depends_on = [vcd_rights_bundle.rde_type_bundle]
}
`

const testAccVcdRdeStep4 = testAccVcdRdeStep3 + `
# skip-binary-test - This should fail
resource "vcd_rde" "rde_naughty-clone" {
  rde_type_vendor    = vcd_rde_type.rde_type.vendor
  rde_type_namespace = vcd_rde_type.rde_type.namespace
  rde_type_version   = vcd_rde_type.rde_type.version
  name               = "{{.Name}}naughty"
  resolve            = {{.Resolve}}
  entity             = file("{{.EntityPath}}")

  depends_on = [vcd_rights_bundle.rde_type_bundle]
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

func importStateIdRde(vendor, namespace, version, name, position string, list bool) resource.ImportStateIdFunc {
	return func(*terraform.State) (string, error) {
		commonIdPart := vendor +
			ImportSeparator +
			namespace +
			ImportSeparator +
			version +
			ImportSeparator +
			name
		if list {
			return "list@" + commonIdPart, nil
		}
		return commonIdPart + ImportSeparator + position, nil
	}
}

// TestAccVcdRdeMetadata tests metadata CRUD on Runtime Defined Entities.
func TestAccVcdRdeMetadata(t *testing.T) {
	skipIfNotSysAdmin(t)
	vcdClient := createTemporaryVCDConnection(true)
	if vcdClient != nil && vcdClient.Client.APIVCDMaxVersionIs("< 37.0") {
		t.Skip("skipped as metadata for vcd_rde is only supported since VCD 10.4.0")
	}
	testOpenApiMetadataEntryCRUD(t,
		testAccCheckVcdRdeMetadata, "vcd_rde.test-rde",
		testAccCheckVcdRdeMetadataDatasource, "data.vcd_rde.test-rde-ds",
		StringMap{})
}

const testAccCheckVcdRdeMetadata = `
data "vcd_rde_type" "rde_type" {
  vendor    = "vmware"
  namespace = "tkgcluster"
  version   = "1.0.0"
}

resource "vcd_rde" "test-rde" {
  rde_type_id = data.vcd_rde_type.rde_type.id
  name        = "{{.Name}}"
  entity      = "{\"foo\":\"bar\"}" # We are just testing metadata so we don't care about entity state

  {{.Metadata}}
}
`

const testAccCheckVcdRdeMetadataDatasource = `
data "vcd_rde" "test-rde-ds" {
  name        = vcd_rde.test-rde.name
  rde_type_id = data.vcd_rde_type.rde_type.id
}
`
