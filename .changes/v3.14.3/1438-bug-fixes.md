* Fix [Issue 1414](https://github.com/vmware/terraform-provider-vcd/issues/1414) in `vcd_resource_list`
  with `resource_type = "vcd_vapp_vm"` and `list_mode = "hierarchy"`: without `parent`, the list had an
  empty field before the vApp name [GH-1438]
