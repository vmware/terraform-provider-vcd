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

To start installing CSE v4.0 in a VCD appliance, you must use **v3.9.0 or above** of the VCD Terraform Provider
configured as **System administrator**, as you'll need to create provider-scoped items such as Runtime Defined Entities,
Roles, Compute Policies, etc.

You can use the following example to start composing your Terraform configuration:

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

### Set up the Organizations

In this guide we will configure CSE v4.0 making use of two different [Organizations][org]:

- Solutions [Organization][org]: This [Organization][org] will host all provider-scoped items, such as the CSE Server.
  It should only be accessible to CSE administrators.
- Cluster [Organization][org]: This [Organization][org] will host the Kubernetes clusters for the users of this tenant to consume them.

This setup is just a proposal, you can have more cluster [organizations][org] or reuse an existing one.
In the sample HCL below you can find these two [Organizations][org] configured with no lease for vApps nor vApp Templates.
You can adjust it to fit with the requirements of your service:

```hcl
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
```

As mentioned, if you already have some [Organizations][org] available, you can fetch them with a data source instead:

```hcl
data "vcd_org" "solutions_organization" {
  name = "solutions_org"
}

data "vcd_org" "cluster_organization" {
  name = "cluster_org"
}
```

### Create the needed VM Sizing Policies

CSE v4.0 requires a specific set of [VM Sizing Policies][sizing] to be able to dimension the Kubernetes clusters.
You must create them with the HCL snippet below.

~> Apply this HCL as it is. In other words, names, descriptions and CPU/Memory specifications should **not** be modified.

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

You can of course create more [VM Sizing Policies][sizing], the ones specified above are just **the minimum required**
for CSE to work:

```hcl
# We can create as many policies as we need
resource "vcd_vm_sizing_policy" "other_policy" {
  name        = "Other policy"
  description = "Other useful Sizing Policy"
}
```


### Set up the VDCs

In this guide we will configure a single [VDC][vdc] per Organization:

- One VDC for the Solutions Organization.
- One VDC for the Cluster Organization.

This setup is just a proposal, you can have more [VDCs][vdc] or reuse some existing [VDCs][vdc].

In the sample HCL below you can find these two [VDCs][vdc]:

```hcl
# We fetch some required information like Provider VDC, Edge Clusters, etc
data "vcd_provider_vdc" "nsxt_pvdc" {
  name = "providerVdc1"
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
  network_pool_name = "NSX-T pool"
  provider_vdc_name = data.vcd_provider_vdc.nsxt_pvdc.name # Use a valid Provider VDC retrieved above
  edge_cluster_id   = data.vcd_nsxt_edge_cluster.cluster_edgecluster.id # Use a valid Edge Cluster retrieved above

  # You can tune these arguments to your fit your needs
  network_quota = 1000
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

  # Make sure you specify the VM Sizing Policies created previously
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

# The VDC that will host the CSE appliance and other provider-level items
resource "vcd_org_vdc" "solutions_vdc" {
  name        = "solutions_vdc"
  description = "Solutions VDC"
  org         = vcd_org.solutions_organization.name

  allocation_model  = "AllocationVApp" # You can use other models
  network_pool_name = "NSX-T pool"
  provider_vdc_name = data.vcd_provider_vdc.nsxt_pvdc.name # Use a valid Provider VDC retrieved above
  edge_cluster_id   = data.vcd_nsxt_edge_cluster.solutions_edgecluster.id # Use a valid Edge Cluster retrieved above

  # You can tune these arguments to your fit your needs
  network_quota = 1000
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

  # You can create more Sizing Policies and add them to the VDC:
  default_compute_policy_id = vcd_vm_sizing_policy.other_policy.id
  vm_sizing_policy_ids = [
    vcd_vm_sizing_policy.other_policy.id
  ]
}
```

### Create Catalogs and upload OVAs

We need to create some [Catalogs][catalog] to be able to store and retrieve CSE Server OVAs and maintain a repository of Kubernetes Template OVAs.
In this step, we will create two [Catalogs][catalog]:

- One Catalog in the Solutions Organization to upload CSE Server OVA for easy access.
- One shared Catalog in the Solutions Organization that will contain the Kubernetes Template OVAs.

Here's a sample HCL that can help you to achieve this setup:

