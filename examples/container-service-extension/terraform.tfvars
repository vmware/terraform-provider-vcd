# Provider config
admin-user       = "******************************************"
admin-password   = "******************************************"
admin-org        = "System"
vcd-url          = "https://myvcd.company.com/api"
vcenter-name     = "vcenter1"
vcenter-username = "******************************************"
vcenter-password = "******************************************"

# CSE config
org-name = "cse_org"
storage-profile = "*"
vdc-name       = "cse_vdc"
vdc-allocation = "AllocationVApp"
vdc-provider   = "providerVdcNsxt1"
vdc-netpool    = "NSX-T Overlay 1"

tier0-manager        = "nsxManager1"
tier0-router         = "VCD T0"
tier0-gateway-ip     = "10.12.123.123"
tier0-gateway-prefix = "19"
tier0-gateway-ip-ranges = [
  ["10.12.234.226", "10.12.234.228"],
  ["10.12.234.230", "10.12.234.232"],
  ["10.12.234.235", "10.12.234.237"]
]

edge-gateway-ip     = "10.12.123.123"
edge-gateway-prefix = "19"
edge-gateway-ip-ranges = [
  ["10.12.234.226", "10.12.234.228"],
  ["10.12.234.230", "10.12.234.232"],
  ["10.12.234.235", "10.12.234.237"]
]

routed-gateway  = "192.168.7.1"
routed-prefix   = "24"
routed-ip-range = ["192.168.7.2", "192.168.7.100"]
routed-dns      = ["8.8.8.8", "8.8.8.4"]

avi-controller-name  = "aviController1"
avi-importable-cloud = "NSXT https://my-importable-cloud.company.com"

# OVA name should be something like "ubuntu-2022-kube-v1.22.0+vmware.1-tkg.1-1234567891234567890"
tkgm-ova-name   = "ubuntu-2004-kube-v1.21.2+vmware.1-tkg.1-7832907791984498322"
tkgm-ova-folder = "/Users/username/Downloads"
