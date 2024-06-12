//go:build vdc || ALL || functional

package vcd

import (
	"fmt"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// TestAccVcdVdcTemplate tests the creation and update of 4 VDC templates with vcd_org_vdc_template, each one using a different Allocation model.
// Also tests the data source and import features.
func TestAccVcdVdcTemplate(t *testing.T) {
	preTestChecks(t)
	skipIfNotSysAdmin(t)

	var params = StringMap{
		"OrgToPublish":    testConfig.VCD.Org,
		"ProviderVdc":     testConfig.VCD.NsxtProviderVdc.Name,
		"ExternalNetwork": testConfig.Nsxt.ExternalNetwork,
		"NetworkPool":     testConfig.VCD.NsxtProviderVdc.NetworkPool,
		"StorageProfile":  testConfig.VCD.NsxtProviderVdc.StorageProfile,
		"Name1":           t.Name() + "1",
		"Name2":           t.Name() + "2",
		"Name3":           t.Name() + "3",
		"Name4":           t.Name() + "4",
	}
	testParamsNotEmpty(t, params)

	step1 := templateFill(testAccVdcTemplateStep1, params)
	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	debugPrintf("#[DEBUG] CONFIGURATION - Step 1: %s", step1)

	template1 := "vcd_org_vdc_template.template1"

	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckVdcTemplateDestroyed(params["Name1"].(string)),
			testAccCheckVdcTemplateDestroyed(params["Name2"].(string)),
			testAccCheckVdcTemplateDestroyed(params["Name3"].(string)),
			testAccCheckVdcTemplateDestroyed(params["Name4"].(string)),
		),
		Steps: []resource.TestStep{
			{
				Config: step1,
				Check:  resource.ComposeTestCheckFunc(),
			},
			{
				ResourceName:            template1,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdFunc:       importStateIdTopHierarchy(params["Name1"].(string)),
				ImportStateVerifyIgnore: []string{"bindings.%", "bindings"},
			},
		},
	})
	postTestChecks(t)
}

// Checks that a VDC Template is correctly destroyed
func testAccCheckVdcTemplateDestroyed(vdcTemplateName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*VCDClient)
		var err error
		var id string
		for _, rs := range s.RootModule().Resources {
			if rs.Type == "vcd_org_vdc_template" && rs.Primary.Attributes["name"] == vdcTemplateName {
				id = rs.Primary.ID
			}
		}

		if id == "" {
			return fmt.Errorf("vcd_vdc_template with name %s was not found in tfstate", vdcTemplateName)
		}

		_, err = conn.GetVdcTemplateById(id)
		if err != nil && !govcd.ContainsNotFound(err) {
			return fmt.Errorf("VDC Template '%s' still exists, but got an error when retrieving: %s", id, err)
		}
		if err == nil {
			return fmt.Errorf("VDC Template '%s' still exists", id)
		}

		return nil
	}
}

const testAccVdcTemplateStep1 = `
data "vcd_org" "org" {
  name = "{{.OrgToPublish}}"
}

data "vcd_provider_vdc" "pvdc" {
  name = "{{.ProviderVdc}}"
}

data "vcd_external_network_v2" "ext_net" {
  name = "{{.ExternalNetwork}}"
}

data "vcd_network_pool" "np" {
  name = "{{.NetworkPool}}"
}

resource "vcd_org_vdc_template" "template1" {
  name               = "{{.Name1}}"
  tenant_name        = "{{.Name1}}_tenant"
  description        = "{{.Name1}}_description"
  tenant_description = "{{.Name1}}_tenant_description"
  allocation_model   = "AllocationVApp"

  compute_configuration {
    cpu_limit         = 0
    cpu_guaranteed    = 20
    cpu_speed         = 256
    memory_limit      = 1024
    memory_guaranteed = 30
  }

  provider_vdc {
    id                  = data.vcd_provider_vdc.pvdc.id
    external_network_id = data.vcd_external_network_v2.ext_net.id
  }

  storage_profile {
    name    = "{{.StorageProfile}}"
    default = true
    limit   = 1024
  }

  network_pool_id = data.vcd_network_pool.np.id

  readable_by_org_ids = [
    data.vcd_org.org.id
  ]
}

resource "vcd_org_vdc_template" "template2" {
  name               = "{{.Name2}}"
  tenant_name        = "{{.Name2}}_tenant"
  description        = "{{.Name2}}_description"
  tenant_description = "{{.Name2}}_tenant_description"
  allocation_model   = "AllocationPool"

  compute_configuration {
	cpu_allocated     = 256
    cpu_guaranteed    = 20
    cpu_speed         = 1000
    memory_allocated  = 1024
    memory_guaranteed = 30
  }

  provider_vdc {
    id                  = data.vcd_provider_vdc.pvdc.id
    external_network_id = data.vcd_external_network_v2.ext_net.id
  }

  storage_profile {
    name    = "{{.StorageProfile}}"
    default = true
    limit   = 1024
  }

  network_pool_id = data.vcd_network_pool.np.id

  readable_by_org_ids = [
    data.vcd_org.org.id
  ]
}

resource "vcd_org_vdc_template" "template3" {
  name               = "{{.Name3}}"
  tenant_name        = "{{.Name3}}_tenant"
  description        = "{{.Name3}}_description"
  tenant_description = "{{.Name3}}_tenant_description"
  allocation_model   = "ReservationPool"

  compute_configuration {
    cpu_allocated    = 256
    cpu_limit        = 0
    memory_allocated = 1024
    memory_limit     = 0
  }

  provider_vdc {
    id                  = data.vcd_provider_vdc.pvdc.id
    external_network_id = data.vcd_external_network_v2.ext_net.id
  }

  storage_profile {
    name    = "{{.StorageProfile}}"
    default = true
    limit   = 1024
  }

  network_pool_id = data.vcd_network_pool.np.id

  readable_by_org_ids = [
    data.vcd_org.org.id
  ]
}

resource "vcd_org_vdc_template" "template4" {
  name               = "{{.Name4}}"
  tenant_name        = "{{.Name4}}_tenant"
  description        = "{{.Name4}}_description"
  tenant_description = "{{.Name4}}_tenant_description"
  allocation_model   = "Flex"

  compute_configuration {
    cpu_allocated     = 256
    cpu_limit         = 0
    cpu_guaranteed    = 20
    cpu_speed         = 256
    memory_allocated  = 1024
    memory_limit      = 0
    memory_guaranteed = 30
  }

  provider_vdc {
    id                  = data.vcd_provider_vdc.pvdc.id
    external_network_id = data.vcd_external_network_v2.ext_net.id
  }

  storage_profile {
    name    = "{{.StorageProfile}}"
    default = true
    limit   = 1024
  }

  network_pool_id = data.vcd_network_pool.np.id

  readable_by_org_ids = [
    data.vcd_org.org.id
  ]
}
`
