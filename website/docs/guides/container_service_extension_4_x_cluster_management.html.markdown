---
layout: "vcd"
page_title: "VMware Cloud Director: Container Service Extension v4.1 Kubernetes clusters management"
sidebar_current: "docs-vcd-guides-cse-4-x-cluster-management"
description: |-
  Provides guidance on provisioning Kubernetes clusters using Container Service Extension v4.1
---

# Container Service Extension v4.1 Kubernetes clusters management

## About

This guide explains how to create, update and delete **Tanzu Kubernetes Grid multicloud (TKGm)** clusters in a VCD appliance with Container Service Extension v4.1
installed, using Terraform.

We will use the Runtime Defined Entity (RDE) resource [`vcd_rde`][rde] for this purpose.

~> This section assumes that the CSE installation was done following the [CSE v4.1 installation guide][cse_install_guide].
That is, CSE Server should be up and running and all elements must be working.

## Pre-requisites

-> Please read also the pre-requisites section in the [CSE documentation][cse_docs].

In order to complete the steps described in this guide, please be aware:

* CSE v4.1 must be installed. Read the [installation guide][cse_install_guide] for more information.
* Terraform VMWare Cloud Director provider needs to be v3.11.0 or above.
* CSE Server must be up and running.

## Creating a Kubernetes cluster

-> Please have a look at a working example of a TKGm cluster [here][cluster]. It is encouraged to read the following
section to understand how it works.

To be able to create a TKGm cluster, one needs to prepare a [`vcd_rde`][rde] resource. In the [proposed example][cluster],
this RDE is named `k8s_cluster_instance`. The important arguments to take into account are:

* **`resolve` must be always `false`**, because the CSE Server is responsible for performing the RDE resolution when the
  TKGm cluster is completely provisioned, so Terraform should not interfere with this process.
