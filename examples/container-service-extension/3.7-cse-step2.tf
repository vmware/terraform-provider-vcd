# --------------------------------------------------------------------------------------------------------
# WARNING: This is done after running `cse install` command.
# --------------------------------------------------------------------------------------------------------

# Here we create a new rights bundle for CSE, with the rights assigned already to the Default Rights Bundle (hence the
# data source) plus new ones coming from `cse install` command.

data "vcd_rights_bundle" "default-rb" {
  name = "Default Rights Bundle"
}

resource "vcd_rights_bundle" "cse-rb" {
  name        = "CSE Rights Bundle"
  description = "Rights bundle to manage CSE"
  rights = setunion(data.vcd_rights_bundle.default-rb.rights, [
    "API Tokens: Manage",
    "Organization vDC Shared Named Disk: Create",
    "cse:nativeCluster: View",
    "cse:nativeCluster: Full Access",
    "cse:nativeCluster: Modify"
  ])
  publish_to_all_tenants = true # Here we publish to all tenants for simplicity, but you can select the tenant in which CSE is used
}

# Here we fetch the rights bundle that `cse install` creates. It is not published by default, so first we
# create the resource in Terraform. In step 3 we'll import the state from VCD and publish it to the tenants.

data "vcd_rights_bundle" "cse-rights-bundle" {
  name = "cse:nativeCluster Entitlement"
}

resource "vcd_rights_bundle" "published-cse-rights-bundle" {
  name                   = data.vcd_rights_bundle.cse-rights-bundle.name
  description            = data.vcd_rights_bundle.cse-rights-bundle.description
  rights                 = data.vcd_rights_bundle.cse-rights-bundle.rights
  publish_to_all_tenants = false
}

# Create a new role for CSE, with the new rights to create clusters and manage them.

data "vcd_role" "vapp_author" {
  org  = vcd_org.cse_org.name
  name = "vApp Author"
}

resource "vcd_role" "cluster_author" {
  org         = vcd_org.cse_org.name
  name        = "Cluster Author"
  description = "Can read and create clusters"
  rights = setunion(data.vcd_role.vapp_author.rights, [
    "API Tokens: Manage",
    "Organization vDC Shared Named Disk: Create",
    "Organization vDC Gateway: View",
    "Organization vDC Gateway: View Load Balancer",
    "Organization vDC Gateway: Configure Load Balancer",
    "Organization vDC Gateway: View NAT",
    "Organization vDC Gateway: Configure NAT",
    "cse:nativeCluster: View",
    "cse:nativeCluster: Full Access",
    "cse:nativeCluster: Modify",
    "Certificate Library: View" # Implicit role needed
  ])

  depends_on = [vcd_rights_bundle.cse-rb]
}

# Create a user with that role.

resource "vcd_org_user" "cse_user" {
  org = vcd_org.cse_org.name

  name        = "cse_user"
  description = "Cluster author"
  role        = vcd_role.cluster_author.name
  password    = "****"
}
