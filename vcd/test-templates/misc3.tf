# Remove the first '#' from the next two lines to enable options for terraform executable 
# apply-options -parallelism=1
# destroy-options -parallelism=1

# Edge gateway load balancer configuration
# v2.4.0+

variable "service_monitor_count" {
  default = 20
}

resource "vcd_lb_service_monitor" "test" {
  count = "${var.service_monitor_count}"

  org          = "{{.Org}}"
  vdc          = "{{.Vdc}}"
  edge_gateway = "{{.EdgeGateway}}"

  name        = "test-monitor-${count.index}"
  interval    = 5
  timeout     = 5
  max_retries = 3
  type        = "http"
  method      = "POST"
  send        = "{\"key\": \"value\"}"
  expected    = "HTTP/1.1"
  receive     = "OK"

  extension = {
    "content-type" = "application/json"
    "no-body"      = ""
  }
}


data "vcd_lb_service_monitor" "ds-lb" {
  count = "${var.service_monitor_count}"

  org          = "{{.Org}}"
  vdc          = "{{.Vdc}}"
  edge_gateway = "{{.EdgeGateway}}"
  name         = "${vcd_lb_service_monitor.test[count.index].name}"
}
