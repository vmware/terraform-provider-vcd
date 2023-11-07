# ------------------------------------------------------------------------------------------------------------
# WARNING: This CSE installation method is deprecated in favor of CSE v4.x. Please have a look at
#          https://registry.terraform.io/providers/vmware/vcd/latest/docs/guides/container_service_extension_4_x_install
# ------------------------------------------------------------------------------------------------------------

# These variables are for configuring the CSE installation

variable "org-name" {
  description = "The Organization where CSE will be installed"
  type        = string
}

variable "vdc-name" {
  description = "The VDC name where CSE will be installed"
  type        = string
}

variable "vdc-allocation" {
  description = "The VDC allocation model"
  default     = "AllocationVApp"
  type        = string
}

variable "vdc-provider" {
  description = "The provider VDC"
  type        = string
}

variable "vdc-netpool" {
  description = "The networking pool name for the VDC"
  type        = string
}

variable "tier0-manager" {
  description = "The NSX manager for creating the Tier 0 Gateway"
  type        = string
}

variable "tier0-router" {
  description = "The router for creating the Tier 0 Gateway"
  type        = string
}

variable "tier0-gateway-ip" {
  description = "The Tier 0 Gateway IP"
  type        = string
}

variable "tier0-gateway-prefix" {
  description = "The Tier 0 Gateway prefix length"
  type        = string
}

variable "tier0-gateway-ip-ranges" {
  description = "The Tier 0 Gateway available IPs for static IP pool. It's a list of IP pairs (ranges)"
  type        = list(list(string))
}

variable "edge-gateway-ip" {
  description = "The Edge Gateway IP"
  type        = string
}

variable "edge-gateway-prefix" {
  description = "The Edge Gateway prefix length"
  type        = string
}

variable "edge-gateway-ip-ranges" {
  description = "The Edge Gateway available IPs for allocation.  It's a list of IP pairs (ranges)"
  type        = list(list(string))
}

variable "snat-external-ip" {
  description = "External IP to map to an internal one through a SNAT mapping"
  type        = string
}

variable "routed-gateway" {
  description = "The routed network gateway IP"
  type        = string
}

variable "routed-prefix" {
  description = "The routed network prefix for IPs"
  type        = string
}

variable "routed-ip-range" {
  description = "The routed network available IP range"
  type        = list(string)
}

variable "routed-dns" {
  description = "The routed network DNS servers (two)"
  type        = list(string)
}

variable "storage-profile" {
  description = "Storage profile to use"
  default     = "*"
  type        = string
}

variable "tkgm-ova-name" {
  description = "TKGm OVA name"
  type        = string
}

variable "tkgm-ova-folder" {
  description = "Folder where TKGm OVA is located"
  type        = string
}

variable "avi-controller-name" {
  description = "AVI controller name"
  type        = string
}

variable "avi-importable-cloud" {
  description = "AVI importable cloud"
  type        = string
}

variable "avi-virtual-service-ip" {
  description = "IP for the virtual service on AVI load balancer"
  type        = string
}