# ------------------------------------------------------------------------------------------------------------
# WARNING: This CSE installation method is deprecated in favor of CSE v4.x. Please have a look at
#          https://registry.terraform.io/providers/vmware/vcd/latest/docs/guides/container_service_extension_4_x_install
# ------------------------------------------------------------------------------------------------------------

# ------------------------------------------------------------------------------------------------------------
# CSE 3.1.x installation example HCL:
#
# * This HCL depends on 'vcd' and 'cse' CLI. Install and configure these following the instructions on
#   http://vmware.github.io/vcd-cli/install.html
#   https://vmware.github.io/container-service-extension/cse3_0/INSTALLATION.html#getting_cse
#
# * This example allows to create TKGm clusters using VCD UI once applied. Read the guide present in
#   https://registry.terraform.io/providers/vmware/vcd/latest/docs/guides/container_service_extension_3_1_x for more
#   information.
#
# * Some resources and data sources from this HCL are run as System administrator, as it involves creating provider
#   elements such as Organizations, VDCs or Provider Gateways. CSE items are created by the CSE service account defined
#   below.
#
# * Please customize the values present in this file to your needs. Also check `terraform.tfvars.example`
#   for customisation.
# ------------------------------------------------------------------------------------------------------------

terraform {
  required_providers {
    vcd = {
      source  = "vmware/vcd"
      version = ">= 3.8.0"
    }
  }
}

# Create a service role at provider level with the minimum rights required to install and manage CSE items.
# This role will be used by the CSE Service Account defined below.

resource "vcd_role" "cse-service-role" {
  name        = "CSE Service Role"
  description = "CSE Service Role has all the rights necessary for CSE to operate"

  rights = [
    "Access Control List: View",
    "Access Control List: Manage",
    "AMQP Settings: View",
    "Catalog: Add vApp from My Cloud",
    "Catalog: Create / Delete a Catalog",
    "Catalog: Edit Properties",
    "Catalog: Publish",
    "Catalog: Sharing",
    "Catalog: View ACL",
    "Catalog: View Private and Shared Catalogs",
    "Catalog: View Published Catalogs",
    "Content Library System Settings: View",
    "Custom entity: Create custom entity definitions",
    "Custom entity: Delete custom entity definitions",
    "Custom entity: Edit custom entity definitions",
    "Custom entity: View custom entity definitions",
    "Extension Services: View",
    "Extensions: View",
    "External Service: Manage",
    "External Service: View",
    "General: View Error Details",
    "Group / User: View",
    "Host: View",
    "Kerberos Settings: View",
    "Organization Network: View",
    "Organization vDC Compute Policy: Admin View",
    "Organization vDC Compute Policy: Manage",
    "Organization vDC Compute Policy: View",
    "Organization vDC Kubernetes Policy: Edit",
    "Organization vDC Network: Edit Properties",
    "Organization vDC Network: View Properties",
    "Organization vDC: Extended Edit",
    "Organization vDC: Extended View",
    "Organization vDC: View",
    "Organization: Perform Administrator Queries",
    "Organization: View",
    "Provider Network: View",
    "Provider vDC Compute Policy: Manage",
    "Provider vDC Compute Policy: View",
    "Provider vDC: View",
    "Right: Manage",
    "Right: View",
    "Rights Bundle: View",
    "Rights Bundle: Edit",
    "Role: Create, Edit, Delete, or Copy",
    "Service Configuration: Manage",
    "Service Configuration: View",
    "System Settings: View",
    "Task: Resume, Abort, or Fail",
    "Task: Update",
    "Task: View Tasks",
    "Token: Manage",
    "UI Plugins: Define, Upload, Modify, Delete, Associate or Disassociate",
    "UI Plugins: View",
    "vApp Template / Media: Copy",
    "vApp Template / Media: Create / Upload",
    "vApp Template / Media: Edit",
    "vApp Template / Media: View",
    "vApp Template: Checkout",
    "vApp Template: Import",
    "vApp: Allow All Extra Config",
    "vApp: Allow Ethernet Coalescing Extra Config",
    "vApp: Allow Latency Extra Config",
    "vApp: Allow Matching Extra Config",
    "vApp: Allow NUMA Node Affinity Extra Config",
    "vApp: Create / Reconfigure",
    "vApp: Delete",
    "vApp: Edit Properties",
    "vApp: Edit VM CPU and Memory reservation settings in all VDC types",
    "vApp: Edit VM CPU",
    "vApp: Edit VM Compute Policy",
    "vApp: Edit VM Hard Disk",
    "vApp: Edit VM Memory",
    "vApp: Edit VM Network",
    "vApp: Edit VM Properties",
    "vApp: Manage VM Password Settings",
    "vApp: Power Operations",
    "vApp: Shadow VM View",
    "vApp: Upload",
    "vApp: Use Console",
    "vApp: VM Boot Options",
    "vApp: VM Check Compliance",
    "vApp: VM Migrate, Force Undeploy, Relocate, Consolidate",
    "vApp: View VM and VM's Disks Encryption Status",
    "vApp: View VM metrics",
    "vCenter: View",
    "vSphere Server: View",
    "vmware:tkgcluster: Administrator Full access",
    "vmware:tkgcluster: Administrator View",
    "vmware:tkgcluster: Full Access",
    "vmware:tkgcluster: Modify",
    "vmware:tkgcluster: View",
    # These rights are only needed after CSE is completely installed
    # "cse:nativeCluster: Administrator Full access",
    # "cse:nativeCluster: Administrator View",
    # "cse:nativeCluster: Full Access",
    # "cse:nativeCluster: Modify",
    # "cse:nativeCluster: View"
  ]
}

