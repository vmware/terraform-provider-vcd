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
# CSE administrator user configuration
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
# CSE Runtime Defined Entities setup
# ------------------------------------------------
variable "capvcd_rde_version" {
  type        = string
  description = "Version of the CAPVCD Runtime Defined Entity Type"
  default     = "1.2.0"
}
