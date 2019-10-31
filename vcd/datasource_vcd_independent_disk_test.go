// +build disk ALL functional

package vcd

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

// Test independent disk data resource
// Using a disk data source we reference a disk data source
func TestAccVcdDataSourceIndependentDisk(t *testing.T) {
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
	}

	configText := templateFill(testAccCheckVcdDataSourceIndependentDisk, params)
	params["FuncName"] = t.Name() + "-withId"
	configTextWithId := templateFill(testAccCheckVcdDataSourceIndependentDiskWithId, params)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testDiskResourcesDestroyed,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDiskCreated("vcd_independent_disk."+resourceName),
					resource.TestCheckResourceAttr("data.vcd_independent_disk."+datasourceName, "name", diskName),
					resource.TestCheckResourceAttr("data.vcd_independent_disk."+datasourceName, "description", diskName+"description"),
					//resource.TestCheckResourceAttr("data.vcd_independent_disk."+datasourceName, "size_in_bytes", "5242880"),
					resource.TestCheckResourceAttr("data.vcd_independent_disk."+datasourceName, "bus_type", "SCSI"),
					resource.TestCheckResourceAttr("data.vcd_independent_disk."+datasourceName, "bus_sub_type", "lsilogicsas"),
					resource.TestCheckResourceAttr("data.vcd_independent_disk."+datasourceName, "storage_profile", "*"),
					resource.TestMatchOutput("owner_name", regexp.MustCompile(`^\S+`)),
					resource.TestMatchOutput("datastore_name", regexp.MustCompile(`^\S+`)),
					testCheckDiskNonStringOutputs(),
				),
			},
			resource.TestStep{
				Config: configTextWithId,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDiskCreated("vcd_independent_disk."+resourceName),
					resource.TestCheckResourceAttr("data.vcd_independent_disk."+datasourceNameWithId, "name", diskName+"WithId"),
					resource.TestCheckResourceAttr("data.vcd_independent_disk."+datasourceNameWithId, "description", diskName+"description"),
					//resource.TestCheckResourceAttr("data.vcd_independent_disk."+datasourceNameWithId, "size_in_bytes", "5242880"),
					resource.TestCheckResourceAttr("data.vcd_independent_disk."+datasourceNameWithId, "bus_type", "SCSI"),
					resource.TestCheckResourceAttr("data.vcd_independent_disk."+datasourceNameWithId, "bus_sub_type", "lsilogicsas"),
					resource.TestCheckResourceAttr("data.vcd_independent_disk."+datasourceNameWithId, "storage_profile", "*"),
					resource.TestMatchOutput("owner_name", regexp.MustCompile(`^\S+`)),
					resource.TestMatchOutput("datastore_name", regexp.MustCompile(`^\S+`)),
					testCheckDiskNonStringOutputs(),
				),
			},
		},
	})
}

func testCheckDiskNonStringOutputs() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		outputs := s.RootModule().Outputs

		if outputs["is_attached"].Value.(bool) != false {
			return fmt.Errorf("is_attached value didn't match")
		}

		if regexp.MustCompile(`^\d+$`).MatchString(fmt.Sprintf("%s", outputs["iops"].Value)) {
			return fmt.Errorf("iops value isn't int")
		}

		return nil
	}
}

const testAccCheckVcdDataSourceIndependentDisk = `
resource "vcd_independent_disk" "{{.ResourceName}}" {
  vdc             = "{{.Vdc}}"
  name            = "{{.name}}"
  description     = "{{.description}}"
  size            = "{{.size}}" 
  #size_in_bytes   = "{{.size}}"
  bus_type        = "{{.busType}}"
  bus_sub_type    = "{{.busSubType}}"
  storage_profile = "{{.storageProfileName}}"
}

data "vcd_independent_disk" "{{.dataSourceName}}" {
  name    = "${vcd_independent_disk.{{.ResourceName}}.name}" 
}

output "iops" {
  value = data.vcd_independent_disk.{{.dataSourceName}}.iops
  depends_on = ["data.vcd_independent_disk.{{.dataSourceName}}"]
}
output "owner_name" {
  value = data.vcd_independent_disk.{{.dataSourceName}}.owner_name
  depends_on = [data.vcd_independent_disk.{{.dataSourceName}}]
}
output "datastore_name" {
  value = data.vcd_independent_disk.{{.dataSourceName}}.datastore_name
  depends_on = [data.vcd_independent_disk.{{.dataSourceName}}]
}
output "is_attached" {
  value = data.vcd_independent_disk.{{.dataSourceName}}.is_attached
  depends_on = [data.vcd_independent_disk.{{.dataSourceName}}]
}
`

const testAccCheckVcdDataSourceIndependentDiskWithId = `
resource "vcd_independent_disk" "{{.ResourceName}}" {
  vdc             = "{{.Vdc}}"
  name            = "{{.name}}WithId"
  description     = "{{.description}}"
  size            = "{{.size}}"
  #size_in_bytes  = "{{.size}}"
  bus_type        = "{{.busType}}"
  bus_sub_type    = "{{.busSubType}}"
  storage_profile = "{{.storageProfileName}}"
}

data "vcd_independent_disk" "{{.datasourceNameWithId}}" {
  id         = "${vcd_independent_disk.{{.ResourceName}}.id}"
}

output "iops" {
  value      = data.vcd_independent_disk.{{.datasourceNameWithId}}.iops
  depends_on = ["data.vcd_independent_disk.{{.datasourceNameWithId}}"]
}
output "owner_name" {
  value      = data.vcd_independent_disk.{{.datasourceNameWithId}}.owner_name
  depends_on = [data.vcd_independent_disk.{{.datasourceNameWithId}}]
}
output "datastore_name" {
  value      = data.vcd_independent_disk.{{.datasourceNameWithId}}.datastore_name
  depends_on = [data.vcd_independent_disk.{{.datasourceNameWithId}}]
}
output "is_attached" {
  value      = data.vcd_independent_disk.{{.datasourceNameWithId}}.is_attached
  depends_on = [data.vcd_independent_disk.{{.datasourceNameWithId}}]
}
`
