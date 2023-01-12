* Change `vcd_catalog`, `vcd_catalog_media`, `vcd_catalog_vapp_template`, and `vcd_catalog_item` to access their entities without the need to use a full Org object, thus allowing the access to shared catalogs from other organizations (Issue #960) [GH-972]
* Remove unnecessary URL checks from `vcd_subscribed_catalog` creation, to allow subscribing to non-VCD entities, such as vSphere shared library [GH-972]

