* Deprecated `template_name` in favor of `vapp_template_id` in `vcd_vapp_vm` and `vcd_vm` to be able to use unique URNs instead
  of catalog dependent names [GH-931]
* Deprecated `boot_image` in favor of `boot_image_id` in `vcd_vapp_vm` and `vcd_vm` to be able to use URNs instead
  of catalog dependent names [GH-931]
* Deprecated `catalog_name` in favor of `vapp_template_id` or `boot_image_id`, which don't require a catalog name anymore [GH-931]
