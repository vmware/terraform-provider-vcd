---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_tm_edge_cluster_qos"
sidebar_current: "docs-vcd-resource-tm-edge-cluster-qos"
description: |-
  Provides a VMware Cloud Foundation Tenant Manager Edge Cluster QoS resource.
---

# vcd\_tm\_edge\_cluster\_qos

Provides a VMware Cloud Foundation Tenant Manager Edge Cluster QoS resource.

-> This resource does not create an Edge Cluster QoS entity, but configures QoS for a given
`edge_cluster_id`. Similarly, `terraform destroy` operation does not remove Edge Cluster, but resets
QoS settings to default (unlimited). 

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

resource "vcd_tm_edge_cluster_qos" "demo" {
  edge_cluster_id = data.vcd_tm_edge_cluster.demo.id

  egress_committed_bandwidth_mbps  = 1
  egress_burst_size_bytes          = 2
  ingress_committed_bandwidth_mbps = 3
  ingress_burst_size_bytes         = 4
}
```

## Argument Reference

The following arguments are supported:

* `edge_cluster_id` - (Required) An ID of Edge Cluster. Can be looked up using
  [vcd_tm_edge_cluster](/providers/vmware/vcd/latest/docs/data-sources/tm_edge_cluster) data source
* `egress_committed_bandwidth_mbps` - (Optional) Committed egress bandwidth specified in Mbps.
  Bandwidth is limited to line rate. Traffic exceeding bandwidth will be dropped. Required with
  `egress_burst_size_bytes` 
* `egress_burst_size_bytes` - (Optional) Egress burst size in bytes. Required with
  `egress_committed_bandwidth_mbps`
* `ingress_committed_bandwidth_mbps` - (Optional) Committed ingress bandwidth specified in Mbps.
  Bandwidth is limited to line rate. Traffic exceeding bandwidth will be dropped. Required with
  `ingress_burst_size_bytes`
* `ingress_burst_size_bytes` - (Optional) Ingress burst size in bytes. Required with
  `ingress_committed_bandwidth_mbps`

## Importing

~> **Note:** The current implementation of Terraform import can only import resources into the
state. It does not generate configuration. However, an experimental feature in Terraform 1.5+ allows
also code generation. See [Importing resources][importing-resources] for more information.

An existing IP Space configuration can be [imported][docs-import] into this resource via supplying
path for it. An example is below:

[docs-import]: https://www.terraform.io/docs/import/

```
terraform import vcd_tm_edge_cluster_qos.imported my-region-name.my-edge-cluster-name
```

The above would import the `my-edge-cluster-name` Edge Cluster QoS settings that is in
`my-region-name` Region.
