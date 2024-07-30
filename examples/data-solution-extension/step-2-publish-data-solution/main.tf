provider "vcd" {
  user                 = var.vcd_admin
  password             = var.vcd_password
  auth_type            = "integrated"
  url                  = var.vcd_url
  sysorg               = var.vcd_sysorg
  org                  = var.vcd_tenant_org
  allow_unverified_ssl = "true"
  max_retry_timeout    = 600
  logging              = true
  logging_file         = "go-vcloud-director.log"
}

data "vcd_org" "dse-consumer" {
  name = var.vcd_tenant_org
}

# Create a combination of rights for Container Service Extention (CSE) and Data Solution Extension
# (DSE)

data "vcd_rights_bundle" "dse-rb" {
  name = "vmware:dataSolutionsRightsBundle"
}

data "vcd_rights_bundle" "k8s-rights" {
  name = "Kubernetes Clusters Rights Bundle"
}

resource "vcd_global_role" "dse" {
  name                   = "DSE Role"
  description            = "Global role for consuming DSE"
  rights                 = setunion(data.vcd_rights_bundle.k8s-rights.rights, data.vcd_rights_bundle.dse-rb.rights)
  publish_to_all_tenants = false
  tenants = [
    data.vcd_org.dse-consumer.name
  ]
}

resource "vcd_org_user" "my-org-admin" {
  org = data.vcd_org.dse-consumer.name

  name        = var.vcd_tenant_user
  description = "DSE User"
  role        = vcd_global_role.dse.name
  password    = var.vcd_tenant_password

  depends_on = [vcd_global_role.dse]
}


# Configure VCD Data Solutions (DSO) and Mongo DB Community version with default repositories
resource "vcd_dse_registry_configuration" "dso" {
  name               = "VCD Data Solutions"
  use_default_values = true
}

resource "vcd_dse_registry_configuration" "mongodb-community" {
  name               = "MongoDB Community"
  use_default_values = true
}

# Publish Mongo DB Data Solution to tenant
resource "vcd_dse_solution_publish" "mongodb-community" {
  data_solution_id = vcd_dse_registry_configuration.mongodb-community.id

  org_id = data.vcd_org.dse-consumer.id
}

output "vcd_url" {
  value = replace(var.vcd_url, "/api", "/tenant/${data.vcd_org.dse-consumer.name}")
}

output "tenant_org" {
  value = data.vcd_org.dse-consumer.name
}

output "tenant_user" {
  value = vcd_org_user.my-org-admin.name
}

output "tenant_password" {
  value     = vcd_org_user.my-org-admin.password
  sensitive = true
}

