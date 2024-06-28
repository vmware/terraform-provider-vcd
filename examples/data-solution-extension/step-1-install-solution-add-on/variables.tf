
variable "vcd_url" {
   type = string
}

variable "vcd_admin" {
  type = string
  default = "administrator"
}

variable "vcd_sysorg" {
  type = string
  default = "System"
}

variable "vcd_password" {
  type = string
}

variable "vcd_solutions_org" {
  type = string
}

variable "vcd_solutions_vdc" {
  type = string
}

variable "vcd_solutions_vdc_routed_network" {
  type = string
}

variable "vcd_solutions_vdc_storage_profile_name" {
  type = string
}

variable "vcd_dse_add_on_iso_path" {
  type = string
}

variable "vcd_tenant_org" {
  type = string
}
