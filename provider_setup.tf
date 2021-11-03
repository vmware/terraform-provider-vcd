
terraform {
  required_providers {
    vcd = {
      source  = "vmware/vcd"
      version = "~> 3.4"
    }
    nsxt = {
      source = "vmware/nsxt"
    }
  }
  required_version = ">= 0.13"
}
