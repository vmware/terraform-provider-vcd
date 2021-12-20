//go:build network || nsxt || ALL || functional
// +build network nsxt ALL functional

package vcd

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// TestAccVcdNsxtAppPortProfileTenant tests possible options for tenant scope
func TestAccVcdNsxtAppPortProfileTenant(t *testing.T) {
	preTestChecks(t)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	vcdClient := createTemporaryVCDConnection(false)
	if vcdClient.Client.APIVCDMaxVersionIs("< 34.0") {
		t.Skip(t.Name() + " requires at least API v34.0 (vCD 10.1.1+)")
	}
	skipNoNsxtConfiguration(t)

	var params = StringMap{
		"Org":     testConfig.VCD.Org,
		"NsxtVdc": testConfig.Nsxt.Vdc,
		"Tags":    "nsxt network",
	}

	configText1 := templateFill(testAccVcdNsxtAppPortProfileTenantStep1, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)

	params["FuncName"] = t.Name() + "-step2"
	configText2 := templateFill(testAccVcdNsxtAppPortProfileTenantStep1Ds, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText2)

	params["FuncName"] = t.Name() + "-step3"
	configText3 := templateFill(testAccVcdNsxtAppPortProfileTenantStep3, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText3)

	params["FuncName"] = t.Name() + "-step4"
	configText4 := templateFill(testAccVcdNsxtAppPortProfileTenantStep4, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 3: %s", configText4)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckOpenApiNsxtAppPortDestroy("custom_app_prof", "PROVIDER"),
			testAccCheckOpenApiNsxtAppPortDestroy("custom_app_prof", "PROVIDER"),
			testAccCheckOpenApiNsxtAppPortDestroy("custom_app_prof", "TENANT"),
			testAccCheckOpenApiNsxtAppPortDestroy("custom_app_prof", "TENANT"),
			testAccCheckOpenApiNsxtAppPortDestroy("custom_app_prof", "SYSTEM"),
			testAccCheckOpenApiNsxtAppPortDestroy("custom_app_prof", "SYSTEM"),
		),
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_nsxt_app_port_profile.custom", "id"),
					resource.TestCheckResourceAttr("vcd_nsxt_app_port_profile.custom", "name", "custom_app_prof"),
					resource.TestCheckResourceAttr("vcd_nsxt_app_port_profile.custom", "description", "Application port profile for custom"),
					resource.TestCheckResourceAttr("vcd_nsxt_app_port_profile.custom", "scope", "TENANT"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_app_port_profile.custom", "app_port.*", map[string]string{
						"protocol": "ICMPv4",
					}),
				),
			},
			resource.TestStep{
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_nsxt_app_port_profile.custom", "id"),
					resource.TestCheckResourceAttr("vcd_nsxt_app_port_profile.custom", "name", "custom_app_prof"),
					resource.TestCheckResourceAttr("vcd_nsxt_app_port_profile.custom", "description", "Application port profile for custom"),
					resource.TestCheckResourceAttr("vcd_nsxt_app_port_profile.custom", "scope", "TENANT"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_app_port_profile.custom", "app_port.*", map[string]string{
						"protocol": "ICMPv4",
					}),
					resourceFieldsEqual("vcd_nsxt_app_port_profile.custom", "data.vcd_nsxt_app_port_profile.custom", []string{}),
				),
			},
			resource.TestStep{
				Config: configText3,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_nsxt_app_port_profile.custom", "id"),
					resource.TestCheckResourceAttr("vcd_nsxt_app_port_profile.custom", "name", "custom_app_prof-updated"),
					resource.TestCheckResourceAttr("vcd_nsxt_app_port_profile.custom", "description", "Application port profile for custom-updated"),
					resource.TestCheckResourceAttr("vcd_nsxt_app_port_profile.custom", "scope", "TENANT"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_app_port_profile.custom", "app_port.*", map[string]string{
						"protocol": "ICMPv6",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_app_port_profile.custom", "app_port.*", map[string]string{
						"protocol": "TCP",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_app_port_profile.custom", "app_port.*", map[string]string{
						"protocol": "UDP",
					}),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_app_port_profile.custom", "app_port.*.port.*", "2000"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_app_port_profile.custom", "app_port.*.port.*", "2010-2020"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_app_port_profile.custom", "app_port.*.port.*", "12345"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_app_port_profile.custom", "app_port.*.port.*", "65000"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_app_port_profile.custom", "app_port.*.port.*", "40000-60000"),
				),
			},
			resource.TestStep{
				Config: configText4,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_nsxt_app_port_profile.custom", "id"),
					resource.TestCheckResourceAttr("vcd_nsxt_app_port_profile.custom", "name", "custom_app_prof-updated"),
					resource.TestCheckResourceAttr("vcd_nsxt_app_port_profile.custom", "description", ""),
					resource.TestCheckResourceAttr("vcd_nsxt_app_port_profile.custom", "scope", "TENANT"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_app_port_profile.custom", "app_port.*", map[string]string{
						"protocol": "ICMPv6",
					}),
				),
			},
			resource.TestStep{
				ResourceName:      "vcd_nsxt_app_port_profile.custom",
				ImportState:       true,
				ImportStateVerify: true,
				// This will generate import name of org_name.vdc_name.app_profile_name
				ImportStateIdFunc: importStateIdOrgNsxtVdcObject(testConfig, "custom_app_prof-updated"),
			},
		},
	})
	postTestChecks(t)
}

