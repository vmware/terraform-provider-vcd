# This file is the same as ./ResourceExamples/vcd_test.tf
# converted into a template to make it easy to run.

resource "vcd_network_routed" "net" {
  name         = "my-nt"
  org          = "{{.Org}}"
  vdc          = "{{.Vdc}}"
  edge_gateway = "{{.EdgeGateway}}"
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
  name = "test-tf-2"
  org  = "{{.Org}}"
  vdc  = "{{.Vdc}}"
}

resource "vcd_org" "test5" {
  name              = "test5"
  full_name         = "test5"
  is_enabled        = "true"
  stored_vm_quota   = 10
  deployed_vm_quota = 10
  delete_force      = "true"
  delete_recursive  = "true"
}

resource "vcd_org" "test4" {
  name             = "test4"
  full_name        = "test4"
  is_enabled       = "true"
  stored_vm_quota  = 10
  delete_force     = "true"
  delete_recursive = "true"
}