```hcl
resource "vcd_catalog" "cse_catalog" {
  org  = vcd_org.solutions_organization.name # References the Solutions Organization created previously
  name = "cse_catalog"

  delete_force     = "true"
  delete_recursive = "true"
  
  # In this guide, everything is created from scratch, so it is needed to wait for the VDC to be available, so the
  # Catalog can be created.
  depends_on = [
    vcd_org_vdc.solutions_vdc
  ]
}

resource "vcd_catalog" "tkgm_catalog" {
  org  = vcd_org.solutions_organization.name # References the Solutions Organization created previously
  name = "tkgm_catalog"

  delete_force     = "true"
  delete_recursive = "true"
  
  # In this guide, everything is created from scratch, so it is needed to wait for the VDC to be available, so the
  # Catalog can be created.
  depends_on = [
    vcd_org_vdc.solutions_vdc
  ]
}

# We share the TKGm Catalog with the Cluster Organization created previously.
resource "vcd_catalog_access_control" "tkgm_catalog_ac" {
  org                  = vcd_org.solutions_organization.name # References the Solutions Organization created previously
  catalog_id           = vcd_catalog.tkgm_catalog.id
  shared_with_everyone = false
  shared_with {
    org_id       = vcd_org.cluster_organization.id # Shared with the Cluster Organization
    access_level = "ReadOnly"
  }
}
```

If you have already some [Catalogs][catalog] available, you can fetch them with a data source instead:

```hcl
data "vcd_catalog" "cse_catalog" {
  org  = vcd_org.solutions_organization.name
  name = "cse_catalog"
}

# This should be shared with the Cluster organization if it belongs to the Solutions Organization.
# If it is not shared, please look at the `vcd_catalog_access_control` from the snippet above.
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

CSE 4.0 requires some new [Runtime Defined Entity Interfaces][rde_interface] and [Runtime Defined Entity Types][rde_type]
to be created in VCD.

The required schemas to create [Runtime Defined Entity Types][rde_type] are present in this repository and are already
referenced in the snippet below.

It will create the following items in your VCD appliance:

- "VCDKEConfig" type, this is required to instantiate an entity that will contain the CSE Server required configuration.
- "CAPVCD" type, this is required to instantiate Kubernetes clusters.
- "VCDKEConfig" interface, required by the VCDKEConfig type.

~> Apply this HCL as it is. In other words, the [RDE Interfaces][rde_interface] and [RDE Types][rde_type] should have the specified `name`, `version`, `vendor` and `nss` to work with CSE v4.0

```hcl
resource "vcd_rde_interface" "vcdkeconfig_interface" {
  name    = "VCDKEConfig"
  version = "1.0.0"
  vendor  = "vmware"
  nss     = "VCDKEConfig"
}

# This one exists in VCD, so we just fetch it with a data source
data "vcd_rde_interface" "kubernetes_interface" {
  vendor  = "vmware"
  nss     = "k8s"
  version = "1.0.0"
}

resource "vcd_rde_type" "vcdkeconfig_type" {
  name          = "VCD-KE RDE Schema"
  nss           = "VCDKEConfig"
  version       = "1.0.0"
  schema_url    = "https://raw.githubusercontent.com/vmware/terraform-provider-vcd/main/examples/container-service-extension-4.0/schemas/vcdkeconfig-type-schema.json" 
  vendor        = "vmware"
  interface_ids = [vcd_rde_interface.vcdkeconfig_interface.id]
}

