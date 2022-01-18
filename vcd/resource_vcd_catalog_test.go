//go:build catalog || ALL || functional
// +build catalog ALL functional

package vcd

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"
	"time"

	"github.com/vmware/go-vcloud-director/v2/govcd"
	"github.com/vmware/go-vcloud-director/v2/types/v56"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	testingTags["catalog"] = "resource_vcd_catalog_test.go"
}

var TestAccVcdCatalogName = "TestAccVcdCatalog"
var TestAccVcdCatalogDescription = "TestAccVcdCatalogBasicDescription"

func TestAccVcdCatalog(t *testing.T) {
	preTestChecks(t)
	var params = StringMap{
		"Org":            testConfig.VCD.Org,
		"CatalogName":    TestAccVcdCatalogName,
		"Description":    TestAccVcdCatalogDescription,
		"StorageProfile": testConfig.VCD.ProviderVdc.StorageProfile,
		"Tags":           "catalog",
	}

	configText := templateFill(testAccCheckVcdCatalog, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

	params["FuncName"] = t.Name() + "step1"
	params["Description"] = "TestAccVcdCatalogBasicDescription-description"
	configText1 := templateFill(testAccCheckVcdCatalogStep1, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText1)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resourceAddress := "vcd_catalog.test-catalog"
	// Use field value caching function across multiple test steps to ensure object wasn't recreated (ID did not change)
	cachedId := &testCachedFieldValue{}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckCatalogDestroy,
		Steps: []resource.TestStep{
			// Provision catalog without storage profile
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					cachedId.cacheTestResourceFieldValue(resourceAddress, "id"),
					resource.TestCheckResourceAttr(resourceAddress, "name", TestAccVcdCatalogName),
					resource.TestCheckResourceAttr(resourceAddress, "description", TestAccVcdCatalogDescription),
					resource.TestCheckResourceAttr(resourceAddress, "storage_profile_id", ""),
					testAccCheckVcdCatalogExists(resourceAddress),
				),
			},
			// Set storage profile for existing catalog
			resource.TestStep{
				Config: configText1,
				Check: resource.ComposeTestCheckFunc(
					cachedId.testCheckCachedResourceFieldValue(resourceAddress, "id"),
					resource.TestCheckResourceAttr(resourceAddress, "name", TestAccVcdCatalogName),
					resource.TestCheckResourceAttr(resourceAddress, "description", "TestAccVcdCatalogBasicDescription-description"),
					resource.TestMatchResourceAttr(resourceAddress, "storage_profile_id",
						regexp.MustCompile(`^urn:vcloud:vdcstorageProfile:`)),
					testAccCheckVcdCatalogExists(resourceAddress),
				),
			},
			// Remove storage profile just like it was provisioned in step 0
			resource.TestStep{

				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					cachedId.testCheckCachedResourceFieldValue(resourceAddress, "id"),
					resource.TestCheckResourceAttr(resourceAddress, "name", TestAccVcdCatalogName),
					resource.TestCheckResourceAttr(resourceAddress, "description", TestAccVcdCatalogDescription),
					resource.TestCheckResourceAttr(resourceAddress, "storage_profile_id", ""),
					testAccCheckVcdCatalogExists(resourceAddress),
				),
			},
			resource.TestStep{
				ResourceName:      resourceAddress,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdOrgObject(testConfig, TestAccVcdCatalogName),
				// These fields can't be retrieved from catalog data
				ImportStateVerifyIgnore: []string{"delete_force", "delete_recursive"},
			},
		},
	})
	postTestChecks(t)
}

