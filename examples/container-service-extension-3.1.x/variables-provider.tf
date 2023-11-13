# ------------------------------------------------------------------------------------------------------------
# WARNING: This CSE installation method is deprecated in favor of CSE v4.x. Please have a look at
#          https://registry.terraform.io/providers/vmware/vcd/latest/docs/guides/container_service_extension_4_x_install
# ------------------------------------------------------------------------------------------------------------

# These variables are for configuring the VCD provider

variable "admin-user" {
  description = "The System administrator user that will create basic infrastructure and the CSE service account"
  default     = "administrator"
  type        = string
}

variable "admin-password" {
  description = "The System administrator password"
  sensitive   = true
  type        = string
}

variable "admin-org" {
  description = "The System administrator organization"
  default     = "System"
  sensitive   = true
  type        = string
}

variable "service-account-user" {
  description = "The CSE service account user name that will install CSE. It should be a new user."
  default     = "cse_service_account"
  type        = string
}

variable "service-account-password" {
  description = "The CSE service account password to put to the new CSE service account."
  default     = "cse_service_account"
  sensitive   = true
  type        = string
}

variable "vcd-url" {
  description = "The target VCD url, like 'https://my-vcd.company.com/api'"
  type        = string
}
