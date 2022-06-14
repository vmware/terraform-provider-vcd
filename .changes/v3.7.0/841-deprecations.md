* Deprecate `vdc` field in NSX-T Edge Gateway child entities. This field is no longer precise as
  with introduction of VDC Group support an Edge Gateway can be bound either to a VDC, either to a
  VDC Group. Parent VDC or VDC Group is now inherited from `edge_gateway_id` field. Impacted
  resources and data sources are: `vcd_nsxt_firewall`, `vcd_nsxt_nat_rule`,
  `vcd_nsxt_ipsec_vpn_tunnel`, `,vcd_nsxt_alb_settings`,
  `vcd_nsxt_alb_edgegateway_service_engine_group`, `vcd_nsxt_alb_virtual_service`,
  `vcd_nsxt_alb_pool` [GH-841]