// TestAccVcdCatalogWithStorageProfile is very similar to TestAccVcdCatalog, but it ensure that a catalog can be created
// using specific storage profile
func TestAccVcdCatalogWithStorageProfile(t *testing.T) {
	preTestChecks(t)
	var params = StringMap{
		"Org":            testConfig.VCD.Org,
		"Vdc":            testConfig.VCD.Vdc,
		"CatalogName":    TestAccVcdCatalogName,
		"Description":    TestAccVcdCatalogDescription,
		"StorageProfile": testConfig.VCD.ProviderVdc.StorageProfile,
		"Tags":           "catalog",
	}

	configText := templateFill(testAccCheckVcdCatalogStep1, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resourceAddress := "vcd_catalog.test-catalog"
	dataSourceAddress := "data.vcd_storage_profile.sp"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckCatalogDestroy,
		Steps: []resource.TestStep{
			// Provision with storage profile
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdCatalogExists(resourceAddress),
					resource.TestCheckResourceAttr(resourceAddress, "name", TestAccVcdCatalogName),
					resource.TestCheckResourceAttr(resourceAddress, "description", TestAccVcdCatalogDescription),
					resource.TestMatchResourceAttr(resourceAddress, "storage_profile_id",
						regexp.MustCompile(`^urn:vcloud:vdcstorageProfile:`)),
					resource.TestCheckResourceAttrPair(resourceAddress, "storage_profile_id", dataSourceAddress, "id"),
					checkStorageProfileOriginatesInParentVdc(dataSourceAddress,
						params["StorageProfile"].(string),
						params["Org"].(string),
						params["Vdc"].(string)),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccCheckVcdCatalogPublished = `
resource "vcd_catalog" "test-catalog" {
  org = "{{.Org}}" 
  
  name        = "{{.CatalogName}}"
  description = "{{.Description}}"

  delete_force      = "true"
  delete_recursive  = "true"

  publish_enabled               = "{{.PublishEnabled}}"
  cache_enabled                 = "{{.CacheEnabled}}"
  preserve_identity_information = "{{.PreserveIdentityInformation}}"
  password                      = "superUnknown"
}
`

const testAccCheckVcdCatalogPublishedUpdate1 = `
resource "vcd_catalog" "test-catalog" {
  org = "{{.Org}}" 
  
  name        = "{{.CatalogName}}"
  description = "{{.Description}}"

  delete_force      = "true"
  delete_recursive  = "true"

  publish_enabled               = "{{.PublishEnabledUpdate1}}"
  cache_enabled                 = "{{.CacheEnabledUpdate1}}"
  preserve_identity_information = "{{.PreserveIdentityInformationUpdate1}}"
  password                      = "superUnknown"
}
`

const testAccCheckVcdCatalogPublishedUpdate2 = `
resource "vcd_catalog" "test-catalog" {
  org = "{{.Org}}" 
  
  name        = "{{.CatalogName}}"
  description = "{{.Description}}"

  delete_force      = "true"
  delete_recursive  = "true"

  publish_enabled               = "{{.PublishEnabledUpdate2}}"
  cache_enabled                 = "{{.CacheEnabledUpdate2}}"
  preserve_identity_information = "{{.PreserveIdentityInformationUpdate2}}"
  password                      = "superUnknown"
}
`

// TestAccVcdCatalogPublishedToExternalOrg is very similar to TestAccVcdCatalog, but it ensures that a catalog can be
// published to external Org
func TestAccVcdCatalogPublishedToExternalOrg(t *testing.T) {
	preTestChecks(t)
	var params = StringMap{
		"Org":                                testConfig.VCD.Org,
		"Vdc":                                testConfig.VCD.Vdc,
		"CatalogName":                        TestAccVcdCatalogName,
		"Description":                        TestAccVcdCatalogDescription,
		"Tags":                               "catalog",
		"PublishEnabled":                     true,
		"PublishEnabledUpdate1":              true,
		"PublishEnabledUpdate2":              false,
		"CacheEnabled":                       true,
		"CacheEnabledUpdate1":                false,
		"CacheEnabledUpdate2":                false,
		"PreserveIdentityInformation":        true,
		"PreserveIdentityInformationUpdate1": false,
		"PreserveIdentityInformationUpdate2": false,
	}

	configText := templateFill(testAccCheckVcdCatalogPublished, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)
	params["FuncName"] = t.Name() + "step1"
	configTextUpd1 := templateFill(testAccCheckVcdCatalogPublishedUpdate1, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)
	params["FuncName"] = t.Name() + "step2"
	configTextUpd2 := templateFill(testAccCheckVcdCatalogPublishedUpdate2, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	resourceAddress := "vcd_catalog.test-catalog"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckCatalogDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdCatalogExists(resourceAddress),
					resource.TestCheckResourceAttr(resourceAddress, "name", TestAccVcdCatalogName),
					resource.TestCheckResourceAttr(resourceAddress, "description", TestAccVcdCatalogDescription),
					resource.TestCheckResourceAttr(resourceAddress, "publish_enabled",
						strconv.FormatBool(params["PublishEnabled"].(bool))),
					resource.TestCheckResourceAttr(resourceAddress, "preserve_identity_information",
						strconv.FormatBool(params["PreserveIdentityInformation"].(bool))),
					resource.TestCheckResourceAttr(resourceAddress, "cache_enabled",
						strconv.FormatBool(params["CacheEnabled"].(bool))),
					//resource.TestCheckResourceAttr(resourceAddress, "password", params[]),
				),
			},
			resource.TestStep{
				Config: configTextUpd1,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdCatalogExists(resourceAddress),
					resource.TestCheckResourceAttr(resourceAddress, "name", TestAccVcdCatalogName),
					resource.TestCheckResourceAttr(resourceAddress, "description", TestAccVcdCatalogDescription),
					resource.TestCheckResourceAttr(resourceAddress, "publish_enabled",
						strconv.FormatBool(params["PublishEnabledUpdate1"].(bool))),
					resource.TestCheckResourceAttr(resourceAddress, "preserve_identity_information",
						strconv.FormatBool(params["PreserveIdentityInformationUpdate1"].(bool))),
					resource.TestCheckResourceAttr(resourceAddress, "cache_enabled",
						strconv.FormatBool(params["CacheEnabledUpdate1"].(bool))),
					//resource.TestCheckResourceAttr(resourceAddress, "password", params[]),
				),
			},
			resource.TestStep{
				Config: configTextUpd2,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdCatalogExists(resourceAddress),
					resource.TestCheckResourceAttr(resourceAddress, "name", TestAccVcdCatalogName),
					resource.TestCheckResourceAttr(resourceAddress, "description", TestAccVcdCatalogDescription),
					resource.TestCheckResourceAttr(resourceAddress, "publish_enabled",
						strconv.FormatBool(params["PublishEnabledUpdate2"].(bool))),
					resource.TestCheckResourceAttr(resourceAddress, "preserve_identity_information",
						strconv.FormatBool(params["PreserveIdentityInformationUpdate2"].(bool))),
					resource.TestCheckResourceAttr(resourceAddress, "cache_enabled",
						strconv.FormatBool(params["CacheEnabledUpdate2"].(bool))),
					//resource.TestCheckResourceAttr(resourceAddress, "password", params[]),
				),
			},
		},
	})
	postTestChecks(t)
}

func testAccCheckVcdCatalogExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no Org ID is set")
		}

		conn := testAccProvider.Meta().(*VCDClient)

		adminOrg, err := conn.GetAdminOrg(testConfig.VCD.Org)
		if err != nil {
			return fmt.Errorf(errorRetrievingOrg, testConfig.VCD.Org+" and error: "+err.Error())
		}

		_, err = adminOrg.GetCatalogByNameOrId(rs.Primary.ID, false)
		if err != nil {
			return fmt.Errorf("catalog %s does not exist (%s)", rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccCheckCatalogDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*VCDClient)
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "vcd_catalog" && rs.Primary.Attributes["name"] != TestAccVcdCatalogName {
			continue
		}

		adminOrg, err := conn.GetAdminOrg(testConfig.VCD.Org)
		if err != nil {
			return fmt.Errorf(errorRetrievingOrg, testConfig.VCD.Org+" and error: "+err.Error())
		}

		_, err = adminOrg.GetCatalogByName(rs.Primary.ID, false)

		if err == nil {
			return fmt.Errorf("catalog %s still exists", rs.Primary.ID)
		}

	}

	return nil
}

const testAccCheckVcdCatalog = `
resource "vcd_catalog" "test-catalog" {
  org = "{{.Org}}" 
  
  name        = "{{.CatalogName}}"
  description = "{{.Description}}"

  delete_force      = "true"
  delete_recursive  = "true"
}
`

