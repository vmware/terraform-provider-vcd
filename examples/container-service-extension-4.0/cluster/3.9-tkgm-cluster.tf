# ------------------------------------------------------------------------------------------------------------
# TKGm cluster creation
#
# * Please read the guide present at https://registry.terraform.io/providers/vmware/vcd/latest/docs/guides/container_service_extension_4_0
#   before applying this configuration.
#
# * The installation process is split into two steps as Providers will need to generate an API token for the created
#   CSE administrator user, in order to use it with the CSE Server that will be deployed in the second step.
#
# * This step will only create the required Runtime Defined Entity (RDE) Interfaces, Types, Role and finally
#   the CSE administrator user.
#
# * Rename "terraform.tfvars.example" to "terraform.tfvars" and adapt the values to your needs.
#   Other than that, this snippet should be applied as it is.
#   You can check the comments on each resource/data source for more help and context.
# ------------------------------------------------------------------------------------------------------------

# VCD Provider configuration. It must be at least v3.9.0 and configured with a System administrator account.
terraform {
  required_providers {
    vcd = {
      source  = "vmware/vcd"
      version = ">= 3.9"
    }
  }
}

provider "vcd" {
  url                  = "${var.vcd_url}/api"
  user                 = var.administrator_user
  password             = var.administrator_password
  auth_type            = "integrated"
  sysorg               = var.administrator_org
  org                  = var.administrator_org
  allow_unverified_ssl = var.insecure_login
  logging              = true
  logging_file         = "cse_install_step1.log"
}

locals {
  capvcd_yaml_rendered = templatefile("/users/bob/capvcd-templates/cluster-template-v1.22.9.yaml", {
    CLUSTER_NAME     = "my-cluster"
    TARGET_NAMESPACE = "my-cluster-ns"

    VCD_SITE                     = var.vcd_url
    VCD_ORGANIZATION             = "cluster_org"
    VCD_ORGANIZATION_VDC         = "cluster_vdc"
    VCD_ORGANIZATION_VDC_NETWORK = "cluster_routed_network"

    VCD_USERNAME_B64      = base64encode(var.k8s_cluster_user)
    VCD_REFRESH_TOKEN_B64 = base64encode(var.k8s_cluster_api_token)
    SSH_PUBLIC_KEY        = ""

    CONTROL_PLANE_MACHINE_COUNT        = 1
    VCD_CONTROL_PLANE_SIZING_POLICY    =
    VCD_CONTROL_PLANE_PLACEMENT_POLICY = ""
    VCD_CONTROL_PLANE_STORAGE_PROFILE  = ""

    VCD_WORKER_STORAGE_PROFILE  = ""
    VCD_WORKER_SIZING_POLICY    = vcd_vm_sizing_policy.default_policy.name
    VCD_WORKER_PLACEMENT_POLICY = ""
    WORKER_MACHINE_COUNT        = 1

    DISK_SIZE         = "20Gi"
    VCD_CATALOG       = vcd_catalog.cse_catalog.name
    VCD_TEMPLATE_NAME = replace(var.tkgm_ova_name, ".ova", "")

    POD_CIDR     = "100.96.0.0/11"
    SERVICE_CIDR = "100.64.0.0/13"
  })
}