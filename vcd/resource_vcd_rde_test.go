//go:build rde || ALL || functional

package vcd

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
)

// TestAccVcdRde tests the behaviour of RDE instances.
func TestAccVcdRde(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	var params = StringMap{
		"FuncName":       t.Name() + "-Step1-and-2",
		"Org":            testConfig.VCD.Org,
		"ProviderSystem": providerVcdSystem,
		"ProviderOrg1":   providerVcdOrg1,
		"Nss":            "nss",
		"Version":        "1.0.0",
		"Vendor":         "vendor",
		"Name":           t.Name(),
		"Resolve":        false,
		"SchemaPath":     getCurrentDir() + "/../test-resources/rde_type.json",
		"EntityPath":     getCurrentDir() + "/../test-resources/rde_instance.json",
		"EntityUrl":      "https://raw.githubusercontent.com/vmware/terraform-provider-vcd/main/test-resources/rde_instance.json",
		"Tags":           "rde",
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
	params["FuncName"] = t.Name() + "-DS"
	stepDataSource := templateFill(testAccVcdRdeDS, params)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	debugPrintf("#[DEBUG] CONFIGURATION preReqs: %s\n", preReqsConfig)
	debugPrintf("#[DEBUG] CONFIGURATION init: %s\n", stepInit)
	debugPrintf("#[DEBUG] CONFIGURATION resolve: %s\n", stepResolve)
	debugPrintf("#[DEBUG] CONFIGURATION fix wrong RDE: %s\n", stepFixWrongRde)
	debugPrintf("#[DEBUG] CONFIGURATION data source: %s\n", stepDataSource)

	rdeUrnRegexp := fmt.Sprintf(`urn:vcloud:entity:%s:%s:[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$`, params["Vendor"].(string), params["Nss"].(string))
	rdeType := "vcd_rde_type.rde_type"
	rdeFromFile := "vcd_rde.rde_file"
	rdeFromUrl := "vcd_rde.rde_url"
	rdeWrong := "vcd_rde.rde_naughty"
	rdeTenant := "vcd_rde.rde_tenant"
	rdeDataSource1 := "data.vcd_rde.existing_rde1"
	rdeDataSource2 := "data.vcd_rde.existing_rde2"

	// We will cache some RDE identifiers, so we can use them later for importing and other steps
	cachedIds := make([]testCachedFieldValue, 2)

	vcdClient := createTemporaryVCDConnection(true)
	if vcdClient == nil || vcdClient.VCDClient == nil {
		t.Errorf("could not get a VCD connection to add rights to tenant user")
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: buildMultipleProviders(),
		CheckDestroy:      testAccCheckRdeDestroy(rdeType, rdeFromFile, rdeFromUrl),
		Steps: []resource.TestStep{
			// Preconfigure the test: We fetch an existing interface and create an RDE Type.
			// Creating an RDE Type results in the creation of an associated Rights Bundle that is NOT published to all tenants by default.
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
					// This function needs to be called with fresh clients (no cached ones), as it modifies
					// rights of the tenant user.
					addRightsToTenantUser(t, vcdClient, params["Vendor"].(string), params["Nss"].(string))
					// We need to invalidate existing client cache and start a new one as the rights for the tenant user have changed, hence
					// we can't reuse existing sessions
					cachedVCDClients.reset()
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					// We cache some IDs to use it on later steps
					cachedIds[0].cacheTestResourceFieldValue(rdeFromFile, "id"),
					cachedIds[1].cacheTestResourceFieldValue(rdeFromUrl, "id"),

					resource.TestMatchResourceAttr(rdeFromFile, "id", regexp.MustCompile(rdeUrnRegexp)),
					resource.TestCheckResourceAttr(rdeFromFile, "name", t.Name()+"file"),
					resource.TestCheckResourceAttrPair(rdeFromFile, "rde_type_id", rdeType, "id"),
					resource.TestMatchResourceAttr(rdeFromFile, "computed_entity", regexp.MustCompile("{.*\"stringValue\".*}")),
					resource.TestCheckResourceAttr(rdeFromFile, "state", "PRE_CREATED"),
					resource.TestMatchResourceAttr(rdeFromFile, "org_id", getUuidRegex("urn:vcloud:org:", "$")),
					resource.TestMatchResourceAttr(rdeFromFile, "owner_user_id", getUuidRegex("urn:vcloud:user:", "$")),
					resource.TestCheckResourceAttr(rdeFromFile, "entity_in_sync", "true"),

					resource.TestMatchResourceAttr(rdeFromUrl, "id", regexp.MustCompile(rdeUrnRegexp)),
					resource.TestCheckResourceAttr(rdeFromUrl, "name", t.Name()+"url"),
					resource.TestCheckResourceAttrPair(rdeFromUrl, "rde_type_id", rdeType, "id"),
					resource.TestCheckResourceAttrPair(rdeFromUrl, "computed_entity", rdeFromFile, "computed_entity"),
					resource.TestCheckResourceAttr(rdeFromUrl, "state", "PRE_CREATED"),
					resource.TestCheckResourceAttrPair(rdeFromUrl, "org_id", rdeFromFile, "org_id"),
					resource.TestCheckResourceAttrPair(rdeFromUrl, "owner_user_id", rdeFromFile, "owner_user_id"),
					resource.TestCheckResourceAttr(rdeFromUrl, "entity_in_sync", "true"),

					resource.TestMatchResourceAttr(rdeWrong, "id", regexp.MustCompile(rdeUrnRegexp)),
					resource.TestCheckResourceAttr(rdeWrong, "name", t.Name()+"naughty"),
					resource.TestCheckResourceAttrPair(rdeWrong, "rde_type_id", rdeType, "id"),
					resource.TestCheckResourceAttr(rdeWrong, "computed_entity", "{\"this_json_is_bad\":\"yes\"}"),
					resource.TestCheckResourceAttr(rdeWrong, "state", "PRE_CREATED"),
					resource.TestCheckResourceAttrPair(rdeWrong, "org_id", rdeFromFile, "org_id"),
					resource.TestCheckResourceAttrPair(rdeWrong, "owner_user_id", rdeFromFile, "owner_user_id"),
					resource.TestCheckResourceAttr(rdeWrong, "entity_in_sync", "true"),

					resource.TestMatchResourceAttr(rdeTenant, "id", regexp.MustCompile(rdeUrnRegexp)),
					resource.TestCheckResourceAttr(rdeTenant, "name", t.Name()+"tenant"),
					resource.TestCheckResourceAttrPair(rdeTenant, "rde_type_id", rdeType, "id"),
					resource.TestCheckResourceAttrPair(rdeTenant, "computed_entity", rdeFromFile, "computed_entity"),
					resource.TestCheckResourceAttr(rdeTenant, "state", "PRE_CREATED"),
					resource.TestCheckResourceAttrPair(rdeTenant, "org_id", rdeFromFile, "org_id"),
					resource.TestMatchResourceAttr(rdeTenant, "owner_user_id", getUuidRegex("urn:vcloud:user:", "$")), // Owner is different in this case
					resource.TestCheckResourceAttr(rdeTenant, "entity_in_sync", "true"),
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
					manipulateRde(t, vcdClient, cachedIds[0].fieldValue) // Changes the RDE from file
					manipulateRde(t, vcdClient, cachedIds[1].fieldValue) // Changes the RDE from URL
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestMatchResourceAttr(rdeFromFile, "computed_entity", regexp.MustCompile(`.*stringValueChanged.*`)),
					resource.TestCheckResourceAttr(rdeFromFile, "entity_in_sync", "false"),
					resource.TestMatchResourceAttr(rdeFromUrl, "computed_entity", regexp.MustCompile(`.*stringValueChanged.*`)),
					resource.TestCheckResourceAttr(rdeFromUrl, "entity_in_sync", "false"),
				),
			},
			{
				Config: stepDataSource,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(rdeDataSource1, "id", rdeFromFile, "id"),
					resource.TestCheckResourceAttrPair(rdeDataSource1, "external_id", rdeFromFile, "external_id"),
					resource.TestCheckResourceAttrPair(rdeDataSource1, "entity", rdeFromFile, "computed_entity"),
					resource.TestCheckResourceAttrPair(rdeDataSource1, "state", rdeFromFile, "state"),
					resource.TestCheckResourceAttrPair(rdeDataSource1, "org_id", rdeFromFile, "org_id"),
					resource.TestCheckResourceAttrPair(rdeDataSource1, "owner_user_id", rdeFromFile, "owner_user_id"),
					resource.TestCheckResourceAttrPair(rdeDataSource2, "id", rdeFromFile, "id"),
					resource.TestCheckResourceAttrPair(rdeDataSource2, "external_id", rdeFromFile, "external_id"),
					resource.TestCheckResourceAttrPair(rdeDataSource2, "entity", rdeFromFile, "computed_entity"),
					resource.TestCheckResourceAttrPair(rdeDataSource2, "state", rdeFromFile, "state"),
					resource.TestCheckResourceAttrPair(rdeDataSource2, "org_id", rdeFromFile, "org_id"),
					resource.TestCheckResourceAttrPair(rdeDataSource2, "owner_user_id", rdeFromFile, "owner_user_id"),
				),
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
					return cachedIds[0].fieldValue, nil
				},
				ImportStateVerifyIgnore: []string{"resolve", "input_entity", "input_entity_url"},
			},
			// Import with the list option, it should return the RDE that we cached
			{
				ResourceName:      rdeFromFile,
				ImportState:       true,
				ImportStateIdFunc: importStateIdRde(params["Vendor"].(string), params["Nss"].(string), params["Version"].(string), t.Name()+"file-updated", "1", true),
				ExpectError:       regexp.MustCompile(`.*` + cachedIds[0].fieldValue + `.*`),
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
# skip-binary-test - This would require additional rights in tenant user
resource "vcd_rde" "rde_file" {
  provider = {{.ProviderSystem}}

  org          = "{{.Org}}"
  rde_type_id  = vcd_rde_type.rde_type.id
  name         = "{{.Name}}file"
  resolve      = {{.Resolve}}
  input_entity = file("{{.EntityPath}}")

  depends_on = [vcd_rights_bundle.rde_type_bundle]
}

resource "vcd_rde" "rde_url" {
  provider = {{.ProviderSystem}}

  org                = "{{.Org}}"
  rde_type_id        = vcd_rde_type.rde_type.id
  name               = "{{.Name}}url"
  resolve            = {{.Resolve}}
  input_entity_url   = "{{.EntityUrl}}"
  resolve_on_removal = false
  depends_on = [vcd_rights_bundle.rde_type_bundle]
}

resource "vcd_rde" "rde_naughty" {
  provider = {{.ProviderSystem}}

  org                = "{{.Org}}"
  rde_type_id        = vcd_rde_type.rde_type.id
  name               = "{{.Name}}naughty"
  resolve      		 = {{.Resolve}}
  input_entity       = "{ \"this_json_is_bad\": \"yes\"}"
  resolve_on_removal = false

  depends_on = [vcd_rights_bundle.rde_type_bundle]
}

resource "vcd_rde" "rde_tenant" {
  provider = {{.ProviderOrg1}}

  org                = "{{.Org}}"
  rde_type_id        = vcd_rde_type.rde_type.id
  name               = "{{.Name}}tenant"
  resolve            = {{.Resolve}}
  input_entity       = file("{{.EntityPath}}")
  resolve_on_removal = false

  depends_on = [vcd_rights_bundle.rde_type_bundle]
}
`

const testAccVcdRde2 = testAccVcdRdePrerequisites + `
# skip-binary-test - Deletion should fail
resource "vcd_rde" "rde_file" {
  provider = {{.ProviderSystem}}

  org                = "{{.Org}}"
  rde_type_id        = vcd_rde_type.rde_type.id
  name               = "{{.Name}}file-updated" # Updated name
  resolve            = {{.Resolve}}
  input_entity       = file("{{.EntityPath}}")
  resolve_on_removal = false

  depends_on = [vcd_rights_bundle.rde_type_bundle]
}

resource "vcd_rde" "rde_url" {
  provider = {{.ProviderSystem}}

  org                = "{{.Org}}"
  rde_type_id        = vcd_rde_type.rde_type.id
  name               = "{{.Name}}url-updated" # Updated name
  resolve            = {{.Resolve}}
  input_entity_url   = "{{.EntityUrl}}"
  resolve_on_removal = false

  depends_on = [vcd_rights_bundle.rde_type_bundle]
}
`

const testAccVcdRde3 = testAccVcdRdePrerequisites + `
# skip-binary-test - This would require additional rights in tenant user
resource "vcd_rde" "rde_file" {
  provider = {{.ProviderSystem}}

  org                = "{{.Org}}"
  rde_type_id        = vcd_rde_type.rde_type.id
  name               = "{{.Name}}file-updated" # Updated name
  resolve            = {{.Resolve}}
  input_entity       = file("{{.EntityPath}}")
  resolve_on_removal = false

  depends_on = [vcd_rights_bundle.rde_type_bundle]
}

resource "vcd_rde" "rde_url" {
  provider = {{.ProviderSystem}}

  org                = "{{.Org}}"
  rde_type_id        = vcd_rde_type.rde_type.id
  name               = "{{.Name}}url-updated" # Updated name
  resolve            = {{.Resolve}}
  input_entity_url   = "{{.EntityUrl}}"
  resolve_on_removal = false

  depends_on = [vcd_rights_bundle.rde_type_bundle]
}

resource "vcd_rde" "rde_naughty" {
  provider = {{.ProviderSystem}}

  org                = "{{.Org}}"
  rde_type_id        = vcd_rde_type.rde_type.id
  name               = "{{.Name}}naughty"
  resolve            = {{.Resolve}}
  input_entity       = file("{{.EntityPath}}") # Updated to a correct JSON
  resolve_on_removal = false

  depends_on = [vcd_rights_bundle.rde_type_bundle]
}

resource "vcd_rde" "rde_tenant" {
  provider = {{.ProviderOrg1}}

  org                = "{{.Org}}"
  rde_type_id        = vcd_rde_type.rde_type.id
  name               = "{{.Name}}tenant-updated" # Updated name
  resolve            = {{.Resolve}}
  input_entity       = file("{{.EntityPath}}")
  resolve_on_removal = false

  depends_on = [vcd_rights_bundle.rde_type_bundle]
}
`

const testAccVcdRdeDS = testAccVcdRde3 + `
# skip-binary-test - Contains data source referencing a resource
data "vcd_rde" "existing_rde1" {
  provider = {{.ProviderSystem}}

  rde_type_id = vcd_rde_type.rde_type.id
  name        = "{{.Name}}file-updated"
}

data "vcd_rde" "existing_rde2" {
  provider = {{.ProviderOrg1}}

  rde_type_id = vcd_rde_type.rde_type.id
  name        = "{{.Name}}file-updated"
}
`

// testAccCheckRdeDestroy checks that the RDE instances defined by their identifiers no longer
// exist in VCD.
func testAccCheckRdeDestroy(rdeId string, identifiers ...string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, identifier := range identifiers {
			rde, ok := s.RootModule().Resources[rdeId]
			if !ok {
				return fmt.Errorf("not found: %s", identifier)
			}

			if rde.Primary.ID == "" {
				return fmt.Errorf("no RDE ID is set")
			}

			conn := testAccProvider.Meta().(*VCDClient)

			_, err := conn.VCDClient.GetRdeById(rde.Primary.ID)

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
		t.Fatalf("could not get Organization Administrator global role: %s", err)
	}
	rightsBundleName := fmt.Sprintf("%s:%s Entitlement", vendor, nss)
	rightsBundle, err := vcdClient.VCDClient.Client.GetRightsBundleByName(rightsBundleName)
	if err != nil {
		t.Fatalf("could not get '%s' rights bundle: %s", rightsBundleName, err)
	}
	err = rightsBundle.PublishAllTenants()
	if err != nil {
		t.Fatalf("could not publish '%s' rights bundle to all tenants: %s", rightsBundleName, err)
	}

	rights, err := rightsBundle.GetRights(nil)
	if err != nil {
		t.Fatalf("could not get rights from '%s' rights bundle: %s", rightsBundleName, err)
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
		t.Fatalf("could not add rights '%v' to role '%s'", rightsToAdd, role.GlobalRole.Name)
	}
}

// manipulateRde mimics a 3rd party member that changes an RDE in VCD side. This is a common use-case in RDEs
func manipulateRde(t *testing.T, vcdClient *VCDClient, rdeId string) {
	rde, err := vcdClient.GetRdeById(rdeId)
	if err != nil {
		t.Fatalf("could not get RDE with ID '%s': %s", rdeId, err)
	}

	rde.DefinedEntity.Entity["bar"] = "stringValueChanged"

	err = rde.Update(*rde.DefinedEntity)
	if err != nil {
		t.Fatalf("could not update RDE with ID '%s': %s", rdeId, err)
	}
}

// TestAccVcdRdeMetadata tests metadata CRUD on Runtime Defined Entities.
func TestAccVcdRdeMetadata(t *testing.T) {
	skipIfNotSysAdmin(t)
	testOpenApiMetadataEntryCRUD(t,
		testAccCheckVcdRdeMetadata, "vcd_rde.test_rde",
		testAccCheckVcdRdeMetadataDatasource, "data.vcd_rde.test_rde_ds",
		StringMap{})
}

const testAccCheckVcdRdeMetadata = `
data "vcd_rde_type" "rde_type" {
  vendor  = "vmware"
  nss     = "tkgcluster"
  version = "1.0.0"
}
resource "vcd_rde" "test_rde" {
  org          = "System"
  rde_type_id  = data.vcd_rde_type.rde_type.id
  name         = "{{.Name}}"
  input_entity = "{\"foo\":\"bar\"}" # We are just testing metadata so we don't care about entity state
  resolve      = true
  {{.Metadata}}
}
`

const testAccCheckVcdRdeMetadataDatasource = `
data "vcd_rde" "test_rde_ds" {
  org         = "System"
  rde_type_id = vcd_rde.test_rde.rde_type_id
  name        = vcd_rde.test_rde.name
}
`

func TestAccVcdRdeMetadataIgnore(t *testing.T) {
	skipIfNotSysAdmin(t)

	getObjectById := func(vcdClient *VCDClient, id string) (openApiMetadataCompatible, error) {
		rde, err := vcdClient.GetRdeById(id)
		if err != nil {
			return nil, fmt.Errorf("could not retrieve RDE '%s': %s", id, err)
		}
		return rde, nil
	}

	testOpenApiMetadataEntryIgnore(t,
		testAccCheckVcdRdeMetadata, "vcd_rde.test_rde",
		testAccCheckVcdRdeMetadataDatasource, "data.vcd_rde.test_rde_ds",
		getObjectById, StringMap{})
}

// TestAccVcdRdeTenantMetadata tests that a tenant user cannot read metadata that was created in the RDE with domain = "PROVIDER".
// It will only be able to read those entries with domain = "TENANT".
func TestAccVcdRdeTenantMetadata(t *testing.T) {
	skipIfNotSysAdmin(t)
	preTestChecks(t)
	var params = StringMap{
		"FuncName":        t.Name() + "-Step1",
		"ProviderVcdOrg1": providerVcdOrg1,
		"Org":             testConfig.VCD.Org,
		"Name":            t.Name(),
		"Metadata":        getOpenApiMetadataTestingHcl(1, 0, 0, 0, 0, 3, 0),
		"SchemaPath":      getCurrentDir() + "/../test-resources/rde_type.json",
	}
	testParamsNotEmpty(t, params)

	step1 := templateFill(testAccCheckVcdRdeTenantMetadata, params)
	debugPrintf("#[DEBUG] CONFIGURATION Step 1: %s", step1)

	params["FuncName"] = t.Name() + "-Step2"
	step2 := templateFill(testAccCheckVcdRdeTenantMetadata+testAccCheckVcdRdeMetadataTenantDatasource, params)
	debugPrintf("#[DEBUG] CONFIGURATION Step 2: %s", step2)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	resourceName := "vcd_rde.test_rde"
	datasourceName := "data.vcd_rde.test_rde_ds"

	// Required to manipulate rights
	vcdClient := createTemporaryVCDConnection(true)
	if vcdClient == nil || vcdClient.VCDClient == nil {
		t.Errorf("could not get a VCD connection to add rights to tenant user")
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: buildMultipleProviders(),
		Steps: []resource.TestStep{
			{
				Config: step1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", t.Name()),
					resource.TestCheckResourceAttr(resourceName, "metadata_entry.#", "4"),
					testCheckOpenApiMetadataEntrySetElemNestedAttrs(resourceName, "stringKey1", "stringValue1", types.OpenApiMetadataStringEntry, "TENANT", "", "false", "false"),
					testCheckOpenApiMetadataEntrySetElemNestedAttrs(resourceName, "provider1", "provider1", types.OpenApiMetadataStringEntry, "PROVIDER", "", "false", "false"),
					testCheckOpenApiMetadataEntrySetElemNestedAttrs(resourceName, "provider2", "provider2", types.OpenApiMetadataStringEntry, "PROVIDER", "", "false", "false"),
					testCheckOpenApiMetadataEntrySetElemNestedAttrs(resourceName, "provider3", "provider3", types.OpenApiMetadataStringEntry, "PROVIDER", "", "false", "false"),
				),
			},
			{
				Config: step2,
				PreConfig: func() {
					// Rights will be deleted with the destruction of the RDE Type.
					addRightsToTenantUser(t, vcdClient, "vmware", params["Name"].(string))
					// We need to invalidate existing client cache and start a new one as the rights for the tenant user have changed, hence
					// we can't reuse existing sessions
					cachedVCDClients.reset()
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, "name", t.Name()),
					resource.TestCheckResourceAttr(datasourceName, "metadata_entry.#", "1"),
					testCheckOpenApiMetadataEntrySetElemNestedAttrs(resourceName, "stringKey1", "stringValue1", types.OpenApiMetadataStringEntry, "TENANT", "", "false", "false"),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccCheckVcdRdeTenantMetadata = `
# skip-binary: This is already tested by TestAccVcdRdeMetadata
resource "vcd_rde_type" "rde_type" {
  name    = "{{.Name}}"
  vendor  = "vmware"
  nss     = "{{.Name}}"
  version = "1.0.0"
  schema  = file("{{.SchemaPath}}")
}

# This is required because the organization where the RDE is created
# is lacking of the rights created by the type above.
# This bundle will be automatically removed by VCD with the destruction of the type.
resource "vcd_rights_bundle" "rde_type_bundle" {
  name                   = "{{.Name}} bundle"
  description            = "{{.Name}} bundle"
  publish_to_all_tenants = true
  rights = [
    "vmware:{{.Name}}: Administrator Full access",
    "vmware:{{.Name}}: Full Access",
    "vmware:{{.Name}}: Modify",
    "vmware:{{.Name}}: View",
    "vmware:{{.Name}}: Administrator View",
  ]
  depends_on = [vcd_rde_type.rde_type]
}

resource "vcd_rde" "test_rde" {
  org          = "{{.Org}}"
  rde_type_id  = vcd_rde_type.rde_type.id
  name         = "{{.Name}}"
  input_entity = "{\"foo\":\"bar\"}" # We are just testing metadata so we don't care about entity state
  resolve      = true
  {{.Metadata}}

  depends_on = [vcd_rights_bundle.rde_type_bundle] # We need to wait for the rights to be published, otherwise creation will fail
}
`

const testAccCheckVcdRdeMetadataTenantDatasource = `
# skip-binary-test: This requires manual publishing of rights
data "vcd_rde" "test_rde_ds" {
  provider = {{.ProviderVcdOrg1}}

  org         = vcd_rde.test_rde.org
  rde_type_id = vcd_rde.test_rde.rde_type_id
  name        = vcd_rde.test_rde.name
}
`
