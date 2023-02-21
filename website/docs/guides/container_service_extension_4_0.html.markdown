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

You can of course create more policies that suit to your needs. The above policies are **just the minimum required**.

### Set up the VDCs

In this step we will create a specific [VDC](/providers/vmware/vcd/latest/docs/resources/org_vdc) that will host
the CSE appliance and its configuration, called "Solutions VDC", and a second [VDC](/providers/vmware/vcd/latest/docs/resources/org_vdc)
that will host the clusters for the tenants to use, called "Cluster VDC".

You can customise the following sample HCL snippet to your needs. It creates these two [VDCs](/providers/vmware/vcd/latest/docs/resources/org_vdc)

```hcl
# We fetch some required information like Provider VDC, Edge Clusters, etc
data "vcd_provider_vdc" "nsxt_pvdc" {
  name = "nsxTPvdc1"
}

data "vcd_nsxt_edge_cluster" "cluster_edgecluster" {
  org             = vcd_org.cluster_organization.name
  provider_vdc_id = data.vcd_provider_vdc.nsxt_pvdc.id
  name            = "edgeCluster1"
}

# The VDC that will host the Kubernetes clusters
resource "vcd_org_vdc" "cluster_vdc" {
  name        = "cluster_vdc"
  description = "Cluster VDC"
  org         = vcd_org.cluster_organization.name

  allocation_model  = "AllocationVApp" # You can use other models
  network_pool_name = "NSX-T Overlay 1"
  provider_vdc_name = data.vcd_provider_vdc.nsxt_pvdc.name
  edge_cluster_id   = data.vcd_nsxt_edge_cluster.cluster_edgecluster.id

  # You can tune these arguments to your needs
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

data "vcd_nsxt_edge_cluster" "solutions_edgecluster" {
  org             = vcd_org.cluster_organization.name
  provider_vdc_id = data.vcd_provider_vdc.nsxt_pvdc.id
  name            = "edgeCluster2"
}

# The VDC that will host the CSE appliance
resource "vcd_org_vdc" "solutions_vdc" {
  name        = "solutions_vdc"
  description = "Solutions VDC"
  org         = vcd_org.solutions_organization.name

  allocation_model  = "AllocationVApp" # You can use other models
  network_pool_name = "NSX-T Overlay 1"
  provider_vdc_name = data.vcd_provider_vdc.nsxt_pvdc.name
  edge_cluster_id   = data.vcd_nsxt_edge_cluster.solutions_edgecluster.id

  # You can tune these arguments to your needs
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

  enabled                  = true
  enable_thin_provisioning = true
  enable_fast_provisioning = true
  delete_force             = true
  delete_recursive         = true

  # You can create more Sizing Policies and add them to the VDC:
  default_compute_policy_id = vcd_vm_sizing_policy.other_policy.id
  vm_sizing_policy_ids = [
    vcd_vm_sizing_policy.other_policy.id
  ]
}
```

### Create Catalogs and upload OVAs

We need to create some [Catalogs](/providers/vmware/vcd/latest/docs/resources/catalog) to be able to store and retrieve
CSE Server OVAs and maintain a repository of Kubernetes Template OVAs.
In this step, we will create two [Catalogs](/providers/vmware/vcd/latest/docs/resources/catalog):

- One Catalog in the Solutions Organization to upload CSE Server OVA for easy access.
- One shared Catalog in the Solutions Organization that will contain the Kubernetes Template OVAs.

Here's a sample HCL that can help you to achieve this setup:

