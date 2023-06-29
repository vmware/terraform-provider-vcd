//go:build disk || ALL || functional

package vcd

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/vmware/go-vcloud-director/v2/govcd"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var resourceName = "TestAccVcdIndependentDiskBasic_1"
var resourceNameSecond = "TestAccVcdIndependentDiskBasic_2"
var resourceNameThird = "TestAccVcdIndependentDiskBasic_3"
var name = "TestAccVcdIndependentDiskBasic"

func TestAccVcdIndependentDiskBasic(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	if testConfig.VCD.NsxtProviderVdc.StorageProfile == "" || testConfig.VCD.NsxtProviderVdc.StorageProfile2 == "" {
		t.Skip("Both variables testConfig.VCD.ProviderVdc.StorageProfile and testConfig.VCD.ProviderVdc.StorageProfile2 must be set")
	}

	var params = StringMap{
		"Org":                testConfig.VCD.Org,
		"Vdc":                testConfig.Nsxt.Vdc,
		"name":               name,
		"description":        "independent disk description",
		"secondName":         name + "second",
		"size":               "5000",
		"busType":            "SCSI",
		"busSubType":         "lsilogicsas",
		"storageProfileName": testConfig.VCD.NsxtProviderVdc.StorageProfile,
		"ResourceName":       resourceName,
		"secondResourceName": resourceNameSecond,
		"thirdResourceName":  resourceNameThird,
		"Tags":               "disk",
		"sizeUpdate":         "6000",
		"busTypeNvme":        "NVME",
		"busSubTypeNvme":     "nvmecontroller",
		"VmName":             t.Name(),
		"Catalog":            testSuiteCatalogName,
		"CatalogItem":        testSuiteCatalogOVAItem,
		"metadataKey":        "key1",
		"metadataValue":      "value1",
	}
	testParamsNotEmpty(t, params)

	params["FuncName"] = t.Name() + "-Compatibility"
	configTextForCompatibility := templateFill(testAccCheckVcdIndependentDiskForCompatibility, params)
	params["FuncName"] = t.Name() + "-WithoutOptionals"
	configTextWithoutOptionals := templateFill(testAccCheckVcdIndependentDiskWithoutOptionals, params)
	params["FuncName"] = t.Name() + "-Nvme"
	configTextNvme := templateFill(testAccCheckVcdIndependentDiskNvmeType, params)
	params["FuncName"] = t.Name() + "-attachedToVm"
	configTextAttachedToVm := templateFill(testAccCheckVcdIndependentDiskAttachedToVm, params)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	debugPrintf("#[DEBUG] CONFIGURATION: %s", configTextForCompatibility)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testDiskResourcesDestroyed,
		Steps: []resource.TestStep{
			{
				Config: configTextForCompatibility,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDiskCreated("vcd_independent_disk."+resourceName),
					resource.TestMatchResourceAttr("vcd_independent_disk."+resourceName, "owner_name", regexp.MustCompile(`^\S+`)),
					resource.TestMatchResourceAttr("vcd_independent_disk."+resourceName, "datastore_name", regexp.MustCompile(`^\S+`)),
					resource.TestMatchResourceAttr("vcd_independent_disk."+resourceName, "iops", regexp.MustCompile(`^\d+$`)),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "bus_type", params["busType"].(string)),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "bus_sub_type", params["busSubType"].(string)),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "is_attached", "false"),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "size_in_mb", params["size"].(string)),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "name", params["name"].(string)),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "description", params["description"].(string)),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "storage_profile", params["storageProfileName"].(string)),
					resource.TestMatchResourceAttr("vcd_independent_disk."+resourceName, "uuid", regexp.MustCompile(`^\S+`)),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "sharing_type", "None"),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "encrypted", "false"),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "attached_vm_ids.#", "0"),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "metadata."+params["metadataKey"].(string), params["metadataValue"].(string)),
				),
			},
			{
				ResourceName:            "vcd_independent_disk." + resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdFunc:       importStateIdByDisk("vcd_independent_disk." + resourceName),
				ImportStateVerifyIgnore: []string{"org", "vdc"},
			},
			{
				Config: configTextWithoutOptionals,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDiskCreated("vcd_independent_disk."+resourceNameSecond),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceNameSecond, "size_in_mb", params["size"].(string)),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceNameSecond, "bus_type", "SCSI"),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceNameSecond, "bus_sub_type", "lsilogic"),
					resource.TestMatchResourceAttr("vcd_independent_disk."+resourceNameSecond, "owner_name", regexp.MustCompile(`^\S+`)),
					resource.TestMatchResourceAttr("vcd_independent_disk."+resourceNameSecond, "datastore_name", regexp.MustCompile(`^\S+`)),
					resource.TestMatchResourceAttr("vcd_independent_disk."+resourceNameSecond, "iops", regexp.MustCompile(`^\d+$`)),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceNameSecond, "is_attached", "false"),
				),
			},
			{
				Config: configTextNvme,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_independent_disk."+resourceName, "owner_name", regexp.MustCompile(`^\S+`)),
					resource.TestMatchResourceAttr("vcd_independent_disk."+resourceName, "datastore_name", regexp.MustCompile(`^\S+`)),
					resource.TestMatchResourceAttr("vcd_independent_disk."+resourceName, "iops", regexp.MustCompile(`^\d+$`)),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "bus_type", params["busTypeNvme"].(string)),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "bus_sub_type", params["busSubTypeNvme"].(string)),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "is_attached", "false"),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "size_in_mb", params["size"].(string)),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "name", params["name"].(string)),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "description", params["description"].(string)),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "storage_profile", params["storageProfileName"].(string)),
					resource.TestMatchResourceAttr("vcd_independent_disk."+resourceName, "uuid", regexp.MustCompile(`^\S+`)),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "sharing_type", "None"),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "encrypted", "false"),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "attached_vm_ids.#", "0"),
				),
			},
			{
				Config: configTextAttachedToVm,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_independent_disk."+resourceName, "owner_name", regexp.MustCompile(`^\S+`)),
					resource.TestMatchResourceAttr("vcd_independent_disk."+resourceName, "datastore_name", regexp.MustCompile(`^\S+`)),
					resource.TestMatchResourceAttr("vcd_independent_disk."+resourceName, "iops", regexp.MustCompile(`^\d+$`)),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "bus_type", params["busType"].(string)),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "bus_sub_type", params["busSubType"].(string)),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "is_attached", "false"),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "size_in_mb", params["size"].(string)),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "name", params["name"].(string)),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "description", params["description"].(string)),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "storage_profile", params["storageProfileName"].(string)),
					resource.TestMatchResourceAttr("vcd_independent_disk."+resourceName, "uuid", regexp.MustCompile(`^\S+`)),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "sharing_type", "None"),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "encrypted", "false"),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "attached_vm_ids.#", "0"),

					resource.TestMatchResourceAttr("vcd_independent_disk."+resourceNameThird, "owner_name", regexp.MustCompile(`^\S+`)),
					resource.TestMatchResourceAttr("vcd_independent_disk."+resourceNameThird, "datastore_name", regexp.MustCompile(`^\S+`)),
					resource.TestMatchResourceAttr("vcd_independent_disk."+resourceNameThird, "iops", regexp.MustCompile(`^\d+$`)),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceNameThird, "bus_type", params["busType"].(string)),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceNameThird, "bus_sub_type", params["busSubType"].(string)),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceNameThird, "is_attached", "false"),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceNameThird, "size_in_mb", params["size"].(string)),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceNameThird, "name", resourceNameThird),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceNameThird, "description", params["description"].(string)),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceNameThird, "storage_profile", params["storageProfileName"].(string)),
					resource.TestMatchResourceAttr("vcd_independent_disk."+resourceNameThird, "uuid", regexp.MustCompile(`^\S+`)),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceNameThird, "sharing_type", "None"),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceNameThird, "encrypted", "false"),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceNameThird, "attached_vm_ids.#", "0"),
				),
			},
		},
	})
	postTestChecks(t)
}

