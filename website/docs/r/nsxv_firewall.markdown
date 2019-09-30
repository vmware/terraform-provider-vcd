---
layout: "vcd"
page_title: "vCloudDirector: vcd_nsxv_firewall"
sidebar_current: "docs-vcd-resource-nsxv-firewall"
description: |-
  Provides a vCloud Director firewall resource for advanced edge gateways (NSX-V). This can be used
  to create, modify, and delete firewall rules.
---

# vcd\_nsxv\_firewall

Provides a vCloud Director firewall resource for advanced edge gateways (NSX-V). This can be used to
create, modify, and delete firewall rules. Replaces
[`vcd_firewall_rules`](/docs/providers/vcd/r/firewall_rules.html) resource.

~> **Note:** This resource requires advanced edge gateway. For non-advanced edge gateways please
use the [`vcd_firewall_rules`](/docs/providers/vcd/r/firewall_rules.html) resource.

## Example Usage 1 (Minimal input)

```hcl
resource "vcd_nsxv_firewall" "my-rule-1" {
  org          = "my-org"
  vdc          = "my-vdc"
  edge_gateway = "my-edge-gateway"

  source {
    ip_addresses = ["any"]
  }

  destination {
    ip_addresses = ["192.168.1.110"]
  }

  service {
    protocol = "any"
  }
}
```


## Example Usage 2 (Multiple services)

```hcl
resource "vcd_nsxv_firewall" "my-rule-1" {
  org          = "my-org"
  vdc          = "my-vdc"
  edge_gateway = "my-edge-gateway"

  source {
    ip_addresses       = ["any"]
    gateway_interfaces = ["internal"]
  }

  destination {
    ip_addresses = ["192.168.1.110"]
  }

  service {
    protocol = "icmp"
  }

  service {
    protocol = "tcp"
    port     = "443"
  }
}
```

## Example Usage 3 (Use exclusion in source)

```hcl
resource "vcd_nsxv_firewall" "my-rule-1" {
  org          = "my-org"
  vdc          = "my-vdc"
  edge_gateway = "my-edge-gateway"

  source {
    exclude            = true
    gateway_interfaces = ["internal"]
  }

  destination {
    ip_addresses = ["any"]
  }

  service {
    protocol = "icmp"
  }
}
```

## Example Usage 4 (Deny rule using exclusion and priority set)

```hcl
resource "vcd_nsxv_firewall" "my-rule-1" {
  org          = "my-org"
  vdc          = "my-vdc"
  edge_gateway = "my-edge-gateway"

  logging_enabled = "true"
  action          = "deny"

  source {
    ip_addresses = ["30.10.10.0/24", "31.10.10.0/24"]
    org_networks = ["org-net-1", "org-net-2"]
  }

  destination {
    ip_addresses = ["any"]
  }

  service {
    protocol = "icmp"
  }
}

resource "vcd_nsxv_firewall" "my-rule-2" {
  org          = "my-org"
  vdc          = "my-vdc"
  edge_gateway = "my-edge-gateway"

  # This attribute allows to ensure rule is inserted above the referred one
  # in rule processing engine
  above_rule_id = "${vcd_nsxv_firewall.my-rule-1.id}"
  name          = "my-friendly-name"

  source {
    ip_addresses = ["30.10.10.0/24", "31.10.10.0/24"]
    org_networks = ["org-net-1", "org-net-2"]
  }

  destination {
    ip_addresses = ["any"]
  }

  service {
    protocol = "icmp"
  }
}
```


## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to use, optional if defined at provider level. Useful
when connected as sysadmin working across different organisations.
* `vdc` - (Optional) The name of VDC to use, optional if defined at provider level.
* `edge_gateway` - (Required) The name of the edge gateway on which to apply the firewall rule.
* `action` - (Optional) Defines if the rule is set to `accept` or `deny` traffic. Default `accept`
* `enabled` - (Optional) Defines if the rule is enabaled. Default `true`.
* `logging_enabled` - (Optional) Defines if the logging for this rule is enabaled. Default `false`.
* `name` - (Optional) Free text name. Can be duplicate.
* `rule_tag` - (Optional) This can be used to specify user-controlled rule tag. If not specified,
it will report rule ID after creation. Must be between 65537-131072.
* `above_rule_id` - (Optional) This can be used to alter default rule placement order. By default
every rule is appended to the end of firewall rule list. When a value of another rule is set - this
rule will be placed above the specified rule.
* `source` - (Required) Exactly one block to define source criteria for firewall. See
[Endpoint](#endpoint) and example for usage details.
* `destination` - (Required) Exactly one block to define source criteria for firewall. See
[Endpoint](#endpoint) and example for usage details.
* `service` - (Required) One or more blocks to define protocol and port details. Use multiple blocks
if you want to define multiple port/protocol combinations for the same rule. See
[Service](#service) and example for usage details.


<a id="endpoint"></a>
## Endpoint (source or destination)

* `exclude` - (Optional) When the toggle exclusion is selected, the rule is applied to
traffic on all sources except for the locations you excluded. When the toggle exclusion is not
selected, the rule applies to traffic you specified. Default `false`
* `ip_addresses` - (Optional) A set of IP addresses, CIDRs or ranges. A keyword `any` is also
accepted as a parameter.
* `gateway_interfaces` - (Optional) A set of with either three keywords `vse` (UI names it as `any`), `internal`, `external` or an org network name. It automatically looks up vNic in the backend.
* `virtual_machine_ids` - (Optional) A set of `.id` fields of `vcd_vapp_vm` resources.
* `org_networks` - (Optional) A set of org network names.


<a id="service"></a>
## Service

* `protocol` - (Required) One of `any`, `tcp`, `udp`, `icmp` to apply.
* `port` - (Optional) Port number or range separated by `-` for port number. Default 'any'.
* `source_port` - (Optional) Port number or range separated by `-` for port number. Default 'any'.

## Attribute Reference

The following additional attributes are exported:

* `rule_type` - Possible values - `user`, `internal_high`.

## Importing

~> **Note:** The current implementation of Terraform import can only import resources into the state.
It does not generate configuration. [More information.](https://www.terraform.io/docs/import/)



An existing firewall rule can be [imported][docs-import] into this resource
via supplying the full dot separated path for firewall rule. An example is below:

[docs-import]: https://www.terraform.io/docs/import/

```
terraform import vcd_nsxv_firewall.imported my-org.my-org-vdc.my-edge-gw.my-firewall-rule-id
```

The above would import the application rule named `my-firewall-rule-id` that is defined on edge
gateway `my-edge-gw` which is configured in organization named `my-org` and vDC named `my-org-vdc`.


!> **Warning:** The UI does not show real firewall rule IDs, but their order from 1 to n.
Real firewall rules have IDs with larger integer numbers like `132730`. See below how to find real
IDs.

### Listing real firewall rule IDs

To list the real IDs there is a
special command **`terraform import vcd_nsxv_firewall.imported list@my-org.my-org-vdc.my-edge-gw`**
where `my-org` is the organization used, `my-org-vdc` is vDC name and `my-edge-gw` is edge gateway
name. The output for this command should look similar to below one:

```shell
$ terraform import vcd_nsxv_firewall.import list@my-org.my-org-vdc.my-edge-gw
vcd_nsxv_firewall.import: Importing from ID "list@my-org.my-org-vdc.my-edge-gw"...
Retrieving all firewall rules
UI ID   ID      Name                                    Action  Type
-----   --      ----                                    ------  ----
1       132589  firewall                                accept  internal_high
2       132730  My deny rule                            deny    user
3       132729  My accept rule                          accept  user
4       132588  default rule for ingress traffic        deny    default_policy

Error: Resource was not imported! Please use the above ID to format the command as:
terraform import vcd_nsxv_firewall.resource-name org.vdc.edge-gw.firewall-rule-id
```

Now to import rule with UI ID 2 (real ID 132730) one could supply this command:

```shell
$ terraform import vcd_nsxv_firewall.import my-org.my-org-vdc.my-edge-gw.132730
vcd_nsxv_firewall.import: Importing from ID "my-org.my-org-vdc.my-edge-gw.132730"...
vcd_nsxv_firewall.import: Import prepared!
  Prepared vcd_nsxv_firewall for import
vcd_nsxv_firewall.import: Refreshing state... [id=132730]

Import successful!

The resources that were imported are shown above. These resources are now in
your Terraform state and will henceforth be managed by Terraform.
```