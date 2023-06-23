---
layout: "vcd"
page_title: "VMware Cloud Director: IP Spaces"
sidebar_current: "docs-vcd-guides-ip-spaces"
description: |-
  Provides guidance to IP Spaces in VCD 10.4.1+
---

# IP Spaces

IP Spaces provide structured approach to allocating public and private IP addresses by preventing
the use of overlapping IP addresses across organizations and organization VDCs.

Starting with **VMware Cloud Director 10.4.1**, you can use IP Spaces to manage your IP address
allocation needs. IP Spaces provide structured approach to allocating *public* and *private* IP
addresses by preventing the use of overlapping IP addresses across organizations and organization
VDCs.

An IP space consists of a set of defined non-overlapping IP ranges and small CIDR blocks that are
reserved and used during the consumption aspect of the IP space life cycle. An IP space can be
either IPv4 or IPv6, but not both.

Every IP space has an internal scope and an external scope. The internal scope of an IP space is a
list of CIDR notations that defines the exact span of IP addresses in which all ranges and blocks
must be contained in. The external scope defines the total span of IP addresses to which the IP
space has access, for example the internet or a WAN. The internal and external scopes are used to
define default NAT rules and BGP prefixes.

As a service provider, you create public, shared, or private IP spaces and assign them to provider
gateways by creating IP space uplinks. After creating an IP space, you can assign to it IP prefixes
for networks and floating IP addresses for network services.

Organization administrators can view general information about the IP spaces in their organization
has access, and manage the IP spaces available to them.

There are three types of IP spaces that you can create.

* Public IP Space - A public IP Space is used by multiple organizations and is controlled by the
  service provider through a quota-based system. 
* Shared IP Space - An IP Space for services and management networks that are required in the tenant
  space, but as a service provider, you don't want to expose it to organizations in your
  environment. 
* Private IP Space - Private IP Spaces are dedicated to a single tenant - a private IP space is used
  by only one organization that is specified during the space creation. For this organization, IP
  consumption is unlimited.


## New resources

* `vcd_ip_space` - provides IP Space definition capability
* `vcd_ip_space_uplink` - provides capability to assign IP Space Uplink for Provider Gateways
* `vcd_ip_space_ip_allocation` - provides capability to allocation floating IPs or IP Prefix
* `vcd_ip_space_custom_quota` - provides capability to set Org specified Custom Quotas to override
  default ones defined in `vcd_ip_space`

## Modified resources

* `vcd_external_network_v2` - new fields `use_ip_spaces` and `dedicated_org_id` (applicable only to
  T0 or T0 VRF backed networks also known as Provider Gateways in UI)
* `vcd_nsxt_edgegateway` - none of the fields `subnet_with_total_ip_count`, `subnet`,
  `subnet_with_ip_count` are no longer mandatory when specifying `external_network_id` that is using
  IP Spaces. As a result they will not be populated after read operations together with
  `used_ip_count` and `unused_ip_count`. Additional computed flag `uses_ip_spaces` to tell if the
  Edge Gateway is using IP Spaces (is backed by Provider Gateway that has IP Space Uplinks)

## Sample configuration using IP Spaces

