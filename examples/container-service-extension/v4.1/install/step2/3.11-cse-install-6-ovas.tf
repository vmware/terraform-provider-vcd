# ------------------------------------------------------------------------------------------------------------
# CSE v4.1 installation:
#
# * Please read the guide at https://registry.terraform.io/providers/vmware/vcd/latest/docs/guides/container_service_extension_4_x_install
#   before applying this configuration.
#
# * Rename "terraform.tfvars.example" to "terraform.tfvars" and adapt the values to your needs.
#   Other than that, this snippet should be applied as it is.
#   You can check the comments on each resource/data source for more help and context.
# ------------------------------------------------------------------------------------------------------------

# In this section we create two Catalogs, one to host all CSE Server OVAs and another one to host TKGm OVAs.
# They are created in the Solutions organization and only the TKGm will be shared as read-only. This will guarantee
# that only CSE admins can manage OVAs.
resource "vcd_catalog" "cse_catalog" {
  org  = vcd_org.solutions_organization.name # References the Solutions Organization created previously
  name = "cse_catalog"

  delete_force     = "true"
  delete_recursive = "true"

  # In this example, everything is created from scratch, so it is needed to wait for the VDC to be available, so the
  # Catalog can be created.
  depends_on = [
    vcd_org_vdc.solutions_vdc
  ]
}

resource "vcd_catalog" "tkgm_catalog" {
  org  = vcd_org.solutions_organization.name # References the Solutions Organization
  name = "tkgm_catalog"

  delete_force     = "true"
  delete_recursive = "true"

  # In this example, everything is created from scratch, so it is needed to wait for the VDC to be available, so the
  # Catalog can be created.
  depends_on = [
    vcd_org_vdc.solutions_vdc
  ]
}

# We share the TKGm Catalog with the Tenant Organization created previously.
resource "vcd_catalog_access_control" "tkgm_catalog_ac" {
  org                  = vcd_org.solutions_organization.name # References the Solutions Organization created previously
  catalog_id           = vcd_catalog.tkgm_catalog.id
  shared_with_everyone = false
  shared_with {
    org_id       = vcd_org.tenant_organization.id # Shared with the Tenant Organization
    access_level = "ReadOnly"
  }
}

# We upload a minimum set of OVAs for CSE to work. Read the official documentation to check
# where to find the OVAs:
# https://docs.vmware.com/en/VMware-Cloud-Director-Container-Service-Extension/index.html
resource "vcd_catalog_vapp_template" "tkgm_ova" {
  for_each   = toset(var.tkgm_ova_files)
  org        = vcd_org.solutions_organization.name # References the Solutions Organization created previously
  catalog_id = vcd_catalog.tkgm_catalog.id         # References the TKGm Catalog created previously

  name        = replace(each.key, ".ova", "")
  description = replace(each.key, ".ova", "")
  ova_path    = format("%s/%s", var.tkgm_ova_folder, each.key)
}

resource "vcd_catalog_vapp_template" "cse_ova" {
  org        = vcd_org.solutions_organization.name # References the Solutions Organization created previously
  catalog_id = vcd_catalog.cse_catalog.id          # References the CSE Catalog created previously

  name        = replace(var.cse_ova_file, ".ova", "")
  description = replace(var.cse_ova_file, ".ova", "")
  ova_path    = format("%s/%s", var.cse_ova_folder, var.cse_ova_file)
}

