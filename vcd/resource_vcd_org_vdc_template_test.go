//go:build vdc || ALL || functional

package vcd

import (
	"fmt"
	"github.com/vmware/go-vcloud-director/v2/govcd"
	"regexp"
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
		"ProviderVcdSystem":   providerVcdSystem,
		"ProviderVcdTenant":   providerVcdOrg1,
		"OrgToPublish":        testConfig.VCD.Org,
		"ProviderVdc":         testConfig.VCD.NsxtProviderVdc.Name,
		"Vdc":                 testConfig.Nsxt.Vdc,
		"EdgeCluster":         testConfig.Nsxt.NsxtEdgeCluster,
		"ExternalNetwork":     testConfig.Nsxt.ExternalNetwork,
		"NetworkPool":         testConfig.VCD.NsxtProviderVdc.NetworkPool,
		"StorageProfile":      testConfig.VCD.NsxtProviderVdc.StorageProfile,
		"StorageProfile2":     testConfig.VCD.NsxtProviderVdc.StorageProfile2,
		"Name1":               t.Name() + "1",
		"Name2":               t.Name() + "2",
		"Name3":               t.Name() + "3",
		"Name4":               t.Name() + "4",
		"FuncName":            t.Name() + "Step1",
		"StorageProfileLimit": 1024,
		"CpuLimit":            0,
		"CpuGuaranteed":       20,
		"CpuSpeed":            1000,
		"CpuAllocated":        256,
		"MemoryLimit":         0,
		"MemoryGuaranteed":    50,
		"MemoryAllocated":     1024,
	}
	testParamsNotEmpty(t, params)

	step1 := templateFill(testAccVdcTemplateResource, params)
	debugPrintf("#[DEBUG] CONFIGURATION - Step 1: %s", step1)
	params["FuncName"] = t.Name() + "Step2"
	params["StorageProfileLimit"] = 2048
	params["CpuLimit"] = 512
	params["CpuGuaranteed"] = 100
	params["CpuSpeed"] = 1500
	params["CpuAllocated"] = 512
	params["MemoryLimit"] = 2048
	params["MemoryGuaranteed"] = 60
	params["MemoryAllocated"] = 1536
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
		ProviderFactories: buildMultipleProviders(),
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckVdcTemplateDestroyed(params["Name1"].(string)),
			testAccCheckVdcTemplateDestroyed(params["Name2"].(string)),
			testAccCheckVdcTemplateDestroyed(params["Name3"].(string)),
			testAccCheckVdcTemplateDestroyed(params["Name4"].(string)),
		),
		Steps: []resource.TestStep{
			{
				Config: step1,
				Check: resource.ComposeAggregateTestCheckFunc(
					// First VDC template
					resource.TestCheckResourceAttr(template+"1", "name", params["Name1"].(string)),
					resource.TestCheckResourceAttr(template+"1", "description", params["Name1"].(string)+"_description"),
					resource.TestCheckResourceAttr(template+"1", "tenant_name", params["Name1"].(string)+"_tenant"),
					resource.TestCheckResourceAttr(template+"1", "tenant_description", params["Name1"].(string)+"_tenant_description"),
					resource.TestMatchTypeSetElemNestedAttrs(template+"1", "provider_vdc.*", map[string]*regexp.Regexp{
						"id":                  regexp.MustCompile(`^urn:vcloud:providervdc:.+$`),
						"external_network_id": regexp.MustCompile(`^urn:vcloud:network:.+$`),
					}),
					resource.TestCheckResourceAttr(template+"1", "allocation_model", "AllocationVApp"),
					resource.TestCheckTypeSetElemNestedAttrs(template+"1", "compute_configuration.*", map[string]string{
						"cpu_allocated":              "0", // Not used
						"cpu_limit":                  "0",
						"cpu_guaranteed":             "20",
						"cpu_speed":                  "1000",
						"memory_allocated":           "0", // Not used
						"memory_guaranteed":          "50",
						"memory_limit":               "0",
						"elasticity":                 "true",  // Computed
						"include_vm_memory_overhead": "false", // Not used
					}),
					resource.TestCheckTypeSetElemNestedAttrs(template+"1", "storage_profile.*", map[string]string{
						"name":    params["StorageProfile"].(string),
						"default": "true",
						"limit":   "1024",
					}),
					resource.TestCheckResourceAttr(template+"1", "enable_fast_provisioning", "false"),
					resource.TestCheckResourceAttr(template+"1", "enable_thin_provisioning", "false"),
					resource.TestCheckTypeSetElemNestedAttrs(template+"1", "storage_profile.*", map[string]string{
						"name":    params["StorageProfile"].(string),
						"default": "true",
						"limit":   "1024",
					}),
					resource.TestCheckNoResourceAttr(template+"1", "edge_gateway.0"),
					resource.TestMatchResourceAttr(template+"1", "network_pool_id", regexp.MustCompile(`^urn:vcloud:networkpool:.+$`)),
					resource.TestCheckResourceAttr(template+"1", "nic_quota", "100"),
					resource.TestCheckResourceAttr(template+"1", "vm_quota", "0"),
					resource.TestCheckResourceAttr(template+"1", "provisioned_network_quota", "1000"),
					resource.TestCheckResourceAttr(template+"1", "readable_by_org_ids.#", "1"),

					// Second VDC template
					resource.TestCheckResourceAttr(template+"2", "name", params["Name2"].(string)),
					resource.TestCheckResourceAttr(template+"2", "description", params["Name2"].(string)+"_description"),
					resource.TestCheckResourceAttr(template+"2", "tenant_name", params["Name2"].(string)+"_tenant"),
					resource.TestCheckResourceAttr(template+"2", "tenant_description", params["Name2"].(string)+"_tenant_description"),
					resource.TestMatchTypeSetElemNestedAttrs(template+"2", "provider_vdc.*", map[string]*regexp.Regexp{
						"id":                  regexp.MustCompile(`^urn:vcloud:providervdc:.+$`),
						"external_network_id": regexp.MustCompile(`^urn:vcloud:network:.+$`),
					}),
					resource.TestCheckResourceAttr(template+"2", "allocation_model", "AllocationPool"),
					resource.TestCheckTypeSetElemNestedAttrs(template+"2", "compute_configuration.*", map[string]string{
						"cpu_allocated":              "256",
						"cpu_limit":                  "0", // Not used
						"cpu_guaranteed":             "20",
						"cpu_speed":                  "1000",
						"memory_allocated":           "1024",
						"memory_guaranteed":          "50",
						"memory_limit":               "0",     // Not used
						"elasticity":                 "false", // Not used
						"include_vm_memory_overhead": "true",  // Computed
					}),
					resource.TestCheckTypeSetElemNestedAttrs(template+"2", "storage_profile.*", map[string]string{
						"name":    params["StorageProfile"].(string),
						"default": "true",
						"limit":   "1024",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(template+"2", "storage_profile.*", map[string]string{
						"name":    params["StorageProfile2"].(string),
						"default": "false",
						"limit":   "128",
					}),
					resource.TestCheckResourceAttr(template+"2", "enable_fast_provisioning", "true"),
					resource.TestCheckResourceAttr(template+"2", "enable_thin_provisioning", "true"),
					resource.TestCheckResourceAttr(template+"2", "edge_gateway.0.name", "edgy2"),
					resource.TestCheckResourceAttr(template+"2", "edge_gateway.0.ip_allocation_count", "10"),
					resource.TestCheckResourceAttr(template+"2", "edge_gateway.0.routed_network_name", "net2"),
					resource.TestCheckResourceAttr(template+"2", "edge_gateway.0.routed_network_gateway_cidr", "1.2.3.4/24"),
					resource.TestCheckTypeSetElemNestedAttrs(template+"2", "edge_gateway.0.static_ip_pool.*", map[string]string{
						"start_address": "1.2.3.4",
						"end_address":   "1.2.3.4",
					}),
					resource.TestCheckNoResourceAttr(template+"2", "network_pool_id"),
					resource.TestCheckResourceAttr(template+"2", "nic_quota", "100"),
					resource.TestCheckResourceAttr(template+"2", "vm_quota", "0"),
					resource.TestCheckResourceAttr(template+"2", "provisioned_network_quota", "1000"),
					resource.TestCheckResourceAttr(template+"2", "readable_by_org_ids.#", "0"),

					// Third VDC template
					resource.TestCheckResourceAttr(template+"3", "name", params["Name3"].(string)),
					resource.TestCheckResourceAttr(template+"3", "description", params["Name3"].(string)+"_description"),
					resource.TestCheckResourceAttr(template+"3", "tenant_name", params["Name3"].(string)+"_tenant"),
					resource.TestCheckResourceAttr(template+"3", "tenant_description", params["Name3"].(string)+"_tenant_description"),
					resource.TestMatchTypeSetElemNestedAttrs(template+"3", "provider_vdc.*", map[string]*regexp.Regexp{
						"id":                      regexp.MustCompile(`^urn:vcloud:providervdc:.+$`),
						"external_network_id":     regexp.MustCompile(`^urn:vcloud:network:.+$`),
						"gateway_edge_cluster_id": regexp.MustCompile(`^.+$`),
					}),
					resource.TestCheckResourceAttr(template+"3", "allocation_model", "ReservationPool"),
					resource.TestCheckTypeSetElemNestedAttrs(template+"3", "compute_configuration.*", map[string]string{
						"cpu_allocated":              "256",
						"cpu_limit":                  "0", // Unlimited
						"cpu_guaranteed":             "0", // Not used
						"cpu_speed":                  "0", // Not used
						"memory_allocated":           "1024",
						"memory_guaranteed":          "0", // Not used
						"memory_limit":               "0", // Not used
						"elasticity":                 "false",
						"include_vm_memory_overhead": "true", // Computed
					}),
					resource.TestCheckTypeSetElemNestedAttrs(template+"3", "storage_profile.*", map[string]string{
						"name":    params["StorageProfile"].(string),
						"default": "true",
						"limit":   "1024",
					}),
					resource.TestCheckResourceAttr(template+"3", "enable_fast_provisioning", "false"),
					resource.TestCheckResourceAttr(template+"3", "enable_thin_provisioning", "false"),
					resource.TestCheckResourceAttr(template+"3", "edge_gateway.0.name", "edgy3"),
					resource.TestCheckResourceAttr(template+"3", "edge_gateway.0.ip_allocation_count", "15"),
					resource.TestCheckResourceAttr(template+"3", "edge_gateway.0.routed_network_name", "net3"),
					resource.TestCheckResourceAttr(template+"3", "edge_gateway.0.routed_network_gateway_cidr", "1.1.1.1/2"),
					resource.TestCheckTypeSetElemNestedAttrs(template+"3", "edge_gateway.0.static_ip_pool.*", map[string]string{
						"start_address": "1.1.1.1",
						"end_address":   "1.1.1.1",
					}),
					resource.TestMatchResourceAttr(template+"3", "network_pool_id", regexp.MustCompile(`^urn:vcloud:networkpool:.+$`)),
					resource.TestCheckResourceAttr(template+"3", "nic_quota", "100"),
					resource.TestCheckResourceAttr(template+"3", "vm_quota", "100"),
					resource.TestCheckResourceAttr(template+"3", "provisioned_network_quota", "20"),
					resource.TestCheckResourceAttr(template+"1", "readable_by_org_ids.#", "1"),

					// Fourth VDC template
					resource.TestCheckResourceAttr(template+"4", "name", params["Name4"].(string)),
					resource.TestCheckResourceAttr(template+"4", "description", params["Name4"].(string)+"_description"),
					resource.TestCheckResourceAttr(template+"4", "tenant_name", params["Name4"].(string)+"_tenant"),
					resource.TestCheckResourceAttr(template+"4", "tenant_description", params["Name4"].(string)+"_tenant_description"),
					resource.TestMatchTypeSetElemNestedAttrs(template+"4", "provider_vdc.*", map[string]*regexp.Regexp{
						"id":                  regexp.MustCompile(`^urn:vcloud:providervdc:.+$`),
						"external_network_id": regexp.MustCompile(`^urn:vcloud:network:.+$`),
					}),
					resource.TestCheckResourceAttr(template+"4", "allocation_model", "Flex"),
					resource.TestCheckTypeSetElemNestedAttrs(template+"4", "compute_configuration.*", map[string]string{
						"cpu_allocated":              "256",
						"cpu_limit":                  "0",
						"cpu_guaranteed":             "20",
						"cpu_speed":                  "1000",
						"memory_allocated":           "1024",
						"memory_guaranteed":          "50",
						"memory_limit":               "0",
						"elasticity":                 "true",
						"include_vm_memory_overhead": "true",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(template+"4", "storage_profile.*", map[string]string{
						"name":    params["StorageProfile"].(string),
						"default": "true",
						"limit":   "1024",
					}),
					resource.TestCheckResourceAttr(template+"4", "enable_fast_provisioning", "false"),
					resource.TestCheckResourceAttr(template+"4", "enable_thin_provisioning", "false"),
					resource.TestCheckNoResourceAttr(template+"4", "edge_gateway.0"),
					resource.TestMatchResourceAttr(template+"4", "network_pool_id", regexp.MustCompile(`^urn:vcloud:networkpool:.+$`)),
					resource.TestCheckResourceAttr(template+"4", "nic_quota", "10"),
					resource.TestCheckResourceAttr(template+"4", "vm_quota", "0"),
					resource.TestCheckResourceAttr(template+"4", "provisioned_network_quota", "1000"),
					resource.TestCheckResourceAttr(template+"4", "readable_by_org_ids.#", "1"),
				),
			},
			{
				Config: step2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(template+"1", "name", params["Name1"].(string)),
					resource.TestCheckTypeSetElemNestedAttrs(template+"1", "compute_configuration.*", map[string]string{
						"cpu_allocated":              "0", // Not used
						"cpu_limit":                  "512",
						"cpu_guaranteed":             "100",
						"cpu_speed":                  "1500",
						"memory_allocated":           "0", // Not used
						"memory_guaranteed":          "60",
						"memory_limit":               "2048",
						"elasticity":                 "true",  // Computed
						"include_vm_memory_overhead": "false", // Not used
					}),

					resource.TestCheckResourceAttr(template+"2", "name", params["Name2"].(string)),
					resource.TestCheckTypeSetElemNestedAttrs(template+"2", "compute_configuration.*", map[string]string{
						"cpu_allocated":              "512",
						"cpu_limit":                  "0", // Not used
						"cpu_guaranteed":             "100",
						"cpu_speed":                  "1500",
						"memory_allocated":           "1536",
						"memory_guaranteed":          "60",
						"memory_limit":               "0",     // Not used
						"elasticity":                 "false", // Not used
						"include_vm_memory_overhead": "true",  // Computed
					}),
					resource.TestCheckResourceAttr(template+"3", "name", params["Name3"].(string)),
					resource.TestCheckTypeSetElemNestedAttrs(template+"3", "compute_configuration.*", map[string]string{
						"cpu_allocated":              "512",
						"cpu_limit":                  "512",
						"cpu_guaranteed":             "0", // Not used
						"cpu_speed":                  "0", // Not used
						"memory_allocated":           "1536",
						"memory_guaranteed":          "0", // Not used
						"memory_limit":               "0", // Not used
						"elasticity":                 "false",
						"include_vm_memory_overhead": "true", // Computed
					}),
					resource.TestCheckResourceAttr(template+"4", "name", params["Name4"].(string)),
					resource.TestCheckTypeSetElemNestedAttrs(template+"4", "compute_configuration.*", map[string]string{
						"cpu_allocated":              "512",
						"cpu_limit":                  "512",
						"cpu_guaranteed":             "100",
						"cpu_speed":                  "1500",
						"memory_allocated":           "1536",
						"memory_guaranteed":          "60",
						"memory_limit":               "2048",
						"elasticity":                 "true",
						"include_vm_memory_overhead": "true",
					}),
				),
			},
			{
				Config: step3,
				Check: resource.ComposeAggregateTestCheckFunc(
					resourceFieldsEqual(template+"1", dsTemplate+"1", nil),
					resourceFieldsEqual(template+"2", dsTemplate+"2", nil),
					resourceFieldsEqual(template+"3", dsTemplate+"3", nil),
					resourceFieldsEqual(template+"4", dsTemplate+"4", nil),

					// This one is read by a tenant, so we ignore 'readable_by_org_ids' completely
					resourceFieldsEqualCustom(template+"4", dsTemplate+"5", []string{"readable_by_org_ids"}, stringInSlicePartially),
				),
			},
			{
				ResourceName:      template + "1",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateIdFunc: importStateIdTopHierarchy(params["Name1"].(string)),
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
  provider = {{.ProviderVcdSystem}}

  name = "{{.OrgToPublish}}"
}

data "vcd_org_vdc" "vdc" {
  provider = {{.ProviderVcdSystem}}

  org  = data.vcd_org.org.name
  name = "{{.Vdc}}"
}

data "vcd_nsxt_edge_cluster" "ec" {
  provider = {{.ProviderVcdSystem}}

  org    = data.vcd_org.org.name
  vdc_id = data.vcd_org_vdc.vdc.id
  name   = "{{.EdgeCluster}}"
}

data "vcd_provider_vdc" "pvdc" {
  provider = {{.ProviderVcdSystem}}

  name = "{{.ProviderVdc}}"
}

data "vcd_external_network_v2" "ext_net" {
  provider = {{.ProviderVcdSystem}}

  name = "{{.ExternalNetwork}}"
}

data "vcd_network_pool" "np" {
  provider = {{.ProviderVcdSystem}}

  name = "{{.NetworkPool}}"
}

resource "vcd_org_vdc_template" "template1" {
  provider = {{.ProviderVcdSystem}}

  name               = "{{.Name1}}"
  tenant_name        = "{{.Name1}}_tenant"
  description        = "{{.Name1}}_description"
  tenant_description = "{{.Name1}}_tenant_description"
  allocation_model   = "AllocationVApp"

  compute_configuration {
    cpu_limit         = {{.CpuLimit}}
    cpu_guaranteed    = {{.CpuGuaranteed}}
    cpu_speed         = {{.CpuSpeed}}
    memory_limit      = {{.MemoryLimit}}
    memory_guaranteed = {{.MemoryGuaranteed}}
  }

  provider_vdc {
    id                  = data.vcd_provider_vdc.pvdc.id
    external_network_id = data.vcd_external_network_v2.ext_net.id
  }

  storage_profile {
    name    = "{{.StorageProfile}}"
    default = true
    limit   = {{.StorageProfileLimit}}
  }

  network_pool_id = data.vcd_network_pool.np.id

  readable_by_org_ids = [
    data.vcd_org.org.id
  ]
}

resource "vcd_org_vdc_template" "template2" {
  provider = {{.ProviderVcdSystem}}

  name               = "{{.Name2}}"
  tenant_name        = "{{.Name2}}_tenant"
  description        = "{{.Name2}}_description"
  tenant_description = "{{.Name2}}_tenant_description"
  allocation_model   = "AllocationPool"

  compute_configuration {
	cpu_allocated     = {{.CpuAllocated}}
    cpu_guaranteed    = {{.CpuGuaranteed}}
    cpu_speed         = {{.CpuSpeed}}
    memory_allocated  = {{.MemoryAllocated}}
    memory_guaranteed = {{.MemoryGuaranteed}}
  }

  edge_gateway {
    name                        = "edgy2"
    ip_allocation_count         = 10
    routed_network_name         = "net2"
    routed_network_gateway_cidr = "1.2.3.4/24"
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
    limit   = {{.StorageProfileLimit}}
  }
  storage_profile {
    name    = "{{.StorageProfile2}}"
    default = false
    limit   = 128
  }

  enable_fast_provisioning = true
  enable_thin_provisioning = true
}

resource "vcd_org_vdc_template" "template3" {
  provider = {{.ProviderVcdSystem}}

  name               = "{{.Name3}}"
  tenant_name        = "{{.Name3}}_tenant"
  description        = "{{.Name3}}_description"
  tenant_description = "{{.Name3}}_tenant_description"
  allocation_model   = "ReservationPool"

  compute_configuration {
    cpu_allocated    = {{.CpuAllocated}}
    cpu_limit        = {{.CpuLimit}}
    memory_allocated = {{.MemoryAllocated}}
  }

  edge_gateway {
    name                        = "edgy3"
    ip_allocation_count         = 15
    routed_network_name         = "net3"
    routed_network_gateway_cidr = "1.1.1.1/2"
    static_ip_pool {
      start_address = "1.1.1.1"
      end_address   = "1.1.1.1"
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
    limit   = {{.StorageProfileLimit}}
  }

  network_pool_id = data.vcd_network_pool.np.id

  nic_quota                 = 100
  vm_quota                  = 100
  provisioned_network_quota = 20

  readable_by_org_ids = [
    data.vcd_org.org.id
  ]
}

resource "vcd_org_vdc_template" "template4" {
  provider = {{.ProviderVcdSystem}}

  name               = "{{.Name4}}"
  tenant_name        = "{{.Name4}}_tenant"
  description        = "{{.Name4}}_description"
  tenant_description = "{{.Name4}}_tenant_description"
  allocation_model   = "Flex"

  compute_configuration {
    cpu_allocated     = {{.CpuAllocated}}
    cpu_limit         = {{.CpuLimit}}
    cpu_guaranteed    = {{.CpuGuaranteed}}
    cpu_speed         = {{.CpuSpeed}}
    memory_allocated  = {{.MemoryAllocated}}
    memory_limit      = {{.MemoryLimit}}
    memory_guaranteed = {{.MemoryGuaranteed}}
    
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
    limit   = {{.StorageProfileLimit}}
  }

  network_pool_id = data.vcd_network_pool.np.id

  readable_by_org_ids = [
    data.vcd_org.org.id
  ]
}
`

const testAccVdcTemplateResourceAndDatasource = testAccVdcTemplateResource + `
data "vcd_org_vdc_template" "ds_template1" {
  provider = {{.ProviderVcdSystem}}

  name = vcd_org_vdc_template.template1.name
}

data "vcd_org_vdc_template" "ds_template2" {
  provider = {{.ProviderVcdSystem}}

  name = vcd_org_vdc_template.template2.name
}

data "vcd_org_vdc_template" "ds_template3" {
  provider = {{.ProviderVcdSystem}}

  name = vcd_org_vdc_template.template3.name
}

data "vcd_org_vdc_template" "ds_template4" {
  provider = {{.ProviderVcdSystem}}

  name = vcd_org_vdc_template.template4.name
}

data "vcd_org_vdc_template" "ds_template5" {
  provider = {{.ProviderVcdTenant}}

  # Careful, uses tenant_name as we use tenant configuration now
  name = vcd_org_vdc_template.template4.tenant_name
}
`
