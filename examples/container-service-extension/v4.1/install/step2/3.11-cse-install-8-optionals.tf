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

# This resource installs the UI Plugin. It can be useful for tenant users that are not familiar with
# Terraform.
resource "vcd_ui_plugin" "k8s_container_clusters_ui_plugin" {
  count       = var.k8s_container_clusters_ui_plugin_path == "" ? 0 : 1
  plugin_path = var.k8s_container_clusters_ui_plugin_path
  enabled     = true
  tenant_ids = [
    data.vcd_org.system_org.id,
    vcd_org.solutions_organization.id,
    vcd_org.tenant_organization.id,
  ]
}

data "vcd_org" "system_org" {
  name = var.administrator_org
}
