---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_tm_edge_cluster_qos"
sidebar_current: "docs-vcd-data-source-tm-edge-cluster-qos"
description: |-
  Provides a VMware Cloud Foundation Tenant Manager Edge Cluster QoS data source.
---

# vcd\_tm\_edge\_cluster\_qos

Provides a VMware Cloud Foundation Tenant Manager Edge Cluster QoS data source.

## Example Usage

```hcl
data "vcd_tm_region" "demo" {
  name = "region-one"
}

data "vcd_tm_edge_cluster" "demo" {
  name             = "my-edge-cluster"
  region_id        = data.vcd_tm_region.demo.id
  sync_before_read = true
}

data "vcd_tm_edge_cluster_qos" "demo" {
  edge_cluster_id = data.vcd_tm_edge_cluster.demo.id
}
```

## Argument Reference

The following arguments are supported:

* `edge_cluster_id` - (Required) An ID of Edge Cluster. Can be looked up using
  [vcd_tm_edge_cluster](/providers/vmware/vcd/latest/docs/data-sources/tm_edge_cluster) data source

## Attribute Reference

All the arguments and attributes defined in
[`vcd_tm_edge_cluster_qos`](/providers/vmware/vcd/latest/docs/resources/tm_edge_cluster_qos) resource are available.