* Fix Issue #769 "Plugin did not respond : terraform-provider-vcd may crash with Terraform 1.1+ on some OSes".
  The consequences of this fix are that some messages that were directed at the standard output (such as
  progress percentage during uploads or suggestions when using outdated resources) are now written to the regular
  log file (`go-vcloud-director.log`) using the special tag `[SCREEN]` for easy filtering. [GH-771]
  
