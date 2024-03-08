# ------------------------------------------------------------------------------------------------------------
# CSE v4.1 installation:
#
# * Please read the guide at https://registry.terraform.io/providers/vmware/vcd/latest/docs/guides/container_service_extension_4_x_install
#   before applying this configuration.
#
# * Rename "terraform.tfvars.example" to "terraform.tfvars" and adapt the values to your needs.
#
# * Please review this file carefully, as it shapes the structure of your organization, hence you should customise
#   it to your needs.
#   You can check the comments on each resource/data source for more help and context.
# ------------------------------------------------------------------------------------------------------------

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

# Fetch the VM Sizing Policies created in step 1
data "vcd_vm_sizing_policy" "tkg_s" {
  name = "TKG small"
}

data "vcd_vm_sizing_policy" "tkg_m" {
  name = "TKG medium"
}

data "vcd_vm_sizing_policy" "tkg_l" {
  name = "TKG large"
}

data "vcd_vm_sizing_policy" "tkg_xl" {
  name = "TKG extra-large"
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

  # Make sure you specify the required VM Sizing Policies managed by the data sources specified above.
  default_compute_policy_id = data.vcd_vm_sizing_policy.tkg_s.id
  vm_sizing_policy_ids = [
    data.vcd_vm_sizing_policy.tkg_xl.id,
    data.vcd_vm_sizing_policy.tkg_l.id,
    data.vcd_vm_sizing_policy.tkg_m.id,
    data.vcd_vm_sizing_policy.tkg_s.id,
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
