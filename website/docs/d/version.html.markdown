---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_version"
sidebar_current: "docs-vcd-data-source-version"
description: |-
  Provides a VCD version data source.
---

# vcd\_version

Provides a VMware Cloud Director version data source to fetch the VCD version, the maximum API version and perform some optional
checks with version constraints.

Supported in provider *v3.12+*. Requires System Administrator privileges.

## Example Usage

```hcl
# This data source will assert that the VCD version is exactly 10.5.1, otherwise it will fail
data "vcd_version" "gte_1051" {
  condition         = "= 10.5.1"
  fail_if_not_match = true
}
```

## Argument Reference

The following arguments are supported:

* `condition` - (Optional) A version constraint to check against the VCD version
* `fail_if_not_match` - (Optional) Required if `condition` is set. Throws an error if the version constraint set in `condition` is not met

## Attribute Reference

* `matches_condition` - It is true if the VCD version matches the constraint set in `condition`
* `vcd_version` - The VCD version
* `api_version` - The maximum supported API version
