# Note: all resources are created inside a NSX-T VDC

data "vcd_nsxt_edgegateway" "existing" {
  org  = "datacloud"
  vdc  = "nsxt-vdc-datacloud"
  name = "nsxt-gw-datacloud"
}

resource "vcd_network_routed_v2" "net_r_v2" {
  name            = "net_r_v2"
  org             = "datacloud"
  vdc             = "nsxt-vdc-datacloud"
  edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id
  gateway         = "10.10.102.1"
  prefix_length   = 24

  static_ip_pool {
    start_address = "10.10.102.2"
    end_address   = "10.10.102.200"
  }
}

resource "vcd_network_isolated_v2" "net_i_v2" {
  name            = "net_i_v2"
  org             = "datacloud"
  vdc             = "nsxt-vdc-datacloud"
  
  gateway         = "110.10.102.1"
  prefix_length   = 26

  static_ip_pool {
    start_address = "110.10.102.2"
    end_address   = "110.10.102.20"
  }
}

resource "vcd_nsxt_network_dhcp" "net_r_dhcp" {
  org             = "datacloud"
  vdc             = "nsxt-vdc-datacloud"
  
  org_network_id  = vcd_network_routed_v2.net_r_v2.id

  pool {
    start_address = "10.10.102.210"
    end_address   = "10.10.102.220"
  }

  pool {
    start_address = "10.10.102.230"
    end_address   = "10.10.102.240"
  }
}

resource "vcd_nsxt_network_imported" "imported-test" {
  name            = "imported-test"
  org             = "datacloud"
  vdc             = "nsxt-vdc-datacloud"
  gateway         = "12.12.2.1"
  prefix_length   = 24

  nsxt_logical_switch_name = "segment-datacloud"

  static_ip_pool {
    start_address = "12.12.2.10"
    end_address   = "12.12.2.15"
  }
}

resource "vcd_vm" "standaloneVm" {
  org           = "datacloud"
  vdc           = "nsxt-vdc-datacloud"
  name          = "standaloneVm"
  computer_name = "standaloneVm-unique"
  catalog_name  = "cat-datacloud"
  template_name = "photon-hw11"
  description   = "test standalone VM"
  memory        = 1024
  cpus          = 2
  cpu_cores     = 1

  network_dhcp_wait_seconds = 10

  network {
    type               = "org"
    name               = vcd_network_routed_v2.net_r_v2.name
    ip_allocation_mode = "MANUAL"
    ip                 = "10.10.102.161"
  }

  network {
    type               = "org"
    name               = vcd_network_routed_v2.net_r_v2.name
    ip_allocation_mode = "DHCP"
  }

  network {
    type               = "org"
    name               = vcd_network_routed_v2.net_r_v2.name
    ip_allocation_mode = "POOL"
  }

  network {
    type               = "org"
    name               = vcd_network_isolated_v2.net_i_v2.name
    ip_allocation_mode = "POOL"
  }

  network {
    type               = "org"
    name               = vcd_nsxt_network_imported.imported-test.name
    ip_allocation_mode = "POOL"
  }
}

