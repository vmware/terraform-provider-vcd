package vcd

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func testAccCheckVcdVmDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*VCDClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "vcd_vm" {
			continue
		}

		_, err := conn.OrgVdc.GetVMByHREF(rs.Primary.ID)

		if err == nil {
			return fmt.Errorf("VPCs still exist")
		}

		return nil
	}

	return nil
}

func TestAccVcdVm_Basic(t *testing.T) {
	// var vapp govcd.VApp
	// var vm govcd.VM

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVcdVmDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: fmt.Sprintf(testAccCheckVcdVm_basic0),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm", "name", "test"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm", "memory", "512"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm", "cpus", "1"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm", "power_on", "true"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm", "storage_profile", "Silver"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm", "nested_hypervisor_enabled", "false"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm", "network.2.ip", "192.168.2.100"),

					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "name", "kdalby-dev-vapp-test-only-vm"),
					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "organization_network.#", "2"),
					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "organization_network.0", "FCI-IRT_ISN6_ORG-SRV"),
					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "organization_network.1", "FCI-IRT_ISN6_ORG-MGT"),
					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "vapp_network.#", "2"),
					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "vapp_network.0.name", "test"),
					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "vapp_network.1.name", "test2"),
				),
			},
			resource.TestStep{
				Config: fmt.Sprintf(testAccCheckVcdVm_basic1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm", "name", "test"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm", "memory", "1024"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm", "cpus", "2"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm", "power_on", "false"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm", "storage_profile", "Silver"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm", "nested_hypervisor_enabled", "true"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm", "network.2.ip", "192.168.2.100"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm", "network.#", "4"),
					// resource.Test(
					// 	"vcd_vm.test-vm", "network.2", "192.168.3.100"),

					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "name", "kdalby-dev-vapp-test-only-vm"),
					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "organization_network.#", "2"),
					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "organization_network.0", "FCI-IRT_ISN6_ORG-SRV"),
					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "organization_network.1", "FCI-IRT_ISN6_ORG-MGT"),
					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "vapp_network.#", "2"),
					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "vapp_network.0.name", "test"),
					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "vapp_network.1.name", "test2"),
				),
			},
			resource.TestStep{
				Config: fmt.Sprintf(testAccCheckVcdVm_basic2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm", "name", "test2"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm", "memory", "512"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm", "cpus", "4"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm", "power_on", "true"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm", "storage_profile", "Silver"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm", "nested_hypervisor_enabled", "false"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm", "network.0.name", "FCI-IRT_ISN6_ORG-SRV"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm", "network.1.name", "test"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm", "network.1.ip", "192.168.2.100"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm", "network.#", "2"),
					// resource.Test(
					// 	"vcd_vm.test-vm", "network.2", "192.168.3.100"),

					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "name", "kdalby-dev-vapp-test-only-vm"),
					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "organization_network.#", "2"),
					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "organization_network.0", "FCI-IRT_ISN6_ORG-SRV"),
					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "organization_network.1", "FCI-IRT_ISN6_ORG-MGT"),
					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "vapp_network.#", "2"),
					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "vapp_network.0.name", "test"),
					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "vapp_network.1.name", "test2"),
				),
			},
		},
	})

}

