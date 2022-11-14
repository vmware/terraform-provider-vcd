terraform {
  required_providers {
    vcd = {
      source  = "vmware/vcd"
      version = ">= 3.8.0"
    }
  }
  required_version = ">= 1.3.0"
}
