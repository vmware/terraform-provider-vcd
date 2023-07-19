# ------------------------------------------------------------------------------------------------------------
# CSE 4.0 installation, step 1:
#
# * Please read the guide present at https://registry.terraform.io/providers/vmware/vcd/latest/docs/guides/container_service_extension_4_0_install
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

# VCD Provider configuration. It must be at least v3.10.0 and configured with a System administrator account.
terraform {
  required_providers {
    vcd = {
      source  = "vmware/vcd"
      version = ">= 3.10"
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

# This is the interface required to create the "VCDKEConfig" Runtime Defined Entity Type.
resource "vcd_rde_interface" "vcdkeconfig_interface" {
  vendor  = "vmware"
  nss     = "VCDKEConfig"
  version = "1.0.0"
  name    = "VCDKEConfig"
}

# This resource will manage the "VCDKEConfig" RDE Type required to instantiate the CSE Server configuration.
# The schema URL points to the JSON schema hosted in the terraform-provider-vcd repository.
resource "vcd_rde_type" "vcdkeconfig_type" {
  vendor        = "vmware"
  nss           = "VCDKEConfig"
  version       = "1.0.0"
  name          = "VCD-KE RDE Schema"
  schema_url    = "https://raw.githubusercontent.com/vmware/terraform-provider-vcd/main/examples/container-service-extension-4.0/schemas/vcdkeconfig-type-schema.json"
  interface_ids = [vcd_rde_interface.vcdkeconfig_interface.id]
}

# This RDE Interface exists in VCD, so it must be fetched with a RDE Interface data source. This RDE Interface is used to be
# able to create the "capvcdCluster" RDE Type.
data "vcd_rde_interface" "kubernetes_interface" {
  vendor  = "vmware"
  nss     = "k8s"
  version = "1.0.0"
}

# This RDE Interface will create the "capvcdCluster" RDE Type required to create Kubernetes clusters.
# The schema URL points to the JSON schema hosted in the terraform-provider-vcd repository.
resource "vcd_rde_type" "capvcdcluster_type" {
  vendor        = "vmware"
  nss           = "capvcdCluster"
  version       = var.capvcd_rde_version
  name          = "CAPVCD Cluster"
  schema_url    = "https://raw.githubusercontent.com/vmware/terraform-provider-vcd/main/examples/container-service-extension-4.0/schemas/capvcd-type-schema.json"
  interface_ids = [data.vcd_rde_interface.kubernetes_interface.id]
}

# This role is having only the minimum set of rights required for the CSE Server to function.
# It is created in the "System" provider organization scope.
resource "vcd_role" "cse_admin_role" {
  org         = var.administrator_org
  name        = "CSE Admin Role"
  description = "Used for administrative purposes"
  rights = [
    "API Tokens: Manage",
    "${vcd_rde_type.vcdkeconfig_type.vendor}:${vcd_rde_type.vcdkeconfig_type.nss}: Administrator Full access",
    "${vcd_rde_type.vcdkeconfig_type.vendor}:${vcd_rde_type.vcdkeconfig_type.nss}: Administrator View",
    "${vcd_rde_type.vcdkeconfig_type.vendor}:${vcd_rde_type.vcdkeconfig_type.nss}: Full Access",
    "${vcd_rde_type.vcdkeconfig_type.vendor}:${vcd_rde_type.vcdkeconfig_type.nss}: Modify",
    "${vcd_rde_type.vcdkeconfig_type.vendor}:${vcd_rde_type.vcdkeconfig_type.nss}: View",
    "${vcd_rde_type.capvcdcluster_type.vendor}:${vcd_rde_type.capvcdcluster_type.nss}: Administrator Full access",
    "${vcd_rde_type.capvcdcluster_type.vendor}:${vcd_rde_type.capvcdcluster_type.nss}: Administrator View",
    "${vcd_rde_type.capvcdcluster_type.vendor}:${vcd_rde_type.capvcdcluster_type.nss}: Full Access",
    "${vcd_rde_type.capvcdcluster_type.vendor}:${vcd_rde_type.capvcdcluster_type.nss}: Modify",
    "${vcd_rde_type.capvcdcluster_type.vendor}:${vcd_rde_type.capvcdcluster_type.nss}: View"
  ]
}

# This will allow to have a user with a limited set of rights that can access the Provider area of VCD.
# This user will be used by the CSE Server, with an API token that must be created afterwards.
resource "vcd_org_user" "cse_admin" {
  org      = var.administrator_org
  name     = var.cse_admin_username
  password = var.cse_admin_password
  role     = vcd_role.cse_admin_role.name
}

# This will output the username that you need to create an API token for.
output "ask_to_create_api_token" {
  value = "Please go to ${var.vcd_url}/provider/administration/settings/user-preferences, logged in as '${vcd_org_user.cse_admin.name}' and create an API token, as it will be required for step 2"
}
