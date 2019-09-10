// +build vdc ALL functional

package vcd

import (
	"fmt"
	"github.com/hashicorp/terraform/terraform"
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

	datasourceVdc := "vcd_org_vdc.asDatasource"
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { preRunChecks(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVdcDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: configText,
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

const testAccCheckVcdVdcDatasource_basic = `
data "vcd_org_vdc" "asDatasource" {
  org  = "{{.OrgName}}"
  name = "{{.ExistingVdcName}}"
}

resource "vcd_org_vdc" "{{.VdcName}}" {
  name = "{{.VdcName}}"
  org  = "{{.OrgName}}"

  allocation_model  = "${data.vcd_org_vdc.asDatasource.allocation_model}"
  network_pool_name = "${data.vcd_org_vdc.asDatasource.network_pool_name}"
  provider_vdc_name = "${data.vcd_org_vdc.asDatasource.provider_vdc_name}"

  compute_capacity {
    cpu {
     allocated = "${tolist(tolist(data.vcd_org_vdc.asDatasource.compute_capacity)[0].cpu)[0].allocated}"
     limit     = "${tolist(tolist(data.vcd_org_vdc.asDatasource.compute_capacity)[0].cpu)[0].limit}"
    }

    memory {
     allocated = "${tolist(tolist(data.vcd_org_vdc.asDatasource.compute_capacity)[0].memory)[0].allocated}"
     limit     = "${tolist(tolist(data.vcd_org_vdc.asDatasource.compute_capacity)[0].memory)[0].limit}"
    }
  }

  storage_profile {
    name    = "${data.vcd_org_vdc.asDatasource.storage_profile[0].name}"
    enabled = "${data.vcd_org_vdc.asDatasource.storage_profile[0].enabled}"
    limit   = "${data.vcd_org_vdc.asDatasource.storage_profile[0].limit}"
    default = "${data.vcd_org_vdc.asDatasource.storage_profile[0].default}"
  }

  metadata = {
    vdc_metadata = "VDC Metadata"
  }

  enabled                  = "${data.vcd_org_vdc.asDatasource.enabled}"
  enable_thin_provisioning = "${data.vcd_org_vdc.asDatasource.enable_thin_provisioning}"
  enable_fast_provisioning = "${data.vcd_org_vdc.asDatasource.enable_fast_provisioning}"
  delete_force             = true
  delete_recursive         = true
}
`
