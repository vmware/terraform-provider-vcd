---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_tm_edge_cluster_qos"
sidebar_current: "docs-vcd-resource-tm-edge-cluster-qos"
description: |-
  Provides a VMware Cloud Foundation Tenant Manager Edge Cluster QoS.
---

# vcd\_tm\_edge\_cluster\_qos

Provides a VMware Cloud Foundation Tenant Manager Edge Cluster QoS.

-> This resource does not actually create an Edge Cluster QoS, but configures QoS for a given
`edge_cluster_id`. Similarly, `terraform destroy` operation does not remove Edge Cluster, but resets
QoS settings to unlimited. 

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

  egress_committed_bandwidth_mbps = 1
  egress_burst_size_bytes = 2
  ingress_committed_bandwidth_mbps = 3
  ingress_burst_size_bytes = 4
}
```

## Argument Reference

The following arguments are supported:

* `edge_cluster_id` - (Required) An ID of Edge Cluster. Can be looked up using
  [vcd_tm_edge_cluster](/providers/vmware/vcd/latest/docs/data-sources/tm_edge_cluster) data source
* `egress_committed_bandwidth_mbps` - (Optional) Committed egress bandwidth specified in Mbps. Bandwidth is
	limited to line rate when the value configured is greater than line rate. Traffic exceeding
	bandwidth will be dropped. 
* `egress_burst_size_bytes` - (Optional) Egress burst size in bytes
* `ingress_committed_bandwidth_mbps` - (Optional) Committed ingress bandwidth specified in Mbps.
	Bandwidth is limited to line rate when the value configured is greater than line rate. Traffic
	exceeding bandwidth will be dropped. 
* `ingress_burst_size_bytes` - (Optional) Ingres burst size in bytes

-> Choosing to set _egress_ or _ingress_ requires both of that traffic direction must be set

## Attribute Reference

The following attributes are exported on this resource:

* `max_virtual_services` - Maximum number of virtual services this NSX-T ALB Service Engine Group can run
* `reserved_virtual_services` - Number of reserved virtual services
* `deployed_virtual_services` - Number of deployed virtual services
* `ha_mode` defines High Availability Mode for Service Engine Group. One off:
  * ELASTIC_N_PLUS_M_BUFFER - Service Engines will scale out to N active nodes with M nodes as buffer.
  * ELASTIC_ACTIVE_ACTIVE - Active-Active with scale out.
  * LEGACY_ACTIVE_STANDBY - Traditional single Active-Standby configuration

## Importing

~> **Note:** The current implementation of Terraform import can only import resources into the
state. It does not generate configuration. However, an experimental feature in Terraform 1.5+ allows
also code generation. See [Importing resources][importing-resources] for more information.

An existing IP Space configuration can be [imported][docs-import] into this resource via supplying
path for it. An example is below:

[docs-import]: https://www.terraform.io/docs/import/

```
terraform import vcd_tm_ip_space.imported my-ip-space-name
```

The above would import the `my-ip-space-name` IP Space.