// TestAccVcdIndependentDiskBasicWithUpdates is very similar to TestAccVcdIndependentDiskBasic, but also tests updating the disks.
func TestAccVcdIndependentDiskBasicWithUpdates(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	// The test is being skipped due to a known bug in the current versions of VCD
	// when updating the disk resources, thus the test will only be run on versions
	// released after v10.4.1.
	vcdClient := createTemporaryVCDConnection(false)
	if vcdClient.Client.APIVCDMaxVersionIs("<= 37.1") {
		t.Skip("This test may fail on versions up to VCD 10.4.1 (API V37.1) because of a known bug. Skipping.")
	}

	if testConfig.VCD.NsxtProviderVdc.StorageProfile == "" || testConfig.VCD.NsxtProviderVdc.StorageProfile2 == "" {
		t.Skip("Both variables testConfig.VCD.ProviderVdc.StorageProfile and testConfig.VCD.ProviderVdc.StorageProfile2 must be set")
	}

	var params = StringMap{
		"Org":                      testConfig.VCD.Org,
		"Vdc":                      testConfig.Nsxt.Vdc,
		"name":                     name,
		"description":              "independent disk description",
		"secondName":               name + "second",
		"size":                     "5000",
		"busType":                  "SCSI",
		"busSubType":               "lsilogicsas",
		"storageProfileName":       testConfig.VCD.NsxtProviderVdc.StorageProfile,
		"ResourceName":             resourceName,
		"secondResourceName":       resourceNameSecond,
		"thirdResourceName":        resourceNameThird,
		"Tags":                     "disk",
		"descriptionUpdate":        "independent disk description updated",
		"sizeUpdate":               "6000",
		"storageProfileNameUpdate": testConfig.VCD.NsxtProviderVdc.StorageProfile2,
		"busTypeNvme":              "NVME",
		"busSubTypeNvme":           "nvmecontroller",
		"VmName":                   t.Name(),
		"Catalog":                  testSuiteCatalogName,
		"CatalogItem":              testSuiteCatalogOVAItem,
		"metadataKey":              "key1",
		"metadataValue":            "value1",
		"metadataKeyUpdate":        "key2",
		"metadataValueUpdate":      "value2",
	}
	testParamsNotEmpty(t, params)

	params["FuncName"] = t.Name() + "-Compatibility"
	configTextForCompatibility := templateFill(testAccCheckVcdIndependentDiskForCompatibility, params)
	params["FuncName"] = t.Name() + "-WithoutOptionals"
	configTextWithoutOptionals := templateFill(testAccCheckVcdIndependentDiskWithoutOptionals, params)
	params["FuncName"] = t.Name() + "-Update"
	configTextForUpdate := templateFill(testAccCheckVcdIndependentDiskForUpdate, params)

	var configTextNvme string
	var configTextNvmeUpdate string

	params["FuncName"] = t.Name() + "-Nvme"
	configTextNvme = templateFill(testAccCheckVcdIndependentDiskNvmeType, params)
	params["FuncName"] = t.Name() + "-NvmeUpdate"
	configTextNvmeUpdate = templateFill(testAccCheckVcdIndependentDiskNvmeTypeUpdate, params)

	params["FuncName"] = t.Name() + "-attachedToVm"
	configTextAttachedToVm := templateFill(testAccCheckVcdIndependentDiskAttachedToVm, params)
	params["FuncName"] = t.Name() + "-attachedToVmUpdate"
	configTextAttachedToVmUpdate := templateFill(testAccCheckVcdIndependentDiskAttachedToVmUpdate, params)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	debugPrintf("#[DEBUG] CONFIGURATION: %s", configTextForCompatibility)

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy:      testDiskResourcesDestroyed,
		Steps: []resource.TestStep{
			{
				Config: configTextForCompatibility,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDiskCreated("vcd_independent_disk."+resourceName),
					resource.TestMatchResourceAttr("vcd_independent_disk."+resourceName, "owner_name", regexp.MustCompile(`^\S+`)),
					resource.TestMatchResourceAttr("vcd_independent_disk."+resourceName, "datastore_name", regexp.MustCompile(`^\S+`)),
					resource.TestMatchResourceAttr("vcd_independent_disk."+resourceName, "iops", regexp.MustCompile(`^\d+$`)),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "bus_type", params["busType"].(string)),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "bus_sub_type", params["busSubType"].(string)),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "is_attached", "false"),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "size_in_mb", params["size"].(string)),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "name", params["name"].(string)),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "description", params["description"].(string)),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "storage_profile", params["storageProfileName"].(string)),
					resource.TestMatchResourceAttr("vcd_independent_disk."+resourceName, "uuid", regexp.MustCompile(`^\S+`)),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "sharing_type", "None"),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "encrypted", "false"),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "attached_vm_ids.#", "0"),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "metadata."+params["metadataKey"].(string), params["metadataValue"].(string)),
				),
			},
			{
				Config: configTextForUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_independent_disk."+resourceName, "owner_name", regexp.MustCompile(`^\S+`)),
					resource.TestMatchResourceAttr("vcd_independent_disk."+resourceName, "datastore_name", regexp.MustCompile(`^\S+`)),
					resource.TestMatchResourceAttr("vcd_independent_disk."+resourceName, "iops", regexp.MustCompile(`^\d+$`)),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "bus_type", params["busType"].(string)),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "bus_sub_type", params["busSubType"].(string)),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "is_attached", "false"),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "size_in_mb", params["sizeUpdate"].(string)),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "name", params["name"].(string)),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "description", params["descriptionUpdate"].(string)),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "storage_profile", params["storageProfileNameUpdate"].(string)),
					resource.TestMatchResourceAttr("vcd_independent_disk."+resourceName, "uuid", regexp.MustCompile(`^\S+`)),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "sharing_type", "None"),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "encrypted", "false"),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "attached_vm_ids.#", "0"),
					resource.TestCheckNoResourceAttr("vcd_independent_disk."+resourceName, "metadata."+params["metadataKey"].(string)),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "metadata."+params["metadataKeyUpdate"].(string), params["metadataValueUpdate"].(string)),
				),
			},
			{
				ResourceName:            "vcd_independent_disk." + resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdFunc:       importStateIdByDisk("vcd_independent_disk." + resourceName),
				ImportStateVerifyIgnore: []string{"org", "vdc"},
			},
			{
				Config: configTextWithoutOptionals,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDiskCreated("vcd_independent_disk."+resourceNameSecond),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceNameSecond, "size_in_mb", params["size"].(string)),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceNameSecond, "bus_type", "SCSI"),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceNameSecond, "bus_sub_type", "lsilogic"),
					resource.TestMatchResourceAttr("vcd_independent_disk."+resourceNameSecond, "owner_name", regexp.MustCompile(`^\S+`)),
					resource.TestMatchResourceAttr("vcd_independent_disk."+resourceNameSecond, "datastore_name", regexp.MustCompile(`^\S+`)),
					resource.TestMatchResourceAttr("vcd_independent_disk."+resourceNameSecond, "iops", regexp.MustCompile(`^\d+$`)),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceNameSecond, "is_attached", "false"),
				),
			},
			{
				Config: configTextNvme,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_independent_disk."+resourceName, "owner_name", regexp.MustCompile(`^\S+`)),
					resource.TestMatchResourceAttr("vcd_independent_disk."+resourceName, "datastore_name", regexp.MustCompile(`^\S+`)),
					resource.TestMatchResourceAttr("vcd_independent_disk."+resourceName, "iops", regexp.MustCompile(`^\d+$`)),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "bus_type", params["busTypeNvme"].(string)),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "bus_sub_type", params["busSubTypeNvme"].(string)),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "is_attached", "false"),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "size_in_mb", params["size"].(string)),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "name", params["name"].(string)),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "description", params["description"].(string)),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "storage_profile", params["storageProfileName"].(string)),
					resource.TestMatchResourceAttr("vcd_independent_disk."+resourceName, "uuid", regexp.MustCompile(`^\S+`)),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "sharing_type", "None"),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "encrypted", "false"),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "attached_vm_ids.#", "0"),
				),
			},
			{
				Config: configTextNvmeUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_independent_disk."+resourceName, "owner_name", regexp.MustCompile(`^\S+`)),
					resource.TestMatchResourceAttr("vcd_independent_disk."+resourceName, "datastore_name", regexp.MustCompile(`^\S+`)),
					resource.TestMatchResourceAttr("vcd_independent_disk."+resourceName, "iops", regexp.MustCompile(`^\d+$`)),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "bus_type", params["busTypeNvme"].(string)),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "bus_sub_type", params["busSubTypeNvme"].(string)),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "is_attached", "false"),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "size_in_mb", params["sizeUpdate"].(string)),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "name", params["name"].(string)),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "description", params["descriptionUpdate"].(string)),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "storage_profile", params["storageProfileNameUpdate"].(string)),
					resource.TestMatchResourceAttr("vcd_independent_disk."+resourceName, "uuid", regexp.MustCompile(`^\S+`)),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "sharing_type", "None"),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "encrypted", "false"),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "attached_vm_ids.#", "0"),
				),
			},
			{
				Config: configTextAttachedToVm,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_independent_disk."+resourceName, "owner_name", regexp.MustCompile(`^\S+`)),
					resource.TestMatchResourceAttr("vcd_independent_disk."+resourceName, "datastore_name", regexp.MustCompile(`^\S+`)),
					resource.TestMatchResourceAttr("vcd_independent_disk."+resourceName, "iops", regexp.MustCompile(`^\d+$`)),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "bus_type", params["busType"].(string)),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "bus_sub_type", params["busSubType"].(string)),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "is_attached", "false"),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "size_in_mb", params["size"].(string)),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "name", params["name"].(string)),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "description", params["description"].(string)),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "storage_profile", params["storageProfileName"].(string)),
					resource.TestMatchResourceAttr("vcd_independent_disk."+resourceName, "uuid", regexp.MustCompile(`^\S+`)),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "sharing_type", "None"),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "encrypted", "false"),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "attached_vm_ids.#", "0"),

					resource.TestMatchResourceAttr("vcd_independent_disk."+resourceNameThird, "owner_name", regexp.MustCompile(`^\S+`)),
					resource.TestMatchResourceAttr("vcd_independent_disk."+resourceNameThird, "datastore_name", regexp.MustCompile(`^\S+`)),
					resource.TestMatchResourceAttr("vcd_independent_disk."+resourceNameThird, "iops", regexp.MustCompile(`^\d+$`)),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceNameThird, "bus_type", params["busType"].(string)),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceNameThird, "bus_sub_type", params["busSubType"].(string)),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceNameThird, "is_attached", "false"),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceNameThird, "size_in_mb", params["size"].(string)),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceNameThird, "name", resourceNameThird),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceNameThird, "description", params["description"].(string)),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceNameThird, "storage_profile", params["storageProfileName"].(string)),
					resource.TestMatchResourceAttr("vcd_independent_disk."+resourceNameThird, "uuid", regexp.MustCompile(`^\S+`)),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceNameThird, "sharing_type", "None"),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceNameThird, "encrypted", "false"),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceNameThird, "attached_vm_ids.#", "0"),
				),
			},
			{
				Config: configTextAttachedToVmUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("vcd_independent_disk."+resourceName, "owner_name", regexp.MustCompile(`^\S+`)),
					resource.TestMatchResourceAttr("vcd_independent_disk."+resourceName, "datastore_name", regexp.MustCompile(`^\S+`)),
					resource.TestMatchResourceAttr("vcd_independent_disk."+resourceName, "iops", regexp.MustCompile(`^\d+$`)),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "bus_type", params["busType"].(string)),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "bus_sub_type", params["busSubType"].(string)),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "is_attached", "true"),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "size_in_mb", params["sizeUpdate"].(string)),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "name", params["name"].(string)),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "description", params["descriptionUpdate"].(string)),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "storage_profile", params["storageProfileNameUpdate"].(string)),
					resource.TestMatchResourceAttr("vcd_independent_disk."+resourceName, "uuid", regexp.MustCompile(`^\S+`)),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "sharing_type", "None"),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "encrypted", "false"),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "attached_vm_ids.#", "1"),

					resource.TestMatchResourceAttr("vcd_independent_disk."+resourceNameThird, "owner_name", regexp.MustCompile(`^\S+`)),
					resource.TestMatchResourceAttr("vcd_independent_disk."+resourceNameThird, "datastore_name", regexp.MustCompile(`^\S+`)),
					resource.TestMatchResourceAttr("vcd_independent_disk."+resourceNameThird, "iops", regexp.MustCompile(`^\d+$`)),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceNameThird, "bus_type", params["busType"].(string)),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceNameThird, "bus_sub_type", params["busSubType"].(string)),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceNameThird, "is_attached", "true"),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceNameThird, "size_in_mb", params["sizeUpdate"].(string)),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceNameThird, "name", resourceNameThird),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceNameThird, "description", params["descriptionUpdate"].(string)),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceNameThird, "storage_profile", params["storageProfileNameUpdate"].(string)),
					resource.TestMatchResourceAttr("vcd_independent_disk."+resourceNameThird, "uuid", regexp.MustCompile(`^\S+`)),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceNameThird, "sharing_type", "None"),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceNameThird, "encrypted", "false"),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceNameThird, "attached_vm_ids.#", "1"),
				),
			},
		},
	})
	postTestChecks(t)
}

