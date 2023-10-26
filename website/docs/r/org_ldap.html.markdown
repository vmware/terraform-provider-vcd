---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_org_ldap"
sidebar_current: "docs-vcd-resource-org-ldap"
description: |-
  Provides a VMware Cloud Director Organization LDAP resource. This can be used to create, delete, and update LDAP configuration for an organization .
---

# vcd\_org\_ldap

Provides a VMware Cloud Director Org LDAP resource. This can be used to create, update, and delete LDAP configuration for an organization.

Supported in provider *v3.8+*

-> **Note:** This resource requires system administrator privileges.

## Example Usage 1 - Custom configuration

```hcl
provider "vcd" {
  user     = var.admin_user
  password = var.admin_password
  org      = "System"
  url      = "https://AcmeVcd/api"
}

data "vcd_org" "my-org" {
  name = "my-org"
}

# The settings below (except the server IP) are taken from the LDAP docker testing image
# https://github.com/rroemhild/docker-test-openldap
resource "vcd_org_ldap" "my-org-ldap" {
  org_id    = data.vcd_org.my-org.id
  ldap_mode = "CUSTOM"
  custom_settings {
    server                  = "192.168.1.172"
    port                    = 389
    is_ssl                  = false
    username                = "cn=admin,dc=planetexpress,dc=com"
    password                = "GoodNewsEveryone"
    authentication_method   = "SIMPLE"
    base_distinguished_name = "dc=planetexpress,dc=com"
    connector_type          = "OPEN_LDAP"
    user_attributes {
      object_class                = "inetOrgPerson"
      unique_identifier           = "uid"
      display_name                = "cn"
      username                    = "uid"
      given_name                  = "givenName"
      surname                     = "sn"
      telephone                   = "telephoneNumber"
      group_membership_identifier = "dn"
      email                       = "mail"
    }
    group_attributes {
      name                        = "cn"
      object_class                = "group"
      membership                  = "member"
      unique_identifier           = "cn"
      group_membership_identifier = "dn"
    }
  }
}
```

-> **Note** 
The password value never gets returned by GET. Therefore, if we want `terraform plan` to return a clean state, we need
to add a `lifecycle` block at the end of the resource definition, after creating or updating it.
And we need to remove the `lifecycle` block _if we want to change the password_.

```hcl
resource "vcd_org_ldap" "my-org-ldap" {
  # all other fields
  # ...
  lifecycle {
    # password value does not get returned by GET
    ignore_changes = [custom_settings[0].password]
  }
}
```

## Example Usage 2 - Using system configuration

```hcl
data "vcd_org" "my-org" {
  name = "my-org"
}

resource "vcd_org_ldap" "my-org-ldap" {
  org_id         = data.vcd_org.my-org.id
  ldap_mode      = "SYSTEM"
  custom_user_ou = "ou=Foo,dc=domain,dc=local base DN"
}
```

## Argument Reference

The following arguments are supported:

* `org_id` - (Required) Org ID: there is only one LDAP configuration available for an organization. Thus, the resource can be identified by the Org.
* `ldap_mode` - (Required) One of `NONE`, `CUSTOM`, `SYSTEM`. Note that using `NONE` has the effect of removing the LDAP settings
* `custom_user_ou` - (Optional; *v3.11+*) If `ldap_mode` is `SYSTEM`, specifies a LDAP `attribute=value` pair to use for OU (organizational unit)
* `custom_settings` - (Optional) LDAP server configuration. Becomes mandatory if `ldap_mode` is set to `CUSTOM`. See [Custom Settings](#custom-settings) below for details

<a id="custom-settings"></a>
## Custom Settings

The `custom_settings` section contains the configuration for the LDAP server

* `server` - (Required) The IP address or host name of the server providing the LDAP service
* `port` - (Required) Port number of the LDAP server (usually 389 for LDAP, 636 for LDAPS)
* `authentication_method` - (Required) Authentication method: one of `SIMPLE`, `MD5DIGEST`, `NTLM`
* `connector_type` - (Required) Type of connector: one of `OPEN_LDAP`, `ACTIVE_DIRECTORY`
* `base_distinguished_name` - (Required) LDAP search base
* `is_ssl` - (Optional) True if the LDAP service requires an SSL connection
* `username` - (Optional) _Username_ to use when logging in to LDAP, specified using LDAP attribute=value pairs 
  (for example: cn="ldap-admin", c="example", dc="com")
* `password` - (Optional) _Password_ for the user identified by UserName. This value is never returned by GET. 
   It is inspected on create and modify. On modify, the absence of this element indicates that the password should not be changed

* `user_attributes` - (Required) User settings when `ldap_mode` is `CUSTOM` See [User Attributes](#user-attributes) below for details
* `group_attributes` - (Required) Group settings when `ldap_mode` is `CUSTOM` See [Group Attributes](#group-attributes) below for details

<a id="user-attributes"></a>
### User Attributes

* `object_class` - (Required)  LDAP _objectClass_ of which imported users are members. For example, _user_ or _person_
* `unique_identifier` - (Required) LDAP attribute to use as the unique identifier for a user. For example, _objectGuid_
* `username` - (Required) LDAP attribute to use when looking up a username to import. For example, _userPrincipalName_ or _samAccountName_
* `email` - (Required) LDAP attribute to use for the user's email address. For example, _mail_
* `display_name` - (Required) LDAP attribute to use for the user's full name. For example, _displayName_
* `given_name` - (Required) LDAP attribute to use for the user's given name. For example, _givenName_
* `surname` - (Required) LDAP attribute to use for the user's surname. For example, _sn_
* `telephone` - (Required) LDAP attribute to use for the user's telephone number. For example, _telephoneNumber_
* `group_membership_identifier` - (Required) LDAP attribute that identifies a user as a member of a group. For example, _dn_
* `group_back_link_identifier` - (Optional) LDAP attribute that returns the identifiers of all the groups of which the user is a member

<a id="group-attributes"></a>
### Group Attributes

* `object_class` - (Required) LDAP _objectClass_ of which imported groups are members. For example, _group_
* `unique_identifier` - (Required) LDAP attribute to use as the unique identifier for a group. For example, _objectGuid_
* `name` - (Required) LDAP attribute to use for the group name. For example, _cn_
* `membership` - (Required) LDAP attribute to use when getting the members of a group. For example, _member_
* `group_membership_identifier` - (Required) LDAP attribute that identifies a group as a member of another group. For example, _dn_
* `group_back_link_identifier` - (Optional) LDAP group attribute used to identify a group member

## Importing

~> **Note:** The current implementation of Terraform import can only import resources into the state. It does not generate
configuration. [More information.][docs-import]

An existing LDAP configuration for an Org can be [imported][docs-import] into this resource via supplying the path for an Org. Since the Org is
at the top of the vCD hierarchy, the path corresponds to the Org name.
For example, using this structure, representing an existing LDAP configuration that was **not** created using Terraform:

```hcl
data "vcd_org" "my-org" {
  name = "my-org"
}

resource "vcd_org_ldap" "my-org-ldap" {
  org_id = data.vcd_org.my-org.id
}
```

You can import such LDAP configuration into terraform state using one of the following commands

```
# EITHER
terraform import vcd_org_ldap.my-org-ldap organization_name
# OR
terraform import vcd_org_ldap.my-org-ldap organization_id
```

After that, you must expand the configuration file before you can either update or delete the LDAP configuration. Running `terraform plan`
at this stage will show the difference between the minimal configuration file and the stored properties.
