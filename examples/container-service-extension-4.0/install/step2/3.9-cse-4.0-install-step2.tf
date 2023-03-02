# ------------------------------------------------------------------------------------------------------------
# CSE 4.0 installation example HCL:
#
# * This example allows to create TKGm clusters using Terraform once applied. Read the guide present in
#   https://registry.terraform.io/providers/vmware/vcd/latest/docs/guides/container_service_extension_4.0 for more
#   information.
#
# * Some resources and data sources from this HCL are run as System administrator, as it involves creating provider
#   elements such as Organizations, VDCs or Provider Gateways. CSE Server is run with a more limited set of rights
#   with a System-scoped user and an API token managed by the `refresh_token.sh` script.
#
# * Please customize the values present in this file to your needs. Also check `terraform.tfvars.example`
#   for customisation.
# ------------------------------------------------------------------------------------------------------------

# VCD Provider configuration. It must be at least v3.9.0 and configured with a System administrator account.
# This is needed to build the minimum setup for CSE v4.0 to work, like Organizations, VDCs, Provider Gateways, etc.

terraform {
  required_providers {
    vcd = {
      source  = "vmware/vcd"
      version = ">= 3.9"
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
  logging_file         = "cse_${var.administrator_user}.log"
}

# In this example HCL we will create two Organizations:
# - The Solutions Organization will host the CSE Server and will be used for administrative tasks.
# - The Cluster Organization will host Kubernetes clusters and will be accessed by the tenants.
#
# Both organizations are created with unlimited lease, as CSE Server must be always up and running.

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

# The VM Sizing Policies defined below must be created as they are. Nothing should be changed here.

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

# This section will create one VDC per organization. To create the VDCs we need to fetch some elements like
# Provider VDC, Edge Clusters, etc.

data "vcd_provider_vdc" "nsxt_pvdc" {
  name = var.provider_vdc_name
}

data "vcd_nsxt_edge_cluster" "cluster_edgecluster" {
  org             = vcd_org.cluster_organization.name
  provider_vdc_id = data.vcd_provider_vdc.nsxt_pvdc.id
  name            = var.nsxt_edge_cluster_name
}

# The VDC that will host the Kubernetes clusters
resource "vcd_org_vdc" "cluster_vdc" {
  name        = "cluster_vdc"
  description = "Cluster VDC"
  org         = vcd_org.cluster_organization.name

  allocation_model  = "AllocationVApp" # You can use other models
  network_pool_name = var.network_pool_name
  provider_vdc_name = data.vcd_provider_vdc.nsxt_pvdc.name
  edge_cluster_id   = data.vcd_nsxt_edge_cluster.cluster_edgecluster.id

  # You can tune these arguments to your fit your needs
  network_quota = 50
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

  # Make sure you specify the required VM Sizing Policies
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
  edge_cluster_id   = data.vcd_nsxt_edge_cluster.cluster_edgecluster.id

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

# We upload a minimum set of OVAs for CSE to work. Read the official documentation to check
# where to find the OVAs:
# https://docs.vmware.com/en/VMware-Cloud-Director-Container-Service-Extension/4.0/VMware-Cloud-Director-Container-Service-Extension-Install-provider-4.0/GUID-519D73E8-5459-439E-AB92-83076F556E53.html#GUID-519D73E8-5459-439E-AB92-83076F556E53

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

# In the following section we create the required RDE Interfaces and RDE Types.

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
  schema_url    = "https://raw.githubusercontent.com/adambarreiro/terraform-provider-vcd/add-cse40-guide/examples/container-service-extension-4.0/schemas/vcdkeconfig-type-schema.json"
  vendor        = "vmware"
  interface_ids = [vcd_rde_interface.vcdkeconfig_interface.id]
}

resource "vcd_rde_type" "capvcd_cluster_type" {
  name          = "CAPVCD Cluster"
  nss           = "capvcdCluster"
  version       = "1.1.0"
  schema_url    = "https://raw.githubusercontent.com/adambarreiro/terraform-provider-vcd/add-cse40-guide/examples/container-service-extension-4.0/schemas/capvcd-type-schema.json"
  vendor        = "vmware"
  interface_ids = [data.vcd_rde_interface.kubernetes_interface.id]
}

resource "vcd_role" "cse_admin_role" {
  org         = "System"
  name        = "CSE Admin Role"
  description = "Used for administrative purposes"
  rights = [
    "API Tokens: Manage",
    "${vcd_rde_type.vcdkeconfig_type.vendor}:${vcd_rde_type.vcdkeconfig_type.nss}: Administrator Full access",
    "${vcd_rde_type.vcdkeconfig_type.vendor}:${vcd_rde_type.vcdkeconfig_type.nss}: Administrator View",
    "${vcd_rde_type.vcdkeconfig_type.vendor}:${vcd_rde_type.vcdkeconfig_type.nss}: Full Access",
    "${vcd_rde_type.vcdkeconfig_type.vendor}:${vcd_rde_type.vcdkeconfig_type.nss}: Modify",
    "${vcd_rde_type.vcdkeconfig_type.vendor}:${vcd_rde_type.vcdkeconfig_type.nss}: View",
    "${vcd_rde_type.capvcd_cluster_type.vendor}:${vcd_rde_type.capvcd_cluster_type.nss}: Administrator Full access",
    "${vcd_rde_type.capvcd_cluster_type.vendor}:${vcd_rde_type.capvcd_cluster_type.nss}: Administrator View",
    "${vcd_rde_type.capvcd_cluster_type.vendor}:${vcd_rde_type.capvcd_cluster_type.nss}: Full Access",
    "${vcd_rde_type.capvcd_cluster_type.vendor}:${vcd_rde_type.capvcd_cluster_type.nss}: Modify",
    "${vcd_rde_type.capvcd_cluster_type.vendor}:${vcd_rde_type.capvcd_cluster_type.nss}: View"
  ]
}

resource "vcd_org_user" "cse_admin" {
  org      = "System"
  name     = var.cse_admin_user
  password = var.cse_admin_password
  role     = vcd_role.cse_admin_role.name
}

resource "null_resource" "cse_api_token_script" {
  triggers = {
    # Trigger the installation only if the attributes that force a replacement are changed. In other words, this only needs
    # to be triggered if the user is deleted and re-created.
    config_has_changed = join("_", [
      vcd_org_user.cse_admin.name,
      vcd_org_user.cse_admin.org,
      vcd_org_user.cse_admin.is_external,
    ])
  }

  provisioner "local-exec" {
    when = create
    command = "VCD_PASSWORD=${var.cse_admin_password} ./refresh_token.sh create ${var.vcd_url} ${var.cse_admin_user} System ${var.cse_admin_user}"
  }

  provisioner "local-exec" {
    when = destroy
    command = "VCD_PASSWORD=${var.cse_admin_password} ./refresh_token.sh destroy ${var.vcd_url} ${var.cse_admin_user} System ${var.cse_admin_user}"
  }
}

data "local_sensitive_file" "cse_api_token" {
  filename = "${path.module}/.${vcd_org_user.cse_admin.name}_${vcd_org_user.cse_admin.name}"
}

