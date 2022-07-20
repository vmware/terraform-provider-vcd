# These variables are for configuring the VCD provider

variable "admin-user" {
  description = "The System administrator user that will create the CSE service account role and user"
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
  description = "The CSE service account user name that will install CSE. It will be created."
  default     = "cse_service_account"
  type        = string
}

variable "service-account-password" {
  description = "The CSE service account password to put to the CSE service account."
  default     = "cse_service_account"
  sensitive   = true
  type        = string
}

variable "vcd-url" {
  description = "The target VCD url, like 'https://my-vcd.company.com/api'"
  type        = string
}
