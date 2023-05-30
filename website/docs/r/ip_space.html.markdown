---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_ip_space"
sidebar_current: "docs-vcd-resource-ip-space"
description: |-
  Provides a resource to manage IP Spaces for IP address management needs. IP Spaces provide 
  structured approach to allocating public and private IP addresses by preventing the use of 
  overlapping IP addresses across organizations and organization VDCs.
---

# vcd\_ip\_space

Provides a resource to manage IP Spaces for IP address management needs. IP Spaces provide
structured approach to allocating public and private IP addresses by preventing the use of
overlapping IP addresses across organizations and organization VDCs.

IP Spaces require VCD 10.4.1+ with NSX-T.

## Example Usage (Private)

```hcl
resource "vcd_ip_space" "space1" {
  name        = "org-owned-ip-space"
  description = "description of IP Space"
  type        = "PRIVATE"
  org_id      = data.vcd_org.org1.id

  internal_scope = ["192.168.1.0/24","10.10.10.0/24", "11.11.11.0/24"]

  route_advertisement_enabled = false

  ip_prefix {
	  default_quota = -1 # unlimited

	  prefix {
	  	first_ip = "192.168.1.100"
	  	prefix_length = 30
	  	prefix_count = 4
	  }
  
	  prefix {
	  	first_ip = "192.168.1.200"
	  	prefix_length = 30
	  	prefix_count = 4
	  }
  }

  ip_prefix {
	  default_quota = -1 # unlimited

	  prefix {
	  	first_ip = "10.10.10.96"
	  	prefix_length = 29
	  	prefix_count = 4
	  }
  }

  ip_range {
	  start_address = "11.11.11.100"
	  end_address   = "11.11.11.110"
  }

  ip_range {
	  start_address = "11.11.11.120"
	  end_address   = "11.11.11.123"
  }
}
```

## Example Usage (Public)

```hcl
resource "vcd_ip_space" "space1" {
  name        = "Public-Tokyo"
  type        = "PUBLIC"

  internal_scope = ["192.168.1.0/24","10.10.10.0/24", "11.11.11.0/24"]

  route_advertisement_enabled = true

  ip_prefix {
	  default_quota = 2

	  prefix {
	  	first_ip = "192.168.1.100"
	  	prefix_length = 30
	  	prefix_count = 4
	  }
  }

  ip_prefix {
	  default_quota = -1

	  prefix {
	  	first_ip = "10.10.10.96"
	  	prefix_length = 29
	  	prefix_count = 4
	  }
  }

  ip_range {
	  start_address = "11.11.11.100"
	  end_address   = "11.11.11.110"
  }
}
```

## Example Usage (Shared)

```hcl
resource "vcd_ip_space" "space1" {
  name        = "Backup-network"
  description = "Network used for backups"
  type        = "SHARED_SERVICES"

  internal_scope = ["192.168.1.0/24","10.10.10.0/24", "11.11.11.0/24"]

  route_advertisement_enabled = false

  ip_prefix {
	  default_quota = 0 # no quota

	  prefix {
	  	first_ip = "192.168.1.100"
	  	prefix_length = 30
	  	prefix_count = 4
	  }
  }

  ip_prefix {
	  default_quota = 0 # no quota

	  prefix {
	  	first_ip = "10.10.10.96"
	  	prefix_length = 29
	  	prefix_count = 4
	  }
  }
}

```

## Argument Reference

The following arguments are supported:

* `org_id` - (Optional) Required for `PRIVATE` type
* `name` - (Required) A name for IP Space
* `description` - (Optional) - Description of IP Space
* `type` - (Required) One of `PUBLIC`, `SHARED_SERVICES`, `PRIVATE`
  * `PUBLIC` - A public IP space is *used by multiple organizations* and is *controlled by the service
    provider* through a quota-based system. 
  * `SHARED_SERVICES` - An IP space for services and management networks that are required in the
    tenant space, but as a service provider, you don't want to expose it to organizations in your
    environment. The main difference from `PUBLIC` network is that IPs cannot be allocated by tenants.
  * `PRIVATE` - Private IP spaces are dedicated to a single tenant - a private IP space is used by
    only one organization that is specified during the space creation. For this organization, IP
    consumption is unlimited.

* `internal_scope` - (Required) The internal scope of an IP space is a list of CIDR notations that
  defines the exact span of IP addresses in which all ranges and blocks must be contained in.
* `external_scope` - (Optional) The external scope defines the total span of IP addresses to which the IP
  space has access, for example the internet or a WAN. 
* `ip_range` - (Optional) One or more [ip_range](#ipspace-ip-range) for floating IP address
  allocation. (Floating IP addresses are just IP addresses taken from the defined range) 
* `ip_range_quota` - (Optional) If you entered at least one IP Range (`ip_range`), enter a
  number of floating IP addresses to allocate individually. `-1` is unlimited, while `0` means that
  no IPs can be allocated.
* `ip_prefix` - (Optional) One or more IP prefixes (blocks) [ip_prefix](#ipspace-ip-prefix)

* `route_advertisement_enabled` - (Optional) Toggle on the route advertisement option to
  enable advertising networks with IP prefixes from this IP space (default `false`)

<a id="ipspace-ip-range"></a>

## ip_range block

* `start_address` - (Required) - Start IP address of a range
* `end_address` - (Required) - End IP address of a range

```hcl
ip_range {
 start_address = "11.11.11.120"
 end_address   = "11.11.11.123"
}
```

<a id="ipspace-ip-prefix"></a>

## ip_prefix block

* `default_quota` 
* `prefix` - IP block definition as detail [below](#ipspace-ip-prefix-prefix)

<a id="ipspace-ip-prefix-prefix"></a>

## prefix block

Defines blocks of IPs. Blocks must fall into subnets defined in `internal_scope` and not clash with
IP ranges defined in `ip_range` 

* `first_ip` - (Required) - First IP of the prefix
* `prefix_length` - (Required) - Prefix length
* `prefix_count` - (Required) - Number of prefixes 

```hcl
 ip_prefix {
  default_quota = 2

  prefix {
    first_ip      = "192.168.1.100"
    prefix_length = 30
    prefix_count  = 4
  }

  prefix {
    first_ip      = "192.168.1.200"
    prefix_length = 30
    prefix_count  = 4
  }
}
```


## Importing

~> The current implementation of Terraform import can only import resources into the state.
It does not generate configuration. [More information.](https://www.terraform.io/docs/import/)

An IP Space can be [imported][docs-import] into this resource via supplying path for it. An example
is below:

[docs-import]: https://www.terraform.io/docs/import/

```
terraform import vcd_ip_space.imported ip-space-name
```

The above would import the `ip-space-name` IP Space defined at provider
level.


or 

```
terraform import vcd_ip_space.imported org-name.ip-space-name
```

The above would import the `ip-space-name` IP Space defined for Org `org-name`.
