//go:build vapp || vm || standaloneVm || ALL || functional

package vcd

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccVcdVAppVm_4types_Ipv6(t *testing.T) {
	preTestChecks(t)

	var params = StringMap{
		"TestName":        t.Name(),
		"Org":             testConfig.VCD.Org,
		"Vdc":             testConfig.Nsxt.Vdc,
		"Catalog":         testConfig.VCD.Catalog.NsxtBackedCatalogName,
		"CatalogItem":     testConfig.VCD.Catalog.CatalogItemWithMultiVms,
		"Media":           testConfig.Media.NsxtBackedMediaName,
		"NsxtEdgeGateway": testConfig.Nsxt.EdgeGateway,

		"Tags": "vapp vm",
	}
	testParamsNotEmpty(t, params)

	configTextStep1 := templateFill(testAccVcdVAppVm_4types_ipv6Step1, params)
	debugPrintf("#[DEBUG] CONFIGURATION Step 1: %s\n", configTextStep1)

	params["FuncName"] = t.Name() + "-step2"
	configTextStep2 := templateFill(testAccVcdVAppVm_4types_ipv6Step2DS, params)
	debugPrintf("#[DEBUG] CONFIGURATION Step 2: %s\n", configTextStep2)

	params["FuncName"] = t.Name() + "-step3"
	configTextStep3 := templateFill(testAccVcdVAppVm_4types_ipv6Step3, params)
	debugPrintf("#[DEBUG] CONFIGURATION Step 3: %s\n", configTextStep3)

	params["FuncName"] = t.Name() + "-step4"
	configTextStep4 := templateFill(testAccVcdVAppVm_4types_ipv6Step4DS, params)
	debugPrintf("#[DEBUG] CONFIGURATION Step 4: %s\n", configTextStep4)

	if vcdShortTest {
		t.Skip(acceptanceTestsSkipped)
		return
	}

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
					resource.TestCheckResourceAttr("vcd_vapp_vm.template-vm", "network.0.secondary_ip_allocation_mode", "POOL"),
					resource.TestMatchResourceAttr("vcd_vapp_vm.template-vm", "network.0.secondary_ip", regexp.MustCompile("^2002:0:0:1234:abcd:ffff:c0a6:")),
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
					resource.TestMatchResourceAttr("vcd_vapp_vm.empty-vm", "network.0.secondary_ip", regexp.MustCompile("^2002:0:0:1234:abcd:ffff:c0a6:")),
					resource.TestCheckResourceAttr("vcd_vapp_vm.empty-vm", "network.0.secondary_ip_allocation_mode", "POOL"),
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
					resource.TestCheckResourceAttr("vcd_vm.template-vm", "network.0.secondary_ip_allocation_mode", "POOL"),
					resource.TestMatchResourceAttr("vcd_vm.template-vm", "network.0.secondary_ip", regexp.MustCompile("^2002:0:0:1234:abcd:ffff:c0a6:")),
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
					resource.TestMatchResourceAttr("vcd_vm.empty-vm", "network.0.secondary_ip", regexp.MustCompile("^2002:0:0:1234:abcd:ffff:c0a6:")),
					resource.TestCheckResourceAttr("vcd_vm.empty-vm", "network.0.secondary_ip_allocation_mode", "POOL"),
					resource.TestCheckResourceAttr("vcd_vm.empty-vm", "network.1.type", "org"),
					resource.TestCheckResourceAttr("vcd_vm.empty-vm", "network.1.adapter_type", "E1000E"),
					resource.TestCheckResourceAttr("vcd_vm.empty-vm", "network.1.ip_allocation_mode", "POOL"),
					resource.TestCheckResourceAttr("vcd_vm.empty-vm", "network.1.mac", "00:00:00:22:33:44"),
					resource.TestCheckResourceAttr("vcd_vm.empty-vm", "power_on", "false"),
					resource.TestCheckResourceAttr("vcd_vm.empty-vm", "status", "8"), // 8 - means POWERED OFF
					resource.TestCheckResourceAttr("vcd_vm.empty-vm", "status_text", "POWERED_OFF"),
					testAccCheckVcdVMPowerState(testConfig.VCD.Org, testConfig.Nsxt.Vdc, "", t.Name()+"-empty-standalone-vm", "POWERED_OFF"),

					// VM Copy checks

					// vApp checks
					resource.TestCheckResourceAttr("vcd_vapp.vm-copy-destination-template-vm", "name", t.Name()+"-vm-copy-template-destination"),
					resource.TestCheckResourceAttr("vcd_vapp.vm-copy-destination-template-vm", "description", "vApp destination for VM Copy"),
					resource.TestCheckResourceAttr("vcd_vapp.vm-copy-destination-template-vm", "power_on", "false"),
					resource.TestCheckResourceAttr("vcd_vapp.vm-copy-destination-template-vm", "status", "1"), // 1 - means RESOLVED
					resource.TestCheckResourceAttr("vcd_vapp.vm-copy-destination-template-vm", "status_text", "RESOLVED"),
					testAccCheckVcdVappPowerState(testConfig.VCD.Org, testConfig.Nsxt.Vdc, t.Name()+"-vm-copy-template-destination", []string{"POWERED_OFF"}),

					resource.TestCheckResourceAttr("vcd_vapp.vm-copy-destination-empty-vm", "name", t.Name()+"-vm-copy-empty-destination"),
					resource.TestCheckResourceAttr("vcd_vapp.vm-copy-destination-empty-vm", "description", "vApp destination for VM Copy"),
					resource.TestCheckResourceAttr("vcd_vapp.vm-copy-destination-empty-vm", "power_on", "false"),
					resource.TestCheckResourceAttr("vcd_vapp.vm-copy-destination-empty-vm", "status", "1"), // 1 - means RESOLVED
					resource.TestCheckResourceAttr("vcd_vapp.vm-copy-destination-empty-vm", "status_text", "RESOLVED"),
					testAccCheckVcdVappPowerState(testConfig.VCD.Org, testConfig.Nsxt.Vdc, t.Name()+"-vm-copy-empty-destination", []string{"POWERED_OFF"}),

					// Template vApp VM checks
					resource.TestCheckResourceAttr("vcd_vapp_vm.template-vm-copy", "vm_type", "vcd_vapp_vm"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.template-vm-copy", "name", t.Name()+"-template-vapp-vm-copy"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.template-vm-copy", "description", t.Name()+"-template-vapp-vm"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.template-vm-copy", "cpu_hot_add_enabled", "false"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.template-vm-copy", "memory_hot_add_enabled", "false"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.template-vm-copy", "expose_hardware_virtualization", "false"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.template-vm-copy", "prevent_update_power_off", "true"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.template-vm-copy", "network.#", "2"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.template-vm-copy", "network.0.type", "org"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.template-vm-copy", "network.0.adapter_type", "VMXNET3"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.template-vm-copy", "network.0.ip_allocation_mode", "POOL"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.template-vm-copy", "network.0.secondary_ip_allocation_mode", "POOL"),
					resource.TestMatchResourceAttr("vcd_vapp_vm.template-vm-copy", "network.0.secondary_ip", regexp.MustCompile("^2002:0:0:1234:abcd:ffff:c0a6:")),
					resource.TestCheckResourceAttr("vcd_vapp_vm.template-vm-copy", "network.1.type", "vapp"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.template-vm-copy", "network.1.adapter_type", "E1000"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.template-vm-copy", "network.1.ip_allocation_mode", "POOL"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.template-vm-copy", "network.1.mac", "00:00:00:AA:AC:CC"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.template-vm-copy", "power_on", "false"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.template-vm-copy", "status", "8"), // 8 - means POWERED OFF
					resource.TestCheckResourceAttr("vcd_vapp_vm.template-vm-copy", "status_text", "POWERED_OFF"),
					testAccCheckVcdVMPowerState(testConfig.VCD.Org, testConfig.Nsxt.Vdc, t.Name()+"-vm-copy-template-destination", t.Name()+"-template-vapp-vm-copy", "POWERED_OFF"),

					// Empty vApp VM checks
					resource.TestCheckResourceAttr("vcd_vapp_vm.empty-vm-copy", "vm_type", "vcd_vapp_vm"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.empty-vm-copy", "name", t.Name()+"-empty-vapp-vm-copy"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.empty-vm-copy", "description", t.Name()+"-empty-vapp-vm"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.empty-vm-copy", "computer_name", "vapp-vm"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.empty-vm-copy", "cpus", "1"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.empty-vm-copy", "memory", "1024"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.empty-vm-copy", "os_type", "sles10_64Guest"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.empty-vm-copy", "hardware_version", "vmx-14"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.empty-vm-copy", "cpu_hot_add_enabled", "false"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.empty-vm-copy", "memory_hot_add_enabled", "false"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.empty-vm-copy", "expose_hardware_virtualization", "false"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.empty-vm-copy", "prevent_update_power_off", "true"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.empty-vm-copy", "network.#", "2"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.empty-vm-copy", "network.0.type", "org"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.empty-vm-copy", "network.0.adapter_type", "VMXNET3"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.empty-vm-copy", "network.0.ip_allocation_mode", "POOL"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.empty-vm-copy", "network.0.secondary_ip_allocation_mode", "POOL"),
					resource.TestMatchResourceAttr("vcd_vapp_vm.empty-vm-copy", "network.0.secondary_ip", regexp.MustCompile("^2002:0:0:1234:abcd:ffff:c0a6:")),
					resource.TestCheckResourceAttr("vcd_vapp_vm.empty-vm-copy", "network.1.type", "vapp"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.empty-vm-copy", "network.1.adapter_type", "E1000"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.empty-vm-copy", "network.1.ip_allocation_mode", "POOL"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.empty-vm-copy", "network.1.mac", "00:00:00:AA:BB:FC"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.empty-vm-copy", "power_on", "false"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.empty-vm-copy", "status", "8"), // 8 - means POWERED OFF
					resource.TestCheckResourceAttr("vcd_vapp_vm.empty-vm-copy", "status_text", "POWERED_OFF"),
					testAccCheckVcdVMPowerState(testConfig.VCD.Org, testConfig.Nsxt.Vdc, t.Name()+"-vm-copy-empty-destination", t.Name()+"-empty-vapp-vm-copy", "POWERED_OFF"),

					// Standalone template VM checks
					resource.TestCheckResourceAttr("vcd_vm.template-vm-copy", "vm_type", "vcd_vm"),
					resource.TestCheckResourceAttr("vcd_vm.template-vm-copy", "name", t.Name()+"-template-standalone-vm-copy"),
					resource.TestCheckResourceAttr("vcd_vm.template-vm-copy", "description", t.Name()+"-template-standalone-vm"),
					resource.TestCheckResourceAttr("vcd_vm.template-vm-copy", "cpu_hot_add_enabled", "false"),
					resource.TestCheckResourceAttr("vcd_vm.template-vm-copy", "memory_hot_add_enabled", "false"),
					resource.TestCheckResourceAttr("vcd_vm.template-vm-copy", "expose_hardware_virtualization", "false"),
					resource.TestCheckResourceAttr("vcd_vm.template-vm-copy", "prevent_update_power_off", "true"),
					resource.TestCheckResourceAttr("vcd_vm.template-vm-copy", "network.#", "2"),
					resource.TestCheckResourceAttr("vcd_vm.template-vm-copy", "network.0.type", "org"),
					resource.TestCheckResourceAttr("vcd_vm.template-vm-copy", "network.0.adapter_type", "VMXNET3"),
					resource.TestCheckResourceAttr("vcd_vm.template-vm-copy", "network.0.ip_allocation_mode", "POOL"),
					resource.TestCheckResourceAttr("vcd_vm.template-vm-copy", "network.0.secondary_ip_allocation_mode", "POOL"),
					resource.TestMatchResourceAttr("vcd_vm.template-vm-copy", "network.0.secondary_ip", regexp.MustCompile("^2002:0:0:1234:abcd:ffff:c0a6:")),
					resource.TestCheckResourceAttr("vcd_vm.template-vm-copy", "network.1.type", "org"),
					resource.TestCheckResourceAttr("vcd_vm.template-vm-copy", "network.1.adapter_type", "E1000E"),
					resource.TestCheckResourceAttr("vcd_vm.template-vm-copy", "network.1.ip_allocation_mode", "POOL"),
					resource.TestCheckResourceAttr("vcd_vm.template-vm-copy", "network.1.mac", "00:00:00:11:FF:33"),
					resource.TestCheckResourceAttr("vcd_vm.template-vm-copy", "power_on", "false"),
					resource.TestCheckResourceAttr("vcd_vm.template-vm-copy", "status", "8"), // 8 - means POWERED OFF
					resource.TestCheckResourceAttr("vcd_vm.template-vm-copy", "status_text", "POWERED_OFF"),
					testAccCheckVcdVMPowerState(testConfig.VCD.Org, testConfig.Nsxt.Vdc, "", t.Name()+"-template-standalone-vm-copy", "POWERED_OFF"),

					// Standalone empty VM checks
					resource.TestCheckResourceAttr("vcd_vm.empty-vm-copy", "vm_type", "vcd_vm"),
					resource.TestCheckResourceAttr("vcd_vm.empty-vm-copy", "name", t.Name()+"-empty-standalone-vm-copy"),
					resource.TestCheckResourceAttr("vcd_vm.empty-vm-copy", "description", t.Name()+"-standalone"),
					resource.TestCheckResourceAttr("vcd_vm.empty-vm-copy", "cpus", "1"),
					resource.TestCheckResourceAttr("vcd_vm.empty-vm-copy", "memory", "1024"),
					resource.TestCheckResourceAttr("vcd_vm.empty-vm-copy", "os_type", "sles10_64Guest"),
					resource.TestCheckResourceAttr("vcd_vm.empty-vm-copy", "hardware_version", "vmx-14"),
					resource.TestCheckResourceAttr("vcd_vm.empty-vm-copy", "cpu_hot_add_enabled", "false"),
					resource.TestCheckResourceAttr("vcd_vm.empty-vm-copy", "memory_hot_add_enabled", "false"),
					resource.TestCheckResourceAttr("vcd_vm.empty-vm-copy", "expose_hardware_virtualization", "false"),
					resource.TestCheckResourceAttr("vcd_vm.empty-vm-copy", "prevent_update_power_off", "true"),
					resource.TestCheckResourceAttr("vcd_vm.empty-vm-copy", "network.#", "2"),
					resource.TestCheckResourceAttr("vcd_vm.empty-vm-copy", "network.0.type", "org"),
					resource.TestCheckResourceAttr("vcd_vm.empty-vm-copy", "network.0.adapter_type", "VMXNET3"),
					resource.TestCheckResourceAttr("vcd_vm.empty-vm-copy", "network.0.ip_allocation_mode", "POOL"),
					resource.TestCheckResourceAttr("vcd_vm.empty-vm-copy", "network.0.secondary_ip_allocation_mode", "POOL"),
					resource.TestMatchResourceAttr("vcd_vm.empty-vm-copy", "network.0.secondary_ip", regexp.MustCompile("^2002:0:0:1234:abcd:ffff:c0a6:")),
					resource.TestCheckResourceAttr("vcd_vm.empty-vm-copy", "network.1.type", "org"),
					resource.TestCheckResourceAttr("vcd_vm.empty-vm-copy", "network.1.adapter_type", "E1000E"),
					resource.TestCheckResourceAttr("vcd_vm.empty-vm-copy", "network.1.ip_allocation_mode", "POOL"),
					resource.TestCheckResourceAttr("vcd_vm.empty-vm-copy", "network.1.mac", "00:00:00:22:33:44"),
					resource.TestCheckResourceAttr("vcd_vm.empty-vm-copy", "power_on", "false"),
					resource.TestCheckResourceAttr("vcd_vm.empty-vm-copy", "status", "8"), // 8 - means POWERED OFF
					resource.TestCheckResourceAttr("vcd_vm.empty-vm-copy", "status_text", "POWERED_OFF"),
					testAccCheckVcdVMPowerState(testConfig.VCD.Org, testConfig.Nsxt.Vdc, "", t.Name()+"-empty-standalone-vm-copy", "POWERED_OFF"),
				),
			},
			{
				Config: configTextStep2,
				Check: resource.ComposeAggregateTestCheckFunc(
					resourceFieldsEqual("data.vcd_vapp_vm.template-vm", "vcd_vapp_vm.template-vm", []string{"%"}),
					resourceFieldsEqual("data.vcd_vapp_vm.empty-vm", "vcd_vapp_vm.empty-vm", []string{"%"}),
					resourceFieldsEqual("data.vcd_vm.template-vm", "vcd_vm.template-vm", []string{"%"}),
					resourceFieldsEqual("data.vcd_vm.empty-vm", "vcd_vm.empty-vm", []string{"%"}),
					resourceFieldsEqual("data.vcd_vapp_vm.template-vm-copy", "vcd_vapp_vm.template-vm-copy", []string{"%"}),
					resourceFieldsEqual("data.vcd_vapp_vm.empty-vm-copy", "vcd_vapp_vm.empty-vm-copy", []string{"%"}),
					resourceFieldsEqual("data.vcd_vm.template-vm-copy", "vcd_vm.template-vm-copy", []string{"%"}),
					resourceFieldsEqual("data.vcd_vm.empty-vm-copy", "vcd_vm.empty-vm-copy", []string{"%"}),
				),
			},
			{
				Config: configTextStep3,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Template vApp VM checks
					resource.TestCheckResourceAttr("vcd_vapp_vm.template-vm", "network.#", "2"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.template-vm", "network.0.type", "org"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.template-vm", "network.0.adapter_type", "VMXNET3"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.template-vm", "network.0.ip_allocation_mode", "POOL"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.template-vm", "network.0.secondary_ip_allocation_mode", "MANUAL"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.template-vm", "network.0.secondary_ip", "2002:0:0:1234:abcd:ffff:c0a6:138"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.template-vm", "network.1.type", "vapp"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.template-vm", "network.1.adapter_type", "E1000"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.template-vm", "network.1.ip_allocation_mode", "POOL"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.template-vm", "network.1.mac", "00:00:00:AA:BB:CC"),

					// Empty vApp VM checks
					resource.TestCheckResourceAttr("vcd_vapp_vm.empty-vm", "network.#", "2"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.empty-vm", "network.0.type", "org"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.empty-vm", "network.0.adapter_type", "VMXNET3"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.empty-vm", "network.0.ip_allocation_mode", "POOL"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.empty-vm", "network.0.secondary_ip_allocation_mode", "MANUAL"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.empty-vm", "network.0.secondary_ip", "2002:0:0:1234:abcd:ffff:c0a6:139"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.empty-vm", "network.1.type", "vapp"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.empty-vm", "network.1.adapter_type", "E1000"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.empty-vm", "network.1.ip_allocation_mode", "POOL"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.empty-vm", "network.1.mac", "00:00:00:BB:AA:CC"),

					// Standalone template VM checks
					resource.TestCheckResourceAttr("vcd_vm.template-vm", "network.#", "2"),
					resource.TestCheckResourceAttr("vcd_vm.template-vm", "network.0.type", "org"),
					resource.TestCheckResourceAttr("vcd_vm.template-vm", "network.0.adapter_type", "VMXNET3"),
					resource.TestCheckResourceAttr("vcd_vm.template-vm", "network.0.ip_allocation_mode", "POOL"),
					resource.TestCheckResourceAttr("vcd_vm.template-vm", "network.0.secondary_ip_allocation_mode", "MANUAL"),
					resource.TestCheckResourceAttr("vcd_vm.template-vm", "network.0.secondary_ip", "2002:0:0:1234:abcd:ffff:c0a6:13a"),
					resource.TestCheckResourceAttr("vcd_vm.template-vm", "network.1.type", "org"),
					resource.TestCheckResourceAttr("vcd_vm.template-vm", "network.1.adapter_type", "E1000E"),
					resource.TestCheckResourceAttr("vcd_vm.template-vm", "network.1.ip_allocation_mode", "POOL"),
					resource.TestCheckResourceAttr("vcd_vm.template-vm", "network.1.mac", "00:00:00:11:22:33"),

					// Standalone empty VM checks
					resource.TestCheckResourceAttr("vcd_vm.empty-vm", "network.#", "2"),
					resource.TestCheckResourceAttr("vcd_vm.empty-vm", "network.0.type", "org"),
					resource.TestCheckResourceAttr("vcd_vm.empty-vm", "network.0.adapter_type", "VMXNET3"),
					resource.TestCheckResourceAttr("vcd_vm.empty-vm", "network.0.ip_allocation_mode", "POOL"),
					resource.TestCheckResourceAttr("vcd_vm.empty-vm", "network.0.secondary_ip_allocation_mode", "MANUAL"),
					resource.TestCheckResourceAttr("vcd_vm.empty-vm", "network.0.secondary_ip", "2002:0:0:1234:abcd:ffff:c0a6:13b"),
					resource.TestCheckResourceAttr("vcd_vm.empty-vm", "network.1.type", "org"),
					resource.TestCheckResourceAttr("vcd_vm.empty-vm", "network.1.adapter_type", "E1000E"),
					resource.TestCheckResourceAttr("vcd_vm.empty-vm", "network.1.ip_allocation_mode", "POOL"),
					resource.TestCheckResourceAttr("vcd_vm.empty-vm", "network.1.mac", "00:00:00:22:33:44"),

					// Template vApp VM checks
					resource.TestCheckResourceAttr("vcd_vapp_vm.template-vm-copy", "network.#", "2"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.template-vm-copy", "network.0.type", "org"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.template-vm-copy", "network.0.adapter_type", "VMXNET3"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.template-vm-copy", "network.0.ip_allocation_mode", "POOL"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.template-vm-copy", "network.0.secondary_ip_allocation_mode", "MANUAL"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.template-vm-copy", "network.0.secondary_ip", "2002:0:0:1234:abcd:ffff:c0a6:13c"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.template-vm-copy", "network.1.type", "vapp"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.template-vm-copy", "network.1.adapter_type", "E1000"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.template-vm-copy", "network.1.ip_allocation_mode", "POOL"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.template-vm-copy", "network.1.mac", "00:00:00:AA:AC:CC"),

					// Empty vApp VM checks
					resource.TestCheckResourceAttr("vcd_vapp_vm.empty-vm-copy", "network.#", "2"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.empty-vm-copy", "network.0.type", "org"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.empty-vm-copy", "network.0.adapter_type", "VMXNET3"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.empty-vm-copy", "network.0.ip_allocation_mode", "POOL"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.empty-vm-copy", "network.0.secondary_ip_allocation_mode", "MANUAL"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.empty-vm-copy", "network.0.secondary_ip", "2002:0:0:1234:abcd:ffff:c0a6:13d"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.empty-vm-copy", "network.1.type", "vapp"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.empty-vm-copy", "network.1.adapter_type", "E1000"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.empty-vm-copy", "network.1.ip_allocation_mode", "POOL"),
					resource.TestCheckResourceAttr("vcd_vapp_vm.empty-vm-copy", "network.1.mac", "00:00:00:AA:BB:FC"),

					// Standalone template VM checks
					resource.TestCheckResourceAttr("vcd_vm.template-vm-copy", "network.#", "2"),
					resource.TestCheckResourceAttr("vcd_vm.template-vm-copy", "network.0.type", "org"),
					resource.TestCheckResourceAttr("vcd_vm.template-vm-copy", "network.0.adapter_type", "VMXNET3"),
					resource.TestCheckResourceAttr("vcd_vm.template-vm-copy", "network.0.ip_allocation_mode", "POOL"),
					resource.TestCheckResourceAttr("vcd_vm.template-vm-copy", "network.0.secondary_ip_allocation_mode", "MANUAL"),
					resource.TestCheckResourceAttr("vcd_vm.template-vm-copy", "network.0.secondary_ip", "2002:0:0:1234:abcd:ffff:c0a6:13e"),
					resource.TestCheckResourceAttr("vcd_vm.template-vm-copy", "network.1.type", "org"),
					resource.TestCheckResourceAttr("vcd_vm.template-vm-copy", "network.1.adapter_type", "E1000E"),
					resource.TestCheckResourceAttr("vcd_vm.template-vm-copy", "network.1.ip_allocation_mode", "POOL"),
					resource.TestCheckResourceAttr("vcd_vm.template-vm-copy", "network.1.mac", "00:00:00:11:FF:33"),

					// Standalone empty VM checks
					resource.TestCheckResourceAttr("vcd_vm.empty-vm-copy", "network.#", "2"),
					resource.TestCheckResourceAttr("vcd_vm.empty-vm-copy", "network.0.type", "org"),
					resource.TestCheckResourceAttr("vcd_vm.empty-vm-copy", "network.0.adapter_type", "VMXNET3"),
					resource.TestCheckResourceAttr("vcd_vm.empty-vm-copy", "network.0.ip_allocation_mode", "POOL"),
					resource.TestCheckResourceAttr("vcd_vm.empty-vm-copy", "network.0.secondary_ip_allocation_mode", "MANUAL"),
					resource.TestCheckResourceAttr("vcd_vm.empty-vm-copy", "network.0.secondary_ip", "2002:0:0:1234:abcd:ffff:c0a6:13f"),
					resource.TestCheckResourceAttr("vcd_vm.empty-vm-copy", "network.1.type", "org"),
					resource.TestCheckResourceAttr("vcd_vm.empty-vm-copy", "network.1.adapter_type", "E1000E"),
					resource.TestCheckResourceAttr("vcd_vm.empty-vm-copy", "network.1.ip_allocation_mode", "POOL"),
					resource.TestCheckResourceAttr("vcd_vm.empty-vm-copy", "network.1.mac", "00:00:00:22:33:44"),
				),
			},
			{
				Config: configTextStep4,
				Check: resource.ComposeAggregateTestCheckFunc(
					resourceFieldsEqual("data.vcd_vapp_vm.template-vm", "vcd_vapp_vm.template-vm", []string{"%"}),
					resourceFieldsEqual("data.vcd_vapp_vm.empty-vm", "vcd_vapp_vm.empty-vm", []string{"%"}),
					resourceFieldsEqual("data.vcd_vm.template-vm", "vcd_vm.template-vm", []string{"%"}),
					resourceFieldsEqual("data.vcd_vm.empty-vm", "vcd_vm.empty-vm", []string{"%"}),
					resourceFieldsEqual("data.vcd_vapp_vm.template-vm-copy", "vcd_vapp_vm.template-vm-copy", []string{"%"}),
					resourceFieldsEqual("data.vcd_vapp_vm.empty-vm-copy", "vcd_vapp_vm.empty-vm-copy", []string{"%"}),
					resourceFieldsEqual("data.vcd_vm.template-vm-copy", "vcd_vm.template-vm-copy", []string{"%"}),
					resourceFieldsEqual("data.vcd_vm.empty-vm-copy", "vcd_vm.empty-vm-copy", []string{"%"}),
				),
			},
		},
	})
	postTestChecks(t)
}