# Create the CSE Service Account to manage CSE installation. We use it for the operations related to CSE installation and management.

resource "vcd_org_user" "cse-service-account" {
  name     = var.service-account-user
  password = var.service-account-password
  role     = vcd_role.cse-service-role.name
}

# This provider definition uses the created CSE Service Account to perform all CSE related operations with the alias
# "cse-service-account" (more info about aliases: https://www.terraform.io/language/providers/configuration#alias-multiple-provider-configurations)
# This allows to both use the System Administrator for operations that require full access, and the service account for
# CSE.

provider "vcd" {
  alias                = "cse-service-account"
  user                 = vcd_org_user.cse-service-account.name
  password             = vcd_org_user.cse-service-account.password
  auth_type            = "integrated"
  url                  = var.vcd-url
  org                  = var.admin-org
  logging_file         = "cse-install.log"
  allow_unverified_ssl = true
}

# The default provider configuration uses the System Administrator to create the service role, the CSE Service Account
# and the minimum objects required (Organization, VDC, ALB, etc).

provider "vcd" {
  user                 = var.admin-user
  password             = var.admin-password
  auth_type            = "integrated"
  url                  = var.vcd-url
  org                  = var.admin-org
  logging_file         = "cse-install-administrator.log"
  allow_unverified_ssl = true
}

# If you have already an Organization, remove the `vcd_org` resource from this HCL file and either
# configure it in the provider settings or fetch it with the `vcd_org` data source, as follows:
#
#   data "vcd_org" "cse_org" {
#     name = var.organization
#   }
#
# If you remove this resource, you need to adapt all `org = vcd_org.cse_org.name` occurrences to your needs.

resource "vcd_org" "cse_org" {
  name             = var.org-name
  full_name        = var.org-name
  is_enabled       = true
  delete_force     = true
  delete_recursive = true
}

# If you have already a VDC, remove the `vcd_org_vdc` resource from this HCL file and
# fetch it with the `vcd_org_vdc` data source, as follows:
#
#   data "vcd_org_vdc" "cse_vdc" {
#     org  = data.vcd_org.cse_org
#     name = var.vdc
#   }
# If you remove this resource, you need to change `owner_id`/`vdc` in the affected resources.

