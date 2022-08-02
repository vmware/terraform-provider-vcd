* Fix a bug that causes `vcd_vapp_vm` to fail on creation if attribute `sizing_policy_id` is set and corresponds to a
Sizing Policy with CPU or memory defined, `template_name` is used and `power_on` is `true` [GH-883]
