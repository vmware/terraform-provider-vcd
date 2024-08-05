//go:build rde || ALL || functional

package vcd

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// TestAccVcdRdeTypeBehaviorAndAcl tests both RDE Type Behaviors resource/data-source and the Access Level resource/data-source
func TestAccVcdRdeTypeBehaviorAndAcl(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	var params = StringMap{
		"Nss":                   "nss1",
		"Version":               "1.0.0",
		"Vendor":                "vendor1",
		"Name":                  t.Name(),
		"Description":           t.Name(),
		"SchemaPath":            getCurrentDir() + "/../test-resources/rde_type.json",
		"ExecutionId":           "MyActivity",
		"ExecutionType":         "noop",
		"TypeAccessLevels":      "\"urn:vcloud:accessLevel:FullControl\"",
		"InterfaceAccessLevels": "\"urn:vcloud:accessLevel:FullControl\"",
		"HookEvent":             "PostCreate",
	}
	testParamsNotEmpty(t, params)

	configText1 := templateFill(testAccVcdRdeTypeBehavior, params)
	debugPrintf("#[DEBUG] CONFIGURATION 1: %s\n", configText1)

	params["FuncName"] = t.Name() + "-Step2"
	params["Description"] = t.Name() + "Updated"
	params["ExecutionId"] = "MyActivityUpdated"
	params["TypeAccessLevels"] = "\"urn:vcloud:accessLevel:ReadOnly\""
	params["InterfaceAccessLevels"] = "\"urn:vcloud:accessLevel:FullControl\""
	params["HookEvent"] = "PreDelete"
	configText2 := templateFill(testAccVcdRdeTypeBehavior, params)
	debugPrintf("#[DEBUG] CONFIGURATION 2: %s\n", configText2)

	params["FuncName"] = t.Name() + "-Step3"
	configText3 := templateFill(testAccVcdRdeTypeBehavior+testAccVcdRdeTypeBehaviorDS, params)
	debugPrintf("#[DEBUG] CONFIGURATION 3: %s\n", configText3)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	interfaceBehavior1 := "vcd_rde_interface_behavior.behavior1"
	interfaceBehavior2 := "vcd_rde_interface_behavior.behavior2"
	interfaceBehavior3 := "vcd_rde_interface_behavior.behavior3"
	rdeTypeBehavior := "vcd_rde_type_behavior.behavior_override"
	rdeTypeBehavior2 := "vcd_rde_type_behavior.behavior_override2"
	rdeTypeBehaviorDS := "data.vcd_rde_type_behavior.behavior_override_ds"
	rdeTypeBehaviorInterfaceDS := "data.vcd_rde_type_behavior.behavior_interface_ds"
	interfaceBehaviorAcl := "vcd_rde_type_behavior_acl.interface_acl"
	interfaceBehaviorAclDS := "data.vcd_rde_type_behavior_acl.interface_acl_ds"
	rdeTypeBehaviorAcl := "vcd_rde_type_behavior_acl.type_acl"
	rdeTypeBehaviorAclDS := "data.vcd_rde_type_behavior_acl.type_acl_ds"
	rdeType := "vcd_rde_type.type"
	rdeTypeWithHooks := "vcd_rde_type.type_with_hooks"

	cachedId := &testCachedFieldValue{}
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckRdeTypesDestroy(rdeType, rdeTypeWithHooks), // If the RDE Type is destroyed, the Behavior is also destroyed.
		Steps: []resource.TestStep{
			{
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					cachedId.cacheTestResourceFieldValue(interfaceBehavior1, "id"),
					// RDE Type Behavior
					resource.TestCheckResourceAttrPair(rdeTypeBehavior, "ref", interfaceBehavior1, "id"),
					resource.TestCheckResourceAttr(rdeTypeBehavior, "name", t.Name()+"1"),
					resource.TestCheckResourceAttr(rdeTypeBehavior, "description", t.Name()+"Override"),
					resource.TestCheckResourceAttr(rdeTypeBehavior, "execution.id", "MyActivityOverride"),
					resource.TestCheckResourceAttr(rdeTypeBehavior, "execution.type", "noop"),
					resource.TestMatchResourceAttr(rdeTypeBehavior, "execution_json", regexp.MustCompile(`"type":.*"noop"`)),             // Because it's a simple map
					resource.TestMatchResourceAttr(rdeTypeBehavior, "execution_json", regexp.MustCompile(`"id":.*"MyActivityOverride"`)), // Because it's a simple map

					// RDE Type Behavior with complex JSON
					resource.TestCheckResourceAttrPair(rdeTypeBehavior2, "ref", interfaceBehavior3, "id"),
					resource.TestCheckResourceAttr(rdeTypeBehavior2, "name", t.Name()+"3"),
					resource.TestCheckResourceAttr(rdeTypeBehavior2, "description", t.Name()+"Override3"),
					resource.TestMatchResourceAttr(rdeTypeBehavior2, "execution_json", regexp.MustCompile(`"type":.*"noop"`)),
					resource.TestMatchResourceAttr(rdeTypeBehavior2, "execution_json", regexp.MustCompile(`"id":.*"MyActivityOverride3"`)),
					resource.TestCheckResourceAttr(rdeTypeBehavior2, "execution.%", "2"), // Because it's a simple map

					// Interface Access Levels
					resource.TestCheckResourceAttrPair(interfaceBehaviorAcl, "id", interfaceBehavior2, "id"),
					resource.TestCheckResourceAttrPair(interfaceBehaviorAcl, "behavior_id", interfaceBehaviorAcl, "id"),
					resource.TestCheckResourceAttr(interfaceBehaviorAcl, "access_level_ids.#", "1"),
					resource.TestCheckTypeSetElemAttr(interfaceBehaviorAcl, "access_level_ids.*", "urn:vcloud:accessLevel:FullControl"),

					// Type Access Levels
					resource.TestCheckResourceAttrPair(rdeTypeBehaviorAcl, "id", rdeTypeBehavior, "id"),
					resource.TestCheckResourceAttrPair(rdeTypeBehaviorAcl, "behavior_id", rdeTypeBehaviorAcl, "id"),
					resource.TestCheckResourceAttr(rdeTypeBehaviorAcl, "access_level_ids.#", "1"),
					resource.TestCheckTypeSetElemAttr(rdeTypeBehaviorAcl, "access_level_ids.*", "urn:vcloud:accessLevel:FullControl"),

					// RDE Type with Hooks
					resource.TestCheckTypeSetElemNestedAttrs(rdeTypeWithHooks, "hook.*", map[string]string{
						"event":       "PostCreate",
						"behavior_id": cachedId.fieldValue,
					}),
				),
			},
			{
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					// RDE Type Behavior
					resource.TestCheckResourceAttrPair(rdeTypeBehavior, "ref", interfaceBehavior1, "id"),
					resource.TestCheckResourceAttr(rdeTypeBehavior, "description", t.Name()+"UpdatedOverride"),
					resource.TestCheckResourceAttr(rdeTypeBehavior, "execution.id", "MyActivityUpdatedOverride"),
					resource.TestCheckResourceAttr(rdeTypeBehavior, "execution.type", "noop"),
					resource.TestMatchResourceAttr(rdeTypeBehavior, "execution_json", regexp.MustCompile(`"type":.*"noop"`)),                    // Because it's a simple map
					resource.TestMatchResourceAttr(rdeTypeBehavior, "execution_json", regexp.MustCompile(`"id":.*"MyActivityUpdatedOverride"`)), // Because it's a simple map

					// RDE Type Behavior with complex JSON
					resource.TestCheckResourceAttrPair(rdeTypeBehavior2, "ref", interfaceBehavior3, "id"),
					resource.TestCheckResourceAttr(rdeTypeBehavior2, "description", t.Name()+"UpdatedOverride3"),
					resource.TestMatchResourceAttr(rdeTypeBehavior2, "execution_json", regexp.MustCompile(`"type":.*"noop"`)),
					resource.TestMatchResourceAttr(rdeTypeBehavior2, "execution_json", regexp.MustCompile(`"id":.*"MyActivityUpdatedOverride3"`)),
					resource.TestCheckResourceAttr(rdeTypeBehavior2, "execution.%", "2"), // Because it's a simple map

					// Interface Access Levels
					resource.TestCheckResourceAttrPair(interfaceBehaviorAcl, "id", interfaceBehavior2, "id"),
					resource.TestCheckResourceAttrPair(interfaceBehaviorAcl, "behavior_id", interfaceBehaviorAcl, "id"),
					resource.TestCheckResourceAttr(interfaceBehaviorAcl, "access_level_ids.#", "1"),
					resource.TestCheckTypeSetElemAttr(interfaceBehaviorAcl, "access_level_ids.*", "urn:vcloud:accessLevel:FullControl"),

					// Type Access Levels
					resource.TestCheckResourceAttrPair(rdeTypeBehaviorAcl, "id", rdeTypeBehavior, "id"),
					resource.TestCheckResourceAttrPair(rdeTypeBehaviorAcl, "behavior_id", rdeTypeBehaviorAcl, "id"),
					resource.TestCheckResourceAttr(rdeTypeBehaviorAcl, "access_level_ids.#", "1"),
					resource.TestCheckTypeSetElemAttr(rdeTypeBehaviorAcl, "access_level_ids.*", "urn:vcloud:accessLevel:ReadOnly"),

					// RDE Type with Hooks
					resource.TestCheckTypeSetElemNestedAttrs(rdeTypeWithHooks, "hook.*", map[string]string{
						"event":       "PreDelete",
						"behavior_id": cachedId.fieldValue,
					}),
				),
			},
			{
				Config: configText3,
				Check: resource.ComposeAggregateTestCheckFunc(
					// RDE Type Behavior
					resource.TestCheckResourceAttrPair(rdeTypeBehaviorDS, "id", rdeTypeBehavior, "id"),
					resource.TestCheckResourceAttrPair(rdeTypeBehaviorDS, "ref", rdeTypeBehavior, "ref"),
					resource.TestCheckResourceAttrPair(rdeTypeBehaviorDS, "description", rdeTypeBehavior, "description"),
					resource.TestCheckResourceAttrPair(rdeTypeBehaviorDS, "execution.%", rdeTypeBehavior, "execution.%"),
					resource.TestCheckResourceAttrPair(rdeTypeBehaviorDS, "execution.id", rdeTypeBehavior, "execution.id"),
					resource.TestCheckResourceAttrPair(rdeTypeBehaviorDS, "execution.type", rdeTypeBehavior, "execution.type"),
					resource.TestCheckResourceAttrPair(rdeTypeBehaviorDS, "execution_json", rdeTypeBehavior, "execution_json"),

					// RDE Type Behavior from an Interface
					resource.TestCheckResourceAttrPair(rdeTypeBehaviorInterfaceDS, "id", interfaceBehavior2, "id"),
					resource.TestCheckResourceAttrPair(rdeTypeBehaviorInterfaceDS, "ref", interfaceBehavior2, "id"),
					resource.TestCheckResourceAttrPair(rdeTypeBehaviorInterfaceDS, "description", interfaceBehavior2, "description"),
					resource.TestCheckResourceAttrPair(rdeTypeBehaviorInterfaceDS, "execution.%", interfaceBehavior2, "execution.%"),
					resource.TestCheckResourceAttrPair(rdeTypeBehaviorInterfaceDS, "execution.id", interfaceBehavior2, "execution.id"),
					resource.TestCheckResourceAttrPair(rdeTypeBehaviorInterfaceDS, "execution.type", interfaceBehavior2, "execution.type"),
					resource.TestCheckResourceAttrPair(rdeTypeBehaviorInterfaceDS, "execution_json", interfaceBehavior2, "execution_json"),

					// Interface Access Levels
					resource.TestCheckResourceAttrPair(interfaceBehaviorAclDS, "id", interfaceBehaviorAcl, "id"),
					resource.TestCheckResourceAttrPair(interfaceBehaviorAclDS, "access_level_ids.#", interfaceBehaviorAcl, "access_level_ids.#"),
					resource.TestCheckTypeSetElemAttr(interfaceBehaviorAclDS, "access_level_ids.*", "urn:vcloud:accessLevel:FullControl"),

					// Type Access Levels
					resource.TestCheckResourceAttrPair(rdeTypeBehaviorAclDS, "id", rdeTypeBehaviorAcl, "id"),
					resource.TestCheckResourceAttrPair(rdeTypeBehaviorAclDS, "access_level_ids.#", interfaceBehaviorAcl, "access_level_ids.#"),
					resource.TestCheckTypeSetElemAttr(rdeTypeBehaviorAclDS, "access_level_ids.*", "urn:vcloud:accessLevel:ReadOnly"),
				),
			},
			{
				ResourceName:            rdeTypeBehavior,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdFunc:       importStateIdInterfaceBehavior(params["Vendor"].(string), params["Nss"].(string), params["Version"].(string), params["Name"].(string)+"1"),
				ImportStateVerifyIgnore: []string{"always_update_secure_execution_properties"}, // Cannot be imported, it's just a flag
			},
			{
				ResourceName:      interfaceBehaviorAcl,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdInterfaceBehavior(params["Vendor"].(string), params["Nss"].(string), params["Version"].(string), params["Name"].(string)+"1"),
			},
			{
				ResourceName:      rdeTypeBehaviorAcl,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdInterfaceBehavior(params["Vendor"].(string), params["Nss"].(string), params["Version"].(string), params["Name"].(string)+"1"),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdRdeTypeBehavior = `
resource "vcd_rde_interface" "interface" {
  nss     = "{{.Nss}}"
  version = "{{.Version}}"
  vendor  = "{{.Vendor}}"
  name    = "{{.Name}}"
}

resource "vcd_rde_interface_behavior" "behavior1" {
  rde_interface_id = vcd_rde_interface.interface.id
  name             = "{{.Name}}1"
  description      = "{{.Description}}"
  execution = {
    "id" : "{{.ExecutionId}}1"
    "type" : "{{.ExecutionType}}"
  }
}

resource "vcd_rde_interface_behavior" "behavior2" {
  rde_interface_id = vcd_rde_interface.interface.id
  name             = "{{.Name}}2"
  description      = "{{.Description}}"
  execution_json = jsonencode({
    "id" : "{{.ExecutionId}}2"
    "type" : "{{.ExecutionType}}"
  })
}

resource "vcd_rde_interface_behavior" "behavior3" {
  rde_interface_id = vcd_rde_interface.interface.id
  name             = "{{.Name}}3"
  description      = "{{.Description}}"
  execution_json = jsonencode({
    "id" : "{{.ExecutionId}}3"
    "type" : "{{.ExecutionType}}"
  })
}

resource "vcd_rde_type" "type" {
  nss           = "{{.Nss}}"
  version       = "{{.Version}}"
  vendor        = "{{.Vendor}}"
  name          = "{{.Name}}"
  description   = "{{.Description}}"
  interface_ids = [vcd_rde_interface.interface.id]
  schema        = file("{{.SchemaPath}}")

  # Behaviors can't be created after the RDE Interface is used by a RDE Type
  # so we need to depend on the Behaviors to wait for them to be created first.
  depends_on = [vcd_rde_interface_behavior.behavior1, vcd_rde_interface_behavior.behavior2]
}

resource "vcd_rde_type" "type_with_hooks" {
  nss           = "{{.Nss}}Hooks"
  version       = "{{.Version}}"
  vendor        = "{{.Vendor}}Hooks"
  name          = "{{.Name}}Hooks"
  description   = "{{.Description}}Hooks"
  interface_ids = [vcd_rde_interface.interface.id]
  schema        = file("{{.SchemaPath}}")
  hook {
    event       = "{{.HookEvent}}"
    behavior_id = vcd_rde_interface_behavior.behavior1.id
  }
}

resource "vcd_rde_type_behavior" "behavior_override" {
  rde_type_id               = vcd_rde_type.type.id
  rde_interface_behavior_id = vcd_rde_interface_behavior.behavior1.id
  description               = "{{.Description}}Override"
  execution = {
    "id" : "{{.ExecutionId}}Override"
    "type" : "{{.ExecutionType}}"
  }
}

resource "vcd_rde_type_behavior" "behavior_override2" {
  rde_type_id               = vcd_rde_type.type.id
  rde_interface_behavior_id = vcd_rde_interface_behavior.behavior3.id
  description               = "{{.Description}}Override3"
  execution_json = jsonencode({
    "id" : "{{.ExecutionId}}Override3"
    "type" : "{{.ExecutionType}}"
  })
}

resource "vcd_rde_type_behavior_acl" "type_acl" {
  rde_type_id = vcd_rde_type.type.id
  behavior_id = vcd_rde_type_behavior.behavior_override.id
  access_level_ids = [{{.TypeAccessLevels}}]
}

resource "vcd_rde_type_behavior_acl" "interface_acl" {
  rde_type_id = vcd_rde_type.type.id
  behavior_id = vcd_rde_interface_behavior.behavior2.id
  access_level_ids = [{{.InterfaceAccessLevels}}]
}
`

const testAccVcdRdeTypeBehaviorDS = `
# We fetch an RDE Type Behavior
data "vcd_rde_type_behavior" "behavior_override_ds" {
  rde_type_id = vcd_rde_type_behavior.behavior_override.rde_type_id
  behavior_id = vcd_rde_type_behavior.behavior_override.id
}

# In this case we fetch a RDE Interface Behavior inherited automatically
data "vcd_rde_type_behavior" "behavior_interface_ds" {
  rde_type_id = vcd_rde_type_behavior.behavior_override.rde_type_id
  behavior_id = vcd_rde_interface_behavior.behavior2.id
}

data "vcd_rde_type_behavior_acl" "interface_acl_ds" {
  rde_type_id = vcd_rde_type_behavior_acl.interface_acl.rde_type_id
  behavior_id = vcd_rde_type_behavior_acl.interface_acl.behavior_id
}

data "vcd_rde_type_behavior_acl" "type_acl_ds" {
  rde_type_id = vcd_rde_type_behavior_acl.type_acl.rde_type_id
  behavior_id = vcd_rde_type_behavior_acl.type_acl.behavior_id
}
`
