// +build vdc ALL functional

package vcd

import (
	"fmt"
	"github.com/hashicorp/terraform/terraform"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccVcdVdcDatasource(t *testing.T) {
	validateConfiguration(t)

	allocationModel := "ReservationPool"
	vdcName := TestAccVcdVdc + "ForDataSourceTest"

	var params = StringMap{
		"VdcName":                   vdcName,
		"OrgName":                   testConfig.VCD.Org,
		"AllocationModel":           allocationModel,
		"ProviderVdc":               testConfig.VCD.ProviderVdc.Name,
		"NetworkPool":               testConfig.VCD.ProviderVdc.NetworkPool,
		"ProviderVdcStorageProfile": testConfig.VCD.ProviderVdc.StorageProfile,
		"Tags":                      "vdc",
		"FuncName":                  "TestAccVcdOrgVdcReservationPool",
		// cause vDC ignores empty values and use default
		"MemoryGuaranteed": "1",
		"CpuGuaranteed":    "1",
	}

	if !usingSysAdmin() {
		t.Skip("TestAccVcdVdcDatasource requires system admin privileges")
		return
	}

	params["FuncName"] = t.Name() + "-Datasource"
	configText := templateFill(testAccCheckVcdVdcDatasource_basic, params)

	debugPrintf("#[DEBUG] CONFIGURATION: %s", configText)

	datasourceVdc := "vcd_org_vdc.asDatasource"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { preRunChecks(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVdcDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config:      configText,
				ExpectError: regexp.MustCompile(`After applying this step and refreshing, the plan was not empty`),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVcdVdcExists("vcd_org_vdc."+vdcName),
					resource.ComposeTestCheckFunc(testAccDataSourceVcdOrgVdc("data."+datasourceVdc, vdcName)),
				),
			},
			resource.TestStep{
				ResourceName:      "vcd_org_vdc." + vdcName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdByVdc(vdcName),
				// These fields can't be retrieved
				ImportStateVerifyIgnore: []string{"delete_force", "delete_recursive"},
			},
		},
	})
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

		if attr["name"] != vdcResource.Primary.Attributes["name"] {
			return fmt.Errorf("name is %s; want %s", attr["name"], vdcResource.Primary.Attributes["name"])
		}

		if attr["root_resource_id"] != vdcResource.Primary.Attributes["root_resource_id"] {
			return fmt.Errorf("root_resource_id is %s; want %s", attr["root_resource_id"], vdcResource.Primary.Attributes["root_resource_id"])
		}

		if attr["org"] != vdcResource.Primary.Attributes["org"] {
			return fmt.Errorf("org is %s; want %s", attr["org"], vdcResource.Primary.Attributes["org"])
		}

		if attr["allocation_model"] != vdcResource.Primary.Attributes["allocation_model"] {
			return fmt.Errorf("allocation_model is %s; want %s", attr["allocation_model"], vdcResource.Primary.Attributes["allocation_model"])
		}

		if attr["network_pool_name"] != vdcResource.Primary.Attributes["network_pool_name"] {
			return fmt.Errorf("network_pool_name is %s; want %s", attr["network_pool_name"], vdcResource.Primary.Attributes["network_pool_name"])
		}

		if attr["provider_vdc_name"] != vdcResource.Primary.Attributes["provider_vdc_name"] {
			return fmt.Errorf("provider_vdc_name is %s; want %s", attr["provider_vdc_name"], vdcResource.Primary.Attributes["provider_vdc_name"])
		}

		if attr["enabled"] != vdcResource.Primary.Attributes["enabled"] {
			return fmt.Errorf("enabled is %s; want %s", attr["enabled"], vdcResource.Primary.Attributes["enabled"])
		}

		if attr["enable_thin_provisioning"] != vdcResource.Primary.Attributes["enable_thin_provisioning"] {
			return fmt.Errorf("enable_thin_provisioning is %s; want %s", attr["enable_thin_provisioning"], vdcResource.Primary.Attributes["enable_thin_provisioning"])
		}

		if attr["enable_fast_provisioning"] != vdcResource.Primary.Attributes["enable_fast_provisioning"] {
			return fmt.Errorf("enable_fast_provisioning is %s; want %s", attr["enable_fast_provisioning"], vdcResource.Primary.Attributes["enable_fast_provisioning"])
		}

		if attr["compute_capacity.0.cpu.0.allocated"] == vdcResource.Primary.Attributes["compute_capacity.0.cpu.0.allocated"] {
			return fmt.Errorf("compute_capacity.0.cpu.0.allocated is %#v; want %#v", attr["compute_capacity.0.cpu.0.allocated"], vdcResource.Primary.Attributes["compute_capacity.0.cpu.0.allocated"])
		}

		if attr["compute_capacity.0.cpu.0.limit"] == vdcResource.Primary.Attributes["compute_capacity.0.cpu.0.limit"] {
			return fmt.Errorf("compute_capacity.0.cpu.0.limit is %#v; want %#v", attr["compute_capacity.0.cpu.0.limit"], vdcResource.Primary.Attributes["compute_capacity.0.cpu.0.limit"])
		}

		if attr["compute_capacity.0.cpu.0.overhead"] == vdcResource.Primary.Attributes["compute_capacity.0.cpu.0.overhead"] {
			return fmt.Errorf("compute_capacity.0.cpu.0.overhead is %#v; want %#v", attr["compute_capacity.0.cpu.0.overhead"], vdcResource.Primary.Attributes["compute_capacity.0.cpu.0.overhead"])
		}

		if attr["compute_capacity.0.cpu.0.reserved"] == vdcResource.Primary.Attributes["compute_capacity.0.cpu.0.reserved"] {
			return fmt.Errorf("compute_capacity.0.cpu.0.reserved is %#v; want %#v", attr["compute_capacity.0.cpu.0.reserved"], vdcResource.Primary.Attributes["compute_capacity.0.cpu.0.reserved"])
		}

		if attr["compute_capacity.0.cpu.0.used"] == vdcResource.Primary.Attributes["compute_capacity.0.cpu.0.used"] {
			return fmt.Errorf("compute_capacity.0.cpu.0.used is %#v; want %#v", attr["compute_capacity.0.cpu.0.used"], vdcResource.Primary.Attributes["compute_capacity.0.cpu.0.used"])
		}

		if attr["compute_capacity.0.memory.0.allocated"] == vdcResource.Primary.Attributes["compute_capacity.0.memory.0.allocated"] {
			return fmt.Errorf("compute_capacity.0.memory.0.allocated is %#v; want %#v", attr["compute_capacity.0.memory.0.allocated"], vdcResource.Primary.Attributes["compute_capacity.0.memory.0.allocated"])
		}

		if attr["compute_capacity.0.memory.0.limit"] == vdcResource.Primary.Attributes["compute_capacity.0.memory.0.limit"] {
			return fmt.Errorf("compute_capacity.0.memory.0.limit is %#v; want %#v", attr["compute_capacity.0.memory.0.limit"], vdcResource.Primary.Attributes["compute_capacity.0.memory.0.limit"])
		}

		if attr["compute_capacity.0.memory.0.overhead"] == vdcResource.Primary.Attributes["compute_capacity.0.memory.0.overhead"] {
			return fmt.Errorf("compute_capacity.0.memory.0.overhead is %#v; want %#v", attr["compute_capacity.0.memory.0.overhead"], vdcResource.Primary.Attributes["compute_capacity.0.memory.0.overhead"])
		}

		if attr["compute_capacity.0.memory.0.reserved"] == vdcResource.Primary.Attributes["compute_capacity.0.memory.0.reserved"] {
			return fmt.Errorf("compute_capacity.0.memory.0.reserved is %#v; want %#v", attr["compute_capacity.0.memory.0.reserved"], vdcResource.Primary.Attributes["compute_capacity.0.memory.0.reserved"])
		}

		if attr["compute_capacity.0.memory.0.used"] == vdcResource.Primary.Attributes["compute_capacity.0.memory.0.used"] {
			return fmt.Errorf("compute_capacity.0.memory.0.used is %#v; want %#v", attr["compute_capacity.0.memory.0.used"], vdcResource.Primary.Attributes["compute_capacity.0.memory.0.used"])
		}

		if attr["storage_profile.0.name"] != vdcResource.Primary.Attributes["storage_profile.0.name"] {
			return fmt.Errorf("storage_profile.0.name is %s; want %s", attr["storage_profile.0.name"], vdcResource.Primary.Attributes["storage_profile.0.name"])
		}

		if attr["storage_profile.0.limit"] != vdcResource.Primary.Attributes["storage_profile.0.limit"] {
			return fmt.Errorf("storage_profile.0.limit is %s; want %s", attr["storage_profile.0.limit"], vdcResource.Primary.Attributes["storage_profile.0.limit"])
		}

		if attr["storage_profile.0.enabled"] != vdcResource.Primary.Attributes["storage_profile.0.enabled"] {
			return fmt.Errorf("storage_profile.0.enabled is %s; want %s", attr["storage_profile.0.enabled"], vdcResource.Primary.Attributes["storage_profile.0.enabled"])
		}

		if attr["storage_profile.0.default"] != vdcResource.Primary.Attributes["storage_profile.0.default"] {
			return fmt.Errorf("storage_profile.0.default is %s; want %s", attr["storage_profile.0.default"], vdcResource.Primary.Attributes["storage_profile.0.default"])
		}

		if attr["metadata.vdc_metadata"] != vdcResource.Primary.Attributes["metadata.vdc_metadata"] {
			return fmt.Errorf("metadata.vdc_metadata is %s; want %s", attr["metadata.vdc_metadata"], vdcResource.Primary.Attributes["metadata.vdc_metadata"])
		}

		return nil
	}
}

