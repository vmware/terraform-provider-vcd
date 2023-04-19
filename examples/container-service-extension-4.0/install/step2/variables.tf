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
# CSE administrator user details
# ------------------------------------------------

variable "cse_admin_user" {
  description = "The CSE administrator user created in previous step (Example: 'cse-admin')"
  type        = string
}

variable "cse_admin_api_token" {
  description = "The CSE administrator API token that should have been created before running this installation step"
  type        = string
  sensitive   = true
}

# ------------------------------------------------
# VDC setup
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

# ------------------------------------------------
# Catalog and OVAs
# ------------------------------------------------

variable "tkgm_ova_folder" {
  description = "Absolute path to the TKGm OVA file, with no file name (Example: '/home/bob/Downloads/tkgm')"
  type        = string
}

variable "tkgm_ova_file" {
  description = "TKGm OVA file name, with no path (Example: 'ubuntu-2004-kube-v1.22.9+vmware.1-tkg.1-2182cbabee08edf480ee9bc5866d6933.ova')"
  type        = string
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
# Networking
# ------------------------------------------------

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

# ------------------------------------------------
# ALB
# ------------------------------------------------
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
# CSE Server
# ------------------------------------------------
variable "vcdkeconfig_template_filepath" {
  type        = string
  description = "Path to the VCDKEConfig JSON template"
  default     = "../../entities/vcdkeconfig-template.json"
}

variable "capvcd_version" {
  type        = string
  description = "Version of CAPVCD"
  default     = "1.0.0"
}

variable "capvcd_rde_version" {
  type        = string
  description = "Version of the CAPVCD Runtime Defined Entity Type"
  default     = "1.1.0"
}

variable "cpi_version" {
  type        = string
  description = "VCDKEConfig: Cloud Provider Interface version"
  default     = "1.2.0"
}

variable "csi_version" {
  type        = string
  description = "VCDKEConfig: Container Storage Interface version"
  default     = "1.3.0"
}

variable "github_personal_access_token" {
  type        = string
  description = "VCDKEConfig: Prevents potential github rate limiting errors during cluster creation and deletion"
  sensitive   = true
}

variable "no_proxy" {
  type        = string
  description = "VCDKEConfig: List of comma-separated domains without spaces"
  default     = "localhost,127.0.0.1,cluster.local,.svc"
}

variable "http_proxy" {
  type        = string
  description = "VCDKEConfig: Address of your HTTP proxy server"
  default     = ""
}

variable "https_proxy" {
  type        = string
  description = "VCDKEConfig: Address of your HTTPS proxy server"
  default     = ""
}

variable "syslog_host" {
  type        = string
  description = "VCDKEConfig: Domain for system logs"
  default     = ""
}

variable "syslog_port" {
  type        = string
  description = "VCDKEConfig: Port for system logs"
  default     = ""
}
