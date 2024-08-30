* Fix [Issue 1216](https://github.com/vmware/terraform-provider-vcd/issues/1216) in `vcd_org_vdc`
  that would fail on creation when `vm_placement_policy_ids` are set but one does not want to declare a `default_compute_policy_id`
  (to use the System default instead) [GH-1313]
