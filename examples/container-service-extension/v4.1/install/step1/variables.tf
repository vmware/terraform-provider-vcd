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

# ------------------------------------------------
# CSE Server Settings
# ------------------------------------------------

variable "vcdkeconfig_template_filepath" {
  type        = string
  description = "Path to the VCDKEConfig JSON template"
  default     = "../../entities/vcdkeconfig.json.template"
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
  default     = ""
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
  description = "VCDKEConfig: Node will be considered unhealthy and remediated if joining the cluster takes longer than this timeout (seconds)"
  default     = "900"
}

variable "node_not_ready_timeout" {
  type        = string
  description = "VCDKEConfig: A newly joined node will be considered unhealthy and remediated if it cannot host workloads for longer than this timeout (seconds)"
  default     = "300"
}

variable "node_unknown_timeout" {
  type        = string
  description = "VCDKEConfig: A healthy node will be considered unhealthy and remediated if it is unreachable for longer than this timeout (seconds)"
  default     = "300"
}

variable "max_unhealthy_node_percentage" {
  type        = number
  description = "VCDKEConfig: Remediation will be suspended when the number of unhealthy nodes exceeds this percentage. (100% means that unhealthy nodes will always be remediated, while 0% means that unhealthy nodes will never be remediated)"
  default     = 100
  validation {
    condition     = var.max_unhealthy_node_percentage >= 0 && var.max_unhealthy_node_percentage <= 100
    error_message = "The value must be a percentage, hence between 0 and 100"
  }
}

variable "container_registry_url" {
  type        = string
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
