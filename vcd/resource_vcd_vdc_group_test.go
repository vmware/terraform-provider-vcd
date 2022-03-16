//go:build vdcGroup || ALL || functional
// +build vdcGroup ALL functional

package vcd

import (
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// TestAccVcdVdcGroupResource tests that VDC group can be managed
func TestAccVcdVdcGroupResource(t *testing.T) {
	preTestChecks(t)

	// This test requires access to the vCD before filling templates
	// Thus it won't run in the short test
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	vcdClient := createTemporaryVCDConnection(false)

	if !vcdClient.Client.IsSysAdmin {
		t.Skip(t.Name() + " only System Administrator can run test of VDC group")
	}

	if testConfig.Nsxt.Vdc == "" || testConfig.VCD.NsxtProviderVdc.Name == "" ||
		testConfig.VCD.NsxtProviderVdc.NetworkPool == "" || testConfig.VCD.ProviderVdc.StorageProfile == "" {
		t.Skip("Variables Nsxt.Vdc, VCD.NsxtProviderVdc.NetworkPool, VCD.NsxtProviderVdc.Name," +
			" VCD.ProviderVdc.StorageProfile  must be set")
	}

	// String map to fill the template
	var params = StringMap{
		"Org":                       testConfig.VCD.Org,
		"VDC":                       testConfig.Nsxt.Vdc,
		"Name":                      "TestAccVcdVdcGroupResource",
		"NameUpdated":               "TestAccVcdVdcGroupResourceUpdated",
		"Description":               "myDescription",
		"DescriptionUpdate":         "myDescription updated",
		"ProviderVdc":               testConfig.VCD.NsxtProviderVdc.Name,
		"NetworkPool":               testConfig.VCD.NsxtProviderVdc.NetworkPool,
		"Allocated":                 "1024",
		"Limit":                     "1024",
		"ProviderVdcStorageProfile": testConfig.VCD.ProviderVdc.StorageProfile,
		"OrgUser":                   testConfig.TestEnvBuild.OrgUser,
		"OrgUserPassword":           testConfig.TestEnvBuild.OrgUserPassword,
		"VcdUrl":                    testConfig.Provider.Url,
		"OrgUserProvider":           "",
		"Dfw":                       "false",
		"DefaultPolicy":             "false",
		"DfwUpdated":                "true",
		"DefaultPolicyUpdated":      "true",
		"DfwUpdated2":               "false",
		"DefaultPolicyUpdated2":     "false",
		"DfwUpdated3":               "true",
		"DefaultPolicyUpdated3":     "false",
		"DfwUpdated4":               "false",
		"DefaultPolicyUpdated4":     "true",
		"DfwUpdated5":               "true",
		"DefaultPolicyUpdated5":     "true",
		"Tags":                      "vdc vdcGroup",
	}

	runVdcGroupTest(t, params)
}

// TestAccVcdVdcGroupResourceAsOrgUser tests that VDC group can be managed by Org user
func TestAccVcdVdcGroupResourceAsOrgUser(t *testing.T) {
	preTestChecks(t)

	vcdClient := createTemporaryVCDConnection(true)
	if vcdClient == nil {
		t.Skip(t.Name() + " requires a connection to set the tests")
	}

	if !vcdClient.Client.IsSysAdmin {
		t.Skip(t.Name() + " only System Administrator can run test of VDC group")
	}

	if testConfig.Nsxt.Vdc == "" || testConfig.VCD.NsxtProviderVdc.Name == "" ||
		testConfig.VCD.NsxtProviderVdc.NetworkPool == "" || testConfig.VCD.ProviderVdc.StorageProfile == "" {
		t.Skip("Variables Nsxt.Vdc, VCD.NsxtProviderVdc.NetworkPool, VCD.NsxtProviderVdc.Name," +
			" VCD.ProviderVdc.StorageProfile  must be set")
	}

	// Check if needed rights are configured
	checkRights(t, vcdClient)

	// String map to fill the template
	var params = StringMap{
		"Org":                       testConfig.VCD.Org,
		"VDC":                       testConfig.Nsxt.Vdc,
		"Name":                      "TestAccVcdVdcGroupResource",
		"NameUpdated":               "TestAccVcdVdcGroupResourceUpdated",
		"Description":               "myDescription",
		"DescriptionUpdate":         "myDescription updated",
		"ProviderVdc":               testConfig.VCD.NsxtProviderVdc.Name,
		"NetworkPool":               testConfig.VCD.NsxtProviderVdc.NetworkPool,
		"Allocated":                 "1024",
		"Limit":                     "1024",
		"ProviderVdcStorageProfile": testConfig.VCD.ProviderVdc.StorageProfile,
		"OrgUser":                   testConfig.TestEnvBuild.OrgUser,
		"OrgUserPassword":           testConfig.TestEnvBuild.OrgUserPassword,
		"VcdUrl":                    testConfig.Provider.Url,
		"OrgUserProvider":           "provider = vcd.orguser",
		"Dfw":                       "false",
		"DefaultPolicy":             "false",
		"DfwUpdated":                "true",
		"DefaultPolicyUpdated":      "true",
		"DfwUpdated2":               "false",
		"DefaultPolicyUpdated2":     "false",
		"DfwUpdated3":               "true",
		"DefaultPolicyUpdated3":     "false",
		"DfwUpdated4":               "false",
		"DefaultPolicyUpdated4":     "true",
		"DfwUpdated5":               "true",
		"DefaultPolicyUpdated5":     "true",
		"Tags":                      "vdc vdcGroup",
	}

	// run as Org user
	runVdcGroupTest(t, params)
}

func checkRights(t *testing.T, vcdClient *VCDClient) {
	var missingRights []string
	defaultRightsBundle, err := vcdClient.Client.GetRightsBundleByName("Default Rights Bundle")
	if err != nil {
		t.Errorf("%s error fetch default rights bundle: %s", t.Name(), err)
	}

	rightsBeforeChange, err := defaultRightsBundle.GetRights(nil)
	if err != nil {
		t.Errorf("%s error fetching rights: %s", t.Name(), err)
	}

	for _, rightName := range []string{
		"vDC Group: Configure",
		"vDC Group: Configure Logging",
		"vDC Group: View",
		"Organization vDC Distributed Firewall: Enable/Disable",
		//"Security Tag Edit", 10.2 doesn't have it and for this kind testing not needed
	} {
		newRight, err := vcdClient.Client.GetRightByName(rightName)
		if err != nil {
			t.Errorf("%s error fetching right %s: %s", t.Name(), rightName, err)
		}
		foundRight := false
		for _, old := range rightsBeforeChange {
			if old.Name == rightName {
				foundRight = true
			}
		}
		if !foundRight {
			missingRights = append(missingRights, newRight.Name)
		}
	}

	if len(missingRights) > 0 {
		t.Skip(t.Name() + "missing rights to run test: " + strings.Join(missingRights, ", "))
	}

	orgAdminGlobalRole, err := vcdClient.Client.GetGlobalRoleByName("Organization Administrator")
	if err != nil {
		t.Errorf("%s error fetching global role Org Administrator: %s", t.Name(), err)
	}

	globalRoleRights, err := orgAdminGlobalRole.GetRights(nil)
	if err != nil {
		t.Errorf("%s error fetching rights: %s", t.Name(), err)
	}

	rightName := "Organization vDC Distributed Firewall: Enable/Disable"
	foundRight := false
	for _, globalRoleRight := range globalRoleRights {
		if globalRoleRight.Name == rightName {
			foundRight = true
		}
	}

	if !foundRight {
		t.Skip(t.Name() + "missing rights to run test:" + rightName)
	}
}

func runVdcGroupTest(t *testing.T, params StringMap) {

	params["FuncName"] = t.Name() + "-newVdc"
	configTextPre := templateFill(testAccVcdVdcGroupNewVdc, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 1: %s", configTextPre)

	params["FuncName"] = t.Name()
	configText1 := templateFill(testAccVcdVdcGroupResource, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 2: %s", configText1)

	params["FuncName"] = t.Name() + "-update"
	configText2 := templateFill(testAccVcdVdcGroupResourceUpdate, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 3: %s", configText2)

	params["FuncName"] = t.Name() + "-update2"
	configText3 := templateFill(testAccVcdVdcGroupResourceUpdate2, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 4: %s", configText3)

	params["FuncName"] = t.Name() + "-update3"
	configText4 := templateFill(testAccVcdVdcGroupResourceUpdate3, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 5: %s", configText4)

	params["FuncName"] = t.Name() + "-update4"
	configText5 := templateFill(testAccVcdVdcGroupResourceUpdate4, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 6: %s", configText5)

	params["FuncName"] = t.Name() + "-update5"
	configText6 := templateFill(testAccVcdVdcGroupResourceUpdate5, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 7: %s", configText6)

	params["FuncName"] = t.Name() + "-datasource"
	configText7 := templateFill(testAccVcdVdcGroupDatasource, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 8: %s", configText7)

	params["FuncName"] = t.Name() + "-provider"
	configTextProvider := templateFill(testAccVcdVdcGroupOrgProvider, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 9: %s", configTextProvider)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resourceAddressVdcGroup := "vcd_vdc_group.fromUnitTest"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		PreCheck:          func() { testAccPreCheck(t) },

		Steps: []resource.TestStep{
			// initialize new VDC, this done separately as otherwise randomly fail due choose wrong connection
			resource.TestStep{
				Config: configTextPre,
			},
			resource.TestStep{
				Config: configText1,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddressVdcGroup, "name", params["Name"].(string)),
					resource.TestMatchResourceAttr(resourceAddressVdcGroup, "id", regexp.MustCompile(`^\S+`)),
					resource.TestCheckResourceAttr(resourceAddressVdcGroup, "description", params["Description"].(string)),
					resource.TestMatchResourceAttr(resourceAddressVdcGroup, "starting_vdc_id", regexp.MustCompile(`^\S+`)),
					resource.TestCheckOutput("participatingVdcCount", "2"),
					resource.TestCheckResourceAttr(resourceAddressVdcGroup, "dfw_enabled", params["Dfw"].(string)),
					resource.TestCheckResourceAttr(resourceAddressVdcGroup, "default_policy_status", params["DefaultPolicy"].(string)),
				),
			},
			resource.TestStep{
				Config: configText2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddressVdcGroup, "name", params["NameUpdated"].(string)),
					resource.TestMatchResourceAttr(resourceAddressVdcGroup, "id", regexp.MustCompile(`^\S+`)),
					resource.TestCheckResourceAttr(resourceAddressVdcGroup, "description", params["DescriptionUpdate"].(string)),
					resource.TestMatchResourceAttr(resourceAddressVdcGroup, "starting_vdc_id", regexp.MustCompile(`^\S+`)),
					resource.TestCheckOutput("participatingVdcCount", "1"),
					resource.TestCheckResourceAttr(resourceAddressVdcGroup, "dfw_enabled", params["DfwUpdated"].(string)),
					resource.TestCheckResourceAttr(resourceAddressVdcGroup, "default_policy_status", params["DefaultPolicyUpdated"].(string)),
				),
			},
			resource.TestStep{
				Config: configText3,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddressVdcGroup, "name", params["NameUpdated"].(string)),
					resource.TestMatchResourceAttr(resourceAddressVdcGroup, "id", regexp.MustCompile(`^\S+`)),
					resource.TestCheckResourceAttr(resourceAddressVdcGroup, "description", params["DescriptionUpdate"].(string)),
					resource.TestMatchResourceAttr(resourceAddressVdcGroup, "starting_vdc_id", regexp.MustCompile(`^\S+`)),
					resource.TestCheckOutput("participatingVdcCount", "1"),
					resource.TestCheckResourceAttr(resourceAddressVdcGroup, "dfw_enabled", params["DfwUpdated2"].(string)),
					resource.TestCheckResourceAttr(resourceAddressVdcGroup, "default_policy_status", params["DefaultPolicyUpdated2"].(string)),
				),
			},
			resource.TestStep{
				Config: configText4,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddressVdcGroup, "name", params["NameUpdated"].(string)),
					resource.TestMatchResourceAttr(resourceAddressVdcGroup, "id", regexp.MustCompile(`^\S+`)),
					resource.TestCheckResourceAttr(resourceAddressVdcGroup, "description", params["DescriptionUpdate"].(string)),
					resource.TestMatchResourceAttr(resourceAddressVdcGroup, "starting_vdc_id", regexp.MustCompile(`^\S+`)),
					resource.TestCheckOutput("participatingVdcCount", "1"),
					resource.TestCheckResourceAttr(resourceAddressVdcGroup, "dfw_enabled", params["DfwUpdated3"].(string)),
					resource.TestCheckResourceAttr(resourceAddressVdcGroup, "default_policy_status", params["DefaultPolicyUpdated3"].(string)),
				),
			},
			resource.TestStep{
				Config:      configText5,
				ExpectError: regexp.MustCompile("`default_policy_status` must be `false` when `dfw_enabled` is `false`."),
			},
			resource.TestStep{
				Config: configText6,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceAddressVdcGroup, "name", params["NameUpdated"].(string)),
					resource.TestMatchResourceAttr(resourceAddressVdcGroup, "id", regexp.MustCompile(`^\S+`)),
					resource.TestCheckResourceAttr(resourceAddressVdcGroup, "description", params["DescriptionUpdate"].(string)),
					resource.TestMatchResourceAttr(resourceAddressVdcGroup, "starting_vdc_id", regexp.MustCompile(`^\S+`)),
					resource.TestCheckOutput("participatingVdcCount", "1"),
					resource.TestCheckResourceAttr(resourceAddressVdcGroup, "dfw_enabled", params["DfwUpdated5"].(string)),
					resource.TestCheckResourceAttr(resourceAddressVdcGroup, "default_policy_status", params["DefaultPolicyUpdated5"].(string)),
				),
			},
			resource.TestStep{
				Config: configText7,
				Check: resource.ComposeAggregateTestCheckFunc(
					resourceFieldsEqual(resourceAddressVdcGroup, "data.vcd_vdc_group.fetchCreated", []string{"participating_vdc_ids.#",
						"starting_vdc_id", "%", "participating_vdc_ids.0", "default_policy_status"}),
				),
			},
			resource.TestStep{
				ResourceName:            resourceAddressVdcGroup,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdFunc:       importStateIdOrgObject(testConfig, params["NameUpdated"].(string)),
				ImportStateVerifyIgnore: []string{"starting_vdc_id"},
			},
			// for clean destroy, otherwise randomly fail due choose wrong connection
			resource.TestStep{
				Config: configTextProvider,
			},
			// for clean destroy, otherwise randomly fail due choose wrong connection
			resource.TestStep{
				Config: configTextPre,
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdVdcGroupNewVdc = `
resource "vcd_org_vdc" "newVdc" {
  provider = vcd

  name = "newVdc"
  org  = "{{.Org}}"

  allocation_model  = "Flex"
  network_pool_name = "{{.NetworkPool}}"
  provider_vdc_name = "{{.ProviderVdc}}"

  compute_capacity {
    cpu {
      allocated = "{{.Allocated}}"
      limit     = "{{.Limit}}"
    }

    memory {
      allocated = "{{.Allocated}}"
      limit     = "{{.Limit}}"
    }
  }

  storage_profile {
    name    = "{{.ProviderVdcStorageProfile}}"
    enabled = true
    limit   = 10240
    default = true
  }

  metadata = {
    vdc_metadata = "VDC Metadata"
  }

  enabled                    = true
  enable_thin_provisioning   = true
  enable_fast_provisioning   = true
  delete_force               = true
  delete_recursive           = true
  elasticity      			 = true
  include_vm_memory_overhead = true
}
`
const testAccVcdVdcGroupOrgProvider = testAccVcdVdcGroupNewVdc + `
provider "vcd" {
  alias                = "orguser"
  user                 = "{{.OrgUser}}"
  password             = "{{.OrgUserPassword}}"
  auth_type            = "integrated"
  url                  = "{{.VcdUrl}}"
  sysorg               = "{{.Org}}"
  org                  = "{{.Org}}"
  vdc                  = "{{.VDC}}"
  allow_unverified_ssl = "true"
  max_retry_timeout    = 600
  logging              = true
  logging_file         = "go-vcloud-director-org.log"
}
`

const testAccVcdVdcGroupResource = testAccVcdVdcGroupOrgProvider + `
data "vcd_org_vdc" "startVdc"{
  {{if .OrgUserProvider}}{{.OrgUserProvider}}{{end}}

  org  = "{{.Org}}"
  name = "{{.VDC}}"
}

resource "vcd_vdc_group" "fromUnitTest" {
  {{if .OrgUserProvider}}{{.OrgUserProvider}}{{end}}

  org                   = "{{.Org}}"
  name                  = "{{.Name}}"
  description           = "{{.Description}}"
  starting_vdc_id       = data.vcd_org_vdc.startVdc.id
  participating_vdc_ids = [data.vcd_org_vdc.startVdc.id, vcd_org_vdc.newVdc.id]

  dfw_enabled           = "{{.Dfw}}"
}

output "participatingVdcCount" {
  value = length(vcd_vdc_group.fromUnitTest.participating_vdc_ids)
}
`

const testAccVcdVdcGroupResourceUpdate = testAccVcdVdcGroupOrgProvider + `
data "vcd_org_vdc" "startVdc"{
  {{if .OrgUserProvider}}{{.OrgUserProvider}}{{end}}

  org  = "{{.Org}}"
  name = "{{.VDC}}"
}

resource "vcd_vdc_group" "fromUnitTest" {
  {{if .OrgUserProvider}}{{.OrgUserProvider}}{{end}}

  org                   = "{{.Org}}"
  name                  = "{{.NameUpdated}}"
  description           = "{{.DescriptionUpdate}}"
  starting_vdc_id       = data.vcd_org_vdc.startVdc.id
  participating_vdc_ids = [data.vcd_org_vdc.startVdc.id]

  dfw_enabled           = "{{.DfwUpdated}}"
  default_policy_status = "{{.DefaultPolicyUpdated}}"
}

output "participatingVdcCount" {
  value = length(vcd_vdc_group.fromUnitTest.participating_vdc_ids)
}
`

const testAccVcdVdcGroupResourceUpdate2 = testAccVcdVdcGroupOrgProvider + `
data "vcd_org_vdc" "startVdc"{
  {{if .OrgUserProvider}}{{.OrgUserProvider}}{{end}}

  org  = "{{.Org}}"
  name = "{{.VDC}}"
}

resource "vcd_vdc_group" "fromUnitTest" {
  {{if .OrgUserProvider}}{{.OrgUserProvider}}{{end}}

  org                   = "{{.Org}}"
  name                  = "{{.NameUpdated}}"
  description           = "{{.DescriptionUpdate}}"
  starting_vdc_id       = data.vcd_org_vdc.startVdc.id
  participating_vdc_ids = [data.vcd_org_vdc.startVdc.id]

  dfw_enabled           = "{{.DfwUpdated2}}"
  default_policy_status = "{{.DefaultPolicyUpdated2}}"
}

output "participatingVdcCount" {
  value = length(vcd_vdc_group.fromUnitTest.participating_vdc_ids)
}
`

const testAccVcdVdcGroupResourceUpdate3 = testAccVcdVdcGroupOrgProvider + `
data "vcd_org_vdc" "startVdc"{
  {{if .OrgUserProvider}}{{.OrgUserProvider}}{{end}}

  org  = "{{.Org}}"
  name = "{{.VDC}}"
}

resource "vcd_vdc_group" "fromUnitTest" {
  {{if .OrgUserProvider}}{{.OrgUserProvider}}{{end}}

  org                   = "{{.Org}}"
  name                  = "{{.NameUpdated}}"
  description           = "{{.DescriptionUpdate}}"
  starting_vdc_id       = data.vcd_org_vdc.startVdc.id
  participating_vdc_ids = [data.vcd_org_vdc.startVdc.id]


  dfw_enabled           = "{{.DfwUpdated3}}"
  default_policy_status = "{{.DefaultPolicyUpdated3}}"
}

output "participatingVdcCount" {
  value = length(vcd_vdc_group.fromUnitTest.participating_vdc_ids)
}
`

const testAccVcdVdcGroupResourceUpdate4 = testAccVcdVdcGroupOrgProvider + `
data "vcd_org_vdc" "startVdc"{
  {{if .OrgUserProvider}}{{.OrgUserProvider}}{{end}}

  org  = "{{.Org}}"
  name = "{{.VDC}}"
}

# skip-binary-test: checking if error is thrown
resource "vcd_vdc_group" "fromUnitTest" {
  {{if .OrgUserProvider}}{{.OrgUserProvider}}{{end}}

  org                   = "{{.Org}}"
  name                  = "{{.NameUpdated}}"
  description           = "{{.DescriptionUpdate}}"
  starting_vdc_id       = data.vcd_org_vdc.startVdc.id
  participating_vdc_ids = [data.vcd_org_vdc.startVdc.id]


  dfw_enabled           = "{{.DfwUpdated4}}"
  default_policy_status = "{{.DefaultPolicyUpdated4}}"
}

output "participatingVdcCount" {
  value = length(vcd_vdc_group.fromUnitTest.participating_vdc_ids)
}
`

const testAccVcdVdcGroupResourceUpdate5 = testAccVcdVdcGroupOrgProvider + `
data "vcd_org_vdc" "startVdc"{
  {{if .OrgUserProvider}}{{.OrgUserProvider}}{{end}}

  org  = "{{.Org}}"
  name = "{{.VDC}}"
}

resource "vcd_vdc_group" "fromUnitTest" {
  {{if .OrgUserProvider}}{{.OrgUserProvider}}{{end}}

  org                   = "{{.Org}}"
  name                  = "{{.NameUpdated}}"
  description           = "{{.DescriptionUpdate}}"
  starting_vdc_id       = data.vcd_org_vdc.startVdc.id
  participating_vdc_ids = [data.vcd_org_vdc.startVdc.id]


  dfw_enabled           = "{{.DfwUpdated5}}"
  default_policy_status = "{{.DefaultPolicyUpdated5}}"
}

output "participatingVdcCount" {
  value = length(vcd_vdc_group.fromUnitTest.participating_vdc_ids)
}
`

const testAccVcdVdcGroupDatasource = testAccVcdVdcGroupResourceUpdate3 + `
# skip-binary-test: data source test only works in acceptance tests
data "vcd_vdc_group" "fetchCreated" {
  {{if .OrgUserProvider}}{{.OrgUserProvider}}{{end}}

  org  = "{{.Org}}"
  name = vcd_vdc_group.fromUnitTest.name
}
`
