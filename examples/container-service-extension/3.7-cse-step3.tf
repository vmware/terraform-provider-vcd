# --------------------------------------------------------------------------------------------------------
# WARNING: This is done after running `cse install` command.
# --------------------------------------------------------------------------------------------------------

# Here we publish the rights bundle that `cse install` creates to the desired tenants. For that, first we
# create the resource in Terraform, to have its state created. Next, we'll import the state from VCD.

data "vcd_rights_bundle" "cse-rights-bundle" {
  name = "cse:nativeCluster Entitlement"
}

resource "vcd_rights_bundle" "published-cse-rights-bundle" {
  name                   = data.vcd_rights_bundle.cse-rights-bundle.name
  description            = data.vcd_rights_bundle.cse-rights-bundle.description
  rights                 = data.vcd_rights_bundle.cse-rights-bundle.rights
  publish_to_all_tenants = false
}

# After being created, you need to execute:
#     terraform import vcd_rights_bundle.published-cse-rights-bundle "cse:nativeCluster Entitlement"
