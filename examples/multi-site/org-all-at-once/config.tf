provider "vcd" {
  user                 = var.vcd_admin
  password             = var.vcd_password
  token                = ""
  api_token            = ""
  auth_type            = "integrated"
  saml_adfs_rpt_id     = ""
  url                  = var.vcd_url
  sysorg               = var.vcd_sysorg
  org                  = var.vcd_org1
  allow_unverified_ssl = "true"
  max_retry_timeout    = 600
  logging              = true
  logging_file         = "go-vcloud-director.log"
}

data "vcd_org" "org1" {
  name     = var.vcd_org1
}
data "vcd_org" "org2" {
  name     = var.vcd_org2
}

data "vcd_multisite_org_data" "org1-data" {
  org_id           = data.vcd_org.org1.id
}

data "vcd_multisite_org_data" "org2-data" {
  org_id           = data.vcd_org.org2.id
}

data "vcd_resource_list" "orgs" {
  name          = "org_associations"
  resource_type = "vcd_multisite_org_association"
  list_mode     = "name_id"
}

resource "vcd_multisite_org_association" "org1-org2" {
  org_id                  = data.vcd_org.org1.id
  association_data        = data.vcd_multisite_org_data.org2-data.association_data
  connection_timeout_mins = 2
}

resource "vcd_multisite_org_association" "org2-org1" {
  org_id                  = data.vcd_org.org2.id
  association_data        = data.vcd_multisite_org_data.org1-data.association_data
  connection_timeout_mins = 2
}

output "org2-org1" {
  value = vcd_multisite_org_association.org2-org1
}


output "org1-org2" {
  value = vcd_multisite_org_association.org1-org2
}

output "org_associations" {
  value = data.vcd_resource_list.orgs
}
