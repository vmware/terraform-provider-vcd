---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_nsxt_alb_virtual_service_http_sec_rules"
sidebar_current: "docs-vcd-resource-nsxt-alb-virtual-service-http-sec-rules"
description: |-
  Provides a resource to manage ALB Service Engine Groups policies for HTTP security rules. HTTP security 
  rules allow users to configure allowing or denying certain requests, to close the TCP connection, 
  to redirect a request to HTTPS, or to apply a rate limit.
---

# vcd\_nsxt\_alb\_virtual\_service\_http\_sec\_rules

Supported in provider *v3.14+* and VCD 10.5+ with NSX-T and ALB.

Provides a resource to manage ALB Service Engine Groups policies for HTTP security rules. HTTP security 
rules allow users to configure allowing or denying certain requests, to close the TCP connection, 
to redirect a request to HTTPS, or to apply a rate limit.

## Example Usage

```hcl
resource "vcd_nsxt_alb_virtual_service_http_sec_rules" "example" {
  virtual_service_id = vcd_nsxt_alb_virtual_service.test.id

  rule {
    name    = "sec-redirect-to-https"
    active  = true
    logging = true
    match_criteria {
      protocol_type = "HTTP"
    }

    actions {
      redirect_to_https = "80"
    }
  }

  rule {
    name    = "sec-connection-allow"
    active  = true
    logging = false
    match_criteria {
      client_ip_address {
        criteria     = "IS_IN"
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
      connections = "ALLOW"
    }
  }

  rule {
    name    = "sec-connection-close"
    active  = true
    logging = true
    match_criteria {
      path {
        criteria = "CONTAINS"
        paths    = ["/123", "/234"]
      }
    }

    actions {
      connections = "CLOSE"
    }
  }

  rule {
    name    = "sec-response"
    active  = true
    logging = true
    match_criteria {
      client_ip_address {
        criteria     = "IS_NOT_IN"
        ip_addresses = ["2.1.1.1", "6.2.2.2"]
      }

      service_ports {
        criteria = "IS_IN"
        ports    = [80, 81]
      }

      protocol_type = "HTTP"
    }

    actions {
      send_response {
        content      = base64encode("PERMISSION DENIED")
        content_type = "text/plain"
        status_code  = "403"
      }
    }
  }

  rule {
    name    = "sec-rate-limit-close-connection"
    active  = true
    logging = true
    match_criteria {
      request_headers {
        criteria = "DOES_NOT_BEGIN_WITH"
        name     = "X"
        values   = ["value1", "value2"]
      }
    }

    actions {
      rate_limit {
        count                   = "10000"
        period                  = "2000"
        action_close_connection = true
      }
    }
  }

  rule {
    name    = "sec-rate-limit-redirect"
    active  = true
    logging = true
    match_criteria {
      request_headers {
        criteria = "DOES_NOT_BEGIN_WITH"
        name     = "X"
        values   = ["value1", "value2"]
      }
    }

    actions {
      rate_limit {
        count  = "10000"
        period = "2000"
        action_redirect {
          protocol    = "HTTPS"
          port        = 80
          status_code = 302
          host        = "other-host"
          path        = "/"
          keep_query  = true
        }
      }
    }
  }

  rule {
    name    = "sec-rate-limit-local-resp"
    active  = true
    logging = true
    match_criteria {
      query = ["546", "666"]
    }

    actions {
      rate_limit {
        count  = "10000"
        period = "2000"
        action_local_response {
          content      = base64encode("PERMISSION DENIED")
          content_type = "text/plain"
          status_code  = "403"
        }
      }
    }
  }

  rule {
    name    = "one-criteria"
    active  = true
    logging = true
    match_criteria {
      cookie {
        criteria = "DOES_NOT_END_WITH"
        name     = "does-not-name"
        value    = "does-not-value"
      }
    }

    actions {
      redirect_to_https = "80"
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
 * `criteria` - (Required) One of `BEGINS_WITH`, `DOES_NOT_BEGIN_WITH`, `CONTAINS`,
   `DOES_NOT_CONTAIN`, `ENDS_WITH`, `DOES_NOT_END_WITH`, `EQUALS`, `DOES_NOT_EQUAL`, `REGEX_MATCH`,
   `REGEX_DOES_NOT_MATCH`
 * `paths` - (Required) A set of patchs to match given criteria
* `query` - (Optional) HTTP request query strings to match
* `request_headers` - (Optional) One or more blocks of request headers to match
 * `criteria` - (Required) One of `BEGINS_WITH`, `DOES_NOT_BEGIN_WITH`, `CONTAINS`,
   `DOES_NOT_CONTAIN`, `ENDS_WITH`, `DOES_NOT_END_WITH`, `EQUALS`, `DOES_NOT_EQUAL`, `REGEX_MATCH`,
   `REGEX_DOES_NOT_MATCH`
 * `name` - (Required) Name of the header to match
 * `values` - (Required) A set of strings values that should match header value
* `cookie` - (Optional) A block 
 * `criteria` - (Required) One of `BEGINS_WITH`, `DOES_NOT_BEGIN_WITH`, `CONTAINS`,
   `DOES_NOT_CONTAIN`, `ENDS_WITH`, `DOES_NOT_END_WITH`, `EQUALS`, `DOES_NOT_EQUAL`, `REGEX_MATCH`,
   `REGEX_DOES_NOT_MATCH`
 * `name` - (Required) Name of the HTTP cookie to match
 * `value` - (Required) A cookie value to match


<a id="rule-action-block"></a>
## Actions

One of the below actions should be specified per rule

* `redirect_to_https` - (Optional) Port number that should be redirected to HTTPS
* `connections` - (Optional) One of `ALLOW` or `CLOSE`
* `rate_limit` - (Optional) Rate based action
 * `count` - (Required) Amount of connections that are permitted within each period before triggering the action
 * `period` - (Required) Time value in seconds to enforce rate count.
 * `action_close_connection` - (Optional) Boolean value to mark that the connection should be closed
   if `count` limit is hit.
 * `action_redirect` - (Optional)
      * `protocol` - (Required) One of `HTTP`, `HTTPS`
      * `port` - (Required) Destination port for redirect
      * `status_code` - (Required) Status code to use for redirect. One of `301`, `302`, `307`
      * `host` - (Required) Host, to which the request should be redirected
      * `path` - (Required) Path to which the request should be redirected
      * `keep_query` - (Required) Boolean value to mark if query part be preserved or not
 * `action_local_response` - (Optional)
      * `content` - (Required) Base64 encoded content. Terraform function `base64encode` can be used.
      * `content_type` - (Required) Mime type of content. E.g. `text/plain`
      * `status_code` - (Required) Status code that should be sent. E.g. 403
* `send_response` - (Optional) Send a customized response
 * `content` - Base64 encoded content. Terraform function `base64encode` can be used.
 * `content_type` - Mime type of content. E.g. `text/plain`
 * `status_code` - (Required) Status code that should be sent. E.g. 403

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
