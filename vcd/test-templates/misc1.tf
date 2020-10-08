# Remove the first '#' from the next two lines to enable options for terraform executable 
## apply-options -parallelism=1
## destroy-options -parallelism=1

# Org VDC network
resource "vcd_network_routed" "tf_net" {
  name         = "TfNet"
  edge_gateway = "{{.EdgeGateway}}"

  gateway = "192.168.0.1"

  static_ip_pool {
    start_address = "192.168.0.2"
    end_address   = "192.168.0.100"
  }

  dhcp_pool {
    start_address = "192.168.0.101"
    end_address   = "192.168.0.200"
  }
}

resource "vcd_vapp" "tf_vapp" {
  name = "TfVApp"

  depends_on = ["vcd_network_routed.tf_net"]
}

# v2.1.0
resource "vcd_vapp_network" "tf_vapp_net" {
  name       = "TfVAppNet"
  vapp_name  = "${vcd_vapp.tf_vapp.name}"
  gateway    = "192.168.2.1"
  netmask    = "255.255.255.0"
  dns1       = "192.168.2.1"
  dns2       = "192.168.2.2"
  dns_suffix = "{{.Org}}.org"

  static_ip_pool {
    start_address = "192.168.2.51"
    end_address   = "192.168.2.100"
  }

  dhcp_pool {
    start_address = "192.168.2.2"
    end_address   = "192.168.2.50"
  }

  depends_on = ["vcd_vapp.tf_vapp"]
}

resource "vcd_vapp_vm" "tf_vm_11" {
  vapp_name  = "${vcd_vapp.tf_vapp.name}"
  name          = "TfVM11"
  catalog_name  = "{{.Catalog}}"
  template_name = "{{.CatalogItem}}"
  memory        = 384
  cpus          = 1

  # v2.2.0+
  network {
    type               = "vapp"
    name               = "${vcd_vapp_network.tf_vapp_net.name}"
    ip_allocation_mode = "POOL"
    is_primary         = false
  }

  # v2.2.0+
  network {
    type               = "org"
    name               = "TfNet"
    ip                 = "192.168.0.11"
    ip_allocation_mode = "MANUAL"
    is_primary         = true
  }

  # v2.2.0+
  network {
    type               = "none"
    ip_allocation_mode = "NONE"
  }

  # v2.2.0+
  expose_hardware_virtualization = true

  # v2.2.0+
  metadata = {
    role    = "test"
    env     = "staging"
    version = "v2.2.0"
  }

  accept_all_eulas = "true"
  depends_on       = ["vcd_network_routed.tf_net", "vcd_vapp.tf_vapp"]
}
