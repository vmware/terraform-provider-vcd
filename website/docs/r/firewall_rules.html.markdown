---
layout: "vcd"
page_title: "vCloudDirector: vcd_firewall_rules"
sidebar_current: "docs-vcd-resource-firewall-rules"
description: |-
  Provides a vCloud Director Firewall resource. This can be used to create, modify, and delete firewall settings and rules.
---

# vcd\_firewall\_rules

Provides a vCloud Director Firewall resource. This can be used to create,
modify, and delete firewall settings and rules.

~> **Note:** DEPRECATED: Please use the improved [`vcd_nsxv_firewall_rule`](/docs/providers/vcd/r/nsxv_firewall_rule.html)
resource with advanced edge gateways (NSX-V).

~> **Note:** Using this resource automatically enables default firewall rule logging. This may cause
[`vcd_edgegateway`](/docs/providers/vcd/r/edgegateway.html) resource to report changes for field
 `fw_default_rule_logging_enabled` during `plan`/`apply` phases.

## Example Usage

```hcl
resource "vcd_firewall_rules" "fw" {
  edge_gateway   = "Edge Gateway Name"
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

resource "vcd_vapp" "web" {
  # ...
}

resource "vcd_firewall_rules" "fw-web" {
  edge_gateway   = "Edge Gateway Name"
  default_action = "drop"

  rule {
    description      = "allow-web"
    policy           = "allow"
    protocol         = "tcp"
    destination_port = "80"
    destination_ip   = "${vcd_vapp.web.ip}"
    source_port      = "any"
    source_ip        = "any"
  }
}
```

## Argument Reference

The following arguments are supported:

* `edge_gateway` - (Required) The name of the edge gateway on which to apply the Firewall Rules
* `default_action` - (Required) Either "allow" or "drop". Specifies what to do should none of the rules match
* `rule` - (Optional) Configures a firewall rule; see [Rules](#rules) below for details.
* `org` - (Optional; *v2.0+*) The name of organization to use, optional if defined at provider level. Useful when connected as sysadmin working across different organisations
* `vdc` - (Optional; *v2.0+*) The name of VDC to use, optional if defined at provider level

<a id="rules"></a>
## Rules

Each firewall rule supports the following attributes:

* `description` - (Required) Description of the fireall rule
* `policy` - (Required) Specifies what to do when this rule is matched. Either "allow" or "drop"
* `protocol` - (Required) The protocol to match. One of "tcp", "udp", "icmp" or "any"
* `destination_port` - (Required) The destination port to match. Either a port number or "any"
* `destination_ip` - (Required) The destination IP to match. Either an IP address, IP range or "any"
* `source_port` - (Required) The source port to match. Either a port number or "any"
* `source_ip` - (Required) The source IP to match. Either an IP address, IP range or "any"
