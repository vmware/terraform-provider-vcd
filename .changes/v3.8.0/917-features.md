* Added attribute `metadata_entry` to the following data sources:
  `vcd_catalog`, `vcd_catalog_media`, `vcd_independent_disk`, `vcd_network_direct`, `vcd_network_isolated`,
  `vcd_network_isolated_v2`, `vcd_network_routed`, `vcd_network_routed_v2`, `vcd_org`, `vcd_org_vdc`, `vcd_provider_vdc`,
  `vcd_storage_profile`, `vcd_vapp`, `vcd_vapp_vm`. This new attribute replaces `metadata`
  to add support of metadata visibility (user access levels), all the available types and domains for every metadata
  entry. [GH-917]
* Added attribute `metadata_entry` to the following resources:
  `vcd_catalog`, `vcd_catalog_media`, `vcd_independent_disk`, `vcd_network_direct`, `vcd_network_isolated`,
  `vcd_network_isolated_v2`, `vcd_network_routed`, `vcd_network_routed_v2`, `vcd_org`, `vcd_org_vdc`, `vcd_vapp`,
  `vcd_vapp_vm`. This new attribute replaces `metadata` to add support of metadata visibility (user access levels),
  all the available types and domains for every metadata entry. [GH-917]