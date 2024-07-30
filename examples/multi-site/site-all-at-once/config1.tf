provider "vcd" {
  alias                = "vcd1"
  user                 = var.vcd_admin1
  password             = var.vcd_password1
  token                = ""
  api_token            = ""
  auth_type            = "integrated"
  saml_adfs_rpt_id     = ""
  url                  = var.vcd_url1
  sysorg               = var.vcd_sysorg1
  org                  = var.vcd_org1
  allow_unverified_ssl = "true"
  max_retry_timeout    = 600
  logging              = true
  logging_file         = "go-vcloud-director-1.log"
}

data "vcd_org" "org1" {
  provider = vcd.vcd1
  name     = var.vcd_org1
}

data "vcd_multisite_org_data" "org1-data" {
  provider         = vcd.vcd1
  org_id           = data.vcd_org.org1.id
  download_to_file = "${var.vcd_org1}.xml"
}

data "vcd_multisite_site_data" "site1-data" {
  provider         = vcd.vcd1
  download_to_file = "${var.site_name1}-site.xml"
}

# The data files are needed because the data will be read by resources handled by a different user, which can't read
# directly the data source that is the origin of these files

data "local_file" "site1" {
  filename   = "${var.site_name1}-site.xml"
  depends_on = [data.vcd_multisite_site_data.site1-data]
}

data "local_file" "org1" {
  filename   = "${var.vcd_org1}.xml"
  depends_on = [data.vcd_multisite_org_data.org1-data]
}
