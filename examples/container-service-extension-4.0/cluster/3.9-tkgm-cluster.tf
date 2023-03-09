# ------------------------------------------------------------------------------------------------------------
# TKGm cluster creation
#
# * Please read the guide present at https://registry.terraform.io/providers/vmware/vcd/latest/docs/guides/container_service_extension_4_0
#   before applying this configuration.
#
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
  sysorg               = "System"
  org                  = "cluster_org"
  allow_unverified_ssl = var.insecure_login
  logging              = true
  logging_file         = "cluster_${var.administrator_user}.log"
}

data "vcd_vm_sizing_policy" "cluster_sizing_policy" {
  name = "TKG small"
}

data "vcd_catalog" "cluster_catalog" {
  org  = "solutions_org"
  name = "tkgm_catalog"
}

data "vcd_catalog_vapp_template" "cluster_tkgm_ova" {
  catalog_id = data.vcd_catalog.cluster_catalog.id
  name       = "ubuntu-2004-kube-v1.22.9+vmware.1-tkg.1-2182cbabee08edf480ee9bc5866d6933"
}

data "vcd_role" "cluster_author_role" {
  name = "Kubernetes Cluster Author"
}

resource "vcd_org_user" "cluster_author" {
  name     = "cluster_author"
  password = "change-me"
  role     = data.vcd_role.cluster_author_role.name
}

provider "vcd" {
  alias                = "cluster_author"
  url                  = "${var.vcd_url}/api"
  user                 = vcd_org_user.cluster_author.name
  password             = vcd_org_user.cluster_author.password
  auth_type            = "integrated"
  org                  = "cluster_org"
  allow_unverified_ssl = var.insecure_login
  logging              = true
  logging_file         = "cluster_${vcd_org_user.cluster_author.name}.log"
}

locals {
  capvcd_yaml_rendered = templatefile("./cluster-template-v1.22.9.yaml", {
    CLUSTER_NAME     = "my-cluster"
    TARGET_NAMESPACE = "my-cluster-ns"

    VCD_SITE                     = var.vcd_url
    VCD_ORGANIZATION             = "cluster_org"
    VCD_ORGANIZATION_VDC         = "cluster_vdc"
    VCD_ORGANIZATION_VDC_NETWORK = "cluster_routed_network"

    VCD_USERNAME_B64      = vcd_org_user.cluster_author.name
    VCD_PASSWORD_B64      = vcd_org_user.cluster_author.password
    VCD_REFRESH_TOKEN_B64 = ""
    SSH_PUBLIC_KEY        = ""

    CONTROL_PLANE_MACHINE_COUNT        = 1
    VCD_CONTROL_PLANE_SIZING_POLICY    = data.vcd_vm_sizing_policy.cluster_sizing_policy.name
    VCD_CONTROL_PLANE_PLACEMENT_POLICY = ""
    VCD_CONTROL_PLANE_STORAGE_PROFILE  = ""

    VCD_WORKER_STORAGE_PROFILE  = ""
    VCD_WORKER_SIZING_POLICY    = data.vcd_vm_sizing_policy.cluster_sizing_policy.name
    VCD_WORKER_PLACEMENT_POLICY = ""
    WORKER_MACHINE_COUNT        = 1

    DISK_SIZE         = "20Gi"
    VCD_CATALOG       = data.vcd_catalog.cluster_catalog.name
    VCD_TEMPLATE_NAME = data.vcd_catalog_vapp_template.cluster_tkgm_ova.name

    POD_CIDR     = "100.96.0.0/11"
    SERVICE_CIDR = "100.64.0.0/13"
  })
}