func testAccCheckDiskCreated(itemName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		injectItemRs, ok := s.RootModule().Resources[itemName]
		if !ok {
			return fmt.Errorf("not found: %s", itemName)
		}

		if injectItemRs.Primary.ID == "" {
			return fmt.Errorf("no disk insert ID is set")
		}

		conn := testAccProvider.Meta().(*VCDClient)

		_, vdc, err := conn.GetOrgAndVdc(testConfig.VCD.Org, testConfig.Nsxt.Vdc)
		if err != nil {
			return fmt.Errorf(errorRetrievingVdcFromOrg, testConfig.VCD.Vdc, testConfig.VCD.Org, err)
		}

		_, err = vdc.GetDiskById(injectItemRs.Primary.ID, true)
		if err != nil {
			return fmt.Errorf("independent disk %s isn't exist and error: %#v", itemName, err)
		}

		return nil
	}
}

func testDiskResourcesDestroyed(s *terraform.State) error {
	conn := testAccProvider.Meta().(*VCDClient)
	for _, rs := range s.RootModule().Resources {
		itemName := rs.Primary.Attributes["name"]
		if rs.Type != "vcd_independent_disk" && itemName != name {
			continue
		}

		_, vdc, err := conn.GetOrgAndVdc(testConfig.VCD.Org, testConfig.Nsxt.Vdc)
		if err != nil {
			return fmt.Errorf(errorRetrievingVdcFromOrg, testConfig.VCD.Vdc, testConfig.VCD.Org, err)
		}

		_, err = vdc.GetDisksByName(name, true)
		if !govcd.IsNotFound(err) {
			return fmt.Errorf("independent disk %s still exist and error: %#v", itemName, err)
		}

	}
	return nil
}

