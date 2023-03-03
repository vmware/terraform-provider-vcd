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
  description = "The CSE administrator user (Example: 'cse-admin')"
  type        = string
}

variable "cse_admin_api_token" {
  description = "The CSE administrator API token"
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

# ------------------------------------------------
# Networking
# ------------------------------------------------

variable "nsxt_manager_name" {
  description = "NSX-T manager name"
  type        = string
}

variable "nsxt_tier0_router_name" {
  description = "NSX-T tier-0 router name"
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

variable "cluster_provider_gateway_gateway_ip" {
  description = "Gateway IP for the Cluster Provider Gateway"
  type        = string
}

variable "cluster_provider_gateway_gateway_prefix_length" {
  description = "Prefix length for the Cluster Provider Gateway"
  type        = string
}

variable "cluster_provider_gateway_static_ip_ranges" {
  type        = list(list(string))
  description = "List of pairs of public IPs for the Solutions Provider Gateway"
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

variable "cluster_routed_network_gateway_ip" {
  description = "Gateway IP for the Cluster routed network"
  type        = string
}

variable "cluster_routed_network_prefix_length" {
  description = "Prefix length for the Cluster routed network"
  type        = string
}

variable "cluster_routed_network_ip_pool_start_address" {
  description = "Start address for the IP pool of the Cluster routed network"
  type        = string
}

variable "cluster_routed_network_ip_pool_end_address" {
  description = "End address for the IP pool of the Cluster routed network"
  type        = string
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
  description = "VCDKEConfig: CAPVCD version"
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