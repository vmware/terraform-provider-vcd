//go:build rde || ALL || functional

package vcd

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"regexp"
	"strings"
	"testing"
)

// TODO: Test resolve = false
// Create non-resolved RDE
// Delete should fail
// Update schema of resolved entity?
// Update to resolve = false when resolved??

// TestAccVcdRde tests the behaviour of RDE instances.
func TestAccVcdRde(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	var params = StringMap{
		"FuncName":       t.Name() + "-Step1-and-2",
		"ProviderSystem": providerVcdSystem,
		"ProviderOrg1":   providerVcdOrg1,
		"Namespace":      "namespace",
		"Version":        "1.0.0",
		"Vendor":         "vendor",
		"Name":           t.Name(),
		"Resolve":        true,
		"SchemaPath":     getCurrentDir() + "/../test-resources/rde_type.json",
		"EntityPath":     getCurrentDir() + "/../test-resources/rde_instance.json",
		"EntityUrl":      "https://raw.githubusercontent.com/adambarreiro/terraform-provider-vcd/add-rde-support-3/test-resources/rde_instance.json", // FIXME
	}
	testParamsNotEmpty(t, params)

	params["FuncName"] = t.Name() + "-Prereqs"
	preReqsConfig := templateFill(testAccVcdRdePrerequisites, params)
	params["FuncName"] = t.Name() + "-Step1And2"
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
	rdeTenant := "vcd_rde.rde_tenant"

	// We will cache an RDE identifier, so we can use it later for importing
	cachedId := &testCachedFieldValue{}

	resource.Test(t, resource.TestCase{
		ProviderFactories: buildMultipleProviders(),
		CheckDestroy:      testAccCheckRdeDestroy(rdeType, rdeFromFile, rdeFromUrl),
		Steps: []resource.TestStep{
			// Step 1: Preconfigure the test: We fetch an existing interface and create an RDE Type.
			// Creating an RDE Type creates an associated Rights Bundle that is not published to all tenants by default.
			// To move forward we need these rights published, so we create a new published bundle with the same rights.
			{
				Config: preReqsConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(rdeType, "vendor", params["Vendor"].(string)),
					resource.TestCheckResourceAttr(rdeType, "namespace", params["Namespace"].(string)),
				),
			},
			// Step 2: Create 4 RDEs:
			// - From a file with Sysadmin using tenant context.
			// - From a URL with Sysadmin using tenant context.
			// - With a wrong JSON schema, it should be created.
			// - From a file with Tenant user as Organization Administrator.
			// For the Tenant user to be able to use RDEs, we need to add rights to the Organization Administrator global
			// role, which is done in `addRightsToTenantUser`.
			{
				Config: step1and2,
				PreConfig: func() {
					addRightsToTenantUser(t, params["Vendor"].(string), params["Namespace"].(string))
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					// We cache the ID to use it on later steps
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

					resource.TestMatchResourceAttr(rdeFromUrl, "id", regexp.MustCompile(rdeUrnRegexp)),
					resource.TestCheckResourceAttr(rdeFromUrl, "name", t.Name()+"url"),
					resource.TestCheckResourceAttrPair(rdeFromUrl, "rde_type_vendor", rdeFromFile, "rde_type_vendor"),
					resource.TestCheckResourceAttrPair(rdeFromUrl, "rde_type_namespace", rdeFromFile, "rde_type_namespace"),
					resource.TestCheckResourceAttrPair(rdeFromUrl, "rde_type_version", rdeFromFile, "rde_type_version"),
					resource.TestCheckResourceAttrPair(rdeFromUrl, "entity", rdeFromFile, "entity"),
					resource.TestCheckResourceAttr(rdeFromUrl, "state", "RESOLVED"),
					resource.TestCheckResourceAttrPair(rdeFromUrl, "org_id", rdeFromFile, "org_id"),
					resource.TestCheckResourceAttrPair(rdeFromUrl, "owner_id", rdeFromFile, "owner_id"),

					resource.TestMatchResourceAttr(rdeWrong, "id", regexp.MustCompile(rdeUrnRegexp)),
					resource.TestCheckResourceAttr(rdeWrong, "name", t.Name()+"naughty"),
					resource.TestCheckResourceAttrPair(rdeWrong, "rde_type_vendor", rdeFromFile, "rde_type_vendor"),
					resource.TestCheckResourceAttrPair(rdeWrong, "rde_type_namespace", rdeFromFile, "rde_type_namespace"),
					resource.TestCheckResourceAttrPair(rdeWrong, "rde_type_version", rdeFromFile, "rde_type_version"),
					resource.TestCheckResourceAttr(rdeWrong, "entity", "{\"this_json_is_bad\":\"yes\"}"),
					resource.TestCheckResourceAttr(rdeWrong, "state", "RESOLUTION_ERROR"),
					resource.TestCheckResourceAttrPair(rdeFromUrl, "org_id", rdeFromFile, "org_id"),
					resource.TestCheckResourceAttrPair(rdeFromUrl, "owner_id", rdeFromFile, "owner_id"),

					resource.TestMatchResourceAttr(rdeTenant, "id", regexp.MustCompile(rdeUrnRegexp)),
					resource.TestCheckResourceAttr(rdeTenant, "name", t.Name()+"tenant"),
					resource.TestCheckResourceAttrPair(rdeTenant, "rde_type_vendor", rdeFromFile, "rde_type_vendor"),
					resource.TestCheckResourceAttrPair(rdeTenant, "rde_type_namespace", rdeFromFile, "rde_type_namespace"),
					resource.TestCheckResourceAttrPair(rdeTenant, "rde_type_version", rdeFromFile, "rde_type_version"),
					resource.TestCheckResourceAttrPair(rdeTenant, "entity", rdeFromFile, "entity"),
					resource.TestCheckResourceAttr(rdeTenant, "state", "RESOLVED"),
					resource.TestCheckResourceAttrPair(rdeTenant, "org_id", rdeFromFile, "org_id"),
					resource.TestMatchResourceAttr(rdeTenant, "owner_id", regexp.MustCompile(`urn:vcloud:user:[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$`)), // Owner is different in this case
				),
			},
			// Step 3: Taint the wrong RDE to test that we can delete a resolved RDE with wrong JSONs
			{
				Config: step1and2,
				Taint:  []string{rdeWrong}, // We force a deletion of a wrongly resolved RDE.
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(rdeWrong, "state", "RESOLUTION_ERROR"),
				),
			},
			// Step 3: Fixes the wrong RDE and resolves it. It also updates the names of the other RDEs.
			{
				Config: step3,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(rdeFromFile, "name", t.Name()+"file-updated"),
					resource.TestCheckResourceAttr(rdeFromUrl, "name", t.Name()+"url-updated"),
					resource.TestCheckResourceAttr(rdeTenant, "name", t.Name()+"tenant-updated"),
					resource.TestCheckResourceAttr(rdeWrong, "state", "RESOLVED"),
				),
			},
			// Step 4: The provider doesn't allow creating more than one RDE with same name.
			{
				Config:      step4,
				ExpectError: regexp.MustCompile(".*found other Runtime Defined Entities with same name.*"),
			},
			// Import by vendor + namespace + version + name + position
			{
				ResourceName:            rdeFromFile,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdFunc:       importStateIdRde(params["Vendor"].(string), params["Namespace"].(string), params["Version"].(string), t.Name()+"file-updated", "1", false),
				ImportStateVerifyIgnore: []string{"resolve"},
			},
			// Import using the cached RDE ID
			{
				ResourceName:      rdeFromFile,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					return cachedId.fieldValue, nil
				},
				ImportStateVerifyIgnore: []string{"resolve"},
			},
			// Import with the list option, it should return the RDE that we cached
			{
				ResourceName:      rdeFromFile,
				ImportState:       true,
				ImportStateIdFunc: importStateIdRde(params["Vendor"].(string), params["Namespace"].(string), params["Version"].(string), t.Name()+"file-updated", "1", true),
				ExpectError:       regexp.MustCompile(`.*` + cachedId.fieldValue + `.*`),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdRdePrerequisites = `
data "vcd_rde_interface" "existing_interface" {
  provider = {{.ProviderSystem}}

  namespace = "k8s"
  version   = "1.0.0"
  vendor    = "vmware"
}

resource "vcd_rde_type" "rde_type" {
  provider = {{.ProviderSystem}}

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
  provider = {{.ProviderSystem}}

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
  provider = {{.ProviderSystem}}

  rde_type_vendor    = vcd_rde_type.rde_type.vendor
  rde_type_namespace = vcd_rde_type.rde_type.namespace
  rde_type_version   = vcd_rde_type.rde_type.version
  name               = "{{.Name}}file"
  resolve            = {{.Resolve}}
  entity             = file("{{.EntityPath}}")

  depends_on = [vcd_rights_bundle.rde_type_bundle]
}

resource "vcd_rde" "rde_url" {
  provider = {{.ProviderSystem}}

  rde_type_vendor    = vcd_rde_type.rde_type.vendor
  rde_type_namespace = vcd_rde_type.rde_type.namespace
  rde_type_version   = vcd_rde_type.rde_type.version
  name               = "{{.Name}}url"
  resolve            = {{.Resolve}}
  entity_url         = "{{.EntityUrl}}"

  depends_on = [vcd_rights_bundle.rde_type_bundle]
}

resource "vcd_rde" "rde_naughty" {
  provider = {{.ProviderSystem}}

  rde_type_vendor    = vcd_rde_type.rde_type.vendor
  rde_type_namespace = vcd_rde_type.rde_type.namespace
  rde_type_version   = vcd_rde_type.rde_type.version
  name               = "{{.Name}}naughty"
  resolve            = {{.Resolve}}
  entity             = "{ \"this_json_is_bad\": \"yes\"}"

  depends_on = [vcd_rights_bundle.rde_type_bundle]
}

resource "vcd_rde" "rde_tenant" {
  provider = {{.ProviderOrg1}}

  rde_type_vendor    = vcd_rde_type.rde_type.vendor
  rde_type_namespace = vcd_rde_type.rde_type.namespace
  rde_type_version   = vcd_rde_type.rde_type.version
  name               = "{{.Name}}tenant"
  resolve            = {{.Resolve}}
  entity             = file("{{.EntityPath}}")

  depends_on = [vcd_rights_bundle.rde_type_bundle]
}
`

const testAccVcdRdeStep3 = testAccVcdRdePrerequisites + `
resource "vcd_rde" "rde_file" {
  provider = {{.ProviderSystem}}

  rde_type_vendor    = vcd_rde_type.rde_type.vendor
  rde_type_namespace = vcd_rde_type.rde_type.namespace
  rde_type_version   = vcd_rde_type.rde_type.version
  name               = "{{.Name}}file-updated" # Updated name
  resolve            = {{.Resolve}}
  entity             = file("{{.EntityPath}}")

  depends_on = [vcd_rights_bundle.rde_type_bundle]
}

resource "vcd_rde" "rde_url" {
  provider = {{.ProviderSystem}}

  rde_type_vendor    = vcd_rde_type.rde_type.vendor
  rde_type_namespace = vcd_rde_type.rde_type.namespace
  rde_type_version   = vcd_rde_type.rde_type.version
  name               = "{{.Name}}url-updated" # Updated name
  resolve            = {{.Resolve}}
  entity_url         = "{{.EntityUrl}}"

  depends_on = [vcd_rights_bundle.rde_type_bundle]
}

resource "vcd_rde" "rde_naughty" {
  provider = {{.ProviderSystem}}

  rde_type_vendor    = vcd_rde_type.rde_type.vendor
  rde_type_namespace = vcd_rde_type.rde_type.namespace
  rde_type_version   = vcd_rde_type.rde_type.version
  name               = "{{.Name}}naughty"
  resolve            = {{.Resolve}}
  entity             = file("{{.EntityPath}}") # Updated to a correct JSON

  depends_on = [vcd_rights_bundle.rde_type_bundle]
}

resource "vcd_rde" "rde_tenant" {
  provider = {{.ProviderOrg1}}

  rde_type_vendor    = vcd_rde_type.rde_type.vendor
  rde_type_namespace = vcd_rde_type.rde_type.namespace
  rde_type_version   = vcd_rde_type.rde_type.version
  name               = "{{.Name}}tenant-updated" # Updated name
  resolve            = {{.Resolve}}
  entity             = file("{{.EntityPath}}")

  depends_on = [vcd_rights_bundle.rde_type_bundle]
}
`

const testAccVcdRdeStep4 = testAccVcdRdeStep3 + `
# skip-binary-test - This should fail
resource "vcd_rde" "rde_naughty-clone" {
  provider = {{.ProviderSystem}}

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

// addRightsToTenantUser adds the RDE type (specified by vendor and namespace) rights to the
// Organization Administrator global role, so a tenant user with this role can perform CRUD operations
// on RDEs.
// NOTE: We don't need to remove the added rights after the test is run, because the RDE Type and the Rights Bundle
// are destroyed and the rights disappear with them gone.
func addRightsToTenantUser(t *testing.T, vendor, namespace string) {
	vcdClient := createTemporaryVCDConnection(true)
	if vcdClient == nil || vcdClient.VCDClient == nil {
		t.Errorf("could not get a VCD connection to add rights to tenant user")
	}
	role, err := vcdClient.VCDClient.Client.GetGlobalRoleByName("Organization Administrator")
	if err != nil {
		t.Errorf("could not get Organization Administrator global role: %s", err)
	}
	rightsBundleName := fmt.Sprintf("%s:%s Entitlement", vendor, namespace)
	rightsBundle, err := vcdClient.VCDClient.Client.GetRightsBundleByName(rightsBundleName)
	if err != nil {
		t.Errorf("could not get %s rights bundle: %s", rightsBundleName, err)
	}
	rights, err := rightsBundle.GetRights(nil)
	if err != nil {
		t.Errorf("could not get rights from %s rights bundle: %s", rightsBundleName, err)
	}
	var rightsToAdd []types.OpenApiReference
	for _, right := range rights {
		if strings.Contains(strings.ToLower(right.Name), fmt.Sprintf("%s:%s", vendor, namespace)) {
			rightsToAdd = append(rightsToAdd, types.OpenApiReference{
				Name: right.Name,
				ID:   right.ID,
			})
		}
	}
	err = role.AddRights(rightsToAdd)
	if err != nil {
		t.Errorf("could not add rights %v to role %s", rightsToAdd, role.GlobalRole.Name)
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
