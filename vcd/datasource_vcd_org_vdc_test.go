// +build vdc ALL functional

package vcd

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	"regexp"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccVcdVdcDatasource(t *testing.T) {
	validateConfiguration(t)

	vdcName := TestAccVcdVdc + "ForDataSourceTest"

	var params = StringMap{
		"ExistingVdcName": testConfig.VCD.Vdc,
		"VdcName":         vdcName,
		"OrgName":         testConfig.VCD.Org,
		"FuncName":        "TestAccVcdVdcDatasource",
	}

	if !usingSysAdmin() {
		t.Skip("TestAccVcdVdcDatasource requires system admin privileges")
		return
	}

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	configText := templateFill(testAccCheckVcdVdcDatasource_basic, params)

	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

	datasourceVdc := "vcd_org_vdc.existingVdc"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { preRunChecks(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVdcDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
				// Dont' handle very well Sets or computed fields
				ExpectError: regexp.MustCompile(`After applying this step, the plan was not empty`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdVdcExists("vcd_org_vdc."+vdcName),
					resource.TestCheckResourceAttrPair("data."+datasourceVdc, "org", "vcd_org_vdc."+vdcName, "org"),
					resource.TestCheckResourceAttrPair("data."+datasourceVdc, "allocation_model", "vcd_org_vdc."+vdcName, "allocation_model"),
					resource.TestCheckResourceAttrPair("data."+datasourceVdc, "network_pool_name", "vcd_org_vdc."+vdcName, "network_pool_name"),
					resource.TestCheckResourceAttrPair("data."+datasourceVdc, "provider_vdc_name", "vcd_org_vdc."+vdcName, "provider_vdc_name"),
					resource.TestCheckResourceAttrPair("data."+datasourceVdc, "enabled", "vcd_org_vdc."+vdcName, "enabled"),
					resource.TestCheckResourceAttrPair("data."+datasourceVdc, "enable_thin_provisioning", "vcd_org_vdc."+vdcName, "enable_thin_provisioning"),
					resource.TestCheckResourceAttrPair("data."+datasourceVdc, "storage_profile.0.enabled", "vcd_org_vdc."+vdcName, "storage_profile.0.enabled"),
					resource.TestCheckResourceAttrPair("data."+datasourceVdc, "storage_profile.0.default", "vcd_org_vdc."+vdcName, "storage_profile.0.default"),
					resource.TestCheckResourceAttr("vcd_org_vdc."+vdcName, "metadata.vdc_metadata", "VDC Metadata"),
					resource.ComposeTestCheckFunc(testAccDataSourceVcdOrgVdc("data."+datasourceVdc, vdcName)),
				),
			},
		},
	})
}

func convertIntWithErrIgnore(value string) int {
	if n, err := strconv.Atoi(value); err == nil {
		return n
	} else {
		fmt.Println(value, "is not an integer.")
	}
	return -1
}

