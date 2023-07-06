---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_ip_space_custom_quota"
sidebar_current: "docs-vcd-resource-ip-space-custom-quota"
description: |-
  Provides a resource to manage Custom Quotas for a given Org in a particular IP Space if one wants 
  to override default quota set for IP Space.
---

# vcd\_ip\_space\_custom\_quota

Provides a resource to manage Custom Quotas for a given Org in a particular IP Space if one wants to
override default quota set for IP Space.

~> Only `System Administrator` can create this resource.

<a id="example-1"></a>
## Example Usage (Custom IP Space Quota for a particular Org)

```hcl
resource "vcd_ip_space_custom_quota" "q1" {
  org_id      = data.vcd_org.org1.id
  ip_space_id = vcd_ip_space.space1.id

  ip_range_quota = 23

  ip_prefix_quota {
    prefix_length = 29
    quota         = 17
  }

  ip_prefix_quota {
    prefix_length = 30
    quota         = 15
  }

  # Custom Quota can only be configured once Edge Gateway is created
  depends_on = [vcd_nsxt_edgegateway.ip-space]
}
```

## Argument Reference

The following arguments are supported:

* `ip_space_id` - (Required) - IP Space ID to set Custom Quotas
* `org_id` - (Required) Organization ID, for which the Quota should be customized
* `ip_range_quota` - (Optional) Floating IP Quota. Will inherit the default Quota set in
  `vcd_ip_space` if not set
* `ip_prefix_quota` - (Optional) IP Prefix Quota set in [ip_prefix_quota](#ip-prefix-quota) blocks.
  Will inherit the default Quota set in `vcd_ip_space` if not set

~> The resource `vcd_ip_space_custom_quota` can only be created for an Org after an NSX-T Edge
Gateway backed by Provider Gateway is created within the Org. An explicit `depends_on` constraint
for an Edge Gateway to exist might be required. (See the [example](#example-1).)

<a id="ip-prefix-quota"></a>

## ip_prefix_quota block

* `prefix_length` - (Required) - Prefix length for which the quota should be set (must be an
  existing prefix length in parent IP Space)
* `quota` - (Required) - Quota value for specific *prefix_length*


## Importing

~> The current implementation of Terraform import can only import resources into the state.
It does not generate configuration. [More information.](https://www.terraform.io/docs/import/)

An existing Custom IP Space Quota configuration can be [imported][docs-import] into this resource
via supplying path for it. An example is
below:

[docs-import]: https://www.terraform.io/docs/import/

```
terraform import vcd_ip_space_custom_quota.imported ip-space-name.org-name
```

The above would import the Custom Quota defined for Org `org-name` in IP Space `ip-space-name`.
