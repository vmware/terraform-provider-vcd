//go:build disk || ALL || functional
// +build disk ALL functional

package vcd

import (
	"fmt"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var resourceName = "TestAccVcdIndependentDiskBasic_1"
var resourceNameSecond = "TestAccVcdIndependentDiskBasic_2"
var resourceNameThird = "TestAccVcdIndependentDiskBasic_3"
var name = "TestAccVcdIndependentDiskBasic"

func TestAccVcdIndependentDiskBasic(t *testing.T) {
	preTestChecks(t)
	if !usingSysAdmin() {
		t.Skip("TestAccVcdIndependentDiskBasic requires system admin privileges")
	}

	if testConfig.VCD.ProviderVdc.StorageProfile == "" || testConfig.VCD.ProviderVdc.StorageProfile2 == "" {
		t.Skip("Both variables testConfig.VCD.ProviderVdc.StorageProfile and testConfig.VCD.ProviderVdc.StorageProfile2 must be set")
	}

	var params = StringMap{
		"Org":                      testConfig.VCD.Org,
		"Vdc":                      testConfig.VCD.Vdc,
		"name":                     name,
		"description":              "independent disk description",
		"secondName":               name + "second",
		"size":                     "5000",
		"busType":                  "SCSI",
		"busSubType":               "lsilogicsas",
		"storageProfileName":       testConfig.VCD.ProviderVdc.StorageProfile,
		"ResourceName":             resourceName,
		"secondResourceName":       resourceNameSecond,
		"thirdResourceName":        resourceNameThird,
		"Tags":                     "disk",
		"descriptionUpdate":        "independent disk description updated",
		"sizeUpdate":               "6000",
		"storageProfileNameUpdate": testConfig.VCD.ProviderVdc.StorageProfile2,
		"busTypeNvme":              "NVME",
		"busSubTypeNvme":           "nvmecontroller",
		"VmName":                   t.Name(),
		"Catalog":                  testSuiteCatalogName,
		"CatalogItem":              testSuiteCatalogOVAItem,
	}

	// regexp for empty value
	uuidMatchRegexp := regexp.MustCompile(`^$`)
	vcdClient := createTemporaryVCDConnection(true)
	if vcdClient != nil && vcdClient.Client.APIVCDMaxVersionIs(">= 36") {
		// from 36.0 API version value is returned
		uuidMatchRegexp = regexp.MustCompile(`^\S+`)
	}

	params["FuncName"] = t.Name() + "-Compatibility"
	configTextForCompatibility := templateFill(testAccCheckVcdIndependentDiskForCompatibility, params)
	params["FuncName"] = t.Name() + "-WithoutOptionals"
	configTextWithoutOptionals := templateFill(testAccCheckVcdIndependentDiskWithoutOptionals, params)
	params["FuncName"] = t.Name() + "-Update"
	configTextForUpdate := templateFill(testAccCheckVcdIndependentDiskForUpdate, params)
	params["FuncName"] = t.Name() + "-Nvme"
	configTextNvme := templateFill(testAccCheckVcdIndependentDiskNvmeType, params)
	params["FuncName"] = t.Name() + "-NvmeUpdate"
	configTextNvmeUpdate := templateFill(testAccCheckVcdIndependentDiskNvmeTypeUpdate, params)
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
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testDiskResourcesDestroyed,
		Steps: []resource.TestStep{
			resource.TestStep{
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
					resource.TestMatchResourceAttr("vcd_independent_disk."+resourceName, "uuid", uuidMatchRegexp),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "sharing_type", "None"),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "encrypted", "false"),
				),
			},
			resource.TestStep{
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
					resource.TestMatchResourceAttr("vcd_independent_disk."+resourceName, "uuid", uuidMatchRegexp),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "sharing_type", "None"),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "encrypted", "false"),
				),
			},
			resource.TestStep{
				ResourceName:            "vcd_independent_disk." + resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdFunc:       importStateIdByDisk("vcd_independent_disk." + resourceName),
				ImportStateVerifyIgnore: []string{"org", "vdc"},
			},
			resource.TestStep{
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
			resource.TestStep{
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
					resource.TestMatchResourceAttr("vcd_independent_disk."+resourceName, "uuid", uuidMatchRegexp),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "sharing_type", "None"),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "encrypted", "false"),
				),
			},
			resource.TestStep{
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
					resource.TestMatchResourceAttr("vcd_independent_disk."+resourceName, "uuid", uuidMatchRegexp),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "sharing_type", "None"),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "encrypted", "false"),
				),
			},
			resource.TestStep{
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
					resource.TestMatchResourceAttr("vcd_independent_disk."+resourceName, "uuid", uuidMatchRegexp),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "sharing_type", "None"),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "encrypted", "false"),

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
					resource.TestMatchResourceAttr("vcd_independent_disk."+resourceNameThird, "uuid", uuidMatchRegexp),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceNameThird, "sharing_type", "None"),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceNameThird, "encrypted", "false"),
				),
			},
			resource.TestStep{
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
					resource.TestMatchResourceAttr("vcd_independent_disk."+resourceName, "uuid", uuidMatchRegexp),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "sharing_type", "None"),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "encrypted", "false"),

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
					resource.TestMatchResourceAttr("vcd_independent_disk."+resourceNameThird, "uuid", uuidMatchRegexp),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceNameThird, "sharing_type", "None"),
					resource.TestCheckResourceAttr("vcd_independent_disk."+resourceName, "encrypted", "false"),
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

		_, vdc, err := conn.GetOrgAndVdc(testConfig.VCD.Org, testConfig.VCD.Vdc)
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

		_, vdc, err := conn.GetOrgAndVdc(testConfig.VCD.Org, testConfig.VCD.Vdc)
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

		importId := testConfig.VCD.Org + "." + testConfig.VCD.Vdc + "." + rs.Primary.ID
		if testConfig.VCD.Org == "" || testConfig.VCD.Vdc == "" || rs.Primary.ID == "" {
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