resource "vcd_rde_type" "capvcd_cluster_type" {
  name          = "CAPVCD Cluster"
  nss           = "capvcdCluster"
  version       = "1.1.0"
  schema_url    = "https://raw.githubusercontent.com/vmware/terraform-provider-vcd/main/examples/container-service-extension-4.0/schemas/capvcd-type-schema.json"
  vendor        = "vmware"
  interface_ids = [data.vcd_rde_interface.kubernetes_interface.id]
}
```

### Create CSE Admin Role

Until now, we have created everything as System administrator, but for security reasons we should create another [User][user] with
a specific subset of rights to run the CSE Server and other administrative tasks with it. In order to do that, we will create
a new provider-scoped [Role][role] called **"CSE Admin Role"**, a [User][user] with this [Role][role], and create an API token for it.

~> Apply this HCL as it is. In other words, the created [Role][role] should have the specified name, belong to System, and have the specified set of rights.

```hcl
resource "vcd_role" "cse_admin_role" {
  org = "System"
  name = "CSE Admin Role"
  description = "Used for administrative purposes"
  rights = [
    "API Tokens: Manage",
    "vmware:VCDKEConfig: Administrator Full access",
    "vmware:VCDKEConfig: Administrator View",
    "vmware:VCDKEConfig: Full Access",
    "vmware:VCDKEConfig: Modify",
    "vmware:VCDKEConfig: View",
    "vmware:capvcdCluster: Administrator Full access",
    "vmware:capvcdCluster: Administrator View",
    "vmware:capvcdCluster: Full Access",
    "vmware:capvcdCluster: Modify",
    "vmware:capvcdCluster: View"
  ]

  # This role depends on the created RDE Types as the rights are created after creation of the type.
  depends_on = [
    vcd_rde_type.vcdkeconfig_type,
    vcd_rde_type.capvcd_cluster_type,
  ]
}
```

After creating the [Role][role], we need to create the CSE Administrator user. **Please change the password accordingly**.

```hcl
resource "vcd_org_user" "cse_admin" {
  org      = vcd_org.solutions_organization.name # It should belong to the Solutions Org, as CSE Appliance will live there
  name     = "cse-admin"
  password = "******"
  role     = vcd_role.cse_admin_role.name
}
```

~> Take into account that you need to create an API token for this user using UI or a shell script, in order to configure the CSE Appliance later on.

### Create and publish a "Kubernetes Cluster Author" global role

Apart from the role to administrate the CSE Server created in previous step, we also need a [Global Role][global_role] for the Kubernetes clusters consumers.
It would be similar to the concept of "vApp Author" but for Kubernetes clusters. In order to create the [Global Role][global_role], first we need
to create a new [Rights Bundle][rights_bundle] and publish it to all the tenants:

~> Apply this HCL as it is. In other words, the created [Rights Bundle][rights_bundle] should have the specified name, description and have the specified set of rights.

```hcl
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
    "vmware:capvcdCluster: Administrator Full access",
    "vmware:capvcdCluster: Full Access",
    "vmware:capvcdCluster: Modify",
    "vmware:capvcdCluster: View",
    "vmware:capvcdCluster: Administrator View",
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
  publish_to_all_tenants = true
  
  # As we use rights created by the CAPVCD Type created previously, we need to depend on it
  depends_on = [
    vcd_rde_type.capvcd_cluster_type
  ]
}
```

Now we're in position to create the [Global Role][global_role]:

~> Apply this HCL as it is. In other words, the created [Global Role][global_role] should have the specified name, description and have the specified set of rights.

```hcl
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
    "vmware:capvcdCluster: Administrator Full access",
    "vmware:capvcdCluster: Full Access",
    "vmware:capvcdCluster: Modify",
    "vmware:capvcdCluster: View",
    "vmware:capvcdCluster: Administrator View",
    "vmware:tkgcluster: Full Access",
    "vmware:tkgcluster: Modify",
    "vmware:tkgcluster: View",
    "vmware:tkgcluster: Administrator View",
    "vmware:tkgcluster: Administrator Full access",
  ]

  publish_to_all_tenants = true

  # As we use rights created by the CAPVCD Type created previously, we need to depend on it
  depends_on = [
    vcd_rights_bundle.k8s_clusters_rights_bundle
  ]
}
```

### Set up networking

This step assumes that your VDC doesn't have any networking set up, so it builds a very basic setup that will make CSE v4.0 work.
If you have already networking in place, please skip this step and configure CSE server with an existing network.

#### Provider Gateways

The first step is setting up the [Provider Gateways][provider_gateway] as System administrator, we will set up one per Organization.
These gateways will route public IPs to the [Edge Gateways][edge_gateway] that we will create in the following step.

```hcl
data "vcd_nsxt_manager" "cse_nsxt_manager" {
  name = "nsxtManager1"
}

data "vcd_nsxt_tier0_router" "solutions_tier0_router" {
  name            = "VCD T0 edgeCluster1"
  nsxt_manager_id = data.vcd_nsxt_manager.cse_nsxt_manager.id
}

