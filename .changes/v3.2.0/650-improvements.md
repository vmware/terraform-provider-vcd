* `vcd_edgegateway` resource and datasource throws error (on create, import and datasource read) and refers to `vcd_nsxt_edgegateway` for NSX-T backed VDC [GH-650]
* `vcd_nsxt_edgegateway` resource and datasource throws error (on create, import and datasource read) and refers to `vcd_edgegateway` for NSX-V backed VDC [GH-650]
* `vcd_network_isolated`and `vcd_network_routed` throw warnings on create and errors on import by referring to `vcd_network_isolated_v2`and `vcd_network_routed_v2` for NSX VDCs [GH-650]
* `vcd_vapp_network` throws error when `org_network_name` is specified for NSX-T VDC (because NSX-T networks cannot be attached to vApp networks) [GH-650]
