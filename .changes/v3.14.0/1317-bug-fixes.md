* Fix [Issue 1205](https://github.com/vmware/terraform-provider-vcd/issues/1205) in `vcd_vapp_vm`
  and `vcd_vm` resources where not setting `ip_allocation_mode` in a `network` block would cause a 500 error [GH-1317]
