---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_nsxt_alb_virtual_service_http_req_rules"
sidebar_current: "docs-vcd-resource-nsxt-alb-virtual-service-http-req-rules"
description: |-
  Provides a resource to manage ALB Service Engine Groups policies for HTTP requests. HTTP request 
  rules modify requests before they are either forwarded to the application, used as a basis for 
  content switching, or discarded.
---

# vcd\_nsxt\_alb\_virtual\_service\_http\_req\_rules

Supported in provider *v3.14+* and VCD 10.5+ with NSX-T and ALB.

Provides a resource to manage ALB Service Engine Groups policies for HTTP requests. HTTP request 
rules modify requests before they are either forwarded to the application, used as a basis for 
content switching, or discarded.

## Example Usage

```hcl
resource "vcd_nsxt_alb_virtual_service_http_req_rules" "example" {
  virtual_service_id = data.vcd_nsxt_alb_virtual_service.test.id

  rule {
    name   = "criteria-max-rewrite"
    active = true
    match_criteria {
      client_ip_address {
        criteria     = "IS_NOT_IN"
        ip_addresses = ["1.1.1.1", "2.2.2.2"]
      }

      service_ports {
        criteria = "IS_IN"
        ports    = [80, 81]
      }

      protocol_type = "HTTP"

      http_methods {
        criteria = "IS_IN"
        methods  = ["COPY", "HEAD"]
      }
      path {
        criteria = "CONTAINS"
        paths    = ["/123", "/234"]
      }
      query = ["546", "666"]

      request_headers {
        criteria = "DOES_NOT_BEGIN_WITH"
        name     = "X"
        values   = ["value1", "value2"]
      }

      request_headers {
        criteria = "DOES_NOT_EQUAL"
        name     = "Y-DOES-NOT"
        values   = ["value1", "value2"]
      }

      cookie {
        criteria = "DOES_NOT_END_WITH"
        name     = "does-not-name"
        value    = "does-not-value"
      }

    }

    actions {
      rewrite_url {
        host_header   = "X-HOST-HEADER"
        existing_path = "/123"
        keep_query    = true
        query         = "rewrite"
      }
    }
  }

  rule {
    name   = "other"
    active = true
    match_criteria {
      client_ip_address {
        criteria     = "IS_NOT_IN"
        ip_addresses = ["1.1.1.1", "2.2.2.2"]
      }
    }

    actions {
      modify_header {
        action = "REMOVE"
        name   = "X-REMOVE-HEADER"
      }
      modify_header {
        action = "ADD"
        name   = "X-ADDED-HEADER"
        value  = "value"
      }

      modify_header {
        action = "REPLACE"
        name   = "X-EXISTING-HEADER"
        value  = "new-value"
      }
    }
  }
}
```

## Argument Reference

The following arguments are supported:

