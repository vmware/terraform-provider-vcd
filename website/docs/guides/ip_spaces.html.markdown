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

* Public IP Space - A public IP space is used by multiple organizations and is controlled by the
  service provider through a quota-based system. 
* Shared IP Space - An IP space for services and management networks that are required in the tenant
  space, but as a service provider, you don't want to expose it to organizations in your
  environment. 
* Private IP Space - Private IP spaces are dedicated to a single tenant - a private IP space is used
  by only one organization that is specified during the space creation. For this organization, IP
  consumption is unlimited.



## Impacted resources

* `vcd_external_network_v2` - new fields `use_ip_spaces` and `dedicated_org_id`

## References

* [VMware Cloud Director Documentation for Providers](https://docs.vmware.com/en/VMware-Cloud-Director/10.4/VMware-Cloud-Director-Service-Provider-Admin-Portal-Guide/GUID-46772618-7991-4928-A77B-BC774C45EA33.html)
* [VMware Cloud Director Documentation for Tenants](https://docs.vmware.com/en/VMware-Cloud-Director/10.4/VMware-Cloud-Director-Tenant-Portal-Guide/GUID-FB230D89-ACBC-4345-A11A-D099D359ED1B.html)

