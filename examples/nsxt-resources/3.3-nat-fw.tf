/*
Release 3.3.0 adds support for NSX-T Edge Gateway management capabilities: Security Groups, IP Sets,
Application Port Profiles, Firewalls, and NAT Rules.
Here is a simplified example utilizing mentioned resources (they all have corresponding data sources
should one need to access existing configuration data).
Scenario:
* An NSX-T Routed Org Network for database servers named "database-servers" with a Security Group
  containing it
* An IP Set with defined subnet containing Application servers
* Application Port profile defined for Port TCP 3306 so that Application Servers can access database
  servers
* Firewall rule to ALLOW outgoing traffic from Application Servers (IP set) to Security Group
  containing database servers for TCP Port 3306 (defined by Application Port Profile)
* Sets up NAT (DNAT + SNAT) with external address being primary IP (external port 443 is DNAT'ed)
  address on Edge Gateway so that it can accept public traffic and internal IP Address being in the
  "Application Servers" subnet
Note. Some entities - like VMs are missing here to shorten then example
*/

data "vcd_nsxt_edgegateway" "existing" {
  org  = "sample-org"
  vdc  = "nsxt-vdc"
  name = "nsxt-gw"
}

resource "vcd_network_routed_v2" "db-net" {
  org  = "sample-org"
  vdc  = "nsxt-vdc"
  name = "database-servers"

  edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id

  gateway       = "10.10.10.1"
  prefix_length = 24

  static_ip_pool {
    start_address = "10.10.10.10"
    end_address   = "10.10.10.20"
  }
}

resource "vcd_nsxt_ip_set" "app-servers" {
  org = "sample-org"
  vdc = "nsxt-vdc"

  edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id

  name        = "Application server IP Set"
  description = "Application server IP addresses"

  ip_addresses = [
    "192.168.1.0/28",
  ]
}


resource "vcd_nsxt_security_group" "db-servers" {
  org = "sample-org"
  vdc = "nsxt-vdc"

  edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id

  name        = "db-servers"
  description = "Database servers"

  member_org_network_ids = [vcd_network_routed_v2.db-net.id]
}


resource "vcd_nsxt_app_port_profile" "tcp-3306" {
  org = "sample-org"
  vdc = "nsxt-vdc"

  name        = "TCP 3306"
  description = "Database access"

  scope = "TENANT"

  app_port {
    protocol = "TCP"
    port     = ["3306"]
  }
}

resource "vcd_nsxt_firewall" "testing" {
  org = "sample-org"
  vdc = "nsxt-vdc"

  edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id

  rule {
    action               = "ALLOW"
    name                 = "Allow traffic from App Servers to Database servers"
    direction            = "OUT"
    ip_protocol          = "IPV4"
    source_ids           = [vcd_nsxt_ip_set.app-servers.id]
    destination_ids      = [vcd_nsxt_security_group.db-servers.id]
    app_port_profile_ids = [vcd_nsxt_app_port_profile.tcp-3306.id]
  }
}

resource "vcd_nsxt_nat_rule" "dnat" {
  org = "sample-org"
  vdc = "nsxt-vdc"

  edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id

  name        = "App Server DNAT"
  description = "DNAT rule from primary Edge Gateway IP to single IP"
  rule_type   = "DNAT"

  # Using primary_ip from edge gateway
  external_address = tolist(data.vcd_nsxt_edgegateway.existing.subnet)[0].primary_ip
  # DNAT rule to one of application servers
  internal_address    = "192.168.1.10"
  app_port_profile_id = vcd_nsxt_app_port_profile.tcp-3306.id
  dnat_external_port  = "443"
}

resource "vcd_nsxt_nat_rule" "snat" {
  org = "sample-org"
  vdc = "nsxt-vdc"

  edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id

  name      = "App Server SNAT"
  rule_type = "SNAT"

  external_address = tolist(data.vcd_nsxt_edgegateway.existing.subnet)[0].primary_ip
  internal_address = "192.168.1.10"
}
