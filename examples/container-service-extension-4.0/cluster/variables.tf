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

variable "cluster_author_user" {
  description = "Username of the Kubernetes cluster author"
}

variable "cluster_author_password" {
  description = "Password of the Kubernetes cluster author"
  type        = string
  sensitive   = true
}

variable "administrator_org" {
  description = "The VCD administrator organization (Example: 'System')"
  type        = string
  default     = "System"
}

# ------------------------------------------------
# CSE configuration
# ------------------------------------------------
variable "capvcd_rde_version" {
  type        = string
  description = "Version of the CAPVCD Runtime Defined Entity Type"
  default     = "1.1.0"
}

# ------------------------------------------------
# Kubernetes cluster configuration
# ------------------------------------------------
variable "k8s_cluster_name" {
  description = "The name of the Kubernetes cluster. Name must contain only lowercase alphanumeric characters or '-' start with an alphabetic character, end with an alphanumeric, and contain at most 31 characters (Example: 'MyCluster')"
  type        = string
  validation {
    condition     = can(regex("^[a-z][a-z0-9-]{0,29}[a-z0-9]$", var.k8s_cluster_name))
    error_message = "Name must contain only lowercase alphanumeric characters or '-', start with an alphabetic character, end with an alphanumeric, and contain at most 31 characters."
  }
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

variable "cluster_author_api_token" {
  description = "API token of the Kubernetes cluster author"
  sensitive   = true
}

variable "ssh_public_key" {
  description = "SSH public key to be able to login to the cluster control plane nodes"
  default     = ""
}

variable "control_plane_machine_count" {
  description = "Number of control plane nodes (VMs)"
  type        = number
  validation {
    condition     = var.control_plane_machine_count > 0 && var.control_plane_machine_count % 2 != 0
    error_message = "Must be an odd number and higher than 0"
  }
  default = 3
}

variable "control_plane_sizing_policy" {
  description = "The VM Sizing Policy used for the control plane"
  default     = "tkg_s"
}

variable "control_plane_placement_policy" {
  description = "The VM Placement Policy used for the control plane"
  default     = ""
}

variable "control_plane_storage_profile" {
  description = "The Storage Profile used for the control plane"
  default     = "*"
}

variable "worker_machine_count" {
  description = "Number of worker nodes (VMs)"
  type        = number
  validation {
    condition     = var.worker_machine_count > 0
    error_message = "Must be higher than 0"
  }
  default = 2
}

variable "worker_sizing_policy" {
  description = "The VM Sizing Policy used for the workers"
  default     = "tkg_s"
}

variable "worker_placement_policy" {
  description = "The VM Placement Policy used for the workers"
  default     = ""
}

variable "worker_storage_profile" {
  description = "The Storage Profile used for the workers"
  default     = "*"
}

variable "disk_size" {
  description = "Disk size of every node"
  default     = "20Gi"
}

variable "tkgm_catalog" {
  description = "The TKGm Catalog used to pick the OVAs to create the Kubernetes cluster"
}

variable "tkgm_ova" {
  description = "The TKGm OVA to create the Kubernetes cluster"
}

variable "pod_cidr" {
  description = "The CIDR to use for the pods network"
  default     = "100.96.0.0/11"
}

variable "service_cidr" {
  description = "The CIDR to use for the pods network"
  default     = "100.64.0.0/13"
}

variable "tkr_version" {
  description = "String that defines the Tanzu Kubernetes release version inside the CAPVCD YAML template"
}

variable "tkg_version" {
  description = "String that defines the TKG version inside the CAPVCD YAML template"
}

variable "default_storage_class_filesystem" {
  description = "Filesystem for the default storage class"
  default     = "ext4"
  validation {
    condition     = var.default_storage_class_filesystem == "ext4" || var.default_storage_class_filesystem == "xfs"
    error_message = "Must be 'ext4' or 'xfs'"
  }
}

variable "default_storage_class_name" {
  description = "Name for the default storage class"
  default     = "default-storage-class-1"
  validation {
    condition     = can(regex("^[a-z][a-z0-9-]{0,62}[a-z0-9]$", var.default_storage_class_name))
    error_message = "Name must contain only lowercase alphanumeric characters or '-', start with an alphabetic character, end with an alphanumeric, and contain at most 63 characters."
  }
}

variable "default_storage_class_storage_profile" {
  description = "Storage Profile to use for the default storage class"
  default     = "*"
}

variable "default_storage_class_delete_reclaim_policy" {
  description = "Use a 'Delete' reclaim policy, that deletes the volume when the PersistentVolumeClaim is deleted"
  default     = "true"
}

variable "auto_repair_on_errors" {
  description = "If true, CSE attempts to recreate the clusters in error state. If false, it leaves the cluster in an error state for manual troubleshooting"
  default     = "true"
}