const testAccCheckVcdCatalogStep1 = `
data "vcd_storage_profile" "sp" {
	name = "{{.StorageProfile}}"
}

resource "vcd_catalog" "test-catalog" {
  org = "{{.Org}}" 
  
  name               = "{{.CatalogName}}"
  description        = "{{.Description}}"
  storage_profile_id = data.vcd_storage_profile.sp.id

  delete_force      = "true"
  delete_recursive  = "true"
}
`

// TestAccVcdCatalogSharedAccess is a test to cover bugfix when Organization Administrator is not able to lookup shared
// catalog from another Org
// Because of limited Terraform acceptance test functionality it uses go-vcloud-director SDK to pre-configure
// environment explicitly using System Org (even if it the test is run as Org user). The following objects are created
// using SDK (their cleanup is deferred):
// * Org
// * Vdc inside newly created Org
// * Catalog inside newly created Org. This catalog is shared with Org defined in testConfig.VCD.Org variable
// * Uploads A minimal vApp template to save on upload / VM spawn time
//
// After these objects are pre-created using SDK, terraform definition is used to spawn a VM by using template in a
// catalog from another Org. This test works in both System and Org admin roles but the bug (which was introduced in SDK
// v2.12.0 and terraform-provider-vcd v3.3.0) occurred only for Organization Administrator user.
//
// Original issue -  https://github.com/vmware/terraform-provider-vcd/issues/689
func TestAccVcdCatalogSharedAccess(t *testing.T) {
	preTestChecks(t)
	// This test manipulates VCD during runtime using SDK and is not possible to run as binary or upgrade test
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	// initiate System client ignoring value of `VCD_TEST_ORG_USER` flag and create pre-requisite objects
	systemClient := createSystemTemporaryVCDConnection()
	catalog, vdc, oldOrg, newOrg, err := spawnTestOrgVdcSharedCatalog(systemClient, t.Name())
	if err != nil {
		testOrgVdcSharedCatalogCleanUp(catalog, vdc, oldOrg, newOrg, t)
		t.Fatalf("%s", err)
	}
	// call cleanup ath the end of the test with any of the entities that have been created up to that point
	defer func() { testOrgVdcSharedCatalogCleanUp(catalog, vdc, oldOrg, newOrg, t) }()

	var params = StringMap{
		"Org":               testConfig.VCD.Org,
		"Vdc":               testConfig.VCD.Vdc,
		"TestName":          t.Name(),
		"SharedCatalog":     t.Name(),
		"SharedCatalogItem": "vapp-template",
		"Tags":              "catalog",
	}

	configText1 := templateFill(testAccCheckVcdCatalogShared, params)
	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText1)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy: resource.ComposeTestCheckFunc(
			testAccCheckVcdVAppVmDestroy(t.Name()),
			testAccCheckVcdStandaloneVmDestroy("test-standalone-vm", "", ""),
		),

		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText1,
				Check: resource.ComposeTestCheckFunc(
					// There is no need to check for much resources - the main point is to have the VMs created
					// without failures for catalog lookups or similar problems
					resource.TestCheckResourceAttrSet("vcd_vm.test-vm", "id"),
					resource.TestCheckResourceAttrSet("vcd_vapp.singleVM", "id"),
					resource.TestCheckResourceAttrSet("vcd_vapp_vm.test-vm", "id"),
				),
			},
		},
	})

	postTestChecks(t)
}

const testAccCheckVcdCatalogShared = `
resource "vcd_vm" "test-vm" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  name          = "test-standalone-vm"
  catalog_name  = "{{.SharedCatalog}}"
  template_name = "{{.SharedCatalogItem}}"
  power_on      = false
}

resource "vcd_vapp" "singleVM" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  name = "{{.TestName}}"
}

resource "vcd_vapp_vm" "test-vm" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  vapp_name     = vcd_vapp.singleVM.name
  name          = "test-vapp-vm"
  catalog_name  = "{{.SharedCatalog}}"
  template_name = "{{.SharedCatalogItem}}"
  power_on      = false

  depends_on = [vcd_vapp.singleVM]
}
`