const testAccCheckVcdVm_basic0 = `
resource "vcd_vapp" "test-vapp" {
  name     = "kdalby-dev-vapp-test-only-vm"

  organization_network = [
    "FCI-IRT_ISN6_ORG-SRV",
    "FCI-IRT_ISN6_ORG-MGT",
  ]

  vapp_network {
     name = "test"
     description = ""
     gateway = "192.168.2.1"
     netmask = "255.255.255.0"
     dns1 = "8.8.8.8"
     dns2 = "8.8.4.4"
     start = "192.168.2.100"
     end = "192.168.2.199"
     nat = false
     parent = "FCI-IRT_ISN6_ORG-SRV"
     dhcp = false
  }
  vapp_network {
     name = "test2"
     description = ""
     gateway = "192.168.3.1"
     netmask = "255.255.255.0"
     dns1 = "8.8.8.8"
     dns2 = "8.8.4.4"
     start = "192.168.3.100"
     end = "192.168.3.199"
     nat = true
     parent = "FCI-IRT_ISN6_ORG-SRV"
     dhcp = true
     dhcp_start = "192.168.3.200"
     dhcp_end = "192.168.3.249"
  }

}
resource "vcd_vm" "test-vm"    {
      name          = "test"
      vapp_href     = "${vcd_vapp.test-vapp.id}"
      catalog_name  = "BETA_PUBLIC_IT_DEPARTMENT"
      template_name = "Ubuntu_Server_16.04"
      memory        = 512
      cpus          = 1
      power_on      = true
      storage_profile = "Silver"
      nested_hypervisor_enabled = false


      network = {
        name               = "FCI-IRT_ISN6_ORG-SRV"
        ip_allocation_mode = "POOL"
        is_primary         = true
        adapter_type       = "VMXNET3"
      }

      network = {
        name               = "FCI-IRT_ISN6_ORG-MGT"
        ip_allocation_mode = "POOL"
        adapter_type       = "VMXNET3"
      }

      network = {
        name               = "test"
        ip_allocation_mode = "POOL"
        adapter_type       = "VMXNET3"
      }
    }

`

const testAccCheckVcdVm_basic1 = `
resource "vcd_vapp" "test-vapp" {
  name     = "kdalby-dev-vapp-test-only-vm"

  organization_network = [
    "FCI-IRT_ISN6_ORG-SRV",
    "FCI-IRT_ISN6_ORG-MGT",
  ]

  vapp_network {
     name = "test"
     description = ""
     gateway = "192.168.2.1"
     netmask = "255.255.255.0"
     dns1 = "8.8.8.8"
     dns2 = "8.8.4.4"
     start = "192.168.2.100"
     end = "192.168.2.199"
     nat = false
     parent = "FCI-IRT_ISN6_ORG-SRV"
     dhcp = false
  }
  vapp_network {
     name = "test2"
     description = ""
     gateway = "192.168.3.1"
     netmask = "255.255.255.0"
     dns1 = "8.8.8.8"
     dns2 = "8.8.4.4"
     start = "192.168.3.100"
     end = "192.168.3.199"
     nat = true
     parent = "FCI-IRT_ISN6_ORG-SRV"
     dhcp = true
     dhcp_start = "192.168.3.200"
     dhcp_end = "192.168.3.249"
  }

}
resource "vcd_vm" "test-vm"    {
      name          = "test"
      vapp_href     = "${vcd_vapp.test-vapp.id}"
      catalog_name  = "BETA_PUBLIC_IT_DEPARTMENT"
      template_name = "Ubuntu_Server_16.04"
      memory        = 1024
      cpus          = 2
      power_on      = false
      storage_profile = "Silver"
      nested_hypervisor_enabled = true

      network = {
        name               = "FCI-IRT_ISN6_ORG-SRV"
        ip_allocation_mode = "POOL"
        is_primary         = true
        adapter_type       = "VMXNET3"
      }

      network = {
        name               = "FCI-IRT_ISN6_ORG-MGT"
        ip_allocation_mode = "POOL"
        adapter_type       = "VMXNET3"
      }

      network = {
        name               = "test"
        ip_allocation_mode = "POOL"
        adapter_type       = "VMXNET3"
      }

      network = {
        name               = "test2"
        ip_allocation_mode = "DHCP"
        adapter_type       = "VMXNET3"
      }
    }

`

