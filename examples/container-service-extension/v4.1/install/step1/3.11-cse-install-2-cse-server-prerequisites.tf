# ------------------------------------------------------------------------------------------------------------
# CSE v4.1 installation, step 1:
#
# * Please read the guide at https://registry.terraform.io/providers/vmware/vcd/latest/docs/guides/container_service_extension_4_x_install
#   before applying this configuration.
#
# * The installation process is split into two steps as the first one creates a CSE admin user that needs to be
#   used in a "provider" block in the second one.
#
# * This file contains the same resources created by the "Configure Settings for CSE Server > Set Up Prerequisites" step in the
#   UI wizard.
#
# * Rename "terraform.tfvars.example" to "terraform.tfvars" and adapt the values to your needs.
#   Other than that, this snippet should be applied as it is.
#   You can check the comments on each resource/data source for more help and context.
# ------------------------------------------------------------------------------------------------------------

# This is the RDE Interface required to create the "VCDKEConfig" RDE Type.
# This should not be changed.
resource "vcd_rde_interface" "vcdkeconfig_interface" {
  vendor  = "vmware"
  nss     = "VCDKEConfig"
  version = "1.0.0"
  name    = "VCDKEConfig"
}

# This resource will manage the "VCDKEConfig" RDE Type required to instantiate the CSE Server configuration.
# The schema URL points to the JSON schema hosted in the terraform-provider-vcd repository.
# This should not be changed.
resource "vcd_rde_type" "vcdkeconfig_type" {
  vendor        = "vmware"
  nss           = "VCDKEConfig"
  version       = "1.1.0"
  name          = "VCD-KE RDE Schema"
  schema_url    = "https://raw.githubusercontent.com/vmware/terraform-provider-vcd/main/examples/container-service-extension/v4.1/schemas/vcdkeconfig-type-schema-v1.1.0.json"
  interface_ids = [vcd_rde_interface.vcdkeconfig_interface.id]
}

# This RDE Interface exists in VCD, so it must be fetched with a RDE Interface data source. This RDE Interface is used to be
# able to create the "capvcdCluster" RDE Type.
# This should not be changed.
data "vcd_rde_interface" "kubernetes_interface" {
  vendor  = "vmware"
  nss     = "k8s"
  version = "1.0.0"
}

# This is the interface required to create the "CAPVCD" Runtime Defined Entity Type.
# This should not be changed.
resource "vcd_rde_interface" "cse_interface" {
  vendor  = "cse"
  nss     = "capvcd"
  version = "1.0.0"
  name    = "cseInterface"
}

# This RDE Interface behavior is required to be able to obtain the Kubeconfig and other important information.
# This should not be changed.
resource "vcd_rde_interface_behavior" "capvcd_behavior" {
  rde_interface_id = vcd_rde_interface.cse_interface.id
  name             = "getFullEntity"
  execution = {
    "type" : "noop"
    "id" : "getFullEntity"
  }
}

# This RDE Interface will create the "capvcdCluster" RDE Type required to create Kubernetes clusters.
# The schema URL points to the JSON schema hosted in the terraform-provider-vcd repository.
# This should not be changed.
resource "vcd_rde_type" "capvcdcluster_type" {
  vendor        = "vmware"
  nss           = "capvcdCluster"
  version       = "1.2.0"
  name          = "CAPVCD Cluster"
  schema_url    = "https://raw.githubusercontent.com/vmware/terraform-provider-vcd/main/examples/container-service-extension/v4.1/schemas/capvcd-type-schema-v1.2.0.json"
  interface_ids = [data.vcd_rde_interface.kubernetes_interface.id]

  depends_on = [vcd_rde_interface_behavior.capvcd_behavior] # Interface Behaviors must be created before any RDE Type
}

# Access Level for the CAPVCD Type Behavior
# This should not be changed.
resource "vcd_rde_type_behavior_acl" "capvcd_behavior_acl" {
  rde_type_id      = vcd_rde_type.capvcdcluster_type.id
  behavior_id      = vcd_rde_interface_behavior.capvcd_behavior.id
  access_level_ids = ["urn:vcloud:accessLevel:FullControl"]
}

# This role is having only the minimum set of rights required for the CSE Server to function.
# It is created in the "System" provider organization scope.
# This should not be changed.
resource "vcd_role" "cse_admin_role" {
  org         = var.administrator_org
  name        = "CSE Admin Role"
  description = "Used for administrative purposes"
  rights = concat([
    "API Tokens: Manage",
    "${vcd_rde_type.vcdkeconfig_type.vendor}:${vcd_rde_type.vcdkeconfig_type.nss}: Administrator Full access",
    "${vcd_rde_type.vcdkeconfig_type.vendor}:${vcd_rde_type.vcdkeconfig_type.nss}: Administrator View",
    "${vcd_rde_type.vcdkeconfig_type.vendor}:${vcd_rde_type.vcdkeconfig_type.nss}: Full Access",
    "${vcd_rde_type.vcdkeconfig_type.vendor}:${vcd_rde_type.vcdkeconfig_type.nss}: Modify",
    "${vcd_rde_type.vcdkeconfig_type.vendor}:${vcd_rde_type.vcdkeconfig_type.nss}: View",
    "${vcd_rde_type.capvcdcluster_type.vendor}:${vcd_rde_type.capvcdcluster_type.nss}: Administrator Full access",
    "${vcd_rde_type.capvcdcluster_type.vendor}:${vcd_rde_type.capvcdcluster_type.nss}: Administrator View",
    "${vcd_rde_type.capvcdcluster_type.vendor}:${vcd_rde_type.capvcdcluster_type.nss}: Full Access",
    "${vcd_rde_type.capvcdcluster_type.vendor}:${vcd_rde_type.capvcdcluster_type.nss}: Modify",
    "${vcd_rde_type.capvcdcluster_type.vendor}:${vcd_rde_type.capvcdcluster_type.nss}: View"
  ], data.vcd_version.gte_1051.matches_condition ? ["Organization: Traversal"] : [])
}

