//go:build vdc || ALL || functional

package vcd

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// TestAccVcdVdcTemplateInstance tests the creation of a VDC with a VDC Template.
func TestAccVcdVdcTemplateInstance(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	var params = StringMap{
		"OrgToPublish":    testConfig.VCD.Org,
		"ProviderVdc":     testConfig.VCD.NsxtProviderVdc.Name,
		"ExternalNetwork": testConfig.Nsxt.ExternalNetwork,
		"StorageProfile":  testConfig.VCD.NsxtProviderVdc.StorageProfile,
		"Name":            t.Name(),
	}
	testParamsNotEmpty(t, params)

	step1 := templateFill(testAccVdcTemplateInstanceStep1, params)
	debugPrintf("#[DEBUG] CONFIGURATION - Step 1: %s", step1)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	template := "vcd_org_vdc_template.template"
	instance := "vcd_org_vdc_template_instance.instance"
	vdc := "data.vcd_org_vdc.new_vdc"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckVdcTemplateDestroyed(params["Name"].(string)),
			testAccCheckVdcTemplateInstanceDestroyed(params["OrgToPublish"].(string), params["Name"].(string)),
		),
		Steps: []resource.TestStep{
			{
				Config: step1,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Correctness of the instance
					resource.TestCheckResourceAttr(template, "name", params["Name"].(string)),
					resource.TestCheckResourceAttrPair(instance, "org_vdc_template_id", template, "id"),
					resource.TestCheckResourceAttrPair(instance, "name", template, "name"),
					resource.TestCheckResourceAttrPair(instance, "description", template, "name"),
					resource.TestCheckResourceAttrPair(instance, "org_id", "data.vcd_org.org", "id"),
					resource.TestCheckResourceAttrPair(vdc, "id", instance, "id"),
					resource.TestCheckResourceAttrPair(vdc, "name", instance, "name"),
					resource.TestCheckResourceAttrPair(vdc, "description", instance, "description"),
					resource.TestCheckResourceAttrPair(vdc, "org", "data.vcd_org.org", "name"),

					// Correctness of what we put in the template (and default values)
					resource.TestCheckResourceAttrPair(vdc, "allocation_model", template, "allocation_model"),
					resource.TestCheckResourceAttrPair(vdc, "nic_quota", template, "nic_quota"),
					resource.TestCheckResourceAttrPair(vdc, "vm_quota", template, "vm_quota"),
					resource.TestCheckResourceAttrPair(vdc, "network_quota", template, "provisioned_network_quota"),
					resource.TestCheckResourceAttrPair(vdc, "enable_thin_provisioning", template, "enable_thin_provisioning"),
					resource.TestCheckResourceAttrPair(vdc, "enable_fast_provisioning", template, "enable_fast_provisioning"),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccVdcTemplateInstanceStep1 = `
data "vcd_org" "org" {
  name = "{{.OrgToPublish}}"
}

data "vcd_provider_vdc" "pvdc" {
  name = "{{.ProviderVdc}}"
}

data "vcd_external_network_v2" "ext_net" {
  name = "{{.ExternalNetwork}}"
}

resource "vcd_org_vdc_template" "template" {
  name               = "{{.Name}}"
  tenant_name        = "{{.Name}}"
  allocation_model   = "Flex"

  compute_configuration {
    cpu_allocated     = 0
    cpu_limit         = 256
    cpu_guaranteed    = 50
    cpu_speed         = 1000
    memory_allocated  = 1024
    memory_limit      = 0
    memory_guaranteed = 50
    
    elasticity                 = true
    include_vm_memory_overhead = true
  }

  provider_vdc {
    id                  = data.vcd_provider_vdc.pvdc.id
    external_network_id = data.vcd_external_network_v2.ext_net.id
  }

  storage_profile {
    name    = "{{.StorageProfile}}"
    default = true
    limit   = 0
  }

  enable_thin_provisioning = true
  enable_fast_provisioning = true

  readable_by_org_ids = [
    data.vcd_org.org.id
  ]
}

resource "vcd_org_vdc_template_instance" "instance" {
  org_vdc_template_id = vcd_org_vdc_template.template.id
  name                = vcd_org_vdc_template.template.name
  org_id              = data.vcd_org.org.id
  description         = vcd_org_vdc_template.template.name
}

# This one depends on the VDC Template instance, so it waits for it to be finished creating the VDC
data "vcd_org_vdc" "new_vdc" {
	org  = data.vcd_org.org.name
    name = vcd_org_vdc_template_instance.instance.name
}
`

// Checks that a VDC Template instance is correctly destroyed
func testAccCheckVdcTemplateInstanceDestroyed(orgName, vdcName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*VCDClient)
		var err error
		var id string
		for _, rs := range s.RootModule().Resources {
			if rs.Type == "vcd_org_vdc_template_instance" && rs.Primary.Attributes["name"] == vdcName {
				id = rs.Primary.ID
			}
		}

		if id == "" {
			return fmt.Errorf("vcd_vdc_template_instance with name %s was not found in tfstate", vdcName)
		}

		org, err := conn.GetOrg(orgName)
		if err != nil {
			return fmt.Errorf("error retrieving Org with name '%s': %s", orgName, err)
		}

		_, err = org.GetVDCById(id, false)
		if err != nil && !govcd.ContainsNotFound(err) {
			return fmt.Errorf("VDC '%s' still exists, but got an error when retrieving: %s", id, err)
		}
		if err == nil {
			return fmt.Errorf("VDC '%s' still exists", id)
		}

		return nil
	}
}
