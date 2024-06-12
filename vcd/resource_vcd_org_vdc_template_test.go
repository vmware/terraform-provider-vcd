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
		"Vdc":             testConfig.Nsxt.Vdc,
		"EdgeCluster":     testConfig.Nsxt.NsxtEdgeCluster,
		"ExternalNetwork": testConfig.Nsxt.ExternalNetwork,
		"NetworkPool":     testConfig.VCD.NsxtProviderVdc.NetworkPool,
		"StorageProfile":  testConfig.VCD.NsxtProviderVdc.StorageProfile,
		"Name1":           t.Name() + "1",
		"Name2":           t.Name() + "2",
		"Name3":           t.Name() + "3",
		"Name4":           t.Name() + "4",
		"FuncName":        t.Name() + "Step1",
	}
	testParamsNotEmpty(t, params)

	step1 := templateFill(testAccVdcTemplateResource, params)
	debugPrintf("#[DEBUG] CONFIGURATION - Step 1: %s", step1)
	params["FuncName"] = t.Name() + "Step2"
	step2 := templateFill(testAccVdcTemplateResource, params)
	debugPrintf("#[DEBUG] CONFIGURATION - Step 2: %s", step2)
	params["FuncName"] = t.Name() + "Step3"
	step3 := templateFill(testAccVdcTemplateResourceAndDatasource, params)
	debugPrintf("#[DEBUG] CONFIGURATION - Step 3: %s", step3)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

	template := "vcd_org_vdc_template.template"
	dsTemplate := "data.vcd_org_vdc_template.ds_template"

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
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(template+"1", "name", params["Name1"].(string)),

					resource.TestCheckResourceAttr(template+"2", "name", params["Name2"].(string)),

					resource.TestCheckResourceAttr(template+"3", "name", params["Name3"].(string)),

					resource.TestCheckResourceAttr(template+"4", "name", params["Name4"].(string)),
				),
			},
			{
				Config: step2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(template+"1", "name", params["Name1"].(string)),

					resource.TestCheckResourceAttr(template+"2", "name", params["Name2"].(string)),

					resource.TestCheckResourceAttr(template+"3", "name", params["Name3"].(string)),

					resource.TestCheckResourceAttr(template+"4", "name", params["Name4"].(string)),
				),
			},
			{
				Config: step3,
				Check: resource.ComposeTestCheckFunc(
					resourceFieldsEqual(template+"1", dsTemplate+"1", nil),
					resourceFieldsEqual(template+"2", dsTemplate+"2", nil),
					resourceFieldsEqual(template+"3", dsTemplate+"3", nil),
					resourceFieldsEqual(template+"4", dsTemplate+"4", nil),
				),
			},
			{
				ResourceName:            template + "1",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateIdFunc:       importStateIdTopHierarchy(params["Name1"].(string)),
				ImportStateVerifyIgnore: []string{"bindings.%", "bindings"}, // TODO: Work with this, what happens after import
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

const testAccVdcTemplateResource = `
data "vcd_org" "org" {
  name = "{{.OrgToPublish}}"
}

data "vcd_org_vdc" "vdc" {
  org  = data.vcd_org.org.name
  name = "{{.Vdc}}"
}

data "vcd_nsxt_edge_cluster" "ec" {
  org    = data.vcd_org.org.name
  vdc_id = data.vcd_org_vdc.vdc.id
  name   = "{{.EdgeCluster}}"
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

  edge_gateway {
    name                 = "edgy"
    ip_allocation_count  = 10
    network_name         = "net2"
    network_gateway_cidr = "1.2.3.4/24"
    static_ip_pool {
      start_address = "1.2.3.4"
      end_address   = "1.2.3.4"
    }
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

  edge_gateway {
    name                 = "edgy"
    ip_allocation_count  = 10
    network_name         = "net2"
    network_gateway_cidr = "1.2.3.4/24"
    static_ip_pool {
      start_address = "1.2.3.4"
      end_address   = "1.2.3.4"
    }
  }

  provider_vdc {
    id                      = data.vcd_provider_vdc.pvdc.id
    external_network_id     = data.vcd_external_network_v2.ext_net.id
	gateway_edge_cluster_id = data.vcd_nsxt_edge_cluster.ec.id
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

const testAccVdcTemplateResourceAndDatasource = testAccVdcTemplateResource + `
data "vcd_org_vdc_template" "ds_template1" {
  name = vcd_org_vdc_template.template1.name
}

data "vcd_org_vdc_template" "ds_template2" {
  name = vcd_org_vdc_template.template2.name
}

data "vcd_org_vdc_template" "ds_template3" {
  name = vcd_org_vdc_template.template3.name
}

data "vcd_org_vdc_template" "ds_template4" {
  name = vcd_org_vdc_template.template4.name
}
`
