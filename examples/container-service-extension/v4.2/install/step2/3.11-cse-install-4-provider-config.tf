# ------------------------------------------------------------------------------------------------------------
# CSE 4.2 installation, step 2:
#
# * Please read the guide at https://registry.terraform.io/providers/vmware/vcd/latest/docs/guides/container_service_extension_4_x_install
#   before applying this configuration.
#
# * The installation process is split into two steps as the first one creates a CSE admin user that needs to be
#   used in a "provider" block in the second one.
#
# * Rename "terraform.tfvars.example" to "terraform.tfvars" and adapt the values to your needs.
#   Other than that, this snippet should be applied as it is.
# ------------------------------------------------------------------------------------------------------------

# VCD Provider configuration. It must be v3.14.1 and configured with a System administrator account.
terraform {
  required_providers {
    vcd = {
      source  = "vmware/vcd"
      version = "3.14.1"
    }
    time = {
      source  = "hashicorp/time"
      version = "0.9.1"
    }
    local = {
      source  = "hashicorp/local"
      version = "2.4.0"
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
  logging_file         = "cse_install_step2.log"
}

# Minimum supported version for CSE
data "vcd_version" "cse_minimum_supported" {
  condition         = ">= 10.4.2"
  fail_if_not_match = true
}
