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
  user                 = var.cse_admin_username
  password             = var.cse_admin_password
  auth_type            = "integrated"
  sysorg               = var.administrator_org
  org                  = var.administrator_org
  allow_unverified_ssl = var.insecure_login
  logging              = true
  logging_file         = "cse_cluster_creation.log"
}

data "vcd_org" "cluster_org" {
  name = var.cluster_organization
}

data "vcd_org_vdc" "cluster_org" {
  org = data.vcd_org.cluster_org
  name = var.cluster_vdc
}

locals {
  capvcd_yaml_rendered = templatefile("./cluster-template-v1.22.9.yaml", {
    CLUSTER_NAME     = var.k8s_cluster_name
    TARGET_NAMESPACE = "${var.k8s_cluster_name}-ns"

    VCD_SITE                     = var.vcd_url
    VCD_ORGANIZATION             = data.vcd_org.cluster_org.name
    VCD_ORGANIZATION_VDC         = vcd_org_vdc.cluster_vdc.name
    VCD_ORGANIZATION_VDC_NETWORK = vcd_network_routed_v2.cluster_routed_network.name

    VCD_USERNAME_B64      = base64encode(var.k8s_cluster_user)
    VCD_PASSWORD_B64      = "" # We use an API token, which is recommended
    VCD_REFRESH_TOKEN_B64 = base64encode(var.k8s_cluster_api_token)
    SSH_PUBLIC_KEY        = ""

    CONTROL_PLANE_MACHINE_COUNT        = 1
    VCD_CONTROL_PLANE_SIZING_POLICY    = vcd_vm_sizing_policy.tkg_m.name
    VCD_CONTROL_PLANE_PLACEMENT_POLICY = ""
    VCD_CONTROL_PLANE_STORAGE_PROFILE  = ""

    WORKER_MACHINE_COUNT        = 1
    VCD_WORKER_SIZING_POLICY    = vcd_vm_sizing_policy.tkg_m.name
    VCD_WORKER_PLACEMENT_POLICY = ""
    VCD_WORKER_STORAGE_PROFILE  = ""

    DISK_SIZE         = "20Gi"
    VCD_CATALOG       = vcd_catalog.cse_catalog.name
    VCD_TEMPLATE_NAME = replace(var.tkgm_ova_name, ".ova", "")

    POD_CIDR     = "100.96.0.0/11"
    SERVICE_CIDR = "100.64.0.0/13"
  })
}

# Wait until CSE VM is fully initialized
resource "time_sleep" "wait_120_seconds" {
  depends_on = [
    vcd_vapp_vm.cse_appliance_vm, vcd_catalog_vapp_template.tkgm_ova
  ]
  create_duration = "120s"
}

resource "vcd_rde" "k8s_cluster_instance" {
  org                = "cluster_org"
  name               = "my-cluster"
  rde_type_id        = vcd_rde_type.capvcdcluster_type.id # This must reference the CAPVCD RDE Type
  resolve            = false                              # MUST be false as it is resolved by CSE Server
  resolve_on_removal = true                               # MUST be true as it won't be resolved by Terraform

  # Read the RDE template present in this repository
  input_entity = templatefile("${path.module}/entities/tkgmcluster-template.json", {
    vcd_url = replace(var.vcd_api_endpoint, "/api", "")
    name    = var.k8s_cluster_name
    org     = vcd_org.cluster_organization.name
    vdc     = vcd_org_vdc.cluster_vdc.name

    capi_yaml = replace(replace(local.capvcd_yaml_rendered, "\n", "\\n"), "\"", "\\\"")

    delete                = false # Make this true to delete the cluster
    force_delete          = false # Make this true to forcefully delete the cluster
    auto_repair_on_errors = true  # Change this to false to troubleshoot possible issues
  })

  depends_on = [
    time_sleep.wait_120_seconds
  ]
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