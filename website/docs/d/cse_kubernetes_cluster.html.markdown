---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_cse_kubernetes_cluster"
sidebar_current: "docs-vcd-data-source-cse-kubernetes-cluster"
description: |-
  Provides a resource to read Kubernetes clusters from VMware Cloud Director with Container Service Extension installed and running.
---

# vcd\_cse\_kubernetes\_cluster

Provides a data source to read Kubernetes clusters in VMware Cloud Director with Container Service Extension (CSE) installed and running.

Supported in provider *v3.12+*

Supports the following **Container Service Extension** versions:

* 4.2

-> To install CSE in VMware Cloud Director, please follow [this guide](/providers/vmware/vcd/latest/docs/guides/container_service_extension_4_x_install)

## Example Usage

```hcl
data "vcd_cse_kubernetes_cluster" "my_cluster" {
  cluster_id  = "urn:vcloud:entity:vmware:capvcdCluster:e8e82bcc-50a1-484f-9dd0-20965ab3e865"
}
```

## Argument Reference

The following arguments are supported:

* `cluster_id` - (Required) Unequivocally identifies a cluster in VCD

## Attribute Reference

All attributes defined in [vcd_cse_kubernetes_cluster](/providers/vmware/vcd/latest/docs/resources/cse_kubernetes_cluster) resource are supported.
Also, the resource arguments are also available as read-only attributes.