resource "vcd_external_network_v2" "solutions_tier0" {
  name = "solutions_tier0"

  nsxt_network {
    nsxt_manager_id      = data.vcd_nsxt_manager.cse_nsxt_manager.id
    nsxt_tier0_router_id = data.vcd_nsxt_tier0_router.solutions_tier0_router.id
  }

  ip_scope {
    gateway       = "2.7.4.1"
    prefix_length = 24

    static_ip_pool {
      start_address = "2.7.4.1"
      end_address   = "2.7.4.254"
    }
  }
}

data "vcd_nsxt_tier0_router" "cluster_tier0_router" {
  name            = "VCD T0 edgeCluster2"
  nsxt_manager_id = data.vcd_nsxt_manager.cse_nsxt_manager.id
}

resource "vcd_external_network_v2" "cluster_tier0" {
  name = "cluster_tier0"

  nsxt_network {
    nsxt_manager_id      = data.vcd_nsxt_manager.cse_nsxt_manager.id
    nsxt_tier0_router_id = data.vcd_nsxt_tier0_router.cluster_tier0_router.id
  }

  ip_scope {
    gateway       = "2.7.5.1"
    prefix_length = 24

    static_ip_pool {
      start_address = "2.7.5.1"
      end_address   = "2.7.5.254"
    }
  }
}
```

#### Edge Gateways

```hcl
resource "vcd_nsxt_edgegateway" "solutions_edgegateway" {
  org      = vcd_org.solutions_organization.name
  owner_id = vcd_org_vdc.solutions_vdc.id

  name                      = "solutions_edgegateway"
  external_network_id       = vcd_external_network_v2.solutions_tier0.id
  dedicate_external_network = true

  subnet {
    gateway       = var.gateway_ip
    prefix_length = 19
    primary_ip    = var.solutions_static_ips[0][0] # The first IP provided will be assigned as gateway IP

    dynamic "allocated_ips" {
      for_each = var.solutions_static_ips
      iterator = ip
      content {
        start_address = ip.value[0]
        end_address   = ip.value[1]
      }
    }
  }

  depends_on = [vcd_org_vdc.solutions_vdc]
}

resource "vcd_nsxt_edgegateway" "cluster_edgegateway" {
  org      = vcd_org.cluster_organization.name
  owner_id = vcd_org_vdc.cluster_vdc.id

  name                      = "cluster_edgegateway"
  external_network_id       = vcd_external_network_v2.cluster_tier0.id
  dedicate_external_network = true

  subnet {
    gateway       = var.gateway_ip
    prefix_length = 19
    primary_ip    = var.cluster_static_ips[0][0] # The first IP provided will be assigned as gateway IP

    dynamic "allocated_ips" {
      for_each = var.cluster_static_ips
      iterator = ip
      content {
        start_address = ip.value[0]
        end_address   = ip.value[1]
      }
    }
  }

  depends_on = [vcd_org_vdc.cluster_vdc]
}
```

#### Advanced Load Balancer configuration

-> To learn more about the Advanced Load Balancer capabilities, please read the Terraform guide [here](/providers/vmware/vcd/latest/docs/guides/nsxt_alb).

The following snippet

```hcl
data "vcd_nsxt_alb_controller" "cse_avi_controller" {
  name = "aviController1"
}

data "vcd_nsxt_alb_importable_cloud" "cse_importable_cloud" {
  name          = var.avi_importable_cloud
  controller_id = data.vcd_nsxt_alb_controller.cse_avi_controller.id
}

resource "vcd_nsxt_alb_cloud" "cse_nsxt_alb_cloud" {
  name = "cse_nsxt_alb_cloud"

  controller_id       = data.vcd_nsxt_alb_controller.cse_avi_controller.id
  importable_cloud_id = data.vcd_nsxt_alb_importable_cloud.cse_importable_cloud.id
  network_pool_id     = data.vcd_nsxt_alb_importable_cloud.cse_importable_cloud.network_pool_id
}

resource "vcd_nsxt_alb_service_engine_group" "cse_alb_seg" {
  name                                 = "cse_alb_seg"
  alb_cloud_id                         = vcd_nsxt_alb_cloud.cse_nsxt_alb_cloud.id
  importable_service_engine_group_name = "Default-Group"
  reservation_model                    = "SHARED"
}

