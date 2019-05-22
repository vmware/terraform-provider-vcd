# Remove the first '#' from the next two lines to enable options for terraform executable 
## apply-options -parallelism=1
## destroy-options -parallelism=1

# Independent disk
# v2.1.0
resource "vcd_independent_disk" "terraform_disk" {
  name         = "tf-disk"
  size         = "1024"        # MB
  bus_type     = "SCSI"
  bus_sub_type = "VirtualSCSI"
}

# Org VDC network
resource "vcd_network_routed" "terraform_network" {
  name         = "TerraformNetwork"
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

# vApp in the routed network
resource "vcd_vapp" "terraform_vapp" {
  name = "TerraformVApp"

  # Try adding an Org VDC network
  # network_name = "..."

  depends_on = ["vcd_network_routed.terraform_network"]
}

# vApp network
# v2.1.0
resource "vcd_vapp_network" "terraform_vapp_net" {
  name       = "TerraformVAppNet"
  vapp_name  = "${vcd_vapp.terraform_vapp.name}"
  gateway    = "192.168.2.1"
  netmask    = "255.255.255.0"
  dns1       = "192.168.2.1"
  dns2       = "192.168.2.2"
  dns_suffix = "{{.Org}}.org"

  #guest_vlan_allowed = true

  static_ip_pool {
    start_address = "192.168.2.51"
    end_address   = "192.168.2.100"
  }
  dhcp_pool {
    start_address = "192.168.2.2"
    end_address   = "192.168.2.50"
  }
}

# vApp's VM connected to a network with routed connection to the outside
resource "vcd_vapp_vm" "terraform_vm_first" {
  vapp_name  = "${vcd_vapp.terraform_vapp.name}"
  name          = "TerraformVMFirst"
  catalog_name  = "{{.Catalog}}"
  template_name = "{{.CatalogItem}}"
  memory        = 384
  cpus          = 4

  # Connect to routed network
  network {
    type                 = "org"
    name                 = "${vcd_network_routed.terraform_network.name}"
    ip                   = "192.168.0.7"
    ip_allocation_mode   = "POOL"
  }

  # v2.1.0
  disk {
    name        = "${vcd_independent_disk.terraform_disk.name}"
    bus_number  = 1
    unit_number = 0
  }

  accept_all_eulas = "true"
  depends_on       = ["vcd_network_routed.terraform_network", "vcd_vapp.terraform_vapp", "vcd_independent_disk.terraform_disk", "vcd_vapp_network.terraform_vapp_net"]
}

resource "vcd_catalog_media" "media_for_insertion" {
  catalog = "{{.Catalog}}"

  name                 = "media_for_insertion"
  description          = "media for insertion"
  media_path           = "{{.MediaPath}}"
  upload_piece_size    = {{.MediaUploadPieceSize}}
  show_upload_progress = "{{.MediaUploadProgress}}"
}

# Attach ISO to VM
resource "vcd_inserted_media" "terraform_inserted_iso" {
  catalog    = "{{.Catalog}}"
  name       = "${vcd_catalog_media.media_for_insertion.name}"

  vapp_name  = "${vcd_vapp.terraform_vapp.name}"
  vm_name   = "TerraformVMFirst"

  depends_on = ["vcd_vapp_vm.terraform_vm_first"]
}

resource "vcd_vapp_vm" "terraform_vm_second" {
  vapp_name  = "${vcd_vapp.terraform_vapp.name}"
  name          = "TerraformVMSecond"
  catalog_name  = "{{.Catalog}}"
  template_name = "{{.CatalogItem}}"
  memory        = 384

  # Connect to vApp network
  # v2.1.0

  network {
    type                 = "vapp"
    name                 = "${vcd_vapp_network.terraform_vapp_net.name}"
    ip                   = "192.168.0.9"
    ip_allocation_mode   = "POOL"
  }

  # Reconfigure CPU cores
  # v2.1.0
  cpus = "4"
  cpu_cores        = "2"
  accept_all_eulas = "true"
  depends_on       = ["vcd_vapp.terraform_vapp", "vcd_vapp_network.terraform_vapp_net", "vcd_vapp_vm.terraform_vm_first", "vcd_vapp_network.terraform_vapp_net"]
}

# SNAT rule to let the VMs' traffic out
resource "vcd_snat" "outbound" {
  edge_gateway = "{{.EdgeGateway}}"

  external_ip = "{{.ExternalIp}}/32"
  internal_ip = "192.168.0.0/24"
}

# DNAT rule to SSH the VM from the outside
resource "vcd_dnat" "sshVM" {
  edge_gateway = "{{.EdgeGateway}}"

  external_ip     = "{{.ExternalIp}}/32"
  port            = 2227
  internal_ip     = "192.168.0.7/32"
  translated_port = 22
}

# Firewall rule to allow SSH on NAT port
resource "vcd_firewall_rules" "terraform_fw_rule" {
  edge_gateway   = "{{.EdgeGateway}}"
  default_action = "allow"

  rule {
    description      = "allows_ssh"
    policy           = "allow"
    protocol         = "TCP"
    destination_port = "2227"
    destination_ip   = "10.150.211.101"
    source_port      = "any"
    source_ip        = "any"
  }
}