func testAccDataSourceVcdOrgVdc(name, vdcName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		resources, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("root module has no resource called %s", name)
		}

		vdcResource, ok := s.RootModule().Resources["vcd_org_vdc."+vdcName]
		if !ok {
			return fmt.Errorf("can't find vcd_org_vdc.%s in state", vdcName)
		}

		attr := resources.Primary.Attributes

		internalMapOfCpuValues := map[string]interface{}{"allocated": convertIntWithErrIgnore(attr["compute_capacity.0.cpu.0.allocated"]), "limit": convertIntWithErrIgnore(attr["compute_capacity.0.cpu.0.limit"]),
			"overhead": convertIntWithErrIgnore(attr["compute_capacity.0.cpu.0.overhead"]), "reserved": convertIntWithErrIgnore(attr["compute_capacity.0.cpu.0.reserved"]), "used": convertIntWithErrIgnore(attr["compute_capacity.0.cpu.0.used"])}
		cpuHashInternalValue := hashMapStringForCapacityElements(internalMapOfCpuValues)

		internalMapOfMemoryValues := map[string]interface{}{"allocated": convertIntWithErrIgnore(attr["compute_capacity.0.memory.0.allocated"]), "limit": convertIntWithErrIgnore(attr["compute_capacity.0.memory.0.limit"]),
			"overhead": convertIntWithErrIgnore(attr["compute_capacity.0.memory.0.overhead"]), "reserved": convertIntWithErrIgnore(attr["compute_capacity.0.memory.0.reserved"]), "used": convertIntWithErrIgnore(attr["compute_capacity.0.memory.0.used"])}
		memoryHashInternalValue := hashMapStringForCapacityElements(internalMapOfMemoryValues)

		memoryCapacityArray := make([]interface{}, 0)
		memoryCapacityArray = append(memoryCapacityArray, internalMapOfMemoryValues)
		cpuCapacityArray := make([]interface{}, 0)
		cpuCapacityArray = append(cpuCapacityArray, internalMapOfCpuValues)

		cpu := *schema.NewSet(hashMapStringForCapacityElements, cpuCapacityArray)
		memory := *schema.NewSet(hashMapStringForCapacityElements, memoryCapacityArray)

		mainHashValue := hashMapStringForCapacity(map[string]interface{}{"cpu": &cpu, "memory": &memory})

		if attr["compute_capacity.0.cpu.0.allocated"] != vdcResource.Primary.Attributes[fmt.Sprintf("compute_capacity.%d.cpu.%d.allocated", mainHashValue, cpuHashInternalValue)] {
			return fmt.Errorf("compute_capacity.0.cpu.0.allocated is %#v; want %#v", attr["compute_capacity.0.cpu.0.allocated"], vdcResource.Primary.Attributes[fmt.Sprintf("compute_capacity.%d.cpu.%d.allocated", mainHashValue, cpuHashInternalValue)])
		}

		if attr["compute_capacity.0.cpu.0.limit"] != vdcResource.Primary.Attributes[fmt.Sprintf("compute_capacity.%d.cpu.%d.limit", mainHashValue, cpuHashInternalValue)] {
			return fmt.Errorf("compute_capacity.0.cpu.0.limit is %#v; want %#v", attr["compute_capacity.0.cpu.0.limit"], vdcResource.Primary.Attributes[fmt.Sprintf("compute_capacity.%d.cpu.%d.limit", mainHashValue, cpuHashInternalValue)])
		}

		if attr["compute_capacity.0.cpu.0.overhead"] != vdcResource.Primary.Attributes[fmt.Sprintf("compute_capacity.%d.cpu.%d.overhead", mainHashValue, cpuHashInternalValue)] {
			return fmt.Errorf("compute_capacity.0.cpu.0.overhead is %#v; want %#v", attr["compute_capacity.0.cpu.0.overhead"], vdcResource.Primary.Attributes[fmt.Sprintf("compute_capacity.%d.cpu.%d.overhead", mainHashValue, cpuHashInternalValue)])
		}

		if attr["compute_capacity.0.cpu.0.reserved"] != vdcResource.Primary.Attributes[fmt.Sprintf("compute_capacity.%d.cpu.%d.reserved", mainHashValue, cpuHashInternalValue)] {
			return fmt.Errorf("compute_capacity.0.cpu.0.reserved is %#v; want %#v", attr["compute_capacity.0.cpu.0.reserved"], vdcResource.Primary.Attributes[fmt.Sprintf("compute_capacity.%d.cpu.%d.reserved", mainHashValue, cpuHashInternalValue)])
		}

		if attr["compute_capacity.0.cpu.0.used"] != vdcResource.Primary.Attributes[fmt.Sprintf("compute_capacity.%d.cpu.%d.used", mainHashValue, cpuHashInternalValue)] {
			return fmt.Errorf("compute_capacity.0.cpu.0.used is %#v; want %#v", attr["compute_capacity.0.cpu.0.used"], vdcResource.Primary.Attributes[fmt.Sprintf("compute_capacity.%d.cpu.%d.used", mainHashValue, cpuHashInternalValue)])
		}

		if attr["compute_capacity.0.memory.0.allocated"] != vdcResource.Primary.Attributes[fmt.Sprintf("compute_capacity.%d.memory.%d.allocated", mainHashValue, memoryHashInternalValue)] {
			return fmt.Errorf("compute_capacity.0.memory.0.allocated is %#v; want %#v", attr["compute_capacity.0.memory.0.allocated"], vdcResource.Primary.Attributes[fmt.Sprintf("compute_capacity.%d.memory.%d.allocated", mainHashValue, memoryHashInternalValue)])
		}

		if attr["compute_capacity.0.memory.0.limit"] != vdcResource.Primary.Attributes[fmt.Sprintf("compute_capacity.%d.memory.%d.limit", mainHashValue, memoryHashInternalValue)] {
			return fmt.Errorf("compute_capacity.0.memory.0.limit is %#v; want %#v", attr["compute_capacity.0.memory.0.limit"], vdcResource.Primary.Attributes[fmt.Sprintf("compute_capacity.%d.memory.%d.limit", mainHashValue, memoryHashInternalValue)])
		}

		if attr["compute_capacity.0.memory.0.overhead"] != vdcResource.Primary.Attributes[fmt.Sprintf("compute_capacity.%d.memory.%d.overhead", mainHashValue, memoryHashInternalValue)] {
			return fmt.Errorf("compute_capacity.0.memory.0.overhead is %#v; want %#v", attr["compute_capacity.0.memory.0.overhead"], vdcResource.Primary.Attributes[fmt.Sprintf("compute_capacity.%d.memory.%d.overhead", mainHashValue, memoryHashInternalValue)])
		}

		if attr["compute_capacity.0.memory.0.reserved"] != vdcResource.Primary.Attributes[fmt.Sprintf("compute_capacity.%d.memory.%d.reserved", mainHashValue, memoryHashInternalValue)] {
			return fmt.Errorf("compute_capacity.0.memory.0.reserved is %#v; want %#v", attr["compute_capacity.0.memory.0.reserved"], vdcResource.Primary.Attributes[fmt.Sprintf("compute_capacity.%d.memory.%d.reserved", mainHashValue, memoryHashInternalValue)])
		}

		if attr["compute_capacity.0.memory.0.used"] != vdcResource.Primary.Attributes[fmt.Sprintf("compute_capacity.%d.memory.%d.used", mainHashValue, memoryHashInternalValue)] {
			return fmt.Errorf("compute_capacity.0.memory.0.used is %#v; want %#v", attr["compute_capacity.0.memory.0.used"], vdcResource.Primary.Attributes[fmt.Sprintf("compute_capacity.%d.memory.%d.used", mainHashValue, memoryHashInternalValue)])
		}

		return nil
	}
}