resource "vcd_org_vdc" "cse_vdc" {
  name = var.vdc-name
  org  = vcd_org.cse_org.name # Change this reference if you used a data source to fetch an already existent Org.

  allocation_model  = var.vdc-allocation
  provider_vdc_name = var.vdc-provider
  network_pool_name = var.vdc-netpool
  network_quota     = 50

  compute_capacity {
    cpu {
      limit = 0
    }

    memory {
      limit = 0
    }
  }

  storage_profile {
    name    = var.storage-profile
    enabled = true
    limit   = 0
    default = true
  }

  enabled                  = true
  enable_thin_provisioning = true
  enable_fast_provisioning = true
  delete_force             = true
  delete_recursive         = true
}

# Create a Provider Gateway connected to the outside world network. This will be used to download software
# for the Kubernetes nodes and access the cluster.

data "vcd_nsxt_manager" "main" {
  name = var.tier0-manager
}

data "vcd_nsxt_tier0_router" "router" {
  name            = var.tier0-router
  nsxt_manager_id = data.vcd_nsxt_manager.main.id
}

resource "vcd_external_network_v2" "cse_external_network_nsxt" {
  name        = "nsxt-extnet-cse"
  description = "NSX-T backed network for k8s clusters"

  nsxt_network {
    nsxt_manager_id      = data.vcd_nsxt_manager.main.id
    nsxt_tier0_router_id = data.vcd_nsxt_tier0_router.router.id
  }

  ip_scope {
    gateway       = var.tier0-gateway-ip
    prefix_length = var.tier0-gateway-prefix

    dynamic "static_ip_pool" {
      for_each = var.tier0-gateway-ip-ranges
      iterator = ip
      content {
        start_address = ip.value[0]
        end_address   = ip.value[1]
      }
    }
  }
}

# Create an Edge Gateway that will be used by the cluster as the main router.

resource "vcd_nsxt_edgegateway" "cse_egw" {
  org      = vcd_org.cse_org.name   # Change this reference if you used a data source to fetch an already existent Org.
  owner_id = vcd_org_vdc.cse_vdc.id # Change this reference if you used a data source to fetch an already existent VDC.

  name                = "cse-egw"
  description         = "Edge gateway for CSE to route traffic in the Kubernetes cluster"
  external_network_id = vcd_external_network_v2.cse_external_network_nsxt.id

  subnet {
    gateway       = var.edge-gateway-ip
    prefix_length = var.edge-gateway-prefix
    primary_ip    = var.edge-gateway-ip-ranges[0][0] # The first IP provided will be assigned as gateway IP

    dynamic "allocated_ips" {
      for_each = var.edge-gateway-ip-ranges
      iterator = ip
      content {
        start_address = ip.value[0]
        end_address   = ip.value[1]
      }
    }
  }

  depends_on = [vcd_org_vdc.cse_vdc]
}

# Routed network for the Kubernetes cluster.

resource "vcd_network_routed_v2" "cse_routed" {
  org         = vcd_org.cse_org.name
  name        = "cse_routed_net"
  description = "My routed Org VDC network backed by NSX-T"

  edge_gateway_id = vcd_nsxt_edgegateway.cse_egw.id

  gateway       = var.routed-gateway
  prefix_length = var.routed-prefix

  static_ip_pool {
    start_address = var.routed-ip-range[0]
    end_address   = var.routed-ip-range[1]
  }

  dns1 = var.routed-dns[0]
  dns2 = var.routed-dns[1]
}

# NAT rule to map traffic to internal network IPs.

resource "vcd_nsxt_nat_rule" "snat" {
  org             = vcd_org.cse_org.name # Change this reference if you used a data source to fetch an already existent Org.
  edge_gateway_id = vcd_nsxt_edgegateway.cse_egw.id

  name        = "SNAT rule"
  rule_type   = "SNAT"
  description = "description"

  external_address = var.snat-external-ip
  internal_address = format("%s.%s.%s.0/%s", split(".", var.routed-gateway)[0], split(".", var.routed-gateway)[1], split(".", var.routed-gateway)[2], var.routed-prefix)
  logging          = true
}