const testAccCheckVcdVm_basic2 = `
resource "vcd_vapp" "test-vapp" {
  name     = "kdalby-dev-vapp-test-only-vm"

  organization_network = [
    "FCI-IRT_ISN6_ORG-SRV",
    "FCI-IRT_ISN6_ORG-MGT",
  ]

  vapp_network {
     name = "test"
     description = ""
     gateway = "192.168.2.1"
     netmask = "255.255.255.0"
     dns1 = "8.8.8.8"
     dns2 = "8.8.4.4"
     start = "192.168.2.100"
     end = "192.168.2.199"
     nat = false
     parent = "FCI-IRT_ISN6_ORG-SRV"
     dhcp = false
  }
  vapp_network {
     name = "test2"
     description = ""
     gateway = "192.168.3.1"
     netmask = "255.255.255.0"
     dns1 = "8.8.8.8"
     dns2 = "8.8.4.4"
     start = "192.168.3.100"
     end = "192.168.3.199"
     nat = true
     parent = "FCI-IRT_ISN6_ORG-SRV"
     dhcp = true
     dhcp_start = "192.168.3.200"
     dhcp_end = "192.168.3.249"
  }

}
resource "vcd_vm" "test-vm"    {
      name          = "test2"
      vapp_href     = "${vcd_vapp.test-vapp.id}"
      catalog_name  = "BETA_PUBLIC_IT_DEPARTMENT"
      template_name = "Ubuntu_Server_16.04"
      memory        = 512
      cpus          = 4
      power_on      = true
      storage_profile = "Silver"
      nested_hypervisor_enabled = false

      network = {
        name               = "FCI-IRT_ISN6_ORG-SRV"
        ip_allocation_mode = "POOL"
        is_primary         = true
        adapter_type       = "VMXNET3"
      }

      network = {
        name               = "test"
        ip_allocation_mode = "POOL"
        adapter_type       = "VMXNET3"
      }

    }
`

func TestAccVcdVm_MultiVM(t *testing.T) {
	// var vapp govcd.VApp
	// var vm govcd.VM

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVcdVmDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: fmt.Sprintf(testAccCheckVcdVm_multi_vm0),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm", "name", "test"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm", "memory", "512"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm", "cpus", "1"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm", "power_on", "true"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm", "storage_profile", "Silver"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm", "nested_hypervisor_enabled", "false"),

					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm2", "name", "test2"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm2", "memory", "1024"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm2", "cpus", "2"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm2", "power_on", "true"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm2", "storage_profile", "Silver-Protected"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm2", "nested_hypervisor_enabled", "false"),

					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "name", "kdalby-dev-vapp-test-only-vm-multi"),
					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "organization_network.#", "2"),
					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "organization_network.0", "FCI-IRT_ISN6_ORG-SRV"),
					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "organization_network.1", "FCI-IRT_ISN6_ORG-MGT"),
					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "vapp_network.#", "2"),
					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "vapp_network.0.name", "test"),
					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "vapp_network.1.name", "test2"),
				),
			},
			resource.TestStep{
				Config: fmt.Sprintf(testAccCheckVcdVm_multi_vm1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm", "name", "test"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm", "memory", "512"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm", "cpus", "1"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm", "power_on", "true"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm", "storage_profile", "Silver"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm", "nested_hypervisor_enabled", "false"),

					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm2", "name", "test2"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm2", "memory", "1024"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm2", "cpus", "2"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm2", "power_on", "false"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm2", "storage_profile", "Silver-Protected"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm2", "nested_hypervisor_enabled", "false"),

					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm3", "name", "test3"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm3", "memory", "512"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm3", "cpus", "1"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm3", "power_on", "true"),

					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm4", "name", "test4"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm4", "memory", "1024"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm4", "cpus", "2"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm4", "power_on", "false"),

					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm5", "name", "test5"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm5", "memory", "512"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm5", "cpus", "1"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm5", "power_on", "true"),

					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm6", "name", "test6"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm6", "memory", "1024"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm6", "cpus", "2"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm6", "power_on", "false"),

					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "name", "kdalby-dev-vapp-test-only-vm-multi"),
					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "organization_network.#", "2"),
					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "organization_network.0", "FCI-IRT_ISN6_ORG-SRV"),
					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "organization_network.1", "FCI-IRT_ISN6_ORG-MGT"),
					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "vapp_network.#", "2"),
					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "vapp_network.0.name", "test"),
					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "vapp_network.1.name", "test2"),
				),
			},
			resource.TestStep{
				Config: fmt.Sprintf(testAccCheckVcdVm_multi_vm2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm", "name", "taste"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm", "memory", "1024"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm", "cpus", "4"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm", "power_on", "false"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm", "storage_profile", "Silver-Protected"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm", "nested_hypervisor_enabled", "false"),

					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm2", "name", "taste2"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm2", "memory", "512"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm2", "cpus", "1"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm2", "power_on", "true"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm2", "storage_profile", "Silver"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm2", "nested_hypervisor_enabled", "false"),

					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm3", "name", "taste3"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm3", "memory", "1024"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm3", "cpus", "4"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm3", "storage_profile", "Silver-Protected"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm3", "power_on", "false"),

					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm4", "name", "taste4"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm4", "memory", "512"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm4", "cpus", "1"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm4", "storage_profile", "Silver"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm4", "power_on", "true"),

					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm5", "name", "taste5"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm5", "memory", "1024"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm5", "cpus", "4"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm5", "storage_profile", "Silver-Protected"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm5", "power_on", "false"),

					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm6", "name", "taste6"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm6", "memory", "512"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm6", "cpus", "1"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm6", "storage_profile", "Silver"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm6", "power_on", "true"),

					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "name", "kdalby-dev-vapp-test-only-vm-multi"),
					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "organization_network.#", "2"),
					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "organization_network.0", "FCI-IRT_ISN6_ORG-SRV"),
					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "organization_network.1", "FCI-IRT_ISN6_ORG-MGT"),
					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "vapp_network.#", "2"),
					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "vapp_network.0.name", "test"),
					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "vapp_network.1.name", "test2"),
				),
			},
		},
	})

}

