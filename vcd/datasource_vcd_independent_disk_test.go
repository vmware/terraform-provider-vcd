//go:build disk || ALL || functional
// +build disk ALL functional

package vcd

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// Test independent disk data resource
// Using a disk data source we reference a disk data source
func TestAccVcdDataSourceIndependentDisk(t *testing.T) {
	preTestChecks(t)
	if !usingSysAdmin() {
		t.Skip("TestAccVcdDataSourceIndependentDisk requires system admin privileges")
	}
	resourceName := "TestAccVcdDataSourceIndependentDisk_1"
	datasourceName := "TestAccVcdDataSourceIndependentDisk_Data"
	datasourceNameWithId := "TestAccVcdDataSourceIndependentDiskWithId_Data"
	diskName := "TestAccVcdDataSourceIndependentDisk"

	var params = StringMap{
		"Vdc":                  testConfig.VCD.Vdc,
		"name":                 diskName,
		"description":          diskName + "description",
		"size":                 "52",
		"busType":              "SCSI",
		"busSubType":           "lsilogicsas",
		"storageProfileName":   "*",
		"ResourceName":         resourceName,
		"Tags":                 "disk",
		"dataSourceName":       datasourceName,
		"datasourceNameWithId": datasourceNameWithId,
		"metadataKey":          "key1",
		"metadataValue":        "value1",
	}

	// Updated parameters for step2
	var updateParams = make(StringMap)
	for k, v := range params {
		updateParams[k] = v
	}
	updateParams["metadataKey"] = "key2"
	updateParams["metadataValue"] = "value2"

	// regexp for empty value
	uuidMatchRegexp := regexp.MustCompile(`^$`)
	vcdClient := createTemporaryVCDConnection(true)
	sharingType := ""
	if vcdClient != nil && vcdClient.Client.APIVCDMaxVersionIs(">= 36") {
		// from 36.0 API version value is returned
		uuidMatchRegexp = regexp.MustCompile(`^\S+`)
		sharingType = "None"
	}

	params["FuncName"] = t.Name() + "-Step1"
	configText := templateFill(testAccCheckVcdDataSourceIndependentDisk, params)
	updateParams["FuncName"] = t.Name() + "-Step2"
	configText2 := templateFill(testAccCheckVcdDataSourceIndependentDisk, updateParams)
	params["FuncName"] = t.Name() + "-withId"
	configTextWithId := templateFill(testAccCheckVcdDataSourceIndependentDiskWithId, params)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testDiskResourcesDestroyed,
		Steps: []resource.TestStep{
			{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDiskCreated("vcd_independent_disk."+resourceName),
					resource.TestCheckResourceAttr("data.vcd_independent_disk."+datasourceName, "name", diskName),
					resource.TestCheckResourceAttr("data.vcd_independent_disk."+datasourceName, "description", diskName+"description"),
					resource.TestCheckResourceAttr("data.vcd_independent_disk."+datasourceName, "size_in_mb", params["size"].(string)),
					resource.TestCheckResourceAttr("data.vcd_independent_disk."+datasourceName, "bus_type", "SCSI"),
					resource.TestCheckResourceAttr("data.vcd_independent_disk."+datasourceName, "bus_sub_type", "lsilogicsas"),
					resource.TestCheckResourceAttr("data.vcd_independent_disk."+datasourceName, "storage_profile", "*"),
					resource.TestCheckResourceAttr("data.vcd_independent_disk."+datasourceName, "metadata."+params["metadataKey"].(string), params["metadataValue"].(string)),
					resource.TestMatchOutput("owner_name", regexp.MustCompile(`^\S+`)),
					resource.TestMatchOutput("datastore_name", regexp.MustCompile(`^\S+`)),
					resource.TestMatchOutput("uuid", uuidMatchRegexp),
					resource.TestCheckOutput("sharing_type", sharingType),
					resource.TestCheckOutput("encrypted", "false"),
					resource.TestCheckOutput("attached_vm_ids", "0"),
					testCheckDiskNonStringOutputs(),
				),
			},
			{
				Config: configText2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr("data.vcd_independent_disk."+datasourceName, "metadata."+params["metadataKey"].(string)),
					resource.TestCheckResourceAttr("data.vcd_independent_disk."+datasourceName, "metadata."+updateParams["metadataKey"].(string), updateParams["metadataValue"].(string)),
				),
			},
			{
				Config: configTextWithId,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDiskCreated("vcd_independent_disk."+resourceName),
					resource.TestCheckResourceAttr("data.vcd_independent_disk."+datasourceNameWithId, "name", diskName+"WithId"),
					resource.TestCheckResourceAttr("data.vcd_independent_disk."+datasourceNameWithId, "description", diskName+"description"),
					resource.TestCheckResourceAttr("data.vcd_independent_disk."+datasourceNameWithId, "size_in_mb", params["size"].(string)),
					resource.TestCheckResourceAttr("data.vcd_independent_disk."+datasourceNameWithId, "bus_type", "SCSI"),
					resource.TestCheckResourceAttr("data.vcd_independent_disk."+datasourceNameWithId, "bus_sub_type", "lsilogicsas"),
					resource.TestCheckResourceAttr("data.vcd_independent_disk."+datasourceNameWithId, "storage_profile", "*"),
					resource.TestMatchOutput("owner_name", regexp.MustCompile(`^\S+`)),
					resource.TestMatchOutput("datastore_name", regexp.MustCompile(`^\S+`)),
					resource.TestMatchOutput("uuid", uuidMatchRegexp),
					resource.TestCheckOutput("sharing_type", sharingType),
					resource.TestCheckOutput("encrypted", "false"),
					resource.TestCheckOutput("attached_vm_ids", "0"),
					testCheckDiskNonStringOutputs(),
				),
			},
		},
	})
	postTestChecks(t)
}

