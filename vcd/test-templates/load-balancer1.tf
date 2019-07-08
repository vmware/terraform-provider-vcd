# Remove the first '#' from the next two lines to enable options for terraform executable 
## apply-options -parallelism=1
## destroy-options -parallelism=1

# Edge gateway load balancer configuration with all separate components
# and their datasources.

# v2.4.0+

variable "component_count" {
  default = 10
}

resource "vcd_lb_service_monitor" "test" {
  count = "${var.component_count}"

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
  count = "${var.component_count}"

  org          = "{{.Org}}"
  vdc          = "{{.Vdc}}"
  edge_gateway = "{{.EdgeGateway}}"
  name         = "${vcd_lb_service_monitor.test[count.index].name}"
}


resource "vcd_lb_server_pool" "server-pool" {
  count = "${var.component_count}"

  org          = "{{.Org}}"
  vdc          = "{{.Vdc}}"
  edge_gateway = "{{.EdgeGateway}}"

  name = "test-pool-${count.index}"
  algorithm = "round-robin"
  enable_transparency = "true"

  monitor_id = "${vcd_lb_service_monitor.test[count.index].id}"

  member {
    condition = "enabled"
    name = "member1"
    ip_address = "1.1.1.1"
    port = 8443
    monitor_port = 9000
    weight = 1
    min_connections = 0
    max_connections = 100
  }

  member {
    condition = "drain"
    name = "member2"
    ip_address = "2.2.2.2"
    port = 7000
    monitor_port = 4000
    weight = 2
    min_connections = 6
    max_connections = 8
  }

  member {
    condition = "disabled"
    name = "member3"
    ip_address = "3.3.3.3"
    port = 3333
    monitor_port = 4444
    weight = 6
    min_connections = 3
    max_connections = 3
  }

  member {
    condition = "disabled"
    name = "member4"
    ip_address = "4.4.4.4"
    port = 3333
    monitor_port = 4444
    weight = 6
  }
}



data "vcd_lb_server_pool" "ds-pool" {
  count = "${var.component_count}"

  org          = "{{.Org}}"
  vdc          = "{{.Vdc}}"
  edge_gateway = "{{.EdgeGateway}}"
  name         = "${vcd_lb_server_pool.server-pool[count.index].name}"
}


resource "vcd_lb_app_profile" "test" {
  count = "${var.component_count}"

	org          = "{{.Org}}"
	vdc          = "{{.Vdc}}"
	edge_gateway = "{{.EdgeGateway}}"
  
	name           = "test-app-profile-${count.index}"
	type           = "TCP"
}

data "vcd_lb_app_profile" "test" {
  count = "${var.component_count}"
  
	org          = "{{.Org}}"
	vdc          = "{{.Vdc}}"
	edge_gateway = "{{.EdgeGateway}}"
	name         = "${vcd_lb_app_profile.test[count.index].name}"
}
