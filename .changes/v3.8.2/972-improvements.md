* Add `catalog_id` to resource and data source `vcd_catalog_media` to allow handling similarly to `vcd_catalog_vapp_template` [GH-972]
* Change `vcd_catalog`, `vcd_catalog_media`, `vcd_catalog_vapp_template`, and `vcd_catalog_item` to access theis entities without the need to use a full Org object, thus allowing the access to shared catalogs from other orgs.
