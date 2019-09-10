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
					resource.TestCheckResourceAttrPair("data."+datasourceVdc, "name", "vcd_org_vdc."+vdcName, "name"),
					resource.TestCheckResourceAttrPair("data."+datasourceVdc, "root_resource_id", "vcd_org_vdc."+vdcName, "root_resource_id"),
					resource.TestCheckResourceAttrPair("data."+datasourceVdc, "org", "vcd_org_vdc."+vdcName, "org"),
					resource.TestCheckResourceAttrPair("data."+datasourceVdc, "allocation_model", "vcd_org_vdc."+vdcName, "allocation_model"),
					resource.TestCheckResourceAttrPair("data."+datasourceVdc, "network_pool_name", "vcd_org_vdc."+vdcName, "network_pool_name"),
					resource.TestCheckResourceAttrPair("data."+datasourceVdc, "provider_vdc_name", "vcd_org_vdc."+vdcName, "provider_vdc_name"),
					resource.TestCheckResourceAttrPair("data."+datasourceVdc, "enabled", "vcd_org_vdc."+vdcName, "enabled"),
					resource.TestCheckResourceAttrPair("data."+datasourceVdc, "enable_thin_provisioning", "vcd_org_vdc."+vdcName, "enable_thin_provisioning"),
					resource.TestCheckResourceAttrPair("data."+datasourceVdc, "storage_profile.0.enabled", "vcd_org_vdc."+vdcName, "storage_profile.0.enabled"),
					resource.TestCheckResourceAttrPair("data."+datasourceVdc, "storage_profile.0.default", "vcd_org_vdc."+vdcName, "storage_profile.0.default"),
					resource.TestCheckResourceAttrPair("data."+datasourceVdc, "metadata.vdc_metadata", "vcd_org_vdc."+vdcName, "metadata.vdc_metadata"),
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
