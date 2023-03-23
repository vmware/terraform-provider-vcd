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

func TestAccVcdRdeType(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	var params = StringMap{
		"ProviderVcdSystem":   providerVcdSystem,
		"ProviderVcdOrg1":     providerVcdOrg1,
		"Nss":                 "nss",
		"Version":             "1.0.0",
		"Vendor":              "vendor",
		"Name":                t.Name(),
		"Description":         "Created by " + t.Name(),
		"InterfaceReferences": "vcd_rde_interface.rde_interface1.id",
		"SchemaPath":          getCurrentDir() + "/../test-resources/rde_type.json",
		"SchemaUrl":           "https://raw.githubusercontent.com/vmware/terraform-provider-vcd/main/test-resources/rde_type.json",
	}
	testParamsNotEmpty(t, params)

	configTextCreate := templateFill(testAccVcdRdeType, params)
	params["FuncName"] = t.Name() + "-Update"
	params["Name"] = params["FuncName"]
	params["Description"] = "Created by" + params["FuncName"].(string)
	params["InterfaceReferences"] = "vcd_rde_interface.rde_interface1.id, vcd_rde_interface.rde_interface2.id"
	configTextUpdate := templateFill(testAccVcdRdeType, params)
	params["FuncName"] = t.Name() + "-WithTenantDS"
	configTextTenantDS := templateFill(testAccVcdRdeTypeTenantDS, params)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	debugPrintf("#[DEBUG] CONFIGURATION create: %s\n", configTextCreate)
	debugPrintf("#[DEBUG] CONFIGURATION update: %s\n", configTextUpdate)
	debugPrintf("#[DEBUG] CONFIGURATION with data source: %s\n", configTextTenantDS)

	rdeTypeFromFile := "vcd_rde_type.rde_type_file"
	rdeTypeFromUrl := "vcd_rde_type.rde_type_url"
	rdeTypeFromDS := "data.vcd_rde_type.rde_type_ds"

	vcdClient := createTemporaryVCDConnection(true)
	if vcdClient == nil || vcdClient.VCDClient == nil {
		t.Errorf("could not get a VCD connection to add rights to tenant user")
	}

	resource.Test(t, resource.TestCase{
		ProviderFactories: buildMultipleProviders(),
		CheckDestroy:      testAccCheckRdeTypesDestroy(rdeTypeFromFile, rdeTypeFromUrl),
		Steps: []resource.TestStep{
			{
				Config: configTextCreate,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(rdeTypeFromFile, "nss", params["Nss"].(string)+"file"),
					resource.TestCheckResourceAttr(rdeTypeFromFile, "version", params["Version"].(string)),
					resource.TestCheckResourceAttr(rdeTypeFromFile, "vendor", params["Vendor"].(string)+"file"),
					resource.TestCheckResourceAttr(rdeTypeFromFile, "name", t.Name()),
					resource.TestCheckResourceAttr(rdeTypeFromFile, "description", "Created by "+t.Name()),
					resource.TestCheckResourceAttrPair(rdeTypeFromFile, "interface_ids.0", "vcd_rde_interface.rde_interface1", "id"),
					resource.TestMatchResourceAttr(rdeTypeFromFile, "schema", regexp.MustCompile("{.*\"foo\".*\"bar\".*}")),

					resource.TestCheckResourceAttr(rdeTypeFromUrl, "nss", params["Nss"].(string)+"url"),
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
					resource.TestCheckResourceAttr(rdeTypeFromFile, "nss", params["Nss"].(string)+"file"),
					resource.TestCheckResourceAttr(rdeTypeFromFile, "version", params["Version"].(string)),
					resource.TestCheckResourceAttr(rdeTypeFromFile, "vendor", params["Vendor"].(string)+"file"),
					resource.TestCheckResourceAttr(rdeTypeFromFile, "name", t.Name()+"-Update"),
					resource.TestCheckResourceAttr(rdeTypeFromFile, "description", "Created by"+t.Name()+"-Update"),
					resource.TestCheckResourceAttr(rdeTypeFromFile, "interface_ids.#", "2"),

					resource.TestCheckResourceAttr(rdeTypeFromUrl, "nss", params["Nss"].(string)+"url"),
					resource.TestCheckResourceAttr(rdeTypeFromUrl, "vendor", params["Vendor"].(string)+"url"),

					resource.TestCheckResourceAttrPair(rdeTypeFromUrl, "version", rdeTypeFromFile, "version"),
					resource.TestCheckResourceAttrPair(rdeTypeFromUrl, "name", rdeTypeFromFile, "name"),
					resource.TestCheckResourceAttrPair(rdeTypeFromUrl, "description", rdeTypeFromFile, "description"),
					resource.TestCheckResourceAttrPair(rdeTypeFromUrl, "interface_ids.#", rdeTypeFromFile, "interface_ids.#"),
					resource.TestCheckResourceAttrPair(rdeTypeFromUrl, "schema", rdeTypeFromFile, "schema"),
				),
			},
			// With this step we check that a tenant with enough rights can read a RDE Type with a data source
			{
				Config: configTextTenantDS,
				PreConfig: func() {
					addRdeTypeRightsToTenantUser(t, vcdClient, params["Vendor"].(string)+"file", params["Nss"].(string)+"file")
				},
				Check: resourceFieldsEqual(rdeTypeFromDS, rdeTypeFromFile, []string{"%", "schema_url"}), // Exclude % as we don't have `schema_url`
			},
			{
				ResourceName:      rdeTypeFromFile,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdDefinedInterface(params["Vendor"].(string)+"file", params["Nss"].(string)+"file", params["Version"].(string)),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdRdeType = `
resource "vcd_rde_interface" "rde_interface1" {
  provider = {{.ProviderVcdSystem}}

  nss     = "namespace1"
  version = "1.0.0"
  vendor  = "vendor1"
  name    = "name1"
}

resource "vcd_rde_interface" "rde_interface2" {
  provider = {{.ProviderVcdSystem}}

  nss     = "namespace2"
  version = "2.0.0"
  vendor  = "vendor2"
  name    = "name2"
}

resource "vcd_rde_type" "rde_type_file" {
  provider = {{.ProviderVcdSystem}}

  nss           = "{{.Nss}}file"
  version       = "{{.Version}}"
  vendor        = "{{.Vendor}}file"
  name          = "{{.Name}}"
  description   = "{{.Description}}"
  interface_ids = [{{.InterfaceReferences}}]
  schema        = file("{{.SchemaPath}}")
}

resource "vcd_rde_type" "rde_type_url" {
  provider = {{.ProviderVcdSystem}}

  nss           = "{{.Nss}}url"
  version       = "{{.Version}}"
  vendor        = "{{.Vendor}}url"
  name          = "{{.Name}}"
  description   = "{{.Description}}"
  interface_ids = [{{.InterfaceReferences}}]
  schema_url    = "{{.SchemaUrl}}"
}
`

const testAccVcdRdeTypeTenantDS = testAccVcdRdeType + `
# skip-binary-test: Using a data source that references a resource to be created
data "vcd_rde_type" "rde_type_ds" {
  provider = {{.ProviderVcdOrg1}}

  nss     = "{{.Nss}}file"
  version = "{{.Version}}"
  vendor  = "{{.Vendor}}file"
}
`

// addRdeTypeRightsToTenantUser adds the RDE type (specified by vendor and nss) rights to the
// Organization Administrator global role, so a tenant user with this role can perform CRUD operations
// on RDEs.
// NOTE: We don't need to remove the added rights after the test is run, because the RDE Type and the Rights Bundle
// are destroyed and the rights disappear with them gone.
func addRdeTypeRightsToTenantUser(t *testing.T, vcdClient *VCDClient, vendor, nss string) {
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
	role, err := vcdClient.VCDClient.Client.GetGlobalRoleByName("Organization Administrator")
	if err != nil {
		t.Errorf("could not get Organization Administrator global role: %s", err)
	}
	err = role.AddRights(rightsToAdd)
	if err != nil {
		t.Errorf("could not add rights '%v' to role '%s'", rightsToAdd, role.GlobalRole.Name)
	}
}

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

// TestAccVcdRdeTypeValidation tests the validation rules for the RDE Type resource
func TestAccVcdRdeTypeValidation(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	var params = StringMap{
		"Nss":        "wrong%%%%",
		"Version":    "1.0.0",
		"Vendor":     "Vendor_0-9",
		"Name":       t.Name(),
		"SchemaPath": getCurrentDir() + "/../test-resources/rde_type.json",
	}
	testParamsNotEmpty(t, params)

	config1 := templateFill(testAccVcdRdeTypeWrongFields, params)
	params["FuncName"] = t.Name() + "2"
	params["Nss"] = "Nss_0-9"
	params["Vendor"] = "wrong%%%%"
	config2 := templateFill(testAccVcdRdeTypeWrongFields, params)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	debugPrintf("#[DEBUG] CONFIGURATION 1: %s\n", config1)
	debugPrintf("#[DEBUG] CONFIGURATION 2: %s\n", config2)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      config1,
				ExpectError: regexp.MustCompile(".*only alphanumeric characters, underscores and hyphens allowed.*"),
			},
			{
				Config:      config2,
				ExpectError: regexp.MustCompile(".*only alphanumeric characters, underscores and hyphens allowed.*"),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdRdeTypeWrongFields = `
# skip-binary-test - This test checks early failure validations
resource "vcd_rde_type" "rde_type_validation" {
  nss           = "{{.Nss}}"
  version       = "{{.Version}}"
  vendor        = "{{.Vendor}}"
  name          = "{{.Name}}"
  description   = "{{.Description}}"
  schema        = file("{{.SchemaPath}}")
}
`
