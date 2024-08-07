* Fix [Issue 1236](https://github.com/vmware/terraform-provider-vcd/issues/1236):
  `list_mode="import"` of data source `vcd_resource_list` created wrong import statements when VCD items names have special
  characters [GH-1302]
* Fix [Issue 1236](https://github.com/vmware/terraform-provider-vcd/issues/1236):
  `list_mode="hierarchy"` of data source `vcd_resource_list` repeated the parent element twice when obtaining the hierarchy [GH-1302]