func testCheckDiskNonStringOutputs() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		outputs := s.RootModule().Outputs

		if outputs["is_attached"].Value.(string) != "false" {
			return fmt.Errorf("is_attached value didn't match")
		}

		iops := outputs["iops"].Value.(string)
		reNumber := regexp.MustCompile(`^\d+$`)
		if !reNumber.MatchString(iops) {
			return fmt.Errorf("iops value isn't an integer")
		}
		return nil
	}
}

const testAccCheckVcdDataSourceIndependentDisk = `
resource "vcd_independent_disk" "{{.ResourceName}}" {
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

data "vcd_independent_disk" "{{.dataSourceName}}" {
  name       = vcd_independent_disk.{{.ResourceName}}.name
}

output "iops" {
  value = data.vcd_independent_disk.{{.dataSourceName}}.iops
}
output "owner_name" {
  value = data.vcd_independent_disk.{{.dataSourceName}}.owner_name
}
output "datastore_name" {
  value = data.vcd_independent_disk.{{.dataSourceName}}.datastore_name
}
output "is_attached" {
  value = data.vcd_independent_disk.{{.dataSourceName}}.is_attached
}
output "encrypted" {
  value = data.vcd_independent_disk.{{.dataSourceName}}.encrypted
}
output "sharing_type" {
  value = data.vcd_independent_disk.{{.dataSourceName}}.sharing_type
}
output "uuid" {
  value = data.vcd_independent_disk.{{.dataSourceName}}.uuid
}
output "attached_vm_ids" {
  value = length(tolist(data.vcd_independent_disk.{{.dataSourceName}}.attached_vm_ids))
}
`

const testAccCheckVcdDataSourceIndependentDiskWithId = `
resource "vcd_independent_disk" "{{.ResourceName}}" {
  vdc             = "{{.Vdc}}"
  name            = "{{.name}}WithId"
  description     = "{{.description}}"
  size_in_mb      = "{{.size}}"
  bus_type        = "{{.busType}}"
  bus_sub_type    = "{{.busSubType}}"
  storage_profile = "{{.storageProfileName}}"
}

data "vcd_independent_disk" "{{.datasourceNameWithId}}" {
  id         = vcd_independent_disk.{{.ResourceName}}.id
}

output "iops" {
  value = data.vcd_independent_disk.{{.datasourceNameWithId}}.iops
}
output "owner_name" {
  value = data.vcd_independent_disk.{{.datasourceNameWithId}}.owner_name
}
output "datastore_name" {
  value = data.vcd_independent_disk.{{.datasourceNameWithId}}.datastore_name
}
output "is_attached" {
  value = data.vcd_independent_disk.{{.datasourceNameWithId}}.is_attached
}
output "encrypted" {
  value = data.vcd_independent_disk.{{.datasourceNameWithId}}.encrypted
}
output "sharing_type" {
  value = data.vcd_independent_disk.{{.datasourceNameWithId}}.sharing_type
}
output "uuid" {
  value = data.vcd_independent_disk.{{.datasourceNameWithId}}.uuid
}
output "attached_vm_ids" {
  value = length(tolist(data.vcd_independent_disk.{{.datasourceNameWithId}}.attached_vm_ids))
}
`
