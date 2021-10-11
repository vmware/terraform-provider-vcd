provider "vcd" {
  user                 = var.adminUser
  password             = var.adminPassword
  auth_type            = "integrated"
  url                  = var.url
  sysorg               = var.sysOrg
  org                  = var.org
  allow_unverified_ssl = "true"
  max_retry_timeout    = 600
  logging              = true
  logging_file         = "go-vcloud-director-provider.log"
}

// PRE-REQUISITE:
// For this script to work correctly, the rights bundle "Defaults Rights Bundle" must be modified
// – before running this script – to NOT PUBLISH TO ALL TENANTS, but to publish explicitly to existing tenants only.

// Creates a new org
resource "vcd_org" "another-org" {
  name             = var.org
  full_name        = "another org"
  description      = "Organization ${var.org}"
  delete_force     = "true"
  delete_recursive = "true"
}

// Creates a new user. It is used for the credentials of the tenant script
resource "vcd_org_user" "another-user" {
  org            = vcd_org.another-org.name
  name           = var.orgUser
  password       = var.orgPassword
  role           = "Organization Administrator"
  take_ownership = true
}

// Gets the data of the defaults rights
data "vcd_rights_bundle" "defaults" {
  name = "Default Rights Bundle"
}

// Creates a new defaults rights bundle, published only to the new org
resource "vcd_rights_bundle" "new-defaults" {
  name                   = "new-defaults"
  description            = "new defaults rights"
  publish_to_all_tenants = false

  # the tenant will have an extra right
  rights = setunion(
    data.vcd_rights_bundle.defaults.rights, # rights from existing rights bundle
    ["API Explorer: View"]                  # rights to be added
  )
  tenants = [vcd_org.another-org.name]
}

# Gets the existing global role for "vApp Author"
data "vcd_global_role" "vapp-author" {
  name = "vApp Author"
}

# Gets the existing global role for "Catalog Author"
data "vcd_global_role" "catalog-author" {
  name = "Catalog Author"
}

// Makes a new global role combining vApp Author and Catalog Author
resource "vcd_global_role" "super-vapp-author" {
  name                   = "super-vapp-author"
  description            = "A global role from CLI"
  publish_to_all_tenants = false
  rights = setunion(
    data.vcd_global_role.vapp-author.rights,    # rights from existing global role
    data.vcd_global_role.catalog-author.rights, # rights from existing global role
    ["API Explorer: View"],                     # more rights to be added
  )
  tenants = [vcd_org.another-org.name]
}
