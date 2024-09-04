* Fix [Issue 1216](https://github.com/vmware/terraform-provider-vcd/issues/1216) in `vcd_org_vdc`
  which failed on creation when `vm_placement_policy_ids` were set but `default_compute_policy_id`
  was not declared (System default was used instead) [GH-1313]
