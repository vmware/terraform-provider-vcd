---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_tm_edge_cluster"
sidebar_current: "docs-vcd-data-source-tm-edge-cluster"
description: |-
  Provides a VMware Cloud Foundation Tenant Manager Edge Cluster data source.
---

# vcd\_tm\_edge\_cluster

Provides a VMware Cloud Foundation Tenant Manager Edge Cluster data source.

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
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of Edge Cluster
* `region_id` - (Required) The ID of parent region. Can be looked up using
  [`vcd_tm_region`](/providers/vmware/vcd/latest/docs/data-sources/tm_region) data source
* `sync_before_read` - (Optional) Set to true to trigger Sync before attempting to search for Edge
  Cluster . Default `false`.

## Attribute Reference

* `node_count` - Number of transport nodes in the Edge Cluster. If this information is not
  available, it will be set to `-1`
* `org_count` - Number of organizations using this Edge Cluster
* `vpc_count` - Number of VPCs using this Edge Cluster
* `average_cpu_usage_percentage` - Average CPU utilization percentage across all member nodes
* `average_memory_usage_percentage` - Average RAM utilization percentage across all member nodes
* `health_status` - Current health status of Edge Cluster. One of:
 * `UP` - The Edge Cluster is healthy
 * `DOWN` - The Edge Cluster is down
 * `DEGRADED` - The Edge Cluster is not operating at capacity. One or more member nodes are down or inactive
 * `UNKNOWN` - The Edge Cluster state is unknown. If UNKNOWN, `average_cpu_usage_percentage` and `average_memory_usage_percentage` will be not be set
* `status` - Represents current status of the networking entity. One of:
 * `PENDING` - Desired entity configuration has been received by system and is pending realization
 * `CONFIGURING` - The system is in process of realizing the entity
 * `REALIZED` - The entity is successfully realized in the system
 * `REALIZATION_FAILED` - There are some issues and the system is not able to realize the entity
 * `UNKNOWN` - Current state of entity is unknown
* `deployment_type` - Deployment type for transport nodes in the Edge Cluster. Possible values are:
 * `VIRTUAL_MACHINE` - If all members are of type _VIRTUAL_MACHINE_
 * `PHYSICAL_MACHINE` - If all members are of type _PHYSICAL_MACHINE_
 * `UNKNOWN` - If there are no members or their type is not known