# Cluster requires network traffic is open, to download required dependencies to create nodes. Adapt this firewall
# rule to your organization security requirements, as this is just an example.

resource "vcd_nsxt_firewall" "firewall" {
  org             = vcd_org.cse_org.name # Change this reference if you used a data source to fetch an already existent Org.
  edge_gateway_id = vcd_nsxt_edgegateway.cse_egw.id

  rule {
    action      = "ALLOW"
    name        = "allow all IPv4 traffic for CSE clusters"
    direction   = "IN_OUT"
    ip_protocol = "IPV4"
  }
}

# Catalog to upload the TKGm OVAs.

data "vcd_storage_profile" "cse_sp" {
  org  = vcd_org.cse_org.name     # Change this reference if you used a data source to fetch an already existent Org.
  vdc  = vcd_org_vdc.cse_vdc.name # Change this reference if you used a data source to fetch an already existent VDC.
  name = var.storage-profile

  depends_on = [vcd_org.cse_org, vcd_org_vdc.cse_vdc]
}

resource "vcd_catalog" "cat-cse" {
  org         = vcd_org.cse_org.name # Change this reference if you used a data source to fetch an already existent Org.
  name        = "cat-cse"
  description = "CSE catalog to store TKGm OVA files"

  storage_profile_id = data.vcd_storage_profile.cse_sp.id

  delete_force     = true
  delete_recursive = true
}

# TKGm OVA upload. The `catalog_item_metadata` is required for CSE to detect the OVAs.

resource "vcd_catalog_item" "tkgm_ova" {
  provider = vcd.cse-service-account # Using CSE Service Account for this resource

  org     = vcd_org.cse_org.name # Change this reference if you used a data source to fetch an already existent Org.
  catalog = vcd_catalog.cat-cse.name

  name              = replace(var.tkgm-ova-name, ".ova", "")
  description       = replace(var.tkgm-ova-name, ".ova", "")
  ova_path          = format("%s/%s", var.tkgm-ova-folder, var.tkgm-ova-name)
  upload_piece_size = 100

  metadata_entry {
    key         = "kind"
    value       = "TKGm" # This value is always the same
    type        = "MetadataStringValue"
    user_access = "READWRITE"
    is_system   = false
  }

  metadata_entry {
    key         = "kubernetes"
    value       = "TKGm" # This value is always the same
    type        = "MetadataStringValue"
    user_access = "READWRITE"
    is_system   = false
  }

  metadata_entry {
    key         = "kubernetes_version"
    value       = split("-", var.tkgm-ova-name)[3] # The version comes in the OVA name downloaded from Customer Connect
    type        = "MetadataStringValue"
    user_access = "READWRITE"
    is_system   = false
  }

  metadata_entry {
    key         = "name"
    value       = replace(var.tkgm-ova-name, ".ova", "") # The name as it was in the OVA downloaded from Customer Connect
    type        = "MetadataStringValue"
    user_access = "READWRITE"
    is_system   = false
  }

  metadata_entry {
    key         = "os"
    value       = split("-", var.tkgm-ova-name)[0] # The OS comes in the OVA name downloaded from Customer Connect
    type        = "MetadataStringValue"
    user_access = "READWRITE"
    is_system   = false
  }

  metadata_entry {
    key         = "revision"
    value       = "1" # This value is always the same
    type        = "MetadataStringValue"
    user_access = "READWRITE"
    is_system   = false
  }

}

# AVI configuration for Kubernetes services, this allows the cluster to create Kubernetes services of type Load Balancer.

data "vcd_nsxt_alb_controller" "cse_alb_controller" {
  name = var.avi-controller-name
}

data "vcd_nsxt_alb_importable_cloud" "cse_importable_cloud" {
  name          = var.avi-importable-cloud
  controller_id = data.vcd_nsxt_alb_controller.cse_alb_controller.id
}