func importStateIdByDisk(resource string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resource]
		if !ok {
			return "", fmt.Errorf("not found resource: %s", resource)
		}

		if rs.Primary.ID == "" {
			return "", fmt.Errorf("no ID is set for %s resource", resource)
		}

		importId := testConfig.VCD.Org + "." + testConfig.Nsxt.Vdc + "." + rs.Primary.ID
		if testConfig.VCD.Org == "" || testConfig.Nsxt.Vdc == "" || rs.Primary.ID == "" {
			return "", fmt.Errorf("missing information to generate import path: %s", importId)
		}
		return importId, nil
	}
}

func init() {
	testingTags["disk"] = "resource_vcd_independent_disk_test.go"
}

const testAccCheckVcdIndependentDiskForCompatibility = `
resource "vcd_independent_disk" "{{.ResourceName}}" {
  org             = "{{.Org}}"
  vdc             = "{{.Vdc}}"
  name            = "{{.name}}"
  description     = "{{.description}}"
  size_in_mb      = "{{.size}}"
  bus_type        = "{{.busType}}"
  bus_sub_type    = "{{.busSubType}}"
  storage_profile = "{{.storageProfileName}}"
  metadata = {
    {{.metadataKey}} = "{{.metadataValue}}"
  }
}
`

const testAccCheckVcdIndependentDiskForUpdate = `
# skip-binary-test: only for updates
resource "vcd_independent_disk" "{{.ResourceName}}" {
  org             = "{{.Org}}"
  vdc             = "{{.Vdc}}"
  name            = "{{.name}}"
  description     = "{{.descriptionUpdate}}"
  size_in_mb      = "{{.sizeUpdate}}"
  bus_type        = "{{.busType}}"
  bus_sub_type    = "{{.busSubType}}"
  storage_profile = "{{.storageProfileNameUpdate}}"
  metadata = {
    {{.metadataKeyUpdate}} = "{{.metadataValueUpdate}}"
  }
}
`

