variable "vcd_url" {
  description = "The VCD URL (Example: 'https://vcd.my-company.com')"
  type        = string
}

variable "insecure_login" {
  description = "Allow unverified SSL connections when operating with VCD"
  type        = bool
  default     = false
}

variable "administrator_user" {
  description = "The VCD administrator password (Example: 'administrator')"
  default     = "administrator"
  type        = string
}

variable "administrator_password" {
  description = "The VCD administrator password"
  type        = string
  sensitive   = true
}

variable "administrator_org" {
  description = "The VCD administrator organization (Example: 'System')"
  type        = string
  default     = "System"
}

variable "provider_vdc_name" {
  description = "The Provider VDC that will be used to create the required VDCs"
  type        = string
}

variable "nsxt_edge_cluster_name" {
  description = "The NSX-T Edge Cluster name, that relates to the specified Provider VDC"
  type        = string
}

variable "network_pool_name" {
  description = "The network pool to be used on VDC creation"
  type        = string
}

variable "tkgm_ova_folder" {
  description = "Path to the TKGm OVA file, with no file name (Example: '/home/bob/Downloads/tkgm')"
  type        = string
}

variable "tkgm_ova_file" {
  description = "TKGm file name, with no path (Example: 'ubuntu-2004-kube-v1.22.9+vmware.1-tkg.1-2182cbabee08edf480ee9bc5866d6933.ova')"
  type        = string
}

variable "cse_ova_folder" {
  description = "Path to the CSE OVA file, with no file name (Example: '/home/bob/Downloads/cse')"
  type        = string
}

variable "cse_ova_file" {
  description = "CSE file name, with no path (Example: 'VMware_Cloud_Director_Container_Service_Extension-4.0.1.62-21109756.ova')"
  type        = string
}