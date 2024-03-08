---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_cse_kubernetes_cluster"
sidebar_current: "docs-vcd-data-source-cse-kubernetes-cluster"
description: |-
  Provides a data source to read Kubernetes clusters from VMware Cloud Director with Container Service Extension installed and running.
---

# vcd\_cse\_kubernetes\_cluster

Provides a data source to read Kubernetes clusters in VMware Cloud Director with Container Service Extension (CSE) installed and running.

Supported in provider *v3.12+*

Supports the following **Container Service Extension** versions:

* 4.1.0
* 4.1.1
* 4.2.0
* 4.2.1

-> To install CSE in VMware Cloud Director, please follow [this guide](/providers/vmware/vcd/latest/docs/guides/container_service_extension_4_x_install)

## Example Usage with ID

The cluster ID identifies unequivocally the cluster within VCD, and can be obtained with the CSE Kubernetes Clusters UI Plugin, by selecting
the desired cluster and obtaining the ID from the displayed information.

```hcl
data "vcd_cse_kubernetes_cluster" "my_cluster" {
  cluster_id = "urn:vcloud:entity:vmware:capvcdCluster:e8e82bcc-50a1-484f-9dd0-20965ab3e865"
}
```

## Example Usage with Name

Sometimes using the cluster ID is not convenient, so this data source allows using the cluster name.
As VCD allows to have multiple clusters with the same name, this option must be used with caution as it will fail
if there is more than one Kubernetes cluster with the same name in the same Organization:

```hcl
locals {
  my_clusters = toset(["my-cluster-1", "my-cluster-2", "my-cluster-3"])
}

data "vcd_cse_kubernetes_cluster" "my_clusters" {
  for_each    = local.my_clusters
  org_id      = data.vcd_org.org.id
  cse_version = "4.2.1"
  name        = each.key
}
```

## Argument Reference

The following arguments are supported:

* `cluster_id` - (Optional) Unequivocally identifies a cluster in VCD. Either `cluster_id` or `name` must be set.
* `name` - (Optional) Allows to find a Kubernetes cluster by name inside the given Organization with ID `org_id`. Either `cluster_id` or `name` must be set. This argument requires `cse_version` and `org_id` to be set.
* `org_id` - (Optional) The ID of the Organization to which the Kubernetes cluster belongs. Only used if `cluster_id` is not set. Must be present if `name` is used.
* `cse_version` - (Optional) Specifies the CSE Version of the cluster to find when `name` is used instead of `cluster_id`.

## Attribute Reference

All attributes defined in [vcd_cse_kubernetes_cluster](/providers/vmware/vcd/latest/docs/resources/cse_kubernetes_cluster) resource are supported.
Also, the resource arguments are also available as read-only attributes.
