---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_nsxt_edge_cluster"
sidebar_current: "docs-vcd-data-source-nsxt-edge-cluster"
description: |-
  Provides a data source for available NSX-T Edge Clusters.
---

# vcd\_nsxt\_edge\_cluster

Provides a data source for available NSX-T Edge Clusters.

Supported in provider *v3.1+*

~> **Note:** This resource uses new VMware Cloud Director
[OpenAPI](https://code.vmware.com/docs/11982/getting-started-with-vmware-cloud-director-openapi) and
requires at least VCD *10.1.1+* and NSX-T *3.1+*.

## Example Usage 

```hcl
data "vcd_nsxt_edge_cluster" "first" {
  name = "edge-cluster-one"
}
```


## Argument Reference

The following arguments are supported:

* `name` - (Required) NSX-T Edge Cluster name.

## Attribute reference

* `description` - Edge Cluster description in NSX-T manager.
* `node_count` - Number of nodes in Edge Cluster.
* `node_type` - Type of nodes in Edge Cluster.
* `deployment_type` - Deployment type of Edge Cluster.