const testAccVcdVAppVm_4types_ipv6Common = `
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

  dual_stack_enabled      = true
  secondary_gateway       = "2002:0:0:1234:abcd:ffff:c0a6:121"
  secondary_prefix_length = 123

  secondary_static_ip_pool {
    start_address = "2002:0:0:1234:abcd:ffff:c0a6:122"
    end_address   = "2002:0:0:1234:abcd:ffff:c0a6:137"
  }
}
`

const testAccVcdVAppVm_4types_ipv6Step1 = testAccVcdVAppVm_4types_ipv6Common + `
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
	
	secondary_ip_allocation_mode = "POOL"
  }

  network {
	type               = "vapp"
	name               = (vcd_vapp_network.template.id == "always-not-equal" ? null : vcd_vapp_network.template.name)
	adapter_type       = "E1000"
	ip_allocation_mode = "POOL"
	mac                = "00:00:00:AA:BB:CC"
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

	secondary_ip_allocation_mode = "POOL"
  }

  network {
	type               = "vapp"
	name               = (vcd_vapp_network.empty-vm.id == "always-not-equal" ? null : vcd_vapp_network.empty-vm.name)
	adapter_type       = "E1000"
	ip_allocation_mode = "POOL"
	mac                = "00:00:00:BB:AA:CC"
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

	secondary_ip_allocation_mode = "POOL"
  }

  network {
	type               = "org"
	name               = (vcd_network_routed_v2.nsxt-backed.id == "always-not-equal" ? null : vcd_network_routed_v2.nsxt-backed.name)
	adapter_type       = "E1000E"
	ip_allocation_mode = "POOL"
	mac                = "00:00:00:11:22:33"
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

	secondary_ip_allocation_mode = "POOL"
  }

  network {
	type               = "org"
	name               = (vcd_network_routed_v2.nsxt-backed.id == "always-not-equal" ? null : vcd_network_routed_v2.nsxt-backed.name)
	adapter_type       = "E1000E"
	ip_allocation_mode = "POOL"
	mac                = "00:00:00:22:33:44"
  }

  prevent_update_power_off = true
}

# VM Copy from here
resource "vcd_vapp" "vm-copy-destination-template-vm" {
  org         = "{{.Org}}"
  vdc         = "{{.Vdc}}"
  name        = "{{.TestName}}-vm-copy-template-destination"
  description = "vApp destination for VM Copy"
  power_on    = false
}

resource "vcd_vapp_network" "template-copy" {
  org  = "{{.Org}}"
  vdc  = "{{.Vdc}}"
  name = "{{.TestName}}-template-vm-copy"

  vapp_name          = vcd_vapp.vm-copy-destination-template-vm.name
  gateway            = "192.168.3.1"
  netmask            = "255.255.255.0"

  static_ip_pool {
	start_address = "192.168.3.51"
	end_address   = "192.168.3.100"
  }

  depends_on = [vcd_vapp.template-vm]
}

resource "vcd_vapp_org_network" "template-vapp-copy" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  vapp_name        = vcd_vapp.vm-copy-destination-template-vm.name
  org_network_name = (vcd_network_routed_v2.nsxt-backed.id == "always-not-equal" ? null : vcd_network_routed_v2.nsxt-backed.name)
}

resource "vcd_vapp" "vm-copy-destination-empty-vm" {
  org         = "{{.Org}}"
  vdc         = "{{.Vdc}}"
  name        = "{{.TestName}}-vm-copy-empty-destination"
  description = "vApp destination for VM Copy"
  power_on    = false
}

resource "vcd_vapp_network" "empty-vm-copy" {
  org  = "{{.Org}}"
  vdc  = "{{.Vdc}}"
  name = "{{.TestName}}-empty-vm-copy"

  vapp_name          = vcd_vapp.vm-copy-destination-empty-vm.name
  gateway            = "192.168.2.1"
  netmask            = "255.255.255.0"

  static_ip_pool {
	start_address = "192.168.2.51"
	end_address   = "192.168.2.100"
  }

  depends_on = [vcd_vapp.empty-vm]
}

resource "vcd_vapp_org_network" "empty-vapp-copy" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  vapp_name        = vcd_vapp.vm-copy-destination-empty-vm.name
  org_network_name = (vcd_network_routed_v2.nsxt-backed.id == "always-not-equal" ? null : vcd_network_routed_v2.nsxt-backed.name)
}

resource "vcd_vapp_vm" "template-vm-copy" {
  org  = "{{.Org}}"
  vdc  = "{{.Vdc}}"
  
  copy_from_vm_id = vcd_vapp_vm.template-vm.id
  vapp_name       = vcd_vapp.vm-copy-destination-template-vm.name
  name            = "{{.TestName}}-template-vapp-vm-copy"
  description     = "{{.TestName}}-template-vapp-vm"
  power_on        = false

  network {
	type               = "org"
	name               = vcd_vapp_org_network.template-vapp-copy.org_network_name
	adapter_type       = "VMXNET3"
	ip_allocation_mode = "POOL"

	secondary_ip_allocation_mode = "POOL"
  }

  network {
	type               = "vapp"
	name               = vcd_vapp_network.template-copy.name
	adapter_type       = "E1000"
	ip_allocation_mode = "POOL"
	mac                = "00:00:00:AA:AC:CC"
  }

  prevent_update_power_off = true

  depends_on = [vcd_vapp_org_network.template-vapp-copy, vcd_vapp_network.template-copy]
}

resource "vcd_vapp_vm" "empty-vm-copy" {
  org  = "{{.Org}}"
  vdc  = "{{.Vdc}}"
  
  vapp_name     = vcd_vapp.vm-copy-destination-empty-vm.name
  name          = "{{.TestName}}-empty-vapp-vm-copy"
  description   = "{{.TestName}}-empty-vapp-vm"
  computer_name = "vapp-vm"
  power_on      = false

  cpus   = 1
  memory = 1024

  copy_from_vm_id = vcd_vapp_vm.empty-vm.id

  network {
	type               = "org"
	name               = vcd_vapp_org_network.empty-vapp-copy.org_network_name
	adapter_type       = "VMXNET3"
	ip_allocation_mode = "POOL"

	secondary_ip_allocation_mode = "POOL"
  }

  network {
	type               = "vapp"
	name               = vcd_vapp_network.empty-vm-copy.name
	adapter_type       = "E1000"
	ip_allocation_mode = "POOL"
	mac                = "00:00:00:AA:BB:FC"
  }

  prevent_update_power_off = true
  depends_on = [vcd_vapp_org_network.template-vapp-copy, vcd_vapp_network.template-copy]
}

resource "vcd_vm" "template-vm-copy" {
  org  = "{{.Org}}"
  vdc  = "{{.Vdc}}"

  copy_from_vm_id = vcd_vm.template-vm.id
  
  name        = "{{.TestName}}-template-standalone-vm-copy"
  description = "{{.TestName}}-template-standalone-vm"
  power_on    = false

  network {
	type               = "org"
	name               = (vcd_network_routed_v2.nsxt-backed.id == "always-not-equal" ? null : vcd_network_routed_v2.nsxt-backed.name)
	adapter_type       = "VMXNET3"
	ip_allocation_mode = "POOL"

	secondary_ip_allocation_mode = "POOL"
  }

  network {
	type               = "org"
	name               = (vcd_network_routed_v2.nsxt-backed.id == "always-not-equal" ? null : vcd_network_routed_v2.nsxt-backed.name)
	adapter_type       = "E1000E"
	ip_allocation_mode = "POOL"
	mac                = "00:00:00:11:FF:33"
  }

  prevent_update_power_off = true
}

resource "vcd_vm" "empty-vm-copy" {
  org  = "{{.Org}}"
  vdc  = "{{.Vdc}}"

  name          = "{{.TestName}}-empty-standalone-vm-copy"
  description   = "{{.TestName}}-standalone"
  computer_name = "standalone"
  power_on      = false

  cpus   = 1
  memory = 1024

  copy_from_vm_id = vcd_vm.empty-vm.id

  network {
	type               = "org"
	name               = (vcd_network_routed_v2.nsxt-backed.id == "always-not-equal" ? null : vcd_network_routed_v2.nsxt-backed.name)
	adapter_type       = "VMXNET3"
	ip_allocation_mode = "POOL"

	secondary_ip_allocation_mode = "POOL"
  }

  network {
	type               = "org"
	name               = (vcd_network_routed_v2.nsxt-backed.id == "always-not-equal" ? null : vcd_network_routed_v2.nsxt-backed.name)
	adapter_type       = "E1000E"
	ip_allocation_mode = "POOL"
	mac                = "00:00:00:22:33:44"
  }

  prevent_update_power_off = true
}
`

