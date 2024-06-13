* Fix [Issue #1258](https://github.com/vmware/terraform-provider-vcd/issues/1258): `vcd_cse_kubernetes_cluster` fails
during creation when the chosen network belongs to a VDC Group [GH-1266]
* Fix [Issue #1265](https://github.com/vmware/terraform-provider-vcd/issues/1265): The `kubeconfig` attribute from
  `vcd_cse_kubernetes_cluster` resource and data source is now marked as sensitive [GH-1266]