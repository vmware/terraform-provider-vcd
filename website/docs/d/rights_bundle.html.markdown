---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_rights_bundle"
sidebar_current: "docs-vcd-data-source-rights-bundle"
description: |-
 Provides a VMware Cloud Director rights bundle data source. This can be used to read rights bundles.
---

# vcd\_rights\_bundle

Provides a VMware Cloud Director rights bundle data source. This can be used to read rights bundles.

Supported in provider *v3.3+*

## Example Usage

```hcl
data "vcd_rights_bundle" "default-set" {
  name = "Default Rights Bundle"
}
```

```
Sample output:

default-set = {
  "bundle_key" = "com.vmware.vcloud.undefined.key"
  "description" = "Default set of tenant rights"
  "id" = "urn:vcloud:rightsBundle:31482daa-c398-4938-895e-5dc6911a0641"
  "name" = "Default Rights Bundle"
  "publish_to_all_tenants" = false
  "read_only" = false
  "rights" = toset([
    "Access All Organization VDCs",
    "Catalog: Add vApp from My Cloud",
    "Catalog: CLSP Publish Subscribe",
    "Catalog: Change Owner",
    "Catalog: Create / Delete a Catalog",
    "Catalog: Edit Properties",
    "Catalog: Publish",
    "Catalog: Sharing",
    "Catalog: View ACL",
    "Catalog: View Private and Shared Catalogs",
    "Catalog: View Published Catalogs",
    "Certificate Library: View",
    "Custom entity: View all custom entity instances in org",
    "Custom entity: View custom entity instance",
    "General: Administrator Control",
    "General: Administrator View",
    "General: Send Notification",
    "Group / User: View",
    "Hybrid Cloud Operations: Acquire control ticket",
    "Hybrid Cloud Operations: Acquire from-the-cloud tunnel ticket",
    "Hybrid Cloud Operations: Acquire to-the-cloud tunnel ticket",
    "Hybrid Cloud Operations: Create from-the-cloud tunnel",
    "Hybrid Cloud Operations: Create to-the-cloud tunnel",
    "Hybrid Cloud Operations: Delete from-the-cloud tunnel",
    "Hybrid Cloud Operations: Delete to-the-cloud tunnel",
    "Hybrid Cloud Operations: Update from-the-cloud tunnel endpoint tag",
    "Hybrid Cloud Operations: View from-the-cloud tunnel",
    "Hybrid Cloud Operations: View to-the-cloud tunnel",
    "Organization Network: Edit Properties",
    "Organization Network: View",
    "Organization VDC: view metrics",
    "Organization vDC Compute Policy: View",
    "Organization vDC Distributed Firewall: Configure Rules",
    "Organization vDC Distributed Firewall: View Rules",
    "Organization vDC Gateway: Configure DHCP",
    "Organization vDC Gateway: Configure DNS",
    "Organization vDC Gateway: Configure Firewall",
    "Organization vDC Gateway: Configure IPSec VPN",
    "Organization vDC Gateway: Configure Load Balancer",
    "Organization vDC Gateway: Configure NAT",
    "Organization vDC Gateway: Configure Static Routing",
    "Organization vDC Gateway: Configure Syslog",
    "Organization vDC Gateway: Convert to Advanced Networking",
    "Organization vDC Gateway: View",
    "Organization vDC Gateway: View DHCP",
    "Organization vDC Gateway: View DNS",
    "Organization vDC Gateway: View Firewall",
    "Organization vDC Gateway: View IPSec VPN",
    "Organization vDC Gateway: View Load Balancer",
    "Organization vDC Gateway: View NAT",
    "Organization vDC Gateway: View Static Routing",
    "Organization vDC Named Disk: Change Owner",
    "Organization vDC Named Disk: Create",
    "Organization vDC Named Disk: Delete",
    "Organization vDC Named Disk: Edit Properties",
    "Organization vDC Named Disk: View Properties",
    "Organization vDC Network: Edit Properties",
    "Organization vDC Network: View Properties",
    "Organization vDC Storage Profile: Set Default",
    "Organization vDC: Edit",
    "Organization vDC: Edit ACL",
    "Organization vDC: Manage Firewall",
    "Organization vDC: VM-VM Affinity Edit",
    "Organization vDC: View",
    "Organization vDC: View ACL",
    "Organization: Edit Association Settings",
    "Organization: Edit Federation Settings",
    "Organization: Edit Leases Policy",
    "Organization: Edit OAuth Settings",
    "Organization: Edit Password Policy",
    "Organization: Edit Properties",
    "Organization: Edit Quotas Policy",
    "Organization: Edit SMTP Settings",
    "Organization: Import User/Group from IdP while Editing VDC ACL",
    "Organization: View",
    "Organization: view metrics",
    "Quota Policy Capabilities: View",
    "Role: Create, Edit, Delete, or Copy",
    "SSL: Test Connection",
    "Service Library: View service libraries",
    "Truststore: Manage",
    "Truststore: View",
    "UI Plugins: View",
    "VAPP_VM_METADATA_TO_VCENTER",
    "VDC Template: Instantiate",
    "VDC Template: View",
    "vApp Template / Media: Copy",
    "vApp Template / Media: Create / Upload",
    "vApp Template / Media: Edit",
    "vApp Template / Media: View",
    "vApp Template: Change Owner",
    "vApp Template: Checkout",
    "vApp Template: Download",
    "vApp: Change Owner",
    "vApp: Copy",
    "vApp: Create / Reconfigure",
    "vApp: Delete",
    "vApp: Download",
    "vApp: Edit Properties",
    "vApp: Edit VM CPU",
    "vApp: Edit VM Hard Disk",
    "vApp: Edit VM Memory",
    "vApp: Edit VM Network",
    "vApp: Edit VM Properties",
    "vApp: Manage VM Password Settings",
    "vApp: Power Operations",
    "vApp: Sharing",
    "vApp: Snapshot Operations",
    "vApp: Upload",
    "vApp: Use Console",
    "vApp: VM Boot Options",
    "vApp: View ACL",
    "vApp: View VM metrics",
  ])
  "tenants" = toset([
    "org1",
    "org2"
  ])
}
```


## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the rights bundle.

## Attribute Reference

* `description` - A description of the rights bundle
* `bundle_key` - Key used for internationalization.
* `rights` - List of rights assigned to this role
* `publish_to_all_tenants` - When true, publishes the rights bundle to all tenants
* `tenants` - List of tenants to which this rights bundle gets published. Ignored if `publish_to_all_tenants` is true.
* `read_only` - Whether this rights bundle is read-only

## More information

See [Roles management](/docs/providers/vcd/guides/roles_management.html) for a broader description of how rights bundles and
rights work together.
