---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_lb_app_profile"
sidebar_current: "docs-vcd-resource-lb-app-profile"
description: |-
  Provides an NSX edge gateway load balancer application profile resource.
---

# vcd\_lb\_app\_profile

Provides a VMware Cloud Director Edge Gateway Load Balancer Application Profile resource. An application
profile defines the behavior of the load balancer for a particular type of network traffic. After
configuring a profile, you associate it with a virtual server. The virtual server then processes
traffic according to the values specified in the profile.

~> **Note:** This resource does not currently support attaching  Pool and Virtual
Server certificates. The `enable_pool_side_ssl` only toggles the option, but does not setup
certificates.

~> **Note:** To make load balancing work one must ensure that load balancing is enabled on edge
gateway (edge gateway must be advanced).
This depends on NSX version to work properly. Please refer to [VMware Product Interoperability
Matrices](https://www.vmware.com/resources/compatibility/sim/interop_matrix.php#interop&29=&93=) 
to check supported VMware Cloud Director and NSX for vSphere configurations.

~> **Note:** The VMware Cloud Director API for NSX supports a subset of the operations and objects defined
in the NSX vSphere API Guide. The API supports NSX 6.2, 6.3, and 6.4.

Supported in provider *v2.4+*

## Example Usage 1 (TCP Application Profile)

```hcl
resource "vcd_lb_app_profile" "tcp" {
  org          = "my-org"
  vdc          = "my-org-vdc"
  edge_gateway = "my-edge-gw"

  name = "tcp-app-profile"
  type = "tcp"
}
```

## Example Usage 2 (HTTP Cookie based Application Profile)

```hcl
resource "vcd_lb_app_profile" "http" {
  org          = "my-org"
  vdc          = "my-org-vdc"
  edge_gateway = "my-edge-gw"

  name = "http-profile"
  type = "http"

  http_redirect_url              = "/service-one"
  persistence_mechanism          = "cookie"
  cookie_name                    = "JSESSIONID"
  cookie_mode                    = "insert"
  insert_x_forwarded_http_header = "true"
}
```

## Argument Reference

The following arguments are supported:

* `org` - (Optional) The name of organization to use, optional if defined at provider level. Useful when connected as sysadmin working across different organisations
* `vdc` - (Optional) The name of VDC to use, optional if defined at provider level
* `edge_gateway` - (Required) The name of the edge gateway on which the application profile is to be created
* `name` - (Required) Application profile name
* `type` - (Required) Protocol type used to send requests to the server. One of `tcp`, `udp`,
`http`, or `https`
* `enable_ssl_passthrough` - (Optional) Enable SSL authentication to be passed through to the
virtual server. Otherwise SSL authentication takes place at the destination address
* `http_redirect_url` - (Optional) The URL to which traffic that arrives at the destination address
should be redirected. Only applies for types `http` and `https`
* `persistence_mechanism` - (Optional) Persistence mechanism for the profile. One of 'cookie',
'ssl-sessionid', 'sourceip'
* `cookie_name` - (Optional) Used to uniquely identify the session the first time a client accesses
the site. The load balancer refers to this cookie when connecting subsequent requests in the
session, so that they all go to the same virtual server. Only applies for
`persistence_mechanism` 'cookie'
* `cookie_mode` - (Optional) The mode by which the cookie should be inserted. One of 'insert', 
'prefix', or 'appsession'
* `expiration` - (Optional) Length of time in seconds that persistence stays in effect
* `insert_x_forwarded_http_header` - (Optional) Enables 'X-Forwarded-For' header for identifying
the originating IP address of a client connecting to a Web server through the load balancer.
Only applies for types `http` and `https`
* `enable_pool_side_ssl` - (Optional) Enable to define the certificate, CAs, or CRLs used to
authenticate the load balancer from the server side. **Note:** This resource does not currently
support attaching Pool and Virtual Server certificates therefore this toggle only enables it. To
make it fully work certificates must be currently attached manually.

## Attribute Reference

The following attributes are exported on this resource:

* `id` - The NSX ID of the load balancer application profile

## Importing

~> **Note:** The current implementation of Terraform import can only import resources into the state.
It does not generate configuration. [More information.](https://www.terraform.io/docs/import/)

An existing load balancer application profile can be [imported][docs-import] into this resource
via supplying the full dot separated path for load balancer application profile. An example is
below:

[docs-import]: https://www.terraform.io/docs/import/

```
terraform import vcd_lb_app_profile.imported my-org.my-org-vdc.my-edge-gw.my-lb-app-profile
```

The above would import the application profile named `my-lb-app-profile` that is defined on edge
gateway `my-edge-gw` which is configured in organization named `my-org` and vDC named `my-org-vdc`.
