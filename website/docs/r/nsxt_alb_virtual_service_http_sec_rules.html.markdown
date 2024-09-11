---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_nsxt_alb_virtual_service_http_sec_rules"
sidebar_current: "docs-vcd-resource-nsxt-alb-virtual-service-http-sec-rules"
description: |-
  Provides a resource to manage ALB Service Engine Groups policies for HTTP requests. HTTP security 
  rules allow users to configure allowing or denying certain requests, to close the TCP connection, 
  to redirect a request to HTTPS, or to apply a rate limit.
---

# vcd\_nsxt\_alb\_virtual\_service\_http\_sec\_rules

Supported in provider *v3.14+* and VCD 10.5+ with NSX-T and ALB.

Provides a resource to manage ALB Service Engine Groups policies for HTTP requests. HTTP security 
rules allow users to configure allowing or denying certain requests, to close the TCP connection, 
to redirect a request to HTTPS, or to apply a rate limit.

## Example Usage ()

```hcl

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

* `` - (Optional) 

<a id="rule-action-block"></a>
## Actions

* `` - (Optional) 


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
