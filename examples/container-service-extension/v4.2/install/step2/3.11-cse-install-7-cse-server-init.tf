# ------------------------------------------------------------------------------------------------------------
# CSE 4.2 installation:
#
# * Please read the guide at https://registry.terraform.io/providers/vmware/vcd/latest/docs/guides/container_service_extension_4_x_install
#   before applying this configuration.
#
# * Rename "terraform.tfvars.example" to "terraform.tfvars" and adapt the values to your needs.
#   Other than that, this snippet should be applied as it is.
#   You can check the comments on each resource/data source for more help and context.
# ------------------------------------------------------------------------------------------------------------

# Log in to VCD with the cse_admin username created above. This will be used to provision
# an API token that must be consumed by the CSE Server.
# This should not be changed.
provider "vcd" {
  alias                = "cse_admin"
  url                  = "${var.vcd_url}/api"
  user                 = var.cse_admin_username
  password             = var.cse_admin_password
  auth_type            = "integrated"
  org                  = var.administrator_org
  allow_unverified_ssl = var.insecure_login
  logging              = true
  logging_file         = "cse_install_cse_admin.log"
}

# Generates an API token for the CSE Admin user, that will be used to instantiate the CSE Server.
# This should not be changed.
resource "vcd_api_token" "cse_admin_token" {
  provider         = vcd.cse_admin
  name             = "CSE Admin API Token"
  file_name        = var.cse_admin_api_token_file
  allow_token_file = true
}

data "local_file" "cse_admin_token_file" {
  filename = vcd_api_token.cse_admin_token.file_name
}

# This is the CSE Server vApp
resource "vcd_vapp" "cse_server_vapp" {
  org  = vcd_org.solutions_organization.name
  vdc  = vcd_org_vdc.solutions_vdc.name
  name = "CSE Server vApp"

  lease {
    runtime_lease_in_sec = 0
    storage_lease_in_sec = 0
  }
}

# The CSE Server vApp network that will consume an existing routed network from
# the solutions organization.
resource "vcd_vapp_org_network" "cse_server_network" {
  org = vcd_org.solutions_organization.name
  vdc = vcd_org_vdc.solutions_vdc.name

  vapp_name        = vcd_vapp.cse_server_vapp.name
  org_network_name = vcd_network_routed_v2.solutions_routed_network.name

  reboot_vapp_on_removal = true
}

# The CSE Server VM. It requires guest properties to be introduced for it to work
# properly. You can troubleshoot it by checking the cse.log file.
resource "vcd_vapp_vm" "cse_server_vm" {
  org = vcd_org.solutions_organization.name
  vdc = vcd_org_vdc.solutions_vdc.name

  vapp_name = vcd_vapp.cse_server_vapp.name
  name      = "CSE Server VM"

  vapp_template_id = vcd_catalog_vapp_template.cse_ova.id

  network {
    type               = "org"
    name               = vcd_vapp_org_network.cse_server_network.org_network_name
    ip_allocation_mode = "POOL"
  }

  guest_properties = {

    # VCD host
    "cse.vcdHost" = var.vcd_url

    # CSE Server org
    "cse.vAppOrg" = vcd_org.solutions_organization.name

    # CSE admin account's Access Token
    "cse.vcdRefreshToken" = jsondecode(data.local_file.cse_admin_token_file.content)["refresh_token"]

    # CSE admin account's username
    "cse.vcdUsername" = var.cse_admin_username

    # CSE admin account's org
    "cse.userOrg" = vcd_org.solutions_organization.name
  }

  customization {
    force                      = false
    enabled                    = true
    allow_local_admin_password = true
    auto_generate_password     = true
  }
}
