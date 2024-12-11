---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_cse_kubernetes_cluster"
sidebar_current: "docs-vcd-resource-cse-kubernetes-cluster"
description: |-
  Provides a resource to manage Kubernetes clusters in VMware Cloud Director with Container Service Extension installed and running.
---

# vcd\_cse\_kubernetes\_cluster

Provides a resource to manage Kubernetes clusters in VMware Cloud Director with Container Service Extension (CSE) installed and running.

Supported in provider *v3.12+*

Supports the following **Container Service Extension** versions:

* [4.1.0](https://docs.vmware.com/en/VMware-Cloud-Director-Container-Service-Extension/4.1/rn/vmware-cloud-director-container-service-extension-41-release-notes/index.html) (Terraform Provider v3.12+)
* [4.1.1 / 4.1.1a](https://docs.vmware.com/en/VMware-Cloud-Director-Container-Service-Extension/4.1.1/rn/vmware-cloud-director-container-service-extension-411-release-notes/index.html) (Terraform Provider v3.12 or above)
* [4.2.0](https://docs.vmware.com/en/VMware-Cloud-Director-Container-Service-Extension/4.2/rn/vmware-cloud-director-container-service-extension-42-release-notes/index.html) (Terraform Provider v3.12 or above)
* [4.2.1](https://docs.vmware.com/en/VMware-Cloud-Director-Container-Service-Extension/4.2.1/rn/vmware-cloud-director-container-service-extension-421-release-notes/index.html) (Terraform Provider v3.12 or above)
* [4.2.2](https://docs.vmware.com/en/VMware-Cloud-Director-Container-Service-Extension/4.2.2/rn/vmware-cloud-director-container-service-extension-422-release-notes/index.html) (Terraform Provider v3.14.1 or above)
* [4.2.3](https://docs.vmware.com/en/VMware-Cloud-Director-Container-Service-Extension/4.2.3/rn/vmware-cloud-director-container-service-extension-423-release-notes/index.html) (Terraform Provider v3.14.1 or above)

-> To install CSE in VMware Cloud Director, please follow [this guide](/providers/vmware/vcd/latest/docs/guides/container_service_extension_4_x_install)

## Example Usage

```hcl
data "vcd_catalog" "tkg_catalog" {
  org  = "solutions_org" # The catalog is shared with 'tenant_org', so it is visible for tenants
  name = "tkgm_catalog"
}

# Fetch a valid Kubernetes template OVA. If it's not valid, cluster creation will fail.
data "vcd_catalog_vapp_template" "tkg_ova" {
  org        = data.vcd_catalog.tkg_catalog.org
  catalog_id = data.vcd_catalog.tkg_catalog.id
  name       = "ubuntu-2204-kube-v1.30.2+vmware.1-tkg.1-00b380629c7a9c10afaaa9df46ba2283"
}

data "vcd_org_vdc" "vdc" {
  org  = "tenant_org"
  name = "tenant_vdc"
}

data "vcd_nsxt_edgegateway" "egw" {
  org      = data.vcd_org_vdc.vdc.org
  owner_id = data.vcd_org_vdc.vdc.id
  name     = "tenant_edgegateway"
}

data "vcd_network_routed_v2" "routed" {
  org             = data.vcd_nsxt_edgegateway.egw.org
  edge_gateway_id = data.vcd_nsxt_edgegateway.egw.id
  name            = "tenant_net_routed"
}

# Fetch a valid Sizing policy created during CSE installation.
# Refer to the CSE installation guide for more information.
data "vcd_vm_sizing_policy" "tkg_small" {
  name = "TKG small"
}

data "vcd_storage_profile" "sp" {
  org  = data.vcd_org_vdc.vdc.org
  vdc  = data.vcd_org_vdc.vdc.name
  name = "*"
}

# The token file is required, and it should be safely stored
# Versions 4.2.2 and 4.2.3 should NOT use a System Administrator token
resource "vcd_api_token" "token" {
  name             = "myClusterToken"
  file_name        = "/home/Bob/safely_stored_token.json"
  allow_token_file = true
}

resource "vcd_cse_kubernetes_cluster" "my_cluster" {
  cse_version            = "4.2.3"
  runtime                = "tkg"
  name                   = "test2"
  kubernetes_template_id = data.vcd_catalog_vapp_template.tkg_ova.id
  org                    = data.vcd_org_vdc.vdc.org
  vdc_id                 = data.vcd_org_vdc.vdc.id
  network_id             = data.vcd_network_routed_v2.routed.id
  api_token_file         = vcd_api_token.token.file_name

  control_plane {
    machine_count      = 1
    disk_size_gi       = 20
    sizing_policy_id   = data.vcd_vm_sizing_policy.tkg_small.id
    storage_profile_id = data.vcd_storage_profile.sp.id
  }

  worker_pool {
    name               = "node-pool-1"
    machine_count      = 1
    disk_size_gi       = 20
    sizing_policy_id   = data.vcd_vm_sizing_policy.tkg_small.id
    storage_profile_id = data.vcd_storage_profile.sp.id
  }

  default_storage_class {
    name               = "sc-1"
    storage_profile_id = data.vcd_storage_profile.sp.id
    reclaim_policy     = "delete"
    filesystem         = "ext4"
  }

  auto_repair_on_errors = true
  node_health_check     = true

  operations_timeout_minutes = 0
}

output "kubeconfig" {
  value     = vcd_cse_kubernetes_cluster.my_cluster.kubeconfig
  sensitive = true
}
```

## Argument Reference

The following arguments are supported:

* `cse_version` - (Required) Specifies the CSE version to use. Accepted versions: `4.1.0`, `4.1.1` (also for *4.1.1a*),
  `4.2.0`, `4.2.1`, `4.2.2` (VCD Provider *v3.14.1+*) and `4.2.3` (VCD Provider *v3.14.1+*)
* `runtime` - (Optional) Specifies the Kubernetes runtime to use. Defaults to `tkg` (Tanzu Kubernetes Grid)
* `name` - (Required) The name of the Kubernetes cluster. It must contain only lowercase alphanumeric characters or "-",
  start with an alphabetic character, end with an alphanumeric, and contain at most 31 characters
* `kubernetes_template_id` - (Required) The ID of the vApp Template that corresponds to a Kubernetes template OVA
* `org` - (Optional) The name of organization that will host the Kubernetes cluster, optional if defined in the provider configuration
* `vdc_id` - (Required) The ID of the VDC that hosts the Kubernetes cluster
* `network_id` - (Required) The ID of the network that the Kubernetes cluster will use
* `owner` - (Optional) The user that creates the cluster and owns the API token specified in `api_token`.
  It must have the `Kubernetes Cluster Author` role that was created during CSE installation.
  If not specified, it assumes it's the user from the provider configuration

~> Versions 4.2.2 and 4.2.3 should not use the System administrator for the `owner` nor `api_token_file`, as stated in their
[release notes](https://docs.vmware.com/en/VMware-Cloud-Director-Container-Service-Extension/4.2.2/rn/vmware-cloud-director-container-service-extension-422-release-notes/index.html#Known%20Issues),
there is an existing issue that prevents the cluster to be created.

* `api_token_file` - (Required) Must be a file generated by [`vcd_api_token` resource](/providers/vmware/vcd/latest/docs/resources/api_token),
  or a file that follows the same formatting, that stores the API token used to create and manage the cluster,
  owned by the user specified in `owner`. Be careful about this file, as it contains sensitive information
* `ssh_public_key` - (Optional) The SSH public key used to log in into the cluster nodes
* `control_plane` - (Required) See [**Control Plane**](#control-plane)
* `worker_pool` - (Required) See [**Worker Pools**](#worker-pools)
* `default_storage_class` - (Optional) See [**Default Storage Class**](#default-storage-class)
* `pods_cidr` - (Optional) A CIDR block for the pods to use. Defaults to `100.96.0.0/11`
* `services_cidr` - (Optional) A CIDR block for the services to use. Defaults to `100.64.0.0/13`
* `virtual_ip_subnet` - (Optional) A virtual IP subnet for the cluster
* `auto_repair_on_errors` - (Optional) If errors occur before the Kubernetes cluster becomes available, and this argument is `true`,
  CSE Server will automatically attempt to repair the cluster. Defaults to `false`.
  Since CSE 4.1.1, when the cluster is available/provisioned, this flag is set automatically to false.
* `node_health_check` - (Optional) After the Kubernetes cluster becomes available, nodes that become unhealthy will be
  remediated according to unhealthy node conditions and remediation rules. Defaults to `false`
* `operations_timeout_minutes` - (Optional) The time, in minutes, to wait for the cluster operations to be successfully completed.
  For example, during cluster creation, it should be in `provisioned` state before the timeout is reached, otherwise the
  operation will return an error. For cluster deletion, this timeout specifies the time to wait until the cluster is completely deleted.
  Setting this argument to `0` means to wait indefinitely (not recommended as it could hang Terraform if the cluster can't be created
  due to a configuration error if `auto_repair_on_errors=true`). Defaults to `60`

### Control Plane

The `control_plane` block is **required** and unique per resource, meaning that there must be **exactly one** of these
in every resource.

This block asks for the following arguments:

* `machine_count` - (Optional) The number of nodes that the control plane has. Must be an odd number and higher than `0`. Defaults to `3`
* `disk_size_gi` - (Optional) Disk size, in **Gibibytes (Gi)**, for the control plane VMs. Must be at least `20`. Defaults to `20`
* `sizing_policy_id` - (Optional) VM Sizing policy for the control plane VMs. Must be one of the ones made available during CSE installation
* `placement_policy_id` - (Optional) VM Placement policy for the control plane VMs
* `storage_profile_id` - (Optional) Storage profile for the control plane VMs
* `ip` - (Optional) IP for the control plane. It will be automatically assigned during cluster creation if left empty

### Worker Pools

The `worker_pool` block is **required**, and every cluster should have **at least one** of them.

Each block asks for the following arguments:

* `name` - (Required) The name of the worker pool. It must be unique per cluster, and must contain only lowercase alphanumeric characters or "-",
  start with an alphabetic character, end with an alphanumeric, and contain at most 31 characters
* `machine_count` - (Optional) The number of VMs that the worker pool has. Must be higher than `0`, unless `autoscaler_max_replicas` and `autoscaler_min_replicas` are set,
  in this case it must be `0` (in this particular case, this value will not be used). Defaults to `1`
* `disk_size_gi` - (Optional) Disk size, in **Gibibytes (Gi)**, for the worker pool VMs. Must be at least `20`. Defaults to `20`
* `sizing_policy_id` - (Optional) VM Sizing policy for the control plane VMs. Must be one of the ones made available during CSE installation
* `placement_policy_id` - (Optional) VM Placement policy for the worker pool VMs. If this one is set, `vgpu_policy_id` must be empty
* `vgpu_policy_id` - (Optional) vGPU policy for the worker pool VMs. If this one is set, `placement_policy_id` must be empty
* `storage_profile_id` - (Optional) Storage profile for the worker pool VMs
* `autoscaler_max_replicas` - (Optional; *v3.13+*) Together with `autoscaler_min_replicas`, and **only when `machine_count=0` (or unset)**, defines the maximum number of nodes that
  the Kubernetes Autoscaler will deploy for this worker pool. Read the section below for details.
* `autoscaler_min_replicas` - (Optional; *v3.13+*) Together with `autoscaler_max_replicas`, and **only when `machine_count=0` (or unset)**, defines the minimum number of nodes that
  the Kubernetes Autoscaler will deploy for this worker pool. Read the section below for details.

#### Worker pools with Kubernetes Autoscaler enabled

-> Supported in provider *v3.13+*

~> The Autoscaler can only work in clusters with Internet access, as it needs to download the Docker image from
k8s.gcr.io/autoscaling/cluster-autoscaler

The **Kubernetes Autoscaler** is a component that automatically adjusts the size of a Kubernetes Cluster so that all pods have a
place to run and there are no unneeded nodes. You can read more about it [here](https://github.com/kubernetes/autoscaler/blob/master/cluster-autoscaler/cloudprovider/clusterapi/README.md).

This provider has two arguments for the `worker_pool` block since version v3.13.0: `autoscaler_max_replicas` and `autoscaler_min_replicas`.
They allow to define the maximum and minimum amount of nodes of a pool, respectively. They specify the autoscaling
capabilities of the given Worker Pool as defined [here](https://www.vmware.com/content/dam/digitalmarketing/vmware/en/pdf/docs/vmw-whitepaper-cluster-auto-scaler.pdf).

If at least **one** `worker_pool` block has `autoscaler_max_replicas` and `autoscaler_min_replicas` defined (and subsequently, `machine_count=0`),
the provider will deploy the Kubernetes Autoscaler in the cluster `kube-system` namespace, with the following components:

* A `Deployment`, this is the Autoscaler deployment definition
* Dependencies for the Autoscaler: A `ServiceAccount`, a `ClusterRole`, a `Role` and a `ClusterRoleBinding`

These elements are the ones defined in [the documentation](https://www.vmware.com/content/dam/digitalmarketing/vmware/en/pdf/docs/vmw-whitepaper-cluster-auto-scaler.pdf).

The Kubernetes Autoscaler will be deployed only **once**, as soon as **one** `worker_pool` requires it, and it will be scaled up/down
depending on the requirements of the worker pools throughout their lifecycle: If **all** of the `worker_pool` blocks unset the autoscaling
arguments during following updates, the Autoscaler deployment will be **scaled down to 0 replicas**.
If one of the `worker_pool` blocks requires autoscaling again, it will be **scaled up to 1 replica**.

~> From the [FAQ](https://github.com/kubernetes/autoscaler/blob/master/cluster-autoscaler/FAQ.md), be aware that "by default, Cluster Autoscaler does
not enforce the node group size. If your cluster is below the minimum number of nodes configured for CA, it will be scaled up only in presence of
unschedulable pods. On the other hand, if your cluster is above the minimum number of nodes configured for CA, it will be scaled down only if it
has unneeded nodes."

```hcl
resource "vcd_cse_kubernetes_cluster" "my_cluster" {
  name = "test"
  # ... Omitted

  control_plane {
    # ... Omitted
  }

  worker_pool {
    name               = "node-pool-1"
    disk_size_gi       = 20
    sizing_policy_id   = data.vcd_vm_sizing_policy.tkg_small.id
    storage_profile_id = data.vcd_storage_profile.sp.id

    # Enables the Kubernetes Autoscaler for this Worker Pool
    autoscaler_max_replicas = 10
    autoscaler_min_replicas = 2
  }
  worker_pool {
    name               = "node-pool-1"
    machine_count      = 1 # Regular static replicas
    disk_size_gi       = 20
    sizing_policy_id   = data.vcd_vm_sizing_policy.tkg_small.id
    storage_profile_id = data.vcd_storage_profile.sp.id
  }
}
```

### Default Storage Class

The `default_storage_class` block is **optional**, and every cluster should have **at most one** of them.

If defined, the block asks for the following arguments:

* `name` - (Required) The name of the default storage class. It must contain only lowercase alphanumeric characters or "-",
  start with an alphabetic character, end with an alphanumeric, and contain at most 31 characters
* `storage_profile_id` - (Required) Storage profile for the default storage class
* `reclaim_policy` - (Required) A value of `delete` deletes the volume when the PersistentVolumeClaim is deleted. `retain` does not,
  and the volume can be manually reclaimed
* `filesystem` - (Required) Filesystem of the storage class, can be either `ext4` or `xfs`

## Attribute Reference

The following attributes are available for consumption as read-only attributes after a successful cluster creation:

* `kubernetes_version` - The version of Kubernetes installed in this cluster
* `tkg_product_version` - The version of TKG installed in this cluster
* `capvcd_version` - The version of CAPVCD used by this cluster
* `cluster_resource_set_bindings` - The cluster resource set bindings of this cluster
* `cpi_version` - The version of the Cloud Provider Interface used by this cluster
* `csi_version` - The version of the Container Storage Interface used by this cluster
* `state` - The Kubernetes cluster status, can be `provisioning` when it is being created, `provisioned` when it was successfully
  created and ready to use, or `error` when an error occurred. `provisioning` can only be obtained when a timeout happens during
  cluster creation. `error` can only be obtained either with a timeout or when `auto_repair_on_errors=false`.
* `kubeconfig` - The ready-to-use Kubeconfig file **contents** as a raw string. Only available when `state=provisioned`
* `supported_upgrades` - A set of vApp Template names that can be fetched with a
  [`vcd_catalog_vapp_template` data source](/providers/vmware/vcd/latest/docs/data-sources/catalog_vapp_template) to upgrade the cluster.
* `events` - A set of events that happened during the Kubernetes cluster lifecycle. They're ordered from most recent to least. Each event has:
  * `name` - Name of the event
  * `resource_id` - ID of the resource that caused the event
  * `type` - Type of the event, either `event` or `error`
  * `details` - Details of the event
  * `occurred_at` - When the event happened

## Updating

Only the following arguments can be updated:

* `kubernetes_template_id`: The cluster must allow upgrading to the new TKG version. You can check `supported_upgrades` attribute to know
  the available OVAs. Upgrading the Kubernetes version will also upgrade the Cluster Autoscaler to its corresponding minor version, if it is being used by any `worker_pool`.
* `machine_count` of the `control_plane`: Supports scaling up and down. Nothing else can be updated.
* `machine_count` of any `worker_pool`: Supports scaling up and down. Use caution when resizing down to 0 nodes.
  The cluster must always have at least 1 running node, or else the cluster will enter an unrecoverable error state.
* `auto_repair_on_errors`: Can only be updated in CSE 4.1.0, and it is recommended to set it to `false` when the cluster is created.
  In versions higher than 4.1.0, this is automatically done by the CSE Server, so this flag cannot be updated.
* `node_health_check`: Can be turned on/off.
* `operations_timeout_minutes`: Does not require modifying the existing cluster

You can also add more `worker_pool` blocks to add more Worker Pools to the cluster. **You can't delete Worker Pools**, but they can
be scaled down to zero.

Updating any other argument will delete the existing cluster and create a new one, when the Terraform plan is applied.

Modifying the CSE version of a cluster with `cse_version` is not supported.

## Accessing the Kubernetes cluster

To retrieve the Kubeconfig of a created cluster, you may set it as an output:

```hcl
output "kubeconfig" {
  value     = vcd_cse_kubernetes_cluster.my_cluster.kubeconfig
  sensitive = true
}
```

Then, creating a file turns out to be trivial:

```shell
terraform output -raw kubeconfig > $HOME/kubeconfig
```

The Kubeconfig can now be used with `kubectl` and the Kubernetes cluster can be used.

## Importing

An existing Kubernetes cluster can be [imported][docs-import] into this resource via supplying the **Cluster ID** for it.
The ID can be easily obtained in VCD UI, in the CSE Kubernetes Container Clusters plugin.

An example is below. During import, none of the mentioned arguments are required, but they will be in subsequent Terraform commands
such as `terraform plan`. Each comment in the code gives some context about how to obtain them to have a completely manageable cluster:

```hcl
# This is just a snippet of code that will host the imported cluster that already exists in VCD.
# This must NOT be created with Terraform beforehand, it is just a shell that will receive the information
# None of the arguments are required during the Import phase, but they will be asked when operating it afterwards
resource "vcd_cse_kubernetes_cluster" "imported_cluster" {
  name                   = "test2"                                   # The name of the existing cluster
  cse_version            = "4.2.3"                                   # The CSE version installed in your VCD
  kubernetes_template_id = data.vcd_catalog_vapp_template.tkg_ova.id # See below data sources
  vdc_id                 = data.vcd_org_vdc.vdc.id                   # See below data sources
  network_id             = data.vcd_network_routed_v2.routed.id      # See below data sources
  node_health_check      = true                                      # Whether the existing cluster has Machine Health Check enabled or not, this can be checked in UI

  control_plane {
    machine_count      = 5                                      # This is optional, but not setting it to the current value will make subsequent plans to try to scale our existing cluster to the default one
    sizing_policy_id   = data.vcd_vm_sizing_policy.tkg_small.id # See below data sources
    storage_profile_id = data.vcd_storage_profile.sp.id         # See below data sources
  }

  worker_pool {
    name               = "node-pool-1"                          # The name of the existing worker pool of the existing cluster. Retrievable from UI
    machine_count      = 40                                     # This is optional, but not setting it to the current value will make subsequent plans to try to scale our existing cluster to the default one
    sizing_policy_id   = data.vcd_vm_sizing_policy.tkg_small.id # See below data sources
    storage_profile_id = data.vcd_storage_profile.sp.id         # See below data sources
  }

  # While optional, we cannot change the Default Storage Class after an import, so we need
  # to set the information of the existing cluster to avoid re-creation.
  # The information can be retrieved from UI
  default_storage_class {
    filesystem         = "ext4"
    name               = "sc-1"
    reclaim_policy     = "delete"
    storage_profile_id = data.vcd_storage_profile.sp.id # See below data sources
  }
}

# The below data sources are needed to retrieve the required IDs. They are not needed
# during the Import phase, but they will be asked when operating it afterwards

# The VDC and Organization where the existing cluster is located
data "vcd_org_vdc" "vdc" {
  org  = "tenant_org"
  name = "tenant_vdc"
}

# The OVA that the existing cluster is using. You can obtain the OVA by inspecting
# the existing cluster TKG/Kubernetes version.
data "vcd_catalog_vapp_template" "tkg_ova" {
  org        = data.vcd_catalog.tkg_catalog.org
  catalog_id = data.vcd_catalog.tkg_catalog.id
  name       = "ubuntu-2204-kube-v1.30.2+vmware.1-tkg.1-00b380629c7a9c10afaaa9df46ba2283"
}

# The network that the existing cluster is using
data "vcd_network_routed_v2" "routed" {
  org             = data.vcd_nsxt_edgegateway.egw.org
  edge_gateway_id = data.vcd_nsxt_edgegateway.egw.id
  name            = "tenant_net_routed"
}

# The VM Sizing Policy of the existing cluster nodes
data "vcd_vm_sizing_policy" "tkg_small" {
  name = "TKG small"
}

# The Storage Profile that the existing cluster uses
data "vcd_storage_profile" "sp" {
  org  = data.vcd_org_vdc.vdc.org
  vdc  = data.vcd_org_vdc.vdc.name
  name = "*"
}

data "vcd_catalog" "tkg_catalog" {
  org  = "solutions_org" # The Organization that shares the TKGm OVAs with the tenants
  name = "tkgm_catalog"  # The Catalog name
}

data "vcd_nsxt_edgegateway" "egw" {
  org      = data.vcd_org_vdc.vdc.org
  owner_id = data.vcd_org_vdc.vdc.id
  name     = "tenant_edgegateway"
}
```

```sh
terraform import vcd_cse_kubernetes_cluster.imported_cluster urn:vcloud:entity:vmware:capvcdCluster:1d24af33-6e5a-4d47-a6ea-06d76f3ee5c9
```

-> The ID is required as it is the only way to unequivocally identify a Kubernetes cluster inside VCD. To obtain the ID
you can check the Kubernetes Container Clusters UI plugin, where all the available clusters are listed.

After that, you can expand the configuration file and either update or delete the Kubernetes cluster. Running `terraform plan`
at this stage will show the difference between the minimal configuration file and the Kubernetes cluster stored properties.

### Importing with Import blocks (Terraform v1.5+)

~> Terraform warns that this procedure is considered **experimental**. Read more [here](/providers/vmware/vcd/latest/docs/guides/importing_resources)

Given a Cluster ID, like `urn:vcloud:entity:vmware:capvcdCluster:f2d88194-3745-47ef-a6e1-5ee0bbce38f6`, you can write
the following HCL block in your Terraform configuration:

```hcl
import {
  to = vcd_cse_kubernetes_cluster.imported_cluster
  id = "urn:vcloud:entity:vmware:capvcdCluster:f2d88194-3745-47ef-a6e1-5ee0bbce38f6"
}
```

Instead of using the suggested snippet in the section above, executing the command
`terraform plan -generate-config-out=generated_resources.tf` will generate a similar code, automatically.

Once the code is validated, running `terraform apply` will perform the import operation and save the Kubernetes cluster
into the Terraform state. The Kubernetes cluster can now be operated with Terraform.

[docs-import]:https://www.terraform.io/docs/import/