const testAccVcdVAppVm_4types_ipv6Step2DS = testAccVcdVAppVm_4types_ipv6DS + testAccVcdVAppVm_4types_ipv6Step1

const testAccVcdVAppVm_4types_ipv6Step3 = testAccVcdVAppVm_4types_ipv6Common + `
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
	
	secondary_ip_allocation_mode = "MANUAL"
	secondary_ip                 = "2002:0:0:1234:abcd:ffff:c0a6:138"
  }

  network {
	type               = "vapp"
	name               = (vcd_vapp_network.template.id == "always-not-equal" ? null : vcd_vapp_network.template.name)
	adapter_type       = "E1000"
	ip_allocation_mode = "POOL"
	mac                = "00:00:00:AA:BB:CC"
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

	secondary_ip_allocation_mode = "MANUAL"
	secondary_ip                 = "2002:0:0:1234:abcd:ffff:c0a6:139"
  }

  network {
	type               = "vapp"
	name               = (vcd_vapp_network.empty-vm.id == "always-not-equal" ? null : vcd_vapp_network.empty-vm.name)
	adapter_type       = "E1000"
	ip_allocation_mode = "POOL"
	mac                = "00:00:00:BB:AA:CC"
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

	secondary_ip_allocation_mode = "MANUAL"
	secondary_ip                 = "2002:0:0:1234:abcd:ffff:c0a6:13a"
  }

  network {
	type               = "org"
	name               = (vcd_network_routed_v2.nsxt-backed.id == "always-not-equal" ? null : vcd_network_routed_v2.nsxt-backed.name)
	adapter_type       = "E1000E"
	ip_allocation_mode = "POOL"
	mac                = "00:00:00:11:22:33"
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

	secondary_ip_allocation_mode = "MANUAL"
	secondary_ip                 = "2002:0:0:1234:abcd:ffff:c0a6:13b"
  }

  network {
	type               = "org"
	name               = (vcd_network_routed_v2.nsxt-backed.id == "always-not-equal" ? null : vcd_network_routed_v2.nsxt-backed.name)
	adapter_type       = "E1000E"
	ip_allocation_mode = "POOL"
	mac                = "00:00:00:22:33:44"
  }

  prevent_update_power_off = true
}

# VM Copy from here
resource "vcd_vapp" "vm-copy-destination-template-vm" {
  org         = "{{.Org}}"
  vdc         = "{{.Vdc}}"
  name        = "{{.TestName}}-vm-copy-template-destination"
  description = "vApp destination for VM Copy"
  power_on    = false
}

resource "vcd_vapp_network" "template-copy" {
  org  = "{{.Org}}"
  vdc  = "{{.Vdc}}"
  name = "{{.TestName}}-template-vm-copy"

  vapp_name          = vcd_vapp.vm-copy-destination-template-vm.name
  gateway            = "192.168.3.1"
  netmask            = "255.255.255.0"

  static_ip_pool {
	start_address = "192.168.3.51"
	end_address   = "192.168.3.100"
  }

  depends_on = [vcd_vapp.template-vm]
}

resource "vcd_vapp_org_network" "template-vapp-copy" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  vapp_name        = vcd_vapp.vm-copy-destination-template-vm.name
  org_network_name = (vcd_network_routed_v2.nsxt-backed.id == "always-not-equal" ? null : vcd_network_routed_v2.nsxt-backed.name)
}

resource "vcd_vapp" "vm-copy-destination-empty-vm" {
  org         = "{{.Org}}"
  vdc         = "{{.Vdc}}"
  name        = "{{.TestName}}-vm-copy-empty-destination"
  description = "vApp destination for VM Copy"
  power_on    = false
}

resource "vcd_vapp_network" "empty-vm-copy" {
  org  = "{{.Org}}"
  vdc  = "{{.Vdc}}"
  name = "{{.TestName}}-empty-vm-copy"

  vapp_name          = vcd_vapp.vm-copy-destination-empty-vm.name
  gateway            = "192.168.2.1"
  netmask            = "255.255.255.0"

  static_ip_pool {
	start_address = "192.168.2.51"
	end_address   = "192.168.2.100"
  }

  depends_on = [vcd_vapp.empty-vm]
}

resource "vcd_vapp_org_network" "empty-vapp-copy" {
  org = "{{.Org}}"
  vdc = "{{.Vdc}}"

  vapp_name        = vcd_vapp.vm-copy-destination-empty-vm.name
  org_network_name = (vcd_network_routed_v2.nsxt-backed.id == "always-not-equal" ? null : vcd_network_routed_v2.nsxt-backed.name)
}

resource "vcd_vapp_vm" "template-vm-copy" {
  org  = "{{.Org}}"
  vdc  = "{{.Vdc}}"
  
  copy_from_vm_id = vcd_vapp_vm.template-vm.id
  vapp_name       = vcd_vapp.vm-copy-destination-template-vm.name
  name            = "{{.TestName}}-template-vapp-vm-copy"
  description     = "{{.TestName}}-template-vapp-vm"
  power_on        = false

  network {
	type               = "org"
	name               = vcd_vapp_org_network.template-vapp-copy.org_network_name
	adapter_type       = "VMXNET3"
	ip_allocation_mode = "POOL"

	secondary_ip_allocation_mode = "MANUAL"
	secondary_ip                 = "2002:0:0:1234:abcd:ffff:c0a6:13c"
  }

  network {
	type               = "vapp"
	name               = vcd_vapp_network.template-copy.name
	adapter_type       = "E1000"
	ip_allocation_mode = "POOL"
	mac                = "00:00:00:AA:AC:CC"
  }

  prevent_update_power_off = true

  depends_on = [vcd_vapp_org_network.template-vapp-copy, vcd_vapp_network.template-copy]
}

resource "vcd_vapp_vm" "empty-vm-copy" {
  org  = "{{.Org}}"
  vdc  = "{{.Vdc}}"
  
  vapp_name     = vcd_vapp.vm-copy-destination-empty-vm.name
  name          = "{{.TestName}}-empty-vapp-vm-copy"
  description   = "{{.TestName}}-empty-vapp-vm"
  computer_name = "vapp-vm"
  power_on      = false

  cpus   = 1
  memory = 1024

  copy_from_vm_id = vcd_vapp_vm.empty-vm.id

  network {
	type               = "org"
	name               = vcd_vapp_org_network.empty-vapp-copy.org_network_name
	adapter_type       = "VMXNET3"
	ip_allocation_mode = "POOL"

	secondary_ip_allocation_mode = "MANUAL"
	secondary_ip                 = "2002:0:0:1234:abcd:ffff:c0a6:13d"
  }

  network {
	type               = "vapp"
	name               = vcd_vapp_network.empty-vm-copy.name
	adapter_type       = "E1000"
	ip_allocation_mode = "POOL"
	mac                = "00:00:00:AA:BB:FC"
  }

  prevent_update_power_off = true
  depends_on = [vcd_vapp_org_network.template-vapp-copy, vcd_vapp_network.template-copy]
}

resource "vcd_vm" "template-vm-copy" {
  org  = "{{.Org}}"
  vdc  = "{{.Vdc}}"

  copy_from_vm_id = vcd_vm.template-vm.id
  
  name        = "{{.TestName}}-template-standalone-vm-copy"
  description = "{{.TestName}}-template-standalone-vm"
  power_on    = false

  network {
	type               = "org"
	name               = (vcd_network_routed_v2.nsxt-backed.id == "always-not-equal" ? null : vcd_network_routed_v2.nsxt-backed.name)
	adapter_type       = "VMXNET3"
	ip_allocation_mode = "POOL"

	secondary_ip_allocation_mode = "MANUAL"
	secondary_ip                 = "2002:0:0:1234:abcd:ffff:c0a6:13e"
  }

  network {
	type               = "org"
	name               = (vcd_network_routed_v2.nsxt-backed.id == "always-not-equal" ? null : vcd_network_routed_v2.nsxt-backed.name)
	adapter_type       = "E1000E"
	ip_allocation_mode = "POOL"
	mac                = "00:00:00:11:FF:33"
  }

  prevent_update_power_off = true
}

resource "vcd_vm" "empty-vm-copy" {
  org  = "{{.Org}}"
  vdc  = "{{.Vdc}}"

  name          = "{{.TestName}}-empty-standalone-vm-copy"
  description   = "{{.TestName}}-standalone"
  computer_name = "standalone"
  power_on      = false

  cpus   = 1
  memory = 1024

  copy_from_vm_id = vcd_vm.empty-vm.id

  network {
	type               = "org"
	name               = (vcd_network_routed_v2.nsxt-backed.id == "always-not-equal" ? null : vcd_network_routed_v2.nsxt-backed.name)
	adapter_type       = "VMXNET3"
	ip_allocation_mode = "POOL"

	secondary_ip_allocation_mode = "MANUAL"
	secondary_ip                 = "2002:0:0:1234:abcd:ffff:c0a6:13f"
  }

  network {
	type               = "org"
	name               = (vcd_network_routed_v2.nsxt-backed.id == "always-not-equal" ? null : vcd_network_routed_v2.nsxt-backed.name)
	adapter_type       = "E1000E"
	ip_allocation_mode = "POOL"
	mac                = "00:00:00:22:33:44"
  }

  prevent_update_power_off = true
}
`

