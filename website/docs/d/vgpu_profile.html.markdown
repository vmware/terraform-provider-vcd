---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_vgpu_profile"
sidebar_current: "docs-vcd-datasource-vgpu-policy"
description: |-
  Provides a datasource to read vGPU profiles in VMware Cloud Director.
---

# vcd\_vm\_vgpu\_profile

Supported in provider *3.11* and VCD *10.4.0+*.

Provides a datasource to read vGPU profiles in VMware Cloud Director.

## Example Usage

```hcl
data "vcd_vgpu_profile" "profile-name" {
  name = "my-profile"
}

output "profileId" {
  value = data.vcd_vgpu_profile.profile-name.id
}
```
## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the vGPU profile.

## Attribute Reference

* `id` - ID of the vGPU profile.
* `tenant_facing_name` - Tenant facing name of the vGPU profile.
* `instructions` - Instructions for the vGPU profile.

