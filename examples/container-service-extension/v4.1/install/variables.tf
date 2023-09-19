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
# CSE Server Pre-requisites
# ------------------------------------------------

variable "cse_admin_username" {
  description = "The CSE administrator user that will be created (Example: 'cse-admin')"
  type        = string
}

variable "cse_admin_password" {
  description = "The password to set for the CSE administrator to be created"
  type        = string
  sensitive   = true
}

variable "cse_admin_api_token_file" {
  description = "The file where the API token for the CSE administrator is stored"
  type        = string
  default     = "cse_admin_api_token.json"
}

# ------------------------------------------------
# CSE Server Settings
# ------------------------------------------------

variable "vcdkeconfig_template_filepath" {
  type        = string
  description = "Path to the VCDKEConfig JSON template"
  default     = "../entities/vcdkeconfig.json.template"
}

variable "capvcd_version" {
  type        = string
  description = "Version of CAPVCD"
  default     = "1.1.0"
}

variable "cpi_version" {
  type        = string
  description = "VCDKEConfig: Cloud Provider Interface version"
  default     = "1.4.0"
}

variable "csi_version" {
  type        = string
  description = "VCDKEConfig: Container Storage Interface version"
  default     = "1.4.0"
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

variable "node_startup_timeout" {
  type        = string
  description = "VCDKEConfig: Node will be considered unhealthy and remediated if joining the cluster takes longer than this timeout"
  default     = "900"
}

variable "node_not_ready_timeout" {
  type        = string
  description = "VCDKEConfig: A newly joined node will be considered unhealthy and remediated if it cannot host workloads for longer than this timeout"
  default     = "300"
}

variable "node_unknown_timeout" {
  type        = string
  description = "VCDKEConfig: A healthy node will be considered unhealthy and remediated if it is unreachable for longer than this timeout"
  default     = "300"
}

variable "max_unhealthy_node_percentage" {
  type        = number
  description = "VCDKEConfig: Remediation will be suspended when the number of unhealthy nodes exceeds this percentage. (100% means that unhealthy nodes will always be remediated, while 0% means that unhealthy nodes will never be remediated)"
  default     = 100
}

variable "container_registry_url" {
  type        = number
  description = "VCDKEConfig: URL from where TKG clusters will fetch container images"
  default     = "projects.registry.vmware.com"
}

variable "bootstrap_vm_certificates" {
  type        = list(string)
  description = "VCDKEConfig: Certificate(s) to allow the ephemeral VM (created during cluster creation) to authenticate with. For example, when pulling images from a container registry. (Copy and paste .cert file contents)"
  default     = []
}

variable "k8s_cluster_certificates" {
  type        = list(string)
  description = "VCDKEConfig: Certificate(s) to allow clusters to authenticate with. For example, when pulling images from a container registry. (Copy and paste .cert file contents)"
  default     = []
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
# Other configuration
# ------------------------------------------------
variable "k8s_container_clusters_ui_plugin_path" {
  type        = string
  description = "Path to the Kubernetes Container Clusters UI Plugin zip file"
  default     = ""
}
