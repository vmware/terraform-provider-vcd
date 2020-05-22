---
layout: "vcd"
page_title: "vCloudDirector: vcd_vapp_firewall_rules"
sidebar_current: "docs-vcd-resource-vapp-firewall-rules"
description: |-
  Provides a vCloud Director vApp Firewall resource. This can be used to create, modify, and delete firewall settings and rules.
---

# vcd\_vapp\_firewall\_rules

Provides a vCloud Director vApp Firewall resource. This can be used to create,
modify, and delete firewall settings and rules in a [vApp network](/docs/providers/vcd/r/vapp_network.html).

!> **Warning:** Using this resource overrides any existing firewall rules on vApp network.

## Example Usage

```hcl
resource "vcd_vapp" "web" {
  name = "web"
}
​
resource "vcd_vapp_network" "vapp-net" {
  name               = "my-vapp-net"
  vapp_name          = vcd_vapp.web.name
  org_network_name   = "my-vdc-int-net"
  gateway            = "192.168.2.1"
  netmask            = "255.255.255.0"
  dns1               = "192.168.2.1"
​
  static_ip_pool {
    start_address = "192.168.2.51"
    end_address   = "192.168.2.100"
  }
}

resource "vcd_vapp_firewall_rules" "vapp_fw" {
  vapp_id        = vcd_vapp.TestAccVcdVAppForInsert.id
  network_id     = vcd_vapp_network.vapp-net.id
  default_action = "drop"

  rule {
    description      = "drop-ftp-out"
    policy           = "drop"
    protocol         = "tcp"
    destination_port = "21"
    destination_ip   = "any"
    source_port      = "any"
    source_ip        = "10.10.0.0/24"
  }

  rule {
    description      = "allow-outbound"
    policy           = "allow"
    protocol         = "any"
    destination_port = "any"
    destination_ip   = "any"
    source_port      = "any"
    source_ip        = "10.10.0.0/24"
  }
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional; *v2.0+*) The name of organization to use, optional if defined at provider level. Useful when connected as sysadmin working across different organisations.
* `vdc` - (Optional; *v2.0+*) The name of VDC to use, optional if defined at provider level.
* `vapp_id` - (Required) The identifier of [vApp](/docs/providers/vcd/r/vapp.html).
* `network_id` - (Required) The identifier of [vApp network](/docs/providers/vcd/r/vapp_network.html).
* `default_action` - (Required) Either 'allow' or 'drop'. Specifies what to do should none of the rules match.
* `log_default_action` - (Optional) Flag to enable logging for default action. Default value is `false`.
* `rule` - (Optional) Configures a firewall rule; see [Rules](#rules) below for details.

<a id="rules"></a>
## Rules

Each firewall rule supports the following attributes:

* `description` - (Optional) Name of the firewall rule.
* `enabled` - (Optional) `true` value will enable firewall rule.
* `policy` - (Optional) Specifies what to do when this rule is matched. Either `allow` or `drop`.
* `protocol` - (Optional) The protocol to match. One of `tcp`, `udp`, `icmp`, `any` or `tcp&udp`.
* `destination_port` - (Optional) The destination port to match. Either a port number or `any`.
* `destination_ip` - (Optional) The destination IP to match. Either an IP address, IP range or `any`.
* `destination_vm_id` - (Optional) Destination VM identifier.
* `destination_vm_ip_type` - (Optional) The value can be one of: `assigned` - use assigned internal IP, `NAT` - use NATed external IP.
* `destination_vm_nic_id` - (Optional) VM NIC ID to which this rule applies.
* `source_port` - (Optional) The source port to match. Either a port number or `any`.
* `source_ip` - (Optional) The source IP to match. Either an IP address, IP range or `any`.
* `source_vm_id` - (Optional) Source VM identifier.
* `source_vm_ip_type` - (Optional) The value can be one of: `assigned` - use assigned internal IP, `NAT` - use NATed external IP.
* `source_vm_nic_id` - (Optional) VM NIC ID to which this rule applies.
* `enable_logging`- (Optional) `true` value will enable rule logging. Default is `false`.

## Importing

~> **Note:** The current implementation of Terraform import can only import resources into the state.
It does not generate configuration. [More information.](https://www.terraform.io/docs/import/)

An existing an vApp network firewall rules can be [imported][docs-import] into this resource
via supplying the full dot separated path to vapp network. An example is
below:

```
terraform import vcd_vapp_firewall_rules.my-rules my-org.my-vdc.vapp_name.network_name
```
or using IDs:
```
terraform import vcd_vapp_firewall_rules.my-rules my-org.my-vdc.vapp_id.network_id
```

NOTE: the default separator (.) can be changed using Provider.import_separator or variable VCD_IMPORT_SEPARATOR

[docs-import]:https://www.terraform.io/docs/import/

After that, you can expand the configuration file and either update or delete the vApp network rules as needed. Running `terraform plan`
at this stage will show the difference between the minimal configuration file and the vApp network rules stored properties.
