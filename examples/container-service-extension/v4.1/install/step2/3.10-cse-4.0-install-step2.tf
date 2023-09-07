# ------------------------------------------------------------------------------------------------------------
# CSE 4.0 installation, step 2:
#
# * Please read the guide present at https://registry.terraform.io/providers/vmware/vcd/latest/docs/guides/container_service_extension_4_0_install
#   before applying this configuration.
#
# * Please apply "3.10-cse-4.0-install-step1.tf" first, located at
#   https://github.com/vmware/terraform-provider-vcd/tree/main/examples/container-service-extension-4.0/install/step1
#
# * Please review this HCL configuration before applying, to change the settings to the ones that fit best with your organization.
#   For example, network settings such as firewall rules, network subnets, VDC allocation modes, ALB feature set, etc should be
#   carefully reviewed.
#
# * Rename "terraform.tfvars.example" to "terraform.tfvars" and adapt the values to your needs.
#   You can check the comments on each resource/data source for more help and context.
# ------------------------------------------------------------------------------------------------------------

# VCD Provider configuration. It must be at least v3.10.0 and configured with a System administrator account.
# This is needed to build the minimum setup for CSE v4.0 to work, like Organizations, VDCs, Provider Gateways, etc.
terraform {
  required_providers {
    vcd = {
      source  = "vmware/vcd"
      version = ">= 3.10"
    }
    time = {
      source  = "hashicorp/time"
      version = ">= 0.9"
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

# The two resources below will create the two Organizations mentioned in the CSE documentation:
# https://docs.vmware.com/en/VMware-Cloud-Director-Container-Service-Extension/index.html

# The Solutions Organization will host the CSE Server and its intended to be used by CSE Administrators only.
# The TKGm clusters are NOT placed here. The attributes related to lease are set to unlimited, as the CSE
# Server should be always up and running in order to process requests.
resource "vcd_org" "solutions_organization" {
  name             = "solutions_org"
  full_name        = "Solutions Organization"
  is_enabled       = true
  delete_force     = true
  delete_recursive = true

  vapp_lease {
    maximum_runtime_lease_in_sec          = 0
    power_off_on_runtime_lease_expiration = false
    maximum_storage_lease_in_sec          = 0
    delete_on_storage_lease_expiration    = false
  }

  vapp_template_lease {
    maximum_storage_lease_in_sec       = 0
    delete_on_storage_lease_expiration = false
  }
}

# The Tenant Organization will host the TKGm clusters and its intended to be used by tenants.
# The TKGm clusters must be placed here. The attributes related to lease are set to unlimited, as the TKGm clusters vApps
# should not be powered off.
resource "vcd_org" "tenant_organization" {
  name             = "tenant_org"
  full_name        = "Tenant Organization"
  is_enabled       = true
  delete_force     = true
  delete_recursive = true

  vapp_lease {
    maximum_runtime_lease_in_sec          = 0
    power_off_on_runtime_lease_expiration = false
    maximum_storage_lease_in_sec          = 0
    delete_on_storage_lease_expiration    = false
  }

  vapp_template_lease {
    maximum_storage_lease_in_sec       = 0
    delete_on_storage_lease_expiration = false
  }
}

# The VM Sizing Policies defined below MUST be created as they are specified in this HCL. These are the default
# policies required by CSE to create TKGm clusters, hence nothing should be modified here.
resource "vcd_vm_sizing_policy" "tkg_xl" {
  name        = "TKG extra-large"
  description = "Extra-large VM sizing policy for a Kubernetes cluster node (8 CPU, 32GB memory)"
  cpu {
    count = 8
  }
  memory {
    size_in_mb = "32768"
  }
}

resource "vcd_vm_sizing_policy" "tkg_l" {
  name        = "TKG large"
  description = "Large VM sizing policy for a Kubernetes cluster node (4 CPU, 16GB memory)"
  cpu {
    count = 4
  }
  memory {
    size_in_mb = "16384"
  }
}

resource "vcd_vm_sizing_policy" "tkg_m" {
  name        = "TKG medium"
  description = "Medium VM sizing policy for a Kubernetes cluster node (2 CPU, 8GB memory)"
  cpu {
    count = 2
  }
  memory {
    size_in_mb = "8192"
  }
}

resource "vcd_vm_sizing_policy" "tkg_s" {
  name        = "TKG small"
  description = "Small VM sizing policy for a Kubernetes cluster node (2 CPU, 4GB memory)"
  cpu {
    count = 2
  }
  memory {
    size_in_mb = "4048"
  }
}

# This section will create one VDC per organization. To create the VDCs we need to fetch some elements like
# Provider VDC, Edge Clusters, etc.
data "vcd_provider_vdc" "nsxt_pvdc" {
  name = var.provider_vdc_name
}

data "vcd_nsxt_edge_cluster" "nsxt_edgecluster" {
  org             = vcd_org.tenant_organization.name
  provider_vdc_id = data.vcd_provider_vdc.nsxt_pvdc.id
  name            = var.nsxt_edge_cluster_name
}

# The VDC that will host the Kubernetes clusters.
resource "vcd_org_vdc" "tenant_vdc" {
  name        = "tenant_vdc"
  description = "Tenant VDC"
  org         = vcd_org.tenant_organization.name

  allocation_model  = "AllocationVApp" # You can use other models.
  network_pool_name = var.network_pool_name
  provider_vdc_name = data.vcd_provider_vdc.nsxt_pvdc.name
  edge_cluster_id   = data.vcd_nsxt_edge_cluster.nsxt_edgecluster.id

  # You can tune these arguments to your fit your needs.
  network_quota = 50
  compute_capacity {
    cpu {
      allocated = 0
    }

    memory {
      allocated = 0
    }
  }

  # You can tune these arguments to your fit your needs.
  storage_profile {
    name    = "*"
    limit   = 0
    default = true
  }

  # You can tune these arguments to your fit your needs.
  enabled                  = true
  enable_thin_provisioning = true
  enable_fast_provisioning = true
  delete_force             = true
  delete_recursive         = true

  # Make sure you specify the required VM Sizing Policies managed by the resources specified above.
  default_compute_policy_id = vcd_vm_sizing_policy.tkg_s.id
  vm_sizing_policy_ids = [
    vcd_vm_sizing_policy.tkg_xl.id,
    vcd_vm_sizing_policy.tkg_l.id,
    vcd_vm_sizing_policy.tkg_m.id,
    vcd_vm_sizing_policy.tkg_s.id,
  ]
}

# The VDC that will host the CSE server and other provider-level items
resource "vcd_org_vdc" "solutions_vdc" {
  name        = "solutions_vdc"
  description = "Solutions VDC"
  org         = vcd_org.solutions_organization.name

  allocation_model  = "AllocationVApp" # You can use other models
  network_pool_name = var.network_pool_name
  provider_vdc_name = data.vcd_provider_vdc.nsxt_pvdc.name
  edge_cluster_id   = data.vcd_nsxt_edge_cluster.nsxt_edgecluster.id

  # You can tune these arguments to your fit your needs
  network_quota = 10
  compute_capacity {
    cpu {
      allocated = 0
    }

    memory {
      allocated = 0
    }
  }

  # You can tune these arguments to your fit your needs
  storage_profile {
    name    = "*"
    limit   = 0
    default = true
  }

  # You can tune these arguments to your fit your needs
  enabled                  = true
  enable_thin_provisioning = true
  enable_fast_provisioning = true
  delete_force             = true
  delete_recursive         = true
}

# In this section we create two Catalogs, one to host all CSE Server OVAs and another one to host TKGm OVAs.
# They are created in the Solutions organization and only the TKGm will be shared as read-only. This will guarantee
# that only CSE admins can manage OVAs.
resource "vcd_catalog" "cse_catalog" {
  org  = vcd_org.solutions_organization.name # References the Solutions Organization created previously
  name = "cse_catalog"

  delete_force     = "true"
  delete_recursive = "true"

  # In this example, everything is created from scratch, so it is needed to wait for the VDC to be available, so the
  # Catalog can be created.
  depends_on = [
    vcd_org_vdc.solutions_vdc
  ]
}

resource "vcd_catalog" "tkgm_catalog" {
  org  = vcd_org.solutions_organization.name # References the Solutions Organization
  name = "tkgm_catalog"

  delete_force     = "true"
  delete_recursive = "true"

  # In this example, everything is created from scratch, so it is needed to wait for the VDC to be available, so the
  # Catalog can be created.
  depends_on = [
    vcd_org_vdc.solutions_vdc
  ]
}

# We share the TKGm Catalog with the Tenant Organization created previously.
resource "vcd_catalog_access_control" "tkgm_catalog_ac" {
  org                  = vcd_org.solutions_organization.name # References the Solutions Organization created previously
  catalog_id           = vcd_catalog.tkgm_catalog.id
  shared_with_everyone = false
  shared_with {
    org_id       = vcd_org.tenant_organization.id # Shared with the Tenant Organization
    access_level = "ReadOnly"
  }
}

# We upload a minimum set of OVAs for CSE to work. Read the official documentation to check
# where to find the OVAs:
# https://docs.vmware.com/en/VMware-Cloud-Director-Container-Service-Extension/index.html
resource "vcd_catalog_vapp_template" "tkgm_ova" {
  org        = vcd_org.solutions_organization.name # References the Solutions Organization created previously
  catalog_id = vcd_catalog.tkgm_catalog.id         # References the TKGm Catalog created previously

  name        = replace(var.tkgm_ova_file, ".ova", "")
  description = replace(var.tkgm_ova_file, ".ova", "")
  ova_path    = format("%s/%s", var.tkgm_ova_folder, var.tkgm_ova_file)
}

resource "vcd_catalog_vapp_template" "cse_ova" {
  org        = vcd_org.solutions_organization.name # References the Solutions Organization created previously
  catalog_id = vcd_catalog.cse_catalog.id          # References the CSE Catalog created previously

  name        = replace(var.cse_ova_file, ".ova", "")
  description = replace(var.cse_ova_file, ".ova", "")
  ova_path    = format("%s/%s", var.cse_ova_folder, var.cse_ova_file)
}

# Fetch the RDE Type created in 3.10-cse-4.0-install-step1.tf. This is required to be able to create the following
# Rights Bundle.
data "vcd_rde_type" "existing_capvcdcluster_type" {
  vendor  = "vmware"
  nss     = "capvcdCluster"
  version = var.capvcd_rde_version
}

# This resource manages the Rights Bundle required by tenants to create and consume Kubernetes clusters.
resource "vcd_rights_bundle" "k8s_clusters_rights_bundle" {
  name        = "Kubernetes Clusters Rights Bundle"
  description = "Rights bundle with required rights for managing Kubernetes clusters"
  rights = [
    "API Tokens: Manage",
    "vApp: Allow All Extra Config",
    "Catalog: View Published Catalogs",
    "Organization vDC Shared Named Disk: Create",
    "Organization vDC Gateway: View",
    "Organization vDC Gateway: View NAT",
    "Organization vDC Gateway: Configure NAT",
    "Organization vDC Gateway: View Load Balancer",
    "Organization vDC Gateway: Configure Load Balancer",
    "${data.vcd_rde_type.existing_capvcdcluster_type.vendor}:${data.vcd_rde_type.existing_capvcdcluster_type.nss}: Administrator Full access",
    "${data.vcd_rde_type.existing_capvcdcluster_type.vendor}:${data.vcd_rde_type.existing_capvcdcluster_type.nss}: Full Access",
    "${data.vcd_rde_type.existing_capvcdcluster_type.vendor}:${data.vcd_rde_type.existing_capvcdcluster_type.nss}: Modify",
    "${data.vcd_rde_type.existing_capvcdcluster_type.vendor}:${data.vcd_rde_type.existing_capvcdcluster_type.nss}: View",
    "${data.vcd_rde_type.existing_capvcdcluster_type.vendor}:${data.vcd_rde_type.existing_capvcdcluster_type.nss}: Administrator View",
    "General: Administrator View",
    "Certificate Library: Manage",
    "Access All Organization VDCs",
    "Certificate Library: View",
    "Organization vDC Named Disk: Create",
    "Organization vDC Named Disk: Edit Properties",
    "Organization vDC Named Disk: View Properties",
    "vmware:tkgcluster: Full Access",
    "vmware:tkgcluster: Modify",
    "vmware:tkgcluster: View",
    "vmware:tkgcluster: Administrator View",
    "vmware:tkgcluster: Administrator Full access",
  ]
  publish_to_all_tenants = true # This needs to be published to all the Organizations
}

# With the Rights Bundle specified above, we need also a new Role for tenant users who want to create and manage
# Kubernetes clusters.
resource "vcd_global_role" "k8s_cluster_author" {
  name        = "Kubernetes Cluster Author"
  description = "Role to create Kubernetes clusters"
  rights = [
    "API Tokens: Manage",
    "Access All Organization VDCs",
    "Catalog: Add vApp from My Cloud",
    "Catalog: View Private and Shared Catalogs",
    "Catalog: View Published Catalogs",
    "Certificate Library: View",
    "Organization vDC Compute Policy: View",
    "Organization vDC Gateway: Configure Load Balancer",
    "Organization vDC Gateway: Configure NAT",
    "Organization vDC Gateway: View",
    "Organization vDC Gateway: View Load Balancer",
    "Organization vDC Gateway: View NAT",
    "Organization vDC Named Disk: Create",
    "Organization vDC Named Disk: Delete",
    "Organization vDC Named Disk: Edit Properties",
    "Organization vDC Named Disk: View Properties",
    "Organization vDC Network: View Properties",
    "Organization vDC Shared Named Disk: Create",
    "Organization vDC: VM-VM Affinity Edit",
    "Organization: View",
    "UI Plugins: View",
    "VAPP_VM_METADATA_TO_VCENTER",
    "vApp Template / Media: Copy",
    "vApp Template / Media: Edit",
    "vApp Template / Media: View",
    "vApp Template: Checkout",
    "vApp: Allow All Extra Config",
    "vApp: Copy",
    "vApp: Create / Reconfigure",
    "vApp: Delete",
    "vApp: Download",
    "vApp: Edit Properties",
    "vApp: Edit VM CPU",
    "vApp: Edit VM Hard Disk",
    "vApp: Edit VM Memory",
    "vApp: Edit VM Network",
    "vApp: Edit VM Properties",
    "vApp: Manage VM Password Settings",
    "vApp: Power Operations",
    "vApp: Sharing",
    "vApp: Snapshot Operations",
    "vApp: Upload",
    "vApp: Use Console",
    "vApp: VM Boot Options",
    "vApp: View ACL",
    "vApp: View VM metrics",
    "${data.vcd_rde_type.existing_capvcdcluster_type.vendor}:${data.vcd_rde_type.existing_capvcdcluster_type.nss}: Administrator Full access",
    "${data.vcd_rde_type.existing_capvcdcluster_type.vendor}:${data.vcd_rde_type.existing_capvcdcluster_type.nss}: Full Access",
    "${data.vcd_rde_type.existing_capvcdcluster_type.vendor}:${data.vcd_rde_type.existing_capvcdcluster_type.nss}: Modify",
    "${data.vcd_rde_type.existing_capvcdcluster_type.vendor}:${data.vcd_rde_type.existing_capvcdcluster_type.nss}: View",
    "${data.vcd_rde_type.existing_capvcdcluster_type.vendor}:${data.vcd_rde_type.existing_capvcdcluster_type.nss}: Administrator View",
    "vmware:tkgcluster: Full Access",
    "vmware:tkgcluster: Modify",
    "vmware:tkgcluster: View",
    "vmware:tkgcluster: Administrator View",
    "vmware:tkgcluster: Administrator Full access",
  ]

  publish_to_all_tenants = true # This needs to be published to all the Organizations

  # As we use rights created by the CAPVCD Type created previously, we need to depend on it
  depends_on = [
    vcd_rights_bundle.k8s_clusters_rights_bundle
  ]
}

# The networking setup specified below will configure one Provider Gateway + Edge Gateway + Routed network per
# organization. You can customise this section according to your needs.

data "vcd_nsxt_manager" "cse_nsxt_manager" {
  name = var.nsxt_manager_name
}

data "vcd_nsxt_tier0_router" "solutions_tier0_router" {
  name            = var.solutions_nsxt_tier0_router_name
  nsxt_manager_id = data.vcd_nsxt_manager.cse_nsxt_manager.id
}

resource "vcd_external_network_v2" "solutions_tier0" {
  name = "solutions_tier0"

  nsxt_network {
    nsxt_manager_id      = data.vcd_nsxt_manager.cse_nsxt_manager.id
    nsxt_tier0_router_id = data.vcd_nsxt_tier0_router.solutions_tier0_router.id
  }

  ip_scope {
    gateway       = var.solutions_provider_gateway_gateway_ip
    prefix_length = var.solutions_provider_gateway_gateway_prefix_length

    dynamic "static_ip_pool" {
      for_each = var.solutions_provider_gateway_static_ip_ranges
      iterator = ip
      content {
        start_address = ip.value[0]
        end_address   = ip.value[1]
      }
    }
  }
}

data "vcd_nsxt_tier0_router" "tenant_tier0_router" {
  name            = var.tenant_nsxt_tier0_router_name
  nsxt_manager_id = data.vcd_nsxt_manager.cse_nsxt_manager.id
}

resource "vcd_external_network_v2" "tenant_tier0" {
  name = "tenant_tier0"

  nsxt_network {
    nsxt_manager_id      = data.vcd_nsxt_manager.cse_nsxt_manager.id
    nsxt_tier0_router_id = data.vcd_nsxt_tier0_router.tenant_tier0_router.id
  }

  ip_scope {
    gateway       = var.tenant_provider_gateway_gateway_ip
    prefix_length = var.tenant_provider_gateway_gateway_prefix_length

    dynamic "static_ip_pool" {
      for_each = var.tenant_provider_gateway_static_ip_ranges
      iterator = ip
      content {
        start_address = ip.value[0]
        end_address   = ip.value[1]
      }
    }
  }
}

# This Edge Gateway will consume automatically the available IPs from the Provider Gateway.
resource "vcd_nsxt_edgegateway" "solutions_edgegateway" {
  org      = vcd_org.solutions_organization.name
  owner_id = vcd_org_vdc.solutions_vdc.id

  name                = "solutions_edgegateway"
  external_network_id = vcd_external_network_v2.solutions_tier0.id

  subnet {
    gateway       = var.solutions_provider_gateway_gateway_ip
    prefix_length = var.solutions_provider_gateway_gateway_prefix_length
    primary_ip    = var.solutions_provider_gateway_static_ip_ranges[0][0]

    dynamic "allocated_ips" {
      for_each = var.solutions_provider_gateway_static_ip_ranges
      iterator = ip
      content {
        start_address = ip.value[0]
        end_address   = ip.value[1]
      }
    }
  }
}

# This Edge Gateway will consume automatically the available IPs from the Provider Gateway.
resource "vcd_nsxt_edgegateway" "tenant_edgegateway" {
  org      = vcd_org.tenant_organization.name
  owner_id = vcd_org_vdc.tenant_vdc.id

  name                = "tenant_edgegateway"
  external_network_id = vcd_external_network_v2.tenant_tier0.id

  subnet {
    gateway       = var.tenant_provider_gateway_gateway_ip
    prefix_length = var.tenant_provider_gateway_gateway_prefix_length
    primary_ip    = var.tenant_provider_gateway_static_ip_ranges[0][0]

    dynamic "allocated_ips" {
      for_each = var.tenant_provider_gateway_static_ip_ranges
      iterator = ip
      content {
        start_address = ip.value[0]
        end_address   = ip.value[1]
      }
    }
  }
}

# CSE requires ALB to be configured to support the LoadBalancers that are deployed by the CPI of VMware Cloud Director.
resource "vcd_nsxt_alb_controller" "cse_avi_controller" {
  name     = "cse_alb_controller"
  username = var.alb_controller_username
  password = var.alb_controller_password
  url      = var.alb_controller_url
}

data "vcd_nsxt_alb_importable_cloud" "cse_importable_cloud" {
  name          = var.alb_importable_cloud_name
  controller_id = vcd_nsxt_alb_controller.cse_avi_controller.id
}

resource "vcd_nsxt_alb_cloud" "cse_nsxt_alb_cloud" {
  name = "cse_nsxt_alb_cloud"

  controller_id       = vcd_nsxt_alb_controller.cse_avi_controller.id
  importable_cloud_id = data.vcd_nsxt_alb_importable_cloud.cse_importable_cloud.id
  network_pool_id     = data.vcd_nsxt_alb_importable_cloud.cse_importable_cloud.network_pool_id
}

resource "vcd_nsxt_alb_service_engine_group" "cse_alb_seg" {
  name                                 = "cse_alb_seg"
  alb_cloud_id                         = vcd_nsxt_alb_cloud.cse_nsxt_alb_cloud.id
  importable_service_engine_group_name = "Default-Group"
  reservation_model                    = "SHARED"
}

# We introduce a sleep to wait for the provider part of ALB to be ready before the assignment to the Edge gateways
resource "time_sleep" "cse_alb_wait" {
  depends_on      = [vcd_nsxt_alb_service_engine_group.cse_alb_seg]
  create_duration = "30s"
}

## ALB for solutions edge gateway
resource "vcd_nsxt_alb_settings" "solutions_alb_settings" {
  org             = vcd_org.solutions_organization.name
  edge_gateway_id = vcd_nsxt_edgegateway.solutions_edgegateway.id
  is_active       = true

  # This dependency is required to make sure that provider part of operations is done
  depends_on = [time_sleep.cse_alb_wait]
}

resource "vcd_nsxt_alb_edgegateway_service_engine_group" "solutions_assignment" {
  org                       = vcd_org.solutions_organization.name
  edge_gateway_id           = vcd_nsxt_alb_settings.solutions_alb_settings.edge_gateway_id
  service_engine_group_id   = vcd_nsxt_alb_service_engine_group.cse_alb_seg.id
  reserved_virtual_services = 50
  max_virtual_services      = 50
}

resource "vcd_nsxt_alb_edgegateway_service_engine_group" "tenant_assignment" {
  org                       = vcd_org.tenant_organization.name
  edge_gateway_id           = vcd_nsxt_alb_settings.tenant_alb_settings.edge_gateway_id
  service_engine_group_id   = vcd_nsxt_alb_service_engine_group.cse_alb_seg.id
  reserved_virtual_services = 50
  max_virtual_services      = 50
}

## ALB for tenant edge gateway
resource "vcd_nsxt_alb_settings" "tenant_alb_settings" {
  org             = vcd_org.tenant_organization.name
  edge_gateway_id = vcd_nsxt_edgegateway.tenant_edgegateway.id
  is_active       = true

  # This dependency is required to make sure that provider part of operations is done
  depends_on = [time_sleep.cse_alb_wait]
}

# We create a Routed network in the Solutions organization that will be used by the CSE Server.
resource "vcd_network_routed_v2" "solutions_routed_network" {
  org         = vcd_org.solutions_organization.name
  name        = "solutions_routed_network"
  description = "Solutions routed network"

  edge_gateway_id = vcd_nsxt_edgegateway.solutions_edgegateway.id

  gateway       = var.solutions_routed_network_gateway_ip
  prefix_length = var.solutions_routed_network_prefix_length

  static_ip_pool {
    start_address = var.solutions_routed_network_ip_pool_start_address
    end_address   = var.solutions_routed_network_ip_pool_end_address
  }

  dns1       = var.solutions_routed_network_dns
  dns_suffix = var.solutions_routed_network_dns_suffix
}

# We create a Routed network in the Tenant organization that will be used by the Kubernetes clusters.
resource "vcd_network_routed_v2" "tenant_routed_network" {
  org         = vcd_org.tenant_organization.name
  name        = "tenant_net_routed"
  description = "Routed network for the K8s clusters"

  edge_gateway_id = vcd_nsxt_edgegateway.tenant_edgegateway.id

  gateway       = var.tenant_routed_network_gateway_ip
  prefix_length = var.tenant_routed_network_prefix_length

  static_ip_pool {
    start_address = var.tenant_routed_network_ip_pool_start_address
    end_address   = var.tenant_routed_network_ip_pool_end_address
  }

  dns1       = var.tenant_routed_network_dns
  dns_suffix = var.tenant_routed_network_dns_suffix
}

# We need SNAT rules in both networks to provide with Internet connectivity.
resource "vcd_nsxt_nat_rule" "solutions_nat" {
  org             = vcd_org.solutions_organization.name
  edge_gateway_id = vcd_nsxt_edgegateway.solutions_edgegateway.id

  name        = "Solutions SNAT rule"
  rule_type   = "SNAT"
  description = "Solutions SNAT rule"

  external_address = var.solutions_snat_external_ip
  internal_address = var.solutions_snat_internal_network_cidr
  logging          = true
}

resource "vcd_nsxt_nat_rule" "tenant_nat" {
  org             = vcd_org.solutions_organization.name
  edge_gateway_id = vcd_nsxt_edgegateway.tenant_edgegateway.id

  name        = "Tenant SNAT rule"
  rule_type   = "SNAT"
  description = "Tenant SNAT rule"

  external_address = var.tenant_snat_external_ip
  internal_address = var.tenant_snat_internal_network_cidr
  logging          = true
}

# WARNING: Please adjust this rule to your needs. The CSE Server requires Internet access to be configured.
resource "vcd_nsxt_firewall" "solutions_firewall" {
  org             = vcd_org.solutions_organization.name
  edge_gateway_id = vcd_nsxt_edgegateway.solutions_edgegateway.id

  rule {
    action      = "ALLOW"
    name        = "Allow all traffic"
    direction   = "IN_OUT"
    ip_protocol = "IPV4_IPV6"
  }
}

# WARNING: Please adjust this rule to your needs. The Bootstrap clusters and final Kubernetes clusters require Internet access to be configured.
resource "vcd_nsxt_firewall" "tenant_firewall" {
  org             = vcd_org.tenant_organization.name
  edge_gateway_id = vcd_nsxt_edgegateway.tenant_edgegateway.id

  rule {
    action      = "ALLOW"
    name        = "Allow all traffic"
    direction   = "IN_OUT"
    ip_protocol = "IPV4_IPV6"
  }
}

# Fetch the RDE Type created in 3.10-cse-4.0-install-step1.tf, as we need to create the configuration instance.
data "vcd_rde_type" "existing_vcdkeconfig_type" {
  vendor  = "vmware"
  nss     = "VCDKEConfig"
  version = "1.0.0"
}

# This RDE should be applied as it is.
resource "vcd_rde" "vcdkeconfig_instance" {
  org         = var.administrator_org
  name        = "vcdKeConfig"
  rde_type_id = data.vcd_rde_type.existing_vcdkeconfig_type.id
  resolve     = true
  input_entity = templatefile(var.vcdkeconfig_template_filepath, {
    capvcd_version                  = var.capvcd_version
    capvcd_rde_version              = var.capvcd_rde_version
    cpi_version                     = var.cpi_version
    csi_version                     = var.csi_version
    github_personal_access_token    = var.github_personal_access_token
    bootstrap_cluster_sizing_policy = vcd_vm_sizing_policy.tkg_s.name # References the small VM Sizing Policy
    no_proxy                        = var.no_proxy
    http_proxy                      = var.http_proxy
    https_proxy                     = var.https_proxy
    syslog_host                     = var.syslog_host
    syslog_port                     = var.syslog_port
  })
}

resource "vcd_vapp" "cse_server_vapp" {
  org  = vcd_org.solutions_organization.name
  vdc  = vcd_org_vdc.solutions_vdc.name
  name = "CSE Server vApp"

  lease {
    runtime_lease_in_sec = 0
    storage_lease_in_sec = 0
  }
}

resource "vcd_vapp_org_network" "cse_server_network" {
  org = vcd_org.solutions_organization.name
  vdc = vcd_org_vdc.solutions_vdc.name

  vapp_name        = vcd_vapp.cse_server_vapp.name
  org_network_name = vcd_network_routed_v2.solutions_routed_network.name

  reboot_vapp_on_removal = true
}

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
    "cse.vcdRefreshToken" = var.cse_admin_api_token

    # CSE admin account's username
    "cse.vcdUsername" = var.cse_admin_user

    # CSE admin account's org
    "cse.userOrg" = var.administrator_org
  }

  customization {
    force                      = false
    enabled                    = true
    allow_local_admin_password = true
    auto_generate_password     = true
  }

  depends_on = [
    vcd_rde.vcdkeconfig_instance
  ]
}

data "vcd_org" "system_org" {
  name = var.administrator_org
}

resource vcd_ui_plugin "k8s_container_clusters_ui_plugin" {
  count = var.k8s_container_clusters_ui_plugin_path == "" ? 0 : 1
  plugin_path = var.k8s_container_clusters_ui_plugin_path
  enabled = true
  tenant_ids = [
    data.vcd_org.system_org.id,
    vcd_org.solutions_organization.id,
    vcd_org.tenant_organization.id,
  ]
}