```hcl
resource "vcd_catalog" "cse_catalog" {
  org  = vcd_org.solutions_organization.name # References the Solutions Organization created previously
  name = "cse_catalog"

  delete_force     = "true"
  delete_recursive = "true"
  # You can use a specific `storage_profile_id` argument here to use the same storage as the Solutions VDC
}

resource "vcd_catalog" "tkgm_catalog" {
  org  = vcd_org.solutions_organization.name # References the Solutions Organization created previously
  name = "tkgm_catalog"

  delete_force     = "true"
  delete_recursive = "true"
  # You can use a specific `storage_profile_id` argument here to use the same storage as the Solutions VDC
}

# We share the TKGm Catalog with the Cluster Organization created previously.
resource "vcd_catalog_access_control" "tkgm_catalog_ac" {
  org                  = vcd_org.solutions_organization.name # References the Solutions Organization created previously
  catalog_id           = vcd_catalog.tkgm_catalog.id
  shared_with_everyone = false
  shared_with {
    org_id       = vcd_org.cluster_organization.id
    access_level = "ReadOnly"
  }
}
```

If you have already some [Catalogs](/providers/vmware/vcd/latest/docs/data-sources/catalog) available, you can fetch them
with a data source instead:

```hcl
data "vcd_catalog" "cse_catalog" {
  org  = vcd_org.solutions_organization.name
  name = "cse_catalog"
}

# This should be shared if it belongs to the Solutions Organization
data "vcd_catalog" "tkgm_catalog" {
  org  = vcd_org.solutions_organization.name
  name = "tkgm_catalog"
}
```

To upload both CSE and TKGm OVAs, you can use the following sample HCL snippets:

```hcl
resource "vcd_catalog_vapp_template" "tkgm_ova" {
  org        = vcd_org.solutions_organization.name  # References the Solutions Organization created previously
  catalog_id = vcd_catalog.tkgm_catalog.id  # References the TKGm Catalog created previously

  name        = replace(var.tkgm_ova_file, ".ova", "")
  description = replace(var.tkgm_ova_file, ".ova", "")
  ova_path    = format("%s/%s", var.tkgm_ova_folder, var.tkgm_ova_file)
}

resource "vcd_catalog_vapp_template" "cse_ova" {
  org        = vcd_org.solutions_organization.name  # References the Solutions Organization created previously
  catalog_id = vcd_catalog.cse_catalog.id # References the CSE Catalog created previously

  name        = replace(var.cse_ova_file, ".ova", "")
  description = replace(var.cse_ova_file, ".ova", "")
  ova_path    = format("%s/%s", var.cse_ova_folder, var.cse_ova_file)
}
```

As you can see, the `name`, `description` and `ova_path` are taken from variables. This is just a suggestion. You can
use other ways of retrieving the OVAs that are better suited to your needs.

### Register Interfaces and Entity Types

It is required that you add the following [Runtime Defined Entity Interfaces](/providers/vmware/vcd/latest/docs/resources/rde_interface)
and [Runtime Defined Entity Types](/providers/vmware/vcd/latest/docs/data-sources/rde_type) to VCD:

```hcl
resource "vcd_rde_interface" "vcd_ke_config_interface" {
  name    = "VCDKEConfig"
  version = "1.0.0"
  vendor  = "vmware"
  nss     = "VCDKEConfig"
}

# This one exists in VCD
data "vcd_rde_interface" "kubernetes_interface" {
  vendor  = "vmware"
  nss     = "k8s"
  version = "1.0.0"
}

resource "vcd_rde_type" "vcd_ke_config_type" {
  name          = "VCD-KE RDE Schema"
  nss           = "VCDKEConfig"
  version       = "1.0.0"
  schema        = file("${path.module}/schemas/vcdkeconfig-type-schema.json")
  vendor        = "vmware"
  interface_ids = [vcd_rde_interface.vcd_ke_config_interface.id]
}

resource "vcd_rde_type" "capvcd_cluster_type" {
  name          = "CAPVCD Cluster"
  nss           = "capvcdCluster"
  version       = "1.1.0"
  schema        = file("${path.module}/schemas/capvcd-type-schema.json")
  vendor        = "vmware"
  interface_ids = [data.vcd_rde_interface.kubernetes_interface.id]
}
```

### Create CSE Admin Role

### Create and Publish 'Kubernetes Clusters Rights Bundle'

### Create and Publish 'Kubernetes Cluster Author' global role