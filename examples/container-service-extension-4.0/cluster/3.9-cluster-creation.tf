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
  user                 = var.cluster_author_user
  password             = var.cluster_author_password
  auth_type            = "integrated"
  org                  = var.cluster_organization
  allow_unverified_ssl = var.insecure_login
  logging              = true
  logging_file         = "cse_cluster_creation.log"
}

locals {
  capvcd_yaml_rendered = templatefile("./cluster-template-v1.22.9.yaml", {
    CLUSTER_NAME     = var.k8s_cluster_name
    TARGET_NAMESPACE = "${var.k8s_cluster_name}-ns"

    VCD_SITE                     = var.vcd_url
    VCD_ORGANIZATION             = var.cluster_organization
    VCD_ORGANIZATION_VDC         = var.cluster_vdc
    VCD_ORGANIZATION_VDC_NETWORK = var.cluster_routed_network

    VCD_USERNAME_B64      = base64encode(var.cluster_author_user)
    VCD_PASSWORD_B64      = "" # We use an API token, which is recommended
    VCD_REFRESH_TOKEN_B64 = base64encode(var.cluster_author_api_token)
    SSH_PUBLIC_KEY        = ""

    CONTROL_PLANE_MACHINE_COUNT        = 1
    VCD_CONTROL_PLANE_SIZING_POLICY    = var.cluster_sizing_policy
    VCD_CONTROL_PLANE_PLACEMENT_POLICY = "\"\"" # Not using any placement policy, but it is required in the template
    VCD_CONTROL_PLANE_STORAGE_PROFILE  = ""

    WORKER_MACHINE_COUNT        = 1
    VCD_WORKER_SIZING_POLICY    = var.cluster_sizing_policy
    VCD_WORKER_PLACEMENT_POLICY = "\"\"" # Not using any placement policy, but it is required in the template
    VCD_WORKER_STORAGE_PROFILE  = ""

    DISK_SIZE         = "20Gi"
    VCD_CATALOG       = var.tkgm_catalog
    VCD_TEMPLATE_NAME = var.tkgm_ova

    POD_CIDR     = "100.96.0.0/11"
    SERVICE_CIDR = "100.64.0.0/13"
  })
}

data "vcd_rde_type" "capvcdcluster_type" {
  vendor  = "vmware"
  nss     = "capvcdCluster"
  version = "1.0.0"
}

resource "vcd_rde" "k8s_cluster_instance" {
  name               = var.k8s_cluster_name
  rde_type_id        = data.vcd_rde_type.capvcdcluster_type.id # This must reference the CAPVCD RDE Type
  resolve            = false                                   # MUST be false as it is resolved by CSE Server
  resolve_on_removal = true                                    # MUST be true as it won't be resolved by Terraform

  # Read the RDE template present in this repository
  input_entity = templatefile("../entities/tkgmcluster-template.json", {
    vcd_url = var.vcd_url
    name    = var.k8s_cluster_name
    org     = var.cluster_organization
    vdc     = var.cluster_vdc

    capi_yaml = replace(replace(local.capvcd_yaml_rendered, "\n", "\\n"), "\"", "\\\"")

    delete                = false # Make this true to delete the cluster
    force_delete          = false # Make this true to forcefully delete the cluster
    auto_repair_on_errors = true  # Change this to false to troubleshoot possible issues
  })
}

output "computed_k8s_cluster_id" {
  value = vcd_rde.k8s_cluster_instance.id
}

output "computed_k8s_cluster_capvcdyaml" {
  value = jsondecode(vcd_rde.k8s_cluster_instance.computed_entity)["spec"]["capiYaml"]
}

# output "kubeconfig" {
#   value = jsondecode(vcd_rde.k8s_cluster_instance.computed_entity)["status"]["capvcd"]["private"]["kubeConfig"]
# }