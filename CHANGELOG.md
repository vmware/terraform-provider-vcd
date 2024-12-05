## 3.14.1 (December 5, 2024)

### IMPROVEMENTS
* Add support of CSE 4.2.2 and 4.2.3 versions by improving `vcd_cse_kubernetes cluster` and updating
  the installation guide and examples ([#1357](https://github.com/vmware/terraform-provider-vcd/pull/1357))

### NOTES
* Bump [`go-vcloud-director`](https://github.com/vmware/go-vcloud-director/tree/release/v2.x) to v2.26.1
(SDK this provider uses for low level access to the VCD) ([#1363](https://github.com/vmware/terraform-provider-vcd/pull/1363))

## 3.14.0 (September 17, 2024)

### FEATURES
* **New Resource:** `vcd_external_endpoint` to manage External Endpoints ([#1295](https://github.com/vmware/terraform-provider-vcd/pull/1295), [#1322](https://github.com/vmware/terraform-provider-vcd/pull/1322))
* **New Data Source:** `vcd_external_endpoint` to read External Endpoints ([#1295](https://github.com/vmware/terraform-provider-vcd/pull/1295), [#1322](https://github.com/vmware/terraform-provider-vcd/pull/1322))
* **New Resource:** `vcd_api_filter` to manage API Filters ([#1295](https://github.com/vmware/terraform-provider-vcd/pull/1295), [#1322](https://github.com/vmware/terraform-provider-vcd/pull/1322))
* **New Data Source:** `vcd_api_filter` to read API Filters ([#1295](https://github.com/vmware/terraform-provider-vcd/pull/1295), [#1322](https://github.com/vmware/terraform-provider-vcd/pull/1322))
* **New Data Source:** `vcd_nsxt_tier0_router_interface` to read Tier-0 Router Interfaces that can
  be assigned to IP Space Uplinks ([#1311](https://github.com/vmware/terraform-provider-vcd/pull/1311))
* **New Data Source:** `vcd_catalog_access_control` to read Catalog access controls ([#1315](https://github.com/vmware/terraform-provider-vcd/pull/1315))
* **New Resource:** `vcd_nsxt_alb_virtual_service_http_req_rules` to manage ALB Virtual Service Request Rules ([#1320](https://github.com/vmware/terraform-provider-vcd/pull/1320))
* **New Data Source:** `vcd_nsxt_alb_virtual_service_http_req_rules` to read ALB Virtual Service Request Rules ([#1320](https://github.com/vmware/terraform-provider-vcd/pull/1320))
* **New Resource:** `vcd_nsxt_alb_virtual_service_http_resp_rules` to manage ALB Virtual Service Response Rules ([#1320](https://github.com/vmware/terraform-provider-vcd/pull/1320))
* **New Data Source:** `vcd_nsxt_alb_virtual_service_http_resp_rules` to read ALB Virtual Service Response Rules ([#1320](https://github.com/vmware/terraform-provider-vcd/pull/1320))
* **New Resource:** `vcd_nsxt_alb_virtual_service_http_sec_rules` to manage ALB Virtual Service Security Rules ([#1320](https://github.com/vmware/terraform-provider-vcd/pull/1320))
* **New Data Source:** `vcd_nsxt_alb_virtual_service_http_sec_rules` to read ALB Virtual Service Security Rules ([#1320](https://github.com/vmware/terraform-provider-vcd/pull/1320))

### IMPROVEMENTS
* Add new argument `execution_json` to resources and data sources `vcd_rde_interface_behavior` and `vcd_rde_type_behavior`
  to define complex behavior executions that could not be specified with the `execution` map ([#1131](https://github.com/vmware/terraform-provider-vcd/pull/1131))
* Add new argument `arguments_json` and `metadata_json` to data source `vcd_rde_behavior_invocation` to be able to
  invoke behaviors that have complex execution definitions ([#1131](https://github.com/vmware/terraform-provider-vcd/pull/1131))
* Resources and data sources `vcd_vapp_vm` and `vcd_vm` support IPv6 via `secondary_ip_allocation_mode`
  and `secondary_ip` fields ([#1292](https://github.com/vmware/terraform-provider-vcd/pull/1292))
* Add `vcd_nsxt_alb_edgegateway_service_engine_group` and `vcd_nsxt_alb_service_engine_group` resource types to
  `vcd_resource_list` data source ([#1296](https://github.com/vmware/terraform-provider-vcd/pull/1296), [#1322](https://github.com/vmware/terraform-provider-vcd/pull/1322))
* Add support for NSX-T Non-distributed Org VDC networks via flags `non_distributed_routing_enabled`
  in `vcd_nsxt_edgegateway` and `interface_type=non_distributed` in `vcd_network_routed_v2`
  ([#1297](https://github.com/vmware/terraform-provider-vcd/pull/1297), [#1322](https://github.com/vmware/terraform-provider-vcd/pull/1322))
* Add provider option `saml_adfs_cookie` that can help lookup of ADFS server ([#1298](https://github.com/vmware/terraform-provider-vcd/pull/1298))
* Use Bearer token in SAML ADFS auth flow instead the old one ([#1298](https://github.com/vmware/terraform-provider-vcd/pull/1298))
* Resource `vcd_nsxt_edgegateway` adds support for Distributed-only `deployment_mode` ([#1300](https://github.com/vmware/terraform-provider-vcd/pull/1300))
* Add `account_lockout` block to `vcd_org` resource and data source, that specifies the account locking mechanism with the
  sub-arguments `enabled`, `invalid_logins_before_lockout` and `lockout_interval_minutes` ([#1304](https://github.com/vmware/terraform-provider-vcd/pull/1304))
* Add IP Space locks for `vcd_ip_space_ip_allocation` to prevent concurrent modification error in
  API ([#1305](https://github.com/vmware/terraform-provider-vcd/pull/1305))
* Resource `vcd_ip_space_uplink` adds support for managing IP Space Uplink Associated Interfaces via
  `associated_interface_ids` and new data source `vcd_nsxt_tier0_interface` ([#1311](https://github.com/vmware/terraform-provider-vcd/pull/1311))

### BUG FIXES
* Fix [Issue 1287](https://github.com/vmware/terraform-provider-vcd/issues/1287) Read-only org sharing prevents sharing to users ([#1291](https://github.com/vmware/terraform-provider-vcd/pull/1291))
* Fix [Issue 1183](https://github.com/vmware/terraform-provider-vcd/issues/1183) where updates might
  fail for vcd_external_network_v2 when NSX-T edge gateway has `dedicate_external_network=true`
  ([#1301](https://github.com/vmware/terraform-provider-vcd/pull/1301))
* Fix [Issue 1236](https://github.com/vmware/terraform-provider-vcd/issues/1236):
  `list_mode="import"` of data source `vcd_resource_list` created wrong import statements when VCD items names have special
  characters ([#1302](https://github.com/vmware/terraform-provider-vcd/pull/1302))
* Fix [Issue 1236](https://github.com/vmware/terraform-provider-vcd/issues/1236):
  `list_mode="hierarchy"` of data source `vcd_resource_list` repeated the parent element twice when obtaining the hierarchy ([#1302](https://github.com/vmware/terraform-provider-vcd/pull/1302))
* Fix [Issue 1243](https://github.com/vmware/terraform-provider-vcd/issues/1243):
  Allow unlimited `limit_in_mhz` in `vcd_vm_sizing_policy` resource ([#1303](https://github.com/vmware/terraform-provider-vcd/pull/1303), [#1318](https://github.com/vmware/terraform-provider-vcd/pull/1318))
* Fix [Issue 1307](https://github.com/vmware/terraform-provider-vcd/issues/1307) in `vcd_vapp_vm`
  and `vcd_vm` resources where `firmware=efi` field wouldn't be applied for template based
  VMs with `firmware=bios` on creation ([#1308](https://github.com/vmware/terraform-provider-vcd/pull/1308))
* Fix [Issue 1262](https://github.com/vmware/terraform-provider-vcd/issues/1262): Panic when creating a VM with `vcd_vm` and `vcd_vapp_vm`
  when the VCD provider is configured with a user without "Organization vDC Disk: View/Edit IOPS" rights ([#1312](https://github.com/vmware/terraform-provider-vcd/pull/1312))
* Fix [Issue 1216](https://github.com/vmware/terraform-provider-vcd/issues/1216) in `vcd_org_vdc`
  which failed on creation when `vm_placement_policy_ids` were set but `default_compute_policy_id`
  was not declared (System default was used instead) ([#1313](https://github.com/vmware/terraform-provider-vcd/pull/1313))
* Fix [Issue 1205](https://github.com/vmware/terraform-provider-vcd/issues/1205) in `vcd_vapp_vm`
  and `vcd_vm` resources where not setting `ip_allocation_mode` in a `network` block would cause a 500 error ([#1317](https://github.com/vmware/terraform-provider-vcd/pull/1317))

### DEPRECATIONS
* Data source `vcd_rde_behavior_invocation` deprecate `arguments` and `metadata` arguments in favor of `arguments_json` and
  `metadata_json`, that allow to invoke behaviors with complex execution definitions ([#1131](https://github.com/vmware/terraform-provider-vcd/pull/1131))
* Data source `vcd_rde_interface_behavior` deprecate `execution` in favor of `execution_json`, which allow to read
  complex execution definitions from an existing behavior ([#1131](https://github.com/vmware/terraform-provider-vcd/pull/1131))
* Data source `vcd_rde_type_behavior` deprecate `execution` in favor of `execution_json`, which allow to read
  complex execution definitions from an existing behavior ([#1131](https://github.com/vmware/terraform-provider-vcd/pull/1131))

### NOTES
* Bump [`go-vcloud-director`](https://github.com/vmware/go-vcloud-director) to v2.26.0 (SDK this provider uses for low level access to the VCD) ([#1325](https://github.com/vmware/terraform-provider-vcd/pull/1325))
* Reduce memory usage of test `TestAccVcdStandaloneVmWithVmSizing` to avoid errors on tiny VCD testing setups ([#1306](https://github.com/vmware/terraform-provider-vcd/pull/1306))
* Correct a malformed HCL snippet in the *"Connecting with a Service Account API token"* section of the documentation ([#1322](https://github.com/vmware/terraform-provider-vcd/pull/1322))

## 3.13.0 (July 2, 2024)

### FEATURES
* Add support for **VCD 10.6** ([#1279](https://github.com/vmware/terraform-provider-vcd/pull/1279))
* **New Guide** `Data Solution Extension and Solution Add-On management` ([#1286](https://github.com/vmware/terraform-provider-vcd/pull/1286))
* **New Guide** `Site and Org associations` to describe association operations for sites and organizations ([#1260](https://github.com/vmware/terraform-provider-vcd/pull/1260))
* **New Resource:** `vcd_solution_landing_zone` to manage Solution Add-On Landing Zone ([#1251](https://github.com/vmware/terraform-provider-vcd/pull/1251))
* **New Data Source:** `vcd_solution_landing_zone` to read Solution Add-On Landing Zone ([#1251](https://github.com/vmware/terraform-provider-vcd/pull/1251))
* **New Resource:** `vcd_solution_add_on` to manage Solution Add-Ons ([#1256](https://github.com/vmware/terraform-provider-vcd/pull/1256))
* **New Data Source:** `vcd_solution_add_on` to read Solution Add-Ons ([#1256](https://github.com/vmware/terraform-provider-vcd/pull/1256))
* **New Resource:** `vcd_solution_add_on_instance` to manage Solution Add-On Instances ([#1272](https://github.com/vmware/terraform-provider-vcd/pull/1272))
* **New Data Source:** `vcd_solution_add_on_instance` to read existing Solution Add-On Instances
  ([#1272](https://github.com/vmware/terraform-provider-vcd/pull/1272))
* **New Resource:** `vcd_solution_add_on_instance_publish` to manage publishing settings for
  Solution Add-On Instances  ([#1272](https://github.com/vmware/terraform-provider-vcd/pull/1272))
* **New Data Source:** `vcd_solution_add_on_instance_publish` to read publishing settings for
  Solution Add-On Instances ([#1272](https://github.com/vmware/terraform-provider-vcd/pull/1272))
* **New Resource:** `vcd_dse_registry_configuration` to manage Data Solution Extension (DSE)
  Registry Configuration ([#1284](https://github.com/vmware/terraform-provider-vcd/pull/1284),[#1286](https://github.com/vmware/terraform-provider-vcd/pull/1286))
* **New Data Source:** `vcd_dse_registry_configuration` to read Data Solution Extension (DSE)
  Registry Configuration ([#1284](https://github.com/vmware/terraform-provider-vcd/pull/1284),[#1286](https://github.com/vmware/terraform-provider-vcd/pull/1286))
* **New Resource:** `vcd_dse_solution_publish` to manage DSE Solution publishing ([#1284](https://github.com/vmware/terraform-provider-vcd/pull/1284))
* **New Data Source:** `vcd_dse_solution_publish` to read DSE Solution publishing ([#1284](https://github.com/vmware/terraform-provider-vcd/pull/1284))
* **New Data Source:** `vcd_multisite_site` to read the state and associations of current site ([#1260](https://github.com/vmware/terraform-provider-vcd/pull/1260))
* **New Data Source:** `vcd_multisite_site_data` to produce the association data needed to start a site association ([#1260](https://github.com/vmware/terraform-provider-vcd/pull/1260))
* **New Data Source:** `vcd_multisite_site_association` to read the details of a site association ([#1260](https://github.com/vmware/terraform-provider-vcd/pull/1260))
* **New Resource:** `vcd_multisite_site_association` to associate the current site with a remote one ([#1260](https://github.com/vmware/terraform-provider-vcd/pull/1260))
* **New Data Source:** `vcd_multisite_org_data` to produce the association data needed to start an organization association ([#1260](https://github.com/vmware/terraform-provider-vcd/pull/1260))
* **New Data Source:** `vcd_multisite_org_association` to read the details of an organization association ([#1260](https://github.com/vmware/terraform-provider-vcd/pull/1260))
* **New Resource:** `vcd_multisite_org_association` to associate a local organization with a remote one ([#1260](https://github.com/vmware/terraform-provider-vcd/pull/1260))
* **New Resource:** `vcd_org_oidc` to manage the Open ID Connect settings for an Organization ([#1263](https://github.com/vmware/terraform-provider-vcd/pull/1263))
* **New Data Source:** `vcd_org_oidc` to read the Open ID Connect settings from an Organization ([#1263](https://github.com/vmware/terraform-provider-vcd/pull/1263))
* **New Resource:** `vcd_org_vdc_template` to manage VDC Templates ([#1276](https://github.com/vmware/terraform-provider-vcd/pull/1276), [#1280](https://github.com/vmware/terraform-provider-vcd/pull/1280))
* **New Data Source:** `vcd_org_vdc_template` to read VDC Templates ([#1276](https://github.com/vmware/terraform-provider-vcd/pull/1276))
* **New Resource:** `vcd_org_vdc_template_instance` to instantiate VDC Templates ([#1280](https://github.com/vmware/terraform-provider-vcd/pull/1280))

### IMPROVEMENTS
* Resource and data source `vcd_vapp` add fields `vm_names`, `vapp_network_names`, `vapp_org_network_names` to list VMs and vApp networks. ([#1235](https://github.com/vmware/terraform-provider-vcd/pull/1235))
* Data source `vcd_resource_list` adds ability to list `vcd_vapp_network`, `vcd_vapp_org_network`, `vcd_vapp_all_network` to list vApp networks ([#1235](https://github.com/vmware/terraform-provider-vcd/pull/1235))
* Resource and data source `vcd_external_network_v2` add support for Provider Gateway Topology
  intentions in VCD 10.5.1+ via fields `nat_and_firewall_service_intention` and
  `route_advertisement_intention` ([#1239](https://github.com/vmware/terraform-provider-vcd/pull/1239))
* Resource `vcd_nsxt_firewall` supports `REJECT` action ([#1240](https://github.com/vmware/terraform-provider-vcd/pull/1240))
* Resources `vcd_vapp_vm` and `vcd_vm` add property `set_extra_config` to add, modify, or remove VM extra configuration items ([#1253](https://github.com/vmware/terraform-provider-vcd/pull/1253), [#1288](https://github.com/vmware/terraform-provider-vcd/pull/1288))
* Resources and data sources `vcd_vapp_vm` and `vcd_vm` add property `extra_config` to read existing VM extra configuration ([#1253](https://github.com/vmware/terraform-provider-vcd/pull/1253), [#1288](https://github.com/vmware/terraform-provider-vcd/pull/1288))
* Resource and data source `vcd_catalog_media` exposed additional attribute `catalog_item_id` to
  expose catalog item ID ([#1256](https://github.com/vmware/terraform-provider-vcd/pull/1256))
* Data source `vcd_resource_list` can now list site and organization associations ([#1260](https://github.com/vmware/terraform-provider-vcd/pull/1260))
* The `worker_pool` block from `vcd_cse_kubernetes_cluster` resource allows to configure the
  [cluster autoscaler](https://github.com/kubernetes/autoscaler/tree/master/cluster-autoscaler) with
  `autoscaler_max_replicas` and `autoscaler_min_replicas` arguments ([#1269](https://github.com/vmware/terraform-provider-vcd/pull/1269))
* Data source `vcd_resource_list` can list VDC Templates ([#1276](https://github.com/vmware/terraform-provider-vcd/pull/1276))
* Improve `rde_type_behavior_acl` documentation to state that redundant access levels should be avoided, especially
  in VCD 10.6+ to prevent undesired updates-in-place during plans ([#1277](https://github.com/vmware/terraform-provider-vcd/pull/1277))

### BUG FIXES
* Fix a missing Edge Gateway filter bug in `vcd_nsxt_alb_edgegateway_service_engine_group` resource
  (issue [#1245](https://github.com/vmware/terraform-provider-vcd/issues/1245)) ([#1246](https://github.com/vmware/terraform-provider-vcd/pull/1246))
* Fix [Issue #1258](https://github.com/vmware/terraform-provider-vcd/issues/1258): `vcd_cse_kubernetes_cluster` fails
during creation when the chosen network belongs to a VDC Group ([#1266](https://github.com/vmware/terraform-provider-vcd/pull/1266))
* Fix [Issue #1265](https://github.com/vmware/terraform-provider-vcd/issues/1265): The `kubeconfig` attribute from
  `vcd_cse_kubernetes_cluster` resource and data source is now marked as sensitive ([#1266](https://github.com/vmware/terraform-provider-vcd/pull/1266))
* Fix a bug where `vcd_nsxt_distributed_firewall_rule` resource could report incorrect firewall rule
  when using `above_rule_id` (issue
  [#1202](https://github.com/vmware/terraform-provider-vcd/issues/1202)) ([#1268](https://github.com/vmware/terraform-provider-vcd/pull/1268))
* Fix `vcd_catalog_media` resource so it doesn't wait indefinitely to the upload task to reach 100% progress,
  by checking also its status, to decide that the upload is complete or aborted ([#1273](https://github.com/vmware/terraform-provider-vcd/pull/1273))
* Fix Issue [1170](https://github.com/vmware/terraform-provider-vcd/issues/1170) where an imported VM complains about missing or altered fields and Terraform tries to re-create the resource ([#1274](https://github.com/vmware/terraform-provider-vcd/pull/1274))
* Fix [Issue #1202](https://github.com/vmware/terraform-provider-vcd/issues/1270) - Resource and
  data source `vcd_nsxt_edgegateway` may crash due to exhausting memory while counting huge IPv6
  subnets by adding count limit defined in`ip_count_read_limit` field ([#1275](https://github.com/vmware/terraform-provider-vcd/pull/1275))
* Fix `vcd_nsxt_ipsec_vpn_tunnel` update operations, that failed in VCD 10.6+ when a `security_profile_customization` block is added
  to the IPSec VPN tunnel ([#1282](https://github.com/vmware/terraform-provider-vcd/pull/1282))
* Fix resource `vcd_nsxt_alb_settings` so update operations don't fail in VCD 10.6+ ([#1283](https://github.com/vmware/terraform-provider-vcd/pull/1283))

### NOTES
* Bump [`go-vcloud-director`](https://github.com/vmware/go-vcloud-director) to v2.25.0 (SDK this provider uses for low level access to the VCD) ([#1289](https://github.com/vmware/terraform-provider-vcd/pull/1289))
* Bump `terraform-plugin-sdk` to v2.34.0 ([#1271](https://github.com/vmware/terraform-provider-vcd/pull/1271))
* Amend the test `TestAccVcdRdeDuplicate` so it doesn't fail on VCD 10.6+. Since this version, whenever a RDE is created
  in a tenant by the System Administrator, the owner is not `"administrator"` anymore, but `"system"` ([#1278](https://github.com/vmware/terraform-provider-vcd/pull/1278))
* Tests for FLEX Org VDC must set `memory_guaranteed` when `include_vm_memory_overhead=true`
  ([#1281](https://github.com/vmware/terraform-provider-vcd/pull/1281))
* Amend `TestAccVcdOrgOidc` to check the `redirect_uri` in a case-insensitive way ([#1282](https://github.com/vmware/terraform-provider-vcd/pull/1282))
* Amend `TestAccVcdCatalogSharedAccess`, it failed in VCD 10.6+ as the used VDC was missing the
  `ResourceGuaranteedMemory` parameter (Flex allocation model) ([#1283](https://github.com/vmware/terraform-provider-vcd/pull/1283))
* Amend test `TestAccVcdSubscribedCatalog` to be compatible with VCD 10.6.0 ([#1285](https://github.com/vmware/terraform-provider-vcd/pull/1285))
* Update `vcd_subscribed_catalog` resource documentation to state that `metadata` attribute is not available in
VCD 10.6.0 ([#1285](https://github.com/vmware/terraform-provider-vcd/pull/1285))

## 3.12.1 (April 19, 2024)

### IMPROVEMENTS
* Improve page links for authentication methods in main documentation page ([#1241](https://github.com/vmware/terraform-provider-vcd/pull/1241))
* Rename VMware NSX Advanced Load Balancer (Avi) to VMware Avi Load Balancer ([#1241](https://github.com/vmware/terraform-provider-vcd/pull/1241))

### BUG FIXES
* Fix [Issue #1242](https://github.com/vmware/terraform-provider-vcd/issues/1242): panic when edge gateway IP count returns empty ([#1244](https://github.com/vmware/terraform-provider-vcd/pull/1244))
* Fix [Issue #1248](https://github.com/vmware/terraform-provider-vcd/issues/1248) that prevents CSE Kubernetes clusters from being upgraded to an OVA with higher Kubernetes version but same TKG version, and to an OVA with a higher patch version of Kubernetes ([#1247](https://github.com/vmware/terraform-provider-vcd/pull/1247))
* Fix [Issue #1248](https://github.com/vmware/terraform-provider-vcd/issues/1248) that prevents CSE Kubernetes clusters from being upgraded to TKG v2.5.0 with Kubernetes v1.26.11 as it performed an invalid upgrade of CoreDNS ([#1247](https://github.com/vmware/terraform-provider-vcd/pull/1247))
* Fix [Issue #1252](https://github.com/vmware/terraform-provider-vcd/issues/1248) that prevents reading the SSH Public Key from provisioned CSE Kubernetes clusters ([#1247](https://github.com/vmware/terraform-provider-vcd/pull/1247))

### NOTES

* Bump [`go-vcloud-director`](https://github.com/vmware/go-vcloud-director) to v2.24.0 (SDK this provider uses for low level access to the VCD) ([#1247](https://github.com/vmware/terraform-provider-vcd/pull/1247))

## 3.12.0 (March 22, 2024)

### FEATURES
* **New Resource:** `vcd_cse_kubernetes_cluster` to create and manage Kubernetes clusters in a VCD with Container Service Extension
  4.2.1, 4.2.0, 4.1.1 or 4.1.0 installed and running ([#1195](https://github.com/vmware/terraform-provider-vcd/pull/1195), [#1218](https://github.com/vmware/terraform-provider-vcd/pull/1218), [#1222](https://github.com/vmware/terraform-provider-vcd/pull/1222))
* **New Data Source:** `vcd_cse_kubernetes_cluster` to read Kubernetes clusters from a VCD with Container Service Extension
  4.2.1, 4.2.0, 4.1.1 or 4.1.0 installed and running ([#1195](https://github.com/vmware/terraform-provider-vcd/pull/1195), [#1218](https://github.com/vmware/terraform-provider-vcd/pull/1218), [#1222](https://github.com/vmware/terraform-provider-vcd/pull/1222))
* **New Data Source:** `vcd_version` to get the VCD version and perform additional checks with version constraints ([#1195](https://github.com/vmware/terraform-provider-vcd/pull/1195), [#1218](https://github.com/vmware/terraform-provider-vcd/pull/1218))

### IMPROVEMENTS
* Resource `vcd_provider_vdc` adds ability of creating a provider VDC without network pool or NSX-T manager (issue [#1186](https://github.com/vmware/terraform-provider-vcd/issues/1186)) ([#1220](https://github.com/vmware/terraform-provider-vcd/pull/1220))
* Add route advertisement support to `vcd_network_routed_v2` via field `route_advertisement_enabled`
  ([#1203](https://github.com/vmware/terraform-provider-vcd/pull/1203))
* `vcd_vapp_vm` and `vcd_vm` add field `consolidate_disks_on_create` that  helps to change template
  disk sizes using `override_template_disk` in fast provisioned VDCs ([#1206](https://github.com/vmware/terraform-provider-vcd/pull/1206))
* `vcd_vapp_vm` and `vcd_vm` add support for VM Copy operation by using `copy_from_vm_id` field
  ([#1210](https://github.com/vmware/terraform-provider-vcd/pull/1210), [#1218](https://github.com/vmware/terraform-provider-vcd/pull/1218), [#1219](https://github.com/vmware/terraform-provider-vcd/pull/1219))
* Resources and data sources `vcd_vapp_vm`/`vcd_vm` expose computed field `vapp_id` ([#1215](https://github.com/vmware/terraform-provider-vcd/pull/1215))
* Resource `vcd_catalog_vapp_template` supports creating templates from existing vApps and
  standalone VMs using new `capture_vapp` configuration block ([#1215](https://github.com/vmware/terraform-provider-vcd/pull/1215))
* Resource `vcd_catalog_vapp_template` exposes attribute `catalog_item_id` for related Catalog Item
  ID ([#1215](https://github.com/vmware/terraform-provider-vcd/pull/1215), [#1219](https://github.com/vmware/terraform-provider-vcd/pull/1219))

### BUG FIXES
* Fix [Issue #1121](https://github.com/vmware/terraform-provider-vcd/issues/1221) Portgroup backed network pool can't have a data source ([#1220](https://github.com/vmware/terraform-provider-vcd/pull/1220))

### DEPRECATIONS
* Resource `vcd_cse_kubernetes_cluster` deprecates the Container Service Extension cluster management guide,
  so users should not use `vcd_rde` resources to create a Kubernetes cluster anymore ([#1195](https://github.com/vmware/terraform-provider-vcd/pull/1195))

### NOTES
* Bump terraform-plugin-sdk to v2.31.0 ([#1193](https://github.com/vmware/terraform-provider-vcd/pull/1193))
* Bump [`go-vcloud-director`](https://github.com/vmware/go-vcloud-director) to v2.23.0 (SDK this provider uses for low level access 
  to the VCD) ([#1225](https://github.com/vmware/terraform-provider-vcd/pull/1225))

## 3.11.0 (December 12, 2023)

### FEATURES
* Add support for VMware Cloud Director **10.5.1**
* Add support for **Container Service Extension v4.1** by updating both the installation and the cluster management
  guides ([#1063](https://github.com/vmware/terraform-provider-vcd/pull/1063), [#1139](https://github.com/vmware/terraform-provider-vcd/pull/1139))
* **New Resource:** `vcd_network_pool` to create and manage network pools ([#1115](https://github.com/vmware/terraform-provider-vcd/pull/1115))
* **New Data Source:** `vcd_nsxt_segment_ip_discovery_profile` to read NSX-T IP Discovery Segment Profiles ([#1120](https://github.com/vmware/terraform-provider-vcd/pull/1120))
* **New Data Source:** `vcd_nsxt_segment_mac_discovery_profile` to read NSX-T MAC Discovery Segment Profiles ([#1120](https://github.com/vmware/terraform-provider-vcd/pull/1120))
* **New Data Source:** `vcd_nsxt_segment_spoof_guard_profile` to read NSX-T Spoof Guard Profiles ([#1120](https://github.com/vmware/terraform-provider-vcd/pull/1120))
* **New Data Source:** `vcd_nsxt_segment_qos_profile` to read NSX-T QoS Profiles ([#1120](https://github.com/vmware/terraform-provider-vcd/pull/1120))
* **New Data Source:** `vcd_nsxt_segment_security_profile` to read NSX-T Segment Security Profiles ([#1120](https://github.com/vmware/terraform-provider-vcd/pull/1120))
* **New Resource:** `vcd_nsxt_segment_profile_template` to manage NSX-T Segment Profile Templates ([#1120](https://github.com/vmware/terraform-provider-vcd/pull/1120))
* **New Data Source:** `vcd_nsxt_segment_profile_template` to read NSX-T Segment Profile Templates ([#1120](https://github.com/vmware/terraform-provider-vcd/pull/1120))
* **New Resource:** `vcd_nsxt_global_default_segment_profile_template` to manage NSX-T Global Default Segment Profile Templates ([#1120](https://github.com/vmware/terraform-provider-vcd/pull/1120))
* **New Data Source:** `vcd_nsxt_global_default_segment_profile_template` to read NSX-T Global Default Segment Profile Templates ([#1120](https://github.com/vmware/terraform-provider-vcd/pull/1120))
* **New Resource:** `vcd_org_vdc_nsxt_network_profile` to manage default Segment Profile Templates for NSX-T VDCs ([#1120](https://github.com/vmware/terraform-provider-vcd/pull/1120))
* **New Data Source:** `vcd_org_vdc_nsxt_network_profile` to read default Segment Profile Templates for NSX-T VDCs ([#1120](https://github.com/vmware/terraform-provider-vcd/pull/1120))
* **New Resource:** `vcd_nsxt_network_segment_profile` to manage individual Segment Profiles or Segment Profile Templates for NSX-T Org VDC Networks ([#1120](https://github.com/vmware/terraform-provider-vcd/pull/1120))
* **New Data Source:** `vcd_nsxt_network_segment_profile` to read individual Segment Profiles or Segment Profile Templates for NSX-T Org VDC Networks ([#1120](https://github.com/vmware/terraform-provider-vcd/pull/1120))
* **New Resource:** `vcd_nsxt_edgegateway_l2_vpn_tunnel` to manage Edge Gateway L2 VPN Tunnel sessions ([#1061](https://github.com/vmware/terraform-provider-vcd/pull/1061))
* **New Data Source:** `vcd_nsxt_edgegateway_l2_vpn_tunnel` to read Edge Gateway L2 VPN Tunnel sessions ([#1061](https://github.com/vmware/terraform-provider-vcd/pull/1061))
* **New Resource:** `vcd_nsxt_edgegateway_dns` to manage Edge Gateway DNS forwarder configuration ([#1154](https://github.com/vmware/terraform-provider-vcd/pull/1154))
* **New Data Source:** `vcd_nsxt_edgegateway_dns` to read Edge Gateway DNS forwarder configuration ([#1154](https://github.com/vmware/terraform-provider-vcd/pull/1154))

### EXPERIMENTAL FEATURES
* New guide `Importing resources` on how to import resources with new experimental [Terraform import blocks](https://developer.hashicorp.com/terraform/language/import) ([#1104](https://github.com/vmware/terraform-provider-vcd/pull/1104))
* New example `Importing catalog contents` showing how to import shared catalogs ([#1104](https://github.com/vmware/terraform-provider-vcd/pull/1104))
* New example `Importing cloned vApps` showing how to import vApps and VMs from cloned vApps ([#1104](https://github.com/vmware/terraform-provider-vcd/pull/1104))
* **New Data Source:** `vcd_rde_behavior_invocation` to invoke a Behavior of a given RDE ([#1117](https://github.com/vmware/terraform-provider-vcd/pull/1117), [#1136](https://github.com/vmware/terraform-provider-vcd/pull/1136))
* **New Resource:** `vcd_vm_vgpu_policy` to manage VM vGPU compute policy configuration ([#1167](https://github.com/vmware/terraform-provider-vcd/pull/1167))
* **New Data Source:** `vcd_vm_vgpu_policy` to read VM vGPU compute policies ([#1167](https://github.com/vmware/terraform-provider-vcd/pull/1167))

### IMPROVEMENTS
* Add `metadata_entry` attribute to `vcd_rde` resource and data source to manage metadata entries of type
  `String`, `Number` and `Bool` in Runtime Defined Entities ([#1018](https://github.com/vmware/terraform-provider-vcd/pull/1018), [#1164](https://github.com/vmware/terraform-provider-vcd/pull/1164))
* Resource `vcd_catalog_access_control` adds property `read_only_shared_with_all_orgs` to share the catalog as read-only with all organizations ([#1020](https://github.com/vmware/terraform-provider-vcd/pull/1020))
* Resource and data source `vcd_org` add properties `number_of_vdcs`, `number_of_catalogs`, `list_of_vdcs`, `list_of_catalogs` ([#1020](https://github.com/vmware/terraform-provider-vcd/pull/1020))
* Resources `vcd_vapp_network` and `vcd_vapp_org_network` will additionally check if vApp is in
  `RESOLVED` (in addition to already checked `POWERED_OFF`) state before attempting a reboot when
  `reboot_vapp_on_removal` flag is set to `true` ([#1092](https://github.com/vmware/terraform-provider-vcd/pull/1092))
* Resource `vcd_vdc_group` supports force deletion using new parameter `force_delete` ([#1071](https://github.com/vmware/terraform-provider-vcd/pull/1071))
* Add fields `name_regex` and `import_file_name` to `vcd_resource_list` to facilitate creation of import blocks ([#1104](https://github.com/vmware/terraform-provider-vcd/pull/1104))
* Properties `delete_force` and `delete_recursive` in `vcd_org`, `vcd_org_vdc`, and `vcd_catalog` are now optional, to facilitate import operations ([#1104](https://github.com/vmware/terraform-provider-vcd/pull/1104))
* Properties `ova_path` and `ovf_url` in `vcd_catalog_item` and `vcd_vapp_template` are now optional, to facilitate import operations ([#1104](https://github.com/vmware/terraform-provider-vcd/pull/1104))
* Property `ova_path` in `vcd_catalog_media` is now optional, to facilitate import operations ([#1104](https://github.com/vmware/terraform-provider-vcd/pull/1104))
* Add field `ssl_enabled` to resource and data source `vcd_nsxt_alb_pool` to set SSL support on demand ([#1108](https://github.com/vmware/terraform-provider-vcd/pull/1108))
* Introduce new attributes `firmware` and `boot_options` to `vcd_vm` and `vcd_vapp_vm`, allowing to specify boot options of a VM (VCD 10.4.1+) ([#1109](https://github.com/vmware/terraform-provider-vcd/pull/1109))
* Resource and data source `vcd_nsxt_edgegateway` support attachment of NSX-T Segment backed
  External Networks via `external_network` block ([#1111](https://github.com/vmware/terraform-provider-vcd/pull/1111), [#1172](https://github.com/vmware/terraform-provider-vcd/pull/1172))
* Data source `vcd_resource_list` can now list network pools, vCenters, NSX-T transfer zones, distributed switches, and importable port groups ([#1115](https://github.com/vmware/terraform-provider-vcd/pull/1115))
* Data source `vcd_network_pool` includes all properties of the corresponding resource ([#1115](https://github.com/vmware/terraform-provider-vcd/pull/1115))
* Field `rde_type_id` from resource `vcd_rde` does not force a deletion when updated, to allow easier RDE Type version upgrades ([#1117](https://github.com/vmware/terraform-provider-vcd/pull/1117))
* Resource `vcd_rde_type` supports Behavior hooks with the new `hook` blocks, that allow to automatically invoke
  Behaviors on certain RDE lifecycle events ([#1122](https://github.com/vmware/terraform-provider-vcd/pull/1122))
* Data source `vcd_rde_type` supports reading Behavior hooks from VCD and store their information in the new `hook` blocks ([#1122](https://github.com/vmware/terraform-provider-vcd/pull/1122))
* Add property `upload_any_file` to resource `vcd_catalog_media` to allow uploading any file as catalog media item ([#1123](https://github.com/vmware/terraform-provider-vcd/pull/1123))
* Add property `download_to_file` to data source `vcd_catalog_media` to allow downloading a catalog media item into a file ([#1124](https://github.com/vmware/terraform-provider-vcd/pull/1124))
* Resource `vcd_provider_vdc` supports metadata with `metadata_entry` blocks ([#1126](https://github.com/vmware/terraform-provider-vcd/pull/1126))
* Resource and data source `vcd_catalog_vapp_template` add property `lease` with field `storage_lease_in_sec` to handle
  the VApp Template lease ([#1130](https://github.com/vmware/terraform-provider-vcd/pull/1130))
* Add property `custom_user_ou` to `vcd_org_ldap` to specify custom attributes when `ldap_mode = "SYSTEM"` ([#1142](https://github.com/vmware/terraform-provider-vcd/pull/1142))
* Add support to the metadata that gets automatically created on `vcd_vapp_vm` and `vcd_vm` when they are created by a VM from a vApp Template in VCD 10.5.1+,
  with the new `inherited_metadata` computed map. Example of metadata entries of this kind: `vm.origin.id`, `vm.origin.name`, `vm.origin.type` ([#1146](https://github.com/vmware/terraform-provider-vcd/pull/1146), [#1173](https://github.com/vmware/terraform-provider-vcd/pull/1173))
* Add support to the metadata that gets automatically created on `vcd_vapp` when it is created by a vApp Template or another vApp in VCD 10.5.1+,
  with the new `inherited_metadata` computed map. Example of metadata entries of this kind: `vapp.origin.id`, `vapp.origin.name`, `vapp.origin.type` ([#1146](https://github.com/vmware/terraform-provider-vcd/pull/1146), [#1173](https://github.com/vmware/terraform-provider-vcd/pull/1173))
* Add missing property `value` to `vcd_ip_space_ip_allocation` to specify IP or Prefix value on VCD
  10.4.2+  ([#1147](https://github.com/vmware/terraform-provider-vcd/pull/1147))
* Add `vcd_independent_disk` to the resources retrieved by `vcd_resource_list` ([#1155](https://github.com/vmware/terraform-provider-vcd/pull/1155))
* Resource and data source `vcd_ip_space` support NAT and Firewall creation configuration using
  fields `default_firewall_rule_creation_enabled`, `default_no_snat_rule_creation_enabled`,
  `default_snat_rule_creation_enabled` ([#1156](https://github.com/vmware/terraform-provider-vcd/pull/1156))

### BUG FIXES
* Minimize risk of latency-induced test failures in TestAccVcdSubscribedCatalog ([#1101](https://github.com/vmware/terraform-provider-vcd/pull/1101))
* Fix Issue [#1112](https://github.com/vmware/terraform-provider-vcd/issues/1112): Data source `vcd_vcenter` fails when name contains spaces ([#1115](https://github.com/vmware/terraform-provider-vcd/pull/1115))
* Fix a bug that made impossible to delete `vcd_rde_type_behavior_acl` resources when the Access Level is the last one
  in the Behavior ([#1117](https://github.com/vmware/terraform-provider-vcd/pull/1117))
* Fix the resource `vcd_rde_type_behavior_acl` to avoid race conditions when creating, updating or deleting more than one
  Access Level ([#1117](https://github.com/vmware/terraform-provider-vcd/pull/1117))
* Fix media item detection in `vcd_resource_list`: it was incorrectly listing also vApp templates ([#1119](https://github.com/vmware/terraform-provider-vcd/pull/1119))
* Fix Issue [#1127](https://github.com/vmware/terraform-provider-vcd/issues/1127) (Incorrect behavior of vcd_resource_list, which can retrieve Edge Gateways belonging to a VDC, but not belonging to a VDC Group) ([#1129](https://github.com/vmware/terraform-provider-vcd/pull/1129))
* Fix test TestAccVcdRightsContainers and expand it to test most available items ([#1135](https://github.com/vmware/terraform-provider-vcd/pull/1135))
* Fix a bug in `vcd_rde` that caused a RDE created in a certain Organization to be unreachable by a user
  belonging to a different Organization despite having the required rights ([#1139](https://github.com/vmware/terraform-provider-vcd/pull/1139), [#1164](https://github.com/vmware/terraform-provider-vcd/pull/1164))
* Fix organization retrieval in `vcd_resource_list` when users fill the `"parent"` field instead of `"org"` ([#1140](https://github.com/vmware/terraform-provider-vcd/pull/1140))
* Fix organization retrieval in `vcd_resource_list` when field `"org"` from the provider block was not used ([#1140](https://github.com/vmware/terraform-provider-vcd/pull/1140))
* Fix Issue [1134](https://github.com/vmware/terraform-provider-vcd/issues/1134) : Can't use SYSTEM `ldap_mode` ([#1142](https://github.com/vmware/terraform-provider-vcd/pull/1142))
* Fix a bug in `ignore_metadata_changes` provider configuration block when `conflict_action = warn`, that caused
  an operation to fail immediately instead of continuing without an error when a conflict was found ([#1164](https://github.com/vmware/terraform-provider-vcd/pull/1164), [#1173](https://github.com/vmware/terraform-provider-vcd/pull/1173))
* Fix usage example for datasource `vcd_nsxt_edgegateway_bgp_ip_prefix_list` in registry
  documentation ([#1169](https://github.com/vmware/terraform-provider-vcd/pull/1169))

### DEPRECATIONS
* Resource `vcd_org_vdc` deprecates `edge_cluster_id` in favor of new resource
  `vcd_org_vdc_nsxt_network_profile` that can configure NSX-T Edge Clusters and default Segment
  Profile Templates for NSX-T VDCs ([#1120](https://github.com/vmware/terraform-provider-vcd/pull/1120))

### NOTES
* Drop support for VCD 10.3.x ([#1108](https://github.com/vmware/terraform-provider-vcd/pull/1108))
* Add ability to split the test suite across several VCDs ([#1110](https://github.com/vmware/terraform-provider-vcd/pull/1110))
* Patch tests `TestAccVcdVAppVmCustomizationSettings` and
  `TestAccVcdStandaloneVmCustomizationSettings` to use valid guest customization settings ([#1113](https://github.com/vmware/terraform-provider-vcd/pull/1113))
* Improve `TestAccVcdIpv6Support` to avoid subnet clashes ([#1133](https://github.com/vmware/terraform-provider-vcd/pull/1133))
* Bump `terraform-plugin-sdk` to `v2.29.0` ([#1148](https://github.com/vmware/terraform-provider-vcd/pull/1148))
* Bump minimum Go requirement to `1.20` because such version is required by `terraform-plugin-sdk`
  `v2.29.0` ([#1148](https://github.com/vmware/terraform-provider-vcd/pull/1148))
* Improve documentation of the `provider_scoped` and `tenant_scoped` attributes from `vcd_ui_plugin` resource ([#1180](https://github.com/vmware/terraform-provider-vcd/pull/1180))

## 3.10.0 (July 20, 2023)

### FEATURES
* Add a **new guide** to create and manage Kubernetes Clusters using Container Service Extension v4.0 ([#1030](https://github.com/vmware/terraform-provider-vcd/pull/1030))
* **New Resource:** `vcd_nsxt_edgegateway_dhcp_forwarding` to manage NSX-T Edge Gateway DHCP Forwarding configuration ([#1056](https://github.com/vmware/terraform-provider-vcd/pull/1056))
* **New Data Source:** `vcd_nsxt_edgegateway_dhcp_forwarding` to read NSX-T Edge Gateway DHCP Forwarding configuration ([#1056](https://github.com/vmware/terraform-provider-vcd/pull/1056))
* **New Resource:** `vcd_ui_plugin` to programmatically install and manage UI Plugins ([#1059](https://github.com/vmware/terraform-provider-vcd/pull/1059))
* **New Data Source:** `vcd_ui_plugin` to fetch existing UI Plugins ([#1059](https://github.com/vmware/terraform-provider-vcd/pull/1059))
* **New Resource:** `vcd_ip_space` to manage IP Spaces in VCD 10.4.1+ ([#1061](https://github.com/vmware/terraform-provider-vcd/pull/1061))
* **New Data Source:** `vcd_ip_space` to read IP Spaces in VCD 10.4.1+ ([#1061](https://github.com/vmware/terraform-provider-vcd/pull/1061))
* **New Resource:** `vcd_ip_space_uplink` to manage IP Space Uplinks for External Networks (Provider
  gateways) in VCD 10.4.1+ ([#1062](https://github.com/vmware/terraform-provider-vcd/pull/1062))
* **New Data Source:** `vcd_ip_space_uplink` to read IP Space Uplinks for External Networks
  (Provider gateways) in VCD 10.4.1+ ([#1062](https://github.com/vmware/terraform-provider-vcd/pull/1062))
* **New Resource:** `vcd_ip_space_ip_allocation` to manage IP Space IP Allocations in VCD 10.4.1+
  ([#1062](https://github.com/vmware/terraform-provider-vcd/pull/1062))
* **New Data Source:** `vcd_ip_space_ip_allocation` to read IP Space IP Allocations in VCD 10.4.1+
  ([#1062](https://github.com/vmware/terraform-provider-vcd/pull/1062))
* **New Resource:** `vcd_ip_space_custom_quota` to manage Custom IP Space Quotas for individual
  Organizations in VCD 10.4.1+ ([#1062](https://github.com/vmware/terraform-provider-vcd/pull/1062))
* **New Data Source:** `vcd_ip_space_custom_quota` to read Custom IP Space Quotas for individual
  Organizations in VCD 10.4.1+ ([#1062](https://github.com/vmware/terraform-provider-vcd/pull/1062))
* **New Resource:** `vcd_org_saml` to manage an organization SAML configuration ([#1064](https://github.com/vmware/terraform-provider-vcd/pull/1064))
* **New Data Source:** `vcd_org_saml` to read an organization SAML configuration ([#1064](https://github.com/vmware/terraform-provider-vcd/pull/1064))
* **New Data Source:** `vcd_org_saml_metadata` to read an organization SAML service provider metadata ([#1064](https://github.com/vmware/terraform-provider-vcd/pull/1064))
* **New Resource:** `vcd_api_token` to manage API tokens ([#1070](https://github.com/vmware/terraform-provider-vcd/pull/1070))
* **New Resource:** `vcd_service_account` to manage Service Accounts ([#1070](https://github.com/vmware/terraform-provider-vcd/pull/1070))
* **New Data Source:** `vcd_service_account` to read Service Accounts ([#1070](https://github.com/vmware/terraform-provider-vcd/pull/1070))
* **New Resource:** `vcd_nsxt_edgegateway_dhcpv6` to manage NSX-T Edge Gateway DHCPv6 configuration
  ([#1071](https://github.com/vmware/terraform-provider-vcd/pull/1071),[#1083](https://github.com/vmware/terraform-provider-vcd/pull/1083))
* **New Data Source:** `vcd_nsxt_edgegateway_dhcpv6` to read NSX-T Edge Gateway DHCPv6 configuration
  ([#1071](https://github.com/vmware/terraform-provider-vcd/pull/1071),[#1083](https://github.com/vmware/terraform-provider-vcd/pull/1083))
* **New Resource:** `vcd_provider_vdc` to manage provider VDCs ([#1073](https://github.com/vmware/terraform-provider-vcd/pull/1073))
* **New Data Source:** `vcd_resource_pool` to read vCenter Resource Pools ([#1073](https://github.com/vmware/terraform-provider-vcd/pull/1073))
* **New Data Source:** `vcd_network_pool` to read Network Pools ([#1073](https://github.com/vmware/terraform-provider-vcd/pull/1073))
* **New Resource:** `vcd_rde_interface_behavior` to manage RDE Interface Behaviors, which can be invoked by RDEs and
  overridden by RDE Types ([#1074](https://github.com/vmware/terraform-provider-vcd/pull/1074))
* **New Data Source:** `vcd_rde_interface_behavior` to read RDE Interface Behaviors, so they can be used
  in RDE Type Behavior overrides ([#1074](https://github.com/vmware/terraform-provider-vcd/pull/1074))
* **New Resource:** `vcd_rde_type_behavior` to manage Behaviors in RDE Types, which can override those defined
  in RDE Interfaces ([#1074](https://github.com/vmware/terraform-provider-vcd/pull/1074))
* **New Data Source:** `vcd_rde_type_behavior` to read RDE Type Behaviors ([#1074](https://github.com/vmware/terraform-provider-vcd/pull/1074))
* **New Resource:** `vcd_rde_type_behavior_acl` to manage the access to Behaviors in RDE Types and RDE Interfaces ([#1074](https://github.com/vmware/terraform-provider-vcd/pull/1074))
* **New Data Source:** `vcd_rde_type_behavior_acl` to read Access Levels from Behaviors of RDE Types and RDE Interfaces ([#1074](https://github.com/vmware/terraform-provider-vcd/pull/1074))
* **New Resource:** `vcd_nsxt_edgegateway_static_route` to manage NSX-T Edge Gateway Static Routes
  on VCD 10.4.0+ ([#1075](https://github.com/vmware/terraform-provider-vcd/pull/1075))
* **New Data Source:** `vcd_nsxt_edgegateway_static_route` to read NSX-T Edge Gateway Static Routes
  on VCD 10.4.0+ ([#1075](https://github.com/vmware/terraform-provider-vcd/pull/1075))
* **New Resource:** `vcd_nsxt_distributed_firewall_rule` to manage NSX-T Distributed Firewall one by
  one. Rules will *not be created in parallel* because the API provides no direct endpoint to create
  a single rule and this functionality uses a custom-made function that abstracts the "update all"
  endpoint ([#1076](https://github.com/vmware/terraform-provider-vcd/pull/1076))
* **New Data Source:** `vcd_nsxt_distributed_firewall_rule` to read NSX-T Distributed Firewall one
  by one ([#1076](https://github.com/vmware/terraform-provider-vcd/pull/1076))
* **New Resource:** `vcd_cloned_vapp` to create a vApp from either a vApp template or another vApp ([#1081](https://github.com/vmware/terraform-provider-vcd/pull/1081))

### EXPERIMENTAL
(_Experimental features and improvements may change in future releases, until declared stable._)
* Add `ignore_metadata_changes` argument to the Provider configuration to be able to specify metadata entries that should **not**
  be managed by Terraform when using `metadata_entry` configuration blocks ([#1057](https://github.com/vmware/terraform-provider-vcd/pull/1057), [#1089](https://github.com/vmware/terraform-provider-vcd/pull/1089))

### IMPROVEMENTS
* The guide to install the Container Service Extension v4.0 now additionally explains how to install the
  Kubernetes Container Clusters UI Plugin ([#1059](https://github.com/vmware/terraform-provider-vcd/pull/1059))
* `vcd_external_network_v2` resource and data source support IP Spaces on VCD 10.4.1+ by adding
  `use_ip_spaces` and `dedicated_org_id` fields ([#1062](https://github.com/vmware/terraform-provider-vcd/pull/1062))
* `vcd_nsxt_edgegateway` resource supports IP Spaces by not requiring `subnet` specification
  ([#1062](https://github.com/vmware/terraform-provider-vcd/pull/1062))
* Resource and data source `vcd_nsxt_alb_virtual_service` support IPv6 on VCD 10.4.0+ via new field
  `ipv6_virtual_ip_address` ([#1071](https://github.com/vmware/terraform-provider-vcd/pull/1071))
* Resource and data source `vcd_network_routed_v2` support Dual-Stack mode using
  `dual_stack_enabled` and `secondary_gateway`, `secondary_prefix_length`,
  `secondary_static_ip_pool` fields ([#1071](https://github.com/vmware/terraform-provider-vcd/pull/1071))
* Resource and data source `vcd_network_isolated_v2` support Dual-Stack mode using
  `dual_stack_enabled` and `secondary_gateway`, `secondary_prefix_length`,
  `secondary_static_ip_pool` fields ([#1071](https://github.com/vmware/terraform-provider-vcd/pull/1071))
* Resource and data source `vcd_nsxt_network_imported` support Dual-Stack mode using
  `dual_stack_enabled` and `secondary_gateway`, `secondary_prefix_length`,
  `secondary_static_ip_pool` fields ([#1071](https://github.com/vmware/terraform-provider-vcd/pull/1071))
* Resource and data source `vcd_nsxt_network_dhcp_binding` support `dhcp_v6_config` config ([#1071](https://github.com/vmware/terraform-provider-vcd/pull/1071))
* Validate possibility to perform end to end IPv6 configuration via additional tests ([#1071](https://github.com/vmware/terraform-provider-vcd/pull/1071))
* Resource `vcd_vdc_group` adds new field `remove_default_firewall_rule` to remove default
  Distributed Firewall Rule after creation ([#1076](https://github.com/vmware/terraform-provider-vcd/pull/1076))
* The attribute `description` of `vcd_vm_placement_policy` is now Computed, as latest VCD versions set a default description
automatically if it is not set ([#1082](https://github.com/vmware/terraform-provider-vcd/pull/1082))

### BUG FIXES
* Fix [Issue #1058](https://github.com/vmware/terraform-provider-vcd/issues/1058) - Multiple `SYSTEM` scope data source `vcd_nsxt_app_port_profile` when multiple NSX-T managers are configured ([#1065](https://github.com/vmware/terraform-provider-vcd/pull/1065))
* Fix [Issue #1069](https://github.com/vmware/terraform-provider-vcd/issues/1069) - Inconsistent `security_profile_customization` field during `vcd_nsxt_ipsec_vpn_tunnel` update ([#1072](https://github.com/vmware/terraform-provider-vcd/pull/1072))
* Fix [Issue #1066](https://github.com/vmware/terraform-provider-vcd/issues/1066) - Not possible to handle more than 128 storage profiles ([#1073](https://github.com/vmware/terraform-provider-vcd/pull/1073))
* Fix a bug that caused `vcd_catalog` creation to fail if it is created with deprecated `metadata` argument in VCD 10.5  ([#1085](https://github.com/vmware/terraform-provider-vcd/pull/1085))

### NOTES
* Internal - replaces `takeBoolPointer`, `takeIntPointer`, `takeInt64Pointer` with generic `addrOf`
  ([#1055](https://github.com/vmware/terraform-provider-vcd/pull/1055))
* Bump `terraform-plugin-sdk` to v2.27.0 ([#1079](https://github.com/vmware/terraform-provider-vcd/pull/1079))
* Resource `vcd_nsxt_edgegateway_bgp_configuration` will send existing `GracefulRestart` to avoid
  API validation errors in VCD 10.5.0+ ([#1083](https://github.com/vmware/terraform-provider-vcd/pull/1083))
* [`go-vcloud-director`](https://github.com/vmware/go-vcloud-director), the SDK this provider uses for low level access to the VCD, released with version v2.21.0

## 3.9.0 (April 27, 2023)

### FEATURES
* New guide to install **Container Service Extension (CSE)** v4.0 in VCD 10.4+ ([#1003](https://github.com/vmware/terraform-provider-vcd/pull/1003), [#1053](https://github.com/vmware/terraform-provider-vcd/pull/1053))
* **New Resource:** `vcd_rde_interface` to manage Runtime Defined Entity Interfaces
  which are required for using Runtime Defined Entity (RDE) types ([#965](https://github.com/vmware/terraform-provider-vcd/pull/965))
* **New Data Source:** `vcd_rde_interface` to fetch existing Runtime Defined Entity Interfaces ([#965](https://github.com/vmware/terraform-provider-vcd/pull/965))
* **New Resource:** `vcd_rde_type` to manage Runtime Defined Entity Types
  which are required for using Runtime Defined Entities (RDEs) ([#973](https://github.com/vmware/terraform-provider-vcd/pull/973))
* **New Data Source:** `vcd_rde_type` to fetch existing Runtime Defined Entity Types ([#973](https://github.com/vmware/terraform-provider-vcd/pull/973))
* **New Resource:** `vcd_rde` to manage Runtime Defined Entities ([#977](https://github.com/vmware/terraform-provider-vcd/pull/977))
* **New Data Source:** `vcd_rde` to fetch existing Runtime Defined Entities ([#977](https://github.com/vmware/terraform-provider-vcd/pull/977))
* **New Resource:** `vcd_nsxv_distributed_firewall` to create and manage NSX-V distributed firewall ([#988](https://github.com/vmware/terraform-provider-vcd/pull/988))
* **New Data Source:** `vcd_nsxv_distributed_firewall` to fetch existing NSX-V distributed firewall ([#988](https://github.com/vmware/terraform-provider-vcd/pull/988))
* **New Data Source:** `vcd_nsxv_application_finder` to search applications and application groups to use with a NSX-V distributed firewall ([#988](https://github.com/vmware/terraform-provider-vcd/pull/988))
* **New Data Source:** `vcd_nsxv_application` to fetch existing application to use with a NSX-V distributed firewall ([#988](https://github.com/vmware/terraform-provider-vcd/pull/988))
* **New Data Source:** `vcd_nsxv_application_group` to fetch existing application_group to use with a NSX-V distributed firewall ([#988](https://github.com/vmware/terraform-provider-vcd/pull/988))
* **New Resource:** `vcd_nsxt_network_dhcp_binding` to manage NSX-T DHCP Bindings ([#1039](https://github.com/vmware/terraform-provider-vcd/pull/1039))
* **New Data Source:** `vcd_nsxt_network_dhcp_binding` to read NSX-T DHCP Bindings ([#1039](https://github.com/vmware/terraform-provider-vcd/pull/1039))
* **New Resource:** `vcd_nsxt_edgegateway_rate_limiting` to manage NSX-T Edge Gateway Rate Limiting ([#1042](https://github.com/vmware/terraform-provider-vcd/pull/1042))
* **New Data Source:** `vcd_nsxt_edgegateway_rate_limiting` to read NSX-T Edge Gateway Rate Limiting ([#1042](https://github.com/vmware/terraform-provider-vcd/pull/1042))
* **New Data Source:** `vcd_nsxt_edgegateway_qos_profile` to read QoS profiles available for
  `vcd_nsxt_edgegateway_rate_limiting` resource ([#1042](https://github.com/vmware/terraform-provider-vcd/pull/1042))

### IMPROVEMENTS
* `vcd_external_network_v2` allows setting DNS fields `dns1`, `dns2` and `dns_suffix` for NSX-T
  backed entities so that it can be inherited by direct Org VDC networks ([#984](https://github.com/vmware/terraform-provider-vcd/pull/984))
* `vcd_org_vdc` includes a property `enable_nsxv_distributed_firewall` to enable or disable a NSX-V distributed firewall ([#988](https://github.com/vmware/terraform-provider-vcd/pull/988))
* `vcd_nsxt_edgegateway` resource and data source got automatic IP allocation support using new
  configuration fields `subnet_with_total_ip_count`, `subnet_with_ip_count` and `total_allocated_ip_count` fields ([#991](https://github.com/vmware/terraform-provider-vcd/pull/991))
* `vcd_nsxt_edgegateway` resource and data source expose `used_ip_count` and `unused_ip_count`
  attributes ([#991](https://github.com/vmware/terraform-provider-vcd/pull/991), [#1047](https://github.com/vmware/terraform-provider-vcd/pull/1047))
* `vcd_nsxt_alb_settings` resource and data source adds two new fields `is_transparent_mode_enabled`
  and `ipv6_service_network_specification` ([#996](https://github.com/vmware/terraform-provider-vcd/pull/996))
* Resources `vcd_vapp_network` and `vcd_vapp_org_network` add convenience flag
  `reboot_vapp_on_removal`. When enabled, it will power off parent vApp (and power back on after
  if it was before) during vApp network removal. This improves workflows with VCD 10.4.1+ which
  returns an error when removing vApp networks from powered on vApps ([#1004](https://github.com/vmware/terraform-provider-vcd/pull/1004))
* `vcd_vapp_vm` and `vcd_vm` resources support security tag management via new field `security_tags` ([#1006](https://github.com/vmware/terraform-provider-vcd/pull/1006), [#1046](https://github.com/vmware/terraform-provider-vcd/pull/1046))
* Resource `vcd_nsxt_ipsec_vpn_tunnel` adds support for custom `remote_id` field and certificate
  based auth via fields `authentication_mode`, `certificate_id`, `ca_certificate_id` ([#1010](https://github.com/vmware/terraform-provider-vcd/pull/1010))
* `vcd_org_group` adds `OAUTH` as an option to argument `provider_type` ([#1013](https://github.com/vmware/terraform-provider-vcd/pull/1013))
* Resource and data source `vcd_nsxt_alb_virtual_service` add support for Transparent mode in VCD
  10.4.1+ via field `is_transparent_mode_enabled` ([#1024](https://github.com/vmware/terraform-provider-vcd/pull/1024))
* Resource and data source `vcd_nsxt_alb_pool` add support for Pool Group Membership via field
  `member_group_id` ([#1024](https://github.com/vmware/terraform-provider-vcd/pull/1024))
* Resource and data source `vcd_nsxt_network_imported` support Distributed Virtual Port Group (DVPG)
  backed Org VDC network ([#1043](https://github.com/vmware/terraform-provider-vcd/pull/1043))
* Support provider authentication using Active Service Accounts ([#1040](https://github.com/vmware/terraform-provider-vcd/pull/1040))

### BUG FIXES
* Fix a bug that prevented returning a specific error while authenticating provider with invalid
  password ([#962](https://github.com/vmware/terraform-provider-vcd/pull/962))
* Add `prefix_length` field to `vcd_vapp_network` as creating IPv6 vApp networks was not supported due to the lack of a suitable subnet representation (Issue #999) ([#1007](https://github.com/vmware/terraform-provider-vcd/pull/1007), [#1031](https://github.com/vmware/terraform-provider-vcd/pull/1031))
* Remove incorrect default value from `vcd_vapp_network` `netmask` field, as it prevents using IPv6 networks. Users of already defined resources need to add a `netmask = "255.255.255.0"` when using IPv4 ([#1007](https://github.com/vmware/terraform-provider-vcd/pull/1007))

### DEPRECATIONS
* Deprecate `netmask` in favor of `prefix_length` for `vcd_vapp_network` ([#1007](https://github.com/vmware/terraform-provider-vcd/pull/1007))

### NOTES
* Add missing test name fields for `TestAccVcdNsxtEdgeBgpConfigIntegrationVdc` and
  `TestAccVcdNsxtEdgeBgpConfigIntegrationVdcGroup` ([#958](https://github.com/vmware/terraform-provider-vcd/pull/958))
* Create `TestAccVcdCatalogRename`, which checks that renaming a catalog works correctly ([#992](https://github.com/vmware/terraform-provider-vcd/pull/992))
* Removed disk update steps from `TestAccVcdIndependentDiskBasic`, as it was sometimes failing due to a bug in VCD. Created a new one `TestAccVcdIndependentDiskBasicWithUpdates` which will be run only on new releases of VCD (>=v10.4.1) ([#1014](https://github.com/vmware/terraform-provider-vcd/pull/1014))
* Increased sleep in between testing steps in `TestAccVcdNsxtDynamicSecurityGroupVdcGroupCriteriaWithVms` from 15s to 25s to let VMs get created ([#1014](https://github.com/vmware/terraform-provider-vcd/pull/1014))
* Added skipping of `TestAccVcdVsphereSubscriber` and `TestAccVcdSubscribedCatalog` if VCD version is older than v10.4.0 as there was a bug with catalog sharing rights that caused the tests to fail ([#1014](https://github.com/vmware/terraform-provider-vcd/pull/1014))
* Update `CODING_GUIDELINES.md` with documentation notes ([#1015](https://github.com/vmware/terraform-provider-vcd/pull/1015))
* Bump `terraform-plugin-sdk` to v2.26.1 ([#1002](https://github.com/vmware/terraform-provider-vcd/pull/1002), [#1023](https://github.com/vmware/terraform-provider-vcd/pull/1023))
* Bump `golang.org/x/net` to v0.7.0 to address vulnerability reports ([#1002](https://github.com/vmware/terraform-provider-vcd/pull/1002))
* Add support for Go 1.20 in testing workflows ([#1034](https://github.com/vmware/terraform-provider-vcd/pull/1034))
* Bump `staticcheck` to 2023.1.3 ([#1034](https://github.com/vmware/terraform-provider-vcd/pull/1034))

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
