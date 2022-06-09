# These variables are for configuring the VCD provider

variable "admin-user" {
  description = "The System administrator user that will install CSE"
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

variable "vcd-url" {
  description = "The target VCD url, like 'https://my-vcd.company.com/api'"
  type        = string
}

variable "vcenter-name" {
  description = "The name of the vCenter used in VCD, for CSE installation process"
  type        = string
}

variable "vcenter-username" {
  description = "The name of the vCenter username, for CSE installation process"
  type        = string
}

variable "vcenter-password" {
  description = "The name of the vCenter username, for CSE installation process"
  sensitive   = true
  type        = string
}
