---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_nsxt_alb_virtual_service_http_req_rules"
sidebar_current: "docs-vcd-datasource-nsxt-alb-virtual-service-http-req-rules"
description: |-
  Provides a data source to read ALB Service Engine Groups policies for HTTP requests. HTTP request 
  rules modify requests before they are either forwarded to the application, used as a basis for 
  content switching, or discarded.
---

# vcd\_nsxt\_alb\_virtual\_service\_http\_req\_rules

Supported in provider *v3.14+* and VCD 10.5+ with NSX-T and ALB.

Provides a data source to read ALB Service Engine Groups policies for HTTP requests. HTTP request 
rules modify requests before they are either forwarded to the application, used as a basis for 
content switching, or discarded.

## Example Usage

```hcl
data "vcd_nsxt_alb_virtual_service_http_req_rules" "request-rules" {
  virtual_service_id = vcd_nsxt_alb_virtual_service.test.id
}
```

## Argument Reference

The following arguments are supported:

* `virtual_service_id` - (Required) An ID of existing ALB Virtual Service.

## Attribute Reference

All the arguments and attributes defined in
[`vcd_nsxt_alb_virtual_service_http_req_rules.html`](/providers/vmware/vcd/latest/docs/resources/nsxt_alb_virtual_service_http_req_rules)
resource are available.