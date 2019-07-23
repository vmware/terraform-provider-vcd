---
layout: "vcd"
page_title: "vCloudDirector: vcd_lb_virtual_server"
sidebar_current: "docs-vcd-resource-lb-virtual-server"
description: |-
  Provides an NSX edge gateway load balancer virtual server resource.
---

# vcd\_lb\_virtual\_server

Provides a vCloud Director edge gateway load balancer virtual server resource. Adds an edge gateway
internal or uplink interface as a virtual server. A virtual server has a public IP address and services all incoming client requests. 

~> **Note:** To make load balancing work one must ensure that load balancing is enabled on edge gateway (edge gateway must be advanced).
This depends on NSX version to work properly. Please refer to [VMware Product Interoperability Matrices](https://www.vmware.com/resources/compatibility/sim/interop_matrix.php#interop&29=&93=) 
to check supported vCloud director and NSX for vSphere configurations.

~> **Note:** The vCloud Director API for NSX supports a subset of the operations and objects defined in the NSX vSphere 
API Guide. The API supports NSX 6.2, 6.3, and 6.4.

Supported in provider *v2.4+*

## Example Usage (HTTP virtual server)

```hcl
provider "vcd" {
  user     = "${var.admin_user}"
  password = "${var.admin_password}"
  org      = "System"
  url      = "https://AcmeVcd/api"
}

resource "vcd_lb_virtual_server" "http" {
  org          = "my-org"
  vdc          = "my-org-vdc"
  edge_gateway = "my-edge-gw"
  
  name       = "http-virtual-server"
  ip_address = "1.1.1.1" # Edge gateway uplink interface IP
  protocol   = "http"    # Must be the same as specified in application profile
  port       = 80
  
  app_profile_id = "${vcd_lb_app_profile.http.id}"
  server_pool_id = "${vcd_lb_server_pool.web-servers.id}"
  app_rule_ids   = ["${vcd_lb_app_rule.redirect.id}", "${vcd_lb_app_rule.language.id}"]
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to use, optional if defined at provider level. Useful
when connected as sysadmin working across different organisations
* `vdc` - (Optional) The name of VDC to use, optional if defined at provider level
* `edge_gateway` - (Required) The name of the edge gateway on which the virtual server is to be
created
* `name` - (Required) Virtual server name
* `description` - (Optional) Virtual server description
* `enabled` - (Optional) Defines if the virtual server is enabled. Default `true`
* `enable_acceleration` - (Optional) Defines if the virtual server uses acceleration. Default
`false`
* `ip_address` - (Required) Set the IP address that the load balancer listens on
* `protocol` - (Required) Select the protocol that the virtual server accepts. One of `tcp`, `udp`,
`http`, or `https` **Note**: You must select the same protocol used by the selected
**Application Profile**
* `port` - (Required) The port number that the load balancer listens on
* `connection_limit` - (Optional) Maximum concurrent connections that the virtual server can process
* `connection_rate_limit` - (Optional) Maximum incoming new connection requests per second
* `server_pool_id` - (Optional) The server pool that the load balancer will use
* `app_profile_id` - (Optional) Application profile ID to be associated with the virtual server
* `app_rule_ids` - (Optional) List of attached application rule IDs

## Attribute Reference

The following attributes are exported on the base level of this resource:

* `id` - The NSX ID of the load balancer virtual server

## Importing

~> **Note:** The current implementation of Terraform import can only import resources into the state. It does not generate
configuration. [More information.](https://www.terraform.io/docs/import/)

An existing load balancer virtual server can be [imported][docs-import] into this resource
via supplying the full dot separated path for load balancer virtual server. An example is below:

[docs-import]: https://www.terraform.io/docs/import/

```
terraform import vcd_lb_virtual_server.imported my-org.my-org-vdc.my-edge-gw.my-lb-virtual-server
```

The above would import the virtual server named `my-lb-virtual-server` that is defined on edge gateway
`my-edge-gw` which is configured in organization named `my-org` and vDC named `my-org-vdc`.