// spawnTestOrgVdcSharedCatalog spawns an Org to be used in tests
func spawnTestOrgVdcSharedCatalog(client *VCDClient, name string) (govcd.AdminCatalog, *govcd.Vdc, *govcd.AdminOrg, *govcd.AdminOrg, error) {
	fmt.Println("# Setting up prerequisites using SDK (non Terraform definitions)")
	fmt.Printf("# Using user 'System' (%t) to prepare environment\n", client.Client.IsSysAdmin)

	existingOrg, err := client.GetAdminOrgByName(testConfig.VCD.Org)
	if err != nil {
		return govcd.AdminCatalog{}, nil, nil, nil, fmt.Errorf("error getting existing Org '%s': %s", testConfig.VCD.Org, err)
	}
	task, err := govcd.CreateOrg(client.VCDClient, name, name, name, existingOrg.AdminOrg.OrgSettings, true)
	if err != nil {
		return govcd.AdminCatalog{}, nil, existingOrg, nil, fmt.Errorf("error creating Org '%s': %s", name, err)
	}
	err = task.WaitTaskCompletion()
	if err != nil {
		return govcd.AdminCatalog{}, nil, existingOrg, nil, fmt.Errorf("task failed for Org '%s' creation: %s", name, err)
	}
	newAdminOrg, err := client.GetAdminOrgByName(name)
	if err != nil {
		return govcd.AdminCatalog{}, nil, existingOrg, nil, fmt.Errorf("error getting new Org '%s': %s", name, err)
	}
	fmt.Printf("# Created new Org '%s'\n", newAdminOrg.AdminOrg.Name)

	existingVdc, err := existingOrg.GetAdminVDCByName(testConfig.VCD.Vdc, false)
	if err != nil {
		return govcd.AdminCatalog{}, nil, existingOrg, newAdminOrg, fmt.Errorf("error retrieving existing VDC '%s': %s", testConfig.VCD.Vdc, err)

	}
	vdcConfiguration := &types.VdcConfiguration{
		Name:            name + "-VDC",
		Xmlns:           types.XMLNamespaceVCloud,
		AllocationModel: "Flex",
		ComputeCapacity: []*types.ComputeCapacity{
			&types.ComputeCapacity{
				CPU: &types.CapacityWithUsage{
					Units:     "MHz",
					Allocated: 1024,
					Limit:     1024,
				},
				Memory: &types.CapacityWithUsage{
					Allocated: 1024,
					Limit:     1024,
					Units:     "MB",
				},
			},
		},
		VdcStorageProfile: []*types.VdcStorageProfileConfiguration{&types.VdcStorageProfileConfiguration{
			Enabled: true,
			Units:   "MB",
			Limit:   1024,
			Default: true,
			ProviderVdcStorageProfile: &types.Reference{
				HREF: getVdcProviderVdcStorageProfileHref(client),
			},
		},
		},
		NetworkPoolReference: &types.Reference{
			HREF: existingVdc.AdminVdc.NetworkPoolReference.HREF,
		},
		ProviderVdcReference: &types.Reference{
			HREF: existingVdc.AdminVdc.ProviderVdcReference.HREF,
		},
		IsEnabled:             true,
		IsThinProvision:       true,
		UsesFastProvisioning:  true,
		IsElastic:             takeBoolPointer(true),
		IncludeMemoryOverhead: takeBoolPointer(true),
	}

	vdc, err := newAdminOrg.CreateOrgVdc(vdcConfiguration)
	if err != nil {
		return govcd.AdminCatalog{}, nil, existingOrg, newAdminOrg, err
	}
	fmt.Printf("# Created new Vdc '%s' inside Org '%s'\n", vdc.Vdc.Name, newAdminOrg.AdminOrg.Name)

	catalog, err := newAdminOrg.CreateCatalog(name, name)
	if err != nil {
		return govcd.AdminCatalog{}, vdc, existingOrg, newAdminOrg, err
	}
	fmt.Printf("# Created new Catalog '%s' inside Org '%s'\n", catalog.AdminCatalog.Name, newAdminOrg.AdminOrg.Name)

	// Share new Catalog in newOrgName1 with default test Org vcd.Org
	readOnly := "ReadOnly"
	accessControl := &types.ControlAccessParams{
		IsSharedToEveryone:  false,
		EveryoneAccessLevel: &readOnly,
		AccessSettings: &types.AccessSettingList{
			AccessSetting: []*types.AccessSetting{&types.AccessSetting{
				Subject: &types.LocalSubject{
					HREF: existingOrg.AdminOrg.HREF,
					Name: existingOrg.AdminOrg.Name,
					Type: types.MimeOrg,
				},
				AccessLevel: "ReadOnly",
			}},
		},
	}
	err = catalog.SetAccessControl(accessControl, false)
	if err != nil {
		return catalog, vdc, existingOrg, newAdminOrg, err
	}
	fmt.Printf("# Shared new Catalog '%s' with existing Org '%s'\n",
		catalog.AdminCatalog.Name, existingOrg.AdminOrg.Name)

	uploadTask, err := catalog.UploadOvf(testConfig.Ova.OvaPath, "vapp-template", "upload from test", 1024)
	if err != nil {
		return catalog, vdc, existingOrg, newAdminOrg, fmt.Errorf("error uploading template: %s", err)
	}

	err = uploadTask.WaitTaskCompletion()
	if err != nil {
		return catalog, vdc, existingOrg, newAdminOrg, fmt.Errorf("error uploading template: %s", err)
	}
	fmt.Printf("# Uploaded vApp template '%s' to shared new Catalog '%s' in new Org '%s' with existing Org '%s'\n",
		"vapp-template", catalog.AdminCatalog.Name, newAdminOrg.AdminOrg.Name, existingOrg.AdminOrg.Name)

	return catalog, vdc, existingOrg, newAdminOrg, nil
}

