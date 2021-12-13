//go:build vdcGroup || ALL || functional
// +build vdcGroup ALL functional

package vcd

import (
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"log"
	"regexp"
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

	vcdClient := createTemporaryVCDConnection()
	if vcdClient.Client.APIVCDMaxVersionIs("< 35.0") {
		t.Skip(t.Name() + " requires at least API v35.0 (vCD 10.2+)")
	}

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
		"SkipBinary":                "",
	}

	runVdcGroupTest(t, params)
}

// TestAccVcdVdcGroupResourceAsOrgUser tests that VDC group can be managed by Org user
func TestAccVcdVdcGroupResourceAsOrgUser(t *testing.T) {
	preTestChecks(t)

	// This test requires access to the vCD before filling templates
	// Thus it won't run in the short test
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	vcdClient := createTemporaryVCDConnection()
	if vcdClient.Client.APIVCDMaxVersionIs("< 35.0") {
		t.Skip(t.Name() + " requires at least API v35.0 (vCD 10.2+)")
	}

	if !vcdClient.Client.IsSysAdmin {
		t.Skip(t.Name() + " only System Administrator can run test of VDC group")
	}

	if testConfig.Nsxt.Vdc == "" || testConfig.VCD.NsxtProviderVdc.Name == "" ||
		testConfig.VCD.NsxtProviderVdc.NetworkPool == "" || testConfig.VCD.ProviderVdc.StorageProfile == "" {
		t.Skip("Variables Nsxt.Vdc, VCD.NsxtProviderVdc.NetworkPool, VCD.NsxtProviderVdc.Name," +
			" VCD.ProviderVdc.StorageProfile  must be set")
	}

	// Add needed right for Org User
	rightsToAdd := addRights(t, vcdClient)

	//remove added rights at the end of test
	defer cleanupRightsAndBundle(t, vcdClient, rightsToAdd)

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
		"SkipBinary":                "# skip-binary-test: in binary user rights aren't changed to be correct",
	}

	// run as Org user
	runVdcGroupTest(t, params)
}

func addRights(t *testing.T, vcdClient *VCDClient) []types.OpenApiReference {
	// add needed rights
	var rightsToAdd []types.OpenApiReference
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
		if foundRight {
			log.Printf("Right %s already in Default Rights Bundle\n", rightName)
			// ignore
		} else {
			rightsToAdd = append(rightsToAdd, types.OpenApiReference{Name: newRight.Name, ID: newRight.ID})
		}
	}

	if len(rightsToAdd) > 0 {
		err = defaultRightsBundle.AddRights(rightsToAdd)
		if err != nil {
			t.Errorf("%s error adding rights %s: %s", t.Name(), rightsToAdd, err)
		}
		err = defaultRightsBundle.PublishAllTenants()
		if err != nil {
			t.Errorf("%s error publishing to tenants: %s", t.Name(), err)
		}
	}

	orgAdminGlobalRole, err := vcdClient.Client.GetGlobalRoleByName("Organization Administrator")
	if err != nil {
		t.Errorf("%s error fetching global role Org Administrator: %s", t.Name(), err)
	}
	missingRight, err := vcdClient.Client.GetRightByName("Organization vDC Distributed Firewall: Enable/Disable")
	if err != nil {
		t.Errorf("%s error fetching right: %s", t.Name(), err)
	}
	err = orgAdminGlobalRole.AddRights([]types.OpenApiReference{{Name: missingRight.Name, ID: missingRight.ID}})
	if err != nil {
		t.Errorf("%s error adding right: %s", t.Name(), err)
	}
	err = orgAdminGlobalRole.PublishAllTenants()
	if err != nil {
		t.Errorf("%s error publishing to tenants: %s", t.Name(), err)
	}

	return rightsToAdd
}

func cleanupRightsAndBundle(t *testing.T, vcdClient *VCDClient, rightsToRemove []types.OpenApiReference) {
	if len(rightsToRemove) > 0 {
		defaultRightsBundle, err := vcdClient.Client.GetRightsBundleByName("Default Rights Bundle")
		if err != nil {
			t.Errorf("%s error fetch default rights bundle: %s", t.Name(), err)
		}
		err = defaultRightsBundle.RemoveRights(rightsToRemove)
		if err != nil {
			t.Errorf("%s error removing rights %s: %s", t.Name(), rightsToRemove, err)
		}
		err = defaultRightsBundle.PublishAllTenants()
		if err != nil {
			t.Errorf("%s error unpublishing to tenants: %s", t.Name(), err)
		}
	}

	orgAdminGlobalRole, err := vcdClient.Client.GetGlobalRoleByName("Organization Administrator")
	if err != nil {
		t.Errorf("%s error fetching global role Org Administrator: %s", t.Name(), err)
	}
	missingRight, err := vcdClient.Client.GetRightByName("Organization vDC Distributed Firewall: Enable/Disable")
	if err != nil {
		t.Errorf("%s error fetching right: %s", t.Name(), err)
	}
	err = orgAdminGlobalRole.RemoveRights([]types.OpenApiReference{{Name: missingRight.Name, ID: missingRight.ID}})
	if err != nil {
		t.Errorf("%s error adding right: %s", t.Name(), err)
	}
	err = orgAdminGlobalRole.PublishAllTenants()
	if err != nil {
		t.Errorf("%s error publishing to tenants: %s", t.Name(), err)
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

	params["FuncName"] = t.Name() + "-datasource"
	configText5 := templateFill(testAccVcdVdcGroupDatasource, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 6: %s", configText5)

	params["FuncName"] = t.Name() + "-provider"
	configTextProvider := templateFill(testAccVcdVdcGroupOrgProvider, params)
	debugPrintf("#[DEBUG] CONFIGURATION for step 7: %s", configTextProvider)

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
				Config: configText5,
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
{{if .SkipBinary}}{{.SkipBinary}}{{end}}
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
{{if .SkipBinary}}{{.SkipBinary}}{{end}}
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
{{if .SkipBinary}}{{.SkipBinary}}{{end}}
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
{{if .SkipBinary}}{{.SkipBinary}}{{end}}
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
{{if .SkipBinary}}{{.SkipBinary}}{{end}}
data "vcd_org_vdc" "startVdc"{
  {{if .OrgUserProvider}}{{.OrgUserProvider}}{{end}}

  org  = "{{.Org}}"
  name = "{{.VDC}}"
}

{{if .SkipBinary}}{{.SkipBinary}}{{end}}
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

{{if .SkipBinary}}{{.SkipBinary}}{{end}}
output "participatingVdcCount" {
  value = length(vcd_vdc_group.fromUnitTest.participating_vdc_ids)
}
`

const testAccVcdVdcGroupResourceUpdate3 = testAccVcdVdcGroupOrgProvider + `
{{if .SkipBinary}}{{.SkipBinary}}{{end}}
data "vcd_org_vdc" "startVdc"{
  {{if .OrgUserProvider}}{{.OrgUserProvider}}{{end}}

  org  = "{{.Org}}"
  name = "{{.VDC}}"
}

{{if .SkipBinary}}{{.SkipBinary}}{{end}}
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

{{if .SkipBinary}}{{.SkipBinary}}{{end}}
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
