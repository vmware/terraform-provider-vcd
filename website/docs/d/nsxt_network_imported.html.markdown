---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_nsxt_network_imported"
sidebar_current: "docs-vcd-data-source-nsxt-network-imported"
description: |-
  Provides a VMware Cloud Director Org VDC NSX-T Imported Network data source to read data or reference existing network.
---

# vcd\_nsxt\_network\_imported

Provides a VMware Cloud Director Org VDC NSX-T Imported Network data source to read data or reference existing network.

Supported in provider *v3.2+* for NSX-T VDCs only.

-> This is **not Terraform imported** data source, but a special **Imported** type of **Org VDC network** in NSX-T VDC. Read more about Imported Network in [official VCD documentation]((https://docs.vmware.com/en/VMware-Cloud-Director/10.2/VMware-Cloud-Director-Tenant-Portal-Guide/GUID-FB303D62-67EA-4209-BE4D-C3746481BCC8.html).

## Example Usage

```hcl
data "vcd_nsxt_network_imported" "net" {
  org  = "my-org" # Optional
  vdc  = "my-vdc" # Optional
  name = "my-net"
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to use, optional if defined at provider level
* `vdc` - (Optional) The name of VDC to use, optional if defined at provider level
* `name` - (Required) A unique name for the network (optional when `filter` is used)
* `filter` - (Optional) Retrieves the data source using one or more filter parameters

## Attribute reference

All attributes defined in [imported network resource](/docs/providers/vcd/r/nsxt_network_imported.html#attribute-reference) are supported.

## Filter arguments

* `name_regex` (Optional) matches the name using a regular expression.
* `ip` (Optional) matches the IP of the resource using a regular expression.

See [Filters reference](/docs/providers/vcd/guides/data_source_filters.html) for details and examples.