## ALB for solutions edge gateway

resource "vcd_nsxt_alb_settings" "solutions_alb_settings" {
  org             = vcd_org.solutions_organization.name
  edge_gateway_id = vcd_nsxt_edgegateway.solutions_edgegateway.id
  is_active       = true

  # This dependency is required to make sure that provider part of operations is done
  depends_on = [vcd_nsxt_alb_service_engine_group.cse_alb_seg]
}

resource "vcd_nsxt_alb_edgegateway_service_engine_group" "solutions_assignment" {
  org                       = vcd_org.solutions_organization.name
  edge_gateway_id           = vcd_nsxt_alb_settings.solutions_alb_settings.edge_gateway_id
  service_engine_group_id   = vcd_nsxt_alb_service_engine_group.cse_alb_seg.id
  reserved_virtual_services = 50
  max_virtual_services      = 50
}

resource "vcd_nsxt_alb_edgegateway_service_engine_group" "cluster_assignment" {
  org                       = vcd_org.cluster_organization.name
  edge_gateway_id           = vcd_nsxt_alb_settings.cluster_alb_settings.edge_gateway_id
  service_engine_group_id   = vcd_nsxt_alb_service_engine_group.cse_alb_seg.id
  reserved_virtual_services = 50
  max_virtual_services      = 50
}

resource "vcd_nsxt_alb_settings" "cluster_alb_settings" {
  org             = vcd_org.cluster_organization.name
  edge_gateway_id = vcd_nsxt_edgegateway.cluster_edgegateway.id
  is_active       = true

  depends_on = [vcd_nsxt_alb_service_engine_group.cse_alb_seg]
}
```

#### Organization networks

```hcl
resource "vcd_network_routed_v2" "solutions_routed_network" {
  org         = vcd_org.solutions_organization.name
  name        = "solutions_routed_network"
  description = "Solutions routed network"

  edge_gateway_id = vcd_nsxt_edgegateway.solutions_edgegateway.id

  gateway       = "192.168.0.1"
  prefix_length = 24

  static_ip_pool {
    start_address = "192.168.0.2"
    end_address   = "192.168.0.10"
  }

  dns1       = "10.84.54.20"
  dns2       = "1.1.1.1"
  dns_suffix = "eng.vmware.com"
}

resource "vcd_network_routed_v2" "cluster_routed_network" {
  org         = vcd_org.cluster_organization.name
  name        = "cluster_net_routed"
  description = "Routed network for the K8s clusters"

  edge_gateway_id = vcd_nsxt_edgegateway.cluster_edgegateway.id

  gateway       = "10.0.0.1"
  prefix_length = 16

  static_ip_pool {
    start_address = "10.0.0.2"
    end_address   = "10.0.255.254"
  }

  dns1       = "10.84.54.20"
  dns2       = "1.1.1.1"
  dns_suffix = "eng.vmware.com"
}

resource "vcd_nsxt_route_advertisement" "solutions_routing_advertisement" {
  edge_gateway_id = vcd_nsxt_edgegateway.solutions_edgegateway.id
  enabled         = true
  subnets         = ["192.168.0.0/24"]
}

resource "vcd_nsxt_route_advertisement" "cluster_routing_advertisement" {
  edge_gateway_id = vcd_nsxt_edgegateway.cluster_edgegateway.id
  enabled         = true
  subnets         = ["10.0.0.0/16"]
}

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

resource "vcd_nsxt_firewall" "cluster_firewall" {
  org             = vcd_org.cluster_organization.name
  edge_gateway_id = vcd_nsxt_edgegateway.cluster_edgegateway.id

  rule {
    action      = "ALLOW"
    name        = "Allow all traffic"
    direction   = "IN_OUT"
    ip_protocol = "IPV4_IPV6"
  }
}
```

### Configure CSE server

```hcl
# We read the entity JSON of the VCDKEConfig as template as some fields are references to Terraform resources.
# The inputs are taken from UI.
data "template_file" "vcdkeconfig_instance_template" {
  template = file("${path.module}/entities/vcdkeconfig-template.json")
  vars = {
    capvcd_version                  = var.capvcd_version
    cpi_version                     = var.cpi_version
    csi_version                     = var.csi_version
    github_personal_access_token    = var.github_personal_access_token
    bootstrap_cluster_sizing_policy = vcd_vm_sizing_policy.tkg_s.name
    no_proxy                        = var.no_proxy
    http_proxy                      = var.http_proxy
    https_proxy                     = var.https_proxy
    syslog_host                     = var.syslog_host
    syslog_port                     = var.syslog_port
  }
}

