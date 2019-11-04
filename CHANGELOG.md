## 2.6.0 (Unreleased)

IMPROVEMENTS:

* Switch to Terraform terraform-plugin-sdk v1.0.0 as per recent [HashiCorp
  recommendation](https://www.terraform.io/docs/extend/plugin-sdk.html) - [GH-382]

BUG FIXES:
* Fix `vcd_org_vdc` datasource read. When user was Organization administrator datasource failed.

## 2.5.0 (October 28, 2019)

FEATURES:
* **New Resource:** `vcd_nsxv_dnat` DNAT for advanced edge gateways using proxied NSX-V API - ([#328](https://github.com/terraform-providers/terraform-provider-vcd/issues/328))
* **New Resource:** `vcd_nsxv_snat`  SNAT for advanced edge gateways using proxied NSX-V API - ([#328](https://github.com/terraform-providers/terraform-provider-vcd/issues/328))
* **New Resource:** `vcd_nsxv_firewall_rule`  firewall for advanced edge gateways using proxied NSX-V API - ([#341](https://github.com/terraform-providers/terraform-provider-vcd/issues/341), [#358](https://github.com/terraform-providers/terraform-provider-vcd/issues/358))
* **New Data Source:** `vcd_org` Organization - ([#218](https://github.com/terraform-providers/terraform-provider-vcd/issues/218))
* **New Data Source:** `vcd_catalog` Catalog - ([#218](https://github.com/terraform-providers/terraform-provider-vcd/issues/218))
* **New Data Source:** `vcd_catalog_item` CatalogItem - ([#218](https://github.com/terraform-providers/terraform-provider-vcd/issues/218))
* **New Data Source:** `vcd_org_vdc` Organization VDC - ([#324](https://github.com/terraform-providers/terraform-provider-vcd/issues/324))
* **New Data Source:** `vcd_external_network` External Network - ([#218](https://github.com/terraform-providers/terraform-provider-vcd/issues/218))
* **New Data Source:** `vcd_edgegateway` Edge Gateway - ([#218](https://github.com/terraform-providers/terraform-provider-vcd/issues/218))
* **New Data Source:** `vcd_network_routed` Routed Network - ([#218](https://github.com/terraform-providers/terraform-provider-vcd/issues/218))
* **New Data Source:** `vcd_network_isolated` Isolated Network - ([#218](https://github.com/terraform-providers/terraform-provider-vcd/issues/218))
* **New Data Source:** `vcd_network_direct` Direct Network - ([#218](https://github.com/terraform-providers/terraform-provider-vcd/issues/218))
* **New Data Source:** `vcd_vapp` vApp - ([#218](https://github.com/terraform-providers/terraform-provider-vcd/issues/218))
* **New Data Source:** `vcd_nsxv_dnat` DNAT for advanced edge gateways using proxied NSX-V API - ([#328](https://github.com/terraform-providers/terraform-provider-vcd/issues/328))
* **New Data Source:** `vcd_nsxv_snat` SNAT for advanced edge gateways using proxied NSX-V API - ([#328](https://github.com/terraform-providers/terraform-provider-vcd/issues/328))
* **New Data Source:** `vcd_nsxv_firewall_rule` firewall for advanced edge gateways using proxied NSX-V API - ([#341](https://github.com/terraform-providers/terraform-provider-vcd/issues/341))
* **New Data Source:** `vcd_independent_disk` Independent disk - ([#349](https://github.com/terraform-providers/terraform-provider-vcd/issues/349))
* **New Data Source:** `vcd_catalog_media` Media item - ([#340](https://github.com/terraform-providers/terraform-provider-vcd/issues/340))

IMPROVEMENTS:

* Add support for vCD 10.0
* `resource/vcd_org` Add import capability and full read support ([#218](https://github.com/terraform-providers/terraform-provider-vcd/issues/218))
* `resource/vcd_catalog` Add import capability and full read support ([#218](https://github.com/terraform-providers/terraform-provider-vcd/issues/218))
* `resource/vcd_catalog_item` Add import capability and full read support ([#218](https://github.com/terraform-providers/terraform-provider-vcd/issues/218))
* `resource/vcd_external_network` Add import capability and full read support ([#218](https://github.com/terraform-providers/terraform-provider-vcd/issues/218))
* `resource/vcd_edgegateway` Add import capability and full read support ([#218](https://github.com/terraform-providers/terraform-provider-vcd/issues/218))
* `resource/vcd_network_routed` Add import capability and full read support ([#218](https://github.com/terraform-providers/terraform-provider-vcd/issues/218))
* `resource/vcd_network_isolated` Add import capability and full read support ([#218](https://github.com/terraform-providers/terraform-provider-vcd/issues/218))
* `resource/vcd_network_direct` Add import capability and full read support ([#218](https://github.com/terraform-providers/terraform-provider-vcd/issues/218))
* `resource/vcd_vapp` Add import capability and full read support ([#218](https://github.com/terraform-providers/terraform-provider-vcd/issues/218))
* `resource/vcd_network_direct`: Direct network state ID changed from network name to vCD ID 
* `resource/vcd_network_isolated`: Isolated network state ID changed from network name to vCD ID 
* `resource/vcd_network_routed`: Routed network state ID changed from network name to vCD ID 
* `resource/vcd_vapp`: vApp state ID changed from vApp name to vCD ID
* `resource/vcd_vapp`: Add properties `status` and `status_text`
* `resource/catalog_item` added catalog item metadata support [[#285](https://github.com/terraform-providers/terraform-provider-vcd/issues/285)] 
* `resource/vcd_catalog`: Catalog state ID changed from catalog name to vCD ID 
* `resource/vcd_catalog_item`: CatalogItem state ID changed from colon separated list of catalog name and item name to vCD ID 
* `resource/catalog_item` added catalog item metadata support [[#298](https://github.com/terraform-providers/terraform-provider-vcd/issues/298)] 
* `resource/catalog_media` added catalog media item metadata support ([#298](https://github.com/terraform-providers/terraform-provider-vcd/issues/298))
* `resource/vcd_vapp_vm` supports update for `network` block ([#310](https://github.com/terraform-providers/terraform-provider-vcd/issues/310))
* `resource/vcd_vapp_vm` allows to force guest customization ([#310](https://github.com/terraform-providers/terraform-provider-vcd/issues/310))
* `resource/vcd_vapp` supports guest properties ([#319](https://github.com/terraform-providers/terraform-provider-vcd/issues/319))
* `resource/vcd_vapp_vm` supports guest properties ([#319](https://github.com/terraform-providers/terraform-provider-vcd/issues/319))
* `resource/vcd_network_direct` Add computed properties (external network gateway, netmask, DNS, and DNS suffix) ([#330](https://github.com/terraform-providers/terraform-provider-vcd/issues/330))
* `vcd_org_vdc` Add import capability and full read support ([#218](https://github.com/terraform-providers/terraform-provider-vcd/issues/218))
* Upgrade Terraform SDK dependency to 0.12.8 ([#320](https://github.com/terraform-providers/terraform-provider-vcd/issues/320))
* `resource/vcd_vapp_vm` has new field `computer_name` ([#334](https://github.com/terraform-providers/terraform-provider-vcd/issues/334))
* Import functions can now use custom separators instead of "." ([#343](https://github.com/terraform-providers/terraform-provider-vcd/issues/343))
* `resource/vcd_independent_disk` Add computed properties (`iops`, `owner_name`, `datastore_name`, `is_attached`) and read support for all fields except the `size` ([#349](https://github.com/terraform-providers/terraform-provider-vcd/issues/349))
* `resource/vcd_independent_disk` Disk state ID changed from name of disk to vCD ID ([#349](https://github.com/terraform-providers/terraform-provider-vcd/issues/349))
* Import functions can now use a custom separator instead of "." ([#343](https://github.com/terraform-providers/terraform-provider-vcd/issues/343))
* `resource/vcd_catalog_media` Add computed properties (`is_iso`, `owner_name`, `is_published`, `creation_date`, `size`, `status`, `storage_profile_name`) and full read support ([#340](https://github.com/terraform-providers/terraform-provider-vcd/issues/340))
* `resource/vcd_catalog_media` MediaItem state ID changed from colon separated list of catalog name and media name to vCD ID ([#340](https://github.com/terraform-providers/terraform-provider-vcd/issues/340))
* Import functions can now use custom separators instead of "." ([#343](https://github.com/terraform-providers/terraform-provider-vcd/issues/343))
* `resource/vcd_vapp_vm` has new field `computer_name` ([#334](https://github.com/terraform-providers/terraform-provider-vcd/issues/334), [#347](https://github.com/terraform-providers/terraform-provider-vcd/issues/347))

BUG FIXES:
* Change default value for `vcd_org.deployed_vm_quota` and `vcd_org.stored_vm_quota`. It was incorrectly set at `-1` instead of `0`.
* Change Org ID from partial task ID to real Org ID during creation.
* Wait for task completion on creation and update, where tasks were not handled at all.
* `resource/vcd_firewall_rules` force recreation of the resource when attributes of the sub-element `rule` are changed (fixes a situation when it tried to update a rule).
* `resource/vcd_network_isolated` Fix definition of DHCP, which was created automatically with leftovers from static IP pool even when not requested.
* `resource/vcd_network_routed` Fix retrieval with early vCD versions. ([#344](https://github.com/terraform-providers/terraform-provider-vcd/issues/344))
* `resource/vcd_edgegateway_vpn` Required replacement every time for `shared_secret` field.  ([#361](https://github.com/terraform-providers/terraform-provider-vcd/issues/361))

DEPRECATIONS
* The ability of deploying a VM implicitly within a vApp is deprecated. Users are encouraged to set an empty vApp and
add explicit VM resources `vcd_vapp_vm`.
  For this reason, the following fields in `vcd_vapp` are deprecated:
  * `template_name`
  * `catalog_name`
  * `network_name`
  * `ip`
  * `cpus`
  * `memory`
  * `network_name`
  * `initscript`
  * `ovf`
  * `accept_all_eulas`

NOTES:
* Drop support for vCD 8.20

## 2.4.0 (July 29, 2019)

FEATURES:

* **New Resource:** `vcd_lb_service_monitor`  Load Balancer Service Monitor - ([#256](https://github.com/terraform-providers/terraform-provider-vcd/issues/256), [#290](https://github.com/terraform-providers/terraform-provider-vcd/issues/290))
* **New Resource:** `vcd_edgegateway` creates and deletes edge gateways, manages general load balancing settings - ([#262](https://github.com/terraform-providers/terraform-provider-vcd/issues/262), [#288](https://github.com/terraform-providers/terraform-provider-vcd/issues/288))
* **New Resource:** `vcd_lb_server_pool` Load Balancer Server Pool - ([#268](https://github.com/terraform-providers/terraform-provider-vcd/issues/268), [#290](https://github.com/terraform-providers/terraform-provider-vcd/issues/290), [#297](https://github.com/terraform-providers/terraform-provider-vcd/issues/297))
* **New Resource:** `vcd_lb_app_profile` Load Balancer Application profile - ([#274](https://github.com/terraform-providers/terraform-provider-vcd/issues/274), [#290](https://github.com/terraform-providers/terraform-provider-vcd/issues/290), [#297](https://github.com/terraform-providers/terraform-provider-vcd/issues/297))
* **New Resource:** `vcd_lb_app_rule` Load Balancer Application rule - ([#278](https://github.com/terraform-providers/terraform-provider-vcd/issues/278), [#290](https://github.com/terraform-providers/terraform-provider-vcd/issues/290))
* **New Resource:** `vcd_lb_virtual_server` Load Balancer Virtual Server - ([#284](https://github.com/terraform-providers/terraform-provider-vcd/issues/284), [#290](https://github.com/terraform-providers/terraform-provider-vcd/issues/290), [#297](https://github.com/terraform-providers/terraform-provider-vcd/issues/297))
* **New Resource:** `vcd_org_user`  Organization User - ([#279](https://github.com/terraform-providers/terraform-provider-vcd/issues/279))
* **New Data Source:** `vcd_lb_service_monitor` Load Balancer Service Monitor  - ([#256](https://github.com/terraform-providers/terraform-provider-vcd/issues/256), [#290](https://github.com/terraform-providers/terraform-provider-vcd/issues/290))
* **New Data Source:** `vcd_lb_server_pool`  Load Balancer Server Pool - ([#268](https://github.com/terraform-providers/terraform-provider-vcd/issues/268), [#290](https://github.com/terraform-providers/terraform-provider-vcd/issues/290), [#297](https://github.com/terraform-providers/terraform-provider-vcd/issues/297))
* **New Data Source:** `vcd_lb_app_profile` Load Balancer Application profile - ([#274](https://github.com/terraform-providers/terraform-provider-vcd/issues/274), [#290](https://github.com/terraform-providers/terraform-provider-vcd/issues/290), [#297](https://github.com/terraform-providers/terraform-provider-vcd/issues/297))
* **New Data Source:** `vcd_lb_app_rule` Load Balancer Application rule - ([#278](https://github.com/terraform-providers/terraform-provider-vcd/issues/278), [#290](https://github.com/terraform-providers/terraform-provider-vcd/issues/290))
* **New Data Source:** `vcd_lb_virtual_server` Load Balancer Virtual Server - ([#284](https://github.com/terraform-providers/terraform-provider-vcd/issues/284), [#290](https://github.com/terraform-providers/terraform-provider-vcd/issues/290), [#297](https://github.com/terraform-providers/terraform-provider-vcd/issues/297))
* **New build commands** `make test-env-init` and `make test-env-apply` can configure an empty vCD to run the test suite. See `TESTING.md` for details.
* `resource/vcd_org_vdc` added Org VDC update and full state read - ([#275](https://github.com/terraform-providers/terraform-provider-vcd/issues/275))
* `resource/vcd_org_vdc` added Org VDC metadata support - ([#276](https://github.com/terraform-providers/terraform-provider-vcd/issues/276))
* `resource/vcd_snat` added ability to choose network name and type. [[#282](https://github.com/terraform-providers/terraform-provider-vcd/issues/282)] 
* `resource/vcd_dnat` added ability to choose network name and type. ([#282](https://github.com/terraform-providers/terraform-provider-vcd/issues/282), [#292](https://github.com/terraform-providers/terraform-provider-vcd/issues/292), [#293](https://github.com/terraform-providers/terraform-provider-vcd/issues/293))

IMPROVEMENTS:
* `resource/vcd_org_vdc`: Fix ignoring of resource guarantee values - ([#265](https://github.com/terraform-providers/terraform-provider-vcd/issues/265))
* `resource/vcd_org_vdc`: Org VDC state ID changed from name to vCD ID - ([#275](https://github.com/terraform-providers/terraform-provider-vcd/issues/275))
* Change resource handling to use locking mechanism when resource parallel handling is not supported by vCD. [[#255](https://github.com/terraform-providers/terraform-provider-vcd/issues/255)] 
* Fix issue when vApp is power cycled during member VM deletion. ([#261](https://github.com/terraform-providers/terraform-provider-vcd/issues/261))
* `resource/vcd_dnat`, `resource/vcd_snat` has got full read functionality. This means that on the next `plan/apply` it will detect if configuration has changed in vCD and propose to update it.

BUG FIXES:

* `resource/vcd_dnat and resource/vcd_snat` - fix resource destroy as it would still leave NAT rule in edge gateway. Fix works if network_name and network_type is used. ([#282](https://github.com/terraform-providers/terraform-provider-vcd/issues/282))

NOTES:
* `resource/vcd_dnat` `protocol` requires lower case values to be consistent with the underlying NSX API. This may result in invalid configuration if upper case was used previously!
* `resource/vcd_dnat` default value for `protocol` field changed from upper case `TCP` to lower case `tcp`, which may result in a single update when running `plan` on a configuration with a state file from an older version.

## 2.3.0 (May 29, 2019)

IMPROVEMENTS:

* Switch to Terraform 0.12 SDK which is required for Terraform 0.12 support. HCL (HashiCorp configuration language) 
parsing behaviour may have changed as a result of changes made by the new SDK version ([#254](https://github.com/terraform-providers/terraform-provider-vcd/issues/254))

NOTES:

* Provider plugin will still work with Terraform 0.11 executable

## 2.2.0 (May 16, 2019)

FEATURES:

* `vcd_vapp_vm` - Ability to add metadata to a VM. For previous behaviour please see `BACKWARDS INCOMPATIBILITIES` ([#158](https://github.com/terraform-providers/terraform-provider-vcd/issues/158))
* `vcd_vapp_vm` - Ability to enable hardware assisted CPU virtualization for VM. It allows hypervisor nesting. ([#219](https://github.com/terraform-providers/terraform-provider-vcd/issues/219))
* **New Resource:** external network - `vcd_external_network` - ([#230](https://github.com/terraform-providers/terraform-provider-vcd/issues/230))
* **New Resource:** VDC resource `vcd_org_vdc` - ([#236](https://github.com/terraform-providers/terraform-provider-vcd/issues/236))
* resource/vcd_vapp_vm: Add `network` argument for multiple NIC support and more flexible configuration ([#233](https://github.com/terraform-providers/terraform-provider-vcd/issues/233))
* resource/vcd_vapp_vm: Add `mac` argument to store MAC address in state file ([#233](https://github.com/terraform-providers/terraform-provider-vcd/issues/233))

BUG FIXES:

* `vcd_vapp` - Ability to add metadata to empty vApp. For previous behaviour please see `BACKWARDS INCOMPATIBILITIES` ([#158](https://github.com/terraform-providers/terraform-provider-vcd/issues/158))

BACKWARDS INCOMPATIBILITIES / NOTES:

* `vcd_vapp` - Metadata is no longer added to first VM in vApp it will be added to vApp directly instead. ([#158](https://github.com/terraform-providers/terraform-provider-vcd/issues/158))
* Tests files are now all tagged. Running them through Makefile works as before, but manual execution requires specific tags. Run `go test -v .` for tags list.
* `vcd_vapp_vm` - Deprecated attributes `network_name`, `vapp_network_name`, `network_href` and `ip` in favor of `network` ([#118](https://github.com/terraform-providers/terraform-provider-vcd/issues/118))

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