const testAccCheckVcdIndependentDiskWithoutOptionals = `
resource "vcd_independent_disk" "{{.secondResourceName}}" {
  name            = "{{.secondName}}"
  size_in_mb      = "{{.size}}"
}
`

const testAccCheckVcdIndependentDiskNvmeType = `
resource "vcd_independent_disk" "{{.ResourceName}}" {
  org             = "{{.Org}}"
  vdc             = "{{.Vdc}}"
  name            = "{{.name}}"
  description     = "{{.description}}"
  size_in_mb      = "{{.size}}"
  bus_type        = "{{.busTypeNvme}}"
  bus_sub_type    = "{{.busSubTypeNvme}}"
  storage_profile = "{{.storageProfileName}}"
}
`

const testAccCheckVcdIndependentDiskNvmeTypeUpdate = `
# skip-binary-test: only for updates
resource "vcd_independent_disk" "{{.ResourceName}}" {
  org             = "{{.Org}}"
  vdc             = "{{.Vdc}}"
  name            = "{{.name}}"
  description     = "{{.descriptionUpdate}}"
  size_in_mb      = "{{.sizeUpdate}}"
  bus_type        = "{{.busTypeNvme}}"
  bus_sub_type    = "{{.busSubTypeNvme}}"
  storage_profile = "{{.storageProfileNameUpdate}}"
}
`

