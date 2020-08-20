---
layout: "vcd"
page_title: "vCloudDirector: vcd_org_user"
sidebar_current: "docs-vcd-datasource-org-user"
description: |-
  Provides a vCloud Director Organization user data source. This can be used to read organization users.
---

# vcd\_org\_user

Provides a vCloud Director Org User data source. This can be used to read organization users, including org administrators.

Supported in provider *v2.10+*


## Example Usage

```hcl
data "vcd_org_user" "my-org-admin" {
  org  = "my-org"
  name = "my-org-admin"
}

data "vcd_org_user" "my-vapp-creator" {
  org     = "my-org"
  user_id = "urn:vcloud:user:c311eb35-6984-4d26-3ee9-0000deadbeef"
}

output "admin_user" {
  value = data.vcd_org_user.my-org-admin
}

output "vapp_creator_user" {
  value = data.vcd_org_user.my-vapp-creator
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to which the user belongs. Optional if defined at provider level.
* `name` - (Optional) The name of the user. Required if `user_id` is not set
* `user_id` - (Optional) The ID of the user. Required if `name` is not set

## Attribute reference

~> **Note:** passwords are never returned with the user data.

* `provider_type` - Identity provider type for this user. 
* `role` - The role of the user. 
* `full_name` - The full name of the user.
* `description` - An optional description of the user.
* `telephone` - The Org User telephone number.
* `email_address` - The Org User email address.
* `instant_messaging` - The Org User instant messaging.
* `enabled` - True if the user is enabled and can log in.
* `is_group_role` - True if this user has a group role.
* `is_locked` - If the user account has been locked due to too many invalid login attempts, the value will be true. 
* `deployed_vm_quota` - Quota of vApps that this user can deploy. A value of 0 specifies an unlimited quota.
* `stored_vm_quota` -  Quota of vApps that this user can store. A value of 0 specifies an unlimited quota.
* `id` - The ID of the Organization user

