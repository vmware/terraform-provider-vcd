variable "user" {}
variable "password" {}
variable "vcd_url" {}
variable "org" {}
variable "vdc" {}
variable "ova_path" {}

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
  allow_unverified_ssl = "true"
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

resource "vcd_catalog" "test-publisher" {
  org                = var.org
  name               = "test-publisher"
  description        = "test publisher catalog"
  storage_profile_id = data.vcd_storage_profile.storage_profile.id

  delete_force     = "true"
  delete_recursive = "true"

  metadata = {
    identity     = "published catalog"
    origin       = var.org
    host_version = "10.4.1"
  }
  publish_enabled               = "true"
  cache_enabled                 = "true"
  preserve_identity_information = "false"
  password                      = chomp(data.local_sensitive_file.password_file.content)
}

resource "vcd_catalog_vapp_template" "test_vt" {
  count      = 10
  org        = var.org
  catalog_id = vcd_catalog.test-publisher.id

  name              = "test-vt-${count.index}"
  description       = "test vapp template test-vt-${count.index}"
  ova_path          = "${var.ova_path}/test_vapp_template.ova"
  upload_piece_size = 5

  metadata = {
    identity       = "published catalog item"
    origin         = var.org
    parent         = vcd_catalog.test-publisher.name
    parent_created = vcd_catalog.test-publisher.created
    host_version   = "10.4.1"
  }
}

resource "vcd_catalog_media" "test_media" {
  count   = 20
  org     = var.org
  catalog = vcd_catalog.test-publisher.name

  name              = "test-media-${count.index}"
  description       = "test media item test-media-${count.index}"
  media_path        = "${var.ova_path}/test.iso"
  upload_piece_size = 5
  metadata = {
    identity       = "published catalog item"
    origin         = var.org
    parent         = vcd_catalog.test-publisher.name
    parent_created = vcd_catalog.test-publisher.created
    host_version   = "10.4.1"
  }
}

output "publishing_url" {
  value = vcd_catalog.test-publisher.publish_subscription_url
}

output "num_published_templates" {
  value = vcd_catalog.test-publisher.number_of_vapp_templates
}

output "num_published_media" {
  value = vcd_catalog.test-publisher.number_of_media
}