* `virtual_service_id` - (Required) An ID of existing ALB Virtual Service.
* `rule` - (Required) One or more [rule](#rule-block) blocks with matching criteria and actions.

<a id="rule-block"></a>
## Rule

* `name` - (Required) Name of the rule
* `active` - (Optional) Defines if the rule is active. Default `true`
* `logging` - (Optional) Defines if the requests that match should be logged. Default `false`
* `match_criteria` - (Required) A block of [criteria](#rule-criteria-block) to match the requests
* `actions` - (Required) A block of [actions](#rule-action-block) to perform with requests that match

<a id="rule-criteria-block"></a>
## Match Criteria

One or more of criteria can be specified to match traffic. At least one criteria is required

* `client_ip_address` - (Optional) Match the rule based on client IP address rules
 * `criteria` - (Required) One of `IS_IN`, `IS_NOT_IN`
 * `ip_addresses` - (Required) A set of IP addresses to match
* `service_ports` - (Optional) Match the rule based on service ports
 * `criteria` - (Required) One of `IS_IN`, `IS_NOT_IN`
 * `ports` - (Required) A set of ports to match
* `protocol_type` - (Optional) One of `HTTP` or `HTTPS`
* `http_methods` - (Optional) Defines HTTP methods that should be matched
 * `criteria` - (Required) One of `IS_IN`, `IS_NOT_IN`
 * `methods` - (Required) A set of HTTP methods from the list: `GET`, `PUT`, `POST`, `DELETE`,
   `HEAD`, `OPTIONS`, `TRACE`, `CONNECT`, `PATCH`, `PROPFIND`, `PROPPATCH`, `MKCOL`, `COPY`, `MOVE`,
   `LOCK`, `UNLOCK`
* `path` - (Optional) 
 * `criteria` - (Required) One of `BEGINS_WITH`, `DOES_NOT_BEGIN_WITH`, `CONTAINS`, `DOES_NOT_CONTAIN`, `ENDS_WITH`, `DOES_NOT_END_WITH`, `EQUALS`, `DOES_NOT_EQUAL`, `REGEX_MATCH`, `REGEX_DOES_NOT_MATCH`
 * `paths` - (Required) A set of patchs to match given criteria
* `query` - (Optional) HTTP request query strings to match
* `request_headers` - (Optional) One or more blocks of request headers to match
 * `criteria` - (Required) One of `BEGINS_WITH`, `DOES_NOT_BEGIN_WITH`, `CONTAINS`, `DOES_NOT_CONTAIN`, `ENDS_WITH`, `DOES_NOT_END_WITH`, `EQUALS`, `DOES_NOT_EQUAL`, `REGEX_MATCH`, `REGEX_DOES_NOT_MATCH`
 * `name` - (Required) Name of the header to match
 * `values` - (Required) A set of strings values that should match header value
* `cookie` - (Optional) A block 
 * `criteria` - (Required) One of `BEGINS_WITH`, `DOES_NOT_BEGIN_WITH`, `CONTAINS`, `DOES_NOT_CONTAIN`, `ENDS_WITH`, `DOES_NOT_END_WITH`, `EQUALS`, `DOES_NOT_EQUAL`, `REGEX_MATCH`, `REGEX_DOES_NOT_MATCH`
 * `name` - (Required) Name of the HTTP cookie to match
 * `value` - (Required) A cookie value to match

<a id="rule-action-block"></a>
## Actions

One or more of action blocks can be specified. At least one is required. Some may prohibit others
and API validation will return errors

* `redirect` - (Optional) Redirects the request to different location
 * `protocol` - (Required) One of `HTTP`, `HTTPS`
 * `port` - (Required) Destination port for redirect
 * `status_code` - (Required) Status code to use for redirect. One of `301`, `302`, `307`
 * `host` - (Optional) Host, to which the request should be redirected
 * `path` - (Optional) Path to which the request should be redirected
 * `keep_query` - (Optional) Boolean value to mark if query part be preserved or not
* `modify_header` - (Optional) An optional block of header modification actions
 * `action` - (Required) One of `ADD`, `REMOVE`, `REPLACE`
 * `name` - (Required) Name of the header to modify
 * `value` - (Optional) Value (only application to `ADD` and `REPLACE` actions)
* `rewrite_url` - (Optional) URL Rewrite rules, at most one block
 * `host_header` - (Required) Host to use for the rewritten URL
 * `existing_path` - (Required) Path to use for the rewritten URL
 * `keep_query` - (Optional)Whether or not to keep the existing query string when rewriting the URL
 * `query` - (Optional) Query string to use or append to the existing query string in the rewritten URL

## Importing

~> The current implementation of Terraform import can only import resources into the state.
It does not generate configuration. [More information.](https://www.terraform.io/docs/import/)

An existing ALB Virtual Service configuration can be [imported][docs-import] into this resource
via supplying path for it. An example is below:

[docs-import]: https://www.terraform.io/docs/import/

```
terraform import vcd_nsxt_alb_virtual_service_http_req_rules.imported my-org.my-org-vdc-org-vdc-group-name.my-edge-gateway.my-virtual-service-name
```

The above would import the `my-virtual-service-name` ALB Virtual Service Policy rules that are
defined in NSX-T Edge Gateway `my-edge-gateway` inside Org `my-org` and VDC or VDC Group
`my-org-vdc-org-vdc-group-name`.