func TestAccVcdNsxtAppPortProfileProvider(t *testing.T) {
	preTestChecks(t)
	skipNoNsxtConfiguration(t)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	vcdClient := createTemporaryVCDConnection(false)
	if !vcdClient.Client.IsSysAdmin {
		t.Skip(t.Name() + " only System Administrator can create Provider scoped Application Port Profiles")
	}

	if vcdClient.Client.APIVCDMaxVersionIs("< 34.0") {
		t.Skip(t.Name() + " requires at least API v34.0 (vCD 10.1.1+)")
	}

	var params = StringMap{
		"Org":         "System",
		"NsxtVdc":     testConfig.Nsxt.Vdc,
		"NsxtManager": testConfig.Nsxt.Manager,
		"Tags":        "nsxt network",
	}

	configText1 := templateFill(testAccVcdNsxtAppPortProfileProviderStep1, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)

	params["FuncName"] = t.Name() + "-step2"
	configText2 := templateFill(testAccVcdNsxtAppPortProfileProviderStep1AndDS, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText2)

	params["FuncName"] = t.Name() + "-step3"
	configText3 := templateFill(testAccVcdNsxtAppPortProfileProviderStep2, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 3: %s", configText3)

	params["FuncName"] = t.Name() + "-step4"
	configText4 := templateFill(testAccVcdNsxtAppPortProfileProviderStep4, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 4: %s", configText4)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		PreCheck:          func() { testAccPreCheck(t) },
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckOpenApiNsxtAppPortDestroy("custom_app_prof", "PROVIDER"),
			testAccCheckOpenApiNsxtAppPortDestroy("custom_app_prof", "PROVIDER"),
			testAccCheckOpenApiNsxtAppPortDestroy("custom_app_prof", "TENANT"),
			testAccCheckOpenApiNsxtAppPortDestroy("custom_app_prof", "TENANT"),
			testAccCheckOpenApiNsxtAppPortDestroy("custom_app_prof", "SYSTEM"),
			testAccCheckOpenApiNsxtAppPortDestroy("custom_app_prof", "SYSTEM"),
		),
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_nsxt_app_port_profile.custom", "id"),
					resource.TestCheckResourceAttr("vcd_nsxt_app_port_profile.custom", "name", "custom_app_prof"),
					resource.TestCheckResourceAttr("vcd_nsxt_app_port_profile.custom", "description", "Application port profile for custom"),
					resource.TestCheckResourceAttr("vcd_nsxt_app_port_profile.custom", "scope", "PROVIDER"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_app_port_profile.custom", "app_port.*", map[string]string{
						"protocol": "ICMPv4",
					}),
				),
			},
			resource.TestStep{
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_nsxt_app_port_profile.custom", "id"),
					resource.TestCheckResourceAttr("vcd_nsxt_app_port_profile.custom", "org", "System"),
					resource.TestCheckResourceAttr("vcd_nsxt_app_port_profile.custom", "name", "custom_app_prof"),
					resource.TestCheckResourceAttr("vcd_nsxt_app_port_profile.custom", "description", "Application port profile for custom"),
					resource.TestCheckResourceAttr("vcd_nsxt_app_port_profile.custom", "scope", "PROVIDER"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_app_port_profile.custom", "app_port.*", map[string]string{
						"protocol": "ICMPv4",
					}),
					resource.TestCheckResourceAttrPair("vcd_nsxt_app_port_profile.custom", "id", "data.vcd_nsxt_app_port_profile.custom", "id"),
					// GET does not return nsxt_manager_id in the object therefore it cannot be set during read
					resourceFieldsEqual("vcd_nsxt_app_port_profile.custom", "data.vcd_nsxt_app_port_profile.custom", []string{"nsxt_manager_id"}),
				),
			},
			resource.TestStep{
				Config: configText3,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_nsxt_app_port_profile.custom", "id"),
					resource.TestCheckResourceAttr("vcd_nsxt_app_port_profile.custom", "org", "System"),
					resource.TestCheckResourceAttr("vcd_nsxt_app_port_profile.custom", "name", "custom_app_prof-updated"),
					resource.TestCheckResourceAttr("vcd_nsxt_app_port_profile.custom", "description", "Application port profile for custom-updated"),
					resource.TestCheckResourceAttr("vcd_nsxt_app_port_profile.custom", "scope", "PROVIDER"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_app_port_profile.custom", "app_port.*", map[string]string{
						"protocol": "ICMPv6",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_app_port_profile.custom", "app_port.*", map[string]string{
						"protocol": "TCP",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_app_port_profile.custom", "app_port.*", map[string]string{
						"protocol": "UDP",
					}),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_app_port_profile.custom", "app_port.*.port.*", "2000"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_app_port_profile.custom", "app_port.*.port.*", "2010-2020"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_app_port_profile.custom", "app_port.*.port.*", "12345"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_app_port_profile.custom", "app_port.*.port.*", "65000"),
					resource.TestCheckTypeSetElemAttr("vcd_nsxt_app_port_profile.custom", "app_port.*.port.*", "40000-60000"),
				),
			},
			resource.TestStep{
				Config: configText4,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vcd_nsxt_app_port_profile.custom", "id"),
					resource.TestCheckResourceAttr("vcd_nsxt_app_port_profile.custom", "org", "System"),
					resource.TestCheckResourceAttr("vcd_nsxt_app_port_profile.custom", "name", "custom_app_prof-updated"),
					resource.TestCheckResourceAttr("vcd_nsxt_app_port_profile.custom", "description", ""),
					resource.TestCheckResourceAttr("vcd_nsxt_app_port_profile.custom", "scope", "PROVIDER"),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_nsxt_app_port_profile.custom", "app_port.*", map[string]string{
						"protocol": "ICMPv6",
					}),
				),
			},
			resource.TestStep{
				ResourceName:      "vcd_nsxt_app_port_profile.custom",
				ImportState:       true,
				ImportStateVerify: true,
				// This will generate import name of org_name.vdc_name.app_profile_name
				ImportStateIdFunc: importStateIdNsxtManagerObject(testConfig, "custom_app_prof-updated"),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdNsxtAppPortProfileTenantStep1 = `
resource "vcd_nsxt_app_port_profile" "custom" {
  org  = "{{.Org}}"
  vdc  = "{{.NsxtVdc}}"
  name = "custom_app_prof"

  description = "Application port profile for custom"
  scope       = "TENANT"

  app_port {
    protocol = "ICMPv4"
  }
}
`

const testAccVcdNsxtAppPortProfileTenantStep1Ds = testAccVcdNsxtAppPortProfileTenantStep1 + `
# skip-binary-test: data source test only works in acceptance tests
data "vcd_nsxt_app_port_profile" "custom" {
  org   = "{{.Org}}"
  vdc   = "{{.NsxtVdc}}"
  name  = "custom_app_prof"
  scope = "TENANT"
}
`

const testAccVcdNsxtAppPortProfileTenantStep3 = `
resource "vcd_nsxt_app_port_profile" "custom" {
  org  = "{{.Org}}"
  vdc  = "{{.NsxtVdc}}"
  name = "custom_app_prof-updated"

  description = "Application port profile for custom-updated"
  scope       = "TENANT"

  app_port {
    protocol = "ICMPv6"
  }

  app_port {
    protocol = "TCP"
    port     = ["2000", "2010-2020", "12345", "65000"]
  }

  app_port {
    protocol = "UDP"
    port     = ["40000-60000"]
  }
}
`

const testAccVcdNsxtAppPortProfileTenantStep4 = `
resource "vcd_nsxt_app_port_profile" "custom" {
  org  = "{{.Org}}"
  vdc  = "{{.NsxtVdc}}"
  name = "custom_app_prof-updated"

  scope = "TENANT"

  app_port {
    protocol = "ICMPv6"
  }
}
`

const testAccVcdNsxtAppPortProfileProviderNsxtManagerDS = `
data "vcd_nsxt_manager" "main" {
  name = "{{.NsxtManager}}"
}
`

const testAccVcdNsxtAppPortProfileProviderStep1 = testAccVcdNsxtAppPortProfileProviderNsxtManagerDS + `
resource "vcd_nsxt_app_port_profile" "custom" {
  org  = "System"
  name = "custom_app_prof"

  description     = "Application port profile for custom"
  scope           = "PROVIDER"
  nsxt_manager_id = data.vcd_nsxt_manager.main.id

  app_port {
    protocol = "ICMPv4"
  }
}
`

const testAccVcdNsxtAppPortProfileProviderStep1AndDS = testAccVcdNsxtAppPortProfileProviderStep1 + `
# skip-binary-test: data source test only works in acceptance tests
data "vcd_nsxt_app_port_profile" "custom" {
  org   = "System"
  name  = "custom_app_prof"
  scope = "PROVIDER"
}
`

const testAccVcdNsxtAppPortProfileProviderStep2 = testAccVcdNsxtAppPortProfileProviderNsxtManagerDS + `
resource "vcd_nsxt_app_port_profile" "custom" {
  org  = "System"
  name = "custom_app_prof-updated"

  description     = "Application port profile for custom-updated"
  scope           = "PROVIDER"
  nsxt_manager_id = data.vcd_nsxt_manager.main.id

  app_port {
    protocol = "ICMPv6"
  }

  app_port {
    protocol = "TCP"
    port     = ["2000", "2010-2020", "12345", "65000"]
  }

  app_port {
    protocol = "UDP"
    port     = ["40000-60000"]
  }
}
`

const testAccVcdNsxtAppPortProfileProviderStep4 = testAccVcdNsxtAppPortProfileProviderNsxtManagerDS + `
resource "vcd_nsxt_app_port_profile" "custom" {
  org  = "System"
  name = "custom_app_prof-updated"

  scope           = "PROVIDER"
  nsxt_manager_id = data.vcd_nsxt_manager.main.id

  app_port {
    protocol = "ICMPv6"
  }
}
`