const testAccCheckVcdVm_multi_vm_vapp = `
resource "vcd_vapp" "test-vapp" {
  name     = "kdalby-dev-vapp-test-only-vm-multi"

  organization_network = [
    "FCI-IRT_ISN6_ORG-SRV",
    "FCI-IRT_ISN6_ORG-MGT",
  ]

  vapp_network {
     name = "test"
     description = ""
     gateway = "192.168.2.1"
     netmask = "255.255.255.0"
     dns1 = "8.8.8.8"
     dns2 = "8.8.4.4"
     start = "192.168.2.100"
     end = "192.168.2.199"
     nat = false
     parent = "FCI-IRT_ISN6_ORG-SRV"
     dhcp = false
  }
  vapp_network {
     name = "test2"
     description = ""
     gateway = "192.168.3.1"
     netmask = "255.255.255.0"
     dns1 = "8.8.8.8"
     dns2 = "8.8.4.4"
     start = "192.168.3.100"
     end = "192.168.3.199"
     nat = true
     parent = "FCI-IRT_ISN6_ORG-SRV"
     dhcp = true
     dhcp_start = "192.168.3.200"
     dhcp_end = "192.168.3.249"
  }

}`

var testAccCheckVcdVm_multi_vm0 = fmt.Sprintf(`
%s
resource "vcd_vm" "test-vm"    {
  name          = "test"
  vapp_href     = "${vcd_vapp.test-vapp.id}"
  catalog_name  = "BETA_PUBLIC_IT_DEPARTMENT"
  template_name = "Ubuntu_Server_16.04"
  memory        = 512
  cpus          = 1
  power_on      = true
  storage_profile = "Silver"
  nested_hypervisor_enabled = false

}
resource "vcd_vm" "test-vm2"    {
  name          = "test2"
  vapp_href     = "${vcd_vapp.test-vapp.id}"
  catalog_name  = "BETA_PUBLIC_IT_DEPARTMENT"
  template_name = "Ubuntu_Server_16.04"
  memory        = 1024
  cpus          = 2
  power_on      = true
  storage_profile = "Silver-Protected"
  nested_hypervisor_enabled = false

}
`, testAccCheckVcdVm_multi_vm_vapp)