const testAccCheckVcdIndependentDiskAttachedToVm = `
resource "vcd_independent_disk" "{{.ResourceName}}" {
  org             = "{{.Org}}"
  vdc             = "{{.Vdc}}"
  name            = "{{.name}}"
  description     = "{{.description}}"
  size_in_mb      = "{{.size}}"
  bus_type        = "{{.busType}}"
  bus_sub_type    = "{{.busSubType}}"
  storage_profile = "{{.storageProfileName}}"
}

resource "vcd_independent_disk" "{{.thirdResourceName}}" {
  org             = "{{.Org}}"
  vdc             = "{{.Vdc}}"
  name            = "{{.thirdResourceName}}"
  description     = "{{.description}}"
  size_in_mb      = "{{.size}}"
  bus_type        = "{{.busType}}"
  bus_sub_type    = "{{.busSubType}}"
  storage_profile = "{{.storageProfileName}}"
}


resource "vcd_vapp" "{{.ResourceName}}" {
  name = "{{.ResourceName}}"
}

resource "vcd_vapp_vm" "{{.VmName}}" {
  vapp_name     = vcd_vapp.{{.ResourceName}}.name
  name          = "{{.VmName}}"
  catalog_name  = "{{.Catalog}}"
  template_name = "{{.CatalogItem}}"
  memory        = 1024
  cpus          = 1
  power_on      = "false"

  hardware_version = "vmx-13"

  disk {
    name        = vcd_independent_disk.{{.ResourceName}}.name
    bus_number  = 1
    unit_number = 0
  }

  disk {
    name        = vcd_independent_disk.{{.thirdResourceName}}.name
    bus_number  = 1
    unit_number = 1
  }

}
`