# This will allow to have a user with a limited set of rights that can access the Provider area of VCD.
# This user will be used by the CSE Server, with an API token that must be created in Step 2.
# This should not be changed.
resource "vcd_org_user" "cse_admin" {
  org      = var.administrator_org
  name     = var.cse_admin_username
  password = var.cse_admin_password
  role     = vcd_role.cse_admin_role.name
}

# This resource manages the Rights Bundle required by tenants to create and consume Kubernetes clusters.
# This should not be changed.
resource "vcd_rights_bundle" "k8s_clusters_rights_bundle" {
  name        = "Kubernetes Clusters Rights Bundle"
  description = "Rights bundle with required rights for managing Kubernetes clusters"
  rights = [
    "API Tokens: Manage",
    "Access All Organization VDCs",
    "Catalog: View Published Catalogs",
    "Certificate Library: Manage",
    "Certificate Library: View",
    "General: Administrator View",
    "Organization vDC Gateway: Configure Load Balancer",
    "Organization vDC Gateway: Configure NAT",
    "Organization vDC Gateway: View Load Balancer",
    "Organization vDC Gateway: View NAT",
    "Organization vDC Gateway: View",
    "Organization vDC Named Disk: Create",
    "Organization vDC Named Disk: Edit Properties",
    "Organization vDC Named Disk: View Properties",
    "Organization vDC Shared Named Disk: Create",
    "vApp: Allow All Extra Config",
    "${vcd_rde_type.vcdkeconfig_type.vendor}:${vcd_rde_type.vcdkeconfig_type.nss}: View",
    "${vcd_rde_type.capvcdcluster_type.vendor}:${vcd_rde_type.capvcdcluster_type.nss}: Administrator Full access",
    "${vcd_rde_type.capvcdcluster_type.vendor}:${vcd_rde_type.capvcdcluster_type.nss}: Full Access",
    "${vcd_rde_type.capvcdcluster_type.vendor}:${vcd_rde_type.capvcdcluster_type.nss}: Modify",
    "${vcd_rde_type.capvcdcluster_type.vendor}:${vcd_rde_type.capvcdcluster_type.nss}: View",
    "${vcd_rde_type.capvcdcluster_type.vendor}:${vcd_rde_type.capvcdcluster_type.nss}: Administrator View",
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
# This should not be changed.
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
    "Organization vDC Disk: View IOPS",
    "Organization vDC Gateway: Configure Load Balancer",
    "Organization vDC Gateway: Configure NAT",
    "Organization vDC Gateway: View",
    "Organization vDC Gateway: View Load Balancer",
    "Organization vDC Gateway: View NAT",
    "Organization vDC Named Disk: Create",
    "Organization vDC Named Disk: Delete",
    "Organization vDC Named Disk: Edit Properties",
    "Organization vDC Named Disk: View Encryption Status",
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
    "vApp: Edit VM Compute Policy",
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
    "vApp: View VM and VM's Disks Encryption Status",
    "vApp: View VM metrics",
    "${vcd_rde_type.vcdkeconfig_type.vendor}:${vcd_rde_type.vcdkeconfig_type.nss}: View",
    "${vcd_rde_type.capvcdcluster_type.vendor}:${vcd_rde_type.capvcdcluster_type.nss}: Administrator Full access",
    "${vcd_rde_type.capvcdcluster_type.vendor}:${vcd_rde_type.capvcdcluster_type.nss}: Full Access",
    "${vcd_rde_type.capvcdcluster_type.vendor}:${vcd_rde_type.capvcdcluster_type.nss}: Modify",
    "${vcd_rde_type.capvcdcluster_type.vendor}:${vcd_rde_type.capvcdcluster_type.nss}: View",
    "${vcd_rde_type.capvcdcluster_type.vendor}:${vcd_rde_type.capvcdcluster_type.nss}: Administrator View",
    "vmware:tkgcluster: Full Access",
    "vmware:tkgcluster: Modify",
    "vmware:tkgcluster: View",
  ]

  publish_to_all_tenants = true # This needs to be published to all the Organizations

  # As we use rights created by the CAPVCD Type created previously, we need to depend on it
  depends_on = [
    vcd_rights_bundle.k8s_clusters_rights_bundle
  ]
}

# The VM Sizing Policies defined below MUST be created as they are specified in this HCL. These are the default
# policies required by CSE to create TKGm clusters.
# This should not be changed.
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
