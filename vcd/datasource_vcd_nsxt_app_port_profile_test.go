//go:build network || nsxt || ALL || functional
// +build network nsxt ALL functional

package vcd

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// TestAccVcdNsxtAppPortProfileDsSystem tests if a built-in SYSTEM scope application port profile can be read
func TestAccVcdNsxtAppPortProfileDsSystem(t *testing.T) {
	preTestChecks(t)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	skipNoNsxtConfiguration(t)

	var params = StringMap{
		"Org":         testConfig.VCD.Org,
		"NsxtVdc":     testConfig.Nsxt.Vdc,
		"ProfileName": "Active Directory Server", // Existing System built-in Application Port Profile
		"Scope":       "SYSTEM",
		"Tags":        "nsxt network",
	}

	configText1 := templateFill(testAccVcdNsxtAppPortProfileSystemDSStep1, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		PreCheck:          func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_app_port_profile.custom", "id"),
					resource.TestCheckResourceAttr("data.vcd_nsxt_app_port_profile.custom", "name", "Active Directory Server"),
					resource.TestCheckResourceAttr("data.vcd_nsxt_app_port_profile.custom", "scope", "SYSTEM"),
					resource.TestCheckTypeSetElemAttr("data.vcd_nsxt_app_port_profile.custom", "app_port.*.port.*", "464"),
					resource.TestCheckTypeSetElemNestedAttrs("data.vcd_nsxt_app_port_profile.custom", "app_port.*", map[string]string{
						"protocol": "TCP",
					}),
				),
			},
		},
	})
	postTestChecks(t)
}

// TestAccVcdNsxtAppPortProfileDsSystem tests if "Active Directory Server" Application Port Profile is not found in
// PROVIDER context (because it is defined in SYSTEM context)
func TestAccVcdNsxtAppPortProfileDsProviderNotFound(t *testing.T) {
	preTestChecks(t)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	skipNoNsxtConfiguration(t)

	var params = StringMap{
		"Org":         testConfig.VCD.Org,
		"NsxtVdc":     testConfig.Nsxt.Vdc,
		"ProfileName": "Active Directory Server",
		"Scope":       "PROVIDER",
		"Tags":        "nsxt network",
	}

	configText1 := templateFill(testAccVcdNsxtAppPortProfileSystemDSStep1, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		PreCheck:          func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config:      configText1,
				ExpectError: regexp.MustCompile(`.*ENF.*`),
			},
		},
	})
	postTestChecks(t)
}

