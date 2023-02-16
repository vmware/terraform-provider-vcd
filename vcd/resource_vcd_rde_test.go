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

// TestAccVcdRde tests the behaviour of RDE instances.
func TestAccVcdRde(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	var params = StringMap{
		"FuncName":       t.Name() + "-Step1-and-2",
		"ProviderSystem": providerVcdSystem,
		"ProviderOrg1":   providerVcdOrg1,
		"Nss":            "nss",
		"Version":        "1.0.0",
		"Vendor":         "vendor",
		"Name":           t.Name(),
		"Resolve":        false,
		"SchemaPath":     getCurrentDir() + "/../test-resources/rde_type.json",
		"EntityPath":     getCurrentDir() + "/../test-resources/rde_instance.json",
		"EntityUrl":      "https://raw.githubusercontent.com/adambarreiro/terraform-provider-vcd/add-rde-support-3/test-resources/rde_instance.json", // FIXME
	}
	testParamsNotEmpty(t, params)

	params["FuncName"] = t.Name() + "-Prereqs"
	preReqsConfig := templateFill(testAccVcdRdePrerequisites, params)
	params["FuncName"] = t.Name() + "-Init"
	stepInit := templateFill(testAccVcdRde1, params)
	params["FuncName"] = t.Name() + "-DeleteFail"
	stepDeleteFail := templateFill(testAccVcdRde2, params)
	params["FuncName"] = t.Name() + "-Resolve"
	params["Resolve"] = true
	stepResolve := templateFill(testAccVcdRde1, params)
	params["FuncName"] = t.Name() + "-FixWrongRde"
	stepFixWrongRde := templateFill(testAccVcdRde3, params)
	params["FuncName"] = t.Name() + "-Duplicate"
	stepCreateDuplicate := templateFill(testAccVcdRde4, params)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	debugPrintf("#[DEBUG] CONFIGURATION preReqs: %s\n", preReqsConfig)
	debugPrintf("#[DEBUG] CONFIGURATION init: %s\n", stepInit)
	debugPrintf("#[DEBUG] CONFIGURATION resolve: %s\n", stepResolve)
	debugPrintf("#[DEBUG] CONFIGURATION fix wrong RDE: %s\n", stepFixWrongRde)
	debugPrintf("#[DEBUG] CONFIGURATION duplicate: %s\n", stepCreateDuplicate)

	rdeUrnRegexp := fmt.Sprintf(`urn:vcloud:entity:%s:%s:[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$`, params["Vendor"].(string), params["Nss"].(string))
	rdeType := "vcd_rde_type.rde_type"
	rdeFromFile := "vcd_rde.rde_file"
	rdeFromUrl := "vcd_rde.rde_url"
	rdeWrong := "vcd_rde.rde_naughty"
	rdeTenant := "vcd_rde.rde_tenant"

	// We will cache an RDE identifier, so we can use it later for importing
	cachedId := &testCachedFieldValue{}

	vcdClient := createTemporaryVCDConnection(true)
	if vcdClient == nil || vcdClient.VCDClient == nil {
		t.Errorf("could not get a VCD connection to add rights to tenant user")
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: buildMultipleProviders(),
		CheckDestroy:      testAccCheckRdeDestroy(rdeType, rdeFromFile, rdeFromUrl),
		Steps: []resource.TestStep{
			// Preconfigure the test: We fetch an existing interface and create an RDE Type.
			// Creating an RDE Type creates an associated Rights Bundle that is NOT published to all tenants by default.
			// To move forward we need these rights published, so we create a new published bundle with the same rights.
			{
				Config: preReqsConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(rdeType, "vendor", params["Vendor"].(string)),
					resource.TestCheckResourceAttr(rdeType, "nss", params["Nss"].(string)),
				),
			},
			// Create 4 RDEs in non-resolved state (pre-created):
			// - From a file with Sysadmin using tenant context.
			// - From a URL with Sysadmin using tenant context.
			// - With a wrong JSON schema, it should be created.
			// - From a file with Tenant user as Organization Administrator.
			// For the Tenant user to be able to use RDEs, we need to add the RDE Type rights to the Organization Administrator global
			// role, which is done in `addRightsToTenantUser`.
			{
				Config: stepInit,
				PreConfig: func() {
					addRightsToTenantUser(t, vcdClient, params["Vendor"].(string), params["Nss"].(string))
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					// We cache the ID to use it on later steps
					cachedId.cacheTestResourceFieldValue(rdeFromFile, "id"),

					resource.TestMatchResourceAttr(rdeFromFile, "id", regexp.MustCompile(rdeUrnRegexp)),
					resource.TestCheckResourceAttr(rdeFromFile, "name", t.Name()+"file"),
					resource.TestCheckResourceAttrPair(rdeFromFile, "rde_type_vendor", rdeType, "vendor"),
					resource.TestCheckResourceAttrPair(rdeFromFile, "rde_type_nss", rdeType, "nss"),
					resource.TestCheckResourceAttrPair(rdeFromFile, "rde_type_version", rdeType, "version"),
					resource.TestMatchResourceAttr(rdeFromFile, "computed_entity", regexp.MustCompile("{.*\"stringValue\".*}")),
					resource.TestCheckResourceAttr(rdeFromFile, "state", "PRE_CREATED"),
					resource.TestMatchResourceAttr(rdeFromFile, "org_id", regexp.MustCompile(`urn:vcloud:org:[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$`)),
					resource.TestMatchResourceAttr(rdeFromFile, "owner_id", regexp.MustCompile(`urn:vcloud:user:[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$`)),

					resource.TestMatchResourceAttr(rdeFromUrl, "id", regexp.MustCompile(rdeUrnRegexp)),
					resource.TestCheckResourceAttr(rdeFromUrl, "name", t.Name()+"url"),
					resource.TestCheckResourceAttrPair(rdeFromUrl, "rde_type_vendor", rdeFromFile, "rde_type_vendor"),
					resource.TestCheckResourceAttrPair(rdeFromUrl, "rde_type_nss", rdeFromFile, "rde_type_nss"),
					resource.TestCheckResourceAttrPair(rdeFromUrl, "rde_type_version", rdeFromFile, "rde_type_version"),
					resource.TestCheckResourceAttrPair(rdeFromUrl, "computed_entity", rdeFromFile, "computed_entity"),
					resource.TestCheckResourceAttr(rdeFromUrl, "state", "PRE_CREATED"),
					resource.TestCheckResourceAttrPair(rdeFromUrl, "org_id", rdeFromFile, "org_id"),
					resource.TestCheckResourceAttrPair(rdeFromUrl, "owner_id", rdeFromFile, "owner_id"),

					resource.TestMatchResourceAttr(rdeWrong, "id", regexp.MustCompile(rdeUrnRegexp)),
					resource.TestCheckResourceAttr(rdeWrong, "name", t.Name()+"naughty"),
					resource.TestCheckResourceAttrPair(rdeWrong, "rde_type_vendor", rdeFromFile, "rde_type_vendor"),
					resource.TestCheckResourceAttrPair(rdeWrong, "rde_type_nss", rdeFromFile, "rde_type_nss"),
					resource.TestCheckResourceAttrPair(rdeWrong, "rde_type_version", rdeFromFile, "rde_type_version"),
					resource.TestCheckResourceAttr(rdeWrong, "computed_entity", "{\"this_json_is_bad\":\"yes\"}"),
					resource.TestCheckResourceAttr(rdeWrong, "state", "PRE_CREATED"),
					resource.TestCheckResourceAttrPair(rdeFromUrl, "org_id", rdeFromFile, "org_id"),
					resource.TestCheckResourceAttrPair(rdeFromUrl, "owner_id", rdeFromFile, "owner_id"),

					resource.TestMatchResourceAttr(rdeTenant, "id", regexp.MustCompile(rdeUrnRegexp)),
					resource.TestCheckResourceAttr(rdeTenant, "name", t.Name()+"tenant"),
					resource.TestCheckResourceAttrPair(rdeTenant, "rde_type_vendor", rdeFromFile, "rde_type_vendor"),
					resource.TestCheckResourceAttrPair(rdeTenant, "rde_type_nss", rdeFromFile, "rde_type_nss"),
					resource.TestCheckResourceAttrPair(rdeTenant, "rde_type_version", rdeFromFile, "rde_type_version"),
					resource.TestCheckResourceAttrPair(rdeTenant, "computed_entity", rdeFromFile, "computed_entity"),
					resource.TestCheckResourceAttr(rdeTenant, "state", "PRE_CREATED"),
					resource.TestCheckResourceAttrPair(rdeTenant, "org_id", rdeFromFile, "org_id"),
					resource.TestMatchResourceAttr(rdeTenant, "owner_id", regexp.MustCompile(`urn:vcloud:user:[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$`)), // Owner is different in this case
				),
			},
			// Delete the RDE that has wrong JSON. It should fail because the RDE is not resolved.
			// NOTE: We cannot use SDK Taint here because it will be permanently tainted, hence we can't update it
			// to a correct state in next steps (we can't un-taint)
			{
				Config:      stepDeleteFail,
				ExpectError: regexp.MustCompile(".*could not delete the Runtime Defined Entity.*"),
			},
			// Updates all RDEs to resolve them
			{
				Config: stepResolve,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(rdeFromFile, "state", "RESOLVED"),
					resource.TestCheckResourceAttr(rdeFromUrl, "state", "RESOLVED"),
					resource.TestCheckResourceAttr(rdeTenant, "state", "RESOLVED"),
					resource.TestCheckResourceAttr(rdeWrong, "state", "RESOLUTION_ERROR"),
				),
			},
			// Taint the RDE that has wrong JSON. It should be removed now as it was resolved (with errors)
			{
				Config: stepResolve,
				Taint:  []string{rdeWrong}, // This time is resolved, but still wrong
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(rdeWrong, "state", "RESOLUTION_ERROR"),
				),
			},
			// Fixes the wrong RDE and resolves it. It also updates the names of the other RDEs.
			{
				Config: stepFixWrongRde,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(rdeFromFile, "name", t.Name()+"file-updated"),
					resource.TestCheckResourceAttr(rdeFromUrl, "name", t.Name()+"url-updated"),
					resource.TestCheckResourceAttr(rdeTenant, "name", t.Name()+"tenant-updated"),
					resource.TestCheckResourceAttr(rdeWrong, "state", "RESOLVED"),
				),
			},
			// We test the use case where a 3rd party member changes the RDE json in VCD. The computed entity should change,
			// while the input entity should remain the same.
			{
				Config: stepFixWrongRde,
				PreConfig: func() {
					manipulateRde(t, vcdClient, cachedId.fieldValue)
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(rdeFromFile, "input_entity", rdeWrong, "input_entity"),
					resource.TestMatchResourceAttr(rdeFromFile, "computed_entity", regexp.MustCompile(`.*stringValueChanged.*`)),
				),
			},
			// The provider doesn't allow creating more than one RDE with same name.
			{
				Config:      stepCreateDuplicate,
				ExpectError: regexp.MustCompile(".*found other Runtime Defined Entities with same name.*"),
			},
			// Import by vendor + nss + version + name + position
			{
				ResourceName:            rdeFromFile,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdFunc:       importStateIdRde(params["Vendor"].(string), params["Nss"].(string), params["Version"].(string), t.Name()+"file-updated", "1", false),
				ImportStateVerifyIgnore: []string{"resolve", "input_entity", "input_entity_url"},
			},
			// Import using the cached RDE ID
			{
				ResourceName:      rdeFromFile,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: func(state *terraform.State) (string, error) {
					return cachedId.fieldValue, nil
				},
				ImportStateVerifyIgnore: []string{"resolve", "input_entity", "input_entity_url"},
			},
			// Import with the list option, it should return the RDE that we cached
			{
				ResourceName:      rdeFromFile,
				ImportState:       true,
				ImportStateIdFunc: importStateIdRde(params["Vendor"].(string), params["Nss"].(string), params["Version"].(string), t.Name()+"file-updated", "1", true),
				ExpectError:       regexp.MustCompile(`.*` + cachedId.fieldValue + `.*`),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdRdePrerequisites = `
data "vcd_rde_interface" "existing_interface" {
  provider = {{.ProviderSystem}}

  nss     = "k8s"
  version = "1.0.0"
  vendor  = "vmware"
}

resource "vcd_rde_type" "rde_type" {
  provider = {{.ProviderSystem}}

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
  provider = {{.ProviderSystem}}

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
`

const testAccVcdRde1 = testAccVcdRdePrerequisites + `
resource "vcd_rde" "rde_file" {
  provider = {{.ProviderSystem}}

  rde_type_vendor  = vcd_rde_type.rde_type.vendor
  rde_type_nss     = vcd_rde_type.rde_type.nss
  rde_type_version = vcd_rde_type.rde_type.version
  name             = "{{.Name}}file"
  resolve          = {{.Resolve}}
  input_entity     = file("{{.EntityPath}}")

  depends_on = [vcd_rights_bundle.rde_type_bundle]
}

resource "vcd_rde" "rde_url" {
  provider = {{.ProviderSystem}}

  rde_type_vendor  = vcd_rde_type.rde_type.vendor
  rde_type_nss     = vcd_rde_type.rde_type.nss
  rde_type_version = vcd_rde_type.rde_type.version
  name             = "{{.Name}}url"
  resolve          = {{.Resolve}}
  input_entity_url = "{{.EntityUrl}}"

  depends_on = [vcd_rights_bundle.rde_type_bundle]
}

resource "vcd_rde" "rde_naughty" {
  provider = {{.ProviderSystem}}

  rde_type_vendor  = vcd_rde_type.rde_type.vendor
  rde_type_nss     = vcd_rde_type.rde_type.nss
  rde_type_version = vcd_rde_type.rde_type.version
  name             = "{{.Name}}naughty"
  resolve          = {{.Resolve}}
  input_entity     = "{ \"this_json_is_bad\": \"yes\"}"

  depends_on = [vcd_rights_bundle.rde_type_bundle]
}

resource "vcd_rde" "rde_tenant" {
  provider = {{.ProviderOrg1}}

  rde_type_vendor  = vcd_rde_type.rde_type.vendor
  rde_type_nss     = vcd_rde_type.rde_type.nss
  rde_type_version = vcd_rde_type.rde_type.version
  name             = "{{.Name}}tenant"
  resolve          = {{.Resolve}}
  input_entity     = file("{{.EntityPath}}")

  depends_on = [vcd_rights_bundle.rde_type_bundle]
}
`

const testAccVcdRde2 = testAccVcdRdePrerequisites + `
resource "vcd_rde" "rde_file" {
  provider = {{.ProviderSystem}}

  rde_type_vendor  = vcd_rde_type.rde_type.vendor
  rde_type_nss     = vcd_rde_type.rde_type.nss
  rde_type_version = vcd_rde_type.rde_type.version
  name             = "{{.Name}}file-updated" # Updated name
  resolve          = {{.Resolve}}
  input_entity     = file("{{.EntityPath}}")

  depends_on = [vcd_rights_bundle.rde_type_bundle]
}

resource "vcd_rde" "rde_url" {
  provider = {{.ProviderSystem}}

  rde_type_vendor  = vcd_rde_type.rde_type.vendor
  rde_type_nss     = vcd_rde_type.rde_type.nss
  rde_type_version = vcd_rde_type.rde_type.version
  name             = "{{.Name}}url-updated" # Updated name
  resolve          = {{.Resolve}}
  input_entity_url = "{{.EntityUrl}}"

  depends_on = [vcd_rights_bundle.rde_type_bundle]
}
`

const testAccVcdRde3 = testAccVcdRdePrerequisites + `
resource "vcd_rde" "rde_file" {
  provider = {{.ProviderSystem}}

  rde_type_vendor  = vcd_rde_type.rde_type.vendor
  rde_type_nss     = vcd_rde_type.rde_type.nss
  rde_type_version = vcd_rde_type.rde_type.version
  name             = "{{.Name}}file-updated" # Updated name
  resolve          = {{.Resolve}}
  input_entity     = file("{{.EntityPath}}")

  depends_on = [vcd_rights_bundle.rde_type_bundle]
}

resource "vcd_rde" "rde_url" {
  provider = {{.ProviderSystem}}

  rde_type_vendor  = vcd_rde_type.rde_type.vendor
  rde_type_nss     = vcd_rde_type.rde_type.nss
  rde_type_version = vcd_rde_type.rde_type.version
  name             = "{{.Name}}url-updated" # Updated name
  resolve          = {{.Resolve}}
  input_entity_url = "{{.EntityUrl}}"

  depends_on = [vcd_rights_bundle.rde_type_bundle]
}

resource "vcd_rde" "rde_naughty" {
  provider = {{.ProviderSystem}}

  rde_type_vendor  = vcd_rde_type.rde_type.vendor
  rde_type_nss     = vcd_rde_type.rde_type.nss
  rde_type_version = vcd_rde_type.rde_type.version
  name             = "{{.Name}}naughty"
  resolve          = {{.Resolve}}
  input_entity     = file("{{.EntityPath}}") # Updated to a correct JSON

  depends_on = [vcd_rights_bundle.rde_type_bundle]
}

resource "vcd_rde" "rde_tenant" {
  provider = {{.ProviderOrg1}}

  rde_type_vendor  = vcd_rde_type.rde_type.vendor
  rde_type_nss     = vcd_rde_type.rde_type.nss
  rde_type_version = vcd_rde_type.rde_type.version
  name             = "{{.Name}}tenant-updated" # Updated name
  resolve          = {{.Resolve}}
  input_entity     = file("{{.EntityPath}}")

  depends_on = [vcd_rights_bundle.rde_type_bundle]
}
`

const testAccVcdRde4 = testAccVcdRde3 + `
# skip-binary-test - This should fail
resource "vcd_rde" "rde_naughty-clone" {
  provider = {{.ProviderSystem}}

  rde_type_vendor  = vcd_rde_type.rde_type.vendor
  rde_type_nss     = vcd_rde_type.rde_type.nss
  rde_type_version = vcd_rde_type.rde_type.version
  name             = "{{.Name}}naughty"
  resolve          = {{.Resolve}}
  input_entity     = file("{{.EntityPath}}")

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

func importStateIdRde(vendor, nss, version, name, position string, list bool) resource.ImportStateIdFunc {
	return func(*terraform.State) (string, error) {
		commonIdPart := vendor +
			ImportSeparator +
			nss +
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

// addRightsToTenantUser adds the RDE type (specified by vendor and nss) rights to the
// Organization Administrator global role, so a tenant user with this role can perform CRUD operations
// on RDEs.
// NOTE: We don't need to remove the added rights after the test is run, because the RDE Type and the Rights Bundle
// are destroyed and the rights disappear with them gone.
func addRightsToTenantUser(t *testing.T, vcdClient *VCDClient, vendor, nss string) {
	role, err := vcdClient.VCDClient.Client.GetGlobalRoleByName("Organization Administrator")
	if err != nil {
		t.Errorf("could not get Organization Administrator global role: %s", err)
	}
	rightsBundleName := fmt.Sprintf("%s:%s Entitlement", vendor, nss)
	rightsBundle, err := vcdClient.VCDClient.Client.GetRightsBundleByName(rightsBundleName)
	if err != nil {
		t.Errorf("could not get '%s' rights bundle: %s", rightsBundleName, err)
	}
	err = rightsBundle.PublishAllTenants()
	if err != nil {
		t.Errorf("could not publish '%s' rights bundle to all tenants: %s", rightsBundleName, err)
	}

	rights, err := rightsBundle.GetRights(nil)
	if err != nil {
		t.Errorf("could not get rights from '%s' rights bundle: %s", rightsBundleName, err)
	}
	var rightsToAdd []types.OpenApiReference
	for _, right := range rights {
		if strings.Contains(right.Name, fmt.Sprintf("%s:%s", vendor, nss)) {
			rightsToAdd = append(rightsToAdd, types.OpenApiReference{
				Name: right.Name,
				ID:   right.ID,
			})
		}
	}
	err = role.AddRights(rightsToAdd)
	if err != nil {
		t.Errorf("could not add rights '%v' to role '%s'", rightsToAdd, role.GlobalRole.Name)
	}
}

// manipulateRde mimics a 3rd party member that changes an RDE in VCD side. This is a common use-case in RDEs
func manipulateRde(t *testing.T, vcdClient *VCDClient, rdeId string) {
	rde, err := vcdClient.GetRdeById(rdeId)
	if err != nil {
		t.Errorf("could not get RDE with ID '%s': %s", rdeId, err)
	}

	rde.DefinedEntity.Entity["bar"] = "stringValueChanged"

	err = rde.Update(*rde.DefinedEntity)
	if err != nil {
		t.Errorf("could not update RDE with ID '%s': %s", rdeId, err)
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
		"", "",
		StringMap{})
}

const testAccCheckVcdRdeMetadata = `
data "vcd_rde_type" "rde_type" {
  vendor  = "vmware"
  nss     = "tkgcluster"
  version = "1.0.0"
}

resource "vcd_rde" "test-rde" {
  rde_type_vendor  = data.vcd_rde_type.rde_type.vendor
  rde_type_nss     = data.vcd_rde_type.rde_type.nss
  rde_type_version = data.vcd_rde_type.rde_type.version

  name         = "{{.Name}}"
  input_entity = "{\"foo\":\"bar\"}" # We are just testing metadata so we don't care about entity state

  {{.Metadata}}
}
`