const testAccCheckVcdIndependentDiskAttachedToVmUpdate = `
# skip-binary-test: only for updates
resource "vcd_independent_disk" "{{.ResourceName}}" {
  org             = "{{.Org}}"
  vdc             = "{{.Vdc}}"
  name            = "{{.name}}"
  description     = "{{.descriptionUpdate}}"
  size_in_mb      = "{{.sizeUpdate}}"
  bus_type        = "{{.busType}}"
  bus_sub_type    = "{{.busSubType}}"
  storage_profile = "{{.storageProfileNameUpdate}}"
}

resource "vcd_independent_disk" "{{.thirdResourceName}}" {
  org             = "{{.Org}}"
  vdc             = "{{.Vdc}}"
  name            = "{{.thirdResourceName}}"
  description     = "{{.descriptionUpdate}}"
  size_in_mb      = "{{.sizeUpdate}}"
  bus_type        = "{{.busType}}"
  bus_sub_type    = "{{.busSubType}}"
  storage_profile = "{{.storageProfileNameUpdate}}"
}


resource "vcd_vapp" "{{.ResourceName}}" {
  name = "{{.ResourceName}}"
}

resource "vcd_vapp_vm" "{{.VmName}}" {
  vapp_name     = vcd_vapp.{{.ResourceName}}.name
  name          = "{{.VmName}}"
  catalog_name  = "{{.Catalog}}"
  template_name = "{{.CatalogItem}}"
  memory        = 1024
  cpus          = 1
  power_on      = "false"

  hardware_version = "vmx-13"

  disk {
    name        = vcd_independent_disk.{{.ResourceName}}.name
    bus_number  = 1
    unit_number = 0
  }

  disk {
    name        = vcd_independent_disk.{{.thirdResourceName}}.name
    bus_number  = 1
    unit_number = 1
  }

}
`

