* Make `license_type` attribute on **vcd_nsxt_alb_controller** optional as it is not used from VCD v10.4 onwards [GH-878]
* Add `supported_feature_set` to **vcd_nsxt_alb_service_engine_group** resource and data source to be compatible with VCD v10.4, which replaces the **vcd_nsxt_alb_controller** `license_type` [GH-878]
* Add `supported_feature_set` to **vcd_nsxt_alb_settings** resource and data source to be compatible with VCD v10.4, which replaces the **vcd_nsxt_alb_controller** `license_type` [GH-878]
