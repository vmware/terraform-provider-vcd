* `resource/vcd_vapp_vm` and `datasource/vcd_vapp_vm` can report `network.X.is_primary` attribute
  incorrectly when VM is imported to Terraform and NIC indexes in vCD do not start with 0. [[#512](https://github.com/terraform-providers/terraform-provider-vcd/issues/512)] 