const testAccVcdVAppVm_4types_ipv6Step4DS = testAccVcdVAppVm_4types_ipv6DS + testAccVcdVAppVm_4types_ipv6Step3

const testAccVcdVAppVm_4types_ipv6DS = `
# skip-binary-test: Data Source test
data "vcd_vapp_vm" "template-vm" {
  vapp_name = vcd_vapp.template-vm.name
  name      = vcd_vapp_vm.template-vm.name
}

data "vcd_vapp_vm" "empty-vm" {
  vapp_name = vcd_vapp.empty-vm.name
  name      = vcd_vapp_vm.empty-vm.name
}

data "vcd_vm" "template-vm" {
  name = vcd_vm.template-vm.name
}

data "vcd_vm" "empty-vm" {
  name = vcd_vm.empty-vm.name
}


data "vcd_vapp_vm" "template-vm-copy" {
  vapp_name = vcd_vapp.vm-copy-destination-template-vm.name
  name      = vcd_vapp_vm.template-vm-copy.name
}

data "vcd_vapp_vm" "empty-vm-copy" {
  vapp_name = vcd_vapp.vm-copy-destination-empty-vm.name
  name      = vcd_vapp_vm.empty-vm-copy.name
}

data "vcd_vm" "template-vm-copy" {
  name = vcd_vm.template-vm-copy.name
}

data "vcd_vm" "empty-vm-copy" {
  name = vcd_vm.empty-vm-copy.name
}
`
