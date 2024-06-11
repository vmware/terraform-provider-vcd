
data "vcd_resource_list" "sites" {
  provider      = vcd.vcd1
  name          = "site_associations"
  resource_type = "vcd_multisite_site_association"
  list_mode     = "name_id"
}

data "vcd_resource_list" "orgs" {
  provider      = vcd.vcd1
  name          = "org_associations"
  resource_type = "vcd_multisite_org_association"
  list_mode     = "name_id"
}

resource "vcd_multisite_site_association" "site1-site2" {
  provider                = vcd.vcd1
  association_data        = data.local_file.site2.content
  connection_timeout_mins = 2
}

resource "vcd_multisite_org_association" "org1-org2" {
  provider                = vcd.vcd1
  org_id                  = data.vcd_org.org1.id
  association_data        = data.local_file.org2.content
  connection_timeout_mins = 2
  depends_on              = [vcd_multisite_site_association.site1-site2]
}

output "site1-site2" {
  value = vcd_multisite_site_association.site1-site2
}

output "org1-org2" {
  value = vcd_multisite_org_association.org1-org2
}

output "site_associations" {
  value = data.vcd_resource_list.sites
}
output "org_associations" {
  value = data.vcd_resource_list.orgs
}
