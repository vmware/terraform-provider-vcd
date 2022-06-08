# If you have already an Organization, you can remove the `vcd_org` resource from this HCL file and simply
# configure it in the provider settings.

resource "vcd_org" "cse_org" {
  name              = "cse_org"
  full_name         = "cse_org"
  is_enabled        = "true"
  stored_vm_quota   = 50
  deployed_vm_quota = 50
  delete_force      = "true"
  delete_recursive  = "true"

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

# If you have already a VDC, you can remove the `vcd_org_vdc` resource from this HCL file and simply
# configure it in the provider settings. The VDC must be backed by NSX-T.

resource "vcd_org_vdc" "cse_vdc" {
  name = "cse_vdc"
  org  = vcd_org.cse_org.name

  allocation_model  = "AllocationVApp"
  provider_vdc_name = "NSXT_PVDC"
  network_pool_name = "NSX-T Overlay 1"
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
    name    = "*"
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

# Here we create a Tier 0 Gateway connected to the outside world network. This will be used to download software
# for the Kubernetes nodes and access the cluster.

data "vcd_nsxt_manager" "main" {
  name = "my-nsx-manager"
}

data "vcd_nsxt_tier0_router" "router" {
  name            = "VCD T0 Edge"
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
    gateway       = "10.0.0.1"
    prefix_length = "16"

    static_ip_pool {
      start_address = "10.0.0.2"
      end_address   = "10.0.255.254"
    }
  }
}

# Create an Edge Gateway that will be used by the cluster as the main router.

resource "vcd_nsxt_edgegateway" "cse_egw" {
  org      = vcd_org.cse_org.name
  owner_id = vcd_org_vdc.cse_vdc.id

  name                = "cse-egw"
  description         = "CSE edge gateway"
  external_network_id = vcd_external_network_v2.cse_external_network_nsxt.id

  subnet {
    gateway       = "10.0.0.1"
    prefix_length = "24"
    primary_ip    = "10.0.0.2"

    allocated_ips {
      start_address = "10.0.0.2"
      end_address   = "10.0.0.254"
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

  gateway       = "192.168.7.1"
  prefix_length = 24

  static_ip_pool {
    start_address = "192.168.7.2"
    end_address   = "192.168.7.100"
  }

  dns1 = "8.8.8.8"
  dns2 = "8.8.8.4"
}

# NAT rule to map traffic to internal network IPs.

resource "vcd_nsxt_nat_rule" "snat" {
  org             = vcd_org.cse_org.name
  edge_gateway_id = vcd_nsxt_edgegateway.cse_egw.id

  name        = "SNAT rule"
  rule_type   = "SNAT"
  description = "description"

  external_address = "10.0.0.3"
  internal_address = "192.168.7.0/24"
  logging          = true
}

# Cluster requires network traffic is open, to download required dependencies to create nodes. Adapt this firewall
# rule to your organization security requirements, as this is just an example.

resource "vcd_nsxt_firewall" "firewall" {
  org             = vcd_org.cse_org.name
  edge_gateway_id = vcd_nsxt_edgegateway.cse_egw.id

  rule {
    action      = "ALLOW"
    name        = "allow all IPv4 traffic"
    direction   = "IN_OUT"
    ip_protocol = "IPV4"
  }
}

# Catalog to upload the TKGm OVAs.

data "vcd_storage_profile" "cse_sp" {
  org  = vcd_org.cse_org.name
  vdc  = vcd_org_vdc.cse_vdc.name
  name = "*"

  depends_on = [vcd_org.cse_org, vcd_org_vdc.cse_vdc]
}

resource "vcd_catalog" "cat-cse" {
  org         = vcd_org.cse_org.name
  name        = "cat-cse"
  description = "CSE catalog"

  storage_profile_id = data.vcd_storage_profile.cse_sp.id

  delete_force     = "true"
  delete_recursive = "true"
  depends_on       = [vcd_org_vdc.cse_vdc]
}

# TKGm OVA upload. The `catalog_item_metadata` is required for CSE to detect the OVAs.

resource "vcd_catalog_item" "tkgm_ova" {
  org     = vcd_org.cse_org.name
  catalog = vcd_catalog.cat-cse.name

  name                 = "ubuntu-2004-kube-v1.21.2+vmware.1-tkg.1-7832907791984498322"
  description          = "ubuntu-2004-kube-v1.21.2+vmware.1-tkg.1-7832907791984498322"
  ova_path             = "/Users/bob/CSE/TKGm/ubuntu-2004-kube-v1.21.2+vmware.1-tkg.1-7832907791984498322.iso"
  upload_piece_size    = 100
  show_upload_progress = true

  catalog_item_metadata = {
    "kind"               = "TKGm" # This value is always the same
    "kubernetes"         = "TKGm" # This value is always the same
    "kubernetes_version" = "v1.21.2+vmware.1" # The version comes in the OVA name downloaded from Customer Connect
    "name"               = "ubuntu-2004-kube-v1.21.2+vmware.1-tkg.1-7832907791984498322" # The name as it was in the OVA downloaded from Customer Connect
    "os"                 = "ubuntu" # The OS comes in the OVA name downloaded from Customer Connect
    "revision"           = "1" # This value is always the same
  }
}

# AVI configuration for Kubernetes services, this allows the cluster to create Kubernetes services of type Load Balancer.

data "vcd_nsxt_alb_controller" "cse_alb_controller" {
  name = "aviController1"
}

data "vcd_nsxt_alb_importable_cloud" "cse_importable_cloud" {
  name          = "NSXT avi.foo.com"
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
  org             = vcd_org.cse_org.name
  edge_gateway_id = vcd_nsxt_edgegateway.cse_egw.id
  is_active       = true

  # This dependency is required to make sure that provider part of operations is done
  depends_on = [vcd_nsxt_alb_service_engine_group.cse_alb_seg]
}

resource "vcd_nsxt_alb_edgegateway_service_engine_group" "assignment" {
  org                     = vcd_org.cse_org.name
  edge_gateway_id         = vcd_nsxt_edgegateway.cse_egw.id
  service_engine_group_id = vcd_nsxt_alb_service_engine_group.cse_alb_seg.id
}

resource "vcd_nsxt_alb_pool" "cse_alb_pool" {
  org             = vcd_org.cse_org.name
  edge_gateway_id = vcd_nsxt_edgegateway.cse_egw.id
  name            = "cse-avi-pool"
}

resource "vcd_nsxt_alb_virtual_service" "cse-virtual-service" {
  org             = vcd_org.cse_org.name
  edge_gateway_id = vcd_nsxt_edgegateway.cse_egw.id
  name            = "cse-virtual-service"

  pool_id                  = vcd_nsxt_alb_pool.cse_alb_pool.id
  service_engine_group_id  = vcd_nsxt_alb_edgegateway_service_engine_group.assignment.service_engine_group_id
  virtual_ip_address       = "192.168.8.88"
  application_profile_type = "HTTP"
  service_port {
    start_port = 80
    type       = "TCP_PROXY"
  }
}
