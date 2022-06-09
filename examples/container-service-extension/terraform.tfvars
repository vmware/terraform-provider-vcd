# Provider config
admin-user       = "administrator"
admin-password   = "ca$hc0w"
admin-org        = "System"
vcd-url          = "https://atl1-vcd-static-99-234.eng.vmware.com/api"
vcenter-name     = "vc1"
vcenter-username = "administrator@vsphere.local"
vcenter-password = "Welcome@123"

# CSE config
org-name = "cse_org"
storage-profile = "*"
vdc-name       = "cse_vdc"
vdc-allocation = "AllocationVApp"
vdc-provider   = "nsxTPvdc1"
vdc-netpool    = "NSX-T Overlay 1"

tier0-manager        = "nsxManager1"
tier0-router         = "VCD T0 edgeCluster1"
tier0-gateway-ip     = "10.89.127.253"
tier0-gateway-prefix = "19"
tier0-gateway-ip-ranges = [
  ["10.89.99.226", "10.89.99.226"],
  ["10.89.99.228", "10.89.99.228"],
  ["10.89.99.230", "10.89.99.233"]
]

edge-gateway-ip     = "10.89.127.253"
edge-gateway-prefix = "19"
edge-gateway-ip-ranges = [
  ["10.89.99.226", "10.89.99.226"],
  ["10.89.99.228", "10.89.99.228"],
  ["10.89.99.230", "10.89.99.233"]
]

routed-gateway  = "192.168.7.1"
routed-prefix   = "24"
routed-ip-range = ["192.168.7.2", "192.168.7.100"]
routed-dns      = ["8.8.8.8", "8.8.8.4"]

avi-controller-name  = "aviController1"
avi-importable-cloud = "NSXT atl1-vcd-static-99-236.eng.vmware.com"

# OVA name should be something like "ubuntu-2022-kube-v1.22.0+vmware.1-tkg.1-1234567891234567890"
tkgm-ova-name   = "ubuntu-2004-kube-v1.21.2+vmware.1-tkg.1-7832907791984498322"
tkgm-ova-folder = "/Users/abarreiro/Downloads"