var testAccCheckVcdVm_multi_vm1 = fmt.Sprintf(`
%s
resource "vcd_vm" "test-vm"    {
  name          = "test"
  vapp_href     = "${vcd_vapp.test-vapp.id}"
  catalog_name  = "BETA_PUBLIC_IT_DEPARTMENT"
  template_name = "Ubuntu_Server_16.04"
  memory        = 512
  cpus          = 1
  power_on      = true
  storage_profile = "Silver"
  nested_hypervisor_enabled = false

}
resource "vcd_vm" "test-vm2"    {
  name          = "test2"
  vapp_href     = "${vcd_vapp.test-vapp.id}"
  catalog_name  = "BETA_PUBLIC_IT_DEPARTMENT"
  template_name = "Ubuntu_Server_16.04"
  memory        = 1024
  cpus          = 2
  power_on      = false
  storage_profile = "Silver-Protected"
  nested_hypervisor_enabled = false

}
resource "vcd_vm" "test-vm3"    {
  name          = "test3"
  vapp_href     = "${vcd_vapp.test-vapp.id}"
  catalog_name  = "BETA_PUBLIC_IT_DEPARTMENT"
  template_name = "Ubuntu_Server_16.04"
  memory        = 512
  cpus          = 1
  power_on      = true
  storage_profile = "Silver"
  nested_hypervisor_enabled = false

}
resource "vcd_vm" "test-vm4"    {
  name          = "test4"
  vapp_href     = "${vcd_vapp.test-vapp.id}"
  catalog_name  = "BETA_PUBLIC_IT_DEPARTMENT"
  template_name = "Ubuntu_Server_16.04"
  memory        = 1024
  cpus          = 2
  power_on      = false
  storage_profile = "Silver-Protected"
  nested_hypervisor_enabled = false

}
resource "vcd_vm" "test-vm5"    {
  name          = "test5"
  vapp_href     = "${vcd_vapp.test-vapp.id}"
  catalog_name  = "BETA_PUBLIC_IT_DEPARTMENT"
  template_name = "Ubuntu_Server_16.04"
  memory        = 512
  cpus          = 1
  power_on      = true
  storage_profile = "Silver"
  nested_hypervisor_enabled = false

}
resource "vcd_vm" "test-vm6"    {
  name          = "test6"
  vapp_href     = "${vcd_vapp.test-vapp.id}"
  catalog_name  = "BETA_PUBLIC_IT_DEPARTMENT"
  template_name = "Ubuntu_Server_16.04"
  memory        = 1024
  cpus          = 2
  power_on      = false
  storage_profile = "Silver-Protected"
  nested_hypervisor_enabled = false

}
`, testAccCheckVcdVm_multi_vm_vapp)

var testAccCheckVcdVm_multi_vm2 = fmt.Sprintf(`
%s
resource "vcd_vm" "test-vm"    {
  name          = "taste"
  vapp_href     = "${vcd_vapp.test-vapp.id}"
  catalog_name  = "BETA_PUBLIC_IT_DEPARTMENT"
  template_name = "Ubuntu_Server_16.04"
  memory        = 1024
  cpus          = 4
  power_on      = false
  storage_profile = "Silver-Protected"
  nested_hypervisor_enabled = false

}
resource "vcd_vm" "test-vm2"    {
  name          = "taste2"
  vapp_href     = "${vcd_vapp.test-vapp.id}"
  catalog_name  = "BETA_PUBLIC_IT_DEPARTMENT"
  template_name = "Ubuntu_Server_16.04"
  memory        = 512
  cpus          = 1
  power_on      = true
  storage_profile = "Silver"
  nested_hypervisor_enabled = false

}
resource "vcd_vm" "test-vm3"    {
  name          = "taste3"
  vapp_href     = "${vcd_vapp.test-vapp.id}"
  catalog_name  = "BETA_PUBLIC_IT_DEPARTMENT"
  template_name = "Ubuntu_Server_16.04"
  memory        = 1024
  cpus          = 4
  power_on      = false
  storage_profile = "Silver-Protected"
  nested_hypervisor_enabled = false

}
resource "vcd_vm" "test-vm4"    {
  name          = "taste4"
  vapp_href     = "${vcd_vapp.test-vapp.id}"
  catalog_name  = "BETA_PUBLIC_IT_DEPARTMENT"
  template_name = "Ubuntu_Server_16.04"
  memory        = 512
  cpus          = 1
  power_on      = true
  storage_profile = "Silver"
  nested_hypervisor_enabled = false

}
resource "vcd_vm" "test-vm5"    {
  name          = "taste5"
  vapp_href     = "${vcd_vapp.test-vapp.id}"
  catalog_name  = "BETA_PUBLIC_IT_DEPARTMENT"
  template_name = "Ubuntu_Server_16.04"
  memory        = 1024
  cpus          = 4
  power_on      = false
  storage_profile = "Silver-Protected"
  nested_hypervisor_enabled = false

}
resource "vcd_vm" "test-vm6"    {
  name          = "taste6"
  vapp_href     = "${vcd_vapp.test-vapp.id}"
  catalog_name  = "BETA_PUBLIC_IT_DEPARTMENT"
  template_name = "Ubuntu_Server_16.04"
  memory        = 512
  cpus          = 1
  power_on      = true
  storage_profile = "Silver"
  nested_hypervisor_enabled = false

}
`, testAccCheckVcdVm_multi_vm_vapp)