func testOrgVdcSharedCatalogCleanUp(catalog govcd.AdminCatalog, vdc *govcd.Vdc, existingOrg, newAdminOrg *govcd.AdminOrg, t *testing.T) {
	fmt.Println("# Cleaning up")
	var err error
	if catalog != (govcd.AdminCatalog{}) {
		err = catalog.Delete(true, true)
		if err != nil {
			t.Errorf("error cleaning up catalog: %s", err)
		}
		// The catalog.Delete ignores the task returned and does not wait for the operation to complete. This code
		// was made for a particular bugfix therefore other parts of code were not altered/fixed.
		for i := 0; i < 30; i++ {
			_, err := existingOrg.GetAdminCatalogById(catalog.AdminCatalog.ID, true)
			if govcd.ContainsNotFound(err) {
				break
			} else {
				time.Sleep(time.Second)
			}
		}
	}

	if vdc != nil {
		err = vdc.DeleteWait(true, true)
		if err != nil {
			t.Errorf("error cleaning up VDC: %s", err)
		}
	}

	if newAdminOrg != nil {
		err = newAdminOrg.Refresh()
		if err != nil {
			t.Errorf("error refreshing Org: %s", err)
		}
		err = newAdminOrg.Delete(true, true)
		if err != nil {
			t.Errorf("error cleaning up Org: %s", err)
		}
	}
}

func getVdcProviderVdcStorageProfileHref(client *VCDClient) string {
	results, _ := client.QueryWithNotEncodedParams(nil, map[string]string{
		"type":   "providerVdcStorageProfile",
		"filter": fmt.Sprintf("name==%s", testConfig.VCD.ProviderVdc.StorageProfile2),
	})
	providerVdcStorageProfileHref := results.Results.ProviderVdcStorageProfileRecord[0].HREF
	return providerVdcStorageProfileHref
}
