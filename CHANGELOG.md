## 2.9.0 (Unreleased)

IMPROVEMENTS:

* `resource/vcd_vapp_vm` allows creating empty VM. New fields added `boot_image`, `os_type` and `hardware_version`. Also, supports `description` updates. [GH-484]
* Removed code that handled specific cases for API 29.0 and 30.0. This library now supports VCD versions from 9.5 to 10.1 included [GH-499]
* Added command line flags to test suite, corresponding to environment variables listed in TESTING.md [GH-505]
* `resource/vcd_vapp_vm` allows creating VM from multi VM vApp template [GH-501]

NOTES:

* Dropped support for vCD 9.1 [GH-492]

## 2.8.0 (April 16, 2020)

IMPROVEMENTS:

* `resource/vcd_network_routed`, `resource/vcd_network_direct`, and `resource/vcd_network_isolated` now support in place updates. ([#465](https://github.com/terraform-providers/terraform-provider-vcd/issues/465))
* `vcd_vapp_network`, `vcd_vapp_org_network` has now missing import documentation [[#481](https://github.com/terraform-providers/terraform-provider-vcd/issues/481)] 
* `resource/vcd_vapp_vm` and `datasource/vcd_vapp_vm` simplifies network adapter validation when it
  is not attached to network (`network.x.type=none` and `network.x.ip_allocation_mode=none`) and
  `network_dhcp_wait_seconds` is defined. This is required for vCD 10.1 support ([#485](https://github.com/terraform-providers/terraform-provider-vcd/issues/485))

BUG FIXES
* Using wrong defaults for `vcd_network_isolated` and `vcd_network_routed` DNS ([#434](https://github.com/terraform-providers/terraform-provider-vcd/issues/434))
* `external_network_gateway` not filled in datasource `vcd_network_direct` ([#450](https://github.com/terraform-providers/terraform-provider-vcd/issues/450))
* `resource/vcd_vapp_vm` sometimes reports incorrect `vcd_vapp_vm.ip` and `vcd_vapp_vm.mac` fields in deprecated network
configuration (when using `vcd_vapp_vm.network_name` and `vcd_vapp_vm.vapp_network_name` parameters instead of
`vcd_vapp_vm.network` blocks) ([#478](https://github.com/terraform-providers/terraform-provider-vcd/issues/478))
* `resource/vcd_vapp_org_network` fix potential error 'NAT rule cannot be configured for nics with
  DHCP addressing mode' during removal ([#489](https://github.com/terraform-providers/terraform-provider-vcd/issues/489))
* `resource/vcd_org_vdc` supports vCD 10.1 and resolves "no provider VDC found" errors ([#489](https://github.com/terraform-providers/terraform-provider-vcd/issues/489))

DEPRECATIONS:
* vCD 9.1 support is deprecated. Next version will require at least version 9.5 ([#489](https://github.com/terraform-providers/terraform-provider-vcd/issues/489))

NOTES:

* Bump terraform-plugin-sdk to v1.8.0 ([#479](https://github.com/terraform-providers/terraform-provider-vcd/issues/479))
* Update Travis to use Go 1.14 ([#479](https://github.com/terraform-providers/terraform-provider-vcd/issues/479))

## 2.7.0 (March 13, 2020)

FEATURES:

* **New Resource:** `vcd_vm_internal_disk` VM internal disk configuration ([#412](https://github.com/terraform-providers/terraform-provider-vcd/issues/412))
* **New Resource:** `vcd_vapp_org_network` vApp organization network ([#455](https://github.com/terraform-providers/terraform-provider-vcd/issues/455))
* **New Data Source:** `vcd_vapp_org_network` vApp org network ([#455](https://github.com/terraform-providers/terraform-provider-vcd/issues/455))
* **New Data Source:** `vcd_vapp_network` vApp network ([#455](https://github.com/terraform-providers/terraform-provider-vcd/issues/455))

IMPROVEMENTS:

* `vcd_vapp_network` supports isolated network and vApp network connected to Org VDC networks
  ([#455](https://github.com/terraform-providers/terraform-provider-vcd/issues/455),[#468](https://github.com/terraform-providers/terraform-provider-vcd/issues/468))
* `vcd_vapp_network` does not default `dns1` and `dns2` fields to 8.8.8.8 and 8.8.4.4 respectively
  ([#455](https://github.com/terraform-providers/terraform-provider-vcd/issues/455),[#468](https://github.com/terraform-providers/terraform-provider-vcd/issues/468))
* `vcd_org_vdc` can be created with Flex allocation model in vCD 9.7 and later. Also two new fields added
  for Flex - `elasticity`, `include_vm_memory_overhead` ([#443](https://github.com/terraform-providers/terraform-provider-vcd/issues/443))
* `resource/vcd_org` and `datasource/vcd_org` include a section `vapp_lease` and a section
  `vapp_template_lease` to define lease related parameters of depending entities - ([#432](https://github.com/terraform-providers/terraform-provider-vcd/issues/432))
* `resource/vcd_vapp_vm` Internal disks in VM template can be edited by `override_template_disk`
  field ([#412](https://github.com/terraform-providers/terraform-provider-vcd/issues/412))
* `vcd_vapp_vm` `disk` has new attribute `size_in_mb` ([#433](https://github.com/terraform-providers/terraform-provider-vcd/issues/433))
* `resource/vcd_vapp_vm` and `datasource/vcd_vapp_vm` get optional `network_dhcp_wait_seconds` field
  to ensure `ip` is reported when `ip_allocation_mode=DHCP` is used ([#436](https://github.com/terraform-providers/terraform-provider-vcd/issues/436))
* `resource/vcd_vapp_vm` and `datasource/vcd_vapp_vm` include a field `adapter_type` in `network`
  definition to specify NIC type - ([#441](https://github.com/terraform-providers/terraform-provider-vcd/issues/441))
* `resource/vcd_vapp_vm` and `datasource/vcd_vapp_vm` `customization` block supports all available
  features ([#462](https://github.com/terraform-providers/terraform-provider-vcd/issues/462), [#469](https://github.com/terraform-providers/terraform-provider-vcd/issues/469), [#477](https://github.com/terraform-providers/terraform-provider-vcd/issues/477))
* `datasource/*` - all data sources return an error when object is not found ([#446](https://github.com/terraform-providers/terraform-provider-vcd/issues/446), [#470](https://github.com/terraform-providers/terraform-provider-vcd/issues/470))
* `vcd_vapp_vm` allows to add routed vApp network, not only isolated one. `network.name` can reference 
  `vcd_vapp_network.name` of a vApp network with `org_network_name` set ([#472](https://github.com/terraform-providers/terraform-provider-vcd/issues/472))

DEPRECATIONS:
* `resource/vcd_vapp_vm` `network.name` deprecates automatic attachment of vApp Org network when
  `network.type=org` and it is not attached with `vcd_vapp_org_network` before referencing it.
  ([#455](https://github.com/terraform-providers/terraform-provider-vcd/issues/455))
* `resource/vcd_vapp_vm` field `initscript` is now deprecated in favor of
  `customization.0.initscript` to group all guest customization settings ([#462](https://github.com/terraform-providers/terraform-provider-vcd/issues/462))

BUG FIXES:

* `resource/vcd_vapp_vm` read - independent disks where losing `bus_number` and `unit_number`
  values after refresh ([#433](https://github.com/terraform-providers/terraform-provider-vcd/issues/433))
* `datasource/vcd_nsxv_dhcp_relay` crashes if no DHCP relay settings are present in Edge Gateway
  ([#446](https://github.com/terraform-providers/terraform-provider-vcd/issues/446))
* `resource/vcd_vapp_vm` `network` block changes caused MAC address changes in existing NICs
  ([#436](https://github.com/terraform-providers/terraform-provider-vcd/issues/436),[#407](https://github.com/terraform-providers/terraform-provider-vcd/issues/407))
* Fix a potential data race in client connection caching when VCD_CACHE is enabled ([#453](https://github.com/terraform-providers/terraform-provider-vcd/issues/453))
* `resource/vcd_vapp_vm` when `customization.0.force=false` crashes with `interface {} is nil` ([#462](https://github.com/terraform-providers/terraform-provider-vcd/issues/462))
* `resource/vcd_vapp_vm` `customization.0.force=true` could have skipped "Forced customization" on
  each apply ([#462](https://github.com/terraform-providers/terraform-provider-vcd/issues/462), [#477](https://github.com/terraform-providers/terraform-provider-vcd/issues/477))

NOTES:

* Drop support for vCD 9.0
* Bump terraform-plugin-sdk to v1.5.0 ([#442](https://github.com/terraform-providers/terraform-provider-vcd/issues/442))
* `make seqtestacc` and `make test-binary` use `-race` flags for `go test` to check if there are no
  data races. Additionally GNUMakefile supports `make installrace` and `make buildrace` to build
  binary with race detection enabled. ([#453](https://github.com/terraform-providers/terraform-provider-vcd/issues/453))
* Add `make test-upgrade-prepare` directive ([#462](https://github.com/terraform-providers/terraform-provider-vcd/issues/462))

## 2.6.0 (December 13, 2019)

FEATURES:

* **New Resource:** `vcd_nsxv_dhcp_relay` Edge gateway DHCP relay configuration - ([#416](https://github.com/terraform-providers/terraform-provider-vcd/issues/416))
* **New Resource:** `vcd_nsxv_ip_set` IP set - ([#406](https://github.com/terraform-providers/terraform-provider-vcd/issues/406),[#411](https://github.com/terraform-providers/terraform-provider-vcd/issues/411))
* **New Data Source:** `vcd_nsxv_dhcp_relay` Edge gateway DHCP relay configuration - ([#416](https://github.com/terraform-providers/terraform-provider-vcd/issues/416))
* **New Data Source:** `vcd_vapp_vm` VM - ([#218](https://github.com/terraform-providers/terraform-provider-vcd/issues/218))
* **New Data Source:** `vcd_nsxv_ip_set` IP set - ([#406](https://github.com/terraform-providers/terraform-provider-vcd/issues/406),[#411](https://github.com/terraform-providers/terraform-provider-vcd/issues/411))
* **New build command:** `make test-upgrade` to run an upgrade test from the previous released version

IMPROVEMENTS:

* Switch to Terraform terraform-plugin-sdk v1.3.0 as per recent [HashiCorp
  recommendation](https://www.terraform.io/docs/extend/plugin-sdk.html) - ([#382](https://github.com/terraform-providers/terraform-provider-vcd/issues/382), [#406](https://github.com/terraform-providers/terraform-provider-vcd/issues/406))
* `resource/vcd_vapp_vm` VM state ID changed from VM name to vCD ID
* `resource/vcd_vapp_vm` Add properties `description` and `storage_profile`
* `resource/vcd_vapp_vm` Add import capability and full read support ([#218](https://github.com/terraform-providers/terraform-provider-vcd/issues/218))

* `resource/vcd_nsxv_dnat` and `resource/vcd_nsxv_dnat` more precise error message when network is
  not found - ([#384](https://github.com/terraform-providers/terraform-provider-vcd/issues/384))
* `resource/vcd_nsxv_dnat` and `resource/vcd_nsxv_dnat` `rule_tag` must be int to avoid vCD internal
  exception passthrough - ([#384](https://github.com/terraform-providers/terraform-provider-vcd/issues/384))
* `resource/vcd_nsxv_dnat` put correct name in doc example - ([#384](https://github.com/terraform-providers/terraform-provider-vcd/issues/384))
* `resource/vcd_nsxv_dnat` and `resource/vcd_nsxv_dnat` avoid rule replacement because of changed
  `rule_tag` when rule is altered via UI - ([#384](https://github.com/terraform-providers/terraform-provider-vcd/issues/384))
* `resource/vcd_nsxv_firewall_rule` add explicit protocol validation to avoid odd NSX-V API error -
  ([#384](https://github.com/terraform-providers/terraform-provider-vcd/issues/384))
* `resource/vcd_nsxv_firewall_rule` `rule_tag` must be int to avoid vCD internal exception
  passthrough - ([#384](https://github.com/terraform-providers/terraform-provider-vcd/issues/384))
* Fix code warnings from `staticcheck` and add command `make static` to Travis tests
* `resource/vcd_edge_gateway` and `datasource/vcd_edge_gateway` add `default_external_network_ip`
  and `external_network_ips` fields to export default edge gateway IP address and other external
  network IPs used on gateway interfaces - ([#389](https://github.com/terraform-providers/terraform-provider-vcd/issues/389), [#401](https://github.com/terraform-providers/terraform-provider-vcd/issues/401))
* Add `token` to the `vcd` provider for the ability of connecting with an authorization token - ([#280](https://github.com/terraform-providers/terraform-provider-vcd/issues/280))
* Add command `make token` to create an authorization token from testing credentials
* Clean up interpolation-only expressions from tests (as allowed in terraform v0.12.11+)
* Increment vCD API version used from 27.0 to 29.0 ([#396](https://github.com/terraform-providers/terraform-provider-vcd/issues/396))
* `resource/vcd_network_routed` Add properties `description` and `interface_type` ([#321](https://github.com/terraform-providers/terraform-provider-vcd/issues/321),[#342](https://github.com/terraform-providers/terraform-provider-vcd/issues/342),[#374](https://github.com/terraform-providers/terraform-provider-vcd/issues/374))
* `resource/vcd_network_isolated` Add property `description` ([#373](https://github.com/terraform-providers/terraform-provider-vcd/issues/373))
* `resource/vcd_network_direct` Add property `description`
* `resource/vcd_network_routed` Add check for valid IPs ([#374](https://github.com/terraform-providers/terraform-provider-vcd/issues/374))
* `resource/vcd_network_isolated` Add check for valid IPs ([#373](https://github.com/terraform-providers/terraform-provider-vcd/issues/373))
* `resource/vcd_nsxv_firewall_rule` Add support for IP sets ([#411](https://github.com/terraform-providers/terraform-provider-vcd/issues/411))
* `resource/vcd_edgegateway` new fields `fips_mode_enabled`, `use_default_route_for_dns_relay`
  - ([#401](https://github.com/terraform-providers/terraform-provider-vcd/issues/401),[#414](https://github.com/terraform-providers/terraform-provider-vcd/issues/414))
* `resource/vcd_edgegateway`  new `external_network` block for advanced configurations of external
  networks including multiple subnets, IP pool sub-allocation and rate limits - ([#401](https://github.com/terraform-providers/terraform-provider-vcd/issues/401),[#418](https://github.com/terraform-providers/terraform-provider-vcd/issues/418))
* `resource/vcd_edgegateway` enables read support for field `distributed_routing` after switch to
  vCD API v29.0 - ([#401](https://github.com/terraform-providers/terraform-provider-vcd/issues/401))
* `vcd_nsxv_firewall_rule` - improve internal lookup mechanism for `gateway_interfaces` field in
  source and/or destination ([#419](https://github.com/terraform-providers/terraform-provider-vcd/issues/419))

BUG FIXES:

* Fix `vcd_org_vdc` datasource read. When user was Organization administrator datasource failed. Fields provider_vdc_name, storage_profile, memory_guaranteed, cpu_guaranteed, cpu_speed, enable_thin_provisioning, enable_fast_provisioning, network_pool_name won't have values for org admin.
* Removed `power_on` property from data source `vcd_vapp`, as it is a directive used during vApp build.
  Its state is never updated and the fields `status` and `status_text` already provide the necessary information.
  ([#379](https://github.com/terraform-providers/terraform-provider-vcd/issues/379))
* Fix `vcd_independent_disk` reapply issue, which was seen when optional `bus_sub_type` and `bus_type` wasn't used - ([#394](https://github.com/terraform-providers/terraform-provider-vcd/issues/394))
* Fix `vcd_vapp_network` apply issue, where the property `guest_vlan_allowed` was applied only to the last of multiple networks.
* `datasource/vcd_network_direct` is now readable by Org User (previously it was only by Sys Admin), as this change made it possible to get the details of External Network as Org User ([#408](https://github.com/terraform-providers/terraform-provider-vcd/issues/408))

DEPRECATIONS:

* Deprecated property `storage_profile` in resource `vcd_vapp`, as the corresponding field is now enabled in `vcd_vapp_vm`
* `resource/vcd_edgegateway` deprecates fields `external_networks` and `default_gateway_network` in
  favor of new `external_network` block(s) - [[#401](https://github.com/terraform-providers/terraform-provider-vcd/issues/401)] 

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
