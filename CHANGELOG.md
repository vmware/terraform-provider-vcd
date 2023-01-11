## 3.8.2 (January 12th, 2023)

### IMPROVEMENTS
* Add `catalog_id` to resource and data source `vcd_catalog_media` to allow handling similarly to `vcd_catalog_vapp_template` ([#972](https://github.com/vmware/terraform-provider-vcd/pull/972))

### BUG FIXES
* Change `vcd_catalog`, `vcd_catalog_media`, `vcd_catalog_vapp_template`, and `vcd_catalog_item` to access their entities without the need to use a full Org object, thus allowing the access to shared catalogs from other organizations (Issue #960) ([#972](https://github.com/vmware/terraform-provider-vcd/pull/972))
* Fix a bug that caused inconsistent plan when using `group_id` in `vcd_catalog_access_control`,
  `vcd_org_vdc_access_control` and `vcd_vapp_access_control` resources ([#963](https://github.com/vmware/terraform-provider-vcd/pull/963))
* Remove unnecessary URL checks from `vcd_subscribed_catalog` creation, to allow subscribing to non-VCD entities, such as vSphere shared library ([#972](https://github.com/vmware/terraform-provider-vcd/pull/972))
* Remove unnecessary validation that prevents attaching NSX-T Org network to vApp using
  `org_network_name` field in `vcd_vapp_network` resource ([#975](https://github.com/vmware/terraform-provider-vcd/pull/975))

### DEPRECATIONS
* Deprecate usage of `catalog` in favor of `catalog_id` in `vcd_catalog_media` ([#972](https://github.com/vmware/terraform-provider-vcd/pull/972))

### NOTES
* Add mini-framework for running tests with several Organizations ([#972](https://github.com/vmware/terraform-provider-vcd/pull/972))
* Try to amend quirky test `TestAccVcdNsxtDynamicSecurityGroupVdcGroupCriteriaWithVm` that sometimes fails due to a bad filter.
  It now uses a shorter name for the Dynamic Security Groups to try to not break the resulting filter chain ([#980](https://github.com/vmware/terraform-provider-vcd/pull/980))

## 3.8.1 (December 14, 2022)

### IMPROVEMENTS
* Add `vdc_id` to data source `vcd_vm_placement_policy` to allow tenant users to fetch VM Placement Policies from
  the ones assigned to VDCs ([#948](https://github.com/vmware/terraform-provider-vcd/pull/948))
* Resource and data source `vcd_catalog` and `vcd_subscribed_catalog` introduce new computed field `is_local` to specify
  whether the catalog originated from the current org ([#949](https://github.com/vmware/terraform-provider-vcd/pull/949))
* Improve usage of `org` field in `vcd_catalog` to accept sharing Org name for shared catalogs and improve error messages ([#949](https://github.com/vmware/terraform-provider-vcd/pull/949))

### BUG FIXES
* Fix a bug that caused `vcd_vm_group` data source to fail when the backing Provider VDC had multiple resource pools ([#948](https://github.com/vmware/terraform-provider-vcd/pull/948))
* Fix issue #944 - shared catalog datasource not accessible to Org users ([#949](https://github.com/vmware/terraform-provider-vcd/pull/949))
* Fix issue #672 - update Org with invalid or extended LDAP settings ([#952](https://github.com/vmware/terraform-provider-vcd/pull/952), [#955](https://github.com/vmware/terraform-provider-vcd/pull/955))

## 3.8.0 (November 25, 2022)

### FEATURES
* **New Resource:** `vcd_catalog_vapp_template` to manage the upload and usage of vApp Templates ([#899](https://github.com/vmware/terraform-provider-vcd/pull/899))
* **New Data Source:** `vcd_catalog_vapp_template` to fetch existing vApp Templates ([#899](https://github.com/vmware/terraform-provider-vcd/pull/899))
* **New Resource:** `vcd_vm_placement_policy` that allows creating VM Placement Policies ([#904](https://github.com/vmware/terraform-provider-vcd/pull/904), [#911](https://github.com/vmware/terraform-provider-vcd/pull/911))
* **New Data Source:** `vcd_vm_placement_policy` that allows fetching existing VM Placement Policies ([#904](https://github.com/vmware/terraform-provider-vcd/pull/904), [#911](https://github.com/vmware/terraform-provider-vcd/pull/911))
* **New Data Source:** `vcd_provider_vdc` that allows fetching existing Provider VDCs ([#904](https://github.com/vmware/terraform-provider-vcd/pull/904))
* **New Data Source:** `vcd_vm_group` that allows fetching existing VM Groups, to be able to create VM Placement Policies ([#904](https://github.com/vmware/terraform-provider-vcd/pull/904))
* **New Resource**: `vcd_org_ldap` that allows configuring LDAP settings for an organization ([#909](https://github.com/vmware/terraform-provider-vcd/pull/909))
* **New Data Source**: `vcd_org_ldap` that allows exploring LDAP settings for an organization ([#909](https://github.com/vmware/terraform-provider-vcd/pull/909))
* **New Resource:** `vcd_catalog_access_control` that allows sharing a catalog with users, groups, or Orgs ([#915](https://github.com/vmware/terraform-provider-vcd/pull/915))
* **New Resource**: `vcd_subscribed_catalog` that allows subscribing to a published catalog ([#916](https://github.com/vmware/terraform-provider-vcd/pull/916))
* **New Data Source**: `vcd_subscribed_catalog` that allows reading a subscribed catalog ([#916](https://github.com/vmware/terraform-provider-vcd/pull/916))
* **New Data Source**: `vcd_task` that allows reading a VCD task ([#916](https://github.com/vmware/terraform-provider-vcd/pull/916))

### IMPROVEMENTS
* Add attribute `metadata_entry` to the following data sources:
  `vcd_catalog`, `vcd_catalog_media`, `vcd_independent_disk`, `vcd_network_direct`, `vcd_network_isolated`,
  `vcd_network_isolated_v2`, `vcd_network_routed`, `vcd_network_routed_v2`, `vcd_org`, `vcd_org_vdc`, `vcd_provider_vdc`,
  `vcd_storage_profile`, `vcd_vapp`, `vcd_vapp_vm`. This new attribute replaces `metadata`
  to add support of metadata visibility (user access levels), all the available types and domains for every metadata
  entry ([#917](https://github.com/vmware/terraform-provider-vcd/pull/917))
* Add attribute `metadata_entry` to the following resources:
  `vcd_catalog`, `vcd_catalog_media`, `vcd_independent_disk`, `vcd_network_direct`, `vcd_network_isolated`,
  `vcd_network_isolated_v2`, `vcd_network_routed`, `vcd_network_routed_v2`, `vcd_org`, `vcd_org_vdc`, `vcd_vapp`,
  `vcd_vapp_vm`. This new attribute replaces `metadata` to add support of metadata visibility (user access levels),
  all the available types and domains for every metadata entry ([#917](https://github.com/vmware/terraform-provider-vcd/pull/917))
* Add `placement_policy_id` attribute to `vcd_vapp_vm` and `vcd_vm` resource and data source,
  to support the usage of VM Placement Policies in VMs ([#922](https://github.com/vmware/terraform-provider-vcd/pull/922))
* Resources and data sources `vcd_vapp_vm` and `vcd_vm` have new computed fields `status` and
  `status_text` ([#901](https://github.com/vmware/terraform-provider-vcd/pull/901))
* Add `vm_placement_policy_ids` attribute to `vcd_org_vdc` resource and data source to assign existing
  VM Placement Policies to VDCs ([#904](https://github.com/vmware/terraform-provider-vcd/pull/904), [#911](https://github.com/vmware/terraform-provider-vcd/pull/911))
* Add `default_compute_policy_id` attribute to `vcd_org_vdc` resource and data source to specify a default
  VM Sizing Policy, VM Placement Policy or vGPU Policy for the VDC ([#904](https://github.com/vmware/terraform-provider-vcd/pull/904), [#911](https://github.com/vmware/terraform-provider-vcd/pull/911))
* Add attributes `href`, `vapp_template_list`, `media_item_list`, and `publishing_url` to `vcd_catalog` resource and data source to show published items ([#916](https://github.com/vmware/terraform-provider-vcd/pull/916))
* Add `subscribed_catalog` to examples ([#916](https://github.com/vmware/terraform-provider-vcd/pull/916))
* Upgrade Terraform SDK dependency to v2.24.1 ([#920](https://github.com/vmware/terraform-provider-vcd/pull/920), [#930](https://github.com/vmware/terraform-provider-vcd/pull/930))
* Resource and data source `vcd_org_vdc` introduce new field `edge_cluster_id` to specify NSX-T
  Edge Cluster for VDC ([#921](https://github.com/vmware/terraform-provider-vcd/pull/921))
* Resources, that are removed outside of Terraform control are removed from state and recreated
  instead of returning error ([#925](https://github.com/vmware/terraform-provider-vcd/pull/925))
  * `vcd_edgegateway_settings`
  * `vcd_vapp_network`
  * `vcd_vm_internal_disk`
  * `vcd_nsxv_dhcp_relay`
  * `vcd_vapp_static_routing`
  * `vcd_vapp_nat_rules`
  * `vcd_vapp_firewall_rules`
  * `vcd_vapp_access_control`
  * `vcd_nsxt_alb_edgegateway_service_engine_group`
  * `vcd_org_vdc`
  * `vcd_org_user`
  * `vcd_external_network`
* Resource and data source `vcd_nsxt_network_dhcp` support Isolated networks, different DHCP modes
  ('EDGE', 'NETWORK', 'RELAY') and lease time ([#929](https://github.com/vmware/terraform-provider-vcd/pull/929))
* Add the new attributes `vapp_template_id`, `boot_image_id` to the resources `vcd_vapp_vm` and `vcd_vm` to be able
  to use unique URNs to reference vApp Templates and Media items through data sources, to build strong dependencies
  in Terraform configuration ([#931](https://github.com/vmware/terraform-provider-vcd/pull/931))
* Data source `vcd_nsxt_edge_cluster` supports NSX-T Edge Cluster filtering by `vdc_id`, `vdc_group_id`, 
  and `provider_vdc_id` ([#921](https://github.com/vmware/terraform-provider-vcd/pull/921))

### BUG FIXES
* Fix bug where VM was power cycled multiple times during creation ([#901](https://github.com/vmware/terraform-provider-vcd/pull/901))
* Fix bug where storage_profile is ignored for empty (non template) VM ([#901](https://github.com/vmware/terraform-provider-vcd/pull/901))
* `resource/vcd_nsxt_alb_edgegateway_service_engine_group` field `reserved_virtual_services` accepts
  "0" as value ([#923](https://github.com/vmware/terraform-provider-vcd/pull/923))
* Fix a bug in `resource/vcd_vapp` that would prevent to Power off vApp when previous state was
  `power_on=true` ([#932](https://github.com/vmware/terraform-provider-vcd/pull/932))

### DEPRECATIONS
* Deprecate `vcd_external_network` in favor of `vcd_external_network_v2` ([#903](https://github.com/vmware/terraform-provider-vcd/pull/903))
* Deprecate `default_vm_sizing_policy_id` field in `vcd_org_vdc` resource and data source. This field is misleading as it
  can contain not only VM Sizing Policies but also VM Placement Policies or vGPU Policies.
  Its replacement is the `default_compute_policy_id` attribute ([#904](https://github.com/vmware/terraform-provider-vcd/pull/904))
* Deprecate attribute `metadata` in favor of `metadata_entry` in the following data sources:
  `vcd_catalog`, `vcd_catalog_media`, `vcd_catalog_vapp_template`, `vcd_independent_disk`, `vcd_network_direct`,
  `vcd_network_isolated`, `vcd_network_isolated_v2`, `vcd_network_routed`, `vcd_network_routed_v2`, `vcd_org`,
  `vcd_org_vdc`, `vcd_provider_vdc`, `vcd_storage_profile`, `vcd_vapp`, `vcd_vapp_vm` ([#917](https://github.com/vmware/terraform-provider-vcd/pull/917))
* Deprecate attribute `metadata` in favor of `metadata_entry` in the following resources:
  `vcd_catalog`, `vcd_catalog_media`, `vcd_catalog_vapp_template`, `vcd_independent_disk`, `vcd_network_direct`,
  `vcd_network_isolated`, `vcd_network_isolated_v2`, `vcd_network_routed`, `vcd_network_routed_v2`, `vcd_org`,
  `vcd_org_vdc`, `vcd_vapp`, `vcd_vapp_vm` ([#917](https://github.com/vmware/terraform-provider-vcd/pull/917))
* Deprecate attribute `catalog_item_metadata` in favor of `metadata_entry` in the `vcd_catalog_item` resource
  and data source. ([#917](https://github.com/vmware/terraform-provider-vcd/pull/917))
* Deprecate `template_name` in favor of `vapp_template_id` in `vcd_vapp_vm` and `vcd_vm` to be able to use unique URNs instead
  of catalog dependent names ([#931](https://github.com/vmware/terraform-provider-vcd/pull/931))
* Deprecate `boot_image` in favor of `boot_image_id` in `vcd_vapp_vm` and `vcd_vm` to be able to use URNs instead
  of catalog dependent names ([#931](https://github.com/vmware/terraform-provider-vcd/pull/931))
* Deprecate `catalog_name` in favor of `vapp_template_id` or `boot_image_id`, which don't require a catalog name anymore ([#931](https://github.com/vmware/terraform-provider-vcd/pull/931))
* Data source `vcd_nsxt_edge_cluster` deprecates `vdc` field in favor of three new fields to define
  NSX-T Edge Cluster lookup scope - `vdc_id`, `vdc_group_id`, and `provider_vdc_id` ([#921](https://github.com/vmware/terraform-provider-vcd/pull/921))

### NOTES
* Drop support for EOL VCD 10.2.x ([#903](https://github.com/vmware/terraform-provider-vcd/pull/903))
* Add a guide and examples on [Catalog subscribing and sharing](https://registry.terraform.io/providers/vmware/vcd/latest/docs/guides/catalog_subscription_and_sharing) to the documentation ([#916](https://github.com/vmware/terraform-provider-vcd/pull/916))
* All non-NSX-V resources and data sources use the new SDK signatures with Context and Diagnostics ([#895](https://github.com/vmware/terraform-provider-vcd/pull/895))
* Refactor VM Creation code, which should result in identifiable parts creation for all types of
  VMs. Behind the scenes, there are 4 different types of VMs with respective different API calls as
  listed below ([#901](https://github.com/vmware/terraform-provider-vcd/pull/901))
  * `vcd_vapp_vm` built from vApp template
  * `vcd_vm` built from vApp template
  * `vcd_vapp_vm` built without vApp template (empty VM)
  * `vcd_vm` built without vApp template (empty VM)
* Bump Go to 1.19 in `go.mod` as the minimum required version. ([#902](https://github.com/vmware/terraform-provider-vcd/pull/902), [#916](https://github.com/vmware/terraform-provider-vcd/pull/916))
* Code documentation formatting is adjusted using Go 1.19 (`make fmt`) ([#902](https://github.com/vmware/terraform-provider-vcd/pull/902))
* Adjust GitHub actions in pipeline to use the latest code ([#902](https://github.com/vmware/terraform-provider-vcd/pull/902))
* `staticcheck` switched version naming from `2021.1.2` to `v0.3.3` in downloads section. This PR
  also updates the code to fetch correct staticcheck ([#902](https://github.com/vmware/terraform-provider-vcd/pull/902))
* package `io/ioutil` is deprecated as of Go 1.16. `staticcheck` started complaining about usage of
  deprecated packages. As a result this PR switches packages to either `io` or `os` (still the same
  functions are used) ([#902](https://github.com/vmware/terraform-provider-vcd/pull/902))
* Add a new GitHub Action to run `gosec` on every push and pull request ([#928](https://github.com/vmware/terraform-provider-vcd/pull/928))


## 3.7.0 (August 2, 2022)

### FEATURES
* Add a guide and examples to configure and install **Container Service Extension (CSE)** 3.1.x using Terraform ([#856](https://github.com/vmware/terraform-provider-vcd/pull/856), [#882](https://github.com/vmware/terraform-provider-vcd/pull/882), [#876](https://github.com/vmware/terraform-provider-vcd/pull/876))
* **New Resource:** `vcd_security_tag` that allows creating security tags ([#845](https://github.com/vmware/terraform-provider-vcd/pull/845))
* **New Resource:** `vcd_org_vdc_access_control` to manage VDC access control ([#850](https://github.com/vmware/terraform-provider-vcd/pull/850))
* **New Resource:** `vcd_nsxt_route_advertisement` that allows NSX-T Edge Gateway to advertise subnets to Tier-0 Gateway ([#858](https://github.com/vmware/terraform-provider-vcd/pull/858), [#888](https://github.com/vmware/terraform-provider-vcd/pull/888))
* **New Data Source:** `vcd_nsxt_route_advertisement` that reads the NSX-T Edge Gateway routes that are being advertised ([#858](https://github.com/vmware/terraform-provider-vcd/pull/858), [#888](https://github.com/vmware/terraform-provider-vcd/pull/888))
* **New Resource:** `vcd_nsxt_edgegateway_bgp_configuration` for NSX-T Edge Gateway BGP
  Configuration ([#798](https://github.com/vmware/terraform-provider-vcd/pull/798), [#887](https://github.com/vmware/terraform-provider-vcd/pull/887))
* **New Data Source:** `vcd_nsxt_edgegateway_bgp_configuration` for reading NSX-T Edge Gateway BGP
  Configuration ([#798](https://github.com/vmware/terraform-provider-vcd/pull/798), [#887](https://github.com/vmware/terraform-provider-vcd/pull/887))
* **New Resource:** `vcd_nsxt_dynamic_security_group` to manage dynamic security groups ([#877](https://github.com/vmware/terraform-provider-vcd/pull/877))
* **New Data Source:** `vcd_nsxt_dynamic_security_group` to lookup existing dynamic security groups
  ([#877](https://github.com/vmware/terraform-provider-vcd/pull/877))
* **New Resource:** `vcd_nsxt_edgegateway_bgp_ip_prefix_list` allows users to configure NSX-T Edge Gateway BGP IP Prefix Lists ([#879](https://github.com/vmware/terraform-provider-vcd/pull/879), [#887](https://github.com/vmware/terraform-provider-vcd/pull/887), [#888](https://github.com/vmware/terraform-provider-vcd/pull/888))
* **New Data Source:** `vcd_nsxt_edgegateway_bgp_ip_prefix_list` allows users to read NSX-T Edge Gateway BGP IP Prefix Lists ([#879](https://github.com/vmware/terraform-provider-vcd/pull/879), [#887](https://github.com/vmware/terraform-provider-vcd/pull/887), [#888](https://github.com/vmware/terraform-provider-vcd/pull/888))
* **New Resource:** `vcd_nsxt_edgegateway_bgp_neighbor` allows users to configure NSX-T Edge Gateway BGP Neighbors ([#879](https://github.com/vmware/terraform-provider-vcd/pull/879), [#887](https://github.com/vmware/terraform-provider-vcd/pull/887), [#888](https://github.com/vmware/terraform-provider-vcd/pull/888))
* **New Data Source:** `vcd_nsxt_edgegateway_bgp_neighbor` allows users to read NSX-T Edge Gateway BGP Neighbors ([#879](https://github.com/vmware/terraform-provider-vcd/pull/879), [#887](https://github.com/vmware/terraform-provider-vcd/pull/887), [#888](https://github.com/vmware/terraform-provider-vcd/pull/888))

### IMPROVEMENTS
* `resource/vcd_nsxt_network_dhcp` and `datasource/vcd_nsxt_firewall` now support `dns_servers` ([#830](https://github.com/vmware/terraform-provider-vcd/pull/830))
* Add VDC Group compatibility for `resource/vcd_nsxt_firewall` and `datasource/vcd_nsxt_firewall` ([#841](https://github.com/vmware/terraform-provider-vcd/pull/841), [#888](https://github.com/vmware/terraform-provider-vcd/pull/888))
* Add VDC Group compatibility for `resource/vcd_nsxt_nat_rule` and `datasource/vcd_nsxt_nat_rule` ([#841](https://github.com/vmware/terraform-provider-vcd/pull/841), [#888](https://github.com/vmware/terraform-provider-vcd/pull/888))
* Add VDC Group compatibility for `resource/vcd_nsxt_ipsec_vpn_tunnel` and `datasource/vcd_nsxt_ipsec_vpn_tunnel` ([#841](https://github.com/vmware/terraform-provider-vcd/pull/841), [#888](https://github.com/vmware/terraform-provider-vcd/pull/888))
* Add VDC Group compatibility for `resource/vcd_nsxt_alb_settings` and `datasource/vcd_nsxt_alb_settings` ([#841](https://github.com/vmware/terraform-provider-vcd/pull/841))
* Add VDC Group compatibility for `resource/vcd_nsxt_alb_edgegateway_service_engine_group` and `datasource/vcd_nsxt_alb_edgegateway_service_engine_group` ([#841](https://github.com/vmware/terraform-provider-vcd/pull/841), [#854](https://github.com/vmware/terraform-provider-vcd/pull/854))
* Add VDC Group compatibility for `resource/vcd_nsxt_alb_virtual_service` and `datasource/vcd_nsxt_alb_virtual_service` ([#841](https://github.com/vmware/terraform-provider-vcd/pull/841))
* Add VDC Group compatibility for `resource/vcd_nsxt_alb_pool` and `datasource/vcd_nsxt_alb_pool` ([#841](https://github.com/vmware/terraform-provider-vcd/pull/841))
* `resource/vcd_vm_sizing_policy`: remove (deprecate) unneeded `org` property ([#843](https://github.com/vmware/terraform-provider-vcd/pull/843))
* `datasource/vcd_vm_sizing_policy`: remove (deprecate) unneeded `org` property ([#843](https://github.com/vmware/terraform-provider-vcd/pull/843))
* Add changes to allow running tests on CDS and make NSX-V configuration optional ([#848](https://github.com/vmware/terraform-provider-vcd/pull/848))
* `resource/vcd_catalog_item` and `datasource/vcd_catalog_item` now support metadata for the CatalogItem entity with `catalog_item_metadata` attribute ([#851](https://github.com/vmware/terraform-provider-vcd/pull/851))
* `metadata` attribute on every compatible resource and data source is now more performant when adding and updating metadata ([#853](https://github.com/vmware/terraform-provider-vcd/pull/853))
* Make `license_type` attribute on **vcd_nsxt_alb_controller** optional as it is not used from VCD v10.4 onwards ([#878](https://github.com/vmware/terraform-provider-vcd/pull/878))
* Add `supported_feature_set` to **vcd_nsxt_alb_service_engine_group** resource and data source to be compatible with VCD v10.4, which replaces the **vcd_nsxt_alb_controller** `license_type` ([#878](https://github.com/vmware/terraform-provider-vcd/pull/878))
* Add `supported_feature_set` to **vcd_nsxt_alb_settings** resource and data source to be compatible with VCD v10.4, which replaces the **vcd_nsxt_alb_controller** `license_type` ([#878](https://github.com/vmware/terraform-provider-vcd/pull/878))

### BUG FIXES
* Fix typo in documentation for `resource/vcd_vapp_vm` to fix broken attribute rendering ([#828](https://github.com/vmware/terraform-provider-vcd/pull/828))
* Skip binary and upgrade tests for NVMe in VCD < 10.2.2 ([#838](https://github.com/vmware/terraform-provider-vcd/pull/838))
* Network lookup could return incorrect ID for `vcd_vapp_network` and `vcd_vapp_org_network` ([#838](https://github.com/vmware/terraform-provider-vcd/pull/838))
* Set missing `org` and `vdc` fields during import of `vcd_independent_disk` ([#838](https://github.com/vmware/terraform-provider-vcd/pull/838))
* Add missing `None` mode in independent disk `sharing_type` ([#849](https://github.com/vmware/terraform-provider-vcd/pull/849))
* `vcd_vm_internal_disk` now uses default IOPS value from storage profile when custom IOPS value isn't provided ([#863](https://github.com/vmware/terraform-provider-vcd/pull/863))
* Fix `vcd_inserted_media` locking mechanism to avoid race condition with `vcd_vm_internal_disk` ([#870](https://github.com/vmware/terraform-provider-vcd/pull/870), [#888](https://github.com/vmware/terraform-provider-vcd/pull/888))
* Fix a bug that causes `vcd_vapp_vm` to fail on creation if attribute `sizing_policy_id` is set and corresponds to a
Sizing Policy with CPU or memory defined, `template_name` is used and `power_on` is `true` ([#883](https://github.com/vmware/terraform-provider-vcd/pull/883))

### DEPRECATIONS
* Deprecate `vdc` field in NSX-T Edge Gateway child entities. This field is no longer precise as
  with introduction of VDC Group support an Edge Gateway can be bound either to a VDC, either to a
  VDC Group. Parent VDC or VDC Group is now inherited from `edge_gateway_id` field. Impacted
  resources and data sources are: `vcd_nsxt_firewall`, `vcd_nsxt_nat_rule`,
  `vcd_nsxt_ipsec_vpn_tunnel`, `,vcd_nsxt_alb_settings`,
  `vcd_nsxt_alb_edgegateway_service_engine_group`, `vcd_nsxt_alb_virtual_service`,
  `vcd_nsxt_alb_pool` ([#841](https://github.com/vmware/terraform-provider-vcd/pull/841))
* `resource/vcd_nsxt_network_dhcp` and `datasource/vcd_nsxt_network_dhcp` deprecate `vdc` field to
  make consumption more friendly with VDC Groups ([#846](https://github.com/vmware/terraform-provider-vcd/pull/846))

### NOTES
* Apply `gofmt -s -w .` to cleanup code ([#833](https://github.com/vmware/terraform-provider-vcd/pull/833))
* Adjust `role` data source and resource documentation for `rights` attribute to reflect its *Set* nature ([#834](https://github.com/vmware/terraform-provider-vcd/pull/834))
* Adjust `rights_bundle` data source and resource documentation for `rights` and `tenants` attribute to reflect its *Set* nature ([#834](https://github.com/vmware/terraform-provider-vcd/pull/834))
* Add an example about using CloudInit to Guest Customization guides page ([#852](https://github.com/vmware/terraform-provider-vcd/pull/852))
* Upgrade Terraform SDK dependency to v2.17.0 [#844, #853]
* Testing infrastructure: make NSX-T VDC primary for tests instead of NSX-V one (required for CDS
  certification) ([#886](https://github.com/vmware/terraform-provider-vcd/pull/886))
* Add Cloud Director Service (CDS) as supported ([#890](https://github.com/vmware/terraform-provider-vcd/pull/890))


## 3.6.0 (April 14, 2022)

### FEATURES
* **New Data Source:** `vcd_org_group` allows to fetch an Organization Group to use it with other resources ([#798](https://github.com/vmware/terraform-provider-vcd/pull/798))
* Add `catalog_version`, `number_of_vapp_templates`, `number_of_media`, `is_shared`, `is_published`, `publish_subscription_type` computed fields to catalog resource and datasource  ([#800](https://github.com/vmware/terraform-provider-vcd/pull/800))
* Update `vcd_catalog` datasource so now it can take org from provider level or datasource level like modern resources/datasources ([#800](https://github.com/vmware/terraform-provider-vcd/pull/800))
* **New Resource:** `vcd_nsxt_distributed_firewall` manages Distributed Firewall Rules in VDC Groups
  ([#810](https://github.com/vmware/terraform-provider-vcd/pull/810))
* **New Data Source:** `vcd_nsxt_distributed_firewall` provides lookup functionality for Distributed
  Firewall rules in VDC Groups ([#810](https://github.com/vmware/terraform-provider-vcd/pull/810))
* **New Data Source:** `vcd_nsxt_network_context_profile` allows user to lookup Network Context
  Profile ID for usage in `vcd_nsxt_distributed_firewall` ([#810](https://github.com/vmware/terraform-provider-vcd/pull/810))


### IMPROVEMENTS
* `resource/vcd_catalog` add support for `metadata` so this field can be set when creating/updating catalogs. ([#780](https://github.com/vmware/terraform-provider-vcd/pull/780))
* `datasource/vcd_catalog` add support for `metadata` so this field can be retrieved when reading from catalogs. ([#780](https://github.com/vmware/terraform-provider-vcd/pull/780))
* `data/vcd_storage_profile` add IOPS settings properties `iops_limiting_enabled`, `maximum_disk_iops`, `default_disk_iops`, `disk_iops_per_gb_max`, `iops_limit` and other data attributes `limit`, `units`, `used_storage`, `default`, `enabled`, `iops_allocated` to the state ([#782](https://github.com/vmware/terraform-provider-vcd/pull/782))
* `resource/vcd_nsxt_edgegateway` and `datasource/vcd_nsxt_edgegateway` support VDC Groups via new
  field `owner_id` replacing `vdc` ([#793](https://github.com/vmware/terraform-provider-vcd/pull/793))
* Update codebase to be compatible with changes in go-vcloud-director due to bump to VCD API V35.0 ([#795](https://github.com/vmware/terraform-provider-vcd/pull/795))
* `vcd_org` resource adds support for `metadata` so this field can be set when creating/updating organizations. ([#797](https://github.com/vmware/terraform-provider-vcd/pull/797))
* `vcd_org` data source adds support for `metadata` so this field can be retrieved when reading from organizations. ([#797](https://github.com/vmware/terraform-provider-vcd/pull/797))
* `vcd_independent_disk` resource adds support for `metadata` so this field can be set when creating/updating independent disks. ([#797](https://github.com/vmware/terraform-provider-vcd/pull/797))
* `vcd_independent_disk` data source adds support for `metadata` so this field can be retrieved when reading from independent disks. ([#797](https://github.com/vmware/terraform-provider-vcd/pull/797))
* `vcd_org_user` resource and data source have now `is_external` attribute to support the importing of LDAP users into the Organization ([#798](https://github.com/vmware/terraform-provider-vcd/pull/798))
* `vcd_org_user` resource does not have a default value for `deployed_vm_quota` and `stored_vm_quota`. Local users will have unlimited quota by default, imported from LDAP will have no quota ([#798](https://github.com/vmware/terraform-provider-vcd/pull/798))
* `vcd_org_user` resource and data source have now `group_names` attribute to list group names if the user comes from an LDAP group ([#798](https://github.com/vmware/terraform-provider-vcd/pull/798))
* `vcd_org_group` resource and data source have now `user_names` attribute to list user names if the user was imported from LDAP ([#798](https://github.com/vmware/terraform-provider-vcd/pull/798))
* `resource/vcd_network_routed_v2` and `datasource/vcd_network_routed_v2` support VDC Groups by
  inheriting parent VDC or VDC Group from Edge Gateway  ([#801](https://github.com/vmware/terraform-provider-vcd/pull/801))
* `resource/vcd_network_isolated_v2` and `datasource/vcd_network_isolated_v2` support VDC Groups via
  new field `owner_id` replacing `vdc` ([#801](https://github.com/vmware/terraform-provider-vcd/pull/801))
* `resource/vcd_nsxt_network_imported` and `datasource/vcd_nsxt_network_imported` support VDC Groups
  via new field `owner_id` replacing `vdc`  ([#801](https://github.com/vmware/terraform-provider-vcd/pull/801))
* Add support for `can_publish_external_catalogs` and `can_subscribe_external_catalogs` in `datasource_vcd_org` and `resource_vcd_org` ([#803](https://github.com/vmware/terraform-provider-vcd/pull/803))
* `resource/vcd_network_direct` add support for `metadata` so this field can be set when creating/updating direct networks. ([#804](https://github.com/vmware/terraform-provider-vcd/pull/804))
* `datasource/vcd_network_direct` add support for `metadata` so this field can be retrieved when reading from direct networks. ([#804](https://github.com/vmware/terraform-provider-vcd/pull/804))
* `resource/vcd_network_isolated` add support for `metadata` so this field can be set when creating/updating isolated networks. ([#804](https://github.com/vmware/terraform-provider-vcd/pull/804))
* `datasource/vcd_network_isolated` add support for `metadata` so this field can be retrieved when reading from isolated networks. ([#804](https://github.com/vmware/terraform-provider-vcd/pull/804))
* `resource/vcd_network_routed` add support for `metadata` so this field can be set when creating/updating routed networks. ([#804](https://github.com/vmware/terraform-provider-vcd/pull/804))
* `datasource/vcd_network_routed` add support for `metadata` so this field can be retrieved when reading from routed networks. ([#804](https://github.com/vmware/terraform-provider-vcd/pull/804))
* `resource/vcd_nsxt_ip_set` and `datasource/vcd_nsxt_ip_set` support VDC Groups by inheriting parent VDC
or VDC Group from Edge Gateway  ([#809](https://github.com/vmware/terraform-provider-vcd/pull/809))
* Data source `vcd_storage_profile` supports `metadata` so this field can be populated when reading VDC storage profiles. ([#811](https://github.com/vmware/terraform-provider-vcd/pull/811))
* `resource/vcd_nsxt_app_port_profile` and `datasource/vcd_nsxt_app_port_profile` add support
  for VDC Groups with new field `context_id` ([#812](https://github.com/vmware/terraform-provider-vcd/pull/812))
* `resource/vcd_nsxt_security_group` and `datasource/vcd_nsxt_security_group` support VDC Groups by inheriting parent VDC
or VDC Group from Edge Gateway  ([#814](https://github.com/vmware/terraform-provider-vcd/pull/814))
* `resource/vcd_network_isolated_v2` add support for `metadata` so this field can be set when creating/updating isolated NSX-T networks. ([#816](https://github.com/vmware/terraform-provider-vcd/pull/816))
* `datasource/vcd_network_isolated_v2` add support for `metadata` so this field can be retrieved when reading from isolated NSX-T networks. ([#816](https://github.com/vmware/terraform-provider-vcd/pull/816))
* `resource/vcd_network_routed_v2` add support for `metadata` so this field can be set when creating/updating routed NSX-T networks. ([#816](https://github.com/vmware/terraform-provider-vcd/pull/816))
* `datasource/vcd_network_routed_v2` add support for `metadata` so this field can be retrieved when reading from routed NSX-T networks. ([#816](https://github.com/vmware/terraform-provider-vcd/pull/816))
* `vcd_catalog_item` allows using `ovf_url` to upload vApp template from URL ([#770](https://github.com/vmware/terraform-provider-vcd/pull/770))
* `vcd_catalog_item` update allows changing `name` and `description` ([#770](https://github.com/vmware/terraform-provider-vcd/pull/770))
* `vcd_catalog` allows to publish a catalog externally to make its vApp templates and media files available for subscription by organizations outside the Cloud Director installation ([#772](https://github.com/vmware/terraform-provider-vcd/pull/772)], [[#773](https://github.com/vmware/terraform-provider-vcd/pull/773))
* `vcd_vapp_vm`, `vcd_vm` allows configuring advanced compute settings (shares and reservations) for VM ([#779](https://github.com/vmware/terraform-provider-vcd/pull/779))
* `vcd_independent_disk` allows creating additionally shared disks types by configuring `sharing_type` (VCD 10.3+). Also, add update support. Add new disk type `NVME` and new attributes `encrypted` and `uuid`. Import now allows listing all independent disks in VDC ([#789](https://github.com/vmware/terraform-provider-vcd/pull/789), [#810](https://github.com/vmware/terraform-provider-vcd/pull/810))

### BUG FIXES
* Fixes Issue #754 where VDC creation with storage profile `enabled=false` wasn't working ([#781](https://github.com/vmware/terraform-provider-vcd/pull/781))
* Fix Issue #611 when read of `vcd_vapp_vm` and `vcd_vm` resource failed when VM isn't found. Now allows Terraform to recreate resource when it isn't found. ([#783](https://github.com/vmware/terraform-provider-vcd/pull/783))
* Fix Issue #759 where enable_ip_masquerade handling in vcd_vapp_nat_rules resource wasn't correct ([#784](https://github.com/vmware/terraform-provider-vcd/pull/784))
* Fix bug in `datasource/vcd_nsxt_app_port_profile` where a lookup of a TENANT scope profile could
  fail finding exact Application Port Profile in case Org has multiple VDCs ([#812](https://github.com/vmware/terraform-provider-vcd/pull/812))

### NOTES
* Default values for `deployed_vm_quota` and `stored_vm_quota` for `org_user` have changed from 10 to 0 (unlimited) ([#798](https://github.com/vmware/terraform-provider-vcd/pull/798))
* Internal functions `lockParentEdgeGtw`, `unLockParentEdgeGtw`, `lockEdgeGateway`,
  `unlockEdgeGateway` were converted to use just their ID for lock key instead of full path
  `org:vdc:edge_id`. This is done because paths for VDC and VDC Groups can differ, but UUID is
  unique so it makes it simpler to manage ([#801](https://github.com/vmware/terraform-provider-vcd/pull/801))
* Additional locking mechanisms `lockIfOwnerIsVdcGroup`, `unLockIfOwnerIsVdcGroup`, `lockById`,
  `unlockById` ([#801](https://github.com/vmware/terraform-provider-vcd/pull/801))
* Bump `staticheck` tool to `2022.1` to support Go 1.18 and fix newly detected errors ([#813](https://github.com/vmware/terraform-provider-vcd/pull/813))
* Improve docs for `vcd_nsxt_network_dhcp` VDC Group support ([#818](https://github.com/vmware/terraform-provider-vcd/pull/818))
* Improves VDC Group guide documentation ([#818](https://github.com/vmware/terraform-provider-vcd/pull/818))

## 3.5.1 (January 13, 2022)

### BUG FIXES
Fix Issue #769 "Plugin did not respond: terraform-provider-vcd may crash with Terraform 1.1+ on some OSes".

The consequences of this fix are that some messages that were directed at the standard output (such as
progress percentage during uploads or suggestions when using outdated resources) are now written to the regular
log file (`go-vcloud-director.log`) using the special tag `[SCREEN]` for easy filtering. [#771](https://github.com/vmware/terraform-provider-vcd/pull/771)

## 3.5.0 (January 7, 2022)

### FEATURES
* **New Resource:** `vcd_nsxt_alb_settings` for managing NSX-T ALB on NSX-T Edge Gateways ([#726](https://github.com/vmware/terraform-provider-vcd/pull/726))
* **New Data source:** `vcd_nsxt_alb_settings` for reading NSX-T ALB on NSX-T Edge Gateways  ([#726](https://github.com/vmware/terraform-provider-vcd/pull/726))
* **New Resource:** `vcd_library_certificate` for managing certificates in library (provider configuration) ([#733](https://github.com/vmware/terraform-provider-vcd/pull/733))
* **New Data source:** `vcd_library_certificate` for reading certificates in library ([#733](https://github.com/vmware/terraform-provider-vcd/pull/733))
* **New Resource:** `vcd_nsxt_alb_edgegateway_service_engine_group` for managing NSX-T ALB Service Engine Groups
  assignments to Edge Gateways ([#738](https://github.com/vmware/terraform-provider-vcd/pull/738), [#764](https://github.com/vmware/terraform-provider-vcd/pull/764))
* **New Data source:** `vcd_nsxt_alb_edgegateway_service_engine_group` for reading NSX-T ALB Service Engine Groups
  assignments to Edge Gateways ([#738](https://github.com/vmware/terraform-provider-vcd/pull/738), [#764](https://github.com/vmware/terraform-provider-vcd/pull/764))
* **New Resource:** `vcd_vdc_group` for managing VDC Groups ([#752](https://github.com/vmware/terraform-provider-vcd/pull/752))
* **New Data source:** `vcd_vdc_group` for reading VDC Groups ([#752](https://github.com/vmware/terraform-provider-vcd/pull/752))
* **New Resource:** `vcd_nsxt_alb_pool` for NSX-T Load Balancer pools ([#756](https://github.com/vmware/terraform-provider-vcd/pull/756))
* **New Data source:** `vcd_nsxt_alb_pool` for reading NSX-T Load Balancer pools ([#756](https://github.com/vmware/terraform-provider-vcd/pull/756))
* **New Resource:** `vcd_nsxt_alb_virtual_service` for managing NSX-T ALB Virtual Service on NSX-T Edge Gateways
  ([#757](https://github.com/vmware/terraform-provider-vcd/pull/757), [#764](https://github.com/vmware/terraform-provider-vcd/pull/764))
* **New Data source:** `vcd_nsxt_alb_virtual_service` for reading NSX-T ALB Virtual Service on NSX-T Edge Gateways
  ([#757](https://github.com/vmware/terraform-provider-vcd/pull/757), [#764](https://github.com/vmware/terraform-provider-vcd/pull/764))

### IMPROVEMENTS
* Remove Coverity warnings from code ([#734](https://github.com/vmware/terraform-provider-vcd/pull/734))
* `resource/vcd_vapp_vm`, `resource/vcd_vm_internal_disk` add support for disk controller type nvme 
  ([#680](https://github.com/vmware/terraform-provider-vcd/pull/680),[#739](https://github.com/vmware/terraform-provider-vcd/pull/739))
* Add property `api_token` to provider, supporting API token authentication ([#742](https://github.com/vmware/terraform-provider-vcd/pull/742))
* Bump Terraform SDK to 2.10.0 ([#751](https://github.com/vmware/terraform-provider-vcd/pull/751))
* `resource/vcd_vapp` add support for `lease` settings management. [[#762](https://github.com/vmware/terraform-provider-vcd/pull/762))
* `datasource/vcd_vapp` add support for `lease` settings visualization. ([#762](https://github.com/vmware/terraform-provider-vcd/pull/762))

### BUG FIXES
* Fix bootable media connection in `vcd_vm` by sending media reference in `CreateVM` ([#714](https://github.com/vmware/terraform-provider-vcd/pull/714))
* Fix broken documentation links ([#721](https://github.com/vmware/terraform-provider-vcd/pull/721))
* Fix bug where using `vcd_vm_internal_disk` could remove `description` field in `vcd_vapp_vm` and `vcd_vm` ([#758](https://github.com/vmware/terraform-provider-vcd/pull/758))
* Skip tests for resources not available in 10.1 ([#761](https://github.com/vmware/terraform-provider-vcd/pull/761))
* Fix wrong references in `TestAccVcdVAppVmMultiNIC` ([#761](https://github.com/vmware/terraform-provider-vcd/pull/761))

### NOTES
* Add Guest Customization docs with examples to "guides" section ([#729](https://github.com/vmware/terraform-provider-vcd/pull/729))
* Improve HCL samples in documentation to only contain single newline spaces and adjust automated check to catch it
  ([#747](https://github.com/vmware/terraform-provider-vcd/pull/747))
* Add support for Terraform 1.1 CLI by upgrading Terraform Plugin SDK from v2.7.0 to v2.10.0 
  (there was a crash with older SDK version) ([#751](https://github.com/vmware/terraform-provider-vcd/pull/751))

## 3.4.0 (September 30, 2021)

### FEATURES
* **New Resource:** `vcd_nsxt_alb_controller` for managing NSX-T ALB Controllers (provider configuration) ([#709](https://github.com/vmware/terraform-provider-vcd/pull/709))
* **New Data source:** `vcd_nsxt_alb_controller` for reading NSX-T ALB Controllers (provider configuration)  ([#709](https://github.com/vmware/terraform-provider-vcd/pull/709),[#718](https://github.com/vmware/terraform-provider-vcd/pull/718))
* **New Resource:** `vcd_nsxt_alb_cloud` for managing NSX-T ALB Clouds (provider configuration)  ([#709](https://github.com/vmware/terraform-provider-vcd/pull/709))
* **New Data source:** `vcd_nsxt_alb_cloud` for reading NSX-T ALB Clouds (provider configuration)  ([#709](https://github.com/vmware/terraform-provider-vcd/pull/709),[#718](https://github.com/vmware/terraform-provider-vcd/pull/718))
* **New Data source:** `vcd_nsxt_alb_importable_cloud` for reading NSX-T ALB Importable Clouds (provider configuration)  ([#709](https://github.com/vmware/terraform-provider-vcd/pull/709),[#718](https://github.com/vmware/terraform-provider-vcd/pull/718))
* **New Resource:** `vcd_nsxt_alb_service_engine_group` for managing NSX-T ALB Service Engine Groups (provider configuration)  ([#709](https://github.com/vmware/terraform-provider-vcd/pull/709))
* **New Data source:** `vcd_nsxt_alb_service_engine_group` for reading NSX-T ALB Service Engine Groups (provider configuration)  ([#709](https://github.com/vmware/terraform-provider-vcd/pull/709),[#718](https://github.com/vmware/terraform-provider-vcd/pull/718))

### IMPROVEMENTS
* **Resource** and **Data source** `vcd_external_network_v2` support NSX-T Segment backed network ([#711](https://github.com/vmware/terraform-provider-vcd/pull/711))
* `vcd_org_vdc`: it is now possible adding and removing storage profiles ([#698](https://github.com/vmware/terraform-provider-vcd/pull/698))
* `data/vcd_nsxt_edge_cluster` add documentation for missing fields ([#700](https://github.com/vmware/terraform-provider-vcd/pull/700))
* Bump Terraform SDK to 2.7.0 ([#701](https://github.com/vmware/terraform-provider-vcd/pull/701))
* Formatted HCL docs using `terraform fmt`  ([#705](https://github.com/vmware/terraform-provider-vcd/pull/705))
* Updated attribute syntax to use first class expressions ([#705](https://github.com/vmware/terraform-provider-vcd/pull/705))
* Align build tags to match go fmt with Go 1.17 ([#707](https://github.com/vmware/terraform-provider-vcd/pull/707))
* Improve `test-tags.sh` script to handle new build tag format ([#707](https://github.com/vmware/terraform-provider-vcd/pull/707))
* Prevent invalid space in the base64 encoded authentication string in the token scripts specific for Linux ([#708](https://github.com/vmware/terraform-provider-vcd/pull/708))

### BUG FIXES
* Fix Issue #696: Catalog deletion failure returns as success ([#698](https://github.com/vmware/terraform-provider-vcd/pull/698))
* Fix Issue #648 `vcd_org_vdc`: adding a storage profile requires the vdc to be replaced ([#698](https://github.com/vmware/terraform-provider-vcd/pull/698))
* Primary NIC removal for `vcd_vapp_vm` and `vcd_vapp` is done in cold fashion ([#716](https://github.com/vmware/terraform-provider-vcd/pull/716))
* Change broken docs reference links with final docs format

### NOTES
* Drop support for VCD 10.0 ([#704](https://github.com/vmware/terraform-provider-vcd/pull/704))

## 3.3.1 (July 5, 2021)

### BUG FIXES
* (Issue #689) `resource/vcd_vm` and `resource/vcd_vapp_vm` cannot find templates in shared catalogs from other Orgs 
  ([#691](https://github.com/vmware/terraform-provider-vcd/pull/691))

## 3.3.0 (June 30, 2021)

### FEATURES
* **New Resource:** `vcd_nsxt_security_group` for NSX-T Edge Gateways ([#663](https://github.com/vmware/terraform-provider-vcd/pull/663))
* **New Data source:** `vcd_nsxt_security_group` for NSX-T Edge Gateways ([#663](https://github.com/vmware/terraform-provider-vcd/pull/663))
* **New Resource:** `vcd_nsxt_ip_set` for NSX-T IP Set management ([#668](https://github.com/vmware/terraform-provider-vcd/pull/668))
* **New Data source:** `vcd_nsxt_ip_set` for NSX-T IP Set management ([#668](https://github.com/vmware/terraform-provider-vcd/pull/668))
* **New Resource:** `vcd_nsxt_app_port_profile` for NSX-T Application Port Profile management ([#673](https://github.com/vmware/terraform-provider-vcd/pull/673))
* **New Data source:** `vcd_nsxt_app_port_profile` for NSX-T Application Port Profile management ([#673](https://github.com/vmware/terraform-provider-vcd/pull/673))
* **New Resource:** `vcd_nsxt_firewall` for NSX-T Edge Gateways ([#663](https://github.com/vmware/terraform-provider-vcd/pull/663))
* **New Data Source:** `vcd_nsxt_firewall` for NSX-T Edge Gateways ([#663](https://github.com/vmware/terraform-provider-vcd/pull/663))
* **New Resource:** `vcd_nsxt_nat_rule` for NSX-T Edge Gateways ([#676](https://github.com/vmware/terraform-provider-vcd/pull/676))
* **New Data source:** `vcd_nsxt_nat_rule` for NSX-T Edge Gateways ([#676](https://github.com/vmware/terraform-provider-vcd/pull/676))
* **New Resource:** `vcd_role` for provider and tenant role management ([#677](https://github.com/vmware/terraform-provider-vcd/pull/677))
* **New Resource:** `vcd_global_role` for provider role management ([#677](https://github.com/vmware/terraform-provider-vcd/pull/677))
* **New Resource:** `vcd_rights_bundle` for provider role management ([#677](https://github.com/vmware/terraform-provider-vcd/pull/677))
* **New Data source:** `vcd_right` for provider and tenant role management ([#677](https://github.com/vmware/terraform-provider-vcd/pull/677))
* **New Data source:** `vcd_role` for provider and tenant role management ([#677](https://github.com/vmware/terraform-provider-vcd/pull/677))
* **New Data source:** `vcd_global_role` for provider role management ([#677](https://github.com/vmware/terraform-provider-vcd/pull/677))
* **New Data source:** `vcd_rights_bundle` for provider role management ([#677](https://github.com/vmware/terraform-provider-vcd/pull/677))
* **New Resource:** `vcd_nsxt_ipsec_vpn_tunnel` for NSX-T IPsec VPN Tunnel management ([#678](https://github.com/vmware/terraform-provider-vcd/pull/678))
* **New Data source:** `vcd_nsxt_ipsec_vpn_tunnel` for NSX-T IPsec VPN Tunnel management ([#678](https://github.com/vmware/terraform-provider-vcd/pull/678))

### IMPROVEMENTS
* Fix a few typos and add example in Routed Network V2 docs ([#663](https://github.com/vmware/terraform-provider-vcd/pull/663))
* Add compatibility with Terraform 0.15.0 for `TestAccVcdNsxtStandaloneVmTemplate` and
  `TestAccVcdStandaloneVmTemplate` ([#663](https://github.com/vmware/terraform-provider-vcd/pull/663))
* Tests: add optional test config variable Misc.LdapContainer to override default LDAP container location.  
  (can be used to overcome Docker registry throttling issues for pulling testing LDAP container) ([#673](https://github.com/vmware/terraform-provider-vcd/pull/673))
* `datasource/vcd_resource_list` adds `vcd_right`, `vcd_role`, `vcd_global_role`, `vcd_rights_bundle` to the supported resource types ([#677](https://github.com/vmware/terraform-provider-vcd/pull/677))
* Replace `vdc.ComposeRawVApp` with `vdc.CreateRawVApp` ([#681](https://github.com/vmware/terraform-provider-vcd/pull/681))
* Change "vCloud Director" or "vCloudDirector" to "VMware Cloud Director" in all docs. ([#682](https://github.com/vmware/terraform-provider-vcd/pull/682))
* Update `org_user` docs to include creation of system users. ([#682](https://github.com/vmware/terraform-provider-vcd/pull/682))

### BUG FIXES
* `resource/vcd_edgegateway` not retrieved when there are more than 25 items ([#660](https://github.com/vmware/terraform-provider-vcd/pull/660))
* (Issue #633) vApp description was ignored in creation and update. ([#664](https://github.com/vmware/terraform-provider-vcd/pull/664))
* (Issue #634) Setting CPU count to less than what the template has throws error. ([#673](https://github.com/vmware/terraform-provider-vcd/pull/673))
* Tests: LDAP related tests had incorrect port mapping after image update ([#673](https://github.com/vmware/terraform-provider-vcd/pull/673))
* `resource/vcd_org_vdc` complained about storage profile name update from empty to specified on some VCD versions ([#676](https://github.com/vmware/terraform-provider-vcd/pull/676))

### NOTES
* Bump terraform-plugin-sdk to 2.4.4 ([#657](https://github.com/vmware/terraform-provider-vcd/pull/657))
* Add VCDClient.GetNsxtEdgeGateway, VCDClient.GetNsxtEdgeGatewayById, and
  GetNsxtEdgeGatewayFromResourceById convenience functions ([#663](https://github.com/vmware/terraform-provider-vcd/pull/663))
* Add `importStateIdNsxtEdgeGatewayObject` function for import testing of NSX-T Edge Gateway child
  objects ([#663](https://github.com/vmware/terraform-provider-vcd/pull/663))

## 3.2.0 (March 11, 2021)

### FEATURES
* **New Resource:** `vcd_network_routed_v2` for NSX-T and NSX-V routed networks ([#628](https://github.com/vmware/terraform-provider-vcd/issues/628))
* **New Data source:** `vcd_network_routed_v2` for NSX-T and NSX-V routed networks ([#628](https://github.com/vmware/terraform-provider-vcd/issues/628))
* **New Resource:** `vcd_network_isolated_v2` for NSX-T and NSX-V isolated networks ([#636](https://github.com/vmware/terraform-provider-vcd/issues/636))
* **New Data source:** `vcd_network_isolated_v2` for NSX-T and NSX-V isolated networks ([#636](https://github.com/vmware/terraform-provider-vcd/issues/636))
* **New Resource:** `vcd_vm` - Standalone VM ([#638](https://github.com/vmware/terraform-provider-vcd/issues/638))
* **New Data source:** `vcd_vm` - Standalone VM ([#638](https://github.com/vmware/terraform-provider-vcd/issues/638))
* **New Resource:** `vcd_nsxt_network_imported` for NSX-T imported networks ([#645](https://github.com/vmware/terraform-provider-vcd/issues/645))
* **New Data source:** `vcd_nsxt_network_imported` for NSX-T imported networks ([#645](https://github.com/vmware/terraform-provider-vcd/issues/645))
* **New Resource:** `vcd_nsxt_network_dhcp` for NSX-T routed network DHCP configuration ([#650](https://github.com/vmware/terraform-provider-vcd/issues/650))
* **New Data source:** `vcd_nsxt_network_dhcp` for NSX-T routed network DHCP configuration ([#650](https://github.com/vmware/terraform-provider-vcd/issues/650))

### IMPROVEMENTS
* `make install` will use lightweight tags for build version injection ([#628](https://github.com/vmware/terraform-provider-vcd/issues/628))
* `datasource/vcd_resource_list` adds `vcd_vm` to the supported resource types ([#638](https://github.com/vmware/terraform-provider-vcd/issues/638))
* `datasource/vcd_resource_list` adds `vcd_network_routed_v2` to the supported resource types ([#628](https://github.com/vmware/terraform-provider-vcd/issues/628))
* `datasource/vcd_resource_list` adds `vcd_network_isolated_v2` to the supported resource types ([#636](https://github.com/vmware/terraform-provider-vcd/issues/636))
* `datasource/vcd_resource_list` adds `vcd_nsxt_network_imported` to the supported resource types ([#645](https://github.com/vmware/terraform-provider-vcd/issues/645))
* `vcd_edgegateway` resource and datasource throws error (on create, import and datasource read) and refers to `vcd_nsxt_edgegateway` for NSX-T backed VDC ([#650](https://github.com/vmware/terraform-provider-vcd/issues/650))
* `vcd_network_isolated`and `vcd_network_routed` throw warnings on create and errors on import by referring to `vcd_network_isolated_v2`and `vcd_network_routed_v2` for NSX VDCs ([#650](https://github.com/vmware/terraform-provider-vcd/issues/650))
* `vcd_vapp_network` throws error when `org_network_name` is specified for NSX-T VDC (because NSX-T networks cannot be attached to vApp networks) ([#650](https://github.com/vmware/terraform-provider-vcd/issues/650))

### NOTES
* Internal functions `lockParentEdgeGtw` and `unLockParentEdgeGtw` will handle Edge Gateway locks when `name` or `id` reference is used ([#628](https://github.com/vmware/terraform-provider-vcd/issues/628))

## 3.1.0 (December 18, 2020)

FEATURES

* **New Resource:** `vcd_nsxt_edgegateway` - NSX-T edge gateway ([#600](https://github.com/vmware/terraform-provider-vcd/issues/600), [#608](https://github.com/vmware/terraform-provider-vcd/issues/608))
* **New Data source:** `vcd_nsxt_edgegateway` - NSX-T edge gateway ([#600](https://github.com/vmware/terraform-provider-vcd/issues/600))
* **New Data source:** `vcd_nsxt_edge_cluster` - allows to lookup ID and other details of NSX-T Edge Cluster ([#600](https://github.com/vmware/terraform-provider-vcd/issues/600))
* **New Data source:** `vcd_resource_list` info about VCD entities ([#594](https://github.com/vmware/terraform-provider-vcd/issues/594))
* **New Data source:** `vcd_resource_schema` definition of VCD resources schema ([#594](https://github.com/vmware/terraform-provider-vcd/issues/594))
* **New Data Source**: `vcd_storage_profile` for storage profile lookup ([#602](https://github.com/vmware/terraform-provider-vcd/issues/602))

IMPROVEMENTS

* `resource/vcd_vapp_vm` adds support to update `storage_profile` ([#580](https://github.com/vmware/terraform-provider-vcd/issues/580))
* `resource/vcd_org_vdc` adds support to update `storage_profile` ([#583](https://github.com/vmware/terraform-provider-vcd/issues/583))
* `resource/vcd_org_vdc` new computed field `storage_used_in_mb` ([#583](https://github.com/vmware/terraform-provider-vcd/issues/583))
* `resource/vcd_catalog` allows to set and update `storage_profile_id` ([#602](https://github.com/vmware/terraform-provider-vcd/issues/602))
* `resource/vcd_catalog` adds support to update `description` ([#602](https://github.com/vmware/terraform-provider-vcd/issues/602))
* `datasource/vcd_catalog` exports `storage_profile_id` ([#602](https://github.com/vmware/terraform-provider-vcd/issues/602))
* Provider: add support for bearer tokens in addition to authorization tokens ([#590](https://github.com/vmware/terraform-provider-vcd/issues/590))
* Provider: automatically use `/cloudapi/1.0.0/sessions/provider` when `/api/sessions` is disabled ([#590](https://github.com/vmware/terraform-provider-vcd/issues/590))

BUG FIXES

* `resource/vcd_external_network_v2` may fail when creating NSX-V network backed by standard vSwitch port group ([#600](https://github.com/vmware/terraform-provider-vcd/issues/600))

NOTES

* Bump terraform-plugin-sdk to v2.2.0 ([#576](https://github.com/vmware/terraform-provider-vcd/issues/576), [#594](https://github.com/vmware/terraform-provider-vcd/issues/594))

## 3.0.0 (October 16, 2020)

FEATURES

* **New Resource**: `vcd_vapp_access_control` Access control for vApps [#543](https://github.com/vmware/terraform-provider-vcd/pull/543)
* **New Data Source**: `vcd_org_user` Org User [#543](https://github.com/vmware/terraform-provider-vcd/pull/543)
* **New Resource**: `vcd_vm_sizing_policy` VDC VM sizing policy [#553](https://github.com/vmware/terraform-provider-vcd/pull/553)
* **New Data Source**: `vcd_vm_sizing_policy` VDC VM sizing policy [#553](https://github.com/vmware/terraform-provider-vcd/pull/553)
* **New Resource**: `vcd_edgegateway_settings` Changes LB and FW global settings for Edge Gateway [#557](https://github.com/vmware/terraform-provider-vcd/pull/557)
* **New Resource**: `vcd_external_network_v2` with NSX-T support [#560](https://github.com/vmware/terraform-provider-vcd/pull/560)
* **New Data Source**: `vcd_external_network_v2` with NSX-T support [#560](https://github.com/vmware/terraform-provider-vcd/pull/560)
* **New Data Source**: `vcd_vcenter` useful for `vcd_external_network_v2` resource when used with NSX-V [#560](https://github.com/vmware/terraform-provider-vcd/pull/560)
* **New Data Source**: `vcd_portgroup` useful for `vcd_external_network_v2` resource when used with NSX-V [#560](https://github.com/vmware/terraform-provider-vcd/pull/560)
* **New Data Source**: `vcd_nsxt_manager` useful for `vcd_external_network_v2` resource when used with NSX-T [#560](https://github.com/vmware/terraform-provider-vcd/pull/560)
* **New Data Source**: `vcd_nsxt_tier0_router` useful for `vcd_external_network_v2` resource when used with NSX-T [#560](https://github.com/vmware/terraform-provider-vcd/pull/560)

IMPROVEMENTS

* Added command `make tagverify` to check testing tags isolation. It also runs when calling `make test` [#532](https://github.com/vmware/terraform-provider-vcd/pull/532)
* Added directory `.changes` and script `./scripts/make-changelog.sh` to handle CHANGELOG entries [#534](https://github.com/vmware/terraform-provider-vcd/pull/534)
* `resource/vcd_vapp_vm` allows toggle network connection with `network.connected` [#535](https://github.com/vmware/terraform-provider-vcd/pull/535)
* `resource/vcd_vapp_vm` allows toggle memory and vCPU hot add with `cpu_hot_add_enabled` and `memory_hot_add_enabled` [#536](https://github.com/vmware/terraform-provider-vcd/pull/536)
* `resource/vcd_vapp_vm` allows change `network` parameters without VM power off [#536](https://github.com/vmware/terraform-provider-vcd/pull/536)
* Repository has a new home! Moved from https://github.com/terraform-providers/terraform-provider-vcd to https://github.com/vmware/terraform-provider-vcd [#542](https://github.com/vmware/terraform-provider-vcd/pull/542)
* Added support for NSX-T Org VDC [#550](https://github.com/vmware/terraform-provider-vcd/pull/550)
* `resource/vcd_org_vdc` new fields for assigning VM sizing policies `vm_sizing_policy_ids` and `default_vm_sizing_policy_id` [#553](https://github.com/vmware/terraform-provider-vcd/pull/553)
* `resource/vcd_vapp_vm` new field `sizing_policy_id` uses VM sizing policy [#553](https://github.com/vmware/terraform-provider-vcd/pull/553)

BUG FIXES

* `resource/vcd_vapp_vm` removed default value for `cpus` and `cpu_cores` [#553](https://github.com/vmware/terraform-provider-vcd/pull/553)
* `resource/vcd_vapp_vm` fix ignoring `is_primary=false` [#556](https://github.com/vmware/terraform-provider-vcd/pull/556)

NOTES

* Added support for VCD 10.2 [#544](https://github.com/vmware/terraform-provider-vcd/pull/544)
* Dropped support for VCD 9.5 [#544](https://github.com/vmware/terraform-provider-vcd/pull/544)
* `resource/vcd_nsxv_firewall_rule` `virtual_machine_ids` renamed to `vm_ids` [#558](https://github.com/vmware/terraform-provider-vcd/pull/558)
* `resource/vcd_vm_affinity_rule` `virtual_machine_ids` renamed to `vm_ids` [#558](https://github.com/vmware/terraform-provider-vcd/pull/558)
* Provider will send HTTP User-Agent while performing API calls [#566](https://github.com/vmware/terraform-provider-vcd/pull/566)
* Added conditional skips for some checks in test `TestAccVcdVAppVmDhcpWait`

REMOVALS

* Fixed `vcd_independent_disk.size` issue, new field `size_in_mb` replaces the `size` [#588](https://github.com/vmware/terraform-provider-vcd/pull/588)
* Removed deprecated resource `vcd_network` [#543](https://github.com/vmware/terraform-provider-vcd/pull/543)
* Removed deprecated resources `vcd_dnat`, `vcd_snat`, and `vcd_firewall_rules` [#557](https://github.com/vmware/terraform-provider-vcd/pull/557)
* Removed deprecated attributes `ip, network_name, vapp_network_name, network_href, mac, initscript` from `vcd_vapp_vm` [#563](https://github.com/vmware/terraform-provider-vcd/pull/563)
* Removed deprecated attributes `external_networks, default_gateway_network, advaced` from `vcd_edgegateway` [#588](https://github.com/vmware/terraform-provider-vcd/pull/588)
* Removed `vcd_independent_disk.size` in favor of `vcd_independent_disk.size_in_mb` [#588](https://github.com/vmware/terraform-provider-vcd/pull/588)
* Removed deprecated attributes `template_name, catalog_name, network_name, memory, cpus, ip, storage_profile, initscript, ovf, accept_all_eulas` from `vcd_vapp` [#588](https://github.com/vmware/terraform-provider-vcd/pull/588)

## 2.9.0 (June 30, 2020)
FEATURES:

* **New Resource**: `vcd_vm_affinity_rule` VM affinity and anti-affinity rules ([#514](https://github.com/vmware/terraform-provider-vcd/issues/514))
* **New Data Source**: `vcd_vm_affinity_rule` VM affinity and anti-affinity rules ([#514](https://github.com/vmware/terraform-provider-vcd/issues/514))
* Add support for SAML auth with Active Directory Federation Services (ADFS) as IdP using
  "/adfs/services/trust/13/usernamemixed" endpoint usin auth_type="saml_adfs". ([#504](https://github.com/vmware/terraform-provider-vcd/issues/504))
* Add support for LDAP authentication using auth_type="integrated". ([#504](https://github.com/vmware/terraform-provider-vcd/issues/504))
* **New Resource:** `vcd_org_group` Org Group management ([#513](https://github.com/vmware/terraform-provider-vcd/issues/513))
* **New Resource:** `resource/vcd_vapp_firewall_rules` vApp network firewall rules ([#511](https://github.com/vmware/terraform-provider-vcd/issues/511))
* **New Resource:** `resource/vcd_vapp_nat_rules` vApp network NAT rules ([#518](https://github.com/vmware/terraform-provider-vcd/issues/518))
* **New Resource:** `resource/vcd_vapp_static_routing` vApp network static routing rules ([#520](https://github.com/vmware/terraform-provider-vcd/issues/520))

IMPROVEMENTS:

* `resource/vcd_vapp_vm` allows creating empty VM. New fields added `boot_image`, `os_type` and `hardware_version`. Also, supports `description` updates. ([#484](https://github.com/vmware/terraform-provider-vcd/issues/484))
* Data sources `vcd_edgegateway`, `vcd_catalog`, `vcd_catalog_item`, `vcd_catalog_media`, `vcd_network_*` allow search by filter. ([#487](https://github.com/vmware/terraform-provider-vcd/issues/487))
* Removed code that handled specific cases for API 29.0 and 30.0. This library now supports VCD versions from 9.5 to 10.1 included ([#499](https://github.com/vmware/terraform-provider-vcd/issues/499))
* Added command line flags to test suite, corresponding to environment variables listed in TESTING.md ([#505](https://github.com/vmware/terraform-provider-vcd/issues/505))
* `resource/vcd_vapp_vm` allows creating VM from multi VM vApp template ([#501](https://github.com/vmware/terraform-provider-vcd/issues/501))

BUG FIXES:
* `resource/vcd_vapp_vm` and `datasource/vcd_vapp_vm` can report `network.X.is_primary` attribute
  incorrectly when VM is imported to Terraform and NIC indexes in vCD do not start with 0. [[#512](https://github.com/vmware/terraform-provider-vcd/issues/512)] 
* Rename docs files from `.markdown` to `.html.markdown` (Add test to check file name consistency) ([#522](https://github.com/vmware/terraform-provider-vcd/issues/522))
* `nat_enabled` and `firewall_enabled` were incorrectly added to `vcd_vapp_network` and would collide with the depending resources. 
Now moved to respective resources `vcd_vapp_nat_rules` and `vcd_vapp_firewall_rules`.

DEPRECATIONS:

* Deprecated `vcd_snat` (replaced by `vcd_nsxv_snat`), `vcd_dnat` (replaced by `vcd_nsxv_dnat`), and `vcd_firewall_rules` (replaced by `vcd_nsxv_firewall_rule`) ([#518](https://github.com/vmware/terraform-provider-vcd/issues/518))
  The deprecated resources are to be used only with non-advanced edge gateway.


NOTES:

* Dropped support for vCD 9.1 ([#492](https://github.com/vmware/terraform-provider-vcd/issues/492))

## 2.8.0 (April 16, 2020)

IMPROVEMENTS:

* `resource/vcd_network_routed`, `resource/vcd_network_direct`, and `resource/vcd_network_isolated` now support in place updates. ([#465](https://github.com/vmware/terraform-provider-vcd/issues/465))
* `vcd_vapp_network`, `vcd_vapp_org_network` has now missing import documentation [[#481](https://github.com/vmware/terraform-provider-vcd/issues/481)] 
* `resource/vcd_vapp_vm` and `datasource/vcd_vapp_vm` simplifies network adapter validation when it
  is not attached to network (`network.x.type=none` and `network.x.ip_allocation_mode=none`) and
  `network_dhcp_wait_seconds` is defined. This is required for vCD 10.1 support ([#485](https://github.com/vmware/terraform-provider-vcd/issues/485))

BUG FIXES
* Using wrong defaults for `vcd_network_isolated` and `vcd_network_routed` DNS ([#434](https://github.com/vmware/terraform-provider-vcd/issues/434))
* `external_network_gateway` not filled in datasource `vcd_network_direct` ([#450](https://github.com/vmware/terraform-provider-vcd/issues/450))
* `resource/vcd_vapp_vm` sometimes reports incorrect `vcd_vapp_vm.ip` and `vcd_vapp_vm.mac` fields in deprecated network
configuration (when using `vcd_vapp_vm.network_name` and `vcd_vapp_vm.vapp_network_name` parameters instead of
`vcd_vapp_vm.network` blocks) ([#478](https://github.com/vmware/terraform-provider-vcd/issues/478))
* `resource/vcd_vapp_org_network` fix potential error 'NAT rule cannot be configured for nics with
  DHCP addressing mode' during removal ([#489](https://github.com/vmware/terraform-provider-vcd/issues/489))
* `resource/vcd_org_vdc` supports vCD 10.1 and resolves "no provider VDC found" errors ([#489](https://github.com/vmware/terraform-provider-vcd/issues/489))

DEPRECATIONS:
* vCD 9.1 support is deprecated. Next version will require at least version 9.5 ([#489](https://github.com/vmware/terraform-provider-vcd/issues/489))

NOTES:

* Bump terraform-plugin-sdk to v1.8.0 ([#479](https://github.com/vmware/terraform-provider-vcd/issues/479))
* Update Travis to use Go 1.14 ([#479](https://github.com/vmware/terraform-provider-vcd/issues/479))

## 2.7.0 (March 13, 2020)

FEATURES:

* **New Resource:** `vcd_vm_internal_disk` VM internal disk configuration ([#412](https://github.com/vmware/terraform-provider-vcd/issues/412))
* **New Resource:** `vcd_vapp_org_network` vApp organization network ([#455](https://github.com/vmware/terraform-provider-vcd/issues/455))
* **New Data Source:** `vcd_vapp_org_network` vApp org network ([#455](https://github.com/vmware/terraform-provider-vcd/issues/455))
* **New Data Source:** `vcd_vapp_network` vApp network ([#455](https://github.com/vmware/terraform-provider-vcd/issues/455))

IMPROVEMENTS:

* `vcd_vapp_network` supports isolated network and vApp network connected to Org VDC networks
  ([#455](https://github.com/vmware/terraform-provider-vcd/issues/455),[#468](https://github.com/vmware/terraform-provider-vcd/issues/468))
* `vcd_vapp_network` does not default `dns1` and `dns2` fields to 8.8.8.8 and 8.8.4.4 respectively
  ([#455](https://github.com/vmware/terraform-provider-vcd/issues/455),[#468](https://github.com/vmware/terraform-provider-vcd/issues/468))
* `vcd_org_vdc` can be created with Flex allocation model in vCD 9.7 and later. Also two new fields added
  for Flex - `elasticity`, `include_vm_memory_overhead` ([#443](https://github.com/vmware/terraform-provider-vcd/issues/443))
* `resource/vcd_org` and `datasource/vcd_org` include a section `vapp_lease` and a section
  `vapp_template_lease` to define lease related parameters of depending entities - ([#432](https://github.com/vmware/terraform-provider-vcd/issues/432))
* `resource/vcd_vapp_vm` Internal disks in VM template can be edited by `override_template_disk`
  field ([#412](https://github.com/vmware/terraform-provider-vcd/issues/412))
* `vcd_vapp_vm` `disk` has new attribute `size_in_mb` ([#433](https://github.com/vmware/terraform-provider-vcd/issues/433))
* `resource/vcd_vapp_vm` and `datasource/vcd_vapp_vm` get optional `network_dhcp_wait_seconds` field
  to ensure `ip` is reported when `ip_allocation_mode=DHCP` is used ([#436](https://github.com/vmware/terraform-provider-vcd/issues/436))
* `resource/vcd_vapp_vm` and `datasource/vcd_vapp_vm` include a field `adapter_type` in `network`
  definition to specify NIC type - ([#441](https://github.com/vmware/terraform-provider-vcd/issues/441))
* `resource/vcd_vapp_vm` and `datasource/vcd_vapp_vm` `customization` block supports all available
  features ([#462](https://github.com/vmware/terraform-provider-vcd/issues/462), [#469](https://github.com/vmware/terraform-provider-vcd/issues/469), [#477](https://github.com/vmware/terraform-provider-vcd/issues/477))
* `datasource/*` - all data sources return an error when object is not found ([#446](https://github.com/vmware/terraform-provider-vcd/issues/446), [#470](https://github.com/vmware/terraform-provider-vcd/issues/470))
* `vcd_vapp_vm` allows to add routed vApp network, not only isolated one. `network.name` can reference 
  `vcd_vapp_network.name` of a vApp network with `org_network_name` set ([#472](https://github.com/vmware/terraform-provider-vcd/issues/472))

DEPRECATIONS:
* `resource/vcd_vapp_vm` `network.name` deprecates automatic attachment of vApp Org network when
  `network.type=org` and it is not attached with `vcd_vapp_org_network` before referencing it.
  ([#455](https://github.com/vmware/terraform-provider-vcd/issues/455))
* `resource/vcd_vapp_vm` field `initscript` is now deprecated in favor of
  `customization.0.initscript` to group all guest customization settings ([#462](https://github.com/vmware/terraform-provider-vcd/issues/462))

BUG FIXES:

* `resource/vcd_vapp_vm` read - independent disks where losing `bus_number` and `unit_number`
  values after refresh ([#433](https://github.com/vmware/terraform-provider-vcd/issues/433))
* `datasource/vcd_nsxv_dhcp_relay` crashes if no DHCP relay settings are present in Edge Gateway
  ([#446](https://github.com/vmware/terraform-provider-vcd/issues/446))
* `resource/vcd_vapp_vm` `network` block changes caused MAC address changes in existing NICs
  ([#436](https://github.com/vmware/terraform-provider-vcd/issues/436),[#407](https://github.com/vmware/terraform-provider-vcd/issues/407))
* Fix a potential data race in client connection caching when VCD_CACHE is enabled ([#453](https://github.com/vmware/terraform-provider-vcd/issues/453))
* `resource/vcd_vapp_vm` when `customization.0.force=false` crashes with `interface {} is nil` ([#462](https://github.com/vmware/terraform-provider-vcd/issues/462))
* `resource/vcd_vapp_vm` `customization.0.force=true` could have skipped "Forced customization" on
  each apply ([#462](https://github.com/vmware/terraform-provider-vcd/issues/462), [#477](https://github.com/vmware/terraform-provider-vcd/issues/477))

NOTES:

* Drop support for vCD 9.0
* Bump terraform-plugin-sdk to v1.5.0 ([#442](https://github.com/vmware/terraform-provider-vcd/issues/442))
* `make seqtestacc` and `make test-binary` use `-race` flags for `go test` to check if there are no
  data races. Additionally GNUMakefile supports `make installrace` and `make buildrace` to build
  binary with race detection enabled. ([#453](https://github.com/vmware/terraform-provider-vcd/issues/453))
* Add `make test-upgrade-prepare` directive ([#462](https://github.com/vmware/terraform-provider-vcd/issues/462))

## 2.6.0 (December 13, 2019)

FEATURES:

* **New Resource:** `vcd_nsxv_dhcp_relay` Edge gateway DHCP relay configuration - ([#416](https://github.com/vmware/terraform-provider-vcd/issues/416))
* **New Resource:** `vcd_nsxv_ip_set` IP set - ([#406](https://github.com/vmware/terraform-provider-vcd/issues/406),[#411](https://github.com/vmware/terraform-provider-vcd/issues/411))
* **New Data Source:** `vcd_nsxv_dhcp_relay` Edge gateway DHCP relay configuration - ([#416](https://github.com/vmware/terraform-provider-vcd/issues/416))
* **New Data Source:** `vcd_vapp_vm` VM - ([#218](https://github.com/vmware/terraform-provider-vcd/issues/218))
* **New Data Source:** `vcd_nsxv_ip_set` IP set - ([#406](https://github.com/vmware/terraform-provider-vcd/issues/406),[#411](https://github.com/vmware/terraform-provider-vcd/issues/411))
* **New build command:** `make test-upgrade` to run an upgrade test from the previous released version

IMPROVEMENTS:

* Switch to Terraform terraform-plugin-sdk v1.3.0 as per recent [HashiCorp
  recommendation](https://www.terraform.io/docs/extend/plugin-sdk.html) - ([#382](https://github.com/vmware/terraform-provider-vcd/issues/382), [#406](https://github.com/vmware/terraform-provider-vcd/issues/406))
* `resource/vcd_vapp_vm` VM state ID changed from VM name to vCD ID
* `resource/vcd_vapp_vm` Add properties `description` and `storage_profile`
* `resource/vcd_vapp_vm` Add import capability and full read support ([#218](https://github.com/vmware/terraform-provider-vcd/issues/218))

* `resource/vcd_nsxv_dnat` and `resource/vcd_nsxv_dnat` more precise error message when network is
  not found - ([#384](https://github.com/vmware/terraform-provider-vcd/issues/384))
* `resource/vcd_nsxv_dnat` and `resource/vcd_nsxv_dnat` `rule_tag` must be int to avoid vCD internal
  exception passthrough - ([#384](https://github.com/vmware/terraform-provider-vcd/issues/384))
* `resource/vcd_nsxv_dnat` put correct name in doc example - ([#384](https://github.com/vmware/terraform-provider-vcd/issues/384))
* `resource/vcd_nsxv_dnat` and `resource/vcd_nsxv_dnat` avoid rule replacement because of changed
  `rule_tag` when rule is altered via UI - ([#384](https://github.com/vmware/terraform-provider-vcd/issues/384))
* `resource/vcd_nsxv_firewall_rule` add explicit protocol validation to avoid odd NSX-V API error -
  ([#384](https://github.com/vmware/terraform-provider-vcd/issues/384))
* `resource/vcd_nsxv_firewall_rule` `rule_tag` must be int to avoid vCD internal exception
  passthrough - ([#384](https://github.com/vmware/terraform-provider-vcd/issues/384))
* Fix code warnings from `staticcheck` and add command `make static` to Travis tests
* `resource/vcd_edge_gateway` and `datasource/vcd_edge_gateway` add `default_external_network_ip`
  and `external_network_ips` fields to export default edge gateway IP address and other external
  network IPs used on gateway interfaces - ([#389](https://github.com/vmware/terraform-provider-vcd/issues/389), [#401](https://github.com/vmware/terraform-provider-vcd/issues/401))
* Add `token` to the `vcd` provider for the ability of connecting with an authorization token - ([#280](https://github.com/vmware/terraform-provider-vcd/issues/280))
* Add command `make token` to create an authorization token from testing credentials
* Clean up interpolation-only expressions from tests (as allowed in terraform v0.12.11+)
* Increment vCD API version used from 27.0 to 29.0 ([#396](https://github.com/vmware/terraform-provider-vcd/issues/396))
* `resource/vcd_network_routed` Add properties `description` and `interface_type` ([#321](https://github.com/vmware/terraform-provider-vcd/issues/321),[#342](https://github.com/vmware/terraform-provider-vcd/issues/342),[#374](https://github.com/vmware/terraform-provider-vcd/issues/374))
* `resource/vcd_network_isolated` Add property `description` ([#373](https://github.com/vmware/terraform-provider-vcd/issues/373))
* `resource/vcd_network_direct` Add property `description`
* `resource/vcd_network_routed` Add check for valid IPs ([#374](https://github.com/vmware/terraform-provider-vcd/issues/374))
* `resource/vcd_network_isolated` Add check for valid IPs ([#373](https://github.com/vmware/terraform-provider-vcd/issues/373))
* `resource/vcd_nsxv_firewall_rule` Add support for IP sets ([#411](https://github.com/vmware/terraform-provider-vcd/issues/411))
* `resource/vcd_edgegateway` new fields `fips_mode_enabled`, `use_default_route_for_dns_relay`
  - ([#401](https://github.com/vmware/terraform-provider-vcd/issues/401),[#414](https://github.com/vmware/terraform-provider-vcd/issues/414))
* `resource/vcd_edgegateway`  new `external_network` block for advanced configurations of external
  networks including multiple subnets, IP pool sub-allocation and rate limits - ([#401](https://github.com/vmware/terraform-provider-vcd/issues/401),[#418](https://github.com/vmware/terraform-provider-vcd/issues/418))
* `resource/vcd_edgegateway` enables read support for field `distributed_routing` after switch to
  vCD API v29.0 - ([#401](https://github.com/vmware/terraform-provider-vcd/issues/401))
* `vcd_nsxv_firewall_rule` - improve internal lookup mechanism for `gateway_interfaces` field in
  source and/or destination ([#419](https://github.com/vmware/terraform-provider-vcd/issues/419))

BUG FIXES:

* Fix `vcd_org_vdc` datasource read. When user was Organization administrator datasource failed. Fields provider_vdc_name, storage_profile, memory_guaranteed, cpu_guaranteed, cpu_speed, enable_thin_provisioning, enable_fast_provisioning, network_pool_name won't have values for org admin.
* Removed `power_on` property from data source `vcd_vapp`, as it is a directive used during vApp build.
  Its state is never updated and the fields `status` and `status_text` already provide the necessary information.
  ([#379](https://github.com/vmware/terraform-provider-vcd/issues/379))
* Fix `vcd_independent_disk` reapply issue, which was seen when optional `bus_sub_type` and `bus_type` wasn't used - ([#394](https://github.com/vmware/terraform-provider-vcd/issues/394))
* Fix `vcd_vapp_network` apply issue, where the property `guest_vlan_allowed` was applied only to the last of multiple networks.
* `datasource/vcd_network_direct` is now readable by Org User (previously it was only by Sys Admin), as this change made it possible to get the details of External Network as Org User ([#408](https://github.com/vmware/terraform-provider-vcd/issues/408))

DEPRECATIONS:

* Deprecated property `storage_profile` in resource `vcd_vapp`, as the corresponding field is now enabled in `vcd_vapp_vm`
* `resource/vcd_edgegateway` deprecates fields `external_networks` and `default_gateway_network` in
  favor of new `external_network` block(s) - [[#401](https://github.com/vmware/terraform-provider-vcd/issues/401)] 

## 2.5.0 (October 28, 2019)

FEATURES:

* **New Resource:** `vcd_nsxv_dnat` DNAT for advanced edge gateways using proxied NSX-V API - ([#328](https://github.com/vmware/terraform-provider-vcd/issues/328))
* **New Resource:** `vcd_nsxv_snat`  SNAT for advanced edge gateways using proxied NSX-V API - ([#328](https://github.com/vmware/terraform-provider-vcd/issues/328))
* **New Resource:** `vcd_nsxv_firewall_rule`  firewall for advanced edge gateways using proxied NSX-V API - ([#341](https://github.com/vmware/terraform-provider-vcd/issues/341), [#358](https://github.com/vmware/terraform-provider-vcd/issues/358))
* **New Data Source:** `vcd_org` Organization - ([#218](https://github.com/vmware/terraform-provider-vcd/issues/218))
* **New Data Source:** `vcd_catalog` Catalog - ([#218](https://github.com/vmware/terraform-provider-vcd/issues/218))
* **New Data Source:** `vcd_catalog_item` CatalogItem - ([#218](https://github.com/vmware/terraform-provider-vcd/issues/218))
* **New Data Source:** `vcd_org_vdc` Organization VDC - ([#324](https://github.com/vmware/terraform-provider-vcd/issues/324))
* **New Data Source:** `vcd_external_network` External Network - ([#218](https://github.com/vmware/terraform-provider-vcd/issues/218))
* **New Data Source:** `vcd_edgegateway` Edge Gateway - ([#218](https://github.com/vmware/terraform-provider-vcd/issues/218))
* **New Data Source:** `vcd_network_routed` Routed Network - ([#218](https://github.com/vmware/terraform-provider-vcd/issues/218))
* **New Data Source:** `vcd_network_isolated` Isolated Network - ([#218](https://github.com/vmware/terraform-provider-vcd/issues/218))
* **New Data Source:** `vcd_network_direct` Direct Network - ([#218](https://github.com/vmware/terraform-provider-vcd/issues/218))
* **New Data Source:** `vcd_vapp` vApp - ([#218](https://github.com/vmware/terraform-provider-vcd/issues/218))
* **New Data Source:** `vcd_nsxv_dnat` DNAT for advanced edge gateways using proxied NSX-V API - ([#328](https://github.com/vmware/terraform-provider-vcd/issues/328))
* **New Data Source:** `vcd_nsxv_snat` SNAT for advanced edge gateways using proxied NSX-V API - ([#328](https://github.com/vmware/terraform-provider-vcd/issues/328))
* **New Data Source:** `vcd_nsxv_firewall_rule` firewall for advanced edge gateways using proxied NSX-V API - ([#341](https://github.com/vmware/terraform-provider-vcd/issues/341))
* **New Data Source:** `vcd_independent_disk` Independent disk - ([#349](https://github.com/vmware/terraform-provider-vcd/issues/349))
* **New Data Source:** `vcd_catalog_media` Media item - ([#340](https://github.com/vmware/terraform-provider-vcd/issues/340))

IMPROVEMENTS:

* Add support for vCD 10.0
* `resource/vcd_org` Add import capability and full read support ([#218](https://github.com/vmware/terraform-provider-vcd/issues/218))
* `resource/vcd_catalog` Add import capability and full read support ([#218](https://github.com/vmware/terraform-provider-vcd/issues/218))
* `resource/vcd_catalog_item` Add import capability and full read support ([#218](https://github.com/vmware/terraform-provider-vcd/issues/218))
* `resource/vcd_external_network` Add import capability and full read support ([#218](https://github.com/vmware/terraform-provider-vcd/issues/218))
* `resource/vcd_edgegateway` Add import capability and full read support ([#218](https://github.com/vmware/terraform-provider-vcd/issues/218))
* `resource/vcd_network_routed` Add import capability and full read support ([#218](https://github.com/vmware/terraform-provider-vcd/issues/218))
* `resource/vcd_network_isolated` Add import capability and full read support ([#218](https://github.com/vmware/terraform-provider-vcd/issues/218))
* `resource/vcd_network_direct` Add import capability and full read support ([#218](https://github.com/vmware/terraform-provider-vcd/issues/218))
* `resource/vcd_vapp` Add import capability and full read support ([#218](https://github.com/vmware/terraform-provider-vcd/issues/218))
* `resource/vcd_network_direct`: Direct network state ID changed from network name to vCD ID 
* `resource/vcd_network_isolated`: Isolated network state ID changed from network name to vCD ID 
* `resource/vcd_network_routed`: Routed network state ID changed from network name to vCD ID 
* `resource/vcd_vapp`: vApp state ID changed from vApp name to vCD ID
* `resource/vcd_vapp`: Add properties `status` and `status_text`
* `resource/catalog_item` added catalog item metadata support [[#285](https://github.com/vmware/terraform-provider-vcd/issues/285)] 
* `resource/vcd_catalog`: Catalog state ID changed from catalog name to vCD ID 
* `resource/vcd_catalog_item`: CatalogItem state ID changed from colon separated list of catalog name and item name to vCD ID 
* `resource/catalog_item` added catalog item metadata support [[#298](https://github.com/vmware/terraform-provider-vcd/issues/298)] 
* `resource/catalog_media` added catalog media item metadata support ([#298](https://github.com/vmware/terraform-provider-vcd/issues/298))
* `resource/vcd_vapp_vm` supports update for `network` block ([#310](https://github.com/vmware/terraform-provider-vcd/issues/310))
* `resource/vcd_vapp_vm` allows to force guest customization ([#310](https://github.com/vmware/terraform-provider-vcd/issues/310))
* `resource/vcd_vapp` supports guest properties ([#319](https://github.com/vmware/terraform-provider-vcd/issues/319))
* `resource/vcd_vapp_vm` supports guest properties ([#319](https://github.com/vmware/terraform-provider-vcd/issues/319))
* `resource/vcd_network_direct` Add computed properties (external network gateway, netmask, DNS, and DNS suffix) ([#330](https://github.com/vmware/terraform-provider-vcd/issues/330))
* `vcd_org_vdc` Add import capability and full read support ([#218](https://github.com/vmware/terraform-provider-vcd/issues/218))
* Upgrade Terraform SDK dependency to 0.12.8 ([#320](https://github.com/vmware/terraform-provider-vcd/issues/320))
* `resource/vcd_vapp_vm` has new field `computer_name` ([#334](https://github.com/vmware/terraform-provider-vcd/issues/334))
* Import functions can now use custom separators instead of "." ([#343](https://github.com/vmware/terraform-provider-vcd/issues/343))
* `resource/vcd_independent_disk` Add computed properties (`iops`, `owner_name`, `datastore_name`, `is_attached`) and read support for all fields except the `size` ([#349](https://github.com/vmware/terraform-provider-vcd/issues/349))
* `resource/vcd_independent_disk` Disk state ID changed from name of disk to vCD ID ([#349](https://github.com/vmware/terraform-provider-vcd/issues/349))
* Import functions can now use a custom separator instead of "." ([#343](https://github.com/vmware/terraform-provider-vcd/issues/343))
* `resource/vcd_catalog_media` Add computed properties (`is_iso`, `owner_name`, `is_published`, `creation_date`, `size`, `status`, `storage_profile_name`) and full read support ([#340](https://github.com/vmware/terraform-provider-vcd/issues/340))
* `resource/vcd_catalog_media` MediaItem state ID changed from colon separated list of catalog name and media name to vCD ID ([#340](https://github.com/vmware/terraform-provider-vcd/issues/340))
* Import functions can now use custom separators instead of "." ([#343](https://github.com/vmware/terraform-provider-vcd/issues/343))
* `resource/vcd_vapp_vm` has new field `computer_name` ([#334](https://github.com/vmware/terraform-provider-vcd/issues/334), [#347](https://github.com/vmware/terraform-provider-vcd/issues/347))


BUG FIXES:

* Change default value for `vcd_org.deployed_vm_quota` and `vcd_org.stored_vm_quota`. It was incorrectly set at `-1` instead of `0`.
* Change Org ID from partial task ID to real Org ID during creation.
* Wait for task completion on creation and update, where tasks were not handled at all.
* `resource/vcd_firewall_rules` force recreation of the resource when attributes of the sub-element `rule` are changed (fixes a situation when it tried to update a rule).
* `resource/vcd_network_isolated` Fix definition of DHCP, which was created automatically with leftovers from static IP pool even when not requested.
* `resource/vcd_network_routed` Fix retrieval with early vCD versions. ([#344](https://github.com/vmware/terraform-provider-vcd/issues/344))
* `resource/vcd_edgegateway_vpn` Required replacement every time for `shared_secret` field.  ([#361](https://github.com/vmware/terraform-provider-vcd/issues/361))

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

* **New Resource:** `vcd_lb_service_monitor`  Load Balancer Service Monitor - ([#256](https://github.com/vmware/terraform-provider-vcd/issues/256), [#290](https://github.com/vmware/terraform-provider-vcd/issues/290))
* **New Resource:** `vcd_edgegateway` creates and deletes edge gateways, manages general load balancing settings - ([#262](https://github.com/vmware/terraform-provider-vcd/issues/262), [#288](https://github.com/vmware/terraform-provider-vcd/issues/288))
* **New Resource:** `vcd_lb_server_pool` Load Balancer Server Pool - ([#268](https://github.com/vmware/terraform-provider-vcd/issues/268), [#290](https://github.com/vmware/terraform-provider-vcd/issues/290), [#297](https://github.com/vmware/terraform-provider-vcd/issues/297))
* **New Resource:** `vcd_lb_app_profile` Load Balancer Application profile - ([#274](https://github.com/vmware/terraform-provider-vcd/issues/274), [#290](https://github.com/vmware/terraform-provider-vcd/issues/290), [#297](https://github.com/vmware/terraform-provider-vcd/issues/297))
* **New Resource:** `vcd_lb_app_rule` Load Balancer Application rule - ([#278](https://github.com/vmware/terraform-provider-vcd/issues/278), [#290](https://github.com/vmware/terraform-provider-vcd/issues/290))
* **New Resource:** `vcd_lb_virtual_server` Load Balancer Virtual Server - ([#284](https://github.com/vmware/terraform-provider-vcd/issues/284), [#290](https://github.com/vmware/terraform-provider-vcd/issues/290), [#297](https://github.com/vmware/terraform-provider-vcd/issues/297))
* **New Resource:** `vcd_org_user`  Organization User - ([#279](https://github.com/vmware/terraform-provider-vcd/issues/279))
* **New Data Source:** `vcd_lb_service_monitor` Load Balancer Service Monitor  - ([#256](https://github.com/vmware/terraform-provider-vcd/issues/256), [#290](https://github.com/vmware/terraform-provider-vcd/issues/290))
* **New Data Source:** `vcd_lb_server_pool`  Load Balancer Server Pool - ([#268](https://github.com/vmware/terraform-provider-vcd/issues/268), [#290](https://github.com/vmware/terraform-provider-vcd/issues/290), [#297](https://github.com/vmware/terraform-provider-vcd/issues/297))
* **New Data Source:** `vcd_lb_app_profile` Load Balancer Application profile - ([#274](https://github.com/vmware/terraform-provider-vcd/issues/274), [#290](https://github.com/vmware/terraform-provider-vcd/issues/290), [#297](https://github.com/vmware/terraform-provider-vcd/issues/297))
* **New Data Source:** `vcd_lb_app_rule` Load Balancer Application rule - ([#278](https://github.com/vmware/terraform-provider-vcd/issues/278), [#290](https://github.com/vmware/terraform-provider-vcd/issues/290))
* **New Data Source:** `vcd_lb_virtual_server` Load Balancer Virtual Server - ([#284](https://github.com/vmware/terraform-provider-vcd/issues/284), [#290](https://github.com/vmware/terraform-provider-vcd/issues/290), [#297](https://github.com/vmware/terraform-provider-vcd/issues/297))
* **New build commands** `make test-env-init` and `make test-env-apply` can configure an empty vCD to run the test suite. See `TESTING.md` for details.
* `resource/vcd_org_vdc` added Org VDC update and full state read - ([#275](https://github.com/vmware/terraform-provider-vcd/issues/275))
* `resource/vcd_org_vdc` added Org VDC metadata support - ([#276](https://github.com/vmware/terraform-provider-vcd/issues/276))
* `resource/vcd_snat` added ability to choose network name and type. [[#282](https://github.com/vmware/terraform-provider-vcd/issues/282)] 
* `resource/vcd_dnat` added ability to choose network name and type. ([#282](https://github.com/vmware/terraform-provider-vcd/issues/282), [#292](https://github.com/vmware/terraform-provider-vcd/issues/292), [#293](https://github.com/vmware/terraform-provider-vcd/issues/293))

IMPROVEMENTS:

* `resource/vcd_org_vdc`: Fix ignoring of resource guarantee values - ([#265](https://github.com/vmware/terraform-provider-vcd/issues/265))
* `resource/vcd_org_vdc`: Org VDC state ID changed from name to vCD ID - ([#275](https://github.com/vmware/terraform-provider-vcd/issues/275))
* Change resource handling to use locking mechanism when resource parallel handling is not supported by vCD. [[#255](https://github.com/vmware/terraform-provider-vcd/issues/255)] 
* Fix issue when vApp is power cycled during member VM deletion. ([#261](https://github.com/vmware/terraform-provider-vcd/issues/261))
* `resource/vcd_dnat`, `resource/vcd_snat` has got full read functionality. This means that on the next `plan/apply` it will detect if configuration has changed in vCD and propose to update it.

BUG FIXES:

* `resource/vcd_dnat and resource/vcd_snat` - fix resource destroy as it would still leave NAT rule in edge gateway. Fix works if network_name and network_type is used. ([#282](https://github.com/vmware/terraform-provider-vcd/issues/282))

NOTES:

* `resource/vcd_dnat` `protocol` requires lower case values to be consistent with the underlying NSX API. This may result in invalid configuration if upper case was used previously!
* `resource/vcd_dnat` default value for `protocol` field changed from upper case `TCP` to lower case `tcp`, which may result in a single update when running `plan` on a configuration with a state file from an older version.

## 2.3.0 (May 29, 2019)

IMPROVEMENTS:

* Switch to Terraform 0.12 SDK which is required for Terraform 0.12 support. HCL (HashiCorp configuration language) 
parsing behaviour may have changed as a result of changes made by the new SDK version ([#254](https://github.com/vmware/terraform-provider-vcd/issues/254))

NOTES:

* Provider plugin will still work with Terraform 0.11 executable

## 2.2.0 (May 16, 2019)

FEATURES:

* `vcd_vapp_vm` - Ability to add metadata to a VM. For previous behaviour please see `BACKWARDS INCOMPATIBILITIES` ([#158](https://github.com/vmware/terraform-provider-vcd/issues/158))
* `vcd_vapp_vm` - Ability to enable hardware assisted CPU virtualization for VM. It allows hypervisor nesting. ([#219](https://github.com/vmware/terraform-provider-vcd/issues/219))
* **New Resource:** external network - `vcd_external_network` - ([#230](https://github.com/vmware/terraform-provider-vcd/issues/230))
* **New Resource:** VDC resource `vcd_org_vdc` - ([#236](https://github.com/vmware/terraform-provider-vcd/issues/236))
* resource/vcd_vapp_vm: Add `network` argument for multiple NIC support and more flexible configuration ([#233](https://github.com/vmware/terraform-provider-vcd/issues/233))
* resource/vcd_vapp_vm: Add `mac` argument to store MAC address in state file ([#233](https://github.com/vmware/terraform-provider-vcd/issues/233))

BUG FIXES:

* `vcd_vapp` - Ability to add metadata to empty vApp. For previous behaviour please see `BACKWARDS INCOMPATIBILITIES` ([#158](https://github.com/vmware/terraform-provider-vcd/issues/158))

BACKWARDS INCOMPATIBILITIES / NOTES:

* `vcd_vapp` - Metadata is no longer added to first VM in vApp it will be added to vApp directly instead. ([#158](https://github.com/vmware/terraform-provider-vcd/issues/158))
* Tests files are now all tagged. Running them through Makefile works as before, but manual execution requires specific tags. Run `go test -v .` for tags list.
* `vcd_vapp_vm` - Deprecated attributes `network_name`, `vapp_network_name`, `network_href` and `ip` in favor of `network` ([#118](https://github.com/vmware/terraform-provider-vcd/issues/118))

## 2.1.0 (March 27, 2019)

NOTES:
* Please look for "v2.1+" keyword in documentation which is used to emphasize new features.
* Project switched to using Go modules, while `vendor` is left for backwards build compatibility only. It is worth having a
look at [README.md](README.md) to understand how Go modules impact build and development ([#178](https://github.com/vmware/terraform-provider-vcd/issues/178))
* Project dependency of github.com/hashicorp/terraform updated from v0.10.0 to v0.11.13 ([#181](https://github.com/vmware/terraform-provider-vcd/issues/181))
* MaxRetryTimeout is shared with underlying SDK `go-vcloud-director` ([#189](https://github.com/vmware/terraform-provider-vcd/issues/189))
* Improved testing functionality ([#166](https://github.com/vmware/terraform-provider-vcd/issues/166))

FEATURES:

* **New Resource:** disk resource - `vcd_independent_disk` ([#188](https://github.com/vmware/terraform-provider-vcd/issues/188))
* resource/vcd_vapp_vm has ability to attach independent disk ([#188](https://github.com/vmware/terraform-provider-vcd/issues/188))
* **New Resource:** vApp network - `vcd_vapp_network` ([#155](https://github.com/vmware/terraform-provider-vcd/issues/155))
* resource/vcd_vapp_vm has ability to use vApp network ([#155](https://github.com/vmware/terraform-provider-vcd/issues/155))

IMPROVEMENTS:

* resource/vcd_inserted_media now supports force ejecting on running VM ([#184](https://github.com/vmware/terraform-provider-vcd/issues/184))
* resource/vcd_vapp_vm now support CPU cores configuration ([#174](https://github.com/vmware/terraform-provider-vcd/issues/174))

BUG FIXES:

* resource/vcd_vapp, resource/vcd_vapp_vm add vApp status handling when environment is very fast ([#68](https://github.com/vmware/terraform-provider-vcd/issues/68))
* resource/vcd_vapp_vm add additional validation to check if vApp template is OK [[#157](https://github.com/vmware/terraform-provider-vcd/issues/157)] 

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

* `vcd_vapp` - Fixes an issue with Networks in vApp templates being required, also introduced in 0.1.2 ([#38](https://github.com/vmware/terraform-provider-vcd/issues/38))

FEATURES:

* `vcd_vapp` - Add support for defining shared vcd_networks ([#46](https://github.com/vmware/terraform-provider-vcd/pull/46))
* `vcd_vapp` - Added options to configure dhcp lease times ([#47](https://github.com/vmware/terraform-provider-vcd/pull/47))

## 1.0.0 (August 17, 2017)

IMPROVEMENTS:

* `vcd_vapp` - Fixes an issue with storage profiles introduced in 0.1.2 ([#39](https://github.com/vmware/terraform-provider-vcd/issues/39))

BACKWARDS INCOMPATIBILITIES / NOTES:

* provider: Deprecate `maxRetryTimeout` in favour of `max_retry_timeout` ([#40](https://github.com/vmware/terraform-provider-vcd/issues/40))

## 0.1.3 (August 09, 2017)

IMPROVEMENTS:

* `vcd_vapp` - Setting the computer name regardless of init script ([#31](https://github.com/vmware/terraform-provider-vcd/issues/31))
* `vcd_vapp` - Fixes the power_on issue introduced in 0.1.2 ([#33](https://github.com/vmware/terraform-provider-vcd/issues/33))
* `vcd_vapp` - Fixes issue with allocated IP address affecting tf plan ([#17](https://github.com/vmware/terraform-provider-vcd/issues/17) & [#29](https://github.com/vmware/terraform-provider-vcd/issues/29))
* `vcd_vapp_vm` - Setting the computer name regardless of init script ([#31](https://github.com/vmware/terraform-provider-vcd/issues/31))
* `vcd_firewall_rules` - corrected typo in docs ([#35](https://github.com/vmware/terraform-provider-vcd/issues/35))

## 0.1.2 (August 03, 2017)

IMPROVEMENTS:

* Possibility to add OVF parameters to a vApp ([#1](https://github.com/vmware/terraform-provider-vcd/pull/1))
* Added storage profile support  ([#23](https://github.com/vmware/terraform-provider-vcd/pull/23))

## 0.1.1 (June 28, 2017)

FEATURES:

* **New VM Resource**: `vcd_vapp_vm` ([#9](https://github.com/vmware/terraform-provider-vcd/issues/9))
* **New VPN Resource**: `vcd_edgegateway_vpn`

IMPROVEMENTS:

* resource/vcd_dnat: Added a new (optional) param translated_port ([#14](https://github.com/vmware/terraform-provider-vcd/issues/14))

## 0.1.0 (June 20, 2017)

NOTES:

* Same functionality as that of Terraform 0.9.8. Repacked as part of [Provider Splitout](https://www.hashicorp.com/blog/upcoming-provider-changes-in-terraform-0-10/)
