* Added the new attributes `vapp_template_id`, `boot_image_id` to the resources `vcd_vapp_vm` and `vcd_vm` to be able
  to use unique URNs to reference vApp Templates and Media items through data sources, to build strong dependencies
  in Terraform configuration. [GH-931]