* **`resolve_on_removal` must be always `true`**, because the RDE is resolved by the CSE Server and not by Terraform. If one
  wants to execute a Terraform destroy without the RDE being resolved, the operation will fail. Being `true` assures that Terraform
  can perform a Terraform destroy in every case. See ["Deleting a Kubernetes cluster"](#deleting-a-kubernetes-cluster) section for more info.

The [`vcd_rde`][rde] argument `input_entity` is taking the output of the Terraform built-in function `templatefile`, that references
a JSON template that is located at [here][tkgmcluster_template]. This function will set the correct values to the following
placeholders that can be found in that file:

* `vcd_url`: The VCD URL, the same that was used during CSE installation.
* `name`: This will be the TKGm cluster name. It must contain only lowercase alphanumeric characters or '-',
  start with an alphabetic character, end with an alphanumeric, and contain at most 31 characters.
* `org`: The Organization in which the TKGm clusters will be created. In this guide it was created as `tenant_org` and named
  "Tenant Organization" during CSE installation phase.
* `vdc`: The VDC in which the TKGm clusters will be created. In this guide it was created as `tenant_vdc` and named
  "Tenant VDC" during CSE installation phase.
* `api_token`: The API token that corresponds to the user that will create the cluster. This is created with a [`vcd_api_token`][api_token] resource.
  One can customise the path to the JSON file where the API token is stored using `cluster_author_token_file` variable.
* `capi_yaml`: This must be set with a valid CAPVCD YAML, that is explained below.
* `delete`: This is used to delete a cluster. See ["Deleting a Kubernetes cluster"](#deleting-a-kubernetes-cluster) section for more info.
  During creation, it should be always `false`.
* `force_delete`: This is used to forcefully delete a cluster. See ["Deleting a Kubernetes cluster"](#deleting-a-kubernetes-cluster) section for more info.
  During creation, it should be always `false`.
* `auto_repair_on_errors`: Setting this to `true` will make the CSE Server to automatically repair the TKGm cluster on errors.

The following four placeholders are **only needed if one wants to provide a default storage class** with the TKGm cluster.
If this is not needed, please remove the whole `defaultStorageClassOptions` block from [the JSON template][tkgmcluster_template]:

* `default_storage_class_filesystem`: Filesystem for the default storage class. Only `ext4` or `xfs` are valid.
* `default_storage_class_name`: Name of the default storage class, it must contain only lowercase alphanumeric characters or '-',
  start with an alphabetic character, end with an alphanumeric, and contain at most 63 characters.
* `default_storage_class_storage_profile`: Storage profile to use for the default storage class, for example `*`.
* `default_storage_class_delete_reclaim_policy`: Set this to `true` to use a "Delete" reclaim policy, that deletes the volume when the PersistentVolumeClaim is deleted.

To create a valid input for the `capi_yaml` placeholder, a [CAPVCD][capvcd] YAML is required, which describes the TKGm cluster to be
created. In order to craft it, we need to follow these steps:

* First, we need to download a YAML template from the [CAPVCD repository][capvcd_templates].
  We should choose the template that matches the TKGm OVA. For example, if we uploaded the `ubuntu-2004-kube-v1.25.7+vmware.2-tkg.1-8a74b9f12e488c54605b3537acb683bc.ova`
  and we want to use it, the template that we need to obtain corresponds to v1.25.7, that is `cluster-template-v1.25.7.yaml`.

* This template requires some extra elements to be added to the `kind: Cluster` block, inside `metadata`. These elements are `labels` and
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

* The template also requires a few extra snippets to be added that are not present in the original YAML and are essential for
  CSE v4.1 to work properly. These are already present in the [proposed example][cluster] cluster YAML:

  * In `kind: KubeadmConfigTemplate`: One must add the `preKubeadmCommands` and `useExperimentalRetryJoin` blocks under the `spec > users` section:
```yaml
      preKubeadmCommands:
        - mv /etc/ssl/certs/custom_certificate_*.crt
          /usr/local/share/ca-certificates && update-ca-certificates
      useExperimentalRetryJoin: true
```

* In `kind: KubeadmControlPlane`: One must add the `preKubeadmCommands` and `controllerManager` blocks under the `kubeadmConfigSpec` section:
```yaml
      preKubeadmCommands:
        - mv /etc/ssl/certs/custom_certificate_*.crt
          /usr/local/share/ca-certificates && update-ca-certificates
      controllerManager:
        extraArgs:
          enable-hostpath-provisioner: "true"
```

* The downloaded template does not have any **Node Health Checks**, which is a feature introduced in v4.1. You must add the following 
  block of kind `MachineHealthCheck` to be able to use it. The placeholders are already being populated in the [proposed example][cluster] with
  the CSE Server configuration parameters set during the installation process:

```yaml
apiVersion: cluster.x-k8s.io/v1beta1
kind: MachineHealthCheck
metadata:
  name: ${CLUSTER_NAME}
  namespace: ${TARGET_NAMESPACE}
  labels:
    clusterctl.cluster.x-k8s.io: ""
    clusterctl.cluster.x-k8s.io/move: ""
spec:
  clusterName: ${CLUSTER_NAME}
  maxUnhealthy: ${MAX_UNHEALTHY_NODE_PERCENTAGE}%
  nodeStartupTimeout: ${NODE_STARTUP_TIMEOUT}s
  selector:
    matchLabels:
      cluster.x-k8s.io/cluster-name: ${CLUSTER_NAME}
  unhealthyConditions:
    - type: Ready
      status: Unknown
      timeout: ${NODE_UNKNOWN_TIMEOUT}s
    - type: Ready
      status: "False"
      timeout: ${NODE_NOT_READY_TIMEOUT}s
---
```

* To use a Control Plane IP and/or a Virtual IP Subnet, you must add the following snippets to the kind `VCDCluster` spec section. The [proposed example][cluster]
  is not using this feature, so you must add a correct value for the `CONTROL_PLANE_IP` and/or `VIRTUAL_IP_SUBNET` placeholders:

```yaml
  controlPlaneEndpoint:
    host: ${CONTROL_PLANE_IP}
    port: 6443
```
```yaml
  loadBalancerConfigSpec:
    vipSubnet: ${VIRTUAL_IP_SUBNET}
```

* The downloaded template has a single worker pool (to see an example with **two** worker pools, please check the [proposed example][cluster]).
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

* `CLUSTER_NAME`: This will be the TKGm cluster name. It must contain only lowercase alphanumeric characters or '-',
start with an alphabetic character, end with an alphanumeric, and contain at most 31 characters.
* `TARGET_NAMESPACE`: This will be the TKGm cluster namespace. In [the example][cluster] the value is
`"${var.k8s_cluster_name}-ns"`, which mimics the UI behaviour, as the namespace is the name of the TKGm cluster concatenated with `-ns`.
* `VCD_SITE`: The VCD URL, the same that was used during CSE installation.
* `VCD_ORGANIZATION`: The Organization in which the TKGm clusters will be created. In this guide it was created as `tenant_org` and named
"Tenant Organization" during CSE installation phase.
* `VCD_ORGANIZATION_VDC`: The VDC in which the TKGm clusters will be created. In this guide it was created as `tenant_vdc` and named
  "Tenant VDC" during CSE installation phase.
* `VCD_ORGANIZATION_VDC_NETWORK`: The VDC network that the TKGm clusters will use. In this guide it was created as a Routed
  network called `tenant_net_routed`.
* `VCD_USERNAME_B64`: The name of a user with the "Kubernetes Cluster Author" role (`k8s_cluster_author`) that was created during CSE installation.
It must be encoded in Base64.
* `VCD_PASSWORD_B64` (**Discouraged in favor of `VCD_REFRESH_TOKEN_B64`**): The password of the user above.
  It must be encoded in Base64. Please do **not** use this value (by setting it to `""`) and use `VCD_REFRESH_TOKEN_B64` instead.
* `VCD_REFRESH_TOKEN_B64`: An API token that belongs to the user above. In UI, the API tokens can be generated in the user preferences
  in the top right, then go to the API tokens section, add a new one. Or we can visit `/tenant/<TENANT-NAME>/administration/settings/user-preferences`
  in the target VCD, logged in as the cluster author user in the desired tenant. It must be encoded in Base64.
* `SSH_PUBLIC_KEY`: We can set a public SSH key to be able to debug the TKGm control plane nodes. It can be empty (`""`)
* `CONTROL_PLANE_MACHINE_COUNT`: Number of control plane nodes (VMs). **Must be an odd number and higher than 0**.
* `VCD_CONTROL_PLANE_SIZING_POLICY`: Name of an existing VM Sizing Policy, created during CSE installation. Can be empty to use the VDC default (`""`)
* `VCD_CONTROL_PLANE_PLACEMENT_POLICY` : Name of an existing VM Placement Policy. Can be empty (`""`)
* `VCD_CONTROL_PLANE_STORAGE_PROFILE`: Name of an existing Storage Profile, for example `"*"` to use the default.
* `WORKER_MACHINE_COUNT`: Number of worker nodes (VMs). **Must be higher than 0**.
* `VCD_WORKER_SIZING_POLICY`: Name of an existing VM Sizing Policy, created during CSE installation. Can be empty to use the VDC default (`""`)
* `VCD_WORKER_PLACEMENT_POLICY`: Name of an existing VM Placement Policy. Can be empty (`""`)
* `VCD_WORKER_STORAGE_PROFILE`: Name of an existing Storage Profile, for example `"*"` to use the default.
* `DISK_SIZE`: Specifies the storage size for each node (VM). It uses the [same units as every other Kubernetes resource](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/),
  for example `"1Gi"` to use 1 gibibyte (1024 MiB), or `"1G"` for 1 gigabyte (1000 MB).
* `VCD_CATALOG`: The catalog where the TKGm OVAs are. For example, in the above CSE installation it was `tkgm_catalog`.
* `VCD_TEMPLATE_NAME` = The TKGm OVA name, for example `ubuntu-2004-kube-v1.25.7+vmware.2-tkg.1-8a74b9f12e488c54605b3537acb683bc`
* `POD_CIDR`: The CIDR used for Pod networking, for example `"100.96.0.0/11"`.
* `SERVICE_CIDR`: The CIDR used for Service networking, for example `"100.64.0.0/13"`.

There are some extra variables that are obtained from the `VCDKEConfig` RDE that was created during CSE installation.
In the [proposed example][cluster] they are read with a `vcd_rde` data source:

* `NODE_STARTUP_TIMEOUT`: A node will be considered unhealthy and remediated if joining the cluster takes longer than this timeout (seconds, defaults to 900 in the sample configuration).
* `NODE_NOT_READY_TIMEOUT`: A newly joined node will be considered unhealthy and remediated if it cannot host workloads for longer than this timeout (seconds, defaults to 300 in the sample configuration).
* `NODE_UNKNOWN_TIMEOUT`: A healthy node will be considered unhealthy and remediated if it is unreachable for longer than this timeout (seconds, defaults to 300 in the sample configuration).
* `MAX_UNHEALTHY_NODE_PERCENTAGE`: Remediation will be suspended when the number of unhealthy nodes exceeds this percentage.
  (100% means that unhealthy nodes will always be remediated, while 0% means that unhealthy nodes will never be remediated). Defaults to 100 in the sample configuration.
* `CONTAINER_REGISTRY_URL`: URL from where TKG clusters will fetch container images, useful for VCD appliances that are completely isolated from Internet. Defaults to "projects.registry.vmware.com" in the sample configuration.

There are two additional variables that we added manually: `TKR_VERSION` and `TKGVERSION`.
To know their values, one can use [this script](https://github.com/vmware/cluster-api-provider-cloud-director/blob/main/docs/WORKLOAD_CLUSTER.md#script-to-get-kubernetes-etcd-coredns-versions-from-tkg-ova),
or check the following table with some script outputs:

| OVA name                                                 | TKR_VERSION               | TKGVERSION |
|----------------------------------------------------------|---------------------------|------------|
| v1.25.7+vmware.2-tkg.1-8a74b9f12e488c54605b3537acb683bc  | v1.25.7---vmware.2-tkg.1  | v2.2.0     |
| v1.24.11+vmware.1-tkg.1-2ccb2a001f8bd8f15f1bfbc811071830 | v1.24.11---vmware.1-tkg.1 | v2.2.0     |
| v1.24.10+vmware.1-tkg.1-765d418b72c247c2310384e640ee075e | v1.24.10---vmware.1-tkg.2 | v2.1.1     |
| v1.23.17+vmware.1-tkg.1-ee4d95d5d08cd7f31da47d1480571754 | v1.23.17---vmware.1-tkg.1 | v2.2.0     |
| v1.23.16+vmware.1-tkg.1-eb0de9755338b944ea9652e6f758b3ce | v1.23.16---vmware.1-tkg.1 | v2.1.1     |
| v1.22.17+vmware.1-tkg.1-df08b304658a6cf17f5e74dc0ab7543c | v1.22.17---vmware.1-tkg.1 | v2.1.1     |

The final two variables are optional, `CONTROL_PLANE_IP` and `VIRTUAL_IP_SUBNET`.
If one decided to add the Control Plane IP and the Virtual IP Subnet to the `kind: VCDCluster` block in the CAPVCD YAML
that was explained above, then these will be mandatory and must be set with a valid IP and CIDR respectively.

In [the TKGm cluster creation example][cluster], the built-in Terraform function `templatefile` is used to substitute every placeholder
mentioned above with its final value. The returned value is the CAPVCD YAML payload that needs to be set in the `capi_yaml` placeholder in the
JSON template.

~> Notice that we need to use the Terraform built-in function `jsonencode` with the final CAPVCD YAML payload, so all 
special characters are correctly escaped in the payload. This is also done in [the mentioned example][cluster].

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
the Kubeconfig is available and ready to use. It can now be retrieved it with the `vcd_rde_behavior_invocation` data source by adding
the following snippet to the existing configuration:

```hcl
data "vcd_rde_behavior_invocation" "get_kubeconfig" {
  rde_id      = vcd_rde.k8s_cluster_instance.id
  behavior_id = "urn:vcloud:behavior-interface:getFullEntity:cse:capvcd:1.0.0"
}

output "kubeconfig" {
  value = jsondecode(data.vcd_rde_behavior_invocation.get_kubeconfig.result)["entity"]["status"]["capvcd"]["private"]["kubeConfig"]
}
```

Then, running `terraform output -no-color -raw kubeconfig > kubeconfig.yaml` should give you a completely operational KubeConfig YAML file.

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

* TKGm OVA: If there is a newer version of TKGm, we can modify the referenced OVA.
* Number of worker nodes. Remember this must be higher than 0.
* Number of control plane nodes. Remember this must be an odd number and higher than 0.

### Upgrade a cluster to CSE v4.1

To upgrade a cluster from CSE v4.0 to v4.1, first of all you need to change the RDE Type that the TKGm cluster `vcd_rde` uses:

```
rde_type_id = data.vcd_rde_type.capvcdcluster_type_v1_2_0.id # This must reference the CAPVCD RDE Type v1.2.0
```

Then, before updating, please revisit the ["Creating a Kubernetes cluster"](#creating-a-kubernetes-cluster) and
["Updating a Kubernetes cluster"](#updating-a-kubernetes-cluster) sections, to be sure that you consider the new features
and requirements of v4.1 that need to be included in the CAPVCD YAML:

* Adding the `MachineHealthCheck` section to the cluster template YAML to use CSE v4.1 health checking capabilities.
* Adding the needed `preKubeadmCommands` sections to the cluster template YAML.
* Updating to a supported TKGm OVA (see the table above with the supported versions).

With the new CAPVCD YAML, you need to get the actual cluster state from the `vcd_rde` `computed_entity` attribute and
create a new value for the input `entity` argument. Follow the steps mentioned in ["Updating a Kubernetes cluster"](#updating-a-kubernetes-cluster).

## Deleting a Kubernetes cluster

~> Do **NOT** remove the cluster from the HCL configuration! This will leave dangling resources that the CSE Server creates
when the TKGm cluster is created, such as vApps, networks, virtual services, etc. Please follow the procedure described in
this section to destroy a cluster entirely.

To delete an existing TKGm cluster, one needs to mark it for deletion in the `vcd_rde` resource. In the [example configuration][cluster],
there are two keys `delete` and `force_delete` that correspond to the CAPVCD [RDE Type][rde_type] schema fields `markForDelete`
and `forceDelete` respectively.

* Setting `delete = true` will make the CSE Server remove the cluster from VCD and eventually the RDE.
* Setting also `force_delete = true` will force the CSE Server to delete the cluster, and their associated resources
  that are not fully complete and that are in an unremovable state.

Follow the instructions described in ["Updating a Kubernetes cluster"](#updating-a-kubernetes-cluster) to learn how to perform
the update of these two properties.

Once updated, one can monitor the `vcd_rde` resource to check the deletion process. Eventually, the RDE won't exist anymore in VCD and Terraform will
ask for creation again. It can be now removed from the HCL configuration.

[api_token]: /providers/vmware/vcd/latest/docs/resources/api_token
[capvcd]: https://github.com/vmware/cluster-api-provider-cloud-director
[capvcd_templates]: https://github.com/vmware/cluster-api-provider-cloud-director/tree/main/templates
[cluster]: https://github.com/vmware/terraform-provider-vcd/tree/main/examples/container-service-extension/v4.1/cluster
[cse_install_guide]: /providers/vmware/vcd/latest/docs/guides/container_service_extension_4_x_install
[cse_docs]: https://docs.vmware.com/en/VMware-Cloud-Director-Container-Service-Extension/index.html
[rde]: /providers/vmware/vcd/latest/docs/resources/rde
[rde_input_vs_computed]: /providers/vmware/vcd/latest/docs/resources/rde#input-entity-vs-computed-entity
[rde_type]: /providers/vmware/vcd/latest/docs/resources/rde_type
[tkgmcluster_template]: https://github.com/vmware/terraform-provider-vcd/tree/main/examples/container-service-extension/v4.1/entities/tkgmcluster.json.template