func TestAccVcdVm_MultiNic(t *testing.T) {
	// var vapp govcd.VApp
	// var vm govcd.VM

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVcdVmDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: fmt.Sprintf(testAccCheckVcdVm_multi_nic0),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm", "name", "test"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm", "memory", "512"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm", "cpus", "1"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm", "power_on", "false"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm", "storage_profile", "Silver"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm", "nested_hypervisor_enabled", "false"),

					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "name", "kdalby-dev-vapp-test-only-vm-multi-nic"),
					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "organization_network.#", "2"),
					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "organization_network.0", "FCI-IRT_ISN6_ORG-SRV"),
					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "organization_network.1", "FCI-IRT_ISN6_ORG-MGT"),
					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "vapp_network.#", "2"),
					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "vapp_network.0.name", "test"),
					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "vapp_network.1.name", "test2"),

					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm", "network.#", "1"),

					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm", "network.0.name", "FCI-IRT_ISN6_ORG-SRV"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm", "network.0.ip_allocation_mode", "POOL"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm", "network.0.is_primary", "true"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm", "network.0.adapter_type", "VMXNET3"),
				),
			},
			resource.TestStep{
				Config: fmt.Sprintf(testAccCheckVcdVm_multi_nic1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm", "name", "test"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm", "memory", "512"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm", "cpus", "1"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm", "power_on", "false"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm", "storage_profile", "Silver"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm", "nested_hypervisor_enabled", "false"),

					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "name", "kdalby-dev-vapp-test-only-vm-multi-nic"),
					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "organization_network.#", "2"),
					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "organization_network.0", "FCI-IRT_ISN6_ORG-SRV"),
					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "organization_network.1", "FCI-IRT_ISN6_ORG-MGT"),
					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "vapp_network.#", "2"),
					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "vapp_network.0.name", "test"),
					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "vapp_network.1.name", "test2"),

					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm", "network.#", "2"),

					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm", "network.0.name", "FCI-IRT_ISN6_ORG-SRV"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm", "network.0.ip_allocation_mode", "DHCP"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm", "network.0.ip", ""),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm", "network.0.is_primary", "true"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm", "network.0.adapter_type", "VMXNET3"),

					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm", "network.1.name", "test"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm", "network.1.ip_allocation_mode", "POOL"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm", "network.1.is_primary", "false"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm", "network.1.adapter_type", "E1000E"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm", "network.1.ip", "192.168.2.100"),
				),
			},
			resource.TestStep{
				Config: fmt.Sprintf(testAccCheckVcdVm_multi_nic2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm", "name", "test"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm", "memory", "512"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm", "cpus", "1"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm", "power_on", "false"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm", "storage_profile", "Silver"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm", "nested_hypervisor_enabled", "false"),

					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "name", "kdalby-dev-vapp-test-only-vm-multi-nic"),
					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "organization_network.#", "2"),
					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "organization_network.0", "FCI-IRT_ISN6_ORG-SRV"),
					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "organization_network.1", "FCI-IRT_ISN6_ORG-MGT"),
					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "vapp_network.#", "2"),
					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "vapp_network.0.name", "test"),
					resource.TestCheckResourceAttr(
						"vcd_vapp.test-vapp", "vapp_network.1.name", "test2"),

					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm", "network.#", "3"),

					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm", "network.0.name", "FCI-IRT_ISN6_ORG-SRV"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm", "network.0.ip_allocation_mode", "DHCP"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm", "network.0.ip", ""),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm", "network.0.is_primary", "false"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm", "network.0.adapter_type", "VMXNET3"),

					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm", "network.1.name", "FCI-IRT_ISN6_ORG-MGT"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm", "network.1.ip_allocation_mode", "POOL"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm", "network.1.is_primary", "true"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm", "network.1.adapter_type", "E1000E"),

					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm", "network.2.name", "test"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm", "network.2.ip_allocation_mode", "POOL"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm", "network.2.is_primary", "false"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm", "network.2.adapter_type", "E1000"),
					resource.TestCheckResourceAttr(
						"vcd_vm.test-vm", "network.2.ip", "192.168.2.100"),
				),
			},
		},
	})

}

