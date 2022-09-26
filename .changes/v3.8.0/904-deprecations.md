* Deprecated `default_vm_sizing_policy_id` field in `vcd_org_vdc` resource and data source. This field is misleading as it
  can contain not only VM Sizing Policies but also VM Placement Policies or vGPU Policies.
  Its replacement is the `default_compute_policy_id` attribute [GH-904]
