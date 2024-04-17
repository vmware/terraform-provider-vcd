//go:build vapp || vm || standaloneVm || ALL || functional

package vcd

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/vmware/go-vcloud-director/v2/types/v56"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcdVAppVm_extraconfig(t *testing.T) {
	preTestChecks(t)

	initialValue := "value"
	updatedValue := "another value"
	var params = StringMap{
		"TestName":          t.Name(),
		"Org":               testConfig.VCD.Org,
		"Vdc":               testConfig.Nsxt.Vdc,
		"Catalog":           testConfig.VCD.Catalog.NsxtBackedCatalogName,
		"CatalogItem":       testConfig.VCD.Catalog.CatalogItemWithMultiVms,
		"Media":             testConfig.Media.NsxtBackedMediaName,
		"NsxtEdgeGateway":   testConfig.Nsxt.EdgeGateway,
		"ExtraConfigValue1": initialValue + "1",
		"ExtraConfigValue2": initialValue + "2",
		"FuncName":          t.Name() + "-step1",
		"Tags":              "vapp vm",
		"SkipMessage":       " ",
	}
	testParamsNotEmpty(t, params)

	configTextStep1 := templateFill(testAccVcdVAppVm_extraconfig, params)

	params["SkipMessage"] = "# skip-binary-test: only for updates"
	params["ExtraConfigValue1"] = updatedValue + "1"
	params["ExtraConfigValue2"] = updatedValue + "2"
	params["FuncName"] = t.Name() + "-step2"
	configTextStep2 := templateFill(testAccVcdVAppVm_extraconfig, params)

	params["ExtraConfigValue1"] = ""
	params["ExtraConfigValue2"] = ""
	params["FuncName"] = t.Name() + "-step3"
	configTextStep3 := templateFill(testAccVcdVAppVm_extraconfig, params)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}
	debugPrintf("#[DEBUG] CONFIGURATION: %s\n", configTextStep1)
	resource.Test(t, resource.TestCase{
		ProviderFactories: testAccProviders,
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckVcdNsxtVAppVmDestroy(t.Name()+"-template-vm"),
			testAccCheckVcdNsxtVAppVmDestroy(t.Name()+"-empty-vm"),
			testAccCheckVcdStandaloneVmDestroy(t.Name()+"-template-standalone-vm", testConfig.VCD.Org, testConfig.Nsxt.Vdc),
			testAccCheckVcdStandaloneVmDestroy(t.Name()+"-empty-standalone-vm", testConfig.VCD.Org, testConfig.Nsxt.Vdc),
		),
		Steps: []resource.TestStep{
			{
				Config: configTextStep1,
				Check: resource.ComposeAggregateTestCheckFunc(

					// vApp checks
					resource.TestCheckResourceAttr("vcd_vapp.template-vm", "name", t.Name()+"-template-vm"),
					resource.TestCheckResourceAttr("vcd_vapp.template-vm", "description", "vApp for Template VM description"),
					resource.TestCheckResourceAttr("vcd_vapp.template-vm", "power_on", "false"),
					resource.TestCheckResourceAttr("vcd_vapp.template-vm", "status", "1"), // 1 - means RESOLVED
					resource.TestCheckResourceAttr("vcd_vapp.template-vm", "status_text", "RESOLVED"),
					testAccCheckVcdVappPowerState(testConfig.VCD.Org, testConfig.Nsxt.Vdc, t.Name()+"-template-vm", []string{"POWERED_OFF"}),

					resource.TestCheckResourceAttr("vcd_vapp.empty-vm", "name", t.Name()+"-empty-vm"),
					resource.TestCheckResourceAttr("vcd_vapp.empty-vm", "description", "vApp for Empty VM description"),
					resource.TestCheckResourceAttr("vcd_vapp.empty-vm", "power_on", "false"),
					resource.TestCheckResourceAttr("vcd_vapp.empty-vm", "status", "1"), // 1 - means RESOLVED
					resource.TestCheckResourceAttr("vcd_vapp.empty-vm", "status_text", "RESOLVED"),
					testAccCheckVcdVappPowerState(testConfig.VCD.Org, testConfig.Nsxt.Vdc, t.Name()+"-empty-vm", []string{"POWERED_OFF"}),

					// Template vApp VM checks
					resource.TestCheckResourceAttr("vcd_vapp_vm.template-vm", "vm_type", "vcd_vapp_vm"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.template-vm", "name", t.Name()+"-template-vapp-vm"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.template-vm", "description", t.Name()+"-template-vapp-vm"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.template-vm", "cpu_hot_add_enabled", "false"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.template-vm", "memory_hot_add_enabled", "false"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.template-vm", "expose_hardware_virtualization", "false"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.template-vm", "prevent_update_power_off", "true"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.template-vm", "network.#", "2"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.template-vm", "network.0.type", "org"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.template-vm", "network.0.adapter_type", "VMXNET3"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.template-vm", "network.0.ip_allocation_mode", "POOL"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.template-vm", "network.1.type", "vapp"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.template-vm", "network.1.adapter_type", "E1000"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.template-vm", "network.1.ip_allocation_mode", "POOL"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.template-vm", "network.1.mac", "00:00:00:AA:BB:CC"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.template-vm", "power_on", "false"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.template-vm", "status", "8"), // 8 - means POWERED OFF
					resource.TestCheckResourceAttr("vcd_vapp_vm.template-vm", "status_text", "POWERED_OFF"),
					testAccCheckVcdVMPowerState(testConfig.VCD.Org, testConfig.Nsxt.Vdc, t.Name()+"-template-vm", t.Name()+"-template-vapp-vm", "POWERED_OFF"),

					// Empty vApp VM checks
					resource.TestCheckResourceAttr("vcd_vapp_vm.empty-vm", "vm_type", "vcd_vapp_vm"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.empty-vm", "name", t.Name()+"-empty-vapp-vm"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.empty-vm", "description", t.Name()+"-empty-vapp-vm"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.empty-vm", "computer_name", "vapp-vm"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.empty-vm", "cpus", "1"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.empty-vm", "memory", "1024"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.empty-vm", "os_type", "sles10_64Guest"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.empty-vm", "hardware_version", "vmx-14"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.empty-vm", "cpu_hot_add_enabled", "false"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.empty-vm", "memory_hot_add_enabled", "false"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.empty-vm", "expose_hardware_virtualization", "false"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.empty-vm", "prevent_update_power_off", "true"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.empty-vm", "network.#", "2"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.empty-vm", "network.0.type", "org"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.empty-vm", "network.0.adapter_type", "VMXNET3"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.empty-vm", "network.0.ip_allocation_mode", "POOL"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.empty-vm", "network.1.type", "vapp"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.empty-vm", "network.1.adapter_type", "E1000"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.empty-vm", "network.1.ip_allocation_mode", "POOL"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.empty-vm", "network.1.mac", "00:00:00:BB:AA:CC"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.empty-vm", "power_on", "false"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.empty-vm", "status", "8"), // 8 - means POWERED OFF
					resource.TestCheckResourceAttr("vcd_vapp_vm.empty-vm", "status_text", "POWERED_OFF"),
					testAccCheckVcdVMPowerState(testConfig.VCD.Org, testConfig.Nsxt.Vdc, t.Name()+"-empty-vm", t.Name()+"-empty-vapp-vm", "POWERED_OFF"),

					// Standalone template VM checks
					resource.TestCheckResourceAttr("vcd_vm.template-vm", "vm_type", "vcd_vm"),
					resource.TestCheckResourceAttr("vcd_vm.template-vm", "name", t.Name()+"-template-standalone-vm"),
					resource.TestCheckResourceAttr("vcd_vm.template-vm", "description", t.Name()+"-template-standalone-vm"),
					resource.TestCheckResourceAttr("vcd_vm.template-vm", "cpu_hot_add_enabled", "false"),
					resource.TestCheckResourceAttr("vcd_vm.template-vm", "memory_hot_add_enabled", "false"),
					resource.TestCheckResourceAttr("vcd_vm.template-vm", "expose_hardware_virtualization", "false"),
					resource.TestCheckResourceAttr("vcd_vm.template-vm", "prevent_update_power_off", "true"),
					resource.TestCheckResourceAttr("vcd_vm.template-vm", "network.#", "2"),
					resource.TestCheckResourceAttr("vcd_vm.template-vm", "network.0.type", "org"),
					resource.TestCheckResourceAttr("vcd_vm.template-vm", "network.0.adapter_type", "VMXNET3"),
					resource.TestCheckResourceAttr("vcd_vm.template-vm", "network.0.ip_allocation_mode", "POOL"),
					resource.TestCheckResourceAttr("vcd_vm.template-vm", "network.1.type", "org"),
					resource.TestCheckResourceAttr("vcd_vm.template-vm", "network.1.adapter_type", "E1000E"),
					resource.TestCheckResourceAttr("vcd_vm.template-vm", "network.1.ip_allocation_mode", "POOL"),
					resource.TestCheckResourceAttr("vcd_vm.template-vm", "network.1.mac", "00:00:00:11:22:33"),
					resource.TestCheckResourceAttr("vcd_vm.template-vm", "power_on", "false"),
					resource.TestCheckResourceAttr("vcd_vm.template-vm", "status", "8"), // 8 - means POWERED OFF
					resource.TestCheckResourceAttr("vcd_vm.template-vm", "status_text", "POWERED_OFF"),
					testAccCheckVcdVMPowerState(testConfig.VCD.Org, testConfig.Nsxt.Vdc, "", t.Name()+"-template-standalone-vm", "POWERED_OFF"),

					// Standalone empty VM checks
					resource.TestCheckResourceAttr("vcd_vm.empty-vm", "vm_type", "vcd_vm"),
					resource.TestCheckResourceAttr("vcd_vm.empty-vm", "name", t.Name()+"-empty-standalone-vm"),
					resource.TestCheckResourceAttr("vcd_vm.empty-vm", "description", t.Name()+"-standalone"),
					resource.TestCheckResourceAttr("vcd_vm.empty-vm", "cpus", "1"),
					resource.TestCheckResourceAttr("vcd_vm.empty-vm", "memory", "1024"),
					resource.TestCheckResourceAttr("vcd_vm.empty-vm", "os_type", "sles10_64Guest"),
					resource.TestCheckResourceAttr("vcd_vm.empty-vm", "hardware_version", "vmx-14"),
					resource.TestCheckResourceAttr("vcd_vm.empty-vm", "cpu_hot_add_enabled", "false"),
					resource.TestCheckResourceAttr("vcd_vm.empty-vm", "memory_hot_add_enabled", "false"),
					resource.TestCheckResourceAttr("vcd_vm.empty-vm", "expose_hardware_virtualization", "false"),
					resource.TestCheckResourceAttr("vcd_vm.empty-vm", "prevent_update_power_off", "true"),
					resource.TestCheckResourceAttr("vcd_vm.empty-vm", "network.#", "2"),
					resource.TestCheckResourceAttr("vcd_vm.empty-vm", "network.0.type", "org"),
					resource.TestCheckResourceAttr("vcd_vm.empty-vm", "network.0.adapter_type", "VMXNET3"),
					resource.TestCheckResourceAttr("vcd_vm.empty-vm", "network.0.ip_allocation_mode", "POOL"),
					resource.TestCheckResourceAttr("vcd_vm.empty-vm", "network.1.type", "org"),
					resource.TestCheckResourceAttr("vcd_vm.empty-vm", "network.1.adapter_type", "E1000E"),
					resource.TestCheckResourceAttr("vcd_vm.empty-vm", "network.1.ip_allocation_mode", "POOL"),
					resource.TestCheckResourceAttr("vcd_vm.empty-vm", "network.1.mac", "00:00:00:22:33:44"),
					resource.TestCheckResourceAttr("vcd_vm.empty-vm", "power_on", "false"),
					resource.TestCheckResourceAttr("vcd_vm.empty-vm", "status", "8"), // 8 - means POWERED OFF
					resource.TestCheckResourceAttr("vcd_vm.empty-vm", "status_text", "POWERED_OFF"),
					testAccCheckVcdVMPowerState(testConfig.VCD.Org, testConfig.Nsxt.Vdc, "", t.Name()+"-empty-standalone-vm", "POWERED_OFF"),

					// Check initial extra configuration items
					resource.TestCheckTypeSetElemNestedAttrs("vcd_vapp_vm.template-vm", "extra_config.*", map[string]string{
						"key":   "template-vapp-vm-1",
						"value": initialValue + "1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_vapp_vm.template-vm", "extra_config.*", map[string]string{
						"key":   "template-vapp-vm-2",
						"value": initialValue + "2",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_vm.template-vm", "extra_config.*", map[string]string{
						"key":   "template-vm-1",
						"value": initialValue + "1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_vm.template-vm", "extra_config.*", map[string]string{
						"key":   "template-vm-2",
						"value": initialValue + "2",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_vapp_vm.empty-vm", "extra_config.*", map[string]string{
						"key":   "empty-vapp-vm-1",
						"value": initialValue + "1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_vapp_vm.empty-vm", "extra_config.*", map[string]string{
						"key":   "empty-vapp-vm-2",
						"value": initialValue + "2",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_vm.empty-vm", "extra_config.*", map[string]string{
						"key":   "empty-vm-1",
						"value": initialValue + "1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_vm.empty-vm", "extra_config.*", map[string]string{
						"key":   "empty-vm-2",
						"value": initialValue + "2",
					}),
				),
			},
			// Update extra configuration items
			{
				Config: configTextStep2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckTypeSetElemNestedAttrs("vcd_vapp_vm.template-vm", "extra_config.*", map[string]string{
						"key":   "template-vapp-vm-1",
						"value": updatedValue + "1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_vapp_vm.template-vm", "extra_config.*", map[string]string{
						"key":   "template-vapp-vm-2",
						"value": updatedValue + "2",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_vm.template-vm", "extra_config.*", map[string]string{
						"key":   "template-vm-1",
						"value": updatedValue + "1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_vm.template-vm", "extra_config.*", map[string]string{
						"key":   "template-vm-2",
						"value": updatedValue + "2",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_vapp_vm.empty-vm", "extra_config.*", map[string]string{
						"key":   "empty-vapp-vm-1",
						"value": updatedValue + "1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_vapp_vm.empty-vm", "extra_config.*", map[string]string{
						"key":   "empty-vapp-vm-2",
						"value": updatedValue + "2",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_vm.empty-vm", "extra_config.*", map[string]string{
						"key":   "empty-vm-1",
						"value": updatedValue + "1",
					}),
					resource.TestCheckTypeSetElemNestedAttrs("vcd_vm.empty-vm", "extra_config.*", map[string]string{
						"key":   "empty-vm-2",
						"value": updatedValue + "2",
					}),
				),
			},
			// Remove extra configuration items
			{
				Config: configTextStep3,
				Check: resource.ComposeAggregateTestCheckFunc(
					checkExtraConfigExists(testConfig.Nsxt.Vdc, t.Name()+"-template-vapp-vm", map[string]string{"template-vapp-vm1": ""}, false),
					checkExtraConfigExists(testConfig.Nsxt.Vdc, t.Name()+"-template-standalone-vm", map[string]string{"template-vm1": ""}, false),
					checkExtraConfigExists(testConfig.Nsxt.Vdc, t.Name()+"-empty-vapp-vm", map[string]string{"empty-vapp-vm1": ""}, false),
					checkExtraConfigExists(testConfig.Nsxt.Vdc, t.Name()+"-empty-standalone-vm", map[string]string{"empty-vm1": ""}, false),
					checkExtraConfigExists(testConfig.Nsxt.Vdc, t.Name()+"-template-vapp-vm", map[string]string{"template-vapp-vm2": ""}, false),
					checkExtraConfigExists(testConfig.Nsxt.Vdc, t.Name()+"-template-standalone-vm", map[string]string{"template-vm2": ""}, false),
					checkExtraConfigExists(testConfig.Nsxt.Vdc, t.Name()+"-empty-vapp-vm", map[string]string{"empty-vapp-vm2": ""}, false),
					checkExtraConfigExists(testConfig.Nsxt.Vdc, t.Name()+"-empty-standalone-vm", map[string]string{"empty-vm2": ""}, false),
				),
			},
		},
	})
	postTestChecks(t)
}

func checkExtraConfigExists(vdcName, vmName string, data map[string]string, wantExist bool) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		conn := testAccProvider.Meta().(*VCDClient)

		vms, err := conn.Client.QueryVmList(types.VmQueryFilterOnlyDeployed)
		if err != nil {
			return err
		}
		vmHref := ""

		for _, vmRef := range vms {
			if vmRef.VdcName == vdcName && vmRef.Name == vmName {
				vmHref = vmRef.HREF
				break
			}
		}
		if vmHref == "" {
			return fmt.Errorf("could not find VM %s in VDC %s", vmName, vdcName)
		}

		vm, err := conn.Client.GetVMByHref(vmHref)
		if err != nil {
			return err
		}
		if vm.VM.VirtualHardwareSection == nil || len(vm.VM.VirtualHardwareSection.ExtraConfig) == 0 {
			return fmt.Errorf("no extra configuration found for VM %s", vmName)
		}
		for _, ec := range vm.VM.VirtualHardwareSection.ExtraConfig {
			found := 0
			for key, value := range data {
				if ec.Key == key {
					if wantExist {
						if ec.Value == value {
							found++
						}
					} else {
						return fmt.Errorf("key '%s' should not exist, but found in VM %s", key, vmName)
					}
				}
			}
			if wantExist && found < len(data) {
				return fmt.Errorf("key/value combinations not found in VM extra-config: wanted %d - found %d", len(data), found)
			}
		}
		return nil
	}
}

const testAccVcdVAppVm_extraconfig = `
{{.SkipMessage}}
data "vcd_org_vdc" "nsxt" {
  org  = "{{.Org}}"
  name = "{{.Vdc}}"
}

data "vcd_nsxt_edgegateway" "t1" {
  org      = "{{.Org}}"
  owner_id = data.vcd_org_vdc.nsxt.id
  name     = "{{.NsxtEdgeGateway}}"
}

data "vcd_catalog" "{{.Catalog}}" {
	org  = "{{.Org}}"
	name = "{{.Catalog}}"
}

data "vcd_catalog_vapp_template" "{{.CatalogItem}}" {
	org         = "{{.Org}}"
	catalog_id = data.vcd_catalog.{{.Catalog}}.id
	name       = "{{.CatalogItem}}"
}

data "vcd_catalog_media" "{{.Media}}" {
	org     = "{{.Org}}"
	catalog = data.vcd_catalog.{{.Catalog}}.name
	name    = "{{.Media}}"
}

resource "vcd_network_routed_v2" "nsxt-backed" {
  org = "{{.Org}}"

  edge_gateway_id = data.vcd_nsxt_edgegateway.t1.id

  name = "{{.TestName}}"

  gateway       = "1.1.1.1"
  prefix_length = 24

  static_ip_pool {
    start_address = "1.1.1.10"
    end_address   = "1.1.1.40"
  }
}

resource "vcd_vapp" "template-vm" {
  org         = "{{.Org}}"
  vdc         = "{{.Vdc}}"
  name        = "{{.TestName}}-template-vm"
  description = "vApp for Template VM description"
  power_on    = false
}

resource "vcd_vapp_network" "template" {
  org  = "{{.Org}}"
  vdc  = "{{.Vdc}}"
  name = "{{.TestName}}-template-vm"

  vapp_name          = (vcd_vapp.template-vm.id == "always-not-equal" ? null : vcd_vapp.template-vm.name)
  gateway            = "192.168.3.1"
  netmask            = "255.255.255.0"

  static_ip_pool {
	start_address = "192.168.3.51"
	end_address   = "192.168.3.100"
  }

  depends_on = [vcd_vapp.template-vm]
}

resource "vcd_vapp_org_network" "template-vapp" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  vapp_name        = (vcd_vapp.template-vm.id == "always-not-equal" ? null : vcd_vapp.template-vm.name)
  org_network_name = (vcd_network_routed_v2.nsxt-backed.id == "always-not-equal" ? null : vcd_network_routed_v2.nsxt-backed.name)
}

resource "vcd_vapp" "empty-vm" {
  org         = "{{.Org}}"
  vdc         = "{{.Vdc}}"
  name        = "{{.TestName}}-empty-vm"
  description = "vApp for Empty VM description"
  power_on    = false
}

resource "vcd_vapp_network" "empty-vm" {
  org  = "{{.Org}}"
  vdc  = "{{.Vdc}}"
  name = "{{.TestName}}-empty-vm"

  vapp_name          = (vcd_vapp.empty-vm.id == "always-not-equal" ? null : vcd_vapp.empty-vm.name)
  gateway            = "192.168.2.1"
  netmask            = "255.255.255.0"

  static_ip_pool {
	start_address = "192.168.2.51"
	end_address   = "192.168.2.100"
  }

  depends_on = [vcd_vapp.empty-vm]
}

resource "vcd_vapp_org_network" "empty-vapp" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  vapp_name        = (vcd_vapp.empty-vm.id == "always-not-equal" ? null : vcd_vapp.empty-vm.name)
  org_network_name = (vcd_network_routed_v2.nsxt-backed.id == "always-not-equal" ? null : vcd_network_routed_v2.nsxt-backed.name)
}

resource "vcd_vapp_vm" "template-vm" {
  org  = "{{.Org}}"
  vdc  = "{{.Vdc}}"

  vapp_template_id = data.vcd_catalog_vapp_template.{{.CatalogItem}}.id
  
  vapp_name   = vcd_vapp.template-vm.name
  name        = "{{.TestName}}-template-vapp-vm"
  description = "{{.TestName}}-template-vapp-vm"
  power_on    = false

  network {
	type               = "org"
	name               = (vcd_vapp_org_network.template-vapp.id == "always-not-equal" ? null : vcd_vapp_org_network.template-vapp.org_network_name)
	adapter_type       = "VMXNET3"
	ip_allocation_mode = "POOL"
  }

  network {
	type               = "vapp"
	name               = (vcd_vapp_network.template.id == "always-not-equal" ? null : vcd_vapp_network.template.name)
	adapter_type       = "E1000"
	ip_allocation_mode = "POOL"
	mac                = "00:00:00:AA:BB:CC"
  }

  set_extra_config {
    key   = "template-vapp-vm-1"
    value = "{{.ExtraConfigValue1}}"
  }

  set_extra_config {
    key      = "template-vapp-vm-2"
    value    = "{{.ExtraConfigValue2}}"
    required = true
  }

  depends_on = [vcd_vapp_network.template]

  prevent_update_power_off = true
}

resource "vcd_vapp_vm" "empty-vm" {
  org  = "{{.Org}}"
  vdc  = "{{.Vdc}}"
  
  vapp_name     = vcd_vapp.empty-vm.name
  name          = "{{.TestName}}-empty-vapp-vm"
  description   = "{{.TestName}}-empty-vapp-vm"
  computer_name = "vapp-vm"
  power_on      = false

  cpus   = 1
  memory = 1024

  os_type          = "sles10_64Guest"
  hardware_version = "vmx-14"
  boot_image_id    = data.vcd_catalog_media.{{.Media}}.id

  network {
	type               = "org"
	name               = (vcd_vapp_org_network.empty-vapp.id == "always-not-equal" ? null : vcd_vapp_org_network.empty-vapp.org_network_name)
	adapter_type       = "VMXNET3"
	ip_allocation_mode = "POOL"
  }

  network {
	type               = "vapp"
	name               = (vcd_vapp_network.empty-vm.id == "always-not-equal" ? null : vcd_vapp_network.empty-vm.name)
	adapter_type       = "E1000"
	ip_allocation_mode = "POOL"
	mac                = "00:00:00:BB:AA:CC"
  }

  set_extra_config {
    key   = "empty-vapp-vm-1"
    value = "{{.ExtraConfigValue1}}"
  }

  set_extra_config {
    key      = "empty-vapp-vm-2"
    value    = "{{.ExtraConfigValue2}}"
    required = true
  }

  depends_on = [vcd_vapp_network.empty-vm]

  prevent_update_power_off = true
}

resource "vcd_vm" "template-vm" {
  org  = "{{.Org}}"
  vdc  = "{{.Vdc}}"

  vapp_template_id = data.vcd_catalog_vapp_template.{{.CatalogItem}}.id
  
  name        = "{{.TestName}}-template-standalone-vm"
  description = "{{.TestName}}-template-standalone-vm"
  power_on    = false

  network {
	type               = "org"
	name               = (vcd_network_routed_v2.nsxt-backed.id == "always-not-equal" ? null : vcd_network_routed_v2.nsxt-backed.name)
	adapter_type       = "VMXNET3"
	ip_allocation_mode = "POOL"
  }

  network {
	type               = "org"
	name               = (vcd_network_routed_v2.nsxt-backed.id == "always-not-equal" ? null : vcd_network_routed_v2.nsxt-backed.name)
	adapter_type       = "E1000E"
	ip_allocation_mode = "POOL"
	mac                = "00:00:00:11:22:33"
  }

  set_extra_config {
    key   = "template-vm-1"
    value = "{{.ExtraConfigValue1}}"
  }

  set_extra_config {
    key      = "template-vm-2"
    value    = "{{.ExtraConfigValue2}}"
    required = true
  }

  prevent_update_power_off = true
}

resource "vcd_vm" "empty-vm" {
  org  = "{{.Org}}"
  vdc  = "{{.Vdc}}"

  name          = "{{.TestName}}-empty-standalone-vm"
  description   = "{{.TestName}}-standalone"
  computer_name = "standalone"
  power_on      = false

  cpus   = 1
  memory = 1024

  os_type          = "sles10_64Guest"
  hardware_version = "vmx-14"
  boot_image_id    = data.vcd_catalog_media.{{.Media}}.id

  network {
	type               = "org"
	name               = (vcd_network_routed_v2.nsxt-backed.id == "always-not-equal" ? null : vcd_network_routed_v2.nsxt-backed.name)
	adapter_type       = "VMXNET3"
	ip_allocation_mode = "POOL"
  }

  network {
	type               = "org"
	name               = (vcd_network_routed_v2.nsxt-backed.id == "always-not-equal" ? null : vcd_network_routed_v2.nsxt-backed.name)
	adapter_type       = "E1000E"
	ip_allocation_mode = "POOL"
	mac                = "00:00:00:22:33:44"
  }

  set_extra_config {
    key   = "empty-vm-1"
    value = "{{.ExtraConfigValue1}}"
  }

  set_extra_config {
    key      = "empty-vm-2"
    value    = "{{.ExtraConfigValue2}}"
    required = true
  }


  prevent_update_power_off = true
}
`
