provider "vcd" {
  user                 = var.orgUser
  password             = var.orgPassword
  auth_type            = "integrated"
  url                  = var.url
  sysorg               = var.org
  org                  = var.org
  allow_unverified_ssl = "true"
  max_retry_timeout    = 600
  logging              = true
  logging_file         = "go-vcloud-director-org.log"
}

# Pre requisite:
# run the corresponding provider script first

// The new role, deriving from the global role, is now available in the organization, as seen by the tenant
data "vcd_role" "super-vapp-author" {
  org  = var.org
  name = "super-vapp-author"
}

output "org_role" {
  value = data.vcd_role.super-vapp-author
}


output "super_vapp_author_rights" {
  value = length(data.vcd_role.super-vapp-author.rights)
}