const testAccCheckVcdVdcDatasource_basic = `
data "vcd_org_vdc" "existingVdc" {
  org  = "{{.OrgName}}"
  name = "{{.ExistingVdcName}}"
}

resource "vcd_org_vdc" "{{.VdcName}}" {
  name = "{{.VdcName}}"
  org  = "{{.OrgName}}"

  allocation_model  = "${data.vcd_org_vdc.existingVdc.allocation_model}"
  network_pool_name = "${data.vcd_org_vdc.existingVdc.network_pool_name}"
  provider_vdc_name = "${data.vcd_org_vdc.existingVdc.provider_vdc_name}"

  compute_capacity {
    cpu {
     allocated = "${tolist(tolist(data.vcd_org_vdc.existingVdc.compute_capacity)[0].cpu)[0].allocated}"
     limit     = "${tolist(tolist(data.vcd_org_vdc.existingVdc.compute_capacity)[0].cpu)[0].limit}"
    }

    memory {
     allocated = "${tolist(tolist(data.vcd_org_vdc.existingVdc.compute_capacity)[0].memory)[0].allocated}"
     limit     = "${tolist(tolist(data.vcd_org_vdc.existingVdc.compute_capacity)[0].memory)[0].limit}"
    }
  }

  storage_profile {
    name    = "${data.vcd_org_vdc.existingVdc.storage_profile[0].name}"
    enabled = "${data.vcd_org_vdc.existingVdc.storage_profile[0].enabled}"
    limit   = "${data.vcd_org_vdc.existingVdc.storage_profile[0].limit}"
    default = "${data.vcd_org_vdc.existingVdc.storage_profile[0].default}"
  }

  metadata = {
    vdc_metadata = "VDC Metadata"
  }

  enabled                  = "${data.vcd_org_vdc.existingVdc.enabled}"
  enable_thin_provisioning = "${data.vcd_org_vdc.existingVdc.enable_thin_provisioning}"
  enable_fast_provisioning = "${data.vcd_org_vdc.existingVdc.enable_fast_provisioning}"
  delete_force             = true
  delete_recursive         = true
}
`