```hcl
data "vcd_nsxt_manager" "main" {
  name = "nsxManager1"
}

data "vcd_nsxt_tier0_router" "router" {
  name            = "tier0Router-cloudOrg"
  nsxt_manager_id = data.vcd_nsxt_manager.main.id
}

data "vcd_org" "org1" {
  name = "cloudOrg"
}

data "vcd_org_vdc" "vdc1" {
  org  = "cloudOrg"
  name = "nsxt-vdc-cloudOrg"
}

resource "vcd_ip_space" "space1" {
  name = "IpSpaceIntegration"
  type = "PUBLIC"

  internal_scope = ["192.168.1.0/24", "10.10.10.0/24", "11.11.11.0/24"]
  external_scope = "0.0.0.0/24"

  route_advertisement_enabled = false

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

  ip_prefix {
    default_quota = -1

    prefix {
      first_ip      = "10.10.10.96"
      prefix_length = 29
      prefix_count  = 4
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

resource "vcd_external_network_v2" "provider-gateway" {
  name = "IpSpaceIntegration"

  nsxt_network {
    nsxt_manager_id      = data.vcd_nsxt_manager.main.id
    nsxt_tier0_router_id = data.vcd_nsxt_tier0_router.router.id
  }

  # A flag to use IP Spaces instead of specifying IP allocations at Provider Gateway configuration
  use_ip_spaces = true
}

# The IP Space uplink for external network is the way to assign IP Space to External Network
resource "vcd_ip_space_uplink" "u1" {
  name                = "IpSpaceIntegration"
  external_network_id = vcd_external_network_v2.provider-gateway.id
  ip_space_id         = vcd_ip_space.space1.id
}

# Using an External Network that has an IP Space uplink does not need subnet allocation
resource "vcd_nsxt_edgegateway" "ip-space" {
  org                 = "cloudOrg"
  name                = "IpSpaceIntegration"
  owner_id            = data.vcd_org_vdc.vdc1.id
  external_network_id = vcd_external_network_v2.provider-gateway.id

  depends_on = [vcd_ip_space_uplink.u1]
}

# Explicit resource for Floating IP which is then used in vcd_nsxt_nat_rule (`ip_address` attribute)
resource "vcd_ip_space_ip_allocation" "public-floating-ip" {
  org_id      = data.vcd_org.org1.id
  ip_space_id = vcd_ip_space.space1.id
  type        = "FLOATING_IP"

  depends_on = [vcd_nsxt_edgegateway.ip-space]
}

resource "vcd_ip_space_ip_allocation" "public-floating-ip2" {
  org_id      = data.vcd_org.org1.id
  ip_space_id = vcd_ip_space.space1.id
  type        = "FLOATING_IP"

  depends_on = [vcd_nsxt_edgegateway.ip-space]
}

resource "vcd_nsxt_nat_rule" "dnat-floating-ip" {
  org             = "cloudOrg"
  edge_gateway_id = vcd_nsxt_edgegateway.ip-space.id

  name      = "IpSpaceIntegration"
  rule_type = "DNAT"

  # Using Floating IP From IP Space
  external_address = vcd_ip_space_ip_allocation.public-floating-ip2.ip_address
  internal_address = "77.77.77.1"
  logging          = true
}

# An example IP allocation for manual usage
resource "vcd_ip_space_ip_allocation" "public-floating-ip-manual" {
  org_id      = data.vcd_org.org1.id
  ip_space_id = vcd_ip_space.space1.id
  type        = "FLOATING_IP"
  usage_state = "USED_MANUAL"
  description = "manually used floating IP"

  depends_on = [vcd_nsxt_edgegateway.ip-space]
}

# IP Prefix allocation for routed network within Org - `ip_address` computed attribute is returned 
# in CIDR format (e.g. 192.168.1.0/24). Using built-in Terraform functions `cidrhost` and `split` 
# can be used to split and pick IP Addresses as well as prefix length
resource "vcd_ip_space_ip_allocation" "public-ip-prefix" {
  org_id        = data.vcd_org.org1.id
  ip_space_id   = vcd_ip_space.space1.id
  type          = "IP_PREFIX"
  prefix_length = 29

  depends_on = [vcd_nsxt_edgegateway.ip-space]
}


# ip_address = split("/","10.10.10.1/24")

resource "vcd_network_routed_v2" "using-public-prefix" {
  org             = "cloudOrg"
  name            = "IpSpaceIntegration"
  edge_gateway_id = vcd_nsxt_edgegateway.ip-space.id
  gateway         = cidrhost(vcd_ip_space_ip_allocation.public-ip-prefix.ip_address, 1)
  prefix_length   = vcd_ip_space_ip_allocation.public-ip-prefix.prefix_length

  static_ip_pool {
    start_address = cidrhost(vcd_ip_space_ip_allocation.public-ip-prefix.ip_address, 2)
    end_address   = cidrhost(vcd_ip_space_ip_allocation.public-ip-prefix.ip_address, 4)
  }
}
```

## Sample configuration without IP Spaces (the old way)

```hcl
resource "vcd_external_network_v2" "provider-gateway" {
  name = "IpSpaceIntegration"

  nsxt_network {
    nsxt_manager_id      = data.vcd_nsxt_manager.main.id
    nsxt_tier0_router_id = data.vcd_nsxt_tier0_router.router.id
  }

  ip_scope {
    # enabled       = true # by default
    gateway       = "11.11.11.1"
    prefix_length = "24"

    static_ip_pool {
      start_address = "11.11.11.100"
      end_address   = "11.11.11.110"
    }
  }
}

resource "vcd_nsxt_edgegateway" "ip-space" {
  org                 = "cloudOrg"
  name                = "IpSpaceIntegration"
  owner_id            = data.vcd_org_vdc.vdc1.id
  external_network_id = vcd_external_network_v2.provider-gateway.id

  subnet {
    gateway       = "11.11.11.1"
    prefix_length = "24"
    primary_ip    = "11.11.11.100"

    # IP Allocation occurs in Edge Gateway
    allocated_ips {
      start_address = "11.11.11.100"
      end_address   = "11.11.11.108"
    }
  }
}

resource "vcd_nsxt_nat_rule" "dnat-floating-ip" {
  org             = "cloudOrg"
  edge_gateway_id = vcd_nsxt_edgegateway.ip-space.id

  name      = "IpSpaceIntegration"
  rule_type = "DNAT"

  # Using Floating IP From IP Space
  external_address = tolist(vcd_nsxt_edgegateway.ip-space.subnet)[0].end_address
  internal_address = "77.77.77.1"
  logging          = true
}
```

## References

* [VMware Cloud Director Documentation for Providers](https://docs.vmware.com/en/VMware-Cloud-Director/10.4/VMware-Cloud-Director-Service-Provider-Admin-Portal-Guide/GUID-46772618-7991-4928-A77B-BC774C45EA33.html)
* [VMware Cloud Director Documentation for Tenants](https://docs.vmware.com/en/VMware-Cloud-Director/10.4/VMware-Cloud-Director-Tenant-Portal-Guide/GUID-FB230D89-ACBC-4345-A11A-D099D359ED1B.html)
* [IP Space Uplinks for Provider Gateways](https://docs.vmware.com/en/VMware-Cloud-Director/10.4/VMware-Cloud-Director-Service-Provider-Admin-Portal-Guide/GUID-0D40BD21-CAAA-4FD3-B6ED-78BA8FE2DEF1.html)
