---
layout: "vcd"
page_title: "VMware Cloud Director: Container Service Extension v4.0"
sidebar_current: "docs-vcd-guides-cse"
description: |-
  Provides guidance on configuring VCD to be able to install and use Container Service Extension v4.0
---

# Container Service Extension v4.0

## About

This guide describes the required steps to configure VCD to install the Container Service Extension (CSE) v4.0, that
will allow tenant users to deploy **Tanzu Kubernetes Grid Multi-cloud (TKGm)** clusters on VCD using Terraform.

To know more about CSE v4.0, you can visit [the documentation](https://docs.vmware.com/en/VMware-Cloud-Director-Container-Service-Extension/index.html).

## Pre-requisites

In order to complete the steps described in this guide, please be aware:

* CSE v4.0 is supported from VCD v10.4.0 or above, make sure your VCD appliance matches the criteria.
* Terraform provider needs to be v3.9.0 or above.

## Installation process

To start installing CSE v4.0 in a VCD appliance, you must use **v3.9.0 or above** of the VCD Terraform Provider:

```hcl
terraform {
  required_providers {
    vcd = {
      source  = "vmware/vcd"
      version = ">= 3.9"
    }
  }
}

provider "vcd" {
  user                 = "administrator"
  password             = "*******"
  auth_type            = "integrated"
  sysorg               = "System"
  org                  = "System"
  url                  = "https://vcd.my-company.com/api"
}
```

As you will be creating several administrator-scoped resources like Orgs, VDCs, Provider Gateways, etc; make sure you provide 
**System administrator** credentials.

### Set up the Organizations

In this step we will create a specific [Organization](/providers/vmware/vcd/latest/docs/resources/org) that will host
the CSE appliance and its configuration, called "Solutions Organization", and a second [Organization](/providers/vmware/vcd/latest/docs/resources/org)
that will host the clusters for the tenants to use, called "Cluster Organization".

You can customise the following sample HCL snippet to your needs. It creates these two [Organizations](/providers/vmware/vcd/latest/docs/resources/org)
with no limits on lease:

```hcl
resource "vcd_org" "cluster_organization" {
  name             = "cluster_org"
  full_name        = "Cluster Organization"
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
```

If you have already some [Organizations](/providers/vmware/vcd/latest/docs/data-sources/org) available, you can fetch them
with a data source instead:

```hcl
data "vcd_org" "cluster_organization" {
  name = "cluster_org"
}

data "vcd_org" "solutions_organization" {
  name = "solutions_org"
}
```

### Create the needed Sizing Policies

CSE 4.0 requires a specific set of Sizing Policies to be able to dimension the Kubernetes clusters. You can create them
with the following HCL snippet. The names and descriptions should not be modified from this snippet.

```hcl
resource "vcd_vm_sizing_policy" "tkg_xl" {
  name        = "TKG extra_large"
  description = "Extra large VM sizing policy for a Kubernetes cluster node"
  cpu {
    count = 8
  }
  memory {
    size_in_mb = "32768"
  }
}

resource "vcd_vm_sizing_policy" "tkg_l" {
  name        = "TKG large"
  description = "Large VM sizing policy for a Kubernetes cluster node"
  cpu {
    count = 4
  }
  memory {
    size_in_mb = "16384"
  }
}

resource "vcd_vm_sizing_policy" "tkg_m" {
  name        = "TKG medium"
  description = "Medium VM sizing policy for a Kubernetes cluster node"
  cpu {
    count = 2
  }
  memory {
    size_in_mb = "8192"
  }
}

resource "vcd_vm_sizing_policy" "tkg_s" {
  name        = "TKG small"
  description = "Small VM sizing policy for a Kubernetes cluster node"
  cpu {
    count = 2
  }
  memory {
    size_in_mb = "4048"
  }
}
```

### Set up the VDCs

In this step we will create a specific [VDC](/providers/vmware/vcd/latest/docs/resources/org_vdc) that will host
the CSE appliance and its configuration, called "Solutions VDC", and a second [VDC](/providers/vmware/vcd/latest/docs/resources/org_vdc)
that will host the clusters for the tenants to use, called "Cluster VDC".

You can customise the following sample HCL snippet to your needs. It creates these two [VDCs](/providers/vmware/vcd/latest/docs/resources/org_vdc)

~> The target VDC needs to be backed by **NSX-T** for CSE to work.

```hcl
resource "vcd_org_vdc" "cluster_vdc" {
  name        = "cluster_vdc"
  description = "Cluster VDC"
  org         = vcd_org.cluster_organization.name

  allocation_model  = "AllocationVApp"
  network_pool_name = "NSX-T Overlay 1"
  provider_vdc_name = data.vcd_provider_vdc.nsxt_pvdc.name
  edge_cluster_id   = data.vcd_nsxt_edge_cluster.cluster_edgecluster.id

  network_quota = 1000

  compute_capacity {
    cpu {
      allocated = 0
    }

    memory {
      allocated = 0
    }
  }

  storage_profile {
    name    = "*"
    limit   = 0
    default = true
  }

  storage_profile {
    name    = "Development2"
    limit   = 0
    default = false
  }

  enabled                  = true
  enable_thin_provisioning = true
  enable_fast_provisioning = true
  delete_force             = true
  delete_recursive         = true

  default_compute_policy_id = vcd_vm_sizing_policy.tkg_s.id

  vm_sizing_policy_ids = [
    vcd_vm_sizing_policy.tkg_xl.id,
    vcd_vm_sizing_policy.tkg_l.id,
    vcd_vm_sizing_policy.tkg_m.id,
    vcd_vm_sizing_policy.tkg_s.id,
  ]
}
```

