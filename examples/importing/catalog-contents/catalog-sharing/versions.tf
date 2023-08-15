terraform {
  required_providers {
    vcd = {
      source  = "vmware/vcd"
      version = ">= 3.11.0" # we need the version containing vcd_resource_list enhancements for importing
    }
  }
  required_version = ">= 1.5.0" # terraform versions before 1.5 do not support import blocks
}
