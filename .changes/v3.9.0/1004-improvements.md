* Resources `vcd_vapp_network` and `vcd_vapp_org_network` add convenience flag
  `reboot_vapp_on_removal`. When enabled, it will power off parent vApp (and power back on after
  if it was before) during vApp network removal. This improves workflows with VCD 10.4.1+ which
  returns an error when removing vApp networks from powered on vApps [GH-1004]
