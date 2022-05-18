---
layout: "vcd"
page_title: "VMware Cloud Director: Container Service Extension"
sidebar_current: "docs-vcd-guides-cse"
description: |-
Provides guidance on Container Service Extension for VCD.
---

## Overview

Container Service Extension (CSE) is a VCD extension that helps tenants create and work with Kubernetes clusters.

CSE brings Kubernetes as a Service to VCD, by creating customized VM templates (Kubernetes templates) and enabling tenant users to deploy fully functional Kubernetes clusters as self-contained vApps.

To know more about CSE, you can explore [the official website](https://vmware.github.io/container-service-extension/).

## Installation

To install CSE in a VCD appliance, you can use **v3.7.0 or above** of the VCD Provider:

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

For the Kubernetes clusters to be functional, you need to provide some networking resources to the target VDC. First,
the [External network](/providers/vmware/vcd/latest/docs/resources/external_network_v2) will provide access to the outside world.
This will be required to be able to communicate with Kubernetes API server through `kubectl`.

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

Then the edge gateway that will be used with our Kubernetes clusters:

```hcl
resource "vcd_nsxt_edgegateway" "cse_egw" {
  org      = vcd_org.cse_org.name
  owner_id = vcd_org_vdc.cse_vdc.id

  name                = "cse-egw"
  description         = "CSE edge gateway"
  external_network_id = vcd_external_network_v2.cse_external_network_nsxt.id

  subnet {
    gateway       = local.testbed_gateway_ip
    prefix_length = "19"
    primary_ip    = local.testbed_routable_ips[0]
    allocated_ips {
      start_address = local.testbed_routable_ips[0]
      end_address   = local.testbed_routable_ips[0]
    }
  }

  depends_on = [vcd_org_vdc.cse_vdc]
}
```

Next step is to configure a valid routed network:

```hcl
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

  dns1       = local.testbed_dns
  dns2       = "8.8.8.4"
  dns_suffix = "eng.vmware.com"
}
```