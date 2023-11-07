# ------------------------------------------------
# Provider config
# ------------------------------------------------

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
  description = "The VCD administrator user (Example: 'administrator')"
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

# ------------------------------------------------
# Infrastructure
# ------------------------------------------------

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

variable "nsxt_manager_name" {
  description = "NSX-T manager name, required to create the Provider Gateways"
  type        = string
}

variable "solutions_nsxt_tier0_router_name" {
  description = "Name of an existing NSX-T tier-0 router to create the Solutions Provider Gateway"
  type        = string
}

variable "solutions_provider_gateway_gateway_ip" {
  description = "Gateway IP for the Solutions Provider Gateway"
  type        = string
}

variable "solutions_provider_gateway_gateway_prefix_length" {
  description = "Prefix length for the Solutions Provider Gateway"
  type        = string
}

variable "solutions_provider_gateway_static_ip_ranges" {
  type        = list(list(string))
  description = "List of pairs of public IPs for the Solutions Provider Gateway"
}

variable "tenant_nsxt_tier0_router_name" {
  description = "Name of an existing NSX-T tier-0 router to create the Tenant Provider Gateway"
  type        = string
}

variable "tenant_provider_gateway_gateway_ip" {
  description = "Gateway IP for the Tenant Provider Gateway"
  type        = string
}

variable "tenant_provider_gateway_gateway_prefix_length" {
  description = "Prefix length for the Tenant Provider Gateway"
  type        = string
}

variable "tenant_provider_gateway_static_ip_ranges" {
  type        = list(list(string))
  description = "List of pairs of public IPs for the Tenant Provider Gateway"
}

variable "solutions_routed_network_gateway_ip" {
  description = "Gateway IP for the Solutions routed network"
  type        = string
}

variable "solutions_routed_network_prefix_length" {
  description = "Prefix length for the Solutions routed network"
  type        = string
}

variable "solutions_routed_network_ip_pool_start_address" {
  description = "Start address for the IP pool of the Solutions routed network"
  type        = string
}

variable "solutions_routed_network_ip_pool_end_address" {
  description = "End address for the IP pool of the Solutions routed network"
  type        = string
}

variable "solutions_snat_external_ip" {
  description = "Used to create a SNAT rule to allow connectivity. This specifies the external IP, which should be one of the Provider Gateway available IPs"
  type        = string
}

variable "solutions_snat_internal_network_cidr" {
  description = "Used to create a SNAT rule to allow connectivity. This specifies the internal subnet CIDR, which should correspond to the routed network IPs"
  type        = string
}

variable "solutions_routed_network_dns" {
  description = "Custom DNS server IP to use for the Solutions routed network"
  type        = string
  default     = ""
}

variable "solutions_routed_network_dns_suffix" {
  description = "Custom DNS suffix to use for the Solutions routed network"
  type        = string
  default     = ""
}

variable "tenant_routed_network_gateway_ip" {
  description = "Gateway IP for the Tenant routed network"
  type        = string
}

variable "tenant_routed_network_prefix_length" {
  description = "Prefix length for the Tenant routed network"
  type        = string
}

variable "tenant_routed_network_ip_pool_start_address" {
  description = "Start address for the IP pool of the Tenant routed network"
  type        = string
}

variable "tenant_routed_network_ip_pool_end_address" {
  description = "End address for the IP pool of the Tenant routed network"
  type        = string
}

variable "tenant_snat_external_ip" {
  description = "Used to create a SNAT rule to allow connectivity. This specifies the external IP, which should be one of the Provider Gateway available IPs"
  type        = string
}

variable "tenant_snat_internal_network_cidr" {
  description = "Used to create a SNAT rule to allow connectivity. This specifies the internal subnet CIDR, which should correspond to the routed network IPs"
  type        = string
}

variable "tenant_routed_network_dns" {
  description = "Custom DNS server IP to use for the Tenant routed network"
  type        = string
  default     = ""
}

variable "tenant_routed_network_dns_suffix" {
  description = "Custom DNS suffix to use for the Tenant routed network"
  type        = string
  default     = ""
}

variable "alb_controller_username" {
  description = "The user to create an ALB Controller with"
  type        = string
}

variable "alb_controller_password" {
  description = "The password for the user that will be used to create the ALB Controller"
  type        = string
}

variable "alb_controller_url" {
  description = "The URL to create the ALB Controller"
  type        = string
}

variable "alb_importable_cloud_name" {
  description = "Name of an available importable cloud to be able to create an ALB NSX-T Cloud"
  type        = string
}

# ------------------------------------------------
# Catalog and OVAs
# ------------------------------------------------

variable "tkgm_ova_folder" {
  description = "Absolute path to the TKGm OVA files, with no file name (Example: '/home/bob/Downloads/tkgm')"
  type        = string
}

variable "tkgm_ova_files" {
  description = "A set of TKGm OVA file names, with no path (Example: 'ubuntu-2004-kube-v1.25.7+vmware.2-tkg.1-8a74b9f12e488c54605b3537acb683bc.ova')"
  type        = set(string)
}

variable "cse_ova_folder" {
  description = "Absolute path to the CSE OVA file, with no file name (Example: '/home/bob/Downloads/cse')"
  type        = string
}

variable "cse_ova_file" {
  description = "CSE OVA file name, with no path (Example: 'VMware_Cloud_Director_Container_Service_Extension-4.0.1.62-21109756.ova')"
  type        = string
}

# ------------------------------------------------
# CSE Server initialization
# ------------------------------------------------

variable "cse_admin_username" {
  description = "The CSE administrator user that was created in step 1"
  type        = string
}

variable "cse_admin_password" {
  description = "The password to set for the CSE administrator user that was created in step 1"
  type        = string
  sensitive   = true
}

variable "cse_admin_api_token_file" {
  description = "The file where the API Token for the CSE Administrator will be stored"
  type        = string
  default     = "cse_admin_api_token.json"
}

# ------------------------------------------------
# Other configuration
# ------------------------------------------------

variable "k8s_container_clusters_ui_plugin_path" {
  type        = string
  description = "Path to the Kubernetes Container Clusters UI Plugin zip file"
  default     = ""
}
