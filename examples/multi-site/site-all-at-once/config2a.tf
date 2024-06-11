
resource "vcd_multisite_site_association" "site2-site1" {
  provider                = vcd.vcd2
  association_data        = data.local_file.site1.content
  connection_timeout_mins = 2
}

resource "vcd_multisite_org_association" "org2-org1" {
  provider                = vcd.vcd2
  org_id                  = data.vcd_org.org2.id
  association_data        = data.local_file.org1.content
  connection_timeout_mins = 2
  depends_on              = [vcd_multisite_site_association.site2-site1]
}

output "site2-site1" {
  value = vcd_multisite_site_association.site2-site1
}

output "org2-org1" {
  value = vcd_multisite_org_association.org2-org1
}