resource "vcd_rde" "vcdkeconfig_instance" {
  org              = "System"
  name             = "vcdKeConfig"
  rde_type_vendor  = vcd_rde_type.vcdkeconfig_type.vendor
  rde_type_nss     = vcd_rde_type.vcdkeconfig_type.nss
  rde_type_version = vcd_rde_type.vcdkeconfig_type.version
  resolve          = true
  input_entity     = data.template_file.vcdkeconfig_instance_template.rendered
}
```

### Deploy CSE server

```hcl

resource "vcd_vapp" "cse_appliance_vapp" {
  org  = vcd_org.solutions_organization.name
  vdc  = vcd_org_vdc.solutions_vdc.name
  name = "CSE Appliance vApp"

  lease {
    runtime_lease_in_sec = 0
    storage_lease_in_sec = 0
  }
}

resource "vcd_vapp_org_network" "cse_appliance_network" {
  org = vcd_org.solutions_organization.name
  vdc = vcd_org_vdc.solutions_vdc.name

  vapp_name        = vcd_vapp.cse_appliance_vapp.name
  org_network_name = vcd_network_routed_v2.solutions_routed_network.name
}

resource "vcd_vapp_vm" "cse_appliance_vm" {
  org = vcd_org.solutions_organization.name
  vdc = vcd_org_vdc.solutions_vdc.name

  vapp_name = vcd_vapp.cse_appliance_vapp.name
  name      = "CSE Appliance VM"

  vapp_template_id = vcd_catalog_vapp_template.cse_ova.id

  network {
    type               = "org"
    name               = vcd_vapp_org_network.cse_appliance_network.org_network_name
    ip_allocation_mode = "POOL"
  }

  guest_properties = {

    # VCD host
    "cse.vcdHost" = replace(var.vcd_api_endpoint, "/api", "")

    # CSE service account's org
    "cse.AppOrg" = vcd_org.solutions_organization.name

    # CSE service account's Access Token
    "cse.vcdRefreshToken" = var.service_account_access_token

    # CSE service account's username
    "cse.vcdUsername" = var.service_account_username

    # CSE service vApp's org
    "cse.userOrg" = "System"
  }

  customization {
    force                      = false
    enabled                    = true
    allow_local_admin_password = true
    auto_generate_password     = false
    admin_password             = var.cse_vm_password # In the guide it says to auto generate, but for simplicity it is hardcoded
  }

  depends_on = [
    vcd_rde.vcdkeconfig_instance
  ]
}
```

## CSE upgrade process

Coming soon

## Cluster operations

Coming soon

### Create a cluster

Coming soon

```hcl
data "template_file" "k8s_cluster_yaml_template" {
  template = file("${path.module}/capvcd_templates/v1.22.9.yaml")
  vars = {
    cluster_name = var.k8s_cluster_name

    vcd_url     = replace(var.vcd_api_endpoint, "/api", "")
    org         = vcd_org.cluster_organization.name
    vdc         = vcd_org_vdc.cluster_vdc.name
    org_network = vcd_network_routed_v2.cluster_routed_network.name

    base64_username  = base64encode(var.k8s_cluster_user)
    base64_api_token = base64encode(var.k8s_cluster_api_token)

    ssh_public_key                   = ""
    control_plane_machine_count      = 1
    control_plane_sizing_policy_name = vcd_vm_sizing_policy.default_policy.name
    control_plane_sizing_policy_name = ""
    control_plane_placement_policy   = ""
    control_plane_storage_profile    = ""
    control_plane_disk_size          = "20Gi"

    worker_storage_policy   = ""
    worker_sizing_policy = vcd_vm_sizing_policy.default_policy.name
    worker_placement_policy = ""
    worker_machine_count    = 1
    worker_disk_size        = "20Gi"

    catalog_name = vcd_catalog.cse_catalog.name
    tkgm_ova     = replace(var.tkgm_ova_name, ".ova", "")

    pods_cidr     = "100.96.0.0/11"
    services_cidr = "100.64.0.0/13"
  }
}

