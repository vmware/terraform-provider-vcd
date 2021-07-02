/*
VMware Cloud Director (starting with 10.1) supports IPSec VPN. IPSec VPN offers site-to-site
connectivity between an edge gateway and remote sites which also use NSX-T Data Center or which have
either third-party hardware routers or VPN gateways that support IPSec.
Here is a quick minimal example to configure IPSec VPN Tunnel on NSX-T Edge Gateway using Terraform:
*/

data "vcd_nsxt_edgegateway" "existing" {
  org  = "org"
  vdc  = "nsxt-vdc"
  name = "nsxt-gateway"
}

resource "vcd_nsxt_ipsec_vpn_tunnel" "first-tunnel" {
  org = "org"
  vdc = "nsxt-vdc"

  edge_gateway_id = data.vcd_nsxt_edgegateway.existing.id

  name = "IPSec VPN tunnel 3.3.0"
  # The pre-shared key must be the same on the other end of the IPSec VPN tunnel. 
  pre_shared_key = "secret-shared-key"

  # Primary IP address extracted from Edge Gateway data source
  local_ip_address = tolist(data.vcd_nsxt_edgegateway.existing.subnet)[0].primary_ip
  local_networks   = ["10.10.10.0/24"]
  # Remote peer IP address
  remote_ip_address = "1.2.3.4"
  remote_networks   = ["192.168.1.0/24", "192.168.10.0/24", "192.168.20.0/28"]
}

/*
This example uses default security profile, but it can be customized using 
security_profile_customization block (https://registry.terraform.io/providers/vmware/vcd/latest/docs/resources/nsxt_ipsec_vpn_tunnel#security-profile-customization).
  */
