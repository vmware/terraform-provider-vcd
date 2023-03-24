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

variable "cse_admin_username" {
  description = "The CSE administrator user"
  type        = string
}

variable "cse_admin_password" {
  description = "The CSE administrator password"
  type        = string
  sensitive   = true
}

variable "administrator_org" {
  description = "The VCD administrator organization (Example: 'System')"
  type        = string
  default     = "System"
}

# ------------------------------------------------
# Kubernetes cluster configuration
# ------------------------------------------------
variable "k8s_cluster_name" {
  description = "The name of the Kubernetes cluster. Name must contain only lowercase alphanumeric characters or '-' start with an alphabetic character, end with an alphanumeric, and contain at most 31 characters (Example: 'MyCluster')"
  type        = string
}

variable "cluster_organization" {
  description = "The Organization that will host the Kubernetes cluster"
}

variable "cluster_vdc" {
  description = "The VDC that will host the Kubernetes cluster"
}

variable "cluster_routed_network" {
  description = "The routed network used for the Kubernetes cluster"
}

variable "cluster_sizing_policy" {
  description = "The VM sizing policy used for the Kubernetes cluster"
}

variable "cluster_author_user" {
  description = "Username of the Kubernetes cluster author"
}

variable "cluster_author_api_token" {
  description = "API token of the Kubernetes cluster author"
  sensitive   = true
}

variable "tkgm_catalog" {
  description = "The TKGm Catalog used to pick the OVAs to create the Kubernetes cluster"
}

variable "tkgm_ova" {
  description = "The TKGm OVA to create the Kubernetes cluster"
}