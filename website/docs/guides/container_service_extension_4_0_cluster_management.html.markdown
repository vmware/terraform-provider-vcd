---
layout: "vcd"
page_title: "VMware Cloud Director: Container Service Extension v4.0 Kubernetes clusters management"
sidebar_current: "docs-vcd-guides-cse-4-0-cluster-management"
description: |-
  Provides guidance on provisioning Kubernetes clusters using Container Service Extension v4.0
---

# Container Service Extension v4.0 Kubernetes clusters management

## About

This guide explains how to create, update and delete TKGm clusters in a VCD appliance with Container Service Extension v4.0
installed, using Terraform.

We will use the [`vcd_rde`][rde] resource for this purpose.

~> This section assumes that the CSE installation was done following the [CSE v4.0 installation guide][cse_install_guide].
That is, CSE Server should be up and running and all elements must be working.

## Pre-requisites

-> Please read also the pre-requisites section in the [CSE documentation][cse_docs].

In order to complete the steps described in this guide, please be aware:

* CSE v4.0 must be installed. Read the [installation guide][cse_install_guide] for more information.
* Terraform VMWare Cloud Director provider needs to be v3.9.0 or above.
* CSE Server must be up and running.

## Creating a Kubernetes cluster

-> Please have a look at a working example of a TKGm cluster [here][cluster]. It is encouraged to read the following
section to understand how it works.

To be able to create a TKGm cluster, one needs to prepare a [`vcd_rde`][rde] resource. In the [proposed example][cluster],
this RDE is named `k8s_cluster_instance`. The important arguments to take into account are:

- **`resolve` must be always `false`**, because the CSE Server is responsible for performing the RDE resolution when the
  TKGm cluster is completely provisioned, so Terraform should not interfere with this process.
