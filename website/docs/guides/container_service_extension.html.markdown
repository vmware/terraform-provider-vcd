---
layout: "vcd"
page_title: "VMware Cloud Director: Container Service Extension"
sidebar_current: "docs-vcd-guides-cse"
description: |-
Provides guidance on Container Service Extension for VCD.
---

# Container Service Extension

## About

Container Service Extension (CSE) is a VCD extension that helps tenants create and work with Kubernetes clusters.

CSE brings Kubernetes as a Service to VCD, by creating customized VM templates (Kubernetes templates) and enabling tenant users to deploy fully functional Kubernetes clusters as self-contained vApps.

To know more about CSE, you can explore [the official website](https://vmware.github.io/container-service-extension/).

## Requirements

* CSE is supported from VCD 10.3.0 or above
* Used Terraform provider needs to be v3.7.0 or above
* All CSE elements use NSX-T backed resources, **no** NSX-V is supported

## Installation process

To start installing CSE in a VCD appliance, you can use **v3.7.0 or above** of the VCD Provider:

```hcl
terraform {
  required_providers {
    vcd = {
      source  = "vmware/vcd"
      version = ">= 3.7.0"
    }
  }
}

provider "vcd" {
  user                 = "administrator"
  password             = var.vcd_pass
  auth_type            = "integrated"
  sysorg               = "System"
  url                  = var.vcd_url
  max_retry_timeout    = var.vcd_max_retry_timeout
  allow_unverified_ssl = var.vcd_allow_unverified_ssl
}
```

### Step 1: Initialization

This step assumes that you want to install CSE in a brand new [Organization](/providers/vmware/vcd/latest/docs/resources/org)
with no [VDCs](/providers/vmware/vcd/latest/docs/resources/org_vdc), or that is a fresh installation of VCD.
Otherwise, please skip this step and configure `org` and `vdc` attributes in the provider configuration above.

The VDC needs to be backed by **NSX-T** for CSE to work. Here is an example that creates both the Organization and the VDC:

```hcl
resource "vcd_org" "cse_org" {
  name              = "cse_org"
  full_name         = "cse_org"
  is_enabled        = "true"
  delete_force      = "true"
  delete_recursive  = "true"
}

resource "vcd_org_vdc" "cse_vdc" {
  name = "cse_vdc"
  org  = vcd_org.cse_org.name

  allocation_model  = "AllocationVApp"
  network_pool_name = "NSX-T Overlay"
  provider_vdc_name = "nsxTPvdc1"

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
```

## Step 2: Configure networking

For the Kubernetes clusters to be functional, you need to provide some networking resources to the target VDC:

* [Tier-0 Gateway](/providers/vmware/vcd/latest/docs/resources/external_network_v2)
* [Edge Gateway](/providers/vmware/vcd/latest/docs/resources/nsxt_edgegateway)
* [Routed Network](/providers/vmware/vcd/latest/docs/resources/network_routed_v2)
* [SNAT rule](/providers/vmware/vcd/latest/docs/resources/nsxt_nat_rule)

The [Tier-0 Gateway](/providers/vmware/vcd/latest/docs/resources/external_network_v2) will provide access to the
outside world. For example, this will allow cluster users to communicate with Kubernetes API server through `kubectl`.

Here is an example on how to configure this resource:

```hcl
data "vcd_nsxt_manager" "main" {
  name = "my-nsxt-manager"
}

data "vcd_nsxt_tier0_router" "router" {
  name            = "VCD T0 edgeCluster"
  nsxt_manager_id = data.vcd_nsxt_manager.main.id
}

resource "vcd_external_network_v2" "cse_external_network_nsxt" {
  name        = "extnet-cse"
  description = "NSX-T backed network for k8s clusters"

  nsxt_network {
    nsxt_manager_id      = data.vcd_nsxt_manager.main.id
    nsxt_tier0_router_id = data.vcd_nsxt_tier0_router.router.id
  }

  ip_scope {
    gateway       = "88.88.88.1"
    prefix_length = "24"

    static_ip_pool {
      start_address = "88.88.88.88"
      end_address   = "88.88.88.100"
    }
  }
}
```

Create also an [Edge Gateway](/providers/vmware/vcd/latest/docs/resources/nsxt_edgegateway) that will use the recently created
external network. This will act as the main router connecting our nodes in the internal network to the external (Tier 0 Gateway) network:

```hcl
resource "vcd_nsxt_edgegateway" "cse_egw" {
  org      = vcd_org.cse_org.name
  owner_id = vcd_org_vdc.cse_vdc.id

  name                = "cse-egw"
  description         = "CSE edge gateway"
  external_network_id = vcd_external_network_v2.cse_external_network_nsxt.id

  subnet {
    gateway       = "88.88.88.1"
    prefix_length = "24"
    primary_ip    = "88.88.88.88"
    allocated_ips {
      start_address = "88.88.88.88"
      end_address   = "88.88.88.100"
    }
  }
  depends_on = [vcd_org_vdc.cse_vdc]
}
```

This will create a basic Edge Gateway, you can of course add more complex [firewall rules](/providers/vmware/vcd/latest/docs/resources/nsxt_firewall)
or other configurations to fit with your organization requirements.

Create a [Routed Network](/providers/vmware/vcd/latest/docs/resources/network_routed_v2) that will be using the recently
created Edge Gateway. This network is the one used by all the Kubernetes nodes in the cluster:

```hcl
resource "vcd_network_routed_v2" "cse_routed" {
  org         = vcd_org.cse_org.name
  name        = "cse_routed_net"
  description = "My routed Org VDC network backed by NSX-T"

  edge_gateway_id = vcd_nsxt_edgegateway.cse_egw.id

  gateway       = "192.168.7.0"
  prefix_length = 24

  # This network will allow us to have 256 Kubernetes nodes
  static_ip_pool {
    start_address = "192.168.7.1"
    end_address   = "192.168.7.100"
  }

  dns1       = "8.8.8.8"
  dns2       = "8.8.8.4"
}
```

To be able to reach the Kubernetes nodes within the routed network, you need also a [SNAT rule](/providers/vmware/vcd/latest/docs/resources/nsxt_nat_rule):

```hcl
resource "vcd_nsxt_nat_rule" "snat" {
  org = vcd_org.cse_org.name
  edge_gateway_id = vcd_nsxt_edgegateway.cse_egw.id

  name        = "SNAT rule"
  rule_type   = "SNAT"
  description = "description"
  
  external_address = "88.88.88.89" # A public IP from the external network
  internal_address = "192.168.7.0/24" # This is the routed network CIDR
  logging          = true
}
```

## Step 3: Configure ALB

Advanced Load Balancers are required for CSE to be able to handle Kubernetes services and other casuistic.
You need the following resources:

* [ALB Controller](/providers/vmware/vcd/latest/docs/resources/nsxt_alb_controller)
* [ALB Cloud](/providers/vmware/vcd/latest/docs/resources/nsxt_alb_cloud)
* [ALB Service Engine Group](/providers/vmware/vcd/latest/docs/resources/nsxt_alb_service_engine_group)
* [ALB Settings](/providers/vmware/vcd/latest/docs/resources/nsxt_alb_settings)
* [ALB Edge Gateway SEG](/providers/vmware/vcd/latest/docs/resources/nsxt_alb_edgegateway_service_engine_group)
* [ALB Pool](/providers/vmware/vcd/latest/docs/resources/nsxt_alb_pool)
* [ALB Virtual Service](/providers/vmware/vcd/latest/docs/resources/nsxt_alb_virtual_service)

You can have a look at [this guide](/providers/vmware/vcd/latest/docs/guides/nsxt_alb) as it explains every resource
and provides some examples of how to setup ALB in VCD.

## Step 4: Configure catalogs and OVAs

You need to have a catalog for vApp Templates and some OVA files to be able to create Kubernetes clusters. Here is an
example of how to create the [Catalog](/providers/vmware/vcd/latest/docs/resources/catalog):

```hcl
data "vcd_storage_profile" "cse_storage_profile" {
  org  = vcd_org.cse_org.name
  vdc  = vcd_org_vdc.cse_vdc.name
  name = "*"
  depends_on = [vcd_org.cse_org, vcd_org_vdc.cse_vdc]
}

resource "vcd_catalog" "cat-cse" {
  org         = vcd_org.cse_org.name
  name        = "cat-cse"
  description = "CSE catalog"

  storage_profile_id = data.vcd_storage_profile.cse_storage_profile.id

  delete_force     = "true"
  delete_recursive = "true"
  depends_on       = [vcd_org_vdc.cse_vdc]
}
```

Then we can upload TKGm (Tanzu Kubernetes Grid) OVAs. These can be downloaded from VMware Customer Connect.
To upload them, use the [Catalog Item](/providers/vmware/vcd/latest/docs/resources/catalog_item) resource:

-> Note that CSE is **not compatible** yet with PhotonOS

```hcl
resource "vcd_catalog_item" "tkgm_ova" {
  org     = vcd_org.cse_org.name
  catalog = vcd_catalog.cat-cse.name

  name                 = "ubuntu-2004-kube-v1.21.2+vmware.1-tkg.1-7832907791984498322"
  description          = "ubuntu-2004-kube-v1.21.2+vmware.1-tkg.1-7832907791984498322"
  ova_path             = "/Users/johndoe/Download/ubuntu-2004-kube-v1.21.2+vmware.1-tkg.1-7832907791984498322"
  upload_piece_size    = 100
  show_upload_progress = true

  catalog_item_metadata = {
    "cni"                       = "antrea"
    "cni_version"               = "0.0.0"
    "container_runtime"         = "containerd"
    "container_runtime_version" = "v1.4.6+vmware.1"
    "cse_version"               = "3.1.2"
    "kind"                      = "TKGm"
    "kubernetes"                = "TKGm"
    "kubernetes_version"        = "v1.21.2+vmware.1"
    "name"                      = "ubuntu-2004-kube-v1.21.2+vmware.1-tkg.1-7832907791984498322"
    "os"                        = "ubuntu"
    "os_version"                = "20.04"
    "revision"                  = "1"
  }
}
```

All the metadata is required for CSE to fetch the OVA file.
Alternatively, you can upload the OVA file using `cse-cli`, explained in the next step.

## Step 5: CSE command cli

This is the only step that must be done without any Terraform script.
You need to [install CSE command line interface](https://vmware.github.io/container-service-extension/cse3_0/INSTALLATION.html#getting_cse)
and then provide a config.yaml with the entities that were created by Terraform.

When you execute the `cse install` command, CSE will be almost ready to be used, you only need to publish new right bundles and rights
to the organization.

## Step 6: Rights and roles

```hcl
# Default Rights bundle upgrade
data "vcd_rights_bundle" "default-rb" {
  name = "Default Rights Bundle"
}

data "vcd_rights_bundle" "cse-rights-bundle" {
  name = "cse:nativeCluster Entitlement"
}

# vApp Author role
data "vcd_role" "vapp_author" {
  org  = vcd_org.cse_org.name
  name = "vApp Author"
}

# If the below resources fail, do:
#
# terraform import vcd_rights_bundle.published-cse-rights-bundle "cse:nativeCluster Entitlement"
#
# Then retry again

resource "vcd_rights_bundle" "published-cse-rights-bundle" {
  name                   = data.vcd_rights_bundle.cse-rights-bundle.name
  description            = data.vcd_rights_bundle.cse-rights-bundle.description
  rights                 = data.vcd_rights_bundle.cse-rights-bundle.rights
  publish_to_all_tenants = true
}

resource "vcd_rights_bundle" "cse-rb" {
  name        = "CSE Rights Bundle"
  description = "Rights bundle to manage CSE"
  rights = setunion(data.vcd_rights_bundle.default-rb.rights, [
    "API Tokens: Manage",
    "Organization vDC Shared Named Disk: Create",
    "cse:nativeCluster: View",
    "cse:nativeCluster: Full Access",
    "cse:nativeCluster: Modify"
  ])
  publish_to_all_tenants = true
}

resource "vcd_role" "cluster_author" {
  org         = vcd_org.cse_org.name
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

resource "vcd_org_user" "cse_user" {
  org = vcd_org.cse_org.name

  name        = "cse_user"
  description = "Cluster author"
  role        = vcd_role.cluster_author.name
  password    = "ca$hc0w"
}
```