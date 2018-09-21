
provider "vcd" {
  user                 = "root"
  password             = "root"
  org                  = "System"
  url                  = "https://api.vcd.com/api"
  max_retry_timeout    = "60"
  allow_unverified_ssl = "true"
}

resource "vcd_dnat" "web2" {

  edge_gateway = "test_edge_3"
  org = "au"
  vdc = "au-vdc"
  external_ip  = "10.x.x.x"
  port         = 80
  internal_ip  = "10.x.x.x"
  translated_port = 8080
}

resource "vcd_snat" "outbound" {
  edge_gateway = "test_edge_3"
  org = "au"
  vdc = "au-vdc"
  external_ip  = "10.x.x.x"
  internal_ip  = "10.x.x.x"
}
resource "vcd_network" "net" {
  name         = "my-nt"
  org = "au"
  vdc = "au-vdc"
  edge_gateway = "test_edge_3"
  gateway      = "10.10.1.1"

  dhcp_pool {
    start_address = "10.10.1.2"
    end_address   = "10.10.1.100"
  }

  static_ip_pool {
    start_address = "10.10.1.152"
    end_address   = "10.10.1.254"
  }
}

resource "vcd_vapp" "test-tf-2" {
  name          = "test-tf-2"
  org           = "au"
  vdc           = "au-vdc"
 
}

resource "vcd_org" "test5" {
  name = "test5"
  full_name = "test5"
  is_enabled = "true"
  stored_vm_quota = 10
  deployed_vm_quota = 10
  force = "true"
  recursive = "true"
}

resource "vcd_org" "test4" {
  name = "test4"
  full_name = "test4"
  is_enabled = "true"
  stored_vm_quota = 10
  force = "true"
  recursive = "true"
}
