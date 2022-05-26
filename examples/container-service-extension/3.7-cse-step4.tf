# --------------------------------------------------------------------------------------------------------
# WARNING: This is done after applying `3.7-cse-step3`.
# --------------------------------------------------------------------------------------------------------

# Here we finish what we started on `3.7-cse-step3`. Notice the publish_to_all_tenants is now true.
# Please make sure you executed the `terraform import` from previous step before applying this HCL.

data "vcd_rights_bundle" "cse-rights-bundle" {
  name = "cse:nativeCluster Entitlement"
}

resource "vcd_rights_bundle" "published-cse-rights-bundle" {
  name                   = data.vcd_rights_bundle.cse-rights-bundle.name
  description            = data.vcd_rights_bundle.cse-rights-bundle.description
  rights                 = data.vcd_rights_bundle.cse-rights-bundle.rights
  publish_to_all_tenants = true
}