- **`resolve_on_removal` must be always `true`**, because the RDE is resolved by the CSE Server and not by Terraform. If one
  wants to execute a Terraform destroy without the RDE being resolved, the operation will fail. Being `true` assures that Terraform
  can perform a Terraform destroy in every case. See ["Deleting a Kubernetes cluster"](#deleting-a-kubernetes-cluster) section for more info.

The [`vcd_rde`][rde] argument `input_entity` is taking the output of the Terraform built-in function `templatefile`, that references
a JSON template that is located at [here][tkgmcluster_template]. This function will set the correct values to the following
placeholders that can be found in that file:

- `vcd_url`: The VCD URL, the same that was used during CSE installation.
- `name`: This will be the TKGm cluster name. It must contain only lowercase alphanumeric characters or '-',
  start with an alphabetic character, end with an alphanumeric, and contain at most 31 characters.
- `org`: The Organization in which the TKGm clusters will be created. In this guide it was created as `tenant_org` and named
  "Tenant Organization" during CSE installation phase.
- `vdc`: The VDC in which the TKGm clusters will be created. In this guide it was created as `tenant_vdc` and named
  "Tenant VDC" during CSE installation phase.
- `capi_yaml`: This must be set with a valid single-lined CAPVCD YAML, that will be explained next.
- `delete`: This is used to delete a cluster. See ["Deleting a Kubernetes cluster"](#deleting-a-kubernetes-cluster) section for more info.
  During creation it should be always `false`.
- `force_delete`: This is used to forcefully delete a cluster. See ["Deleting a Kubernetes cluster"](#deleting-a-kubernetes-cluster) section for more info.
  During creation it should be always `false`.
- `auto_repair_on_errors`: Setting this to `true` will make the CSE Server to automatically repair the TKGm cluster on errors.

The following four placeholders are **only needed if one wants to provide a default storage class** with the TKGm cluster.
If this is not needed, please remove the whole `defaultStorageClassOptions` block from the JSON template:

- `default_storage_class_filesystem`: Filesystem for the default storage class. Only `ext4` or `xfs` are valid.
- `default_storage_class_name`: Name of the default storage class, it must contain only lowercase alphanumeric characters or '-',
  start with an alphabetic character, end with an alphanumeric, and contain at most 63 characters.
- `default_storage_class_storage_profile`: Storage profile to use for the default storage class, for example `*`.
- `default_storage_class_delete_reclaim_policy`: Set this to `true` to use a "Delete" reclaim policy, that deletes the volume when the PersistentVolumeClaim is deleted.

To create a valid input for the `capi_yaml` placeholder, a [CAPVCD][capvcd] YAML is required, which describes the TKGm cluster to be
created. In order to craft it, we need to follow these steps:

- First, we need to download a YAML template from the [CAPVCD repository][capvcd_templates].
  We should choose the template that matches the TKGm OVA. For example, if we uploaded the `ubuntu-2004-kube-v1.22.9+vmware.1-tkg.1-2182cbabee08edf480ee9bc5866d6933.ova`
  and we want to use it, the template that we need to obtain corresponds to v1.22.9, that is `cluster-template-v1.22.9.yaml`.

- This template requires some extra elements to be added to the `kind: Cluster` block, inside `metadata`. These elements are `labels` and
  `annotations`, that are required by the CSE Server to be able to provision the cluster correctly. In other words, **cluster creation will fail
  if these are not added**:

```yaml
apiVersion: cluster.x-k8s.io/v1beta1
kind: Cluster
metadata:
  name: ${CLUSTER_NAME}
  namespace: ${TARGET_NAMESPACE}
  labels: # This `labels` block should be added, with its nested elements
    cluster-role.tkg.tanzu.vmware.com/management: ""
    tanzuKubernetesRelease: ${TKR_VERSION}
    tkg.tanzu.vmware.com/cluster-name: ${CLUSTER_NAME}
  annotations: # This `annotations` block should be added, with its nested element
    TKGVERSION: ${TKGVERSION}
# ...
```

- The downloaded template has a single worker pool (to see an example with **two** worker pools, please check the [proposed example][cluster]).
  If we need to have **more than one worker pool**, we have to add more objects of kind `VCDMachineTemplate`, `KubeadmConfigTemplate` and
  `MachineDeployment`. In the downloaded template, they look like this:

```yaml
# ...
apiVersion: infrastructure.cluster.x-k8s.io/v1beta1
kind: VCDMachineTemplate
metadata:
  name: ${CLUSTER_NAME}-md-0
# ...
apiVersion: bootstrap.cluster.x-k8s.io/v1beta1
kind: KubeadmConfigTemplate
metadata:
  name: ${CLUSTER_NAME}-md-0
  namespace: ${TARGET_NAMESPACE}
# ...
apiVersion: cluster.x-k8s.io/v1beta1
kind: MachineDeployment
metadata:
  name: ${CLUSTER_NAME}-md-0
  namespace: ${TARGET_NAMESPACE}
```

  Notice that the default worker pool is named `${CLUSTER_NAME}-md-0`. To add an extra one, we need to duplicate these
  blocks and name them differently, for example `name: ${CLUSTER_NAME}-md-1`.
  If we want to specify a different number of worker nodes per worker pool, we need to modify the original template, otherwise
  they will share the `${WORKER_MACHINE_COUNT}` placeholder located in the `MachineDeployment` object.
  The same happens with the **VM Sizing Policy, VM Placement Policy and Storage Profile**.
  See the explanation below for every placeholder to better understand how to adjust them.

Now that the YAML template is ready, one needs to understand the meaning of all the placeholders.
Below is the explanation of each one of them. See also [the working example][cluster] to observe the final result.

- `CLUSTER_NAME`: This will be the TKGm cluster name. It must contain only lowercase alphanumeric characters or '-',
start with an alphabetic character, end with an alphanumeric, and contain at most 31 characters.
- `TARGET_NAMESPACE`: This will be the TKGm cluster namespace. In [the example][cluster] the value is
`"${var.k8s_cluster_name}-ns"`, which mimics the UI behaviour, as the namespace is the name of the TKGm cluster concatenated with `-ns`.
- `VCD_SITE`: The VCD URL, the same that was used during CSE installation.
- `VCD_ORGANIZATION`: The Organization in which the TKGm clusters will be created. In this guide it was created as `tenant_org` and named
"Tenant Organization" during CSE installation phase.
- `VCD_ORGANIZATION_VDC`: The VDC in which the TKGm clusters will be created. In this guide it was created as `tenant_vdc` and named
  "Tenant VDC" during CSE installation phase.
- `VCD_ORGANIZATION_VDC_NETWORK`: The VDC network that the TKGm clusters will use. In this guide it was created as a Routed
  network called `tenant_net_routed`.
- `VCD_USERNAME_B64`: The name of a user with the "Kubernetes Cluster Author" role (`k8s_cluster_author`) that was created during CSE installation.
It must be encoded in Base64.
- `VCD_PASSWORD_B64` (**Discouraged in favor of `VCD_REFRESH_TOKEN_B64`**): The password of the user above.
  It must be encoded in Base64. Please do **not** use this value (by setting it to `""`) and use `VCD_REFRESH_TOKEN_B64` instead.
- `VCD_REFRESH_TOKEN_B64`: An API token that belongs to the user above. In UI, the API tokens can be generated in the user preferences
  in the top right, then go to the API tokens section, add a new one. Or we can visit `/tenant/<TENANT-NAME>/administration/settings/user-preferences`
  in the target VCD, logged in as the cluster author user in the desired tenant. It must be encoded in Base64.
- `SSH_PUBLIC_KEY`: We can set a public SSH key to be able to debug the TKGm control plane nodes. It can be empty (`""`)
- `CONTROL_PLANE_MACHINE_COUNT`: Number of control plane nodes (VMs). **Must be an odd number and higher than 0**.
- `VCD_CONTROL_PLANE_SIZING_POLICY`: Name of an existing VM Sizing Policy, created during CSE installation. Can be empty to use the VDC default (`""`)
- `VCD_CONTROL_PLANE_PLACEMENT_POLICY` : Name of an existing VM Placement Policy. Can be empty (`""`)
- `VCD_CONTROL_PLANE_STORAGE_PROFILE`: Name of an existing Storage Profile, for example `"*"` to use the default.
- `WORKER_MACHINE_COUNT`: Number of worker nodes (VMs). **Must be higher than 0**.
- `VCD_WORKER_SIZING_POLICY`: Name of an existing VM Sizing Policy, created during CSE installation. Can be empty to use the VDC default (`""`)
- `VCD_WORKER_PLACEMENT_POLICY`: Name of an existing VM Placement Policy. Can be empty (`""`)
- `VCD_WORKER_STORAGE_PROFILE`: Name of an existing Storage Profile, for example `"*"` to use the default.
- `DISK_SIZE`: Specifies the storage size for each node (VM). It uses the [same units as every other Kubernetes resource](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/),
  for example `"1Gi"` to use 1 gibibyte (1024 MiB), or `"1G"` for 1 gigabyte (1000 MB).
- `VCD_CATALOG`: The catalog where the TKGm OVAs are. For example, in the above CSE installation it was `tkgm_catalog`.
- `VCD_TEMPLATE_NAME` = The TKGm OVA name, for example `ubuntu-2004-kube-v1.22.9+vmware.1-tkg.1-2182cbabee08edf480ee9bc5866d6933`
- `POD_CIDR`: The CIDR used for Pod networking, for example `"100.96.0.0/11"`.
- `SERVICE_CIDR`: The CIDR used for Service networking, for example `"100.64.0.0/13"`.

There are three additional variables that we added manually: `TKR_VERSION` and `TKGVERSION`.
To know their values, one can use [this script](https://github.com/vmware/cluster-api-provider-cloud-director/blob/main/docs/WORKLOAD_CLUSTER.md#script-to-get-kubernetes-etcd-coredns-versions-from-tkg-ova),
or check the following table with some script outputs:

| OVA name                                                 | TKR_VERSION               | TKGVERSION |
|----------------------------------------------------------|---------------------------|------------|
| v1.19.16+vmware.1-tkg.2-fba68db15591c15fcd5f26b512663a42 | v1.19.16---vmware.1-tkg.2 | v1.4.3     |
| v1.20.14+vmware.1-tkg.2-5a5027ce2528a6229acb35b38ff8084e | v1.20.14---vmware.1-tkg.2 | v1.4.3     |
| v1.20.15+vmware.1-tkg.2-839faf7d1fa7fa356be22b72170ce1a8 | v1.20.15---vmware.1-tkg.2 | v1.5.4     |
| v1.21.8+vmware.1-tkg.2-ed3c93616a02968be452fe1934a1d37c  | v1.21.8---vmware.1-tkg.2  | v1.4.3     |
| v1.21.11+vmware.1-tkg.2-d788dbbb335710c0a0d1a28670057896 | v1.21.11---vmware.1-tkg.2 | v1.5.4     |
| v1.22.9+vmware.1-tkg.1-2182cbabee08edf480ee9bc5866d6933  | v1.22.9---vmware.1-tkg.1  | v1.5.4     |
| v1.22.13+vmware.1-tkg.2-ea08b304658a6cf17f5e74dc0ab7544f | v1.22.13---vmware.1-tkg.1 | v1.6.1     |
| v1.21.14+vmware.2-tkg.5-d793afae5aa18e50bd9175e339904496 | v1.21.14---vmware.2-tkg.5 | v1.6.1     |
| v1.23.10+vmware.1-tkg.2-b53d41690f8742e7388f2c553fd9a181 | v1.23.10---vmware.1-tkg.1 | v1.6.1     |

In [the TKGm cluster creation example][cluster], the built-in Terraform function `templatefile` is used to substitute every placeholder
mentioned above with its final value. The returned value is the CAPVCD YAML payload that needs to be set in the `capi_yaml` placeholder in the
JSON template.

~> Notice that we need to replace `\n` with `\\n` and also `\"` to `\\\"` to avoid breaking the JSON contents when the YAML is set, as
line breaks and double quotes would not be interpreted correctly otherwise. This is also done in [the mentioned example][cluster].

When we have set all the required values, a `terraform apply` should trigger a TKGm cluster creation. The operation is asynchronous, meaning that
we need to monitor the RDE `computed_entity` value to see the status of the cluster provisioning. Some interesting output examples:

```hcl
locals {
  k8s_cluster_computed = jsondecode(vcd_rde.k8s_cluster_instance.computed_entity)
  has_status           = lookup(local.k8s_cluster_computed, "status", null) != null
}

# Outputs the TKGm Cluster creation status
output "computed_k8s_cluster_status" {
  value = local.has_status ? local.k8s_cluster_computed["status"]["vcdKe"]["state"] : null
}

# Outputs the TKGm Cluster creation events
output "computed_k8s_cluster_events" {
  value = local.has_status ? local.k8s_cluster_computed["status"]["vcdKe"]["eventSet"] : null
}

```

When the status displayed by `computed_k8s_cluster_status` is `provisioned`, it will mean that the TKGm cluster is successfully provisioned and
the Kubeconfig is available and ready to use. It can be retrieved it with:

```hcl
locals {
  is_k8s_cluster_provisioned = local.has_status ? local.k8s_cluster_computed["status"]["vcdKe"]["state"] == "provisioned" ? lookup(local.k8s_cluster_computed["status"], "capvcd", null) != null : false : false
}

output "computed_k8s_cluster_kubeconfig" {
  value = local.is_k8s_cluster_provisioned ? local.k8s_cluster_computed["status"]["capvcd"]["private"]["kubeConfig"] : null
}
```

## Updating a Kubernetes cluster

We can perform a Terraform update to resize a TKGm cluster, for example. In order to do that, we must take into account how the
[`vcd_rde`][rde] resource works. We should read [its documentation][rde_input_vs_computed] to better understand how updates work
in this specific case.

To apply a correct update, we need to take the most recent state of the TKGm cluster, which is reflected in the contents of
the `computed_entity` attribute. Copy the value of this attribute, edit the properties that we would like to modify, and place the
final result inside `input_entity`. Now the changes can be applied with `terraform apply`.

~> Do **NOT** use the initial `input_entity` contents to perform an update, as the CSE Server puts vital information in
the RDE contents (which is reflected in the `computed_entity` attribute) that were not in the first JSON payload.
If this information is not sent back, **the cluster will be broken**.

Upgradeable items:

- TKGm OVA: If there is a newer version of TKGm, we can modify the referenced OVA.
- Number of worker nodes. Remember this must be higher than 0.
- Number of control plane nodes. Remember this must be an odd number and higher than 0.

## Deleting a Kubernetes cluster

~> Do **NOT** remove the cluster from the HCL configuration! This will leave dangling resources that the CSE Server creates
when the TKGm cluster is created, such as vApps, networks, virtual services, etc. Please follow the procedure described in
this section to destroy a cluster entirely.

To delete an existing TKGm cluster, one needs to mark it for deletion in the `vcd_rde` resource. In the [example configuration][cluster],
there are two keys `delete` and `force_delete` that correspond to the CAPVCD [RDE Type][rde_type] schema fields `markForDelete`
and `forceDelete` respectively.

- Setting `delete = true` will make the CSE Server remove the cluster from VCD and eventually the RDE.
- Setting also `force_delete = true` will force the CSE Server to delete the cluster, and their associated resources
  that are not fully complete and that are in an unremovable state.

Follow the instructions described in ["Updating a Kubernetes cluster"](#updating-a-kubernetes-cluster) to learn how to perform
the update of these two properties.

Once updated, one can monitor the `vcd_rde` resource to check the deletion process. Eventually, the RDE won't exist anymore in VCD and Terraform will
ask for creation again. It can be now removed from the HCL configuration.

[capvcd]: https://github.com/vmware/cluster-api-provider-cloud-director
[capvcd_templates]: https://github.com/vmware/cluster-api-provider-cloud-director/tree/main/templates
[cluster]: https://github.com/vmware/terraform-provider-vcd/tree/main/examples/container-service-extension-4.0/cluster
[cse_install_guide]: /providers/vmware/vcd/latest/docs/guides/container_service_extension_4_0_install
[cse_docs]: https://docs.vmware.com/en/VMware-Cloud-Director-Container-Service-Extension/index.html
[rde]: /providers/vmware/vcd/latest/docs/resources/rde
[rde_input_vs_computed]: /providers/vmware/vcd/latest/docs/resources/rde#input-entity-vs-computed-entity
[rde_type]: /providers/vmware/vcd/latest/docs/resources/rde_type
[tkgmcluster_template]: https://github.com/vmware/terraform-provider-vcd/tree/main/examples/container-service-extension-4.0/entities/tkgmcluster-template.json