const testAccCheckVcdVApp_multi_nic_vapp = `
resource "vcd_vapp" "test-vapp" {
  name     = "kdalby-dev-vapp-test-only-vm-multi-nic"

  organization_network = [
    "FCI-IRT_ISN6_ORG-SRV",
    "FCI-IRT_ISN6_ORG-MGT",
  ]

  vapp_network {
     name = "test"
     description = ""
     gateway = "192.168.2.1"
     netmask = "255.255.255.0"
     dns1 = "8.8.8.8"
     dns2 = "8.8.4.4"
     start = "192.168.2.100"
     end = "192.168.2.199"
     nat = false
     parent = "FCI-IRT_ISN6_ORG-SRV"
     dhcp = false
  }
  vapp_network {
     name = "test2"
     description = ""
     gateway = "192.168.3.1"
     netmask = "255.255.255.0"
     dns1 = "8.8.8.8"
     dns2 = "8.8.4.4"
     start = "192.168.3.100"
     end = "192.168.3.199"
     nat = true
     parent = "FCI-IRT_ISN6_ORG-SRV"
     dhcp = true
     dhcp_start = "192.168.3.200"
     dhcp_end = "192.168.3.249"
  }

}`

var testAccCheckVcdVm_multi_nic0 = fmt.Sprintf(`
%s
resource "vcd_vm" "test-vm"    {
  name          = "test"
  vapp_href     = "${vcd_vapp.test-vapp.id}"
  catalog_name  = "BETA_PUBLIC_IT_DEPARTMENT"
  template_name = "Ubuntu_Server_16.04"
  memory        = 512
  cpus          = 1
  power_on      = false
  storage_profile = "Silver"
  nested_hypervisor_enabled = false

  network = {
    name               = "FCI-IRT_ISN6_ORG-SRV"
    ip_allocation_mode = "POOL"
    is_primary         = true
    adapter_type       = "VMXNET3"
  }
}
`, testAccCheckVcdVApp_multi_nic_vapp)

var testAccCheckVcdVm_multi_nic1 = fmt.Sprintf(`
%s
resource "vcd_vm" "test-vm"    {
  name          = "test"
  vapp_href     = "${vcd_vapp.test-vapp.id}"
  catalog_name  = "BETA_PUBLIC_IT_DEPARTMENT"
  template_name = "Ubuntu_Server_16.04"
  memory        = 512
  cpus          = 1
  power_on      = false
  storage_profile = "Silver"
  nested_hypervisor_enabled = false

  network = {
    name               = "FCI-IRT_ISN6_ORG-SRV"
    ip_allocation_mode = "DHCP"
    is_primary         = true
    adapter_type       = "VMXNET3"
  }

  network = {
    name               = "test"
    ip_allocation_mode = "POOL"
    adapter_type       = "E1000E"
  }
}
`, testAccCheckVcdVApp_multi_nic_vapp)

var testAccCheckVcdVm_multi_nic2 = fmt.Sprintf(`
%s
resource "vcd_vm" "test-vm"    {
  name          = "test"
  vapp_href     = "${vcd_vapp.test-vapp.id}"
  catalog_name  = "BETA_PUBLIC_IT_DEPARTMENT"
  template_name = "Ubuntu_Server_16.04"
  memory        = 512
  cpus          = 1
  power_on      = false
  storage_profile = "Silver"
  nested_hypervisor_enabled = false

  network = {
    name               = "FCI-IRT_ISN6_ORG-SRV"
    ip_allocation_mode = "DHCP"
    adapter_type       = "VMXNET3"
  }
  network = {
    name               = "FCI-IRT_ISN6_ORG-MGT"
    ip_allocation_mode = "POOL"
    is_primary         = true
    adapter_type       = "E1000E"
  }
  network = {
    name               = "test"
    ip_allocation_mode = "POOL"
    adapter_type       = "E1000"
  }
}
`, testAccCheckVcdVApp_multi_nic_vapp)