// TestAccVcdNsxtAppPortProfileDsTenantNotFound tests if "Active Directory Server" Application Port Profile is not found in
// TENANT context (because it is defined in SYSTEM context)
func TestAccVcdNsxtAppPortProfileDsTenantNotFound(t *testing.T) {
	preTestChecks(t)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	skipNoNsxtConfiguration(t)

	var params = StringMap{
		"Org":         testConfig.VCD.Org,
		"NsxtVdc":     testConfig.Nsxt.Vdc,
		"ProfileName": "Active Directory Server",
		"Scope":       "TENANT",
		"Tags":        "nsxt network",
	}

	configText1 := templateFill(testAccVcdNsxtAppPortProfileSystemDSStep1, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		PreCheck:          func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config:      configText1,
				ExpectError: regexp.MustCompile(`.*ENF.*`),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdNsxtAppPortProfileSystemDSStep1 = `
data "vcd_nsxt_app_port_profile" "custom" {
  org  = "{{.Org}}"
  vdc  = "{{.NsxtVdc}}"

  name  = "{{.ProfileName}}"
  scope = "{{.Scope}}"
}
`

// TestAccVcdNsxtAppPortProfileMultiOrg tests that TENANT Application Port Profile lookup works well
// when multiple Orgs exist in VCD. The test does the following:
// * Step 1 - creates another Org with one NSX-T VDC. Creates 2 application port profiles - one in 1st VDC
// * Step 2 - defines 4 data sources for application port profiles in VDC 1 in Org 1 and VDC2 in Org
// 2 using both newer configuration using `context_id` and legacy configured `org` and `vdc`
// This test is done to replicate and fix https://github.com/vmware/terraform-provider-vcd/issues/778
func TestAccVcdNsxtAppPortProfileMultiOrg(t *testing.T) {
	preTestChecks(t)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	skipNoNsxtConfiguration(t)

	var params = StringMap{
		"Org":            testConfig.VCD.Org,
		"NsxtVdc":        testConfig.Nsxt.Vdc,
		"ProfileName":    "Active Directory Server",
		"Scope":          "TENANT",
		"OrgName1":       testConfig.VCD.Org,
		"OrgName2":       t.Name(),
		"VdcName":        t.Name(),
		"MetadataKey":    "k",
		"MetadataValue":  "v",
		"StorageProfile": testConfig.VCD.ProviderVdc.StorageProfile,

		"Tags": "nsxt network",
	}

	configText1 := templateFill(testAccVcdNsxtAppPortProfileMultiOrgPreCreate, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configText1)

	params["FuncName"] = t.Name() + "-step2"
	configText2 := templateFill(testAccVcdNsxtAppPortProfileMultiOrg, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText2)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		PreCheck:          func() { testAccPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: configText1,
			},
			{
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(

					// sleepTester(),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_app_port_profile.custom", "id"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_app_port_profile.custom-legacy-config", "id"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_app_port_profile.custom2", "id"),
					resource.TestCheckResourceAttrSet("data.vcd_nsxt_app_port_profile.custom-legacy-config2", "id"),

					resource.TestCheckResourceAttrPair("data.vcd_nsxt_app_port_profile.custom", "id", "data.vcd_nsxt_app_port_profile.custom-legacy-config", "id"),
					resource.TestCheckResourceAttrPair("data.vcd_nsxt_app_port_profile.custom2", "id", "data.vcd_nsxt_app_port_profile.custom-legacy-config2", "id"),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdNsxtAppPortProfileMultiOrgPreCreate = testAccCheckVcdDatasourceOrg + `
data "vcd_org_vdc" "existingVdc" {
  org  = "{{.Org}}"
  name = "{{.NsxtVdc}}"
}

resource "vcd_org_vdc" "vdc-in-org2" { 
  name = "{{.VdcName}}"
  org  = "{{.OrgName2}}"

  allocation_model  = data.vcd_org_vdc.existingVdc.allocation_model
  network_pool_name = data.vcd_org_vdc.existingVdc.network_pool_name
  provider_vdc_name = data.vcd_org_vdc.existingVdc.provider_vdc_name

  compute_capacity {
    cpu {
     allocated = data.vcd_org_vdc.existingVdc.compute_capacity[0].cpu[0].allocated
     limit     = data.vcd_org_vdc.existingVdc.compute_capacity[0].cpu[0].limit
    }

    memory {
     allocated = data.vcd_org_vdc.existingVdc.compute_capacity[0].memory[0].allocated
     limit     = data.vcd_org_vdc.existingVdc.compute_capacity[0].memory[0].limit
    }
  }

  storage_profile {
    name    = "{{.StorageProfile}}"
    enabled = true
    limit   = 0
    default = true
  }

  metadata = {
    vdc_metadata = "VDC Metadata"
  }

  enabled                  = data.vcd_org_vdc.existingVdc.enabled
  enable_thin_provisioning = data.vcd_org_vdc.existingVdc.enable_thin_provisioning
  enable_fast_provisioning = data.vcd_org_vdc.existingVdc.enable_fast_provisioning
  delete_force             = true
  delete_recursive         = true

  depends_on = [vcd_org.{{.OrgName2}}]
}


data "vcd_org_vdc" "v1" {
  org  = "{{.Org}}"
  name = "{{.NsxtVdc}}"
}

resource "vcd_nsxt_app_port_profile" "custom-org1" {
  org  = "{{.Org}}"
  name = "custom_app_prof"

  context_id = data.vcd_org_vdc.v1.id

  description = "Application port profile for custom"
  scope       = "TENANT"

  app_port {
    protocol = "ICMPv4"
  }
}

resource "vcd_nsxt_app_port_profile" "custom-org2" {
  org  = vcd_org.{{.OrgName2}}.name
  name = "custom_app_prof"

  context_id = vcd_org_vdc.vdc-in-org2.id

  description = "Application port profile for custom"
  scope       = "TENANT"

  app_port {
    protocol = "ICMPv4"
  }

  depends_on = [vcd_org.{{.OrgName2}}]
}
`

const testAccVcdNsxtAppPortProfileMultiOrg = testAccVcdNsxtAppPortProfileMultiOrgPreCreate + `
# skip-binary-test: Data Source test
data "vcd_nsxt_app_port_profile" "custom" {
  org  = "{{.Org}}"
  
  context_id = data.vcd_org_vdc.v1.id

  name  = vcd_nsxt_app_port_profile.custom-org1.name
  scope = "{{.Scope}}"
}

data "vcd_nsxt_app_port_profile" "custom-legacy-config" {
  org  = "{{.Org}}"
  vdc  = "{{.NsxtVdc}}"

  name  = vcd_nsxt_app_port_profile.custom-org1.name
  scope = "{{.Scope}}"
}

data "vcd_nsxt_app_port_profile" "custom2" {
  org  = vcd_org.{{.OrgName2}}.name
	
  context_id = vcd_org_vdc.vdc-in-org2.id
  
  name  = vcd_nsxt_app_port_profile.custom-org2.name
  scope = "{{.Scope}}"
}
  
data "vcd_nsxt_app_port_profile" "custom-legacy-config2" {
  org  = vcd_org.{{.OrgName2}}.name
  vdc  = vcd_org_vdc.vdc-in-org2.name

  name  = vcd_nsxt_app_port_profile.custom-org2.name
  scope = "{{.Scope}}"
}
`
