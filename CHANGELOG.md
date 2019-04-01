## 2.2.0 (Unreleased)

FEATURES:

* `vcd_vapp_vm` - Ability to add metadata to a VM. For previous behaviour please see `BACKWARDS INCOMPATIBILITIES` ([#158](https://github.com/terraform-providers/terraform-provider-vcd/issues/158))

BUG FIXES:

* `vcd_vapp` - Ability to add metadata to empty vApp. For previous behaviour please see `BACKWARDS INCOMPATIBILITIES` ([#158](https://github.com/terraform-providers/terraform-provider-vcd/issues/158))

BACKWARDS INCOMPATIBILITIES / NOTES:

* `vcd_vapp` - Metadata is no longer added to first VM in vApp it will be added to vApp directly instead. ([#158](https://github.com/terraform-providers/terraform-provider-vcd/issues/158))

## 2.1.0 (March 27, 2019)

NOTES:
* Please look for "v2.1+" keyword in documentation which is used to emphasize new features.
* Project switched to using Go modules, while `vendor` is left for backwards build compatibility only. It is worth having a
look at [README.md](README.md) to understand how Go modules impact build and development ([#178](https://github.com/terraform-providers/terraform-provider-vcd/issues/178))
* Project dependency of github.com/hashicorp/terraform updated from v0.10.0 to v0.11.13 ([#181](https://github.com/terraform-providers/terraform-provider-vcd/issues/181))
* MaxRetryTimeout is shared with underlying SDK `go-vcloud-director` ([#189](https://github.com/terraform-providers/terraform-provider-vcd/issues/189))
* Improved testing functionality ([#166](https://github.com/terraform-providers/terraform-provider-vcd/issues/166))

FEATURES:

* **New Resource:** disk resource - `vcd_independent_disk` ([#188](https://github.com/terraform-providers/terraform-provider-vcd/issues/188))
* resource/vcd_vapp_vm has ability to attach independent disk ([#188](https://github.com/terraform-providers/terraform-provider-vcd/issues/188))
* **New Resource:** vApp network - `vcd_vapp_network` ([#155](https://github.com/terraform-providers/terraform-provider-vcd/issues/155))
* resource/vcd_vapp_vm has ability to use vApp network ([#155](https://github.com/terraform-providers/terraform-provider-vcd/issues/155))

IMPROVEMENTS:

* resource/vcd_inserted_media now supports force ejecting on running VM ([#184](https://github.com/terraform-providers/terraform-provider-vcd/issues/184))
* resource/vcd_vapp_vm now support CPU cores configuration ([#174](https://github.com/terraform-providers/terraform-provider-vcd/issues/174))

BUG FIXES:

* resource/vcd_vapp, resource/vcd_vapp_vm add vApp status handling when environment is very fast ([#68](https://github.com/terraform-providers/terraform-provider-vcd/issues/68))
* resource/vcd_vapp_vm add additional validation to check if vApp template is OK [[#157](https://github.com/terraform-providers/terraform-provider-vcd/issues/157)] 

## 2.0.0 (January 30, 2019)

Please look for "v2.0+" keyword in documentation which is used to emphasize changes and new features.

ARCHITECTURAL:

* Vendor (vCD Golang SDK) switch from the old govcloudair to the newly supported [go-vcloud-director](https://github.com/vmware/go-vcloud-director)

FEATURES:

* vCD 8.2, 9.0, 9.1 and 9.5 version support
* Sys admin login support (required to support new higher privileged operations) - `provider.org = "System"` or `provider.sysorg = "System"`
* Ability to select Org and VDC at resource level - `org` and `vdc` parameters
* New Org resource - `vcd_org`
* New Catalog resource - `vcd_catalog`
* New Catalog item resource (upload OVA) - `vcd_catalog_item`
* New Catalog media resource (upload ISO) - `vcd_catalog_media`
* New direct and isolated Org VDC network resources (complements the routed network) - `vcd_network_direct`, `vcd_network_isolated` and `vcd_network_routed`
* DNAT protocol and ICMP sub type setting - `vcd_dnat.protocol` and `vcd_dnat.icmp_sub_type`
* Ability to accept EULAs when deploying VM - `vcd_vapp_vm.accept_all_eulas`
* Setting to log API calls for troubleshooting - `provider.logging` and `provider.logging_file`

IMPROVEMENTS:

* Fixes for guest customization issues
* Improvements to error handling and error messages
* New tests and test framework improvements
* Provisional support for connection caching (disabled by default)

BACKWARDS INCOMPATIBILITIES / NOTES:

* Resource `vcd_network` deprecated in favor of a new name `vcd_network_routed`
* Previously deprecated parameter `provider.maxRetryTimeout` removed completely in favor of `provider.max_retry_timeout`

TESTS:

* Test configuration is now included in a file (create `vcd_test_config.json` from `sample_vcd_test_config.json`) instead of being defined by environment variables

IMPROVEMENTS:

* `vcd_vapp` - Fixes an issue with Networks in vApp templates being required, also introduced in 0.1.2 ([#38](https://github.com/terraform-providers/terraform-provider-vcd/issues/38))

FEATURES:

* `vcd_vapp` - Add support for defining shared vcd_networks ([#46](https://github.com/terraform-providers/terraform-provider-vcd/pull/46))
* `vcd_vapp` - Added options to configure dhcp lease times ([#47](https://github.com/terraform-providers/terraform-provider-vcd/pull/47))


## 1.0.0 (August 17, 2017)

IMPROVEMENTS:

* `vcd_vapp` - Fixes an issue with storage profiles introduced in 0.1.2 ([#39](https://github.com/terraform-providers/terraform-provider-vcd/issues/39))

BACKWARDS INCOMPATIBILITIES / NOTES:

* provider: Deprecate `maxRetryTimeout` in favour of `max_retry_timeout` ([#40](https://github.com/terraform-providers/terraform-provider-vcd/issues/40))

## 0.1.3 (August 09, 2017)

IMPROVEMENTS:

* `vcd_vapp` - Setting the computer name regardless of init script ([#31](https://github.com/terraform-providers/terraform-provider-vcd/issues/31))
* `vcd_vapp` - Fixes the power_on issue introduced in 0.1.2 ([#33](https://github.com/terraform-providers/terraform-provider-vcd/issues/33))
* `vcd_vapp` - Fixes issue with allocated IP address affecting tf plan ([#17](https://github.com/terraform-providers/terraform-provider-vcd/issues/17) & [#29](https://github.com/terraform-providers/terraform-provider-vcd/issues/29))
* `vcd_vapp_vm` - Setting the computer name regardless of init script ([#31](https://github.com/terraform-providers/terraform-provider-vcd/issues/31))
* `vcd_firewall_rules` - corrected typo in docs ([#35](https://github.com/terraform-providers/terraform-provider-vcd/issues/35))


## 0.1.2 (August 03, 2017)

IMPROVEMENTS:

* Possibility to add OVF parameters to a vApp ([#1](https://github.com/terraform-providers/terraform-provider-vcd/pull/1))
* Added storage profile support  ([#23](https://github.com/terraform-providers/terraform-provider-vcd/pull/23))

## 0.1.1 (June 28, 2017)

FEATURES:

* **New VM Resource**: `vcd_vapp_vm` ([#9](https://github.com/terraform-providers/terraform-provider-vcd/issues/9))
* **New VPN Resource**: `vcd_edgegateway_vpn`

IMPROVEMENTS:

* resource/vcd_dnat: Added a new (optional) param translated_port ([#14](https://github.com/terraform-providers/terraform-provider-vcd/issues/14))

## 0.1.0 (June 20, 2017)

NOTES:

* Same functionality as that of Terraform 0.9.8. Repacked as part of [Provider Splitout](https://www.hashicorp.com/blog/upcoming-provider-changes-in-terraform-0-10/)