# sample_cluster.yaml file should convert \n into \\n and " into \" first
data "template_file" "rde_k8s_cluster_instance_template" {
  template = file("${path.module}/entities/k8s_cluster.json")
  vars = {
    vcd_url   = replace(var.vcd_api_endpoint, "/api", "")
    name      = var.k8s_cluster_name
    org       = vcd_org.cluster_organization.name
    vdc       = vcd_org_vdc.cluster_vdc.name
    capi_yaml = replace(replace(data.template_file.k8s_cluster_yaml_template.rendered, "\n", "\\n"), "\"", "\\\"")

    delete                = false # Make this true to delete the cluster
    resolve_on_destroy    = false # Make this true to forcefully delete the cluster
    auto_repair_on_errors = false
  }
}

resource "vcd_rde" "k8s_cluster_instance" {
  org              = vcd_org.cluster_organization.name
  name             = var.k8s_cluster_name
  rde_type_vendor  = vcd_rde_type.capvcd_cluster_type.vendor
  rde_type_nss     = vcd_rde_type.capvcd_cluster_type.nss
  rde_type_version = vcd_rde_type.capvcd_cluster_type.version
  resolve          = false # MUST be false as it is resolved by CSE appliance
  resolve_on_destroy     = true  # MUST be true as it won't be resolved by Terraform
  input_entity     = data.template_file.rde_k8s_cluster_instance_template.rendered

  depends_on = [
    vcd_vapp_vm.cse_appliance_vm, vcd_catalog_vapp_template.tkgm_ova
  ]
}

output "computed_k8s_cluster_id" {
  value = vcd_rde.k8s_cluster_instance.id
}

output "computed_k8s_cluster_capvcdyaml" {
  value = jsondecode(vcd_rde.k8s_cluster_instance.computed_entity)["spec"]["capiYaml"]
}
```

### Retrieve a cluster Kubeconfig

Coming soon

```hcl
# output "kubeconfig" {  
#   value = jsondecode(vcd_rde.k8s_cluster_instance.computed_entity)["status"]["capvcd"]["private"]["kubeConfig"]
# }
```

### Upgrade a cluster

Coming soon

### Delete a cluster

Coming soon

~> Don't remove the resource from HCL as this will trigger a destroy operation, which will leave things behind in VCD.
Follow the mentioned steps instead.

## Uninstall CSE

Before uninstalling CSE, make sure you perform an update operation to mark all clusters for deletion.

~> Don't remove the K8s cluster resources from HCL as this will trigger a destroy operation, which will leave things behind in VCD.
Follow the mentioned steps instead.

Once all clusters are removed in the background by CSE Server, you may destroy the remaining infrastructure.


[org]: </providers/vmware/vcd/latest/docs/resources/org> (vcd_org)
[sizing]: </providers/vmware/vcd/latest/docs/resources/vm_sizing_policy> (vcd_vm_sizing_policy)
[vdc]: </providers/vmware/vcd/latest/docs/resources/org_vdc> (vcd_org_vdc)
[catalog]: </providers/vmware/vcd/latest/docs/resources/catalog> (vcd_catalog)
[rde_interface]: </providers/vmware/vcd/latest/docs/resources/rde_interface> (vcd_rde_interface)
[rde_type]: </providers/vmware/vcd/latest/docs/resources/rde_type> (vcd_rde_type)
[role]: </providers/vmware/vcd/latest/docs/resources/role> (vcd_role)
[user]: </providers/vmware/vcd/latest/docs/resources/user> (vcd_user)
[rights_bundle]: </providers/vmware/vcd/latest/docs/resources/rights_bundle> (vcd_rights_bundle)
[global_role]: </providers/vmware/vcd/latest/docs/resources/global_role> (vcd_global_role)
[provider_gateway]: </providers/vmware/vcd/latest/docs/resources/external_network_v2> (vcd_external_network_v2)
[edge_gateway]: </providers/vmware/vcd/latest/docs/resources/nsxt_edgegateway> (vcd_nsxt_edgegateway)