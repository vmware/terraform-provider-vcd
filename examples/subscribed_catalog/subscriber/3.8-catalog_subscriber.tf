variable "user" {}
variable "password" {}
variable "vcd_url" {}
variable "org" {}
variable "vdc" {}

provider "vcd" {
  user                 = var.user
  password             = var.password
  token                = ""
  api_token            = ""
  auth_type            = "integrated"
  saml_adfs_rpt_id     = ""
  url                  = "${var.vcd_url}/api"
  sysorg               = "System"
  org                  = var.org
  vdc                  = var.vdc
  allow_unverified_ssl = true
  max_retry_timeout    = 600
  logging              = true
  logging_file         = "go-vcloud-director.log"
}

data "vcd_storage_profile" "storage_profile" {
  org  = var.org
  vdc  = var.vdc
  name = "*"
}

data "local_sensitive_file" "password_file" {
  filename = pathexpand("subscription_password.txt")
}

resource "vcd_subscribed_catalog" "remote-subscriber" {
  org                = var.org
  name               = "remote-subscriber"
  storage_profile_id = data.vcd_storage_profile.storage_profile.id

  delete_force     = true
  delete_recursive = true

  subscription_url      = var.subscription_url
  make_local_copy       = true
  subscription_password = chomp(data.local_sensitive_file.password_file.content)

  sync_on_refresh = true
  sync_catalog    = true
}

output "num_subscribed_templates" {
  value = vcd_subscribed_catalog.remote-subscriber.number_of_vapp_templates
}
output "num_subscribed_media" {
  value = vcd_subscribed_catalog.remote-subscriber.number_of_media
}

/** /
data "vcd_catalog_vapp_template" "subscribed_templates" {
  count      = length(vcd_subscribed_catalog.remote-subscriber.vapp_template_list)
  org        = var.org
  catalog_id = vcd_subscribed_catalog.remote-subscriber.id
  name       = vcd_subscribed_catalog.remote-subscriber.vapp_template_list[count.index]
}

data "vcd_catalog_media" "subscribed_media" {
  count   = length(vcd_subscribed_catalog.remote-subscriber.media_item_list)
  org     = var.org
  catalog = vcd_subscribed_catalog.remote-subscriber.name
  name    = vcd_subscribed_catalog.remote-subscriber.media_item_list[count.index]
}

output "templates" {
  value = data.vcd_catalog_vapp_template.subscribed_templates
}

output "media" {
  value = data.vcd_catalog_media.subscribed_media
}

/**/
