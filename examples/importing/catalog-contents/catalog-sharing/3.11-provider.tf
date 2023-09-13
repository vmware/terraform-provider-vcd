variable "user" {}
variable "password" {}
variable "vcd_url" {}
variable "org" {}
variable "catalog" {}
variable "ova_path" {}
variable "new_user" {}
variable "new_password" {}

provider "vcd" {
  user                 = var.user
  password             = var.password
  token                = ""
  api_token            = ""
  auth_type            = "integrated"
  saml_adfs_rpt_id     = ""
  url                  = "${var.vcd_url}/api"
  sysorg               = var.org
  org                  = var.org
  vdc                  = ""
  allow_unverified_ssl = true
  max_retry_timeout    = 600
  logging              = true
  logging_file         = "go-vcloud-director.log"
}
