# ------------------------------------------------------------------------------------------------------------
# CSE v4.1 installation, step 1:
#
# * Please read the guide at https://registry.terraform.io/providers/vmware/vcd/latest/docs/guides/container_service_extension_4_x_install
#   before applying this configuration.
#
# * The installation process is split into two steps as the first one creates a CSE admin user that needs to be
#   used in a "provider" block in the second one.
#
# * This file contains the same resources created by the "Configure Settings for CSE Server > Set Configuration Parameters" step in the
#   UI wizard.
#
# * Rename "terraform.tfvars.example" to "terraform.tfvars" and adapt the values to your needs.
#   Other than that, this snippet should be applied as it is.
#   You can check the comments on the resource for context.
# ------------------------------------------------------------------------------------------------------------

# This RDE configures the CSE Server. It can be customised through variables, and the bootstrap_cluster_sizing_policy
# can also be changed.
# Other than that, this should be applied as it is.
resource "vcd_rde" "vcdkeconfig_instance" {
  org         = var.administrator_org
  name        = "vcdKeConfig"
  rde_type_id = vcd_rde_type.vcdkeconfig_type.id
  resolve     = true
  input_entity = templatefile(var.vcdkeconfig_template_filepath, {
    capvcd_version                = var.capvcd_version
    cpi_version                   = var.cpi_version
    csi_version                   = var.csi_version
    github_personal_access_token  = var.github_personal_access_token
    bootstrap_vm_sizing_policy    = vcd_vm_sizing_policy.tkg_s.name # References the small VM Sizing Policy, it can be changed.
    no_proxy                      = var.no_proxy
    http_proxy                    = var.http_proxy
    https_proxy                   = var.https_proxy
    syslog_host                   = var.syslog_host
    syslog_port                   = var.syslog_port
    node_startup_timeout          = var.node_startup_timeout
    node_not_ready_timeout        = var.node_not_ready_timeout
    node_unknown_timeout          = var.node_unknown_timeout
    max_unhealthy_node_percentage = var.max_unhealthy_node_percentage
    container_registry_url        = var.container_registry_url
    k8s_cluster_certificates      = join(",", var.k8s_cluster_certificates)
    bootstrap_vm_certificates     = join(",", var.bootstrap_vm_certificates)
  })
}