resource "vcd_nsxt_alb_cloud" "cse_alb_cloud" {
  name        = "cse_alb_cloud"
  description = "cse alb cloud"

  controller_id       = data.vcd_nsxt_alb_controller.cse_alb_controller.id
  importable_cloud_id = data.vcd_nsxt_alb_importable_cloud.cse_importable_cloud.id
  network_pool_id     = data.vcd_nsxt_alb_importable_cloud.cse_importable_cloud.network_pool_id
}

resource "vcd_nsxt_alb_service_engine_group" "cse_alb_seg" {
  name                                 = "cse_alb_seg"
  alb_cloud_id                         = vcd_nsxt_alb_cloud.cse_alb_cloud.id
  importable_service_engine_group_name = "Default-Group"
  reservation_model                    = "DEDICATED"
}

resource "vcd_nsxt_alb_settings" "cse_alb_settings" {
  org             = vcd_org.cse_org.name # Change this reference if you used a data source to fetch an already existent Org.
  edge_gateway_id = vcd_nsxt_edgegateway.cse_egw.id
  is_active       = true

  # This dependency is required to make sure that provider part of operations is done
  depends_on = [vcd_nsxt_alb_service_engine_group.cse_alb_seg]
}

resource "vcd_nsxt_alb_edgegateway_service_engine_group" "assignment" {
  org                     = vcd_org.cse_org.name # Change this reference if you used a data source to fetch an already existent Org.
  edge_gateway_id         = vcd_nsxt_alb_settings.cse_alb_settings.edge_gateway_id
  service_engine_group_id = vcd_nsxt_alb_service_engine_group.cse_alb_seg.id
}

resource "vcd_nsxt_alb_pool" "cse_alb_pool" {
  org             = vcd_org.cse_org.name # Change this reference if you used a data source to fetch an already existent Org.
  edge_gateway_id = vcd_nsxt_alb_settings.cse_alb_settings.edge_gateway_id
  name            = "cse-avi-pool"
}

resource "vcd_nsxt_alb_virtual_service" "cse-virtual-service" {
  org             = vcd_org.cse_org.name # Change this reference if you used a data source to fetch an already existent Org.
  edge_gateway_id = vcd_nsxt_alb_settings.cse_alb_settings.edge_gateway_id
  name            = "cse-virtual-service"

  pool_id                  = vcd_nsxt_alb_pool.cse_alb_pool.id
  service_engine_group_id  = vcd_nsxt_alb_edgegateway_service_engine_group.assignment.service_engine_group_id
  virtual_ip_address       = var.avi-virtual-service-ip
  application_profile_type = "HTTP"
  service_port {
    start_port = 80
    type       = "TCP_PROXY"
  }
}

# CSE installation process. With this resource we fetch the `config.yaml.template` file present next to this example HCL, and
# fill the template variables with the ones generated here, such as VDC, Org, Catalog, etc; then we invoke the CSE install command
# with the given template.

resource "null_resource" "cse-install-script" {
  triggers = {
    # Trigger the installation only if one of the configuration values change.
    config_has_changed = join(",", [
      vcd_org_user.cse-service-account.name,
      vcd_org_user.cse-service-account.password,
      vcd_catalog.cat-cse.name,
      vcd_network_routed_v2.cse_routed.name,
      vcd_network_routed_v2.cse_routed.name,
      vcd_org.cse_org.name,
      vcd_org_vdc.cse_vdc.name,
      data.vcd_storage_profile.cse_sp.name
    ])
  }

  provisioner "local-exec" {
    on_failure = continue # Ignores failures to allow re-creating the whole HCL after a destroy, as cse doesn't have an uninstall option.
    command = format("printf '%s' > config.yaml && ./cse-install.sh", templatefile("${path.module}/config.yaml.template", {
      vcd_url         = replace(replace(var.vcd-url, "/api", ""), "/http.*\\/\\//", "")
      vcd_username    = vcd_org_user.cse-service-account.name # Using CSE Service Account
      vcd_password    = vcd_org_user.cse-service-account.password
      catalog         = vcd_catalog.cat-cse.name
      network         = vcd_network_routed_v2.cse_routed.name
      org             = vcd_org.cse_org.name     # Change this reference if you used a data source to fetch an already existent Org.
      vdc             = vcd_org_vdc.cse_vdc.name # Change this reference if you used a data source to fetch an already existent VDC.
      storage_profile = data.vcd_storage_profile.cse_sp.name
    }))
  }
}

