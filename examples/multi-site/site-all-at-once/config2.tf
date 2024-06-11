provider "vcd" {
  alias                = "vcd2"
  user                 = var.vcd_admin2
  password             = var.vcd_password2
  token                = ""
  api_token            = ""
  auth_type            = "integrated"
  saml_adfs_rpt_id     = ""
  url                  = var.vcd_url2
  sysorg               = var.vcd_sysorg2
  org                  = var.vcd_org2
  allow_unverified_ssl = "true"
  max_retry_timeout    = 600
  logging              = true
  logging_file         = "go-vcloud-director-2.log"
}

data "vcd_org" "org2" {
  provider = vcd.vcd2
  name     = var.vcd_org2
}

data "vcd_multisite_org_data" "org2-data" {
  provider         = vcd.vcd2
  org_id           = data.vcd_org.org2.id
  download_to_file = "${var.vcd_org2}.xml"
}

data "vcd_multisite_site_data" "site2-data" {
  provider         = vcd.vcd2
  download_to_file = "${var.site_name2}-site.xml"
}

# The data files are needed because the data will be read by resources handled by a different user, which can't read
# directly the data source that is the origin of these files

data "local_file" "site2" {
  filename   = "${var.site_name2}-site.xml"
  depends_on = [data.vcd_multisite_site_data.site2-data]
}

data "local_file" "org2" {
  filename   = "${var.vcd_org2}.xml"
  depends_on = [data.vcd_multisite_org_data.org2-data]
}
