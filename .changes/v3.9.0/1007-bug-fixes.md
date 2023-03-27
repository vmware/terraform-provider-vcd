* Add `prefix_length` field to `vcd_vapp_network` as creating IPv6 vApp networks was not supported due to the lack of a suitable subnet representation (Issue #999) [GH-1007, GH-1031]
* Remove incorrect default value from `vcd_vapp_network` `netmask` field, as it prevents using IPV6 networks. Users of already defined resources need to add a `netmask = "255.255.255.0"` when using Ipv4 [GH-1007]