# Here we create a new rights bundle for CSE, with the rights assigned already to the Default Rights Bundle (hence the
# data source) plus new ones coming from `cse install` command.

data "vcd_rights_bundle" "default-rb" {
  name = "Default Rights Bundle"
}

resource "vcd_rights_bundle" "cse-rb" {
  provider = vcd.cse-service-account # Using CSE Service Account for this resource

  name        = "CSE Rights Bundle"
  description = "Rights bundle to manage CSE"
  rights = setunion(data.vcd_rights_bundle.default-rb.rights, [
    "API Tokens: Manage",
    "Organization vDC Shared Named Disk: Create",
    "cse:nativeCluster: View",
    "cse:nativeCluster: Full Access",
    "cse:nativeCluster: Modify"
  ])
  publish_to_all_tenants = false
  tenants                = [vcd_org.cse_org.name]

  depends_on = [null_resource.cse-install-script]
}

# Here we fetch the rights bundle that `cse install` creates. As we can't update/destroy it, we simply clone
# into a new and published bundle.

data "vcd_rights_bundle" "cse-native-cluster-entl" {
  provider = vcd.cse-service-account # Using CSE Service Account for this data source

  name = "cse:nativeCluster Entitlement"

  depends_on = [null_resource.cse-install-script]
}

resource "vcd_rights_bundle" "published-cse-rights-bundle" {
  provider = vcd.cse-service-account # Using CSE Service Account for this resource

  name                   = "cse:nativeCluster Entitlement Published"
  description            = data.vcd_rights_bundle.cse-native-cluster-entl.description
  rights                 = data.vcd_rights_bundle.cse-native-cluster-entl.rights
  publish_to_all_tenants = false
  tenants                = [vcd_org.cse_org.name]
}

# Create a new role for CSE, with the new rights to create clusters and manage them.

data "vcd_role" "vapp_author" {
  org  = vcd_org.cse_org.name # Change this reference if you used a data source to fetch an already existent Org.
  name = "vApp Author"
}

resource "vcd_role" "cluster_author" {
  provider = vcd.cse-service-account # Using CSE Service Account for this resource

  org         = vcd_org.cse_org.name # Change this reference if you used a data source to fetch an already existent Org.
  name        = "Cluster Author"
  description = "Can read and create clusters"
  rights = setunion(data.vcd_role.vapp_author.rights, [
    "API Tokens: Manage",
    "Organization vDC Shared Named Disk: Create",
    "Organization vDC Gateway: View",
    "Organization vDC Gateway: View Load Balancer",
    "Organization vDC Gateway: Configure Load Balancer",
    "Organization vDC Gateway: View NAT",
    "Organization vDC Gateway: Configure NAT",
    "cse:nativeCluster: View",
    "cse:nativeCluster: Full Access",
    "cse:nativeCluster: Modify",
    "Certificate Library: View" # Implicit role needed
  ])

  depends_on = [vcd_rights_bundle.cse-rb]
}

output "finish-message-1" {
  value = "Publish Container UI Plugin for CSE plugin on VCD in 'Customize Portal' section to create clusters via UI"
}

output "finish-message-2" {
  value = "Run 'cse run -s -c config.yaml' to start CSE server and create clusters on VCD using published Container UI Plugin"
}

output "finish-message-3" {
  value = "If you need to execute 'terraform destroy', make sure all Kubernetes clusters are deleted first"
}
