
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

variable "vcd_tenant_org" {
  type = string
}

variable "vcd_tenant_user" {
  type = string
}

variable "vcd_tenant_password" {
  type = string
}
