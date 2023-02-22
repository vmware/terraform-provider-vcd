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

In this guide we will configure CSE v4.0 making use of two different [Organizations][r_org]:

- Solutions [Organization][r_org]: This [Organization][r_org] will host all provider-scoped items, such as the CSE Appliance vApp.
- Cluster [Organization][r_org]: This [Organization][r_org] will host the Kubernetes clusters for the tenants to use.

This setup is just a proposal, you can have more cluster [organizations][r_org] or reuse an existing [organization][d_org].
In the sample HCL below you can find these two [Organizations][r_org] configured with no lease for vApps nor vApp Templates.
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

As mentioned, if you already have some [Organizations][d_org] available, you can fetch them with a data source instead:

```hcl
data "vcd_org" "solutions_organization" {
  name = "solutions_org"
}

data "vcd_org" "cluster_organization" {
  name = "cluster_org"
}
```

### Create the needed Sizing Policies

CSE v4.0 requires a specific set of [Sizing Policies][r_sizing] to be able to dimension the Kubernetes clusters.
You must create them with the HCL snippet below.

~> Apply this HCL as it is. In other words, the names, descriptions and CPU/Memory specifications should **not** be modified.

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

You can of course create more [Sizing policies][r_sizing], the ones specified above are just **the minimum required**
for CSE to work.

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

This is a provider-level role in 'System' org.
VMware Cloud Director Container Service Extension Server v4 uses this role to process cluster operations

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

  depends_on = [
    vcd_rde_type.vcd_ke_config_type,
    vcd_rde_type.capvcd_cluster_type,
  ]
}
```

Create a user:

```hcl
resource "vcd_org_user" "cse_admin" {
  org      = vcd_org.solutions_organization.name
  name     = "cse-admin"
  password = "ca$hc0w"
  role     = vcd_role.cse_admin_role.name
}
```

~> You need to create an API token for this user

### Create and Publish 'Kubernetes Clusters Rights Bundle'

Create and publish:

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
  depends_on = [
    vcd_rde_type.capvcd_cluster_type
  ]
}
```

### Create and Publish 'Kubernetes Cluster Author' global role

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

  depends_on = [
    vcd_rights_bundle.k8s_clusters_rights_bundle
  ]
}
```

### Set up networking

This step assumes that your VDC doesn't have any networking set up. If you have already networking in place, please
skip this step.


### Configure CSE server

```hcl
# We read the entity JSON of the VCDKEConfig as template as some fields are references to Terraform resources.
# The inputs are taken from UI.
data "template_file" "vcd_ke_config_instance_template" {
  template = file("${path.module}/entities/vcdkeconfig.json")
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

resource "vcd_rde" "vcd_ke_config_instance" {
  # org         = "System"
  name             = "vcdKeConfig"
  rde_type_vendor  = vcd_rde_type.vcd_ke_config_type.vendor
  rde_type_nss     = vcd_rde_type.vcd_ke_config_type.nss
  rde_type_version = vcd_rde_type.vcd_ke_config_type.version
  resolve          = true
  input_entity     = data.template_file.vcd_ke_config_instance_template.rendered
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
}
```

## CSE upgrade process

## Cluster operations

### Create a cluster

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
    force_delete          = false # Make this true to forcefully delete the cluster
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
  force_delete     = true  # MUST be true as it won't be resolved by Terraform
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

### Retrieve its kubeconfig

```hcl
# output "kubeconfig" {  
#   value = jsondecode(vcd_rde.k8s_cluster_instance.computed_entity)["status"]["capvcd"]["private"]["kubeConfig"]
# }
```

### Upgrade a cluster

### Delete a cluster

~> Don't remove the resource from HCL as this will trigger a destroy operation, which will leave things behind in VCD.
Follow the mentioned steps instead.

## Uninstall CSE

Before uninstalling CSE, make sure you perform an update operation to mark all clusters for deletion.

~> Don't remove the K8s cluster resources from HCL as this will trigger a destroy operation, which will leave things behind in VCD.
Follow the mentioned steps instead.

Once all clusters are removed in the background by CSE Server, you may destroy the remaining infrastructure.


[r_org]: </providers/vmware/vcd/latest/docs/resources/org> (vcd_org)
[d_org]: </providers/vmware/vcd/latest/docs/data-sources/org> (vcd_org)
[r_sizing]: </providers/vmware/vcd/latest/docs/resources/vm_sizing_policy> (vcd_vm_sizing_policy)