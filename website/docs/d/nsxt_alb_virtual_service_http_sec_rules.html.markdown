---
layout: "vcd"
page_title: "VMware Cloud Director: vcd_nsxt_alb_virtual_service_http_sec_rules"
sidebar_current: "docs-vcd-datasource-nsxt-alb-virtual-service-http-sec-rules"
description: |-
  Provides a data source to read ALB Service Engine Groups policies for HTTP requests. HTTP security 
  rules allow users to configure allowing or denying certain requests, to close the TCP connection, 
  to redirect a request to HTTPS, or to apply a rate limit.
---

# vcd\_nsxt\_alb\_virtual\_service\_http\_sec\_rules

Supported in provider *v3.14+* and VCD 10.5+ with NSX-T and ALB.

Provides a data source to read ALB Service Engine Groups policies for HTTP requests. HTTP security 
rules allow users to configure allowing or denying certain requests, to close the TCP connection, 
to redirect a request to HTTPS, or to apply a rate limit.

## Example Usage

```hcl
data "vcd_nsxt_alb_virtual_service_http_sec_rules" "security-rules" {
  virtual_service_id = vcd_nsxt_alb_virtual_service.test.id
}
```

## Argument Reference

The following arguments are supported:

* `virtual_service_id` - (Required) An ID of existing ALB Virtual Service.

## Attribute Reference

All the arguments and attributes defined in
[`vcd_nsxt_alb_virtual_service_http_sec_rules`](/providers/vmware/vcd/latest/docs/resources/nsxt_alb_virtual_service_http_sec_rules)
resource are available.