// TestAccVcdIndependentDiskMetadata tests metadata CRUD on independent disks
func TestAccVcdIndependentDiskMetadata(t *testing.T) {
	testMetadataEntryCRUD(t,
		testAccCheckVcdIndependentDiskMetadata, "vcd_independent_disk.test-independent-disk",
		testAccCheckVcdIndependentDiskMetadataDatasource, "data.vcd_independent_disk.test-independent-disk-ds",
		StringMap{
			"StorageProfile": testConfig.VCD.NsxtProviderVdc.StorageProfile,
		})
}

const testAccCheckVcdIndependentDiskMetadata = `
resource "vcd_independent_disk" "test-independent-disk" {
  org             = "{{.Org}}"
  vdc             = "{{.Vdc}}"
  name            = "{{.Name}}"
  size_in_mb      = "1024"
  bus_type        = "SCSI"
  bus_sub_type    = "VirtualSCSI"
  storage_profile = "{{.StorageProfile}}"
  {{.Metadata}}
}
`

const testAccCheckVcdIndependentDiskMetadataDatasource = `
data "vcd_independent_disk" "test-independent-disk-ds" {
  org     = vcd_independent_disk.test-independent-disk.org
  name    = vcd_independent_disk.test-independent-disk.name
}
`

func TestAccVcdIndependentDiskMetadataIgnore(t *testing.T) {
	skipIfNotSysAdmin(t)

	getObjectById := func(vcdClient *VCDClient, id string) (metadataCompatible, error) {
		adminOrg, err := vcdClient.GetAdminOrgByName(testConfig.VCD.Org)
		if err != nil {
			return nil, fmt.Errorf("could not retrieve Org '%s': %s", testConfig.VCD.Org, err)
		}
		vdc, err := adminOrg.GetVDCByName(testConfig.Nsxt.Vdc, true)
		if err != nil {
			return nil, fmt.Errorf("could not retrieve VDC '%s': %s", testConfig.Nsxt.Vdc, err)
		}
		disk, err := vdc.GetDiskById(id, true)
		if err != nil {
			return nil, fmt.Errorf("could not retrieve Independent Disk '%s': %s", id, err)
		}
		return disk, nil
	}

	testMetadataEntryIgnore(t,
		testAccCheckVcdIndependentDiskMetadata, "vcd_independent_disk.test-independent-disk",
		testAccCheckVcdIndependentDiskMetadataDatasource, "data.vcd_independent_disk.test-independent-disk-ds",
		getObjectById, StringMap{
			"StorageProfile": testConfig.VCD.NsxtProviderVdc.StorageProfile,
		})
}