func importStateIdByVdc(objectName string) resource.ImportStateIdFunc {
	return func(*terraform.State) (string, error) {
		importId := testConfig.VCD.Org + "." + objectName
		if testConfig.VCD.Org == "" || objectName == "" {
			return "", fmt.Errorf("missing information to generate import path: %s", importId)
		}
		return importId, nil
	}
}

const testAccCheckVcdVdcDatasource_basic = `
resource "vcd_org_vdc" "{{.VdcName}}" {
  name = "{{.VdcName}}"
  org  = "{{.OrgName}}"

  allocation_model = "{{.AllocationModel}}"
  network_pool_name     = "{{.NetworkPool}}"
  provider_vdc_name     = "{{.ProviderVdc}}"

  compute_capacity {
    cpu {
      allocated = 2048
      limit     = 2048
    }

    memory {
      allocated = 2048
      limit     = 2048
    }
  }

  storage_profile {
    name = "{{.ProviderVdcStorageProfile}}"
    enabled  = true
    limit    = 10240
    default  = true
  }

  metadata = {
    vdc_metadata = "VDC Metadata"
  }

  enabled                  = true
  enable_thin_provisioning = true
  enable_fast_provisioning = true
  delete_force             = true
  delete_recursive         = true
}

data "vcd_org_vdc" "asDatasource"{
  name = "{{.VdcName}}"
  org  = "{{.OrgName}}"
  depends_on      = ["vcd_org_vdc.{{.VdcName}}"]
}

`
