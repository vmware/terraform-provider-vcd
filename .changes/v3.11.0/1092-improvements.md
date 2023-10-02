* Resources `vcd_vapp_network` and `vcd_vapp_org_network` will additionally check if vApp is in
  `RESOLVED` (in addition to already checked `POWERED_OFF`) state before attempting a reboot when
  `reboot_vapp_on_removal` flag is set to `true` [GH-1092]
